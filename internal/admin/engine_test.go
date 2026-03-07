package admin

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCustomMenuConfig_Structure(t *testing.T) {
	prefix := "/admin-panel/pages"
	items := CustomMenuConfig(prefix)

	if len(items) != 5 {
		t.Fatalf("expected 5 menu items, got %d", len(items))
	}

	// Dashboard: first item with home icon.
	dash := items[0]
	if dash.Title != "Дашборд" {
		t.Errorf("dashboard title: got %q", dash.Title)
	}
	if dash.Icon != "fa-dashboard" {
		t.Errorf("dashboard icon: got %q", dash.Icon)
	}
	if dash.URI != prefix+"/" {
		t.Errorf("dashboard URI: got %q, want %q", dash.URI, prefix+"/")
	}
	if dash.Order != 1 {
		t.Errorf("dashboard order: got %d, want 1", dash.Order)
	}

	// Operations group.
	ops := items[1]
	if ops.Title != "Операции" {
		t.Errorf("operations title: got %q", ops.Title)
	}
	if ops.Header != "Операции" {
		t.Errorf("operations header: got %q", ops.Header)
	}
	if ops.URI != "" {
		t.Errorf("operations URI should be empty, got %q", ops.URI)
	}

	// Moderation child.
	mod := items[2]
	if mod.Title != "Модерация отзывов" {
		t.Errorf("moderation title: got %q", mod.Title)
	}
	if mod.URI != prefix+"/moderation" {
		t.Errorf("moderation URI: got %q", mod.URI)
	}

	// Analytics child.
	analytics := items[3]
	if analytics.Title != "Аналитика" {
		t.Errorf("analytics title: got %q", analytics.Title)
	}
	if analytics.URI != prefix+"/analytics" {
		t.Errorf("analytics URI: got %q", analytics.URI)
	}

	// Health child.
	health := items[4]
	if health.Title != "Мониторинг" {
		t.Errorf("health title: got %q", health.Title)
	}
	if health.URI != prefix+"/health" {
		t.Errorf("health URI: got %q", health.URI)
	}
}

func TestCustomMenuConfig_CustomPrefix(t *testing.T) {
	prefix := "/custom/admin/pages"
	items := CustomMenuConfig(prefix)

	for _, item := range items {
		if item.URI != "" && item.URI[:len(prefix)] != prefix {
			t.Errorf("item %q URI %q does not start with prefix %q", item.Title, item.URI, prefix)
		}
	}
}

func TestCustomMenuConfig_DashboardIsFirst(t *testing.T) {
	items := CustomMenuConfig("/pages")
	if items[0].Order != 1 {
		t.Errorf("dashboard should have order 1, got %d", items[0].Order)
	}
	// Operations group should come after dashboard.
	if items[1].Order <= items[0].Order {
		t.Errorf("operations order %d should be > dashboard order %d", items[1].Order, items[0].Order)
	}
}

func TestCustomMenuConfig_ChildOrdering(t *testing.T) {
	items := CustomMenuConfig("/pages")
	// Children (items 2-4) should have ascending order.
	for i := 3; i < len(items); i++ {
		if items[i].Order <= items[i-1].Order {
			t.Errorf("item %q (order %d) should come after %q (order %d)",
				items[i].Title, items[i].Order, items[i-1].Title, items[i-1].Order)
		}
	}
}

func TestPagesRouter_AllRoutesRegistered(t *testing.T) {
	routes := []struct {
		method string
		path   string
		status int
	}{
		{"GET", "/", http.StatusOK},
		{"GET", "/dashboard", http.StatusOK},
		{"GET", "/moderation", http.StatusOK},
		{"GET", "/analytics", http.StatusOK},
		{"GET", "/health", http.StatusOK},
	}

	handler := PagesRouter(
		&mockDashProvider{},
		&mockModProvider{},
		&mockAnalyticsProvider{},
		&mockHealthProvider{},
		testLogger(),
		"/admin-panel",
	)

	for _, tc := range routes {
		t.Run(tc.method+" "+tc.path, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, nil)
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)
			if rec.Code != tc.status {
				t.Errorf("got status %d, want %d", rec.Code, tc.status)
			}
		})
	}
}

func TestPagesRouter_ModerationAPIRoutes(t *testing.T) {
	routes := []struct {
		method string
		path   string
	}{
		{"POST", "/moderation/api/approve?id=00000000-0000-0000-0000-000000000001"},
		{"POST", "/moderation/api/reject?id=00000000-0000-0000-0000-000000000001"},
		{"POST", "/moderation/api/batch-approve"},
		{"POST", "/moderation/api/batch-reject"},
	}

	handler := PagesRouter(
		&mockDashProvider{},
		&mockModProvider{},
		&mockAnalyticsProvider{},
		&mockHealthProvider{},
		testLogger(),
		"/admin-panel",
	)

	for _, tc := range routes {
		t.Run(tc.method+" "+tc.path, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, nil)
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)
			// These should not 404 (route exists).
			if rec.Code == http.StatusNotFound || rec.Code == http.StatusMethodNotAllowed {
				t.Errorf("route not found: got status %d", rec.Code)
			}
		})
	}
}

func TestPagesRouter_HealthEndpoint(t *testing.T) {
	handler := PagesRouter(
		&mockDashProvider{},
		&mockModProvider{},
		&mockAnalyticsProvider{},
		&mockHealthProvider{},
		testLogger(),
		"/admin-panel",
	)

	req := httptest.NewRequest("GET", "/health", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("health page: got status %d, want 200", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "text/html; charset=utf-8" {
		t.Errorf("health content-type: got %q", ct)
	}
}
