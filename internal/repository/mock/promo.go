package mock

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/repository"
)

type PromoCodeRepo struct {
	mu     sync.RWMutex
	promos map[uuid.UUID]*domain.PromoCode
	usages map[uuid.UUID]*domain.PromoUsage
}

func NewPromoCodeRepo() repository.PromoCodeRepository {
	return &PromoCodeRepo{
		promos: make(map[uuid.UUID]*domain.PromoCode),
		usages: make(map[uuid.UUID]*domain.PromoUsage),
	}
}

func (r *PromoCodeRepo) Create(_ context.Context, promo *domain.PromoCode) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if promo.ID == uuid.Nil {
		promo.ID = uuid.New()
	}

	for _, p := range r.promos {
		if p.Code == promo.Code {
			return domain.ErrAlreadyExists
		}
	}

	if promo.CreatedAt.IsZero() {
		promo.CreatedAt = time.Now()
	}

	cp := *promo
	r.promos[promo.ID] = &cp
	return nil
}

func (r *PromoCodeRepo) GetByID(_ context.Context, id uuid.UUID) (*domain.PromoCode, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	promo, ok := r.promos[id]
	if !ok {
		return nil, domain.ErrPromoNotFound
	}
	cp := *promo
	return &cp, nil
}

func (r *PromoCodeRepo) GetByCode(_ context.Context, code string) (*domain.PromoCode, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, promo := range r.promos {
		if promo.Code == code {
			cp := *promo
			return &cp, nil
		}
	}
	return nil, domain.ErrPromoNotFound
}

func (r *PromoCodeRepo) Update(_ context.Context, promo *domain.PromoCode) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.promos[promo.ID]; !ok {
		return domain.ErrPromoNotFound
	}

	for _, p := range r.promos {
		if p.Code == promo.Code && p.ID != promo.ID {
			return domain.ErrAlreadyExists
		}
	}

	cp := *promo
	r.promos[promo.ID] = &cp
	return nil
}

func (r *PromoCodeRepo) ListByBathhouse(_ context.Context, bathhouseID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.PromoCode], error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var filtered []domain.PromoCode
	for _, promo := range r.promos {
		if promo.BathhouseID != nil && *promo.BathhouseID == bathhouseID {
			cp := *promo
			filtered = append(filtered, cp)
		}
	}

	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].CreatedAt.After(filtered[j].CreatedAt)
	})

	return paginate(filtered, page, pageSize), nil
}

func (r *PromoCodeRepo) ListByCreator(_ context.Context, creatorID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.PromoCode], error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var filtered []domain.PromoCode
	for _, promo := range r.promos {
		if promo.CreatorID == creatorID {
			cp := *promo
			filtered = append(filtered, cp)
		}
	}

	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].CreatedAt.After(filtered[j].CreatedAt)
	})

	return paginate(filtered, page, pageSize), nil
}

func (r *PromoCodeRepo) IncrementUses(_ context.Context, id uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	promo, ok := r.promos[id]
	if !ok {
		return domain.ErrPromoNotFound
	}
	if promo.MaxUses > 0 && promo.CurrentUses >= promo.MaxUses {
		return domain.ErrPromoMaxUses
	}
	promo.CurrentUses++
	return nil
}

func (r *PromoCodeRepo) RecordUsage(_ context.Context, usage *domain.PromoUsage) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if usage.ID == uuid.Nil {
		usage.ID = uuid.New()
	}
	if usage.UsedAt.IsZero() {
		usage.UsedAt = time.Now()
	}

	cp := *usage
	r.usages[usage.ID] = &cp
	return nil
}

func (r *PromoCodeRepo) ApplyUsage(_ context.Context, id uuid.UUID, usage *domain.PromoUsage) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	promo, ok := r.promos[id]
	if !ok {
		return domain.ErrPromoNotFound
	}
	if promo.MaxUses > 0 && promo.CurrentUses >= promo.MaxUses {
		return domain.ErrPromoMaxUses
	}
	promo.CurrentUses++

	if usage.ID == uuid.Nil {
		usage.ID = uuid.New()
	}
	if usage.UsedAt.IsZero() {
		usage.UsedAt = time.Now()
	}

	cp := *usage
	r.usages[usage.ID] = &cp
	return nil
}

func (r *PromoCodeRepo) GetUsageByBookingID(_ context.Context, bookingID uuid.UUID) (*domain.PromoUsage, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, usage := range r.usages {
		if usage.BookingID == bookingID {
			cp := *usage
			return &cp, nil
		}
	}
	return nil, nil
}

func (r *PromoCodeRepo) DecrementUses(_ context.Context, id uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	promo, ok := r.promos[id]
	if !ok {
		return domain.ErrPromoNotFound
	}
	if promo.CurrentUses == 0 {
		return domain.ErrPromoNotFound
	}
	promo.CurrentUses--
	return nil
}

func (r *PromoCodeRepo) DeleteUsage(_ context.Context, usageID uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.usages[usageID]; !ok {
		return domain.ErrPromoNotFound
	}
	delete(r.usages, usageID)
	return nil
}

func (r *PromoCodeRepo) RefundUsage(_ context.Context, promoCodeID uuid.UUID, usageID uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	promo, ok := r.promos[promoCodeID]
	if !ok {
		return domain.ErrPromoNotFound
	}
	if promo.CurrentUses == 0 {
		return domain.ErrPromoNotFound
	}

	if _, ok := r.usages[usageID]; !ok {
		return domain.ErrPromoNotFound
	}

	promo.CurrentUses--
	delete(r.usages, usageID)
	return nil
}

func (r *PromoCodeRepo) DeactivateExpired(_ context.Context, before time.Time) (int64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	var count int64
	for _, promo := range r.promos {
		if promo.IsActive && !promo.ValidUntil.IsZero() && promo.ValidUntil.Before(before) {
			promo.IsActive = false
			count++
		}
	}
	return count, nil
}
