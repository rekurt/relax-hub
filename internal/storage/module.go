package storage

import (
	"github.com/nikitaaldaev/bani/config"
	"github.com/nikitaaldaev/bani/internal/logger"
	"go.uber.org/fx"
)

var Module = fx.Module("storage",
	fx.Provide(provideFileStorage),
)

func provideFileStorage(cfg *config.Config, log *logger.Logger) (FileStorage, error) {
	s, err := NewS3Storage(cfg, log)
	if err != nil && cfg.Environment != "production" {
		log.Warn("S3 storage unavailable, storage operations will be no-ops", "error", err)
		return (*S3Storage)(nil), nil
	}
	return s, err
}
