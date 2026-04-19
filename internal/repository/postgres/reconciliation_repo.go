package postgres

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rekurt/relax-hub/internal/domain"
)

type reconciliationRepo struct {
	pool *pgxpool.Pool
}

func NewReconciliationRepo(pool *pgxpool.Pool) *reconciliationRepo {
	return &reconciliationRepo{pool: pool}
}

func (r *reconciliationRepo) CreateFloatSnapshot(ctx context.Context, s *domain.FloatSnapshot) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	if s.CreatedAt.IsZero() {
		s.CreatedAt = time.Now()
	}
	_, err := r.pool.Exec(ctx, `
		INSERT INTO float_snapshots (
			id, client_wallets_total, owner_wallets_total, escrow_held_total,
			wallet_holds_total, expected_total, actual_total, discrepancy,
			status, client_wallets_count, owner_wallets_count, escrow_count,
			notes, snapshot_date, created_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)`,
		s.ID, s.ClientWalletsTotal, s.OwnerWalletsTotal, s.EscrowHeldTotal,
		s.WalletHoldsTotal, s.ExpectedTotal, s.ActualTotal, s.Discrepancy,
		s.Status, s.ClientWalletsCount, s.OwnerWalletsCount, s.EscrowCount,
		s.Notes, s.SnapshotDate, s.CreatedAt,
	)
	return err
}

func (r *reconciliationRepo) GetLatestFloatSnapshot(ctx context.Context) (*domain.FloatSnapshot, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, client_wallets_total, owner_wallets_total, escrow_held_total,
			wallet_holds_total, expected_total, actual_total, discrepancy,
			status, client_wallets_count, owner_wallets_count, escrow_count,
			notes, snapshot_date, created_at
		FROM float_snapshots ORDER BY snapshot_date DESC LIMIT 1`)
	return scanFloatSnapshot(row)
}

func (r *reconciliationRepo) ListFloatSnapshots(ctx context.Context, from, to time.Time, page, pageSize int) (*domain.PaginatedResult[domain.FloatSnapshot], error) {
	var totalCount int64
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM float_snapshots WHERE snapshot_date >= $1 AND snapshot_date <= $2`,
		from, to,
	).Scan(&totalCount)
	if err != nil {
		return nil, err
	}

	offset := (page - 1) * pageSize
	rows, err := r.pool.Query(ctx, `
		SELECT id, client_wallets_total, owner_wallets_total, escrow_held_total,
			wallet_holds_total, expected_total, actual_total, discrepancy,
			status, client_wallets_count, owner_wallets_count, escrow_count,
			notes, snapshot_date, created_at
		FROM float_snapshots
		WHERE snapshot_date >= $1 AND snapshot_date <= $2
		ORDER BY snapshot_date DESC
		LIMIT $3 OFFSET $4`,
		from, to, pageSize, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []domain.FloatSnapshot
	for rows.Next() {
		s, err := scanFloatSnapshotRow(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, *s)
	}

	return &domain.PaginatedResult[domain.FloatSnapshot]{
		Items:      items,
		TotalCount: totalCount,
		Page:       page,
		PageSize:   pageSize,
	}, nil
}

func (r *reconciliationRepo) CreateReconciliationReport(ctx context.Context, rpt *domain.ReconciliationReport) error {
	if rpt.ID == uuid.Nil {
		rpt.ID = uuid.New()
	}
	if rpt.CreatedAt.IsZero() {
		rpt.CreatedAt = time.Now()
	}
	_, err := r.pool.Exec(ctx, `
		INSERT INTO reconciliation_reports (
			id, period_start, period_end,
			internal_payments_sum, internal_payments_count,
			internal_refunds_sum, internal_refunds_count,
			provider_payments_sum, provider_payments_count,
			provider_refunds_sum, provider_refunds_count,
			payment_discrepancy, refund_discrepancy,
			status, mismatch_details, error_message, created_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17)`,
		rpt.ID, rpt.PeriodStart, rpt.PeriodEnd,
		rpt.InternalPaymentsSum, rpt.InternalPaymentsCount,
		rpt.InternalRefundsSum, rpt.InternalRefundsCount,
		rpt.ProviderPaymentsSum, rpt.ProviderPaymentsCount,
		rpt.ProviderRefundsSum, rpt.ProviderRefundsCount,
		rpt.PaymentDiscrepancy, rpt.RefundDiscrepancy,
		rpt.Status, rpt.MismatchDetails, rpt.ErrorMessage, rpt.CreatedAt,
	)
	return err
}

func (r *reconciliationRepo) GetLatestReconciliationReport(ctx context.Context) (*domain.ReconciliationReport, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, period_start, period_end,
			internal_payments_sum, internal_payments_count,
			internal_refunds_sum, internal_refunds_count,
			provider_payments_sum, provider_payments_count,
			provider_refunds_sum, provider_refunds_count,
			payment_discrepancy, refund_discrepancy,
			status, mismatch_details, error_message, created_at
		FROM reconciliation_reports ORDER BY created_at DESC LIMIT 1`)
	return scanReconciliationReport(row)
}

func (r *reconciliationRepo) ListReconciliationReports(ctx context.Context, page, pageSize int) (*domain.PaginatedResult[domain.ReconciliationReport], error) {
	var totalCount int64
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM reconciliation_reports`).Scan(&totalCount)
	if err != nil {
		return nil, err
	}

	offset := (page - 1) * pageSize
	rows, err := r.pool.Query(ctx, `
		SELECT id, period_start, period_end,
			internal_payments_sum, internal_payments_count,
			internal_refunds_sum, internal_refunds_count,
			provider_payments_sum, provider_payments_count,
			provider_refunds_sum, provider_refunds_count,
			payment_discrepancy, refund_discrepancy,
			status, mismatch_details, error_message, created_at
		FROM reconciliation_reports
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2`,
		pageSize, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []domain.ReconciliationReport
	for rows.Next() {
		rpt, err := scanReconciliationReportRow(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, *rpt)
	}

	return &domain.PaginatedResult[domain.ReconciliationReport]{
		Items:      items,
		TotalCount: totalCount,
		Page:       page,
		PageSize:   pageSize,
	}, nil
}

// SumWalletBalancesByRole sums wallet balances for users with the given role.
// role is "client" or "owner".
func (r *reconciliationRepo) SumWalletBalancesByRole(ctx context.Context, role string) (total int64, count int, err error) {
	err = r.pool.QueryRow(ctx, `
		SELECT COALESCE(SUM(w.balance), 0), COUNT(w.id)
		FROM wallets w
		JOIN users u ON u.id = w.user_id
		WHERE u.role = $1 AND w.status != 'archived'`,
		role,
	).Scan(&total, &count)
	return
}

// SumEscrowHeld sums all escrows in "held" or "disputed" status.
func (r *reconciliationRepo) SumEscrowHeld(ctx context.Context) (total int64, count int, err error) {
	err = r.pool.QueryRow(ctx, `
		SELECT COALESCE(SUM(amount), 0), COUNT(id)
		FROM escrows
		WHERE status IN ('held', 'disputed')`).Scan(&total, &count)
	return
}

// SumActiveWalletHolds sums all active wallet holds.
func (r *reconciliationRepo) SumActiveWalletHolds(ctx context.Context) (total int64, err error) {
	err = r.pool.QueryRow(ctx, `
		SELECT COALESCE(SUM(amount), 0)
		FROM wallet_holds
		WHERE status = 'active'`).Scan(&total)
	return
}

// SumPaymentsForPeriod sums succeeded payments in a time period.
func (r *reconciliationRepo) SumPaymentsForPeriod(ctx context.Context, from, to time.Time) (sum int64, count int, err error) {
	err = r.pool.QueryRow(ctx, `
		SELECT COALESCE(SUM(amount), 0), COUNT(id)
		FROM payments
		WHERE status IN ('succeeded', 'partially_refunded', 'refunded')
		AND created_at >= $1 AND created_at < $2`,
		from, to,
	).Scan(&sum, &count)
	return
}

// SumRefundsForPeriod sums refund amounts in a time period.
func (r *reconciliationRepo) SumRefundsForPeriod(ctx context.Context, from, to time.Time) (sum int64, count int, err error) {
	err = r.pool.QueryRow(ctx, `
		SELECT COALESCE(SUM(refund_amount), 0), COUNT(id)
		FROM payments
		WHERE refund_amount > 0
		AND refunded_at >= $1 AND refunded_at < $2`,
		from, to,
	).Scan(&sum, &count)
	return
}

// scanners

func scanFloatSnapshot(row pgx.Row) (*domain.FloatSnapshot, error) {
	var s domain.FloatSnapshot
	err := row.Scan(
		&s.ID, &s.ClientWalletsTotal, &s.OwnerWalletsTotal, &s.EscrowHeldTotal,
		&s.WalletHoldsTotal, &s.ExpectedTotal, &s.ActualTotal, &s.Discrepancy,
		&s.Status, &s.ClientWalletsCount, &s.OwnerWalletsCount, &s.EscrowCount,
		&s.Notes, &s.SnapshotDate, &s.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func scanFloatSnapshotRow(rows pgx.Rows) (*domain.FloatSnapshot, error) {
	var s domain.FloatSnapshot
	err := rows.Scan(
		&s.ID, &s.ClientWalletsTotal, &s.OwnerWalletsTotal, &s.EscrowHeldTotal,
		&s.WalletHoldsTotal, &s.ExpectedTotal, &s.ActualTotal, &s.Discrepancy,
		&s.Status, &s.ClientWalletsCount, &s.OwnerWalletsCount, &s.EscrowCount,
		&s.Notes, &s.SnapshotDate, &s.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func scanReconciliationReport(row pgx.Row) (*domain.ReconciliationReport, error) {
	var rpt domain.ReconciliationReport
	err := row.Scan(
		&rpt.ID, &rpt.PeriodStart, &rpt.PeriodEnd,
		&rpt.InternalPaymentsSum, &rpt.InternalPaymentsCount,
		&rpt.InternalRefundsSum, &rpt.InternalRefundsCount,
		&rpt.ProviderPaymentsSum, &rpt.ProviderPaymentsCount,
		&rpt.ProviderRefundsSum, &rpt.ProviderRefundsCount,
		&rpt.PaymentDiscrepancy, &rpt.RefundDiscrepancy,
		&rpt.Status, &rpt.MismatchDetails, &rpt.ErrorMessage, &rpt.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &rpt, nil
}

func scanReconciliationReportRow(rows pgx.Rows) (*domain.ReconciliationReport, error) {
	var rpt domain.ReconciliationReport
	err := rows.Scan(
		&rpt.ID, &rpt.PeriodStart, &rpt.PeriodEnd,
		&rpt.InternalPaymentsSum, &rpt.InternalPaymentsCount,
		&rpt.InternalRefundsSum, &rpt.InternalRefundsCount,
		&rpt.ProviderPaymentsSum, &rpt.ProviderPaymentsCount,
		&rpt.ProviderRefundsSum, &rpt.ProviderRefundsCount,
		&rpt.PaymentDiscrepancy, &rpt.RefundDiscrepancy,
		&rpt.Status, &rpt.MismatchDetails, &rpt.ErrorMessage, &rpt.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &rpt, nil
}
