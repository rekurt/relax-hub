package payment

import (
	"context"
	"testing"
)

func TestMockProvider_CreatePayment(t *testing.T) {
	provider := NewMockProvider()
	ctx := context.Background()

	result, err := provider.CreatePayment(ctx, 10050, "RUB", "Test payment", "https://example.com/return", map[string]string{"booking_id": "123"})
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

	result, _ := provider.CreatePayment(ctx, 10000, "RUB", "Test", "https://example.com", nil)

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

func TestMockProvider_MultiplePayments(t *testing.T) {
	provider := NewMockProvider()
	ctx := context.Background()

	result1, _ := provider.CreatePayment(ctx, 5000, "RUB", "Payment 1", "https://example.com", nil)
	result2, _ := provider.CreatePayment(ctx, 7500, "RUB", "Payment 2", "https://example.com", nil)

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

// Verify MockProvider implements PaymentProvider interface
var _ PaymentProvider = (*MockProvider)(nil)

// Verify YooKassaProvider implements PaymentProvider interface
var _ PaymentProvider = (*YooKassaProvider)(nil)
