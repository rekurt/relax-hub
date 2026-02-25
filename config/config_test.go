package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoad_Defaults(t *testing.T) {
	// Load from a non-existent path to test defaults
	cfg, err := Load("/nonexistent/path/config.yaml")
	if err == nil {
		t.Log("No config file found, defaults should apply")
	}
	// When file not found, we get an error - let's test with empty dir
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(cfgPath, []byte(""), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err = Load(cfgPath)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Server.Host != "0.0.0.0" {
		t.Errorf("expected server.host = 0.0.0.0, got %s", cfg.Server.Host)
	}
	if cfg.Server.Port != 8080 {
		t.Errorf("expected server.port = 8080, got %d", cfg.Server.Port)
	}
	if cfg.Redis.Addr != "localhost:6379" {
		t.Errorf("expected redis.addr = localhost:6379, got %s", cfg.Redis.Addr)
	}
	if cfg.Redis.DB != 0 {
		t.Errorf("expected redis.db = 0, got %d", cfg.Redis.DB)
	}
	if cfg.JWT.Secret != "change-me-in-production" {
		t.Errorf("expected jwt.secret = change-me-in-production, got %s", cfg.JWT.Secret)
	}
	if cfg.JWT.TokenTTL != 24*time.Hour {
		t.Errorf("expected jwt.token_ttl = 24h, got %s", cfg.JWT.TokenTTL)
	}
}

func TestLoad_FromYAML(t *testing.T) {
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "config.yaml")

	yamlContent := `
server:
  host: "127.0.0.1"
  port: 9090
database:
  dsn: "postgres://user:pass@db:5432/testdb?sslmode=disable"
redis:
  addr: "redis:6380"
  password: "secret"
  db: 1
jwt:
  secret: "my-secret-key"
  token_ttl: "2h"
`
	if err := os.WriteFile(cfgPath, []byte(yamlContent), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(cfgPath)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Server.Host != "127.0.0.1" {
		t.Errorf("expected server.host = 127.0.0.1, got %s", cfg.Server.Host)
	}
	if cfg.Server.Port != 9090 {
		t.Errorf("expected server.port = 9090, got %d", cfg.Server.Port)
	}
	if cfg.Database.DSN != "postgres://user:pass@db:5432/testdb?sslmode=disable" {
		t.Errorf("expected database.dsn = postgres://..., got %s", cfg.Database.DSN)
	}
	if cfg.Redis.Addr != "redis:6380" {
		t.Errorf("expected redis.addr = redis:6380, got %s", cfg.Redis.Addr)
	}
	if cfg.Redis.Password != "secret" {
		t.Errorf("expected redis.password = secret, got %s", cfg.Redis.Password)
	}
	if cfg.Redis.DB != 1 {
		t.Errorf("expected redis.db = 1, got %d", cfg.Redis.DB)
	}
	if cfg.JWT.Secret != "my-secret-key" {
		t.Errorf("expected jwt.secret = my-secret-key, got %s", cfg.JWT.Secret)
	}
	if cfg.JWT.TokenTTL != 2*time.Hour {
		t.Errorf("expected jwt.token_ttl = 2h, got %s", cfg.JWT.TokenTTL)
	}
}

func TestLoad_EnvOverride(t *testing.T) {
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(cfgPath, []byte(""), 0644); err != nil {
		t.Fatal(err)
	}

	t.Setenv("BANI_SERVER_PORT", "3000")
	t.Setenv("BANI_JWT_SECRET", "env-secret")

	cfg, err := Load(cfgPath)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Server.Port != 3000 {
		t.Errorf("expected server.port = 3000 from env, got %d", cfg.Server.Port)
	}
	if cfg.JWT.Secret != "env-secret" {
		t.Errorf("expected jwt.secret = env-secret from env, got %s", cfg.JWT.Secret)
	}
}

func TestLoad_InvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "config.yaml")

	if err := os.WriteFile(cfgPath, []byte("invalid: [yaml: content"), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := Load(cfgPath)
	if err == nil {
		t.Error("expected error for invalid YAML, got nil")
	}
}
