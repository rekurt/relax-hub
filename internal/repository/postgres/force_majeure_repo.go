package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/repository"
)

type forceMajeureRepo struct {
	pool *pgxpool.Pool
}

func NewForceMajeureRepository(pool *pgxpool.Pool) repository.ForceMajeureRepository {
	return &forceMajeureRepo{pool: pool}
}

func (r *forceMajeureRepo) Create(ctx context.Context, event *domain.ForceMajeureEvent) error {
	if event.ID == uuid.Nil {
		event.ID = uuid.New()
	}

	query := `
		INSERT INTO force_majeure_events (id, admin_id, region, date_from, date_to, reason, affected_count, total_refund, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, now())
		RETURNING created_at`

	err := r.pool.QueryRow(ctx, query,
		event.ID, event.AdminID, event.Region, event.DateFrom, event.DateTo,
		event.Reason, event.AffectedCount, event.TotalRefund,
	).Scan(&event.CreatedAt)
	if err != nil {
		return fmt.Errorf("create force majeure event: %w", err)
	}
	return nil
}

func (r *forceMajeureRepo) List(ctx context.Context) ([]domain.ForceMajeureEvent, error) {
	query := `
		SELECT id, admin_id, region, date_from, date_to, reason, affected_count, total_refund, created_at
		FROM force_majeure_events
		ORDER BY created_at DESC`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list force majeure events: %w", err)
	}
	defer rows.Close()

	var events []domain.ForceMajeureEvent
	for rows.Next() {
		var e domain.ForceMajeureEvent
		err := rows.Scan(&e.ID, &e.AdminID, &e.Region, &e.DateFrom, &e.DateTo,
			&e.Reason, &e.AffectedCount, &e.TotalRefund, &e.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("scan force majeure event: %w", err)
		}
		events = append(events, e)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate force majeure events: %w", err)
	}
	return events, nil
}
