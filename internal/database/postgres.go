package database

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nikitaaldaev/bani/config"
	"go.uber.org/fx"
)

var PostgresModule = fx.Module("postgres",
	fx.Provide(NewPostgresPool),
)

func NewPostgresPool(lc fx.Lifecycle, cfg *config.Config) (*pgxpool.Pool, error) {
	poolCfg, err := pgxpool.ParseConfig(cfg.Database.DSN)
	if err != nil {
		return nil, fmt.Errorf("parse postgres config: %w", err)
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), poolCfg)
	if err != nil {
		return nil, fmt.Errorf("create postgres pool: %w", err)
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			return pool.Ping(ctx)
		},
		OnStop: func(ctx context.Context) error {
			pool.Close()
			return nil
		},
	})

	return pool, nil
}
