package payment

import (
	"context"
	"testing"
)

func TestMockProvider_CreatePayment(t *testing.T) {
	provider := NewMockProvider()
	ctx := context.Background()

	result, err := provider.CreatePayment(ctx, CreatePaymentRequest{
		Amount:      10050,
		Currency:    "RUB",
		Description: "Test payment",
		ReturnURL:   "https://example.com/return",
		Metadata:    map[string]string{"booking_id": "123"},
		Method:      "card",
		Capture:     true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.ExternalID == "" {
		t.Error("expected non-empty external ID")
	}
	if result.ConfirmationURL == "" {
		t.Error("expected non-empty confirmation URL")
	}

	// Verify status is pending after creation
	status, err := provider.GetPaymentStatus(ctx, result.ExternalID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if status != "pending" {
		t.Errorf("expected status pending, got %s", status)
	}
}

func TestMockProvider_CreatePayment_SBP(t *testing.T) {
	provider := NewMockProvider()
	ctx := context.Background()

	result, err := provider.CreatePayment(ctx, CreatePaymentRequest{
		Amount:      5000,
		Currency:    "RUB",
		Description: "SBP payment",
		ReturnURL:   "https://example.com/return",
		Method:      "sbp",
		Capture:     true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.ExternalID == "" {
		t.Error("expected non-empty external ID")
	}
	if result.ConfirmationURL == "" {
		t.Error("expected non-empty confirmation URL")
	}

	method := provider.GetLastPaymentMethod()
	if method != "sbp" {
		t.Errorf("expected method sbp, got %s", method)
	}
}

func TestMockProvider_GetPaymentStatus_NotFound(t *testing.T) {
	provider := NewMockProvider()
	ctx := context.Background()

	_, err := provider.GetPaymentStatus(ctx, "nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent payment")
	}
}

func TestMockProvider_CreateRefund(t *testing.T) {
	provider := NewMockProvider()
	ctx := context.Background()

	result, _ := provider.CreatePayment(ctx, CreatePaymentRequest{
		Amount:   10000,
		Currency: "RUB",
		Method:   "card",
		Capture:  true,
	})

	// Refund should fail for pending payment
	err := provider.CreateRefund(ctx, result.ExternalID, 10000)
	if err == nil {
		t.Error("expected error refunding pending payment")
	}

	// Set status to succeeded and retry
	provider.SetPaymentStatus(result.ExternalID, "succeeded")

	err = provider.CreateRefund(ctx, result.ExternalID, 10000)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestMockProvider_CreateRefund_NotFound(t *testing.T) {
	provider := NewMockProvider()
	ctx := context.Background()

	err := provider.CreateRefund(ctx, "nonexistent", 5000)
	if err == nil {
		t.Error("expected error for nonexistent payment")
	}
}

func TestMockProvider_CapturePayment(t *testing.T) {
	provider := NewMockProvider()
	ctx := context.Background()

	result, _ := provider.CreatePayment(ctx, CreatePaymentRequest{
		Amount:   10000,
		Currency: "RUB",
		Method:   "card",
		Capture:  false,
	})

	// Set to waiting_for_capture (simulating provider state)
	provider.SetPaymentStatus(result.ExternalID, "waiting_for_capture")

	err := provider.CapturePayment(ctx, result.ExternalID, 10000)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	status, _ := provider.GetPaymentStatus(ctx, result.ExternalID)
	if status != "succeeded" {
		t.Errorf("expected status succeeded after capture, got %s", status)
	}

	if !provider.WasCaptured(result.ExternalID) {
		t.Error("expected payment to be marked as captured")
	}
}

func TestMockProvider_CancelPayment(t *testing.T) {
	provider := NewMockProvider()
	ctx := context.Background()

	result, _ := provider.CreatePayment(ctx, CreatePaymentRequest{
		Amount:   10000,
		Currency: "RUB",
		Method:   "card",
		Capture:  false,
	})

	err := provider.CancelPayment(ctx, result.ExternalID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	status, _ := provider.GetPaymentStatus(ctx, result.ExternalID)
	if status != "canceled" {
		t.Errorf("expected status canceled after cancel, got %s", status)
	}

	if !provider.WasCancelled(result.ExternalID) {
		t.Error("expected payment to be marked as cancelled")
	}
}

func TestMockProvider_MultiplePayments(t *testing.T) {
	provider := NewMockProvider()
	ctx := context.Background()

	result1, _ := provider.CreatePayment(ctx, CreatePaymentRequest{Amount: 5000, Currency: "RUB", Method: "card", Capture: true})
	result2, _ := provider.CreatePayment(ctx, CreatePaymentRequest{Amount: 7500, Currency: "RUB", Method: "sbp", Capture: true})

	if result1.ExternalID == result2.ExternalID {
		t.Error("expected unique external IDs for different payments")
	}

	// Both should have independent statuses
	provider.SetPaymentStatus(result1.ExternalID, "succeeded")

	s1, _ := provider.GetPaymentStatus(ctx, result1.ExternalID)
	s2, _ := provider.GetPaymentStatus(ctx, result2.ExternalID)

	if s1 != "succeeded" {
		t.Errorf("expected succeeded for payment 1, got %s", s1)
	}
	if s2 != "pending" {
		t.Errorf("expected pending for payment 2, got %s", s2)
	}
}

func TestKopecksToString(t *testing.T) {
	tests := []struct {
		kopecks  int64
		expected string
	}{
		{10050, "100.50"},
		{500, "5.00"},
		{100, "1.00"},
		{1, "0.01"},
		{99, "0.99"},
		{10000, "100.00"},
		{123456, "1234.56"},
	}

	for _, tt := range tests {
		result := kopecksToString(tt.kopecks)
		if result != tt.expected {
			t.Errorf("kopecksToString(%d) = %s, want %s", tt.kopecks, result, tt.expected)
		}
	}
}

func TestMockProvider_CreatePayment_ApplePay(t *testing.T) {
	provider := NewMockProvider()
	ctx := context.Background()

	result, err := provider.CreatePayment(ctx, CreatePaymentRequest{
		Amount:       5000,
		Currency:     "RUB",
		Description:  "Apple Pay payment",
		ReturnURL:    "https://example.com/return",
		Method:       "apple_pay",
		Capture:      true,
		PaymentToken: "apple-pay-token-data",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.ExternalID == "" {
		t.Error("expected non-empty external ID")
	}

	method := provider.GetLastPaymentMethod()
	if method != "apple_pay" {
		t.Errorf("expected method apple_pay, got %s", method)
	}
}

func TestMockProvider_CreatePayment_GooglePay(t *testing.T) {
	provider := NewMockProvider()
	ctx := context.Background()

	result, err := provider.CreatePayment(ctx, CreatePaymentRequest{
		Amount:       7500,
		Currency:     "RUB",
		Description:  "Google Pay payment",
		ReturnURL:    "https://example.com/return",
		Method:       "google_pay",
		Capture:      true,
		PaymentToken: "google-pay-token-data",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.ExternalID == "" {
		t.Error("expected non-empty external ID")
	}

	method := provider.GetLastPaymentMethod()
	if method != "google_pay" {
		t.Errorf("expected method google_pay, got %s", method)
	}
}

func TestCreatePaymentRequest_PaymentToken(t *testing.T) {
	req := CreatePaymentRequest{
		Amount:       10000,
		Currency:     "RUB",
		Method:       "apple_pay",
		Capture:      true,
		PaymentToken: "test-token",
	}

	if req.PaymentToken != "test-token" {
		t.Errorf("expected payment token 'test-token', got '%s'", req.PaymentToken)
	}
}

// Verify MockProvider implements PaymentProvider interface
var _ PaymentProvider = (*MockProvider)(nil)

// Verify YooKassaProvider implements PaymentProvider interface
var _ PaymentProvider = (*YooKassaProvider)(nil)
