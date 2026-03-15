package pages

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// mockHealthProvider is a test implementation of HealthDataProvider.
type mockHealthProvider struct {
	data *HealthData
	err  error
}

func (m *mockHealthProvider) GetHealthData(_ context.Context) (*HealthData, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.data, nil
}

func sampleHealthData() *HealthData {
	return &HealthData{
		Services: []ServiceStatus{
			{Name: "PostgreSQL", Status: "up", Latency: 2 * time.Millisecond},
			{Name: "Redis", Status: "up", Latency: 1 * time.Millisecond},
		},
		WebSocketClients: 42,
		ModerationBacklog: ModerationBacklog{
			Over24h: 5,
			Over48h: 2,
			Over72h: 1,
		},
		SubscriptionHealth: SubscriptionHealth{
			ExpiringThisWeek: 3,
		},
		System: SystemMetrics{
			Uptime:        2*time.Hour + 30*time.Minute,
			GoVersion:     "go1.23.0",
			NumGoroutines: 15,
			NumCPU:        8,
			MemAlloc:      52428800,  // 50 MB
			MemTotalAlloc: 104857600, // 100 MB
			MemSys:        73400320,  // 70 MB
			NumGC:         42,
		},
		Pool: PoolStats{
			AcquiredConns:     3,
			IdleConns:         7,
			TotalConns:        10,
			MaxConns:          20,
			AcquireCount:      1500,
			EmptyAcquireCount: 5,
			AcquireDuration:   150 * time.Millisecond,
		},
		DatabaseSize:    "256 MB",
		FilesystemOK:    true,
		FilesystemError: "",
		GeneratedAt:     time.Date(2026, 3, 7, 15, 0, 0, 0, time.UTC),
	}
}

func TestHealthHandler_ServeHTTP(t *testing.T) {
	tests := []struct {
		name       string
		provider   *mockHealthProvider
		wantStatus int
	}{
		{
			name:       "success with data",
			provider:   &mockHealthProvider{data: sampleHealthData()},
			wantStatus: http.StatusOK,
		},
		{
			name: "success with all services down",
			provider: &mockHealthProvider{
				data: &HealthData{
					Services: []ServiceStatus{
						{Name: "PostgreSQL", Status: "down", Error: "connection refused"},
						{Name: "Redis", Status: "down", Error: "timeout"},
					},
					GeneratedAt: time.Now(),
				},
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "success with empty data",
			provider: &mockHealthProvider{
				data: &HealthData{GeneratedAt: time.Now()},
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "provider error",
			provider:   &mockHealthProvider{err: context.DeadlineExceeded},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewHealthHandler(tt.provider, testLogger(), "/admin-panel/pages", "/admin-panel")
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/health", nil)

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

func TestHealthHandler_RendersServiceStatus(t *testing.T) {
	data := sampleHealthData()
	provider := &mockHealthProvider{data: data}
	handler := NewHealthHandler(provider, testLogger(), "/admin-panel/pages", "/admin-panel")

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	handler.ServeHTTP(rec, req)

	body := rec.Body.String()
	checks := []string{
		"PostgreSQL",
		"Redis",
		"Работает",
		"42", // WebSocket clients
	}
	for _, c := range checks {
		if !strings.Contains(body, c) {
			t.Errorf("body missing expected value %q", c)
		}
	}
}

func TestHealthHandler_RendersDownServices(t *testing.T) {
	data := &HealthData{
		Services: []ServiceStatus{
			{Name: "PostgreSQL", Status: "down", Error: "connection refused"},
		},
		GeneratedAt: time.Now(),
	}
	provider := &mockHealthProvider{data: data}
	handler := NewHealthHandler(provider, testLogger(), "/admin-panel/pages", "/admin-panel")

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	handler.ServeHTTP(rec, req)

	body := rec.Body.String()
	if !strings.Contains(body, "Недоступен") {
		t.Error("body missing 'Недоступен' badge for down service")
	}
	// Real error message should be shown in collapsed details
	if !strings.Contains(body, "connection refused") {
		t.Error("body missing real error message 'connection refused'")
	}
}

func TestHealthHandler_RendersUnconfiguredRedis(t *testing.T) {
	data := &HealthData{
		Services: []ServiceStatus{
			{Name: "PostgreSQL", Status: "up", Latency: time.Millisecond},
			{Name: "Redis", Status: "unconfigured"},
		},
		GeneratedAt: time.Now(),
	}
	provider := &mockHealthProvider{data: data}
	handler := NewHealthHandler(provider, testLogger(), "/admin-panel/pages", "/admin-panel")

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	handler.ServeHTTP(rec, req)

	body := rec.Body.String()

	if !strings.Contains(body, "Не настроен") {
		t.Error("body missing 'Не настроен' badge for unconfigured service")
	}

	if !strings.Contains(body, "unconfigured") {
		t.Error("body missing 'unconfigured' CSS class")
	}

	if strings.Count(body, "Недоступен") > 0 {
		t.Error("unconfigured Redis should not show 'Недоступен'")
	}
}

func TestHealthHandler_ShowsRealErrorMessage(t *testing.T) {
	data := &HealthData{
		Services: []ServiceStatus{
			{Name: "PostgreSQL", Status: "down", Error: "dial tcp 127.0.0.1:5432: connect: connection refused"},
		},
		GeneratedAt: time.Now(),
	}
	provider := &mockHealthProvider{data: data}
	handler := NewHealthHandler(provider, testLogger(), "/admin-panel/pages", "/admin-panel")

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	handler.ServeHTTP(rec, req)

	body := rec.Body.String()

	if !strings.Contains(body, "dial tcp 127.0.0.1:5432") {
		t.Error("body missing real error message")
	}
	if !strings.Contains(body, "<details>") {
		t.Error("expected error to be in a collapsible details element")
	}
}

func TestHealthHandler_RendersModerationBacklog(t *testing.T) {
	data := sampleHealthData()
	provider := &mockHealthProvider{data: data}
	handler := NewHealthHandler(provider, testLogger(), "/admin-panel/pages", "/admin-panel")

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	handler.ServeHTTP(rec, req)

	body := rec.Body.String()
	// Backlog values should be present with severity classes
	if !strings.Contains(body, "severity-warning") {
		t.Error("expected 'severity-warning' class for >24h backlog count 5")
	}
}

func TestHealthHandler_RendersAutoRefresh(t *testing.T) {
	data := sampleHealthData()
	provider := &mockHealthProvider{data: data}
	handler := NewHealthHandler(provider, testLogger(), "/admin-panel/pages", "/admin-panel")

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	handler.ServeHTTP(rec, req)

	body := rec.Body.String()
	if !strings.Contains(body, "countdown") {
		t.Error("expected auto-refresh countdown element")
	}
	if !strings.Contains(body, "location.reload") {
		t.Error("expected auto-refresh reload script")
	}
	// Check for play/pause button
	if !strings.Contains(body, "refresh-toggle") {
		t.Error("expected play/pause toggle button for auto-refresh")
	}
	// Check for interval selector
	if !strings.Contains(body, "interval-select") {
		t.Error("expected interval selector for auto-refresh")
	}
	// Check for countdown bar
	if !strings.Contains(body, "countdown-fill") {
		t.Error("expected countdown bar element")
	}
	// Check for last-updated time
	if !strings.Contains(body, "Обновлено:") {
		t.Error("expected 'Обновлено:' last-updated text")
	}
}

func TestHealthHandler_ZeroBacklogShowsOk(t *testing.T) {
	data := &HealthData{
		Services: []ServiceStatus{
			{Name: "PostgreSQL", Status: "up", Latency: time.Millisecond},
		},
		ModerationBacklog: ModerationBacklog{
			Over24h: 0,
			Over48h: 0,
			Over72h: 0,
		},
		GeneratedAt: time.Now(),
	}
	provider := &mockHealthProvider{data: data}
	handler := NewHealthHandler(provider, testLogger(), "/admin-panel/pages", "/admin-panel")

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	handler.ServeHTTP(rec, req)

	body := rec.Body.String()
	if !strings.Contains(body, "severity-ok") {
		t.Error("expected 'severity-ok' class for zero backlog counts")
	}
}

func TestHealthHandler_RendersBaseLayout(t *testing.T) {
	data := sampleHealthData()
	provider := &mockHealthProvider{data: data}
	handler := NewHealthHandler(provider, testLogger(), "/admin-panel/pages", "/admin-panel")

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	handler.ServeHTTP(rec, req)

	body := rec.Body.String()

	checks := []string{
		"sidebar",
		"breadcrumb",
		"<title>Мониторинг платформы",
		"Последнее обновление:",
		"/admin-panel/pages/health",
		"Назад в GoAdmin",
	}
	for _, c := range checks {
		if !strings.Contains(body, c) {
			t.Errorf("body missing layout element %q", c)
		}
	}
}

func TestHealthHandler_RendersSystemMetrics(t *testing.T) {
	data := sampleHealthData()
	provider := &mockHealthProvider{data: data}
	handler := NewHealthHandler(provider, testLogger(), "/admin-panel/pages", "/admin-panel")

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	handler.ServeHTTP(rec, req)

	body := rec.Body.String()

	checks := []string{
		"Uptime",
		"2 ч. 30 мин.",
		"go1.23.0",
		"Горутины",
		"15",
		"50.0 MB",
		"Циклов GC",
	}
	for _, c := range checks {
		if !strings.Contains(body, c) {
			t.Errorf("body missing system metric %q", c)
		}
	}
}

func TestHealthHandler_RendersPoolStats(t *testing.T) {
	data := sampleHealthData()
	provider := &mockHealthProvider{data: data}
	handler := NewHealthHandler(provider, testLogger(), "/admin-panel/pages", "/admin-panel")

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	handler.ServeHTTP(rec, req)

	body := rec.Body.String()

	checks := []string{
		"Connection Pool",
		"3 / 20",  // AcquiredConns / MaxConns
		"10 / 20", // TotalConns / MaxConns
		"pool-ok",
		"Idle",
	}
	for _, c := range checks {
		if !strings.Contains(body, c) {
			t.Errorf("body missing pool stat %q", c)
		}
	}
}

func TestHealthHandler_RendersAdditionalChecks(t *testing.T) {
	data := sampleHealthData()
	provider := &mockHealthProvider{data: data}
	handler := NewHealthHandler(provider, testLogger(), "/admin-panel/pages", "/admin-panel")

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	handler.ServeHTTP(rec, req)

	body := rec.Body.String()

	// Database size
	if !strings.Contains(body, "256 MB") {
		t.Error("body missing database size")
	}

	// Filesystem check
	if !strings.Contains(body, "Файловая система") {
		t.Error("body missing filesystem check section")
	}
	if !strings.Contains(body, "OK") {
		t.Error("body missing filesystem OK status")
	}
}

func TestHealthHandler_RendersFilesystemError(t *testing.T) {
	data := sampleHealthData()
	data.FilesystemOK = false
	data.FilesystemError = "permission denied"
	provider := &mockHealthProvider{data: data}
	handler := NewHealthHandler(provider, testLogger(), "/admin-panel/pages", "/admin-panel")

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	handler.ServeHTTP(rec, req)

	body := rec.Body.String()

	if !strings.Contains(body, "Ошибка") {
		t.Error("body missing filesystem error status")
	}
	if !strings.Contains(body, "permission denied") {
		t.Error("body missing filesystem error detail")
	}
}

func TestHealthHandler_BacklogClickableLinks(t *testing.T) {
	data := sampleHealthData()
	provider := &mockHealthProvider{data: data}
	handler := NewHealthHandler(provider, testLogger(), "/admin-panel/pages", "/admin-panel")

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	handler.ServeHTTP(rec, req)

	body := rec.Body.String()

	// Backlog counts should be wrapped in links to moderation page
	if !strings.Contains(body, "/admin-panel/pages/moderation?status=pending") {
		t.Error("backlog counts should link to moderation page with pending filter")
	}
}

func TestHealthHandler_StatusDotIndicators(t *testing.T) {
	data := &HealthData{
		Services: []ServiceStatus{
			{Name: "PostgreSQL", Status: "up", Latency: time.Millisecond},
			{Name: "Redis", Status: "down", Error: "timeout"},
		},
		GeneratedAt: time.Now(),
	}
	provider := &mockHealthProvider{data: data}
	handler := NewHealthHandler(provider, testLogger(), "/admin-panel/pages", "/admin-panel")

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	handler.ServeHTTP(rec, req)

	body := rec.Body.String()

	if !strings.Contains(body, "status-dot") {
		t.Error("expected status dot indicators")
	}
	if !strings.Contains(body, "pulse-green") {
		t.Error("expected pulsing animation for up status")
	}
}

func TestHealthHandler_FormattedLatency(t *testing.T) {
	data := &HealthData{
		Services: []ServiceStatus{
			{Name: "PostgreSQL", Status: "up", Latency: 2 * time.Millisecond},
		},
		GeneratedAt: time.Now(),
	}
	provider := &mockHealthProvider{data: data}
	handler := NewHealthHandler(provider, testLogger(), "/admin-panel/pages", "/admin-panel")

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	handler.ServeHTTP(rec, req)

	body := rec.Body.String()

	// Should show human-readable format "2 ms" instead of Go format "2ms"
	if !strings.Contains(body, "2 ms") {
		t.Error("expected human-readable latency format '2 ms'")
	}
}

// --- Template function unit tests ---

func TestFormatLatency(t *testing.T) {
	tests := []struct {
		name string
		d    time.Duration
		want string
	}{
		{"microseconds", 500 * time.Microsecond, "500 µs"},
		{"milliseconds", 2 * time.Millisecond, "2 ms"},
		{"many milliseconds", 150 * time.Millisecond, "150 ms"},
		{"seconds", 2500 * time.Millisecond, "2.5 s"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatLatency(tt.d)
			if got != tt.want {
				t.Errorf("FormatLatency(%v) = %q, want %q", tt.d, got, tt.want)
			}
		})
	}
}

func TestBacklogSeverity(t *testing.T) {
	tests := []struct {
		count int64
		want  string
	}{
		{0, "severity-ok"},
		{1, "severity-warning"},
		{5, "severity-warning"},
		{6, "severity-elevated"},
		{20, "severity-elevated"},
		{21, "severity-critical"},
		{100, "severity-critical"},
	}
	for _, tt := range tests {
		got := BacklogSeverity(tt.count)
		if got != tt.want {
			t.Errorf("BacklogSeverity(%d) = %q, want %q", tt.count, got, tt.want)
		}
	}
}

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		bytes uint64
		want  string
	}{
		{500, "500 B"},
		{1024, "1.0 KB"},
		{52428800, "50.0 MB"},
		{1073741824, "1.0 GB"},
		{2684354560, "2.5 GB"},
	}
	for _, tt := range tests {
		got := FormatBytes(tt.bytes)
		if got != tt.want {
			t.Errorf("FormatBytes(%d) = %q, want %q", tt.bytes, got, tt.want)
		}
	}
}

func TestFormatUptime(t *testing.T) {
	tests := []struct {
		d    time.Duration
		want string
	}{
		{5 * time.Minute, "5 мин."},
		{2*time.Hour + 30*time.Minute, "2 ч. 30 мин."},
		{26*time.Hour + 15*time.Minute, "1 дн. 2 ч. 15 мин."},
		{72 * time.Hour, "3 дн. 0 ч. 0 мин."},
	}
	for _, tt := range tests {
		got := FormatUptime(tt.d)
		if got != tt.want {
			t.Errorf("FormatUptime(%v) = %q, want %q", tt.d, got, tt.want)
		}
	}
}

func TestPoolUsageClass(t *testing.T) {
	tests := []struct {
		current, max int32
		want         string
	}{
		{0, 20, "pool-ok"},
		{5, 20, "pool-ok"},
		{9, 20, "pool-ok"},
		{10, 20, "pool-warning"}, // 50% = warning threshold
		{11, 20, "pool-warning"},
		{15, 20, "pool-warning"},
		{16, 20, "pool-danger"},
		{20, 20, "pool-danger"},
		{5, 0, "pool-ok"}, // edge case: max=0
	}
	for _, tt := range tests {
		got := PoolUsageClass(tt.current, tt.max)
		if got != tt.want {
			t.Errorf("PoolUsageClass(%d, %d) = %q, want %q", tt.current, tt.max, got, tt.want)
		}
	}
}

func TestPoolPercent(t *testing.T) {
	tests := []struct {
		current, max int32
		want         int
	}{
		{0, 20, 0},
		{10, 20, 50},
		{20, 20, 100},
		{5, 0, 0},   // edge case: max=0
		{25, 20, 100}, // over 100% capped
	}
	for _, tt := range tests {
		got := PoolPercent(tt.current, tt.max)
		if got != tt.want {
			t.Errorf("PoolPercent(%d, %d) = %d, want %d", tt.current, tt.max, got, tt.want)
		}
	}
}

func TestHealthHandler_PoolHighUsage(t *testing.T) {
	data := sampleHealthData()
	data.Pool.AcquiredConns = 18
	data.Pool.TotalConns = 19
	data.Pool.MaxConns = 20
	provider := &mockHealthProvider{data: data}
	handler := NewHealthHandler(provider, testLogger(), "/admin-panel/pages", "/admin-panel")

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	handler.ServeHTTP(rec, req)

	body := rec.Body.String()
	if !strings.Contains(body, "pool-danger") {
		t.Error("expected pool-danger class for high pool usage (90%)")
	}
}

func TestHealthHandler_BacklogSeverityThresholds(t *testing.T) {
	tests := []struct {
		name     string
		over24h  int64
		wantCSS  string
	}{
		{"zero shows ok", 0, "severity-ok"},
		{"low shows warning", 3, "severity-warning"},
		{"medium shows elevated", 10, "severity-elevated"},
		{"high shows critical", 25, "severity-critical"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := &HealthData{
				Services:          []ServiceStatus{{Name: "PostgreSQL", Status: "up"}},
				ModerationBacklog: ModerationBacklog{Over24h: tt.over24h, Over48h: 0, Over72h: 0},
				GeneratedAt:       time.Now(),
			}
			provider := &mockHealthProvider{data: data}
			handler := NewHealthHandler(provider, testLogger(), "/admin-panel/pages", "/admin-panel")

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/health", nil)
			handler.ServeHTTP(rec, req)

			body := rec.Body.String()
			if !strings.Contains(body, tt.wantCSS) {
				t.Errorf("expected %q CSS class for backlog count %d", tt.wantCSS, tt.over24h)
			}
		})
	}
}
