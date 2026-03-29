package server

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/nikitaaldaev/bani/config"
	"github.com/nikitaaldaev/bani/internal/admin"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/handler"
	"github.com/nikitaaldaev/bani/internal/logger"
	"github.com/nikitaaldaev/bani/internal/middleware"
	"github.com/nikitaaldaev/bani/internal/repository"
	httpSwagger "github.com/swaggo/http-swagger/v2"
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
	r.Use(middleware.RecoveryMiddleware(middleware.IsDevEnvironment(p.Config.Environment), p.Log))
	r.Use(p.CORS.Handler)

	r.Get("/health", p.HealthHandler.Health)
	r.Get("/ready", p.HealthHandler.Ready)
	r.Get("/sitemap.xml", p.SitemapHandler.Sitemap)
	r.Get("/calendar/{token}.ics", p.CalendarHandler.ExportICalByToken)

	// Pre-rendered pages for search engine bots (SEO)
	r.Get("/prerender/catalog", p.PrerenderHandler.MainListing)
	r.Get("/prerender/cities/{citySlug}", p.PrerenderHandler.CityListing)
	r.Get("/prerender/bathhouses/{slug}", p.PrerenderHandler.BathhouseDetail)
	r.Get("/prerender/bathhouses/{slug}/reviews", p.PrerenderHandler.BathhouseReviews)

	// Swagger UI (disabled in production)
	if middleware.IsDevEnvironment(p.Config.Environment) {
		r.Get("/swagger/*", httpSwagger.Handler(
			httpSwagger.URL("/swagger/doc.json"),
		))
	}

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
		// WebSocket (auth via query parameter)
		r.Get("/ws/notifications", p.WSHandler.HandleWS)

		// Webhooks (public, called by payment providers)
		r.With(middleware.RateLimit(webhookRateLimiter, 0.5)).Post("/webhooks/yookassa", p.PaymentHandler.HandleWebhook)       // 30/min
		r.With(middleware.RateLimit(webhookRateLimiter, 0.5)).Post("/webhooks/bepaid", p.PaymentHandler.HandleBePaidWebhook) // 30/min

		// Auth (public, rate-limited)
		r.With(middleware.RateLimit(authRegisterRateLimiter, 5.0/60.0)).Post("/auth/register", p.AuthHandler.Register) // 5/min
		r.With(middleware.RateLimit(authLoginRateLimiter, 10.0/60.0)).Post("/auth/login", p.AuthHandler.Login)         // 10/min
		r.With(middleware.RateLimit(authRegisterRateLimiter, 5.0/60.0)).Post("/auth/register-phone", p.AuthHandler.RegisterPhone)
		r.With(middleware.RateLimit(authLoginRateLimiter, 10.0/60.0)).Post("/auth/login-phone", p.AuthHandler.LoginPhone)
		r.With(middleware.RateLimit(authLoginRateLimiter, 10.0/60.0)).Post("/auth/verify-phone", p.AuthHandler.VerifyPhone)
		r.With(auth).Get("/auth/me", p.AuthHandler.Me)
		r.With(auth).Put("/auth/me", p.AuthHandler.UpdateProfile)
		r.With(auth).Post("/auth/me/avatar", p.AuthHandler.UploadAvatar)
		r.With(auth).Delete("/auth/me/avatar", p.AuthHandler.DeleteAvatar)

		// Password reset (public, rate-limited)
		r.With(middleware.RateLimit(authRegisterRateLimiter, 3.0/60.0)).Post("/auth/forgot-password", p.AuthHandler.ForgotPassword) // 3/min
		r.With(middleware.RateLimit(authLoginRateLimiter, 10.0/60.0)).Post("/auth/reset-password", p.AuthHandler.ResetPassword)     // 10/min

		// Account deletion (authenticated)
		r.With(auth).Post("/auth/delete-account", p.AuthHandler.DeleteAccount)
		r.With(auth).Post("/auth/restore-account", p.AuthHandler.RestoreAccount)

		// 2FA (authenticated)
		r.With(auth).Post("/auth/2fa/totp/enable", p.AuthHandler.EnableTOTP)
		r.With(auth).Post("/auth/2fa/totp/verify", p.AuthHandler.VerifyAndActivateTOTP)
		r.With(auth).Delete("/auth/2fa/totp", p.AuthHandler.DisableTOTP)
		r.With(auth).Post("/auth/2fa/sms/enable", p.AuthHandler.EnableSMS2FA)

		// 2FA login verification (public, rate-limited)
		r.With(middleware.RateLimit(authLoginRateLimiter, 10.0/60.0)).Post("/auth/2fa/verify", p.AuthHandler.Verify2FALogin)

		// OAuth (public)
		r.Get("/auth/oauth/{provider}", p.OAuthHandler.OAuthRedirect)
		r.Get("/auth/oauth/{provider}/callback", p.OAuthHandler.OAuthCallback)

		// OAuth (authenticated)
		r.With(auth).Post("/auth/link/{provider}", p.OAuthHandler.LinkSocialAccount)
		r.With(auth).Delete("/auth/link/{provider}", p.OAuthHandler.UnlinkSocialAccount)
		r.With(auth).Get("/auth/me/social-accounts", p.OAuthHandler.ListSocialAccounts)

		// User profiles (public)
		r.Get("/users/{id}/profile", p.AuthHandler.GetPublicProfile)

		// Search suggestions (public)
		r.Get("/search/suggestions", p.SearchHandler.GetSuggestions)

		// Public platform settings (whitelisted keys only)
		r.Get("/settings/{key}", p.PlatformSettingsHandler.GetPublic)

		// Isochrone (public)
		r.Get("/isochrone", p.IsochroneHandler.GetIsochrone)

		// Bathhouse comparison (public)
		r.Post("/bathhouses/compare", p.ComparisonHandler.Compare)

		// Bathhouses (public, with optional auth for is_favorite)
		r.With(optionalAuth).Get("/bathhouses", p.BHHandler.Search)
		r.With(optionalAuth).Get("/bathhouses/by-slug/{slug}", p.BHHandler.GetBySlug)
		r.With(optionalAuth).Get("/bathhouses/{id}", p.BHHandler.GetByID)
		r.Get("/bathhouses/{id}/available-slots", p.BHHandler.GetAvailableSlots)
		r.Get("/bathhouses/{id}/meta", p.BHHandler.GetMeta)
		r.Get("/bathhouses/{id}/transport", p.TransportHandler.GetTransport)
		r.Get("/bathhouses/{id}/schema", p.SitemapHandler.GetSchema)

		// City bathhouses (public, SEO-friendly)
		r.With(optionalAuth).Get("/cities/{slug}/bathhouses", p.BHHandler.SearchByCitySlug)

		// Bathhouses (authenticated)
		r.With(auth, middleware.RequireRole(domain.RoleOwner)).Post("/bathhouses", p.BHHandler.Create)
		r.With(auth, middleware.RequireOwnerOrRepresentative()).Put("/bathhouses/{id}", p.BHHandler.Update)
		r.With(auth, middleware.RequireRole(domain.RoleOwner)).Delete("/bathhouses/{id}", p.BHHandler.Delete)

		// My bathhouses (owner/representative)
		r.With(auth, middleware.RequireOwnerOrRepresentative()).Get("/my/bathhouses", p.BHHandler.MyBathhouses)
		r.With(auth, middleware.RequireOwnerOrRepresentative()).Get("/my/bathhouses/{id}/completeness", p.BHHandler.CheckCompleteness)
		r.With(auth, middleware.RequireOwnerOrRepresentative()).Post("/my/bathhouses/{id}/submit", p.BHHandler.SubmitForModeration)
		r.With(auth, middleware.RequireOwnerOrRepresentative()).Post("/my/bathhouses/{id}/duplicate", p.BHHandler.Duplicate)
		r.With(auth, middleware.RequireOwnerOrRepresentative()).Post("/my/bathhouses/{id}/deactivate", p.BHHandler.Deactivate)
		r.With(auth, middleware.RequireOwnerOrRepresentative()).Post("/my/bathhouses/{id}/activate", p.BHHandler.Activate)
		r.With(auth, middleware.RequireOwnerOrRepresentative()).Delete("/my/bathhouses/{id}/archive", p.BHHandler.Archive)
		r.With(auth, middleware.RequireOwnerOrRepresentative()).Get("/my/bathhouses/{id}/history", p.AuditLogHandler.ListByBathhouse)

		// Add-ons (owner/representative)
		r.With(auth, middleware.RequireOwnerOrRepresentative()).Post("/my/bathhouses/{id}/addons", p.AddOnHandler.Create)
		r.With(auth, middleware.RequireOwnerOrRepresentative()).Get("/my/bathhouses/{id}/addons", p.AddOnHandler.ListByBathhouse)
		r.With(auth, middleware.RequireOwnerOrRepresentative()).Put("/my/addons/{id}", p.AddOnHandler.Update)
		r.With(auth, middleware.RequireOwnerOrRepresentative()).Delete("/my/addons/{id}", p.AddOnHandler.Delete)

		// Add-ons (public)
		r.Get("/bathhouses/{id}/addons", p.AddOnHandler.ListPublic)

		// Bathhouse photos (owner/representative)
		r.With(auth, middleware.RequireOwnerOrRepresentative()).Get("/my/bathhouses/{id}/photos", p.PhotoHandler.ListByBathhouseOwner)
		r.With(auth, middleware.RequireOwnerOrRepresentative()).Post("/my/bathhouses/{id}/photos", p.PhotoHandler.Upload)
		r.With(auth, middleware.RequireOwnerOrRepresentative()).Delete("/photos/{id}", p.PhotoHandler.Delete)
		r.With(auth, middleware.RequireOwnerOrRepresentative()).Put("/my/bathhouses/{id}/photos/reorder", p.PhotoHandler.Reorder)

		// Photo orders (owner/representative)
		r.With(auth, middleware.RequireOwnerOrRepresentative()).Post("/my/bathhouses/{id}/photo-order", p.PhotoOrderHandler.Create)
		r.With(auth, middleware.RequireOwnerOrRepresentative()).Get("/my/photo-orders", p.PhotoOrderHandler.ListByOwner)
		r.With(auth, middleware.RequireOwnerOrRepresentative()).Get("/my/photo-orders/{id}", p.PhotoOrderHandler.GetByID)
		r.With(auth, middleware.RequireOwnerOrRepresentative()).Post("/my/photo-orders/{id}/cancel", p.PhotoOrderHandler.OwnerCancel)

		// Bathhouse photos (public)
		r.Get("/bathhouses/{id}/photos", p.PhotoHandler.ListByBathhouse)

		// Promo codes (owner/representative)
		r.With(auth, middleware.RequireOwnerOrRepresentative()).Post("/my/bathhouses/{id}/promo-codes", p.PromoHandler.CreateForBathhouse)
		r.With(auth, middleware.RequireOwnerOrRepresentative()).Get("/my/bathhouses/{id}/promo-codes", p.PromoHandler.ListByBathhouse)
		r.With(auth, middleware.RequireRole(domain.RoleOwner, domain.RoleRepresentative, domain.RoleAdmin)).Delete("/promo-codes/{id}", p.PromoHandler.Deactivate)

		// Promo codes (validation - public, rate-limited)
		r.With(middleware.RateLimit(promoRateLimiter, 20.0/60.0)).Post("/promo-codes/validate", p.PromoHandler.Validate) // 20/min

		// Bookings (authenticated)
		r.With(auth, middleware.RequireRole(domain.RoleClient)).Post("/bookings", p.BookingHandler.Create)
		r.With(auth).Get("/bookings", p.BookingHandler.ListByUser)
		r.With(auth).Patch("/bookings/{id}/cancel", p.BookingHandler.Cancel)
		r.With(auth).Post("/bookings/{id}/pay", p.PaymentHandler.InitiatePayment)
		r.With(auth).Get("/bookings/{id}/payment", p.PaymentHandler.GetBookingPayment)
		r.With(auth, middleware.RequireOwnerOrRepresentative()).Patch("/bookings/{id}/confirm", p.BookingHandler.Confirm)
		r.With(auth, middleware.RequireOwnerOrRepresentative()).Patch("/bookings/{id}/reject", p.BookingHandler.Reject)
		r.With(auth, middleware.RequireOwnerOrRepresentative()).Patch("/bookings/{id}/approve", p.BookingHandler.Approve)
		r.With(auth, middleware.RequireOwnerOrRepresentative()).Patch("/bookings/{id}/complete", p.BookingHandler.Complete)
		r.With(auth, middleware.RequireOwnerOrRepresentative()).Patch("/bookings/{id}/check-in", p.BookingHandler.CheckIn)
		r.With(auth, middleware.RequireOwnerOrRepresentative()).Patch("/bookings/{id}/check-out", p.BookingHandler.CheckOut)
		r.With(auth).Put("/bookings/{id}/modify", p.BookingHandler.Modify)
		r.With(auth).Post("/bookings/{id}/extend", p.BookingHandler.Extend)
		r.With(auth).Post("/bookings/{id}/dispute-noshow", p.BookingHandler.DisputeNoShow)
		r.With(auth).Get("/bookings/{id}/rebook-data", p.BookingHandler.GetRebookData)

		// Booking modification requests (two-party approval)
		r.With(auth, middleware.RequireRole(domain.RoleClient)).Post("/bookings/{id}/modification-request", p.BookingModificationHandler.RequestModification)
		r.With(auth).Get("/bookings/{id}/modification-requests", p.BookingModificationHandler.ListModificationRequests)
		r.With(auth, middleware.RequireOwnerOrRepresentative()).Patch("/bookings/modification-requests/{id}/approve", p.BookingModificationHandler.ApproveModification)
		r.With(auth, middleware.RequireOwnerOrRepresentative()).Patch("/bookings/modification-requests/{id}/reject", p.BookingModificationHandler.RejectModification)

		// Booking extension requests (two-party approval)
		r.With(auth, middleware.RequireRole(domain.RoleClient)).Post("/bookings/{id}/extension-request", p.BookingExtensionHandler.RequestExtension)
		r.With(auth).Get("/bookings/{id}/extension-requests", p.BookingExtensionHandler.ListExtensionRequests)
		r.With(auth, middleware.RequireOwnerOrRepresentative()).Patch("/bookings/extension-requests/{id}/approve", p.BookingExtensionHandler.ApproveExtension)
		r.With(auth, middleware.RequireOwnerOrRepresentative()).Patch("/bookings/extension-requests/{id}/reject", p.BookingExtensionHandler.RejectExtension)

		// Booking share links
		r.With(auth).Post("/bookings/share", p.ShareHandler.CreateShareLink)
		r.Get("/share/booking/{token}", p.ShareHandler.ResolveShareLink)

		// Bathhouse bookings (owner/representative/admin)
		r.With(auth, middleware.RequireOwnerOrRepresentative()).Get("/bathhouses/{id}/bookings", p.BookingHandler.ListByBathhouse)

		// Reviews
		r.Get("/bathhouses/{id}/reviews", p.ReviewHandler.ListByBathhouse)
		r.With(auth, middleware.RequireRole(domain.RoleClient)).Post("/bathhouses/{id}/reviews", p.ReviewHandler.Create)
		r.With(auth).Put("/reviews/{id}", p.ReviewHandler.Update)
		r.With(auth).Delete("/reviews/{id}", p.ReviewHandler.Delete)
		r.With(auth, middleware.RequireOwnerOrRepresentative()).Post("/reviews/{id}/response", p.ReviewHandler.AddOwnerResponse)

		// Review media
		r.With(auth).Post("/reviews/{id}/media", p.ReviewHandler.UploadMedia)
		r.With(auth).Delete("/media/{id}", p.ReviewHandler.DeleteMedia)

		// Client reviews (owner rates client)
		r.With(auth, middleware.RequireOwnerOrRepresentative()).Post("/client-reviews", p.ClientReviewHandler.Create)
		r.With(auth).Get("/bookings/{id}/client-review", p.ClientReviewHandler.GetByBooking)
		r.With(auth).Get("/my/client-reviews", p.ClientReviewHandler.ListMyClientReviews)

		// Bathhouse gallery (public)
		r.Get("/bathhouses/{id}/gallery", p.MediaHandler.BathhouseGallery)

		// Complaints / Reports (authenticated)
		r.With(auth).Post("/reviews/{id}/report", p.ComplaintHandler.ReportReview)
		r.With(auth).Post("/bathhouses/{id}/report", p.ComplaintHandler.ReportBathhouse)
		r.With(auth).Post("/users/{id}/report", p.ComplaintHandler.ReportUser)

		// Favorites (authenticated)
		r.With(auth).Post("/bathhouses/{id}/favorite", p.FavHandler.Toggle)
		r.With(auth).Get("/my/favorites", p.FavHandler.List)

		// Recently viewed & saved searches (authenticated)
		r.With(auth).Get("/my/recently-viewed", p.SavedSearchHandler.ListRecentlyViewed)
		r.With(auth).Post("/my/saved-searches", p.SavedSearchHandler.CreateSavedSearch)
		r.With(auth).Get("/my/saved-searches", p.SavedSearchHandler.ListSavedSearches)
		r.With(auth).Delete("/my/saved-searches/{id}", p.SavedSearchHandler.DeleteSavedSearch)

		// Saved cards (authenticated)
		r.With(auth).Post("/my/saved-cards", p.SavedCardHandler.CreateSavedCard)
		r.With(auth).Get("/my/saved-cards", p.SavedCardHandler.ListSavedCards)
		r.With(auth).Delete("/my/saved-cards/{id}", p.SavedCardHandler.DeleteSavedCard)
		r.With(auth).Post("/my/saved-cards/{id}/default", p.SavedCardHandler.SetDefaultCard)

		// Recommendations
		r.With(auth).Get("/recommendations", p.RecommendationHandler.GetPersonalized)
		r.Get("/bathhouses/{id}/similar", p.RecommendationHandler.GetSimilar)
		r.Get("/popular", p.RecommendationHandler.GetPopular)
		r.With(auth).Get("/my/preferences", p.RecommendationHandler.GetPreferences)
		r.With(auth).Put("/my/preferences", p.RecommendationHandler.UpdatePreferences)

		// Subscriptions (authenticated owner)
		r.With(auth, middleware.RequireOwnerOrRepresentative()).Post("/my/bathhouses/{id}/subscription", p.SubscriptionHandler.Subscribe)
		r.With(auth, middleware.RequireOwnerOrRepresentative()).Get("/my/bathhouses/{id}/subscription", p.SubscriptionHandler.GetSubscription)
		r.With(auth, middleware.RequireOwnerOrRepresentative()).Delete("/my/bathhouses/{id}/subscription", p.SubscriptionHandler.CancelSubscription)
		r.With(auth, middleware.RequireOwnerOrRepresentative()).Get("/my/subscriptions", p.SubscriptionHandler.ListSubscriptions)

		// Promotions (authenticated owner)
		r.With(auth, middleware.RequireOwnerOrRepresentative()).Post("/my/bathhouses/{id}/promotion", p.SubscriptionHandler.CreatePromotion)
		r.With(auth, middleware.RequireOwnerOrRepresentative()).Get("/my/bathhouses/{id}/promotion", p.SubscriptionHandler.GetPromotion)
		r.With(auth, middleware.RequireOwnerOrRepresentative()).Get("/my/bathhouses/{id}/promotions", p.SubscriptionHandler.ListPromotions)
		r.With(auth, middleware.RequireOwnerOrRepresentative()).Put("/my/promotions/{id}", p.SubscriptionHandler.UpdatePromotion)
		r.With(auth, middleware.RequireOwnerOrRepresentative()).Post("/my/promotions/{id}/pause", p.SubscriptionHandler.PausePromotion)
		r.With(auth, middleware.RequireOwnerOrRepresentative()).Post("/my/promotions/{id}/resume", p.SubscriptionHandler.ResumePromotion)

		// Pricing rules (authenticated owner)
		r.With(auth, middleware.RequireOwnerOrRepresentative()).Post("/my/bathhouses/{id}/pricing-rules", p.PricingHandler.CreateRule)
		r.With(auth, middleware.RequireOwnerOrRepresentative()).Get("/my/bathhouses/{id}/pricing-rules", p.PricingHandler.ListRules)
		r.With(auth, middleware.RequireOwnerOrRepresentative()).Put("/pricing-rules/{id}", p.PricingHandler.UpdateRule)
		r.With(auth, middleware.RequireOwnerOrRepresentative()).Delete("/pricing-rules/{id}", p.PricingHandler.DeleteRule)

		// Seasonal tariffs (authenticated owner)
		r.With(auth, middleware.RequireOwnerOrRepresentative()).Post("/my/bathhouses/{id}/seasonal-tariffs", p.PricingHandler.CreateSeasonalTariff)
		r.With(auth, middleware.RequireOwnerOrRepresentative()).Get("/my/bathhouses/{id}/seasonal-tariffs", p.PricingHandler.ListSeasonalTariffs)
		r.With(auth, middleware.RequireOwnerOrRepresentative()).Put("/seasonal-tariffs/{id}", p.PricingHandler.UpdateSeasonalTariff)
		r.With(auth, middleware.RequireOwnerOrRepresentative()).Delete("/seasonal-tariffs/{id}", p.PricingHandler.DeleteSeasonalTariff)

		// Smart pricing recommendation (authenticated owner)
		r.With(auth, middleware.RequireOwnerOrRepresentative()).Get("/my/bathhouses/{id}/price-recommendation", p.PricingHandler.GetPriceRecommendation)

		// Holiday multiplier (authenticated owner)
		r.With(auth, middleware.RequireOwnerOrRepresentative()).Put("/my/bathhouses/{id}/holiday-multiplier", p.HolidayHandler.SetBathhouseMultiplier)

		// Widget API keys (authenticated owner)
		r.With(auth, middleware.RequireOwnerOrRepresentative()).Get("/my/bathhouses/{id}/widget-key", p.BHHandler.GetWidgetKey)
		r.With(auth, middleware.RequireOwnerOrRepresentative()).Post("/my/bathhouses/{id}/widget-key/regenerate", p.BHHandler.RegenerateWidgetKey)
		r.With(auth, middleware.RequireOwnerOrRepresentative()).Get("/my/bathhouses/{id}/widget-code", p.BHHandler.GetWidgetCode)

		// Price calculator (public)
		r.Get("/bathhouses/{id}/price-calculator", p.PricingHandler.CalculatePrice)

		// Widget API (public, API key based)
		r.With(widgetRateLimit).Get("/widget/{api_key}/bathhouse", p.WidgetHandler.GetBathhouse)
		r.With(widgetRateLimit).Get("/widget/{api_key}/slots", p.WidgetHandler.GetAvailableSlots)
		r.With(widgetRateLimit).Post("/widget/{api_key}/booking", p.WidgetHandler.CreateBooking)

		// Widget static files (public)
		r.Get("/widget.js", p.WidgetHandler.ServeScript)
		r.Get("/widget.css", p.WidgetHandler.ServeStyles)

		// User profile and statistics (authenticated)
		r.With(auth).Get("/my/stats", p.AuthHandler.GetMyStats)
		r.With(auth).Get("/my/profile-completeness", p.AuthHandler.GetProfileCompleteness)
		r.With(auth).Post("/my/onboarding/complete", p.AuthHandler.CompleteOnboarding)

		// Sessions (authenticated)
		r.With(auth).Get("/my/sessions", p.SessionHandler.ListSessions)
		r.With(auth).Delete("/my/sessions", p.SessionHandler.TerminateAllOtherSessions)
		r.With(auth).Delete("/my/sessions/{id}", p.SessionHandler.TerminateSession)

		// Region (authenticated)
		r.With(auth).Get("/my/region", p.RegionHandler.GetRegion)
		r.With(auth).Put("/my/region", p.RegionHandler.SwitchRegion)

		// Wallet (authenticated)
		r.With(auth).Get("/my/wallet", p.WalletHandler.GetWallet)
		r.With(auth).Get("/my/wallet/transactions", p.WalletHandler.ListTransactions)
		r.With(auth).Post("/my/wallet/topup", p.WalletHandler.TopUp)
		r.With(auth).Get("/my/wallet/holds", p.WalletHandler.ListHolds)
		r.With(auth).Get("/my/wallet/export", p.FinancialReportHandler.ExportWalletTransactions)

		// Financial reports (authenticated owner)
		r.With(auth, middleware.RequireRole(domain.RoleOwner)).Get("/my/finance/acts/{bathhouse_id}", p.FinancialReportHandler.GenerateAct)
		r.With(auth, middleware.RequireRole(domain.RoleOwner)).Get("/my/finance/export-xml", p.FinancialReportHandler.ExportXML1C)

		// KYC (authenticated owner)
		r.With(auth, middleware.RequireRole(domain.RoleOwner)).Post("/my/kyc", p.KYCHandler.SubmitKYC)
		r.With(auth, middleware.RequireRole(domain.RoleOwner)).Get("/my/kyc", p.KYCHandler.GetKYCStatus)

		// Offer acceptance (authenticated owner)
		r.With(auth, middleware.RequireRole(domain.RoleOwner)).Post("/my/offer/accept", p.OfferHandler.AcceptOffer)
		r.With(auth, middleware.RequireRole(domain.RoleOwner)).Get("/my/offer/status", p.OfferHandler.GetOfferStatus)

		// Payment details (authenticated owner)
		r.With(auth, middleware.RequireRole(domain.RoleOwner)).Put("/my/payment-details", p.PaymentDetailsHandler.SetPaymentDetails)
		r.With(auth, middleware.RequireRole(domain.RoleOwner)).Get("/my/payment-details", p.PaymentDetailsHandler.GetPaymentDetails)

		// Listing drafts (authenticated owner)
		r.With(auth, middleware.RequireRole(domain.RoleOwner)).Post("/my/listing-drafts", p.ListingDraftHandler.CreateDraft)
		r.With(auth, middleware.RequireRole(domain.RoleOwner)).Get("/my/listing-drafts", p.ListingDraftHandler.ListDrafts)
		r.With(auth, middleware.RequireRole(domain.RoleOwner)).Get("/my/listing-drafts/{id}", p.ListingDraftHandler.GetDraft)
		r.With(auth, middleware.RequireRole(domain.RoleOwner)).Put("/my/listing-drafts/{id}/step/{step}", p.ListingDraftHandler.SaveStep)
		r.With(auth, middleware.RequireRole(domain.RoleOwner)).Post("/my/listing-drafts/{id}/submit", p.ListingDraftHandler.SubmitDraft)
		r.With(auth, middleware.RequireRole(domain.RoleOwner)).Delete("/my/listing-drafts/{id}", p.ListingDraftHandler.DeleteDraft)

		// Listing import CSV/XLSX (authenticated owner)
		r.With(auth, middleware.RequireRole(domain.RoleOwner)).Post("/my/listings/import", p.ListingImportHandler.Import)
		r.With(auth, middleware.RequireRole(domain.RoleOwner)).Get("/my/listings/import/template", p.ListingImportHandler.GetImportTemplate)

		// Payouts (authenticated owner)
		r.With(auth, middleware.RequireRole(domain.RoleOwner)).Post("/my/wallet/payout", p.PayoutHandler.RequestPayout)
		r.With(auth, middleware.RequireRole(domain.RoleOwner)).Put("/my/wallet/auto-payout", p.PayoutHandler.SetAutoPayoutThreshold)
		r.With(auth, middleware.RequireRole(domain.RoleOwner)).Get("/my/wallet/payouts", p.PayoutHandler.ListPayouts)
		r.With(auth, middleware.RequireRole(domain.RoleOwner)).Get("/my/wallet/payouts/export", p.FinancialReportHandler.ExportPayouts)

		// Loyalty program (authenticated)
		r.With(auth).Get("/my/loyalty", p.LoyaltyHandler.GetAccount)
		r.With(auth).Get("/my/loyalty/transactions", p.LoyaltyHandler.ListTransactions)
		r.With(auth).Get("/my/loyalty/levels", p.LoyaltyHandler.GetLevels)

		// Payments (authenticated)
		r.With(auth).Get("/my/payments", p.PaymentHandler.ListUserPayments)

		// Referral program (authenticated)
		r.With(auth).Get("/my/referral", p.ReferralHandler.GetCode)
		r.With(auth).Get("/my/referral/stats", p.ReferralHandler.GetStats)
		r.With(auth).Get("/my/referral/balance", p.ReferralHandler.GetBalance)

		// Gift certificates
		r.With(optionalAuth).Post("/certificates/purchase", p.CertificateHandler.Purchase)
		r.With(auth).Post("/certificates/redeem", p.CertificateHandler.Redeem)
		r.Get("/certificates/{code}/balance", p.CertificateHandler.GetBalance)
		r.With(auth).Get("/my/certificates", p.CertificateHandler.ListMyCertificates)

		// Chat (authenticated)
		r.With(auth, middleware.RequireRole(domain.RoleClient, domain.RoleAdmin)).Post("/bathhouses/{id}/chat", p.ChatHandler.StartConversation)
		r.With(auth).Get("/my/conversations", p.ChatHandler.ListConversations)
		r.With(auth).Get("/conversations/{id}/messages", p.ChatHandler.ListMessages)
		r.With(auth).Post("/conversations/{id}/messages", p.ChatHandler.SendMessage)
		r.With(auth).Patch("/conversations/{id}/read", p.ChatHandler.MarkAsRead)
		r.With(auth).Get("/my/unread-messages-count", p.ChatHandler.GetUnreadCount)

		// Analytics (authenticated owner/representative)
		r.With(auth, middleware.RequireOwnerOrRepresentative()).Get("/my/bathhouses/{id}/analytics", p.AnalyticsHandler.GetOwnerDashboard)
		r.With(auth, middleware.RequireOwnerOrRepresentative()).Get("/my/bathhouses/{id}/analytics/daily", p.AnalyticsHandler.GetOwnerDailyStats)
		r.With(auth, middleware.RequireOwnerOrRepresentative()).Get("/my/bathhouses/{id}/analytics/performance", p.AnalyticsHandler.GetOwnerPerformance)

		// CRM Guest Cards (owner/representative)
		r.With(auth, middleware.RequireOwnerOrRepresentative()).Get("/my/crm/guests", p.GuestCardHandler.ListGuests)
		r.With(auth, middleware.RequireOwnerOrRepresentative()).Put("/my/crm/guests/{id}", p.GuestCardHandler.UpdateGuestNotes)
		r.With(auth, middleware.RequireOwnerOrRepresentative()).Get("/my/crm/guests/export", p.GuestCardHandler.ExportCSV)
		r.With(auth, middleware.RequireOwnerOrRepresentative()).Get("/my/crm/stats", p.GuestCardHandler.GetStats)

		// CRM Segments (owner/representative)
		r.With(auth, middleware.RequireOwnerOrRepresentative()).Get("/my/crm/segments", p.GuestCardHandler.ListSegments)
		r.With(auth, middleware.RequireOwnerOrRepresentative()).Get("/my/crm/segments/{slug}/guests", p.GuestCardHandler.GetGuestsInSegment)

		// CRM RFM Analysis & Custom Segments (owner/representative)
		r.With(auth, middleware.RequireOwnerOrRepresentative()).Get("/my/crm/rfm", p.RFMHandler.GetRFMAnalysis)
		r.With(auth, middleware.RequireOwnerOrRepresentative()).Get("/my/crm/segments/custom", p.RFMHandler.ListCustomSegments)
		r.With(auth, middleware.RequireOwnerOrRepresentative()).Post("/my/crm/segments/custom", p.RFMHandler.CreateCustomSegment)
		r.With(auth, middleware.RequireOwnerOrRepresentative()).Put("/my/crm/segments/custom/{id}", p.RFMHandler.UpdateCustomSegment)
		r.With(auth, middleware.RequireOwnerOrRepresentative()).Delete("/my/crm/segments/custom/{id}", p.RFMHandler.DeleteCustomSegment)
		r.With(auth, middleware.RequireOwnerOrRepresentative()).Get("/my/crm/segments/custom/{id}/guests", p.RFMHandler.GetCustomSegmentGuests)

		// CRM Broadcasts (owner/representative)
		r.With(auth, middleware.RequireOwnerOrRepresentative()).Post("/my/crm/broadcasts", p.BroadcastHandler.CreateBroadcast)
		r.With(auth, middleware.RequireOwnerOrRepresentative()).Get("/my/crm/broadcasts", p.BroadcastHandler.ListBroadcasts)
		r.With(auth, middleware.RequireOwnerOrRepresentative()).Get("/my/crm/broadcasts/{id}", p.BroadcastHandler.GetBroadcast)
		r.With(auth, middleware.RequireOwnerOrRepresentative()).Post("/my/crm/broadcasts/{id}/send", p.BroadcastHandler.SendBroadcast)

		// CRM Auto-scenarios (owner/representative)
		r.With(auth, middleware.RequireOwnerOrRepresentative()).Get("/my/crm/auto-scenarios", p.AutoScenarioHandler.ListAutoScenarios)
		r.With(auth, middleware.RequireOwnerOrRepresentative()).Put("/my/crm/auto-scenarios/{type}", p.AutoScenarioHandler.UpdateAutoScenario)

		// CRM Response Templates (owner/representative)
		r.With(auth, middleware.RequireOwnerOrRepresentative()).Post("/my/crm/templates", p.TemplateHandler.CreateTemplate)
		r.With(auth, middleware.RequireOwnerOrRepresentative()).Get("/my/crm/templates", p.TemplateHandler.ListTemplates)
		r.With(auth, middleware.RequireOwnerOrRepresentative()).Put("/my/crm/templates/{id}", p.TemplateHandler.UpdateTemplate)
		r.With(auth, middleware.RequireOwnerOrRepresentative()).Delete("/my/crm/templates/{id}", p.TemplateHandler.DeleteTemplate)

		// FAQ Bot (authenticated)
		r.With(auth).Post("/my/support/faq-match", p.FAQHandler.MatchFAQ)

		// Webhooks (owner/representative)
		r.With(auth, middleware.RequireOwnerOrRepresentative()).Post("/my/webhooks", p.WebhookHandler.CreateWebhook)
		r.With(auth, middleware.RequireOwnerOrRepresentative()).Get("/my/webhooks", p.WebhookHandler.ListWebhooks)
		r.With(auth, middleware.RequireOwnerOrRepresentative()).Get("/my/webhooks/{id}", p.WebhookHandler.GetWebhook)
		r.With(auth, middleware.RequireOwnerOrRepresentative()).Put("/my/webhooks/{id}", p.WebhookHandler.UpdateWebhook)
		r.With(auth, middleware.RequireOwnerOrRepresentative()).Delete("/my/webhooks/{id}", p.WebhookHandler.DeleteWebhook)
		r.With(auth, middleware.RequireOwnerOrRepresentative()).Get("/my/webhooks/{id}/deliveries", p.WebhookHandler.ListDeliveries)
		r.With(auth, middleware.RequireOwnerOrRepresentative()).Post("/my/webhooks/{id}/test", p.WebhookHandler.TestWebhook)

		// PMS Connections (owner/representative)
		r.With(auth, middleware.RequireOwnerOrRepresentative()).Post("/my/pms-connections", p.PMSHandler.CreatePMSConnection)
		r.With(auth, middleware.RequireOwnerOrRepresentative()).Get("/my/pms-connections", p.PMSHandler.ListPMSConnections)
		r.With(auth, middleware.RequireOwnerOrRepresentative()).Get("/my/pms-connections/{id}", p.PMSHandler.GetPMSConnection)
		r.With(auth, middleware.RequireOwnerOrRepresentative()).Put("/my/pms-connections/{id}", p.PMSHandler.UpdatePMSConnection)
		r.With(auth, middleware.RequireOwnerOrRepresentative()).Delete("/my/pms-connections/{id}", p.PMSHandler.DeletePMSConnection)
		r.With(auth, middleware.RequireOwnerOrRepresentative()).Post("/my/pms-connections/{id}/test", p.PMSHandler.TestPMSConnection)
		r.With(auth, middleware.RequireOwnerOrRepresentative()).Post("/my/pms-connections/{id}/sync", p.PMSHandler.SyncPMSConnection)
		r.With(auth, middleware.RequireOwnerOrRepresentative()).Get("/my/pms-connections/{id}/logs", p.PMSHandler.ListPMSSyncLogs)

		// Support Tickets (authenticated)
		r.With(auth).Post("/my/tickets", p.TicketHandler.CreateTicket)
		r.With(auth).Get("/my/tickets", p.TicketHandler.ListUserTickets)
		r.With(auth).Get("/my/tickets/{id}", p.TicketHandler.GetTicket)
		r.With(auth).Post("/my/tickets/{id}/messages", p.TicketHandler.AddUserMessage)
		r.With(auth).Get("/my/tickets/{id}/messages", p.TicketHandler.ListMessages)
		r.With(auth).Post("/my/tickets/{id}/csat", p.TicketHandler.SubmitCSAT)

		// Disputes (authenticated)
		r.With(auth).Post("/bookings/{id}/dispute", p.DisputeHandler.OpenDispute)
		r.With(auth).Get("/my/disputes", p.DisputeHandler.ListUserDisputes)
		r.With(auth).Get("/my/disputes/{id}", p.DisputeHandler.GetDispute)
		r.With(auth).Post("/my/disputes/{id}/evidence", p.DisputeHandler.SubmitEvidence)
		r.With(auth).Get("/my/disputes/{id}/evidence", p.DisputeHandler.ListEvidence)
		r.With(auth).Post("/my/disputes/{id}/appeal", p.DisputeHandler.AppealDispute)

		// Notifications (authenticated)
		r.With(auth).Get("/my/notifications", p.NotifHandler.List)
		r.With(auth).Get("/my/notifications/unread-count", p.NotifHandler.UnreadCount)
		r.With(auth).Patch("/my/notifications/{id}/read", p.NotifHandler.MarkAsRead)
		r.With(auth).Patch("/my/notifications/read-all", p.NotifHandler.MarkAllAsRead)
		r.With(auth).Get("/my/notification-preferences", p.NotifHandler.GetPreferences)
		r.With(auth).Put("/my/notification-preferences", p.NotifHandler.UpdatePreferences)
		r.With(auth).Get("/my/notification-preferences/events", p.NotifHandler.GetEventPreferences)
		r.With(auth).Put("/my/notification-preferences/events", p.NotifHandler.UpdateEventPreferences)

		// Device tokens for push notifications (authenticated)
		r.With(auth).Post("/device-tokens", p.DeviceTokenHandler.Register)
		r.With(auth).Delete("/device-tokens/{id}", p.DeviceTokenHandler.Delete)

		// Calendar export (owner/representative)
		r.With(auth, middleware.RequireOwnerOrRepresentative()).Get("/my/bathhouses/{id}/calendar.ics", p.CalendarHandler.ExportICal)
		r.With(auth, middleware.RequireOwnerOrRepresentative()).Get("/my/bathhouses/{id}/calendar-token", p.CalendarHandler.GetCalendarToken)

		// External calendars (owner/representative)
		r.With(auth, middleware.RequireOwnerOrRepresentative()).Post("/my/bathhouses/{id}/external-calendars", p.CalendarHandler.AddExternalCalendar)
		r.With(auth, middleware.RequireOwnerOrRepresentative()).Get("/my/bathhouses/{id}/external-calendars", p.CalendarHandler.ListExternalCalendars)
		r.With(auth, middleware.RequireOwnerOrRepresentative()).Delete("/my/external-calendars/{id}", p.CalendarHandler.RemoveExternalCalendar)
		r.With(auth, middleware.RequireOwnerOrRepresentative()).Post("/my/bathhouses/{id}/external-calendars/sync", p.CalendarHandler.SyncExternalCalendars)
		r.With(auth, middleware.RequireOwnerOrRepresentative()).Get("/my/bathhouses/{id}/calendar-conflicts", p.CalendarHandler.GetCalendarConflicts)

		// Slot blocks (owner/representative)
		r.With(auth, middleware.RequireOwnerOrRepresentative()).Post("/my/bathhouses/{id}/slot-blocks", p.CalendarHandler.CreateSlotBlock)
		r.With(auth, middleware.RequireOwnerOrRepresentative()).Delete("/my/slot-blocks/{id}", p.CalendarHandler.DeleteSlotBlock)

		// Representatives (owner)
		r.With(auth, middleware.RequireRole(domain.RoleOwner)).Post("/bathhouses/{id}/representatives", p.RepHandler.Invite)
		r.With(auth, middleware.RequireRole(domain.RoleOwner)).Get("/bathhouses/{id}/representatives", p.RepHandler.ListByBathhouse)
		r.With(auth, middleware.RequireRole(domain.RoleOwner)).Delete("/representatives/{id}", p.RepHandler.Revoke)

		// Cities (public)
		r.Get("/cities", p.CityHandler.List)

		// Admin
		r.Route("/admin", func(r chi.Router) {
			r.Use(auth)
			r.Use(middleware.RequireRole(domain.RoleAdmin))
			r.Use(middleware.RequireAdmin2FA(p.Admin2FAChecker))
			r.Use(middleware.LoadAdminSubRole(p.AdminSubRoleResolver))
			r.Use(middleware.AdminAudit(p.AuditLogRepo, p.Log))

			// Admin role management (super_admin only)
			r.With(middleware.RequireAdminPermission(domain.PermAdminRolesManage)).Get("/roles", p.AdminRoleHandler.ListAdminUsers)
			r.With(middleware.RequireAdminPermission(domain.PermAdminRolesManage)).Put("/roles/{id}", p.AdminRoleHandler.SetAdminSubRole)
			r.With(middleware.RequireAdminPermission(domain.PermAdminRolesManage)).Get("/roles/permissions", p.AdminRoleHandler.GetPermissionsMatrix)

			// User management
			r.With(middleware.RequireAdminPermission(domain.PermUserManage)).Get("/users", p.AdminHandler.ListUsers)
			r.With(middleware.RequireAdminPermission(domain.PermUserManage)).Patch("/users/{id}/block", p.AdminHandler.BlockUser)
			r.With(middleware.RequireAdminPermission(domain.PermUserManage)).Patch("/users/{id}/unblock", p.AdminHandler.UnblockUser)
			r.With(middleware.RequireAdminPermission(domain.PermUserManage)).Post("/users/batch", p.AdminHandler.BatchUsers)

			// Bathhouse moderation
			r.With(middleware.RequireAdminPermission(domain.PermBathhouseModerate)).Get("/bathhouses", p.AdminHandler.ListBathhouses)
			r.With(middleware.RequireAdminPermission(domain.PermBathhouseModerate)).Patch("/bathhouses/{id}/approve", p.AdminHandler.ApproveBathhouse)
			r.With(middleware.RequireAdminPermission(domain.PermBathhouseModerate)).Patch("/bathhouses/{id}/reject", p.AdminHandler.RejectBathhouse)
			r.With(middleware.RequireAdminPermission(domain.PermBathhouseModerate)).Post("/listings/batch", p.AdminHandler.BatchListings)

			// City management
			r.With(middleware.RequireAdminPermission(domain.PermCityManage)).Post("/cities", p.AdminHandler.CreateCity)
			r.With(middleware.RequireAdminPermission(domain.PermCityManage)).Put("/cities/{id}", p.AdminHandler.UpdateCity)
			r.With(middleware.RequireAdminPermission(domain.PermCityManage)).Delete("/cities/{id}", p.AdminHandler.DeleteCity)

			// Review moderation
			r.With(middleware.RequireAdminPermission(domain.PermReviewModerate)).Get("/reviews", p.AdminHandler.ListReviews)
			r.With(middleware.RequireAdminPermission(domain.PermReviewModerate)).Get("/reviews/pending-count", p.AdminHandler.GetPendingCount)
			r.With(middleware.RequireAdminPermission(domain.PermReviewModerate)).Patch("/reviews/{id}/approve", p.AdminHandler.ApproveReview)
			r.With(middleware.RequireAdminPermission(domain.PermReviewModerate)).Patch("/reviews/{id}/reject", p.AdminHandler.RejectReview)
			r.With(middleware.RequireAdminPermission(domain.PermReviewModerate)).Post("/reviews/batch-approve", p.AdminHandler.BatchApproveReviews)
			r.With(middleware.RequireAdminPermission(domain.PermReviewModerate)).Post("/reviews/batch-reject", p.AdminHandler.BatchRejectReviews)

			// Analytics
			r.With(middleware.RequireAdminPermission(domain.PermAnalyticsView)).Get("/analytics", p.AnalyticsHandler.GetAdminDashboard)
			r.With(middleware.RequireAdminPermission(domain.PermAnalyticsView)).Get("/analytics/top", p.AnalyticsHandler.GetTopBathhouses)
			r.With(middleware.RequireAdminPermission(domain.PermAnalyticsView)).Get("/analytics/funnel", p.AnalyticsHandler.GetConversionFunnel)
			r.With(middleware.RequireAdminPermission(domain.PermAnalyticsView)).Get("/analytics/cohorts", p.AnalyticsHandler.GetCohortAnalysis)
			r.With(middleware.RequireAdminPermission(domain.PermAnalyticsView)).Get("/analytics/geo", p.AnalyticsHandler.GetGeoDemandSupply)
			r.With(middleware.RequireAdminPermission(domain.PermAnalyticsView)).Get("/analytics/wallet", p.AnalyticsHandler.GetWalletMetrics)

			// Complaints
			r.With(middleware.RequireAdminPermission(domain.PermComplaintManage)).Get("/complaints", p.ComplaintHandler.List)
			r.With(middleware.RequireAdminPermission(domain.PermComplaintManage)).Get("/complaints/{id}", p.ComplaintHandler.GetByID)
			r.With(middleware.RequireAdminPermission(domain.PermComplaintManage)).Patch("/complaints/{id}/resolve", p.ComplaintHandler.Resolve)
			r.With(middleware.RequireAdminPermission(domain.PermComplaintManage)).Patch("/complaints/{id}/dismiss", p.ComplaintHandler.Dismiss)

			// Photo verification
			r.With(middleware.RequireAdminPermission(domain.PermPhotoModerate)).Get("/photos/pending", p.PhotoHandler.GetPending)
			r.With(middleware.RequireAdminPermission(domain.PermPhotoModerate)).Patch("/photos/{id}/verify", p.PhotoHandler.Verify)
			r.With(middleware.RequireAdminPermission(domain.PermPhotoModerate)).Patch("/photos/{id}/reject", p.PhotoHandler.Reject)

			// KYC moderation
			r.With(middleware.RequireAdminPermission(domain.PermKYCModerate)).Get("/kyc/pending", p.KYCHandler.ListPendingKYC)
			r.With(middleware.RequireAdminPermission(domain.PermKYCModerate)).Patch("/kyc/{id}/approve", p.KYCHandler.ApproveKYC)
			r.With(middleware.RequireAdminPermission(domain.PermKYCModerate)).Patch("/kyc/{id}/reject", p.KYCHandler.RejectKYC)

			// Audit log
			r.With(middleware.RequireAdminPermission(domain.PermAuditLogView)).Get("/audit-log", p.AuditLogHandler.ListAdmin)
			r.With(middleware.RequireAdminPermission(domain.PermAuditLogView)).Get("/audit-log/actions", p.AuditLogHandler.ListAdminActions)

			// Promo codes
			r.With(middleware.RequireAdminPermission(domain.PermPromoManage)).Post("/promo-codes", p.PromoHandler.CreateGlobal)

			// Service fee
			r.With(middleware.RequireAdminPermission(domain.PermServiceFeeManage)).Get("/service-fee", p.ServiceFeeHandler.List)
			r.With(middleware.RequireAdminPermission(domain.PermServiceFeeManage)).Put("/service-fee", p.ServiceFeeHandler.Upsert)

			// Holidays
			r.With(middleware.RequireAdminPermission(domain.PermHolidayManage)).Get("/holidays", p.HolidayHandler.ListHolidays)
			r.With(middleware.RequireAdminPermission(domain.PermHolidayManage)).Post("/holidays", p.HolidayHandler.CreateHoliday)
			r.With(middleware.RequireAdminPermission(domain.PermHolidayManage)).Put("/holidays/{id}", p.HolidayHandler.UpdateHoliday)
			r.With(middleware.RequireAdminPermission(domain.PermHolidayManage)).Delete("/holidays/{id}", p.HolidayHandler.DeleteHoliday)

			// Amenities
			r.With(middleware.RequireAdminPermission(domain.PermAmenityManage)).Get("/amenities", p.AmenityHandler.ListAmenities)
			r.With(middleware.RequireAdminPermission(domain.PermAmenityManage)).Post("/amenities", p.AmenityHandler.CreateAmenity)
			r.With(middleware.RequireAdminPermission(domain.PermAmenityManage)).Put("/amenities/{id}", p.AmenityHandler.UpdateAmenity)
			r.With(middleware.RequireAdminPermission(domain.PermAmenityManage)).Delete("/amenities/{id}", p.AmenityHandler.DeleteAmenity)

			// Object types
			r.With(middleware.RequireAdminPermission(domain.PermObjectTypeManage)).Get("/object-types", p.ObjectTypeHandler.ListObjectTypes)
			r.With(middleware.RequireAdminPermission(domain.PermObjectTypeManage)).Post("/object-types", p.ObjectTypeHandler.CreateObjectType)
			r.With(middleware.RequireAdminPermission(domain.PermObjectTypeManage)).Put("/object-types/{id}", p.ObjectTypeHandler.UpdateObjectType)
			r.With(middleware.RequireAdminPermission(domain.PermObjectTypeManage)).Delete("/object-types/{id}", p.ObjectTypeHandler.DeleteObjectType)

			// Admin booking management
			r.With(middleware.RequireAdminPermission(domain.PermBookingManage)).Get("/bookings", p.BookingHandler.AdminListBookings)
			r.With(middleware.RequireAdminPermission(domain.PermBookingManage)).Post("/bookings/{id}/cancel", p.BookingHandler.AdminCancel)
			r.With(middleware.RequireAdminPermission(domain.PermBookingManage)).Post("/bookings/{id}/change-status", p.BookingHandler.AdminChangeStatus)
			r.With(middleware.RequireAdminPermission(domain.PermBookingManage)).Post("/bookings/{id}/refund", p.PaymentHandler.AdminRefund)

			// FAQ management
			r.With(middleware.RequireAdminPermission(domain.PermFAQManage)).Get("/faq", p.FAQHandler.AdminListFAQ)
			r.With(middleware.RequireAdminPermission(domain.PermFAQManage)).Post("/faq", p.FAQHandler.AdminCreateFAQ)
			r.With(middleware.RequireAdminPermission(domain.PermFAQManage)).Get("/faq/{id}", p.FAQHandler.AdminGetFAQ)
			r.With(middleware.RequireAdminPermission(domain.PermFAQManage)).Put("/faq/{id}", p.FAQHandler.AdminUpdateFAQ)
			r.With(middleware.RequireAdminPermission(domain.PermFAQManage)).Delete("/faq/{id}", p.FAQHandler.AdminDeleteFAQ)
			r.With(middleware.RequireAdminPermission(domain.PermFAQManage)).Post("/faq/seed", p.FAQHandler.SeedFAQ)

			// Support tickets
			r.With(middleware.RequireAdminPermission(domain.PermTicketManage)).Get("/tickets", p.TicketHandler.AdminListTickets)
			r.With(middleware.RequireAdminPermission(domain.PermTicketManage)).Get("/tickets/stats", p.TicketHandler.AdminGetStats)
			r.With(middleware.RequireAdminPermission(domain.PermTicketManage)).Get("/tickets/{id}", p.TicketHandler.AdminGetTicket)
			r.With(middleware.RequireAdminPermission(domain.PermTicketManage)).Patch("/tickets/{id}/assign", p.TicketHandler.AdminAssignTicket)
			r.With(middleware.RequireAdminPermission(domain.PermTicketManage)).Patch("/tickets/{id}/escalate", p.TicketHandler.AdminEscalateTicket)
			r.With(middleware.RequireAdminPermission(domain.PermTicketManage)).Patch("/tickets/{id}/resolve", p.TicketHandler.AdminResolveTicket)
			r.With(middleware.RequireAdminPermission(domain.PermTicketManage)).Post("/tickets/{id}/messages", p.TicketHandler.AdminAddMessage)

			// Disputes
			r.With(middleware.RequireAdminPermission(domain.PermDisputeManage)).Get("/disputes", p.DisputeHandler.AdminListDisputes)
			r.With(middleware.RequireAdminPermission(domain.PermDisputeManage)).Get("/disputes/{id}", p.DisputeHandler.AdminGetDispute)
			r.With(middleware.RequireAdminPermission(domain.PermDisputeManage)).Patch("/disputes/{id}/assign", p.DisputeHandler.AdminAssignDispute)
			r.With(middleware.RequireAdminPermission(domain.PermDisputeManage)).Patch("/disputes/{id}/resolve", p.DisputeHandler.AdminResolveDispute)
			r.With(middleware.RequireAdminPermission(domain.PermDisputeManage)).Patch("/disputes/{id}/close", p.DisputeHandler.AdminCloseDispute)

			// Anti-fraud
			r.With(middleware.RequireAdminPermission(domain.PermAntiFraudManage)).Get("/antifraud/flags", p.AntiFraudHandler.ListFlags)
			r.With(middleware.RequireAdminPermission(domain.PermAntiFraudManage)).Patch("/antifraud/flags/{id}", p.AntiFraudHandler.UpdateFlag)
			r.With(middleware.RequireAdminPermission(domain.PermAntiFraudManage)).Get("/antifraud/stoplist", p.AntiFraudHandler.ListStoplist)
			r.With(middleware.RequireAdminPermission(domain.PermAntiFraudManage)).Post("/antifraud/stoplist", p.AntiFraudHandler.CreateStoplistEntry)
			r.With(middleware.RequireAdminPermission(domain.PermAntiFraudManage)).Delete("/antifraud/stoplist/{id}", p.AntiFraudHandler.DeleteStoplistEntry)

			// Chat content filtering
			r.With(middleware.RequireAdminPermission(domain.PermAntiFraudManage)).Get("/chat/filtered", p.AntiFraudHandler.ListFilteredMessages)

			// Platform settings
			r.With(middleware.RequireAdminPermission(domain.PermSettingsManage)).Get("/settings", p.PlatformSettingsHandler.List)
			r.With(middleware.RequireAdminPermission(domain.PermSettingsManage)).Put("/settings/{key}", p.PlatformSettingsHandler.Update)

			// Feature flags
			r.With(middleware.RequireAdminPermission(domain.PermFeatureFlagsManage)).Get("/feature-flags", p.FeatureFlagHandler.List)
			r.With(middleware.RequireAdminPermission(domain.PermFeatureFlagsManage)).Put("/feature-flags/{key}", p.FeatureFlagHandler.Update)

			// Force majeure
			r.With(middleware.RequireAdminPermission(domain.PermForceMajeureManage)).Post("/force-majeure", p.ForceMajeureHandler.Activate)
			r.With(middleware.RequireAdminPermission(domain.PermForceMajeureManage)).Get("/force-majeure", p.ForceMajeureHandler.List)

			// Reconciliation
			r.With(middleware.RequireAdminPermission(domain.PermReconciliationView)).Get("/reconciliation/summary", p.ReconciliationHandler.GetFloatSummary)
			r.With(middleware.RequireAdminPermission(domain.PermReconciliationView)).Post("/reconciliation/snapshot", p.ReconciliationHandler.TakeSnapshot)
			r.With(middleware.RequireAdminPermission(domain.PermReconciliationView)).Get("/reconciliation/snapshots", p.ReconciliationHandler.ListSnapshots)
			r.With(middleware.RequireAdminPermission(domain.PermReconciliationView)).Post("/reconciliation/reconcile", p.ReconciliationHandler.Reconcile)
			r.With(middleware.RequireAdminPermission(domain.PermReconciliationView)).Get("/reconciliation/reports", p.ReconciliationHandler.ListReports)

			// Bank statement reconciliation
			r.With(middleware.RequireAdminPermission(domain.PermFinanceManage)).Post("/finance/bank-statement", p.BankReconciliationHandler.UploadBankStatement)
			r.With(middleware.RequireAdminPermission(domain.PermFinanceManage)).Get("/finance/reconciliation", p.BankReconciliationHandler.ListUnmatched)
			r.With(middleware.RequireAdminPermission(domain.PermFinanceManage)).Put("/finance/reconciliation/{id}/match", p.BankReconciliationHandler.ManualMatch)

			// Admin wallet management
			r.With(middleware.RequireAdminPermission(domain.PermWalletManage)).Get("/wallets/{id}", p.WalletHandler.AdminGetWallet)
			r.With(middleware.RequireAdminPermission(domain.PermWalletManage)).Post("/wallets/{id}/credit", p.WalletHandler.AdminCreditWallet)
			r.With(middleware.RequireAdminPermission(domain.PermWalletManage)).Post("/wallets/{id}/debit", p.WalletHandler.AdminDebitWallet)
			r.With(middleware.RequireAdminPermission(domain.PermWalletManage)).Post("/wallets/{id}/freeze", p.WalletHandler.AdminFreezeWallet)
			r.With(middleware.RequireAdminPermission(domain.PermWalletManage)).Post("/wallets/{id}/unfreeze", p.WalletHandler.AdminUnfreezeWallet)
			r.With(middleware.RequireAdminPermission(domain.PermWalletManage)).Post("/wallets/batch-credit", p.WalletHandler.AdminBatchCreditWallets)

			// Photo orders
			r.With(middleware.RequireAdminPermission(domain.PermPhotoOrderManage)).Get("/photo-orders", p.PhotoOrderHandler.AdminList)
			r.With(middleware.RequireAdminPermission(domain.PermPhotoOrderManage)).Put("/photo-orders/{id}", p.PhotoOrderHandler.AdminUpdate)

			// SEO prerender cache invalidation
			r.Post("/prerender/invalidate/{slug}", p.PrerenderHandler.InvalidateBathhouseCache)

			// Admin notifications
			r.With(middleware.RequireAdminPermission(domain.PermAdminNotificationsView)).Get("/notifications", p.AdminNotificationHandler.ListAdminNotifications)
			r.With(middleware.RequireAdminPermission(domain.PermAdminNotificationsView)).Put("/notifications/{id}/read", p.AdminNotificationHandler.MarkNotificationRead)
			r.With(middleware.RequireAdminPermission(domain.PermAdminNotificationsView)).Put("/notifications/read-all", p.AdminNotificationHandler.MarkAllNotificationsRead)
			r.With(middleware.RequireAdminPermission(domain.PermAdminNotificationsView)).Get("/notifications/unread-count", p.AdminNotificationHandler.GetUnreadCount)
		})
	})

	// Pass mux reference to GoAdmin for deferred mounting in OnStart lifecycle.
	// GoAdmin's Engine.Use() requires AddConfig to be called first (which happens in OnStart),
	// so we cannot mount GoAdmin routes here during the fx Provide phase.
	if p.GoAdmin != nil {
		p.GoAdmin.Mux = r
	}

	return r
}
