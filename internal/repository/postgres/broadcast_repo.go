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
	"github.com/rekurt/relax-hub/internal/domain"
	"github.com/rekurt/relax-hub/internal/repository"
)

type broadcastRepo struct {
	pool *pgxpool.Pool
}

func NewBroadcastRepository(pool *pgxpool.Pool) repository.BroadcastRepository {
	return &broadcastRepo{pool: pool}
}

var broadcastColumns = `id, owner_id, segment, title, body, image_url, promo_code_id, channels, status, delivered, read, sent_at, created_at, updated_at`

func scanBroadcast(row pgx.Row) (*domain.Broadcast, error) {
	var b domain.Broadcast
	var channels []string
	err := row.Scan(
		&b.ID, &b.OwnerID, &b.Segment, &b.Title, &b.Body, &b.ImageURL,
		&b.PromoCodeID, &channels, &b.Status, &b.Delivered, &b.Read,
		&b.SentAt, &b.CreatedAt, &b.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	b.Channels = make([]domain.BroadcastChannel, len(channels))
	for i, ch := range channels {
		b.Channels[i] = domain.BroadcastChannel(ch)
	}
	return &b, nil
}

func scanBroadcasts(rows pgx.Rows) ([]domain.Broadcast, error) {
	var broadcasts []domain.Broadcast
	for rows.Next() {
		b, err := scanBroadcast(rows)
		if err != nil {
			return nil, fmt.Errorf("scan broadcast: %w", err)
		}
		broadcasts = append(broadcasts, *b)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate broadcast rows: %w", err)
	}
	return broadcasts, nil
}

func (r *broadcastRepo) Create(ctx context.Context, broadcast *domain.Broadcast) error {
	now := time.Now()
	if broadcast.ID == uuid.Nil {
		broadcast.ID = uuid.New()
	}
	if broadcast.CreatedAt.IsZero() {
		broadcast.CreatedAt = now
	}
	if broadcast.UpdatedAt.IsZero() {
		broadcast.UpdatedAt = now
	}

	channels := make([]string, len(broadcast.Channels))
	for i, ch := range broadcast.Channels {
		channels[i] = string(ch)
	}

	query := `INSERT INTO broadcasts (id, owner_id, segment, title, body, image_url, promo_code_id, channels, status, delivered, read, sent_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)`

	_, err := r.pool.Exec(ctx, query,
		broadcast.ID, broadcast.OwnerID, broadcast.Segment, broadcast.Title, broadcast.Body,
		broadcast.ImageURL, broadcast.PromoCodeID, channels, broadcast.Status,
		broadcast.Delivered, broadcast.Read, broadcast.SentAt,
		broadcast.CreatedAt, broadcast.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("create broadcast: %w", err)
	}
	return nil
}

func (r *broadcastRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Broadcast, error) {
	query := `SELECT ` + broadcastColumns + ` FROM broadcasts WHERE id = $1`
	b, err := scanBroadcast(r.pool.QueryRow(ctx, query, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrBroadcastNotFound
		}
		return nil, fmt.Errorf("get broadcast: %w", err)
	}
	return b, nil
}

func (r *broadcastRepo) ListByOwner(ctx context.Context, filter domain.BroadcastFilter) (*domain.PaginatedResult[domain.Broadcast], error) {
	var conditions []string
	var args []interface{}
	argIdx := 1

	conditions = append(conditions, fmt.Sprintf("owner_id = $%d", argIdx))
	args = append(args, filter.OwnerID)
	argIdx++

	if filter.Status != nil {
		conditions = append(conditions, fmt.Sprintf("status = $%d", argIdx))
		args = append(args, *filter.Status)
		argIdx++
	}

	where := strings.Join(conditions, " AND ")

	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM broadcasts WHERE %s`, where)
	var totalCount int64
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&totalCount); err != nil {
		return nil, fmt.Errorf("count broadcasts: %w", err)
	}

	page := filter.Page
	pageSize := filter.PageSize
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}

	totalPages := int((totalCount + int64(pageSize) - 1) / int64(pageSize))
	offset := (page - 1) * pageSize

	dataQuery := fmt.Sprintf(`SELECT %s FROM broadcasts WHERE %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d`,
		broadcastColumns, where, argIdx, argIdx+1)
	args = append(args, pageSize, offset)

	rows, err := r.pool.Query(ctx, dataQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("list broadcasts: %w", err)
	}
	defer rows.Close()

	broadcasts, err := scanBroadcasts(rows)
	if err != nil {
		return nil, err
	}

	return &domain.PaginatedResult[domain.Broadcast]{
		Items:      broadcasts,
		TotalCount: totalCount,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

func (r *broadcastRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.BroadcastStatus) error {
	var query string
	if status == domain.BroadcastStatusSent || status == domain.BroadcastStatusFailed {
		now := time.Now()
		query = `UPDATE broadcasts SET status = $1, sent_at = $3, updated_at = NOW() WHERE id = $2`
		ct, err := r.pool.Exec(ctx, query, status, id, now)
		if err != nil {
			return fmt.Errorf("update broadcast status: %w", err)
		}
		if ct.RowsAffected() == 0 {
			return domain.ErrBroadcastNotFound
		}
		return nil
	}

	query = `UPDATE broadcasts SET status = $1, updated_at = NOW() WHERE id = $2`
	ct, err := r.pool.Exec(ctx, query, status, id)
	if err != nil {
		return fmt.Errorf("update broadcast status: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return domain.ErrBroadcastNotFound
	}
	return nil
}

func (r *broadcastRepo) UpdateStats(ctx context.Context, id uuid.UUID, delivered, read, clicked int64) error {
	query := `UPDATE broadcasts SET delivered = $1, read = $2, clicked = $3, updated_at = NOW() WHERE id = $4`
	ct, err := r.pool.Exec(ctx, query, delivered, read, clicked, id)
	if err != nil {
		return fmt.Errorf("update broadcast stats: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return domain.ErrBroadcastNotFound
	}
	return nil
}

func (r *broadcastRepo) CountRecentByOwner(ctx context.Context, ownerID uuid.UUID, since time.Time) (int64, error) {
	query := `SELECT COUNT(*) FROM broadcasts WHERE owner_id = $1 AND status IN ('sending', 'sent') AND created_at >= $2`
	var count int64
	if err := r.pool.QueryRow(ctx, query, ownerID, since).Scan(&count); err != nil {
		return 0, fmt.Errorf("count recent broadcasts: %w", err)
	}
	return count, nil
}
