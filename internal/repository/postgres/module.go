package postgres

import (
	"github.com/nikitaaldaev/bani/internal/repository"
	"go.uber.org/fx"
)

var Module = fx.Module("repository",
	fx.Provide(
		fx.Annotate(NewUserRepository, fx.As(new(repository.UserRepository))),
		fx.Annotate(NewCityRepository, fx.As(new(repository.CityRepository))),
		fx.Annotate(NewBathhouseRepository, fx.As(new(repository.BathhouseRepository))),
		fx.Annotate(NewBookingRepository, fx.As(new(repository.BookingRepository))),
		fx.Annotate(NewReviewRepository, fx.As(new(repository.ReviewRepository))),
		fx.Annotate(NewFavoriteRepository, fx.As(new(repository.FavoriteRepository))),
		fx.Annotate(NewRepresentativeRepository, fx.As(new(repository.RepresentativeRepository))),
		fx.Annotate(NewNotificationRepository, fx.As(new(repository.NotificationRepository))),
		fx.Annotate(NewSocialAccountRepository, fx.As(new(repository.SocialAccountRepository))),
		fx.Annotate(NewRecommendationRepository, fx.As(new(repository.RecommendationRepository))),
		fx.Annotate(NewSubscriptionRepository, fx.As(new(repository.SubscriptionRepository))),
		fx.Annotate(NewPromotionRepository, fx.As(new(repository.PromotionRepository))),
		fx.Annotate(NewPricingRuleRepository, fx.As(new(repository.PricingRuleRepository))),
		fx.Annotate(NewLoyaltyRepository, fx.As(new(repository.LoyaltyRepository))),
		fx.Annotate(NewConversationRepository, fx.As(new(repository.ConversationRepository))),
		fx.Annotate(NewMessageRepository, fx.As(new(repository.MessageRepository))),
		fx.Annotate(NewAnalyticsRepository, fx.As(new(repository.AnalyticsRepository))),
		fx.Annotate(NewTelegramLinkRepository, fx.As(new(repository.TelegramLinkRepository))),
		fx.Annotate(NewComplaintRepository, fx.As(new(repository.ComplaintRepository))),
		fx.Annotate(NewReferralRepository, fx.As(new(repository.ReferralRepository))),
	),
)
