package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nikitaaldaev/bani/internal/domain"
)

type BookingShareRepo struct {
	pool *pgxpool.Pool
}

func NewBookingShareRepo(pool *pgxpool.Pool) *BookingShareRepo {
	return &BookingShareRepo{pool: pool}
}

func (r *BookingShareRepo) Create(ctx context.Context, share *domain.BookingShare) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO booking_shares (id, token, created_by, bathhouse_id, start_time, end_time, guest_count, booking_id, created_at, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`, share.ID, share.Token, share.CreatedBy, share.BathhouseID, share.StartTime, share.EndTime, share.GuestCount, share.BookingID, share.CreatedAt, share.ExpiresAt)
	return err
}

func (r *BookingShareRepo) GetByToken(ctx context.Context, token string) (*domain.BookingShare, error) {
	var share domain.BookingShare
	err := r.pool.QueryRow(ctx, `
		SELECT id, token, created_by, bathhouse_id, start_time, end_time, guest_count, booking_id, created_at, expires_at
		FROM booking_shares
		WHERE token = $1
	`, token).Scan(
		&share.ID, &share.Token, &share.CreatedBy, &share.BathhouseID,
		&share.StartTime, &share.EndTime, &share.GuestCount, &share.BookingID,
		&share.CreatedAt, &share.ExpiresAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	return &share, err
}

func (r *BookingShareRepo) DeleteExpired(ctx context.Context) (int64, error) {
	tag, err := r.pool.Exec(ctx, `DELETE FROM booking_shares WHERE expires_at < NOW()`)
	if err != nil {
		return 0, err
	}
	return tag.RowsAffected(), nil
}
