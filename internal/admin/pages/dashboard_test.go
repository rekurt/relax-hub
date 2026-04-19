package pages

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/rekurt/relax-hub/internal/logger"
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

func TestDashboardHandler_RendersLocalizedStatuses(t *testing.T) {
	data := sampleDashboardData()
	provider := &mockDashboardProvider{data: data}
	handler := NewDashboardHandler(provider, testLogger(), "/admin-panel/pages", "/admin-panel")

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
	handler.ServeHTTP(rec, req)

	body := rec.Body.String()

	// Booking statuses should be in Russian
	ruStatuses := []string{
		"Подтверждено", // confirmed
		"Ожидает",      // pending
		"Одобрено",     // approved (review)
	}
	for _, s := range ruStatuses {
		if !strings.Contains(body, s) {
			t.Errorf("body missing localized status %q", s)
		}
	}

	// English statuses should NOT appear as badge text (they still appear as CSS classes)
	// We check that "badge confirmed">confirmed doesn't appear (it should be "badge confirmed">Подтверждено)
	if strings.Contains(body, `confirmed">confirmed`) {
		t.Error("found non-localized 'confirmed' status in badge text")
	}
	if strings.Contains(body, `pending">pending`) {
		t.Error("found non-localized 'pending' status in badge text")
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

func TestTrendData_Percent(t *testing.T) {
	tests := []struct {
		name string
		td   TrendData
		want float64
	}{
		{"positive growth", TrendData{Current: 150, Previous: 100}, 50.0},
		{"negative growth", TrendData{Current: 50, Previous: 100}, -50.0},
		{"no change", TrendData{Current: 100, Previous: 100}, 0.0},
		{"zero previous", TrendData{Current: 100, Previous: 0}, 100.0},
		{"both zero", TrendData{Current: 0, Previous: 0}, 0.0},
		{"zero previous zero current", TrendData{Current: 0, Previous: 0}, 0.0},
		{"100% growth", TrendData{Current: 200, Previous: 100}, 100.0},
		{"small growth", TrendData{Current: 105, Previous: 100}, 5.0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.td.Percent()
			if got != tt.want {
				t.Errorf("TrendData{%d,%d}.Percent() = %f, want %f", tt.td.Current, tt.td.Previous, got, tt.want)
			}
		})
	}
}

func TestTrendClass(t *testing.T) {
	tests := []struct {
		name string
		td   TrendData
		want string
	}{
		{"up", TrendData{Current: 150, Previous: 100}, "trend-up"},
		{"down", TrendData{Current: 50, Previous: 100}, "trend-down"},
		{"neutral", TrendData{Current: 100, Previous: 100}, "trend-neutral"},
		{"zero previous", TrendData{Current: 0, Previous: 0}, "trend-neutral"},
		{"zero previous positive current", TrendData{Current: 10, Previous: 0}, "trend-up"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := TrendClass(tt.td)
			if got != tt.want {
				t.Errorf("TrendClass() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestTrendArrow(t *testing.T) {
	tests := []struct {
		name string
		td   TrendData
		want string
	}{
		{"up", TrendData{Current: 150, Previous: 100}, "↑"},
		{"down", TrendData{Current: 50, Previous: 100}, "↓"},
		{"neutral", TrendData{Current: 100, Previous: 100}, "="},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := TrendArrow(tt.td)
			if got != tt.want {
				t.Errorf("TrendArrow() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFormatTrend(t *testing.T) {
	tests := []struct {
		name string
		td   TrendData
		want string
	}{
		{"50 percent up", TrendData{Current: 150, Previous: 100}, "50%"},
		{"50 percent down", TrendData{Current: 50, Previous: 100}, "50%"},
		{"no change", TrendData{Current: 100, Previous: 100}, "0%"},
		{"100 percent", TrendData{Current: 200, Previous: 100}, "100%"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatTrend(tt.td)
			if got != tt.want {
				t.Errorf("FormatTrend() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestRelativeTime(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name string
		t    time.Time
		want string
	}{
		{"just now", now.Add(-10 * time.Second), "только что"},
		{"minutes ago", now.Add(-5 * time.Minute), "5 мин. назад"},
		{"hours ago", now.Add(-3 * time.Hour), "3 ч. назад"},
		{"yesterday", now.Add(-36 * time.Hour), "вчера"},
		{"days ago", now.Add(-72 * time.Hour), "3 дн. назад"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RelativeTime(tt.t)
			if got != tt.want {
				t.Errorf("RelativeTime() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestTruncateText(t *testing.T) {
	tests := []struct {
		name   string
		s      string
		maxLen int
		want   string
	}{
		{"short text", "Hello", 100, "Hello"},
		{"exact length", "Hello", 5, "Hello"},
		{"truncated", "Hello World!", 5, "Hello..."},
		{"empty", "", 10, ""},
		{"unicode", "Привет мир!", 6, "Привет..."},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := TruncateText(tt.s, tt.maxLen)
			if got != tt.want {
				t.Errorf("TruncateText(%q, %d) = %q, want %q", tt.s, tt.maxLen, got, tt.want)
			}
		})
	}
}

func TestDashboardHandler_RendersClickableKPIs(t *testing.T) {
	data := sampleDashboardData()
	provider := &mockDashboardProvider{data: data}
	handler := NewDashboardHandler(provider, testLogger(), "/admin-panel/pages", "/admin-panel")

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
	handler.ServeHTTP(rec, req)

	body := rec.Body.String()

	// KPI cards should be links
	kpiLinks := []string{
		`href="/admin-panel/info/users"`,       // Users KPI → users list
		`href="/admin-panel/info/bathhouses"`,   // Bathhouses KPI → bathhouses list
		`/admin-panel/pages/analytics?date_from`, // Bookings KPI → analytics
	}
	for _, link := range kpiLinks {
		if !strings.Contains(body, link) {
			t.Errorf("body missing KPI link %q", link)
		}
	}

	// Status cards should be links to moderation/GoAdmin
	statusLinks := []string{
		`href="/admin-panel/info/bathhouses?`,
		`href="/admin-panel/pages/moderation?status=pending"`,
	}
	for _, link := range statusLinks {
		if !strings.Contains(body, link) {
			t.Errorf("body missing status card link %q", link)
		}
	}
}

func TestDashboardHandler_RendersRevenueSubCards(t *testing.T) {
	data := sampleDashboardData()
	provider := &mockDashboardProvider{data: data}
	handler := NewDashboardHandler(provider, testLogger(), "/admin-panel/pages", "/admin-panel")

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
	handler.ServeHTTP(rec, req)

	body := rec.Body.String()

	// Revenue should be in 3 separate sub-cards
	if !strings.Contains(body, "revenue-row") {
		t.Error("body missing revenue-row container")
	}
	if !strings.Contains(body, "revenue-card") {
		t.Error("body missing revenue-card elements")
	}
	// Each period should have its label
	for _, label := range []string{"Сегодня", "За неделю", "За месяц"} {
		if !strings.Contains(body, label) {
			t.Errorf("body missing revenue label %q", label)
		}
	}
}

func TestDashboardHandler_RendersTrends(t *testing.T) {
	data := sampleDashboardData()
	// Set some trend data
	data.KPI.UsersTrend = TrendData{Current: 20, Previous: 10}
	data.KPI.BookingsTodayTrend = TrendData{Current: 12, Previous: 15}
	data.KPI.RevenueTodayTrend = TrendData{Current: 1500000, Previous: 1500000}

	provider := &mockDashboardProvider{data: data}
	handler := NewDashboardHandler(provider, testLogger(), "/admin-panel/pages", "/admin-panel")

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
	handler.ServeHTTP(rec, req)

	body := rec.Body.String()

	// Should contain trend classes
	if !strings.Contains(body, "trend-up") {
		t.Error("body missing trend-up class for growing KPI")
	}
	if !strings.Contains(body, "trend-down") {
		t.Error("body missing trend-down class for declining KPI")
	}
	// Users trend: 100% growth
	if !strings.Contains(body, "100%") {
		t.Error("body missing 100% trend for users")
	}
}

func TestDashboardHandler_RendersEmptyStates(t *testing.T) {
	data := &DashboardData{
		GeneratedAt: time.Now(),
		// No bookings, reviews, or registrations
	}
	provider := &mockDashboardProvider{data: data}
	handler := NewDashboardHandler(provider, testLogger(), "/admin-panel/pages", "/admin-panel")

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
	handler.ServeHTTP(rec, req)

	body := rec.Body.String()

	// Each empty feed should show feed-empty with meaningful message
	emptyChecks := []string{
		"feed-empty",
		"Нет бронирований за последнее время",
		"Нет отзывов за последнее время",
		"Нет новых регистраций за последнее время",
	}
	for _, c := range emptyChecks {
		if !strings.Contains(body, c) {
			t.Errorf("body missing empty-state element %q", c)
		}
	}
}

func TestDashboardHandler_RendersViewAllLinks(t *testing.T) {
	data := sampleDashboardData()
	provider := &mockDashboardProvider{data: data}
	handler := NewDashboardHandler(provider, testLogger(), "/admin-panel/pages", "/admin-panel")

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
	handler.ServeHTTP(rec, req)

	body := rec.Body.String()

	viewAllLinks := []string{
		`/admin-panel/info/bookings`,
		`/admin-panel/pages/moderation`,
		`/admin-panel/info/users`,
	}
	for _, link := range viewAllLinks {
		if !strings.Contains(body, link) {
			t.Errorf("body missing view-all link %q", link)
		}
	}
	if !strings.Contains(body, "view-all") {
		t.Error("body missing view-all CSS class")
	}
}

func TestDashboardHandler_RendersAutoRefresh(t *testing.T) {
	data := sampleDashboardData()
	provider := &mockDashboardProvider{data: data}
	handler := NewDashboardHandler(provider, testLogger(), "/admin-panel/pages", "/admin-panel")

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
	handler.ServeHTTP(rec, req)

	body := rec.Body.String()

	if !strings.Contains(body, "refresh-toggle") {
		t.Error("body missing auto-refresh toggle button")
	}
	if !strings.Contains(body, "refresh-status") {
		t.Error("body missing refresh status display")
	}
}

func TestDashboardHandler_RendersReviewPreview(t *testing.T) {
	data := sampleDashboardData()
	data.Reviews = []RecentReview{
		{
			ID:            "r1",
			UserName:      "Тест",
			BathhouseName: "Баня",
			Rating:        5,
			Text:          "Очень длинный текст отзыва который должен быть обрезан до ста символов потому что полный текст не помещается в таблицу и выглядит плохо в интерфейсе",
			Status:        "approved",
			CreatedAt:     time.Now().Add(-1 * time.Hour),
		},
	}
	provider := &mockDashboardProvider{data: data}
	handler := NewDashboardHandler(provider, testLogger(), "/admin-panel/pages", "/admin-panel")

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
	handler.ServeHTTP(rec, req)

	body := rec.Body.String()

	// The preview should contain truncated text (100 chars + "...")
	if !strings.Contains(body, "review-preview") {
		t.Error("body missing review-preview class")
	}
	// The truncated version with "..." should appear in the visible cell text
	if !strings.Contains(body, "...") {
		t.Error("body missing truncation indicator '...'")
	}
	// Full text should still appear in the title attribute for tooltip
	if !strings.Contains(body, "в интерфейсе") {
		t.Error("body missing full review text in title attribute")
	}
}

func TestDashboardHandler_UserNamesTruncated(t *testing.T) {
	data := sampleDashboardData()
	provider := &mockDashboardProvider{data: data}
	handler := NewDashboardHandler(provider, testLogger(), "/admin-panel/pages", "/admin-panel")

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
	handler.ServeHTTP(rec, req)

	body := rec.Body.String()

	// User names in bookings and registrations should have text-truncate class
	// Count occurrences of text-truncate - should be used for user names and bathhouse names
	count := strings.Count(body, "text-truncate")
	// We expect at least: 2 bookings * 2 (name+bathhouse) + 2 reviews * 2 + 1 registration = 9
	if count < 7 {
		t.Errorf("expected at least 7 text-truncate occurrences for names, got %d", count)
	}
}

func TestDashboardHandler_RendersResponsiveDesign(t *testing.T) {
	data := sampleDashboardData()
	provider := &mockDashboardProvider{data: data}
	handler := NewDashboardHandler(provider, testLogger(), "/admin-panel/pages", "/admin-panel")

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
	handler.ServeHTTP(rec, req)

	body := rec.Body.String()

	// Base layout responsive elements
	if !strings.Contains(body, "hamburger-btn") {
		t.Error("body missing hamburger menu button")
	}
	if !strings.Contains(body, "sidebar-overlay") {
		t.Error("body missing sidebar overlay for mobile")
	}

	// Media queries in base CSS
	if !strings.Contains(body, "@media (max-width: 768px)") {
		t.Error("body missing mobile media query (768px)")
	}
	if !strings.Contains(body, "@media (max-width: 1024px)") {
		t.Error("body missing tablet media query (1024px)")
	}

	// Dashboard-specific responsive: feed-grid should use 280px minmax
	if !strings.Contains(body, "minmax(280px") {
		t.Error("body missing reduced feed-grid minmax (280px)")
	}

	// Table responsive wrappers
	if !strings.Contains(body, "table-responsive") {
		t.Error("body missing table-responsive wrapper for horizontal scrolling")
	}
}
