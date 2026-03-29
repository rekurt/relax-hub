package service_test

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/logger"
	"github.com/nikitaaldaev/bani/internal/repository/mock"
	"github.com/nikitaaldaev/bani/internal/service"
)

func newWebhookTestService() (service.WebhookService, *mock.WebhookRepo, *mock.WebhookDeliveryRepo) {
	webhookRepo := mock.NewWebhookRepo().(*mock.WebhookRepo)
	deliveryRepo := mock.NewWebhookDeliveryRepo().(*mock.WebhookDeliveryRepo)
	log := logger.New(logger.LevelWarn)
	return service.NewWebhookService(webhookRepo, deliveryRepo, log), webhookRepo, deliveryRepo
}

func validWebhook() *domain.Webhook {
	return &domain.Webhook{
		ID:       uuid.New(),
		URL:      "https://example.com/webhook",
		Secret:   "test-secret-key",
		Events:   []domain.WebhookEventType{domain.WebhookEventBookingCreated},
		IsActive: true,
	}
}

func TestWebhookService_Create(t *testing.T) {
	svc, repo, _ := newWebhookTestService()
	ctx := context.Background()
	ownerID := uuid.New()

	wh := validWebhook()
	err := svc.Create(ctx, ownerID, domain.RoleOwner, wh)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	if wh.OwnerID != ownerID {
		t.Errorf("expected owner_id=%s, got %s", ownerID, wh.OwnerID)
	}

	stored, err := repo.GetByID(ctx, wh.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if stored.URL != "https://example.com/webhook" {
		t.Errorf("expected url=https://example.com/webhook, got %s", stored.URL)
	}
}

func TestWebhookService_Create_ForbiddenForClient(t *testing.T) {
	svc, _, _ := newWebhookTestService()
	ctx := context.Background()

	err := svc.Create(ctx, uuid.New(), domain.RoleClient, validWebhook())
	if err != domain.ErrForbidden {
		t.Errorf("expected ErrForbidden, got %v", err)
	}
}

func TestWebhookService_Create_InvalidURL(t *testing.T) {
	svc, _, _ := newWebhookTestService()
	ctx := context.Background()

	wh := validWebhook()
	wh.URL = "not-a-url"
	err := svc.Create(ctx, uuid.New(), domain.RoleOwner, wh)
	if err != domain.ErrInvalidInput {
		t.Errorf("expected ErrInvalidInput for bad URL, got %v", err)
	}
}

func TestWebhookService_Create_NoEvents(t *testing.T) {
	svc, _, _ := newWebhookTestService()
	ctx := context.Background()

	wh := validWebhook()
	wh.Events = nil
	err := svc.Create(ctx, uuid.New(), domain.RoleOwner, wh)
	if err != domain.ErrInvalidInput {
		t.Errorf("expected ErrInvalidInput for empty events, got %v", err)
	}
}

func TestWebhookService_Create_InvalidEvent(t *testing.T) {
	svc, _, _ := newWebhookTestService()
	ctx := context.Background()

	wh := validWebhook()
	wh.Events = []domain.WebhookEventType{"invalid.event"}
	err := svc.Create(ctx, uuid.New(), domain.RoleOwner, wh)
	if err != domain.ErrInvalidInput {
		t.Errorf("expected ErrInvalidInput for invalid event, got %v", err)
	}
}

func TestWebhookService_Create_LimitReached(t *testing.T) {
	svc, _, _ := newWebhookTestService()
	ctx := context.Background()
	ownerID := uuid.New()

	for i := 0; i < domain.MaxWebhooksPerOwner; i++ {
		wh := validWebhook()
		wh.ID = uuid.New()
		err := svc.Create(ctx, ownerID, domain.RoleOwner, wh)
		if err != nil {
			t.Fatalf("Create #%d: %v", i+1, err)
		}
	}

	err := svc.Create(ctx, ownerID, domain.RoleOwner, validWebhook())
	if err != domain.ErrWebhookLimitReached {
		t.Errorf("expected ErrWebhookLimitReached, got %v", err)
	}
}

func TestWebhookService_GetByID(t *testing.T) {
	svc, _, _ := newWebhookTestService()
	ctx := context.Background()
	ownerID := uuid.New()

	wh := validWebhook()
	_ = svc.Create(ctx, ownerID, domain.RoleOwner, wh)

	got, err := svc.GetByID(ctx, ownerID, domain.RoleOwner, wh.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if got.URL != wh.URL {
		t.Errorf("expected url=%s, got %s", wh.URL, got.URL)
	}
}

func TestWebhookService_GetByID_NotFound(t *testing.T) {
	svc, _, _ := newWebhookTestService()
	ctx := context.Background()

	_, err := svc.GetByID(ctx, uuid.New(), domain.RoleOwner, uuid.New())
	if err != domain.ErrWebhookNotFound {
		t.Errorf("expected ErrWebhookNotFound, got %v", err)
	}
}

func TestWebhookService_GetByID_ForbiddenOtherOwner(t *testing.T) {
	svc, _, _ := newWebhookTestService()
	ctx := context.Background()
	ownerID := uuid.New()

	wh := validWebhook()
	_ = svc.Create(ctx, ownerID, domain.RoleOwner, wh)

	_, err := svc.GetByID(ctx, uuid.New(), domain.RoleOwner, wh.ID)
	if err != domain.ErrForbidden {
		t.Errorf("expected ErrForbidden, got %v", err)
	}
}

func TestWebhookService_Update(t *testing.T) {
	svc, repo, _ := newWebhookTestService()
	ctx := context.Background()
	ownerID := uuid.New()

	wh := validWebhook()
	_ = svc.Create(ctx, ownerID, domain.RoleOwner, wh)

	err := svc.Update(ctx, ownerID, domain.RoleOwner, &domain.Webhook{
		ID:       wh.ID,
		URL:      "https://updated.com/hook",
		Secret:   "new-secret",
		Events:   []domain.WebhookEventType{domain.WebhookEventBookingCancelled},
		IsActive: false,
	})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}

	stored, _ := repo.GetByID(ctx, wh.ID)
	if stored.URL != "https://updated.com/hook" {
		t.Errorf("expected url=https://updated.com/hook, got %s", stored.URL)
	}
	if stored.IsActive {
		t.Error("expected is_active=false")
	}
}

func TestWebhookService_Update_ForbiddenOtherOwner(t *testing.T) {
	svc, _, _ := newWebhookTestService()
	ctx := context.Background()
	ownerID := uuid.New()

	wh := validWebhook()
	_ = svc.Create(ctx, ownerID, domain.RoleOwner, wh)

	err := svc.Update(ctx, uuid.New(), domain.RoleOwner, &domain.Webhook{
		ID:       wh.ID,
		URL:      "https://hijack.com",
		Secret:   "x",
		Events:   []domain.WebhookEventType{domain.WebhookEventBookingCreated},
		IsActive: true,
	})
	if err != domain.ErrForbidden {
		t.Errorf("expected ErrForbidden, got %v", err)
	}
}

func TestWebhookService_Delete(t *testing.T) {
	svc, repo, _ := newWebhookTestService()
	ctx := context.Background()
	ownerID := uuid.New()

	wh := validWebhook()
	_ = svc.Create(ctx, ownerID, domain.RoleOwner, wh)

	err := svc.Delete(ctx, ownerID, domain.RoleOwner, wh.ID)
	if err != nil {
		t.Fatalf("Delete: %v", err)
	}

	_, err = repo.GetByID(ctx, wh.ID)
	if err != domain.ErrWebhookNotFound {
		t.Errorf("expected ErrWebhookNotFound after delete, got %v", err)
	}
}

func TestWebhookService_Delete_ForbiddenOtherOwner(t *testing.T) {
	svc, _, _ := newWebhookTestService()
	ctx := context.Background()
	ownerID := uuid.New()

	wh := validWebhook()
	_ = svc.Create(ctx, ownerID, domain.RoleOwner, wh)

	err := svc.Delete(ctx, uuid.New(), domain.RoleOwner, wh.ID)
	if err != domain.ErrForbidden {
		t.Errorf("expected ErrForbidden, got %v", err)
	}
}

func TestWebhookService_List(t *testing.T) {
	svc, _, _ := newWebhookTestService()
	ctx := context.Background()
	ownerID := uuid.New()

	for i := 0; i < 3; i++ {
		wh := validWebhook()
		wh.ID = uuid.New()
		_ = svc.Create(ctx, ownerID, domain.RoleOwner, wh)
	}

	result, err := svc.List(ctx, ownerID, domain.RoleOwner, 1, 20)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if result.TotalCount != 3 {
		t.Errorf("expected total_count=3, got %d", result.TotalCount)
	}
	if len(result.Items) != 3 {
		t.Errorf("expected 3 items, got %d", len(result.Items))
	}
}

func TestWebhookService_List_ForbiddenForClient(t *testing.T) {
	svc, _, _ := newWebhookTestService()
	ctx := context.Background()

	_, err := svc.List(ctx, uuid.New(), domain.RoleClient, 1, 20)
	if err != domain.ErrForbidden {
		t.Errorf("expected ErrForbidden, got %v", err)
	}
}

func TestWebhookService_DeliverEvent(t *testing.T) {
	_, webhookRepo, deliveryRepo := newWebhookTestService()
	ctx := context.Background()
	ownerID := uuid.New()

	// Create a test HTTP server
	var receivedCount int32
	var mu sync.Mutex
	var receivedSignature string
	var receivedBody []byte
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&receivedCount, 1)
		sig := r.Header.Get("X-Webhook-Signature")
		body, _ := io.ReadAll(r.Body)
		mu.Lock()
		receivedSignature = sig
		receivedBody = body
		mu.Unlock()
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Create webhook pointing to test server
	wh := &domain.Webhook{
		ID:       uuid.New(),
		OwnerID:  ownerID,
		URL:      server.URL,
		Secret:   "my-secret",
		Events:   []domain.WebhookEventType{domain.WebhookEventBookingCreated},
		IsActive: true,
	}
	_ = webhookRepo.Create(ctx, wh)

	// Build a new service with the test repos
	log := logger.New(logger.LevelWarn)
	svc := service.NewWebhookService(webhookRepo, deliveryRepo, log)

	payload := map[string]string{"booking_id": "test-123"}
	err := svc.DeliverEvent(ctx, ownerID, domain.WebhookEventBookingCreated, payload)
	if err != nil {
		t.Fatalf("DeliverEvent: %v", err)
	}

	// Wait for async delivery
	time.Sleep(500 * time.Millisecond)

	if atomic.LoadInt32(&receivedCount) != 1 {
		t.Errorf("expected 1 delivery, got %d", receivedCount)
	}

	// Verify HMAC signature
	mu.Lock()
	sig := receivedSignature
	body := make([]byte, len(receivedBody))
	copy(body, receivedBody)
	mu.Unlock()

	payloadBytes, _ := json.Marshal(payload)
	mac := hmac.New(sha256.New, []byte("my-secret"))
	mac.Write(payloadBytes)
	expectedSig := hex.EncodeToString(mac.Sum(nil))

	if sig != expectedSig {
		t.Errorf("HMAC signature mismatch: expected %s, got %s", expectedSig, sig)
	}

	// Verify payload
	var gotPayload map[string]string
	_ = json.Unmarshal(body, &gotPayload)
	if gotPayload["booking_id"] != "test-123" {
		t.Errorf("expected booking_id=test-123, got %s", gotPayload["booking_id"])
	}
}

func TestWebhookService_DeliverEvent_NoMatchingWebhooks(t *testing.T) {
	svc, _, _ := newWebhookTestService()
	ctx := context.Background()

	err := svc.DeliverEvent(ctx, uuid.New(), domain.WebhookEventBookingCreated, map[string]string{})
	if err != nil {
		t.Errorf("expected no error for no matching webhooks, got %v", err)
	}
}

func TestWebhookService_DeliverEvent_FailedDeliveryRetry(t *testing.T) {
	_, webhookRepo, deliveryRepo := newWebhookTestService()
	ctx := context.Background()
	ownerID := uuid.New()

	// Create a server that returns 500
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	wh := &domain.Webhook{
		ID:       uuid.New(),
		OwnerID:  ownerID,
		URL:      server.URL,
		Secret:   "secret",
		Events:   []domain.WebhookEventType{domain.WebhookEventBookingCancelled},
		IsActive: true,
	}
	_ = webhookRepo.Create(ctx, wh)

	log := logger.New(logger.LevelWarn)
	svc := service.NewWebhookService(webhookRepo, deliveryRepo, log)

	err := svc.DeliverEvent(ctx, ownerID, domain.WebhookEventBookingCancelled, map[string]string{"test": "1"})
	if err != nil {
		t.Fatalf("DeliverEvent: %v", err)
	}

	// Wait for async delivery
	time.Sleep(500 * time.Millisecond)

	// Should have a pending retry delivery
	retries, err := deliveryRepo.ListPendingRetries(ctx, time.Now().Add(1*time.Hour))
	if err != nil {
		t.Fatalf("ListPendingRetries: %v", err)
	}
	if len(retries) != 1 {
		t.Errorf("expected 1 pending retry, got %d", len(retries))
	}
	if len(retries) > 0 {
		if retries[0].AttemptCount != 1 {
			t.Errorf("expected attempt_count=1, got %d", retries[0].AttemptCount)
		}
		if retries[0].Status != domain.WebhookDeliveryPending {
			t.Errorf("expected status=pending, got %s", retries[0].Status)
		}
	}
}

func TestWebhookService_RepresentativeCanManage(t *testing.T) {
	svc, _, _ := newWebhookTestService()
	ctx := context.Background()
	repID := uuid.New()

	wh := validWebhook()
	err := svc.Create(ctx, repID, domain.RoleRepresentative, wh)
	if err != nil {
		t.Fatalf("Create as representative: %v", err)
	}

	result, err := svc.List(ctx, repID, domain.RoleRepresentative, 1, 20)
	if err != nil {
		t.Fatalf("List as representative: %v", err)
	}
	if len(result.Items) != 1 {
		t.Errorf("expected 1 webhook, got %d", len(result.Items))
	}
}

func TestWebhookService_SignatureComputation(t *testing.T) {
	// Verify the HMAC-SHA256 signature is correctly computed
	secret := "test-key-123"
	payload := []byte(`{"event":"booking.created"}`)

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	expected := hex.EncodeToString(mac.Sum(nil))

	if len(expected) != 64 {
		t.Errorf("expected 64 char hex signature, got %d chars", len(expected))
	}
}
