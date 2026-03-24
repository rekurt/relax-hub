package storage

import (
	"strings"

	"github.com/nikitaaldaev/bani/config"
	"github.com/nikitaaldaev/bani/internal/logger"
	"go.uber.org/fx"
)

var Module = fx.Module("storage",
	fx.Provide(provideFileStorage),
)

func provideFileStorage(cfg *config.Config, log *logger.Logger) (FileStorage, error) {
	s, err := NewS3Storage(cfg, log)
	if err != nil && !strings.EqualFold(cfg.Environment, "production") {
		log.Warn("S3 storage unavailable, using in-memory mock storage", "error", err)
		return NewMockStorage(), nil
	}
	return s, err
}
