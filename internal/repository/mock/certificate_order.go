package mock

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/rekurt/relax-hub/internal/domain"
	"github.com/rekurt/relax-hub/internal/repository"
)

type CertificateOrderRepo struct {
	mu     sync.RWMutex
	orders map[uuid.UUID]*domain.CertificateOrder
}

func NewCertificateOrderRepo() repository.CertificateOrderRepository {
	return &CertificateOrderRepo{
		orders: make(map[uuid.UUID]*domain.CertificateOrder),
	}
}

func (r *CertificateOrderRepo) Create(_ context.Context, order *domain.CertificateOrder) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if order.ID == uuid.Nil {
		order.ID = uuid.New()
	}
	if order.Status == "" {
		order.Status = domain.CertificateOrderStatusDraft
	}
	if order.CreatedAt.IsZero() {
		order.CreatedAt = time.Now()
	}
	if order.UpdatedAt.IsZero() {
		order.UpdatedAt = order.CreatedAt
	}

	cp := *order
	r.orders[order.ID] = &cp
	return nil
}

func (r *CertificateOrderRepo) GetByID(_ context.Context, id uuid.UUID) (*domain.CertificateOrder, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	order, ok := r.orders[id]
	if !ok {
		return nil, domain.ErrCertificateOrderNotFound
	}
	cp := *order
	return &cp, nil
}

func (r *CertificateOrderRepo) GetByExternalID(_ context.Context, externalID string) (*domain.CertificateOrder, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, order := range r.orders {
		if order.ExternalID == externalID {
			cp := *order
			return &cp, nil
		}
	}
	return nil, domain.ErrCertificateOrderNotFound
}

func (r *CertificateOrderRepo) UpdatePayment(_ context.Context, id uuid.UUID, status domain.CertificateOrderStatus, paymentMethod domain.PaymentMethod, provider, externalID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	order, ok := r.orders[id]
	if !ok {
		return domain.ErrCertificateOrderNotFound
	}
	order.Status = status
	order.PaymentMethod = paymentMethod
	order.Provider = provider
	order.ExternalID = externalID
	order.UpdatedAt = time.Now()
	return nil
}

func (r *CertificateOrderRepo) UpdateStatus(_ context.Context, id uuid.UUID, status domain.CertificateOrderStatus) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	order, ok := r.orders[id]
	if !ok {
		return domain.ErrCertificateOrderNotFound
	}
	order.Status = status
	order.UpdatedAt = time.Now()
	return nil
}

func (r *CertificateOrderRepo) MarkPaid(_ context.Context, id uuid.UUID, certificateID uuid.UUID, paidAt time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	order, ok := r.orders[id]
	if !ok {
		return domain.ErrCertificateOrderNotFound
	}
	order.Status = domain.CertificateOrderStatusPaid
	order.CertificateID = &certificateID
	order.PaidAt = &paidAt
	order.UpdatedAt = paidAt
	return nil
}
