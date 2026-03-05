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

type promotionRepo struct {
	pool *pgxpool.Pool
}

func NewPromotionRepository(pool *pgxpool.Pool) repository.PromotionRepository {
	return &promotionRepo{pool: pool}
}

func (r *promotionRepo) Create(ctx context.Context, promo *domain.Promotion) error {
	if promo.ID == uuid.Nil {
		promo.ID = uuid.New()
	}

	now := time.Now()
	promo.CreatedAt = now
	promo.UpdatedAt = now

	query := `
		INSERT INTO promotions (id, bathhouse_id, budget_kopecks, spent_kopecks, start_date, end_date, target_city_id, status, impression_count, click_count, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`

	_, err := r.pool.Exec(ctx, query,
		promo.ID, promo.BathhouseID, promo.BudgetKopecks, promo.SpentKopecks, promo.StartDate, promo.EndDate, promo.TargetCityID, promo.Status, promo.ImpressionCount, promo.ClickCount, promo.CreatedAt, promo.UpdatedAt,
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
	promo := &domain.Promotion{}
	query := `
		SELECT id, bathhouse_id, budget_kopecks, spent_kopecks, start_date, end_date, target_city_id, status, impression_count, click_count, created_at, updated_at
		FROM promotions
		WHERE id = $1
	`

	err := r.pool.QueryRow(ctx, query, id).Scan(
		&promo.ID, &promo.BathhouseID, &promo.BudgetKopecks, &promo.SpentKopecks, &promo.StartDate, &promo.EndDate, &promo.TargetCityID, &promo.Status, &promo.ImpressionCount, &promo.ClickCount, &promo.CreatedAt, &promo.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("get promotion by id: %w", err)
	}
	return promo, nil
}

func (r *promotionRepo) GetActiveBybathhouse(ctx context.Context, bathhouseID uuid.UUID) (*domain.Promotion, error) {
	promo := &domain.Promotion{}
	query := `
		SELECT id, bathhouse_id, budget_kopecks, spent_kopecks, start_date, end_date, target_city_id, status, impression_count, click_count, created_at, updated_at
		FROM promotions
		WHERE bathhouse_id = $1 AND status = $2
		ORDER BY created_at DESC
		LIMIT 1
	`

	err := r.pool.QueryRow(ctx, query, bathhouseID, domain.PromotionActive).Scan(
		&promo.ID, &promo.BathhouseID, &promo.BudgetKopecks, &promo.SpentKopecks, &promo.StartDate, &promo.EndDate, &promo.TargetCityID, &promo.Status, &promo.ImpressionCount, &promo.ClickCount, &promo.CreatedAt, &promo.UpdatedAt,
	)
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
		SET budget_kopecks = $1, spent_kopecks = $2, start_date = $3, end_date = $4, target_city_id = $5, status = $6, impression_count = $7, click_count = $8, updated_at = now()
		WHERE id = $9
	`

	result, err := r.pool.Exec(ctx, query,
		promo.BudgetKopecks, promo.SpentKopecks, promo.StartDate, promo.EndDate, promo.TargetCityID, promo.Status, promo.ImpressionCount, promo.ClickCount, promo.ID,
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
	query := `
		SELECT p.id, p.bathhouse_id, p.budget_kopecks, p.spent_kopecks, p.start_date, p.end_date, p.target_city_id, p.status, p.impression_count, p.click_count, p.created_at, p.updated_at
		FROM promotions p
		JOIN bathhouses b ON p.bathhouse_id = b.id
		WHERE b.owner_id = $1
		ORDER BY p.created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.pool.Query(ctx, query, ownerID, pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("list promotions: %w", err)
	}
	defer rows.Close()

	var promos []domain.Promotion
	for rows.Next() {
		var promo domain.Promotion
		if err := rows.Scan(&promo.ID, &promo.BathhouseID, &promo.BudgetKopecks, &promo.SpentKopecks, &promo.StartDate, &promo.EndDate, &promo.TargetCityID, &promo.Status, &promo.ImpressionCount, &promo.ClickCount, &promo.CreatedAt, &promo.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan promotion: %w", err)
		}
		promos = append(promos, promo)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate promotion rows: %w", err)
	}

	return &domain.PaginatedResult[domain.Promotion]{
		Items:      promos,
		TotalCount: totalCount,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: int(math.Ceil(float64(totalCount) / float64(pageSize))),
	}, nil
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
