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

type telegramLinkRepo struct {
	pool *pgxpool.Pool
}

func NewTelegramLinkRepository(pool *pgxpool.Pool) repository.TelegramLinkRepository {
	return &telegramLinkRepo{pool: pool}
}

func (r *telegramLinkRepo) Create(ctx context.Context, link *domain.TelegramLink) error {
	if link.ID == uuid.Nil {
		link.ID = uuid.New()
	}

	query := `INSERT INTO telegram_links (id, user_id, telegram_id, telegram_username, linked_at)
		VALUES ($1, $2, $3, $4, $5)`

	_, err := r.pool.Exec(ctx, query,
		link.ID, link.UserID, link.TelegramID, link.TelegramUsername, link.LinkedAt,
	)
	if err != nil {
		if isDuplicateKeyError(err) {
			return domain.ErrAlreadyExists
		}
		return fmt.Errorf("create telegram link: %w", err)
	}
	return nil
}

func (r *telegramLinkRepo) GetByTelegramID(ctx context.Context, telegramID int64) (*domain.TelegramLink, error) {
	query := `SELECT id, user_id, telegram_id, telegram_username, linked_at
		FROM telegram_links WHERE telegram_id = $1`

	var link domain.TelegramLink
	err := r.pool.QueryRow(ctx, query, telegramID).Scan(
		&link.ID, &link.UserID, &link.TelegramID, &link.TelegramUsername, &link.LinkedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("get telegram link by telegram id: %w", err)
	}
	return &link, nil
}

func (r *telegramLinkRepo) GetByUserID(ctx context.Context, userID uuid.UUID) (*domain.TelegramLink, error) {
	query := `SELECT id, user_id, telegram_id, telegram_username, linked_at
		FROM telegram_links WHERE user_id = $1`

	var link domain.TelegramLink
	err := r.pool.QueryRow(ctx, query, userID).Scan(
		&link.ID, &link.UserID, &link.TelegramID, &link.TelegramUsername, &link.LinkedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("get telegram link by user id: %w", err)
	}
	return &link, nil
}

func (r *telegramLinkRepo) Delete(ctx context.Context, userID uuid.UUID) error {
	result, err := r.pool.Exec(ctx,
		`DELETE FROM telegram_links WHERE user_id = $1`,
		userID,
	)
	if err != nil {
		return fmt.Errorf("delete telegram link: %w", err)
	}
	if result.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}
