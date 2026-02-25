package server_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/handler"
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

func testRouterParams() server.RouterParams {
	cors := middleware.NewCORSMiddleware()
	authSvc := &mockAuthServiceForRouter{}

	return server.RouterParams{
		CORS:           cors,
		AuthService:    authSvc,
		AuthHandler:    handler.NewAuthHandler(authSvc, nil),
		BHHandler:      handler.NewBathhouseHandler(nil, nil, nil),
		BookingHandler: handler.NewBookingHandler(nil),
		ReviewHandler:  handler.NewReviewHandler(nil),
		RepHandler:     handler.NewRepresentativeHandler(nil),
		CityHandler:    handler.NewCityHandler(nil),
		AdminHandler:   handler.NewAdminHandler(nil, nil, nil),
	}
}

func TestNewRouter(t *testing.T) {
	router := server.NewRouter(testRouterParams())

	if router == nil {
		t.Fatal("expected non-nil router")
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
