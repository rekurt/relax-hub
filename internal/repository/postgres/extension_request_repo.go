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

type ExtensionRequestRepo struct {
	pool *pgxpool.Pool
}

func NewExtensionRequestRepo(pool *pgxpool.Pool) *ExtensionRequestRepo {
	return &ExtensionRequestRepo{pool: pool}
}

func (r *ExtensionRequestRepo) Create(ctx context.Context, req *domain.BookingExtensionRequest) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO extension_requests (
			id, booking_id, user_id, bathhouse_id, status,
			extra_hours, extension_price, new_end_time, hold_id,
			rejection_reason, created_at, expires_at, resolved_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`, req.ID, req.BookingID, req.UserID, req.BathhouseID, req.Status,
		req.ExtraHours, req.ExtensionPrice, req.NewEndTime, req.HoldID,
		req.RejectionReason, req.CreatedAt, req.ExpiresAt, req.ResolvedAt,
	)
	return err
}

func (r *ExtensionRequestRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.BookingExtensionRequest, error) {
	return r.scanRow(r.pool.QueryRow(ctx, `
		SELECT id, booking_id, user_id, bathhouse_id, status,
			extra_hours, extension_price, new_end_time, hold_id,
			rejection_reason, created_at, expires_at, resolved_at
		FROM extension_requests
		WHERE id = $1
	`, id))
}

func (r *ExtensionRequestRepo) GetPendingByBookingID(ctx context.Context, bookingID uuid.UUID) (*domain.BookingExtensionRequest, error) {
	return r.scanRow(r.pool.QueryRow(ctx, `
		SELECT id, booking_id, user_id, bathhouse_id, status,
			extra_hours, extension_price, new_end_time, hold_id,
			rejection_reason, created_at, expires_at, resolved_at
		FROM extension_requests
		WHERE booking_id = $1 AND status = 'pending'
		ORDER BY created_at DESC
		LIMIT 1
	`, bookingID))
}

func (r *ExtensionRequestRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.ExtensionRequestStatus, reason string) error {
	now := time.Now()
	tag, err := r.pool.Exec(ctx, `
		UPDATE extension_requests
		SET status = $2, rejection_reason = $3, resolved_at = $4
		WHERE id = $1
	`, id, status, reason, now)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrExtensionRequestNotFound
	}
	return nil
}

func (r *ExtensionRequestRepo) ListExpired(ctx context.Context) ([]domain.BookingExtensionRequest, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, booking_id, user_id, bathhouse_id, status,
			extra_hours, extension_price, new_end_time, hold_id,
			rejection_reason, created_at, expires_at, resolved_at
		FROM extension_requests
		WHERE status = 'pending' AND expires_at < NOW()
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []domain.BookingExtensionRequest
	for rows.Next() {
		req, err := r.scanFromRows(rows)
		if err != nil {
			return nil, err
		}
		results = append(results, *req)
	}
	return results, rows.Err()
}

func (r *ExtensionRequestRepo) ListByBookingID(ctx context.Context, bookingID uuid.UUID) ([]domain.BookingExtensionRequest, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, booking_id, user_id, bathhouse_id, status,
			extra_hours, extension_price, new_end_time, hold_id,
			rejection_reason, created_at, expires_at, resolved_at
		FROM extension_requests
		WHERE booking_id = $1
		ORDER BY created_at DESC
	`, bookingID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []domain.BookingExtensionRequest
	for rows.Next() {
		req, err := r.scanFromRows(rows)
		if err != nil {
			return nil, err
		}
		results = append(results, *req)
	}
	return results, rows.Err()
}

func (r *ExtensionRequestRepo) scanRow(row pgx.Row) (*domain.BookingExtensionRequest, error) {
	var req domain.BookingExtensionRequest
	err := row.Scan(
		&req.ID, &req.BookingID, &req.UserID, &req.BathhouseID, &req.Status,
		&req.ExtraHours, &req.ExtensionPrice, &req.NewEndTime, &req.HoldID,
		&req.RejectionReason, &req.CreatedAt, &req.ExpiresAt, &req.ResolvedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrExtensionRequestNotFound
	}
	return &req, err
}

func (r *ExtensionRequestRepo) scanFromRows(rows pgx.Rows) (*domain.BookingExtensionRequest, error) {
	var req domain.BookingExtensionRequest
	err := rows.Scan(
		&req.ID, &req.BookingID, &req.UserID, &req.BathhouseID, &req.Status,
		&req.ExtraHours, &req.ExtensionPrice, &req.NewEndTime, &req.HoldID,
		&req.RejectionReason, &req.CreatedAt, &req.ExpiresAt, &req.ResolvedAt,
	)
	return &req, err
}
