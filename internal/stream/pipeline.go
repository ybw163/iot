package stream

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/IBM/sarama"
	"go.uber.org/zap"

	"iot/internal/pb"
	"iot/internal/repository"
	"iot/pkg/logger"
	pkgredis "iot/pkg/redis"
)

// Processor 流处理器接口（类似 Flink 的算子）
type Processor interface {
	Name() string
	Process(ctx context.Context, msg *pb.TelemetryMessage) error
}

// Pipeline 流处理管道（类似 Flink 的 DAG 管道）
type Pipeline struct {
	processors []Processor
}

func NewPipeline() *Pipeline {
	return &Pipeline{}
}

// AddProcessor 添加处理器到管道
func (p *Pipeline) AddProcessor(proc Processor) *Pipeline {
	p.processors = append(p.processors, proc)
	return p
}

// Execute 执行管道中所有处理器，任一步失败不影响后续步骤（容错）
func (p *Pipeline) Execute(ctx context.Context, msg *pb.TelemetryMessage) {
	for _, proc := range p.processors {
		if err := proc.Process(ctx, msg); err != nil {
			logger.Log.Error("pipeline processor failed",
				zap.String("processor", proc.Name()),
				zap.String("vin", msg.GetVin()),
				zap.Error(err),
			)
		}
	}
}

// --- 数据库 Sink 算子 ---

// DatabaseSink 数据库写入处理器（类似 Flink JDBC Sink）
type DatabaseSink struct {
	repo *repository.DeviceRepository
}

func NewDatabaseSink(repo *repository.DeviceRepository) *DatabaseSink {
	return &DatabaseSink{repo: repo}
}

func (s *DatabaseSink) Name() string { return "DatabaseSink" }

func (s *DatabaseSink) Process(ctx context.Context, msg *pb.TelemetryMessage) error {
	updates := map[string]interface{}{
		"power":  msg.GetPower(),
		"speed":  msg.GetSpeed(),
		"status": msg.GetStatus(),
		"lat":    fmt.Sprintf("%.6f", msg.GetLat()),
		"lon":    fmt.Sprintf("%.6f", msg.GetLon()),
	}

	// 根据 VIN 查找设备并更新
	device, err := s.repo.GetByVin(ctx, msg.GetVin())
	if err != nil {
		return fmt.Errorf("find device by VIN %s failed: %w", msg.GetVin(), err)
	}

	if err := s.repo.UpdateFields(ctx, device.ID, updates); err != nil {
		return fmt.Errorf("update device %d failed: %w", device.ID, err)
	}

	logger.Log.Info("[DatabaseSink] device updated",
		zap.String("vin", msg.GetVin()),
		zap.Uint("device_id", device.ID),
		zap.Int32("speed", msg.GetSpeed()),
		zap.Int32("power", msg.GetPower()),
	)
	return nil
}

// --- Redis Cache Sink 算子 ---

// RedisCacheSink Redis缓存处理器（类似 Flink Redis Sink）
type RedisCacheSink struct{}

func NewRedisCacheSink() *RedisCacheSink {
	return &RedisCacheSink{}
}

func (s *RedisCacheSink) Name() string { return "RedisCacheSink" }

func (s *RedisCacheSink) Process(ctx context.Context, msg *pb.TelemetryMessage) error {
	key := fmt.Sprintf("device:telemetry:%s", msg.GetVin())

	cacheData := map[string]interface{}{
		"vin":       msg.GetVin(),
		"power":     msg.GetPower(),
		"speed":     msg.GetSpeed(),
		"status":    msg.GetStatus(),
		"lat":       msg.GetLat(),
		"lon":       msg.GetLon(),
		"timestamp": msg.GetTimestamp(),
		"updated":   time.Now().Unix(),
	}

	if err := pkgredis.Client.HSet(ctx, key, cacheData).Err(); err != nil {
		return fmt.Errorf("redis HSet %s failed: %w", key, err)
	}

	// 设置过期时间 30 分钟
	if err := pkgredis.Client.Expire(ctx, key, 30*time.Minute).Err(); err != nil {
		return fmt.Errorf("redis Expire %s failed: %w", key, err)
	}

	logger.Log.Info("[RedisCacheSink] cache updated",
		zap.String("vin", msg.GetVin()),
		zap.String("redis_key", key),
	)
	return nil
}

// --- 风控算子 ---

// RiskControlProcessor 风控处理器（类似 Flink CEP 复杂事件处理）
type RiskControlProcessor struct {
	speedThreshold int32
	powerThreshold int32
}

func NewRiskControlProcessor() *RiskControlProcessor {
	return &RiskControlProcessor{
		speedThreshold: 120, // 超速阈值 km/h
		powerThreshold: 80,  // 过载功率阈值
	}
}

func (s *RiskControlProcessor) Name() string { return "RiskControlProcessor" }

func (s *RiskControlProcessor) Process(ctx context.Context, msg *pb.TelemetryMessage) error {
	var alerts []string

	// 规则1：超速检测
	if msg.GetSpeed() > s.speedThreshold {
		alerts = append(alerts, fmt.Sprintf("OVERSPEED: speed=%d > threshold=%d", msg.GetSpeed(), s.speedThreshold))
	}

	// 规则2：过载检测
	if msg.GetPower() > s.powerThreshold {
		alerts = append(alerts, fmt.Sprintf("OVERLOAD: power=%d > threshold=%d", msg.GetPower(), s.powerThreshold))
	}

	// 规则3：异常状态检测（离线但仍在上报数据）
	if msg.GetStatus() == int32(pb.StatusOffline) {
		alerts = append(alerts, fmt.Sprintf("ANOMALY: offline device reporting data, status=%d", msg.GetStatus()))
	}

	// 规则4：坐标异常检测（经纬度全为0）
	if msg.GetLat() == 0 && msg.GetLon() == 0 {
		alerts = append(alerts, "ANOMALY: invalid GPS coordinates (0,0)")
	}

	if len(alerts) > 0 {
		// 将告警写入 Redis 告警列表
		alertKey := fmt.Sprintf("device:alerts:%s", msg.GetVin())
		for _, alert := range alerts {
			alertRecord := map[string]interface{}{
				"vin":       msg.GetVin(),
				"alert":     alert,
				"timestamp": strconv.FormatUint(msg.GetTimestamp(), 10),
				"created":   time.Now().Format(time.RFC3339),
			}
			data, _ := json.Marshal(alertRecord)
			if err := pkgredis.Client.RPush(ctx, alertKey, data).Err(); err != nil {
				logger.Log.Error("[RiskControlProcessor] write alert to redis failed",
					zap.String("vin", msg.GetVin()),
					zap.Error(err),
				)
			}
		}
		// 告警列表保留最近 1000 条
		//pkgredis.Client.LTrim(ctx, alertKey, -1000, -1)
		//pkgredis.Client.Expire(ctx, alertKey, 24*time.Hour)

		logger.Log.Warn("[RiskControlProcessor] risk alert triggered",
			zap.String("vin", msg.GetVin()),
			zap.Strings("alerts", alerts),
		)
	}

	return nil
}

// --- Kafka Source 函数 ---

// HandleKafkaMessage 将 Kafka 消息解码并送入管道处理
func HandleKafkaMessage(pipeline *Pipeline) func(msg *sarama.ConsumerMessage) error {
	return func(msg *sarama.ConsumerMessage) error {
		telemetry := &pb.TelemetryMessage{}
		if err := json.Unmarshal(msg.Value, telemetry); err != nil {
			logger.Log.Error("[Pipeline] decode kafka message failed",
				zap.String("topic", msg.Topic),
				zap.Int64("offset", msg.Offset),
				zap.Error(err),
			)
			return err
		}

		// 送入管道处理
		pipeline.Execute(context.Background(), telemetry)
		return nil
	}
}
