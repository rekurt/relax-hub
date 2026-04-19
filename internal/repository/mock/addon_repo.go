package mock

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/rekurt/relax-hub/internal/domain"
	"github.com/rekurt/relax-hub/internal/repository"
)

type AddOnRepo struct {
	mu            sync.RWMutex
	addons        map[uuid.UUID]*domain.AddOn
	bookingAddOns map[uuid.UUID]*domain.BookingAddOn
}

func NewAddOnRepo() *AddOnRepo {
	return &AddOnRepo{
		addons:        make(map[uuid.UUID]*domain.AddOn),
		bookingAddOns: make(map[uuid.UUID]*domain.BookingAddOn),
	}
}

var _ repository.AddOnRepository = (*AddOnRepo)(nil)

func (r *AddOnRepo) Create(_ context.Context, addon *domain.AddOn) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if addon.ID == uuid.Nil {
		addon.ID = uuid.New()
	}
	now := time.Now()
	if addon.CreatedAt.IsZero() {
		addon.CreatedAt = now
	}
	if addon.UpdatedAt.IsZero() {
		addon.UpdatedAt = now
	}
	if !addon.IsActive {
		addon.IsActive = true
	}

	cp := *addon
	r.addons[addon.ID] = &cp
	return nil
}

func (r *AddOnRepo) Update(_ context.Context, addon *domain.AddOn) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.addons[addon.ID]; !ok {
		return domain.ErrAddOnNotFound
	}
	cp := *addon
	r.addons[addon.ID] = &cp
	return nil
}

func (r *AddOnRepo) Delete(_ context.Context, id uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	a, ok := r.addons[id]
	if !ok {
		return domain.ErrAddOnNotFound
	}
	a.IsActive = false
	a.UpdatedAt = time.Now()
	return nil
}

func (r *AddOnRepo) GetByID(_ context.Context, id uuid.UUID) (*domain.AddOn, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	a, ok := r.addons[id]
	if !ok {
		return nil, domain.ErrAddOnNotFound
	}
	cp := *a
	return &cp, nil
}

func (r *AddOnRepo) ListByBathhouse(_ context.Context, bathhouseID uuid.UUID) ([]domain.AddOn, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []domain.AddOn
	for _, a := range r.addons {
		if a.BathhouseID == bathhouseID {
			cp := *a
			result = append(result, cp)
		}
	}
	return result, nil
}

func (r *AddOnRepo) ListActiveByBathhouse(_ context.Context, bathhouseID uuid.UUID) ([]domain.AddOn, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []domain.AddOn
	for _, a := range r.addons {
		if a.BathhouseID == bathhouseID && a.IsActive {
			cp := *a
			result = append(result, cp)
		}
	}
	return result, nil
}

func (r *AddOnRepo) CountByBathhouse(_ context.Context, bathhouseID uuid.UUID) (int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var count int64
	for _, a := range r.addons {
		if a.BathhouseID == bathhouseID && a.IsActive {
			count++
		}
	}
	return count, nil
}

func (r *AddOnRepo) CreateBookingAddOn(_ context.Context, ba *domain.BookingAddOn) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if ba.ID == uuid.Nil {
		ba.ID = uuid.New()
	}
	if ba.CreatedAt.IsZero() {
		ba.CreatedAt = time.Now()
	}

	cp := *ba
	r.bookingAddOns[ba.ID] = &cp
	return nil
}

func (r *AddOnRepo) ListByBooking(_ context.Context, bookingID uuid.UUID) ([]domain.BookingAddOn, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []domain.BookingAddOn
	for _, ba := range r.bookingAddOns {
		if ba.BookingID == bookingID {
			cp := *ba
			result = append(result, cp)
		}
	}
	return result, nil
}
