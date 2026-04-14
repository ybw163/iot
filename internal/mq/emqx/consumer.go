package emqx

import (
	"context"
	"fmt"

	"encoding/json"
	"iot/internal/mq/kafka"
	"iot/internal/pb"
	"iot/pkg/config"
	"iot/pkg/logger"

	pahomqtt "github.com/eclipse/paho.mqtt.golang"
	"go.uber.org/zap"

	"google.golang.org/protobuf/proto"
)

// StartConsumer 启动 EMQX 消费者，订阅 TBOX 遥测主题并解码 Protobuf 消息
func StartConsumer() error {
	if DefaultClient == nil {
		return fmt.Errorf("emqx client not initialized")
	}

	topic := DefaultClient.cfg.Topic
	if topic == "" {
		topic = "tbox_telemetry"
	}

	handler := func(client pahomqtt.Client, msg pahomqtt.Message) {
		// 解码 Protobuf 消息
		telemetry := &pb.TelemetryMessage{}
		if err := proto.Unmarshal(msg.Payload(), telemetry); err != nil {
			logger.Log.Error("decode protobuf message failed",
				zap.String("topic", msg.Topic()),
				zap.Int("payload_len", len(msg.Payload())),
				zap.Error(err),
			)
			return
		}
		data, _ := json.Marshal(telemetry)

		// 转发消息到 Kafka
		kafkaTopic := config.Conf.Kafka.Topic
		if kafka.DefaultProducer == nil {
			logger.Log.Error("kafka producer not initialized, skip forward")
			return
		}
		_, _, err := kafka.DefaultProducer.SendSync(context.Background(), kafkaTopic, data)
		if err != nil {
			logger.Log.Error("forward to kafka failed",
				zap.String("kafka_topic", kafkaTopic),
				zap.Error(err),
			)
			return
		}
		logger.Log.Info("forwarded tbox telemetry to kafka",
			zap.String("mqtt_topic", msg.Topic()),
			zap.String("kafka_topic", kafkaTopic),
		)
	}

	if err := DefaultClient.Subscribe(topic, handler); err != nil {
		return fmt.Errorf("subscribe tbox telemetry failed: %w", err)
	}

	logger.Log.Info("emqx consumer started", zap.String("topic", topic))
	return nil
}
