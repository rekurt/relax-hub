package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/handler"
	"github.com/nikitaaldaev/bani/internal/middleware"
)

// --- BatchListings Tests ---

func TestAdminHandler_BatchListings_Approve(t *testing.T) {
	adminID := uuid.New()
	authSvc := makeAuthToken(adminID, domain.RoleAdmin)

	approved := make(map[uuid.UUID]bool)
	bhSvc := &mockBathhouseService{
		approveFn: func(_ context.Context, id uuid.UUID) error {
			approved[id] = true
			return nil
		},
	}

	id1, id2 := uuid.New(), uuid.New()
	adminH := handler.NewAdminHandler(nil, bhSvc, nil, nil, &mockAdminNotificationService{})

	router := chi.NewRouter()
	router.With(middleware.RequireAuth(authSvc), middleware.RequireRole(domain.RoleAdmin)).Post("/admin/listings/batch", adminH.BatchListings)

	body, _ := json.Marshal(map[string]interface{}{
		"action": "approve",
		"ids":    []string{id1.String(), id2.String()},
	})
	req := httptest.NewRequest(http.MethodPost, "/admin/listings/batch", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer valid-token")
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp handler.APIResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	data, ok := resp.Data.(map[string]interface{})
	if !ok {
		t.Fatalf("expected data to be map, got %T", resp.Data)
	}

	succeeded, _ := data["succeeded"].([]interface{})
	failed, _ := data["failed"].([]interface{})

	if len(succeeded) != 2 {
		t.Errorf("expected 2 succeeded, got %d", len(succeeded))
	}
	if len(failed) != 0 {
		t.Errorf("expected 0 failed, got %d", len(failed))
	}
	if !approved[id1] || !approved[id2] {
		t.Errorf("expected both ids to be approved")
	}
}

func TestAdminHandler_BatchListings_Reject(t *testing.T) {
	adminID := uuid.New()
	authSvc := makeAuthToken(adminID, domain.RoleAdmin)

	rejected := make(map[uuid.UUID]bool)
	bhSvc := &mockBathhouseService{
		rejectFn: func(_ context.Context, id uuid.UUID) error {
			rejected[id] = true
			return nil
		},
	}

	id1 := uuid.New()
	adminH := handler.NewAdminHandler(nil, bhSvc, nil, nil, &mockAdminNotificationService{})

	router := chi.NewRouter()
	router.With(middleware.RequireAuth(authSvc), middleware.RequireRole(domain.RoleAdmin)).Post("/admin/listings/batch", adminH.BatchListings)

	body, _ := json.Marshal(map[string]interface{}{
		"action": "reject",
		"ids":    []string{id1.String()},
	})
	req := httptest.NewRequest(http.MethodPost, "/admin/listings/batch", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer valid-token")
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var resp handler.APIResponse
	json.NewDecoder(rec.Body).Decode(&resp)
	data := resp.Data.(map[string]interface{})
	succeeded := data["succeeded"].([]interface{})

	if len(succeeded) != 1 {
		t.Errorf("expected 1 succeeded, got %d", len(succeeded))
	}
}

func TestAdminHandler_BatchListings_PartialFailure(t *testing.T) {
	adminID := uuid.New()
	authSvc := makeAuthToken(adminID, domain.RoleAdmin)

	goodID := uuid.New()
	badID := uuid.New()

	bhSvc := &mockBathhouseService{
		approveFn: func(_ context.Context, id uuid.UUID) error {
			if id == badID {
				return domain.ErrNotFound
			}
			return nil
		},
	}

	adminH := handler.NewAdminHandler(nil, bhSvc, nil, nil, &mockAdminNotificationService{})

	router := chi.NewRouter()
	router.With(middleware.RequireAuth(authSvc), middleware.RequireRole(domain.RoleAdmin)).Post("/admin/listings/batch", adminH.BatchListings)

	body, _ := json.Marshal(map[string]interface{}{
		"action": "approve",
		"ids":    []string{goodID.String(), badID.String(), "not-a-uuid"},
	})
	req := httptest.NewRequest(http.MethodPost, "/admin/listings/batch", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer valid-token")
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var resp handler.APIResponse
	json.NewDecoder(rec.Body).Decode(&resp)
	data := resp.Data.(map[string]interface{})
	succeeded := data["succeeded"].([]interface{})
	failed := data["failed"].([]interface{})

	if len(succeeded) != 1 {
		t.Errorf("expected 1 succeeded, got %d", len(succeeded))
	}
	if len(failed) != 2 {
		t.Errorf("expected 2 failed, got %d", len(failed))
	}
}

func TestAdminHandler_BatchListings_InvalidAction(t *testing.T) {
	adminID := uuid.New()
	authSvc := makeAuthToken(adminID, domain.RoleAdmin)

	adminH := handler.NewAdminHandler(nil, nil, nil, nil, &mockAdminNotificationService{})

	router := chi.NewRouter()
	router.With(middleware.RequireAuth(authSvc), middleware.RequireRole(domain.RoleAdmin)).Post("/admin/listings/batch", adminH.BatchListings)

	body, _ := json.Marshal(map[string]interface{}{
		"action": "delete",
		"ids":    []string{uuid.New().String()},
	})
	req := httptest.NewRequest(http.MethodPost, "/admin/listings/batch", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer valid-token")
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestAdminHandler_BatchListings_EmptyIDs(t *testing.T) {
	adminID := uuid.New()
	authSvc := makeAuthToken(adminID, domain.RoleAdmin)

	adminH := handler.NewAdminHandler(nil, nil, nil, nil, &mockAdminNotificationService{})

	router := chi.NewRouter()
	router.With(middleware.RequireAuth(authSvc), middleware.RequireRole(domain.RoleAdmin)).Post("/admin/listings/batch", adminH.BatchListings)

	body, _ := json.Marshal(map[string]interface{}{
		"action": "approve",
		"ids":    []string{},
	})
	req := httptest.NewRequest(http.MethodPost, "/admin/listings/batch", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer valid-token")
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestAdminHandler_BatchListings_ExceedsMax(t *testing.T) {
	adminID := uuid.New()
	authSvc := makeAuthToken(adminID, domain.RoleAdmin)

	adminH := handler.NewAdminHandler(nil, nil, nil, nil, &mockAdminNotificationService{})

	router := chi.NewRouter()
	router.With(middleware.RequireAuth(authSvc), middleware.RequireRole(domain.RoleAdmin)).Post("/admin/listings/batch", adminH.BatchListings)

	ids := make([]string, 1001)
	for i := range ids {
		ids[i] = uuid.New().String()
	}
	body, _ := json.Marshal(map[string]interface{}{
		"action": "approve",
		"ids":    ids,
	})
	req := httptest.NewRequest(http.MethodPost, "/admin/listings/batch", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer valid-token")
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

// --- BatchUsers Tests ---

func TestAdminHandler_BatchUsers_Block(t *testing.T) {
	adminID := uuid.New()
	authSvc := makeAuthToken(adminID, domain.RoleAdmin)

	blocked := make(map[uuid.UUID]bool)
	userSvc := &mockUserService{
		blockFn: func(_ context.Context, id uuid.UUID) error {
			blocked[id] = true
			return nil
		},
	}

	id1, id2 := uuid.New(), uuid.New()
	adminH := handler.NewAdminHandler(userSvc, nil, nil, nil, &mockAdminNotificationService{})

	router := chi.NewRouter()
	router.With(middleware.RequireAuth(authSvc), middleware.RequireRole(domain.RoleAdmin)).Post("/admin/users/batch", adminH.BatchUsers)

	body, _ := json.Marshal(map[string]interface{}{
		"action": "block",
		"ids":    []string{id1.String(), id2.String()},
	})
	req := httptest.NewRequest(http.MethodPost, "/admin/users/batch", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer valid-token")
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp handler.APIResponse
	json.NewDecoder(rec.Body).Decode(&resp)
	data := resp.Data.(map[string]interface{})
	succeeded := data["succeeded"].([]interface{})

	if len(succeeded) != 2 {
		t.Errorf("expected 2 succeeded, got %d", len(succeeded))
	}
	if !blocked[id1] || !blocked[id2] {
		t.Errorf("expected both users to be blocked")
	}
}

func TestAdminHandler_BatchUsers_CannotBlockSelf(t *testing.T) {
	adminID := uuid.New()
	authSvc := makeAuthToken(adminID, domain.RoleAdmin)

	userSvc := &mockUserService{
		blockFn: func(_ context.Context, _ uuid.UUID) error {
			return nil
		},
	}

	otherID := uuid.New()
	adminH := handler.NewAdminHandler(userSvc, nil, nil, nil, &mockAdminNotificationService{})

	router := chi.NewRouter()
	router.With(middleware.RequireAuth(authSvc), middleware.RequireRole(domain.RoleAdmin)).Post("/admin/users/batch", adminH.BatchUsers)

	body, _ := json.Marshal(map[string]interface{}{
		"action": "block",
		"ids":    []string{adminID.String(), otherID.String()},
	})
	req := httptest.NewRequest(http.MethodPost, "/admin/users/batch", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer valid-token")
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var resp handler.APIResponse
	json.NewDecoder(rec.Body).Decode(&resp)
	data := resp.Data.(map[string]interface{})
	succeeded := data["succeeded"].([]interface{})
	failed := data["failed"].([]interface{})

	if len(succeeded) != 1 {
		t.Errorf("expected 1 succeeded, got %d", len(succeeded))
	}
	if len(failed) != 1 {
		t.Errorf("expected 1 failed (self-block), got %d", len(failed))
	}

	failedItem := failed[0].(map[string]interface{})
	errMsg, _ := failedItem["error"].(string)
	if !strings.Contains(errMsg, "cannot block yourself") {
		t.Errorf("expected 'cannot block yourself' error, got '%s'", errMsg)
	}
}

func TestAdminHandler_BatchUsers_Unblock(t *testing.T) {
	adminID := uuid.New()
	authSvc := makeAuthToken(adminID, domain.RoleAdmin)

	unblocked := make(map[uuid.UUID]bool)
	userSvc := &mockUserService{
		unblockFn: func(_ context.Context, id uuid.UUID) error {
			unblocked[id] = true
			return nil
		},
	}

	id1 := uuid.New()
	adminH := handler.NewAdminHandler(userSvc, nil, nil, nil, &mockAdminNotificationService{})

	router := chi.NewRouter()
	router.With(middleware.RequireAuth(authSvc), middleware.RequireRole(domain.RoleAdmin)).Post("/admin/users/batch", adminH.BatchUsers)

	body, _ := json.Marshal(map[string]interface{}{
		"action": "unblock",
		"ids":    []string{id1.String()},
	})
	req := httptest.NewRequest(http.MethodPost, "/admin/users/batch", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer valid-token")
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	if !unblocked[id1] {
		t.Errorf("expected user to be unblocked")
	}
}

func TestAdminHandler_BatchUsers_InvalidAction(t *testing.T) {
	adminID := uuid.New()
	authSvc := makeAuthToken(adminID, domain.RoleAdmin)

	adminH := handler.NewAdminHandler(nil, nil, nil, nil, &mockAdminNotificationService{})

	router := chi.NewRouter()
	router.With(middleware.RequireAuth(authSvc), middleware.RequireRole(domain.RoleAdmin)).Post("/admin/users/batch", adminH.BatchUsers)

	body, _ := json.Marshal(map[string]interface{}{
		"action": "delete",
		"ids":    []string{uuid.New().String()},
	})
	req := httptest.NewRequest(http.MethodPost, "/admin/users/batch", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer valid-token")
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

// --- Chunking Test ---

func TestAdminHandler_BatchListings_Chunking(t *testing.T) {
	adminID := uuid.New()
	authSvc := makeAuthToken(adminID, domain.RoleAdmin)

	approveCount := 0
	bhSvc := &mockBathhouseService{
		approveFn: func(_ context.Context, _ uuid.UUID) error {
			approveCount++
			return nil
		},
	}

	ids := make([]string, 250)
	for i := range ids {
		ids[i] = uuid.New().String()
	}

	adminH := handler.NewAdminHandler(nil, bhSvc, nil, nil, &mockAdminNotificationService{})

	router := chi.NewRouter()
	router.With(middleware.RequireAuth(authSvc), middleware.RequireRole(domain.RoleAdmin)).Post("/admin/listings/batch", adminH.BatchListings)

	body, _ := json.Marshal(map[string]interface{}{
		"action": "approve",
		"ids":    ids,
	})
	req := httptest.NewRequest(http.MethodPost, "/admin/listings/batch", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer valid-token")
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var resp handler.APIResponse
	json.NewDecoder(rec.Body).Decode(&resp)
	data := resp.Data.(map[string]interface{})
	succeeded := data["succeeded"].([]interface{})

	if len(succeeded) != 250 {
		t.Errorf("expected 250 succeeded, got %d", len(succeeded))
	}
	if approveCount != 250 {
		t.Errorf("expected 250 approve calls, got %d", approveCount)
	}
}
