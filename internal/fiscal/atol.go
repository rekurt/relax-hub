package fiscal

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/rekurt/relax-hub/internal/logger"
)

const (
	atolBaseURL     = "https://online.atol.ru/possystem/v4"
	atolTestBaseURL = "https://testonline.atol.ru/possystem/v4"
	atolTokenTTL    = 23 * time.Hour // refresh before 24h expiry
)

// ATOLProvider implements FiscalProvider using the ATOL Online API v4.
type ATOLProvider struct {
	login     string
	password  string
	groupCode string
	baseURL   string
	logger    *logger.Logger

	mu         sync.Mutex
	token      string
	tokenExpAt time.Time
	client     *http.Client
}

// NewATOLProvider creates a new ATOL Online fiscal provider.
func NewATOLProvider(login, password, groupCode string, log *logger.Logger) *ATOLProvider {
	baseURL := atolBaseURL
	// Use test endpoint if login starts with "test" (ATOL convention)
	if len(login) >= 4 && login[:4] == "test" {
		baseURL = atolTestBaseURL
	}
	return &ATOLProvider{
		login:     login,
		password:  password,
		groupCode: groupCode,
		baseURL:   baseURL,
		logger:    log,
		client:    &http.Client{Timeout: 30 * time.Second},
	}
}

// getToken returns a valid ATOL API token, refreshing if expired.
func (p *ATOLProvider) getToken(ctx context.Context) (string, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.token != "" && time.Now().Before(p.tokenExpAt) {
		return p.token, nil
	}

	body, _ := json.Marshal(map[string]string{
		"login": p.login,
		"pass":  p.password,
	})

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.baseURL+"/getToken", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("atol create token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("atol get token: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		Token string `json:"token"`
		Error *struct {
			Code int    `json:"code"`
			Text string `json:"text"`
		} `json:"error"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("atol decode token response: %w", err)
	}
	if result.Error != nil {
		return "", fmt.Errorf("atol auth error %d: %s", result.Error.Code, result.Error.Text)
	}
	if result.Token == "" {
		return "", fmt.Errorf("atol returned empty token")
	}

	p.token = result.Token
	p.tokenExpAt = time.Now().Add(atolTokenTTL)
	return p.token, nil
}

// atolReceiptRequest is the ATOL API v4 receipt body.
type atolReceiptRequest struct {
	ExternalID string          `json:"external_id"`
	Receipt    atolReceipt     `json:"receipt"`
	Timestamp  string          `json:"timestamp"`
	Service    atolServiceInfo `json:"service"`
}

type atolReceipt struct {
	Client atolClient `json:"client"`
	Items  []atolItem `json:"items"`
	Total  float64    `json:"total"`
}

type atolClient struct {
	Email string `json:"email,omitempty"`
	Phone string `json:"phone,omitempty"`
}

type atolItem struct {
	Name            string      `json:"name"`
	Price           float64     `json:"price"`
	Quantity        float64     `json:"quantity"`
	Sum             float64     `json:"sum"`
	PaymentMethod   string      `json:"payment_method"`
	PaymentObject   string      `json:"payment_object"`
	Vat             atolVAT     `json:"vat"`
	MeasurementUnit string      `json:"measurement_unit"`
}

type atolVAT struct {
	Type string `json:"type"`
}

type atolServiceInfo struct {
	CallbackURL string `json:"callback_url,omitempty"`
}

func receiptTypeToOperation(rt ReceiptType) string {
	switch rt {
	case ReceiptRefund:
		return "sell_refund"
	default:
		return "sell"
	}
}

func vatToATOL(vat string) string {
	switch vat {
	case "vat20":
		return "vat20"
	case "vat10":
		return "vat10"
	case "vat0":
		return "vat0"
	default:
		return "none"
	}
}

func taxSystemToATOL(ts TaxSystem) string {
	switch ts {
	case TaxSystemOSN:
		return "osn"
	case TaxSystemUSN:
		return "usn_income"
	case TaxSystemPatent:
		return "patent"
	case TaxSystemNPD:
		return "osn" // ATOL doesn't have NPD; self-employed use simplified
	default:
		return "osn"
	}
}

// CreateReceipt sends a fiscal receipt to ATOL Online and returns the receipt info.
func (p *ATOLProvider) CreateReceipt(ctx context.Context, req ReceiptRequest) (*Receipt, error) {
	token, err := p.getToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("atol auth: %w", err)
	}

	externalID := uuid.New().String()
	operation := receiptTypeToOperation(req.Type)

	taxSystem := req.TaxSystem
	if taxSystem == "" {
		taxSystem = TaxSystemDefault
	}

	items := make([]atolItem, len(req.Items))
	for i, it := range req.Items {
		price := float64(it.Price) / 100.0
		qty := float64(it.Quantity)
		items[i] = atolItem{
			Name:            it.Name,
			Price:           price,
			Quantity:        qty,
			Sum:             price * qty,
			PaymentMethod:   "full_payment",
			PaymentObject:   "service",
			Vat:             atolVAT{Type: vatToATOL(it.VAT)},
			MeasurementUnit: "шт",
		}
	}

	total := float64(req.Amount) / 100.0

	atolReq := atolReceiptRequest{
		ExternalID: externalID,
		Timestamp:  time.Now().Format("02.01.2006 15:04:05"),
		Receipt: atolReceipt{
			Client: atolClient{Email: req.Email, Phone: req.Phone},
			Items:  items,
			Total:  total,
		},
	}

	bodyBytes, err := json.Marshal(atolReq)
	if err != nil {
		return nil, fmt.Errorf("atol marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/%s/%s?token=%s", p.baseURL, p.groupCode, operation, token)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("atol create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("atol send receipt: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	var result struct {
		UUID  string `json:"uuid"`
		Error *struct {
			Code int    `json:"code"`
			Text string `json:"text"`
		} `json:"error"`
		Status string `json:"status"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("atol decode response: %w (body: %s)", err, string(respBody))
	}
	if result.Error != nil && result.Error.Code != 0 {
		p.logger.Error("ATOL receipt error",
			"code", result.Error.Code, "text", result.Error.Text,
			"external_id", externalID, "operation", operation)
		return nil, fmt.Errorf("atol error %d: %s", result.Error.Code, result.Error.Text)
	}

	p.logger.Info("ATOL receipt created",
		"uuid", result.UUID, "external_id", externalID,
		"operation", operation, "amount", req.Amount,
		"tax_system", string(taxSystem))

	return &Receipt{
		ID:         externalID,
		ExternalID: result.UUID,
		Status:     result.Status,
	}, nil
}
