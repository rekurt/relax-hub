package mock

import (
	"context"
	"sync"
	"time"

	"github.com/rekurt/relax-hub/internal/domain"
)

type ReconciliationRepo struct {
	mu        sync.RWMutex
	snapshots []domain.FloatSnapshot
	reports   []domain.ReconciliationReport

	// configurable aggregate results
	ClientWalletsTotal int64
	ClientWalletsCount int
	OwnerWalletsTotal  int64
	OwnerWalletsCount  int
	EscrowHeldTotal    int64
	EscrowCount        int
	WalletHoldsTotal   int64
	PaymentsSum        int64
	PaymentsCount      int
	RefundsSum         int64
	RefundsCount       int
}

func NewReconciliationRepo() *ReconciliationRepo {
	return &ReconciliationRepo{}
}

func (r *ReconciliationRepo) CreateFloatSnapshot(_ context.Context, s *domain.FloatSnapshot) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.snapshots = append(r.snapshots, *s)
	return nil
}

func (r *ReconciliationRepo) GetLatestFloatSnapshot(_ context.Context) (*domain.FloatSnapshot, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if len(r.snapshots) == 0 {
		return nil, nil
	}
	s := r.snapshots[len(r.snapshots)-1]
	return &s, nil
}

func (r *ReconciliationRepo) ListFloatSnapshots(_ context.Context, from, to time.Time, page, pageSize int) (*domain.PaginatedResult[domain.FloatSnapshot], error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var items []domain.FloatSnapshot
	for _, s := range r.snapshots {
		if !s.SnapshotDate.Before(from) && !s.SnapshotDate.After(to) {
			items = append(items, s)
		}
	}
	total := int64(len(items))
	offset := (page - 1) * pageSize
	if offset >= len(items) {
		items = nil
	} else {
		end := offset + pageSize
		if end > len(items) {
			end = len(items)
		}
		items = items[offset:end]
	}
	return &domain.PaginatedResult[domain.FloatSnapshot]{
		Items: items, TotalCount: total, Page: page, PageSize: pageSize,
	}, nil
}

func (r *ReconciliationRepo) CreateReconciliationReport(_ context.Context, rpt *domain.ReconciliationReport) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.reports = append(r.reports, *rpt)
	return nil
}

func (r *ReconciliationRepo) GetLatestReconciliationReport(_ context.Context) (*domain.ReconciliationReport, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if len(r.reports) == 0 {
		return nil, nil
	}
	rpt := r.reports[len(r.reports)-1]
	return &rpt, nil
}

func (r *ReconciliationRepo) ListReconciliationReports(_ context.Context, page, pageSize int) (*domain.PaginatedResult[domain.ReconciliationReport], error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	total := int64(len(r.reports))
	offset := (page - 1) * pageSize
	items := r.reports
	if offset >= len(items) {
		items = nil
	} else {
		end := offset + pageSize
		if end > len(items) {
			end = len(items)
		}
		items = items[offset:end]
	}
	return &domain.PaginatedResult[domain.ReconciliationReport]{
		Items: items, TotalCount: total, Page: page, PageSize: pageSize,
	}, nil
}

func (r *ReconciliationRepo) SumWalletBalancesByRole(_ context.Context, role string) (int64, int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if role == "client" {
		return r.ClientWalletsTotal, r.ClientWalletsCount, nil
	}
	return r.OwnerWalletsTotal, r.OwnerWalletsCount, nil
}

func (r *ReconciliationRepo) SumEscrowHeld(_ context.Context) (int64, int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.EscrowHeldTotal, r.EscrowCount, nil
}

func (r *ReconciliationRepo) SumActiveWalletHolds(_ context.Context) (int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.WalletHoldsTotal, nil
}

func (r *ReconciliationRepo) SumPaymentsForPeriod(_ context.Context, _, _ time.Time) (int64, int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.PaymentsSum, r.PaymentsCount, nil
}

func (r *ReconciliationRepo) SumRefundsForPeriod(_ context.Context, _, _ time.Time) (int64, int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.RefundsSum, r.RefundsCount, nil
}
