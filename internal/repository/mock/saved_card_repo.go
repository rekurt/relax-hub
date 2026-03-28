package mock

import (
	"context"
	"math"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/repository"
)

type SavedCardRepo struct {
	mu    sync.RWMutex
	items map[uuid.UUID]*domain.SavedCard
}

func NewSavedCardRepo() *SavedCardRepo {
	return &SavedCardRepo{
		items: make(map[uuid.UUID]*domain.SavedCard),
	}
}

var _ repository.SavedCardRepository = (*SavedCardRepo)(nil)

func (r *SavedCardRepo) Create(_ context.Context, card *domain.SavedCard) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if card.ID == uuid.Nil {
		card.ID = uuid.New()
	}
	if card.CreatedAt.IsZero() {
		card.CreatedAt = time.Now()
	}
	if card.UpdatedAt.IsZero() {
		card.UpdatedAt = time.Now()
	}
	cp := *card
	r.items[card.ID] = &cp
	return nil
}

func (r *SavedCardRepo) GetByID(_ context.Context, id uuid.UUID) (*domain.SavedCard, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	c, ok := r.items[id]
	if !ok {
		return nil, domain.ErrSavedCardNotFound
	}
	cp := *c
	return &cp, nil
}

func (r *SavedCardRepo) ListByUser(_ context.Context, userID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.SavedCard], error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}

	var all []domain.SavedCard
	for _, c := range r.items {
		if c.UserID == userID {
			cp := *c
			all = append(all, cp)
		}
	}

	sort.Slice(all, func(i, j int) bool {
		if all[i].IsDefault && !all[j].IsDefault {
			return true
		}
		if !all[i].IsDefault && all[j].IsDefault {
			return false
		}
		return all[i].CreatedAt.After(all[j].CreatedAt)
	})

	totalCount := int64(len(all))
	offset := (page - 1) * pageSize
	if offset >= len(all) {
		return &domain.PaginatedResult[domain.SavedCard]{
			Items:      nil,
			TotalCount: totalCount,
			Page:       page,
			PageSize:   pageSize,
			TotalPages: int(math.Ceil(float64(totalCount) / float64(pageSize))),
		}, nil
	}

	end := offset + pageSize
	if end > len(all) {
		end = len(all)
	}

	return &domain.PaginatedResult[domain.SavedCard]{
		Items:      all[offset:end],
		TotalCount: totalCount,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: int(math.Ceil(float64(totalCount) / float64(pageSize))),
	}, nil
}

func (r *SavedCardRepo) Delete(_ context.Context, id uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.items[id]; !ok {
		return domain.ErrSavedCardNotFound
	}
	delete(r.items, id)
	return nil
}

func (r *SavedCardRepo) SetDefault(_ context.Context, userID uuid.UUID, cardID uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	targetCard, ok := r.items[cardID]
	if !ok || targetCard.UserID != userID {
		return domain.ErrSavedCardNotFound
	}

	for _, c := range r.items {
		if c.UserID == userID && c.ID != cardID {
			c.IsDefault = false
		}
	}

	targetCard.IsDefault = true
	targetCard.UpdatedAt = time.Now()
	return nil
}

func (r *SavedCardRepo) CountByUser(_ context.Context, userID uuid.UUID) (int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var count int64
	for _, c := range r.items {
		if c.UserID == userID {
			count++
		}
	}
	return count, nil
}
