package server

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/rekurt/relax-hub/internal/domain"
	"github.com/rekurt/relax-hub/internal/middleware"
)

// mountAdminRoutes registers the /api/v1/admin sub-router with all admin endpoints
func mountAdminRoutes(
	r chi.Router,
	p RouterParams,
	auth func(http.Handler) http.Handler,
) {
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
		r.With(middleware.RequireAdminPermission(domain.PermAnalyticsView)).Get("/analytics/business-metrics", p.AnalyticsHandler.GetBusinessMetrics)
		r.With(middleware.RequireAdminPermission(domain.PermAnalyticsView)).Get("/analytics/pnl", p.AnalyticsHandler.GetPnL)
		r.With(middleware.RequireAdminPermission(domain.PermAnalyticsView)).Get("/analytics/heatmap", p.AnalyticsHandler.GetHeatmap)

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
		r.With(middleware.RequireAdminPermission(domain.PermTicketManage)).Get("/tickets/metrics", p.TicketHandler.AdminGetOperationMetrics)
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
}
