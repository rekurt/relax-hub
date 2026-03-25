package antifraud

import "go.uber.org/fx"

var Module = fx.Module("antifraud",
	fx.Provide(
		fx.Annotate(NewFraudEngine, fx.As(new(FraudEngine))),
		fx.Annotate(NewChatFilter, fx.As(new(ChatFilter))),
	),
)
