package service

import (
	"github.com/nikitaaldaev/bani/internal/middleware"
	"go.uber.org/fx"
)

var Module = fx.Module("service",
	fx.Provide(
		NewAccessChecker,
		fx.Annotate(
			NewAuthService,
			fx.As(new(AuthService)),
			fx.As(new(middleware.AuthService)),
		),
		fx.Annotate(NewUserService, fx.As(new(UserService))),
		fx.Annotate(NewBathhouseService, fx.As(new(BathhouseService))),
		fx.Annotate(NewBookingService, fx.As(new(BookingService))),
		fx.Annotate(NewRepresentativeService, fx.As(new(RepresentativeService))),
		fx.Annotate(NewReviewService, fx.As(new(ReviewService))),
		fx.Annotate(NewFavoriteService, fx.As(new(FavoriteService))),
		fx.Annotate(NewCityService, fx.As(new(CityService))),
		fx.Annotate(NewNotificationService, fx.As(new(NotificationService))),
		fx.Annotate(NewOAuthService, fx.As(new(OAuthService))),
		fx.Annotate(NewRecommendationService, fx.As(new(RecommendationService))),
		fx.Annotate(NewSubscriptionService, fx.As(new(SubscriptionService))),
		fx.Annotate(NewPricingService, fx.As(new(PricingService))),
		fx.Annotate(NewLoyaltyService, fx.As(new(LoyaltyService))),
		fx.Annotate(NewChatService, fx.As(new(ChatService))),
		fx.Annotate(NewAnalyticsService, fx.As(new(AnalyticsService))),
		fx.Annotate(NewTelegramLinkService, fx.As(new(TelegramLinkService))),
		fx.Annotate(NewComplaintService, fx.As(new(ComplaintService))),
	),
)
