package postgres

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/repository"
)

type referralRepo struct {
	pool *pgxpool.Pool
}

func NewReferralRepository(pool *pgxpool.Pool) repository.ReferralRepository {
	return &referralRepo{pool: pool}
}

func (r *referralRepo) Create(ctx context.Context, referral *domain.Referral) error {
	if referral.ID == uuid.Nil {
		referral.ID = uuid.New()
	}
	if referral.Status == "" {
		referral.Status = domain.ReferralStatusPending
	}
	if referral.CreatedAt.IsZero() {
		referral.CreatedAt = time.Now()
	}

	query := `
		INSERT INTO referrals (id, referrer_id, referee_id, referral_code, status, bonus_amount, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`

	_, err := r.pool.Exec(ctx, query,
		referral.ID, referral.ReferrerID, referral.RefereeID,
		referral.ReferralCode, string(referral.Status), referral.BonusAmount, referral.CreatedAt,
	)
	if err != nil {
		if isDuplicateKeyError(err) {
			return domain.ErrAlreadyReferred
		}
		return fmt.Errorf("create referral: %w", err)
	}
	return nil
}

func (r *referralRepo) GetByReferee(ctx context.Context, refereeID uuid.UUID) (*domain.Referral, error) {
	query := `
		SELECT id, referrer_id, referee_id, referral_code, status, bonus_amount, completed_at, created_at
		FROM referrals WHERE referee_id = $1`

	var ref domain.Referral
	err := r.pool.QueryRow(ctx, query, refereeID).Scan(
		&ref.ID, &ref.ReferrerID, &ref.RefereeID, &ref.ReferralCode,
		&ref.Status, &ref.BonusAmount, &ref.CompletedAt, &ref.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("get referral by referee: %w", err)
	}
	return &ref, nil
}

func (r *referralRepo) ListByReferrer(ctx context.Context, referrerID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.Referral], error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}

	var totalCount int64
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM referrals WHERE referrer_id = $1`, referrerID).Scan(&totalCount)
	if err != nil {
		return nil, fmt.Errorf("count referrals: %w", err)
	}

	offset := (page - 1) * pageSize
	query := `
		SELECT id, referrer_id, referee_id, referral_code, status, bonus_amount, completed_at, created_at
		FROM referrals WHERE referrer_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := r.pool.Query(ctx, query, referrerID, pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("list referrals: %w", err)
	}
	defer rows.Close()

	referrals := make([]domain.Referral, 0)
	for rows.Next() {
		var ref domain.Referral
		if err := rows.Scan(
			&ref.ID, &ref.ReferrerID, &ref.RefereeID, &ref.ReferralCode,
			&ref.Status, &ref.BonusAmount, &ref.CompletedAt, &ref.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan referral: %w", err)
		}
		referrals = append(referrals, ref)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate referral rows: %w", err)
	}

	return &domain.PaginatedResult[domain.Referral]{
		Items:      referrals,
		TotalCount: totalCount,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: int(math.Ceil(float64(totalCount) / float64(pageSize))),
	}, nil
}

func (r *referralRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.ReferralStatus, completedAt *time.Time) error {
	query := `UPDATE referrals SET status = $2, completed_at = $3 WHERE id = $1 AND status = 'pending'`
	result, err := r.pool.Exec(ctx, query, id, string(status), completedAt)
	if err != nil {
		return fmt.Errorf("update referral status: %w", err)
	}
	if result.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *referralRepo) RevertToPending(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE referrals SET status = 'pending', completed_at = NULL WHERE id = $1 AND status = 'completed'`
	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("revert referral to pending: %w", err)
	}
	if result.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *referralRepo) GetBalance(ctx context.Context, userID uuid.UUID) (*domain.ReferralBalance, error) {
	query := `
		SELECT user_id, balance, total_earned, updated_at, created_at
		FROM referral_balances WHERE user_id = $1`

	var b domain.ReferralBalance
	err := r.pool.QueryRow(ctx, query, userID).Scan(
		&b.UserID, &b.Balance, &b.TotalEarned, &b.UpdatedAt, &b.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("get referral balance: %w", err)
	}
	return &b, nil
}

func (r *referralRepo) CreateBalance(ctx context.Context, balance *domain.ReferralBalance) error {
	now := time.Now()
	balance.CreatedAt = now
	balance.UpdatedAt = now

	query := `
		INSERT INTO referral_balances (user_id, balance, total_earned, updated_at, created_at)
		VALUES ($1, $2, $3, $4, $5)`

	_, err := r.pool.Exec(ctx, query,
		balance.UserID, balance.Balance, balance.TotalEarned, balance.UpdatedAt, balance.CreatedAt,
	)
	if err != nil {
		if isDuplicateKeyError(err) {
			return domain.ErrAlreadyExists
		}
		return fmt.Errorf("create referral balance: %w", err)
	}
	return nil
}

func (r *referralRepo) UpdateBalance(ctx context.Context, userID uuid.UUID, delta int64, trackEarnings bool) error {
	if delta < 0 {
		// For spending, check sufficient balance
		query := `
			UPDATE referral_balances
			SET balance = balance + $1, updated_at = $2
			WHERE user_id = $3 AND balance >= $4`

		result, err := r.pool.Exec(ctx, query, delta, time.Now(), userID, -delta)
		if err != nil {
			return fmt.Errorf("update referral balance: %w", err)
		}
		if result.RowsAffected() == 0 {
			return domain.ErrInsufficientReferralBalance
		}
		return nil
	}

	var query string
	if trackEarnings {
		query = `
			UPDATE referral_balances
			SET balance = balance + $1, total_earned = total_earned + $1, updated_at = $2
			WHERE user_id = $3`
	} else {
		query = `
			UPDATE referral_balances
			SET balance = balance + $1, updated_at = $2
			WHERE user_id = $3`
	}

	result, err := r.pool.Exec(ctx, query, delta, time.Now(), userID)
	if err != nil {
		return fmt.Errorf("update referral balance: %w", err)
	}
	if result.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *referralRepo) CountByReferrer(ctx context.Context, referrerID uuid.UUID) (int, int, error) {
	query := `
		SELECT
			COUNT(*) AS total_invited,
			COUNT(*) FILTER (WHERE status = 'completed') AS total_completed
		FROM referrals WHERE referrer_id = $1`

	var totalInvited, totalCompleted int
	err := r.pool.QueryRow(ctx, query, referrerID).Scan(&totalInvited, &totalCompleted)
	if err != nil {
		return 0, 0, fmt.Errorf("count referrals by referrer: %w", err)
	}
	return totalInvited, totalCompleted, nil
}
