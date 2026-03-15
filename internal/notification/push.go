package notification

import (
	"context"
	"encoding/json"
	"fmt"

	webpush "github.com/SherClockHolmes/webpush-go"
	"github.com/nikitaaldaev/bani/internal/logger"
)

// PushSender sends push notifications to device tokens.
type PushSender interface {
	Send(ctx context.Context, token, title, body string, data map[string]string) error
}

// WebPushSender sends Web Push notifications using VAPID.
type WebPushSender struct {
	vapidPublicKey  string
	vapidPrivateKey string
	vapidContact    string
	logger          *logger.Logger
}

func NewWebPushSender(publicKey, privateKey, contact string, log *logger.Logger) PushSender {
	return &WebPushSender{
		vapidPublicKey:  publicKey,
		vapidPrivateKey: privateKey,
		vapidContact:    contact,
		logger:          log,
	}
}

func (s *WebPushSender) Send(_ context.Context, token, title, body string, data map[string]string) error {
	// Build the notification payload
	payload := map[string]interface{}{
		"title": title,
		"body":  body,
		"data":  data,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal push payload: %w", err)
	}

	// Parse the subscription from the token (stored as JSON)
	sub := &webpush.Subscription{}
	if err := json.Unmarshal([]byte(token), sub); err != nil {
		return fmt.Errorf("unmarshal push subscription: %w", err)
	}

	resp, err := webpush.SendNotification(payloadBytes, sub, &webpush.Options{
		VAPIDPublicKey:  s.vapidPublicKey,
		VAPIDPrivateKey: s.vapidPrivateKey,
		Subscriber:      s.vapidContact,
		TTL:             86400, // 24 hours
	})
	if err != nil {
		return fmt.Errorf("send web push: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("web push failed with status %d", resp.StatusCode)
	}

	return nil
}
