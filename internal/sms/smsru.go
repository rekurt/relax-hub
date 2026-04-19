package sms

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/rekurt/relax-hub/internal/logger"
)

const (
	smsMaxRetries = 3
	smsBaseDelay  = 1 * time.Second
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

	var lastErr error
	for attempt := 0; attempt < smsMaxRetries; attempt++ {
		if attempt > 0 {
			delay := smsBaseDelay
			for i := 0; i < attempt-1; i++ {
				delay *= 2
			}
			s.logger.Warn("retrying SMS send", "phone", phone, "attempt", attempt+1, "delay", delay)
			timer := time.NewTimer(delay)
			select {
			case <-ctx.Done():
				timer.Stop()
				return fmt.Errorf("send SMS cancelled: %w", ctx.Err())
			case <-timer.C:
			}
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://sms.ru/sms/send?"+params.Encode(), nil)
		if err != nil {
			return fmt.Errorf("create SMS request: %w", err)
		}

		resp, err := s.client.Do(req)
		if err != nil {
			// Network error — retryable
			lastErr = fmt.Errorf("send SMS: %w", err)
			continue
		}

		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		if resp.StatusCode >= 500 {
			// Server error — retryable
			lastErr = fmt.Errorf("SMS.ru returned status %d: %s", resp.StatusCode, string(body))
			continue
		}

		if resp.StatusCode != http.StatusOK {
			// Client error — not retryable
			return fmt.Errorf("SMS.ru returned status %d: %s", resp.StatusCode, string(body))
		}

		s.logger.Info("SMS sent", "phone", phone)
		return nil
	}

	return fmt.Errorf("SMS send failed after %d attempts: %w", smsMaxRetries, lastErr)
}
