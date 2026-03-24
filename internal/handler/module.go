package handler

import (
	"github.com/nikitaaldaev/bani/config"
	"github.com/nikitaaldaev/bani/internal/logger"
	"github.com/nikitaaldaev/bani/internal/service"
	"github.com/redis/go-redis/v9"
	"go.uber.org/fx"
)

var Module = fx.Module("handler",
	fx.Provide(
		NewAuthHandler,
		func(
			bathhouseService service.BathhouseService,
			bookingService service.BookingService,
			representativeService service.RepresentativeService,
			favoriteService service.FavoriteService,
			recommendationService service.RecommendationService,
			analyticsService service.AnalyticsService,
			mediaService service.MediaService,
			promotionService service.PromotionService,
			cityService service.CityService,
			savedSearchService service.SavedSearchService,
			suggestionService service.SearchSuggestionService,
			log *logger.Logger,
			cfg *config.Config,
		) *BathhouseHandler {
			return NewBathhouseHandler(bathhouseService, bookingService, representativeService, favoriteService, recommendationService, analyticsService, mediaService, promotionService, cityService, savedSearchService, suggestionService, log, cfg.BaseURL)
		},
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
		NewPhotoHandler,
		NewPromoHandler,
		NewMediaHandler,
		NewPaymentHandler,
		NewDeviceTokenHandler,
		NewWalletHandler,
		NewPayoutHandler,
		NewSessionHandler,
		NewCalendarHandler,
		func(svc service.ReferralService, cfg *config.Config) *ReferralHandler {
			return NewReferralHandler(svc, cfg.BaseURL)
		},
		func(svc service.BathhouseService, cityService service.CityService, redisClient *redis.Client, log *logger.Logger, cfg *config.Config) *SitemapHandler {
			return NewSitemapHandler(svc, cityService, redisClient, log, cfg.BaseURL)
		},
		NewKYCHandler,
		NewOfferHandler,
		NewPaymentDetailsHandler,
		NewListingDraftHandler,
		NewAuditLogHandler,
		NewAddOnHandler,
		NewSearchHandler,
		NewComparisonHandler,
		NewSavedSearchHandler,
		NewServiceFeeHandler,
	),
)
