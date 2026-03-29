package seo

import (
	"github.com/nikitaaldaev/bani/config"
	"github.com/redis/go-redis/v9"
	"go.uber.org/fx"
)

var Module = fx.Module("seo",
	fx.Provide(func(redisClient *redis.Client, cfg *config.Config) *Renderer {
		return NewRenderer(redisClient, cfg.BaseURL)
	}),
)
