package fiscal

import (
	"github.com/rekurt/relax-hub/config"
	"github.com/rekurt/relax-hub/internal/logger"
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
				log.Info("using BY fiscal provider (receipt logging for manual processing)")
				return NewBYProvider(log)
			default:
				return NewNoOpProvider()
			}
		},
	),
)
