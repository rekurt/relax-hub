package mock

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/rekurt/relax-hub/internal/domain"
	"github.com/rekurt/relax-hub/internal/repository"
)

type FeatureFlagRepo struct {
	mu    sync.RWMutex
	flags map[string]*domain.FeatureFlag
}

func NewFeatureFlagRepo() repository.FeatureFlagRepository {
	return &FeatureFlagRepo{
		flags: make(map[string]*domain.FeatureFlag),
	}
}

func (r *FeatureFlagRepo) Get(_ context.Context, key string) (*domain.FeatureFlag, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	f, ok := r.flags[key]
	if !ok {
		return nil, domain.ErrNotFound
	}
	cp := *f
	return &cp, nil
}

func (r *FeatureFlagRepo) GetAll(_ context.Context) ([]domain.FeatureFlag, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []domain.FeatureFlag
	for _, f := range r.flags {
		cp := *f
		result = append(result, cp)
	}
	return result, nil
}

func (r *FeatureFlagRepo) Set(_ context.Context, key string, enabled bool, region *string, updatedBy *uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	f, ok := r.flags[key]
	if !ok {
		return domain.ErrNotFound
	}
	f.Enabled = enabled
	f.Region = region
	f.UpdatedAt = time.Now()
	if updatedBy != nil {
		str := updatedBy.String()
		f.UpdatedBy = &str
	}
	return nil
}

// Seed adds a feature flag to the mock repo (test helper).
func (r *FeatureFlagRepo) Seed(key string, enabled bool, description string, region *string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.flags[key] = &domain.FeatureFlag{
		Key:         key,
		Enabled:     enabled,
		Description: description,
		Region:      region,
		UpdatedAt:   time.Now(),
	}
}
