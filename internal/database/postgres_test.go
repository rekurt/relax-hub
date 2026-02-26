package database

import (
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nikitaaldaev/bani/config"
	"github.com/nikitaaldaev/bani/internal/logger"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
)

func TestPostgresModule_Provides_Pool(t *testing.T) {
	cfg := &config.Config{
		Database: config.DatabaseConfig{
			DSN: "postgres://postgres:postgres@localhost:5432/bani_test?sslmode=disable",
		},
		Logger: config.LoggerConfig{
			Level: "info",
		},
	}

	var pool *pgxpool.Pool

	app := fxtest.New(t,
		fx.Supply(cfg),
		logger.Module,
		PostgresModule,
		fx.Populate(&pool),
	)
	defer app.RequireStop()

	if pool == nil {
		t.Fatal("expected pgxpool.Pool to be provided, got nil")
	}
}

func TestNewPostgresPool_InvalidDSN(t *testing.T) {
	cfg := &config.Config{
		Database: config.DatabaseConfig{
			DSN: "://invalid",
		},
		Logger: config.LoggerConfig{
			Level: "info",
		},
	}

	log := logger.New(logger.LevelInfo)
	_, err := NewPostgresPool(fx.Lifecycle(nil), cfg, log)
	if err == nil {
		t.Fatal("expected error for invalid DSN, got nil")
	}
}
