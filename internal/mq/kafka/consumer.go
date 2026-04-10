package kafka

import (
	"context"
	"fmt"

	"github.com/IBM/sarama"
	"go.uber.org/zap"

	"iot/pkg/config"
	"iot/pkg/logger"
)

type Consumer struct {
	c   sarama.ConsumerGroup
	cfg config.KafkaConfig
}

var DefaultConsumer *Consumer

// MessageHandler 消息处理函数类型
type MessageHandler func(*sarama.ConsumerMessage) error

func InitConsumer(cfg config.KafkaConfig) error {
	saramaCfg := sarama.NewConfig()
	saramaCfg.Consumer.Group.Rebalance.GroupStrategies = []sarama.BalanceStrategy{sarama.NewBalanceStrategyRange()}
	saramaCfg.Consumer.Offsets.Initial = sarama.OffsetNewest

	c, err := sarama.NewConsumerGroup(cfg.Brokers, cfg.GroupID, saramaCfg)
	if err != nil {
		return fmt.Errorf("create kafka consumer failed: %w", err)
	}

	DefaultConsumer = &Consumer{
		c:   c,
		cfg: cfg,
	}

	logger.Log.Info("kafka consumer created", zap.Strings("brokers", cfg.Brokers), zap.String("group_id", cfg.GroupID))
	return nil
}

// Subscribe 订阅主题
func (c *Consumer) Subscribe(ctx context.Context, topic string, handler MessageHandler) error {
	h := &consumerGroupHandler{
		handler: handler,
		topic:   topic,
	}

	go func() {
		for {
			if err := c.c.Consume(ctx, []string{topic}, h); err != nil {
				logger.Log.Error("consume error", zap.String("topic", topic), zap.Error(err))
			}
			if ctx.Err() != nil {
				return
			}
		}
	}()

	return nil
}

// Start 启动消费者（兼容原 RocketMQ 接口，实际订阅在 Subscribe 中已启动 goroutine）
func (c *Consumer) Start() error {
	logger.Log.Info("kafka consumer started")
	return nil
}

func CloseConsumer() {
	if DefaultConsumer != nil {
		if err := DefaultConsumer.c.Close(); err != nil {
			logger.Log.Error("close kafka consumer failed", zap.Error(err))
		}
	}
}

// consumerGroupHandler 实现 sarama.ConsumerGroupHandler 接口
type consumerGroupHandler struct {
	handler MessageHandler
	topic   string
}

func (h *consumerGroupHandler) Setup(sarama.ConsumerGroupSession) error {
	return nil
}

func (h *consumerGroupHandler) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

func (h *consumerGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		if err := h.handler(msg); err != nil {
			logger.Log.Error("consume message failed",
				zap.String("topic", msg.Topic),
				zap.Int32("partition", msg.Partition),
				zap.Int64("offset", msg.Offset),
				zap.Error(err),
			)
			return err
		}
		session.MarkMessage(msg, "")
	}
	return nil
}
