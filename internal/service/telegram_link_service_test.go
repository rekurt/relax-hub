package service

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/repository/mock"
)

func TestTelegramLinkService_LinkAccount(t *testing.T) {
	repo := mock.NewTelegramLinkRepo()
	service := NewTelegramLinkService(repo)
	ctx := context.Background()

	userID := uuid.New()
	link, err := service.LinkAccount(ctx, userID, 123456789, "testuser")
	if err != nil {
		t.Fatalf("LinkAccount failed: %v", err)
	}

	if link.UserID != userID {
		t.Fatalf("Expected UserID %v, got %v", userID, link.UserID)
	}
	if link.TelegramID != 123456789 {
		t.Fatalf("Expected TelegramID 123456789, got %v", link.TelegramID)
	}
	if link.TelegramUsername != "testuser" {
		t.Fatalf("Expected TelegramUsername testuser, got %v", link.TelegramUsername)
	}
}

func TestTelegramLinkService_LinkAccountDuplicate(t *testing.T) {
	repo := mock.NewTelegramLinkRepo()
	service := NewTelegramLinkService(repo)
	ctx := context.Background()

	userID := uuid.New()
	_, err := service.LinkAccount(ctx, userID, 123456789, "testuser")
	if err != nil {
		t.Fatalf("First LinkAccount failed: %v", err)
	}

	// Try to link again
	_, err = service.LinkAccount(ctx, userID, 123456789, "testuser")
	if err != domain.ErrAlreadyExists {
		t.Fatalf("Expected ErrAlreadyExists, got %v", err)
	}
}

func TestTelegramLinkService_LinkAccountInvalidUserID(t *testing.T) {
	repo := mock.NewTelegramLinkRepo()
	service := NewTelegramLinkService(repo)
	ctx := context.Background()

	_, err := service.LinkAccount(ctx, uuid.Nil, 123456789, "testuser")
	if err == nil {
		t.Fatal("Expected error for nil UserID")
	}
}

func TestTelegramLinkService_GetByTelegramID(t *testing.T) {
	repo := mock.NewTelegramLinkRepo()
	service := NewTelegramLinkService(repo)
	ctx := context.Background()

	userID := uuid.New()
	link, _ := service.LinkAccount(ctx, userID, 123456789, "testuser")

	retrieved, err := service.GetByTelegramID(ctx, 123456789)
	if err != nil {
		t.Fatalf("GetByTelegramID failed: %v", err)
	}

	if retrieved.ID != link.ID {
		t.Fatalf("Expected ID %v, got %v", link.ID, retrieved.ID)
	}
}

func TestTelegramLinkService_GetByUserID(t *testing.T) {
	repo := mock.NewTelegramLinkRepo()
	service := NewTelegramLinkService(repo)
	ctx := context.Background()

	userID := uuid.New()
	link, _ := service.LinkAccount(ctx, userID, 123456789, "testuser")

	retrieved, err := service.GetByUserID(ctx, userID)
	if err != nil {
		t.Fatalf("GetByUserID failed: %v", err)
	}

	if retrieved.ID != link.ID {
		t.Fatalf("Expected ID %v, got %v", link.ID, retrieved.ID)
	}
}

func TestTelegramLinkService_UnlinkAccount(t *testing.T) {
	repo := mock.NewTelegramLinkRepo()
	service := NewTelegramLinkService(repo)
	ctx := context.Background()

	userID := uuid.New()
	_, _ = service.LinkAccount(ctx, userID, 123456789, "testuser")

	err := service.UnlinkAccount(ctx, userID)
	if err != nil {
		t.Fatalf("UnlinkAccount failed: %v", err)
	}

	_, err = service.GetByUserID(ctx, userID)
	if err != domain.ErrNotFound {
		t.Fatalf("Expected ErrNotFound after unlink, got %v", err)
	}
}
