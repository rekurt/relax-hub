package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoad_Defaults(t *testing.T) {
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(cfgPath, []byte(""), 0644); err != nil {
		t.Fatal(err)
	}

	// Default JWT secret is insecure, so Load should return an error
	_, err := Load(cfgPath)
	if err == nil {
		t.Fatal("expected error for default JWT secret, got nil")
	}
}

func TestLoad_DefaultsWithJWTSecret(t *testing.T) {
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(cfgPath, []byte(""), 0644); err != nil {
		t.Fatal(err)
	}

	t.Setenv("BANI_JWT_SECRET", "a-secure-secret-for-testing")

	cfg, err := Load(cfgPath)
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

func TestValidate_MissingDSN(t *testing.T) {
	cfg := &Config{
		Environment: "dev",
		Database:    DatabaseConfig{DSN: ""},
		Redis:       RedisConfig{Addr: "localhost:6379"},
		JWT:         JWTConfig{Secret: "test-secret"},
	}

	err := cfg.Validate()
	if err == nil {
		t.Error("expected error for missing database.dsn, got nil")
	}
	if err.Error() != "database.dsn is required (set BANI_DATABASE_DSN)" {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestValidate_MissingJWTSecret(t *testing.T) {
	cfg := &Config{
		Environment: "dev",
		Database:    DatabaseConfig{DSN: "postgres://localhost/db"},
		Redis:       RedisConfig{Addr: "localhost:6379"},
		JWT:         JWTConfig{Secret: ""},
	}

	err := cfg.Validate()
	if err == nil {
		t.Error("expected error for missing jwt.secret, got nil")
	}
}

func TestValidate_ChangeMe(t *testing.T) {
	cfg := &Config{
		Environment: "dev",
		Database:    DatabaseConfig{DSN: "postgres://localhost/db"},
		Redis:       RedisConfig{Addr: "localhost:6379"},
		JWT:         JWTConfig{Secret: "change-me-in-production"},
	}

	err := cfg.Validate()
	if err == nil {
		t.Error("expected error for placeholder jwt.secret, got nil")
	}
}

func TestValidate_MissingRedisAddr(t *testing.T) {
	cfg := &Config{
		Environment: "dev",
		Database:    DatabaseConfig{DSN: "postgres://localhost/db"},
		Redis:       RedisConfig{Addr: ""},
		JWT:         JWTConfig{Secret: "test-secret"},
	}

	err := cfg.Validate()
	if err == nil {
		t.Error("expected error for missing redis.addr, got nil")
	}
	if err.Error() != "redis.addr is required (set BANI_REDIS_ADDR)" {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestValidate_ProductionShortSecret(t *testing.T) {
	cfg := &Config{
		Environment: "production",
		Database:    DatabaseConfig{DSN: "postgres://localhost/db?sslmode=require"},
		Redis:       RedisConfig{Addr: "localhost:6379"},
		JWT:         JWTConfig{Secret: "short-secret"},
	}

	err := cfg.Validate()
	if err == nil {
		t.Error("expected error for short jwt.secret in production, got nil")
	}
	if len(err.Error()) == 0 || !contains(err.Error(), "32 characters") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestValidate_ProductionNoSSLMode(t *testing.T) {
	secret := "this-is-a-very-long-secret-that-is-definitely-over-32-chars"
	cfg := &Config{
		Environment: "production",
		Database:    DatabaseConfig{DSN: "postgres://localhost/db?sslmode=disable"},
		Redis:       RedisConfig{Addr: "localhost:6379"},
		JWT:         JWTConfig{Secret: secret},
	}

	err := cfg.Validate()
	if err == nil {
		t.Error("expected error for missing sslmode=require in production, got nil")
	}
	if err.Error() != "database.dsn must use sslmode=require in production" {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestValidate_ProductionValid(t *testing.T) {
	secret := "this-is-a-very-long-secret-that-is-definitely-over-32-chars"
	cfg := &Config{
		Environment: "production",
		Database:    DatabaseConfig{DSN: "postgres://localhost/db?sslmode=require"},
		Redis:       RedisConfig{Addr: "localhost:6379"},
		JWT:         JWTConfig{Secret: secret},
		Storage:     StorageConfig{AccessKey: "prod-key", SecretKey: "prod-secret", UseSSL: true},
	}

	err := cfg.Validate()
	if err != nil {
		t.Errorf("expected nil error for valid production config, got %v", err)
	}
}

func TestValidate_DevValid(t *testing.T) {
	cfg := &Config{
		Environment: "dev",
		Database:    DatabaseConfig{DSN: "postgres://localhost/db"},
		Redis:       RedisConfig{Addr: "localhost:6379"},
		JWT:         JWTConfig{Secret: "short-secret"},
	}

	err := cfg.Validate()
	if err != nil {
		t.Errorf("expected nil error for valid dev config, got %v", err)
	}
}

func TestLoad_ProductionEnvironment(t *testing.T) {
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(cfgPath, []byte(""), 0644); err != nil {
		t.Fatal(err)
	}

	secret := "this-is-a-very-long-secret-that-is-definitely-over-32-chars"
	t.Setenv("BANI_ENVIRONMENT", "production")
	t.Setenv("BANI_JWT_SECRET", secret)
	t.Setenv("BANI_DATABASE_DSN", "postgres://user:pass@db:5432/testdb?sslmode=require")
	t.Setenv("BANI_STORAGE_ACCESS_KEY", "prod-access-key")
	t.Setenv("BANI_STORAGE_SECRET_KEY", "prod-secret-key")
	t.Setenv("BANI_STORAGE_USE_SSL", "true")

	cfg, err := Load(cfgPath)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Environment != "production" {
		t.Errorf("expected environment = production, got %s", cfg.Environment)
	}
}

func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || index(s, substr) >= 0))
}

func TestLoad_AdminDefaults(t *testing.T) {
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(cfgPath, []byte(""), 0644); err != nil {
		t.Fatal(err)
	}

	t.Setenv("BANI_JWT_SECRET", "a-secure-secret-for-testing")

	cfg, err := Load(cfgPath)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Admin.Enabled != false {
		t.Errorf("expected admin.enabled = false, got %v", cfg.Admin.Enabled)
	}
	if cfg.Admin.Prefix != "/admin-panel" {
		t.Errorf("expected admin.prefix = /admin-panel, got %s", cfg.Admin.Prefix)
	}
	if cfg.Admin.Language != "ru" {
		t.Errorf("expected admin.language = ru, got %s", cfg.Admin.Language)
	}
	if cfg.Admin.Theme != "adminlte" {
		t.Errorf("expected admin.theme = adminlte, got %s", cfg.Admin.Theme)
	}
}

func TestLoad_AdminEnvOverride(t *testing.T) {
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(cfgPath, []byte(""), 0644); err != nil {
		t.Fatal(err)
	}

	t.Setenv("BANI_JWT_SECRET", "a-secure-secret-for-testing")
	t.Setenv("BANI_ADMIN_ENABLED", "true")
	t.Setenv("BANI_ADMIN_PREFIX", "/custom-admin")
	t.Setenv("BANI_ADMIN_LANGUAGE", "en")
	t.Setenv("BANI_ADMIN_THEME", "sword")

	cfg, err := Load(cfgPath)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Admin.Enabled != true {
		t.Errorf("expected admin.enabled = true, got %v", cfg.Admin.Enabled)
	}
	if cfg.Admin.Prefix != "/custom-admin" {
		t.Errorf("expected admin.prefix = /custom-admin, got %s", cfg.Admin.Prefix)
	}
	if cfg.Admin.Language != "en" {
		t.Errorf("expected admin.language = en, got %s", cfg.Admin.Language)
	}
	if cfg.Admin.Theme != "sword" {
		t.Errorf("expected admin.theme = sword, got %s", cfg.Admin.Theme)
	}
}

func TestLoad_AdminFromYAML(t *testing.T) {
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "config.yaml")

	yamlContent := `
jwt:
  secret: "my-secret-key"
admin:
  enabled: true
  prefix: "/my-admin"
  language: "en"
  theme: "sword"
`
	if err := os.WriteFile(cfgPath, []byte(yamlContent), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(cfgPath)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Admin.Enabled != true {
		t.Errorf("expected admin.enabled = true, got %v", cfg.Admin.Enabled)
	}
	if cfg.Admin.Prefix != "/my-admin" {
		t.Errorf("expected admin.prefix = /my-admin, got %s", cfg.Admin.Prefix)
	}
	if cfg.Admin.Language != "en" {
		t.Errorf("expected admin.language = en, got %s", cfg.Admin.Language)
	}
	if cfg.Admin.Theme != "sword" {
		t.Errorf("expected admin.theme = sword, got %s", cfg.Admin.Theme)
	}
}

func index(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
