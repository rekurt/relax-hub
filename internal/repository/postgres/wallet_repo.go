package postgres

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/repository"
)

type walletRepo struct {
	pool *pgxpool.Pool
}

func NewWalletRepository(pool *pgxpool.Pool) repository.WalletRepository {
	return &walletRepo{pool: pool}
}

func (r *walletRepo) Create(ctx context.Context, wallet *domain.Wallet) error {
	if wallet.ID == uuid.Nil {
		wallet.ID = uuid.New()
	}
	now := time.Now()
	if wallet.CreatedAt.IsZero() {
		wallet.CreatedAt = now
	}
	if wallet.UpdatedAt.IsZero() {
		wallet.UpdatedAt = now
	}
	if wallet.Status == "" {
		wallet.Status = domain.WalletStatusActive
	}
	if wallet.Currency == "" {
		wallet.Currency = domain.WalletCurrencyRUB
	}

	query := `
		INSERT INTO wallets (id, user_id, balance, held_amount, currency, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	_, err := r.pool.Exec(ctx, query,
		wallet.ID, wallet.UserID, wallet.Balance, wallet.HeldAmount,
		wallet.Currency, wallet.Status, wallet.CreatedAt, wallet.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("create wallet: %w", err)
	}
	return nil
}

func (r *walletRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Wallet, error) {
	query := `
		SELECT id, user_id, balance, held_amount, currency, status, created_at, updated_at
		FROM wallets WHERE id = $1`

	return r.scanWallet(ctx, query, id)
}

func (r *walletRepo) GetByUserID(ctx context.Context, userID uuid.UUID) (*domain.Wallet, error) {
	query := `
		SELECT id, user_id, balance, held_amount, currency, status, created_at, updated_at
		FROM wallets WHERE user_id = $1`

	return r.scanWallet(ctx, query, userID)
}

func (r *walletRepo) scanWallet(ctx context.Context, query string, arg interface{}) (*domain.Wallet, error) {
	var w domain.Wallet
	err := r.pool.QueryRow(ctx, query, arg).Scan(
		&w.ID, &w.UserID, &w.Balance, &w.HeldAmount,
		&w.Currency, &w.Status, &w.CreatedAt, &w.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrWalletNotFound
		}
		return nil, fmt.Errorf("scan wallet: %w", err)
	}
	return &w, nil
}

func (r *walletRepo) UpdateBalance(ctx context.Context, walletID uuid.UUID, newBalance int64, newHeldAmount int64) error {
	query := `UPDATE wallets SET balance = $2, held_amount = $3, updated_at = $4 WHERE id = $1`

	tag, err := r.pool.Exec(ctx, query, walletID, newBalance, newHeldAmount, time.Now())
	if err != nil {
		return fmt.Errorf("update wallet balance: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrWalletNotFound
	}
	return nil
}

func (r *walletRepo) UpdateStatus(ctx context.Context, walletID uuid.UUID, status domain.WalletStatus) error {
	query := `UPDATE wallets SET status = $2, updated_at = $3 WHERE id = $1`

	tag, err := r.pool.Exec(ctx, query, walletID, status, time.Now())
	if err != nil {
		return fmt.Errorf("update wallet status: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrWalletNotFound
	}
	return nil
}

// --- Transactions ---

func (r *walletRepo) CreateTransaction(ctx context.Context, tx *domain.WalletTransaction) error {
	if tx.ID == uuid.Nil {
		tx.ID = uuid.New()
	}
	if tx.CreatedAt.IsZero() {
		tx.CreatedAt = time.Now()
	}
	if tx.Status == "" {
		tx.Status = domain.WalletTxStatusCompleted
	}

	query := `
		INSERT INTO wallet_transactions (id, wallet_id, type, amount, balance_after, status, description, reference_type, reference_id, is_bonus, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`

	_, err := r.pool.Exec(ctx, query,
		tx.ID, tx.WalletID, tx.Type, tx.Amount, tx.BalanceAfter,
		tx.Status, tx.Description, tx.ReferenceType, tx.ReferenceID,
		tx.IsBonus, tx.ExpiresAt, tx.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("create wallet transaction: %w", err)
	}
	return nil
}

func (r *walletRepo) ListTransactions(ctx context.Context, filter domain.WalletTransactionFilter) (*domain.PaginatedResult[domain.WalletTransaction], error) {
	page := filter.Page
	pageSize := filter.PageSize
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}

	var conditions []string
	var args []interface{}
	argIdx := 1

	if filter.WalletID != nil {
		conditions = append(conditions, fmt.Sprintf("wallet_id = $%d", argIdx))
		args = append(args, *filter.WalletID)
		argIdx++
	}
	if filter.Type != nil {
		conditions = append(conditions, fmt.Sprintf("type = $%d", argIdx))
		args = append(args, *filter.Type)
		argIdx++
	}
	if filter.IsBonus != nil {
		conditions = append(conditions, fmt.Sprintf("is_bonus = $%d", argIdx))
		args = append(args, *filter.IsBonus)
		argIdx++
	}
	if filter.DateFrom != nil {
		conditions = append(conditions, fmt.Sprintf("created_at >= $%d", argIdx))
		args = append(args, *filter.DateFrom)
		argIdx++
	}
	if filter.DateTo != nil {
		conditions = append(conditions, fmt.Sprintf("created_at <= $%d", argIdx))
		args = append(args, *filter.DateTo)
		argIdx++
	}

	where := ""
	if len(conditions) > 0 {
		where = "WHERE " + strings.Join(conditions, " AND ")
	}

	var totalCount int64
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM wallet_transactions %s", where)
	err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&totalCount)
	if err != nil {
		return nil, fmt.Errorf("count wallet transactions: %w", err)
	}

	offset := (page - 1) * pageSize
	selectQuery := fmt.Sprintf(`
		SELECT id, wallet_id, type, amount, balance_after, status, description, reference_type, reference_id, is_bonus, expires_at, created_at
		FROM wallet_transactions %s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d`, where, argIdx, argIdx+1)
	args = append(args, pageSize, offset)

	rows, err := r.pool.Query(ctx, selectQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("list wallet transactions: %w", err)
	}
	defer rows.Close()

	var txs []domain.WalletTransaction
	for rows.Next() {
		var tx domain.WalletTransaction
		if err := rows.Scan(
			&tx.ID, &tx.WalletID, &tx.Type, &tx.Amount, &tx.BalanceAfter,
			&tx.Status, &tx.Description, &tx.ReferenceType, &tx.ReferenceID,
			&tx.IsBonus, &tx.ExpiresAt, &tx.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan wallet transaction: %w", err)
		}
		txs = append(txs, tx)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate wallet transaction rows: %w", err)
	}

	return &domain.PaginatedResult[domain.WalletTransaction]{
		Items:      txs,
		TotalCount: totalCount,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: int(math.Ceil(float64(totalCount) / float64(pageSize))),
	}, nil
}

func (r *walletRepo) GetExpiringBonuses(ctx context.Context, walletID uuid.UUID, before time.Time) ([]domain.WalletTransaction, error) {
	query := `
		SELECT id, wallet_id, type, amount, balance_after, status, description, reference_type, reference_id, is_bonus, expires_at, created_at
		FROM wallet_transactions
		WHERE wallet_id = $1 AND is_bonus = true AND expires_at IS NOT NULL AND expires_at <= $2
			AND status = 'completed'
		ORDER BY expires_at ASC`

	return r.scanTransactions(ctx, query, walletID, before)
}

func (r *walletRepo) GetBonusTransactionsForSpending(ctx context.Context, walletID uuid.UUID) ([]domain.WalletTransaction, error) {
	query := `
		SELECT id, wallet_id, type, amount, balance_after, status, description, reference_type, reference_id, is_bonus, expires_at, created_at
		FROM wallet_transactions
		WHERE wallet_id = $1 AND is_bonus = true AND status = 'completed'
			AND (expires_at IS NULL OR expires_at > now())
		ORDER BY expires_at ASC NULLS LAST, created_at ASC`

	return r.scanTransactions(ctx, query, walletID)
}

func (r *walletRepo) ExpireBonuses(ctx context.Context, transactionIDs []uuid.UUID) error {
	if len(transactionIDs) == 0 {
		return nil
	}

	query := `UPDATE wallet_transactions SET status = 'cancelled' WHERE id = ANY($1) AND status = 'completed'`
	_, err := r.pool.Exec(ctx, query, transactionIDs)
	if err != nil {
		return fmt.Errorf("expire bonuses: %w", err)
	}
	return nil
}

func (r *walletRepo) GetExpiringBonusesSoon(ctx context.Context, walletID uuid.UUID, from, to time.Time) ([]domain.WalletTransaction, error) {
	query := `
		SELECT id, wallet_id, type, amount, balance_after, status, description, reference_type, reference_id, is_bonus, expires_at, created_at
		FROM wallet_transactions
		WHERE wallet_id = $1 AND is_bonus = true AND expires_at >= $2 AND expires_at <= $3
			AND status = 'completed'
		ORDER BY expires_at ASC`

	return r.scanTransactions(ctx, query, walletID, from, to)
}

func (r *walletRepo) scanTransactions(ctx context.Context, query string, args ...interface{}) ([]domain.WalletTransaction, error) {
	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query wallet transactions: %w", err)
	}
	defer rows.Close()

	var txs []domain.WalletTransaction
	for rows.Next() {
		var tx domain.WalletTransaction
		if err := rows.Scan(
			&tx.ID, &tx.WalletID, &tx.Type, &tx.Amount, &tx.BalanceAfter,
			&tx.Status, &tx.Description, &tx.ReferenceType, &tx.ReferenceID,
			&tx.IsBonus, &tx.ExpiresAt, &tx.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan wallet transaction: %w", err)
		}
		txs = append(txs, tx)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate wallet transaction rows: %w", err)
	}
	return txs, nil
}

// --- Holds ---

func (r *walletRepo) CreateHold(ctx context.Context, hold *domain.WalletHold) error {
	if hold.ID == uuid.Nil {
		hold.ID = uuid.New()
	}
	if hold.CreatedAt.IsZero() {
		hold.CreatedAt = time.Now()
	}
	if hold.Status == "" {
		hold.Status = domain.WalletHoldStatusActive
	}

	query := `
		INSERT INTO wallet_holds (id, wallet_id, amount, status, description, reference_type, reference_id, expires_at, captured_at, released_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`

	_, err := r.pool.Exec(ctx, query,
		hold.ID, hold.WalletID, hold.Amount, hold.Status,
		hold.Description, hold.ReferenceType, hold.ReferenceID,
		hold.ExpiresAt, hold.CapturedAt, hold.ReleasedAt, hold.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("create wallet hold: %w", err)
	}
	return nil
}

func (r *walletRepo) GetHoldByID(ctx context.Context, holdID uuid.UUID) (*domain.WalletHold, error) {
	query := `
		SELECT id, wallet_id, amount, status, description, reference_type, reference_id, expires_at, captured_at, released_at, created_at
		FROM wallet_holds WHERE id = $1`

	var h domain.WalletHold
	err := r.pool.QueryRow(ctx, query, holdID).Scan(
		&h.ID, &h.WalletID, &h.Amount, &h.Status,
		&h.Description, &h.ReferenceType, &h.ReferenceID,
		&h.ExpiresAt, &h.CapturedAt, &h.ReleasedAt, &h.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrHoldNotFound
		}
		return nil, fmt.Errorf("scan wallet hold: %w", err)
	}
	return &h, nil
}

func (r *walletRepo) UpdateHoldStatus(ctx context.Context, holdID uuid.UUID, status domain.WalletHoldStatus, capturedAt, releasedAt *time.Time) error {
	query := `UPDATE wallet_holds SET status = $2, captured_at = $3, released_at = $4 WHERE id = $1`

	tag, err := r.pool.Exec(ctx, query, holdID, status, capturedAt, releasedAt)
	if err != nil {
		return fmt.Errorf("update hold status: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrHoldNotFound
	}
	return nil
}

func (r *walletRepo) GetActiveHolds(ctx context.Context, walletID uuid.UUID) ([]domain.WalletHold, error) {
	query := `
		SELECT id, wallet_id, amount, status, description, reference_type, reference_id, expires_at, captured_at, released_at, created_at
		FROM wallet_holds
		WHERE wallet_id = $1 AND status = 'active'
		ORDER BY created_at DESC`

	return r.scanHolds(ctx, query, walletID)
}

func (r *walletRepo) GetExpiredHolds(ctx context.Context, before time.Time) ([]domain.WalletHold, error) {
	query := `
		SELECT id, wallet_id, amount, status, description, reference_type, reference_id, expires_at, captured_at, released_at, created_at
		FROM wallet_holds
		WHERE status = 'active' AND expires_at <= $1
		ORDER BY expires_at ASC`

	return r.scanHolds(ctx, query, before)
}

func (r *walletRepo) scanHolds(ctx context.Context, query string, args ...interface{}) ([]domain.WalletHold, error) {
	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query wallet holds: %w", err)
	}
	defer rows.Close()

	var holds []domain.WalletHold
	for rows.Next() {
		var h domain.WalletHold
		if err := rows.Scan(
			&h.ID, &h.WalletID, &h.Amount, &h.Status,
			&h.Description, &h.ReferenceType, &h.ReferenceID,
			&h.ExpiresAt, &h.CapturedAt, &h.ReleasedAt, &h.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan wallet hold: %w", err)
		}
		holds = append(holds, h)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate wallet hold rows: %w", err)
	}
	return holds, nil
}
