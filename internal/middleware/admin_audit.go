package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/logger"
	"github.com/nikitaaldaev/bani/internal/repository"
)

// sensitiveFields are JSON field names whose values must be redacted in audit logs.
var sensitiveFields = map[string]bool{
	"password":        true,
	"password_hash":   true,
	"token":           true,
	"access_token":    true,
	"refresh_token":   true,
	"secret":          true,
	"secret_key":      true,
	"api_key":         true,
	"bank_account":    true,
	"card_number":     true,
	"cvv":             true,
	"inn":             true,
	"bik":             true,
	"account_number":  true,
	"otp":             true,
	"totp_secret":     true,
	"client_secret":   true,
	"payment_details": true,
}

// AdminAudit returns middleware that logs mutating admin actions (POST/PUT/PATCH/DELETE)
// as audit log entries with EntityType="admin_action".
func AdminAudit(repo repository.AuditLogRepository, log *logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Only audit mutating methods
			if !isMutatingMethod(r.Method) {
				next.ServeHTTP(w, r)
				return
			}

			adminID := GetUserID(r.Context())
			if adminID == uuid.Nil {
				next.ServeHTTP(w, r)
				return
			}

			// Read and restore request body (limit to 1MB to prevent memory exhaustion)
			var redactedBody json.RawMessage
			if r.Body != nil {
				r.Body = http.MaxBytesReader(nil, r.Body, 1<<20) // 1MB
				bodyBytes, err := io.ReadAll(r.Body)
				r.Body.Close()
				if err == nil && len(bodyBytes) > 0 {
					r.Body = io.NopCloser(bytes.NewReader(bodyBytes))
					redactedBody = redactRequestBody(bodyBytes)
				} else {
					r.Body = io.NopCloser(bytes.NewReader(bodyBytes))
				}
			}

			// Wrap response writer to capture status code
			ww := chiMiddleware.NewWrapResponseWriter(w, r.ProtoMajor)

			next.ServeHTTP(ww, r)

			// Only log successful mutating operations
			status := ww.Status()
			if status < 200 || status >= 300 {
				return
			}

			// Extract target entity ID from URL params
			entityID := extractEntityID(r)

			// Build audit details
			details := map[string]interface{}{
				"method":      r.Method,
				"path":        r.URL.Path,
				"ip":          r.RemoteAddr,
				"status_code": status,
			}
			if redactedBody != nil {
				details["request_body"] = json.RawMessage(redactedBody)
			}
			if targetType := inferTargetType(r.URL.Path); targetType != "" {
				details["target_type"] = targetType
			}

			detailsJSON, err := json.Marshal(details)
			if err != nil {
				log.Error("admin audit: failed to marshal details", "error", err)
				return
			}

			action := httpMethodToAction(r.Method)

			entry := &domain.AuditLog{
				ID:            uuid.New(),
				EntityType:    "admin_action",
				EntityID:      entityID,
				UserID:        adminID,
				Action:        action,
				ChangedFields: detailsJSON,
			}

			if err := repo.Create(r.Context(), entry); err != nil {
				log.Error("admin audit: failed to create audit log",
					"admin_id", adminID,
					"method", r.Method,
					"path", r.URL.Path,
					"error", err,
				)
			}
		})
	}
}

func isMutatingMethod(method string) bool {
	switch method {
	case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
		return true
	}
	return false
}

// extractEntityID tries to find a UUID in the URL path params (typically "id").
func extractEntityID(r *http.Request) uuid.UUID {
	if idStr := chi.URLParam(r, "id"); idStr != "" {
		if id, err := uuid.Parse(idStr); err == nil {
			return id
		}
	}
	return uuid.Nil
}

// inferTargetType extracts the admin sub-resource from the URL path.
// e.g., "/api/v1/admin/bathhouses/xxx/approve" -> "bathhouse"
func inferTargetType(path string) string {
	// Find /admin/ segment and take the next path component
	idx := strings.Index(path, "/admin/")
	if idx < 0 {
		return ""
	}
	rest := path[idx+len("/admin/"):]
	parts := strings.SplitN(rest, "/", 2)
	if len(parts) == 0 || parts[0] == "" {
		return ""
	}
	resource := parts[0]
	// Singularize common admin resources
	resource = strings.TrimSuffix(resource, "s")
	return resource
}

// httpMethodToAction maps HTTP method to AuditAction.
func httpMethodToAction(method string) domain.AuditAction {
	switch method {
	case http.MethodPost:
		return domain.AuditActionCreate
	case http.MethodPut, http.MethodPatch:
		return domain.AuditActionUpdate
	case http.MethodDelete:
		return domain.AuditActionDelete
	default:
		return domain.AuditActionUpdate
	}
}

// redactRequestBody parses the JSON body and redacts sensitive fields.
// Returns nil if the body is not valid JSON.
func redactRequestBody(body []byte) json.RawMessage {
	// Limit body size to prevent excessive memory use
	const maxBodySize = 10 * 1024 // 10KB
	if len(body) > maxBodySize {
		truncated := map[string]string{"_truncated": "request body too large"}
		data, _ := json.Marshal(truncated)
		return data
	}

	var parsed interface{}
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil
	}

	redacted := redactValue(parsed)
	data, err := json.Marshal(redacted)
	if err != nil {
		return nil
	}
	return data
}

func redactValue(v interface{}) interface{} {
	switch val := v.(type) {
	case map[string]interface{}:
		result := make(map[string]interface{}, len(val))
		for k, v := range val {
			if sensitiveFields[strings.ToLower(k)] {
				result[k] = "[REDACTED]"
			} else {
				result[k] = redactValue(v)
			}
		}
		return result
	case []interface{}:
		result := make([]interface{}, len(val))
		for i, item := range val {
			result[i] = redactValue(item)
		}
		return result
	default:
		return v
	}
}
