package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rekurt/relax-hub/internal/domain"
	"github.com/rekurt/relax-hub/internal/repository"
)

type certificateOrderRepo struct {
	pool *pgxpool.Pool
}

func NewCertificateOrderRepository(pool *pgxpool.Pool) repository.CertificateOrderRepository {
	return &certificateOrderRepo{pool: pool}
}

const certificateOrderSelectColumns = `
	id, purchaser_id, purchaser_email, recipient_email, recipient_name, message,
	amount, status, payment_method, provider, external_id, certificate_id, paid_at,
	created_at, updated_at
`

func (r *certificateOrderRepo) Create(ctx context.Context, order *domain.CertificateOrder) error {
	if order.ID == uuid.Nil {
		order.ID = uuid.New()
	}
	if order.Status == "" {
		order.Status = domain.CertificateOrderStatusDraft
	}
	if order.CreatedAt.IsZero() {
		order.CreatedAt = time.Now()
	}
	if order.UpdatedAt.IsZero() {
		order.UpdatedAt = order.CreatedAt
	}

	query := `
		INSERT INTO certificate_orders (
			id, purchaser_id, purchaser_email, recipient_email, recipient_name, message,
			amount, status, payment_method, provider, external_id, certificate_id, paid_at,
			created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
	`

	_, err := r.pool.Exec(ctx, query,
		order.ID, order.PurchaserID, order.PurchaserEmail, order.RecipientEmail, order.RecipientName, order.Message,
		order.Amount, string(order.Status), nullablePaymentMethod(order.PaymentMethod), nullableString(order.Provider), nullableString(order.ExternalID), order.CertificateID, order.PaidAt,
		order.CreatedAt, order.UpdatedAt,
	)
	if err != nil {
		if isDuplicateKeyError(err) {
			return domain.ErrAlreadyExists
		}
		return fmt.Errorf("create certificate order: %w", err)
	}
	return nil
}

func (r *certificateOrderRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.CertificateOrder, error) {
	query := `SELECT ` + certificateOrderSelectColumns + ` FROM certificate_orders WHERE id = $1`
	return r.scanOrder(ctx, query, id)
}

func (r *certificateOrderRepo) GetByExternalID(ctx context.Context, externalID string) (*domain.CertificateOrder, error) {
	query := `SELECT ` + certificateOrderSelectColumns + ` FROM certificate_orders WHERE external_id = $1`
	return r.scanOrder(ctx, query, externalID)
}

func (r *certificateOrderRepo) UpdatePayment(ctx context.Context, id uuid.UUID, status domain.CertificateOrderStatus, paymentMethod domain.PaymentMethod, provider, externalID string) error {
	query := `
		UPDATE certificate_orders
		SET status = $2, payment_method = $3, provider = $4, external_id = $5, updated_at = $6
		WHERE id = $1
	`
	tag, err := r.pool.Exec(ctx, query, id, status, nullablePaymentMethod(paymentMethod), nullableString(provider), nullableString(externalID), time.Now())
	if err != nil {
		return fmt.Errorf("update certificate order payment: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrCertificateOrderNotFound
	}
	return nil
}

func (r *certificateOrderRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.CertificateOrderStatus) error {
	query := `UPDATE certificate_orders SET status = $2, updated_at = $3 WHERE id = $1`
	tag, err := r.pool.Exec(ctx, query, id, status, time.Now())
	if err != nil {
		return fmt.Errorf("update certificate order status: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrCertificateOrderNotFound
	}
	return nil
}

func (r *certificateOrderRepo) MarkPaid(ctx context.Context, id uuid.UUID, certificateID uuid.UUID, paidAt time.Time) error {
	query := `
		UPDATE certificate_orders
		SET status = $2, certificate_id = $3, paid_at = $4, updated_at = $4
		WHERE id = $1
	`
	tag, err := r.pool.Exec(ctx, query, id, domain.CertificateOrderStatusPaid, certificateID, paidAt)
	if err != nil {
		return fmt.Errorf("mark certificate order paid: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrCertificateOrderNotFound
	}
	return nil
}

func (r *certificateOrderRepo) scanOrder(ctx context.Context, query string, arg any) (*domain.CertificateOrder, error) {
	var order domain.CertificateOrder
	var paymentMethod *string
	var provider *string
	var externalID *string
	err := r.pool.QueryRow(ctx, query, arg).Scan(
		&order.ID, &order.PurchaserID, &order.PurchaserEmail, &order.RecipientEmail, &order.RecipientName, &order.Message,
		&order.Amount, &order.Status, &paymentMethod, &provider, &externalID, &order.CertificateID, &order.PaidAt,
		&order.CreatedAt, &order.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrCertificateOrderNotFound
		}
		return nil, fmt.Errorf("scan certificate order: %w", err)
	}
	if paymentMethod != nil {
		order.PaymentMethod = domain.PaymentMethod(*paymentMethod)
	}
	if provider != nil {
		order.Provider = *provider
	}
	if externalID != nil {
		order.ExternalID = *externalID
	}
	return &order, nil
}

func nullableString(v string) any {
	if v == "" {
		return nil
	}
	return v
}

func nullablePaymentMethod(v domain.PaymentMethod) any {
	if v == "" {
		return nil
	}
	return string(v)
}
