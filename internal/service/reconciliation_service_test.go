package service

import (
	"context"
	"testing"
	"time"

	"github.com/rekurt/relax-hub/internal/domain"
	"github.com/rekurt/relax-hub/internal/logger"
	"github.com/rekurt/relax-hub/internal/repository/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestReconciliationService(repo *mock.ReconciliationRepo) ReconciliationService {
	log := logger.New(logger.LevelInfo)
	return NewReconciliationService(repo, nil, nil, log)
}

func TestTakeFloatSnapshot_OK(t *testing.T) {
	repo := mock.NewReconciliationRepo()
	repo.ClientWalletsTotal = 500_000
	repo.ClientWalletsCount = 10
	repo.OwnerWalletsTotal = 300_000
	repo.OwnerWalletsCount = 5
	repo.EscrowHeldTotal = 100_000
	repo.EscrowCount = 3
	repo.WalletHoldsTotal = 50_000

	svc := newTestReconciliationService(repo)
	snapshot, err := svc.TakeFloatSnapshot(context.Background())

	require.NoError(t, err)
	assert.Equal(t, int64(500_000), snapshot.ClientWalletsTotal)
	assert.Equal(t, int64(300_000), snapshot.OwnerWalletsTotal)
	assert.Equal(t, int64(100_000), snapshot.EscrowHeldTotal)
	assert.Equal(t, int64(50_000), snapshot.WalletHoldsTotal)
	assert.Equal(t, int64(900_000), snapshot.ExpectedTotal) // 500k + 300k + 100k
	assert.Equal(t, domain.FloatSnapshotOK, snapshot.Status)
	assert.Equal(t, 10, snapshot.ClientWalletsCount)
	assert.Equal(t, 5, snapshot.OwnerWalletsCount)
	assert.Equal(t, 3, snapshot.EscrowCount)
}

func TestTakeFloatSnapshot_Persisted(t *testing.T) {
	repo := mock.NewReconciliationRepo()
	repo.ClientWalletsTotal = 100_000
	repo.OwnerWalletsTotal = 200_000
	repo.EscrowHeldTotal = 50_000

	svc := newTestReconciliationService(repo)
	_, err := svc.TakeFloatSnapshot(context.Background())
	require.NoError(t, err)

	latest, err := svc.GetLatestSnapshot(context.Background())
	require.NoError(t, err)
	require.NotNil(t, latest)
	assert.Equal(t, int64(350_000), latest.ExpectedTotal)
}

func TestReconcileWithProvider_Matched(t *testing.T) {
	repo := mock.NewReconciliationRepo()
	repo.PaymentsSum = 1_000_000
	repo.PaymentsCount = 50
	repo.RefundsSum = 100_000
	repo.RefundsCount = 5

	svc := newTestReconciliationService(repo)
	from := time.Now().AddDate(0, 0, -1)
	to := time.Now()

	report, err := svc.ReconcileWithProvider(context.Background(), from, to)
	require.NoError(t, err)
	assert.Equal(t, domain.ReconciliationMatched, report.Status)
	assert.Equal(t, int64(1_000_000), report.InternalPaymentsSum)
	assert.Equal(t, 50, report.InternalPaymentsCount)
	assert.Equal(t, int64(100_000), report.InternalRefundsSum)
	assert.Equal(t, 5, report.InternalRefundsCount)
	assert.Equal(t, int64(0), report.PaymentDiscrepancy)
	assert.Equal(t, int64(0), report.RefundDiscrepancy)
}

func TestReconcileWithProvider_Persisted(t *testing.T) {
	repo := mock.NewReconciliationRepo()
	repo.PaymentsSum = 500_000
	repo.PaymentsCount = 20

	svc := newTestReconciliationService(repo)
	from := time.Now().AddDate(0, 0, -1)
	to := time.Now()

	_, err := svc.ReconcileWithProvider(context.Background(), from, to)
	require.NoError(t, err)

	latest, err := svc.GetLatestReport(context.Background())
	require.NoError(t, err)
	require.NotNil(t, latest)
	assert.Equal(t, domain.ReconciliationMatched, latest.Status)
}

func TestGetFloatSummary(t *testing.T) {
	repo := mock.NewReconciliationRepo()
	repo.ClientWalletsTotal = 500_000
	repo.ClientWalletsCount = 10
	repo.OwnerWalletsTotal = 300_000
	repo.OwnerWalletsCount = 5
	repo.EscrowHeldTotal = 100_000
	repo.EscrowCount = 3
	repo.WalletHoldsTotal = 50_000

	svc := newTestReconciliationService(repo)
	summary, err := svc.GetFloatSummary(context.Background())

	require.NoError(t, err)
	assert.Equal(t, int64(500_000), summary.ClientWalletsTotal)
	assert.Equal(t, int64(300_000), summary.OwnerWalletsTotal)
	assert.Equal(t, int64(100_000), summary.EscrowHeldTotal)
	assert.Equal(t, int64(50_000), summary.WalletHoldsTotal)
	assert.Equal(t, int64(900_000), summary.PlatformTotal)
	assert.Nil(t, summary.LastSnapshot)
	assert.Nil(t, summary.LastReport)
}

func TestGetFloatSummary_WithPriorSnapshot(t *testing.T) {
	repo := mock.NewReconciliationRepo()
	repo.ClientWalletsTotal = 100_000
	repo.OwnerWalletsTotal = 200_000
	repo.EscrowHeldTotal = 50_000

	svc := newTestReconciliationService(repo)

	// Take a snapshot first
	_, err := svc.TakeFloatSnapshot(context.Background())
	require.NoError(t, err)

	summary, err := svc.GetFloatSummary(context.Background())
	require.NoError(t, err)
	require.NotNil(t, summary.LastSnapshot)
	assert.Equal(t, domain.FloatSnapshotOK, summary.LastSnapshot.Status)
}

func TestListSnapshots_Pagination(t *testing.T) {
	repo := mock.NewReconciliationRepo()
	repo.ClientWalletsTotal = 100_000

	svc := newTestReconciliationService(repo)

	// Take 3 snapshots
	for i := 0; i < 3; i++ {
		_, err := svc.TakeFloatSnapshot(context.Background())
		require.NoError(t, err)
	}

	from := time.Now().AddDate(0, 0, -1)
	to := time.Now().AddDate(0, 0, 1)

	result, err := svc.ListSnapshots(context.Background(), from, to, 1, 2)
	require.NoError(t, err)
	assert.Equal(t, int64(3), result.TotalCount)
	assert.Len(t, result.Items, 2)

	result2, err := svc.ListSnapshots(context.Background(), from, to, 2, 2)
	require.NoError(t, err)
	assert.Len(t, result2.Items, 1)
}

func TestListReports_Pagination(t *testing.T) {
	repo := mock.NewReconciliationRepo()
	repo.PaymentsSum = 100_000

	svc := newTestReconciliationService(repo)

	from := time.Now().AddDate(0, 0, -1)
	to := time.Now()

	for i := 0; i < 3; i++ {
		_, err := svc.ReconcileWithProvider(context.Background(), from, to)
		require.NoError(t, err)
	}

	result, err := svc.ListReports(context.Background(), 1, 2)
	require.NoError(t, err)
	assert.Equal(t, int64(3), result.TotalCount)
	assert.Len(t, result.Items, 2)
}
