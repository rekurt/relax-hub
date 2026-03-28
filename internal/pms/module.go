package pms

import (
	"go.uber.org/fx"
)

var Module = fx.Module("pms",
	fx.Provide(
		func() *ProviderRegistry {
			return NewProviderRegistry([]PMSProvider{
				NewYclientsProvider(),
				NewRestoplaceProvider(),
			})
		},
	),
)
