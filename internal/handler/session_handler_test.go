package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/middleware"
)

type mockSessionService struct {
	createSessionFn            func(ctx context.Context, userID uuid.UUID, deviceInfo, browser, ip string) (*domain.Session, error)
	listSessionsFn             func(ctx context.Context, userID uuid.UUID) ([]domain.Session, error)
	terminateSessionFn         func(ctx context.Context, userID uuid.UUID, sessionID uuid.UUID) error
	terminateAllExceptCurrentFn func(ctx context.Context, userID uuid.UUID, currentSessionID uuid.UUID) error
	validateSessionFn          func(ctx context.Context, sessionID uuid.UUID) (*domain.Session, error)
	validateAndTouchFn         func(ctx context.Context, sessionID uuid.UUID) error
	cleanExpiredFn             func(ctx context.Context) (int64, error)
}

func (m *mockSessionService) CreateSession(ctx context.Context, userID uuid.UUID, deviceInfo, browser, ip string) (*domain.Session, error) {
	if m.createSessionFn != nil {
		return m.createSessionFn(ctx, userID, deviceInfo, browser, ip)
	}
	return nil, nil
}

func (m *mockSessionService) ListSessions(ctx context.Context, userID uuid.UUID) ([]domain.Session, error) {
	if m.listSessionsFn != nil {
		return m.listSessionsFn(ctx, userID)
	}
	return nil, nil
}

func (m *mockSessionService) TerminateSession(ctx context.Context, userID uuid.UUID, sessionID uuid.UUID) error {
	if m.terminateSessionFn != nil {
		return m.terminateSessionFn(ctx, userID, sessionID)
	}
	return nil
}

func (m *mockSessionService) TerminateAllExceptCurrent(ctx context.Context, userID uuid.UUID, currentSessionID uuid.UUID) error {
	if m.terminateAllExceptCurrentFn != nil {
		return m.terminateAllExceptCurrentFn(ctx, userID, currentSessionID)
	}
	return nil
}

func (m *mockSessionService) TerminateAllSessions(_ context.Context, _ uuid.UUID) error {
	return nil
}

func (m *mockSessionService) ValidateSession(ctx context.Context, sessionID uuid.UUID) (*domain.Session, error) {
	if m.validateSessionFn != nil {
		return m.validateSessionFn(ctx, sessionID)
	}
	return nil, nil
}

func (m *mockSessionService) ValidateAndTouch(ctx context.Context, sessionID uuid.UUID) error {
	if m.validateAndTouchFn != nil {
		return m.validateAndTouchFn(ctx, sessionID)
	}
	return nil
}

func (m *mockSessionService) CleanExpired(ctx context.Context) (int64, error) {
	if m.cleanExpiredFn != nil {
		return m.cleanExpiredFn(ctx)
	}
	return 0, nil
}

func TestSessionHandler_ListSessions(t *testing.T) {
	userID := uuid.New()
	sessionID := uuid.New()
	now := time.Now()

	svc := &mockSessionService{
		listSessionsFn: func(_ context.Context, uid uuid.UUID) ([]domain.Session, error) {
			return []domain.Session{
				{
					ID:           sessionID,
					UserID:       uid,
					DeviceInfo:   "iPhone",
					Browser:      "Safari",
					IP:           "1.2.3.4",
					LastActiveAt: now,
					CreatedAt:    now,
					ExpiresAt:    now.Add(30 * 24 * time.Hour),
				},
			}, nil
		},
	}

	h := NewSessionHandler(svc)

	req := httptest.NewRequest(http.MethodGet, "/my/sessions", nil)
	ctx := middleware.SetUserID(req.Context(), userID)
	ctx = middleware.SetSessionID(ctx, sessionID)
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()

	h.ListSessions(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp APIResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if !resp.Success {
		t.Error("expected success=true")
	}

	data, ok := resp.Data.([]interface{})
	if !ok {
		t.Fatal("expected data to be an array")
	}
	if len(data) != 1 {
		t.Errorf("expected 1 session, got %d", len(data))
	}

	session := data[0].(map[string]interface{})
	if session["is_current"] != true {
		t.Error("expected is_current=true for current session")
	}
}

func TestSessionHandler_TerminateSession(t *testing.T) {
	userID := uuid.New()
	sessionID := uuid.New()

	var terminatedID uuid.UUID
	svc := &mockSessionService{
		terminateSessionFn: func(_ context.Context, _ uuid.UUID, sid uuid.UUID) error {
			terminatedID = sid
			return nil
		},
	}

	h := NewSessionHandler(svc)

	r := chi.NewRouter()
	r.Delete("/my/sessions/{id}", func(w http.ResponseWriter, req *http.Request) {
		ctx := middleware.SetUserID(req.Context(), userID)
		req = req.WithContext(ctx)
		h.TerminateSession(w, req)
	})

	req := httptest.NewRequest(http.MethodDelete, "/my/sessions/"+sessionID.String(), nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if terminatedID != sessionID {
		t.Errorf("expected session %s to be terminated, got %s", sessionID, terminatedID)
	}
}

func TestSessionHandler_TerminateSession_NotFound(t *testing.T) {
	userID := uuid.New()
	sessionID := uuid.New()

	svc := &mockSessionService{
		terminateSessionFn: func(_ context.Context, _ uuid.UUID, _ uuid.UUID) error {
			return domain.ErrSessionNotFound
		},
	}

	h := NewSessionHandler(svc)

	r := chi.NewRouter()
	r.Delete("/my/sessions/{id}", func(w http.ResponseWriter, req *http.Request) {
		ctx := middleware.SetUserID(req.Context(), userID)
		req = req.WithContext(ctx)
		h.TerminateSession(w, req)
	})

	req := httptest.NewRequest(http.MethodDelete, "/my/sessions/"+sessionID.String(), nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestSessionHandler_TerminateAllOtherSessions(t *testing.T) {
	userID := uuid.New()
	currentSessionID := uuid.New()

	var calledExceptID uuid.UUID
	svc := &mockSessionService{
		terminateAllExceptCurrentFn: func(_ context.Context, _ uuid.UUID, exceptID uuid.UUID) error {
			calledExceptID = exceptID
			return nil
		},
	}

	h := NewSessionHandler(svc)

	req := httptest.NewRequest(http.MethodDelete, "/my/sessions", nil)
	ctx := middleware.SetUserID(req.Context(), userID)
	ctx = middleware.SetSessionID(ctx, currentSessionID)
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()

	h.TerminateAllOtherSessions(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if calledExceptID != currentSessionID {
		t.Errorf("expected except ID %s, got %s", currentSessionID, calledExceptID)
	}
}

func TestSessionHandler_TerminateSession_InvalidID(t *testing.T) {
	userID := uuid.New()

	svc := &mockSessionService{}
	h := NewSessionHandler(svc)

	r := chi.NewRouter()
	r.Delete("/my/sessions/{id}", func(w http.ResponseWriter, req *http.Request) {
		ctx := middleware.SetUserID(req.Context(), userID)
		req = req.WithContext(ctx)
		h.TerminateSession(w, req)
	})

	req := httptest.NewRequest(http.MethodDelete, "/my/sessions/not-a-uuid", nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
}
