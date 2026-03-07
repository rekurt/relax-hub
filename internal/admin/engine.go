package admin

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/nikitaaldaev/bani/internal/admin/pages"
	"github.com/nikitaaldaev/bani/internal/logger"
)

// PagesRouter creates a chi router with all custom admin pages.
// Dashboard is mounted at "/" as the default landing page.
func PagesRouter(provider pages.DashboardDataProvider, log *logger.Logger) http.Handler {
	r := chi.NewRouter()

	dashboard := pages.NewDashboardHandler(provider, log)
	r.Get("/", dashboard.ServeHTTP)
	r.Get("/dashboard", dashboard.ServeHTTP)

	return r
}
