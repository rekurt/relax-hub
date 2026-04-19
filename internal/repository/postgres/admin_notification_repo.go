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

type adminNotificationRepo struct {
	pool *pgxpool.Pool
}

func NewAdminNotificationRepository(pool *pgxpool.Pool) repository.AdminNotificationRepository {
	return &adminNotificationRepo{pool: pool}
}

var adminNotifColumns = `id, role, severity, type, title, body, data, is_read, read_at, read_by, created_at`

func scanAdminNotification(row pgx.Row) (*domain.AdminNotification, error) {
	var n domain.AdminNotification
	err := row.Scan(
		&n.ID, &n.Role, &n.Severity, &n.Type, &n.Title, &n.Body,
		&n.Data, &n.IsRead, &n.ReadAt, &n.ReadBy, &n.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &n, nil
}

func scanAdminNotifications(rows pgx.Rows) ([]domain.AdminNotification, error) {
	var notifs []domain.AdminNotification
	for rows.Next() {
		n, err := scanAdminNotification(rows)
		if err != nil {
			return nil, fmt.Errorf("scan admin notification: %w", err)
		}
		notifs = append(notifs, *n)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate admin notification rows: %w", err)
	}
	return notifs, nil
}

func (r *adminNotificationRepo) Create(ctx context.Context, notif *domain.AdminNotification) error {
	if notif.ID == uuid.Nil {
		notif.ID = uuid.New()
	}
	if notif.CreatedAt.IsZero() {
		notif.CreatedAt = time.Now()
	}
	if notif.Data == nil {
		notif.Data = []byte("{}")
	}

	query := `
		INSERT INTO admin_notifications (id, role, severity, type, title, body, data, is_read, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`

	_, err := r.pool.Exec(ctx, query,
		notif.ID, string(notif.Role), string(notif.Severity), string(notif.Type),
		notif.Title, notif.Body, notif.Data, notif.IsRead, notif.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("create admin notification: %w", err)
	}
	return nil
}

func (r *adminNotificationRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.AdminNotification, error) {
	query := fmt.Sprintf(`SELECT %s FROM admin_notifications WHERE id = $1`, adminNotifColumns)
	row := r.pool.QueryRow(ctx, query, id)
	notif, err := scanAdminNotification(row)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("get admin notification by id: %w", err)
	}
	return notif, nil
}

func (r *adminNotificationRepo) List(ctx context.Context, filter domain.AdminNotificationFilter) (*domain.PaginatedResult[domain.AdminNotification], error) {
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PageSize < 1 {
		filter.PageSize = 20
	}

	var conditions []string
	var args []interface{}
	argIdx := 1

	if filter.Role != nil {
		conditions = append(conditions, fmt.Sprintf("role = $%d", argIdx))
		args = append(args, string(*filter.Role))
		argIdx++
	}
	if filter.Severity != nil {
		conditions = append(conditions, fmt.Sprintf("severity = $%d", argIdx))
		args = append(args, string(*filter.Severity))
		argIdx++
	}
	if filter.Type != nil {
		conditions = append(conditions, fmt.Sprintf("type = $%d", argIdx))
		args = append(args, string(*filter.Type))
		argIdx++
	}
	if filter.IsRead != nil {
		conditions = append(conditions, fmt.Sprintf("is_read = $%d", argIdx))
		args = append(args, *filter.IsRead)
		argIdx++
	}

	where := ""
	if len(conditions) > 0 {
		where = " WHERE " + strings.Join(conditions, " AND ")
	}

	var totalCount int64
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM admin_notifications`+where, args...).Scan(&totalCount)
	if err != nil {
		return nil, fmt.Errorf("count admin notifications: %w", err)
	}

	offset := (filter.Page - 1) * filter.PageSize
	args = append(args, filter.PageSize, offset)
	query := fmt.Sprintf(`SELECT %s FROM admin_notifications%s ORDER BY created_at DESC LIMIT $%d OFFSET $%d`,
		adminNotifColumns, where, argIdx, argIdx+1)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list admin notifications: %w", err)
	}
	defer rows.Close()

	notifs, err := scanAdminNotifications(rows)
	if err != nil {
		return nil, err
	}

	return &domain.PaginatedResult[domain.AdminNotification]{
		Items:      notifs,
		TotalCount: totalCount,
		Page:       filter.Page,
		PageSize:   filter.PageSize,
		TotalPages: int(math.Ceil(float64(totalCount) / float64(filter.PageSize))),
	}, nil
}

func (r *adminNotificationRepo) MarkAsRead(ctx context.Context, id uuid.UUID, readBy uuid.UUID) error {
	now := time.Now()
	query := `UPDATE admin_notifications SET is_read = true, read_at = $2, read_by = $3 WHERE id = $1 AND is_read = false`
	result, err := r.pool.Exec(ctx, query, id, now, readBy)
	if err != nil {
		return fmt.Errorf("mark admin notification as read: %w", err)
	}
	if result.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *adminNotificationRepo) MarkAllAsReadByRole(ctx context.Context, role domain.AdminSubRole, readBy uuid.UUID) error {
	now := time.Now()
	query := `UPDATE admin_notifications SET is_read = true, read_at = $2, read_by = $3 WHERE role = $1 AND is_read = false`
	_, err := r.pool.Exec(ctx, query, string(role), now, readBy)
	if err != nil {
		return fmt.Errorf("mark all admin notifications as read: %w", err)
	}
	return nil
}

func (r *adminNotificationRepo) CountUnreadByRole(ctx context.Context, role domain.AdminSubRole) (int64, error) {
	var count int64
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM admin_notifications WHERE role = $1 AND is_read = false`,
		string(role),
	).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count unread admin notifications: %w", err)
	}
	return count, nil
}

func (r *adminNotificationRepo) ListUnreadCriticalByRole(ctx context.Context, role domain.AdminSubRole) ([]domain.AdminNotification, error) {
	query := fmt.Sprintf(`SELECT %s FROM admin_notifications WHERE role = $1 AND is_read = false AND severity IN ('error', 'critical') ORDER BY created_at DESC LIMIT 100`, adminNotifColumns)
	rows, err := r.pool.Query(ctx, query, string(role))
	if err != nil {
		return nil, fmt.Errorf("list unread critical admin notifications: %w", err)
	}
	defer rows.Close()
	return scanAdminNotifications(rows)
}

func (r *adminNotificationRepo) DeleteOlderThan(ctx context.Context, before time.Time) (int64, error) {
	result, err := r.pool.Exec(ctx,
		`DELETE FROM admin_notifications WHERE created_at < $1 AND is_read = true`,
		before,
	)
	if err != nil {
		return 0, fmt.Errorf("delete old admin notifications: %w", err)
	}
	return result.RowsAffected(), nil
}
