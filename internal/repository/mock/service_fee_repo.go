package mock

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/rekurt/relax-hub/internal/domain"
	"github.com/rekurt/relax-hub/internal/repository"
)

type ServiceFeeRepo struct {
	mu      sync.RWMutex
	configs map[uuid.UUID]*domain.ServiceFeeConfig
}

func NewServiceFeeRepo() repository.ServiceFeeRepository {
	return &ServiceFeeRepo{
		configs: make(map[uuid.UUID]*domain.ServiceFeeConfig),
	}
}

func (r *ServiceFeeRepo) GetByRegionAndCategory(_ context.Context, region string, category *string) (*domain.ServiceFeeConfig, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, c := range r.configs {
		if c.Region == region && ptrStringEqual(c.Category, category) {
			cp := *c
			return &cp, nil
		}
	}
	return nil, domain.ErrNotFound
}

func (r *ServiceFeeRepo) GetByRegion(_ context.Context, region string) (*domain.ServiceFeeConfig, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, c := range r.configs {
		if c.Region == region && c.Category == nil {
			cp := *c
			return &cp, nil
		}
	}
	return nil, domain.ErrNotFound
}

func (r *ServiceFeeRepo) GetGlobalDefault(_ context.Context) (*domain.ServiceFeeConfig, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, c := range r.configs {
		if c.Region == "*" && c.Category == nil {
			cp := *c
			return &cp, nil
		}
	}
	return nil, domain.ErrNotFound
}

func (r *ServiceFeeRepo) List(_ context.Context) ([]domain.ServiceFeeConfig, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []domain.ServiceFeeConfig
	for _, c := range r.configs {
		cp := *c
		result = append(result, cp)
	}
	return result, nil
}

func (r *ServiceFeeRepo) Upsert(_ context.Context, config *domain.ServiceFeeConfig) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Check for existing entry with same region+category
	for id, c := range r.configs {
		if c.Region == config.Region && ptrStringEqual(c.Category, config.Category) {
			c.FeePercent = config.FeePercent
			c.UpdatedAt = time.Now()
			config.ID = id
			config.CreatedAt = c.CreatedAt
			config.UpdatedAt = c.UpdatedAt
			return nil
		}
	}

	if config.ID == uuid.Nil {
		config.ID = uuid.New()
	}
	now := time.Now()
	config.CreatedAt = now
	config.UpdatedAt = now
	cp := *config
	r.configs[config.ID] = &cp
	return nil
}

func ptrStringEqual(a, b *string) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}
