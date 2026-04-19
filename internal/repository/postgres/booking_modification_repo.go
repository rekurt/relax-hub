package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rekurt/relax-hub/internal/domain"
)

type BookingModificationRequestRepo struct {
	pool *pgxpool.Pool
}

func NewBookingModificationRequestRepo(pool *pgxpool.Pool) *BookingModificationRequestRepo {
	return &BookingModificationRequestRepo{pool: pool}
}

func (r *BookingModificationRequestRepo) Create(ctx context.Context, req *domain.BookingModificationRequest) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO booking_modification_requests (
			id, booking_id, user_id, bathhouse_id, status,
			old_start_time, old_end_time, old_guest_count, old_total_price,
			proposed_start_time, proposed_end_time, proposed_guest_count, proposed_total_price,
			rejection_reason, created_at, expires_at, resolved_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
	`, req.ID, req.BookingID, req.UserID, req.BathhouseID, req.Status,
		req.OldStartTime, req.OldEndTime, req.OldGuestCount, req.OldTotalPrice,
		req.ProposedStartTime, req.ProposedEndTime, req.ProposedGuestCount, req.ProposedTotalPrice,
		req.RejectionReason, req.CreatedAt, req.ExpiresAt, req.ResolvedAt,
	)
	return err
}

func (r *BookingModificationRequestRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.BookingModificationRequest, error) {
	return r.scanRow(r.pool.QueryRow(ctx, `
		SELECT id, booking_id, user_id, bathhouse_id, status,
			old_start_time, old_end_time, old_guest_count, old_total_price,
			proposed_start_time, proposed_end_time, proposed_guest_count, proposed_total_price,
			rejection_reason, created_at, expires_at, resolved_at
		FROM booking_modification_requests
		WHERE id = $1
	`, id))
}

func (r *BookingModificationRequestRepo) GetPendingByBookingID(ctx context.Context, bookingID uuid.UUID) (*domain.BookingModificationRequest, error) {
	return r.scanRow(r.pool.QueryRow(ctx, `
		SELECT id, booking_id, user_id, bathhouse_id, status,
			old_start_time, old_end_time, old_guest_count, old_total_price,
			proposed_start_time, proposed_end_time, proposed_guest_count, proposed_total_price,
			rejection_reason, created_at, expires_at, resolved_at
		FROM booking_modification_requests
		WHERE booking_id = $1 AND status = 'pending'
		ORDER BY created_at DESC
		LIMIT 1
	`, bookingID))
}

func (r *BookingModificationRequestRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.ModificationRequestStatus, reason string) error {
	now := time.Now()
	tag, err := r.pool.Exec(ctx, `
		UPDATE booking_modification_requests
		SET status = $2, rejection_reason = $3, resolved_at = $4
		WHERE id = $1
	`, id, status, reason, now)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrModificationRequestNotFound
	}
	return nil
}

func (r *BookingModificationRequestRepo) ListExpired(ctx context.Context) ([]domain.BookingModificationRequest, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, booking_id, user_id, bathhouse_id, status,
			old_start_time, old_end_time, old_guest_count, old_total_price,
			proposed_start_time, proposed_end_time, proposed_guest_count, proposed_total_price,
			rejection_reason, created_at, expires_at, resolved_at
		FROM booking_modification_requests
		WHERE status = 'pending' AND expires_at < NOW()
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []domain.BookingModificationRequest
	for rows.Next() {
		req, err := r.scanFromRows(rows)
		if err != nil {
			return nil, err
		}
		results = append(results, *req)
	}
	return results, rows.Err()
}

func (r *BookingModificationRequestRepo) ListByBookingID(ctx context.Context, bookingID uuid.UUID) ([]domain.BookingModificationRequest, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, booking_id, user_id, bathhouse_id, status,
			old_start_time, old_end_time, old_guest_count, old_total_price,
			proposed_start_time, proposed_end_time, proposed_guest_count, proposed_total_price,
			rejection_reason, created_at, expires_at, resolved_at
		FROM booking_modification_requests
		WHERE booking_id = $1
		ORDER BY created_at DESC
	`, bookingID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []domain.BookingModificationRequest
	for rows.Next() {
		req, err := r.scanFromRows(rows)
		if err != nil {
			return nil, err
		}
		results = append(results, *req)
	}
	return results, rows.Err()
}

func (r *BookingModificationRequestRepo) scanRow(row pgx.Row) (*domain.BookingModificationRequest, error) {
	var req domain.BookingModificationRequest
	err := row.Scan(
		&req.ID, &req.BookingID, &req.UserID, &req.BathhouseID, &req.Status,
		&req.OldStartTime, &req.OldEndTime, &req.OldGuestCount, &req.OldTotalPrice,
		&req.ProposedStartTime, &req.ProposedEndTime, &req.ProposedGuestCount, &req.ProposedTotalPrice,
		&req.RejectionReason, &req.CreatedAt, &req.ExpiresAt, &req.ResolvedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrModificationRequestNotFound
	}
	return &req, err
}

func (r *BookingModificationRequestRepo) scanFromRows(rows pgx.Rows) (*domain.BookingModificationRequest, error) {
	var req domain.BookingModificationRequest
	err := rows.Scan(
		&req.ID, &req.BookingID, &req.UserID, &req.BathhouseID, &req.Status,
		&req.OldStartTime, &req.OldEndTime, &req.OldGuestCount, &req.OldTotalPrice,
		&req.ProposedStartTime, &req.ProposedEndTime, &req.ProposedGuestCount, &req.ProposedTotalPrice,
		&req.RejectionReason, &req.CreatedAt, &req.ExpiresAt, &req.ResolvedAt,
	)
	return &req, err
}
