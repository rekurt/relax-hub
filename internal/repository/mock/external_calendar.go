package mock

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
)

// ExternalCalendarRepo is an in-memory mock implementation of repository.ExternalCalendarRepository.
type ExternalCalendarRepo struct {
	mu        sync.RWMutex
	calendars map[uuid.UUID]*domain.ExternalCalendar
}

func NewExternalCalendarRepo() *ExternalCalendarRepo {
	return &ExternalCalendarRepo{calendars: make(map[uuid.UUID]*domain.ExternalCalendar)}
}

func (r *ExternalCalendarRepo) Create(_ context.Context, cal *domain.ExternalCalendar) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if cal.ID == uuid.Nil {
		cal.ID = uuid.New()
	}
	cal.CreatedAt = time.Now()
	cp := *cal
	r.calendars[cal.ID] = &cp
	return nil
}

func (r *ExternalCalendarRepo) GetByID(_ context.Context, id uuid.UUID) (*domain.ExternalCalendar, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	c, ok := r.calendars[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	cp := *c
	return &cp, nil
}

func (r *ExternalCalendarRepo) ListByBathhouse(_ context.Context, bathhouseID uuid.UUID) ([]domain.ExternalCalendar, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []domain.ExternalCalendar
	for _, c := range r.calendars {
		if c.BathhouseID == bathhouseID {
			cp := *c
			result = append(result, cp)
		}
	}
	return result, nil
}

func (r *ExternalCalendarRepo) ListAll(_ context.Context) ([]domain.ExternalCalendar, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []domain.ExternalCalendar
	for _, c := range r.calendars {
		cp := *c
		result = append(result, cp)
	}
	return result, nil
}

func (r *ExternalCalendarRepo) Delete(_ context.Context, id uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.calendars[id]; !ok {
		return domain.ErrNotFound
	}
	delete(r.calendars, id)
	return nil
}

func (r *ExternalCalendarRepo) UpdateSyncStatus(_ context.Context, id uuid.UUID, syncedAt time.Time, lastError string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	c, ok := r.calendars[id]
	if !ok {
		return domain.ErrNotFound
	}
	c.LastSyncAt = &syncedAt
	c.LastError = lastError
	return nil
}
