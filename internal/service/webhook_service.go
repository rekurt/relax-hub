package service

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/rekurt/relax-hub/internal/domain"
	"github.com/rekurt/relax-hub/internal/logger"
	"github.com/rekurt/relax-hub/internal/repository"
)

type WebhookService interface {
	Create(ctx context.Context, userID uuid.UUID, role domain.UserRole, webhook *domain.Webhook) error
	GetByID(ctx context.Context, userID uuid.UUID, role domain.UserRole, id uuid.UUID) (*domain.Webhook, error)
	Update(ctx context.Context, userID uuid.UUID, role domain.UserRole, webhook *domain.Webhook) error
	Delete(ctx context.Context, userID uuid.UUID, role domain.UserRole, id uuid.UUID) error
	List(ctx context.Context, userID uuid.UUID, role domain.UserRole, page, pageSize int) (*domain.PaginatedResult[domain.Webhook], error)
	ListDeliveries(ctx context.Context, userID uuid.UUID, role domain.UserRole, webhookID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.WebhookDelivery], error)
	DeliverEvent(ctx context.Context, ownerID uuid.UUID, eventType domain.WebhookEventType, payload interface{}) error
	TestWebhook(ctx context.Context, userID uuid.UUID, role domain.UserRole, id uuid.UUID) error
	RetryFailedDeliveries(ctx context.Context) (int, error)
}

type webhookService struct {
	webhookRepo  repository.WebhookRepository
	deliveryRepo repository.WebhookDeliveryRepository
	httpClient   *http.Client
	logger       *logger.Logger
}

// safeDialContext prevents SSRF by blocking connections to private/loopback/link-local IPs.
func safeDialContext(ctx context.Context, network, addr string) (net.Conn, error) {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, fmt.Errorf("invalid address: %s", addr)
	}
	ips, err := net.DefaultResolver.LookupIPAddr(ctx, host)
	if err != nil {
		return nil, err
	}
	for _, ip := range ips {
		if ip.IP.IsLoopback() || ip.IP.IsPrivate() || ip.IP.IsLinkLocalUnicast() || ip.IP.IsLinkLocalMulticast() || ip.IP.IsUnspecified() {
			return nil, fmt.Errorf("webhook target resolves to private IP: %s", ip.IP)
		}
	}
	var dialer net.Dialer
	return dialer.DialContext(ctx, network, net.JoinHostPort(ips[0].IP.String(), port))
}

func NewWebhookService(
	webhookRepo repository.WebhookRepository,
	deliveryRepo repository.WebhookDeliveryRepository,
	log *logger.Logger,
) WebhookService {
	return &webhookService{
		webhookRepo:  webhookRepo,
		deliveryRepo: deliveryRepo,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				DialContext: safeDialContext,
			},
		},
		logger: log,
	}
}

// NewWebhookServiceWithClient creates a webhook service with a custom HTTP client (for testing).
func NewWebhookServiceWithClient(
	webhookRepo repository.WebhookRepository,
	deliveryRepo repository.WebhookDeliveryRepository,
	log *logger.Logger,
	client *http.Client,
) WebhookService {
	return &webhookService{
		webhookRepo:  webhookRepo,
		deliveryRepo: deliveryRepo,
		httpClient:   client,
		logger:       log,
	}
}

func (s *webhookService) Create(ctx context.Context, userID uuid.UUID, role domain.UserRole, webhook *domain.Webhook) error {
	if role != domain.RoleOwner && role != domain.RoleRepresentative && role != domain.RoleAdmin {
		return domain.ErrForbidden
	}

	webhook.OwnerID = userID

	if err := webhook.Validate(); err != nil {
		return err
	}

	count, err := s.webhookRepo.CountByOwner(ctx, userID)
	if err != nil {
		return fmt.Errorf("count webhooks: %w", err)
	}
	if count >= int64(domain.MaxWebhooksPerOwner) {
		return domain.ErrWebhookLimitReached
	}

	return s.webhookRepo.Create(ctx, webhook)
}

func (s *webhookService) GetByID(ctx context.Context, userID uuid.UUID, role domain.UserRole, id uuid.UUID) (*domain.Webhook, error) {
	if role != domain.RoleOwner && role != domain.RoleRepresentative && role != domain.RoleAdmin {
		return nil, domain.ErrForbidden
	}

	webhook, err := s.webhookRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if role != domain.RoleAdmin && webhook.OwnerID != userID {
		return nil, domain.ErrForbidden
	}

	return webhook, nil
}

func (s *webhookService) Update(ctx context.Context, userID uuid.UUID, role domain.UserRole, webhook *domain.Webhook) error {
	if role != domain.RoleOwner && role != domain.RoleRepresentative && role != domain.RoleAdmin {
		return domain.ErrForbidden
	}

	existing, err := s.webhookRepo.GetByID(ctx, webhook.ID)
	if err != nil {
		return err
	}

	if role != domain.RoleAdmin && existing.OwnerID != userID {
		return domain.ErrForbidden
	}

	existing.URL = webhook.URL
	existing.Secret = webhook.Secret
	existing.Events = webhook.Events
	existing.IsActive = webhook.IsActive

	if err := existing.Validate(); err != nil {
		return err
	}

	return s.webhookRepo.Update(ctx, existing)
}

func (s *webhookService) Delete(ctx context.Context, userID uuid.UUID, role domain.UserRole, id uuid.UUID) error {
	if role != domain.RoleOwner && role != domain.RoleRepresentative && role != domain.RoleAdmin {
		return domain.ErrForbidden
	}

	existing, err := s.webhookRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if role != domain.RoleAdmin && existing.OwnerID != userID {
		return domain.ErrForbidden
	}

	return s.webhookRepo.Delete(ctx, id)
}

func (s *webhookService) List(ctx context.Context, userID uuid.UUID, role domain.UserRole, page, pageSize int) (*domain.PaginatedResult[domain.Webhook], error) {
	if role != domain.RoleOwner && role != domain.RoleRepresentative && role != domain.RoleAdmin {
		return nil, domain.ErrForbidden
	}
	return s.webhookRepo.ListByOwner(ctx, userID, page, pageSize)
}

func (s *webhookService) ListDeliveries(ctx context.Context, userID uuid.UUID, role domain.UserRole, webhookID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.WebhookDelivery], error) {
	if role != domain.RoleOwner && role != domain.RoleRepresentative && role != domain.RoleAdmin {
		return nil, domain.ErrForbidden
	}

	webhook, err := s.webhookRepo.GetByID(ctx, webhookID)
	if err != nil {
		return nil, err
	}
	if role != domain.RoleAdmin && webhook.OwnerID != userID {
		return nil, domain.ErrForbidden
	}

	return s.deliveryRepo.ListByWebhook(ctx, webhookID, page, pageSize)
}

func (s *webhookService) DeliverEvent(ctx context.Context, ownerID uuid.UUID, eventType domain.WebhookEventType, payload interface{}) error {
	webhooks, err := s.webhookRepo.ListActiveByEvent(ctx, ownerID, eventType)
	if err != nil {
		return fmt.Errorf("list webhooks for event: %w", err)
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal webhook payload: %w", err)
	}

	for _, wh := range webhooks {
		delivery := &domain.WebhookDelivery{
			ID:           uuid.New(),
			WebhookID:    wh.ID,
			EventType:    eventType,
			Payload:      payloadBytes,
			Status:       domain.WebhookDeliveryPending,
			AttemptCount: 0,
		}
		if err := s.deliveryRepo.Create(ctx, delivery); err != nil {
			s.logger.Error("create webhook delivery", "webhook_id", wh.ID, "error", err)
			continue
		}

		go s.attemptDelivery(wh, delivery)
	}

	return nil
}

func (s *webhookService) TestWebhook(ctx context.Context, userID uuid.UUID, role domain.UserRole, id uuid.UUID) error {
	if role != domain.RoleOwner && role != domain.RoleRepresentative && role != domain.RoleAdmin {
		return domain.ErrForbidden
	}

	webhook, err := s.webhookRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if role != domain.RoleAdmin && webhook.OwnerID != userID {
		return domain.ErrForbidden
	}

	testPayload := map[string]interface{}{
		"event": "test",
		"data":  map[string]string{"message": "This is a test webhook delivery"},
	}

	payloadBytes, _ := json.Marshal(testPayload)
	delivery := &domain.WebhookDelivery{
		ID:           uuid.New(),
		WebhookID:    webhook.ID,
		EventType:    "test",
		Payload:      payloadBytes,
		Status:       domain.WebhookDeliveryPending,
		AttemptCount: 0,
	}
	if err := s.deliveryRepo.Create(ctx, delivery); err != nil {
		return fmt.Errorf("create test delivery: %w", err)
	}

	go s.attemptDelivery(*webhook, delivery)
	return nil
}

func (s *webhookService) RetryFailedDeliveries(ctx context.Context) (int, error) {
	deliveries, err := s.deliveryRepo.ListPendingRetries(ctx, time.Now())
	if err != nil {
		return 0, fmt.Errorf("list pending retries: %w", err)
	}

	retried := 0
	for _, d := range deliveries {
		webhook, err := s.webhookRepo.GetByID(ctx, d.WebhookID)
		if err != nil {
			s.logger.Error("get webhook for retry", "webhook_id", d.WebhookID, "error", err)
			continue
		}
		if !webhook.IsActive {
			d.Status = domain.WebhookDeliveryFailed
			d.ErrorMessage = "webhook deactivated"
			_ = s.deliveryRepo.Update(ctx, &d)
			continue
		}

		go s.attemptDelivery(*webhook, &d)
		retried++
	}

	return retried, nil
}

func (s *webhookService) attemptDelivery(webhook domain.Webhook, delivery *domain.WebhookDelivery) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	delivery.AttemptCount++

	signature := computeHMAC(webhook.Secret, delivery.Payload)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, webhook.URL, bytes.NewReader(delivery.Payload))
	if err != nil {
		s.markDeliveryFailed(delivery, 0, fmt.Sprintf("create request: %s", err.Error()))
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Webhook-Signature", signature)
	req.Header.Set("X-Webhook-Event", string(delivery.EventType))
	req.Header.Set("X-Webhook-Delivery", delivery.ID.String())

	resp, err := s.httpClient.Do(req)
	if err != nil {
		s.markDeliveryFailed(delivery, 0, fmt.Sprintf("request failed: %s", err.Error()))
		return
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)

	delivery.HTTPStatus = resp.StatusCode

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		delivery.Status = domain.WebhookDeliverySuccess
		delivery.NextRetryAt = nil
		_ = s.deliveryRepo.Update(context.Background(), delivery)
		return
	}

	s.markDeliveryFailed(delivery, resp.StatusCode, fmt.Sprintf("HTTP %d", resp.StatusCode))
}

func (s *webhookService) markDeliveryFailed(delivery *domain.WebhookDelivery, httpStatus int, errMsg string) {
	delivery.HTTPStatus = httpStatus
	delivery.ErrorMessage = errMsg

	if delivery.AttemptCount >= domain.MaxWebhookRetries {
		delivery.Status = domain.WebhookDeliveryFailed
		delivery.NextRetryAt = nil
	} else {
		delivery.Status = domain.WebhookDeliveryPending
		delay := time.Duration(math.Pow(float64(domain.WebhookRetryBaseDelay), float64(delivery.AttemptCount))) * time.Second
		retryAt := time.Now().Add(delay)
		delivery.NextRetryAt = &retryAt
	}

	if err := s.deliveryRepo.Update(context.Background(), delivery); err != nil {
		s.logger.Error("update delivery after failure", "delivery_id", delivery.ID, "error", err)
	}
}

func computeHMAC(secret string, payload []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	return hex.EncodeToString(mac.Sum(nil))
}
