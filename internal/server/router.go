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
	PaymentDetailsHandler    *handler.PaymentDetailsHandler
	ListingDraftHandler      *handler.ListingDraftHandler
	AuditLogHandler          *handler.AuditLogHandler
	AddOnHandler             *handler.AddOnHandler
	SearchHandler            *handler.SearchHandler
	ComparisonHandler        *handler.ComparisonHandler
	SavedSearchHandler       *handler.SavedSearchHandler
	HolidayHandler           *handler.HolidayHandler
	ServiceFeeHandler        *handler.ServiceFeeHandler
	SessionValidator         middleware.SessionValidator `optional:"true"`
	GoAdmin               *admin.GoAdmin             `optional:"true"`
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
		r.With(middleware.RateLimit(webhookRateLimiter, 0.5)).Post("/webhooks/yookassa", p.PaymentHandler.HandleWebhook) // 30/min

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
		r.Post("/auth/reset-password", p.AuthHandler.ResetPassword)

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

		// Bathhouse comparison (public)
		r.Post("/bathhouses/compare", p.ComparisonHandler.Compare)

		// Bathhouses (public, with optional auth for is_favorite)
		r.With(optionalAuth).Get("/bathhouses", p.BHHandler.Search)
		r.With(optionalAuth).Get("/bathhouses/by-slug/{slug}", p.BHHandler.GetBySlug)
		r.With(optionalAuth).Get("/bathhouses/{id}", p.BHHandler.GetByID)
		r.Get("/bathhouses/{id}/available-slots", p.BHHandler.GetAvailableSlots)
		r.Get("/bathhouses/{id}/meta", p.BHHandler.GetMeta)
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
		r.With(auth).Delete("/photos/{id}", p.PhotoHandler.Delete)
		r.With(auth, middleware.RequireOwnerOrRepresentative()).Put("/my/bathhouses/{id}/photos/reorder", p.PhotoHandler.Reorder)

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
		r.With(auth).Post("/bookings/{id}/extend", p.BookingHandler.Extend)
		r.With(auth).Post("/bookings/{id}/dispute-noshow", p.BookingHandler.DisputeNoShow)

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

		// Pricing rules (authenticated owner)
		r.With(auth, middleware.RequireOwnerOrRepresentative()).Post("/my/bathhouses/{id}/pricing-rules", p.PricingHandler.CreateRule)
		r.With(auth, middleware.RequireOwnerOrRepresentative()).Get("/my/bathhouses/{id}/pricing-rules", p.PricingHandler.ListRules)
		r.With(auth, middleware.RequireOwnerOrRepresentative()).Put("/pricing-rules/{id}", p.PricingHandler.UpdateRule)
		r.With(auth, middleware.RequireOwnerOrRepresentative()).Delete("/pricing-rules/{id}", p.PricingHandler.DeleteRule)

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

		// Sessions (authenticated)
		r.With(auth).Get("/my/sessions", p.SessionHandler.ListSessions)
		r.With(auth).Delete("/my/sessions", p.SessionHandler.TerminateAllOtherSessions)
		r.With(auth).Delete("/my/sessions/{id}", p.SessionHandler.TerminateSession)

		// Wallet (authenticated)
		r.With(auth).Get("/my/wallet", p.WalletHandler.GetWallet)
		r.With(auth).Get("/my/wallet/transactions", p.WalletHandler.ListTransactions)
		r.With(auth).Post("/my/wallet/topup", p.WalletHandler.TopUp)
		r.With(auth).Get("/my/wallet/holds", p.WalletHandler.ListHolds)

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

		// Payouts (authenticated owner)
		r.With(auth, middleware.RequireRole(domain.RoleOwner)).Post("/my/wallet/payout", p.PayoutHandler.RequestPayout)
		r.With(auth, middleware.RequireRole(domain.RoleOwner)).Put("/my/wallet/auto-payout", p.PayoutHandler.SetAutoPayoutThreshold)
		r.With(auth, middleware.RequireRole(domain.RoleOwner)).Get("/my/wallet/payouts", p.PayoutHandler.ListPayouts)

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

		// Notifications (authenticated)
		r.With(auth).Get("/my/notifications", p.NotifHandler.List)
		r.With(auth).Get("/my/notifications/unread-count", p.NotifHandler.UnreadCount)
		r.With(auth).Patch("/my/notifications/{id}/read", p.NotifHandler.MarkAsRead)
		r.With(auth).Patch("/my/notifications/read-all", p.NotifHandler.MarkAllAsRead)
		r.With(auth).Get("/my/notification-preferences", p.NotifHandler.GetPreferences)
		r.With(auth).Put("/my/notification-preferences", p.NotifHandler.UpdatePreferences)

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

			r.Get("/users", p.AdminHandler.ListUsers)
			r.Patch("/users/{id}/block", p.AdminHandler.BlockUser)
			r.Patch("/users/{id}/unblock", p.AdminHandler.UnblockUser)

			r.Get("/bathhouses", p.AdminHandler.ListBathhouses)
			r.Patch("/bathhouses/{id}/approve", p.AdminHandler.ApproveBathhouse)
			r.Patch("/bathhouses/{id}/reject", p.AdminHandler.RejectBathhouse)

			r.Post("/cities", p.AdminHandler.CreateCity)
			r.Put("/cities/{id}", p.AdminHandler.UpdateCity)
			r.Delete("/cities/{id}", p.AdminHandler.DeleteCity)

			// Review moderation
			r.Get("/reviews", p.AdminHandler.ListReviews)
			r.Get("/reviews/pending-count", p.AdminHandler.GetPendingCount)
			r.Patch("/reviews/{id}/approve", p.AdminHandler.ApproveReview)
			r.Patch("/reviews/{id}/reject", p.AdminHandler.RejectReview)
			r.Post("/reviews/batch-approve", p.AdminHandler.BatchApproveReviews)
			r.Post("/reviews/batch-reject", p.AdminHandler.BatchRejectReviews)

			// Analytics (admin only)
			r.Get("/analytics", p.AnalyticsHandler.GetAdminDashboard)
			r.Get("/analytics/top", p.AnalyticsHandler.GetTopBathhouses)

			// Complaints (admin only)
			r.Get("/complaints", p.ComplaintHandler.List)
			r.Get("/complaints/{id}", p.ComplaintHandler.GetByID)
			r.Patch("/complaints/{id}/resolve", p.ComplaintHandler.Resolve)
			r.Patch("/complaints/{id}/dismiss", p.ComplaintHandler.Dismiss)

			// Photo verification (admin only)
			r.Get("/photos/pending", p.PhotoHandler.GetPending)
			r.Patch("/photos/{id}/verify", p.PhotoHandler.Verify)
			r.Patch("/photos/{id}/reject", p.PhotoHandler.Reject)

			// KYC moderation (admin only)
			r.Get("/kyc/pending", p.KYCHandler.ListPendingKYC)
			r.Patch("/kyc/{id}/approve", p.KYCHandler.ApproveKYC)
			r.Patch("/kyc/{id}/reject", p.KYCHandler.RejectKYC)

			// Audit log (admin only)
			r.Get("/audit-log", p.AuditLogHandler.ListAdmin)

			// Promo codes (admin only)
			r.Post("/promo-codes", p.PromoHandler.CreateGlobal)

			// Service fee (admin only)
			r.Get("/service-fee", p.ServiceFeeHandler.List)
			r.Put("/service-fee", p.ServiceFeeHandler.Upsert)

			// Holidays (admin only)
			r.Get("/holidays", p.HolidayHandler.ListHolidays)
			r.Post("/holidays", p.HolidayHandler.CreateHoliday)
			r.Put("/holidays/{id}", p.HolidayHandler.UpdateHoliday)
			r.Delete("/holidays/{id}", p.HolidayHandler.DeleteHoliday)

			// Admin refund
			r.Post("/bookings/{id}/refund", p.PaymentHandler.AdminRefund)
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
