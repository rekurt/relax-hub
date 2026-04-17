package notification

import (
	"github.com/nikitaaldaev/bani/config"
	"github.com/nikitaaldaev/bani/internal/logger"
	"go.uber.org/fx"
)

var Module = fx.Module("notification",
	fx.Provide(
		NewDispatcher,
		NewHub,
		func(cfg *config.Config, log *logger.Logger) EmailSender {
			if cfg.Email.Host == "" {
				log.Warn("SMTP not configured, email notifications disabled (set BANI_EMAIL_HOST)")
				return nil
			}
			return NewSMTPEmailSender(SMTPConfig{
				Host:     cfg.Email.Host,
				Port:     cfg.Email.Port,
				Username: cfg.Email.Username,
				Password: cfg.Email.Password,
				From:     cfg.Email.From,
			}, log)
		},
		func(cfg *config.Config, log *logger.Logger) PushSender {
			if cfg.WebPush.VAPIDPublicKey == "" {
				log.Warn("VAPID keys not configured, push notifications disabled (set BANI_WEBPUSH_VAPID_PUBLIC_KEY)")
				return nil
			}
			return NewWebPushSender(
				cfg.WebPush.VAPIDPublicKey,
				cfg.WebPush.VAPIDPrivateKey,
				cfg.WebPush.VAPIDContact,
				log,
			)
		},
	),
)
