package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/rekurt/relax-hub/internal/domain"
)

type PaymentRepository interface {
	Create(ctx context.Context, payment *domain.Payment) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Payment, error)
	GetByBookingID(ctx context.Context, bookingID uuid.UUID) (*domain.Payment, error)
	GetByExternalID(ctx context.Context, externalID string) (*domain.Payment, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status domain.PaymentStatus, externalID string) error
	UpdateRefund(ctx context.Context, id uuid.UUID, refundAmount int64, refundedAt time.Time, status domain.PaymentStatus) error
	UpdateCapture(ctx context.Context, id uuid.UUID, capturedAt time.Time, status domain.PaymentStatus) error
	Delete(ctx context.Context, id uuid.UUID) error
	ListByUser(ctx context.Context, userID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.Payment], error)
}

type WalletRepository interface {
	Create(ctx context.Context, wallet *domain.Wallet) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Wallet, error)
	GetByUserID(ctx context.Context, userID uuid.UUID) (*domain.Wallet, error)
	UpdateBalance(ctx context.Context, walletID uuid.UUID, oldBalance, newBalance int64, oldHeldAmount, newHeldAmount int64) error
	UpdateStatus(ctx context.Context, walletID uuid.UUID, status domain.WalletStatus) error
	ListAllIDs(ctx context.Context) ([]uuid.UUID, error)

	CreateTransaction(ctx context.Context, tx *domain.WalletTransaction) error
	ListTransactions(ctx context.Context, filter domain.WalletTransactionFilter) (*domain.PaginatedResult[domain.WalletTransaction], error)
	GetExpiringBonuses(ctx context.Context, walletID uuid.UUID, before time.Time) ([]domain.WalletTransaction, error)
	GetBonusTransactionsForSpending(ctx context.Context, walletID uuid.UUID) ([]domain.WalletTransaction, error)
	ExpireBonuses(ctx context.Context, transactionIDs []uuid.UUID) error
	GetExpiringBonusesSoon(ctx context.Context, walletID uuid.UUID, from, to time.Time) ([]domain.WalletTransaction, error)

	CreateHold(ctx context.Context, hold *domain.WalletHold) error
	GetHoldByID(ctx context.Context, holdID uuid.UUID) (*domain.WalletHold, error)
	UpdateHoldStatus(ctx context.Context, holdID uuid.UUID, status domain.WalletHoldStatus, capturedAt, releasedAt *time.Time) error
	GetActiveHolds(ctx context.Context, walletID uuid.UUID) ([]domain.WalletHold, error)
	GetExpiredHolds(ctx context.Context, before time.Time) ([]domain.WalletHold, error)
}

type PayoutRepository interface {
	Create(ctx context.Context, payout *domain.Payout) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Payout, error)
	ListByUser(ctx context.Context, userID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.Payout], error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status domain.PayoutStatus, processedAt *time.Time, failureReason string) error
	UpdateExternalID(ctx context.Context, id uuid.UUID, externalID string) error
	GetDailyTotal(ctx context.Context, userID uuid.UUID, date time.Time) (int64, error)
	GetMonthlyTotal(ctx context.Context, userID uuid.UUID, year int, month time.Month) (int64, error)
	GetPendingTotal(ctx context.Context, userID uuid.UUID) (int64, error)
	GetAutoPayoutSettings(ctx context.Context, userID uuid.UUID) (*domain.AutoPayoutSettings, error)
	UpsertAutoPayoutSettings(ctx context.Context, settings *domain.AutoPayoutSettings) error
	ListActiveAutoPayoutSettings(ctx context.Context) ([]domain.AutoPayoutSettings, error)
}

type SavedCardRepository interface {
	Create(ctx context.Context, card *domain.SavedCard) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.SavedCard, error)
	ListByUser(ctx context.Context, userID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.SavedCard], error)
	Delete(ctx context.Context, id uuid.UUID) error
	SetDefault(ctx context.Context, userID uuid.UUID, cardID uuid.UUID) error
	CountByUser(ctx context.Context, userID uuid.UUID) (int64, error)
}

type ServiceFeeRepository interface {
	GetByRegionAndCategory(ctx context.Context, region string, category *string) (*domain.ServiceFeeConfig, error)
	GetByRegion(ctx context.Context, region string) (*domain.ServiceFeeConfig, error)
	GetGlobalDefault(ctx context.Context) (*domain.ServiceFeeConfig, error)
	List(ctx context.Context) ([]domain.ServiceFeeConfig, error)
	Upsert(ctx context.Context, config *domain.ServiceFeeConfig) error
}

type ReconciliationRepository interface {
	CreateFloatSnapshot(ctx context.Context, snapshot *domain.FloatSnapshot) error
	GetLatestFloatSnapshot(ctx context.Context) (*domain.FloatSnapshot, error)
	ListFloatSnapshots(ctx context.Context, from, to time.Time, page, pageSize int) (*domain.PaginatedResult[domain.FloatSnapshot], error)

	CreateReconciliationReport(ctx context.Context, report *domain.ReconciliationReport) error
	GetLatestReconciliationReport(ctx context.Context) (*domain.ReconciliationReport, error)
	ListReconciliationReports(ctx context.Context, page, pageSize int) (*domain.PaginatedResult[domain.ReconciliationReport], error)

	// Aggregate queries for float calculation
	SumWalletBalancesByRole(ctx context.Context, role string) (total int64, count int, err error)
	SumEscrowHeld(ctx context.Context) (total int64, count int, err error)
	SumActiveWalletHolds(ctx context.Context) (total int64, err error)
	SumPaymentsForPeriod(ctx context.Context, from, to time.Time) (sum int64, count int, err error)
	SumRefundsForPeriod(ctx context.Context, from, to time.Time) (sum int64, count int, err error)
}

// BankReconciliationRepository manages bank statement entries and uploads.
type BankReconciliationRepository interface {
	CreateUpload(ctx context.Context, upload *domain.BankStatementUpload) error
	UpdateUploadCounts(ctx context.Context, id uuid.UUID, matched, pending, ignored int) error

	CreateEntries(ctx context.Context, entries []domain.BankStatementEntry) error
	GetEntryByID(ctx context.Context, id uuid.UUID) (*domain.BankStatementEntry, error)
	ListEntries(ctx context.Context, filter domain.BankStatementFilter) (*domain.PaginatedResult[domain.BankStatementEntry], error)
	MatchEntry(ctx context.Context, entryID uuid.UUID, txID uuid.UUID, txType string) error

	// FindPaymentsByAmountAndDate finds payments matching amount and date range for auto-matching.
	FindPaymentsByAmountAndDate(ctx context.Context, amount int64, dateFrom, dateTo time.Time) ([]domain.Payment, error)
}
