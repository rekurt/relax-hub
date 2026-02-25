package app

import (
	"github.com/nikitaaldaev/bani/config"
	"github.com/nikitaaldaev/bani/internal/database"
	"go.uber.org/fx"
)

func New(cfg *config.Config) *fx.App {
	return fx.New(
		fx.Supply(cfg),
		database.PostgresModule,
		database.RedisModule,
	)
}
