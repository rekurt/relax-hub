package admin

import (
	"context"

	"github.com/GoAdminGroup/go-admin/engine"
	gaconfig "github.com/GoAdminGroup/go-admin/modules/config"
	"github.com/jackc/pgx/v5/pgxpool"

	appconfig "github.com/nikitaaldaev/bani/config"
	"github.com/nikitaaldaev/bani/internal/logger"
	"github.com/nikitaaldaev/bani/internal/middleware"
	"go.uber.org/fx"
)

// GoAdmin wraps the GoAdmin engine and its config for deferred initialization.
// GoAdmin's AddConfig immediately opens a DB connection, so we defer it to fx OnStart.
type GoAdmin struct {
	Engine *engine.Engine
	Config *gaconfig.Config
}

// NewGoAdmin creates a GoAdmin wrapper with engine and config, but does NOT
// connect to the database yet. Database connection happens in fx lifecycle OnStart.
func NewGoAdmin(cfg *appconfig.Config, log *logger.Logger) *GoAdmin {
	log.Info("Preparing GoAdmin engine",
		"prefix", cfg.Admin.Prefix,
		"language", cfg.Admin.Language,
		"theme", cfg.Admin.Theme,
	)

	return &GoAdmin{
		Engine: engine.Default(),
		Config: BuildGoAdminConfig(cfg),
	}
}

// Module provides the GoAdmin components via fx dependency injection.
var Module = fx.Module("goadmin",
	fx.Provide(NewGoAdmin),
	fx.Invoke(registerLifecycle),
)

func registerLifecycle(lc fx.Lifecycle, ga *GoAdmin, pool *pgxpool.Pool, cfg *appconfig.Config, authService middleware.AuthService, log *logger.Logger) {
	processor := NewAuthProcessor(authService, log)
	ga.Engine.AddAuthService(processor)

	pagesPrefix := cfg.Admin.Prefix
	if pagesPrefix == "" {
		pagesPrefix = "/admin-panel"
	}
	pagesPrefix += "/pages"

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			log.Info("Starting GoAdmin engine, connecting to database")
			ga.Engine.AddConfig(ga.Config)

			if err := RegisterCustomMenu(ctx, pool, pagesPrefix, log); err != nil {
				log.Error("Failed to register custom menu", "error", err)
			}

			log.Info("GoAdmin engine started successfully")
			return nil
		},
	})
}

// ProvideConditionalModule returns admin fx options only when admin is enabled.
func ProvideConditionalModule(cfg *appconfig.Config) fx.Option {
	if !cfg.Admin.Enabled {
		return fx.Options()
	}
	return Module
}
