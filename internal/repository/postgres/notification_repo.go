package postgres

import (
	"context"
	"encoding/json"
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

type notificationRepo struct {
	pool *pgxpool.Pool
}

func NewNotificationRepository(pool *pgxpool.Pool) repository.NotificationRepository {
	return &notificationRepo{pool: pool}
}

var notificationColumns = `id, user_id, type, title, body, data, is_read, read_at, created_at`

func scanNotification(row pgx.Row) (*domain.Notification, error) {
	var n domain.Notification
	var dataJSON []byte
	err := row.Scan(
		&n.ID, &n.UserID, &n.Type, &n.Title, &n.Body,
		&dataJSON, &n.IsRead, &n.ReadAt, &n.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	if dataJSON != nil {
		if err := json.Unmarshal(dataJSON, &n.Data); err != nil {
			return nil, fmt.Errorf("unmarshal notification data: %w", err)
		}
	}
	return &n, nil
}

func scanNotifications(rows pgx.Rows) ([]domain.Notification, error) {
	var notifications []domain.Notification
	for rows.Next() {
		var n domain.Notification
		var dataJSON []byte
		if err := rows.Scan(
			&n.ID, &n.UserID, &n.Type, &n.Title, &n.Body,
			&dataJSON, &n.IsRead, &n.ReadAt, &n.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan notification: %w", err)
		}
		if dataJSON != nil {
			if err := json.Unmarshal(dataJSON, &n.Data); err != nil {
				return nil, fmt.Errorf("unmarshal notification data: %w", err)
			}
		}
		notifications = append(notifications, n)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate notification rows: %w", err)
	}
	return notifications, nil
}

func (r *notificationRepo) Create(ctx context.Context, notification *domain.Notification) error {
	query := `
		INSERT INTO notifications (id, user_id, type, title, body, data, is_read, read_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`

	if notification.ID == uuid.Nil {
		notification.ID = uuid.New()
	}
	if notification.CreatedAt.IsZero() {
		notification.CreatedAt = time.Now()
	}

	dataJSON, err := json.Marshal(notification.Data)
	if err != nil {
		return fmt.Errorf("marshal notification data: %w", err)
	}

	_, err = r.pool.Exec(ctx, query,
		notification.ID, notification.UserID, notification.Type,
		notification.Title, notification.Body, dataJSON,
		notification.IsRead, notification.ReadAt, notification.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("create notification: %w", err)
	}
	return nil
}

func (r *notificationRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Notification, error) {
	query := `SELECT ` + notificationColumns + ` FROM notifications WHERE id = $1`

	n, err := scanNotification(r.pool.QueryRow(ctx, query, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("get notification by id: %w", err)
	}
	return n, nil
}

func (r *notificationRepo) ListByUser(ctx context.Context, userID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.Notification], error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}

	var totalCount int64
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM notifications WHERE user_id = $1`, userID).Scan(&totalCount)
	if err != nil {
		return nil, fmt.Errorf("count notifications: %w", err)
	}

	offset := (page - 1) * pageSize
	query := `SELECT ` + notificationColumns + ` FROM notifications WHERE user_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`

	rows, err := r.pool.Query(ctx, query, userID, pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("list notifications: %w", err)
	}
	defer rows.Close()

	notifications, err := scanNotifications(rows)
	if err != nil {
		return nil, err
	}

	return &domain.PaginatedResult[domain.Notification]{
		Items:      notifications,
		TotalCount: totalCount,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: int(math.Ceil(float64(totalCount) / float64(pageSize))),
	}, nil
}

func (r *notificationRepo) MarkAsRead(ctx context.Context, id uuid.UUID) error {
	now := time.Now()
	query := `UPDATE notifications SET is_read = true, read_at = $2 WHERE id = $1 AND is_read = false`
	result, err := r.pool.Exec(ctx, query, id, now)
	if err != nil {
		return fmt.Errorf("mark notification as read: %w", err)
	}
	if result.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *notificationRepo) MarkAllAsRead(ctx context.Context, userID uuid.UUID) error {
	now := time.Now()
	query := `UPDATE notifications SET is_read = true, read_at = $2 WHERE user_id = $1 AND is_read = false`
	_, err := r.pool.Exec(ctx, query, userID, now)
	if err != nil {
		return fmt.Errorf("mark all notifications as read: %w", err)
	}
	return nil
}

func (r *notificationRepo) CountUnread(ctx context.Context, userID uuid.UUID) (int64, error) {
	var count int64
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM notifications WHERE user_id = $1 AND is_read = false`, userID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count unread notifications: %w", err)
	}
	return count, nil
}

func (r *notificationRepo) GetPreferences(ctx context.Context, userID uuid.UUID) (*domain.NotificationPreferences, error) {
	query := `SELECT user_id, in_app, email, push, telegram, booking_events, review_events, promo_events, reminders
		FROM notification_preferences WHERE user_id = $1`

	var p domain.NotificationPreferences
	err := r.pool.QueryRow(ctx, query, userID).Scan(
		&p.UserID, &p.InApp, &p.Email, &p.Push, &p.Telegram,
		&p.BookingEvents, &p.ReviewEvents, &p.PromoEvents, &p.Reminders,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			defaults := domain.DefaultNotificationPreferences(userID)
			return &defaults, nil
		}
		return nil, fmt.Errorf("get notification preferences: %w", err)
	}
	return &p, nil
}

func (r *notificationRepo) UpdatePreferences(ctx context.Context, prefs *domain.NotificationPreferences) error {
	query := `
		INSERT INTO notification_preferences (user_id, in_app, email, push, telegram, booking_events, review_events, promo_events, reminders)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (user_id) DO UPDATE SET
			in_app = EXCLUDED.in_app,
			email = EXCLUDED.email,
			push = EXCLUDED.push,
			telegram = EXCLUDED.telegram,
			booking_events = EXCLUDED.booking_events,
			review_events = EXCLUDED.review_events,
			promo_events = EXCLUDED.promo_events,
			reminders = EXCLUDED.reminders`

	_, err := r.pool.Exec(ctx, query,
		prefs.UserID, prefs.InApp, prefs.Email, prefs.Push, prefs.Telegram,
		prefs.BookingEvents, prefs.ReviewEvents, prefs.PromoEvents, prefs.Reminders,
	)
	if err != nil {
		return fmt.Errorf("update notification preferences: %w", err)
	}
	return nil
}
