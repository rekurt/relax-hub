package mock

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
)

type ExtensionRequestRepo struct {
	mu       sync.RWMutex
	requests map[uuid.UUID]*domain.BookingExtensionRequest
}

func NewExtensionRequestRepo() *ExtensionRequestRepo {
	return &ExtensionRequestRepo{
		requests: make(map[uuid.UUID]*domain.BookingExtensionRequest),
	}
}

func (r *ExtensionRequestRepo) Create(_ context.Context, req *domain.BookingExtensionRequest) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	cp := *req
	r.requests[req.ID] = &cp
	return nil
}

func (r *ExtensionRequestRepo) GetByID(_ context.Context, id uuid.UUID) (*domain.BookingExtensionRequest, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	req, ok := r.requests[id]
	if !ok {
		return nil, domain.ErrExtensionRequestNotFound
	}
	cp := *req
	return &cp, nil
}

func (r *ExtensionRequestRepo) GetPendingByBookingID(_ context.Context, bookingID uuid.UUID) (*domain.BookingExtensionRequest, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, req := range r.requests {
		if req.BookingID == bookingID && req.Status == domain.ExtReqPending {
			cp := *req
			return &cp, nil
		}
	}
	return nil, domain.ErrExtensionRequestNotFound
}

func (r *ExtensionRequestRepo) UpdateStatus(_ context.Context, id uuid.UUID, status domain.ExtensionRequestStatus, reason string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	req, ok := r.requests[id]
	if !ok {
		return domain.ErrExtensionRequestNotFound
	}
	req.Status = status
	req.RejectionReason = reason
	now := time.Now()
	req.ResolvedAt = &now
	return nil
}

func (r *ExtensionRequestRepo) ListExpired(_ context.Context) ([]domain.BookingExtensionRequest, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	now := time.Now()
	var results []domain.BookingExtensionRequest
	for _, req := range r.requests {
		if req.Status == domain.ExtReqPending && req.ExpiresAt.Before(now) {
			results = append(results, *req)
		}
	}
	return results, nil
}

func (r *ExtensionRequestRepo) ListByBookingID(_ context.Context, bookingID uuid.UUID) ([]domain.BookingExtensionRequest, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var results []domain.BookingExtensionRequest
	for _, req := range r.requests {
		if req.BookingID == bookingID {
			results = append(results, *req)
		}
	}
	return results, nil
}
