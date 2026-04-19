package mock

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/rekurt/relax-hub/internal/domain"
	"github.com/rekurt/relax-hub/internal/repository"
)

type WebhookRepo struct {
	mu       sync.RWMutex
	webhooks map[uuid.UUID]*domain.Webhook
}

func NewWebhookRepo() repository.WebhookRepository {
	return &WebhookRepo{
		webhooks: make(map[uuid.UUID]*domain.Webhook),
	}
}

func (r *WebhookRepo) Create(_ context.Context, webhook *domain.Webhook) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	if webhook.ID == uuid.Nil {
		webhook.ID = uuid.New()
	}
	if webhook.CreatedAt.IsZero() {
		webhook.CreatedAt = now
	}
	webhook.UpdatedAt = now

	cp := *webhook
	cp.Events = make([]domain.WebhookEventType, len(webhook.Events))
	copy(cp.Events, webhook.Events)
	r.webhooks[webhook.ID] = &cp
	return nil
}

func (r *WebhookRepo) GetByID(_ context.Context, id uuid.UUID) (*domain.Webhook, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	w, ok := r.webhooks[id]
	if !ok {
		return nil, domain.ErrWebhookNotFound
	}
	cp := *w
	cp.Events = make([]domain.WebhookEventType, len(w.Events))
	copy(cp.Events, w.Events)
	return &cp, nil
}

func (r *WebhookRepo) Update(_ context.Context, webhook *domain.Webhook) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	existing, ok := r.webhooks[webhook.ID]
	if !ok {
		return domain.ErrWebhookNotFound
	}
	existing.URL = webhook.URL
	existing.Secret = webhook.Secret
	existing.Events = make([]domain.WebhookEventType, len(webhook.Events))
	copy(existing.Events, webhook.Events)
	existing.IsActive = webhook.IsActive
	existing.UpdatedAt = time.Now()
	return nil
}

func (r *WebhookRepo) Delete(_ context.Context, id uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.webhooks[id]; !ok {
		return domain.ErrWebhookNotFound
	}
	delete(r.webhooks, id)
	return nil
}

func (r *WebhookRepo) ListByOwner(_ context.Context, ownerID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.Webhook], error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var all []domain.Webhook
	for _, w := range r.webhooks {
		if w.OwnerID == ownerID {
			cp := *w
			cp.Events = make([]domain.WebhookEventType, len(w.Events))
			copy(cp.Events, w.Events)
			all = append(all, cp)
		}
	}

	sort.Slice(all, func(i, j int) bool {
		return all[i].CreatedAt.After(all[j].CreatedAt)
	})

	total := int64(len(all))
	offset := (page - 1) * pageSize
	if offset >= len(all) {
		return &domain.PaginatedResult[domain.Webhook]{
			Items:      []domain.Webhook{},
			TotalCount: total,
			Page:       page,
			PageSize:   pageSize,
		}, nil
	}
	end := offset + pageSize
	if end > len(all) {
		end = len(all)
	}

	return &domain.PaginatedResult[domain.Webhook]{
		Items:      all[offset:end],
		TotalCount: total,
		Page:       page,
		PageSize:   pageSize,
	}, nil
}

func (r *WebhookRepo) CountByOwner(_ context.Context, ownerID uuid.UUID) (int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var count int64
	for _, w := range r.webhooks {
		if w.OwnerID == ownerID {
			count++
		}
	}
	return count, nil
}

func (r *WebhookRepo) ListActiveByEvent(_ context.Context, ownerID uuid.UUID, event domain.WebhookEventType) ([]domain.Webhook, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []domain.Webhook
	for _, w := range r.webhooks {
		if w.OwnerID == ownerID && w.IsActive && w.SubscribedTo(event) {
			cp := *w
			cp.Events = make([]domain.WebhookEventType, len(w.Events))
			copy(cp.Events, w.Events)
			result = append(result, cp)
		}
	}
	return result, nil
}

// WebhookDeliveryRepo is an in-memory mock for WebhookDeliveryRepository.
type WebhookDeliveryRepo struct {
	mu         sync.RWMutex
	deliveries map[uuid.UUID]*domain.WebhookDelivery
}

func NewWebhookDeliveryRepo() repository.WebhookDeliveryRepository {
	return &WebhookDeliveryRepo{
		deliveries: make(map[uuid.UUID]*domain.WebhookDelivery),
	}
}

func (r *WebhookDeliveryRepo) Create(_ context.Context, delivery *domain.WebhookDelivery) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	if delivery.ID == uuid.Nil {
		delivery.ID = uuid.New()
	}
	if delivery.CreatedAt.IsZero() {
		delivery.CreatedAt = now
	}
	delivery.UpdatedAt = now

	cp := *delivery
	cp.Payload = make([]byte, len(delivery.Payload))
	copy(cp.Payload, delivery.Payload)
	r.deliveries[delivery.ID] = &cp
	return nil
}

func (r *WebhookDeliveryRepo) Update(_ context.Context, delivery *domain.WebhookDelivery) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	existing, ok := r.deliveries[delivery.ID]
	if !ok {
		return domain.ErrNotFound
	}
	existing.Status = delivery.Status
	existing.HTTPStatus = delivery.HTTPStatus
	existing.ErrorMessage = delivery.ErrorMessage
	existing.AttemptCount = delivery.AttemptCount
	existing.NextRetryAt = delivery.NextRetryAt
	existing.UpdatedAt = time.Now()
	return nil
}

func (r *WebhookDeliveryRepo) ListByWebhook(_ context.Context, webhookID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.WebhookDelivery], error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var all []domain.WebhookDelivery
	for _, d := range r.deliveries {
		if d.WebhookID == webhookID {
			cp := *d
			all = append(all, cp)
		}
	}

	sort.Slice(all, func(i, j int) bool {
		return all[i].CreatedAt.After(all[j].CreatedAt)
	})

	total := int64(len(all))
	offset := (page - 1) * pageSize
	if offset >= len(all) {
		return &domain.PaginatedResult[domain.WebhookDelivery]{
			Items:      []domain.WebhookDelivery{},
			TotalCount: total,
			Page:       page,
			PageSize:   pageSize,
		}, nil
	}
	end := offset + pageSize
	if end > len(all) {
		end = len(all)
	}

	return &domain.PaginatedResult[domain.WebhookDelivery]{
		Items:      all[offset:end],
		TotalCount: total,
		Page:       page,
		PageSize:   pageSize,
	}, nil
}

func (r *WebhookDeliveryRepo) ListPendingRetries(_ context.Context, before time.Time) ([]domain.WebhookDelivery, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []domain.WebhookDelivery
	for _, d := range r.deliveries {
		if d.Status == domain.WebhookDeliveryPending && d.NextRetryAt != nil && !d.NextRetryAt.After(before) {
			cp := *d
			result = append(result, cp)
		}
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].NextRetryAt.Before(*result[j].NextRetryAt)
	})

	if len(result) > 100 {
		result = result[:100]
	}
	return result, nil
}
