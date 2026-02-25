package middleware

import (
	"net/http"

	"github.com/nikitaaldaev/bani/internal/domain"
)

func RequireRole(roles ...domain.UserRole) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userRole := GetUserRole(r.Context())
			if userRole == "" {
				writeAuthError(w, http.StatusUnauthorized, "authentication required")
				return
			}

			for _, allowed := range roles {
				if userRole == allowed {
					next.ServeHTTP(w, r)
					return
				}
			}

			writeAuthError(w, http.StatusForbidden, "insufficient permissions")
		})
	}
}

func RequireOwnerOrRepresentative() func(http.Handler) http.Handler {
	return RequireRole(domain.RoleOwner, domain.RoleRepresentative, domain.RoleAdmin)
}
