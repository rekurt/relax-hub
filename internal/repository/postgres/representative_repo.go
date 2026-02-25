package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/repository"
)

type representativeRepo struct {
	pool *pgxpool.Pool
}

func NewRepresentativeRepository(pool *pgxpool.Pool) repository.RepresentativeRepository {
	return &representativeRepo{pool: pool}
}

func (r *representativeRepo) Create(ctx context.Context, rep *domain.Representative) error {
	query := `
		INSERT INTO representatives (id, user_id, bathhouse_id, owner_id, created_at)
		VALUES ($1, $2, $3, $4, $5)`

	if rep.ID == uuid.Nil {
		rep.ID = uuid.New()
	}
	rep.CreatedAt = time.Now()

	_, err := r.pool.Exec(ctx, query,
		rep.ID, rep.UserID, rep.BathhouseID, rep.OwnerID, rep.CreatedAt,
	)
	if err != nil {
		if isDuplicateKeyError(err) {
			return domain.ErrAlreadyExists
		}
		return fmt.Errorf("create representative: %w", err)
	}
	return nil
}

func (r *representativeRepo) Delete(ctx context.Context, id uuid.UUID) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM representatives WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete representative: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *representativeRepo) GetByUserAndBathhouse(ctx context.Context, userID, bathhouseID uuid.UUID) (*domain.Representative, error) {
	query := `
		SELECT id, user_id, bathhouse_id, owner_id, created_at
		FROM representatives WHERE user_id = $1 AND bathhouse_id = $2`

	var rep domain.Representative
	err := r.pool.QueryRow(ctx, query, userID, bathhouseID).Scan(
		&rep.ID, &rep.UserID, &rep.BathhouseID, &rep.OwnerID, &rep.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("get representative by user and bathhouse: %w", err)
	}
	return &rep, nil
}

func (r *representativeRepo) ListByBathhouse(ctx context.Context, bathhouseID uuid.UUID) ([]domain.Representative, error) {
	query := `
		SELECT id, user_id, bathhouse_id, owner_id, created_at
		FROM representatives WHERE bathhouse_id = $1 ORDER BY created_at`

	rows, err := r.pool.Query(ctx, query, bathhouseID)
	if err != nil {
		return nil, fmt.Errorf("list representatives by bathhouse: %w", err)
	}
	defer rows.Close()

	var reps []domain.Representative
	for rows.Next() {
		var rep domain.Representative
		if err := rows.Scan(
			&rep.ID, &rep.UserID, &rep.BathhouseID, &rep.OwnerID, &rep.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan representative: %w", err)
		}
		reps = append(reps, rep)
	}
	return reps, nil
}

func (r *representativeRepo) ListByUser(ctx context.Context, userID uuid.UUID) ([]domain.Representative, error) {
	query := `
		SELECT id, user_id, bathhouse_id, owner_id, created_at
		FROM representatives WHERE user_id = $1 ORDER BY created_at`

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("list representatives by user: %w", err)
	}
	defer rows.Close()

	var reps []domain.Representative
	for rows.Next() {
		var rep domain.Representative
		if err := rows.Scan(
			&rep.ID, &rep.UserID, &rep.BathhouseID, &rep.OwnerID, &rep.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan representative: %w", err)
		}
		reps = append(reps, rep)
	}
	return reps, nil
}

func (r *representativeRepo) ListBathhouseIDsByUser(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error) {
	query := `SELECT bathhouse_id FROM representatives WHERE user_id = $1`

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("list bathhouse ids by user: %w", err)
	}
	defer rows.Close()

	var ids []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scan bathhouse id: %w", err)
		}
		ids = append(ids, id)
	}
	return ids, nil
}
