package postgres

import (
	"context"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rekurt/relax-hub/internal/domain"
	"github.com/rekurt/relax-hub/internal/repository"
)

type fraudFlagRepo struct {
	pool *pgxpool.Pool
}

func NewFraudFlagRepository(pool *pgxpool.Pool) repository.FraudFlagRepository {
	return &fraudFlagRepo{pool: pool}
}

var fraudFlagColumns = `id, user_id, rule, severity, status, action, details, created_at, reviewed_at, reviewed_by`

func scanFraudFlag(row pgx.Row) (*domain.FraudFlag, error) {
	var f domain.FraudFlag
	err := row.Scan(
		&f.ID, &f.UserID, &f.Rule, &f.Severity, &f.Status, &f.Action,
		&f.Details, &f.CreatedAt, &f.ReviewedAt, &f.ReviewedBy,
	)
	if err != nil {
		return nil, err
	}
	return &f, nil
}

func scanFraudFlags(rows pgx.Rows) ([]domain.FraudFlag, error) {
	var flags []domain.FraudFlag
	for rows.Next() {
		f, err := scanFraudFlag(rows)
		if err != nil {
			return nil, fmt.Errorf("scan fraud flag: %w", err)
		}
		flags = append(flags, *f)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate fraud flag rows: %w", err)
	}
	return flags, nil
}

func (r *fraudFlagRepo) Create(ctx context.Context, flag *domain.FraudFlag) error {
	query := `
		INSERT INTO fraud_flags (id, user_id, rule, severity, status, action, details, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	if flag.ID == uuid.Nil {
		flag.ID = uuid.New()
	}
	if flag.Status == "" {
		flag.Status = domain.FraudFlagStatusPending
	}
	if flag.CreatedAt.IsZero() {
		flag.CreatedAt = time.Now()
	}

	_, err := r.pool.Exec(ctx, query,
		flag.ID, flag.UserID, string(flag.Rule), string(flag.Severity),
		string(flag.Status), string(flag.Action), flag.Details, flag.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("create fraud flag: %w", err)
	}
	return nil
}

func (r *fraudFlagRepo) ListByUser(ctx context.Context, userID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.FraudFlag], error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}

	var totalCount int64
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM fraud_flags WHERE user_id = $1`, userID).Scan(&totalCount)
	if err != nil {
		return nil, fmt.Errorf("count fraud flags by user: %w", err)
	}

	offset := (page - 1) * pageSize
	query := fmt.Sprintf(`SELECT %s FROM fraud_flags WHERE user_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`, fraudFlagColumns)
	rows, err := r.pool.Query(ctx, query, userID, pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("list fraud flags by user: %w", err)
	}
	defer rows.Close()

	flags, err := scanFraudFlags(rows)
	if err != nil {
		return nil, err
	}

	return &domain.PaginatedResult[domain.FraudFlag]{
		Items:      flags,
		TotalCount: totalCount,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: int(math.Ceil(float64(totalCount) / float64(pageSize))),
	}, nil
}

func (r *fraudFlagRepo) ListPending(ctx context.Context, filter domain.FraudFlagFilter) (*domain.PaginatedResult[domain.FraudFlag], error) {
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PageSize < 1 {
		filter.PageSize = 20
	}

	var conditions []string
	var args []interface{}
	argIdx := 1

	if filter.Status != nil {
		conditions = append(conditions, fmt.Sprintf("status = $%d", argIdx))
		args = append(args, string(*filter.Status))
		argIdx++
	} else {
		conditions = append(conditions, fmt.Sprintf("status = $%d", argIdx))
		args = append(args, string(domain.FraudFlagStatusPending))
		argIdx++
	}
	if filter.UserID != nil {
		conditions = append(conditions, fmt.Sprintf("user_id = $%d", argIdx))
		args = append(args, *filter.UserID)
		argIdx++
	}
	if filter.Rule != nil {
		conditions = append(conditions, fmt.Sprintf("rule = $%d", argIdx))
		args = append(args, string(*filter.Rule))
		argIdx++
	}

	where := ""
	if len(conditions) > 0 {
		where = " WHERE " + strings.Join(conditions, " AND ")
	}

	var totalCount int64
	countQuery := `SELECT COUNT(*) FROM fraud_flags` + where
	err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&totalCount)
	if err != nil {
		return nil, fmt.Errorf("count fraud flags: %w", err)
	}

	offset := (filter.Page - 1) * filter.PageSize
	args = append(args, filter.PageSize, offset)
	query := fmt.Sprintf(`SELECT %s FROM fraud_flags%s ORDER BY created_at DESC LIMIT $%d OFFSET $%d`,
		fraudFlagColumns, where, argIdx, argIdx+1)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list fraud flags: %w", err)
	}
	defer rows.Close()

	flags, err := scanFraudFlags(rows)
	if err != nil {
		return nil, err
	}

	return &domain.PaginatedResult[domain.FraudFlag]{
		Items:      flags,
		TotalCount: totalCount,
		Page:       filter.Page,
		PageSize:   filter.PageSize,
		TotalPages: int(math.Ceil(float64(totalCount) / float64(filter.PageSize))),
	}, nil
}

func (r *fraudFlagRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.FraudFlagStatus, reviewedBy uuid.UUID) error {
	now := time.Now()
	query := `UPDATE fraud_flags SET status = $2, reviewed_at = $3, reviewed_by = $4 WHERE id = $1`
	result, err := r.pool.Exec(ctx, query, id, string(status), now, reviewedBy)
	if err != nil {
		return fmt.Errorf("update fraud flag status: %w", err)
	}
	if result.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *fraudFlagRepo) CountByUserAndRule(ctx context.Context, userID uuid.UUID, rule domain.FraudRuleName, since time.Time) (int64, error) {
	query := `SELECT COUNT(*) FROM fraud_flags WHERE user_id = $1 AND rule = $2 AND created_at >= $3`
	var count int64
	err := r.pool.QueryRow(ctx, query, userID, string(rule), since).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count fraud flags by user and rule: %w", err)
	}
	return count, nil
}
