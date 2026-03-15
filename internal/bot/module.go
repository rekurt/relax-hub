package bot

import (
	"github.com/nikitaaldaev/bani/config"
	"github.com/nikitaaldaev/bani/internal/logger"
	"github.com/nikitaaldaev/bani/internal/notification"
	"go.uber.org/fx"
)

var Module = fx.Module("bot",
	fx.Provide(
		func(cfg *config.Config) *config.TelegramConfig { return &cfg.Telegram },
		NewBot,
		ProvideTelegramSender,
	),
)

// ProvideTelegramSender creates a TelegramSender from the Bot's API client
func ProvideTelegramSender(b *Bot, log *logger.Logger) notification.TelegramSender {
	return NewTelegramSender(b.client, log)
}
