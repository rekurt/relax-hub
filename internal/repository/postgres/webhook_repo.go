package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rekurt/relax-hub/internal/domain"
	"github.com/rekurt/relax-hub/internal/repository"
)

type webhookRepo struct {
	pool *pgxpool.Pool
}

func NewWebhookRepository(pool *pgxpool.Pool) repository.WebhookRepository {
	return &webhookRepo{pool: pool}
}

var webhookColumns = `id, owner_id, url, secret, events, is_active, created_at, updated_at`

func scanWebhook(row pgx.Row) (*domain.Webhook, error) {
	var w domain.Webhook
	var eventsJSON []byte
	err := row.Scan(&w.ID, &w.OwnerID, &w.URL, &w.Secret, &eventsJSON, &w.IsActive, &w.CreatedAt, &w.UpdatedAt)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(eventsJSON, &w.Events); err != nil {
		return nil, fmt.Errorf("unmarshal webhook events: %w", err)
	}
	return &w, nil
}

func (r *webhookRepo) Create(ctx context.Context, webhook *domain.Webhook) error {
	now := time.Now()
	if webhook.ID == uuid.Nil {
		webhook.ID = uuid.New()
	}
	if webhook.CreatedAt.IsZero() {
		webhook.CreatedAt = now
	}
	webhook.UpdatedAt = now

	eventsJSON, err := json.Marshal(webhook.Events)
	if err != nil {
		return fmt.Errorf("marshal webhook events: %w", err)
	}

	query := `INSERT INTO webhooks (id, owner_id, url, secret, events, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	_, err = r.pool.Exec(ctx, query,
		webhook.ID, webhook.OwnerID, webhook.URL, webhook.Secret,
		eventsJSON, webhook.IsActive, webhook.CreatedAt, webhook.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("create webhook: %w", err)
	}
	return nil
}

func (r *webhookRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Webhook, error) {
	query := fmt.Sprintf(`SELECT %s FROM webhooks WHERE id = $1`, webhookColumns)
	w, err := scanWebhook(r.pool.QueryRow(ctx, query, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrWebhookNotFound
		}
		return nil, fmt.Errorf("get webhook: %w", err)
	}
	return w, nil
}

func (r *webhookRepo) Update(ctx context.Context, webhook *domain.Webhook) error {
	webhook.UpdatedAt = time.Now()
	eventsJSON, err := json.Marshal(webhook.Events)
	if err != nil {
		return fmt.Errorf("marshal webhook events: %w", err)
	}

	query := `UPDATE webhooks SET url = $1, secret = $2, events = $3, is_active = $4, updated_at = $5 WHERE id = $6`
	ct, err := r.pool.Exec(ctx, query, webhook.URL, webhook.Secret, eventsJSON, webhook.IsActive, webhook.UpdatedAt, webhook.ID)
	if err != nil {
		return fmt.Errorf("update webhook: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return domain.ErrWebhookNotFound
	}
	return nil
}

func (r *webhookRepo) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM webhooks WHERE id = $1`
	ct, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete webhook: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return domain.ErrWebhookNotFound
	}
	return nil
}

func (r *webhookRepo) ListByOwner(ctx context.Context, ownerID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.Webhook], error) {
	countQuery := `SELECT COUNT(*) FROM webhooks WHERE owner_id = $1`
	var total int64
	if err := r.pool.QueryRow(ctx, countQuery, ownerID).Scan(&total); err != nil {
		return nil, fmt.Errorf("count webhooks: %w", err)
	}

	offset := (page - 1) * pageSize
	query := fmt.Sprintf(`SELECT %s FROM webhooks WHERE owner_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`, webhookColumns)
	rows, err := r.pool.Query(ctx, query, ownerID, pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("list webhooks: %w", err)
	}
	defer rows.Close()

	var items []domain.Webhook
	for rows.Next() {
		w, err := scanWebhook(rows)
		if err != nil {
			return nil, fmt.Errorf("scan webhook: %w", err)
		}
		items = append(items, *w)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate webhook rows: %w", err)
	}

	return &domain.PaginatedResult[domain.Webhook]{
		Items:      items,
		TotalCount: total,
		Page:       page,
		PageSize:   pageSize,
	}, nil
}

func (r *webhookRepo) CountByOwner(ctx context.Context, ownerID uuid.UUID) (int64, error) {
	query := `SELECT COUNT(*) FROM webhooks WHERE owner_id = $1`
	var count int64
	if err := r.pool.QueryRow(ctx, query, ownerID).Scan(&count); err != nil {
		return 0, fmt.Errorf("count webhooks: %w", err)
	}
	return count, nil
}

func (r *webhookRepo) ListActiveByEvent(ctx context.Context, ownerID uuid.UUID, event domain.WebhookEventType) ([]domain.Webhook, error) {
	query := fmt.Sprintf(`SELECT %s FROM webhooks WHERE owner_id = $1 AND is_active = true AND events @> $2`, webhookColumns)
	eventsJSON, _ := json.Marshal([]domain.WebhookEventType{event})

	rows, err := r.pool.Query(ctx, query, ownerID, eventsJSON)
	if err != nil {
		return nil, fmt.Errorf("list active webhooks by event: %w", err)
	}
	defer rows.Close()

	var items []domain.Webhook
	for rows.Next() {
		w, err := scanWebhook(rows)
		if err != nil {
			return nil, fmt.Errorf("scan webhook: %w", err)
		}
		items = append(items, *w)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate webhook rows: %w", err)
	}
	return items, nil
}

// Webhook Delivery Repository

type webhookDeliveryRepo struct {
	pool *pgxpool.Pool
}

func NewWebhookDeliveryRepository(pool *pgxpool.Pool) repository.WebhookDeliveryRepository {
	return &webhookDeliveryRepo{pool: pool}
}

var deliveryColumns = `id, webhook_id, event_type, payload, status, http_status, error_message, attempt_count, next_retry_at, created_at, updated_at`

func scanDelivery(row pgx.Row) (*domain.WebhookDelivery, error) {
	var d domain.WebhookDelivery
	err := row.Scan(
		&d.ID, &d.WebhookID, &d.EventType, &d.Payload, &d.Status,
		&d.HTTPStatus, &d.ErrorMessage, &d.AttemptCount, &d.NextRetryAt,
		&d.CreatedAt, &d.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &d, nil
}

func (r *webhookDeliveryRepo) Create(ctx context.Context, delivery *domain.WebhookDelivery) error {
	now := time.Now()
	if delivery.ID == uuid.Nil {
		delivery.ID = uuid.New()
	}
	if delivery.CreatedAt.IsZero() {
		delivery.CreatedAt = now
	}
	delivery.UpdatedAt = now

	query := `INSERT INTO webhook_deliveries (id, webhook_id, event_type, payload, status, http_status, error_message, attempt_count, next_retry_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`

	_, err := r.pool.Exec(ctx, query,
		delivery.ID, delivery.WebhookID, delivery.EventType, delivery.Payload,
		delivery.Status, delivery.HTTPStatus, delivery.ErrorMessage,
		delivery.AttemptCount, delivery.NextRetryAt, delivery.CreatedAt, delivery.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("create webhook delivery: %w", err)
	}
	return nil
}

func (r *webhookDeliveryRepo) Update(ctx context.Context, delivery *domain.WebhookDelivery) error {
	delivery.UpdatedAt = time.Now()
	query := `UPDATE webhook_deliveries SET status = $1, http_status = $2, error_message = $3, attempt_count = $4, next_retry_at = $5, updated_at = $6 WHERE id = $7`
	_, err := r.pool.Exec(ctx, query,
		delivery.Status, delivery.HTTPStatus, delivery.ErrorMessage,
		delivery.AttemptCount, delivery.NextRetryAt, delivery.UpdatedAt, delivery.ID,
	)
	if err != nil {
		return fmt.Errorf("update webhook delivery: %w", err)
	}
	return nil
}

func (r *webhookDeliveryRepo) ListByWebhook(ctx context.Context, webhookID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.WebhookDelivery], error) {
	countQuery := `SELECT COUNT(*) FROM webhook_deliveries WHERE webhook_id = $1`
	var total int64
	if err := r.pool.QueryRow(ctx, countQuery, webhookID).Scan(&total); err != nil {
		return nil, fmt.Errorf("count webhook deliveries: %w", err)
	}

	offset := (page - 1) * pageSize
	query := fmt.Sprintf(`SELECT %s FROM webhook_deliveries WHERE webhook_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`, deliveryColumns)
	rows, err := r.pool.Query(ctx, query, webhookID, pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("list webhook deliveries: %w", err)
	}
	defer rows.Close()

	var items []domain.WebhookDelivery
	for rows.Next() {
		d, err := scanDelivery(rows)
		if err != nil {
			return nil, fmt.Errorf("scan webhook delivery: %w", err)
		}
		items = append(items, *d)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate webhook delivery rows: %w", err)
	}

	return &domain.PaginatedResult[domain.WebhookDelivery]{
		Items:      items,
		TotalCount: total,
		Page:       page,
		PageSize:   pageSize,
	}, nil
}

func (r *webhookDeliveryRepo) ListPendingRetries(ctx context.Context, before time.Time) ([]domain.WebhookDelivery, error) {
	query := fmt.Sprintf(`SELECT %s FROM webhook_deliveries WHERE status = $1 AND next_retry_at IS NOT NULL AND next_retry_at <= $2 ORDER BY next_retry_at ASC LIMIT 100`, deliveryColumns)
	rows, err := r.pool.Query(ctx, query, domain.WebhookDeliveryPending, before)
	if err != nil {
		return nil, fmt.Errorf("list pending retries: %w", err)
	}
	defer rows.Close()

	var items []domain.WebhookDelivery
	for rows.Next() {
		d, err := scanDelivery(rows)
		if err != nil {
			return nil, fmt.Errorf("scan webhook delivery: %w", err)
		}
		items = append(items, *d)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate webhook delivery rows: %w", err)
	}
	return items, nil
}
