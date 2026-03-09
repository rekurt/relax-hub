package server_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
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

func (m *mockAuthServiceForRouter) Register(_ context.Context, _ service.RegisterInput) (*domain.User, string, error) {
	return nil, "", nil
}

func (m *mockAuthServiceForRouter) Login(_ context.Context, _, _ string) (*domain.User, string, error) {
	return nil, "", nil
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
	cors := middleware.NewCORSMiddleware()
	authSvc := &mockAuthServiceForRouter{}
	log := logger.New(logger.LevelWarn)

	return server.RouterParams{
		Log:            log,
		CORS:           cors,
		AuthService:    authSvc,
		AuthHandler:    handler.NewAuthHandler(authSvc, nil),
		BHHandler:      handler.NewBathhouseHandler(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, ""),
		BookingHandler: handler.NewBookingHandler(nil),
		ReviewHandler:  handler.NewReviewHandler(nil, nil, logger.New(logger.LevelError)),
		FavHandler:     handler.NewFavoriteHandler(nil),
		RepHandler:     handler.NewRepresentativeHandler(nil),
		CityHandler:    handler.NewCityHandler(nil),
		AdminHandler:   handler.NewAdminHandler(nil, nil, nil, nil, nil, &mockAdminNotificationService{}),
		SitemapHandler: handler.NewSitemapHandler(nil, nil, nil, log, ""),
		PromoHandler:   handler.NewPromoHandler(nil),
		MediaHandler:   handler.NewMediaHandler(nil),
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
		t.Logf("Note: WriteTimeout (%v) may be less than ReadTimeout/ReadHeaderTimeout for streaming", server.WriteTimeout)
	}
	if server.IdleTimeout < server.WriteTimeout {
		t.Logf("Note: IdleTimeout (%v) is >= WriteTimeout (%v)", server.IdleTimeout, server.WriteTimeout)
	}
}
