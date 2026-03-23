package sms

import (
	"github.com/nikitaaldaev/bani/config"
	"github.com/nikitaaldaev/bani/internal/logger"
	"go.uber.org/fx"
)

var Module = fx.Module("sms",
	fx.Provide(func(cfg *config.Config, log *logger.Logger) Provider {
		if cfg.SMS.APIKey == "" {
			log.Warn("SMS API key not configured, using noop provider")
			return NewNoopProvider(log)
		}
		return NewSMSRuProvider(cfg.SMS.APIKey, log)
	}),
)
