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

type FraudFlagRepo struct {
	mu    sync.RWMutex
	flags map[uuid.UUID]*domain.FraudFlag
}

func NewFraudFlagRepo() repository.FraudFlagRepository {
	return &FraudFlagRepo{
		flags: make(map[uuid.UUID]*domain.FraudFlag),
	}
}

func (r *FraudFlagRepo) Create(_ context.Context, flag *domain.FraudFlag) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	if flag.ID == uuid.Nil {
		flag.ID = uuid.New()
	}
	if flag.Status == "" {
		flag.Status = domain.FraudFlagStatusPending
	}
	if flag.CreatedAt.IsZero() {
		flag.CreatedAt = now
	}

	cp := *flag
	r.flags[flag.ID] = &cp
	return nil
}

func (r *FraudFlagRepo) ListByUser(_ context.Context, userID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.FraudFlag], error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var filtered []domain.FraudFlag
	for _, f := range r.flags {
		if f.UserID == userID {
			cp := *f
			filtered = append(filtered, cp)
		}
	}

	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].CreatedAt.After(filtered[j].CreatedAt)
	})

	return paginate(filtered, page, pageSize), nil
}

func (r *FraudFlagRepo) ListPending(_ context.Context, filter domain.FraudFlagFilter) (*domain.PaginatedResult[domain.FraudFlag], error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	targetStatus := domain.FraudFlagStatusPending
	if filter.Status != nil {
		targetStatus = *filter.Status
	}

	var filtered []domain.FraudFlag
	for _, f := range r.flags {
		if f.Status != targetStatus {
			continue
		}
		if filter.UserID != nil && f.UserID != *filter.UserID {
			continue
		}
		if filter.Rule != nil && f.Rule != *filter.Rule {
			continue
		}
		cp := *f
		filtered = append(filtered, cp)
	}

	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].CreatedAt.After(filtered[j].CreatedAt)
	})

	return paginate(filtered, filter.Page, filter.PageSize), nil
}

func (r *FraudFlagRepo) UpdateStatus(_ context.Context, id uuid.UUID, status domain.FraudFlagStatus, reviewedBy uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	f, ok := r.flags[id]
	if !ok {
		return domain.ErrNotFound
	}
	f.Status = status
	now := time.Now()
	f.ReviewedAt = &now
	f.ReviewedBy = &reviewedBy
	return nil
}

func (r *FraudFlagRepo) CountByUserAndRule(_ context.Context, userID uuid.UUID, rule domain.FraudRuleName, since time.Time) (int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var count int64
	for _, f := range r.flags {
		if f.UserID == userID && f.Rule == rule && !f.CreatedAt.Before(since) {
			count++
		}
	}
	return count, nil
}
