package postgres

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/repository"
)

type complaintRepo struct {
	pool *pgxpool.Pool
}

func NewComplaintRepository(pool *pgxpool.Pool) repository.ComplaintRepository {
	return &complaintRepo{pool: pool}
}

var complaintColumns = `id, reporter_id, target_type, target_id, reason, description, status, resolved_by_id, resolution, resolved_at, created_at`

func scanComplaint(row pgx.Row) (*domain.Complaint, error) {
	var c domain.Complaint
	err := row.Scan(
		&c.ID, &c.ReporterID, &c.TargetType, &c.TargetID,
		&c.Reason, &c.Description, &c.Status,
		&c.ResolvedByID, &c.Resolution, &c.ResolvedAt, &c.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func scanComplaints(rows pgx.Rows) ([]domain.Complaint, error) {
	var complaints []domain.Complaint
	for rows.Next() {
		var c domain.Complaint
		if err := rows.Scan(
			&c.ID, &c.ReporterID, &c.TargetType, &c.TargetID,
			&c.Reason, &c.Description, &c.Status,
			&c.ResolvedByID, &c.Resolution, &c.ResolvedAt, &c.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan complaint: %w", err)
		}
		complaints = append(complaints, c)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate complaint rows: %w", err)
	}
	return complaints, nil
}

func (r *complaintRepo) Create(ctx context.Context, complaint *domain.Complaint) error {
	query := `
		INSERT INTO complaints (id, reporter_id, target_type, target_id, reason, description, status, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	if complaint.ID == uuid.Nil {
		complaint.ID = uuid.New()
	}
	if complaint.Status == "" {
		complaint.Status = domain.ComplaintStatusPending
	}
	if complaint.CreatedAt.IsZero() {
		complaint.CreatedAt = time.Now()
	}

	_, err := r.pool.Exec(ctx, query,
		complaint.ID, complaint.ReporterID, string(complaint.TargetType), complaint.TargetID,
		string(complaint.Reason), complaint.Description, string(complaint.Status), complaint.CreatedAt,
	)
	if err != nil {
		if isDuplicateKeyError(err) {
			return domain.ErrAlreadyReported
		}
		return fmt.Errorf("create complaint: %w", err)
	}
	return nil
}

func (r *complaintRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Complaint, error) {
	query := `SELECT ` + complaintColumns + ` FROM complaints WHERE id = $1`

	c, err := scanComplaint(r.pool.QueryRow(ctx, query, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrComplaintNotFound
		}
		return nil, fmt.Errorf("get complaint by id: %w", err)
	}
	return c, nil
}

func (r *complaintRepo) List(ctx context.Context, filter domain.ComplaintFilter) (*domain.PaginatedResult[domain.Complaint], error) {
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
	}
	if filter.TargetType != nil {
		conditions = append(conditions, fmt.Sprintf("target_type = $%d", argIdx))
		args = append(args, string(*filter.TargetType))
		argIdx++
	}
	if filter.Reason != nil {
		conditions = append(conditions, fmt.Sprintf("reason = $%d", argIdx))
		args = append(args, string(*filter.Reason))
		argIdx++
	}
	if filter.FromDate != nil {
		conditions = append(conditions, fmt.Sprintf("created_at >= $%d", argIdx))
		args = append(args, *filter.FromDate)
		argIdx++
	}
	if filter.ToDate != nil {
		conditions = append(conditions, fmt.Sprintf("created_at <= $%d", argIdx))
		args = append(args, *filter.ToDate)
		argIdx++
	}

	where := ""
	if len(conditions) > 0 {
		where = " WHERE " + strings.Join(conditions, " AND ")
	}

	var totalCount int64
	countQuery := `SELECT COUNT(*) FROM complaints` + where
	err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&totalCount)
	if err != nil {
		return nil, fmt.Errorf("count complaints: %w", err)
	}

	offset := (filter.Page - 1) * filter.PageSize
	args = append(args, filter.PageSize, offset)
	query := fmt.Sprintf(`SELECT %s FROM complaints%s ORDER BY created_at DESC LIMIT $%d OFFSET $%d`,
		complaintColumns, where, argIdx, argIdx+1)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list complaints: %w", err)
	}
	defer rows.Close()

	complaints, err := scanComplaints(rows)
	if err != nil {
		return nil, err
	}

	return &domain.PaginatedResult[domain.Complaint]{
		Items:      complaints,
		TotalCount: totalCount,
		Page:       filter.Page,
		PageSize:   filter.PageSize,
		TotalPages: int(math.Ceil(float64(totalCount) / float64(filter.PageSize))),
	}, nil
}

func (r *complaintRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.ComplaintStatus, resolvedByID *uuid.UUID, resolution string) error {
	var resolvedAt *time.Time
	if status == domain.ComplaintStatusResolved || status == domain.ComplaintStatusDismissed {
		now := time.Now()
		resolvedAt = &now
	}

	query := `UPDATE complaints SET status = $2, resolved_by_id = $3, resolution = $4, resolved_at = $5 WHERE id = $1`
	result, err := r.pool.Exec(ctx, query, id, string(status), resolvedByID, resolution, resolvedAt)
	if err != nil {
		return fmt.Errorf("update complaint status: %w", err)
	}
	if result.RowsAffected() == 0 {
		return domain.ErrComplaintNotFound
	}
	return nil
}

func (r *complaintRepo) CountByTarget(ctx context.Context, targetType domain.ComplaintTargetType, targetID uuid.UUID) (int64, error) {
	query := `SELECT COUNT(*) FROM complaints WHERE target_type = $1 AND target_id = $2 AND status = 'pending'`
	var count int64
	err := r.pool.QueryRow(ctx, query, string(targetType), targetID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count complaints by target: %w", err)
	}
	return count, nil
}

func (r *complaintRepo) CheckExists(ctx context.Context, reporterID uuid.UUID, targetType domain.ComplaintTargetType, targetID uuid.UUID) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM complaints WHERE reporter_id = $1 AND target_type = $2 AND target_id = $3)`
	var exists bool
	err := r.pool.QueryRow(ctx, query, reporterID, string(targetType), targetID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("check complaint exists: %w", err)
	}
	return exists, nil
}
