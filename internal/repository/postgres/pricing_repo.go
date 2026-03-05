package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lib/pq"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/repository"
)

type pricingRuleRepo struct {
	pool *pgxpool.Pool
}

func NewPricingRuleRepository(pool *pgxpool.Pool) repository.PricingRuleRepository {
	return &pricingRuleRepo{pool: pool}
}

func (r *pricingRuleRepo) Create(ctx context.Context, rule *domain.PricingRule) error {
	if rule.ID == uuid.Nil {
		rule.ID = uuid.New()
	}

	rule.CreatedAt = time.Now()

	query := `
		INSERT INTO pricing_rules (id, bathhouse_id, name, type, multiplier, days_of_week, time_from, time_to, date_from, date_to, priority, is_active, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`

	_, err := r.pool.Exec(ctx, query,
		rule.ID, rule.BathhouseID, rule.Name, rule.Type, rule.Multiplier,
		pq.Array(rule.DaysOfWeek), rule.TimeFrom, rule.TimeTo, rule.DateFrom, rule.DateTo,
		rule.Priority, rule.IsActive, rule.CreatedAt,
	)
	if err != nil {
		if isDuplicateKeyError(err) {
			return domain.ErrAlreadyExists
		}
		return fmt.Errorf("create pricing rule: %w", err)
	}
	return nil
}

func (r *pricingRuleRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.PricingRule, error) {
	rule := &domain.PricingRule{}
	query := `
		SELECT id, bathhouse_id, name, type, multiplier, days_of_week, time_from, time_to, date_from, date_to, priority, is_active, created_at
		FROM pricing_rules
		WHERE id = $1
	`

	err := r.pool.QueryRow(ctx, query, id).Scan(
		&rule.ID, &rule.BathhouseID, &rule.Name, &rule.Type, &rule.Multiplier,
		pq.Array(&rule.DaysOfWeek), &rule.TimeFrom, &rule.TimeTo, &rule.DateFrom, &rule.DateTo,
		&rule.Priority, &rule.IsActive, &rule.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("get pricing rule by id: %w", err)
	}
	return rule, nil
}

func (r *pricingRuleRepo) Update(ctx context.Context, rule *domain.PricingRule) error {
	query := `
		UPDATE pricing_rules
		SET name = $1, type = $2, multiplier = $3, days_of_week = $4, time_from = $5, time_to = $6, date_from = $7, date_to = $8, priority = $9, is_active = $10
		WHERE id = $11
	`

	result, err := r.pool.Exec(ctx, query,
		rule.Name, rule.Type, rule.Multiplier, pq.Array(rule.DaysOfWeek),
		rule.TimeFrom, rule.TimeTo, rule.DateFrom, rule.DateTo, rule.Priority, rule.IsActive, rule.ID,
	)
	if err != nil {
		return fmt.Errorf("update pricing rule: %w", err)
	}
	if result.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *pricingRuleRepo) Delete(ctx context.Context, id uuid.UUID) error {
	result, err := r.pool.Exec(ctx, "DELETE FROM pricing_rules WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("delete pricing rule: %w", err)
	}
	if result.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *pricingRuleRepo) ListByBathhouse(ctx context.Context, bathhouseID uuid.UUID) ([]domain.PricingRule, error) {
	query := `
		SELECT id, bathhouse_id, name, type, multiplier, days_of_week, time_from, time_to, date_from, date_to, priority, is_active, created_at
		FROM pricing_rules
		WHERE bathhouse_id = $1
		ORDER BY priority DESC, created_at ASC
	`

	rows, err := r.pool.Query(ctx, query, bathhouseID)
	if err != nil {
		return nil, fmt.Errorf("list pricing rules: %w", err)
	}
	defer rows.Close()

	var rules []domain.PricingRule
	for rows.Next() {
		var rule domain.PricingRule
		if err := rows.Scan(
			&rule.ID, &rule.BathhouseID, &rule.Name, &rule.Type, &rule.Multiplier,
			pq.Array(&rule.DaysOfWeek), &rule.TimeFrom, &rule.TimeTo, &rule.DateFrom, &rule.DateTo,
			&rule.Priority, &rule.IsActive, &rule.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan pricing rule: %w", err)
		}
		rules = append(rules, rule)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate pricing rule rows: %w", err)
	}

	return rules, nil
}

func (r *pricingRuleRepo) GetActiveRules(ctx context.Context, bathhouseID uuid.UUID) ([]domain.PricingRule, error) {
	query := `
		SELECT id, bathhouse_id, name, type, multiplier, days_of_week, time_from, time_to, date_from, date_to, priority, is_active, created_at
		FROM pricing_rules
		WHERE bathhouse_id = $1 AND is_active = true
		ORDER BY priority DESC, created_at ASC
	`

	rows, err := r.pool.Query(ctx, query, bathhouseID)
	if err != nil {
		return nil, fmt.Errorf("get active pricing rules: %w", err)
	}
	defer rows.Close()

	var rules []domain.PricingRule
	for rows.Next() {
		var rule domain.PricingRule
		if err := rows.Scan(
			&rule.ID, &rule.BathhouseID, &rule.Name, &rule.Type, &rule.Multiplier,
			pq.Array(&rule.DaysOfWeek), &rule.TimeFrom, &rule.TimeTo, &rule.DateFrom, &rule.DateTo,
			&rule.Priority, &rule.IsActive, &rule.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan pricing rule: %w", err)
		}
		rules = append(rules, rule)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate pricing rule rows: %w", err)
	}

	return rules, nil
}
