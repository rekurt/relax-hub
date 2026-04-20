package server

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/rekurt/relax-hub/internal/domain"
	"github.com/rekurt/relax-hub/internal/middleware"
)

// mountAPIRoutes registers /api/v1 routes: auth, public bathhouse reads, bookings, reviews, etc.
// /api/v1/my/* and /api/v1/admin/* are handled by mountMyRoutes and mountAdminRoutes.
func mountAPIRoutes(
	r chi.Router,
	p RouterParams,
	auth func(http.Handler) http.Handler,
	optionalAuth func(http.Handler) http.Handler,
	widgetRateLimit func(http.Handler) http.Handler,
	webhookRateLimiter *middleware.RateLimiter,
	authRegisterRateLimiter *middleware.RateLimiter,
	authLoginRateLimiter *middleware.RateLimiter,
	promoRateLimiter *middleware.RateLimiter,
) {
	// WebSocket (auth via query parameter)
	r.Get("/ws/notifications", p.WSHandler.HandleWS)

	// Webhooks (public, called by payment providers)
	r.With(middleware.RateLimit(webhookRateLimiter, 0.5)).Post("/webhooks/yookassa", p.PaymentHandler.HandleWebhook)
	r.With(middleware.RateLimit(webhookRateLimiter, 0.5)).Post("/webhooks/bepaid", p.PaymentHandler.HandleBePaidWebhook)

	// Apple Pay merchant validation (public, called during Apple Pay session init)
	r.Post("/apple-pay/validate-merchant", p.PaymentHandler.ValidateApplePayMerchant)

	// Auth (public, rate-limited)
	r.With(middleware.RateLimit(authRegisterRateLimiter, 5.0/60.0)).Post("/auth/register", p.AuthHandler.Register)
	r.With(middleware.RateLimit(authLoginRateLimiter, 10.0/60.0)).Post("/auth/login", p.AuthHandler.Login)
	r.With(middleware.RateLimit(authRegisterRateLimiter, 5.0/60.0)).Post("/auth/register-phone", p.AuthHandler.RegisterPhone)
	r.With(middleware.RateLimit(authLoginRateLimiter, 10.0/60.0)).Post("/auth/login-phone", p.AuthHandler.LoginPhone)
	r.With(middleware.RateLimit(authLoginRateLimiter, 10.0/60.0)).Post("/auth/phone/start", p.AuthHandler.StartPhone)
	r.With(middleware.RateLimit(authLoginRateLimiter, 10.0/60.0)).Post("/auth/verify-phone", p.AuthHandler.VerifyPhone)
	r.With(auth).Get("/auth/me", p.AuthHandler.Me)
	r.With(auth).Put("/auth/me", p.AuthHandler.UpdateProfile)
	r.With(auth).Post("/auth/me/avatar", p.AuthHandler.UploadAvatar)
	r.With(auth).Delete("/auth/me/avatar", p.AuthHandler.DeleteAvatar)

	// Password reset (public, rate-limited)
	r.With(middleware.RateLimit(authRegisterRateLimiter, 3.0/60.0)).Post("/auth/forgot-password", p.AuthHandler.ForgotPassword)
	r.With(middleware.RateLimit(authLoginRateLimiter, 10.0/60.0)).Post("/auth/reset-password", p.AuthHandler.ResetPassword)

	// Account deletion (authenticated)
	r.With(auth).Post("/auth/delete-account", p.AuthHandler.DeleteAccount)
	r.With(auth).Post("/auth/restore-account", p.AuthHandler.RestoreAccount)

	// 2FA (authenticated)
	r.With(auth).Post("/auth/2fa/totp/enable", p.AuthHandler.EnableTOTP)
	r.With(auth).Post("/auth/2fa/totp/verify", p.AuthHandler.VerifyAndActivateTOTP)
	r.With(auth).Delete("/auth/2fa/totp", p.AuthHandler.DisableTOTP)
	r.With(auth).Post("/auth/2fa/sms/enable", p.AuthHandler.EnableSMS2FA)
	r.With(middleware.RateLimit(authLoginRateLimiter, 10.0/60.0)).Post("/auth/2fa/verify", p.AuthHandler.Verify2FALogin)

	// OAuth (public)
	r.Get("/auth/oauth/providers", p.OAuthHandler.ListConfiguredProviders)
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

	// FAQ (public)
	r.Get("/faq", p.FAQHandler.PublicListFAQ)

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

	// Add-ons (public)
	r.Get("/bathhouses/{id}/addons", p.AddOnHandler.ListPublic)

	// Bathhouse photos (public)
	r.Get("/bathhouses/{id}/photos", p.PhotoHandler.ListByBathhouse)

	// Promo codes (validation - public, rate-limited)
	r.With(middleware.RateLimit(promoRateLimiter, 20.0/60.0)).Post("/promo-codes/validate", p.PromoHandler.Validate)

	// Promo codes (owner/representative/admin deactivation)
	r.With(auth, middleware.RequireRole(domain.RoleOwner, domain.RoleRepresentative, domain.RoleAdmin)).Delete("/promo-codes/{id}", p.PromoHandler.Deactivate)

	// Promotion banners (public, home page)
	r.Get("/promotions/banners", p.SubscriptionHandler.GetPromotionBanners)

	// Bookings (authenticated)
	r.With(auth, middleware.RequireRole(domain.RoleClient)).Post("/bookings", p.BookingHandler.Create)
	r.With(auth).Get("/bookings", p.BookingHandler.ListByUser)
	r.With(auth).Get("/bookings/{id}", p.BookingHandler.GetByID)
	r.With(auth).Patch("/bookings/{id}/cancel", p.BookingHandler.Cancel)
	r.With(auth).Post("/bookings/{id}/pay", p.PaymentHandler.InitiatePayment)
	r.With(auth).Get("/bookings/{id}/payment", p.PaymentHandler.GetBookingPayment)
	r.With(auth, middleware.RequireOwnerOrRepresentative()).Patch("/bookings/{id}/confirm", p.BookingHandler.Confirm)
	r.With(auth, middleware.RequireOwnerOrRepresentative()).Patch("/bookings/{id}/reject", p.BookingHandler.Reject)
	r.With(auth, middleware.RequireOwnerOrRepresentative()).Patch("/bookings/{id}/approve", p.BookingHandler.Approve)
	r.With(auth, middleware.RequireOwnerOrRepresentative()).Patch("/bookings/{id}/complete", p.BookingHandler.Complete)
	r.With(auth, middleware.RequireOwnerOrRepresentative()).Patch("/bookings/{id}/check-in", p.BookingHandler.CheckIn)
	r.With(auth, middleware.RequireOwnerOrRepresentative()).Patch("/bookings/{id}/check-out", p.BookingHandler.CheckOut)
	r.With(auth, middleware.RequireRole(domain.RoleAdmin)).Put("/bookings/{id}/modify", p.BookingHandler.Modify)
	r.With(auth, middleware.RequireRole(domain.RoleAdmin)).Post("/bookings/{id}/extend", p.BookingHandler.Extend)
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

	// Bathhouse gallery (public)
	r.Get("/bathhouses/{id}/gallery", p.MediaHandler.BathhouseGallery)

	// Complaints / Reports (authenticated)
	r.With(auth).Post("/reviews/{id}/report", p.ComplaintHandler.ReportReview)
	r.With(auth).Post("/bathhouses/{id}/report", p.ComplaintHandler.ReportBathhouse)
	r.With(auth).Post("/users/{id}/report", p.ComplaintHandler.ReportUser)

	// Favorites (authenticated)
	r.With(auth).Post("/bathhouses/{id}/favorite", p.FavHandler.Toggle)

	// Recommendations (public)
	r.Get("/bathhouses/{id}/similar", p.RecommendationHandler.GetSimilar)
	r.Get("/popular", p.RecommendationHandler.GetPopular)

	// Price calculator (public)
	r.Get("/bathhouses/{id}/price-calculator", p.PricingHandler.CalculatePrice)

	// Widget API (public, API key based)
	r.With(widgetRateLimit).Get("/widget/{api_key}/bathhouse", p.WidgetHandler.GetBathhouse)
	r.With(widgetRateLimit).Get("/widget/{api_key}/slots", p.WidgetHandler.GetAvailableSlots)
	r.With(widgetRateLimit).Post("/widget/{api_key}/booking", p.WidgetHandler.CreateBooking)

	// Widget static files (public)
	r.Get("/widget.js", p.WidgetHandler.ServeScript)
	r.Get("/widget.css", p.WidgetHandler.ServeStyles)

	// Disputes (authenticated - opening only; /my/disputes/* are in mountMyRoutes)
	r.With(auth).Post("/bookings/{id}/dispute", p.DisputeHandler.OpenDispute)

	// Gift certificates (optionally authenticated)
	r.With(optionalAuth).Post("/certificates/orders", p.CertificateHandler.CreateOrder)
	r.With(optionalAuth).Post("/certificates/orders/{id}/pay", p.CertificateHandler.InitiateOrderPayment)
	r.Get("/certificates/orders/{id}", p.CertificateHandler.GetOrder)
	r.With(optionalAuth).Post("/certificates/purchase", p.CertificateHandler.Purchase)
	r.Get("/certificates/{code}/balance", p.CertificateHandler.GetBalance)

	// Chat start (authenticated client)
	r.With(auth, middleware.RequireRole(domain.RoleClient, domain.RoleAdmin)).Post("/bathhouses/{id}/chat", p.ChatHandler.StartConversation)

	// Representatives (owner)
	r.With(auth, middleware.RequireRole(domain.RoleOwner)).Post("/bathhouses/{id}/representatives", p.RepHandler.Invite)
	r.With(auth, middleware.RequireRole(domain.RoleOwner)).Get("/bathhouses/{id}/representatives", p.RepHandler.ListByBathhouse)
	r.With(auth, middleware.RequireRole(domain.RoleOwner)).Delete("/representatives/{id}", p.RepHandler.Revoke)

	// Cities (public)
	r.Get("/cities", p.CityHandler.List)
}
