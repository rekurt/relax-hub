package pages

import (
	"context"
	"net/http"
	"net/http/httptest"
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
		GeneratedAt: time.Date(2026, 3, 7, 15, 0, 0, 0, time.UTC),
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
			handler := NewHealthHandler(tt.provider, testLogger())
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
	handler := NewHealthHandler(provider, testLogger())

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
		if !containsStr(body, c) {
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
	handler := NewHealthHandler(provider, testLogger())

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	handler.ServeHTTP(rec, req)

	body := rec.Body.String()
	checks := []string{
		"Недоступен",
		"connection refused",
	}
	for _, c := range checks {
		if !containsStr(body, c) {
			t.Errorf("body missing expected value %q", c)
		}
	}
}

func TestHealthHandler_RendersModerationBacklog(t *testing.T) {
	data := sampleHealthData()
	provider := &mockHealthProvider{data: data}
	handler := NewHealthHandler(provider, testLogger())

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	handler.ServeHTTP(rec, req)

	body := rec.Body.String()
	// Backlog values should be present
	if !containsStr(body, "warn") {
		t.Error("expected 'warn' class for non-zero >24h backlog")
	}
	if !containsStr(body, "danger") {
		t.Error("expected 'danger' class for non-zero >48h or >72h backlog")
	}
}

func TestHealthHandler_RendersAutoRefresh(t *testing.T) {
	data := sampleHealthData()
	provider := &mockHealthProvider{data: data}
	handler := NewHealthHandler(provider, testLogger())

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	handler.ServeHTTP(rec, req)

	body := rec.Body.String()
	if !containsStr(body, "countdown") {
		t.Error("expected auto-refresh countdown element")
	}
	if !containsStr(body, "location.reload") {
		t.Error("expected auto-refresh reload script")
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
	handler := NewHealthHandler(provider, testLogger())

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	handler.ServeHTTP(rec, req)

	body := rec.Body.String()
	if !containsStr(body, "ok") {
		t.Error("expected 'ok' class for zero backlog counts")
	}
}
