package server

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/rekurt/relax-hub/config"
	"github.com/rekurt/relax-hub/internal/admin"
	"github.com/rekurt/relax-hub/internal/handler"
	"github.com/rekurt/relax-hub/internal/logger"
	"github.com/rekurt/relax-hub/internal/middleware"
	"github.com/rekurt/relax-hub/internal/repository"
	"go.uber.org/fx"
)

type RouterParams struct {
	fx.In

	Lifecycle             fx.Lifecycle
	Log                   *logger.Logger
	Config                *config.Config
	CORS                  *middleware.CORSMiddleware
	AuthService           middleware.AuthService
	AuthHandler           *handler.AuthHandler
	BHHandler             *handler.BathhouseHandler
	BookingHandler        *handler.BookingHandler
	ReviewHandler         *handler.ReviewHandler
	FavHandler            *handler.FavoriteHandler
	RepHandler            *handler.RepresentativeHandler
	CityHandler           *handler.CityHandler
	AdminHandler          *handler.AdminHandler
	HealthHandler         *handler.HealthHandler
	WSHandler             *handler.WSHandler
	NotifHandler          *handler.NotificationHandler
	OAuthHandler          *handler.OAuthHandler
	RecommendationHandler *handler.RecommendationHandler
	SubscriptionHandler   *handler.SubscriptionHandler
	PricingHandler        *handler.PricingHandler
	WidgetHandler         *handler.WidgetHandler
	LoyaltyHandler        *handler.LoyaltyHandler
	ChatHandler           *handler.ChatHandler
	AnalyticsHandler      *handler.AnalyticsHandler
	ComplaintHandler      *handler.ComplaintHandler
	ReferralHandler       *handler.ReferralHandler
	CertificateHandler    *handler.CertificateHandler
	SitemapHandler        *handler.SitemapHandler
	PhotoHandler          *handler.PhotoHandler
	PromoHandler          *handler.PromoHandler
	MediaHandler          *handler.MediaHandler
	PaymentHandler        *handler.PaymentHandler
	DeviceTokenHandler    *handler.DeviceTokenHandler
	CalendarHandler       *handler.CalendarHandler
	WalletHandler         *handler.WalletHandler
	PayoutHandler         *handler.PayoutHandler
	SessionHandler        *handler.SessionHandler
	KYCHandler            *handler.KYCHandler
	OfferHandler          *handler.OfferHandler
	PaymentDetailsHandler *handler.PaymentDetailsHandler
	ListingDraftHandler   *handler.ListingDraftHandler
	AuditLogHandler       *handler.AuditLogHandler
	AddOnHandler          *handler.AddOnHandler
	SearchHandler         *handler.SearchHandler
	ComparisonHandler     *handler.ComparisonHandler
	SavedSearchHandler    *handler.SavedSearchHandler
	HolidayHandler        *handler.HolidayHandler
	GuestCardHandler      *handler.GuestCardHandler
	RFMHandler            *handler.RFMHandler
	BroadcastHandler      *handler.BroadcastHandler
	AutoScenarioHandler   *handler.AutoScenarioHandler
	TemplateHandler       *handler.TemplateHandler
	ServiceFeeHandler     *handler.ServiceFeeHandler
	TicketHandler         *handler.TicketHandler
	DisputeHandler        *handler.DisputeHandler
	AntiFraudHandler          *handler.AntiFraudHandler
	PlatformSettingsHandler   *handler.PlatformSettingsHandler
	FeatureFlagHandler        *handler.FeatureFlagHandler
	ForceMajeureHandler       *handler.ForceMajeureHandler
	ClientReviewHandler       *handler.ClientReviewHandler
	RegionHandler             *handler.RegionHandler
	FinancialReportHandler    *handler.FinancialReportHandler
	ReconciliationHandler     *handler.ReconciliationHandler
	ListingImportHandler     *handler.ListingImportHandler
	ShareHandler             *handler.ShareHandler
	AmenityHandler           *handler.AmenityHandler
	ObjectTypeHandler        *handler.ObjectTypeHandler
	SavedCardHandler              *handler.SavedCardHandler
	BankReconciliationHandler    *handler.BankReconciliationHandler
	AdminRoleHandler             *handler.AdminRoleHandler
	AdminNotificationHandler     *handler.AdminNotificationHandler
	FAQHandler                   *handler.FAQHandler
	WebhookHandler               *handler.WebhookHandler
	PMSHandler                   *handler.PMSHandler
	IsochroneHandler             *handler.IsochroneHandler
	TransportHandler             *handler.TransportHandler
	PhotoOrderHandler            *handler.PhotoOrderHandler
	BookingModificationHandler   *handler.BookingModificationHandler
	BookingExtensionHandler      *handler.BookingExtensionHandler
	PrerenderHandler             *handler.PrerenderHandler
	BuildInfo             handler.BuildInfo `optional:"true"`
	AuditLogRepo              repository.AuditLogRepository
	AdminSubRoleResolver  middleware.AdminSubRoleResolver
	Admin2FAChecker       middleware.Admin2FAChecker
	SessionValidator      middleware.SessionValidator `optional:"true"`
	GoAdmin               *admin.GoAdmin              `optional:"true"`
}

func NewRouter(p RouterParams) http.Handler {
	r := chi.NewRouter()

	r.Use(chiMiddleware.RequestID)
	r.Use(middleware.Logging(*p.Log))
	r.Use(middleware.Metrics)
	r.Use(middleware.RecoveryMiddleware(middleware.IsDevEnvironment(p.Config.Environment), p.Log))
	r.Use(p.CORS.Handler)
	r.Use(middleware.SecurityHeaders)
	r.Use(func(next http.Handler) http.Handler {
		return http.MaxBytesHandler(next, 10<<20) // 10MB
	})

	mountPublicRoutes(r, p)

	auth := middleware.RequireAuthWithSession(p.AuthService, p.SessionValidator)
	optionalAuth := middleware.OptionalAuth(p.AuthService)

	// Create rate limiters
	widgetRateLimiter := middleware.NewRateLimiter()
	widgetRateLimit := middleware.WidgetRateLimit(widgetRateLimiter, 10) // 10 req/s per API key

	authRegisterRateLimiter := middleware.NewRateLimiter()
	authLoginRateLimiter := middleware.NewRateLimiter()
	webhookRateLimiter := middleware.NewRateLimiter()
	promoRateLimiter := middleware.NewRateLimiter()

	// Close rate limiters on shutdown to stop cleanup goroutines
	if p.Lifecycle != nil {
		p.Lifecycle.Append(fx.Hook{
			OnStop: func(ctx context.Context) error {
				widgetRateLimiter.Close()
				authRegisterRateLimiter.Close()
				authLoginRateLimiter.Close()
				webhookRateLimiter.Close()
				promoRateLimiter.Close()
				return nil
			},
		})
	}

	r.Route("/api/v1", func(r chi.Router) {
		mountAPIRoutes(r, p, auth, optionalAuth, widgetRateLimit, webhookRateLimiter, authRegisterRateLimiter, authLoginRateLimiter, promoRateLimiter)
		mountMyRoutes(r, p, auth)
		mountAdminRoutes(r, p, auth)
	})

	// Pass mux reference to GoAdmin for deferred mounting in OnStart lifecycle.
	// GoAdmin's Engine.Use() requires AddConfig to be called first (which happens in OnStart),
	// so we cannot mount GoAdmin routes here during the fx Provide phase.
	if p.GoAdmin != nil {
		p.GoAdmin.Mux = r
	}

	return r
}
