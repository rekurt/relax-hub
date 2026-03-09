package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"context"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/handler"
	"github.com/nikitaaldaev/bani/internal/service"
)

// --- Mock payment service ---

type mockPaymentService struct {
	initiateFn       func(ctx context.Context, userID, bookingID uuid.UUID) (string, error)
	handleWebhookFn  func(ctx context.Context, event service.WebhookEvent) error
	refundFn         func(ctx context.Context, bookingID uuid.UUID) error
	getByBookingFn   func(ctx context.Context, userID, bookingID uuid.UUID) (*domain.Payment, error)
	listUserFn       func(ctx context.Context, userID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.Payment], error)
}

func (m *mockPaymentService) InitiatePayment(ctx context.Context, userID, bookingID uuid.UUID) (string, error) {
	if m.initiateFn != nil {
		return m.initiateFn(ctx, userID, bookingID)
	}
	return "", nil
}

func (m *mockPaymentService) HandleWebhook(ctx context.Context, event service.WebhookEvent) error {
	if m.handleWebhookFn != nil {
		return m.handleWebhookFn(ctx, event)
	}
	return nil
}

func (m *mockPaymentService) RefundPayment(ctx context.Context, bookingID uuid.UUID) error {
	if m.refundFn != nil {
		return m.refundFn(ctx, bookingID)
	}
	return nil
}

func (m *mockPaymentService) GetPaymentByBooking(ctx context.Context, userID, bookingID uuid.UUID) (*domain.Payment, error) {
	if m.getByBookingFn != nil {
		return m.getByBookingFn(ctx, userID, bookingID)
	}
	return nil, nil
}

func (m *mockPaymentService) ListUserPayments(ctx context.Context, userID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.Payment], error) {
	if m.listUserFn != nil {
		return m.listUserFn(ctx, userID, page, pageSize)
	}
	return nil, nil
}

var _ service.PaymentService = (*mockPaymentService)(nil)

// --- Tests ---

func TestPaymentHandler_InitiatePayment(t *testing.T) {
	bookingID := uuid.New()
	expectedURL := "https://yookassa.ru/checkout/confirm/abc123"

	svc := &mockPaymentService{
		initiateFn: func(ctx context.Context, uID, bID uuid.UUID) (string, error) {
			if bID != bookingID {
				t.Errorf("expected booking ID %v, got %v", bookingID, bID)
			}
			return expectedURL, nil
		},
	}

	h := handler.NewPaymentHandler(svc)
	router := chi.NewRouter()
	router.Post("/bookings/{id}/pay", h.InitiatePayment)

	req := httptest.NewRequest(http.MethodPost, "/bookings/"+bookingID.String()+"/pay", nil)
	req = req.WithContext(createTestContext(uuid.New(), domain.RoleClient))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	var resp struct {
		Success bool `json:"success"`
		Data    struct {
			ConfirmationURL string `json:"confirmation_url"`
		} `json:"data"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.Data.ConfirmationURL != expectedURL {
		t.Errorf("expected URL %s, got %s", expectedURL, resp.Data.ConfirmationURL)
	}
}

func TestPaymentHandler_InitiatePayment_InvalidID(t *testing.T) {
	svc := &mockPaymentService{}
	h := handler.NewPaymentHandler(svc)
	router := chi.NewRouter()
	router.Post("/bookings/{id}/pay", h.InitiatePayment)

	req := httptest.NewRequest(http.MethodPost, "/bookings/not-a-uuid/pay", nil)
	req = req.WithContext(createTestContext(uuid.New(), domain.RoleClient))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

func TestPaymentHandler_InitiatePayment_AlreadyProcessed(t *testing.T) {
	svc := &mockPaymentService{
		initiateFn: func(ctx context.Context, uID, bID uuid.UUID) (string, error) {
			return "", domain.ErrPaymentAlreadyProcessed
		},
	}

	h := handler.NewPaymentHandler(svc)
	router := chi.NewRouter()
	router.Post("/bookings/{id}/pay", h.InitiatePayment)

	req := httptest.NewRequest(http.MethodPost, "/bookings/"+uuid.New().String()+"/pay", nil)
	req = req.WithContext(createTestContext(uuid.New(), domain.RoleClient))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusConflict {
		t.Errorf("expected status 409, got %d", rec.Code)
	}
}

func TestPaymentHandler_HandleWebhook(t *testing.T) {
	externalID := "ext-123"

	svc := &mockPaymentService{
		handleWebhookFn: func(ctx context.Context, event service.WebhookEvent) error {
			if event.ExternalID != externalID {
				t.Errorf("expected external ID %s, got %s", externalID, event.ExternalID)
			}
			if event.Status != "succeeded" {
				t.Errorf("expected status succeeded, got %s", event.Status)
			}
			return nil
		},
	}

	h := handler.NewPaymentHandler(svc)
	router := chi.NewRouter()
	router.Post("/webhooks/yookassa", h.HandleWebhook)

	body, _ := json.Marshal(map[string]interface{}{
		"event": "payment.succeeded",
		"object": map[string]string{
			"id":     externalID,
			"status": "succeeded",
		},
	})
	req := httptest.NewRequest(http.MethodPost, "/webhooks/yookassa", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
}

func TestPaymentHandler_HandleWebhook_InvalidPayload(t *testing.T) {
	svc := &mockPaymentService{}
	h := handler.NewPaymentHandler(svc)
	router := chi.NewRouter()
	router.Post("/webhooks/yookassa", h.HandleWebhook)

	req := httptest.NewRequest(http.MethodPost, "/webhooks/yookassa", bytes.NewReader([]byte("not json")))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

func TestPaymentHandler_HandleWebhook_MissingID(t *testing.T) {
	svc := &mockPaymentService{}
	h := handler.NewPaymentHandler(svc)
	router := chi.NewRouter()
	router.Post("/webhooks/yookassa", h.HandleWebhook)

	body, _ := json.Marshal(map[string]interface{}{
		"event": "payment.succeeded",
		"object": map[string]string{
			"status": "succeeded",
		},
	})
	req := httptest.NewRequest(http.MethodPost, "/webhooks/yookassa", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

func TestPaymentHandler_ListUserPayments(t *testing.T) {
	userID := uuid.New()
	paymentID := uuid.New()
	bookingID := uuid.New()
	now := time.Now()

	svc := &mockPaymentService{
		listUserFn: func(ctx context.Context, uID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.Payment], error) {
			if uID != userID {
				t.Errorf("expected user ID %v, got %v", userID, uID)
			}
			if page != 1 || pageSize != 20 {
				t.Errorf("expected page 1 size 20, got page %d size %d", page, pageSize)
			}
			return &domain.PaginatedResult[domain.Payment]{
				Items: []domain.Payment{
					{
						ID:        paymentID,
						BookingID: bookingID,
						UserID:    userID,
						Amount:    500000,
						Currency:  "RUB",
						Status:    domain.PaymentSucceeded,
						Provider:  "yookassa",
						CreatedAt: now,
						UpdatedAt: now,
					},
				},
				TotalCount: 1,
				Page:       1,
				PageSize:   20,
				TotalPages: 1,
			}, nil
		},
	}

	h := handler.NewPaymentHandler(svc)
	router := chi.NewRouter()
	router.Get("/my/payments", h.ListUserPayments)

	req := httptest.NewRequest(http.MethodGet, "/my/payments", nil)
	req = req.WithContext(createTestContext(userID, domain.RoleClient))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	var resp struct {
		Success bool `json:"success"`
		Data    []struct {
			ID     string `json:"id"`
			Amount int64  `json:"amount"`
			Status string `json:"status"`
		} `json:"data"`
		Meta struct {
			TotalCount int64 `json:"total_count"`
		} `json:"meta"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(resp.Data) != 1 {
		t.Fatalf("expected 1 payment, got %d", len(resp.Data))
	}
	if resp.Data[0].ID != paymentID.String() {
		t.Errorf("expected payment ID %s, got %s", paymentID, resp.Data[0].ID)
	}
	if resp.Meta.TotalCount != 1 {
		t.Errorf("expected total count 1, got %d", resp.Meta.TotalCount)
	}
}

func TestPaymentHandler_GetBookingPayment(t *testing.T) {
	bookingID := uuid.New()
	paymentID := uuid.New()
	userID := uuid.New()
	now := time.Now()

	svc := &mockPaymentService{
		getByBookingFn: func(ctx context.Context, uID, bID uuid.UUID) (*domain.Payment, error) {
			if bID != bookingID {
				t.Errorf("expected booking ID %v, got %v", bookingID, bID)
			}
			return &domain.Payment{
				ID:        paymentID,
				BookingID: bookingID,
				UserID:    userID,
				Amount:    300000,
				Currency:  "RUB",
				Status:    domain.PaymentSucceeded,
				Provider:  "yookassa",
				CreatedAt: now,
				UpdatedAt: now,
			}, nil
		},
	}

	h := handler.NewPaymentHandler(svc)
	router := chi.NewRouter()
	router.Get("/bookings/{id}/payment", h.GetBookingPayment)

	req := httptest.NewRequest(http.MethodGet, "/bookings/"+bookingID.String()+"/payment", nil)
	req = req.WithContext(createTestContext(userID, domain.RoleClient))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	var resp struct {
		Success bool `json:"success"`
		Data    struct {
			ID     string `json:"id"`
			Amount int64  `json:"amount"`
			Status string `json:"status"`
		} `json:"data"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.Data.ID != paymentID.String() {
		t.Errorf("expected payment ID %s, got %s", paymentID, resp.Data.ID)
	}
	if resp.Data.Amount != 300000 {
		t.Errorf("expected amount 300000, got %d", resp.Data.Amount)
	}
}

func TestPaymentHandler_GetBookingPayment_NotFound(t *testing.T) {
	svc := &mockPaymentService{
		getByBookingFn: func(ctx context.Context, uID, bID uuid.UUID) (*domain.Payment, error) {
			return nil, domain.ErrPaymentNotFound
		},
	}

	h := handler.NewPaymentHandler(svc)
	router := chi.NewRouter()
	router.Get("/bookings/{id}/payment", h.GetBookingPayment)

	req := httptest.NewRequest(http.MethodGet, "/bookings/"+uuid.New().String()+"/payment", nil)
	req = req.WithContext(createTestContext(uuid.New(), domain.RoleClient))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", rec.Code)
	}
}

func TestPaymentHandler_GetBookingPayment_InvalidID(t *testing.T) {
	svc := &mockPaymentService{}
	h := handler.NewPaymentHandler(svc)
	router := chi.NewRouter()
	router.Get("/bookings/{id}/payment", h.GetBookingPayment)

	req := httptest.NewRequest(http.MethodGet, "/bookings/invalid/payment", nil)
	req = req.WithContext(createTestContext(uuid.New(), domain.RoleClient))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}
