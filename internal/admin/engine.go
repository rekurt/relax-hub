package admin

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/nikitaaldaev/bani/internal/admin/pages"
	"github.com/nikitaaldaev/bani/internal/logger"
)

// PagesRouter creates a chi router with all custom admin pages.
// Dashboard is mounted at "/" as the default landing page.
func PagesRouter(dashProvider pages.DashboardDataProvider, modProvider pages.ModerationDataProvider, analyticsProvider pages.AnalyticsDataProvider, log *logger.Logger) http.Handler {
	r := chi.NewRouter()

	dashboard := pages.NewDashboardHandler(dashProvider, log)
	r.Get("/", dashboard.ServeHTTP)
	r.Get("/dashboard", dashboard.ServeHTTP)

	moderation := pages.NewModerationHandler(modProvider, log)
	r.Get("/moderation", moderation.ServeHTTP)
	r.Post("/moderation/api/approve", moderation.HandleApprove)
	r.Post("/moderation/api/reject", moderation.HandleReject)
	r.Post("/moderation/api/batch-approve", moderation.HandleBatchApprove)
	r.Post("/moderation/api/batch-reject", moderation.HandleBatchReject)

	analytics := pages.NewAnalyticsHandler(analyticsProvider, log)
	r.Get("/analytics", analytics.ServeHTTP)

	return r
}
