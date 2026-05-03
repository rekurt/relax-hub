package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/rekurt/relax-hub/internal/domain"
	"github.com/rekurt/relax-hub/internal/logger"
	"github.com/rekurt/relax-hub/internal/repository"
)

const (
	platformSettingsCachePrefix = "platform:settings:"
	platformSettingsCacheTTL    = 5 * time.Minute
)

// PlatformSettingsService manages platform-wide configuration settings.
type PlatformSettingsService interface {
	GetString(ctx context.Context, key string) (string, error)
	GetInt(ctx context.Context, key string) (int, error)
	GetFloat(ctx context.Context, key string) (float64, error)
	GetBool(ctx context.Context, key string) (bool, error)
	Set(ctx context.Context, key, value string, updatedBy uuid.UUID) error
	GetAll(ctx context.Context) ([]domain.PlatformSetting, error)
}

type platformSettingsService struct {
	repo   repository.PlatformSettingsRepository
	redis  *redis.Client
	logger *logger.Logger
}

func NewPlatformSettingsService(
	repo repository.PlatformSettingsRepository,
	redisClient *redis.Client,
	log *logger.Logger,
) PlatformSettingsService {
	return &platformSettingsService{
		repo:   repo,
		redis:  redisClient,
		logger: log,
	}
}

func (s *platformSettingsService) GetString(ctx context.Context, key string) (string, error) {
	return s.getValue(ctx, key)
}

func (s *platformSettingsService) GetInt(ctx context.Context, key string) (int, error) {
	val, err := s.getValue(ctx, key)
	if err != nil {
		return 0, err
	}
	n, err := strconv.Atoi(val)
	if err != nil {
		return 0, fmt.Errorf("setting %q is not a valid int: %w", key, domain.ErrInvalidInput)
	}
	return n, nil
}

func (s *platformSettingsService) GetFloat(ctx context.Context, key string) (float64, error) {
	val, err := s.getValue(ctx, key)
	if err != nil {
		return 0, err
	}
	f, err := strconv.ParseFloat(val, 64)
	if err != nil {
		return 0, fmt.Errorf("setting %q is not a valid float: %w", key, domain.ErrInvalidInput)
	}
	return f, nil
}

func (s *platformSettingsService) GetBool(ctx context.Context, key string) (bool, error) {
	val, err := s.getValue(ctx, key)
	if err != nil {
		return false, err
	}
	b, err := strconv.ParseBool(val)
	if err != nil {
		return false, fmt.Errorf("setting %q is not a valid bool: %w", key, domain.ErrInvalidInput)
	}
	return b, nil
}

func (s *platformSettingsService) Set(ctx context.Context, key, value string, updatedBy uuid.UUID) error {
	// Validate value against the setting's declared type
	setting, err := s.repo.Get(ctx, key)
	if err != nil {
		return err
	}
	if err := validateSettingValue(value, setting.Type); err != nil {
		return err
	}

	if err := s.repo.Set(ctx, key, value, &updatedBy); err != nil {
		return err
	}
	// Invalidate cache (best-effort)
	cacheKey := platformSettingsCachePrefix + key
	if err := s.redis.Del(ctx, cacheKey).Err(); err != nil {
		s.logger.Error("failed to invalidate platform setting cache", "key", key, "error", err)
	}
	return nil
}

func (s *platformSettingsService) GetAll(ctx context.Context) ([]domain.PlatformSetting, error) {
	return s.repo.GetAll(ctx)
}

// getValue retrieves a setting value, checking Redis cache first.
func (s *platformSettingsService) getValue(ctx context.Context, key string) (string, error) {
	cacheKey := platformSettingsCachePrefix + key

	// Try cache first (best-effort)
	cached, err := s.redis.Get(ctx, cacheKey).Result()
	if err == nil {
		return cached, nil
	}

	// Cache miss or error — fetch from DB
	setting, err := s.repo.Get(ctx, key)
	if err != nil {
		return "", err
	}

	// Populate cache (best-effort)
	if cacheErr := s.redis.Set(ctx, cacheKey, setting.Value, platformSettingsCacheTTL).Err(); cacheErr != nil {
		s.logger.Error("failed to cache platform setting", "key", key, "error", cacheErr)
	}

	return setting.Value, nil
}

// validateSettingValue checks that value can be parsed as the declared type.
func validateSettingValue(value string, settingType domain.SettingType) error {
	switch settingType {
	case domain.SettingTypeInt:
		if _, err := strconv.Atoi(value); err != nil {
			return fmt.Errorf("value %q is not a valid int: %w", value, domain.ErrInvalidInput)
		}
	case domain.SettingTypeFloat:
		if _, err := strconv.ParseFloat(value, 64); err != nil {
			return fmt.Errorf("value %q is not a valid float: %w", value, domain.ErrInvalidInput)
		}
	case domain.SettingTypeBool:
		if _, err := strconv.ParseBool(value); err != nil {
			return fmt.Errorf("value %q is not a valid bool: %w", value, domain.ErrInvalidInput)
		}
	case domain.SettingTypeJSON:
		if !json.Valid([]byte(value)) {
			return fmt.Errorf("value is not valid JSON: %w", domain.ErrInvalidInput)
		}
	case domain.SettingTypeString:
		// No validation needed
	default:
		return fmt.Errorf("unknown setting type %q: %w", settingType, domain.ErrInvalidInput)
	}
	return nil
}
