package mock

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/repository"
)

var (
	_ repository.ConversationRepository = (*ConversationRepo)(nil)
	_ repository.MessageRepository      = (*MessageRepo)(nil)
)

func TestConversationRepo_CRUD(t *testing.T) {
	ctx := context.Background()
	repo := NewConversationRepo()

	bathhouseID := uuid.New()
	clientID := uuid.New()

	conv := &domain.Conversation{
		BathhouseID: bathhouseID,
		ClientID:    clientID,
	}

	// Create
	if err := repo.Create(ctx, conv); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if conv.ID == uuid.Nil {
		t.Fatal("ID should be assigned after Create")
	}

	// GetByID
	got, err := repo.GetByID(ctx, conv.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if got.BathhouseID != bathhouseID {
		t.Errorf("BathhouseID = %v, want %v", got.BathhouseID, bathhouseID)
	}

	// GetByParticipants
	got, err = repo.GetByParticipants(ctx, bathhouseID, clientID)
	if err != nil {
		t.Fatalf("GetByParticipants: %v", err)
	}
	if got.ID != conv.ID {
		t.Errorf("ID = %v, want %v", got.ID, conv.ID)
	}

	// GetByParticipants - not found
	_, err = repo.GetByParticipants(ctx, uuid.New(), uuid.New())
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("GetByParticipants nonexistent: want ErrNotFound, got %v", err)
	}

	// Duplicate unique constraint
	dup := &domain.Conversation{BathhouseID: bathhouseID, ClientID: clientID}
	if err := repo.Create(ctx, dup); !errors.Is(err, domain.ErrAlreadyExists) {
		t.Errorf("Create duplicate: want ErrAlreadyExists, got %v", err)
	}

	// GetOrCreate - existing
	got, err = repo.GetOrCreate(ctx, &domain.Conversation{BathhouseID: bathhouseID, ClientID: clientID})
	if err != nil {
		t.Fatalf("GetOrCreate existing: %v", err)
	}
	if got.ID != conv.ID {
		t.Errorf("GetOrCreate should return existing, got ID %v, want %v", got.ID, conv.ID)
	}

	// GetOrCreate - new
	newClientID := uuid.New()
	got, err = repo.GetOrCreate(ctx, &domain.Conversation{BathhouseID: bathhouseID, ClientID: newClientID})
	if err != nil {
		t.Fatalf("GetOrCreate new: %v", err)
	}
	if got.ClientID != newClientID {
		t.Errorf("GetOrCreate new: ClientID = %v, want %v", got.ClientID, newClientID)
	}

	// UpdateLastMessageAt
	now := time.Now()
	if err := repo.UpdateLastMessageAt(ctx, conv.ID, now); err != nil {
		t.Fatalf("UpdateLastMessageAt: %v", err)
	}
	got, _ = repo.GetByID(ctx, conv.ID)
	if got.LastMessageAt == nil {
		t.Fatal("LastMessageAt should be set")
	}

	// ListByUser - as client
	result, err := repo.ListByUser(ctx, clientID, nil, 1, 10)
	if err != nil {
		t.Fatalf("ListByUser client: %v", err)
	}
	if result.TotalCount != 1 {
		t.Errorf("ListByUser client: TotalCount = %d, want 1", result.TotalCount)
	}

	// ListByUser - as owner (by bathhouseIDs)
	result, err = repo.ListByUser(ctx, uuid.New(), []uuid.UUID{bathhouseID}, 1, 10)
	if err != nil {
		t.Fatalf("ListByUser owner: %v", err)
	}
	if result.TotalCount != 2 {
		t.Errorf("ListByUser owner: TotalCount = %d, want 2", result.TotalCount)
	}

	// GetByID - not found
	_, err = repo.GetByID(ctx, uuid.New())
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("GetByID nonexistent: want ErrNotFound, got %v", err)
	}
}

func TestMessageRepo_CRUD(t *testing.T) {
	ctx := context.Background()
	repo := NewMessageRepo()

	convID := uuid.New()
	senderID := uuid.New()
	otherID := uuid.New()

	// Create messages
	msg1 := &domain.Message{ConversationID: convID, SenderID: senderID, Text: "Hello"}
	if err := repo.Create(ctx, msg1); err != nil {
		t.Fatalf("Create msg1: %v", err)
	}
	if msg1.ID == uuid.Nil {
		t.Fatal("ID should be assigned after Create")
	}

	msg2 := &domain.Message{ConversationID: convID, SenderID: otherID, Text: "Hi there"}
	if err := repo.Create(ctx, msg2); err != nil {
		t.Fatalf("Create msg2: %v", err)
	}

	// ListByConversation
	result, err := repo.ListByConversation(ctx, convID, 1, 10)
	if err != nil {
		t.Fatalf("ListByConversation: %v", err)
	}
	if result.TotalCount != 2 {
		t.Errorf("TotalCount = %d, want 2", result.TotalCount)
	}

	// ListByConversation - empty
	result, err = repo.ListByConversation(ctx, uuid.New(), 1, 10)
	if err != nil {
		t.Fatalf("ListByConversation empty: %v", err)
	}
	if result.TotalCount != 0 {
		t.Errorf("TotalCount = %d, want 0", result.TotalCount)
	}

	// CountUnread - senderID sees unread from otherID
	count, err := repo.CountUnread(ctx, senderID, []uuid.UUID{convID})
	if err != nil {
		t.Fatalf("CountUnread: %v", err)
	}
	if count != 1 {
		t.Errorf("CountUnread = %d, want 1", count)
	}

	// CountUnread - empty conversationIDs
	count, err = repo.CountUnread(ctx, senderID, nil)
	if err != nil {
		t.Fatalf("CountUnread nil: %v", err)
	}
	if count != 0 {
		t.Errorf("CountUnread nil = %d, want 0", count)
	}

	// MarkAsRead
	if err := repo.MarkAsRead(ctx, convID, senderID); err != nil {
		t.Fatalf("MarkAsRead: %v", err)
	}

	// After marking as read, unread count should be 0
	count, err = repo.CountUnread(ctx, senderID, []uuid.UUID{convID})
	if err != nil {
		t.Fatalf("CountUnread after mark: %v", err)
	}
	if count != 0 {
		t.Errorf("CountUnread after mark = %d, want 0", count)
	}
}
