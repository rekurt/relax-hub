package app

import (
	"github.com/nikitaaldaev/bani/config"
	"github.com/nikitaaldaev/bani/internal/database"
	"github.com/nikitaaldaev/bani/internal/handler"
	"github.com/nikitaaldaev/bani/internal/logger"
	"github.com/nikitaaldaev/bani/internal/middleware"
	"github.com/nikitaaldaev/bani/internal/moderation"
	"github.com/nikitaaldaev/bani/internal/notification"
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
		moderation.Module,
		notification.Module,
		service.Module,
		handler.Module,
		server.Module,
		// Cross-package interface bindings
		fx.Provide(
			func(hub *notification.Hub) service.ChatBroadcaster { return hub },
		),
	)
}
