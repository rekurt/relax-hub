package sms

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/rekurt/relax-hub/internal/logger"
)

type smsruProvider struct {
	apiKey string
	client *http.Client
	logger *logger.Logger
}

func NewSMSRuProvider(apiKey string, log *logger.Logger) Provider {
	return &smsruProvider{
		apiKey: apiKey,
		client: &http.Client{},
		logger: log,
	}
}

func (s *smsruProvider) SendSMS(ctx context.Context, phone string, message string) error {
	params := url.Values{}
	params.Set("api_id", s.apiKey)
	params.Set("to", phone)
	params.Set("msg", message)
	params.Set("json", "1")

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://sms.ru/sms/send?"+params.Encode(), nil)
	if err != nil {
		return fmt.Errorf("create SMS request: %w", err)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("send SMS: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("SMS.ru returned status %d: %s", resp.StatusCode, string(body))
	}

	s.logger.Info("SMS sent", "phone", phone)
	return nil
}
