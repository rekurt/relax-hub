package mock

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/rekurt/relax-hub/internal/domain"
	"github.com/rekurt/relax-hub/internal/repository"
)

type PhotoOrderRepo struct {
	mu     sync.RWMutex
	orders map[uuid.UUID]*domain.PhotoOrder
}

func NewPhotoOrderRepo() repository.PhotoOrderRepository {
	return &PhotoOrderRepo{
		orders: make(map[uuid.UUID]*domain.PhotoOrder),
	}
}

func (r *PhotoOrderRepo) Create(_ context.Context, order *domain.PhotoOrder) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if order.ID == uuid.Nil {
		order.ID = uuid.New()
	}
	now := time.Now()
	if order.CreatedAt.IsZero() {
		order.CreatedAt = now
	}
	order.UpdatedAt = now

	cp := *order
	r.orders[order.ID] = &cp
	return nil
}

func (r *PhotoOrderRepo) GetByID(_ context.Context, id uuid.UUID) (*domain.PhotoOrder, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	o, ok := r.orders[id]
	if !ok {
		return nil, domain.ErrPhotoOrderNotFound
	}
	cp := *o
	return &cp, nil
}

func (r *PhotoOrderRepo) Update(_ context.Context, order *domain.PhotoOrder) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.orders[order.ID]; !ok {
		return domain.ErrPhotoOrderNotFound
	}
	order.UpdatedAt = time.Now()
	cp := *order
	r.orders[order.ID] = &cp
	return nil
}

func (r *PhotoOrderRepo) List(_ context.Context, filter domain.PhotoOrderFilter) (*domain.PaginatedResult[domain.PhotoOrder], error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var all []domain.PhotoOrder
	for _, o := range r.orders {
		if filter.OwnerID != nil && o.OwnerID != *filter.OwnerID {
			continue
		}
		if filter.BathhouseID != nil && o.BathhouseID != *filter.BathhouseID {
			continue
		}
		if filter.Status != nil && o.Status != *filter.Status {
			continue
		}
		if filter.Region != nil && o.Region != *filter.Region {
			continue
		}
		cp := *o
		all = append(all, cp)
	}

	page := filter.Page
	if page < 1 {
		page = 1
	}
	pageSize := filter.PageSize
	if pageSize < 1 {
		pageSize = 10
	}

	total := int64(len(all))
	start := (page - 1) * pageSize
	if start > int(total) {
		start = int(total)
	}
	end := start + pageSize
	if end > int(total) {
		end = int(total)
	}

	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}

	return &domain.PaginatedResult[domain.PhotoOrder]{
		Items:      all[start:end],
		TotalCount: total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}
