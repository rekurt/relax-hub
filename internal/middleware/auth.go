package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
)

type contextKey string

const (
	userIDKey contextKey = "user_id"
	roleKey   contextKey = "user_role"
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

type AuthService interface {
	ParseToken(ctx context.Context, token string) (uuid.UUID, domain.UserRole, error)
}

func RequireAuth(authService AuthService) func(http.Handler) http.Handler {
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
			userID, role, err := authService.ParseToken(r.Context(), token)
			if err != nil {
				writeAuthError(w, http.StatusUnauthorized, "invalid or expired token")
				return
			}

			ctx := context.WithValue(r.Context(), userIDKey, userID)
			ctx = context.WithValue(ctx, roleKey, role)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func writeAuthError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": false,
		"error": map[string]string{
			"code":    http.StatusText(status),
			"message": message,
		},
	})
}
