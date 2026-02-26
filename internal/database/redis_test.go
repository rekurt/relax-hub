package database

import (
	"testing"

	"github.com/nikitaaldaev/bani/config"
	"github.com/nikitaaldaev/bani/internal/logger"
	"github.com/redis/go-redis/v9"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
)

func TestRedisModule_Provides_Client(t *testing.T) {
	cfg := &config.Config{
		Redis: config.RedisConfig{
			Addr:     "localhost:6379",
			Password: "",
			DB:       0,
		},
		Logger: config.LoggerConfig{
			Level: "info",
		},
	}

	var client *redis.Client

	app := fxtest.New(t,
		fx.Supply(cfg),
		logger.Module,
		RedisModule,
		fx.Populate(&client),
	)
	defer app.RequireStop()

	if client == nil {
		t.Fatal("expected redis.Client to be provided, got nil")
	}
}

func TestNewRedisClient_CreatesClient(t *testing.T) {
	cfg := &config.Config{
		Redis: config.RedisConfig{
			Addr:     "localhost:6379",
			Password: "",
			DB:       0,
		},
		Logger: config.LoggerConfig{
			Level: "info",
		},
	}

	lc := fxtest.NewLifecycle(t)
	log := logger.New(logger.LevelInfo)
	client, err := NewRedisClient(lc, cfg, log)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if client == nil {
		t.Fatal("expected redis client to be created, got nil")
	}
}
