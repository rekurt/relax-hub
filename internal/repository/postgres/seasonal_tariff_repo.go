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

type seasonalTariffRepo struct {
	pool *pgxpool.Pool
}

func NewSeasonalTariffRepository(pool *pgxpool.Pool) repository.SeasonalTariffRepository {
	return &seasonalTariffRepo{pool: pool}
}

func (r *seasonalTariffRepo) Create(ctx context.Context, tariff *domain.SeasonalTariff) error {
	if tariff.ID == uuid.Nil {
		tariff.ID = uuid.New()
	}
	now := time.Now()
	tariff.CreatedAt = now
	tariff.UpdatedAt = now

	query := `
		INSERT INTO seasonal_tariffs (id, bathhouse_id, name, date_from, date_to, multiplier, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	_, err := r.pool.Exec(ctx, query,
		tariff.ID, tariff.BathhouseID, tariff.Name,
		tariff.DateFrom, tariff.DateTo, tariff.Multiplier,
		tariff.IsActive, tariff.CreatedAt, tariff.UpdatedAt,
	)
	if err != nil {
		if isDuplicateKeyError(err) {
			return domain.ErrAlreadyExists
		}
		return fmt.Errorf("create seasonal tariff: %w", err)
	}
	return nil
}

func (r *seasonalTariffRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.SeasonalTariff, error) {
	tariff := &domain.SeasonalTariff{}
	query := `
		SELECT id, bathhouse_id, name, date_from, date_to, multiplier, is_active, created_at, updated_at
		FROM seasonal_tariffs
		WHERE id = $1
	`
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&tariff.ID, &tariff.BathhouseID, &tariff.Name,
		&tariff.DateFrom, &tariff.DateTo, &tariff.Multiplier,
		&tariff.IsActive, &tariff.CreatedAt, &tariff.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("get seasonal tariff by id: %w", err)
	}
	return tariff, nil
}

func (r *seasonalTariffRepo) Update(ctx context.Context, tariff *domain.SeasonalTariff) error {
	tariff.UpdatedAt = time.Now()
	query := `
		UPDATE seasonal_tariffs
		SET name = $1, date_from = $2, date_to = $3, multiplier = $4, is_active = $5, updated_at = $6
		WHERE id = $7
	`
	result, err := r.pool.Exec(ctx, query,
		tariff.Name, tariff.DateFrom, tariff.DateTo, tariff.Multiplier,
		tariff.IsActive, tariff.UpdatedAt, tariff.ID,
	)
	if err != nil {
		return fmt.Errorf("update seasonal tariff: %w", err)
	}
	if result.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *seasonalTariffRepo) Delete(ctx context.Context, id uuid.UUID) error {
	result, err := r.pool.Exec(ctx, "DELETE FROM seasonal_tariffs WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("delete seasonal tariff: %w", err)
	}
	if result.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *seasonalTariffRepo) ListByBathhouse(ctx context.Context, bathhouseID uuid.UUID) ([]domain.SeasonalTariff, error) {
	query := `
		SELECT id, bathhouse_id, name, date_from, date_to, multiplier, is_active, created_at, updated_at
		FROM seasonal_tariffs
		WHERE bathhouse_id = $1
		ORDER BY date_from ASC
	`
	rows, err := r.pool.Query(ctx, query, bathhouseID)
	if err != nil {
		return nil, fmt.Errorf("list seasonal tariffs: %w", err)
	}
	defer rows.Close()

	return r.scanTariffs(rows)
}

func (r *seasonalTariffRepo) GetActiveTariffs(ctx context.Context, bathhouseID uuid.UUID, date time.Time) ([]domain.SeasonalTariff, error) {
	query := `
		SELECT id, bathhouse_id, name, date_from, date_to, multiplier, is_active, created_at, updated_at
		FROM seasonal_tariffs
		WHERE bathhouse_id = $1 AND is_active = true AND date_from <= $2::date AND date_to >= $2::date
		ORDER BY date_from ASC
	`
	rows, err := r.pool.Query(ctx, query, bathhouseID, date)
	if err != nil {
		return nil, fmt.Errorf("get active seasonal tariffs: %w", err)
	}
	defer rows.Close()

	return r.scanTariffs(rows)
}

func (r *seasonalTariffRepo) scanTariffs(rows pgx.Rows) ([]domain.SeasonalTariff, error) {
	var tariffs []domain.SeasonalTariff
	for rows.Next() {
		var t domain.SeasonalTariff
		if err := rows.Scan(
			&t.ID, &t.BathhouseID, &t.Name,
			&t.DateFrom, &t.DateTo, &t.Multiplier,
			&t.IsActive, &t.CreatedAt, &t.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan seasonal tariff: %w", err)
		}
		tariffs = append(tariffs, t)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate seasonal tariff rows: %w", err)
	}
	return tariffs, nil
}
