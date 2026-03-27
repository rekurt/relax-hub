package service_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/repository/mock"
	"github.com/nikitaaldaev/bani/internal/service"

	"github.com/alicebob/miniredis/v2"
	"github.com/nikitaaldaev/bani/internal/logger"
	"github.com/redis/go-redis/v9"
)

func setupPlatformSettings(t *testing.T) (service.PlatformSettingsService, *mock.PlatformSettingsRepo, *miniredis.Miniredis) {
	t.Helper()
	repo := mock.NewPlatformSettingsRepo().(*mock.PlatformSettingsRepo)

	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("miniredis: %v", err)
	}
	t.Cleanup(mr.Close)

	redisClient := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	log := logger.New(logger.LevelInfo)

	svc := service.NewPlatformSettingsService(repo, redisClient, log)
	return svc, repo, mr
}

func TestPlatformSettingsService_GetString(t *testing.T) {
	svc, repo, _ := setupPlatformSettings(t)
	repo.Seed("offer_version", "1.0", "Current offer version", domain.SettingTypeString)

	val, err := svc.GetString(context.Background(), "offer_version")
	if err != nil {
		t.Fatalf("GetString: %v", err)
	}
	if val != "1.0" {
		t.Errorf("expected '1.0', got %q", val)
	}
}

func TestPlatformSettingsService_GetInt(t *testing.T) {
	svc, repo, _ := setupPlatformSettings(t)
	repo.Seed("escrow_claim_hours", "48", "Escrow hold hours", domain.SettingTypeInt)

	val, err := svc.GetInt(context.Background(), "escrow_claim_hours")
	if err != nil {
		t.Fatalf("GetInt: %v", err)
	}
	if val != 48 {
		t.Errorf("expected 48, got %d", val)
	}
}

func TestPlatformSettingsService_GetInt_InvalidValue(t *testing.T) {
	svc, repo, _ := setupPlatformSettings(t)
	repo.Seed("bad_int", "not_a_number", "bad value", domain.SettingTypeInt)

	_, err := svc.GetInt(context.Background(), "bad_int")
	if err == nil {
		t.Fatal("expected error for invalid int")
	}
}

func TestPlatformSettingsService_GetFloat(t *testing.T) {
	svc, repo, _ := setupPlatformSettings(t)
	repo.Seed("service_fee_percent", "10.5", "Service fee", domain.SettingTypeFloat)

	val, err := svc.GetFloat(context.Background(), "service_fee_percent")
	if err != nil {
		t.Fatalf("GetFloat: %v", err)
	}
	if val != 10.5 {
		t.Errorf("expected 10.5, got %f", val)
	}
}

func TestPlatformSettingsService_GetBool(t *testing.T) {
	svc, repo, _ := setupPlatformSettings(t)
	repo.Seed("feature_on", "true", "A feature", domain.SettingTypeBool)
	repo.Seed("feature_off", "false", "Another feature", domain.SettingTypeBool)

	val, err := svc.GetBool(context.Background(), "feature_on")
	if err != nil {
		t.Fatalf("GetBool true: %v", err)
	}
	if !val {
		t.Error("expected true")
	}

	val, err = svc.GetBool(context.Background(), "feature_off")
	if err != nil {
		t.Fatalf("GetBool false: %v", err)
	}
	if val {
		t.Error("expected false")
	}
}

func TestPlatformSettingsService_GetNotFound(t *testing.T) {
	svc, _, _ := setupPlatformSettings(t)

	_, err := svc.GetString(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent key")
	}
}

func TestPlatformSettingsService_Set(t *testing.T) {
	svc, repo, _ := setupPlatformSettings(t)
	repo.Seed("escrow_claim_hours", "48", "Escrow hold hours", domain.SettingTypeInt)

	adminID := uuid.New()
	if err := svc.Set(context.Background(), "escrow_claim_hours", "72", adminID); err != nil {
		t.Fatalf("Set: %v", err)
	}

	// Verify updated value
	val, err := svc.GetInt(context.Background(), "escrow_claim_hours")
	if err != nil {
		t.Fatalf("GetInt after set: %v", err)
	}
	if val != 72 {
		t.Errorf("expected 72 after set, got %d", val)
	}
}

func TestPlatformSettingsService_Set_NotFound(t *testing.T) {
	svc, _, _ := setupPlatformSettings(t)
	adminID := uuid.New()

	err := svc.Set(context.Background(), "nonexistent", "val", adminID)
	if err == nil {
		t.Fatal("expected error for nonexistent key")
	}
}

func TestPlatformSettingsService_GetAll(t *testing.T) {
	svc, repo, _ := setupPlatformSettings(t)
	repo.Seed("key1", "val1", "desc1", domain.SettingTypeString)
	repo.Seed("key2", "val2", "desc2", domain.SettingTypeString)

	settings, err := svc.GetAll(context.Background())
	if err != nil {
		t.Fatalf("GetAll: %v", err)
	}
	if len(settings) != 2 {
		t.Errorf("expected 2 settings, got %d", len(settings))
	}
}

func TestPlatformSettingsService_CacheInvalidation(t *testing.T) {
	svc, repo, mr := setupPlatformSettings(t)
	repo.Seed("cached_key", "original", "test", domain.SettingTypeString)

	// First call populates cache
	val, err := svc.GetString(context.Background(), "cached_key")
	if err != nil {
		t.Fatalf("first GetString: %v", err)
	}
	if val != "original" {
		t.Errorf("expected 'original', got %q", val)
	}

	// Verify cache is populated
	cached, err := mr.Get("platform:settings:cached_key")
	if err != nil {
		t.Fatalf("cache not populated: %v", err)
	}
	if cached != "original" {
		t.Errorf("cache value: expected 'original', got %q", cached)
	}

	// Update the setting - should invalidate cache
	adminID := uuid.New()
	if err := svc.Set(context.Background(), "cached_key", "updated", adminID); err != nil {
		t.Fatalf("Set: %v", err)
	}

	// Cache should be invalidated
	if mr.Exists("platform:settings:cached_key") {
		t.Error("cache should be invalidated after Set")
	}

	// Next get should return updated value from DB
	val, err = svc.GetString(context.Background(), "cached_key")
	if err != nil {
		t.Fatalf("GetString after invalidation: %v", err)
	}
	if val != "updated" {
		t.Errorf("expected 'updated', got %q", val)
	}
}

func TestPlatformSettingsService_Set_TypeValidation(t *testing.T) {
	svc, repo, _ := setupPlatformSettings(t)
	adminID := uuid.New()
	ctx := context.Background()

	// Seed settings of different types
	repo.Seed("int_setting", "42", "An int", domain.SettingTypeInt)
	repo.Seed("float_setting", "3.14", "A float", domain.SettingTypeFloat)
	repo.Seed("bool_setting", "true", "A bool", domain.SettingTypeBool)
	repo.Seed("json_setting", `{"key":"val"}`, "A JSON", domain.SettingTypeJSON)
	repo.Seed("string_setting", "hello", "A string", domain.SettingTypeString)

	// Invalid int
	err := svc.Set(ctx, "int_setting", "not_a_number", adminID)
	if err == nil {
		t.Fatal("expected error setting non-int value for int setting")
	}

	// Valid int should succeed
	if err := svc.Set(ctx, "int_setting", "100", adminID); err != nil {
		t.Fatalf("valid int Set failed: %v", err)
	}

	// Invalid float
	err = svc.Set(ctx, "float_setting", "not_a_float", adminID)
	if err == nil {
		t.Fatal("expected error setting non-float value for float setting")
	}

	// Invalid bool
	err = svc.Set(ctx, "bool_setting", "maybe", adminID)
	if err == nil {
		t.Fatal("expected error setting non-bool value for bool setting")
	}

	// Invalid JSON
	err = svc.Set(ctx, "json_setting", "{invalid", adminID)
	if err == nil {
		t.Fatal("expected error setting invalid JSON for json setting")
	}

	// String accepts anything
	if err := svc.Set(ctx, "string_setting", "anything goes", adminID); err != nil {
		t.Fatalf("string Set failed: %v", err)
	}
}

func TestPlatformSettingsService_CacheHit(t *testing.T) {
	svc, repo, mr := setupPlatformSettings(t)
	repo.Seed("cached_key", "db_value", "test", domain.SettingTypeString)

	// Pre-populate cache with a different value
	mr.Set("platform:settings:cached_key", "cached_value")

	// Should return cached value, not DB value
	val, err := svc.GetString(context.Background(), "cached_key")
	if err != nil {
		t.Fatalf("GetString with cache: %v", err)
	}
	if val != "cached_value" {
		t.Errorf("expected 'cached_value' from cache, got %q", val)
	}
}
