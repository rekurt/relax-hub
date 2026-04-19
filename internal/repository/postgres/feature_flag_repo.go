package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rekurt/relax-hub/internal/domain"
	"github.com/rekurt/relax-hub/internal/repository"
)

type featureFlagRepo struct {
	pool *pgxpool.Pool
}

func NewFeatureFlagRepository(pool *pgxpool.Pool) repository.FeatureFlagRepository {
	return &featureFlagRepo{pool: pool}
}

func (r *featureFlagRepo) Get(ctx context.Context, key string) (*domain.FeatureFlag, error) {
	query := `SELECT key, enabled, description, region, updated_at, updated_by FROM feature_flags WHERE key = $1`

	var f domain.FeatureFlag
	var updatedBy *uuid.UUID
	err := r.pool.QueryRow(ctx, query, key).Scan(
		&f.Key, &f.Enabled, &f.Description, &f.Region, &f.UpdatedAt, &updatedBy,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("get feature flag: %w", err)
	}
	if updatedBy != nil {
		str := updatedBy.String()
		f.UpdatedBy = &str
	}
	return &f, nil
}

func (r *featureFlagRepo) GetAll(ctx context.Context) ([]domain.FeatureFlag, error) {
	query := `SELECT key, enabled, description, region, updated_at, updated_by FROM feature_flags ORDER BY key`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("get all feature flags: %w", err)
	}
	defer rows.Close()

	var flags []domain.FeatureFlag
	for rows.Next() {
		var f domain.FeatureFlag
		var updatedBy *uuid.UUID
		if err := rows.Scan(&f.Key, &f.Enabled, &f.Description, &f.Region, &f.UpdatedAt, &updatedBy); err != nil {
			return nil, fmt.Errorf("scan feature flag: %w", err)
		}
		if updatedBy != nil {
			str := updatedBy.String()
			f.UpdatedBy = &str
		}
		flags = append(flags, f)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate feature flag rows: %w", err)
	}
	return flags, nil
}

func (r *featureFlagRepo) Set(ctx context.Context, key string, enabled bool, region *string, updatedBy *uuid.UUID) error {
	query := `UPDATE feature_flags SET enabled = $2, region = $3, updated_at = NOW(), updated_by = $4 WHERE key = $1`

	tag, err := r.pool.Exec(ctx, query, key, enabled, region, updatedBy)
	if err != nil {
		return fmt.Errorf("set feature flag: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}
