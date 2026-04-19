package mock

import (
	"context"
	"sync"

	"github.com/google/uuid"
	"github.com/rekurt/relax-hub/internal/domain"
	"github.com/rekurt/relax-hub/internal/repository"
)

type PaymentDetailsRepo struct {
	mu      sync.RWMutex
	details map[uuid.UUID]*domain.PaymentDetails // keyed by user_id
}

func NewPaymentDetailsRepo() *PaymentDetailsRepo {
	return &PaymentDetailsRepo{
		details: make(map[uuid.UUID]*domain.PaymentDetails),
	}
}

var _ repository.PaymentDetailsRepository = (*PaymentDetailsRepo)(nil)

func (r *PaymentDetailsRepo) Upsert(_ context.Context, details *domain.PaymentDetails) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	cp := *details
	r.details[details.UserID] = &cp
	return nil
}

func (r *PaymentDetailsRepo) GetByUserID(_ context.Context, userID uuid.UUID) (*domain.PaymentDetails, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	pd, ok := r.details[userID]
	if !ok {
		return nil, domain.ErrPaymentDetailsNotFound
	}
	cp := *pd
	return &cp, nil
}
