package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/rekurt/relax-hub/internal/domain"
	"github.com/rekurt/relax-hub/internal/middleware"
	"github.com/rekurt/relax-hub/internal/service"
)

type mockPaymentDetailsService struct {
	setFn      func(ctx context.Context, userID uuid.UUID, input service.SetPaymentDetailsInput) (*domain.PaymentDetails, error)
	getFn      func(ctx context.Context, userID uuid.UUID) (*domain.PaymentDetails, error)
	validateFn func(ctx context.Context, userID uuid.UUID) error
}

func (m *mockPaymentDetailsService) Set(ctx context.Context, userID uuid.UUID, input service.SetPaymentDetailsInput) (*domain.PaymentDetails, error) {
	if m.setFn != nil {
		return m.setFn(ctx, userID, input)
	}
	return nil, nil
}

func (m *mockPaymentDetailsService) Get(ctx context.Context, userID uuid.UUID) (*domain.PaymentDetails, error) {
	if m.getFn != nil {
		return m.getFn(ctx, userID)
	}
	return nil, nil
}

func (m *mockPaymentDetailsService) Validate(ctx context.Context, userID uuid.UUID) error {
	if m.validateFn != nil {
		return m.validateFn(ctx, userID)
	}
	return nil
}

func TestPaymentDetailsHandler_Set_Success(t *testing.T) {
	userID := uuid.New()
	now := time.Now()

	svc := &mockPaymentDetailsService{
		setFn: func(_ context.Context, uid uuid.UUID, input service.SetPaymentDetailsInput) (*domain.PaymentDetails, error) {
			return &domain.PaymentDetails{
				ID:             uuid.New(),
				UserID:         uid,
				EntityType:     input.EntityType,
				BankCardNumber: input.BankCardNumber,
				CardHolderName: input.CardHolderName,
				IsVerified:     false,
				CreatedAt:      now,
				UpdatedAt:      now,
			}, nil
		},
	}

	h := NewPaymentDetailsHandler(svc)

	body, _ := json.Marshal(map[string]string{
		"entity_type":      "individual",
		"bank_card_number": "4111111111111111",
		"card_holder_name": "Иванов Иван",
	})

	req := httptest.NewRequest(http.MethodPut, "/api/v1/my/payment-details", bytes.NewReader(body))
	req = req.WithContext(middleware.SetUserIDForTesting(req.Context(), userID))
	w := httptest.NewRecorder()

	h.SetPaymentDetails(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}

	var resp APIResponse
	json.NewDecoder(w.Body).Decode(&resp)
	if !resp.Success {
		t.Error("expected success = true")
	}
}

func TestPaymentDetailsHandler_Set_InvalidEntityType(t *testing.T) {
	userID := uuid.New()
	svc := &mockPaymentDetailsService{}
	h := NewPaymentDetailsHandler(svc)

	body, _ := json.Marshal(map[string]string{
		"entity_type": "invalid_type",
	})

	req := httptest.NewRequest(http.MethodPut, "/api/v1/my/payment-details", bytes.NewReader(body))
	req = req.WithContext(middleware.SetUserIDForTesting(req.Context(), userID))
	w := httptest.NewRecorder()

	h.SetPaymentDetails(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestPaymentDetailsHandler_Set_ValidationError(t *testing.T) {
	userID := uuid.New()

	svc := &mockPaymentDetailsService{
		setFn: func(_ context.Context, _ uuid.UUID, _ service.SetPaymentDetailsInput) (*domain.PaymentDetails, error) {
			return nil, domain.ErrInvalidInput
		},
	}

	h := NewPaymentDetailsHandler(svc)

	body, _ := json.Marshal(map[string]string{
		"entity_type":      "individual",
		"bank_card_number": "short",
	})

	req := httptest.NewRequest(http.MethodPut, "/api/v1/my/payment-details", bytes.NewReader(body))
	req = req.WithContext(middleware.SetUserIDForTesting(req.Context(), userID))
	w := httptest.NewRecorder()

	h.SetPaymentDetails(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestPaymentDetailsHandler_Get_Success(t *testing.T) {
	userID := uuid.New()
	now := time.Now()

	svc := &mockPaymentDetailsService{
		getFn: func(_ context.Context, uid uuid.UUID) (*domain.PaymentDetails, error) {
			return &domain.PaymentDetails{
				ID:             uuid.New(),
				UserID:         uid,
				EntityType:     domain.KYCEntityIndividual,
				BankCardNumber: "4111111111111111",
				CardHolderName: "Иванов Иван",
				IsVerified:     false,
				CreatedAt:      now,
				UpdatedAt:      now,
			}, nil
		},
	}

	h := NewPaymentDetailsHandler(svc)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/my/payment-details", nil)
	req = req.WithContext(middleware.SetUserIDForTesting(req.Context(), userID))
	w := httptest.NewRecorder()

	h.GetPaymentDetails(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}

	var resp APIResponse
	json.NewDecoder(w.Body).Decode(&resp)
	if !resp.Success {
		t.Error("expected success = true")
	}

	data, _ := json.Marshal(resp.Data)
	var details paymentDetailsResponse
	json.Unmarshal(data, &details)

	// Card number should be masked
	if details.BankCardNumber != "4111****1111" {
		t.Errorf("card should be masked, got %v", details.BankCardNumber)
	}
}

func TestPaymentDetailsHandler_Get_NotFound(t *testing.T) {
	userID := uuid.New()

	svc := &mockPaymentDetailsService{
		getFn: func(_ context.Context, _ uuid.UUID) (*domain.PaymentDetails, error) {
			return nil, domain.ErrPaymentDetailsNotFound
		},
	}

	h := NewPaymentDetailsHandler(svc)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/my/payment-details", nil)
	req = req.WithContext(middleware.SetUserIDForTesting(req.Context(), userID))
	w := httptest.NewRecorder()

	h.GetPaymentDetails(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", w.Code, http.StatusNotFound)
	}
}
