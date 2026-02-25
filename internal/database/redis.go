package database

import (
	"context"
	"fmt"

	"github.com/nikitaaldaev/bani/config"
	"github.com/redis/go-redis/v9"
	"go.uber.org/fx"
)

var RedisModule = fx.Module("redis",
	fx.Provide(NewRedisClient),
)

func NewRedisClient(lc fx.Lifecycle, cfg *config.Config) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			if err := client.Ping(ctx).Err(); err != nil {
				return fmt.Errorf("redis ping: %w", err)
			}
			return nil
		},
		OnStop: func(ctx context.Context) error {
			return client.Close()
		},
	})

	return client, nil
}
