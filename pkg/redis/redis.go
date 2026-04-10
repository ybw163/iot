package redis

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"iot/pkg/config"
	"iot/pkg/logger"
)

var Client *redis.Client

func Init(cfg config.RedisConfig) error {
	Client = redis.NewClient(&redis.Options{
		Addr:     cfg.Addr(),
		Password: cfg.Password,
		DB:       cfg.DB,
		PoolSize: cfg.PoolSize,
	})

	if err := Client.Ping(context.Background()).Err(); err != nil {
		return fmt.Errorf("connect redis failed: %w", err)
	}

	logger.Log.Info("redis connected", zap.String("addr", cfg.Addr()))
	return nil
}

func Close() {
	if Client != nil {
		_ = Client.Close()
	}
}
