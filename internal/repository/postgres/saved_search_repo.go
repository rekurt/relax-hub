package postgres

import (
	"context"
	"fmt"
	"math"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rekurt/relax-hub/internal/domain"
	"github.com/rekurt/relax-hub/internal/repository"
)

type savedSearchRepo struct {
	pool *pgxpool.Pool
}

func NewSavedSearchRepository(pool *pgxpool.Pool) repository.SavedSearchRepository {
	return &savedSearchRepo{pool: pool}
}

func (r *savedSearchRepo) Create(ctx context.Context, search *domain.SavedSearch) error {
	query := `INSERT INTO saved_searches (id, user_id, name, filters, notify_on_new, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)`

	if search.ID == uuid.Nil {
		search.ID = uuid.New()
	}

	_, err := r.pool.Exec(ctx, query,
		search.ID, search.UserID, search.Name, search.Filters, search.NotifyOnNew, search.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("create saved search: %w", err)
	}
	return nil
}

func (r *savedSearchRepo) ListByUser(ctx context.Context, userID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.SavedSearch], error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}

	var totalCount int64
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM saved_searches WHERE user_id = $1`, userID).Scan(&totalCount)
	if err != nil {
		return nil, fmt.Errorf("count saved searches: %w", err)
	}

	offset := (page - 1) * pageSize
	query := `SELECT id, user_id, name, filters, notify_on_new, created_at
		FROM saved_searches WHERE user_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`

	rows, err := r.pool.Query(ctx, query, userID, pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("list saved searches: %w", err)
	}
	defer rows.Close()

	var items []domain.SavedSearch
	for rows.Next() {
		var s domain.SavedSearch
		if err := rows.Scan(&s.ID, &s.UserID, &s.Name, &s.Filters, &s.NotifyOnNew, &s.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan saved search: %w", err)
		}
		items = append(items, s)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate saved search rows: %w", err)
	}

	return &domain.PaginatedResult[domain.SavedSearch]{
		Items:      items,
		TotalCount: totalCount,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: int(math.Ceil(float64(totalCount) / float64(pageSize))),
	}, nil
}

func (r *savedSearchRepo) Delete(ctx context.Context, id uuid.UUID) error {
	result, err := r.pool.Exec(ctx, `DELETE FROM saved_searches WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete saved search: %w", err)
	}
	if result.RowsAffected() == 0 {
		return domain.ErrSavedSearchNotFound
	}
	return nil
}

func (r *savedSearchRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.SavedSearch, error) {
	query := `SELECT id, user_id, name, filters, notify_on_new, created_at
		FROM saved_searches WHERE id = $1`

	var s domain.SavedSearch
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&s.ID, &s.UserID, &s.Name, &s.Filters, &s.NotifyOnNew, &s.CreatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrSavedSearchNotFound
		}
		return nil, fmt.Errorf("get saved search: %w", err)
	}
	return &s, nil
}

func (r *savedSearchRepo) ListWithNotifications(ctx context.Context) ([]domain.SavedSearch, error) {
	query := `SELECT id, user_id, name, filters, notify_on_new, created_at
		FROM saved_searches WHERE notify_on_new = true ORDER BY user_id`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list saved searches with notifications: %w", err)
	}
	defer rows.Close()

	var items []domain.SavedSearch
	for rows.Next() {
		var s domain.SavedSearch
		if err := rows.Scan(&s.ID, &s.UserID, &s.Name, &s.Filters, &s.NotifyOnNew, &s.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan saved search: %w", err)
		}
		items = append(items, s)
	}
	return items, rows.Err()
}
