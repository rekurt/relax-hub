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

type autoScenarioRepo struct {
	pool *pgxpool.Pool
}

func NewAutoScenarioRepository(pool *pgxpool.Pool) repository.AutoScenarioRepository {
	return &autoScenarioRepo{pool: pool}
}

var autoScenarioColumns = `id, owner_id, type, enabled, custom_text, channel, delay_hours, promo_code_id, created_at, updated_at`

func scanAutoScenario(row pgx.Row) (*domain.AutoScenario, error) {
	var s domain.AutoScenario
	err := row.Scan(
		&s.ID, &s.OwnerID, &s.Type, &s.Enabled, &s.CustomText,
		&s.Channel, &s.DelayHours, &s.PromoCodeID,
		&s.CreatedAt, &s.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *autoScenarioRepo) Upsert(ctx context.Context, scenario *domain.AutoScenario) error {
	now := time.Now()
	if scenario.ID == uuid.Nil {
		scenario.ID = uuid.New()
	}
	if scenario.CreatedAt.IsZero() {
		scenario.CreatedAt = now
	}
	scenario.UpdatedAt = now

	query := `INSERT INTO auto_scenarios (id, owner_id, type, enabled, custom_text, channel, delay_hours, promo_code_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (owner_id, type) DO UPDATE SET
			enabled = EXCLUDED.enabled,
			custom_text = EXCLUDED.custom_text,
			channel = EXCLUDED.channel,
			delay_hours = EXCLUDED.delay_hours,
			promo_code_id = EXCLUDED.promo_code_id,
			updated_at = EXCLUDED.updated_at
		RETURNING id, created_at`

	err := r.pool.QueryRow(ctx, query,
		scenario.ID, scenario.OwnerID, scenario.Type, scenario.Enabled,
		scenario.CustomText, scenario.Channel, scenario.DelayHours,
		scenario.PromoCodeID, scenario.CreatedAt, scenario.UpdatedAt,
	).Scan(&scenario.ID, &scenario.CreatedAt)
	if err != nil {
		return fmt.Errorf("upsert auto scenario: %w", err)
	}
	return nil
}

func (r *autoScenarioRepo) ListByOwner(ctx context.Context, ownerID uuid.UUID) ([]domain.AutoScenario, error) {
	query := `SELECT ` + autoScenarioColumns + ` FROM auto_scenarios WHERE owner_id = $1 ORDER BY type`
	rows, err := r.pool.Query(ctx, query, ownerID)
	if err != nil {
		return nil, fmt.Errorf("list auto scenarios: %w", err)
	}
	defer rows.Close()

	var scenarios []domain.AutoScenario
	for rows.Next() {
		s, err := scanAutoScenario(rows)
		if err != nil {
			return nil, fmt.Errorf("scan auto scenario: %w", err)
		}
		scenarios = append(scenarios, *s)
	}
	return scenarios, rows.Err()
}

func (r *autoScenarioRepo) GetByOwnerAndType(ctx context.Context, ownerID uuid.UUID, scenarioType domain.AutoScenarioType) (*domain.AutoScenario, error) {
	query := `SELECT ` + autoScenarioColumns + ` FROM auto_scenarios WHERE owner_id = $1 AND type = $2`
	s, err := scanAutoScenario(r.pool.QueryRow(ctx, query, ownerID, scenarioType))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrAutoScenarioNotFound
		}
		return nil, fmt.Errorf("get auto scenario: %w", err)
	}
	return s, nil
}

func (r *autoScenarioRepo) ListEnabled(ctx context.Context) ([]domain.AutoScenario, error) {
	query := `SELECT ` + autoScenarioColumns + ` FROM auto_scenarios WHERE enabled = true ORDER BY owner_id, type`
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list enabled auto scenarios: %w", err)
	}
	defer rows.Close()

	var scenarios []domain.AutoScenario
	for rows.Next() {
		s, err := scanAutoScenario(rows)
		if err != nil {
			return nil, fmt.Errorf("scan auto scenario: %w", err)
		}
		scenarios = append(scenarios, *s)
	}
	return scenarios, rows.Err()
}

func (r *autoScenarioRepo) RecordExecution(ctx context.Context, scenarioID, guestCardID uuid.UUID) error {
	query := `INSERT INTO auto_scenario_executions (id, scenario_id, guest_card_id, executed_at)
		VALUES ($1, $2, $3, NOW())
		ON CONFLICT (scenario_id, guest_card_id) DO NOTHING`
	_, err := r.pool.Exec(ctx, query, uuid.New(), scenarioID, guestCardID)
	if err != nil {
		return fmt.Errorf("record auto scenario execution: %w", err)
	}
	return nil
}

func (r *autoScenarioRepo) HasBeenExecuted(ctx context.Context, scenarioID, guestCardID uuid.UUID) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM auto_scenario_executions WHERE scenario_id = $1 AND guest_card_id = $2)`
	var exists bool
	err := r.pool.QueryRow(ctx, query, scenarioID, guestCardID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("check auto scenario execution: %w", err)
	}
	return exists, nil
}
