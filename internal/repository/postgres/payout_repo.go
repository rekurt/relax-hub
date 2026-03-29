package postgres

import (
	"context"
	"errors"
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/repository"
)

type payoutRepo struct {
	pool *pgxpool.Pool
}

func NewPayoutRepository(pool *pgxpool.Pool) repository.PayoutRepository {
	return &payoutRepo{pool: pool}
}

func (r *payoutRepo) Create(ctx context.Context, payout *domain.Payout) error {
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

	query := `
		INSERT INTO payouts (id, user_id, amount, status, payout_method, bank_details, external_id, requested_at, processed_at, failure_reason, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`

	_, err := r.pool.Exec(ctx, query,
		payout.ID, payout.UserID, payout.Amount, payout.Status, payout.PayoutMethod,
		payout.BankDetails, payout.ExternalID, payout.RequestedAt, payout.ProcessedAt,
		payout.FailureReason, payout.CreatedAt,
	)
	return err
}

func (r *payoutRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Payout, error) {
	query := `
		SELECT id, user_id, amount, status, payout_method, bank_details, external_id, requested_at, processed_at, failure_reason, created_at
		FROM payouts WHERE id = $1`

	var p domain.Payout
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&p.ID, &p.UserID, &p.Amount, &p.Status, &p.PayoutMethod,
		&p.BankDetails, &p.ExternalID, &p.RequestedAt, &p.ProcessedAt,
		&p.FailureReason, &p.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrPayoutNotFound
		}
		return nil, err
	}
	return &p, nil
}

func (r *payoutRepo) ListByUser(ctx context.Context, userID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.Payout], error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}

	countQuery := `SELECT COUNT(*) FROM payouts WHERE user_id = $1`
	var total int64
	if err := r.pool.QueryRow(ctx, countQuery, userID).Scan(&total); err != nil {
		return nil, err
	}

	offset := (page - 1) * pageSize
	query := `
		SELECT id, user_id, amount, status, payout_method, bank_details, external_id, requested_at, processed_at, failure_reason, created_at
		FROM payouts WHERE user_id = $1
		ORDER BY requested_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := r.pool.Query(ctx, query, userID, pageSize, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []domain.Payout
	for rows.Next() {
		var p domain.Payout
		if err := rows.Scan(
			&p.ID, &p.UserID, &p.Amount, &p.Status, &p.PayoutMethod,
			&p.BankDetails, &p.ExternalID, &p.RequestedAt, &p.ProcessedAt,
			&p.FailureReason, &p.CreatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, p)
	}

	totalPages := int(math.Ceil(float64(total) / float64(pageSize)))

	return &domain.PaginatedResult[domain.Payout]{
		Items:      items,
		TotalCount: total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

func (r *payoutRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.PayoutStatus, processedAt *time.Time, failureReason string) error {
	query := `
		UPDATE payouts SET status = $2, processed_at = $3, failure_reason = $4
		WHERE id = $1`

	tag, err := r.pool.Exec(ctx, query, id, status, processedAt, failureReason)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrPayoutNotFound
	}
	return nil
}

func (r *payoutRepo) UpdateExternalID(ctx context.Context, id uuid.UUID, externalID string) error {
	query := `UPDATE payouts SET external_id = $2 WHERE id = $1`
	tag, err := r.pool.Exec(ctx, query, id, externalID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrPayoutNotFound
	}
	return nil
}

func (r *payoutRepo) GetDailyTotal(ctx context.Context, userID uuid.UUID, date time.Time) (int64, error) {
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	query := `
		SELECT COALESCE(SUM(amount), 0) FROM payouts
		WHERE user_id = $1 AND status IN ('pending', 'processing', 'completed')
		AND requested_at >= $2 AND requested_at < $3`

	var total int64
	err := r.pool.QueryRow(ctx, query, userID, startOfDay, endOfDay).Scan(&total)
	return total, err
}

func (r *payoutRepo) GetMonthlyTotal(ctx context.Context, userID uuid.UUID, year int, month time.Month) (int64, error) {
	startOfMonth := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)
	endOfMonth := startOfMonth.AddDate(0, 1, 0)

	query := `
		SELECT COALESCE(SUM(amount), 0) FROM payouts
		WHERE user_id = $1 AND status IN ('pending', 'processing', 'completed')
		AND requested_at >= $2 AND requested_at < $3`

	var total int64
	err := r.pool.QueryRow(ctx, query, userID, startOfMonth, endOfMonth).Scan(&total)
	return total, err
}

func (r *payoutRepo) GetPendingTotal(ctx context.Context, userID uuid.UUID) (int64, error) {
	query := `
		SELECT COALESCE(SUM(amount), 0) FROM payouts
		WHERE user_id = $1 AND status IN ('pending', 'processing')`

	var total int64
	err := r.pool.QueryRow(ctx, query, userID).Scan(&total)
	return total, err
}

func (r *payoutRepo) GetAutoPayoutSettings(ctx context.Context, userID uuid.UUID) (*domain.AutoPayoutSettings, error) {
	query := `SELECT user_id, threshold, updated_at FROM auto_payout_settings WHERE user_id = $1`

	var s domain.AutoPayoutSettings
	err := r.pool.QueryRow(ctx, query, userID).Scan(&s.UserID, &s.Threshold, &s.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return &domain.AutoPayoutSettings{UserID: userID, Threshold: 0}, nil
		}
		return nil, err
	}
	return &s, nil
}

func (r *payoutRepo) ListActiveAutoPayoutSettings(ctx context.Context) ([]domain.AutoPayoutSettings, error) {
	query := `SELECT user_id, threshold, updated_at FROM auto_payout_settings WHERE threshold > 0`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []domain.AutoPayoutSettings
	for rows.Next() {
		var s domain.AutoPayoutSettings
		if err := rows.Scan(&s.UserID, &s.Threshold, &s.UpdatedAt); err != nil {
			return nil, err
		}
		result = append(result, s)
	}
	return result, rows.Err()
}

func (r *payoutRepo) UpsertAutoPayoutSettings(ctx context.Context, settings *domain.AutoPayoutSettings) error {
	settings.UpdatedAt = time.Now()
	query := `
		INSERT INTO auto_payout_settings (user_id, threshold, updated_at)
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id) DO UPDATE SET threshold = $2, updated_at = $3`

	_, err := r.pool.Exec(ctx, query, settings.UserID, settings.Threshold, settings.UpdatedAt)
	return err
}
