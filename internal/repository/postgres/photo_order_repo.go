package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/repository"
)

type photoOrderRepo struct {
	pool *pgxpool.Pool
}

func NewPhotoOrderRepository(pool *pgxpool.Pool) repository.PhotoOrderRepository {
	return &photoOrderRepo{pool: pool}
}

var photoOrderColumns = `id, owner_id, bathhouse_id, region, status, photographer_name, price, scheduled_at, notes, admin_notes, created_at, updated_at`

func scanPhotoOrder(row pgx.Row) (*domain.PhotoOrder, error) {
	var o domain.PhotoOrder
	err := row.Scan(
		&o.ID, &o.OwnerID, &o.BathhouseID, &o.Region, &o.Status,
		&o.PhotographerName, &o.Price, &o.ScheduledAt, &o.Notes, &o.AdminNotes,
		&o.CreatedAt, &o.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrPhotoOrderNotFound
		}
		return nil, fmt.Errorf("scan photo order: %w", err)
	}
	return &o, nil
}

func (r *photoOrderRepo) Create(ctx context.Context, order *domain.PhotoOrder) error {
	if order.ID == uuid.Nil {
		order.ID = uuid.New()
	}
	now := time.Now()
	if order.CreatedAt.IsZero() {
		order.CreatedAt = now
	}
	order.UpdatedAt = now

	query := fmt.Sprintf(`INSERT INTO photo_orders (%s)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`, photoOrderColumns)

	_, err := r.pool.Exec(ctx, query,
		order.ID, order.OwnerID, order.BathhouseID, order.Region, order.Status,
		order.PhotographerName, order.Price, order.ScheduledAt, order.Notes, order.AdminNotes,
		order.CreatedAt, order.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("create photo order: %w", err)
	}
	return nil
}

func (r *photoOrderRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.PhotoOrder, error) {
	query := fmt.Sprintf(`SELECT %s FROM photo_orders WHERE id = $1`, photoOrderColumns)
	return scanPhotoOrder(r.pool.QueryRow(ctx, query, id))
}

func (r *photoOrderRepo) Update(ctx context.Context, order *domain.PhotoOrder) error {
	order.UpdatedAt = time.Now()
	_, err := r.pool.Exec(ctx,
		`UPDATE photo_orders SET status = $2, photographer_name = $3, price = $4, scheduled_at = $5,
		 notes = $6, admin_notes = $7, updated_at = $8 WHERE id = $1`,
		order.ID, order.Status, order.PhotographerName, order.Price, order.ScheduledAt,
		order.Notes, order.AdminNotes, order.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("update photo order: %w", err)
	}
	return nil
}

func (r *photoOrderRepo) List(ctx context.Context, filter domain.PhotoOrderFilter) (*domain.PaginatedResult[domain.PhotoOrder], error) {
	var conditions []string
	var args []interface{}
	argIdx := 1

	if filter.OwnerID != nil {
		conditions = append(conditions, fmt.Sprintf("owner_id = $%d", argIdx))
		args = append(args, *filter.OwnerID)
		argIdx++
	}
	if filter.BathhouseID != nil {
		conditions = append(conditions, fmt.Sprintf("bathhouse_id = $%d", argIdx))
		args = append(args, *filter.BathhouseID)
		argIdx++
	}
	if filter.Status != nil {
		conditions = append(conditions, fmt.Sprintf("status = $%d", argIdx))
		args = append(args, string(*filter.Status))
		argIdx++
	}
	if filter.Region != nil {
		conditions = append(conditions, fmt.Sprintf("region = $%d", argIdx))
		args = append(args, *filter.Region)
		argIdx++
	}

	where := ""
	if len(conditions) > 0 {
		where = "WHERE " + strings.Join(conditions, " AND ")
	}

	var total int64
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM photo_orders %s", where)
	err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("count photo orders: %w", err)
	}

	page := filter.Page
	if page < 1 {
		page = 1
	}
	pageSize := filter.PageSize
	if pageSize < 1 {
		pageSize = 10
	}

	offset := (page - 1) * pageSize
	query := fmt.Sprintf(`SELECT %s FROM photo_orders %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d`,
		photoOrderColumns, where, argIdx, argIdx+1)
	args = append(args, pageSize, offset)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list photo orders: %w", err)
	}
	defer rows.Close()

	var orders []domain.PhotoOrder
	for rows.Next() {
		o, err := scanPhotoOrder(rows)
		if err != nil {
			return nil, err
		}
		orders = append(orders, *o)
	}

	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}

	return &domain.PaginatedResult[domain.PhotoOrder]{
		Items:      orders,
		TotalCount: total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}
