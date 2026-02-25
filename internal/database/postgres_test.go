package database

import (
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nikitaaldaev/bani/config"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
)

func TestPostgresModule_Provides_Pool(t *testing.T) {
	cfg := &config.Config{
		Database: config.DatabaseConfig{
			DSN: "postgres://postgres:postgres@localhost:5432/bani_test?sslmode=disable",
		},
	}

	var pool *pgxpool.Pool

	app := fxtest.New(t,
		fx.Supply(cfg),
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
	}

	_, err := NewPostgresPool(fx.Lifecycle(nil), cfg)
	if err == nil {
		t.Fatal("expected error for invalid DSN, got nil")
	}
}
