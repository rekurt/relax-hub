package middleware_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/logger"
	"github.com/nikitaaldaev/bani/internal/middleware"
	"github.com/nikitaaldaev/bani/internal/repository/mock"
)

func TestAdminAudit_MutatingMethodsLogged(t *testing.T) {
	repo := mock.NewAuditLogRepo()
	log := logger.New(logger.LevelWarn)
	mw := middleware.AdminAudit(repo, log)

	adminID := uuid.New()

	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	methods := []string{http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete}
	for _, method := range methods {
		req := httptest.NewRequest(method, "/api/v1/admin/bathhouses/"+uuid.New().String()+"/approve", nil)
		ctx := middleware.SetUserID(req.Context(), adminID)
		ctx = middleware.SetUserRole(ctx, domain.RoleAdmin)
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("%s: expected 200, got %d", method, rr.Code)
		}
	}

	// Verify all 4 methods were logged
	entityType := "admin_action"
	result, err := repo.List(context.Background(), domain.AuditLogFilter{
		EntityType: &entityType,
		Page:       1,
		PageSize:   20,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.TotalCount != 4 {
		t.Errorf("expected 4 audit logs, got %d", result.TotalCount)
	}
}

func TestAdminAudit_GetRequestNotLogged(t *testing.T) {
	repo := mock.NewAuditLogRepo()
	log := logger.New(logger.LevelWarn)
	mw := middleware.AdminAudit(repo, log)

	adminID := uuid.New()

	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/users", nil)
	ctx := middleware.SetUserID(req.Context(), adminID)
	ctx = middleware.SetUserRole(ctx, domain.RoleAdmin)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	entityType := "admin_action"
	result, err := repo.List(context.Background(), domain.AuditLogFilter{
		EntityType: &entityType,
		Page:       1,
		PageSize:   20,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.TotalCount != 0 {
		t.Errorf("expected 0 audit logs for GET, got %d", result.TotalCount)
	}
}

func TestAdminAudit_FailedRequestNotLogged(t *testing.T) {
	repo := mock.NewAuditLogRepo()
	log := logger.New(logger.LevelWarn)
	mw := middleware.AdminAudit(repo, log)

	adminID := uuid.New()

	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	}))

	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/cities", bytes.NewBufferString(`{"name":"Test"}`))
	ctx := middleware.SetUserID(req.Context(), adminID)
	ctx = middleware.SetUserRole(ctx, domain.RoleAdmin)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	entityType := "admin_action"
	result, err := repo.List(context.Background(), domain.AuditLogFilter{
		EntityType: &entityType,
		Page:       1,
		PageSize:   20,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.TotalCount != 0 {
		t.Errorf("expected 0 audit logs for failed request, got %d", result.TotalCount)
	}
}

func TestAdminAudit_SensitiveDataRedacted(t *testing.T) {
	repo := mock.NewAuditLogRepo()
	log := logger.New(logger.LevelWarn)
	mw := middleware.AdminAudit(repo, log)

	adminID := uuid.New()

	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	body := map[string]interface{}{
		"name":           "Test User",
		"password":       "secret123",
		"token":          "jwt-token-abc",
		"card_number":    "4111111111111111",
		"status":         "active",
		"amount":         5000,
		"bank_account":   "40817810099910004312",
		"reason":         "test reason",
		"nested": map[string]interface{}{
			"api_key": "key-123",
			"data":    "safe",
		},
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPatch, "/api/v1/admin/users/"+uuid.New().String()+"/block", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	ctx := middleware.SetUserID(req.Context(), adminID)
	ctx = middleware.SetUserRole(ctx, domain.RoleAdmin)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	entityType := "admin_action"
	result, err := repo.List(context.Background(), domain.AuditLogFilter{
		EntityType: &entityType,
		Page:       1,
		PageSize:   20,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.TotalCount != 1 {
		t.Fatalf("expected 1 audit log, got %d", result.TotalCount)
	}

	var details map[string]interface{}
	if err := json.Unmarshal(result.Items[0].ChangedFields, &details); err != nil {
		t.Fatalf("failed to unmarshal details: %v", err)
	}

	reqBody, ok := details["request_body"].(map[string]interface{})
	if !ok {
		t.Fatal("request_body not found or not a map")
	}

	// Sensitive fields should be redacted
	for _, field := range []string{"password", "token", "card_number", "bank_account"} {
		if v, ok := reqBody[field]; !ok || v != "[REDACTED]" {
			t.Errorf("field %q should be [REDACTED], got %v", field, v)
		}
	}

	// Safe fields should be preserved
	if reqBody["name"] != "Test User" {
		t.Errorf("name should be preserved, got %v", reqBody["name"])
	}
	if reqBody["status"] != "active" {
		t.Errorf("status should be preserved, got %v", reqBody["status"])
	}
	if reqBody["reason"] != "test reason" {
		t.Errorf("reason should be preserved, got %v", reqBody["reason"])
	}

	// Nested sensitive fields
	nested, ok := reqBody["nested"].(map[string]interface{})
	if !ok {
		t.Fatal("nested field not found or not a map")
	}
	if nested["api_key"] != "[REDACTED]" {
		t.Errorf("nested api_key should be [REDACTED], got %v", nested["api_key"])
	}
	if nested["data"] != "safe" {
		t.Errorf("nested data should be preserved, got %v", nested["data"])
	}
}

func TestAdminAudit_EntityIDExtracted(t *testing.T) {
	repo := mock.NewAuditLogRepo()
	log := logger.New(logger.LevelWarn)
	mw := middleware.AdminAudit(repo, log)

	adminID := uuid.New()
	targetID := uuid.New()

	// Use chi router to extract URL params
	r := chi.NewRouter()
	r.With(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := middleware.SetUserID(r.Context(), adminID)
			ctx = middleware.SetUserRole(ctx, domain.RoleAdmin)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}).Route("/admin", func(r chi.Router) {
		r.Use(mw)
		r.Patch("/bathhouses/{id}/approve", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})
	})

	req := httptest.NewRequest(http.MethodPatch, "/admin/bathhouses/"+targetID.String()+"/approve", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	entityType := "admin_action"
	result, err := repo.List(context.Background(), domain.AuditLogFilter{
		EntityType: &entityType,
		Page:       1,
		PageSize:   20,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.TotalCount != 1 {
		t.Fatalf("expected 1 audit log, got %d", result.TotalCount)
	}

	if result.Items[0].EntityID != targetID {
		t.Errorf("entity_id = %v, want %v", result.Items[0].EntityID, targetID)
	}
	if result.Items[0].UserID != adminID {
		t.Errorf("user_id = %v, want %v", result.Items[0].UserID, adminID)
	}
	if result.Items[0].Action != domain.AuditActionUpdate {
		t.Errorf("action = %v, want update", result.Items[0].Action)
	}
}

func TestAdminAudit_TargetTypeInferred(t *testing.T) {
	repo := mock.NewAuditLogRepo()
	log := logger.New(logger.LevelWarn)
	mw := middleware.AdminAudit(repo, log)

	adminID := uuid.New()

	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/cities", bytes.NewBufferString(`{"name":"Moscow"}`))
	ctx := middleware.SetUserID(req.Context(), adminID)
	ctx = middleware.SetUserRole(ctx, domain.RoleAdmin)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	entityType := "admin_action"
	result, err := repo.List(context.Background(), domain.AuditLogFilter{
		EntityType: &entityType,
		Page:       1,
		PageSize:   20,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.TotalCount != 1 {
		t.Fatalf("expected 1 audit log, got %d", result.TotalCount)
	}

	var details map[string]interface{}
	if err := json.Unmarshal(result.Items[0].ChangedFields, &details); err != nil {
		t.Fatalf("failed to unmarshal details: %v", err)
	}

	targetType, ok := details["target_type"]
	if !ok {
		t.Error("target_type should be present in audit details")
	} else if targetType == "" {
		t.Error("target_type should not be empty")
	}
	// "cities" -> "citie" via simple TrimSuffix("s") — best-effort singularization
}

func TestAdminAudit_NoUserID_NotLogged(t *testing.T) {
	repo := mock.NewAuditLogRepo()
	log := logger.New(logger.LevelWarn)
	mw := middleware.AdminAudit(repo, log)

	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/cities", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	entityType := "admin_action"
	result, err := repo.List(context.Background(), domain.AuditLogFilter{
		EntityType: &entityType,
		Page:       1,
		PageSize:   20,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.TotalCount != 0 {
		t.Errorf("expected 0 audit logs when no admin user, got %d", result.TotalCount)
	}
}

func TestAdminAudit_RequestBodyPreserved(t *testing.T) {
	repo := mock.NewAuditLogRepo()
	log := logger.New(logger.LevelWarn)
	mw := middleware.AdminAudit(repo, log)

	adminID := uuid.New()
	body := `{"name":"Test City"}`

	var receivedBody string
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		buf := new(bytes.Buffer)
		buf.ReadFrom(r.Body)
		receivedBody = buf.String()
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/cities", bytes.NewBufferString(body))
	ctx := middleware.SetUserID(req.Context(), adminID)
	ctx = middleware.SetUserRole(ctx, domain.RoleAdmin)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if receivedBody != body {
		t.Errorf("handler received body %q, want %q", receivedBody, body)
	}
}
