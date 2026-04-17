package app

import (
	"github.com/nikitaaldaev/bani/config"
	"github.com/nikitaaldaev/bani/internal/admin"
	"github.com/nikitaaldaev/bani/internal/antifraud"
	"github.com/nikitaaldaev/bani/internal/calendar"
	"github.com/nikitaaldaev/bani/internal/cron"
	"github.com/nikitaaldaev/bani/internal/database"
	"github.com/nikitaaldaev/bani/internal/fiscal"
	"github.com/nikitaaldaev/bani/internal/geo"
	"github.com/nikitaaldaev/bani/internal/handler"
	"github.com/nikitaaldaev/bani/internal/logger"
	"github.com/nikitaaldaev/bani/internal/middleware"
	"github.com/nikitaaldaev/bani/internal/moderation"
	"github.com/nikitaaldaev/bani/internal/notification"
	"github.com/nikitaaldaev/bani/internal/payment"
	"github.com/nikitaaldaev/bani/internal/pms"
	"github.com/nikitaaldaev/bani/internal/seo"
	repopostgres "github.com/nikitaaldaev/bani/internal/repository/postgres"
	"github.com/nikitaaldaev/bani/internal/server"
	"github.com/nikitaaldaev/bani/internal/service"
	"github.com/nikitaaldaev/bani/internal/sms"
	"github.com/nikitaaldaev/bani/internal/storage"
	"go.uber.org/fx"
)

func New(cfg *config.Config) *fx.App {
	return fx.New(
		fx.Supply(cfg),
		logger.Module,
		database.PostgresModule,
		database.RedisModule,
		middleware.CORSModule,
		repopostgres.Module,
		storage.Module,
		moderation.Module,
		notification.Module,
		payment.Module,
		fiscal.Module,
		sms.Module,
		service.Module,
		pms.Module,
		geo.Module,
		antifraud.Module,
		seo.Module,
		handler.Module,
		cron.Module,
		server.Module,
		admin.ProvideConditionalModule(cfg),
		// Cross-package interface bindings
		fx.Provide(
			func(cfg *config.Config, log *logger.Logger) notification.TelegramSender {
				if cfg.Telegram.BotToken == "" {
					log.Warn("Telegram bot token not configured, telegram notifications disabled (set BANI_TELEGRAM_BOT_TOKEN)")
					return nil
				}
				sender, err := notification.NewTelegramSender(cfg.Telegram.BotToken, log)
				if err != nil {
					log.Error("failed to create telegram sender, notifications disabled", "error", err)
					return nil
				}
				return sender
			},
			calendar.NewCalendarSyncService,
		),
	)
}
