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

type kycRepo struct {
	pool *pgxpool.Pool
}

func NewKYCRepository(pool *pgxpool.Pool) repository.KYCRepository {
	return &kycRepo{pool: pool}
}

var kycColumns = `id, user_id, status, entity_type, full_name, inn, ogrnip, company_name, document_urls, rejection_reason, submitted_at, reviewed_at, reviewed_by, expires_at, created_at, updated_at`

func scanKYC(row pgx.Row) (*domain.KYCApplication, error) {
	var k domain.KYCApplication
	err := row.Scan(
		&k.ID, &k.UserID, &k.Status, &k.EntityType,
		&k.FullName, &k.INN, &k.OGRNIP, &k.CompanyName,
		&k.DocumentURLs, &k.RejectionReason,
		&k.SubmittedAt, &k.ReviewedAt, &k.ReviewedBy, &k.ExpiresAt,
		&k.CreatedAt, &k.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &k, nil
}

func scanKYCs(rows pgx.Rows) ([]domain.KYCApplication, error) {
	var result []domain.KYCApplication
	for rows.Next() {
		k, err := scanKYC(rows)
		if err != nil {
			return nil, fmt.Errorf("scan kyc: %w", err)
		}
		result = append(result, *k)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows err: %w", err)
	}
	return result, nil
}

func (r *kycRepo) Create(ctx context.Context, kyc *domain.KYCApplication) error {
	query := `INSERT INTO kyc_applications (id, user_id, status, entity_type, full_name, inn, ogrnip, company_name, document_urls, rejection_reason, submitted_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`
	_, err := r.pool.Exec(ctx, query,
		kyc.ID, kyc.UserID, kyc.Status, kyc.EntityType,
		kyc.FullName, kyc.INN, kyc.OGRNIP, kyc.CompanyName,
		kyc.DocumentURLs, kyc.RejectionReason,
		kyc.SubmittedAt, kyc.CreatedAt, kyc.UpdatedAt,
	)
	return err
}

func (r *kycRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.KYCApplication, error) {
	query := fmt.Sprintf("SELECT %s FROM kyc_applications WHERE id = $1", kycColumns)
	k, err := scanKYC(r.pool.QueryRow(ctx, query, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrKYCNotFound
		}
		return nil, err
	}
	return k, nil
}

func (r *kycRepo) GetByUserID(ctx context.Context, userID uuid.UUID) (*domain.KYCApplication, error) {
	query := fmt.Sprintf("SELECT %s FROM kyc_applications WHERE user_id = $1 ORDER BY created_at DESC LIMIT 1", kycColumns)
	k, err := scanKYC(r.pool.QueryRow(ctx, query, userID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrKYCNotFound
		}
		return nil, err
	}
	return k, nil
}

func (r *kycRepo) Update(ctx context.Context, kyc *domain.KYCApplication) error {
	query := `UPDATE kyc_applications SET status = $2, entity_type = $3, full_name = $4, inn = $5, ogrnip = $6, company_name = $7, document_urls = $8, rejection_reason = $9, reviewed_at = $10, reviewed_by = $11, expires_at = $12, updated_at = $13 WHERE id = $1`
	result, err := r.pool.Exec(ctx, query,
		kyc.ID, kyc.Status, kyc.EntityType,
		kyc.FullName, kyc.INN, kyc.OGRNIP, kyc.CompanyName,
		kyc.DocumentURLs, kyc.RejectionReason,
		kyc.ReviewedAt, kyc.ReviewedBy, kyc.ExpiresAt, kyc.UpdatedAt,
	)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return domain.ErrKYCNotFound
	}
	return nil
}

func (r *kycRepo) ListPending(ctx context.Context, page, pageSize int) (*domain.PaginatedResult[domain.KYCApplication], error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	var totalCount int64
	err := r.pool.QueryRow(ctx, "SELECT COUNT(*) FROM kyc_applications WHERE status = 'pending'").Scan(&totalCount)
	if err != nil {
		return nil, err
	}

	query := fmt.Sprintf("SELECT %s FROM kyc_applications WHERE status = 'pending' ORDER BY submitted_at ASC LIMIT $1 OFFSET $2", kycColumns)
	rows, err := r.pool.Query(ctx, query, pageSize, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items, err := scanKYCs(rows)
	if err != nil {
		return nil, err
	}

	return &domain.PaginatedResult[domain.KYCApplication]{
		Items:      items,
		TotalCount: totalCount,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: int(math.Ceil(float64(totalCount) / float64(pageSize))),
	}, nil
}

func (r *kycRepo) Approve(ctx context.Context, id uuid.UUID, reviewedBy uuid.UUID, expiresAt time.Time) error {
	now := time.Now()
	query := `UPDATE kyc_applications SET status = 'approved', reviewed_at = $2, reviewed_by = $3, expires_at = $4, updated_at = $5 WHERE id = $1 AND status = 'pending'`
	result, err := r.pool.Exec(ctx, query, id, now, reviewedBy, expiresAt, now)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return domain.ErrKYCNotFound
	}
	return nil
}

func (r *kycRepo) Reject(ctx context.Context, id uuid.UUID, reviewedBy uuid.UUID, reason string) error {
	now := time.Now()
	query := `UPDATE kyc_applications SET status = 'rejected', reviewed_at = $2, reviewed_by = $3, rejection_reason = $4, updated_at = $5 WHERE id = $1 AND status = 'pending'`
	result, err := r.pool.Exec(ctx, query, id, now, reviewedBy, reason, now)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return domain.ErrKYCNotFound
	}
	return nil
}

func (r *kycRepo) ListExpiredApproved(ctx context.Context, before time.Time) ([]domain.KYCApplication, error) {
	query := fmt.Sprintf("SELECT %s FROM kyc_applications WHERE status = 'approved' AND expires_at IS NOT NULL AND expires_at < $1", kycColumns)
	rows, err := r.pool.Query(ctx, query, before)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanKYCs(rows)
}
