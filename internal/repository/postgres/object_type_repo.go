package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/repository"
)

type objectTypeRepo struct {
	pool *pgxpool.Pool
}

func NewObjectTypeRepository(pool *pgxpool.Pool) repository.ObjectTypeRepository {
	return &objectTypeRepo{pool: pool}
}

var objectTypeColumns = `id, name, description, sort_order, is_active, created_at, updated_at`

func scanObjectType(row pgx.Row) (*domain.ObjectType, error) {
	var o domain.ObjectType
	err := row.Scan(&o.ID, &o.Name, &o.Description, &o.SortOrder, &o.IsActive, &o.CreatedAt, &o.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &o, nil
}

func (r *objectTypeRepo) Create(ctx context.Context, objType *domain.ObjectType) error {
	if objType.ID == uuid.Nil {
		objType.ID = uuid.New()
	}
	now := time.Now()
	objType.CreatedAt = now
	objType.UpdatedAt = now

	query := `INSERT INTO object_types (id, name, description, sort_order, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`
	_, err := r.pool.Exec(ctx, query,
		objType.ID, objType.Name, objType.Description, objType.SortOrder, objType.IsActive,
		objType.CreatedAt, objType.UpdatedAt)
	return err
}

func (r *objectTypeRepo) Update(ctx context.Context, objType *domain.ObjectType) error {
	objType.UpdatedAt = time.Now()
	query := `UPDATE object_types SET name = $1, description = $2, sort_order = $3, is_active = $4, updated_at = $5 WHERE id = $6`
	tag, err := r.pool.Exec(ctx, query,
		objType.Name, objType.Description, objType.SortOrder, objType.IsActive, objType.UpdatedAt, objType.ID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *objectTypeRepo) Delete(ctx context.Context, id uuid.UUID) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM object_types WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *objectTypeRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.ObjectType, error) {
	query := `SELECT ` + objectTypeColumns + ` FROM object_types WHERE id = $1`
	return scanObjectType(r.pool.QueryRow(ctx, query, id))
}

func (r *objectTypeRepo) ListAll(ctx context.Context) ([]domain.ObjectType, error) {
	query := `SELECT ` + objectTypeColumns + ` FROM object_types ORDER BY sort_order, name`
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var types []domain.ObjectType
	for rows.Next() {
		var o domain.ObjectType
		if err := rows.Scan(&o.ID, &o.Name, &o.Description, &o.SortOrder, &o.IsActive, &o.CreatedAt, &o.UpdatedAt); err != nil {
			return nil, err
		}
		types = append(types, o)
	}
	return types, rows.Err()
}
