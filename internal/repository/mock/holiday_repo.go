package mock

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/rekurt/relax-hub/internal/domain"
	"github.com/rekurt/relax-hub/internal/repository"
)

type HolidayRepo struct {
	mu          sync.RWMutex
	holidays    map[uuid.UUID]*domain.Holiday
	multipliers map[uuid.UUID]float64 // bathhouseID -> multiplier
}

func NewHolidayRepo() repository.HolidayRepository {
	return &HolidayRepo{
		holidays:    make(map[uuid.UUID]*domain.Holiday),
		multipliers: make(map[uuid.UUID]float64),
	}
}

func (r *HolidayRepo) Create(_ context.Context, holiday *domain.Holiday) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if holiday.ID == uuid.Nil {
		holiday.ID = uuid.New()
	}
	now := time.Now()
	holiday.CreatedAt = now
	holiday.UpdatedAt = now

	cp := *holiday
	r.holidays[holiday.ID] = &cp
	return nil
}

func (r *HolidayRepo) Update(_ context.Context, holiday *domain.Holiday) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	existing, ok := r.holidays[holiday.ID]
	if !ok {
		return domain.ErrNotFound
	}

	holiday.UpdatedAt = time.Now()
	holiday.CreatedAt = existing.CreatedAt
	cp := *holiday
	r.holidays[holiday.ID] = &cp
	return nil
}

func (r *HolidayRepo) Delete(_ context.Context, id uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.holidays[id]; !ok {
		return domain.ErrNotFound
	}
	delete(r.holidays, id)
	return nil
}

func (r *HolidayRepo) GetByID(_ context.Context, id uuid.UUID) (*domain.Holiday, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	h, ok := r.holidays[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	cp := *h
	return &cp, nil
}

func (r *HolidayRepo) ListAll(_ context.Context) ([]domain.Holiday, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []domain.Holiday
	for _, h := range r.holidays {
		cp := *h
		result = append(result, cp)
	}
	return result, nil
}

func (r *HolidayRepo) ListByRegion(_ context.Context, region string) ([]domain.Holiday, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []domain.Holiday
	for _, h := range r.holidays {
		if h.Region == region {
			cp := *h
			result = append(result, cp)
		}
	}
	return result, nil
}

func (r *HolidayRepo) IsHoliday(_ context.Context, date time.Time, region string) (*domain.Holiday, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, h := range r.holidays {
		if h.Region != region {
			continue
		}
		// Check exact date match
		if h.Date.Year() == date.Year() && h.Date.Month() == date.Month() && h.Date.Day() == date.Day() {
			cp := *h
			return &cp, nil
		}
		// Check recurring match (same month-day regardless of year)
		if h.IsRecurring && h.Date.Month() == date.Month() && h.Date.Day() == date.Day() {
			cp := *h
			return &cp, nil
		}
	}
	return nil, domain.ErrNotFound
}

func (r *HolidayRepo) GetBathhouseMultiplier(_ context.Context, bathhouseID uuid.UUID) (float64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if m, ok := r.multipliers[bathhouseID]; ok {
		return m, nil
	}
	return domain.DefaultHolidayMultiplier, nil
}

func (r *HolidayRepo) SetBathhouseMultiplier(_ context.Context, bathhouseID uuid.UUID, multiplier float64) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.multipliers[bathhouseID] = multiplier
	return nil
}
