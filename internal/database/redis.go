package database

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/rekurt/relax-hub/config"
	"github.com/rekurt/relax-hub/internal/logger"
	"go.uber.org/fx"
)

const (
	// RedisOperationTimeout is the timeout for Redis operations
	RedisOperationTimeout = 5 * time.Second
)

var RedisModule = fx.Module("redis",
	fx.Provide(NewRedisClient),
)

func NewRedisClient(lc fx.Lifecycle, cfg *config.Config, log *logger.Logger) (*redis.Client, error) {
	log.Info("Initializing Redis client", "addr", cfg.Redis.Addr, "operation_timeout", RedisOperationTimeout)

	client := redis.NewClient(&redis.Options{
		Addr:         cfg.Redis.Addr,
		Password:     cfg.Redis.Password,
		DB:           cfg.Redis.DB,
		ReadTimeout:  RedisOperationTimeout,
		WriteTimeout: RedisOperationTimeout,
		PoolTimeout:  RedisOperationTimeout,
	})

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			// Create a context with timeout for the health check ping
			pingCtx, cancel := context.WithTimeout(ctx, RedisOperationTimeout)
			defer cancel()

			if err := client.Ping(pingCtx).Err(); err != nil {
				log.Error("Failed to ping Redis", "error", err)
				return fmt.Errorf("redis ping: %w", err)
			}
			log.Info("Redis connection established", "operation_timeout", RedisOperationTimeout)
			return nil
		},
		OnStop: func(ctx context.Context) error {
			log.Info("Closing Redis connection")
			return client.Close()
		},
	})

	return client, nil
}
