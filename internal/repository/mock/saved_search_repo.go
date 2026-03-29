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

type SavedSearchRepo struct {
	mu    sync.RWMutex
	items map[uuid.UUID]*domain.SavedSearch
}

func NewSavedSearchRepo() *SavedSearchRepo {
	return &SavedSearchRepo{
		items: make(map[uuid.UUID]*domain.SavedSearch),
	}
}

var _ repository.SavedSearchRepository = (*SavedSearchRepo)(nil)

func (r *SavedSearchRepo) Create(_ context.Context, search *domain.SavedSearch) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if search.ID == uuid.Nil {
		search.ID = uuid.New()
	}
	if search.CreatedAt.IsZero() {
		search.CreatedAt = time.Now()
	}
	cp := *search
	r.items[search.ID] = &cp
	return nil
}

func (r *SavedSearchRepo) ListByUser(_ context.Context, userID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.SavedSearch], error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}

	var all []domain.SavedSearch
	for _, s := range r.items {
		if s.UserID == userID {
			cp := *s
			all = append(all, cp)
		}
	}

	sort.Slice(all, func(i, j int) bool {
		return all[i].CreatedAt.After(all[j].CreatedAt)
	})

	totalCount := int64(len(all))
	offset := (page - 1) * pageSize
	if offset >= len(all) {
		return &domain.PaginatedResult[domain.SavedSearch]{
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

	return &domain.PaginatedResult[domain.SavedSearch]{
		Items:      all[offset:end],
		TotalCount: totalCount,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: int(math.Ceil(float64(totalCount) / float64(pageSize))),
	}, nil
}

func (r *SavedSearchRepo) Delete(_ context.Context, id uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.items[id]; !ok {
		return domain.ErrSavedSearchNotFound
	}
	delete(r.items, id)
	return nil
}

func (r *SavedSearchRepo) GetByID(_ context.Context, id uuid.UUID) (*domain.SavedSearch, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	s, ok := r.items[id]
	if !ok {
		return nil, domain.ErrSavedSearchNotFound
	}
	cp := *s
	return &cp, nil
}

func (r *SavedSearchRepo) ListWithNotifications(_ context.Context) ([]domain.SavedSearch, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []domain.SavedSearch
	for _, s := range r.items {
		if s.NotifyOnNew {
			cp := *s
			result = append(result, cp)
		}
	}
	return result, nil
}
