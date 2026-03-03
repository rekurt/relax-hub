package app

import (
	"github.com/nikitaaldaev/bani/config"
	"github.com/nikitaaldaev/bani/internal/database"
	"github.com/nikitaaldaev/bani/internal/handler"
	"github.com/nikitaaldaev/bani/internal/logger"
	"github.com/nikitaaldaev/bani/internal/middleware"
	repopostgres "github.com/nikitaaldaev/bani/internal/repository/postgres"
	"github.com/nikitaaldaev/bani/internal/server"
	"github.com/nikitaaldaev/bani/internal/service"
	"github.com/nikitaaldaev/bani/internal/storage"
	"go.uber.org/fx"
)

func New(cfg *config.Config) *fx.App {
	return fx.New(
		fx.Supply(cfg),
		logger.Module,
		database.PostgresModule,
		database.RedisModule,
		middleware.CORSModule,
		repopostgres.Module,
		storage.Module,
		service.Module,
		handler.Module,
		server.Module,
	)
}
