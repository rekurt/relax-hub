package pages

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/nikitaaldaev/bani/internal/logger"
)

func testLogger() *logger.Logger {
	return logger.New(logger.LevelError)
}

// mockDashboardProvider is a test implementation of DashboardDataProvider.
type mockDashboardProvider struct {
	data *DashboardData
	err  error
}

func (m *mockDashboardProvider) GetDashboardData(_ context.Context) (*DashboardData, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.data, nil
}

func sampleDashboardData() *DashboardData {
	return &DashboardData{
		KPI: KPICards{
			TotalUsers:      150,
			TotalBathhouses: 42,
			BookingsToday:   12,
			BookingsWeek:    87,
			BookingsMonth:   310,
			RevenueToday:    1500000,
			RevenueWeek:     8700000,
			RevenueMonth:    31000000,
		},
		Status: StatusCards{
			PendingBathhouses:   3,
			PendingReviews:      7,
			ActiveSubscriptions: 15,
		},
		Bookings: []RecentBooking{
			{
				ID:            "b1",
				UserName:      "Иван Иванов",
				BathhouseName: "Русская баня на дровах",
				StartTime:     time.Date(2026, 3, 7, 14, 0, 0, 0, time.UTC),
				TotalPrice:    500000,
				Status:        "confirmed",
			},
			{
				ID:            "b2",
				UserName:      "Петр Петров",
				BathhouseName: "Финская сауна Релакс",
				StartTime:     time.Date(2026, 3, 7, 16, 0, 0, 0, time.UTC),
				TotalPrice:    300000,
				Status:        "pending",
			},
		},
		Reviews: []RecentReview{
			{
				ID:            "r1",
				UserName:      "Мария Сидорова",
				BathhouseName: "Русская баня на дровах",
				Rating:        5,
				Text:          "Отличная баня!",
				Status:        "approved",
				CreatedAt:     time.Date(2026, 3, 7, 10, 0, 0, 0, time.UTC),
			},
			{
				ID:            "r2",
				UserName:      "Алексей Козлов",
				BathhouseName: "Финская сауна Релакс",
				Rating:        3,
				Text:          "Нормально, но можно лучше",
				Status:        "pending",
				CreatedAt:     time.Date(2026, 3, 6, 18, 0, 0, 0, time.UTC),
			},
		},
		Registrations: []RecentRegistration{
			{
				ID:        "u1",
				Name:      "Новый Пользователь",
				Email:     "new@example.com",
				Role:      "client",
				CreatedAt: time.Date(2026, 3, 7, 12, 0, 0, 0, time.UTC),
			},
		},
		GeneratedAt: time.Date(2026, 3, 7, 15, 0, 0, 0, time.UTC),
	}
}

func TestDashboardHandler_ServeHTTP(t *testing.T) {
	tests := []struct {
		name       string
		provider   *mockDashboardProvider
		wantStatus int
	}{
		{
			name: "success with data",
			provider: &mockDashboardProvider{
				data: sampleDashboardData(),
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "success with empty data",
			provider: &mockDashboardProvider{
				data: &DashboardData{
					GeneratedAt: time.Now(),
				},
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "provider error",
			provider: &mockDashboardProvider{
				err: context.DeadlineExceeded,
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewDashboardHandler(tt.provider, testLogger(), "/admin-panel/pages", "/admin-panel")
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/dashboard", nil)

			handler.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, tt.wantStatus)
			}

			if tt.wantStatus == http.StatusOK {
				ct := rec.Header().Get("Content-Type")
				if ct != "text/html; charset=utf-8" {
					t.Errorf("content-type = %q, want text/html", ct)
				}
				body := rec.Body.String()
				if len(body) == 0 {
					t.Error("expected non-empty response body")
				}
			}
		})
	}
}

func TestDashboardHandler_RendersKPIs(t *testing.T) {
	data := sampleDashboardData()
	provider := &mockDashboardProvider{data: data}
	handler := NewDashboardHandler(provider, testLogger(), "/admin-panel/pages", "/admin-panel")

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
	handler.ServeHTTP(rec, req)

	body := rec.Body.String()
	checks := []string{
		"150",       // total users
		"42",        // total bathhouses
		"12",        // bookings today
		"87",        // bookings week
		"310",       // bookings month
		"15 000.00", // revenue today (1500000 kopecks = 15000 rubles)
	}
	for _, c := range checks {
		if !strings.Contains(body, c) {
			t.Errorf("body missing expected value %q", c)
		}
	}
}

func TestDashboardHandler_RendersStatusCards(t *testing.T) {
	data := sampleDashboardData()
	provider := &mockDashboardProvider{data: data}
	handler := NewDashboardHandler(provider, testLogger(), "/admin-panel/pages", "/admin-panel")

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
	handler.ServeHTTP(rec, req)

	body := rec.Body.String()
	if !strings.Contains(body, "warn") {
		t.Error("expected 'warn' class for non-zero pending counts")
	}
}

func TestDashboardHandler_RendersActivityFeed(t *testing.T) {
	data := sampleDashboardData()
	provider := &mockDashboardProvider{data: data}
	handler := NewDashboardHandler(provider, testLogger(), "/admin-panel/pages", "/admin-panel")

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
	handler.ServeHTTP(rec, req)

	body := rec.Body.String()
	checks := []string{
		"Иван Иванов",
		"Русская баня на дровах",
		"Мария Сидорова",
		"Новый Пользователь",
		"new@example.com",
	}
	for _, c := range checks {
		if !strings.Contains(body, c) {
			t.Errorf("body missing expected value %q", c)
		}
	}
}

func TestDashboardHandler_RendersBaseLayout(t *testing.T) {
	data := sampleDashboardData()
	provider := &mockDashboardProvider{data: data}
	handler := NewDashboardHandler(provider, testLogger(), "/admin-panel/pages", "/admin-panel")

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
	handler.ServeHTTP(rec, req)

	body := rec.Body.String()

	// Check sidebar
	sidebarChecks := []string{
		"sidebar",
		"/admin-panel/pages/dashboard",
		"/admin-panel/pages/moderation",
		"/admin-panel/pages/analytics",
		"/admin-panel/pages/health",
		"Назад в GoAdmin",
	}
	for _, c := range sidebarChecks {
		if !strings.Contains(body, c) {
			t.Errorf("body missing sidebar element %q", c)
		}
	}

	// Check breadcrumb
	if !strings.Contains(body, "breadcrumb") {
		t.Error("body missing breadcrumb navigation")
	}
	if !strings.Contains(body, "Главная") {
		t.Error("body missing breadcrumb 'Главная' link")
	}

	// Check dynamic title
	if !strings.Contains(body, "<title>Панель управления") {
		t.Error("body missing dynamic page title")
	}

	// Check footer
	if !strings.Contains(body, "Последнее обновление:") {
		t.Error("body missing footer with last update time")
	}

	// Check active page
	if !strings.Contains(body, `class="nav-link active"`) {
		t.Error("body missing active nav-link class")
	}
}

func TestFormatKopecksToRubles(t *testing.T) {
	tests := []struct {
		input int64
		want  string
	}{
		{0, "0.00"},
		{100, "1.00"},
		{150050, "1 500.50"},
		{1500000, "15 000.00"},
		{999, "9.99"},
		{10000, "100.00"},
		{100000000, "1 000 000.00"},
		{50, "0.50"},
		{1, "0.01"},
		{-15000, "-150.00"},
		{-1500050, "-15 000.50"},
		{-1, "-0.01"},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := FormatKopecksToRubles(tt.input)
			if got != tt.want {
				t.Errorf("FormatKopecksToRubles(%d) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
