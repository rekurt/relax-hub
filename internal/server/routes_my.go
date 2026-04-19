package server

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/rekurt/relax-hub/internal/domain"
	"github.com/rekurt/relax-hub/internal/middleware"
)

// mountMyRoutes registers all /api/v1/my/* authenticated user routes (owner, client, representative)
func mountMyRoutes(
	r chi.Router,
	p RouterParams,
	auth func(http.Handler) http.Handler,
) {
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

	// Promo codes (owner/representative)
	r.With(auth, middleware.RequireOwnerOrRepresentative()).Post("/my/bathhouses/{id}/promo-codes", p.PromoHandler.CreateForBathhouse)
	r.With(auth, middleware.RequireOwnerOrRepresentative()).Get("/my/bathhouses/{id}/promo-codes", p.PromoHandler.ListByBathhouse)

	// Favorites
	r.With(auth).Get("/my/favorites", p.FavHandler.List)

	// Recently viewed & saved searches
	r.With(auth).Get("/my/recently-viewed", p.SavedSearchHandler.ListRecentlyViewed)
	r.With(auth).Post("/my/recently-viewed", p.SavedSearchHandler.RecordRecentlyViewed)
	r.With(auth).Post("/my/saved-searches", p.SavedSearchHandler.CreateSavedSearch)
	r.With(auth).Get("/my/saved-searches", p.SavedSearchHandler.ListSavedSearches)
	r.With(auth).Delete("/my/saved-searches/{id}", p.SavedSearchHandler.DeleteSavedSearch)

	// Saved cards
	r.With(auth).Post("/my/saved-cards", p.SavedCardHandler.CreateSavedCard)
	r.With(auth).Get("/my/saved-cards", p.SavedCardHandler.ListSavedCards)
	r.With(auth).Delete("/my/saved-cards/{id}", p.SavedCardHandler.DeleteSavedCard)
	r.With(auth).Post("/my/saved-cards/{id}/default", p.SavedCardHandler.SetDefaultCard)

	// Recommendations & preferences
	r.With(auth).Get("/recommendations", p.RecommendationHandler.GetPersonalized)
	r.With(auth).Get("/my/preferences", p.RecommendationHandler.GetPreferences)
	r.With(auth).Put("/my/preferences", p.RecommendationHandler.UpdatePreferences)

	// Subscriptions (owner)
	r.With(auth, middleware.RequireOwnerOrRepresentative()).Post("/my/bathhouses/{id}/subscription", p.SubscriptionHandler.Subscribe)
	r.With(auth, middleware.RequireOwnerOrRepresentative()).Get("/my/bathhouses/{id}/subscription", p.SubscriptionHandler.GetSubscription)
	r.With(auth, middleware.RequireOwnerOrRepresentative()).Delete("/my/bathhouses/{id}/subscription", p.SubscriptionHandler.CancelSubscription)
	r.With(auth, middleware.RequireOwnerOrRepresentative()).Get("/my/subscriptions", p.SubscriptionHandler.ListSubscriptions)

	// Promotions (owner)
	r.With(auth, middleware.RequireOwnerOrRepresentative()).Post("/my/bathhouses/{id}/promotion", p.SubscriptionHandler.CreatePromotion)
	r.With(auth, middleware.RequireOwnerOrRepresentative()).Get("/my/bathhouses/{id}/promotion", p.SubscriptionHandler.GetPromotion)
	r.With(auth, middleware.RequireOwnerOrRepresentative()).Get("/my/bathhouses/{id}/promotions", p.SubscriptionHandler.ListPromotions)
	r.With(auth, middleware.RequireOwnerOrRepresentative()).Put("/my/promotions/{id}", p.SubscriptionHandler.UpdatePromotion)
	r.With(auth, middleware.RequireOwnerOrRepresentative()).Post("/my/promotions/{id}/pause", p.SubscriptionHandler.PausePromotion)
	r.With(auth, middleware.RequireOwnerOrRepresentative()).Post("/my/promotions/{id}/resume", p.SubscriptionHandler.ResumePromotion)

	// Pricing rules (owner)
	r.With(auth, middleware.RequireOwnerOrRepresentative()).Post("/my/bathhouses/{id}/pricing-rules", p.PricingHandler.CreateRule)
	r.With(auth, middleware.RequireOwnerOrRepresentative()).Get("/my/bathhouses/{id}/pricing-rules", p.PricingHandler.ListRules)
	r.With(auth, middleware.RequireOwnerOrRepresentative()).Put("/pricing-rules/{id}", p.PricingHandler.UpdateRule)
	r.With(auth, middleware.RequireOwnerOrRepresentative()).Delete("/pricing-rules/{id}", p.PricingHandler.DeleteRule)

	// Seasonal tariffs (owner)
	r.With(auth, middleware.RequireOwnerOrRepresentative()).Post("/my/bathhouses/{id}/seasonal-tariffs", p.PricingHandler.CreateSeasonalTariff)
	r.With(auth, middleware.RequireOwnerOrRepresentative()).Get("/my/bathhouses/{id}/seasonal-tariffs", p.PricingHandler.ListSeasonalTariffs)
	r.With(auth, middleware.RequireOwnerOrRepresentative()).Put("/seasonal-tariffs/{id}", p.PricingHandler.UpdateSeasonalTariff)
	r.With(auth, middleware.RequireOwnerOrRepresentative()).Delete("/seasonal-tariffs/{id}", p.PricingHandler.DeleteSeasonalTariff)

	// Smart pricing recommendation
	r.With(auth, middleware.RequireOwnerOrRepresentative()).Get("/my/bathhouses/{id}/price-recommendation", p.PricingHandler.GetPriceRecommendation)

	// Holiday multiplier
	r.With(auth, middleware.RequireOwnerOrRepresentative()).Put("/my/bathhouses/{id}/holiday-multiplier", p.HolidayHandler.SetBathhouseMultiplier)

	// Widget API keys
	r.With(auth, middleware.RequireOwnerOrRepresentative()).Get("/my/bathhouses/{id}/widget-key", p.BHHandler.GetWidgetKey)
	r.With(auth, middleware.RequireOwnerOrRepresentative()).Post("/my/bathhouses/{id}/widget-key/regenerate", p.BHHandler.RegenerateWidgetKey)
	r.With(auth, middleware.RequireOwnerOrRepresentative()).Get("/my/bathhouses/{id}/widget-code", p.BHHandler.GetWidgetCode)

	// User profile and statistics
	r.With(auth).Get("/my/stats", p.AuthHandler.GetMyStats)
	r.With(auth).Get("/my/profile-completeness", p.AuthHandler.GetProfileCompleteness)
	r.With(auth).Post("/my/onboarding/complete", p.AuthHandler.CompleteOnboarding)

	// Sessions
	r.With(auth).Get("/my/sessions", p.SessionHandler.ListSessions)
	r.With(auth).Delete("/my/sessions", p.SessionHandler.TerminateAllOtherSessions)
	r.With(auth).Delete("/my/sessions/{id}", p.SessionHandler.TerminateSession)

	// Region
	r.With(auth).Get("/my/region", p.RegionHandler.GetRegion)
	r.With(auth).Put("/my/region", p.RegionHandler.SwitchRegion)

	// Wallet
	r.With(auth).Get("/my/wallet", p.WalletHandler.GetWallet)
	r.With(auth).Get("/my/wallet/transactions", p.WalletHandler.ListTransactions)
	r.With(auth).Post("/my/wallet/topup", p.WalletHandler.TopUp)
	r.With(auth).Get("/my/wallet/holds", p.WalletHandler.ListHolds)
	r.With(auth).Get("/my/wallet/export", p.FinancialReportHandler.ExportWalletTransactions)

	// Financial reports (owner)
	r.With(auth, middleware.RequireRole(domain.RoleOwner)).Get("/my/finance/acts/{bathhouse_id}", p.FinancialReportHandler.GenerateAct)
	r.With(auth, middleware.RequireRole(domain.RoleOwner)).Get("/my/finance/export-xml", p.FinancialReportHandler.ExportXML1C)

	// KYC (owner)
	r.With(auth, middleware.RequireRole(domain.RoleOwner)).Post("/my/kyc", p.KYCHandler.SubmitKYC)
	r.With(auth, middleware.RequireRole(domain.RoleOwner)).Get("/my/kyc", p.KYCHandler.GetKYCStatus)

	// Offer acceptance (owner)
	r.With(auth, middleware.RequireRole(domain.RoleOwner)).Post("/my/offer/accept", p.OfferHandler.AcceptOffer)
	r.With(auth, middleware.RequireRole(domain.RoleOwner)).Get("/my/offer/status", p.OfferHandler.GetOfferStatus)

	// Payment details (owner)
	r.With(auth, middleware.RequireRole(domain.RoleOwner)).Put("/my/payment-details", p.PaymentDetailsHandler.SetPaymentDetails)
	r.With(auth, middleware.RequireRole(domain.RoleOwner)).Get("/my/payment-details", p.PaymentDetailsHandler.GetPaymentDetails)

	// Listing drafts (owner)
	r.With(auth, middleware.RequireRole(domain.RoleOwner)).Post("/my/listing-drafts", p.ListingDraftHandler.CreateDraft)
	r.With(auth, middleware.RequireRole(domain.RoleOwner)).Get("/my/listing-drafts", p.ListingDraftHandler.ListDrafts)
	r.With(auth, middleware.RequireRole(domain.RoleOwner)).Get("/my/listing-drafts/{id}", p.ListingDraftHandler.GetDraft)
	r.With(auth, middleware.RequireRole(domain.RoleOwner)).Put("/my/listing-drafts/{id}/step/{step}", p.ListingDraftHandler.SaveStep)
	r.With(auth, middleware.RequireRole(domain.RoleOwner)).Post("/my/listing-drafts/{id}/submit", p.ListingDraftHandler.SubmitDraft)
	r.With(auth, middleware.RequireRole(domain.RoleOwner)).Delete("/my/listing-drafts/{id}", p.ListingDraftHandler.DeleteDraft)

	// Listing import CSV/XLSX (owner)
	r.With(auth, middleware.RequireRole(domain.RoleOwner)).Post("/my/listings/import", p.ListingImportHandler.Import)
	r.With(auth, middleware.RequireRole(domain.RoleOwner)).Get("/my/listings/import/template", p.ListingImportHandler.GetImportTemplate)

	// Payouts (owner)
	r.With(auth, middleware.RequireRole(domain.RoleOwner)).Post("/my/wallet/payout", p.PayoutHandler.RequestPayout)
	r.With(auth, middleware.RequireRole(domain.RoleOwner)).Put("/my/wallet/auto-payout", p.PayoutHandler.SetAutoPayoutThreshold)
	r.With(auth, middleware.RequireRole(domain.RoleOwner)).Get("/my/wallet/payouts", p.PayoutHandler.ListPayouts)
	r.With(auth, middleware.RequireRole(domain.RoleOwner)).Get("/my/wallet/payouts/export", p.FinancialReportHandler.ExportPayouts)

	// Loyalty program
	r.With(auth).Get("/my/loyalty", p.LoyaltyHandler.GetAccount)
	r.With(auth).Get("/my/loyalty/transactions", p.LoyaltyHandler.ListTransactions)
	r.With(auth).Get("/my/loyalty/levels", p.LoyaltyHandler.GetLevels)

	// Payments
	r.With(auth).Get("/my/payments", p.PaymentHandler.ListUserPayments)

	// Referral program
	r.With(auth).Get("/my/referral", p.ReferralHandler.GetCode)
	r.With(auth).Get("/my/referral/stats", p.ReferralHandler.GetStats)
	r.With(auth).Get("/my/referral/balance", p.ReferralHandler.GetBalance)

	// Gift certificates (authenticated)
	r.With(auth).Post("/certificates/redeem", p.CertificateHandler.Redeem)
	r.With(auth).Get("/my/certificates", p.CertificateHandler.ListMyCertificates)

	// Chat
	r.With(auth).Get("/my/conversations", p.ChatHandler.ListConversations)
	r.With(auth).Get("/conversations/{id}/messages", p.ChatHandler.ListMessages)
	r.With(auth).Post("/conversations/{id}/messages", p.ChatHandler.SendMessage)
	r.With(auth).Patch("/conversations/{id}/read", p.ChatHandler.MarkAsRead)
	r.With(auth).Get("/my/unread-messages-count", p.ChatHandler.GetUnreadCount)

	// Analytics (owner/representative)
	r.With(auth, middleware.RequireOwnerOrRepresentative()).Get("/my/bathhouses/{id}/analytics", p.AnalyticsHandler.GetOwnerDashboard)
	r.With(auth, middleware.RequireOwnerOrRepresentative()).Get("/my/bathhouses/{id}/analytics/daily", p.AnalyticsHandler.GetOwnerDailyStats)
	r.With(auth, middleware.RequireOwnerOrRepresentative()).Get("/my/bathhouses/{id}/analytics/performance", p.AnalyticsHandler.GetOwnerPerformance)

	// Client reviews (my)
	r.With(auth).Get("/my/client-reviews", p.ClientReviewHandler.ListMyClientReviews)

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

	// FAQ Bot
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

	// Support Tickets
	r.With(auth).Post("/my/tickets", p.TicketHandler.CreateTicket)
	r.With(auth).Get("/my/tickets", p.TicketHandler.ListUserTickets)
	r.With(auth).Get("/my/tickets/{id}", p.TicketHandler.GetTicket)
	r.With(auth).Post("/my/tickets/{id}/messages", p.TicketHandler.AddUserMessage)
	r.With(auth).Get("/my/tickets/{id}/messages", p.TicketHandler.ListMessages)
	r.With(auth).Post("/my/tickets/{id}/csat", p.TicketHandler.SubmitCSAT)

	// Disputes (my)
	r.With(auth).Get("/my/disputes", p.DisputeHandler.ListUserDisputes)
	r.With(auth).Get("/my/disputes/{id}", p.DisputeHandler.GetDispute)
	r.With(auth).Post("/my/disputes/{id}/evidence", p.DisputeHandler.SubmitEvidence)
	r.With(auth).Get("/my/disputes/{id}/evidence", p.DisputeHandler.ListEvidence)
	r.With(auth).Post("/my/disputes/{id}/appeal", p.DisputeHandler.AppealDispute)

	// Notifications
	r.With(auth).Get("/my/notifications", p.NotifHandler.List)
	r.With(auth).Get("/my/notifications/unread-count", p.NotifHandler.UnreadCount)
	r.With(auth).Patch("/my/notifications/{id}/read", p.NotifHandler.MarkAsRead)
	r.With(auth).Patch("/my/notifications/read-all", p.NotifHandler.MarkAllAsRead)
	r.With(auth).Get("/my/notification-preferences", p.NotifHandler.GetPreferences)
	r.With(auth).Put("/my/notification-preferences", p.NotifHandler.UpdatePreferences)
	r.With(auth).Get("/my/notification-preferences/events", p.NotifHandler.GetEventPreferences)
	r.With(auth).Put("/my/notification-preferences/events", p.NotifHandler.UpdateEventPreferences)

	// Device tokens for push notifications
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
}
