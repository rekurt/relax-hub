package middleware_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/middleware"
)

func requestWithAuth(role domain.UserRole) *http.Request {
	userID := uuid.New()
	auth := &mockAuthService{userID: userID, role: role}

	// Create a request and run it through RequireAuth to set context
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer token")

	var enrichedReq *http.Request
	middleware.RequireAuth(auth)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		enrichedReq = r
	})).ServeHTTP(httptest.NewRecorder(), req)

	return enrichedReq
}

func TestRequireRole_AllowedRole(t *testing.T) {
	tests := []struct {
		name    string
		role    domain.UserRole
		allowed []domain.UserRole
	}{
		{"admin accessing admin route", domain.RoleAdmin, []domain.UserRole{domain.RoleAdmin}},
		{"owner accessing owner route", domain.RoleOwner, []domain.UserRole{domain.RoleOwner}},
		{"client accessing client route", domain.RoleClient, []domain.UserRole{domain.RoleClient}},
		{"admin accessing multi-role route", domain.RoleAdmin, []domain.UserRole{domain.RoleAdmin, domain.RoleOwner}},
		{"owner accessing multi-role route", domain.RoleOwner, []domain.UserRole{domain.RoleAdmin, domain.RoleOwner}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			called := false
			handler := middleware.RequireRole(tt.allowed...)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				called = true
				w.WriteHeader(http.StatusOK)
			}))

			req := requestWithAuth(tt.role)
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			if !called {
				t.Error("handler should have been called")
			}
			if rec.Code != http.StatusOK {
				t.Errorf("expected status 200, got %d", rec.Code)
			}
		})
	}
}

func TestRequireRole_ForbiddenRole(t *testing.T) {
	tests := []struct {
		name    string
		role    domain.UserRole
		allowed []domain.UserRole
	}{
		{"client accessing admin route", domain.RoleClient, []domain.UserRole{domain.RoleAdmin}},
		{"client accessing owner route", domain.RoleClient, []domain.UserRole{domain.RoleOwner}},
		{"owner accessing admin route", domain.RoleOwner, []domain.UserRole{domain.RoleAdmin}},
		{"representative accessing admin route", domain.RoleRepresentative, []domain.UserRole{domain.RoleAdmin}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := middleware.RequireRole(tt.allowed...)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				t.Error("handler should not be called")
			}))

			req := requestWithAuth(tt.role)
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			if rec.Code != http.StatusForbidden {
				t.Errorf("expected status 403, got %d", rec.Code)
			}
		})
	}
}

func TestRequireRole_Unauthenticated(t *testing.T) {
	handler := middleware.RequireRole(domain.RoleAdmin)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called")
	}))

	// Request without auth context
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", rec.Code)
	}
}

func TestRequireOwnerOrRepresentative(t *testing.T) {
	tests := []struct {
		name       string
		role       domain.UserRole
		expectCode int
	}{
		{"owner allowed", domain.RoleOwner, http.StatusOK},
		{"representative allowed", domain.RoleRepresentative, http.StatusOK},
		{"admin allowed", domain.RoleAdmin, http.StatusOK},
		{"client forbidden", domain.RoleClient, http.StatusForbidden},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := middleware.RequireOwnerOrRepresentative()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))

			req := requestWithAuth(tt.role)
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			if rec.Code != tt.expectCode {
				t.Errorf("expected status %d, got %d", tt.expectCode, rec.Code)
			}
		})
	}
}

func TestRequireAuth_BearerCaseInsensitive(t *testing.T) {
	userID := uuid.New()
	auth := &mockAuthService{userID: userID, role: domain.RoleClient}

	handler := middleware.RequireAuth(auth)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Test with lowercase "bearer"
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "bearer valid-token")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d (bearer lowercase should work)", rec.Code)
	}
}

func TestRequireAuth_AllRolesSetInContext(t *testing.T) {
	roles := []domain.UserRole{domain.RoleClient, domain.RoleOwner, domain.RoleRepresentative, domain.RoleAdmin}

	for _, role := range roles {
		t.Run(string(role), func(t *testing.T) {
			userID := uuid.New()
			auth := &mockAuthService{userID: userID, role: role}

			var capturedCtx context.Context
			handler := middleware.RequireAuth(auth)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				capturedCtx = r.Context()
				w.WriteHeader(http.StatusOK)
			}))

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.Header.Set("Authorization", "Bearer token")
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			if middleware.GetUserID(capturedCtx) != userID {
				t.Errorf("expected user ID %s, got %s", userID, middleware.GetUserID(capturedCtx))
			}
			if middleware.GetUserRole(capturedCtx) != role {
				t.Errorf("expected role %s, got %s", role, middleware.GetUserRole(capturedCtx))
			}
		})
	}
}
