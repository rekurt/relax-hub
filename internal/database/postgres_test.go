package database

import (
	"testing"
	"time"

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

func TestDefaultQueryTimeout_IsSet(t *testing.T) {
	// Verify that the default query timeout constant is positive
	if DefaultQueryTimeout <= 0 {
		t.Error("DefaultQueryTimeout must be positive")
	}
}

func TestPostgresTimeoutConfiguration(t *testing.T) {
	// Verify timeout value is reasonable for production (20-30 seconds)
	if DefaultQueryTimeout < 20*time.Second || DefaultQueryTimeout > 30*time.Second {
		t.Errorf("DefaultQueryTimeout (%v) is outside typical production range (20-30s)", DefaultQueryTimeout)
	}
}

func TestNewPostgresPool_AppliesPoolConfig(t *testing.T) {
	cfg := &config.Config{
		Database: config.DatabaseConfig{
			DSN:             "postgres://postgres:postgres@localhost:5432/bani_test?sslmode=disable",
			MaxConns:        25,
			MinConns:        3,
			MaxConnLifetime: 2 * time.Hour,
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

	stat := pool.Stat()
	if stat.MaxConns() != 25 {
		t.Errorf("expected MaxConns = 25, got %d", stat.MaxConns())
	}
}

func TestNewPostgresPool_DefaultPoolConfig(t *testing.T) {
	// When pool config fields are zero, pgx defaults are used
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
