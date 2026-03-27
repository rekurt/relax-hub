package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/logger"
	"github.com/nikitaaldaev/bani/internal/repository"
	"github.com/redis/go-redis/v9"
)

const (
	featureFlagCachePrefix = "feature:flags:"
	featureFlagCacheTTL    = 1 * time.Minute
)

// FeatureFlagService manages feature flag toggling with optional region scoping.
type FeatureFlagService interface {
	IsEnabled(ctx context.Context, key string) bool
	IsEnabledForRegion(ctx context.Context, key, region string) bool
	SetFlag(ctx context.Context, key string, enabled bool, region *string, adminID uuid.UUID) error
	GetAll(ctx context.Context) ([]domain.FeatureFlag, error)
}

type featureFlagService struct {
	repo   repository.FeatureFlagRepository
	redis  *redis.Client
	logger *logger.Logger
}

func NewFeatureFlagService(
	repo repository.FeatureFlagRepository,
	redisClient *redis.Client,
	log *logger.Logger,
) FeatureFlagService {
	return &featureFlagService{
		repo:   repo,
		redis:  redisClient,
		logger: log,
	}
}

func (s *featureFlagService) IsEnabled(ctx context.Context, key string) bool {
	flag, err := s.getFlag(ctx, key)
	if err != nil {
		s.logger.Error("failed to get feature flag", "key", key, "error", err)
		return false
	}
	// A flag with a region restriction is not globally enabled.
	if flag.Region != nil {
		return false
	}
	return flag.Enabled
}

func (s *featureFlagService) IsEnabledForRegion(ctx context.Context, key, region string) bool {
	flag, err := s.getFlag(ctx, key)
	if err != nil {
		s.logger.Error("failed to get feature flag for region", "key", key, "region", region, "error", err)
		return false
	}
	if !flag.Enabled {
		return false
	}
	// Global flag (no region) applies to all regions.
	if flag.Region == nil {
		return true
	}
	return *flag.Region == region
}

func (s *featureFlagService) SetFlag(ctx context.Context, key string, enabled bool, region *string, adminID uuid.UUID) error {
	if err := s.repo.Set(ctx, key, enabled, region, &adminID); err != nil {
		return err
	}
	// Invalidate cache (best-effort)
	cacheKey := featureFlagCachePrefix + key
	if err := s.redis.Del(ctx, cacheKey).Err(); err != nil {
		s.logger.Error("failed to invalidate feature flag cache", "key", key, "error", err)
	}
	return nil
}

func (s *featureFlagService) GetAll(ctx context.Context) ([]domain.FeatureFlag, error) {
	return s.repo.GetAll(ctx)
}

// getFlag retrieves a feature flag, checking Redis cache first.
func (s *featureFlagService) getFlag(ctx context.Context, key string) (*domain.FeatureFlag, error) {
	cacheKey := featureFlagCachePrefix + key

	// Try cache first (best-effort)
	cached, err := s.redis.HGetAll(ctx, cacheKey).Result()
	if err == nil && len(cached) > 0 {
		flag := &domain.FeatureFlag{
			Key:     key,
			Enabled: cached["enabled"] == "1",
		}
		if r, ok := cached["region"]; ok && r != "" {
			flag.Region = &r
		}
		return flag, nil
	}

	// Cache miss or error - fetch from DB
	flag, err := s.repo.Get(ctx, key)
	if err != nil {
		return nil, err
	}

	// Populate cache (best-effort)
	fields := map[string]interface{}{
		"enabled": boolToStr(flag.Enabled),
	}
	if flag.Region != nil {
		fields["region"] = *flag.Region
	} else {
		fields["region"] = ""
	}
	pipe := s.redis.Pipeline()
	pipe.HSet(ctx, cacheKey, fields)
	pipe.Expire(ctx, cacheKey, featureFlagCacheTTL)
	if _, cacheErr := pipe.Exec(ctx); cacheErr != nil {
		s.logger.Error("failed to cache feature flag", "key", key, "error", cacheErr)
	}

	return flag, nil
}

func boolToStr(b bool) string {
	if b {
		return "1"
	}
	return "0"
}

