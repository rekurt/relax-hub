package database

import (
	"testing"

	"github.com/nikitaaldaev/bani/config"
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
	}

	var client *redis.Client

	app := fxtest.New(t,
		fx.Supply(cfg),
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
	}

	lc := fxtest.NewLifecycle(t)
	client, err := NewRedisClient(lc, cfg)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if client == nil {
		t.Fatal("expected redis client to be created, got nil")
	}
}
