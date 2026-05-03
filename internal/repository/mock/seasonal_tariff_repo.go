package mock

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/rekurt/relax-hub/internal/domain"
)

// SeasonalTariffRepo is an in-memory mock implementation of repository.SeasonalTariffRepository.
type SeasonalTariffRepo struct {
	mu      sync.RWMutex
	tariffs map[uuid.UUID]*domain.SeasonalTariff
}

func NewSeasonalTariffRepo() *SeasonalTariffRepo {
	return &SeasonalTariffRepo{tariffs: make(map[uuid.UUID]*domain.SeasonalTariff)}
}

func (r *SeasonalTariffRepo) Create(_ context.Context, tariff *domain.SeasonalTariff) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if tariff.ID == uuid.Nil {
		tariff.ID = uuid.New()
	}
	now := time.Now()
	tariff.CreatedAt = now
	tariff.UpdatedAt = now
	cp := *tariff
	r.tariffs[tariff.ID] = &cp
	return nil
}

func (r *SeasonalTariffRepo) GetByID(_ context.Context, id uuid.UUID) (*domain.SeasonalTariff, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	t, ok := r.tariffs[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	cp := *t
	return &cp, nil
}

func (r *SeasonalTariffRepo) Update(_ context.Context, tariff *domain.SeasonalTariff) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.tariffs[tariff.ID]; !ok {
		return domain.ErrNotFound
	}
	tariff.UpdatedAt = time.Now()
	cp := *tariff
	r.tariffs[tariff.ID] = &cp
	return nil
}

func (r *SeasonalTariffRepo) Delete(_ context.Context, id uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.tariffs[id]; !ok {
		return domain.ErrNotFound
	}
	delete(r.tariffs, id)
	return nil
}

func (r *SeasonalTariffRepo) ListByBathhouse(_ context.Context, bathhouseID uuid.UUID) ([]domain.SeasonalTariff, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []domain.SeasonalTariff
	for _, t := range r.tariffs {
		if t.BathhouseID == bathhouseID {
			cp := *t
			result = append(result, cp)
		}
	}
	return result, nil
}

func (r *SeasonalTariffRepo) GetActiveTariffs(_ context.Context, bathhouseID uuid.UUID, date time.Time) ([]domain.SeasonalTariff, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []domain.SeasonalTariff
	for _, t := range r.tariffs {
		if t.BathhouseID == bathhouseID && t.AppliesToDate(date) {
			cp := *t
			result = append(result, cp)
		}
	}
	return result, nil
}
