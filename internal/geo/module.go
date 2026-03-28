package geo

import (
	"github.com/nikitaaldaev/bani/config"
	"github.com/nikitaaldaev/bani/internal/logger"
	"github.com/redis/go-redis/v9"
	"go.uber.org/fx"
)

var Module = fx.Module("geo",
	fx.Provide(
		func(redisClient *redis.Client, log *logger.Logger, cfg *config.Config) *IsochroneService {
			return NewIsochroneService(redisClient, log, cfg.Geo.IsochroneAPIURL, cfg.Geo.IsochroneAPIKey)
		},
	),
)
