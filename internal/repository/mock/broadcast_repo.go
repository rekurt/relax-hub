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

type BroadcastRepo struct {
	mu         sync.RWMutex
	broadcasts map[uuid.UUID]*domain.Broadcast
}

func NewBroadcastRepo() repository.BroadcastRepository {
	return &BroadcastRepo{
		broadcasts: make(map[uuid.UUID]*domain.Broadcast),
	}
}

func (r *BroadcastRepo) Create(_ context.Context, broadcast *domain.Broadcast) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	if broadcast.ID == uuid.Nil {
		broadcast.ID = uuid.New()
	}
	if broadcast.CreatedAt.IsZero() {
		broadcast.CreatedAt = now
	}
	if broadcast.UpdatedAt.IsZero() {
		broadcast.UpdatedAt = now
	}

	cp := *broadcast
	r.broadcasts[broadcast.ID] = &cp
	return nil
}

func (r *BroadcastRepo) GetByID(_ context.Context, id uuid.UUID) (*domain.Broadcast, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	b, ok := r.broadcasts[id]
	if !ok {
		return nil, domain.ErrBroadcastNotFound
	}
	cp := *b
	return &cp, nil
}

func (r *BroadcastRepo) ListByOwner(_ context.Context, filter domain.BroadcastFilter) (*domain.PaginatedResult[domain.Broadcast], error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var filtered []domain.Broadcast
	for _, b := range r.broadcasts {
		if b.OwnerID != filter.OwnerID {
			continue
		}
		if filter.Status != nil && b.Status != *filter.Status {
			continue
		}
		cp := *b
		filtered = append(filtered, cp)
	}

	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].CreatedAt.After(filtered[j].CreatedAt)
	})

	return paginate(filtered, filter.Page, filter.PageSize), nil
}

func (r *BroadcastRepo) UpdateStatus(_ context.Context, id uuid.UUID, status domain.BroadcastStatus) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	b, ok := r.broadcasts[id]
	if !ok {
		return domain.ErrBroadcastNotFound
	}
	b.Status = status
	b.UpdatedAt = time.Now()
	if status == domain.BroadcastStatusSent || status == domain.BroadcastStatusFailed {
		now := time.Now()
		b.SentAt = &now
	}
	return nil
}

func (r *BroadcastRepo) UpdateStats(_ context.Context, id uuid.UUID, delivered, read, clicked int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	b, ok := r.broadcasts[id]
	if !ok {
		return domain.ErrBroadcastNotFound
	}
	b.Delivered = delivered
	b.Read = read
	b.Clicked = clicked
	b.UpdatedAt = time.Now()
	return nil
}

func (r *BroadcastRepo) CountRecentByOwner(_ context.Context, ownerID uuid.UUID, since time.Time) (int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var count int64
	for _, b := range r.broadcasts {
		if b.OwnerID != ownerID {
			continue
		}
		if (b.Status == domain.BroadcastStatusSending || b.Status == domain.BroadcastStatusSent) && !b.CreatedAt.Before(since) {
			count++
		}
	}
	return count, nil
}
