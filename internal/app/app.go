package app

import (
	"github.com/rekurt/relax-hub/config"
	"github.com/rekurt/relax-hub/internal/admin"
	"github.com/rekurt/relax-hub/internal/antifraud"
	"github.com/rekurt/relax-hub/internal/calendar"
	"github.com/rekurt/relax-hub/internal/cron"
	"github.com/rekurt/relax-hub/internal/database"
	"github.com/rekurt/relax-hub/internal/fiscal"
	"github.com/rekurt/relax-hub/internal/geo"
	"github.com/rekurt/relax-hub/internal/handler"
	"github.com/rekurt/relax-hub/internal/logger"
	"github.com/rekurt/relax-hub/internal/middleware"
	"github.com/rekurt/relax-hub/internal/moderation"
	"github.com/rekurt/relax-hub/internal/notification"
	"github.com/rekurt/relax-hub/internal/payment"
	"github.com/rekurt/relax-hub/internal/pms"
	"github.com/rekurt/relax-hub/internal/seo"
	repopostgres "github.com/rekurt/relax-hub/internal/repository/postgres"
	"github.com/rekurt/relax-hub/internal/server"
	"github.com/rekurt/relax-hub/internal/service"
	"github.com/rekurt/relax-hub/internal/sms"
	"github.com/rekurt/relax-hub/internal/storage"
	"go.uber.org/fx"
)

func New(cfg *config.Config, extraOpts ...fx.Option) *fx.App {
	opts := []fx.Option{
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
	}
	opts = append(opts, extraOpts...)
	return fx.New(opts...)
}
