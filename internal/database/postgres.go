package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rekurt/relax-hub/config"
	"github.com/rekurt/relax-hub/internal/logger"
	"go.uber.org/fx"
)

const (
	// DefaultQueryTimeout is the default timeout for database queries in production
	DefaultQueryTimeout = 30 * time.Second
)

var PostgresModule = fx.Module("postgres",
	fx.Provide(NewPostgresPool),
)

func NewPostgresPool(lc fx.Lifecycle, cfg *config.Config, log *logger.Logger) (*pgxpool.Pool, error) {
	log.Info("Initializing PostgreSQL connection pool")

	poolCfg, err := pgxpool.ParseConfig(cfg.Database.DSN)
	if err != nil {
		log.Error("Failed to parse PostgreSQL config", "error", err)
		return nil, fmt.Errorf("parse postgres config: %w", err)
	}

	// Apply pool tuning from config
	if cfg.Database.MaxConns > 0 {
		poolCfg.MaxConns = cfg.Database.MaxConns
	}
	if cfg.Database.MinConns > 0 {
		poolCfg.MinConns = cfg.Database.MinConns
	}
	if cfg.Database.MaxConnLifetime > 0 {
		poolCfg.MaxConnLifetime = cfg.Database.MaxConnLifetime
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), poolCfg)
	if err != nil {
		log.Error("Failed to create PostgreSQL pool", "error", err)
		return nil, fmt.Errorf("create postgres pool: %w", err)
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			// Create a context with timeout for the health check ping
			pingCtx, cancel := context.WithTimeout(ctx, DefaultQueryTimeout)
			defer cancel()

			if err := pool.Ping(pingCtx); err != nil {
				log.Error("Failed to ping PostgreSQL", "error", err)
				return err
			}
			log.Info("PostgreSQL connection established",
				"query_timeout", DefaultQueryTimeout,
				"max_conns", poolCfg.MaxConns,
				"min_conns", poolCfg.MinConns,
				"max_conn_lifetime", poolCfg.MaxConnLifetime,
			)
			return nil
		},
		OnStop: func(ctx context.Context) error {
			log.Info("Closing PostgreSQL connection pool")
			pool.Close()
			return nil
		},
	})

	return pool, nil
}
