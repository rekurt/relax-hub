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

type ReferralRepo struct {
	mu        sync.RWMutex
	referrals map[uuid.UUID]*domain.Referral
	balances  map[uuid.UUID]*domain.ReferralBalance
}

func NewReferralRepo() repository.ReferralRepository {
	return &ReferralRepo{
		referrals: make(map[uuid.UUID]*domain.Referral),
		balances:  make(map[uuid.UUID]*domain.ReferralBalance),
	}
}

func (r *ReferralRepo) Create(_ context.Context, referral *domain.Referral) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if referral.ID == uuid.Nil {
		referral.ID = uuid.New()
	}

	// Check unique constraint: one referral per referee
	for _, ref := range r.referrals {
		if ref.RefereeID == referral.RefereeID {
			return domain.ErrAlreadyReferred
		}
	}

	if referral.Status == "" {
		referral.Status = domain.ReferralStatusPending
	}
	if referral.CreatedAt.IsZero() {
		referral.CreatedAt = time.Now()
	}

	cp := *referral
	r.referrals[referral.ID] = &cp
	return nil
}

func (r *ReferralRepo) GetByReferee(_ context.Context, refereeID uuid.UUID) (*domain.Referral, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, ref := range r.referrals {
		if ref.RefereeID == refereeID {
			cp := *ref
			return &cp, nil
		}
	}
	return nil, domain.ErrNotFound
}

func (r *ReferralRepo) ListByReferrer(_ context.Context, referrerID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.Referral], error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}

	var filtered []domain.Referral
	for _, ref := range r.referrals {
		if ref.ReferrerID == referrerID {
			cp := *ref
			filtered = append(filtered, cp)
		}
	}

	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].CreatedAt.After(filtered[j].CreatedAt)
	})

	totalCount := int64(len(filtered))
	offset := (page - 1) * pageSize
	end := offset + pageSize
	if offset > int(totalCount) {
		offset = int(totalCount)
	}
	if end > int(totalCount) {
		end = int(totalCount)
	}

	return &domain.PaginatedResult[domain.Referral]{
		Items:      filtered[offset:end],
		TotalCount: totalCount,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: int(math.Ceil(float64(totalCount) / float64(pageSize))),
	}, nil
}

func (r *ReferralRepo) UpdateStatus(_ context.Context, id uuid.UUID, status domain.ReferralStatus, completedAt *time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	ref, ok := r.referrals[id]
	if !ok {
		return domain.ErrNotFound
	}

	// Match postgres behavior: only update if current status is pending
	if ref.Status != domain.ReferralStatusPending {
		return domain.ErrNotFound
	}

	ref.Status = status
	ref.CompletedAt = completedAt
	return nil
}

func (r *ReferralRepo) RevertToPending(_ context.Context, id uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	ref, ok := r.referrals[id]
	if !ok {
		return domain.ErrNotFound
	}

	if ref.Status != domain.ReferralStatusCompleted {
		return domain.ErrNotFound
	}

	ref.Status = domain.ReferralStatusPending
	ref.CompletedAt = nil
	return nil
}

func (r *ReferralRepo) GetBalance(_ context.Context, userID uuid.UUID) (*domain.ReferralBalance, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	b, ok := r.balances[userID]
	if !ok {
		return nil, domain.ErrNotFound
	}
	cp := *b
	return &cp, nil
}

func (r *ReferralRepo) CreateBalance(_ context.Context, balance *domain.ReferralBalance) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.balances[balance.UserID]; ok {
		return domain.ErrAlreadyExists
	}

	now := time.Now()
	balance.CreatedAt = now
	balance.UpdatedAt = now
	cp := *balance
	r.balances[balance.UserID] = &cp
	return nil
}

func (r *ReferralRepo) UpdateBalance(_ context.Context, userID uuid.UUID, delta int64, trackEarnings bool) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	b, ok := r.balances[userID]
	if !ok {
		return domain.ErrNotFound
	}

	if delta < 0 && b.Balance < -delta {
		return domain.ErrInsufficientReferralBalance
	}

	b.Balance += delta
	if delta > 0 && trackEarnings {
		b.TotalEarned += delta
	}
	b.UpdatedAt = time.Now()
	return nil
}

func (r *ReferralRepo) CountByReferrer(_ context.Context, referrerID uuid.UUID) (int, int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var totalInvited, totalCompleted int
	for _, ref := range r.referrals {
		if ref.ReferrerID == referrerID {
			totalInvited++
			if ref.Status == domain.ReferralStatusCompleted {
				totalCompleted++
			}
		}
	}
	return totalInvited, totalCompleted, nil
}
