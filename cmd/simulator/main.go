package main

import (
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"
	"encoding/json"
	"iot/internal/mq/emqx"
	"iot/internal/pb"
	"iot/pkg/config"
	"iot/pkg/logger"

	"google.golang.org/protobuf/proto"
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

	// 初始化 EMQX 客户端
	// 模拟器使用独立的 client_id
	simCfg := config.Conf.EMQX
	simCfg.ClientID = "tbox_simulator"

	if err := emqx.Init(simCfg); err != nil {
		logger.Log.Fatal("init emqx failed", zap.Error(err))
	}
	defer emqx.Close()

	// 模拟 TBOX 参数
	vins := []string{
		"LSVAU2180N2012345",
		"LSVAU2180N2012346",
		"LSVAU2180N2012347",
	}

	// 启动定时发送
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	// 立即发送一次
	sendTelemetry(vins)

	logger.Log.Info("tbox simulator started", zap.Int("vehicle_count", len(vins)), zap.Duration("interval", 5*time.Second))

	// 等待退出信号
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	for {
		select {
		case <-ticker.C:
			sendTelemetry(vins)
		case sig := <-sigCh:
			logger.Log.Info("tbox simulator stopping", zap.String("signal", sig.String()))
			return
		}
	}
}

func sendTelemetry(vins []string) {
	for _, vin := range vins {
		msg := &pb.TelemetryMessage{
			Version:   1,
			MsgType:   pb.MsgTypeTelemetry,
			Vin:       vin,
			Power:     rand.Int31n(100),
			Speed:     rand.Int31n(120),
			Status:    pb.StatusOnline,
			Lat:       31.2304 + rand.Float64()*0.01,
			Lon:       121.4737 + rand.Float64()*0.01,
			Timestamp: uint64(time.Now().UnixMilli()),
		}

		payload, err := proto.Marshal(msg)
		if err != nil {
			logger.Log.Error("marshal protobuf failed", zap.String("vin", vin), zap.Error(err))
			continue
		}

		topic := "tbox_telemetry"
		if err := emqx.DefaultClient.PublishBytes(topic, payload); err != nil {
			logger.Log.Error("publish telemetry failed", zap.String("vin", vin), zap.Error(err))
			continue
		}
		data, _ := json.Marshal(msg)
		logger.Log.Info("sent tbox telemetry",
			zap.String("topic", topic),
			zap.String("vin", vin),
			zap.String("data", string(data)),
		)
	}
}
