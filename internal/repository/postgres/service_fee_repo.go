package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rekurt/relax-hub/internal/domain"
	"github.com/rekurt/relax-hub/internal/repository"
)

type serviceFeeRepo struct {
	pool *pgxpool.Pool
}

func NewServiceFeeRepository(pool *pgxpool.Pool) repository.ServiceFeeRepository {
	return &serviceFeeRepo{pool: pool}
}

var serviceFeeColumns = `id, region, category, fee_percent, created_at, updated_at`

func scanServiceFeeConfig(row pgx.Row) (*domain.ServiceFeeConfig, error) {
	var c domain.ServiceFeeConfig
	err := row.Scan(&c.ID, &c.Region, &c.Category, &c.FeePercent, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &c, nil
}

func (r *serviceFeeRepo) GetByRegionAndCategory(ctx context.Context, region string, category *string) (*domain.ServiceFeeConfig, error) {
	query := `SELECT ` + serviceFeeColumns + ` FROM service_fee_configs WHERE region = $1 AND category = $2`
	return scanServiceFeeConfig(r.pool.QueryRow(ctx, query, region, category))
}

func (r *serviceFeeRepo) GetByRegion(ctx context.Context, region string) (*domain.ServiceFeeConfig, error) {
	query := `SELECT ` + serviceFeeColumns + ` FROM service_fee_configs WHERE region = $1 AND category IS NULL`
	return scanServiceFeeConfig(r.pool.QueryRow(ctx, query, region))
}

func (r *serviceFeeRepo) GetGlobalDefault(ctx context.Context) (*domain.ServiceFeeConfig, error) {
	query := `SELECT ` + serviceFeeColumns + ` FROM service_fee_configs WHERE region = '*' AND category IS NULL`
	return scanServiceFeeConfig(r.pool.QueryRow(ctx, query))
}

func (r *serviceFeeRepo) List(ctx context.Context) ([]domain.ServiceFeeConfig, error) {
	query := `SELECT ` + serviceFeeColumns + ` FROM service_fee_configs ORDER BY region, category NULLS FIRST`
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var configs []domain.ServiceFeeConfig
	for rows.Next() {
		var c domain.ServiceFeeConfig
		if err := rows.Scan(&c.ID, &c.Region, &c.Category, &c.FeePercent, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		configs = append(configs, c)
	}
	return configs, rows.Err()
}

func (r *serviceFeeRepo) Upsert(ctx context.Context, config *domain.ServiceFeeConfig) error {
	now := time.Now()
	var query string
	if config.Category == nil {
		// Use partial unique index for NULL category
		query = `
			INSERT INTO service_fee_configs (region, category, fee_percent, created_at, updated_at)
			VALUES ($1, NULL, $2, $3, $3)
			ON CONFLICT (region) WHERE category IS NULL DO UPDATE SET fee_percent = $2, updated_at = $3
			RETURNING id, created_at, updated_at`
		return r.pool.QueryRow(ctx, query, config.Region, config.FeePercent, now).
			Scan(&config.ID, &config.CreatedAt, &config.UpdatedAt)
	}
	query = `
		INSERT INTO service_fee_configs (region, category, fee_percent, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $4)
		ON CONFLICT (region, category) WHERE category IS NOT NULL DO UPDATE SET fee_percent = $3, updated_at = $4
		RETURNING id, created_at, updated_at`
	return r.pool.QueryRow(ctx, query, config.Region, config.Category, config.FeePercent, now).
		Scan(&config.ID, &config.CreatedAt, &config.UpdatedAt)
}
