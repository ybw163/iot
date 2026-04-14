package main

import (
	"context"
	"fmt"
	"syscall"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"iot/internal/handler"
	"iot/internal/model"
	"iot/internal/repository"
	"iot/internal/router"
	"iot/internal/service"
	"iot/internal/stream"
	"iot/pkg/config"
	"iot/pkg/database"
	"iot/pkg/logger"
	mqemqx "iot/internal/mq/emqx"
	mqkafka "iot/internal/mq/kafka"
	pkgredis "iot/pkg/redis"
)

func main() {
	// 加载配置
	if err := config.Load("config/config.yaml"); err != nil {
		panic(fmt.Sprintf("load config failed: %v", err))
	}

	// 初始化日志
	if err := logger.Init(config.Conf.Log); err != nil {
		panic(fmt.Sprintf("init logger failed: %v", err))
	}
	defer logger.Log.Sync()

	// 初始化数据库
	if err := database.Init(config.Conf.Database); err != nil {
		logger.Log.Fatal("init database failed", zap.Error(err))
	}
	defer database.Close()

	// 自动迁移
	if err := database.DB.AutoMigrate(&model.Device{}); err != nil {
		logger.Log.Fatal("auto migrate failed", zap.Error(err))
	}

	// 初始化 Redis
	if err := pkgredis.Init(config.Conf.Redis); err != nil {
		logger.Log.Fatal("init redis failed", zap.Error(err))
	}
	defer pkgredis.Close()

	// 初始化 Kafka Producer
	if err := mqkafka.InitProducer(config.Conf.Kafka); err != nil {
		logger.Log.Fatal("init kafka producer failed", zap.Error(err))
	}
	defer mqkafka.CloseProducer()

	// 初始化 Kafka Consumer
	if err := mqkafka.InitConsumer(config.Conf.Kafka); err != nil {
		logger.Log.Fatal("init kafka consumer failed", zap.Error(err))
	}
	defer mqkafka.CloseConsumer()

	// 初始化各层
	deviceRepo := repository.NewDeviceRepository(database.DB)
	deviceService := service.NewDeviceService(deviceRepo)
	deviceHandler := handler.NewDeviceHandler(deviceService)

	// 构建流处理管道（类似 Flink DAG）
	pipeline := stream.NewPipeline()
	pipeline.
		AddProcessor(stream.NewDatabaseSink(deviceRepo)).    // Sink: 更新数据库
		AddProcessor(stream.NewRedisCacheSink()).             // Sink: 写入 Redis 缓存
		AddProcessor(stream.NewRiskControlProcessor())        // 算子: 风控检测

	// 启动 Kafka Consumer 订阅，消息送入流处理管道
	kafkaTopic := config.Conf.Kafka.Topic
	if err := mqkafka.DefaultConsumer.Subscribe(context.Background(), kafkaTopic, stream.HandleKafkaMessage(pipeline)); err != nil {
		logger.Log.Fatal("subscribe kafka topic failed", zap.Error(err))
	}
	logger.Log.Info("kafka consumer subscribed with stream pipeline",
		zap.String("topic", kafkaTopic),
		zap.Int("pipeline_stages", 3),
	)

	// 初始化 EMQX
	if err := mqemqx.Init(config.Conf.EMQX); err != nil {
		logger.Log.Fatal("init emqx failed", zap.Error(err))
	}
	defer mqemqx.Close()

	// 启动 EMQX 消费者（订阅 TBOX 遥测消息）
	if err := mqemqx.StartConsumer(); err != nil {
		logger.Log.Fatal("start emqx consumer failed", zap.Error(err))
	}

	// 设置路由
	gin.SetMode(config.Conf.Server.Mode)
	engine := gin.New()

	r := router.NewRouter(deviceHandler)
	r.Setup(engine)

	// 启动服务
	addr := fmt.Sprintf(":%d", config.Conf.Server.Port)
	logger.Log.Info("server starting", zap.String("addr", addr))

	if err := engine.Run(addr); err != nil && err != syscall.EINTR {
		logger.Log.Fatal("server start failed", zap.Error(err))
	}
}
