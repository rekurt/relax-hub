package fiscal

import (
	"github.com/nikitaaldaev/bani/config"
	"github.com/nikitaaldaev/bani/internal/logger"
	"go.uber.org/fx"
)

var Module = fx.Module("fiscal",
	fx.Provide(
		func(cfg *config.Config, log *logger.Logger) FiscalProvider {
			switch cfg.Fiscal.Provider {
			case "atol":
				return NewATOLProvider(
					cfg.Fiscal.ATOLLogin,
					cfg.Fiscal.ATOLPassword,
					cfg.Fiscal.ATOLGroupCode,
					log,
				)
			default:
				return NewNoOpProvider()
			}
		},
	),
)
