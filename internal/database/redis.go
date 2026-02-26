package database

import (
	"context"
	"fmt"

	"github.com/nikitaaldaev/bani/config"
	"github.com/nikitaaldaev/bani/internal/logger"
	"github.com/redis/go-redis/v9"
	"go.uber.org/fx"
)

var RedisModule = fx.Module("redis",
	fx.Provide(NewRedisClient),
)

func NewRedisClient(lc fx.Lifecycle, cfg *config.Config, log *logger.Logger) (*redis.Client, error) {
	log.Info("Initializing Redis client", "addr", cfg.Redis.Addr)

	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			if err := client.Ping(ctx).Err(); err != nil {
				log.Error("Failed to ping Redis", "error", err)
				return fmt.Errorf("redis ping: %w", err)
			}
			log.Info("Redis connection established")
			return nil
		},
		OnStop: func(ctx context.Context) error {
			log.Info("Closing Redis connection")
			return client.Close()
		},
	})

	return client, nil
}
