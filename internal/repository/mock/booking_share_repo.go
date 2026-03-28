package mock

import (
	"context"
	"sync"
	"time"

	"github.com/nikitaaldaev/bani/internal/domain"
)

type BookingShareRepo struct {
	mu     sync.RWMutex
	shares map[string]*domain.BookingShare // token -> share
}

func NewBookingShareRepo() *BookingShareRepo {
	return &BookingShareRepo{shares: make(map[string]*domain.BookingShare)}
}

func (r *BookingShareRepo) Create(_ context.Context, share *domain.BookingShare) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.shares[share.Token] = share
	return nil
}

func (r *BookingShareRepo) GetByToken(_ context.Context, token string) (*domain.BookingShare, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	share, ok := r.shares[token]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return share, nil
}

func (r *BookingShareRepo) DeleteExpired(_ context.Context) (int64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	now := time.Now()
	var count int64
	for token, share := range r.shares {
		if share.ExpiresAt.Before(now) {
			delete(r.shares, token)
			count++
		}
	}
	return count, nil
}
