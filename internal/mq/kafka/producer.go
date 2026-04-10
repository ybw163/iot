package kafka

import (
	"context"
	"fmt"
	"time"

	"github.com/IBM/sarama"
	"go.uber.org/zap"

	"iot/pkg/config"
	"iot/pkg/logger"
)

type Producer struct {
	p   sarama.SyncProducer
	cfg config.KafkaConfig
}

var DefaultProducer *Producer

func InitProducer(cfg config.KafkaConfig) error {
	saramaCfg := sarama.NewConfig()
	saramaCfg.Producer.RequiredAcks = sarama.WaitForAll
	saramaCfg.Producer.Retry.Max = cfg.RetryMax
	saramaCfg.Producer.Timeout = time.Duration(cfg.Timeout) * time.Second
	saramaCfg.Producer.Return.Successes = true

	p, err := sarama.NewSyncProducer(cfg.Brokers, saramaCfg)
	if err != nil {
		return fmt.Errorf("create kafka producer failed: %w", err)
	}

	DefaultProducer = &Producer{
		p:   p,
		cfg: cfg,
	}

	logger.Log.Info("kafka producer started", zap.Strings("brokers", cfg.Brokers))
	return nil
}

// SendSync 同步发送消息
func (p *Producer) SendSync(ctx context.Context, topic string, body []byte) (int32, int64, error) {
	msg := &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.ByteEncoder(body),
	}

	partition, offset, err := p.p.SendMessage(msg)
	if err != nil {
		logger.Log.Error("send message failed", zap.String("topic", topic), zap.Error(err))
		return 0, 0, err
	}

	logger.Log.Info("send message success",
		zap.String("topic", topic),
		zap.Int32("partition", partition),
		zap.Int64("offset", offset),
	)

	return partition, offset, nil
}

func CloseProducer() {
	if DefaultProducer != nil {
		if err := DefaultProducer.p.Close(); err != nil {
			logger.Log.Error("close kafka producer failed", zap.Error(err))
		}
	}
}
