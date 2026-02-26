package middleware_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/middleware"
)

type mockAuthService struct {
	userID uuid.UUID
	role   domain.UserRole
	err    error
}

func (m *mockAuthService) ParseToken(_ context.Context, _ string) (uuid.UUID, domain.UserRole, error) {
	return m.userID, m.role, m.err
}

func TestRequireAuth_ValidToken(t *testing.T) {
	userID := uuid.New()
	auth := &mockAuthService{userID: userID, role: domain.RoleClient}

	handler := middleware.RequireAuth(auth)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotID := middleware.GetUserID(r.Context())
		gotRole := middleware.GetUserRole(r.Context())
		if gotID != userID {
			t.Errorf("expected user ID %s, got %s", userID, gotID)
		}
		if gotRole != domain.RoleClient {
			t.Errorf("expected role %s, got %s", domain.RoleClient, gotRole)
		}
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
}

func TestRequireAuth_MissingHeader(t *testing.T) {
	auth := &mockAuthService{}

	handler := middleware.RequireAuth(auth)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called")
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", rec.Code)
	}
}

func TestRequireAuth_InvalidFormat(t *testing.T) {
	auth := &mockAuthService{}

	handler := middleware.RequireAuth(auth)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called")
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "InvalidFormat")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", rec.Code)
	}
}

func TestRequireAuth_InvalidToken(t *testing.T) {
	auth := &mockAuthService{err: errors.New("invalid token")}

	handler := middleware.RequireAuth(auth)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called")
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer bad-token")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", rec.Code)
	}
}

func TestGetUserID_EmptyContext(t *testing.T) {
	id := middleware.GetUserID(context.Background())
	if id != uuid.Nil {
		t.Errorf("expected nil UUID, got %s", id)
	}
}

func TestGetUserRole_EmptyContext(t *testing.T) {
	role := middleware.GetUserRole(context.Background())
	if role != "" {
		t.Errorf("expected empty role, got %s", role)
	}
}

func TestOptionalAuth_WithValidToken(t *testing.T) {
	userID := uuid.New()
	auth := &mockAuthService{userID: userID, role: domain.RoleClient}

	handler := middleware.OptionalAuth(auth)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotID := middleware.GetUserID(r.Context())
		gotRole := middleware.GetUserRole(r.Context())
		if gotID != userID {
			t.Errorf("expected user ID %s, got %s", userID, gotID)
		}
		if gotRole != domain.RoleClient {
			t.Errorf("expected role %s, got %s", domain.RoleClient, gotRole)
		}
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
}

func TestOptionalAuth_WithoutToken(t *testing.T) {
	auth := &mockAuthService{userID: uuid.New(), role: domain.RoleClient}

	handler := middleware.OptionalAuth(auth)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotID := middleware.GetUserID(r.Context())
		if gotID != uuid.Nil {
			t.Errorf("expected nil UUID without token, got %s", gotID)
		}
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
}

func TestOptionalAuth_WithInvalidToken(t *testing.T) {
	auth := &mockAuthService{err: errors.New("invalid token")}

	handler := middleware.OptionalAuth(auth)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotID := middleware.GetUserID(r.Context())
		if gotID != uuid.Nil {
			t.Errorf("expected nil UUID with invalid token, got %s", gotID)
		}
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer bad-token")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
}

func TestOptionalAuth_WithInvalidFormat(t *testing.T) {
	auth := &mockAuthService{userID: uuid.New(), role: domain.RoleClient}

	handler := middleware.OptionalAuth(auth)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotID := middleware.GetUserID(r.Context())
		if gotID != uuid.Nil {
			t.Errorf("expected nil UUID with invalid format, got %s", gotID)
		}
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "InvalidFormat")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
}
