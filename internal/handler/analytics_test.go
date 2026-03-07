package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/logger"
	"github.com/nikitaaldaev/bani/internal/middleware"
	"github.com/nikitaaldaev/bani/internal/service"
)

type mockAnalyticsService struct {
	recordViewFn         func(ctx context.Context, bathhouseID uuid.UUID, viewerID *uuid.UUID, source domain.ViewSource, ipHash string) error
	getOwnerDashboardFn  func(ctx context.Context, userID uuid.UUID, userRole domain.UserRole, bathhouseID uuid.UUID, period domain.AnalyticsPeriod) (*service.OwnerDashboard, error)
	getAdminDashboardFn  func(ctx context.Context, userRole domain.UserRole, period domain.AnalyticsPeriod) (*service.AdminDashboard, error)
	aggregateDailyFn     func(ctx context.Context) error
}

func (m *mockAnalyticsService) RecordView(ctx context.Context, bathhouseID uuid.UUID, viewerID *uuid.UUID, source domain.ViewSource, ipHash string) error {
	if m.recordViewFn != nil {
		return m.recordViewFn(ctx, bathhouseID, viewerID, source, ipHash)
	}
	return nil
}

func (m *mockAnalyticsService) GetOwnerDashboard(ctx context.Context, userID uuid.UUID, userRole domain.UserRole, bathhouseID uuid.UUID, period domain.AnalyticsPeriod) (*service.OwnerDashboard, error) {
	if m.getOwnerDashboardFn != nil {
		return m.getOwnerDashboardFn(ctx, userID, userRole, bathhouseID, period)
	}
	return &service.OwnerDashboard{}, nil
}

func (m *mockAnalyticsService) GetAdminDashboard(ctx context.Context, userRole domain.UserRole, period domain.AnalyticsPeriod) (*service.AdminDashboard, error) {
	if m.getAdminDashboardFn != nil {
		return m.getAdminDashboardFn(ctx, userRole, period)
	}
	return &service.AdminDashboard{}, nil
}

func (m *mockAnalyticsService) AggregateDaily(ctx context.Context) error {
	if m.aggregateDailyFn != nil {
		return m.aggregateDailyFn(ctx)
	}
	return nil
}

func TestGetAdminDashboard_ValidRequest(t *testing.T) {
	mock := &mockAnalyticsService{}
	mock.getAdminDashboardFn = func(ctx context.Context, userRole domain.UserRole, period domain.AnalyticsPeriod) (*service.AdminDashboard, error) {
		return &service.AdminDashboard{
			Period:          period,
			TotalUsers:      1000,
			NewUsers:        100,
			TotalBathhouses: 50,
			TotalBookings:   5000,
			TotalRevenue:    5000000,
			AvgRating:       4.3,
		}, nil
	}

	handler := NewAnalyticsHandler(mock, &logger.Logger{})

	req := httptest.NewRequest(http.MethodGet, "/admin/analytics?period=30d", nil)
	ctx := req.Context()
	ctx = middleware.SetUserRoleForTesting(ctx, domain.RoleAdmin)
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	handler.GetAdminDashboard(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("got status %d, want %d", w.Code, http.StatusOK)
	}

	var response APIResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Errorf("failed to decode response: %v", err)
	}
	if !response.Success {
		t.Errorf("expected success=true")
	}
}

func TestGetAdminDashboard_InvalidPeriod(t *testing.T) {
	mock := &mockAnalyticsService{}
	handler := NewAnalyticsHandler(mock, &logger.Logger{})

	req := httptest.NewRequest(http.MethodGet, "/admin/analytics?period=99d", nil)
	ctx := req.Context()
	ctx = middleware.SetUserRoleForTesting(ctx, domain.RoleAdmin)
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	handler.GetAdminDashboard(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("got status %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestGetOwnerDashboard_InvalidPeriod(t *testing.T) {
	mock := &mockAnalyticsService{}
	handler := NewAnalyticsHandler(mock, &logger.Logger{})

	req := httptest.NewRequest(http.MethodGet, "/my/bathhouses/550e8400-e29b-41d4-a716-446655440000/analytics?period=99d", nil)
	ctx := req.Context()
	ctx = middleware.SetUserIDForTesting(ctx, uuid.New())
	ctx = middleware.SetUserRoleForTesting(ctx, domain.RoleOwner)
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	handler.GetOwnerDashboard(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("got status %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestGetTopBathhouses_ValidRequest(t *testing.T) {
	mock := &mockAnalyticsService{}
	mock.getAdminDashboardFn = func(ctx context.Context, userRole domain.UserRole, period domain.AnalyticsPeriod) (*service.AdminDashboard, error) {
		return &service.AdminDashboard{
			TopBathhouses: []service.TopBathhouseInfo{
				{
					BathhouseID: uuid.New(),
					Name:        "Test Bathhouse 1",
					Bookings:    100,
					Revenue:     1000000,
				},
			},
		}, nil
	}

	handler := NewAnalyticsHandler(mock, &logger.Logger{})

	req := httptest.NewRequest(http.MethodGet, "/admin/analytics/top?metric=bookings&limit=10", nil)
	ctx := req.Context()
	ctx = middleware.SetUserRoleForTesting(ctx, domain.RoleAdmin)
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	handler.GetTopBathhouses(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("got status %d, want %d", w.Code, http.StatusOK)
	}
}

func TestGetTopBathhouses_InvalidMetric(t *testing.T) {
	mock := &mockAnalyticsService{}
	handler := NewAnalyticsHandler(mock, &logger.Logger{})

	req := httptest.NewRequest(http.MethodGet, "/admin/analytics/top?metric=invalid&limit=10", nil)
	ctx := req.Context()
	ctx = middleware.SetUserRoleForTesting(ctx, domain.RoleAdmin)
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	handler.GetTopBathhouses(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("got status %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestIPHashExtraction(t *testing.T) {
	tests := []struct {
		name      string
		remoteAddr string
		xForwarded string
	}{
		{
			name:       "extract from RemoteAddr",
			remoteAddr: "192.168.1.1:12345",
		},
		{
			name:       "extract from X-Forwarded-For",
			xForwarded: "203.0.113.195, 70.41.3.18",
			remoteAddr: "192.168.1.1:12345",
		},
		{
			name:       "prefer X-Forwarded-For when both present",
			xForwarded: "10.0.0.1",
			remoteAddr: "192.168.1.1:12345",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.RemoteAddr = tt.remoteAddr
			if tt.xForwarded != "" {
				req.Header.Set("X-Forwarded-For", tt.xForwarded)
			}

			hash := getIPHash(req)
			if hash == "" {
				t.Errorf("expected non-empty hash")
			}
			// Hash should be hex string (32 chars for MD5)
			if len(hash) != 32 {
				t.Errorf("expected 32-char hash, got %d", len(hash))
			}
		})
	}
}
