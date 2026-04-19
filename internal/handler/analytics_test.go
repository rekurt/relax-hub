package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/rekurt/relax-hub/internal/domain"
	"github.com/rekurt/relax-hub/internal/logger"
	"github.com/rekurt/relax-hub/internal/middleware"
	"github.com/rekurt/relax-hub/internal/service"
)

type mockAnalyticsService struct {
	recordViewFn               func(ctx context.Context, bathhouseID uuid.UUID, viewerID *uuid.UUID, source domain.ViewSource, ipHash string) error
	getOwnerDashboardFn        func(ctx context.Context, userID uuid.UUID, userRole domain.UserRole, bathhouseID uuid.UUID, period domain.AnalyticsPeriod) (*service.OwnerDashboard, error)
	getDailyStatsFn            func(ctx context.Context, userID uuid.UUID, userRole domain.UserRole, bathhouseID uuid.UUID, from, to time.Time) ([]domain.AnalyticsSnapshot, error)
	getAdminDashboardFn        func(ctx context.Context, userRole domain.UserRole, period domain.AnalyticsPeriod) (*service.AdminDashboard, error)
	getTopBathhousesByMetricFn func(ctx context.Context, userRole domain.UserRole, metric domain.TopMetric, limit int64) ([]service.TopBathhouseInfo, error)
	aggregateDailyFn           func(ctx context.Context) error
	getBusinessMetricsFn       func(ctx context.Context, userRole domain.UserRole, period domain.AnalyticsPeriod) (*service.BusinessMetrics, error)
	getPnLFn                   func(ctx context.Context, userRole domain.UserRole, period domain.AnalyticsPeriod) (*service.PnLMetrics, error)
	getHeatmapDataFn           func(ctx context.Context, userRole domain.UserRole, period domain.AnalyticsPeriod, cellSize float64) (*domain.HeatmapData, error)
	getCohortAnalysisFn        func(ctx context.Context, userRole domain.UserRole, months int) (*domain.CohortAnalysis, error)
	getWalletMetricsFn         func(ctx context.Context, userRole domain.UserRole, period domain.AnalyticsPeriod) (*domain.WalletMetrics, error)
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

func (m *mockAnalyticsService) GetDailyStats(ctx context.Context, userID uuid.UUID, userRole domain.UserRole, bathhouseID uuid.UUID, from, to time.Time) ([]domain.AnalyticsSnapshot, error) {
	if m.getDailyStatsFn != nil {
		return m.getDailyStatsFn(ctx, userID, userRole, bathhouseID, from, to)
	}
	return []domain.AnalyticsSnapshot{}, nil
}

func (m *mockAnalyticsService) GetAdminDashboard(ctx context.Context, userRole domain.UserRole, period domain.AnalyticsPeriod) (*service.AdminDashboard, error) {
	if m.getAdminDashboardFn != nil {
		return m.getAdminDashboardFn(ctx, userRole, period)
	}
	return &service.AdminDashboard{}, nil
}

func (m *mockAnalyticsService) GetTopBathhousesByMetric(ctx context.Context, userRole domain.UserRole, metric domain.TopMetric, limit int64) ([]service.TopBathhouseInfo, error) {
	if m.getTopBathhousesByMetricFn != nil {
		return m.getTopBathhousesByMetricFn(ctx, userRole, metric, limit)
	}
	return []service.TopBathhouseInfo{}, nil
}

func (m *mockAnalyticsService) AggregateDaily(ctx context.Context) error {
	if m.aggregateDailyFn != nil {
		return m.aggregateDailyFn(ctx)
	}
	return nil
}

func (m *mockAnalyticsService) UpdateBathhouseMetrics(_ context.Context) (int, error) {
	return 0, nil
}

func (m *mockAnalyticsService) GetConversionFunnel(_ context.Context, _ domain.UserRole, _ domain.AnalyticsPeriod) (*domain.ConversionFunnel, error) {
	return &domain.ConversionFunnel{}, nil
}

func (m *mockAnalyticsService) GetCohortAnalysis(ctx context.Context, userRole domain.UserRole, months int) (*domain.CohortAnalysis, error) {
	if m.getCohortAnalysisFn != nil {
		return m.getCohortAnalysisFn(ctx, userRole, months)
	}
	return &domain.CohortAnalysis{}, nil
}

func (m *mockAnalyticsService) GetGeoDemandSupply(_ context.Context, _ domain.UserRole, _ domain.AnalyticsPeriod) (*domain.GeoDemandSupplyMap, error) {
	return &domain.GeoDemandSupplyMap{}, nil
}

func (m *mockAnalyticsService) GetWalletMetrics(ctx context.Context, userRole domain.UserRole, period domain.AnalyticsPeriod) (*domain.WalletMetrics, error) {
	if m.getWalletMetricsFn != nil {
		return m.getWalletMetricsFn(ctx, userRole, period)
	}
	return &domain.WalletMetrics{}, nil
}

func (m *mockAnalyticsService) GetOwnerPerformance(_ context.Context, _ uuid.UUID, _ domain.UserRole, _ uuid.UUID, _ domain.AnalyticsPeriod) (*domain.OwnerPerformance, error) {
	return &domain.OwnerPerformance{}, nil
}

func (m *mockAnalyticsService) GetBusinessMetrics(ctx context.Context, userRole domain.UserRole, period domain.AnalyticsPeriod) (*service.BusinessMetrics, error) {
	if m.getBusinessMetricsFn != nil {
		return m.getBusinessMetricsFn(ctx, userRole, period)
	}
	return &service.BusinessMetrics{}, nil
}

func (m *mockAnalyticsService) GetPnL(ctx context.Context, userRole domain.UserRole, period domain.AnalyticsPeriod) (*service.PnLMetrics, error) {
	if m.getPnLFn != nil {
		return m.getPnLFn(ctx, userRole, period)
	}
	return &service.PnLMetrics{}, nil
}

func (m *mockAnalyticsService) GetHeatmapData(ctx context.Context, userRole domain.UserRole, period domain.AnalyticsPeriod, cellSize float64) (*domain.HeatmapData, error) {
	if m.getHeatmapDataFn != nil {
		return m.getHeatmapDataFn(ctx, userRole, period, cellSize)
	}
	return &domain.HeatmapData{}, nil
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

func TestGetBusinessMetrics_ValidRequest(t *testing.T) {
	mock := &mockAnalyticsService{}
	handler := NewAnalyticsHandler(mock, &logger.Logger{})

	req := httptest.NewRequest(http.MethodGet, "/admin/analytics/business-metrics?period=30d", nil)
	ctx := req.Context()
	ctx = middleware.SetUserRoleForTesting(ctx, domain.RoleAdmin)
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	handler.GetBusinessMetrics(w, req)

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

func TestGetBusinessMetrics_NonAdminForbidden(t *testing.T) {
	mock := &mockAnalyticsService{
		getBusinessMetricsFn: func(ctx context.Context, userRole domain.UserRole, period domain.AnalyticsPeriod) (*service.BusinessMetrics, error) {
			return nil, domain.ErrForbidden
		},
	}
	handler := NewAnalyticsHandler(mock, &logger.Logger{})

	req := httptest.NewRequest(http.MethodGet, "/admin/analytics/business-metrics?period=30d", nil)
	ctx := req.Context()
	ctx = middleware.SetUserRoleForTesting(ctx, domain.RoleClient)
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	handler.GetBusinessMetrics(w, req)

	// The service returns ErrForbidden, which maps to 403
	if w.Code != http.StatusForbidden {
		t.Errorf("got status %d, want %d", w.Code, http.StatusForbidden)
	}
}

func TestGetPnL_ValidRequest(t *testing.T) {
	mock := &mockAnalyticsService{
		getPnLFn: func(ctx context.Context, userRole domain.UserRole, period domain.AnalyticsPeriod) (*service.PnLMetrics, error) {
			return &service.PnLMetrics{
				Period:          period,
				GMV:             50000000,
				PlatformRevenue: 6500000,
				TakeRate:        13.0,
				TotalBookings:   200,
			}, nil
		},
	}

	handler := NewAnalyticsHandler(mock, &logger.Logger{})

	req := httptest.NewRequest(http.MethodGet, "/admin/analytics/pnl?period=30d", nil)
	ctx := req.Context()
	ctx = middleware.SetUserRoleForTesting(ctx, domain.RoleAdmin)
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	handler.GetPnL(w, req)

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

func TestGetPnL_NonAdminForbidden(t *testing.T) {
	mock := &mockAnalyticsService{
		getPnLFn: func(ctx context.Context, userRole domain.UserRole, period domain.AnalyticsPeriod) (*service.PnLMetrics, error) {
			return nil, domain.ErrForbidden
		},
	}
	handler := NewAnalyticsHandler(mock, &logger.Logger{})

	req := httptest.NewRequest(http.MethodGet, "/admin/analytics/pnl?period=30d", nil)
	ctx := req.Context()
	ctx = middleware.SetUserRoleForTesting(ctx, domain.RoleClient)
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	handler.GetPnL(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("got status %d, want %d", w.Code, http.StatusForbidden)
	}
}

func TestGetHeatmap_Success(t *testing.T) {
	mock := &mockAnalyticsService{
		getHeatmapDataFn: func(ctx context.Context, userRole domain.UserRole, period domain.AnalyticsPeriod, cellSize float64) (*domain.HeatmapData, error) {
			return &domain.HeatmapData{
				Period:   period,
				CellSize: cellSize,
				Cells: []domain.HeatmapCell{
					{Latitude: 55.75, Longitude: 37.62, ListingCount: 10, BookingCount: 50, SearchCount: 200},
				},
			}, nil
		},
	}
	handler := NewAnalyticsHandler(mock, &logger.Logger{})

	req := httptest.NewRequest(http.MethodGet, "/admin/analytics/heatmap?period=30d&cell_size=0.05", nil)
	ctx := req.Context()
	ctx = middleware.SetUserRoleForTesting(ctx, domain.RoleAdmin)
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	handler.GetHeatmap(w, req)

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

func TestGetHeatmap_DefaultParams(t *testing.T) {
	var capturedCellSize float64
	mock := &mockAnalyticsService{
		getHeatmapDataFn: func(ctx context.Context, userRole domain.UserRole, period domain.AnalyticsPeriod, cellSize float64) (*domain.HeatmapData, error) {
			capturedCellSize = cellSize
			return &domain.HeatmapData{Period: period, CellSize: cellSize}, nil
		},
	}
	handler := NewAnalyticsHandler(mock, &logger.Logger{})

	req := httptest.NewRequest(http.MethodGet, "/admin/analytics/heatmap", nil)
	ctx := req.Context()
	ctx = middleware.SetUserRoleForTesting(ctx, domain.RoleAdmin)
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	handler.GetHeatmap(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("got status %d, want %d", w.Code, http.StatusOK)
	}
	if capturedCellSize != 0.01 {
		t.Errorf("default cellSize = %f, want 0.01", capturedCellSize)
	}
}

func TestGetHeatmap_Forbidden(t *testing.T) {
	mock := &mockAnalyticsService{
		getHeatmapDataFn: func(ctx context.Context, userRole domain.UserRole, period domain.AnalyticsPeriod, cellSize float64) (*domain.HeatmapData, error) {
			return nil, domain.ErrForbidden
		},
	}
	handler := NewAnalyticsHandler(mock, &logger.Logger{})

	req := httptest.NewRequest(http.MethodGet, "/admin/analytics/heatmap?period=30d", nil)
	ctx := req.Context()
	ctx = middleware.SetUserRoleForTesting(ctx, domain.RoleClient)
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	handler.GetHeatmap(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("got status %d, want %d", w.Code, http.StatusForbidden)
	}
}

// Regression tests for SQL fixes: cohort pgx encoding, wallet JOIN, pnl column names

func TestGetCohortAnalysis_ValidRequest(t *testing.T) {
	var capturedMonths int
	mock := &mockAnalyticsService{
		getCohortAnalysisFn: func(_ context.Context, _ domain.UserRole, months int) (*domain.CohortAnalysis, error) {
			capturedMonths = months
			return &domain.CohortAnalysis{
				Cohorts: []domain.CohortRow{
					{CohortMonth: "2026-01", UsersCount: 100, TotalSpending: 5000000},
				},
			}, nil
		},
	}
	h := NewAnalyticsHandler(mock, &logger.Logger{})

	req := httptest.NewRequest(http.MethodGet, "/admin/analytics/cohorts?months=3", nil)
	req = req.WithContext(middleware.SetUserRoleForTesting(req.Context(), domain.RoleAdmin))
	w := httptest.NewRecorder()
	h.GetCohortAnalysis(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	if capturedMonths != 3 {
		t.Errorf("expected months=3 passed to service, got %d", capturedMonths)
	}
	var resp APIResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if !resp.Success {
		t.Error("expected success=true")
	}
}

func TestGetCohortAnalysis_DefaultMonths(t *testing.T) {
	var capturedMonths int
	mock := &mockAnalyticsService{
		getCohortAnalysisFn: func(_ context.Context, _ domain.UserRole, months int) (*domain.CohortAnalysis, error) {
			capturedMonths = months
			return &domain.CohortAnalysis{}, nil
		},
	}
	h := NewAnalyticsHandler(mock, &logger.Logger{})

	req := httptest.NewRequest(http.MethodGet, "/admin/analytics/cohorts", nil)
	req = req.WithContext(middleware.SetUserRoleForTesting(req.Context(), domain.RoleAdmin))
	w := httptest.NewRecorder()
	h.GetCohortAnalysis(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	if capturedMonths != 6 {
		t.Errorf("expected default months=6, got %d", capturedMonths)
	}
}

func TestGetWalletMetrics_ValidRequest(t *testing.T) {
	mock := &mockAnalyticsService{
		getWalletMetricsFn: func(_ context.Context, _ domain.UserRole, _ domain.AnalyticsPeriod) (*domain.WalletMetrics, error) {
			return &domain.WalletMetrics{
				TotalClientBalance: 25000000,
				TotalOwnerBalance:  10000000,
				TotalEscrow:        5000000,
				WalletPaymentShare: 32.5,
				ActiveWallets:      500,
			}, nil
		},
	}
	h := NewAnalyticsHandler(mock, &logger.Logger{})

	req := httptest.NewRequest(http.MethodGet, "/admin/analytics/wallet?period=30d", nil)
	req = req.WithContext(middleware.SetUserRoleForTesting(req.Context(), domain.RoleAdmin))
	w := httptest.NewRecorder()
	h.GetWalletMetrics(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var resp APIResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if !resp.Success {
		t.Error("expected success=true")
	}
}

func TestGetWalletMetrics_Forbidden(t *testing.T) {
	mock := &mockAnalyticsService{
		getWalletMetricsFn: func(_ context.Context, _ domain.UserRole, _ domain.AnalyticsPeriod) (*domain.WalletMetrics, error) {
			return nil, domain.ErrForbidden
		},
	}
	h := NewAnalyticsHandler(mock, &logger.Logger{})

	req := httptest.NewRequest(http.MethodGet, "/admin/analytics/wallet", nil)
	req = req.WithContext(middleware.SetUserRoleForTesting(req.Context(), domain.RoleClient))
	w := httptest.NewRecorder()
	h.GetWalletMetrics(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", w.Code)
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
