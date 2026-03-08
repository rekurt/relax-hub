package server

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/nikitaaldaev/bani/internal/admin"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/handler"
	"github.com/nikitaaldaev/bani/internal/logger"
	"github.com/nikitaaldaev/bani/internal/middleware"
	"go.uber.org/fx"
)

type RouterParams struct {
	fx.In

	Log                    *logger.Logger
	CORS                   *middleware.CORSMiddleware
	AuthService            middleware.AuthService
	AuthHandler            *handler.AuthHandler
	BHHandler              *handler.BathhouseHandler
	BookingHandler         *handler.BookingHandler
	ReviewHandler          *handler.ReviewHandler
	FavHandler             *handler.FavoriteHandler
	RepHandler             *handler.RepresentativeHandler
	CityHandler            *handler.CityHandler
	AdminHandler           *handler.AdminHandler
	HealthHandler          *handler.HealthHandler
	WSHandler              *handler.WSHandler
	NotifHandler           *handler.NotificationHandler
	OAuthHandler           *handler.OAuthHandler
	RecommendationHandler  *handler.RecommendationHandler
	SubscriptionHandler    *handler.SubscriptionHandler
	PricingHandler         *handler.PricingHandler
	WidgetHandler          *handler.WidgetHandler
	LoyaltyHandler         *handler.LoyaltyHandler
	ChatHandler            *handler.ChatHandler
	AnalyticsHandler       *handler.AnalyticsHandler
	ComplaintHandler       *handler.ComplaintHandler
	ReferralHandler        *handler.ReferralHandler
	CertificateHandler     *handler.CertificateHandler
	GoAdmin                *admin.GoAdmin `optional:"true"`
}

func NewRouter(p RouterParams) http.Handler {
	r := chi.NewRouter()

	r.Use(chiMiddleware.RequestID)
	r.Use(middleware.Logging(*p.Log))
	r.Use(middleware.RecoveryMiddleware(middleware.IsDevEnvironment(), p.Log))
	r.Use(p.CORS.Handler)

	r.Get("/health", p.HealthHandler.Health)
	r.Get("/ready", p.HealthHandler.Ready)

	auth := middleware.RequireAuth(p.AuthService)
	optionalAuth := middleware.OptionalAuth(p.AuthService)

	// Create rate limiter for widget endpoints
	widgetRateLimiter := middleware.NewRateLimiter()
	widgetRateLimit := middleware.WidgetRateLimit(widgetRateLimiter, 10) // 10 requests per second per API key

	r.Route("/api/v1", func(r chi.Router) {
		// WebSocket (auth via query parameter)
		r.Get("/ws/notifications", p.WSHandler.HandleWS)

		// Auth (public)
		r.Post("/auth/register", p.AuthHandler.Register)
		r.Post("/auth/login", p.AuthHandler.Login)
		r.With(auth).Get("/auth/me", p.AuthHandler.Me)
		r.With(auth).Put("/auth/me", p.AuthHandler.UpdateProfile)
		r.With(auth).Post("/auth/me/avatar", p.AuthHandler.UploadAvatar)
		r.With(auth).Delete("/auth/me/avatar", p.AuthHandler.DeleteAvatar)

		// OAuth (public)
		r.Get("/auth/oauth/{provider}", p.OAuthHandler.OAuthRedirect)
		r.Get("/auth/oauth/{provider}/callback", p.OAuthHandler.OAuthCallback)

		// OAuth (authenticated)
		r.With(auth).Post("/auth/link/{provider}", p.OAuthHandler.LinkSocialAccount)
		r.With(auth).Delete("/auth/link/{provider}", p.OAuthHandler.UnlinkSocialAccount)
		r.With(auth).Get("/auth/me/social-accounts", p.OAuthHandler.ListSocialAccounts)

		// User profiles (public)
		r.Get("/users/{id}/profile", p.AuthHandler.GetPublicProfile)

		// Bathhouses (public, with optional auth for is_favorite)
		r.With(optionalAuth).Get("/bathhouses", p.BHHandler.Search)
		r.With(optionalAuth).Get("/bathhouses/{id}", p.BHHandler.GetByID)
		r.Get("/bathhouses/{id}/available-slots", p.BHHandler.GetAvailableSlots)

		// Bathhouses (authenticated)
		r.With(auth, middleware.RequireRole(domain.RoleOwner)).Post("/bathhouses", p.BHHandler.Create)
		r.With(auth, middleware.RequireOwnerOrRepresentative()).Put("/bathhouses/{id}", p.BHHandler.Update)
		r.With(auth, middleware.RequireRole(domain.RoleOwner)).Delete("/bathhouses/{id}", p.BHHandler.Delete)

		// My bathhouses (owner/representative)
		r.With(auth, middleware.RequireOwnerOrRepresentative()).Get("/my/bathhouses", p.BHHandler.MyBathhouses)

		// Bookings (authenticated)
		r.With(auth, middleware.RequireRole(domain.RoleClient)).Post("/bookings", p.BookingHandler.Create)
		r.With(auth).Get("/bookings", p.BookingHandler.ListByUser)
		r.With(auth).Patch("/bookings/{id}/cancel", p.BookingHandler.Cancel)
		r.With(auth, middleware.RequireOwnerOrRepresentative()).Patch("/bookings/{id}/confirm", p.BookingHandler.Confirm)
		r.With(auth, middleware.RequireOwnerOrRepresentative()).Patch("/bookings/{id}/reject", p.BookingHandler.Reject)
		r.With(auth, middleware.RequireOwnerOrRepresentative()).Patch("/bookings/{id}/complete", p.BookingHandler.Complete)

		// Bathhouse bookings (owner/representative/admin)
		r.With(auth, middleware.RequireOwnerOrRepresentative()).Get("/bathhouses/{id}/bookings", p.BookingHandler.ListByBathhouse)

		// Reviews
		r.Get("/bathhouses/{id}/reviews", p.ReviewHandler.ListByBathhouse)
		r.With(auth, middleware.RequireRole(domain.RoleClient)).Post("/bathhouses/{id}/reviews", p.ReviewHandler.Create)
		r.With(auth).Put("/reviews/{id}", p.ReviewHandler.Update)
		r.With(auth).Delete("/reviews/{id}", p.ReviewHandler.Delete)
		r.With(auth, middleware.RequireOwnerOrRepresentative()).Post("/reviews/{id}/response", p.ReviewHandler.AddOwnerResponse)

		// Complaints / Reports (authenticated)
		r.With(auth).Post("/reviews/{id}/report", p.ComplaintHandler.ReportReview)
		r.With(auth).Post("/bathhouses/{id}/report", p.ComplaintHandler.ReportBathhouse)
		r.With(auth).Post("/users/{id}/report", p.ComplaintHandler.ReportUser)

		// Favorites (authenticated)
		r.With(auth).Post("/bathhouses/{id}/favorite", p.FavHandler.Toggle)
		r.With(auth).Get("/my/favorites", p.FavHandler.List)

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

		// Loyalty program (authenticated)
		r.With(auth).Get("/my/loyalty", p.LoyaltyHandler.GetAccount)
		r.With(auth).Get("/my/loyalty/transactions", p.LoyaltyHandler.ListTransactions)
		r.With(auth).Get("/my/loyalty/levels", p.LoyaltyHandler.GetLevels)

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
