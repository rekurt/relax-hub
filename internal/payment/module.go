package payment

import (
	"github.com/nikitaaldaev/bani/config"
	"go.uber.org/fx"
)

var Module = fx.Module("payment",
	fx.Provide(
		func(cfg *config.Config) PaymentProvider {
			return NewYooKassaProvider(
				cfg.Payment.YooKassa.ShopID,
				cfg.Payment.YooKassa.SecretKey,
			)
		},
	),
)
