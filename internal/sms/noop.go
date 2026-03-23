package sms

import (
	"context"

	"github.com/nikitaaldaev/bani/internal/logger"
)

type noopProvider struct {
	logger *logger.Logger
}

// NewNoopProvider returns a Provider that logs but does not send SMS.
// Useful for development and testing.
func NewNoopProvider(log *logger.Logger) Provider {
	return &noopProvider{logger: log}
}

func (n *noopProvider) SendSMS(_ context.Context, phone string, message string) error {
	n.logger.Info("SMS (noop)", "phone", phone, "message", message)
	return nil
}
