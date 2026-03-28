package payment

import (
	"github.com/nikitaaldaev/bani/config"
	"go.uber.org/fx"
)

var Module = fx.Module("payment",
	fx.Provide(
		func(cfg *config.Config) PaymentProvider {
			base := NewYooKassaProvider(
				cfg.Payment.YooKassa.ShopID,
				cfg.Payment.YooKassa.SecretKey,
			)
			return NewRetryingProvider(base, DefaultRetryConfig)
		},
	),
)
