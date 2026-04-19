package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/rekurt/relax-hub/internal/domain"
)

// AdminSubRoleResolver loads the admin sub-role for a given user.
type AdminSubRoleResolver interface {
	GetAdminSubRole(ctx context.Context, userID uuid.UUID) (domain.AdminSubRole, error)
}

// LoadAdminSubRole resolves and sets the admin sub-role in context for admin users.
// Must be placed after RequireAuth and RequireRole(admin).
func LoadAdminSubRole(resolver AdminSubRoleResolver) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userRole := GetUserRole(r.Context())
			if userRole != domain.RoleAdmin {
				next.ServeHTTP(w, r)
				return
			}

			userID := GetUserID(r.Context())
			if userID == uuid.Nil {
				next.ServeHTTP(w, r)
				return
			}

			subRole, err := resolver.GetAdminSubRole(r.Context(), userID)
			if err != nil {
				writeAuthError(w, http.StatusForbidden, "failed to resolve admin permissions")
				return
			}

			ctx := context.WithValue(r.Context(), adminSubRoleKey, subRole)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireAdminPermission checks that the current admin user has the specified permission.
// Must be placed after LoadAdminSubRole.
func RequireAdminPermission(perm domain.AdminPermission) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			subRole := GetAdminSubRole(r.Context())
			if subRole == "" {
				writeAuthError(w, http.StatusForbidden, "admin sub-role not assigned")
				return
			}

			if !subRole.HasPermission(perm) {
				writeAuthError(w, http.StatusForbidden, "insufficient admin permissions")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequireAdmin2FA checks that the admin user has 2FA enabled.
// Must be placed after RequireAuth.
func RequireAdmin2FA(resolver Admin2FAChecker) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userRole := GetUserRole(r.Context())
			if userRole != domain.RoleAdmin {
				next.ServeHTTP(w, r)
				return
			}

			userID := GetUserID(r.Context())
			if userID == uuid.Nil {
				next.ServeHTTP(w, r)
				return
			}

			has2FA, err := resolver.HasEnabled2FA(r.Context(), userID)
			if err != nil {
				writeAuthError(w, http.StatusForbidden, "failed to check 2FA status")
				return
			}

			if !has2FA {
				writeAuthError(w, http.StatusForbidden, "two-factor authentication required for admin access")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// Admin2FAChecker checks whether a user has 2FA enabled.
type Admin2FAChecker interface {
	HasEnabled2FA(ctx context.Context, userID uuid.UUID) (bool, error)
}
