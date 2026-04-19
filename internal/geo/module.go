package geo

import (
	"github.com/rekurt/relax-hub/config"
	"github.com/rekurt/relax-hub/internal/logger"
	"github.com/redis/go-redis/v9"
	"go.uber.org/fx"
)

var Module = fx.Module("geo",
	fx.Provide(
		func(redisClient *redis.Client, log *logger.Logger, cfg *config.Config) *IsochroneService {
			return NewIsochroneService(redisClient, log, cfg.Geo.IsochroneAPIURL, cfg.Geo.IsochroneAPIKey)
		},
		func(redisClient *redis.Client, log *logger.Logger, cfg *config.Config) *TransportService {
			return NewTransportService(redisClient, log, cfg.Geo.YandexSearchAPIKey)
		},
	),
)
