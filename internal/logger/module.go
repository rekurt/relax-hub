package logger

import (
	"github.com/nikitaaldaev/bani/config"
	"go.uber.org/fx"
)

var Module = fx.Module("logger",
	fx.Provide(NewLogger),
)

func NewLogger(cfg *config.Config) *Logger {
	level := ParseLogLevel(cfg.Logger.Level)
	return New(level)
}
