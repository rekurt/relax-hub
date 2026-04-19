package app

import (
	"testing"
	"time"

	"github.com/rekurt/relax-hub/config"
	"go.uber.org/fx"
)

func TestNew_CreatesApp(t *testing.T) {
	cfg := &config.Config{
		Server: config.ServerConfig{
			Host: "0.0.0.0",
			Port: 8080,
		},
		Database: config.DatabaseConfig{
			DSN: "postgres://postgres:postgres@localhost:5432/bani_test?sslmode=disable",
		},
		Redis: config.RedisConfig{
			Addr:     "localhost:6379",
			Password: "",
			DB:       0,
		},
		JWT: config.JWTConfig{
			Secret:   "test-secret",
			TokenTTL: 24 * time.Hour,
		},
	}

	app := New(cfg)
	if app == nil {
		t.Fatal("expected fx.App to be created, got nil")
	}
}

func TestNew_ValidatesDependencyGraph(t *testing.T) {
	cfg := &config.Config{
		Database: config.DatabaseConfig{
			DSN: "postgres://postgres:postgres@localhost:5432/bani_test?sslmode=disable",
		},
		Redis: config.RedisConfig{
			Addr:     "localhost:6379",
			Password: "",
			DB:       0,
		},
	}

	err := fx.ValidateApp(
		fx.Supply(cfg),
		fx.Provide(NewTestInvoker),
	)
	if err != nil {
		t.Fatalf("expected valid dependency graph, got error: %v", err)
	}
}

func NewTestInvoker(cfg *config.Config) string {
	return "ok"
}
