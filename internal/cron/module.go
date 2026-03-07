package cron

import (
	"go.uber.org/fx"
)

var Module = fx.Module("cron",
	fx.Provide(NewCronScheduler),
	fx.Invoke(RegisterCron),
)
