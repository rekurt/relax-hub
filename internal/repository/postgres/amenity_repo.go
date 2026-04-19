package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rekurt/relax-hub/internal/domain"
	"github.com/rekurt/relax-hub/internal/repository"
)

type amenityRepo struct {
	pool *pgxpool.Pool
}

func NewAmenityRepository(pool *pgxpool.Pool) repository.AmenityRepository {
	return &amenityRepo{pool: pool}
}

var amenityColumns = `id, name, icon, sort_order, is_active, created_at, updated_at`

func scanAmenity(row pgx.Row) (*domain.Amenity, error) {
	var a domain.Amenity
	err := row.Scan(&a.ID, &a.Name, &a.Icon, &a.SortOrder, &a.IsActive, &a.CreatedAt, &a.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &a, nil
}

func (r *amenityRepo) Create(ctx context.Context, amenity *domain.Amenity) error {
	if amenity.ID == uuid.Nil {
		amenity.ID = uuid.New()
	}
	now := time.Now()
	amenity.CreatedAt = now
	amenity.UpdatedAt = now

	query := `INSERT INTO amenities (id, name, icon, sort_order, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`
	_, err := r.pool.Exec(ctx, query,
		amenity.ID, amenity.Name, amenity.Icon, amenity.SortOrder, amenity.IsActive,
		amenity.CreatedAt, amenity.UpdatedAt)
	return err
}

func (r *amenityRepo) Update(ctx context.Context, amenity *domain.Amenity) error {
	amenity.UpdatedAt = time.Now()
	query := `UPDATE amenities SET name = $1, icon = $2, sort_order = $3, is_active = $4, updated_at = $5 WHERE id = $6`
	tag, err := r.pool.Exec(ctx, query,
		amenity.Name, amenity.Icon, amenity.SortOrder, amenity.IsActive, amenity.UpdatedAt, amenity.ID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *amenityRepo) Delete(ctx context.Context, id uuid.UUID) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM amenities WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *amenityRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Amenity, error) {
	query := `SELECT ` + amenityColumns + ` FROM amenities WHERE id = $1`
	return scanAmenity(r.pool.QueryRow(ctx, query, id))
}

func (r *amenityRepo) ListAll(ctx context.Context) ([]domain.Amenity, error) {
	query := `SELECT ` + amenityColumns + ` FROM amenities ORDER BY sort_order, name`
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var amenities []domain.Amenity
	for rows.Next() {
		var a domain.Amenity
		if err := rows.Scan(&a.ID, &a.Name, &a.Icon, &a.SortOrder, &a.IsActive, &a.CreatedAt, &a.UpdatedAt); err != nil {
			return nil, err
		}
		amenities = append(amenities, a)
	}
	return amenities, rows.Err()
}
