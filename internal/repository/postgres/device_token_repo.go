package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/repository"
)

type deviceTokenRepo struct {
	pool *pgxpool.Pool
}

func NewDeviceTokenRepository(pool *pgxpool.Pool) repository.DeviceTokenRepository {
	return &deviceTokenRepo{pool: pool}
}

func (r *deviceTokenRepo) Create(ctx context.Context, token *domain.DeviceToken) error {
	if token.ID == uuid.Nil {
		token.ID = uuid.New()
	}
	now := time.Now()
	if token.CreatedAt.IsZero() {
		token.CreatedAt = now
	}
	if token.UpdatedAt.IsZero() {
		token.UpdatedAt = now
	}

	query := `
		INSERT INTO device_tokens (id, user_id, token, platform, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (token) DO UPDATE SET user_id = $2, platform = $4, updated_at = $6`

	_, err := r.pool.Exec(ctx, query,
		token.ID, token.UserID, token.Token, token.Platform,
		token.CreatedAt, token.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("create device token: %w", err)
	}
	return nil
}

func (r *deviceTokenRepo) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM device_tokens WHERE id = $1`
	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete device token: %w", err)
	}
	if result.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *deviceTokenRepo) DeleteByToken(ctx context.Context, token string) error {
	query := `DELETE FROM device_tokens WHERE token = $1`
	result, err := r.pool.Exec(ctx, query, token)
	if err != nil {
		return fmt.Errorf("delete device token by token: %w", err)
	}
	if result.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *deviceTokenRepo) ListByUser(ctx context.Context, userID uuid.UUID) ([]domain.DeviceToken, error) {
	query := `
		SELECT id, user_id, token, platform, created_at, updated_at
		FROM device_tokens
		WHERE user_id = $1
		ORDER BY created_at DESC`

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("list device tokens: %w", err)
	}
	defer rows.Close()

	var tokens []domain.DeviceToken
	for rows.Next() {
		var dt domain.DeviceToken
		if err := rows.Scan(&dt.ID, &dt.UserID, &dt.Token, &dt.Platform, &dt.CreatedAt, &dt.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan device token: %w", err)
		}
		tokens = append(tokens, dt)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate device token rows: %w", err)
	}

	return tokens, nil
}
