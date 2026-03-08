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

type bathhousePhotoRepo struct {
	pool *pgxpool.Pool
}

func NewBathhousePhotoRepository(pool *pgxpool.Pool) repository.BathhousePhotoRepository {
	return &bathhousePhotoRepo{pool: pool}
}

var photoColumns = `id, bathhouse_id, url, thumbnail_url, position, status, verified_by_id, verified_at, rejection_reason, uploaded_at`

func scanPhoto(row pgx.Row) (*domain.BathhousePhoto, error) {
	var p domain.BathhousePhoto
	err := row.Scan(
		&p.ID, &p.BathhouseID, &p.URL, &p.ThumbnailURL,
		&p.Position, &p.Status, &p.VerifiedByID,
		&p.VerifiedAt, &p.RejectionReason, &p.UploadedAt,
	)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func scanPhotos(rows pgx.Rows) ([]domain.BathhousePhoto, error) {
	var photos []domain.BathhousePhoto
	for rows.Next() {
		p, err := scanPhoto(rows)
		if err != nil {
			return nil, fmt.Errorf("scan photo: %w", err)
		}
		photos = append(photos, *p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate photo rows: %w", err)
	}
	return photos, nil
}

func (r *bathhousePhotoRepo) Create(ctx context.Context, photo *domain.BathhousePhoto) error {
	if photo.ID == uuid.Nil {
		photo.ID = uuid.New()
	}
	if photo.Status == "" {
		photo.Status = domain.PhotoStatusPending
	}
	if photo.UploadedAt.IsZero() {
		photo.UploadedAt = time.Now()
	}

	query := `
		INSERT INTO bathhouse_photos (id, bathhouse_id, url, thumbnail_url, position, status, uploaded_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`

	_, err := r.pool.Exec(ctx, query,
		photo.ID, photo.BathhouseID, photo.URL, photo.ThumbnailURL,
		photo.Position, string(photo.Status), photo.UploadedAt,
	)
	if err != nil {
		return fmt.Errorf("create bathhouse photo: %w", err)
	}
	return nil
}

func (r *bathhousePhotoRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.BathhousePhoto, error) {
	query := `SELECT ` + photoColumns + ` FROM bathhouse_photos WHERE id = $1`
	p, err := scanPhoto(r.pool.QueryRow(ctx, query, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrPhotoNotFound
		}
		return nil, fmt.Errorf("get photo by id: %w", err)
	}
	return p, nil
}

func (r *bathhousePhotoRepo) Delete(ctx context.Context, id uuid.UUID) error {
	result, err := r.pool.Exec(ctx, `DELETE FROM bathhouse_photos WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete photo: %w", err)
	}
	if result.RowsAffected() == 0 {
		return domain.ErrPhotoNotFound
	}
	return nil
}

func (r *bathhousePhotoRepo) ListByBathhouse(ctx context.Context, bathhouseID uuid.UUID) ([]domain.BathhousePhoto, error) {
	query := `SELECT ` + photoColumns + ` FROM bathhouse_photos WHERE bathhouse_id = $1 ORDER BY position ASC`
	rows, err := r.pool.Query(ctx, query, bathhouseID)
	if err != nil {
		return nil, fmt.Errorf("list photos by bathhouse: %w", err)
	}
	defer rows.Close()
	return scanPhotos(rows)
}

func (r *bathhousePhotoRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.PhotoStatus, verifiedByID *uuid.UUID, rejectionReason string) error {
	var verifiedAt *time.Time
	if status == domain.PhotoStatusVerified || status == domain.PhotoStatusRejected {
		now := time.Now()
		verifiedAt = &now
	}

	query := `UPDATE bathhouse_photos SET status = $2, verified_by_id = $3, verified_at = $4, rejection_reason = $5 WHERE id = $1`
	result, err := r.pool.Exec(ctx, query, id, string(status), verifiedByID, verifiedAt, rejectionReason)
	if err != nil {
		return fmt.Errorf("update photo status: %w", err)
	}
	if result.RowsAffected() == 0 {
		return domain.ErrPhotoNotFound
	}
	return nil
}

func (r *bathhousePhotoRepo) Reorder(ctx context.Context, bathhouseID uuid.UUID, photoIDs []uuid.UUID) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	for i, photoID := range photoIDs {
		result, err := tx.Exec(ctx,
			`UPDATE bathhouse_photos SET position = $1 WHERE id = $2 AND bathhouse_id = $3`,
			i, photoID, bathhouseID,
		)
		if err != nil {
			return fmt.Errorf("reorder photo %d: %w", i, err)
		}
		if result.RowsAffected() == 0 {
			return domain.ErrPhotoNotFound
		}
	}

	return tx.Commit(ctx)
}

func (r *bathhousePhotoRepo) CountByBathhouseAndStatus(ctx context.Context, bathhouseID uuid.UUID, status domain.PhotoStatus) (int64, error) {
	var count int64
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM bathhouse_photos WHERE bathhouse_id = $1 AND status = $2`,
		bathhouseID, string(status),
	).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count photos by status: %w", err)
	}
	return count, nil
}

func (r *bathhousePhotoRepo) ListPending(ctx context.Context, page, pageSize int) (*domain.PaginatedResult[domain.BathhousePhoto], error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}

	var totalCount int64
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM bathhouse_photos WHERE status = 'pending'`,
	).Scan(&totalCount)
	if err != nil {
		return nil, fmt.Errorf("count pending photos: %w", err)
	}

	offset := (page - 1) * pageSize
	query := fmt.Sprintf(`SELECT %s FROM bathhouse_photos WHERE status = 'pending' ORDER BY uploaded_at ASC LIMIT $1 OFFSET $2`, photoColumns)
	rows, err := r.pool.Query(ctx, query, pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("list pending photos: %w", err)
	}
	defer rows.Close()

	photos, err := scanPhotos(rows)
	if err != nil {
		return nil, err
	}

	return &domain.PaginatedResult[domain.BathhousePhoto]{
		Items:      photos,
		TotalCount: totalCount,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: int(math.Ceil(float64(totalCount) / float64(pageSize))),
	}, nil
}
