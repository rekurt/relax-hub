package postgres

import (
	"context"
	"errors"
	"fmt"
	"math"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/repository"
)

type favoriteRepo struct {
	pool *pgxpool.Pool
}

func NewFavoriteRepository(pool *pgxpool.Pool) repository.FavoriteRepository {
	return &favoriteRepo{pool: pool}
}

func (r *favoriteRepo) Add(ctx context.Context, favorite *domain.Favorite) error {
	query := `INSERT INTO favorites (id, user_id, bathhouse_id, created_at) VALUES ($1, $2, $3, $4)`

	if favorite.ID == uuid.Nil {
		favorite.ID = uuid.New()
	}

	_, err := r.pool.Exec(ctx, query,
		favorite.ID, favorite.UserID, favorite.BathhouseID, favorite.CreatedAt,
	)
	if err != nil {
		if isDuplicateKeyError(err) {
			return domain.ErrAlreadyExists
		}
		return fmt.Errorf("add favorite: %w", err)
	}
	return nil
}

func (r *favoriteRepo) Remove(ctx context.Context, userID, bathhouseID uuid.UUID) error {
	result, err := r.pool.Exec(ctx,
		`DELETE FROM favorites WHERE user_id = $1 AND bathhouse_id = $2`,
		userID, bathhouseID,
	)
	if err != nil {
		return fmt.Errorf("remove favorite: %w", err)
	}
	if result.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *favoriteRepo) ListByUser(ctx context.Context, userID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.Favorite], error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}

	var totalCount int64
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM favorites WHERE user_id = $1`, userID).Scan(&totalCount)
	if err != nil {
		return nil, fmt.Errorf("count favorites: %w", err)
	}

	offset := (page - 1) * pageSize
	query := `SELECT id, user_id, bathhouse_id, created_at FROM favorites WHERE user_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`

	rows, err := r.pool.Query(ctx, query, userID, pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("list favorites: %w", err)
	}
	defer rows.Close()

	var favorites []domain.Favorite
	for rows.Next() {
		var f domain.Favorite
		if err := rows.Scan(&f.ID, &f.UserID, &f.BathhouseID, &f.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan favorite: %w", err)
		}
		favorites = append(favorites, f)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate favorite rows: %w", err)
	}

	return &domain.PaginatedResult[domain.Favorite]{
		Items:      favorites,
		TotalCount: totalCount,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: int(math.Ceil(float64(totalCount) / float64(pageSize))),
	}, nil
}

func (r *favoriteRepo) IsFavorite(ctx context.Context, userID, bathhouseID uuid.UUID) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM favorites WHERE user_id = $1 AND bathhouse_id = $2)`,
		userID, bathhouseID,
	).Scan(&exists)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, fmt.Errorf("check favorite: %w", err)
	}
	return exists, nil
}

func (r *favoriteRepo) CountByUser(ctx context.Context, userID uuid.UUID) (int64, error) {
	var count int64
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM favorites WHERE user_id = $1`, userID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count favorites by user: %w", err)
	}
	return count, nil
}
