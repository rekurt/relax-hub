package postgres

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rekurt/relax-hub/internal/domain"
	"github.com/rekurt/relax-hub/internal/repository"
)

type paymentRepo struct {
	pool *pgxpool.Pool
}

func NewPaymentRepository(pool *pgxpool.Pool) repository.PaymentRepository {
	return &paymentRepo{pool: pool}
}

func (r *paymentRepo) Create(ctx context.Context, payment *domain.Payment) error {
	query := `
		INSERT INTO payments (id, booking_id, user_id, amount, currency, status, provider, external_id, payment_method, wallet_amount, card_amount, is_hold, captured_at, refund_amount, refunded_at, metadata, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18)`

	if payment.ID == uuid.Nil {
		payment.ID = uuid.New()
	}
	if payment.CreatedAt.IsZero() {
		payment.CreatedAt = time.Now()
	}
	if payment.UpdatedAt.IsZero() {
		payment.UpdatedAt = time.Now()
	}
	if payment.PaymentMethod == "" {
		payment.PaymentMethod = domain.PaymentMethodCard
	}

	metadata := payment.Metadata
	if metadata == nil {
		metadata = make(map[string]string)
	}

	_, err := r.pool.Exec(ctx, query,
		payment.ID, payment.BookingID, payment.UserID,
		payment.Amount, payment.Currency, payment.Status,
		payment.Provider, payment.ExternalID, payment.PaymentMethod,
		payment.WalletAmount, payment.CardAmount,
		payment.IsHold, payment.CapturedAt,
		payment.RefundAmount, payment.RefundedAt, metadata,
		payment.CreatedAt, payment.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("create payment: %w", err)
	}
	return nil
}

const paymentSelectColumns = `id, booking_id, user_id, amount, currency, status, provider, external_id, payment_method, wallet_amount, card_amount, is_hold, captured_at, refund_amount, refunded_at, metadata, created_at, updated_at`

func (r *paymentRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Payment, error) {
	query := `SELECT ` + paymentSelectColumns + ` FROM payments WHERE id = $1`
	return r.scanPayment(ctx, query, id)
}

func (r *paymentRepo) GetByBookingID(ctx context.Context, bookingID uuid.UUID) (*domain.Payment, error) {
	query := `SELECT ` + paymentSelectColumns + ` FROM payments WHERE booking_id = $1`
	return r.scanPayment(ctx, query, bookingID)
}

func (r *paymentRepo) GetByExternalID(ctx context.Context, externalID string) (*domain.Payment, error) {
	query := `SELECT ` + paymentSelectColumns + ` FROM payments WHERE external_id = $1`
	return r.scanPayment(ctx, query, externalID)
}

func (r *paymentRepo) scanPayment(ctx context.Context, query string, arg interface{}) (*domain.Payment, error) {
	var p domain.Payment
	err := r.pool.QueryRow(ctx, query, arg).Scan(
		&p.ID, &p.BookingID, &p.UserID,
		&p.Amount, &p.Currency, &p.Status,
		&p.Provider, &p.ExternalID, &p.PaymentMethod,
		&p.WalletAmount, &p.CardAmount,
		&p.IsHold, &p.CapturedAt,
		&p.RefundAmount, &p.RefundedAt, &p.Metadata,
		&p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrPaymentNotFound
		}
		return nil, fmt.Errorf("scan payment: %w", err)
	}
	return &p, nil
}

func (r *paymentRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.PaymentStatus, externalID string) error {
	query := `UPDATE payments SET status = $2, external_id = $3, updated_at = $4 WHERE id = $1`

	tag, err := r.pool.Exec(ctx, query, id, status, externalID, time.Now())
	if err != nil {
		return fmt.Errorf("update payment status: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrPaymentNotFound
	}
	return nil
}

func (r *paymentRepo) UpdateRefund(ctx context.Context, id uuid.UUID, refundAmount int64, refundedAt time.Time, status domain.PaymentStatus) error {
	query := `UPDATE payments SET refund_amount = $2, refunded_at = $3, status = $4, updated_at = $5 WHERE id = $1`

	tag, err := r.pool.Exec(ctx, query, id, refundAmount, refundedAt, status, time.Now())
	if err != nil {
		return fmt.Errorf("update payment refund: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrPaymentNotFound
	}
	return nil
}

func (r *paymentRepo) UpdateCapture(ctx context.Context, id uuid.UUID, capturedAt time.Time, status domain.PaymentStatus) error {
	query := `UPDATE payments SET captured_at = $2, is_hold = false, status = $3, updated_at = $4 WHERE id = $1`

	tag, err := r.pool.Exec(ctx, query, id, capturedAt, status, time.Now())
	if err != nil {
		return fmt.Errorf("update payment capture: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrPaymentNotFound
	}
	return nil
}

func (r *paymentRepo) Delete(ctx context.Context, id uuid.UUID) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM payments WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete payment: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrPaymentNotFound
	}
	return nil
}

func (r *paymentRepo) ListByUser(ctx context.Context, userID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.Payment], error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}

	var totalCount int64
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM payments WHERE user_id = $1`, userID).Scan(&totalCount)
	if err != nil {
		return nil, fmt.Errorf("count payments by user: %w", err)
	}

	offset := (page - 1) * pageSize
	query := `SELECT ` + paymentSelectColumns + ` FROM payments WHERE user_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`

	rows, err := r.pool.Query(ctx, query, userID, pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("list payments by user: %w", err)
	}
	defer rows.Close()

	var payments []domain.Payment
	for rows.Next() {
		var p domain.Payment
		if err := rows.Scan(
			&p.ID, &p.BookingID, &p.UserID,
			&p.Amount, &p.Currency, &p.Status,
			&p.Provider, &p.ExternalID, &p.PaymentMethod,
			&p.WalletAmount, &p.CardAmount,
			&p.IsHold, &p.CapturedAt,
			&p.RefundAmount, &p.RefundedAt, &p.Metadata,
			&p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan payment: %w", err)
		}
		payments = append(payments, p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate payment rows: %w", err)
	}

	return &domain.PaginatedResult[domain.Payment]{
		Items:      payments,
		TotalCount: totalCount,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: int(math.Ceil(float64(totalCount) / float64(pageSize))),
	}, nil
}
