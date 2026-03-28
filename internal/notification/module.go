package notification

import (
	"github.com/nikitaaldaev/bani/internal/sms"
	"go.uber.org/fx"
)

var Module = fx.Module("notification",
	fx.Provide(
		NewDispatcher,
		NewHub,
		func() EmailSender { return nil },
		func() PushSender { return nil },
		func() sms.Provider { return nil },
	),
)
