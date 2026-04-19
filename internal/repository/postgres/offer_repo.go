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

type offerRepo struct {
	pool *pgxpool.Pool
}

func NewOfferRepository(pool *pgxpool.Pool) repository.OfferRepository {
	return &offerRepo{pool: pool}
}

var offerColumns = `id, user_id, offer_version, accepted_at, ip_address, user_agent, created_at`

func scanOffer(row pgx.Row) (*domain.OfferAcceptance, error) {
	var o domain.OfferAcceptance
	err := row.Scan(&o.ID, &o.UserID, &o.OfferVersion, &o.AcceptedAt, &o.IPAddress, &o.UserAgent, &o.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &o, nil
}

func (r *offerRepo) Create(ctx context.Context, acceptance *domain.OfferAcceptance) error {
	query := `INSERT INTO offer_acceptances (id, user_id, offer_version, accepted_at, ip_address, user_agent, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`
	_, err := r.pool.Exec(ctx, query,
		acceptance.ID, acceptance.UserID, acceptance.OfferVersion,
		acceptance.AcceptedAt, acceptance.IPAddress, acceptance.UserAgent, acceptance.CreatedAt,
	)
	return err
}

func (r *offerRepo) GetByUserID(ctx context.Context, userID uuid.UUID) (*domain.OfferAcceptance, error) {
	query := fmt.Sprintf("SELECT %s FROM offer_acceptances WHERE user_id = $1 ORDER BY accepted_at DESC LIMIT 1", offerColumns)
	o, err := scanOffer(r.pool.QueryRow(ctx, query, userID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrOfferNotFound
		}
		return nil, err
	}
	return o, nil
}

func (r *offerRepo) GetByUserAndVersion(ctx context.Context, userID uuid.UUID, version string) (*domain.OfferAcceptance, error) {
	query := fmt.Sprintf("SELECT %s FROM offer_acceptances WHERE user_id = $1 AND offer_version = $2", offerColumns)
	o, err := scanOffer(r.pool.QueryRow(ctx, query, userID, version))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrOfferNotFound
		}
		return nil, err
	}
	return o, nil
}

func (r *offerRepo) ListByUser(ctx context.Context, userID uuid.UUID) ([]domain.OfferAcceptance, error) {
	query := fmt.Sprintf("SELECT %s FROM offer_acceptances WHERE user_id = $1 ORDER BY accepted_at DESC", offerColumns)
	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []domain.OfferAcceptance
	for rows.Next() {
		o, err := scanOffer(rows)
		if err != nil {
			return nil, fmt.Errorf("scan offer: %w", err)
		}
		result = append(result, *o)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows err: %w", err)
	}
	return result, nil
}
