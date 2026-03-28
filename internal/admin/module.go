package admin

import (
	"context"
	"fmt"
	"net/http"

	"github.com/GoAdminGroup/go-admin/engine"
	gaconfig "github.com/GoAdminGroup/go-admin/modules/config"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	appconfig "github.com/nikitaaldaev/bani/config"
	"github.com/nikitaaldaev/bani/internal/admin/pages"
	"github.com/nikitaaldaev/bani/internal/logger"
	"github.com/nikitaaldaev/bani/internal/middleware"
	"github.com/nikitaaldaev/bani/internal/repository/postgres"
	"github.com/nikitaaldaev/bani/internal/notification"
	"github.com/redis/go-redis/v9"
	"go.uber.org/fx"
)

// GoAdmin wraps the GoAdmin engine and its config for deferred initialization.
// GoAdmin's AddConfig immediately opens a DB connection, so we defer it to fx OnStart.
// The Mux field is set by the router so that Engine.Use() can be called in OnStart
// after AddConfig (Engine.Use requires a configured engine).
type GoAdmin struct {
	Engine      *engine.Engine
	Config      *gaconfig.Config
	PagesRouter http.Handler
	Mux         *chi.Mux
}

// NewGoAdmin creates a GoAdmin wrapper with engine and config, but does NOT
// connect to the database yet. Database connection happens in fx lifecycle OnStart.
func NewGoAdmin(cfg *appconfig.Config, pool *pgxpool.Pool, redisClient *redis.Client, hub *notification.Hub, log *logger.Logger) *GoAdmin {
	log.Info("Preparing GoAdmin engine",
		"prefix", cfg.Admin.Prefix,
		"language", cfg.Admin.Language,
		"theme", cfg.Admin.Theme,
	)

	dashProvider := pages.NewPostgresDashboardProvider(pool, log)
	modProvider := pages.NewPostgresModerationProvider(pool, log)
	analyticsProvider := pages.NewPostgresAnalyticsProvider(pool, log)
	healthProvider := pages.NewPlatformHealthProvider(pool, redisClient, hub, log)
	financeProvider := pages.NewPostgresFinanceProvider(pool, log)
	advAnalyticsProvider := pages.NewPostgresAdvancedAnalyticsProvider(postgres.NewAnalyticsRepository(pool), log)

	return &GoAdmin{
		Engine:      engine.Default(),
		Config:      BuildGoAdminConfig(cfg),
		PagesRouter: PagesRouter(dashProvider, modProvider, analyticsProvider, healthProvider, financeProvider, advAnalyticsProvider, log, cfg.Admin.Prefix),
	}
}

// Module provides the GoAdmin components via fx dependency injection.
var Module = fx.Module("goadmin",
	fx.Provide(NewGoAdmin),
	fx.Invoke(registerLifecycle),
)

func registerLifecycle(lc fx.Lifecycle, ga *GoAdmin, pool *pgxpool.Pool, cfg *appconfig.Config, authService middleware.AuthService, log *logger.Logger) {
	processor := NewAuthProcessor(authService, log, cfg.Environment == "production")
	ga.Engine.AddAuthService(processor)

	pagesPrefix := cfg.Admin.Prefix + "/pages"

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			log.Info("Starting GoAdmin engine, connecting to database")
			ga.Engine.AddConfig(ga.Config)

			// Mount GoAdmin engine routes on the chi mux.
			// This must happen after AddConfig because Engine.Use requires a configured engine.
			if ga.Mux != nil {
				if err := ga.Engine.Use(ga.Mux); err != nil {
					log.Error("Failed to mount GoAdmin engine", "error", err)
					return fmt.Errorf("mount goadmin engine: %w", err)
				}
				log.Info("GoAdmin panel mounted", "prefix", ga.Config.UrlPrefix)

				// Mount custom admin pages with auth that supports both
				// Authorization header and admin_token cookie for browser access
				ga.Mux.Route(pagesPrefix, func(r chi.Router) {
					r.Use(RequireAdminAuth(authService))
					r.Mount("/", ga.PagesRouter)
				})
				log.Info("Admin custom pages mounted", "prefix", pagesPrefix)
			}

			if err := RegisterCustomMenu(ctx, pool, pagesPrefix, log); err != nil {
				log.Error("Failed to register custom menu", "error", err)
				return fmt.Errorf("register admin menu: %w", err)
			}

			log.Info("GoAdmin engine started successfully")
			return nil
		},
		OnStop: func(ctx context.Context) error {
			log.Info("Closing GoAdmin database connections")
			conn := ga.Engine.DefaultConnection()
			if conn != nil {
				if errs := conn.Close(); len(errs) > 0 {
					log.Error("Errors closing GoAdmin database", "errors", errs)
				}
			}
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
