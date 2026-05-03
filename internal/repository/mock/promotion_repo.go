package mock

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/rekurt/relax-hub/internal/domain"
)

// PromotionRepo is an in-memory mock implementation of repository.PromotionRepository.
type PromotionRepo struct {
	mu         sync.RWMutex
	promotions map[uuid.UUID]*domain.Promotion
}

func NewPromotionRepo() *PromotionRepo {
	return &PromotionRepo{promotions: make(map[uuid.UUID]*domain.Promotion)}
}

func (r *PromotionRepo) Create(_ context.Context, promo *domain.Promotion) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if promo.ID == uuid.Nil {
		promo.ID = uuid.New()
	}

	now := time.Now()
	promo.CreatedAt = now
	promo.UpdatedAt = now
	cp := *promo
	r.promotions[promo.ID] = &cp
	return nil
}

func (r *PromotionRepo) GetByID(_ context.Context, id uuid.UUID) (*domain.Promotion, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	promo, ok := r.promotions[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	cp := *promo
	return &cp, nil
}

func (r *PromotionRepo) GetActiveBybathhouse(_ context.Context, bathhouseID uuid.UUID) (*domain.Promotion, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var latest *domain.Promotion
	for _, promo := range r.promotions {
		if promo.BathhouseID == bathhouseID && promo.Status == domain.PromotionActive {
			if latest == nil || promo.CreatedAt.After(latest.CreatedAt) {
				cp := *promo
				latest = &cp
			}
		}
	}

	if latest == nil {
		return nil, domain.ErrNotFound
	}
	return latest, nil
}

func (r *PromotionRepo) Update(_ context.Context, promo *domain.Promotion) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.promotions[promo.ID]; !ok {
		return domain.ErrNotFound
	}

	cp := *promo
	r.promotions[promo.ID] = &cp
	return nil
}

func (r *PromotionRepo) ListByOwner(_ context.Context, ownerID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.Promotion], error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var items []domain.Promotion
	// In a real app, we'd join with bathhouses to find owner
	// For mock, we return empty since we don't have bathhouse context
	// This would be populated by test setup

	return paginate(items, page, pageSize), nil
}

func (r *PromotionRepo) RecordImpression(_ context.Context, promotionID uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	promo, ok := r.promotions[promotionID]
	if !ok {
		return domain.ErrNotFound
	}
	promo.ImpressionCount++
	return nil
}

func (r *PromotionRepo) RecordClick(_ context.Context, promotionID uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	promo, ok := r.promotions[promotionID]
	if !ok {
		return domain.ErrNotFound
	}
	promo.ClickCount++
	return nil
}

func (r *PromotionRepo) ListByBathhouse(_ context.Context, bathhouseID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.Promotion], error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var items []domain.Promotion
	for _, promo := range r.promotions {
		if promo.BathhouseID == bathhouseID {
			items = append(items, *promo)
		}
	}

	return paginate(items, page, pageSize), nil
}

func (r *PromotionRepo) ListAllActive(_ context.Context) ([]domain.Promotion, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var items []domain.Promotion
	for _, promo := range r.promotions {
		if promo.Status == domain.PromotionActive {
			items = append(items, *promo)
		}
	}
	return items, nil
}
