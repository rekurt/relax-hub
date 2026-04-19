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

type PayoutRepo struct {
	mu              sync.RWMutex
	payouts         map[uuid.UUID]*domain.Payout
	autoPaySettings map[uuid.UUID]*domain.AutoPayoutSettings
}

func NewPayoutRepo() repository.PayoutRepository {
	return &PayoutRepo{
		payouts:         make(map[uuid.UUID]*domain.Payout),
		autoPaySettings: make(map[uuid.UUID]*domain.AutoPayoutSettings),
	}
}

func (r *PayoutRepo) Create(_ context.Context, payout *domain.Payout) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if payout.ID == uuid.Nil {
		payout.ID = uuid.New()
	}
	now := time.Now()
	if payout.RequestedAt.IsZero() {
		payout.RequestedAt = now
	}
	if payout.CreatedAt.IsZero() {
		payout.CreatedAt = now
	}
	if payout.Status == "" {
		payout.Status = domain.PayoutStatusPending
	}

	cp := *payout
	r.payouts[payout.ID] = &cp
	return nil
}

func (r *PayoutRepo) GetByID(_ context.Context, id uuid.UUID) (*domain.Payout, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	p, ok := r.payouts[id]
	if !ok {
		return nil, domain.ErrPayoutNotFound
	}
	cp := *p
	return &cp, nil
}

func (r *PayoutRepo) ListByUser(_ context.Context, userID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.Payout], error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var filtered []domain.Payout
	for _, p := range r.payouts {
		if p.UserID == userID {
			cp := *p
			filtered = append(filtered, cp)
		}
	}

	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].RequestedAt.After(filtered[j].RequestedAt)
	})

	return paginate(filtered, page, pageSize), nil
}

func (r *PayoutRepo) UpdateStatus(_ context.Context, id uuid.UUID, status domain.PayoutStatus, processedAt *time.Time, failureReason string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	p, ok := r.payouts[id]
	if !ok {
		return domain.ErrPayoutNotFound
	}
	p.Status = status
	p.ProcessedAt = processedAt
	p.FailureReason = failureReason
	return nil
}

func (r *PayoutRepo) UpdateExternalID(_ context.Context, id uuid.UUID, externalID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	p, ok := r.payouts[id]
	if !ok {
		return domain.ErrPayoutNotFound
	}
	p.ExternalID = externalID
	return nil
}

func (r *PayoutRepo) GetDailyTotal(_ context.Context, userID uuid.UUID, date time.Time) (int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	var total int64
	for _, p := range r.payouts {
		if p.UserID == userID && !p.RequestedAt.Before(startOfDay) && p.RequestedAt.Before(endOfDay) &&
			(p.Status == domain.PayoutStatusPending || p.Status == domain.PayoutStatusProcessing || p.Status == domain.PayoutStatusCompleted) {
			total += p.Amount
		}
	}
	return total, nil
}

func (r *PayoutRepo) GetMonthlyTotal(_ context.Context, userID uuid.UUID, year int, month time.Month) (int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	startOfMonth := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)
	endOfMonth := startOfMonth.AddDate(0, 1, 0)

	var total int64
	for _, p := range r.payouts {
		if p.UserID == userID && !p.RequestedAt.Before(startOfMonth) && p.RequestedAt.Before(endOfMonth) &&
			(p.Status == domain.PayoutStatusPending || p.Status == domain.PayoutStatusProcessing || p.Status == domain.PayoutStatusCompleted) {
			total += p.Amount
		}
	}
	return total, nil
}

func (r *PayoutRepo) GetPendingTotal(_ context.Context, userID uuid.UUID) (int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var total int64
	for _, p := range r.payouts {
		if p.UserID == userID && (p.Status == domain.PayoutStatusPending || p.Status == domain.PayoutStatusProcessing) {
			total += p.Amount
		}
	}
	return total, nil
}

func (r *PayoutRepo) GetAutoPayoutSettings(_ context.Context, userID uuid.UUID) (*domain.AutoPayoutSettings, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	s, ok := r.autoPaySettings[userID]
	if !ok {
		return &domain.AutoPayoutSettings{UserID: userID, Threshold: 0}, nil
	}
	cp := *s
	return &cp, nil
}

func (r *PayoutRepo) ListActiveAutoPayoutSettings(_ context.Context) ([]domain.AutoPayoutSettings, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []domain.AutoPayoutSettings
	for _, s := range r.autoPaySettings {
		if s.Threshold > 0 {
			cp := *s
			result = append(result, cp)
		}
	}
	return result, nil
}

func (r *PayoutRepo) UpsertAutoPayoutSettings(_ context.Context, settings *domain.AutoPayoutSettings) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	settings.UpdatedAt = time.Now()
	cp := *settings
	r.autoPaySettings[settings.UserID] = &cp
	return nil
}
