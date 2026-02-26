package database

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nikitaaldaev/bani/config"
	"github.com/nikitaaldaev/bani/internal/logger"
	"go.uber.org/fx"
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

	pool, err := pgxpool.NewWithConfig(context.Background(), poolCfg)
	if err != nil {
		log.Error("Failed to create PostgreSQL pool", "error", err)
		return nil, fmt.Errorf("create postgres pool: %w", err)
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			if err := pool.Ping(ctx); err != nil {
				log.Error("Failed to ping PostgreSQL", "error", err)
				return err
			}
			log.Info("PostgreSQL connection established")
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
