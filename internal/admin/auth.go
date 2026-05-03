package admin

import (
	"net/http"
	"strings"

	gacontext "github.com/GoAdminGroup/go-admin/context"
	"github.com/GoAdminGroup/go-admin/plugins/admin/models"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/rekurt/relax-hub/internal/domain"
	"github.com/rekurt/relax-hub/internal/logger"
	"github.com/rekurt/relax-hub/internal/middleware"
)

// RequireAdminAuth is a middleware that authenticates admin users by checking
// both the Authorization header and the admin_token cookie. This is needed because
// admin custom pages are accessed via browser navigation (which sends cookies),
// unlike the main API which uses Authorization headers exclusively.
func RequireAdminAuth(authService middleware.AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := extractToken(r)
			if token == "" {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			userID, role, err := authService.ParseToken(r.Context(), token)
			if err != nil {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			if role != domain.RoleAdmin {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			ctx := middleware.SetUserID(r.Context(), userID)
			ctx = middleware.SetUserRole(ctx, role)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// extractToken gets the JWT token from either the Authorization header or admin_token cookie.
func extractToken(r *http.Request) string {
	header := r.Header.Get("Authorization")
	if header != "" {
		parts := strings.SplitN(header, " ", 2)
		if len(parts) == 2 && strings.EqualFold(parts[0], "bearer") {
			return parts[1]
		}
	}

	cookie, err := r.Cookie("admin_token")
	if err == nil && cookie.Value != "" {
		return cookie.Value
	}

	return ""
}

// NewAuthProcessor creates a GoAdmin auth.Processor that bridges the existing JWT auth system
// to GoAdmin's session-based auth. It extracts the Bearer token from the request,
// validates it via the app's AuthService, and maps admin users to GoAdmin's UserModel.
// The bridge cookie is marked Secure in production or whenever the request arrived over TLS,
// so plain-HTTP local/staging setups can still receive and resend the cookie.
func NewAuthProcessor(authService middleware.AuthService, pool *pgxpool.Pool, log *logger.Logger, isProduction bool) func(ctx *gacontext.Context) (models.UserModel, bool, string) {
	return func(ctx *gacontext.Context) (models.UserModel, bool, string) {
		empty := models.UserModel{}

		if ctx.Request == nil {
			return empty, false, "no request"
		}

		token := extractToken(ctx.Request)
		if token == "" {
			// Distinguish between no auth at all and wrong format
			if ctx.Request.Header.Get("Authorization") != "" {
				return empty, false, "invalid authorization format"
			}
			return empty, false, "no authorization"
		}

		userID, role, err := authService.ParseToken(ctx.Request.Context(), token)
		if err != nil {
			log.Debug("GoAdmin auth: token parse failed", "error", err)
			return empty, false, "invalid token"
		}

		// Only admin users can access GoAdmin
		if role != domain.RoleAdmin {
			log.Debug("GoAdmin auth: non-admin user denied", "user_id", userID, "role", role)
			return empty, false, "access denied: admin role required"
		}

		// Set admin_token cookie so that custom pages (served via chi routes with
		// RequireAdminAuth middleware) can authenticate browser navigation requests.
		// GoAdmin calls this processor on every request, so the cookie stays fresh.
		if ctx.Response != nil {
			// #nosec G124 -- Secure is intentionally conditional: forced on in production
			// or whenever the request arrived over TLS, but allowed off on plain-HTTP
			// local/staging setups where browsers would otherwise drop the cookie.
			ctx.SetCookie(&http.Cookie{
				Name:     "admin_token",
				Value:    token,
				Path:     "/",
				HttpOnly: true,
				SameSite: http.SameSiteStrictMode,
				Secure:   isProduction || ctx.Request.TLS != nil,
			})
		}

		// Look up GoAdmin user from goadmin_users to properly track admin identity.
		goAdminID := 1
		goAdminName := "Admin"
		if pool != nil {
			var dbID int
			var dbName string
			err := pool.QueryRow(ctx.Request.Context(),
				`SELECT id, name FROM goadmin_users WHERE username = $1 LIMIT 1`,
				userID.String(),
			).Scan(&dbID, &dbName)
			if err == nil {
				goAdminID = dbID
				goAdminName = dbName
			}
		}

		user := models.UserModel{
			Id:       int64(goAdminID),
			UserName: userID.String(),
			Name:     goAdminName,
			Roles:    []models.RoleModel{{Id: 1, Name: "administrator", Slug: "administrator"}},
		}

		log.Debug("GoAdmin auth: admin user authenticated", "user_id", userID)
		return user, true, ""
	}
}
