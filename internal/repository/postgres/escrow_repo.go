package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rekurt/relax-hub/internal/domain"
	"github.com/rekurt/relax-hub/internal/repository"
)

type escrowRepo struct {
	pool *pgxpool.Pool
}

func NewEscrowRepository(pool *pgxpool.Pool) repository.EscrowRepository {
	return &escrowRepo{pool: pool}
}

var escrowColumns = `id, booking_id, amount, service_fee, status, claim_period_ends_at, released_at, created_at`

func scanEscrow(row pgx.Row) (*domain.Escrow, error) {
	var e domain.Escrow
	err := row.Scan(&e.ID, &e.BookingID, &e.Amount, &e.ServiceFee, &e.Status, &e.ClaimPeriodEndsAt, &e.ReleasedAt, &e.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrEscrowNotFound
		}
		return nil, err
	}
	return &e, nil
}

func (r *escrowRepo) Create(ctx context.Context, escrow *domain.Escrow) error {
	query := `INSERT INTO escrows (id, booking_id, amount, service_fee, status, claim_period_ends_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`
	_, err := r.pool.Exec(ctx, query,
		escrow.ID, escrow.BookingID, escrow.Amount, escrow.ServiceFee,
		escrow.Status, escrow.ClaimPeriodEndsAt, escrow.CreatedAt)
	return err
}

func (r *escrowRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Escrow, error) {
	query := `SELECT ` + escrowColumns + ` FROM escrows WHERE id = $1`
	return scanEscrow(r.pool.QueryRow(ctx, query, id))
}

func (r *escrowRepo) GetByBookingID(ctx context.Context, bookingID uuid.UUID) (*domain.Escrow, error) {
	query := `SELECT ` + escrowColumns + ` FROM escrows WHERE booking_id = $1`
	return scanEscrow(r.pool.QueryRow(ctx, query, bookingID))
}

func (r *escrowRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.EscrowStatus, releasedAt *time.Time) error {
	query := `UPDATE escrows SET status = $2, released_at = $3 WHERE id = $1`
	tag, err := r.pool.Exec(ctx, query, id, status, releasedAt)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrEscrowNotFound
	}
	return nil
}

func (r *escrowRepo) ListMatured(ctx context.Context) ([]domain.Escrow, error) {
	query := `SELECT ` + escrowColumns + ` FROM escrows WHERE status = 'held' AND claim_period_ends_at < now() ORDER BY claim_period_ends_at`
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var escrows []domain.Escrow
	for rows.Next() {
		var e domain.Escrow
		if err := rows.Scan(&e.ID, &e.BookingID, &e.Amount, &e.ServiceFee, &e.Status, &e.ClaimPeriodEndsAt, &e.ReleasedAt, &e.CreatedAt); err != nil {
			return nil, err
		}
		escrows = append(escrows, e)
	}
	return escrows, rows.Err()
}
