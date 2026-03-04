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

type socialAccountRepo struct {
	pool *pgxpool.Pool
}

func NewSocialAccountRepository(pool *pgxpool.Pool) repository.SocialAccountRepository {
	return &socialAccountRepo{pool: pool}
}

func (r *socialAccountRepo) Create(ctx context.Context, account *domain.SocialAccount) error {
	if account.ID == uuid.Nil {
		account.ID = uuid.New()
	}

	query := `INSERT INTO social_accounts (id, user_id, provider, provider_id, email, name, avatar_url, access_token, linked_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`

	_, err := r.pool.Exec(ctx, query,
		account.ID, account.UserID, account.Provider, account.ProviderID,
		account.Email, account.Name, account.AvatarURL, account.AccessToken, account.LinkedAt,
	)
	if err != nil {
		if isDuplicateKeyError(err) {
			return domain.ErrSocialAccountAlreadyLinked
		}
		return fmt.Errorf("create social account: %w", err)
	}
	return nil
}

func (r *socialAccountRepo) GetByProviderAndID(ctx context.Context, provider domain.OAuthProvider, providerID string) (*domain.SocialAccount, error) {
	query := `SELECT id, user_id, provider, provider_id, email, name, avatar_url, access_token, linked_at
		FROM social_accounts WHERE provider = $1 AND provider_id = $2`

	var a domain.SocialAccount
	err := r.pool.QueryRow(ctx, query, provider, providerID).Scan(
		&a.ID, &a.UserID, &a.Provider, &a.ProviderID,
		&a.Email, &a.Name, &a.AvatarURL, &a.AccessToken, &a.LinkedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrSocialAccountNotFound
		}
		return nil, fmt.Errorf("get social account by provider: %w", err)
	}
	return &a, nil
}

func (r *socialAccountRepo) ListByUser(ctx context.Context, userID uuid.UUID) ([]domain.SocialAccount, error) {
	query := `SELECT id, user_id, provider, provider_id, email, name, avatar_url, linked_at
		FROM social_accounts WHERE user_id = $1 ORDER BY linked_at DESC`

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("list social accounts: %w", err)
	}
	defer rows.Close()

	var accounts []domain.SocialAccount
	for rows.Next() {
		var a domain.SocialAccount
		if err := rows.Scan(&a.ID, &a.UserID, &a.Provider, &a.ProviderID,
			&a.Email, &a.Name, &a.AvatarURL, &a.LinkedAt); err != nil {
			return nil, fmt.Errorf("scan social account: %w", err)
		}
		accounts = append(accounts, a)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate social account rows: %w", err)
	}

	return accounts, nil
}

func (r *socialAccountRepo) Delete(ctx context.Context, userID uuid.UUID, provider domain.OAuthProvider) error {
	result, err := r.pool.Exec(ctx,
		`DELETE FROM social_accounts WHERE user_id = $1 AND provider = $2`,
		userID, provider,
	)
	if err != nil {
		return fmt.Errorf("delete social account: %w", err)
	}
	if result.RowsAffected() == 0 {
		return domain.ErrSocialAccountNotFound
	}
	return nil
}
