package mock

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
)

func TestTelegramLinkRepo_Create(t *testing.T) {
	repo := NewTelegramLinkRepo()
	ctx := context.Background()

	userID := uuid.New()
	link := &domain.TelegramLink{
		UserID:           userID,
		TelegramID:       123456789,
		TelegramUsername: "testuser",
	}

	err := repo.Create(ctx, link)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if link.ID == uuid.Nil {
		t.Fatal("ID was not set")
	}
}

func TestTelegramLinkRepo_CreateDuplicate(t *testing.T) {
	repo := NewTelegramLinkRepo()
	ctx := context.Background()

	userID := uuid.New()
	link := &domain.TelegramLink{
		UserID:           userID,
		TelegramID:       123456789,
		TelegramUsername: "testuser",
	}

	err := repo.Create(ctx, link)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Try to create duplicate
	err = repo.Create(ctx, link)
	if err != domain.ErrAlreadyExists {
		t.Fatalf("Expected ErrAlreadyExists, got %v", err)
	}
}

func TestTelegramLinkRepo_GetByTelegramID(t *testing.T) {
	repo := NewTelegramLinkRepo()
	ctx := context.Background()

	userID := uuid.New()
	link := &domain.TelegramLink{
		UserID:           userID,
		TelegramID:       123456789,
		TelegramUsername: "testuser",
	}

	if err := repo.Create(ctx, link); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	retrieved, err := repo.GetByTelegramID(ctx, 123456789)
	if err != nil {
		t.Fatalf("GetByTelegramID failed: %v", err)
	}

	if retrieved.UserID != userID {
		t.Fatalf("Expected UserID %v, got %v", userID, retrieved.UserID)
	}
	if retrieved.TelegramID != 123456789 {
		t.Fatalf("Expected TelegramID 123456789, got %v", retrieved.TelegramID)
	}
}

func TestTelegramLinkRepo_GetByUserID(t *testing.T) {
	repo := NewTelegramLinkRepo()
	ctx := context.Background()

	userID := uuid.New()
	link := &domain.TelegramLink{
		UserID:           userID,
		TelegramID:       123456789,
		TelegramUsername: "testuser",
	}

	if err := repo.Create(ctx, link); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	retrieved, err := repo.GetByUserID(ctx, userID)
	if err != nil {
		t.Fatalf("GetByUserID failed: %v", err)
	}

	if retrieved.TelegramID != 123456789 {
		t.Fatalf("Expected TelegramID 123456789, got %v", retrieved.TelegramID)
	}
}

func TestTelegramLinkRepo_Delete(t *testing.T) {
	repo := NewTelegramLinkRepo()
	ctx := context.Background()

	userID := uuid.New()
	link := &domain.TelegramLink{
		UserID:           userID,
		TelegramID:       123456789,
		TelegramUsername: "testuser",
	}

	if err := repo.Create(ctx, link); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if err := repo.Delete(ctx, userID); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	_, err := repo.GetByUserID(ctx, userID)
	if err != domain.ErrNotFound {
		t.Fatalf("Expected ErrNotFound after delete, got %v", err)
	}
}

func TestTelegramLinkRepo_DeleteNotFound(t *testing.T) {
	repo := NewTelegramLinkRepo()
	ctx := context.Background()

	err := repo.Delete(ctx, uuid.New())
	if err != domain.ErrNotFound {
		t.Fatalf("Expected ErrNotFound, got %v", err)
	}
}
