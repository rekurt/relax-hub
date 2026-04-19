package fiscal_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/fiscal"
	"github.com/nikitaaldaev/bani/internal/logger"
)

// newMockATOLServer creates a httptest server that simulates ATOL Online API v4.
func newMockATOLServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch {
		case r.URL.Path == "/getToken" && r.Method == http.MethodPost:
			json.NewEncoder(w).Encode(map[string]string{"token": "mock-token"})

		case strings.HasSuffix(r.URL.Path, "/sell") && r.Method == http.MethodPost:
			json.NewEncoder(w).Encode(map[string]interface{}{
				"uuid":   "mock-uuid-sell",
				"status": "pending",
			})

		case strings.HasSuffix(r.URL.Path, "/sell_refund") && r.Method == http.MethodPost:
			json.NewEncoder(w).Encode(map[string]interface{}{
				"uuid":   "mock-uuid-refund",
				"status": "pending",
			})

		default:
			http.Error(w, `{"error":{"code":999,"text":"unknown endpoint"}}`, http.StatusBadRequest)
		}
	}))
}

// newTestATOLProvider creates an ATOLProvider pointed at the mock server.
func newTestATOLProvider(t *testing.T, server *httptest.Server) *fiscal.ATOLProvider {
	t.Helper()
	log := logger.New(logger.LevelWarn)
	provider := fiscal.NewATOLProvider("test_login", "test_pass", "test_group", log)
	provider.SetBaseURL(server.URL)
	return provider
}

func TestATOLProvider_CreateReceipt(t *testing.T) {
	server := newMockATOLServer(t)
	defer server.Close()
	provider := newTestATOLProvider(t, server)

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
	if receipt.ExternalID != "mock-uuid-sell" {
		t.Errorf("expected external ID 'mock-uuid-sell', got %q", receipt.ExternalID)
	}
}

func TestATOLProvider_CreateReceipt_RefundType(t *testing.T) {
	server := newMockATOLServer(t)
	defer server.Close()
	provider := newTestATOLProvider(t, server)

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
	if receipt.ExternalID != "mock-uuid-refund" {
		t.Errorf("expected external ID 'mock-uuid-refund', got %q", receipt.ExternalID)
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
	server := newMockATOLServer(t)
	defer server.Close()
	provider := newTestATOLProvider(t, server)

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
