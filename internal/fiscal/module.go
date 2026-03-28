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
			case "by":
				// Belarus fiscalization placeholder — not yet implemented.
				// Returns no-op provider until a BY fiscal provider is integrated.
				log.Info("using no-op fiscal provider for BY region (placeholder)")
				return NewNoOpProvider()
			default:
				return NewNoOpProvider()
			}
		},
	),
)
