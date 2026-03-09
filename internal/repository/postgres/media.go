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

type mediaRepo struct {
	pool *pgxpool.Pool
}

func NewMediaRepository(pool *pgxpool.Pool) repository.MediaRepository {
	return &mediaRepo{pool: pool}
}

var mediaColumns = `id, owner_type, owner_id, user_id, type, url, thumbnail_url, original_name, size, mime_type, width, height, status, created_at`

func scanMedia(row pgx.Row) (*domain.Media, error) {
	var m domain.Media
	err := row.Scan(
		&m.ID, &m.OwnerType, &m.OwnerID, &m.UserID,
		&m.Type, &m.URL, &m.ThumbnailURL, &m.OriginalName,
		&m.Size, &m.MimeType, &m.Width, &m.Height,
		&m.Status, &m.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func scanMediaRows(rows pgx.Rows) ([]domain.Media, error) {
	var items []domain.Media
	for rows.Next() {
		m, err := scanMedia(rows)
		if err != nil {
			return nil, fmt.Errorf("scan media: %w", err)
		}
		items = append(items, *m)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate media rows: %w", err)
	}
	return items, nil
}

func (r *mediaRepo) Create(ctx context.Context, media *domain.Media) error {
	query := `
		INSERT INTO media (id, owner_type, owner_id, user_id, type, url, thumbnail_url, original_name, size, mime_type, width, height, status, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)`

	if media.ID == uuid.Nil {
		media.ID = uuid.New()
	}
	if media.Status == "" {
		media.Status = domain.MediaStatusPending
	}
	if media.CreatedAt.IsZero() {
		media.CreatedAt = time.Now()
	}

	_, err := r.pool.Exec(ctx, query,
		media.ID, string(media.OwnerType), media.OwnerID, media.UserID,
		string(media.Type), media.URL, media.ThumbnailURL, media.OriginalName,
		media.Size, media.MimeType, media.Width, media.Height,
		string(media.Status), media.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("create media: %w", err)
	}
	return nil
}

func (r *mediaRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Media, error) {
	query := `SELECT ` + mediaColumns + ` FROM media WHERE id = $1`

	m, err := scanMedia(r.pool.QueryRow(ctx, query, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrMediaNotFound
		}
		return nil, fmt.Errorf("get media by id: %w", err)
	}
	return m, nil
}

func (r *mediaRepo) ListByOwner(ctx context.Context, ownerType domain.MediaOwnerType, ownerID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.Media], error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}

	var totalCount int64
	countQuery := `SELECT COUNT(*) FROM media WHERE owner_type = $1 AND owner_id = $2`
	err := r.pool.QueryRow(ctx, countQuery, string(ownerType), ownerID).Scan(&totalCount)
	if err != nil {
		return nil, fmt.Errorf("count media: %w", err)
	}

	offset := (page - 1) * pageSize
	query := fmt.Sprintf(`SELECT %s FROM media WHERE owner_type = $1 AND owner_id = $2 ORDER BY created_at DESC LIMIT $3 OFFSET $4`, mediaColumns)

	rows, err := r.pool.Query(ctx, query, string(ownerType), ownerID, pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("list media by owner: %w", err)
	}
	defer rows.Close()

	items, err := scanMediaRows(rows)
	if err != nil {
		return nil, err
	}

	return &domain.PaginatedResult[domain.Media]{
		Items:      items,
		TotalCount: totalCount,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: int(math.Ceil(float64(totalCount) / float64(pageSize))),
	}, nil
}

func (r *mediaRepo) ListByBathhouseReviews(ctx context.Context, bathhouseID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.Media], error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}

	var totalCount int64
	countQuery := `SELECT COUNT(*) FROM media m JOIN reviews rv ON m.owner_type = 'review' AND m.owner_id = rv.id WHERE rv.bathhouse_id = $1 AND m.type = 'image' AND rv.status = 'approved'`
	err := r.pool.QueryRow(ctx, countQuery, bathhouseID).Scan(&totalCount)
	if err != nil {
		return nil, fmt.Errorf("count bathhouse review media: %w", err)
	}

	offset := (page - 1) * pageSize
	query := `SELECT m.id, m.owner_type, m.owner_id, m.user_id, m.type, m.url, m.thumbnail_url, m.original_name, m.size, m.mime_type, m.width, m.height, m.status, m.created_at FROM media m JOIN reviews rv ON m.owner_type = 'review' AND m.owner_id = rv.id WHERE rv.bathhouse_id = $1 AND m.type = 'image' AND rv.status = 'approved' ORDER BY m.created_at DESC LIMIT $2 OFFSET $3`

	rows, err := r.pool.Query(ctx, query, bathhouseID, pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("list media by bathhouse reviews: %w", err)
	}
	defer rows.Close()

	items, err := scanMediaRows(rows)
	if err != nil {
		return nil, err
	}

	return &domain.PaginatedResult[domain.Media]{
		Items:      items,
		TotalCount: totalCount,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: int(math.Ceil(float64(totalCount) / float64(pageSize))),
	}, nil
}

func (r *mediaRepo) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM media WHERE id = $1`
	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete media: %w", err)
	}
	if result.RowsAffected() == 0 {
		return domain.ErrMediaNotFound
	}
	return nil
}

func (r *mediaRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.MediaStatus) error {
	query := `UPDATE media SET status = $2 WHERE id = $1`
	result, err := r.pool.Exec(ctx, query, id, string(status))
	if err != nil {
		return fmt.Errorf("update media status: %w", err)
	}
	if result.RowsAffected() == 0 {
		return domain.ErrMediaNotFound
	}
	return nil
}

func (r *mediaRepo) CountByOwner(ctx context.Context, ownerType domain.MediaOwnerType, ownerID uuid.UUID, mediaType *domain.MediaType) (int64, error) {
	var count int64
	if mediaType != nil {
		query := `SELECT COUNT(*) FROM media WHERE owner_type = $1 AND owner_id = $2 AND type = $3`
		err := r.pool.QueryRow(ctx, query, string(ownerType), ownerID, string(*mediaType)).Scan(&count)
		if err != nil {
			return 0, fmt.Errorf("count media by owner with type: %w", err)
		}
	} else {
		query := `SELECT COUNT(*) FROM media WHERE owner_type = $1 AND owner_id = $2`
		err := r.pool.QueryRow(ctx, query, string(ownerType), ownerID).Scan(&count)
		if err != nil {
			return 0, fmt.Errorf("count media by owner: %w", err)
		}
	}
	return count, nil
}
