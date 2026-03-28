package handler_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/calendar"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/handler"
)

type mockCalendarService struct {
	exportICalFn               func(ctx context.Context, userID uuid.UUID, userRole domain.UserRole, bathhouseID uuid.UUID) (string, error)
	exportICalByTokenFn        func(ctx context.Context, token string) (string, error)
	getOrCreateCalendarTokenFn func(ctx context.Context, userID uuid.UUID, userRole domain.UserRole, bathhouseID uuid.UUID) (string, error)
	addExternalCalendarFn      func(ctx context.Context, userID uuid.UUID, userRole domain.UserRole, bathhouseID uuid.UUID, calendarURL string, source domain.SlotBlockSource) (*domain.ExternalCalendar, error)
	listExternalCalendarsFn    func(ctx context.Context, userID uuid.UUID, userRole domain.UserRole, bathhouseID uuid.UUID) ([]domain.ExternalCalendar, error)
	removeExternalCalendarFn   func(ctx context.Context, userID uuid.UUID, userRole domain.UserRole, calendarID uuid.UUID) error
	syncExternalCalendarsFn    func(ctx context.Context, userID uuid.UUID, userRole domain.UserRole, bathhouseID uuid.UUID) ([]calendar.CalendarConflict, error)
	getCalendarConflictsFn     func(ctx context.Context, userID uuid.UUID, userRole domain.UserRole, bathhouseID uuid.UUID) ([]calendar.CalendarConflict, error)
	createSlotBlockFn          func(ctx context.Context, userID uuid.UUID, userRole domain.UserRole, bathhouseID uuid.UUID, block *domain.SlotBlock) (*domain.SlotBlock, error)
	deleteSlotBlockFn          func(ctx context.Context, userID uuid.UUID, userRole domain.UserRole, blockID uuid.UUID) error
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

func (m *mockCalendarService) AddExternalCalendar(ctx context.Context, userID uuid.UUID, userRole domain.UserRole, bathhouseID uuid.UUID, calendarURL string, source domain.SlotBlockSource) (*domain.ExternalCalendar, error) {
	return m.addExternalCalendarFn(ctx, userID, userRole, bathhouseID, calendarURL, source)
}

func (m *mockCalendarService) ListExternalCalendars(ctx context.Context, userID uuid.UUID, userRole domain.UserRole, bathhouseID uuid.UUID) ([]domain.ExternalCalendar, error) {
	return m.listExternalCalendarsFn(ctx, userID, userRole, bathhouseID)
}

func (m *mockCalendarService) RemoveExternalCalendar(ctx context.Context, userID uuid.UUID, userRole domain.UserRole, calendarID uuid.UUID) error {
	return m.removeExternalCalendarFn(ctx, userID, userRole, calendarID)
}

func (m *mockCalendarService) SyncExternalCalendars(ctx context.Context, userID uuid.UUID, userRole domain.UserRole, bathhouseID uuid.UUID) ([]calendar.CalendarConflict, error) {
	return m.syncExternalCalendarsFn(ctx, userID, userRole, bathhouseID)
}

func (m *mockCalendarService) GetCalendarConflicts(ctx context.Context, userID uuid.UUID, userRole domain.UserRole, bathhouseID uuid.UUID) ([]calendar.CalendarConflict, error) {
	if m.getCalendarConflictsFn != nil {
		return m.getCalendarConflictsFn(ctx, userID, userRole, bathhouseID)
	}
	return nil, nil
}

func (m *mockCalendarService) CreateSlotBlock(ctx context.Context, userID uuid.UUID, userRole domain.UserRole, bathhouseID uuid.UUID, block *domain.SlotBlock) (*domain.SlotBlock, error) {
	return m.createSlotBlockFn(ctx, userID, userRole, bathhouseID, block)
}

func (m *mockCalendarService) DeleteSlotBlock(ctx context.Context, userID uuid.UUID, userRole domain.UserRole, blockID uuid.UUID) error {
	return m.deleteSlotBlockFn(ctx, userID, userRole, blockID)
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

func TestCalendarHandler_AddExternalCalendar(t *testing.T) {
	calID := uuid.New()
	svc := &mockCalendarService{
		addExternalCalendarFn: func(_ context.Context, _ uuid.UUID, _ domain.UserRole, _ uuid.UUID, url string, source domain.SlotBlockSource) (*domain.ExternalCalendar, error) {
			return &domain.ExternalCalendar{
				ID:          calID,
				BathhouseID: uuid.New(),
				URL:         url,
				Source:      source,
			}, nil
		},
	}
	h := handler.NewCalendarHandler(svc)

	r := chi.NewRouter()
	r.Post("/api/v1/my/bathhouses/{id}/external-calendars", h.AddExternalCalendar)

	bathhouseID := uuid.New()
	body := `{"url":"https://calendar.google.com/feed.ics","source":"google_calendar"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/my/bathhouses/"+bathhouseID.String()+"/external-calendars", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(createTestContext(uuid.New(), domain.RoleOwner))
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("status = %d, want %d; body: %s", w.Code, http.StatusCreated, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), calID.String()) {
		t.Error("expected calendar ID in response")
	}
}

func TestCalendarHandler_AddExternalCalendar_InvalidID(t *testing.T) {
	svc := &mockCalendarService{}
	h := handler.NewCalendarHandler(svc)

	r := chi.NewRouter()
	r.Post("/api/v1/my/bathhouses/{id}/external-calendars", h.AddExternalCalendar)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/my/bathhouses/bad-id/external-calendars", strings.NewReader(`{}`))
	req = req.WithContext(createTestContext(uuid.New(), domain.RoleOwner))
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestCalendarHandler_ListExternalCalendars(t *testing.T) {
	calID := uuid.New()
	svc := &mockCalendarService{
		listExternalCalendarsFn: func(_ context.Context, _ uuid.UUID, _ domain.UserRole, _ uuid.UUID) ([]domain.ExternalCalendar, error) {
			return []domain.ExternalCalendar{
				{ID: calID, URL: "https://example.com/cal.ics", Source: domain.SlotBlockSourceGoogleCalendar},
			}, nil
		},
	}
	h := handler.NewCalendarHandler(svc)

	r := chi.NewRouter()
	r.Get("/api/v1/my/bathhouses/{id}/external-calendars", h.ListExternalCalendars)

	bathhouseID := uuid.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/my/bathhouses/"+bathhouseID.String()+"/external-calendars", nil)
	req = req.WithContext(createTestContext(uuid.New(), domain.RoleOwner))
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d; body: %s", w.Code, http.StatusOK, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), calID.String()) {
		t.Error("expected calendar ID in response")
	}
}

func TestCalendarHandler_RemoveExternalCalendar(t *testing.T) {
	svc := &mockCalendarService{
		removeExternalCalendarFn: func(_ context.Context, _ uuid.UUID, _ domain.UserRole, _ uuid.UUID) error {
			return nil
		},
	}
	h := handler.NewCalendarHandler(svc)

	r := chi.NewRouter()
	r.Delete("/api/v1/my/external-calendars/{id}", h.RemoveExternalCalendar)

	calID := uuid.New()
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/my/external-calendars/"+calID.String(), nil)
	req = req.WithContext(createTestContext(uuid.New(), domain.RoleOwner))
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d; body: %s", w.Code, http.StatusOK, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "deleted") {
		t.Error("expected deleted status in response")
	}
}

func TestCalendarHandler_RemoveExternalCalendar_NotFound(t *testing.T) {
	svc := &mockCalendarService{
		removeExternalCalendarFn: func(_ context.Context, _ uuid.UUID, _ domain.UserRole, _ uuid.UUID) error {
			return domain.ErrNotFound
		},
	}
	h := handler.NewCalendarHandler(svc)

	r := chi.NewRouter()
	r.Delete("/api/v1/my/external-calendars/{id}", h.RemoveExternalCalendar)

	calID := uuid.New()
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/my/external-calendars/"+calID.String(), nil)
	req = req.WithContext(createTestContext(uuid.New(), domain.RoleOwner))
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", w.Code, http.StatusNotFound)
	}
}

func TestCalendarHandler_SyncExternalCalendars(t *testing.T) {
	svc := &mockCalendarService{
		syncExternalCalendarsFn: func(_ context.Context, _ uuid.UUID, _ domain.UserRole, _ uuid.UUID) ([]calendar.CalendarConflict, error) {
			return nil, nil
		},
	}
	h := handler.NewCalendarHandler(svc)

	r := chi.NewRouter()
	r.Post("/api/v1/my/bathhouses/{id}/external-calendars/sync", h.SyncExternalCalendars)

	bathhouseID := uuid.New()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/my/bathhouses/"+bathhouseID.String()+"/external-calendars/sync", nil)
	req = req.WithContext(createTestContext(uuid.New(), domain.RoleOwner))
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d; body: %s", w.Code, http.StatusOK, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "synced") {
		t.Error("expected synced status in response")
	}
}

func TestCalendarHandler_CreateSlotBlock(t *testing.T) {
	blockID := uuid.New()
	svc := &mockCalendarService{
		createSlotBlockFn: func(_ context.Context, _ uuid.UUID, _ domain.UserRole, _ uuid.UUID, block *domain.SlotBlock) (*domain.SlotBlock, error) {
			block.ID = blockID
			return block, nil
		},
	}
	h := handler.NewCalendarHandler(svc)

	r := chi.NewRouter()
	r.Post("/api/v1/my/bathhouses/{id}/slot-blocks", h.CreateSlotBlock)

	bathhouseID := uuid.New()
	start := time.Now().Add(24 * time.Hour).Truncate(time.Second)
	end := start.Add(2 * time.Hour)
	body := `{"start_time":"` + start.Format(time.RFC3339) + `","end_time":"` + end.Format(time.RFC3339) + `","description":"Техобслуживание"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/my/bathhouses/"+bathhouseID.String()+"/slot-blocks", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(createTestContext(uuid.New(), domain.RoleOwner))
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("status = %d, want %d; body: %s", w.Code, http.StatusCreated, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), blockID.String()) {
		t.Error("expected block ID in response")
	}
}

func TestCalendarHandler_CreateSlotBlock_InvalidInput(t *testing.T) {
	svc := &mockCalendarService{
		createSlotBlockFn: func(_ context.Context, _ uuid.UUID, _ domain.UserRole, _ uuid.UUID, _ *domain.SlotBlock) (*domain.SlotBlock, error) {
			return nil, domain.ErrInvalidInput
		},
	}
	h := handler.NewCalendarHandler(svc)

	r := chi.NewRouter()
	r.Post("/api/v1/my/bathhouses/{id}/slot-blocks", h.CreateSlotBlock)

	bathhouseID := uuid.New()
	body := `{"start_time":"2026-03-11T10:00:00Z","end_time":"2026-03-11T08:00:00Z","description":"bad"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/my/bathhouses/"+bathhouseID.String()+"/slot-blocks", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(createTestContext(uuid.New(), domain.RoleOwner))
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestCalendarHandler_DeleteSlotBlock(t *testing.T) {
	svc := &mockCalendarService{
		deleteSlotBlockFn: func(_ context.Context, _ uuid.UUID, _ domain.UserRole, _ uuid.UUID) error {
			return nil
		},
	}
	h := handler.NewCalendarHandler(svc)

	r := chi.NewRouter()
	r.Delete("/api/v1/my/slot-blocks/{id}", h.DeleteSlotBlock)

	blockID := uuid.New()
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/my/slot-blocks/"+blockID.String(), nil)
	req = req.WithContext(createTestContext(uuid.New(), domain.RoleOwner))
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d; body: %s", w.Code, http.StatusOK, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "deleted") {
		t.Error("expected deleted status in response")
	}
}

func TestCalendarHandler_DeleteSlotBlock_Forbidden(t *testing.T) {
	svc := &mockCalendarService{
		deleteSlotBlockFn: func(_ context.Context, _ uuid.UUID, _ domain.UserRole, _ uuid.UUID) error {
			return domain.ErrForbidden
		},
	}
	h := handler.NewCalendarHandler(svc)

	r := chi.NewRouter()
	r.Delete("/api/v1/my/slot-blocks/{id}", h.DeleteSlotBlock)

	blockID := uuid.New()
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/my/slot-blocks/"+blockID.String(), nil)
	req = req.WithContext(createTestContext(uuid.New(), domain.RoleOwner))
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("status = %d, want %d", w.Code, http.StatusForbidden)
	}
}
