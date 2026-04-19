package admin

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	gacontext "github.com/GoAdminGroup/go-admin/context"
	gaconfig "github.com/GoAdminGroup/go-admin/modules/config"
	"github.com/google/uuid"
	appconfig "github.com/rekurt/relax-hub/config"
	"github.com/rekurt/relax-hub/internal/domain"
	"github.com/rekurt/relax-hub/internal/logger"
)

type mockAuthService struct {
	userID uuid.UUID
	role   domain.UserRole
	err    error
}

func (m *mockAuthService) ParseToken(_ context.Context, _ string) (uuid.UUID, domain.UserRole, error) {
	return m.userID, m.role, m.err
}

func (m *mockAuthService) ParseTokenWithSession(ctx context.Context, token string) (uuid.UUID, domain.UserRole, uuid.UUID, error) {
	userID, role, err := m.ParseToken(ctx, token)
	return userID, role, uuid.Nil, err
}

func testLogger() *logger.Logger {
	return logger.New(logger.LevelError)
}

func TestBuildGoAdminConfig_Defaults(t *testing.T) {
	cfg := &appconfig.Config{
		Admin: appconfig.AdminConfig{
			Enabled:  true,
			Prefix:   "/admin-panel",
			Language: "ru",
			Theme:    "adminlte",
		},
		Database: appconfig.DatabaseConfig{
			DSN: "postgres://localhost/test",
		},
	}

	gaCfg := BuildGoAdminConfig(cfg)

	if gaCfg.Theme != "adminlte" {
		t.Errorf("expected theme 'adminlte', got %s", gaCfg.Theme)
	}
	if gaCfg.UrlPrefix != "/admin-panel" {
		t.Errorf("expected prefix '/admin-panel', got %s", gaCfg.UrlPrefix)
	}
	if gaCfg.Language != "ru" {
		t.Errorf("expected language 'ru', got %s", gaCfg.Language)
	}
	if gaCfg.Title != "Бани - Админ-панель" {
		t.Errorf("expected Russian title, got %s", gaCfg.Title)
	}
	if gaCfg.SessionLifeTime != 7200 {
		t.Errorf("expected session lifetime 7200, got %d", gaCfg.SessionLifeTime)
	}
	if gaCfg.Debug {
		t.Error("expected Debug to be false")
	}
}

func TestBuildGoAdminConfig_CustomValues(t *testing.T) {
	cfg := &appconfig.Config{
		Admin: appconfig.AdminConfig{
			Enabled:  true,
			Prefix:   "/custom-admin",
			Language: "en",
			Theme:    "sword",
		},
		Database: appconfig.DatabaseConfig{
			DSN: "postgres://localhost/test",
		},
	}

	gaCfg := BuildGoAdminConfig(cfg)

	if gaCfg.Theme != "sword" {
		t.Errorf("expected theme 'sword', got %s", gaCfg.Theme)
	}
	if gaCfg.UrlPrefix != "/custom-admin" {
		t.Errorf("expected prefix '/custom-admin', got %s", gaCfg.UrlPrefix)
	}
	if gaCfg.Language != "en" {
		t.Errorf("expected language 'en', got %s", gaCfg.Language)
	}
}

func TestBuildGoAdminConfig_DatabaseDSN(t *testing.T) {
	dsn := "postgres://user:pass@db:5432/mydb?sslmode=require"
	cfg := &appconfig.Config{
		Admin:    appconfig.AdminConfig{Enabled: true},
		Database: appconfig.DatabaseConfig{DSN: dsn},
	}

	gaCfg := BuildGoAdminConfig(cfg)

	defaultDB := gaCfg.Databases["default"]
	if defaultDB.Dsn != dsn {
		t.Errorf("expected DSN %s, got %s", dsn, defaultDB.Dsn)
	}
	if defaultDB.Driver != gaconfig.DriverPostgresql {
		t.Errorf("expected driver postgresql, got %s", defaultDB.Driver)
	}
}

func TestBuildGoAdminConfig_HidesInternalUI(t *testing.T) {
	cfg := &appconfig.Config{
		Admin:    appconfig.AdminConfig{Enabled: true},
		Database: appconfig.DatabaseConfig{DSN: "postgres://localhost/test"},
	}

	gaCfg := BuildGoAdminConfig(cfg)

	if !gaCfg.HideConfigCenterEntrance {
		t.Error("expected HideConfigCenterEntrance to be true")
	}
	if !gaCfg.HideToolEntrance {
		t.Error("expected HideToolEntrance to be true")
	}
	if !gaCfg.HideAppInfoEntrance {
		t.Error("expected HideAppInfoEntrance to be true")
	}
	if !gaCfg.HidePluginEntrance {
		t.Error("expected HidePluginEntrance to be true")
	}
	if !gaCfg.HideVisitorUserCenterEntrance {
		t.Error("expected HideVisitorUserCenterEntrance to be true")
	}
}

func TestBuildGoAdminConfig_EnvProduction(t *testing.T) {
	cfg := &appconfig.Config{
		Environment: "production",
		Admin: appconfig.AdminConfig{
			Enabled:  true,
			Prefix:   "/admin-panel",
			Language: "ru",
			Theme:    "adminlte",
		},
		Database: appconfig.DatabaseConfig{
			DSN: "postgres://localhost/test",
		},
	}

	gaCfg := BuildGoAdminConfig(cfg)
	if gaCfg.Env != gaconfig.EnvProd {
		t.Errorf("expected EnvProd for production environment, got %s", gaCfg.Env)
	}
}

func TestBuildGoAdminConfig_EnvDev(t *testing.T) {
	cfg := &appconfig.Config{
		Environment: "dev",
		Admin: appconfig.AdminConfig{
			Enabled:  true,
			Prefix:   "/admin-panel",
			Language: "ru",
			Theme:    "adminlte",
		},
		Database: appconfig.DatabaseConfig{
			DSN: "postgres://localhost/test",
		},
	}

	gaCfg := BuildGoAdminConfig(cfg)
	if gaCfg.Env != gaconfig.EnvLocal {
		t.Errorf("expected EnvLocal for dev environment, got %s", gaCfg.Env)
	}
}

func TestNewAuthProcessor_AdminUser(t *testing.T) {
	adminID := uuid.New()
	auth := &mockAuthService{
		userID: adminID,
		role:   domain.RoleAdmin,
	}
	log := testLogger()

	processor := NewAuthProcessor(auth, nil, log, false)

	req := httptest.NewRequest(http.MethodGet, "/admin-panel/", nil)
	req.Header.Set("Authorization", "Bearer valid-admin-token")

	ctx := &gacontext.Context{Request: req}
	user, exists, msg := processor(ctx)

	if !exists {
		t.Fatalf("expected user to exist, got msg: %s", msg)
	}
	if user.UserName != adminID.String() {
		t.Errorf("expected username %s, got %s", adminID.String(), user.UserName)
	}
	if len(user.Roles) == 0 || user.Roles[0].Slug != "administrator" {
		t.Error("expected administrator role")
	}
}

func TestNewAuthProcessor_SetsAdminTokenCookie(t *testing.T) {
	adminID := uuid.New()
	auth := &mockAuthService{
		userID: adminID,
		role:   domain.RoleAdmin,
	}
	log := testLogger()

	processor := NewAuthProcessor(auth, nil, log, false)

	req := httptest.NewRequest(http.MethodGet, "/admin-panel/", nil)
	req.Header.Set("Authorization", "Bearer test-jwt-token")

	resp := &http.Response{Header: make(http.Header)}
	ctx := &gacontext.Context{Request: req, Response: resp}
	_, exists, msg := processor(ctx)

	if !exists {
		t.Fatalf("expected user to exist, got msg: %s", msg)
	}

	cookies := resp.Header.Values("Set-Cookie")
	found := false
	for _, c := range cookies {
		if strings.Contains(c, "admin_token=test-jwt-token") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected admin_token cookie to be set, got headers: %v", cookies)
	}
}

func TestNewAuthProcessor_NonAdminUser(t *testing.T) {
	auth := &mockAuthService{
		userID: uuid.New(),
		role:   domain.RoleOwner,
	}
	log := testLogger()

	processor := NewAuthProcessor(auth, nil, log, false)

	req := httptest.NewRequest(http.MethodGet, "/admin-panel/", nil)
	req.Header.Set("Authorization", "Bearer valid-owner-token")

	ctx := &gacontext.Context{Request: req}
	_, exists, msg := processor(ctx)

	if exists {
		t.Fatal("expected non-admin user to be denied")
	}
	if msg != "access denied: admin role required" {
		t.Errorf("unexpected message: %s", msg)
	}
}

func TestNewAuthProcessor_NoAuth(t *testing.T) {
	auth := &mockAuthService{}
	log := testLogger()

	processor := NewAuthProcessor(auth, nil, log, false)

	req := httptest.NewRequest(http.MethodGet, "/admin-panel/", nil)
	ctx := &gacontext.Context{Request: req}

	_, exists, msg := processor(ctx)
	if exists {
		t.Fatal("expected unauthenticated request to be denied")
	}
	if msg != "no authorization" {
		t.Errorf("unexpected message: %s", msg)
	}
}

func TestNewAuthProcessor_InvalidToken(t *testing.T) {
	auth := &mockAuthService{
		err: domain.ErrUnauthorized,
	}
	log := testLogger()

	processor := NewAuthProcessor(auth, nil, log, false)

	req := httptest.NewRequest(http.MethodGet, "/admin-panel/", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")

	ctx := &gacontext.Context{Request: req}
	_, exists, msg := processor(ctx)

	if exists {
		t.Fatal("expected invalid token to be denied")
	}
	if msg != "invalid token" {
		t.Errorf("unexpected message: %s", msg)
	}
}

func TestNewAuthProcessor_CookieAuth(t *testing.T) {
	adminID := uuid.New()
	auth := &mockAuthService{
		userID: adminID,
		role:   domain.RoleAdmin,
	}
	log := testLogger()

	processor := NewAuthProcessor(auth, nil, log, false)

	req := httptest.NewRequest(http.MethodGet, "/admin-panel/", nil)
	req.AddCookie(&http.Cookie{Name: "admin_token", Value: "cookie-token"})

	ctx := &gacontext.Context{Request: req}
	user, exists, msg := processor(ctx)

	if !exists {
		t.Fatalf("expected cookie auth to work, got msg: %s", msg)
	}
	if user.UserName != adminID.String() {
		t.Errorf("expected username %s, got %s", adminID.String(), user.UserName)
	}
}

func TestNewAuthProcessor_NilRequest(t *testing.T) {
	auth := &mockAuthService{}
	log := testLogger()

	processor := NewAuthProcessor(auth, nil, log, false)

	ctx := &gacontext.Context{Request: nil}
	_, exists, msg := processor(ctx)

	if exists {
		t.Fatal("expected nil request to be denied")
	}
	if msg != "no request" {
		t.Errorf("unexpected message: %s", msg)
	}
}

func TestNewAuthProcessor_InvalidHeaderFormat(t *testing.T) {
	auth := &mockAuthService{}
	log := testLogger()

	processor := NewAuthProcessor(auth, nil, log, false)

	req := httptest.NewRequest(http.MethodGet, "/admin-panel/", nil)
	req.Header.Set("Authorization", "Basic dXNlcjpwYXNz")

	ctx := &gacontext.Context{Request: req}
	_, exists, msg := processor(ctx)

	if exists {
		t.Fatal("expected Basic auth to be denied")
	}
	if msg != "invalid authorization format" {
		t.Errorf("unexpected message: %s", msg)
	}
}

func TestProvideConditionalModule_Disabled(t *testing.T) {
	cfg := &appconfig.Config{
		Admin: appconfig.AdminConfig{Enabled: false},
	}

	opt := ProvideConditionalModule(cfg)
	if opt == nil {
		t.Fatal("expected non-nil fx.Option even when disabled")
	}
}

func TestProvideConditionalModule_Enabled(t *testing.T) {
	cfg := &appconfig.Config{
		Admin: appconfig.AdminConfig{Enabled: true},
		Database: appconfig.DatabaseConfig{
			DSN: "postgres://localhost/test",
		},
	}

	opt := ProvideConditionalModule(cfg)
	if opt == nil {
		t.Fatal("expected non-nil fx.Option when enabled")
	}
}
