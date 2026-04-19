package mock

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/rekurt/relax-hub/internal/domain"
	"github.com/rekurt/relax-hub/internal/repository"
)

type PaymentRepo struct {
	mu       sync.RWMutex
	payments map[uuid.UUID]*domain.Payment
}

func NewPaymentRepo() repository.PaymentRepository {
	return &PaymentRepo{
		payments: make(map[uuid.UUID]*domain.Payment),
	}
}

func (r *PaymentRepo) Create(_ context.Context, payment *domain.Payment) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if payment.ID == uuid.Nil {
		payment.ID = uuid.New()
	}
	if payment.CreatedAt.IsZero() {
		payment.CreatedAt = time.Now()
	}
	if payment.UpdatedAt.IsZero() {
		payment.UpdatedAt = time.Now()
	}
	if payment.Metadata == nil {
		payment.Metadata = make(map[string]string)
	}

	cp := *payment
	cpMeta := make(map[string]string, len(payment.Metadata))
	for k, v := range payment.Metadata {
		cpMeta[k] = v
	}
	cp.Metadata = cpMeta
	r.payments[payment.ID] = &cp
	return nil
}

func (r *PaymentRepo) GetByID(_ context.Context, id uuid.UUID) (*domain.Payment, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	p, ok := r.payments[id]
	if !ok {
		return nil, domain.ErrPaymentNotFound
	}
	cp := r.copyPayment(p)
	return &cp, nil
}

func (r *PaymentRepo) GetByBookingID(_ context.Context, bookingID uuid.UUID) (*domain.Payment, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, p := range r.payments {
		if p.BookingID == bookingID {
			cp := r.copyPayment(p)
			return &cp, nil
		}
	}
	return nil, domain.ErrPaymentNotFound
}

func (r *PaymentRepo) GetByExternalID(_ context.Context, externalID string) (*domain.Payment, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, p := range r.payments {
		if p.ExternalID == externalID {
			cp := r.copyPayment(p)
			return &cp, nil
		}
	}
	return nil, domain.ErrPaymentNotFound
}

func (r *PaymentRepo) UpdateStatus(_ context.Context, id uuid.UUID, status domain.PaymentStatus, externalID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	p, ok := r.payments[id]
	if !ok {
		return domain.ErrPaymentNotFound
	}

	p.Status = status
	p.ExternalID = externalID
	p.UpdatedAt = time.Now()
	return nil
}

func (r *PaymentRepo) UpdateRefund(_ context.Context, id uuid.UUID, refundAmount int64, refundedAt time.Time, status domain.PaymentStatus) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	p, ok := r.payments[id]
	if !ok {
		return domain.ErrPaymentNotFound
	}

	p.RefundAmount = refundAmount
	p.RefundedAt = &refundedAt
	p.Status = status
	p.UpdatedAt = time.Now()
	return nil
}

func (r *PaymentRepo) UpdateCapture(_ context.Context, id uuid.UUID, capturedAt time.Time, status domain.PaymentStatus) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	p, ok := r.payments[id]
	if !ok {
		return domain.ErrPaymentNotFound
	}

	p.CapturedAt = &capturedAt
	p.IsHold = false
	p.Status = status
	p.UpdatedAt = time.Now()
	return nil
}

func (r *PaymentRepo) Delete(_ context.Context, id uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.payments[id]; !ok {
		return domain.ErrPaymentNotFound
	}
	delete(r.payments, id)
	return nil
}

func (r *PaymentRepo) ListByUser(_ context.Context, userID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.Payment], error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var filtered []domain.Payment
	for _, p := range r.payments {
		if p.UserID == userID {
			filtered = append(filtered, r.copyPayment(p))
		}
	}

	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].CreatedAt.After(filtered[j].CreatedAt)
	})

	return paginate(filtered, page, pageSize), nil
}

func (r *PaymentRepo) copyPayment(p *domain.Payment) domain.Payment {
	cp := *p
	if p.Metadata != nil {
		cpMeta := make(map[string]string, len(p.Metadata))
		for k, v := range p.Metadata {
			cpMeta[k] = v
		}
		cp.Metadata = cpMeta
	}
	if p.CapturedAt != nil {
		t := *p.CapturedAt
		cp.CapturedAt = &t
	}
	if p.RefundedAt != nil {
		t := *p.RefundedAt
		cp.RefundedAt = &t
	}
	return cp
}
