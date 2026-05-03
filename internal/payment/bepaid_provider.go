package payment

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	bepaidBaseURL        = "https://checkout.bepaid.by/ctp/api"
	bepaidCheckoutURL    = "https://checkout.bepaid.by/ctp/api/checkouts"
	bepaidDefaultTimeout = 30 * time.Second
)

// BePaidProvider implements PaymentProvider using the bePaid payment gateway (Belarus).
type BePaidProvider struct {
	shopID    string
	secretKey string
	client    *http.Client
}

// NewBePaidProvider creates a new bePaid provider with the given credentials.
func NewBePaidProvider(shopID, secretKey string) *BePaidProvider {
	return &BePaidProvider{
		shopID:    shopID,
		secretKey: secretKey,
		client: &http.Client{
			Timeout: bepaidDefaultTimeout,
		},
	}
}

// bePaid API request/response types

type bepaidCheckoutRequest struct {
	Checkout bepaidCheckout `json:"checkout"`
}

type bepaidCheckout struct {
	Test            bool                 `json:"test,omitempty"`
	TransactionType string               `json:"transaction_type"`
	Order           bepaidOrder          `json:"order"`
	Settings        bepaidSettings       `json:"settings"`
	PaymentMethod   *bepaidPaymentMethod `json:"payment_method,omitempty"`
}

type bepaidOrder struct {
	Amount      int64             `json:"amount"`
	Currency    string            `json:"currency"`
	Description string            `json:"description"`
	TrackingID  string            `json:"tracking_id"`
	Additional  map[string]string `json:"additional_data,omitempty"`
}

type bepaidSettings struct {
	ReturnURL  string `json:"return_url"`
	SuccessURL string `json:"success_url,omitempty"`
	DeclineURL string `json:"decline_url,omitempty"`
	FailURL    string `json:"fail_url,omitempty"`
	CancelURL  string `json:"cancel_url,omitempty"`
	NotifyURL  string `json:"notification_url,omitempty"`
	AutoReturn int    `json:"auto_return,omitempty"`
}

type bepaidPaymentMethod struct {
	Type string `json:"type"`
}

type bepaidCheckoutResponse struct {
	Checkout struct {
		Token       string `json:"token"`
		RedirectURL string `json:"redirect_url"`
	} `json:"checkout"`
}

type bepaidTransactionResponse struct {
	Transaction struct {
		UID    string `json:"uid"`
		Status string `json:"status"`
		Amount int64  `json:"amount"`
	} `json:"transaction"`
}

// CreatePayment creates a payment checkout in bePaid and returns the redirect URL.
func (p *BePaidProvider) CreatePayment(ctx context.Context, req CreatePaymentRequest) (*PaymentResult, error) {
	txType := "payment"
	if !req.Capture {
		txType = "authorization"
	}

	checkoutReq := bepaidCheckoutRequest{
		Checkout: bepaidCheckout{
			TransactionType: txType,
			Order: bepaidOrder{
				Amount:      req.Amount,
				Currency:    req.Currency,
				Description: req.Description,
				TrackingID:  req.Metadata["booking_id"],
				Additional:  req.Metadata,
			},
			Settings: bepaidSettings{
				ReturnURL: req.ReturnURL,
			},
		},
	}

	// Set payment method type if specified
	switch req.Method {
	case "belkart":
		checkoutReq.Checkout.PaymentMethod = &bepaidPaymentMethod{Type: "belkart"}
	case "erip":
		checkoutReq.Checkout.PaymentMethod = &bepaidPaymentMethod{Type: "erip"}
	}

	body, err := json.Marshal(checkoutReq)
	if err != nil {
		return nil, fmt.Errorf("bepaid marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, bepaidCheckoutURL, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("bepaid create request: %w", err)
	}

	httpReq.SetBasicAuth(p.shopID, p.secretKey)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("bepaid http request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("bepaid read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("bepaid create payment: status %d, body: %s", resp.StatusCode, string(respBody))
	}

	var checkoutResp bepaidCheckoutResponse
	if err := json.Unmarshal(respBody, &checkoutResp); err != nil {
		return nil, fmt.Errorf("bepaid unmarshal response: %w", err)
	}

	return &PaymentResult{
		ExternalID:      checkoutResp.Checkout.Token,
		ConfirmationURL: checkoutResp.Checkout.RedirectURL,
	}, nil
}

// GetPaymentStatus returns the current status of a payment by its external ID (token).
func (p *BePaidProvider) GetPaymentStatus(ctx context.Context, externalID string) (string, error) {
	url := fmt.Sprintf("%s/transactions/%s", bepaidBaseURL, externalID)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("bepaid create request: %w", err)
	}

	httpReq.SetBasicAuth(p.shopID, p.secretKey)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("bepaid http request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("bepaid read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("bepaid get status: status %d, body: %s", resp.StatusCode, string(respBody))
	}

	var txResp bepaidTransactionResponse
	if err := json.Unmarshal(respBody, &txResp); err != nil {
		return "", fmt.Errorf("bepaid unmarshal response: %w", err)
	}

	return mapBePaidStatus(txResp.Transaction.Status), nil
}

// CreateRefund creates a refund for a payment in bePaid.
func (p *BePaidProvider) CreateRefund(ctx context.Context, externalID string, amount int64) error {
	reqBody := map[string]interface{}{
		"transaction": map[string]interface{}{
			"parent_uid": externalID,
			"amount":     amount,
		},
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("bepaid marshal refund: %w", err)
	}

	url := fmt.Sprintf("%s/transactions/refunds", bepaidBaseURL)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("bepaid create refund request: %w", err)
	}

	httpReq.SetBasicAuth(p.shopID, p.secretKey)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("bepaid http refund: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("bepaid refund: status %d, body: %s", resp.StatusCode, string(respBody))
	}

	return nil
}

// CapturePayment captures a previously authorized payment hold in bePaid.
func (p *BePaidProvider) CapturePayment(ctx context.Context, externalID string, amount int64) error {
	reqBody := map[string]interface{}{
		"transaction": map[string]interface{}{
			"parent_uid": externalID,
			"amount":     amount,
		},
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("bepaid marshal capture: %w", err)
	}

	url := fmt.Sprintf("%s/transactions/captures", bepaidBaseURL)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("bepaid create capture request: %w", err)
	}

	httpReq.SetBasicAuth(p.shopID, p.secretKey)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("bepaid http capture: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("bepaid capture: status %d, body: %s", resp.StatusCode, string(respBody))
	}

	return nil
}

// CancelPayment voids a previously authorized payment in bePaid.
func (p *BePaidProvider) CancelPayment(ctx context.Context, externalID string) error {
	reqBody := map[string]interface{}{
		"transaction": map[string]interface{}{
			"parent_uid": externalID,
		},
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("bepaid marshal void: %w", err)
	}

	url := fmt.Sprintf("%s/transactions/voids", bepaidBaseURL)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("bepaid create void request: %w", err)
	}

	httpReq.SetBasicAuth(p.shopID, p.secretKey)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("bepaid http void: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("bepaid void: status %d, body: %s", resp.StatusCode, string(respBody))
	}

	return nil
}

// mapBePaidStatus maps bePaid transaction statuses to our internal status strings.
func mapBePaidStatus(bepaidStatus string) string {
	switch bepaidStatus {
	case "successful":
		return "succeeded"
	case "failed", "expired":
		return "failed"
	case "pending", "incomplete":
		return "pending"
	case "authorized":
		return "waiting_for_capture"
	case "voided":
		return "canceled"
	default:
		return bepaidStatus
	}
}
