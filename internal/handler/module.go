package handler

import (
	"github.com/nikitaaldaev/bani/config"
	"github.com/nikitaaldaev/bani/internal/logger"
	"github.com/nikitaaldaev/bani/internal/repository"
	"github.com/nikitaaldaev/bani/internal/service"
	"github.com/redis/go-redis/v9"
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
		NewCertificateHandler,
		func(svc service.ReferralService, cfg *config.Config) *ReferralHandler {
			return NewReferralHandler(svc, cfg.BaseURL)
		},
		func(svc service.BathhouseService, cityRepo repository.CityRepository, redisClient *redis.Client, log *logger.Logger, cfg *config.Config) *SitemapHandler {
			return NewSitemapHandler(svc, cityRepo, redisClient, log, cfg.BaseURL)
		},
	),
)
