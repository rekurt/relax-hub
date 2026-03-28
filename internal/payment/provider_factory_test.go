package payment

import (
	"context"
	"testing"
)

func TestProviderFactory_ProviderForRegion(t *testing.T) {
	ruProvider := NewMockProvider()
	byProvider := NewMockProvider()

	factory := NewProviderFactory(map[string]PaymentProvider{
		"RU": ruProvider,
		"BY": byProvider,
	})

	// RU region
	p, err := factory.ProviderForRegion("RU")
	if err != nil {
		t.Fatalf("unexpected error for RU: %v", err)
	}
	if p != ruProvider {
		t.Error("expected RU provider")
	}

	// BY region
	p, err = factory.ProviderForRegion("BY")
	if err != nil {
		t.Fatalf("unexpected error for BY: %v", err)
	}
	if p != byProvider {
		t.Error("expected BY provider")
	}

	// Unknown region
	_, err = factory.ProviderForRegion("KZ")
	if err == nil {
		t.Error("expected error for unknown region")
	}
}

func TestProviderFactory_ProviderName(t *testing.T) {
	factory := NewProviderFactory(map[string]PaymentProvider{})

	if name := factory.ProviderName("RU"); name != "yookassa" {
		t.Errorf("expected yookassa for RU, got %s", name)
	}
	if name := factory.ProviderName("BY"); name != "bepaid" {
		t.Errorf("expected bepaid for BY, got %s", name)
	}
	if name := factory.ProviderName(""); name != "yookassa" {
		t.Errorf("expected yookassa for empty, got %s", name)
	}
}

func TestProviderFactory_DefaultFallback(t *testing.T) {
	ruProvider := NewMockProvider()
	factory := NewProviderFactory(map[string]PaymentProvider{
		"RU": ruProvider,
	})

	ctx := context.Background()

	// Factory implements PaymentProvider using RU as fallback
	result, err := factory.CreatePayment(ctx, CreatePaymentRequest{
		Amount:   5000,
		Currency: "RUB",
		Method:   "card",
		Capture:  true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ExternalID == "" {
		t.Error("expected non-empty external ID from fallback")
	}

	// Verify the payment was created in the RU mock
	status, err := ruProvider.GetPaymentStatus(ctx, result.ExternalID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if status != "pending" {
		t.Errorf("expected pending, got %s", status)
	}
}

func TestProviderFactory_NoFallback(t *testing.T) {
	factory := NewProviderFactory(map[string]PaymentProvider{
		"BY": NewMockProvider(),
	})

	ctx := context.Background()

	// No RU provider = no fallback
	_, err := factory.CreatePayment(ctx, CreatePaymentRequest{
		Amount:   5000,
		Currency: "BYN",
		Method:   "card",
		Capture:  true,
	})
	if err == nil {
		t.Error("expected error when no fallback provider")
	}
}

func TestProviderFactory_AllMethodsFallback(t *testing.T) {
	ruProvider := NewMockProvider()
	factory := NewProviderFactory(map[string]PaymentProvider{
		"RU": ruProvider,
	})

	ctx := context.Background()

	// Create a payment to test other methods
	result, _ := factory.CreatePayment(ctx, CreatePaymentRequest{
		Amount: 5000, Currency: "RUB", Method: "card", Capture: false,
	})

	// GetPaymentStatus
	status, err := factory.GetPaymentStatus(ctx, result.ExternalID)
	if err != nil {
		t.Fatalf("GetPaymentStatus error: %v", err)
	}
	if status != "pending" {
		t.Errorf("expected pending, got %s", status)
	}

	// CapturePayment
	ruProvider.SetPaymentStatus(result.ExternalID, "waiting_for_capture")
	if err := factory.CapturePayment(ctx, result.ExternalID, 5000); err != nil {
		t.Fatalf("CapturePayment error: %v", err)
	}

	// CreateRefund
	if err := factory.CreateRefund(ctx, result.ExternalID, 2500); err != nil {
		t.Fatalf("CreateRefund error: %v", err)
	}

	// Create another payment for cancel test
	result2, _ := factory.CreatePayment(ctx, CreatePaymentRequest{
		Amount: 3000, Currency: "RUB", Method: "card", Capture: false,
	})
	if err := factory.CancelPayment(ctx, result2.ExternalID); err != nil {
		t.Fatalf("CancelPayment error: %v", err)
	}
}

// Verify ProviderFactory implements PaymentProvider interface
var _ PaymentProvider = (*ProviderFactory)(nil)
