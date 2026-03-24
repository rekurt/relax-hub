package service

import (
	"context"
	"errors"
	"math"

	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/repository"
)

// ServiceFeeService calculates platform service fees applied to bookings.
type ServiceFeeService interface {
	// GetFeePercent resolves the fee percentage using priority: region+category > region > global default.
	GetFeePercent(ctx context.Context, region string, category *string) (float64, error)
	// CalculateFee computes the service fee in kopecks for a given base price.
	CalculateFee(ctx context.Context, basePrice int64, region string, category *string) (int64, error)
	// ListConfigs returns all service fee configs (admin).
	ListConfigs(ctx context.Context) ([]domain.ServiceFeeConfig, error)
	// UpsertConfig creates or updates a service fee config (admin).
	UpsertConfig(ctx context.Context, config *domain.ServiceFeeConfig) error
}

type serviceFeeService struct {
	repo repository.ServiceFeeRepository
}

func NewServiceFeeService(repo repository.ServiceFeeRepository) ServiceFeeService {
	return &serviceFeeService{repo: repo}
}

func (s *serviceFeeService) GetFeePercent(ctx context.Context, region string, category *string) (float64, error) {
	// Priority 1: region + category
	if category != nil {
		cfg, err := s.repo.GetByRegionAndCategory(ctx, region, category)
		if err == nil {
			return cfg.FeePercent, nil
		}
		if !errors.Is(err, domain.ErrNotFound) {
			return 0, err
		}
	}

	// Priority 2: region only
	cfg, err := s.repo.GetByRegion(ctx, region)
	if err == nil {
		return cfg.FeePercent, nil
	}
	if !errors.Is(err, domain.ErrNotFound) {
		return 0, err
	}

	// Priority 3: global default
	cfg, err = s.repo.GetGlobalDefault(ctx)
	if err != nil {
		return 0, err
	}
	return cfg.FeePercent, nil
}

func (s *serviceFeeService) CalculateFee(ctx context.Context, basePrice int64, region string, category *string) (int64, error) {
	pct, err := s.GetFeePercent(ctx, region, category)
	if err != nil {
		return 0, err
	}
	fee := int64(math.Round(float64(basePrice) * pct / 100))
	return fee, nil
}

func (s *serviceFeeService) ListConfigs(ctx context.Context) ([]domain.ServiceFeeConfig, error) {
	return s.repo.List(ctx)
}

func (s *serviceFeeService) UpsertConfig(ctx context.Context, config *domain.ServiceFeeConfig) error {
	if err := config.Validate(); err != nil {
		return err
	}
	return s.repo.Upsert(ctx, config)
}
