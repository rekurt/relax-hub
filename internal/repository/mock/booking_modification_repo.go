package mock

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/rekurt/relax-hub/internal/domain"
)

type BookingModificationRequestRepo struct {
	mu       sync.RWMutex
	requests map[uuid.UUID]*domain.BookingModificationRequest
}

func NewBookingModificationRequestRepo() *BookingModificationRequestRepo {
	return &BookingModificationRequestRepo{
		requests: make(map[uuid.UUID]*domain.BookingModificationRequest),
	}
}

func (r *BookingModificationRequestRepo) Create(_ context.Context, req *domain.BookingModificationRequest) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	cp := *req
	r.requests[req.ID] = &cp
	return nil
}

func (r *BookingModificationRequestRepo) GetByID(_ context.Context, id uuid.UUID) (*domain.BookingModificationRequest, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	req, ok := r.requests[id]
	if !ok {
		return nil, domain.ErrModificationRequestNotFound
	}
	cp := *req
	return &cp, nil
}

func (r *BookingModificationRequestRepo) GetPendingByBookingID(_ context.Context, bookingID uuid.UUID) (*domain.BookingModificationRequest, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, req := range r.requests {
		if req.BookingID == bookingID && req.Status == domain.ModReqPending {
			cp := *req
			return &cp, nil
		}
	}
	return nil, domain.ErrModificationRequestNotFound
}

func (r *BookingModificationRequestRepo) UpdateStatus(_ context.Context, id uuid.UUID, status domain.ModificationRequestStatus, reason string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	req, ok := r.requests[id]
	if !ok {
		return domain.ErrModificationRequestNotFound
	}
	req.Status = status
	req.RejectionReason = reason
	now := time.Now()
	req.ResolvedAt = &now
	return nil
}

func (r *BookingModificationRequestRepo) ListExpired(_ context.Context) ([]domain.BookingModificationRequest, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	now := time.Now()
	var results []domain.BookingModificationRequest
	for _, req := range r.requests {
		if req.Status == domain.ModReqPending && req.ExpiresAt.Before(now) {
			results = append(results, *req)
		}
	}
	return results, nil
}

func (r *BookingModificationRequestRepo) ListByBookingID(_ context.Context, bookingID uuid.UUID) ([]domain.BookingModificationRequest, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var results []domain.BookingModificationRequest
	for _, req := range r.requests {
		if req.BookingID == bookingID {
			results = append(results, *req)
		}
	}
	return results, nil
}
