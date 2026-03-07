package notification

import (
	"context"

	"go.uber.org/fx"
)

// NoopTelegramSender is a no-op implementation used when Telegram bot is not configured.
type NoopTelegramSender struct{}

func NewNoopTelegramSender() TelegramSender {
	return &NoopTelegramSender{}
}

func (n *NoopTelegramSender) Send(_ context.Context, _ int64, _, _ string) error {
	return nil
}

var Module = fx.Module("notification",
	fx.Provide(
		NewDispatcher,
		NewHub,
		func() EmailSender { return NewNoopEmailSender() },
	),
)
