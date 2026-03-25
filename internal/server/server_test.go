package server_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/config"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/handler"
	"github.com/nikitaaldaev/bani/internal/logger"
	"github.com/nikitaaldaev/bani/internal/middleware"
	"github.com/nikitaaldaev/bani/internal/server"
	"github.com/nikitaaldaev/bani/internal/service"
)

type mockAuthServiceForRouter struct{}

func (m *mockAuthServiceForRouter) ParseToken(_ context.Context, _ string) (uuid.UUID, domain.UserRole, error) {
	return uuid.Nil, "", domain.ErrUnauthorized
}

func (m *mockAuthServiceForRouter) ParseTokenWithSession(ctx context.Context, token string) (uuid.UUID, domain.UserRole, uuid.UUID, error) {
	userID, role, err := m.ParseToken(ctx, token)
	return userID, role, uuid.Nil, err
}

func (m *mockAuthServiceForRouter) ParsePartialToken(_ context.Context, _ string) (uuid.UUID, error) {
	return uuid.Nil, domain.ErrUnauthorized
}

func (m *mockAuthServiceForRouter) Register(_ context.Context, _ service.RegisterInput) (*domain.User, string, error) {
	return nil, "", nil
}

func (m *mockAuthServiceForRouter) Login(_ context.Context, _, _ string) (*service.LoginResult, error) {
	return &service.LoginResult{}, nil
}

func (m *mockAuthServiceForRouter) Complete2FALogin(_ context.Context, _ uuid.UUID) (*domain.User, string, error) {
	return nil, "", nil
}

func (m *mockAuthServiceForRouter) RegisterPhone(_ context.Context, _ service.RegisterPhoneInput) error {
	return nil
}

func (m *mockAuthServiceForRouter) LoginPhone(_ context.Context, _ string) error {
	return nil
}

func (m *mockAuthServiceForRouter) VerifyPhone(_ context.Context, _, _, _ string) (*service.LoginResult, error) {
	return &service.LoginResult{}, nil
}

// noopTwoFAServiceForRouter is a no-op TwoFAService for router tests.
type noopTwoFAServiceForRouter struct{}

func (n *noopTwoFAServiceForRouter) GenerateTOTPSecret(_ context.Context, _ uuid.UUID) (string, string, error) {
	return "", "", nil
}
func (n *noopTwoFAServiceForRouter) EnableTOTP(_ context.Context, _ uuid.UUID, _ string) error {
	return nil
}
func (n *noopTwoFAServiceForRouter) DisableTOTP(_ context.Context, _ uuid.UUID, _ string) error {
	return nil
}
func (n *noopTwoFAServiceForRouter) VerifyTOTP(_ context.Context, _ uuid.UUID, _ string) (bool, error) {
	return false, nil
}
func (n *noopTwoFAServiceForRouter) EnableSMS2FA(_ context.Context, _ uuid.UUID) error  { return nil }
func (n *noopTwoFAServiceForRouter) SendSMS2FA(_ context.Context, _ uuid.UUID) error    { return nil }
func (n *noopTwoFAServiceForRouter) VerifySMS2FA(_ context.Context, _ uuid.UUID, _ string) (bool, error) {
	return false, nil
}

type mockAdminNotificationService struct{}

func (m *mockAdminNotificationService) Send(_ context.Context, _ uuid.UUID, _ domain.NotificationType, _, _ string, _ map[string]string) error {
	return nil
}

func (m *mockAdminNotificationService) List(_ context.Context, _ uuid.UUID, _, _ int) (*domain.PaginatedResult[domain.Notification], error) {
	return &domain.PaginatedResult[domain.Notification]{}, nil
}

func (m *mockAdminNotificationService) MarkAsRead(_ context.Context, _, _ uuid.UUID) error {
	return nil
}

func (m *mockAdminNotificationService) MarkAllAsRead(_ context.Context, _ uuid.UUID) error {
	return nil
}

func (m *mockAdminNotificationService) GetUnreadCount(_ context.Context, _ uuid.UUID) (int64, error) {
	return 0, nil
}

func (m *mockAdminNotificationService) GetPreferences(_ context.Context, _ uuid.UUID) (*domain.NotificationPreferences, error) {
	return &domain.NotificationPreferences{}, nil
}

func (m *mockAdminNotificationService) UpdatePreferences(_ context.Context, _ uuid.UUID, _ *domain.NotificationPreferences) error {
	return nil
}

func testRouterParams() server.RouterParams {
	cors := middleware.NewCORSMiddleware(&config.Config{
		CORS: config.CORSConfig{AllowedOrigins: []string{"*"}},
	})
	authSvc := &mockAuthServiceForRouter{}
	log := logger.New(logger.LevelWarn)

	return server.RouterParams{
		Log:            log,
		Config:         &config.Config{Environment: "dev"},
		CORS:           cors,
		AuthService:    authSvc,
		AuthHandler:    handler.NewAuthHandler(authSvc, nil, &noopTwoFAServiceForRouter{}, nil, nil),
		BHHandler:      handler.NewBathhouseHandler(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, ""),
		BookingHandler: handler.NewBookingHandler(nil, nil),
		ReviewHandler:  handler.NewReviewHandler(nil, nil, logger.New(logger.LevelError)),
		FavHandler:     handler.NewFavoriteHandler(nil),
		RepHandler:     handler.NewRepresentativeHandler(nil),
		CityHandler:    handler.NewCityHandler(nil),
		AdminHandler:   handler.NewAdminHandler(nil, nil, nil, nil, &mockAdminNotificationService{}),
		SitemapHandler: handler.NewSitemapHandler(nil, nil, nil, log, ""),
		PromoHandler:   handler.NewPromoHandler(nil),
		MediaHandler:   handler.NewMediaHandler(nil),
		SessionHandler:     handler.NewSessionHandler(nil),
		SavedSearchHandler: handler.NewSavedSearchHandler(nil, nil),
	}
}

func TestNewRouter(t *testing.T) {
	router := server.NewRouter(testRouterParams())

	if router == nil {
		t.Fatal("expected non-nil router")
	}
}

func TestNewRouter_WithoutGoAdmin(t *testing.T) {
	// Verify router works when GoAdmin is nil (admin disabled)
	p := testRouterParams()
	p.GoAdmin = nil

	router := server.NewRouter(p)
	if router == nil {
		t.Fatal("expected non-nil router even without GoAdmin")
	}

	// Health endpoint should still work
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
}

func TestHealthCheck(t *testing.T) {
	router := server.NewRouter(testRouterParams())

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	body := rec.Body.String()
	if body == "" {
		t.Error("expected non-empty body")
	}
}

func TestHealthCheck_WrongMethod(t *testing.T) {
	router := server.NewRouter(testRouterParams())

	req := httptest.NewRequest(http.MethodPost, "/health", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405, got %d", rec.Code)
	}
}

func TestHTTPServerTimeoutConstants(t *testing.T) {
	// Verify that timeout constants are defined and positive
	if server.ReadHeaderTimeout <= 0 {
		t.Error("ReadHeaderTimeout must be positive")
	}
	if server.ReadTimeout <= 0 {
		t.Error("ReadTimeout must be positive")
	}
	if server.WriteTimeout <= 0 {
		t.Error("WriteTimeout must be positive")
	}
	if server.IdleTimeout <= 0 {
		t.Error("IdleTimeout must be positive")
	}
}

func TestHTTPServerTimeoutOrdering(t *testing.T) {
	// Verify sensible timeout ordering for production
	if server.ReadTimeout < server.ReadHeaderTimeout {
		t.Errorf("ReadTimeout (%v) should be >= ReadHeaderTimeout (%v)", server.ReadTimeout, server.ReadHeaderTimeout)
	}
	if server.WriteTimeout < server.ReadTimeout && server.WriteTimeout < server.ReadHeaderTimeout {
		t.Errorf("WriteTimeout (%v) should be >= ReadTimeout (%v) or ReadHeaderTimeout (%v)", server.WriteTimeout, server.ReadTimeout, server.ReadHeaderTimeout)
	}
	if server.IdleTimeout < server.WriteTimeout {
		t.Errorf("IdleTimeout (%v) should be >= WriteTimeout (%v)", server.IdleTimeout, server.WriteTimeout)
	}
}
