package payment

import (
	"github.com/rekurt/relax-hub/config"
	"go.uber.org/fx"
)

var Module = fx.Module("payment",
	fx.Provide(
		func(cfg *config.Config) PaymentProvider {
			providers := make(map[string]PaymentProvider)

			// RU region: YooKassa
			ruBase := NewYooKassaProvider(
				cfg.Payment.YooKassa.ShopID,
				cfg.Payment.YooKassa.SecretKey,
			)
			providers["RU"] = NewRetryingProvider(ruBase, DefaultRetryConfig)

			// BY region: bePaid (only if configured)
			if cfg.Payment.BePaid.ShopID != "" {
				byBase := NewBePaidProvider(
					cfg.Payment.BePaid.ShopID,
					cfg.Payment.BePaid.SecretKey,
				)
				providers["BY"] = NewRetryingProvider(byBase, DefaultRetryConfig)
			}

			return NewProviderFactory(providers)
		},
		func(cfg *config.Config) PayoutProvider {
			if cfg.Payment.YooKassa.PayoutAgentID != "" {
				return NewYooKassaPayoutProvider(
					cfg.Payment.YooKassa.PayoutAgentID,
					cfg.Payment.YooKassa.PayoutSecretKey,
				)
			}
			// Return a no-op provider when payouts are not configured
			return NewNoopPayoutProvider()
		},
	),
)
