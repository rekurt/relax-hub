package storage

import "go.uber.org/fx"

var Module = fx.Module("storage",
	fx.Provide(
		fx.Annotate(NewS3Storage, fx.As(new(FileStorage))),
	),
)
