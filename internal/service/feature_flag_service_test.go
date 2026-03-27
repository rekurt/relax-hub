package service_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/repository/mock"
	"github.com/nikitaaldaev/bani/internal/service"

	"github.com/alicebob/miniredis/v2"
	"github.com/nikitaaldaev/bani/internal/logger"
	"github.com/redis/go-redis/v9"
)

func setupFeatureFlags(t *testing.T) (service.FeatureFlagService, *mock.FeatureFlagRepo, *miniredis.Miniredis) {
	t.Helper()
	repo := mock.NewFeatureFlagRepo().(*mock.FeatureFlagRepo)

	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("miniredis: %v", err)
	}
	t.Cleanup(mr.Close)

	redisClient := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	log := logger.New(logger.LevelInfo)

	svc := service.NewFeatureFlagService(repo, redisClient, log)
	return svc, repo, mr
}

func TestFeatureFlagService_IsEnabled_True(t *testing.T) {
	svc, repo, _ := setupFeatureFlags(t)
	repo.Seed("wallet_enabled", true, "Enable wallet", nil)

	if !svc.IsEnabled(context.Background(), "wallet_enabled") {
		t.Error("expected wallet_enabled to be true")
	}
}

func TestFeatureFlagService_IsEnabled_False(t *testing.T) {
	svc, repo, _ := setupFeatureFlags(t)
	repo.Seed("wallet_enabled", false, "Enable wallet", nil)

	if svc.IsEnabled(context.Background(), "wallet_enabled") {
		t.Error("expected wallet_enabled to be false")
	}
}

func TestFeatureFlagService_IsEnabled_Nonexistent(t *testing.T) {
	svc, _, _ := setupFeatureFlags(t)

	if svc.IsEnabled(context.Background(), "nonexistent") {
		t.Error("expected nonexistent flag to return false")
	}
}

func TestFeatureFlagService_IsEnabled_RegionScoped_ReturnsFalseGlobally(t *testing.T) {
	svc, repo, _ := setupFeatureFlags(t)
	region := "moscow"
	repo.Seed("sbp_payments_enabled", true, "Enable SBP", &region)

	// A region-scoped flag is NOT globally enabled
	if svc.IsEnabled(context.Background(), "sbp_payments_enabled") {
		t.Error("expected region-scoped flag to return false for global check")
	}
}

func TestFeatureFlagService_IsEnabledForRegion_GlobalFlag(t *testing.T) {
	svc, repo, _ := setupFeatureFlags(t)
	repo.Seed("wallet_enabled", true, "Enable wallet", nil)

	// A global flag applies to all regions
	if !svc.IsEnabledForRegion(context.Background(), "wallet_enabled", "moscow") {
		t.Error("expected global flag to be enabled for any region")
	}
	if !svc.IsEnabledForRegion(context.Background(), "wallet_enabled", "spb") {
		t.Error("expected global flag to be enabled for any region")
	}
}

func TestFeatureFlagService_IsEnabledForRegion_MatchingRegion(t *testing.T) {
	svc, repo, _ := setupFeatureFlags(t)
	region := "moscow"
	repo.Seed("sbp_payments_enabled", true, "Enable SBP", &region)

	if !svc.IsEnabledForRegion(context.Background(), "sbp_payments_enabled", "moscow") {
		t.Error("expected flag to be enabled for matching region")
	}
}

func TestFeatureFlagService_IsEnabledForRegion_NonMatchingRegion(t *testing.T) {
	svc, repo, _ := setupFeatureFlags(t)
	region := "moscow"
	repo.Seed("sbp_payments_enabled", true, "Enable SBP", &region)

	if svc.IsEnabledForRegion(context.Background(), "sbp_payments_enabled", "spb") {
		t.Error("expected flag to be disabled for non-matching region")
	}
}

func TestFeatureFlagService_IsEnabledForRegion_DisabledFlag(t *testing.T) {
	svc, repo, _ := setupFeatureFlags(t)
	repo.Seed("wallet_enabled", false, "Enable wallet", nil)

	if svc.IsEnabledForRegion(context.Background(), "wallet_enabled", "moscow") {
		t.Error("expected disabled flag to return false for region check")
	}
}

func TestFeatureFlagService_SetFlag(t *testing.T) {
	svc, repo, _ := setupFeatureFlags(t)
	repo.Seed("wallet_enabled", false, "Enable wallet", nil)

	adminID := uuid.New()
	if err := svc.SetFlag(context.Background(), "wallet_enabled", true, nil, adminID); err != nil {
		t.Fatalf("SetFlag: %v", err)
	}

	if !svc.IsEnabled(context.Background(), "wallet_enabled") {
		t.Error("expected wallet_enabled to be true after SetFlag")
	}
}

func TestFeatureFlagService_SetFlag_WithRegion(t *testing.T) {
	svc, repo, _ := setupFeatureFlags(t)
	repo.Seed("sbp_payments_enabled", false, "Enable SBP", nil)

	adminID := uuid.New()
	region := "moscow"
	if err := svc.SetFlag(context.Background(), "sbp_payments_enabled", true, &region, adminID); err != nil {
		t.Fatalf("SetFlag: %v", err)
	}

	if !svc.IsEnabledForRegion(context.Background(), "sbp_payments_enabled", "moscow") {
		t.Error("expected flag enabled for moscow")
	}
	if svc.IsEnabledForRegion(context.Background(), "sbp_payments_enabled", "spb") {
		t.Error("expected flag disabled for spb")
	}
}

func TestFeatureFlagService_SetFlag_NotFound(t *testing.T) {
	svc, _, _ := setupFeatureFlags(t)
	adminID := uuid.New()

	err := svc.SetFlag(context.Background(), "nonexistent", true, nil, adminID)
	if err == nil {
		t.Fatal("expected error for nonexistent flag")
	}
}

func TestFeatureFlagService_GetAll(t *testing.T) {
	svc, repo, _ := setupFeatureFlags(t)
	repo.Seed("flag1", true, "desc1", nil)
	repo.Seed("flag2", false, "desc2", nil)

	flags, err := svc.GetAll(context.Background())
	if err != nil {
		t.Fatalf("GetAll: %v", err)
	}
	if len(flags) != 2 {
		t.Errorf("expected 2 flags, got %d", len(flags))
	}
}

func TestFeatureFlagService_CacheInvalidation(t *testing.T) {
	svc, repo, mr := setupFeatureFlags(t)
	repo.Seed("cached_flag", true, "test", nil)

	// First call populates cache
	if !svc.IsEnabled(context.Background(), "cached_flag") {
		t.Error("expected cached_flag to be true")
	}

	// Verify cache is populated
	if !mr.Exists("feature:flags:cached_flag") {
		t.Error("cache not populated")
	}

	// Update the flag - should invalidate cache
	adminID := uuid.New()
	if err := svc.SetFlag(context.Background(), "cached_flag", false, nil, adminID); err != nil {
		t.Fatalf("SetFlag: %v", err)
	}

	// Cache should be invalidated
	if mr.Exists("feature:flags:cached_flag") {
		t.Error("cache should be invalidated after SetFlag")
	}

	// Next check should return updated value from DB
	if svc.IsEnabled(context.Background(), "cached_flag") {
		t.Error("expected cached_flag to be false after update")
	}
}

func TestFeatureFlagService_CacheHit(t *testing.T) {
	svc, repo, mr := setupFeatureFlags(t)
	repo.Seed("cached_flag", false, "test", nil)

	// Pre-populate cache with enabled=true
	mr.HSet("feature:flags:cached_flag", "enabled", "1")
	mr.HSet("feature:flags:cached_flag", "region", "")

	// Should return cached value (true), not DB value (false)
	if !svc.IsEnabled(context.Background(), "cached_flag") {
		t.Error("expected cached value (true) to be returned")
	}
}
