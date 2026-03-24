package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/middleware"
)

type mockCertificateService struct {
	purchaseFn  func(ctx context.Context, amount int64, purchaserID *uuid.UUID, purchaserEmail, recipientEmail, recipientName, message string) (*domain.GiftCertificate, error)
	redeemFn    func(ctx context.Context, code string, userID uuid.UUID) (*domain.GiftCertificate, error)
	applyFn     func(ctx context.Context, certificateID, bookingID uuid.UUID, amount int64) error
	getBalanceFn func(ctx context.Context, code string) (*domain.GiftCertificate, error)
	listByUserFn func(ctx context.Context, userID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.GiftCertificate], error)
}

func (m *mockCertificateService) Purchase(ctx context.Context, amount int64, purchaserID *uuid.UUID, purchaserEmail, recipientEmail, recipientName, message string) (*domain.GiftCertificate, error) {
	if m.purchaseFn != nil {
		return m.purchaseFn(ctx, amount, purchaserID, purchaserEmail, recipientEmail, recipientName, message)
	}
	return nil, nil
}

func (m *mockCertificateService) Redeem(ctx context.Context, code string, userID uuid.UUID) (*domain.GiftCertificate, error) {
	if m.redeemFn != nil {
		return m.redeemFn(ctx, code, userID)
	}
	return nil, nil
}

func (m *mockCertificateService) Apply(ctx context.Context, certificateID, bookingID uuid.UUID, amount int64) error {
	if m.applyFn != nil {
		return m.applyFn(ctx, certificateID, bookingID, amount)
	}
	return nil
}

func (m *mockCertificateService) GetBalance(ctx context.Context, code string) (*domain.GiftCertificate, error) {
	if m.getBalanceFn != nil {
		return m.getBalanceFn(ctx, code)
	}
	return nil, nil
}

func (m *mockCertificateService) ListByUser(ctx context.Context, userID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.GiftCertificate], error) {
	if m.listByUserFn != nil {
		return m.listByUserFn(ctx, userID, page, pageSize)
	}
	return &domain.PaginatedResult[domain.GiftCertificate]{}, nil
}

func (m *mockCertificateService) RefundUsage(_ context.Context, _ uuid.UUID) error {
	return nil
}

func TestCertificateHandler_Purchase(t *testing.T) {
	certID := uuid.New()
	now := time.Now()

	svc := &mockCertificateService{
		purchaseFn: func(ctx context.Context, amount int64, purchaserID *uuid.UUID, purchaserEmail, recipientEmail, recipientName, message string) (*domain.GiftCertificate, error) {
			return &domain.GiftCertificate{
				ID:             certID,
				Code:           "BANI-ABCD-1234",
				PurchaserEmail: purchaserEmail,
				RecipientEmail: recipientEmail,
				RecipientName:  recipientName,
				Amount:         amount,
				Balance:        amount,
				Message:        message,
				Status:         domain.CertificateStatusActive,
				ValidUntil:     now.AddDate(0, 0, 365),
				CreatedAt:      now,
			}, nil
		},
	}

	h := NewCertificateHandler(svc)

	body, _ := json.Marshal(map[string]interface{}{
		"amount":          500000,
		"purchaser_email": "buyer@example.com",
		"recipient_email": "gift@example.com",
		"recipient_name":  "Иван",
		"message":         "С днём рождения!",
	})

	req := httptest.NewRequest(http.MethodPost, "/certificates/purchase", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	r := chi.NewRouter()
	r.Post("/certificates/purchase", h.Purchase)
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d: %s", rec.Code, rec.Body.String())
		return
	}

	var resp APIResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if !resp.Success {
		t.Error("expected success=true")
	}

	data, ok := resp.Data.(map[string]interface{})
	if !ok {
		t.Fatal("expected data to be a map")
	}

	if data["code"] != "BANI-ABCD-1234" {
		t.Errorf("expected code BANI-ABCD-1234, got %v", data["code"])
	}
	if data["amount"].(float64) != 500000 {
		t.Errorf("expected amount 500000, got %v", data["amount"])
	}
}

func TestCertificateHandler_Purchase_InvalidInput(t *testing.T) {
	svc := &mockCertificateService{
		purchaseFn: func(ctx context.Context, amount int64, purchaserID *uuid.UUID, purchaserEmail, recipientEmail, recipientName, message string) (*domain.GiftCertificate, error) {
			return nil, domain.ErrInvalidInput
		},
	}

	h := NewCertificateHandler(svc)

	body, _ := json.Marshal(map[string]interface{}{
		"amount": 0,
	})

	req := httptest.NewRequest(http.MethodPost, "/certificates/purchase", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	r := chi.NewRouter()
	r.Post("/certificates/purchase", h.Purchase)
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

func TestCertificateHandler_Purchase_WithAuth(t *testing.T) {
	userID := uuid.New()
	var capturedPurchaserID *uuid.UUID

	svc := &mockCertificateService{
		purchaseFn: func(ctx context.Context, amount int64, purchaserID *uuid.UUID, purchaserEmail, recipientEmail, recipientName, message string) (*domain.GiftCertificate, error) {
			capturedPurchaserID = purchaserID
			return &domain.GiftCertificate{
				ID:             uuid.New(),
				Code:           "BANI-TEST-1234",
				PurchaserID:    purchaserID,
				PurchaserEmail: purchaserEmail,
				RecipientEmail: recipientEmail,
				Amount:         amount,
				Balance:        amount,
				Status:         domain.CertificateStatusActive,
				ValidUntil:     time.Now().AddDate(0, 0, 365),
				CreatedAt:      time.Now(),
			}, nil
		},
	}

	h := NewCertificateHandler(svc)
	authService := &mockAuthService{userID: userID, role: domain.RoleClient}

	body, _ := json.Marshal(map[string]interface{}{
		"amount":          100000,
		"purchaser_email": "buyer@example.com",
		"recipient_email": "gift@example.com",
		"recipient_name":  "Друг",
	})

	req := httptest.NewRequest(http.MethodPost, "/certificates/purchase", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer valid-token")
	rec := httptest.NewRecorder()

	r := chi.NewRouter()
	r.Use(middleware.OptionalAuth(authService))
	r.Post("/certificates/purchase", h.Purchase)
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d: %s", rec.Code, rec.Body.String())
		return
	}

	if capturedPurchaserID == nil || *capturedPurchaserID != userID {
		t.Error("expected purchaser ID to be set from auth context")
	}
}

func TestCertificateHandler_Redeem(t *testing.T) {
	userID := uuid.New()
	certID := uuid.New()

	svc := &mockCertificateService{
		redeemFn: func(ctx context.Context, code string, uid uuid.UUID) (*domain.GiftCertificate, error) {
			if code == "BANI-ABCD-1234" && uid == userID {
				return &domain.GiftCertificate{
					ID:           certID,
					Code:         code,
					Amount:       500000,
					Balance:      500000,
					Status:       domain.CertificateStatusActive,
					RedeemedByID: &userID,
					ValidUntil:   time.Now().AddDate(0, 0, 365),
					CreatedAt:    time.Now(),
				}, nil
			}
			return nil, domain.ErrCertificateNotFound
		},
	}

	h := NewCertificateHandler(svc)
	authService := &mockAuthService{userID: userID, role: domain.RoleClient}

	body, _ := json.Marshal(map[string]interface{}{
		"code": "BANI-ABCD-1234",
	})

	req := httptest.NewRequest(http.MethodPost, "/certificates/redeem", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer valid-token")
	rec := httptest.NewRecorder()

	r := chi.NewRouter()
	r.Use(middleware.RequireAuth(authService))
	r.Post("/certificates/redeem", h.Redeem)
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
		return
	}

	var resp APIResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if !resp.Success {
		t.Error("expected success=true")
	}
}

func TestCertificateHandler_Redeem_Expired(t *testing.T) {
	userID := uuid.New()

	svc := &mockCertificateService{
		redeemFn: func(ctx context.Context, code string, uid uuid.UUID) (*domain.GiftCertificate, error) {
			return nil, domain.ErrCertificateExpired
		},
	}

	h := NewCertificateHandler(svc)
	authService := &mockAuthService{userID: userID, role: domain.RoleClient}

	body, _ := json.Marshal(map[string]interface{}{
		"code": "BANI-EXPR-0000",
	})

	req := httptest.NewRequest(http.MethodPost, "/certificates/redeem", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer valid-token")
	rec := httptest.NewRecorder()

	r := chi.NewRouter()
	r.Use(middleware.RequireAuth(authService))
	r.Post("/certificates/redeem", h.Redeem)
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

func TestCertificateHandler_Redeem_Unauthorized(t *testing.T) {
	svc := &mockCertificateService{}
	h := NewCertificateHandler(svc)
	authService := &mockAuthService{err: domain.ErrUnauthorized}

	body, _ := json.Marshal(map[string]interface{}{
		"code": "BANI-ABCD-1234",
	})

	req := httptest.NewRequest(http.MethodPost, "/certificates/redeem", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	r := chi.NewRouter()
	r.Use(middleware.RequireAuth(authService))
	r.Post("/certificates/redeem", h.Redeem)
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", rec.Code)
	}
}

func TestCertificateHandler_GetBalance(t *testing.T) {
	svc := &mockCertificateService{
		getBalanceFn: func(ctx context.Context, code string) (*domain.GiftCertificate, error) {
			if code == "BANI-ABCD-1234" {
				return &domain.GiftCertificate{
					ID:       uuid.New(),
					Code:     code,
					Amount:   500000,
					Balance:  300000,
					Status:   domain.CertificateStatusActive,
					ValidUntil: time.Date(2027, 1, 1, 0, 0, 0, 0, time.UTC),
				}, nil
			}
			return nil, domain.ErrCertificateNotFound
		},
	}

	h := NewCertificateHandler(svc)

	req := httptest.NewRequest(http.MethodGet, "/certificates/BANI-ABCD-1234/balance", nil)
	rec := httptest.NewRecorder()

	r := chi.NewRouter()
	r.Get("/certificates/{code}/balance", h.GetBalance)
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
		return
	}

	var resp APIResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if !resp.Success {
		t.Error("expected success=true")
	}

	data, ok := resp.Data.(map[string]interface{})
	if !ok {
		t.Fatal("expected data to be a map")
	}

	if data["code"] != "BANI-ABCD-1234" {
		t.Errorf("expected code BANI-ABCD-1234, got %v", data["code"])
	}
	if data["balance"].(float64) != 300000 {
		t.Errorf("expected balance 300000, got %v", data["balance"])
	}
	if data["amount"].(float64) != 500000 {
		t.Errorf("expected amount 500000, got %v", data["amount"])
	}
}

func TestCertificateHandler_GetBalance_NotFound(t *testing.T) {
	svc := &mockCertificateService{
		getBalanceFn: func(ctx context.Context, code string) (*domain.GiftCertificate, error) {
			return nil, domain.ErrCertificateNotFound
		},
	}

	h := NewCertificateHandler(svc)

	req := httptest.NewRequest(http.MethodGet, "/certificates/BANI-XXXX-9999/balance", nil)
	rec := httptest.NewRecorder()

	r := chi.NewRouter()
	r.Get("/certificates/{code}/balance", h.GetBalance)
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", rec.Code)
	}
}

func TestCertificateHandler_ListMyCertificates(t *testing.T) {
	userID := uuid.New()

	svc := &mockCertificateService{
		listByUserFn: func(ctx context.Context, uid uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.GiftCertificate], error) {
			if uid == userID {
				return &domain.PaginatedResult[domain.GiftCertificate]{
					Items: []domain.GiftCertificate{
						{
							ID:             uuid.New(),
							Code:           "BANI-AAAA-1111",
							Amount:         100000,
							Balance:        100000,
							Status:         domain.CertificateStatusActive,
							PurchaserEmail: "me@example.com",
							RecipientEmail: "friend@example.com",
							ValidUntil:     time.Now().AddDate(0, 0, 365),
							CreatedAt:      time.Now(),
						},
						{
							ID:             uuid.New(),
							Code:           "BANI-BBBB-2222",
							Amount:         200000,
							Balance:        0,
							Status:         domain.CertificateStatusUsed,
							PurchaserEmail: "someone@example.com",
							RecipientEmail: "me@example.com",
							ValidUntil:     time.Now().AddDate(0, 0, 365),
							CreatedAt:      time.Now(),
						},
					},
					TotalCount: 2,
				}, nil
			}
			return nil, domain.ErrNotFound
		},
	}

	h := NewCertificateHandler(svc)
	authService := &mockAuthService{userID: userID, role: domain.RoleClient}

	req := httptest.NewRequest(http.MethodGet, "/my/certificates?page=1&page_size=10", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rec := httptest.NewRecorder()

	r := chi.NewRouter()
	r.Use(middleware.RequireAuth(authService))
	r.Get("/my/certificates", h.ListMyCertificates)
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
		return
	}

	var resp APIResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if !resp.Success {
		t.Error("expected success=true")
	}

	data, ok := resp.Data.([]interface{})
	if !ok {
		t.Fatal("expected data to be a list")
	}

	if len(data) != 2 {
		t.Errorf("expected 2 certificates, got %d", len(data))
	}

	if resp.Meta == nil {
		t.Fatal("expected meta to be present")
	}
}

func TestCertificateHandler_ListMyCertificates_Unauthorized(t *testing.T) {
	svc := &mockCertificateService{}
	h := NewCertificateHandler(svc)
	authService := &mockAuthService{err: domain.ErrUnauthorized}

	req := httptest.NewRequest(http.MethodGet, "/my/certificates", nil)
	rec := httptest.NewRecorder()

	r := chi.NewRouter()
	r.Use(middleware.RequireAuth(authService))
	r.Get("/my/certificates", h.ListMyCertificates)
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", rec.Code)
	}
}
