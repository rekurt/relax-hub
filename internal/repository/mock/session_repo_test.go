package mock_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/rekurt/relax-hub/internal/domain"
	"github.com/rekurt/relax-hub/internal/repository/mock"
)

func TestSessionRepo_CreateAndGet(t *testing.T) {
	repo := mock.NewSessionRepo()
	ctx := context.Background()
	userID := uuid.New()

	session := &domain.Session{
		UserID:     userID,
		DeviceInfo: "Test Device",
		Browser:    "Chrome",
		IP:         "127.0.0.1",
	}

	if err := repo.Create(ctx, session); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if session.ID == uuid.Nil {
		t.Error("expected non-nil session ID after create")
	}

	got, err := repo.GetByID(ctx, session.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.UserID != userID {
		t.Errorf("expected user ID %s, got %s", userID, got.UserID)
	}
	if got.DeviceInfo != "Test Device" {
		t.Errorf("expected device info 'Test Device', got %q", got.DeviceInfo)
	}
}

func TestSessionRepo_GetByID_NotFound(t *testing.T) {
	repo := mock.NewSessionRepo()

	_, err := repo.GetByID(context.Background(), uuid.New())
	if err != domain.ErrSessionNotFound {
		t.Errorf("expected ErrSessionNotFound, got %v", err)
	}
}

func TestSessionRepo_ListByUser(t *testing.T) {
	repo := mock.NewSessionRepo()
	ctx := context.Background()
	userID := uuid.New()

	_ = repo.Create(ctx, &domain.Session{UserID: userID, DeviceInfo: "D1"})
	_ = repo.Create(ctx, &domain.Session{UserID: userID, DeviceInfo: "D2"})
	_ = repo.Create(ctx, &domain.Session{UserID: uuid.New(), DeviceInfo: "D3"})

	sessions, err := repo.ListByUser(ctx, userID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(sessions) != 2 {
		t.Errorf("expected 2, got %d", len(sessions))
	}
}

func TestSessionRepo_Delete(t *testing.T) {
	repo := mock.NewSessionRepo()
	ctx := context.Background()

	session := &domain.Session{UserID: uuid.New()}
	_ = repo.Create(ctx, session)

	if err := repo.Delete(ctx, session.ID); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err := repo.GetByID(ctx, session.ID)
	if err != domain.ErrSessionNotFound {
		t.Errorf("expected ErrSessionNotFound after delete")
	}
}

func TestSessionRepo_Delete_NotFound(t *testing.T) {
	repo := mock.NewSessionRepo()
	err := repo.Delete(context.Background(), uuid.New())
	if err != domain.ErrSessionNotFound {
		t.Errorf("expected ErrSessionNotFound, got %v", err)
	}
}

func TestSessionRepo_DeleteAllExcept(t *testing.T) {
	repo := mock.NewSessionRepo()
	ctx := context.Background()
	userID := uuid.New()

	s1 := &domain.Session{UserID: userID}
	s2 := &domain.Session{UserID: userID}
	_ = repo.Create(ctx, s1)
	_ = repo.Create(ctx, s2)

	if err := repo.DeleteAllExcept(ctx, userID, s1.ID); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	sessions, _ := repo.ListByUser(ctx, userID)
	if len(sessions) != 1 {
		t.Errorf("expected 1 session, got %d", len(sessions))
	}
	if sessions[0].ID != s1.ID {
		t.Error("expected kept session to be s1")
	}
}

func TestSessionRepo_UpdateLastActive(t *testing.T) {
	repo := mock.NewSessionRepo()
	ctx := context.Background()

	session := &domain.Session{UserID: uuid.New()}
	_ = repo.Create(ctx, session)

	newTime := time.Now().Add(1 * time.Hour)
	if err := repo.UpdateLastActive(ctx, session.ID, newTime); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, _ := repo.GetByID(ctx, session.ID)
	if !got.LastActiveAt.Equal(newTime) {
		t.Error("expected updated last_active_at")
	}
}

func TestSessionRepo_DeleteExpired(t *testing.T) {
	repo := mock.NewSessionRepo()
	ctx := context.Background()

	// Create expired session
	expired := &domain.Session{
		ID:        uuid.New(),
		UserID:    uuid.New(),
		ExpiresAt: time.Now().Add(-1 * time.Hour),
		CreatedAt: time.Now().Add(-48 * time.Hour),
	}
	_ = repo.Create(ctx, expired)

	// Create valid session
	valid := &domain.Session{UserID: uuid.New()}
	_ = repo.Create(ctx, valid)

	count, err := repo.DeleteExpired(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 deleted, got %d", count)
	}
}
