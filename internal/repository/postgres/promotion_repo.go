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
	"github.com/rekurt/relax-hub/internal/domain"
	"github.com/rekurt/relax-hub/internal/repository"
)

type promotionRepo struct {
	pool *pgxpool.Pool
}

func NewPromotionRepository(pool *pgxpool.Pool) repository.PromotionRepository {
	return &promotionRepo{pool: pool}
}

const promotionColumns = `id, bathhouse_id, daily_bid_kopecks, budget_kopecks, spent_kopecks, start_date, end_date, target_city_id, status, impression_count, click_count, created_at, updated_at`

func scanPromotion(row pgx.Row) (*domain.Promotion, error) {
	promo := &domain.Promotion{}
	err := row.Scan(
		&promo.ID, &promo.BathhouseID, &promo.DailyBidKopecks, &promo.BudgetKopecks, &promo.SpentKopecks,
		&promo.StartDate, &promo.EndDate, &promo.TargetCityID, &promo.Status,
		&promo.ImpressionCount, &promo.ClickCount, &promo.CreatedAt, &promo.UpdatedAt,
	)
	return promo, err
}

func scanPromotionRows(rows pgx.Rows) ([]domain.Promotion, error) {
	var promos []domain.Promotion
	for rows.Next() {
		var promo domain.Promotion
		if err := rows.Scan(
			&promo.ID, &promo.BathhouseID, &promo.DailyBidKopecks, &promo.BudgetKopecks, &promo.SpentKopecks,
			&promo.StartDate, &promo.EndDate, &promo.TargetCityID, &promo.Status,
			&promo.ImpressionCount, &promo.ClickCount, &promo.CreatedAt, &promo.UpdatedAt,
		); err != nil {
			return nil, err
		}
		promos = append(promos, promo)
	}
	return promos, rows.Err()
}

func (r *promotionRepo) Create(ctx context.Context, promo *domain.Promotion) error {
	if promo.ID == uuid.Nil {
		promo.ID = uuid.New()
	}

	now := time.Now()
	promo.CreatedAt = now
	promo.UpdatedAt = now

	query := `
		INSERT INTO promotions (` + promotionColumns + `)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`

	_, err := r.pool.Exec(ctx, query,
		promo.ID, promo.BathhouseID, promo.DailyBidKopecks, promo.BudgetKopecks, promo.SpentKopecks,
		promo.StartDate, promo.EndDate, promo.TargetCityID, promo.Status,
		promo.ImpressionCount, promo.ClickCount, promo.CreatedAt, promo.UpdatedAt,
	)
	if err != nil {
		if isDuplicateKeyError(err) {
			return domain.ErrAlreadyExists
		}
		return fmt.Errorf("create promotion: %w", err)
	}
	return nil
}

func (r *promotionRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Promotion, error) {
	query := `SELECT ` + promotionColumns + ` FROM promotions WHERE id = $1`
	promo, err := scanPromotion(r.pool.QueryRow(ctx, query, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("get promotion by id: %w", err)
	}
	return promo, nil
}

func (r *promotionRepo) GetActiveBybathhouse(ctx context.Context, bathhouseID uuid.UUID) (*domain.Promotion, error) {
	query := `SELECT ` + promotionColumns + ` FROM promotions WHERE bathhouse_id = $1 AND status = $2 ORDER BY created_at DESC LIMIT 1`
	promo, err := scanPromotion(r.pool.QueryRow(ctx, query, bathhouseID, domain.PromotionActive))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("get active promotion by bathhouse: %w", err)
	}
	return promo, nil
}

func (r *promotionRepo) Update(ctx context.Context, promo *domain.Promotion) error {
	query := `
		UPDATE promotions
		SET daily_bid_kopecks = $1, budget_kopecks = $2, spent_kopecks = $3, start_date = $4, end_date = $5,
		    target_city_id = $6, status = $7, impression_count = $8, click_count = $9, updated_at = now()
		WHERE id = $10
	`

	result, err := r.pool.Exec(ctx, query,
		promo.DailyBidKopecks, promo.BudgetKopecks, promo.SpentKopecks,
		promo.StartDate, promo.EndDate, promo.TargetCityID, promo.Status,
		promo.ImpressionCount, promo.ClickCount, promo.ID,
	)
	if err != nil {
		return fmt.Errorf("update promotion: %w", err)
	}
	if result.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *promotionRepo) ListByOwner(ctx context.Context, ownerID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.Promotion], error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}

	var totalCount int64
	err := r.pool.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM promotions p
		JOIN bathhouses b ON p.bathhouse_id = b.id
		WHERE b.owner_id = $1
	`, ownerID).Scan(&totalCount)
	if err != nil {
		return nil, fmt.Errorf("count promotions: %w", err)
	}

	offset := (page - 1) * pageSize
	query := fmt.Sprintf(`
		SELECT p.%s
		FROM promotions p
		JOIN bathhouses b ON p.bathhouse_id = b.id
		WHERE b.owner_id = $1
		ORDER BY p.created_at DESC
		LIMIT $2 OFFSET $3
	`, promotionColumns)

	rows, err := r.pool.Query(ctx, query, ownerID, pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("list promotions: %w", err)
	}
	defer rows.Close()

	promos, err := scanPromotionRows(rows)
	if err != nil {
		return nil, fmt.Errorf("scan promotions: %w", err)
	}

	return &domain.PaginatedResult[domain.Promotion]{
		Items:      promos,
		TotalCount: totalCount,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: int(math.Ceil(float64(totalCount) / float64(pageSize))),
	}, nil
}

func (r *promotionRepo) ListByBathhouse(ctx context.Context, bathhouseID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.Promotion], error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}

	var totalCount int64
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM promotions WHERE bathhouse_id = $1`, bathhouseID).Scan(&totalCount)
	if err != nil {
		return nil, fmt.Errorf("count promotions by bathhouse: %w", err)
	}

	offset := (page - 1) * pageSize
	query := `SELECT ` + promotionColumns + ` FROM promotions WHERE bathhouse_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`

	rows, err := r.pool.Query(ctx, query, bathhouseID, pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("list promotions by bathhouse: %w", err)
	}
	defer rows.Close()

	promos, err := scanPromotionRows(rows)
	if err != nil {
		return nil, fmt.Errorf("scan promotions: %w", err)
	}

	return &domain.PaginatedResult[domain.Promotion]{
		Items:      promos,
		TotalCount: totalCount,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: int(math.Ceil(float64(totalCount) / float64(pageSize))),
	}, nil
}

func (r *promotionRepo) ListAllActive(ctx context.Context) ([]domain.Promotion, error) {
	query := `SELECT ` + promotionColumns + ` FROM promotions WHERE status = $1 ORDER BY created_at ASC`
	rows, err := r.pool.Query(ctx, query, domain.PromotionActive)
	if err != nil {
		return nil, fmt.Errorf("list active promotions: %w", err)
	}
	defer rows.Close()

	promos, err := scanPromotionRows(rows)
	if err != nil {
		return nil, fmt.Errorf("scan active promotions: %w", err)
	}
	return promos, nil
}

func (r *promotionRepo) RecordImpression(ctx context.Context, promotionID uuid.UUID) error {
	query := `
		UPDATE promotions
		SET impression_count = impression_count + 1, updated_at = now()
		WHERE id = $1
	`
	result, err := r.pool.Exec(ctx, query, promotionID)
	if err != nil {
		return fmt.Errorf("record impression: %w", err)
	}
	if result.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *promotionRepo) RecordClick(ctx context.Context, promotionID uuid.UUID) error {
	query := `
		UPDATE promotions
		SET click_count = click_count + 1, updated_at = now()
		WHERE id = $1
	`
	result, err := r.pool.Exec(ctx, query, promotionID)
	if err != nil {
		return fmt.Errorf("record click: %w", err)
	}
	if result.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}
