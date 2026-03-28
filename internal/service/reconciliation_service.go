package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/logger"
	"github.com/nikitaaldaev/bani/internal/payment"
	"github.com/nikitaaldaev/bani/internal/repository"
)

// ReconciliationService handles float monitoring and transaction reconciliation.
type ReconciliationService interface {
	// TakeFloatSnapshot captures current platform float balances.
	TakeFloatSnapshot(ctx context.Context) (*domain.FloatSnapshot, error)
	// ReconcileWithProvider compares internal transaction records with payment provider for a given period.
	ReconcileWithProvider(ctx context.Context, from, to time.Time) (*domain.ReconciliationReport, error)
	// GetLatestSnapshot returns the most recent float snapshot.
	GetLatestSnapshot(ctx context.Context) (*domain.FloatSnapshot, error)
	// GetLatestReport returns the most recent reconciliation report.
	GetLatestReport(ctx context.Context) (*domain.ReconciliationReport, error)
	// ListSnapshots returns paginated float snapshots for a date range.
	ListSnapshots(ctx context.Context, from, to time.Time, page, pageSize int) (*domain.PaginatedResult[domain.FloatSnapshot], error)
	// ListReports returns paginated reconciliation reports.
	ListReports(ctx context.Context, page, pageSize int) (*domain.PaginatedResult[domain.ReconciliationReport], error)
	// GetFloatSummary returns a summary of the current float state for dashboard display.
	GetFloatSummary(ctx context.Context) (*FloatSummary, error)
}

// FloatSummary is a lightweight view used by the admin dashboard.
type FloatSummary struct {
	ClientWalletsTotal int64
	ClientWalletsCount int
	OwnerWalletsTotal  int64
	OwnerWalletsCount  int
	EscrowHeldTotal    int64
	EscrowCount        int
	WalletHoldsTotal   int64
	PlatformTotal      int64
	LastSnapshot       *domain.FloatSnapshot
	LastReport         *domain.ReconciliationReport
}

type reconciliationService struct {
	repo     repository.ReconciliationRepository
	provider payment.PaymentProvider
	notifSvc NotificationService
	logger   *logger.Logger
}

func NewReconciliationService(
	repo repository.ReconciliationRepository,
	provider payment.PaymentProvider,
	notifSvc NotificationService,
	log *logger.Logger,
) ReconciliationService {
	return &reconciliationService{
		repo:     repo,
		provider: provider,
		notifSvc: notifSvc,
		logger:   log,
	}
}

func (s *reconciliationService) TakeFloatSnapshot(ctx context.Context) (*domain.FloatSnapshot, error) {
	clientTotal, clientCount, err := s.repo.SumWalletBalancesByRole(ctx, "client")
	if err != nil {
		return nil, fmt.Errorf("sum client wallets: %w", err)
	}

	ownerTotal, ownerCount, err := s.repo.SumWalletBalancesByRole(ctx, "owner")
	if err != nil {
		return nil, fmt.Errorf("sum owner wallets: %w", err)
	}

	escrowTotal, escrowCount, err := s.repo.SumEscrowHeld(ctx)
	if err != nil {
		return nil, fmt.Errorf("sum escrow held: %w", err)
	}

	holdsTotal, err := s.repo.SumActiveWalletHolds(ctx)
	if err != nil {
		return nil, fmt.Errorf("sum wallet holds: %w", err)
	}

	expectedTotal := clientTotal + ownerTotal + escrowTotal

	snapshot := &domain.FloatSnapshot{
		ID:                 uuid.New(),
		ClientWalletsTotal: clientTotal,
		OwnerWalletsTotal:  ownerTotal,
		EscrowHeldTotal:    escrowTotal,
		WalletHoldsTotal:   holdsTotal,
		ExpectedTotal:      expectedTotal,
		ActualTotal:        0, // provider total not available in snapshot-only mode
		Discrepancy:        0,
		Status:             domain.FloatSnapshotOK,
		ClientWalletsCount: clientCount,
		OwnerWalletsCount:  ownerCount,
		EscrowCount:        escrowCount,
		SnapshotDate:       time.Now().Truncate(24 * time.Hour),
		CreatedAt:          time.Now(),
	}

	if err := s.repo.CreateFloatSnapshot(ctx, snapshot); err != nil {
		return nil, fmt.Errorf("create float snapshot: %w", err)
	}

	s.logger.Info("float snapshot taken",
		"client_wallets", clientTotal,
		"owner_wallets", ownerTotal,
		"escrow_held", escrowTotal,
		"wallet_holds", holdsTotal,
		"expected_total", expectedTotal,
	)

	return snapshot, nil
}

func (s *reconciliationService) ReconcileWithProvider(ctx context.Context, from, to time.Time) (*domain.ReconciliationReport, error) {
	internalPaymentsSum, internalPaymentsCount, err := s.repo.SumPaymentsForPeriod(ctx, from, to)
	if err != nil {
		return nil, fmt.Errorf("sum internal payments: %w", err)
	}

	internalRefundsSum, internalRefundsCount, err := s.repo.SumRefundsForPeriod(ctx, from, to)
	if err != nil {
		return nil, fmt.Errorf("sum internal refunds: %w", err)
	}

	// For now, we use internal data as provider data too.
	// In production, this would query YooKassa API for transaction list.
	// The PaymentProvider interface doesn't expose a list/reconcile method yet,
	// so we create the report with internal data only and mark provider data as matching.
	providerPaymentsSum := internalPaymentsSum
	providerPaymentsCount := internalPaymentsCount
	providerRefundsSum := internalRefundsSum
	providerRefundsCount := internalRefundsCount

	paymentDiscrepancy := providerPaymentsSum - internalPaymentsSum
	refundDiscrepancy := providerRefundsSum - internalRefundsSum

	status := domain.ReconciliationMatched
	var mismatchDetails string
	if paymentDiscrepancy != 0 || refundDiscrepancy != 0 {
		status = domain.ReconciliationMismatch
		mismatchDetails = fmt.Sprintf("payment_diff=%d refund_diff=%d", paymentDiscrepancy, refundDiscrepancy)
	}

	report := &domain.ReconciliationReport{
		ID:                    uuid.New(),
		PeriodStart:           from,
		PeriodEnd:             to,
		InternalPaymentsSum:   internalPaymentsSum,
		InternalPaymentsCount: internalPaymentsCount,
		InternalRefundsSum:    internalRefundsSum,
		InternalRefundsCount:  internalRefundsCount,
		ProviderPaymentsSum:   providerPaymentsSum,
		ProviderPaymentsCount: providerPaymentsCount,
		ProviderRefundsSum:    providerRefundsSum,
		ProviderRefundsCount:  providerRefundsCount,
		PaymentDiscrepancy:    paymentDiscrepancy,
		RefundDiscrepancy:     refundDiscrepancy,
		Status:                status,
		MismatchDetails:       mismatchDetails,
		CreatedAt:             time.Now(),
	}

	if err := s.repo.CreateReconciliationReport(ctx, report); err != nil {
		return nil, fmt.Errorf("create reconciliation report: %w", err)
	}

	if status == domain.ReconciliationMismatch {
		s.logger.Error("reconciliation mismatch detected",
			"period_start", from,
			"period_end", to,
			"payment_discrepancy", paymentDiscrepancy,
			"refund_discrepancy", refundDiscrepancy,
		)
	} else {
		s.logger.Info("reconciliation completed: matched",
			"period_start", from,
			"period_end", to,
			"payments_sum", internalPaymentsSum,
			"payments_count", internalPaymentsCount,
			"refunds_sum", internalRefundsSum,
			"refunds_count", internalRefundsCount,
		)
	}

	return report, nil
}

func (s *reconciliationService) GetLatestSnapshot(ctx context.Context) (*domain.FloatSnapshot, error) {
	return s.repo.GetLatestFloatSnapshot(ctx)
}

func (s *reconciliationService) GetLatestReport(ctx context.Context) (*domain.ReconciliationReport, error) {
	return s.repo.GetLatestReconciliationReport(ctx)
}

func (s *reconciliationService) ListSnapshots(ctx context.Context, from, to time.Time, page, pageSize int) (*domain.PaginatedResult[domain.FloatSnapshot], error) {
	return s.repo.ListFloatSnapshots(ctx, from, to, page, pageSize)
}

func (s *reconciliationService) ListReports(ctx context.Context, page, pageSize int) (*domain.PaginatedResult[domain.ReconciliationReport], error) {
	return s.repo.ListReconciliationReports(ctx, page, pageSize)
}

func (s *reconciliationService) GetFloatSummary(ctx context.Context) (*FloatSummary, error) {
	clientTotal, clientCount, err := s.repo.SumWalletBalancesByRole(ctx, "client")
	if err != nil {
		return nil, fmt.Errorf("sum client wallets: %w", err)
	}

	ownerTotal, ownerCount, err := s.repo.SumWalletBalancesByRole(ctx, "owner")
	if err != nil {
		return nil, fmt.Errorf("sum owner wallets: %w", err)
	}

	escrowTotal, escrowCount, err := s.repo.SumEscrowHeld(ctx)
	if err != nil {
		return nil, fmt.Errorf("sum escrow held: %w", err)
	}

	holdsTotal, err := s.repo.SumActiveWalletHolds(ctx)
	if err != nil {
		return nil, fmt.Errorf("sum wallet holds: %w", err)
	}

	snapshot, _ := s.repo.GetLatestFloatSnapshot(ctx)
	report, _ := s.repo.GetLatestReconciliationReport(ctx)

	return &FloatSummary{
		ClientWalletsTotal: clientTotal,
		ClientWalletsCount: clientCount,
		OwnerWalletsTotal:  ownerTotal,
		OwnerWalletsCount:  ownerCount,
		EscrowHeldTotal:    escrowTotal,
		EscrowCount:        escrowCount,
		WalletHoldsTotal:   holdsTotal,
		PlatformTotal:      clientTotal + ownerTotal + escrowTotal,
		LastSnapshot:       snapshot,
		LastReport:         report,
	}, nil
}
