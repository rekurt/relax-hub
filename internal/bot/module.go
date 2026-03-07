package bot

import (
	"github.com/nikitaaldaev/bani/config"
	"github.com/nikitaaldaev/bani/internal/logger"
	"github.com/nikitaaldaev/bani/internal/notification"
	"go.uber.org/fx"
)

var Module = fx.Module("bot",
	fx.Provide(
		ProvideBotConfig,
		NewBot,
		ProvideTelegramSender,
	),
)

// ProvideBotConfig extracts telegram config and provides it for the bot
func ProvideBotConfig(cfg *config.Config) *config.TelegramConfig {
	return &cfg.Telegram
}

// ProvideTelegramSender creates a TelegramSender from the Bot's API client
func ProvideTelegramSender(b *Bot, log *logger.Logger) notification.TelegramSender {
	return NewTelegramSender(b.client, log)
}
