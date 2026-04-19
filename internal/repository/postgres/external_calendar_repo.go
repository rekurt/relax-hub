package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rekurt/relax-hub/internal/domain"
	"github.com/rekurt/relax-hub/internal/repository"
)

type externalCalendarRepo struct {
	pool *pgxpool.Pool
}

func NewExternalCalendarRepository(pool *pgxpool.Pool) repository.ExternalCalendarRepository {
	return &externalCalendarRepo{pool: pool}
}

func (r *externalCalendarRepo) Create(ctx context.Context, cal *domain.ExternalCalendar) error {
	if cal.ID == uuid.Nil {
		cal.ID = uuid.New()
	}
	cal.CreatedAt = time.Now()

	query := `INSERT INTO external_calendars (id, bathhouse_id, url, source, last_error, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)`

	_, err := r.pool.Exec(ctx, query,
		cal.ID, cal.BathhouseID, cal.URL, cal.Source, cal.LastError, cal.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("create external calendar: %w", err)
	}
	return nil
}

func (r *externalCalendarRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.ExternalCalendar, error) {
	query := `SELECT id, bathhouse_id, url, source, last_sync_at, last_error, created_at
		FROM external_calendars WHERE id = $1`

	var cal domain.ExternalCalendar
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&cal.ID, &cal.BathhouseID, &cal.URL, &cal.Source,
		&cal.LastSyncAt, &cal.LastError, &cal.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("get external calendar: %w", err)
	}
	return &cal, nil
}

func (r *externalCalendarRepo) ListByBathhouse(ctx context.Context, bathhouseID uuid.UUID) ([]domain.ExternalCalendar, error) {
	query := `SELECT id, bathhouse_id, url, source, last_sync_at, last_error, created_at
		FROM external_calendars WHERE bathhouse_id = $1 ORDER BY created_at`

	rows, err := r.pool.Query(ctx, query, bathhouseID)
	if err != nil {
		return nil, fmt.Errorf("list external calendars: %w", err)
	}
	defer rows.Close()

	var calendars []domain.ExternalCalendar
	for rows.Next() {
		var cal domain.ExternalCalendar
		if err := rows.Scan(&cal.ID, &cal.BathhouseID, &cal.URL, &cal.Source,
			&cal.LastSyncAt, &cal.LastError, &cal.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan external calendar: %w", err)
		}
		calendars = append(calendars, cal)
	}
	return calendars, nil
}

func (r *externalCalendarRepo) ListAll(ctx context.Context) ([]domain.ExternalCalendar, error) {
	query := `SELECT id, bathhouse_id, url, source, last_sync_at, last_error, created_at
		FROM external_calendars ORDER BY created_at`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list all external calendars: %w", err)
	}
	defer rows.Close()

	var calendars []domain.ExternalCalendar
	for rows.Next() {
		var cal domain.ExternalCalendar
		if err := rows.Scan(&cal.ID, &cal.BathhouseID, &cal.URL, &cal.Source,
			&cal.LastSyncAt, &cal.LastError, &cal.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan external calendar: %w", err)
		}
		calendars = append(calendars, cal)
	}
	return calendars, nil
}

func (r *externalCalendarRepo) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM external_calendars WHERE id = $1`
	tag, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete external calendar: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *externalCalendarRepo) UpdateSyncStatus(ctx context.Context, id uuid.UUID, syncedAt time.Time, lastError string) error {
	query := `UPDATE external_calendars SET last_sync_at = $2, last_error = $3 WHERE id = $1`
	tag, err := r.pool.Exec(ctx, query, id, syncedAt, lastError)
	if err != nil {
		return fmt.Errorf("update sync status: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}
