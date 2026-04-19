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

type subscriptionRepo struct {
	pool *pgxpool.Pool
}

func NewSubscriptionRepository(pool *pgxpool.Pool) repository.SubscriptionRepository {
	return &subscriptionRepo{pool: pool}
}

func (r *subscriptionRepo) Create(ctx context.Context, sub *domain.Subscription) error {
	if sub.ID == uuid.Nil {
		sub.ID = uuid.New()
	}

	now := time.Now()
	sub.CreatedAt = now
	sub.UpdatedAt = now

	query := `
		INSERT INTO subscriptions (id, bathhouse_id, owner_id, plan, status, start_date, end_date, auto_renew, price_kopecks, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	_, err := r.pool.Exec(ctx, query,
		sub.ID, sub.BathhouseID, sub.OwnerID, sub.Plan, sub.Status, sub.StartDate, sub.EndDate, sub.AutoRenew, sub.PriceKopecks, sub.CreatedAt, sub.UpdatedAt,
	)
	if err != nil {
		if isDuplicateKeyError(err) {
			return domain.ErrAlreadyExists
		}
		return fmt.Errorf("create subscription: %w", err)
	}
	return nil
}

func (r *subscriptionRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Subscription, error) {
	sub := &domain.Subscription{}
	query := `
		SELECT id, bathhouse_id, owner_id, plan, status, start_date, end_date, auto_renew, price_kopecks, created_at, updated_at
		FROM subscriptions
		WHERE id = $1
	`

	err := r.pool.QueryRow(ctx, query, id).Scan(
		&sub.ID, &sub.BathhouseID, &sub.OwnerID, &sub.Plan, &sub.Status, &sub.StartDate, &sub.EndDate, &sub.AutoRenew, &sub.PriceKopecks, &sub.CreatedAt, &sub.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("get subscription by id: %w", err)
	}
	return sub, nil
}

func (r *subscriptionRepo) GetActiveBybathhouse(ctx context.Context, bathhouseID uuid.UUID) (*domain.Subscription, error) {
	sub := &domain.Subscription{}
	query := `
		SELECT id, bathhouse_id, owner_id, plan, status, start_date, end_date, auto_renew, price_kopecks, created_at, updated_at
		FROM subscriptions
		WHERE bathhouse_id = $1 AND status = $2
		ORDER BY created_at DESC
		LIMIT 1
	`

	err := r.pool.QueryRow(ctx, query, bathhouseID, domain.SubscriptionActive).Scan(
		&sub.ID, &sub.BathhouseID, &sub.OwnerID, &sub.Plan, &sub.Status, &sub.StartDate, &sub.EndDate, &sub.AutoRenew, &sub.PriceKopecks, &sub.CreatedAt, &sub.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("get active subscription by bathhouse: %w", err)
	}
	return sub, nil
}

func (r *subscriptionRepo) Update(ctx context.Context, sub *domain.Subscription) error {
	sub.UpdatedAt = time.Now()
	query := `
		UPDATE subscriptions
		SET plan = $1, status = $2, start_date = $3, end_date = $4, auto_renew = $5, price_kopecks = $6, updated_at = $7
		WHERE id = $8
	`

	result, err := r.pool.Exec(ctx, query,
		sub.Plan, sub.Status, sub.StartDate, sub.EndDate, sub.AutoRenew, sub.PriceKopecks, sub.UpdatedAt, sub.ID,
	)
	if err != nil {
		return fmt.Errorf("update subscription: %w", err)
	}
	if result.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *subscriptionRepo) ListByOwner(ctx context.Context, ownerID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.Subscription], error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}

	var totalCount int64
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM subscriptions WHERE owner_id = $1`, ownerID).Scan(&totalCount)
	if err != nil {
		return nil, fmt.Errorf("count subscriptions: %w", err)
	}

	offset := (page - 1) * pageSize
	query := `
		SELECT id, bathhouse_id, owner_id, plan, status, start_date, end_date, auto_renew, price_kopecks, created_at, updated_at
		FROM subscriptions
		WHERE owner_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.pool.Query(ctx, query, ownerID, pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("list subscriptions: %w", err)
	}
	defer rows.Close()

	var subs []domain.Subscription
	for rows.Next() {
		var sub domain.Subscription
		if err := rows.Scan(&sub.ID, &sub.BathhouseID, &sub.OwnerID, &sub.Plan, &sub.Status, &sub.StartDate, &sub.EndDate, &sub.AutoRenew, &sub.PriceKopecks, &sub.CreatedAt, &sub.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan subscription: %w", err)
		}
		subs = append(subs, sub)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate subscription rows: %w", err)
	}

	return &domain.PaginatedResult[domain.Subscription]{
		Items:      subs,
		TotalCount: totalCount,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: int(math.Ceil(float64(totalCount) / float64(pageSize))),
	}, nil
}

func (r *subscriptionRepo) GetExpiring(ctx context.Context, before time.Time) ([]domain.Subscription, error) {
	query := `
		SELECT id, bathhouse_id, owner_id, plan, status, start_date, end_date, auto_renew, price_kopecks, created_at, updated_at
		FROM subscriptions
		WHERE status = $1 AND end_date IS NOT NULL AND end_date <= $2
		ORDER BY end_date ASC
	`

	rows, err := r.pool.Query(ctx, query, domain.SubscriptionActive, before)
	if err != nil {
		return nil, fmt.Errorf("get expiring subscriptions: %w", err)
	}
	defer rows.Close()

	var subs []domain.Subscription
	for rows.Next() {
		var sub domain.Subscription
		if err := rows.Scan(&sub.ID, &sub.BathhouseID, &sub.OwnerID, &sub.Plan, &sub.Status, &sub.StartDate, &sub.EndDate, &sub.AutoRenew, &sub.PriceKopecks, &sub.CreatedAt, &sub.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan subscription: %w", err)
		}
		subs = append(subs, sub)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate subscription rows: %w", err)
	}

	return subs, nil
}
