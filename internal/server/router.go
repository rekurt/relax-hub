package server

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/handler"
	"github.com/nikitaaldaev/bani/internal/middleware"
	"go.uber.org/fx"
)

type RouterParams struct {
	fx.In

	CORS           *middleware.CORSMiddleware
	AuthService    middleware.AuthService
	AuthHandler    *handler.AuthHandler
	BHHandler      *handler.BathhouseHandler
	BookingHandler *handler.BookingHandler
	ReviewHandler  *handler.ReviewHandler
	RepHandler     *handler.RepresentativeHandler
	CityHandler    *handler.CityHandler
	AdminHandler   *handler.AdminHandler
}

func NewRouter(p RouterParams) http.Handler {
	r := chi.NewRouter()

	r.Use(chiMiddleware.RequestID)
	r.Use(middleware.Logging)
	r.Use(chiMiddleware.Recoverer)
	r.Use(p.CORS.Handler)

	r.Get("/health", healthCheck)

	auth := middleware.RequireAuth(p.AuthService)

	r.Route("/api/v1", func(r chi.Router) {
		// Auth (public)
		r.Post("/auth/register", p.AuthHandler.Register)
		r.Post("/auth/login", p.AuthHandler.Login)
		r.With(auth).Get("/auth/me", p.AuthHandler.Me)

		// Bathhouses (public)
		r.Get("/bathhouses", p.BHHandler.Search)
		r.Get("/bathhouses/{id}", p.BHHandler.GetByID)
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

		// Bathhouse bookings (owner/representative/admin)
		r.With(auth, middleware.RequireOwnerOrRepresentative()).Get("/bathhouses/{id}/bookings", p.BookingHandler.ListByBathhouse)

		// Reviews
		r.Get("/bathhouses/{id}/reviews", p.ReviewHandler.ListByBathhouse)
		r.With(auth, middleware.RequireRole(domain.RoleClient)).Post("/bathhouses/{id}/reviews", p.ReviewHandler.Create)

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
		})
	})

	return r
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}
