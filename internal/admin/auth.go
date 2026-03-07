package admin

import (
	"strings"

	gacontext "github.com/GoAdminGroup/go-admin/context"
	"github.com/GoAdminGroup/go-admin/plugins/admin/models"

	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/logger"
	"github.com/nikitaaldaev/bani/internal/middleware"
)

// NewAuthProcessor creates a GoAdmin auth.Processor that bridges the existing JWT auth system
// to GoAdmin's session-based auth. It extracts the Bearer token from the request,
// validates it via the app's AuthService, and maps admin users to GoAdmin's UserModel.
func NewAuthProcessor(authService middleware.AuthService, log *logger.Logger) func(ctx *gacontext.Context) (models.UserModel, bool, string) {
	return func(ctx *gacontext.Context) (models.UserModel, bool, string) {
		empty := models.UserModel{}

		if ctx.Request == nil {
			return empty, false, "no request"
		}

		// Extract Bearer token from Authorization header
		header := ctx.Request.Header.Get("Authorization")
		if header == "" {
			// Also check for token in cookie (for browser-based admin access)
			cookie, err := ctx.Request.Cookie("admin_token")
			if err != nil || cookie.Value == "" {
				return empty, false, "no authorization"
			}
			header = "Bearer " + cookie.Value
		}

		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
			return empty, false, "invalid authorization format"
		}

		token := parts[1]
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

		user := models.UserModel{
			Id:       1,
			UserName: userID.String(),
			Name:     "Admin",
			Roles:    []models.RoleModel{{Id: 1, Name: "administrator", Slug: "administrator"}},
		}

		log.Debug("GoAdmin auth: admin user authenticated", "user_id", userID)
		return user, true, ""
	}
}
