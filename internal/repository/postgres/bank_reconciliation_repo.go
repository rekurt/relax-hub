package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nikitaaldaev/bani/internal/domain"
)

type bankReconciliationRepo struct {
	pool *pgxpool.Pool
}

func NewBankReconciliationRepo(pool *pgxpool.Pool) *bankReconciliationRepo {
	return &bankReconciliationRepo{pool: pool}
}

func (r *bankReconciliationRepo) CreateUpload(ctx context.Context, upload *domain.BankStatementUpload) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO bank_statement_uploads (id, file_name, format, total_rows, matched_count, pending_count, ignored_count, uploaded_by, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		upload.ID, upload.FileName, upload.Format, upload.TotalRows,
		upload.MatchedCount, upload.PendingCount, upload.IgnoredCount,
		upload.UploadedBy, upload.CreatedAt,
	)
	return err
}

func (r *bankReconciliationRepo) UpdateUploadCounts(ctx context.Context, id uuid.UUID, matched, pending, ignored int) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE bank_statement_uploads
		SET matched_count = $2, pending_count = $3, ignored_count = $4
		WHERE id = $1`,
		id, matched, pending, ignored,
	)
	return err
}

func (r *bankReconciliationRepo) CreateEntries(ctx context.Context, entries []domain.BankStatementEntry) error {
	if len(entries) == 0 {
		return nil
	}

	batch := &pgx.Batch{}
	for _, e := range entries {
		batch.Queue(`
			INSERT INTO bank_statement_entries (id, date, amount, description, counterparty, reference_num, matched_tx_id, matched_tx_type, status, upload_batch_id, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`,
			e.ID, e.Date, e.Amount, e.Description, e.Counterparty, e.ReferenceNum,
			e.MatchedTxID, e.MatchedTxType, string(e.Status), e.UploadBatchID,
			e.CreatedAt, e.UpdatedAt,
		)
	}

	results := r.pool.SendBatch(ctx, batch)
	defer results.Close()

	for range entries {
		if _, err := results.Exec(); err != nil {
			return fmt.Errorf("insert bank entry: %w", err)
		}
	}
	return nil
}

func (r *bankReconciliationRepo) GetEntryByID(ctx context.Context, id uuid.UUID) (*domain.BankStatementEntry, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, date, amount, description, counterparty, reference_num,
			   matched_tx_id, matched_tx_type, status, upload_batch_id, created_at, updated_at
		FROM bank_statement_entries WHERE id = $1`, id)

	return scanBankStatementEntry(row)
}

func (r *bankReconciliationRepo) ListEntries(ctx context.Context, filter domain.BankStatementFilter) (*domain.PaginatedResult[domain.BankStatementEntry], error) {
	where := "WHERE 1=1"
	args := []interface{}{}
	argIdx := 1

	if filter.Status != nil {
		where += fmt.Sprintf(" AND status = $%d", argIdx)
		args = append(args, string(*filter.Status))
		argIdx++
	}
	if filter.UploadBatchID != nil {
		where += fmt.Sprintf(" AND upload_batch_id = $%d", argIdx)
		args = append(args, *filter.UploadBatchID)
		argIdx++
	}
	if filter.DateFrom != nil {
		where += fmt.Sprintf(" AND date >= $%d", argIdx)
		args = append(args, *filter.DateFrom)
		argIdx++
	}
	if filter.DateTo != nil {
		where += fmt.Sprintf(" AND date <= $%d", argIdx)
		args = append(args, *filter.DateTo)
		argIdx++
	}

	var totalCount int64
	countQuery := "SELECT COUNT(*) FROM bank_statement_entries " + where
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&totalCount); err != nil {
		return nil, fmt.Errorf("count bank entries: %w", err)
	}

	page := filter.Page
	if page < 1 {
		page = 1
	}
	pageSize := filter.PageSize
	if pageSize < 1 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	query := fmt.Sprintf(`
		SELECT id, date, amount, description, counterparty, reference_num,
			   matched_tx_id, matched_tx_type, status, upload_batch_id, created_at, updated_at
		FROM bank_statement_entries %s
		ORDER BY date DESC, created_at DESC
		LIMIT $%d OFFSET $%d`,
		where, argIdx, argIdx+1)
	args = append(args, pageSize, offset)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list bank entries: %w", err)
	}
	defer rows.Close()

	var items []domain.BankStatementEntry
	for rows.Next() {
		var e domain.BankStatementEntry
		var status string
		if err := rows.Scan(
			&e.ID, &e.Date, &e.Amount, &e.Description, &e.Counterparty, &e.ReferenceNum,
			&e.MatchedTxID, &e.MatchedTxType, &status, &e.UploadBatchID, &e.CreatedAt, &e.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan bank entry: %w", err)
		}
		e.Status = domain.BankStatementEntryStatus(status)
		items = append(items, e)
	}

	return &domain.PaginatedResult[domain.BankStatementEntry]{
		Items:      items,
		TotalCount: totalCount,
		Page:       page,
		PageSize:   pageSize,
	}, nil
}

func (r *bankReconciliationRepo) MatchEntry(ctx context.Context, entryID uuid.UUID, txID uuid.UUID, txType string) error {
	tag, err := r.pool.Exec(ctx, `
		UPDATE bank_statement_entries
		SET matched_tx_id = $2, matched_tx_type = $3, status = $4, updated_at = now()
		WHERE id = $1 AND status = 'pending'`,
		entryID, txID, txType, string(domain.BankEntryManual),
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrBankEntryNotFound
	}
	return nil
}

func (r *bankReconciliationRepo) FindPaymentsByAmountAndDate(ctx context.Context, amount int64, dateFrom, dateTo time.Time) ([]domain.Payment, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, booking_id, user_id, amount, currency, status, provider, external_id,
			   payment_method, wallet_amount, card_amount, is_hold, captured_at,
			   refund_amount, refunded_at, metadata, created_at, updated_at
		FROM payments
		WHERE amount = $1 AND status = 'succeeded' AND created_at >= $2 AND created_at < $3
		ORDER BY created_at DESC`,
		amount, dateFrom, dateTo,
	)
	if err != nil {
		return nil, fmt.Errorf("find payments by amount: %w", err)
	}
	defer rows.Close()

	var payments []domain.Payment
	for rows.Next() {
		var p domain.Payment
		var metadata map[string]string
		if err := rows.Scan(
			&p.ID, &p.BookingID, &p.UserID, &p.Amount, &p.Currency, &p.Status,
			&p.Provider, &p.ExternalID, &p.PaymentMethod, &p.WalletAmount, &p.CardAmount,
			&p.IsHold, &p.CapturedAt, &p.RefundAmount, &p.RefundedAt,
			&metadata, &p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan payment: %w", err)
		}
		p.Metadata = metadata
		payments = append(payments, p)
	}
	return payments, nil
}

type bankEntryScanner interface {
	Scan(dest ...interface{}) error
}

func scanBankStatementEntry(row bankEntryScanner) (*domain.BankStatementEntry, error) {
	var e domain.BankStatementEntry
	var status string
	err := row.Scan(
		&e.ID, &e.Date, &e.Amount, &e.Description, &e.Counterparty, &e.ReferenceNum,
		&e.MatchedTxID, &e.MatchedTxType, &status, &e.UploadBatchID, &e.CreatedAt, &e.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	e.Status = domain.BankStatementEntryStatus(status)
	return &e, nil
}
