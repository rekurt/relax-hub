package app

import (
	"github.com/nikitaaldaev/bani/config"
	"github.com/nikitaaldaev/bani/internal/database"
	"github.com/nikitaaldaev/bani/internal/middleware"
	"github.com/nikitaaldaev/bani/internal/server"
	"go.uber.org/fx"
)

func New(cfg *config.Config) *fx.App {
	return fx.New(
		fx.Supply(cfg),
		database.PostgresModule,
		database.RedisModule,
		middleware.CORSModule,
		server.Module,
	)
}
