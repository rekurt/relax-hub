package fiscal_test

import (
	"context"
	"testing"

	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/fiscal"
	"github.com/nikitaaldaev/bani/internal/logger"
)

func TestATOLProvider_CreateReceipt(t *testing.T) {
	log := logger.New(logger.LevelWarn)
	provider := fiscal.NewATOLProvider("test_login", "test_pass", "test_group", log)

	req := fiscal.ReceiptRequest{
		Type:   fiscal.ReceiptAdvance,
		Amount: 10000,
		Email:  "test@example.com",
		Phone:  "+79001234567",
		Items: []fiscal.ReceiptItem{
			{
				Name:     "Бронирование бани",
				Quantity: 1,
				Price:    10000,
				VAT:      "none",
			},
		},
	}

	receipt, err := provider.CreateReceipt(context.Background(), req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if receipt == nil {
		t.Fatal("expected receipt, got nil")
	}
	if receipt.ID == "" {
		t.Error("expected receipt ID to be non-empty")
	}
	if receipt.Status != "pending" {
		t.Errorf("expected status 'pending', got %q", receipt.Status)
	}
}

func TestATOLProvider_CreateReceipt_RefundType(t *testing.T) {
	log := logger.New(logger.LevelWarn)
	provider := fiscal.NewATOLProvider("login", "pass", "group", log)

	req := fiscal.ReceiptRequest{
		Type:   fiscal.ReceiptRefund,
		Amount: 5000,
		Items: []fiscal.ReceiptItem{
			{Name: "Возврат", Quantity: 1, Price: 5000, VAT: "none"},
		},
	}

	receipt, err := provider.CreateReceipt(context.Background(), req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if receipt == nil {
		t.Fatal("expected receipt, got nil")
	}
	if receipt.ID == "" {
		t.Error("expected receipt ID to be non-empty")
	}
}

func TestNoOpProvider_CreateReceipt(t *testing.T) {
	provider := fiscal.NewNoOpProvider()

	req := fiscal.ReceiptRequest{
		Type:   fiscal.ReceiptAdvance,
		Amount: 10000,
		Items: []fiscal.ReceiptItem{
			{Name: "Test", Quantity: 1, Price: 10000, VAT: "none"},
		},
	}

	receipt, err := provider.CreateReceipt(context.Background(), req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if receipt != nil {
		t.Errorf("expected nil receipt from no-op provider, got %+v", receipt)
	}
}

func TestFiscalProvider_Interface(t *testing.T) {
	// Verify both providers implement FiscalProvider
	var _ fiscal.FiscalProvider = fiscal.NewATOLProvider("", "", "", logger.New(logger.LevelWarn))
	var _ fiscal.FiscalProvider = fiscal.NewNoOpProvider()
}

func TestTaxInfoForEntityType(t *testing.T) {
	tests := []struct {
		name       string
		entityType domain.KYCEntityType
		wantVAT    string
		wantTax    fiscal.TaxSystem
	}{
		{
			name:       "legal entity uses OSN with VAT 20%",
			entityType: domain.KYCEntityLegalEntity,
			wantVAT:    "vat20",
			wantTax:    fiscal.TaxSystemOSN,
		},
		{
			name:       "sole proprietor uses OSN with VAT 20%",
			entityType: domain.KYCEntitySoleProprietor,
			wantVAT:    "vat20",
			wantTax:    fiscal.TaxSystemOSN,
		},
		{
			name:       "self-employed uses NPD without VAT",
			entityType: domain.KYCEntitySelfEmployed,
			wantVAT:    "none",
			wantTax:    fiscal.TaxSystemNPD,
		},
		{
			name:       "individual uses OSN without VAT",
			entityType: domain.KYCEntityIndividual,
			wantVAT:    "none",
			wantTax:    fiscal.TaxSystemOSN,
		},
		{
			name:       "unknown entity type uses defaults",
			entityType: domain.KYCEntityType("unknown"),
			wantVAT:    "none",
			wantTax:    fiscal.TaxSystemDefault,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vat, taxSystem := fiscal.TaxInfoForEntityType(tt.entityType)
			if vat != tt.wantVAT {
				t.Errorf("VAT = %q, want %q", vat, tt.wantVAT)
			}
			if taxSystem != tt.wantTax {
				t.Errorf("TaxSystem = %q, want %q", taxSystem, tt.wantTax)
			}
		})
	}
}

func TestATOLProvider_CreateReceipt_WithTaxSystem(t *testing.T) {
	log := logger.New(logger.LevelWarn)
	provider := fiscal.NewATOLProvider("login", "pass", "group", log)

	req := fiscal.ReceiptRequest{
		Type:      fiscal.ReceiptAdvance,
		Amount:    10000,
		TaxSystem: fiscal.TaxSystemNPD,
		Items: []fiscal.ReceiptItem{
			{Name: "Бронирование бани", Quantity: 1, Price: 10000, VAT: "none"},
		},
	}

	receipt, err := provider.CreateReceipt(context.Background(), req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if receipt == nil {
		t.Fatal("expected receipt, got nil")
	}
	if receipt.Status != "pending" {
		t.Errorf("expected status 'pending', got %q", receipt.Status)
	}
}
