package payment

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/google/uuid"
)

const yookassaPayoutsURL = "https://api.yookassa.ru/v3/payouts"

// YooKassaPayoutProvider implements PayoutProvider using YooKassa Payouts API.
// The YooKassa SDK does not support payouts, so we use direct HTTP calls.
type YooKassaPayoutProvider struct {
	agentID   string // YooKassa agent_id for payouts
	secretKey string
	client    *http.Client
}

// NewYooKassaPayoutProvider creates a new payout provider for YooKassa.
func NewYooKassaPayoutProvider(agentID, secretKey string) *YooKassaPayoutProvider {
	return &YooKassaPayoutProvider{
		agentID:   agentID,
		secretKey: secretKey,
		client:    &http.Client{},
	}
}

type yooPayoutRequest struct {
	Amount          yooAmount       `json:"amount"`
	PayoutToken     string          `json:"payout_token,omitempty"`
	PayoutMethod    *yooPayoutDest  `json:"payout_destination_data,omitempty"`
	Description     string          `json:"description,omitempty"`
	Metadata        json.RawMessage `json:"metadata,omitempty"`
}

type yooAmount struct {
	Value    string `json:"value"`
	Currency string `json:"currency"`
}

type yooPayoutDest struct {
	Type  string `json:"type"`
	Phone string `json:"phone,omitempty"`
}

type yooPayoutResponse struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

func (p *YooKassaPayoutProvider) CreatePayout(ctx context.Context, req CreatePayoutRequest) (*PayoutResult, error) {
	body := yooPayoutRequest{
		Amount: yooAmount{
			Value:    kopecksToString(req.Amount),
			Currency: req.Currency,
		},
		Description: req.Description,
	}

	switch req.Method {
	case "sbp":
		body.PayoutMethod = &yooPayoutDest{
			Type:  "sbp",
			Phone: req.Phone,
		}
	case "bank_transfer":
		body.PayoutMethod = &yooPayoutDest{
			Type: "bank_card",
		}
	default:
		return nil, fmt.Errorf("unsupported payout method: %s", req.Method)
	}

	if req.Metadata != nil {
		meta, err := json.Marshal(req.Metadata)
		if err != nil {
			return nil, fmt.Errorf("marshal metadata: %w", err)
		}
		body.Metadata = meta
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal payout request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, yookassaPayoutsURL, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("create http request: %w", err)
	}

	httpReq.SetBasicAuth(p.agentID, p.secretKey)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Idempotence-Key", uuid.New().String())

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("yookassa payout request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read payout response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("yookassa payout error (status %d): %s", resp.StatusCode, string(respBody))
	}

	var result yooPayoutResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("unmarshal payout response: %w", err)
	}

	return &PayoutResult{
		ExternalID: result.ID,
		Status:     result.Status,
	}, nil
}

func (p *YooKassaPayoutProvider) GetPayoutStatus(ctx context.Context, externalID string) (string, error) {
	url := fmt.Sprintf("%s/%s", yookassaPayoutsURL, externalID)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("create http request: %w", err)
	}

	httpReq.SetBasicAuth(p.agentID, p.secretKey)

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("yookassa get payout: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read payout response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("yookassa get payout error (status %d): %s", resp.StatusCode, string(respBody))
	}

	var result yooPayoutResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", fmt.Errorf("unmarshal payout response: %w", err)
	}

	return result.Status, nil
}
