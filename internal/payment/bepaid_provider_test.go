package payment

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestBePaidProvider_CreatePayment(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify basic auth
		user, pass, ok := r.BasicAuth()
		if !ok || user != "test-shop" || pass != "test-secret" {
			t.Errorf("invalid basic auth: user=%s, pass=%s, ok=%v", user, pass, ok)
		}

		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}

		// Decode request body
		var req bepaidCheckoutRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}

		if req.Checkout.Order.Amount != 10050 {
			t.Errorf("expected amount 10050, got %d", req.Checkout.Order.Amount)
		}
		if req.Checkout.Order.Currency != "BYN" {
			t.Errorf("expected currency BYN, got %s", req.Checkout.Order.Currency)
		}
		if req.Checkout.TransactionType != "payment" {
			t.Errorf("expected transaction_type payment, got %s", req.Checkout.TransactionType)
		}

		resp := bepaidCheckoutResponse{}
		resp.Checkout.Token = "test-token-123"
		resp.Checkout.RedirectURL = "https://checkout.bepaid.by/pay/test-token-123"

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	provider := NewBePaidProvider("test-shop", "test-secret")
	provider.client = server.Client()

	// Override URL by using a custom transport
	origURL := bepaidCheckoutURL
	// We need to point to test server - use a wrapper approach
	transport := &rewriteTransport{
		base:    http.DefaultTransport,
		baseURL: server.URL,
	}
	provider.client.Transport = transport

	result, err := provider.CreatePayment(context.Background(), CreatePaymentRequest{
		Amount:      10050,
		Currency:    "BYN",
		Description: "Test booking",
		ReturnURL:   "https://example.com/return",
		Metadata:    map[string]string{"booking_id": "abc-123"},
		Method:      "card",
		Capture:     true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.ExternalID != "test-token-123" {
		t.Errorf("expected external ID test-token-123, got %s", result.ExternalID)
	}
	if result.ConfirmationURL == "" {
		t.Error("expected non-empty confirmation URL")
	}

	_ = origURL // suppress unused warning
}

func TestBePaidProvider_CreatePayment_Authorization(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req bepaidCheckoutRequest
		json.NewDecoder(r.Body).Decode(&req)

		if req.Checkout.TransactionType != "authorization" {
			t.Errorf("expected transaction_type authorization, got %s", req.Checkout.TransactionType)
		}

		resp := bepaidCheckoutResponse{}
		resp.Checkout.Token = "auth-token-456"
		resp.Checkout.RedirectURL = "https://checkout.bepaid.by/pay/auth-token-456"
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	provider := NewBePaidProvider("test-shop", "test-secret")
	provider.client = &http.Client{
		Transport: &rewriteTransport{base: http.DefaultTransport, baseURL: server.URL},
	}

	result, err := provider.CreatePayment(context.Background(), CreatePaymentRequest{
		Amount:  5000,
		Currency: "BYN",
		Method:  "belkart",
		Capture: false,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ExternalID != "auth-token-456" {
		t.Errorf("expected external ID auth-token-456, got %s", result.ExternalID)
	}
}

func TestBePaidProvider_CreatePayment_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "internal error"}`))
	}))
	defer server.Close()

	provider := NewBePaidProvider("test-shop", "test-secret")
	provider.client = &http.Client{
		Transport: &rewriteTransport{base: http.DefaultTransport, baseURL: server.URL},
	}

	_, err := provider.CreatePayment(context.Background(), CreatePaymentRequest{
		Amount:   10050,
		Currency: "BYN",
		Method:   "card",
		Capture:  true,
	})
	if err == nil {
		t.Error("expected error for server error response")
	}
}

func TestBePaidProvider_GetPaymentStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}

		resp := bepaidTransactionResponse{}
		resp.Transaction.UID = "tx-123"
		resp.Transaction.Status = "successful"
		resp.Transaction.Amount = 10050

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	provider := NewBePaidProvider("test-shop", "test-secret")
	provider.client = &http.Client{
		Transport: &rewriteTransport{base: http.DefaultTransport, baseURL: server.URL},
	}

	status, err := provider.GetPaymentStatus(context.Background(), "tx-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if status != "succeeded" {
		t.Errorf("expected status succeeded, got %s", status)
	}
}

func TestBePaidProvider_CreateRefund(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"transaction": {"uid": "refund-1", "status": "successful"}}`))
	}))
	defer server.Close()

	provider := NewBePaidProvider("test-shop", "test-secret")
	provider.client = &http.Client{
		Transport: &rewriteTransport{base: http.DefaultTransport, baseURL: server.URL},
	}

	err := provider.CreateRefund(context.Background(), "tx-123", 5000)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestBePaidProvider_CapturePayment(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"transaction": {"uid": "cap-1", "status": "successful"}}`))
	}))
	defer server.Close()

	provider := NewBePaidProvider("test-shop", "test-secret")
	provider.client = &http.Client{
		Transport: &rewriteTransport{base: http.DefaultTransport, baseURL: server.URL},
	}

	err := provider.CapturePayment(context.Background(), "tx-123", 10050)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestBePaidProvider_CancelPayment(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"transaction": {"uid": "void-1", "status": "successful"}}`))
	}))
	defer server.Close()

	provider := NewBePaidProvider("test-shop", "test-secret")
	provider.client = &http.Client{
		Transport: &rewriteTransport{base: http.DefaultTransport, baseURL: server.URL},
	}

	err := provider.CancelPayment(context.Background(), "tx-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestMapBePaidStatus(t *testing.T) {
	tests := []struct {
		bepaid   string
		expected string
	}{
		{"successful", "succeeded"},
		{"failed", "failed"},
		{"expired", "failed"},
		{"pending", "pending"},
		{"incomplete", "pending"},
		{"authorized", "waiting_for_capture"},
		{"voided", "canceled"},
		{"unknown", "unknown"},
	}

	for _, tt := range tests {
		result := mapBePaidStatus(tt.bepaid)
		if result != tt.expected {
			t.Errorf("mapBePaidStatus(%q) = %q, want %q", tt.bepaid, result, tt.expected)
		}
	}
}

// Verify BePaidProvider implements PaymentProvider interface
var _ PaymentProvider = (*BePaidProvider)(nil)

// rewriteTransport rewrites request URLs to point to a test server.
type rewriteTransport struct {
	base    http.RoundTripper
	baseURL string
}

func (t *rewriteTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.URL.Scheme = "http"
	req.URL.Host = t.baseURL[len("http://"):]
	return t.base.RoundTrip(req)
}
