package middleware_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/rekurt/relax-hub/internal/domain"
	"github.com/rekurt/relax-hub/internal/middleware"
)

type mockAdminSubRoleResolver struct {
	subRole domain.AdminSubRole
	err     error
}

func (m *mockAdminSubRoleResolver) GetAdminSubRole(_ context.Context, _ uuid.UUID) (domain.AdminSubRole, error) {
	return m.subRole, m.err
}

type mockAdmin2FAChecker struct {
	has2FA bool
	err    error
}

func (m *mockAdmin2FAChecker) HasEnabled2FA(_ context.Context, _ uuid.UUID) (bool, error) {
	return m.has2FA, m.err
}

func adminContextRequest(userID uuid.UUID, role domain.UserRole, subRole domain.AdminSubRole) *http.Request {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	ctx := middleware.SetUserID(req.Context(), userID)
	ctx = middleware.SetUserRole(ctx, role)
	if subRole != "" {
		ctx = middleware.SetAdminSubRole(ctx, subRole)
	}
	return req.WithContext(ctx)
}

func TestLoadAdminSubRole(t *testing.T) {
	userID := uuid.New()
	resolver := &mockAdminSubRoleResolver{subRole: domain.AdminSubRoleModerator}

	var capturedSubRole domain.AdminSubRole
	handler := middleware.LoadAdminSubRole(resolver)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedSubRole = middleware.GetAdminSubRole(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := adminContextRequest(userID, domain.RoleAdmin, "")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	if capturedSubRole != domain.AdminSubRoleModerator {
		t.Errorf("expected sub-role moderator, got %s", capturedSubRole)
	}
}

func TestLoadAdminSubRole_NonAdmin(t *testing.T) {
	resolver := &mockAdminSubRoleResolver{subRole: domain.AdminSubRoleSuperAdmin}

	called := false
	handler := middleware.LoadAdminSubRole(resolver)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))

	req := adminContextRequest(uuid.New(), domain.RoleClient, "")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if !called {
		t.Error("handler should be called for non-admin users")
	}
}

func TestRequireAdminPermission_Allowed(t *testing.T) {
	called := false
	handler := middleware.RequireAdminPermission(domain.PermBathhouseModerate)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))

	req := adminContextRequest(uuid.New(), domain.RoleAdmin, domain.AdminSubRoleModerator)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if !called {
		t.Error("handler should be called when permission is allowed")
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestRequireAdminPermission_Denied(t *testing.T) {
	handler := middleware.RequireAdminPermission(domain.PermFinanceManage)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called when permission is denied")
	}))

	// Moderator doesn't have finance permission
	req := adminContextRequest(uuid.New(), domain.RoleAdmin, domain.AdminSubRoleModerator)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", rec.Code)
	}
}

func TestRequireAdminPermission_NoSubRole(t *testing.T) {
	handler := middleware.RequireAdminPermission(domain.PermUserManage)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called without sub-role")
	}))

	req := adminContextRequest(uuid.New(), domain.RoleAdmin, "")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", rec.Code)
	}
}

func TestRequireAdminPermission_SuperAdminCanDoAnything(t *testing.T) {
	perms := domain.AllAdminPermissions()
	for _, perm := range perms {
		t.Run(string(perm), func(t *testing.T) {
			called := false
			handler := middleware.RequireAdminPermission(perm)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				called = true
				w.WriteHeader(http.StatusOK)
			}))

			req := adminContextRequest(uuid.New(), domain.RoleAdmin, domain.AdminSubRoleSuperAdmin)
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			if !called {
				t.Error("super_admin should be allowed for all permissions")
			}
		})
	}
}

func TestRequireAdmin2FA_WithEnabled(t *testing.T) {
	checker := &mockAdmin2FAChecker{has2FA: true}

	called := false
	handler := middleware.RequireAdmin2FA(checker)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))

	req := adminContextRequest(uuid.New(), domain.RoleAdmin, domain.AdminSubRoleSuperAdmin)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if !called {
		t.Error("handler should be called when 2FA is enabled")
	}
}

func TestRequireAdmin2FA_WithoutEnabled(t *testing.T) {
	checker := &mockAdmin2FAChecker{has2FA: false}

	handler := middleware.RequireAdmin2FA(checker)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called when 2FA is not enabled")
	}))

	req := adminContextRequest(uuid.New(), domain.RoleAdmin, domain.AdminSubRoleSuperAdmin)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", rec.Code)
	}
}

func TestRequireAdmin2FA_NonAdmin_Passthrough(t *testing.T) {
	checker := &mockAdmin2FAChecker{has2FA: false}

	called := false
	handler := middleware.RequireAdmin2FA(checker)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))

	req := adminContextRequest(uuid.New(), domain.RoleClient, "")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if !called {
		t.Error("handler should pass through for non-admin users")
	}
}
