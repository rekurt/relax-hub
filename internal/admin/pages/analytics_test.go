package pages

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// mockAnalyticsProvider is a test implementation of AnalyticsDataProvider.
type mockAnalyticsProvider struct {
	data *AnalyticsData
	err  error
}

func (m *mockAnalyticsProvider) GetAnalyticsData(_ context.Context, _ AnalyticsFilter) (*AnalyticsData, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.data, nil
}

func sampleAnalyticsData() *AnalyticsData {
	return &AnalyticsData{
		BookingsPerDay: []ChartPoint{
			{Label: "05.03", Value: 12},
			{Label: "06.03", Value: 18},
			{Label: "07.03", Value: 7},
		},
		RevenuePerDay: []ChartPoint{
			{Label: "05.03", Value: 1500000},
			{Label: "06.03", Value: 2200000},
			{Label: "07.03", Value: 850000},
		},
		NewUsersPerDay: []ChartPoint{
			{Label: "05.03", Value: 5},
			{Label: "06.03", Value: 8},
			{Label: "07.03", Value: 3},
		},
		TopByBookings: []RankedItem{
			{Name: "Русская баня на дровах", Value: 45},
			{Name: "Финская сауна Релакс", Value: 30},
		},
		TopByRevenue: []RankedItem{
			{Name: "Русская баня на дровах", Value: 5000000},
			{Name: "Финская сауна Релакс", Value: 3200000},
		},
		BookingStatusDist: []DistributionItem{
			{Label: "confirmed", Value: 120},
			{Label: "pending", Value: 15},
			{Label: "cancelled", Value: 8},
		},
		ReviewRatingDist: []DistributionItem{
			{Label: "5", Value: 50},
			{Label: "4", Value: 30},
			{Label: "3", Value: 10},
			{Label: "2", Value: 5},
			{Label: "1", Value: 2},
		},
		Cities: []CityOption{
			{ID: 1, Name: "Москва"},
			{ID: 2, Name: "Санкт-Петербург"},
		},
		Filter:      AnalyticsFilter{},
		GeneratedAt: time.Date(2026, 3, 7, 15, 0, 0, 0, time.UTC),
	}
}

func TestAnalyticsHandler_ServeHTTP(t *testing.T) {
	tests := []struct {
		name       string
		provider   *mockAnalyticsProvider
		wantStatus int
	}{
		{
			name: "success with data",
			provider: &mockAnalyticsProvider{
				data: sampleAnalyticsData(),
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "success with empty data",
			provider: &mockAnalyticsProvider{
				data: &AnalyticsData{
					GeneratedAt: time.Now(),
				},
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "provider error",
			provider: &mockAnalyticsProvider{
				err: context.DeadlineExceeded,
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewAnalyticsHandler(tt.provider, testLogger(), "/admin-panel/pages")
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/analytics", nil)

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

func TestAnalyticsHandler_RendersChartData(t *testing.T) {
	data := sampleAnalyticsData()
	provider := &mockAnalyticsProvider{data: data}
	handler := NewAnalyticsHandler(provider, testLogger(), "/admin-panel/pages")

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/analytics", nil)
	handler.ServeHTTP(rec, req)

	body := rec.Body.String()
	checks := []string{
		"05.03",
		"06.03",
		"07.03",
		"Русская баня на дровах",
		"Финская сауна Релакс",
		"confirmed",
		"bookingsChart",
		"revenueChart",
		"usersChart",
		"topBookingsChart",
		"topRevenueChart",
		"bookingStatusChart",
		"reviewRatingChart",
	}
	for _, c := range checks {
		if !strings.Contains(body, c) {
			t.Errorf("body missing expected value %q", c)
		}
	}
}

func TestAnalyticsHandler_RendersCities(t *testing.T) {
	data := sampleAnalyticsData()
	provider := &mockAnalyticsProvider{data: data}
	handler := NewAnalyticsHandler(provider, testLogger(), "/admin-panel/pages")

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/analytics", nil)
	handler.ServeHTTP(rec, req)

	body := rec.Body.String()
	checks := []string{
		"Москва",
		"Санкт-Петербург",
		"Все города",
	}
	for _, c := range checks {
		if !strings.Contains(body, c) {
			t.Errorf("body missing expected value %q", c)
		}
	}
}

func TestAnalyticsHandler_ParsesQueryParams(t *testing.T) {
	data := sampleAnalyticsData()
	data.Filter = AnalyticsFilter{
		DateFrom: "2026-03-01",
		DateTo:   "2026-03-07",
		CityID:   "1",
	}
	provider := &mockAnalyticsProvider{data: data}
	handler := NewAnalyticsHandler(provider, testLogger(), "/admin-panel/pages")

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/analytics?date_from=2026-03-01&date_to=2026-03-07&city_id=1", nil)
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	body := rec.Body.String()
	if !strings.Contains(body, "2026-03-01") {
		t.Error("body missing date_from value")
	}
	if !strings.Contains(body, "2026-03-07") {
		t.Error("body missing date_to value")
	}
}

func TestAnalyticsHandler_RendersFilterSection(t *testing.T) {
	data := sampleAnalyticsData()
	provider := &mockAnalyticsProvider{data: data}
	handler := NewAnalyticsHandler(provider, testLogger(), "/admin-panel/pages")

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/analytics", nil)
	handler.ServeHTTP(rec, req)

	body := rec.Body.String()
	checks := []string{
		"date_from",
		"date_to",
		"city_id",
		"Применить",
		"Сброс",
	}
	for _, c := range checks {
		if !strings.Contains(body, c) {
			t.Errorf("body missing expected value %q", c)
		}
	}
}

func TestPostgresAnalyticsProvider_ParseDateRange(t *testing.T) {
	p := &PostgresAnalyticsProvider{}

	// Default range: last 30 days
	from, to := p.parseDateRange(AnalyticsFilter{})
	if to.Before(from) {
		t.Error("default: to should be after from")
	}
	diff := to.Sub(from).Hours() / 24
	if diff < 29 || diff > 30 {
		t.Errorf("default range should be ~30 days, got %.1f", diff)
	}

	// Custom range
	from, to = p.parseDateRange(AnalyticsFilter{
		DateFrom: "2026-03-01",
		DateTo:   "2026-03-07",
	})
	if from.Format("2006-01-02") != "2026-03-01" {
		t.Errorf("expected from 2026-03-01, got %s", from.Format("2006-01-02"))
	}
	if to.Format("2006-01-02") != "2026-03-07" {
		t.Errorf("expected to 2026-03-07, got %s", to.Format("2006-01-02"))
	}

	// Invalid dates fall back to defaults
	from2, _ := p.parseDateRange(AnalyticsFilter{DateFrom: "invalid"})
	from3, _ := p.parseDateRange(AnalyticsFilter{})
	if from2.Format("2006-01-02") != from3.Format("2006-01-02") {
		t.Error("invalid date_from should fall back to default")
	}
}
