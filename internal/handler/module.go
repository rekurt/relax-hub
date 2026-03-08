package handler

import (
	"github.com/nikitaaldaev/bani/config"
	"github.com/nikitaaldaev/bani/internal/service"
	"go.uber.org/fx"
)

var Module = fx.Module("handler",
	fx.Provide(
		NewAuthHandler,
		NewBathhouseHandler,
		NewBookingHandler,
		NewReviewHandler,
		NewFavoriteHandler,
		NewRepresentativeHandler,
		NewCityHandler,
		NewAdminHandler,
		NewHealthHandler,
		NewWSHandler,
		NewNotificationHandler,
		NewOAuthHandler,
		NewRecommendationHandler,
		NewSubscriptionHandler,
		NewPricingHandler,
		NewWidgetHandler,
		NewLoyaltyHandler,
		NewChatHandler,
		NewAnalyticsHandler,
		NewComplaintHandler,
		func(svc service.ReferralService, cfg *config.Config) *ReferralHandler {
			return NewReferralHandler(svc, cfg.BaseURL)
		},
	),
)
