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

type promoRepo struct {
	pool *pgxpool.Pool
}

func NewPromoCodeRepository(pool *pgxpool.Pool) repository.PromoCodeRepository {
	return &promoRepo{pool: pool}
}

func (r *promoRepo) Create(ctx context.Context, promo *domain.PromoCode) error {
	if promo.ID == uuid.Nil {
		promo.ID = uuid.New()
	}
	if promo.CreatedAt.IsZero() {
		promo.CreatedAt = time.Now()
	}

	query := `
		INSERT INTO promo_codes (id, code, type, value, bathhouse_id, creator_id, max_uses, current_uses, min_amount, valid_from, valid_until, is_active, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`

	_, err := r.pool.Exec(ctx, query,
		promo.ID, promo.Code, string(promo.Type), promo.Value,
		promo.BathhouseID, promo.CreatorID, promo.MaxUses, promo.CurrentUses,
		promo.MinAmount, promo.ValidFrom, promo.ValidUntil, promo.IsActive, promo.CreatedAt,
	)
	if err != nil {
		if isDuplicateKeyError(err) {
			return domain.ErrAlreadyExists
		}
		return fmt.Errorf("create promo code: %w", err)
	}
	return nil
}

func (r *promoRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.PromoCode, error) {
	query := `
		SELECT id, code, type, value, bathhouse_id, creator_id, max_uses, current_uses, min_amount, valid_from, valid_until, is_active, created_at
		FROM promo_codes WHERE id = $1`

	return r.scanPromo(ctx, query, id)
}

func (r *promoRepo) GetByCode(ctx context.Context, code string) (*domain.PromoCode, error) {
	query := `
		SELECT id, code, type, value, bathhouse_id, creator_id, max_uses, current_uses, min_amount, valid_from, valid_until, is_active, created_at
		FROM promo_codes WHERE code = $1`

	return r.scanPromo(ctx, query, code)
}

func (r *promoRepo) scanPromo(ctx context.Context, query string, arg any) (*domain.PromoCode, error) {
	var promo domain.PromoCode
	err := r.pool.QueryRow(ctx, query, arg).Scan(
		&promo.ID, &promo.Code, &promo.Type, &promo.Value,
		&promo.BathhouseID, &promo.CreatorID, &promo.MaxUses, &promo.CurrentUses,
		&promo.MinAmount, &promo.ValidFrom, &promo.ValidUntil, &promo.IsActive, &promo.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrPromoNotFound
		}
		return nil, fmt.Errorf("get promo code: %w", err)
	}
	return &promo, nil
}

func (r *promoRepo) Update(ctx context.Context, promo *domain.PromoCode) error {
	query := `
		UPDATE promo_codes
		SET code = $2, type = $3, value = $4, bathhouse_id = $5, max_uses = $6,
			min_amount = $7, valid_from = $8, valid_until = $9, is_active = $10
		WHERE id = $1`

	result, err := r.pool.Exec(ctx, query,
		promo.ID, promo.Code, string(promo.Type), promo.Value,
		promo.BathhouseID, promo.MaxUses, promo.MinAmount,
		promo.ValidFrom, promo.ValidUntil, promo.IsActive,
	)
	if err != nil {
		if isDuplicateKeyError(err) {
			return domain.ErrAlreadyExists
		}
		return fmt.Errorf("update promo code: %w", err)
	}
	if result.RowsAffected() == 0 {
		return domain.ErrPromoNotFound
	}
	return nil
}

func (r *promoRepo) ListByBathhouse(ctx context.Context, bathhouseID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.PromoCode], error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}

	var totalCount int64
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM promo_codes WHERE bathhouse_id = $1`, bathhouseID,
	).Scan(&totalCount)
	if err != nil {
		return nil, fmt.Errorf("count promo codes: %w", err)
	}

	offset := (page - 1) * pageSize
	query := `
		SELECT id, code, type, value, bathhouse_id, creator_id, max_uses, current_uses, min_amount, valid_from, valid_until, is_active, created_at
		FROM promo_codes
		WHERE bathhouse_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`

	return r.scanPromoList(ctx, query, totalCount, page, pageSize, bathhouseID, pageSize, offset)
}

func (r *promoRepo) scanPromoList(ctx context.Context, query string, totalCount int64, page, pageSize int, args ...any) (*domain.PaginatedResult[domain.PromoCode], error) {
	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list promo codes: %w", err)
	}
	defer rows.Close()

	promos := make([]domain.PromoCode, 0)
	for rows.Next() {
		var promo domain.PromoCode
		if err := rows.Scan(
			&promo.ID, &promo.Code, &promo.Type, &promo.Value,
			&promo.BathhouseID, &promo.CreatorID, &promo.MaxUses, &promo.CurrentUses,
			&promo.MinAmount, &promo.ValidFrom, &promo.ValidUntil, &promo.IsActive, &promo.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan promo code: %w", err)
		}
		promos = append(promos, promo)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate promo code rows: %w", err)
	}

	return &domain.PaginatedResult[domain.PromoCode]{
		Items:      promos,
		TotalCount: totalCount,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: int(math.Ceil(float64(totalCount) / float64(pageSize))),
	}, nil
}

func (r *promoRepo) ListByCreator(ctx context.Context, creatorID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.PromoCode], error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}

	var totalCount int64
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM promo_codes WHERE creator_id = $1`, creatorID,
	).Scan(&totalCount)
	if err != nil {
		return nil, fmt.Errorf("count promo codes by creator: %w", err)
	}

	offset := (page - 1) * pageSize
	query := `
		SELECT id, code, type, value, bathhouse_id, creator_id, max_uses, current_uses, min_amount, valid_from, valid_until, is_active, created_at
		FROM promo_codes
		WHERE creator_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`

	return r.scanPromoList(ctx, query, totalCount, page, pageSize, creatorID, pageSize, offset)
}

func (r *promoRepo) IncrementUses(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE promo_codes SET current_uses = current_uses + 1 WHERE id = $1 AND (max_uses = 0 OR current_uses < max_uses)`

	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("increment promo uses: %w", err)
	}
	if result.RowsAffected() == 0 {
		return domain.ErrPromoMaxUses
	}
	return nil
}

func (r *promoRepo) RecordUsage(ctx context.Context, usage *domain.PromoUsage) error {
	if usage.ID == uuid.Nil {
		usage.ID = uuid.New()
	}
	if usage.UsedAt.IsZero() {
		usage.UsedAt = time.Now()
	}

	query := `
		INSERT INTO promo_usages (id, promo_code_id, user_id, booking_id, discount_amount, used_at)
		VALUES ($1, $2, $3, $4, $5, $6)`

	_, err := r.pool.Exec(ctx, query,
		usage.ID, usage.PromoCodeID, usage.UserID, usage.BookingID,
		usage.DiscountAmount, usage.UsedAt,
	)
	if err != nil {
		return fmt.Errorf("record promo usage: %w", err)
	}
	return nil
}

func (r *promoRepo) ApplyUsage(ctx context.Context, id uuid.UUID, usage *domain.PromoUsage) error {
	if usage.ID == uuid.Nil {
		usage.ID = uuid.New()
	}
	if usage.UsedAt.IsZero() {
		usage.UsedAt = time.Now()
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	incrementQuery := `UPDATE promo_codes SET current_uses = current_uses + 1 WHERE id = $1 AND (max_uses = 0 OR current_uses < max_uses)`
	result, err := tx.Exec(ctx, incrementQuery, id)
	if err != nil {
		return fmt.Errorf("increment promo uses: %w", err)
	}
	if result.RowsAffected() == 0 {
		return domain.ErrPromoMaxUses
	}

	usageQuery := `
		INSERT INTO promo_usages (id, promo_code_id, user_id, booking_id, discount_amount, used_at)
		VALUES ($1, $2, $3, $4, $5, $6)`
	_, err = tx.Exec(ctx, usageQuery,
		usage.ID, usage.PromoCodeID, usage.UserID, usage.BookingID,
		usage.DiscountAmount, usage.UsedAt,
	)
	if err != nil {
		return fmt.Errorf("record promo usage: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}
	return nil
}

func (r *promoRepo) GetUsageByBookingID(ctx context.Context, bookingID uuid.UUID) (*domain.PromoUsage, error) {
	query := `
		SELECT id, promo_code_id, user_id, booking_id, discount_amount, used_at
		FROM promo_usages WHERE booking_id = $1`

	var usage domain.PromoUsage
	err := r.pool.QueryRow(ctx, query, bookingID).Scan(
		&usage.ID, &usage.PromoCodeID, &usage.UserID, &usage.BookingID,
		&usage.DiscountAmount, &usage.UsedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("get promo usage by booking: %w", err)
	}
	return &usage, nil
}

func (r *promoRepo) DecrementUses(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE promo_codes SET current_uses = current_uses - 1 WHERE id = $1 AND current_uses > 0`

	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("decrement promo uses: %w", err)
	}
	if result.RowsAffected() == 0 {
		return domain.ErrPromoNotFound
	}
	return nil
}

func (r *promoRepo) DeleteUsage(ctx context.Context, usageID uuid.UUID) error {
	query := `DELETE FROM promo_usages WHERE id = $1`

	result, err := r.pool.Exec(ctx, query, usageID)
	if err != nil {
		return fmt.Errorf("delete promo usage: %w", err)
	}
	if result.RowsAffected() == 0 {
		return domain.ErrPromoNotFound
	}
	return nil
}

func (r *promoRepo) RefundUsage(ctx context.Context, promoCodeID uuid.UUID, usageID uuid.UUID) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	decrementQuery := `UPDATE promo_codes SET current_uses = current_uses - 1 WHERE id = $1 AND current_uses > 0`
	result, err := tx.Exec(ctx, decrementQuery, promoCodeID)
	if err != nil {
		return fmt.Errorf("decrement promo uses: %w", err)
	}
	if result.RowsAffected() == 0 {
		return domain.ErrPromoNotFound
	}

	deleteQuery := `DELETE FROM promo_usages WHERE id = $1`
	result, err = tx.Exec(ctx, deleteQuery, usageID)
	if err != nil {
		return fmt.Errorf("delete promo usage: %w", err)
	}
	if result.RowsAffected() == 0 {
		return domain.ErrPromoNotFound
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}
	return nil
}

func (r *promoRepo) DeactivateExpired(ctx context.Context, before time.Time) (int64, error) {
	query := `UPDATE promo_codes SET is_active = false WHERE is_active = true AND valid_until < $1`
	result, err := r.pool.Exec(ctx, query, before)
	if err != nil {
		return 0, fmt.Errorf("deactivate expired promo codes: %w", err)
	}
	return result.RowsAffected(), nil
}
