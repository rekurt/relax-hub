package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/google/uuid"
	"github.com/rekurt/relax-hub/internal/domain"
)

type contextKey string

const (
	userIDKey       contextKey = "user_id"
	roleKey         contextKey = "user_role"
	sessionIDKey    contextKey = "session_id"
	adminSubRoleKey contextKey = "admin_sub_role"
)

func GetUserID(ctx context.Context) uuid.UUID {
	if id, ok := ctx.Value(userIDKey).(uuid.UUID); ok {
		return id
	}
	return uuid.Nil
}

func GetUserRole(ctx context.Context) domain.UserRole {
	if role, ok := ctx.Value(roleKey).(domain.UserRole); ok {
		return role
	}
	return ""
}

func GetSessionID(ctx context.Context) uuid.UUID {
	if id, ok := ctx.Value(sessionIDKey).(uuid.UUID); ok {
		return id
	}
	return uuid.Nil
}

// SetSessionID sets the session ID in context.
func SetSessionID(ctx context.Context, sessionID uuid.UUID) context.Context {
	return context.WithValue(ctx, sessionIDKey, sessionID)
}

// SetUserID sets the user ID in context.
func SetUserID(ctx context.Context, userID uuid.UUID) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}

// SetUserRole sets the user role in context.
func SetUserRole(ctx context.Context, role domain.UserRole) context.Context {
	return context.WithValue(ctx, roleKey, role)
}

// SetUserIDForTesting is an alias for SetUserID, kept for test readability.
func SetUserIDForTesting(ctx context.Context, userID uuid.UUID) context.Context {
	return SetUserID(ctx, userID)
}

// SetUserRoleForTesting is an alias for SetUserRole, kept for test readability.
func SetUserRoleForTesting(ctx context.Context, role domain.UserRole) context.Context {
	return SetUserRole(ctx, role)
}

// GetAdminSubRole returns the admin sub-role from context.
func GetAdminSubRole(ctx context.Context) domain.AdminSubRole {
	if r, ok := ctx.Value(adminSubRoleKey).(domain.AdminSubRole); ok {
		return r
	}
	return ""
}

// SetAdminSubRole sets the admin sub-role in context.
func SetAdminSubRole(ctx context.Context, subRole domain.AdminSubRole) context.Context {
	return context.WithValue(ctx, adminSubRoleKey, subRole)
}

type AuthService interface {
	ParseToken(ctx context.Context, token string) (uuid.UUID, domain.UserRole, error)
	ParseTokenWithSession(ctx context.Context, token string) (userID uuid.UUID, role domain.UserRole, sessionID uuid.UUID, err error)
}

// SessionValidator validates that a session is still active.
type SessionValidator interface {
	ValidateAndTouch(ctx context.Context, sessionID uuid.UUID) error
}

func RequireAuth(authService AuthService) func(http.Handler) http.Handler {
	return RequireAuthWithSession(authService, nil)
}

func RequireAuthWithSession(authService AuthService, sessionValidator SessionValidator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			if header == "" {
				writeAuthError(w, http.StatusUnauthorized, "missing authorization header")
				return
			}

			parts := strings.SplitN(header, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
				writeAuthError(w, http.StatusUnauthorized, "invalid authorization header format")
				return
			}

			token := parts[1]
			userID, role, sessionID, err := authService.ParseTokenWithSession(r.Context(), token)
			if err != nil {
				writeAuthError(w, http.StatusUnauthorized, "invalid or expired token")
				return
			}

			// Validate session if present and validator is configured
			if sessionValidator != nil && sessionID != uuid.Nil {
				if err := sessionValidator.ValidateAndTouch(r.Context(), sessionID); err != nil {
					writeAuthError(w, http.StatusUnauthorized, "session expired or invalid")
					return
				}
			}

			ctx := context.WithValue(r.Context(), userIDKey, userID)
			ctx = context.WithValue(ctx, roleKey, role)
			if sessionID != uuid.Nil {
				ctx = context.WithValue(ctx, sessionIDKey, sessionID)
			}
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func OptionalAuth(authService AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			if header == "" {
				next.ServeHTTP(w, r)
				return
			}

			parts := strings.SplitN(header, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
				next.ServeHTTP(w, r)
				return
			}

			token := parts[1]
			userID, role, sessionID, err := authService.ParseTokenWithSession(r.Context(), token)
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}

			ctx := context.WithValue(r.Context(), userIDKey, userID)
			ctx = context.WithValue(ctx, roleKey, role)
			if sessionID != uuid.Nil {
				ctx = context.WithValue(ctx, sessionIDKey, sessionID)
			}
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func writeAuthError(w http.ResponseWriter, status int, message string) {
	// Map HTTP status codes to standard error codes
	var errorCode string
	switch status {
	case http.StatusUnauthorized:
		errorCode = "unauthorized"
	case http.StatusForbidden:
		errorCode = "forbidden"
	default:
		errorCode = "error"
	}
	writeAuthErrorCode(w, status, errorCode, message)
}

// writeAuthErrorCode writes the same envelope as writeAuthError but with an
// explicit, caller-supplied error code. Use when the generic status-derived
// code would be too coarse for the frontend to act on (e.g. distinguishing
// "admin needs 2FA" from a generic 403).
func writeAuthErrorCode(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(map[string]interface{}{
		"success": false,
		"error": map[string]string{
			"code":    code,
			"message": message,
		},
	}); err != nil {
		fmt.Fprintf(os.Stderr, "failed to encode auth error response: %v\n", err)
	}
}
