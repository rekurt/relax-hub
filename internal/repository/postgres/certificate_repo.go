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
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/repository"
)

type certificateRepo struct {
	pool *pgxpool.Pool
}

func NewGiftCertificateRepository(pool *pgxpool.Pool) repository.GiftCertificateRepository {
	return &certificateRepo{pool: pool}
}

func (r *certificateRepo) Create(ctx context.Context, cert *domain.GiftCertificate) error {
	if cert.ID == uuid.Nil {
		cert.ID = uuid.New()
	}
	if cert.Status == "" {
		cert.Status = domain.CertificateStatusActive
	}
	if cert.CreatedAt.IsZero() {
		cert.CreatedAt = time.Now()
	}

	query := `
		INSERT INTO gift_certificates (id, code, purchaser_id, purchaser_email, recipient_email, recipient_name, amount, balance, message, status, valid_until, redeemed_by_id, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`

	_, err := r.pool.Exec(ctx, query,
		cert.ID, cert.Code, cert.PurchaserID, cert.PurchaserEmail,
		cert.RecipientEmail, cert.RecipientName, cert.Amount, cert.Balance,
		cert.Message, string(cert.Status), cert.ValidUntil, cert.RedeemedByID, cert.CreatedAt,
	)
	if err != nil {
		if isDuplicateKeyError(err) {
			return domain.ErrAlreadyExists
		}
		return fmt.Errorf("create gift certificate: %w", err)
	}
	return nil
}

func (r *certificateRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.GiftCertificate, error) {
	query := `
		SELECT id, code, purchaser_id, purchaser_email, recipient_email, recipient_name, amount, balance, message, status, valid_until, redeemed_by_id, created_at
		FROM gift_certificates WHERE id = $1`

	return r.scanCertificate(ctx, query, id)
}

func (r *certificateRepo) GetByCode(ctx context.Context, code string) (*domain.GiftCertificate, error) {
	query := `
		SELECT id, code, purchaser_id, purchaser_email, recipient_email, recipient_name, amount, balance, message, status, valid_until, redeemed_by_id, created_at
		FROM gift_certificates WHERE code = $1`

	return r.scanCertificate(ctx, query, code)
}

func (r *certificateRepo) scanCertificate(ctx context.Context, query string, arg any) (*domain.GiftCertificate, error) {
	var cert domain.GiftCertificate
	err := r.pool.QueryRow(ctx, query, arg).Scan(
		&cert.ID, &cert.Code, &cert.PurchaserID, &cert.PurchaserEmail,
		&cert.RecipientEmail, &cert.RecipientName, &cert.Amount, &cert.Balance,
		&cert.Message, &cert.Status, &cert.ValidUntil, &cert.RedeemedByID, &cert.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrCertificateNotFound
		}
		return nil, fmt.Errorf("get gift certificate: %w", err)
	}
	return &cert, nil
}

func (r *certificateRepo) ApplyToBooking(ctx context.Context, id uuid.UUID, usage *domain.CertificateUsage) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck // rollback after commit is a no-op

	updateQuery := `
		UPDATE gift_certificates
		SET balance = balance - $2,
			status = CASE WHEN balance - $2 = 0 THEN 'used' ELSE status END
		WHERE id = $1 AND balance >= $2 AND status = 'active' AND valid_until > NOW()`

	result, err := tx.Exec(ctx, updateQuery, id, usage.Amount)
	if err != nil {
		return fmt.Errorf("update certificate balance: %w", err)
	}
	if result.RowsAffected() == 0 {
		return domain.ErrCertificateInsufficientBalance
	}

	if usage.ID == uuid.Nil {
		usage.ID = uuid.New()
	}
	if usage.UsedAt.IsZero() {
		usage.UsedAt = time.Now()
	}

	insertQuery := `
		INSERT INTO certificate_usages (id, certificate_id, booking_id, amount, used_at)
		VALUES ($1, $2, $3, $4, $5)`

	_, err = tx.Exec(ctx, insertQuery, usage.ID, usage.CertificateID, usage.BookingID, usage.Amount, usage.UsedAt)
	if err != nil {
		return fmt.Errorf("create certificate usage: %w", err)
	}

	return tx.Commit(ctx)
}

func (r *certificateRepo) ListByUser(ctx context.Context, userID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.GiftCertificate], error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}

	var totalCount int64
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM gift_certificates WHERE redeemed_by_id = $1 OR purchaser_id = $1`, userID,
	).Scan(&totalCount)
	if err != nil {
		return nil, fmt.Errorf("count certificates: %w", err)
	}

	offset := (page - 1) * pageSize
	query := `
		SELECT id, code, purchaser_id, purchaser_email, recipient_email, recipient_name, amount, balance, message, status, valid_until, redeemed_by_id, created_at
		FROM gift_certificates
		WHERE redeemed_by_id = $1 OR purchaser_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := r.pool.Query(ctx, query, userID, pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("list certificates: %w", err)
	}
	defer rows.Close()

	certs := make([]domain.GiftCertificate, 0)
	for rows.Next() {
		var cert domain.GiftCertificate
		if err := rows.Scan(
			&cert.ID, &cert.Code, &cert.PurchaserID, &cert.PurchaserEmail,
			&cert.RecipientEmail, &cert.RecipientName, &cert.Amount, &cert.Balance,
			&cert.Message, &cert.Status, &cert.ValidUntil, &cert.RedeemedByID, &cert.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan certificate: %w", err)
		}
		certs = append(certs, cert)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate certificate rows: %w", err)
	}

	return &domain.PaginatedResult[domain.GiftCertificate]{
		Items:      certs,
		TotalCount: totalCount,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: int(math.Ceil(float64(totalCount) / float64(pageSize))),
	}, nil
}

func (r *certificateRepo) Redeem(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	query := `
		UPDATE gift_certificates
		SET redeemed_by_id = $2
		WHERE id = $1 AND redeemed_by_id IS NULL AND status = 'active' AND valid_until > NOW()`

	result, err := r.pool.Exec(ctx, query, id, userID)
	if err != nil {
		return fmt.Errorf("redeem certificate: %w", err)
	}
	if result.RowsAffected() == 0 {
		return domain.ErrCertificateNotFound
	}
	return nil
}

