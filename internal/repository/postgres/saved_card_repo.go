package postgres

import (
	"context"
	"fmt"
	"math"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rekurt/relax-hub/internal/domain"
	"github.com/rekurt/relax-hub/internal/repository"
)

type savedCardRepo struct {
	pool *pgxpool.Pool
}

func NewSavedCardRepository(pool *pgxpool.Pool) repository.SavedCardRepository {
	return &savedCardRepo{pool: pool}
}

var savedCardColumns = `id, user_id, provider_token, last4, brand, expiry_month, expiry_year, is_default, created_at, updated_at`

func scanSavedCard(row pgx.Row) (*domain.SavedCard, error) {
	var c domain.SavedCard
	err := row.Scan(
		&c.ID, &c.UserID, &c.ProviderToken, &c.Last4, &c.Brand,
		&c.ExpiryMonth, &c.ExpiryYear, &c.IsDefault,
		&c.CreatedAt, &c.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *savedCardRepo) Create(ctx context.Context, card *domain.SavedCard) error {
	if card.ID == uuid.Nil {
		card.ID = uuid.New()
	}

	query := `
		INSERT INTO saved_cards (id, user_id, provider_token, last4, brand, expiry_month, expiry_year, is_default, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`

	_, err := r.pool.Exec(ctx, query,
		card.ID, card.UserID, card.ProviderToken, card.Last4, card.Brand,
		card.ExpiryMonth, card.ExpiryYear, card.IsDefault,
		card.CreatedAt, card.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("create saved card: %w", err)
	}
	return nil
}

func (r *savedCardRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.SavedCard, error) {
	query := fmt.Sprintf(`SELECT %s FROM saved_cards WHERE id=$1`, savedCardColumns)
	c, err := scanSavedCard(r.pool.QueryRow(ctx, query, id))
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrSavedCardNotFound
		}
		return nil, fmt.Errorf("get saved card by id: %w", err)
	}
	return c, nil
}

func (r *savedCardRepo) ListByUser(ctx context.Context, userID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.SavedCard], error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}

	var totalCount int64
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM saved_cards WHERE user_id = $1`, userID).Scan(&totalCount)
	if err != nil {
		return nil, fmt.Errorf("count saved cards: %w", err)
	}

	offset := (page - 1) * pageSize
	query := fmt.Sprintf(`SELECT %s FROM saved_cards WHERE user_id = $1 ORDER BY is_default DESC, created_at DESC LIMIT $2 OFFSET $3`, savedCardColumns)

	rows, err := r.pool.Query(ctx, query, userID, pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("list saved cards: %w", err)
	}
	defer rows.Close()

	var items []domain.SavedCard
	for rows.Next() {
		var c domain.SavedCard
		err := rows.Scan(
			&c.ID, &c.UserID, &c.ProviderToken, &c.Last4, &c.Brand,
			&c.ExpiryMonth, &c.ExpiryYear, &c.IsDefault,
			&c.CreatedAt, &c.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan saved card: %w", err)
		}
		items = append(items, c)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate saved card rows: %w", err)
	}

	return &domain.PaginatedResult[domain.SavedCard]{
		Items:      items,
		TotalCount: totalCount,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: int(math.Ceil(float64(totalCount) / float64(pageSize))),
	}, nil
}

func (r *savedCardRepo) Delete(ctx context.Context, id uuid.UUID) error {
	ct, err := r.pool.Exec(ctx, `DELETE FROM saved_cards WHERE id=$1`, id)
	if err != nil {
		return fmt.Errorf("delete saved card: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return domain.ErrSavedCardNotFound
	}
	return nil
}

func (r *savedCardRepo) SetDefault(ctx context.Context, userID uuid.UUID, cardID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `UPDATE saved_cards SET is_default = false, updated_at = now() WHERE user_id = $1 AND is_default = true`, userID)
	if err != nil {
		return fmt.Errorf("unset previous default: %w", err)
	}

	ct, err := r.pool.Exec(ctx, `UPDATE saved_cards SET is_default = true, updated_at = now() WHERE id = $1 AND user_id = $2`, cardID, userID)
	if err != nil {
		return fmt.Errorf("set default card: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return domain.ErrSavedCardNotFound
	}
	return nil
}

func (r *savedCardRepo) CountByUser(ctx context.Context, userID uuid.UUID) (int64, error) {
	var count int64
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM saved_cards WHERE user_id = $1`, userID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count saved cards: %w", err)
	}
	return count, nil
}
