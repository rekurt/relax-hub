package postgres

import (
	"context"
	"fmt"
	"math"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rekurt/relax-hub/internal/domain"
	"github.com/rekurt/relax-hub/internal/repository"
)

type auditLogRepo struct {
	pool *pgxpool.Pool
}

func NewAuditLogRepository(pool *pgxpool.Pool) repository.AuditLogRepository {
	return &auditLogRepo{pool: pool}
}

var auditLogColumns = `id, entity_type, entity_id, user_id, action, changed_fields, created_at`

func scanAuditLog(row pgx.Row) (*domain.AuditLog, error) {
	var a domain.AuditLog
	err := row.Scan(
		&a.ID, &a.EntityType, &a.EntityID, &a.UserID,
		&a.Action, &a.ChangedFields, &a.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func scanAuditLogs(rows pgx.Rows) ([]domain.AuditLog, error) {
	var logs []domain.AuditLog
	for rows.Next() {
		a, err := scanAuditLog(rows)
		if err != nil {
			return nil, fmt.Errorf("scan audit log: %w", err)
		}
		logs = append(logs, *a)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate audit log rows: %w", err)
	}
	return logs, nil
}

func (r *auditLogRepo) Create(ctx context.Context, log *domain.AuditLog) error {
	if log.ID == uuid.Nil {
		log.ID = uuid.New()
	}

	query := `
		INSERT INTO audit_logs (id, entity_type, entity_id, user_id, action, changed_fields, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`

	_, err := r.pool.Exec(ctx, query,
		log.ID, log.EntityType, log.EntityID, log.UserID,
		string(log.Action), log.ChangedFields, log.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("create audit log: %w", err)
	}
	return nil
}

func (r *auditLogRepo) ListByEntity(ctx context.Context, entityType string, entityID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.AuditLog], error) {
	filter := domain.AuditLogFilter{
		EntityType: &entityType,
		EntityID:   &entityID,
		Page:       page,
		PageSize:   pageSize,
	}
	return r.List(ctx, filter)
}

func (r *auditLogRepo) ListByUser(ctx context.Context, userID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.AuditLog], error) {
	filter := domain.AuditLogFilter{
		UserID:   &userID,
		Page:     page,
		PageSize: pageSize,
	}
	return r.List(ctx, filter)
}

func (r *auditLogRepo) List(ctx context.Context, filter domain.AuditLogFilter) (*domain.PaginatedResult[domain.AuditLog], error) {
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PageSize < 1 {
		filter.PageSize = 20
	}

	var conditions []string
	var args []interface{}
	argIdx := 1

	if filter.EntityType != nil {
		conditions = append(conditions, fmt.Sprintf("entity_type = $%d", argIdx))
		args = append(args, *filter.EntityType)
		argIdx++
	}
	if filter.EntityID != nil {
		conditions = append(conditions, fmt.Sprintf("entity_id = $%d", argIdx))
		args = append(args, *filter.EntityID)
		argIdx++
	}
	if filter.UserID != nil {
		conditions = append(conditions, fmt.Sprintf("user_id = $%d", argIdx))
		args = append(args, *filter.UserID)
		argIdx++
	}
	if filter.Action != nil {
		conditions = append(conditions, fmt.Sprintf("action = $%d", argIdx))
		args = append(args, string(*filter.Action))
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
	countQuery := `SELECT COUNT(*) FROM audit_logs` + where
	err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&totalCount)
	if err != nil {
		return nil, fmt.Errorf("count audit logs: %w", err)
	}

	offset := (filter.Page - 1) * filter.PageSize
	args = append(args, filter.PageSize, offset)
	query := fmt.Sprintf(`SELECT %s FROM audit_logs%s ORDER BY created_at DESC LIMIT $%d OFFSET $%d`,
		auditLogColumns, where, argIdx, argIdx+1)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list audit logs: %w", err)
	}
	defer rows.Close()

	logs, err := scanAuditLogs(rows)
	if err != nil {
		return nil, err
	}

	return &domain.PaginatedResult[domain.AuditLog]{
		Items:      logs,
		TotalCount: totalCount,
		Page:       filter.Page,
		PageSize:   filter.PageSize,
		TotalPages: int(math.Ceil(float64(totalCount) / float64(filter.PageSize))),
	}, nil
}
