package server

import (
	"github.com/go-chi/chi/v5"
	"github.com/rekurt/relax-hub/internal/middleware"
	httpSwagger "github.com/swaggo/http-swagger/v2"
)

// mountPublicRoutes registers top-level public routes: health, sitemap, prerender, swagger, calendar token
func mountPublicRoutes(r chi.Router, p RouterParams) {
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
}
