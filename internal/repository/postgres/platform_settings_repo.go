package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/repository"
)

type platformSettingsRepo struct {
	pool *pgxpool.Pool
}

func NewPlatformSettingsRepository(pool *pgxpool.Pool) repository.PlatformSettingsRepository {
	return &platformSettingsRepo{pool: pool}
}

func (r *platformSettingsRepo) Get(ctx context.Context, key string) (*domain.PlatformSetting, error) {
	query := `SELECT key, value, description, type, updated_at, updated_by FROM platform_settings WHERE key = $1`

	var s domain.PlatformSetting
	var updatedBy *uuid.UUID
	err := r.pool.QueryRow(ctx, query, key).Scan(
		&s.Key, &s.Value, &s.Description, &s.Type, &s.UpdatedAt, &updatedBy,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("get platform setting: %w", err)
	}
	if updatedBy != nil {
		str := updatedBy.String()
		s.UpdatedBy = &str
	}
	return &s, nil
}

func (r *platformSettingsRepo) GetAll(ctx context.Context) ([]domain.PlatformSetting, error) {
	query := `SELECT key, value, description, type, updated_at, updated_by FROM platform_settings ORDER BY key`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("get all platform settings: %w", err)
	}
	defer rows.Close()

	var settings []domain.PlatformSetting
	for rows.Next() {
		var s domain.PlatformSetting
		var updatedBy *uuid.UUID
		if err := rows.Scan(&s.Key, &s.Value, &s.Description, &s.Type, &s.UpdatedAt, &updatedBy); err != nil {
			return nil, fmt.Errorf("scan platform setting: %w", err)
		}
		if updatedBy != nil {
			str := updatedBy.String()
			s.UpdatedBy = &str
		}
		settings = append(settings, s)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate platform settings rows: %w", err)
	}
	return settings, nil
}

func (r *platformSettingsRepo) Set(ctx context.Context, key, value string, updatedBy *uuid.UUID) error {
	query := `UPDATE platform_settings SET value = $2, updated_at = NOW(), updated_by = $3 WHERE key = $1`

	tag, err := r.pool.Exec(ctx, query, key, value, updatedBy)
	if err != nil {
		return fmt.Errorf("set platform setting: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}
