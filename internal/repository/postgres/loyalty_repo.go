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

type loyaltyRepo struct {
	pool *pgxpool.Pool
}

func NewLoyaltyRepository(pool *pgxpool.Pool) repository.LoyaltyRepository {
	return &loyaltyRepo{pool: pool}
}

func (r *loyaltyRepo) GetAccount(ctx context.Context, userID uuid.UUID) (*domain.LoyaltyAccount, error) {
	account := &domain.LoyaltyAccount{}
	query := `
		SELECT user_id, level, points, total_earned, total_spent, visit_count, created_at, updated_at
		FROM loyalty_accounts
		WHERE user_id = $1
	`

	err := r.pool.QueryRow(ctx, query, userID).Scan(
		&account.UserID, &account.Level, &account.Points, &account.TotalEarned,
		&account.TotalSpent, &account.VisitCount, &account.CreatedAt, &account.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("get loyalty account: %w", err)
	}
	return account, nil
}

func (r *loyaltyRepo) CreateAccount(ctx context.Context, account *domain.LoyaltyAccount) error {
	now := time.Now()
	account.CreatedAt = now
	account.UpdatedAt = now

	query := `
		INSERT INTO loyalty_accounts (user_id, level, points, total_earned, total_spent, visit_count, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := r.pool.Exec(ctx, query,
		account.UserID, account.Level, account.Points, account.TotalEarned,
		account.TotalSpent, account.VisitCount, account.CreatedAt, account.UpdatedAt,
	)
	if err != nil {
		if isDuplicateKeyError(err) {
			return domain.ErrAlreadyExists
		}
		return fmt.Errorf("create loyalty account: %w", err)
	}
	return nil
}

func (r *loyaltyRepo) AddPoints(ctx context.Context, userID uuid.UUID, amount int64) error {
	query := `
		UPDATE loyalty_accounts
		SET points = points + $1, total_earned = total_earned + $1, updated_at = $2
		WHERE user_id = $3
	`

	result, err := r.pool.Exec(ctx, query, amount, time.Now(), userID)
	if err != nil {
		return fmt.Errorf("add loyalty points: %w", err)
	}
	if result.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *loyaltyRepo) SpendPoints(ctx context.Context, userID uuid.UUID, amount int64) error {
	query := `
		UPDATE loyalty_accounts
		SET points = points - $1, total_spent = total_spent + $1, updated_at = $2
		WHERE user_id = $3 AND points >= $1
	`

	result, err := r.pool.Exec(ctx, query, amount, time.Now(), userID)
	if err != nil {
		return fmt.Errorf("spend loyalty points: %w", err)
	}
	if result.RowsAffected() == 0 {
		return domain.ErrInsufficientPoints
	}
	return nil
}

func (r *loyaltyRepo) IncrementVisitCount(ctx context.Context, userID uuid.UUID) error {
	query := `
		UPDATE loyalty_accounts
		SET visit_count = visit_count + 1, updated_at = $1
		WHERE user_id = $2
	`

	result, err := r.pool.Exec(ctx, query, time.Now(), userID)
	if err != nil {
		return fmt.Errorf("increment visit count: %w", err)
	}
	if result.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *loyaltyRepo) UpdateLevel(ctx context.Context, userID uuid.UUID, level domain.LoyaltyLevel) error {
	query := `
		UPDATE loyalty_accounts
		SET level = $1, updated_at = $2
		WHERE user_id = $3
	`

	result, err := r.pool.Exec(ctx, query, level, time.Now(), userID)
	if err != nil {
		return fmt.Errorf("update loyalty level: %w", err)
	}
	if result.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *loyaltyRepo) ListTransactions(ctx context.Context, userID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.LoyaltyTransaction], error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}

	var totalCount int64
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM loyalty_transactions WHERE user_id = $1`, userID).Scan(&totalCount)
	if err != nil {
		return nil, fmt.Errorf("count loyalty transactions: %w", err)
	}

	offset := (page - 1) * pageSize
	query := `
		SELECT id, user_id, type, amount, booking_id, description, created_at
		FROM loyalty_transactions
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.pool.Query(ctx, query, userID, pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("list loyalty transactions: %w", err)
	}
	defer rows.Close()

	var txns []domain.LoyaltyTransaction
	for rows.Next() {
		var tx domain.LoyaltyTransaction
		if err := rows.Scan(&tx.ID, &tx.UserID, &tx.Type, &tx.Amount, &tx.BookingID, &tx.Description, &tx.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan loyalty transaction: %w", err)
		}
		txns = append(txns, tx)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate loyalty transaction rows: %w", err)
	}

	return &domain.PaginatedResult[domain.LoyaltyTransaction]{
		Items:      txns,
		TotalCount: totalCount,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: int(math.Ceil(float64(totalCount) / float64(pageSize))),
	}, nil
}

func (r *loyaltyRepo) CreateTransaction(ctx context.Context, tx *domain.LoyaltyTransaction) error {
	if tx.ID == uuid.Nil {
		tx.ID = uuid.New()
	}
	tx.CreatedAt = time.Now()

	query := `
		INSERT INTO loyalty_transactions (id, user_id, type, amount, booking_id, description, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err := r.pool.Exec(ctx, query,
		tx.ID, tx.UserID, tx.Type, tx.Amount, tx.BookingID, tx.Description, tx.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("create loyalty transaction: %w", err)
	}
	return nil
}
