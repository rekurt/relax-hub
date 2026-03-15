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
		Filter: AnalyticsFilter{},
		DatePresets: BuildDatePresets(time.Date(2026, 3, 7, 15, 0, 0, 0, time.UTC), AnalyticsFilter{}),
		Summary: AnalyticsSummary{
			TotalBookings: 37,
			TotalRevenue:  4550000,
			TotalNewUsers: 16,
			BookingsTrend: TrendData{Current: 37, Previous: 30},
			RevenueTrend:  TrendData{Current: 4550000, Previous: 3800000},
			UsersTrend:    TrendData{Current: 16, Previous: 20},
		},
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
					DatePresets: BuildDatePresets(time.Now(), AnalyticsFilter{}),
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
			handler := NewAnalyticsHandler(tt.provider, testLogger(), "/admin-panel/pages", "/admin-panel")
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
	handler := NewAnalyticsHandler(provider, testLogger(), "/admin-panel/pages", "/admin-panel")

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
	handler := NewAnalyticsHandler(provider, testLogger(), "/admin-panel/pages", "/admin-panel")

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
	handler := NewAnalyticsHandler(provider, testLogger(), "/admin-panel/pages", "/admin-panel")

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
	handler := NewAnalyticsHandler(provider, testLogger(), "/admin-panel/pages", "/admin-panel")

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

func TestAnalyticsHandler_RendersBaseLayout(t *testing.T) {
	data := sampleAnalyticsData()
	provider := &mockAnalyticsProvider{data: data}
	handler := NewAnalyticsHandler(provider, testLogger(), "/admin-panel/pages", "/admin-panel")

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/analytics", nil)
	handler.ServeHTTP(rec, req)

	body := rec.Body.String()

	checks := []string{
		"sidebar",
		"breadcrumb",
		"<title>Аналитика платформы",
		"Последнее обновление:",
		"/admin-panel/pages/analytics",
		"Назад в GoAdmin",
	}
	for _, c := range checks {
		if !strings.Contains(body, c) {
			t.Errorf("body missing layout element %q", c)
		}
	}
}

func TestAnalyticsHandler_XSSEscapedInChartData(t *testing.T) {
	data := sampleAnalyticsData()
	// Inject XSS payload into bathhouse names
	data.TopByBookings = []RankedItem{
		{Name: `Баня "Огонь & Пар" <script>alert(1)</script>`, Value: 10},
	}
	data.TopByRevenue = []RankedItem{
		{Name: `Test"; alert(document.cookie); "`, Value: 5000},
	}
	provider := &mockAnalyticsProvider{data: data}
	handler := NewAnalyticsHandler(provider, testLogger(), "/admin-panel/pages", "/admin-panel")

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/analytics", nil)
	handler.ServeHTTP(rec, req)

	body := rec.Body.String()

	// Raw XSS payloads must NOT appear in the output
	// html/template automatically escapes in JS string context
	if strings.Contains(body, `<script>alert(1)</script>`) {
		t.Error("XSS: unescaped <script> tag found in chart data")
	}
	if strings.Contains(body, `"; alert(document.cookie); "`) {
		t.Error("XSS: unescaped JS injection found in chart data")
	}

	// html/template uses \u003c escaping for < in JS string context
	if !strings.Contains(body, `\u003cscript\u003e`) {
		t.Error("expected html/template JS-escaped <script> (\\u003c format) in chart data")
	}
	// html/template uses \u0022 for " in JS string context
	if !strings.Contains(body, `\u0022`) {
		t.Error("expected html/template JS-escaped double quote (\\u0022) in chart data")
	}
}

func TestAnalyticsHandler_RendersSummary(t *testing.T) {
	data := sampleAnalyticsData()
	provider := &mockAnalyticsProvider{data: data}
	handler := NewAnalyticsHandler(provider, testLogger(), "/admin-panel/pages", "/admin-panel")

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/analytics", nil)
	handler.ServeHTTP(rec, req)

	body := rec.Body.String()

	checks := []string{
		"summary-bar",
		"Бронирования за период",
		"Выручка за период",
		"Новые пользователи",
		"к пред. периоду",
		"37",  // TotalBookings
		"16",  // TotalNewUsers
		"trend-up",
		"trend-down",
	}
	for _, c := range checks {
		if !strings.Contains(body, c) {
			t.Errorf("body missing summary element %q", c)
		}
	}
}

func TestAnalyticsHandler_RendersDatePresets(t *testing.T) {
	data := sampleAnalyticsData()
	provider := &mockAnalyticsProvider{data: data}
	handler := NewAnalyticsHandler(provider, testLogger(), "/admin-panel/pages", "/admin-panel")

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/analytics", nil)
	handler.ServeHTTP(rec, req)

	body := rec.Body.String()

	presetLabels := []string{
		"Сегодня",
		"7 дней",
		"30 дней",
		"Этот месяц",
		"Прошлый месяц",
		"Этот год",
	}
	for _, label := range presetLabels {
		if !strings.Contains(body, label) {
			t.Errorf("body missing date preset %q", label)
		}
	}

	if !strings.Contains(body, "preset-btn") {
		t.Error("body missing preset-btn class")
	}
}

func TestAnalyticsHandler_RendersCSVButtons(t *testing.T) {
	data := sampleAnalyticsData()
	provider := &mockAnalyticsProvider{data: data}
	handler := NewAnalyticsHandler(provider, testLogger(), "/admin-panel/pages", "/admin-panel")

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/analytics", nil)
	handler.ServeHTTP(rec, req)

	body := rec.Body.String()

	csvTypes := []string{
		"type=bookings",
		"type=revenue",
		"type=users",
		"type=top_bookings",
		"type=top_revenue",
	}
	for _, ct := range csvTypes {
		if !strings.Contains(body, ct) {
			t.Errorf("body missing CSV export link for %q", ct)
		}
	}

	if !strings.Contains(body, "csv-btn") {
		t.Error("body missing csv-btn class")
	}
}

func TestAnalyticsHandler_RendersLocalChartJS(t *testing.T) {
	data := sampleAnalyticsData()
	provider := &mockAnalyticsProvider{data: data}
	handler := NewAnalyticsHandler(provider, testLogger(), "/admin-panel/pages", "/admin-panel")

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/analytics", nil)
	handler.ServeHTTP(rec, req)

	body := rec.Body.String()

	if strings.Contains(body, "cdn.jsdelivr.net") {
		t.Error("Chart.js should be self-hosted, not loaded from CDN")
	}
	if !strings.Contains(body, "/admin-panel/pages/static/chart.min.js") {
		t.Error("body missing self-hosted Chart.js path")
	}
}

func TestAnalyticsHandler_CSVExport(t *testing.T) {
	data := sampleAnalyticsData()
	provider := &mockAnalyticsProvider{data: data}
	handler := NewAnalyticsHandler(provider, testLogger(), "/admin-panel/pages", "/admin-panel")

	tests := []struct {
		name       string
		url        string
		wantStatus int
		wantType   string
		wantChecks []string
	}{
		{
			name:       "bookings csv",
			url:        "/analytics/export?type=bookings&from=2026-03-05&to=2026-03-07",
			wantStatus: http.StatusOK,
			wantType:   "text/csv; charset=utf-8",
			wantChecks: []string{"Дата", "Бронирования", "05.03", "12", "06.03", "18"},
		},
		{
			name:       "revenue csv",
			url:        "/analytics/export?type=revenue&from=2026-03-05&to=2026-03-07",
			wantStatus: http.StatusOK,
			wantType:   "text/csv; charset=utf-8",
			wantChecks: []string{"Дата", "Выручка (руб.)"},
		},
		{
			name:       "users csv",
			url:        "/analytics/export?type=users",
			wantStatus: http.StatusOK,
			wantType:   "text/csv; charset=utf-8",
			wantChecks: []string{"Дата", "Новые пользователи"},
		},
		{
			name:       "top bookings csv",
			url:        "/analytics/export?type=top_bookings",
			wantStatus: http.StatusOK,
			wantType:   "text/csv; charset=utf-8",
			wantChecks: []string{"Название", "Бронирования", "Русская баня на дровах", "45"},
		},
		{
			name:       "top revenue csv",
			url:        "/analytics/export?type=top_revenue",
			wantStatus: http.StatusOK,
			wantType:   "text/csv; charset=utf-8",
			wantChecks: []string{"Название", "Выручка (руб.)", "Русская баня на дровах"},
		},
		{
			name:       "missing type",
			url:        "/analytics/export",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "invalid type",
			url:        "/analytics/export?type=invalid",
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, tt.url, nil)
			handler.HandleCSVExport(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, tt.wantStatus)
			}

			if tt.wantStatus == http.StatusOK {
				ct := rec.Header().Get("Content-Type")
				if ct != tt.wantType {
					t.Errorf("content-type = %q, want %q", ct, tt.wantType)
				}

				disp := rec.Header().Get("Content-Disposition")
				if !strings.Contains(disp, "attachment") {
					t.Error("missing attachment in Content-Disposition")
				}
				if !strings.Contains(disp, ".csv") {
					t.Error("missing .csv in Content-Disposition filename")
				}

				body := rec.Body.String()
				// Check BOM
				if len(body) < 3 || body[0] != 0xEF || body[1] != 0xBB || body[2] != 0xBF {
					t.Error("missing UTF-8 BOM in CSV output")
				}

				for _, c := range tt.wantChecks {
					if !strings.Contains(body, c) {
						t.Errorf("CSV body missing %q", c)
					}
				}
			}
		})
	}
}

func TestAnalyticsHandler_CSVExportProviderError(t *testing.T) {
	provider := &mockAnalyticsProvider{err: context.DeadlineExceeded}
	handler := NewAnalyticsHandler(provider, testLogger(), "/admin-panel/pages", "/admin-panel")

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/analytics/export?type=bookings", nil)
	handler.HandleCSVExport(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}
}

func TestBuildDatePresets(t *testing.T) {
	now := time.Date(2026, 3, 15, 12, 0, 0, 0, time.UTC)

	t.Run("generates all 6 presets", func(t *testing.T) {
		presets := BuildDatePresets(now, AnalyticsFilter{})
		if len(presets) != 6 {
			t.Fatalf("expected 6 presets, got %d", len(presets))
		}

		expectedLabels := []string{"Сегодня", "7 дней", "30 дней", "Этот месяц", "Прошлый месяц", "Этот год"}
		for i, label := range expectedLabels {
			if presets[i].Label != label {
				t.Errorf("preset[%d].Label = %q, want %q", i, presets[i].Label, label)
			}
		}
	})

	t.Run("today preset dates", func(t *testing.T) {
		presets := BuildDatePresets(now, AnalyticsFilter{})
		today := presets[0]
		if today.DateFrom != "2026-03-15" || today.DateTo != "2026-03-15" {
			t.Errorf("today preset: from=%q to=%q, want 2026-03-15 for both", today.DateFrom, today.DateTo)
		}
	})

	t.Run("7 days preset dates", func(t *testing.T) {
		presets := BuildDatePresets(now, AnalyticsFilter{})
		week := presets[1]
		if week.DateFrom != "2026-03-09" || week.DateTo != "2026-03-15" {
			t.Errorf("7 days preset: from=%q to=%q, want 2026-03-09 to 2026-03-15", week.DateFrom, week.DateTo)
		}
	})

	t.Run("this month preset dates", func(t *testing.T) {
		presets := BuildDatePresets(now, AnalyticsFilter{})
		month := presets[3]
		if month.DateFrom != "2026-03-01" || month.DateTo != "2026-03-15" {
			t.Errorf("this month preset: from=%q to=%q, want 2026-03-01 to 2026-03-15", month.DateFrom, month.DateTo)
		}
	})

	t.Run("last month preset dates", func(t *testing.T) {
		presets := BuildDatePresets(now, AnalyticsFilter{})
		lastMonth := presets[4]
		if lastMonth.DateFrom != "2026-02-01" || lastMonth.DateTo != "2026-02-28" {
			t.Errorf("last month preset: from=%q to=%q, want 2026-02-01 to 2026-02-28", lastMonth.DateFrom, lastMonth.DateTo)
		}
	})

	t.Run("this year preset dates", func(t *testing.T) {
		presets := BuildDatePresets(now, AnalyticsFilter{})
		year := presets[5]
		if year.DateFrom != "2026-01-01" || year.DateTo != "2026-03-15" {
			t.Errorf("this year preset: from=%q to=%q, want 2026-01-01 to 2026-03-15", year.DateFrom, year.DateTo)
		}
	})

	t.Run("active preset highlighted", func(t *testing.T) {
		filter := AnalyticsFilter{DateFrom: "2026-03-15", DateTo: "2026-03-15"}
		presets := BuildDatePresets(now, filter)

		if !presets[0].Active {
			t.Error("expected today preset to be active")
		}
		for i := 1; i < len(presets); i++ {
			if presets[i].Active {
				t.Errorf("preset[%d] (%q) should not be active", i, presets[i].Label)
			}
		}
	})

	t.Run("no preset active when filter matches none", func(t *testing.T) {
		filter := AnalyticsFilter{DateFrom: "2026-03-10", DateTo: "2026-03-12"}
		presets := BuildDatePresets(now, filter)

		for i, p := range presets {
			if p.Active {
				t.Errorf("preset[%d] (%q) should not be active for custom range", i, p.Label)
			}
		}
	})
}

func TestServeStatic(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/static/chart.min.js", nil)
	ServeStatic(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	body := rec.Body.String()
	if len(body) == 0 {
		t.Error("expected non-empty static file body")
	}
}

func TestServeStatic_NotFound(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/static/nonexistent.js", nil)
	ServeStatic(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

func TestAnalyticsSummaryTrendPercent(t *testing.T) {
	tests := []struct {
		name     string
		trend    TrendData
		wantText string
	}{
		{"increase", TrendData{Current: 37, Previous: 30}, "23%"},
		{"decrease", TrendData{Current: 16, Previous: 20}, "20%"},
		{"zero previous", TrendData{Current: 10, Previous: 0}, "0%"},
		{"no change", TrendData{Current: 10, Previous: 10}, "0%"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SummaryTrendPercent(tt.trend)
			if got != tt.wantText {
				t.Errorf("SummaryTrendPercent(%v) = %q, want %q", tt.trend, got, tt.wantText)
			}
		})
	}
}

func TestAnalyticsHandler_RendersResponsiveDesign(t *testing.T) {
	data := sampleAnalyticsData()
	provider := &mockAnalyticsProvider{data: data}
	handler := NewAnalyticsHandler(provider, testLogger(), "/admin-panel/pages", "/admin-panel")

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/analytics", nil)
	handler.ServeHTTP(rec, req)

	body := rec.Body.String()

	// Chart grid should use 300px minmax
	if !strings.Contains(body, "minmax(300px") {
		t.Error("body missing reduced chart-grid minmax (300px)")
	}

	// Chart cards should have overflow-x: auto for horizontal scrolling
	if !strings.Contains(body, "overflow-x: auto") {
		t.Error("body missing overflow-x: auto for chart cards")
	}

	// Responsive media queries should be present
	if !strings.Contains(body, "@media (max-width: 768px)") {
		t.Error("body missing mobile media query (768px)")
	}
}
