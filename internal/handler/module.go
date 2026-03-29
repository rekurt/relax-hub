package handler

import (
	"github.com/nikitaaldaev/bani/config"
	"github.com/nikitaaldaev/bani/internal/logger"
	"github.com/nikitaaldaev/bani/internal/repository"
	"github.com/nikitaaldaev/bani/internal/seo"
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
			reviewService service.ReviewService,
			log *logger.Logger,
			cfg *config.Config,
		) *BathhouseHandler {
			return NewBathhouseHandler(bathhouseService, bookingService, representativeService, favoriteService, recommendationService, analyticsService, mediaService, promotionService, cityService, savedSearchService, suggestionService, reviewService, log, cfg.BaseURL)
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
		NewGuestCardHandler,
		NewRFMHandler,
		NewBroadcastHandler,
		NewAutoScenarioHandler,
		NewHolidayHandler,
		NewServiceFeeHandler,
		NewTemplateHandler,
		NewTicketHandler,
		NewDisputeHandler,
		NewAntiFraudHandler,
		NewPlatformSettingsHandler,
		NewFeatureFlagHandler,
		NewForceMajeureHandler,
		NewClientReviewHandler,
		NewRegionHandler,
		NewFinancialReportHandler,
		NewReconciliationHandler,
		NewListingImportHandler,
		func(shareRepo repository.BookingShareRepository, cfg *config.Config) *ShareHandler {
			return NewShareHandler(shareRepo, cfg.FrontendURL)
		},
		NewAmenityHandler,
		NewObjectTypeHandler,
		NewSavedCardHandler,
		NewBankReconciliationHandler,
		NewAdminRoleHandler,
		NewAdminNotificationHandler,
		NewFAQHandler,
		NewWebhookHandler,
		NewPMSHandler,
		NewPhotoOrderHandler,
		NewIsochroneHandler,
		NewTransportHandler,
		func(renderer *seo.Renderer, bathhouseService service.BathhouseService, cityService service.CityService, reviewService service.ReviewService, log *logger.Logger, cfg *config.Config) *PrerenderHandler {
			return NewPrerenderHandler(renderer, bathhouseService, cityService, reviewService, log, cfg.BaseURL)
		},
	),
)
