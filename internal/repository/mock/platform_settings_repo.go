package mock

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/repository"
)

type PlatformSettingsRepo struct {
	mu       sync.RWMutex
	settings map[string]*domain.PlatformSetting
}

func NewPlatformSettingsRepo() repository.PlatformSettingsRepository {
	return &PlatformSettingsRepo{
		settings: make(map[string]*domain.PlatformSetting),
	}
}

func (r *PlatformSettingsRepo) Get(_ context.Context, key string) (*domain.PlatformSetting, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	s, ok := r.settings[key]
	if !ok {
		return nil, domain.ErrNotFound
	}
	cp := *s
	return &cp, nil
}

func (r *PlatformSettingsRepo) GetAll(_ context.Context) ([]domain.PlatformSetting, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []domain.PlatformSetting
	for _, s := range r.settings {
		cp := *s
		result = append(result, cp)
	}
	return result, nil
}

func (r *PlatformSettingsRepo) Set(_ context.Context, key, value string, updatedBy *uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	s, ok := r.settings[key]
	if !ok {
		return domain.ErrNotFound
	}
	s.Value = value
	s.UpdatedAt = time.Now()
	if updatedBy != nil {
		str := updatedBy.String()
		s.UpdatedBy = &str
	}
	return nil
}

// Seed adds a setting to the mock repo (test helper).
func (r *PlatformSettingsRepo) Seed(key, value, description string, settingType domain.SettingType) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.settings[key] = &domain.PlatformSetting{
		Key:         key,
		Value:       value,
		Description: description,
		Type:        settingType,
		UpdatedAt:   time.Now(),
	}
}
