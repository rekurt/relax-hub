package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/rekurt/relax-hub/internal/domain"
	"github.com/rekurt/relax-hub/internal/logger"
	"github.com/rekurt/relax-hub/internal/repository/mock"
	"github.com/rekurt/relax-hub/internal/service"
)

func newSessionService() service.SessionService {
	repo := mock.NewSessionRepo()
	log := logger.New(logger.LevelWarn)
	return service.NewSessionService(repo, log)
}

func TestSessionService_CreateSession(t *testing.T) {
	svc := newSessionService()
	userID := uuid.New()

	session, err := svc.CreateSession(context.Background(), userID, "iPhone 15", "Safari", "192.168.1.1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if session.ID == uuid.Nil {
		t.Error("expected non-nil session ID")
	}
	if session.UserID != userID {
		t.Errorf("expected user ID %s, got %s", userID, session.UserID)
	}
	if session.DeviceInfo != "iPhone 15" {
		t.Errorf("expected device info 'iPhone 15', got %q", session.DeviceInfo)
	}
	if session.Browser != "Safari" {
		t.Errorf("expected browser 'Safari', got %q", session.Browser)
	}
	if session.IP != "192.168.1.1" {
		t.Errorf("expected IP '192.168.1.1', got %q", session.IP)
	}
	if session.ExpiresAt.Before(time.Now().Add(29 * 24 * time.Hour)) {
		t.Error("expected expiry ~30 days in the future")
	}
}

func TestSessionService_ListSessions(t *testing.T) {
	svc := newSessionService()
	userID := uuid.New()
	otherUserID := uuid.New()

	_, _ = svc.CreateSession(context.Background(), userID, "Device 1", "Chrome", "1.1.1.1")
	_, _ = svc.CreateSession(context.Background(), userID, "Device 2", "Firefox", "2.2.2.2")
	_, _ = svc.CreateSession(context.Background(), otherUserID, "Device 3", "Safari", "3.3.3.3")

	sessions, err := svc.ListSessions(context.Background(), userID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(sessions) != 2 {
		t.Errorf("expected 2 sessions, got %d", len(sessions))
	}
}

func TestSessionService_TerminateSession(t *testing.T) {
	svc := newSessionService()
	userID := uuid.New()

	session, _ := svc.CreateSession(context.Background(), userID, "Device", "Chrome", "1.1.1.1")

	err := svc.TerminateSession(context.Background(), userID, session.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	sessions, _ := svc.ListSessions(context.Background(), userID)
	if len(sessions) != 0 {
		t.Errorf("expected 0 sessions after termination, got %d", len(sessions))
	}
}

func TestSessionService_TerminateSession_WrongUser(t *testing.T) {
	svc := newSessionService()
	userID := uuid.New()
	otherUserID := uuid.New()

	session, _ := svc.CreateSession(context.Background(), userID, "Device", "Chrome", "1.1.1.1")

	err := svc.TerminateSession(context.Background(), otherUserID, session.ID)
	if err != domain.ErrForbidden {
		t.Errorf("expected ErrForbidden, got %v", err)
	}
}

func TestSessionService_TerminateAllExceptCurrent(t *testing.T) {
	svc := newSessionService()
	userID := uuid.New()

	s1, _ := svc.CreateSession(context.Background(), userID, "Device 1", "Chrome", "1.1.1.1")
	_, _ = svc.CreateSession(context.Background(), userID, "Device 2", "Firefox", "2.2.2.2")
	_, _ = svc.CreateSession(context.Background(), userID, "Device 3", "Safari", "3.3.3.3")

	err := svc.TerminateAllExceptCurrent(context.Background(), userID, s1.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	sessions, _ := svc.ListSessions(context.Background(), userID)
	if len(sessions) != 1 {
		t.Errorf("expected 1 session, got %d", len(sessions))
	}
	if sessions[0].ID != s1.ID {
		t.Error("expected the current session to remain")
	}
}

func TestSessionService_ValidateAndTouch(t *testing.T) {
	svc := newSessionService()
	userID := uuid.New()

	session, _ := svc.CreateSession(context.Background(), userID, "Device", "Chrome", "1.1.1.1")

	err := svc.ValidateAndTouch(context.Background(), session.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSessionService_ValidateAndTouch_NotFound(t *testing.T) {
	svc := newSessionService()

	err := svc.ValidateAndTouch(context.Background(), uuid.New())
	if err != domain.ErrSessionExpired {
		t.Errorf("expected ErrSessionExpired, got %v", err)
	}
}

func TestSessionService_ValidateSession(t *testing.T) {
	svc := newSessionService()
	userID := uuid.New()

	session, _ := svc.CreateSession(context.Background(), userID, "Device", "Chrome", "1.1.1.1")

	validated, err := svc.ValidateSession(context.Background(), session.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if validated.ID != session.ID {
		t.Error("expected same session ID")
	}
}

func TestSessionService_CleanExpired(t *testing.T) {
	repo := mock.NewSessionRepo()
	log := logger.New(logger.LevelWarn)
	svc := service.NewSessionService(repo, log)
	ctx := context.Background()

	// Create an expired session directly via repo
	expiredSession := &domain.Session{
		ID:           uuid.New(),
		UserID:       uuid.New(),
		DeviceInfo:   "old device",
		Browser:      "old browser",
		IP:           "0.0.0.0",
		LastActiveAt: time.Now().Add(-60 * 24 * time.Hour),
		CreatedAt:    time.Now().Add(-60 * 24 * time.Hour),
		ExpiresAt:    time.Now().Add(-1 * time.Hour), // already expired
	}
	_ = repo.Create(ctx, expiredSession)

	// Create a valid session
	_, _ = svc.CreateSession(ctx, uuid.New(), "new device", "Chrome", "1.1.1.1")

	count, err := svc.CleanExpired(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 cleaned session, got %d", count)
	}
}
