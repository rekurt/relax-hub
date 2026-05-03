package seo

import (
	"github.com/redis/go-redis/v9"
	"github.com/rekurt/relax-hub/config"
	"go.uber.org/fx"
)

var Module = fx.Module("seo",
	fx.Provide(func(redisClient *redis.Client, cfg *config.Config) *Renderer {
		return NewRenderer(redisClient, cfg.BaseURL)
	}),
)
