package mock

import (
	"context"
	"math"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/repository"
)

type KYCRepo struct {
	mu   sync.RWMutex
	apps map[uuid.UUID]*domain.KYCApplication
}

func NewKYCRepo() *KYCRepo {
	return &KYCRepo{
		apps: make(map[uuid.UUID]*domain.KYCApplication),
	}
}

// Ensure KYCRepo implements KYCRepository
var _ repository.KYCRepository = (*KYCRepo)(nil)

func (r *KYCRepo) Create(_ context.Context, kyc *domain.KYCApplication) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if kyc.ID == uuid.Nil {
		kyc.ID = uuid.New()
	}
	now := time.Now()
	if kyc.CreatedAt.IsZero() {
		kyc.CreatedAt = now
	}
	if kyc.UpdatedAt.IsZero() {
		kyc.UpdatedAt = now
	}
	if kyc.SubmittedAt.IsZero() {
		kyc.SubmittedAt = now
	}

	cp := *kyc
	r.apps[kyc.ID] = &cp
	return nil
}

func (r *KYCRepo) GetByID(_ context.Context, id uuid.UUID) (*domain.KYCApplication, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	app, ok := r.apps[id]
	if !ok {
		return nil, domain.ErrKYCNotFound
	}
	cp := *app
	return &cp, nil
}

func (r *KYCRepo) GetByUserID(_ context.Context, userID uuid.UUID) (*domain.KYCApplication, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var latest *domain.KYCApplication
	for _, app := range r.apps {
		if app.UserID == userID {
			if latest == nil || app.CreatedAt.After(latest.CreatedAt) {
				cp := *app
				latest = &cp
			}
		}
	}
	if latest == nil {
		return nil, domain.ErrKYCNotFound
	}
	return latest, nil
}

func (r *KYCRepo) Update(_ context.Context, kyc *domain.KYCApplication) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.apps[kyc.ID]; !ok {
		return domain.ErrKYCNotFound
	}
	cp := *kyc
	r.apps[kyc.ID] = &cp
	return nil
}

func (r *KYCRepo) ListPending(_ context.Context, page, pageSize int) (*domain.PaginatedResult[domain.KYCApplication], error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}

	var pending []domain.KYCApplication
	for _, app := range r.apps {
		if app.Status == domain.KYCStatusPending {
			pending = append(pending, *app)
		}
	}

	totalCount := int64(len(pending))
	offset := (page - 1) * pageSize
	end := offset + pageSize
	if offset > len(pending) {
		offset = len(pending)
	}
	if end > len(pending) {
		end = len(pending)
	}

	return &domain.PaginatedResult[domain.KYCApplication]{
		Items:      pending[offset:end],
		TotalCount: totalCount,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: int(math.Ceil(float64(totalCount) / float64(pageSize))),
	}, nil
}

func (r *KYCRepo) Approve(_ context.Context, id uuid.UUID, reviewedBy uuid.UUID, expiresAt time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	app, ok := r.apps[id]
	if !ok || app.Status != domain.KYCStatusPending {
		return domain.ErrKYCNotFound
	}

	now := time.Now()
	app.Status = domain.KYCStatusApproved
	app.ReviewedAt = &now
	app.ReviewedBy = &reviewedBy
	app.ExpiresAt = &expiresAt
	app.UpdatedAt = now
	return nil
}

func (r *KYCRepo) Reject(_ context.Context, id uuid.UUID, reviewedBy uuid.UUID, reason string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	app, ok := r.apps[id]
	if !ok || app.Status != domain.KYCStatusPending {
		return domain.ErrKYCNotFound
	}

	now := time.Now()
	app.Status = domain.KYCStatusRejected
	app.ReviewedAt = &now
	app.ReviewedBy = &reviewedBy
	app.RejectionReason = reason
	app.UpdatedAt = now
	return nil
}

func (r *KYCRepo) ListExpiredApproved(_ context.Context, before time.Time) ([]domain.KYCApplication, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []domain.KYCApplication
	for _, app := range r.apps {
		if app.Status == domain.KYCStatusApproved && app.ExpiresAt != nil && app.ExpiresAt.Before(before) {
			result = append(result, *app)
		}
	}
	return result, nil
}
