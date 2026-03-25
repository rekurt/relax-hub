package moderation

import (
	"github.com/nikitaaldaev/bani/config"
	"go.uber.org/fx"
)

var Module = fx.Module("moderation",
	fx.Provide(
		NewContentFilterProvider,
		func() TextModerationService {
			return NewRegexTextModerator()
		},
	),
)

func NewContentFilterProvider(cfg *config.Config) *ContentFilter {
	return NewContentFilter(cfg.Moderation.Enabled, cfg.Moderation.AutoApprove)
}
