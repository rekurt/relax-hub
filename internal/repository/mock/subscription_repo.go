package mock

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/rekurt/relax-hub/internal/domain"
)

// SubscriptionRepo is an in-memory mock implementation of repository.SubscriptionRepository.
type SubscriptionRepo struct {
	mu            sync.RWMutex
	subscriptions map[uuid.UUID]*domain.Subscription
}

func NewSubscriptionRepo() *SubscriptionRepo {
	return &SubscriptionRepo{subscriptions: make(map[uuid.UUID]*domain.Subscription)}
}

func (r *SubscriptionRepo) Create(_ context.Context, sub *domain.Subscription) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if sub.ID == uuid.Nil {
		sub.ID = uuid.New()
	}

	// Check if another active subscription exists for this bathhouse
	for _, existing := range r.subscriptions {
		if existing.BathhouseID == sub.BathhouseID && existing.Status == domain.SubscriptionActive {
			return domain.ErrAlreadyExists
		}
	}

	now := time.Now()
	sub.CreatedAt = now
	sub.UpdatedAt = now
	cp := *sub
	r.subscriptions[sub.ID] = &cp
	return nil
}

func (r *SubscriptionRepo) GetByID(_ context.Context, id uuid.UUID) (*domain.Subscription, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	sub, ok := r.subscriptions[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	cp := *sub
	return &cp, nil
}

func (r *SubscriptionRepo) GetActiveBybathhouse(_ context.Context, bathhouseID uuid.UUID) (*domain.Subscription, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var latest *domain.Subscription
	for _, sub := range r.subscriptions {
		if sub.BathhouseID == bathhouseID && sub.Status == domain.SubscriptionActive {
			if latest == nil || sub.CreatedAt.After(latest.CreatedAt) {
				cp := *sub
				latest = &cp
			}
		}
	}

	if latest == nil {
		return nil, domain.ErrNotFound
	}
	return latest, nil
}

func (r *SubscriptionRepo) Update(_ context.Context, sub *domain.Subscription) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.subscriptions[sub.ID]; !ok {
		return domain.ErrNotFound
	}

	sub.UpdatedAt = time.Now()
	cp := *sub
	r.subscriptions[sub.ID] = &cp
	return nil
}

func (r *SubscriptionRepo) ListByOwner(_ context.Context, ownerID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.Subscription], error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var items []domain.Subscription
	for _, sub := range r.subscriptions {
		if sub.OwnerID == ownerID {
			items = append(items, *sub)
		}
	}

	return paginate(items, page, pageSize), nil
}

func (r *SubscriptionRepo) GetExpiring(_ context.Context, before time.Time) ([]domain.Subscription, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []domain.Subscription
	for _, sub := range r.subscriptions {
		if sub.Status == domain.SubscriptionActive && sub.EndDate != nil && sub.EndDate.Before(before) || sub.EndDate.Equal(before) {
			result = append(result, *sub)
		}
	}
	return result, nil
}
