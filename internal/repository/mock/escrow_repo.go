package mock

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/rekurt/relax-hub/internal/domain"
	"github.com/rekurt/relax-hub/internal/repository"
)

type EscrowRepo struct {
	mu      sync.RWMutex
	escrows map[uuid.UUID]*domain.Escrow
}

func NewEscrowRepo() repository.EscrowRepository {
	return &EscrowRepo{
		escrows: make(map[uuid.UUID]*domain.Escrow),
	}
}

func (r *EscrowRepo) Create(_ context.Context, escrow *domain.Escrow) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	cp := *escrow
	r.escrows[escrow.ID] = &cp
	return nil
}

func (r *EscrowRepo) GetByID(_ context.Context, id uuid.UUID) (*domain.Escrow, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	e, ok := r.escrows[id]
	if !ok {
		return nil, domain.ErrEscrowNotFound
	}
	cp := *e
	return &cp, nil
}

func (r *EscrowRepo) GetByBookingID(_ context.Context, bookingID uuid.UUID) (*domain.Escrow, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, e := range r.escrows {
		if e.BookingID == bookingID {
			cp := *e
			return &cp, nil
		}
	}
	return nil, domain.ErrEscrowNotFound
}

func (r *EscrowRepo) UpdateStatus(_ context.Context, id uuid.UUID, status domain.EscrowStatus, releasedAt *time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	e, ok := r.escrows[id]
	if !ok {
		return domain.ErrEscrowNotFound
	}
	e.Status = status
	e.ReleasedAt = releasedAt
	return nil
}

func (r *EscrowRepo) ListMatured(_ context.Context) ([]domain.Escrow, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	now := time.Now()
	var result []domain.Escrow
	for _, e := range r.escrows {
		if e.Status == domain.EscrowHeld && e.ClaimPeriodEndsAt.Before(now) {
			cp := *e
			result = append(result, cp)
		}
	}
	return result, nil
}
