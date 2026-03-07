package bot

import (
	"github.com/nikitaaldaev/bani/config"
	"go.uber.org/fx"
)

var Module = fx.Module("bot",
	fx.Provide(
		ProvideBotConfig,
		NewBot,
	),
)

// ProvideBotConfig extracts telegram config and provides it for the bot
func ProvideBotConfig(cfg *config.Config) *config.TelegramConfig {
	return &cfg.Telegram
}
