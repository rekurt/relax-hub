package handler_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/handler"
)

type mockCalendarService struct {
	exportICalFn               func(ctx context.Context, userID uuid.UUID, userRole domain.UserRole, bathhouseID uuid.UUID) (string, error)
	exportICalByTokenFn        func(ctx context.Context, token string) (string, error)
	getOrCreateCalendarTokenFn func(ctx context.Context, userID uuid.UUID, userRole domain.UserRole, bathhouseID uuid.UUID) (string, error)
}

func (m *mockCalendarService) ExportICal(ctx context.Context, userID uuid.UUID, userRole domain.UserRole, bathhouseID uuid.UUID) (string, error) {
	return m.exportICalFn(ctx, userID, userRole, bathhouseID)
}

func (m *mockCalendarService) ExportICalByToken(ctx context.Context, token string) (string, error) {
	return m.exportICalByTokenFn(ctx, token)
}

func (m *mockCalendarService) GetOrCreateCalendarToken(ctx context.Context, userID uuid.UUID, userRole domain.UserRole, bathhouseID uuid.UUID) (string, error) {
	return m.getOrCreateCalendarTokenFn(ctx, userID, userRole, bathhouseID)
}

func TestCalendarHandler_ExportICal(t *testing.T) {
	icalContent := "BEGIN:VCALENDAR\r\nVERSION:2.0\r\nEND:VCALENDAR\r\n"
	svc := &mockCalendarService{
		exportICalFn: func(_ context.Context, _ uuid.UUID, _ domain.UserRole, _ uuid.UUID) (string, error) {
			return icalContent, nil
		},
	}
	h := handler.NewCalendarHandler(svc)

	r := chi.NewRouter()
	r.Get("/api/v1/my/bathhouses/{id}/calendar.ics", h.ExportICal)

	bathhouseID := uuid.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/my/bathhouses/"+bathhouseID.String()+"/calendar.ics", nil)
	req = req.WithContext(createTestContext(uuid.New(), domain.RoleOwner))
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d; body: %s", w.Code, http.StatusOK, w.Body.String())
	}
	if ct := w.Header().Get("Content-Type"); !strings.Contains(ct, "text/calendar") {
		t.Errorf("Content-Type = %q, want text/calendar", ct)
	}
	if !strings.Contains(w.Body.String(), "BEGIN:VCALENDAR") {
		t.Error("expected iCal content in response body")
	}
}

func TestCalendarHandler_ExportICal_InvalidID(t *testing.T) {
	svc := &mockCalendarService{}
	h := handler.NewCalendarHandler(svc)

	r := chi.NewRouter()
	r.Get("/api/v1/my/bathhouses/{id}/calendar.ics", h.ExportICal)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/my/bathhouses/not-a-uuid/calendar.ics", nil)
	req = req.WithContext(createTestContext(uuid.New(), domain.RoleOwner))
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestCalendarHandler_ExportICalByToken(t *testing.T) {
	icalContent := "BEGIN:VCALENDAR\r\nVERSION:2.0\r\nEND:VCALENDAR\r\n"
	svc := &mockCalendarService{
		exportICalByTokenFn: func(_ context.Context, token string) (string, error) {
			if token == "secret123" {
				return icalContent, nil
			}
			return "", domain.ErrNotFound
		},
	}
	h := handler.NewCalendarHandler(svc)

	r := chi.NewRouter()
	r.Get("/calendar/{token}.ics", h.ExportICalByToken)

	req := httptest.NewRequest(http.MethodGet, "/calendar/secret123.ics", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d; body: %s", w.Code, http.StatusOK, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "BEGIN:VCALENDAR") {
		t.Error("expected iCal content")
	}
}

func TestCalendarHandler_ExportICalByToken_NotFound(t *testing.T) {
	svc := &mockCalendarService{
		exportICalByTokenFn: func(_ context.Context, _ string) (string, error) {
			return "", domain.ErrNotFound
		},
	}
	h := handler.NewCalendarHandler(svc)

	r := chi.NewRouter()
	r.Get("/calendar/{token}.ics", h.ExportICalByToken)

	req := httptest.NewRequest(http.MethodGet, "/calendar/bad-token.ics", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", w.Code, http.StatusNotFound)
	}
}

func TestCalendarHandler_GetCalendarToken(t *testing.T) {
	svc := &mockCalendarService{
		getOrCreateCalendarTokenFn: func(_ context.Context, _ uuid.UUID, _ domain.UserRole, _ uuid.UUID) (string, error) {
			return "my-secret-token", nil
		},
	}
	h := handler.NewCalendarHandler(svc)

	r := chi.NewRouter()
	r.Get("/api/v1/my/bathhouses/{id}/calendar-token", h.GetCalendarToken)

	bathhouseID := uuid.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/my/bathhouses/"+bathhouseID.String()+"/calendar-token", nil)
	req = req.WithContext(createTestContext(uuid.New(), domain.RoleOwner))
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d; body: %s", w.Code, http.StatusOK, w.Body.String())
	}
	body := w.Body.String()
	if !strings.Contains(body, "my-secret-token") {
		t.Errorf("expected token in response, got: %s", body)
	}
	if !strings.Contains(body, "/calendar/my-secret-token.ics") {
		t.Errorf("expected URL in response, got: %s", body)
	}
}
