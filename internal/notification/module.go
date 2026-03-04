package notification

import (
	"go.uber.org/fx"
)

var Module = fx.Module("notification",
	fx.Provide(
		NewDispatcher,
		func() EmailSender { return NewNoopEmailSender() },
	),
)
