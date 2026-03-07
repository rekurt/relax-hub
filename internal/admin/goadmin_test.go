package admin

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	gacontext "github.com/GoAdminGroup/go-admin/context"
	gaconfig "github.com/GoAdminGroup/go-admin/modules/config"
	"github.com/google/uuid"
	appconfig "github.com/nikitaaldaev/bani/config"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/logger"
)

type mockAuthService struct {
	userID uuid.UUID
	role   domain.UserRole
	err    error
}

func (m *mockAuthService) ParseToken(_ context.Context, _ string) (uuid.UUID, domain.UserRole, error) {
	return m.userID, m.role, m.err
}

func testLogger() *logger.Logger {
	return logger.New(logger.LevelError)
}

func TestBuildGoAdminConfig_Defaults(t *testing.T) {
	cfg := &appconfig.Config{
		Admin: appconfig.AdminConfig{
			Enabled: true,
		},
		Database: appconfig.DatabaseConfig{
			DSN: "postgres://localhost/test",
		},
	}

	gaCfg := BuildGoAdminConfig(cfg)

	if gaCfg.Theme != "adminlte" {
		t.Errorf("expected default theme 'adminlte', got %s", gaCfg.Theme)
	}
	if gaCfg.UrlPrefix != "/admin-panel" {
		t.Errorf("expected default prefix '/admin-panel', got %s", gaCfg.UrlPrefix)
	}
	if gaCfg.Language != "ru" {
		t.Errorf("expected default language 'ru', got %s", gaCfg.Language)
	}
	if gaCfg.Title != "Бани - Админ-панель" {
		t.Errorf("expected Russian title, got %s", gaCfg.Title)
	}
	if gaCfg.SessionLifeTime != 7200 {
		t.Errorf("expected session lifetime 7200, got %d", gaCfg.SessionLifeTime)
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

func TestNewGoAdmin(t *testing.T) {
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
	log := testLogger()

	ga := NewGoAdmin(cfg, log)
	if ga == nil {
		t.Fatal("expected non-nil GoAdmin")
	}
	if ga.Engine == nil {
		t.Fatal("expected non-nil engine")
	}
	if ga.Config == nil {
		t.Fatal("expected non-nil config")
	}
}

func TestNewAuthProcessor_AdminUser(t *testing.T) {
	adminID := uuid.New()
	auth := &mockAuthService{
		userID: adminID,
		role:   domain.RoleAdmin,
	}
	log := testLogger()

	processor := NewAuthProcessor(auth, log)

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

func TestNewAuthProcessor_NonAdminUser(t *testing.T) {
	auth := &mockAuthService{
		userID: uuid.New(),
		role:   domain.RoleOwner,
	}
	log := testLogger()

	processor := NewAuthProcessor(auth, log)

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

	processor := NewAuthProcessor(auth, log)

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

	processor := NewAuthProcessor(auth, log)

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

	processor := NewAuthProcessor(auth, log)

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

	processor := NewAuthProcessor(auth, log)

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

	processor := NewAuthProcessor(auth, log)

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
