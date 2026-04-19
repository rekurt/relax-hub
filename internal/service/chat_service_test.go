package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/rekurt/relax-hub/internal/antifraud"
	"github.com/rekurt/relax-hub/internal/domain"
	"github.com/rekurt/relax-hub/internal/logger"
	"github.com/rekurt/relax-hub/internal/notification"
	"github.com/rekurt/relax-hub/internal/repository/mock"
	"github.com/rekurt/relax-hub/internal/service"
)

type chatTestEnv struct {
	svc      service.ChatService
	convRepo *mock.ConversationRepo
	msgRepo  *mock.MessageRepo
	bhRepo   *mock.BathhouseRepo
	repRepo  *mock.RepresentativeRepo
}

func newChatTestEnv() *chatTestEnv {
	convRepo := mock.NewConversationRepo()
	msgRepo := mock.NewMessageRepo(convRepo)
	bhRepo := mock.NewBathhouseRepo()
	repRepo := mock.NewRepresentativeRepo()
	ac := service.NewAccessChecker(repRepo, bhRepo)
	log := logger.New(logger.LevelWarn)
	hub := notification.NewHub(log)
	chatFilter := antifraud.NewChatFilter(log)
	svc := service.NewChatService(convRepo, msgRepo, bhRepo, repRepo, ac, &noopNotifService{}, hub, chatFilter, log)
	return &chatTestEnv{
		svc:      svc,
		convRepo: convRepo,
		msgRepo:  msgRepo,
		bhRepo:   bhRepo,
		repRepo:  repRepo,
	}
}

func TestChatService_StartConversation_Success(t *testing.T) {
	env := newChatTestEnv()
	ownerID := uuid.New()
	clientID := uuid.New()
	bh := createBathhouse(t, env.bhRepo, ownerID)

	conv, err := env.svc.StartConversation(context.Background(), clientID, bh.ID, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if conv.BathhouseID != bh.ID {
		t.Errorf("bathhouse_id = %v, want %v", conv.BathhouseID, bh.ID)
	}
	if conv.ClientID != clientID {
		t.Errorf("client_id = %v, want %v", conv.ClientID, clientID)
	}
}

func TestChatService_StartConversation_WithBooking(t *testing.T) {
	env := newChatTestEnv()
	ownerID := uuid.New()
	clientID := uuid.New()
	bh := createBathhouse(t, env.bhRepo, ownerID)
	bookingID := uuid.New()

	conv, err := env.svc.StartConversation(context.Background(), clientID, bh.ID, &bookingID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if conv.BookingID == nil || *conv.BookingID != bookingID {
		t.Errorf("booking_id = %v, want %v", conv.BookingID, bookingID)
	}
}

func TestChatService_StartConversation_ExistingReturned(t *testing.T) {
	env := newChatTestEnv()
	ownerID := uuid.New()
	clientID := uuid.New()
	bh := createBathhouse(t, env.bhRepo, ownerID)

	conv1, err := env.svc.StartConversation(context.Background(), clientID, bh.ID, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	conv2, err := env.svc.StartConversation(context.Background(), clientID, bh.ID, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if conv1.ID != conv2.ID {
		t.Errorf("expected same conversation, got different IDs: %v vs %v", conv1.ID, conv2.ID)
	}
}

func TestChatService_StartConversation_BathhouseNotFound(t *testing.T) {
	env := newChatTestEnv()
	clientID := uuid.New()

	_, err := env.svc.StartConversation(context.Background(), clientID, uuid.New(), nil)
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got: %v", err)
	}
}

func TestChatService_SendMessage_AsClient(t *testing.T) {
	env := newChatTestEnv()
	ownerID := uuid.New()
	clientID := uuid.New()
	bh := createBathhouse(t, env.bhRepo, ownerID)

	conv, _ := env.svc.StartConversation(context.Background(), clientID, bh.ID, nil)

	msg, err := env.svc.SendMessage(context.Background(), clientID, domain.RoleClient, conv.ID, "Hello!")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if msg.Text != "Hello!" {
		t.Errorf("text = %q, want %q", msg.Text, "Hello!")
	}
	if msg.SenderID != clientID {
		t.Errorf("sender_id = %v, want %v", msg.SenderID, clientID)
	}
}

func TestChatService_SendMessage_AsOwner(t *testing.T) {
	env := newChatTestEnv()
	ownerID := uuid.New()
	clientID := uuid.New()
	bh := createBathhouse(t, env.bhRepo, ownerID)

	conv, _ := env.svc.StartConversation(context.Background(), clientID, bh.ID, nil)

	msg, err := env.svc.SendMessage(context.Background(), ownerID, domain.RoleOwner, conv.ID, "Welcome!")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if msg.SenderID != ownerID {
		t.Errorf("sender_id = %v, want %v", msg.SenderID, ownerID)
	}
}

func TestChatService_SendMessage_AsRepresentative(t *testing.T) {
	env := newChatTestEnv()
	ownerID := uuid.New()
	repUserID := uuid.New()
	clientID := uuid.New()
	bh := createBathhouse(t, env.bhRepo, ownerID)

	rep := &domain.Representative{
		ID:          uuid.New(),
		UserID:      repUserID,
		BathhouseID: bh.ID,
		OwnerID:     ownerID,
		Role:        domain.RepRoleManager,
		CreatedAt:   time.Now(),
	}
	if err := env.repRepo.Create(context.Background(), rep); err != nil {
		t.Fatalf("create rep: %v", err)
	}

	conv, _ := env.svc.StartConversation(context.Background(), clientID, bh.ID, nil)

	msg, err := env.svc.SendMessage(context.Background(), repUserID, domain.RoleRepresentative, conv.ID, "Hi from rep!")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if msg.SenderID != repUserID {
		t.Errorf("sender_id = %v, want %v", msg.SenderID, repUserID)
	}
}

func TestChatService_SendMessage_Forbidden(t *testing.T) {
	env := newChatTestEnv()
	ownerID := uuid.New()
	clientID := uuid.New()
	strangerID := uuid.New()
	bh := createBathhouse(t, env.bhRepo, ownerID)

	conv, _ := env.svc.StartConversation(context.Background(), clientID, bh.ID, nil)

	_, err := env.svc.SendMessage(context.Background(), strangerID, domain.RoleClient, conv.ID, "Intruder!")
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("expected ErrForbidden, got: %v", err)
	}
}

func TestChatService_SendMessage_EmptyText(t *testing.T) {
	env := newChatTestEnv()
	ownerID := uuid.New()
	clientID := uuid.New()
	bh := createBathhouse(t, env.bhRepo, ownerID)

	conv, _ := env.svc.StartConversation(context.Background(), clientID, bh.ID, nil)

	_, err := env.svc.SendMessage(context.Background(), clientID, domain.RoleClient, conv.ID, "")
	if !errors.Is(err, domain.ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got: %v", err)
	}
}

func TestChatService_SendMessage_ConversationNotFound(t *testing.T) {
	env := newChatTestEnv()

	_, err := env.svc.SendMessage(context.Background(), uuid.New(), domain.RoleClient, uuid.New(), "Hello")
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got: %v", err)
	}
}

func TestChatService_ListConversations_Client(t *testing.T) {
	env := newChatTestEnv()
	ownerID := uuid.New()
	clientID := uuid.New()
	bh1 := createBathhouse(t, env.bhRepo, ownerID)
	bh2 := createBathhouse(t, env.bhRepo, ownerID)

	_, _ = env.svc.StartConversation(context.Background(), clientID, bh1.ID, nil)
	_, _ = env.svc.StartConversation(context.Background(), clientID, bh2.ID, nil)

	result, err := env.svc.ListConversations(context.Background(), clientID, domain.RoleClient, 1, 20)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.TotalCount != 2 {
		t.Errorf("total_count = %d, want 2", result.TotalCount)
	}
}

func TestChatService_ListConversations_Owner(t *testing.T) {
	env := newChatTestEnv()
	ownerID := uuid.New()
	client1 := uuid.New()
	client2 := uuid.New()
	bh := createBathhouse(t, env.bhRepo, ownerID)

	_, _ = env.svc.StartConversation(context.Background(), client1, bh.ID, nil)
	_, _ = env.svc.StartConversation(context.Background(), client2, bh.ID, nil)

	result, err := env.svc.ListConversations(context.Background(), ownerID, domain.RoleOwner, 1, 20)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.TotalCount != 2 {
		t.Errorf("total_count = %d, want 2", result.TotalCount)
	}
}

func TestChatService_ListConversations_Representative(t *testing.T) {
	env := newChatTestEnv()
	ownerID := uuid.New()
	repUserID := uuid.New()
	clientID := uuid.New()
	bh := createBathhouse(t, env.bhRepo, ownerID)

	rep := &domain.Representative{
		ID:          uuid.New(),
		UserID:      repUserID,
		BathhouseID: bh.ID,
		OwnerID:     ownerID,
		Role:        domain.RepRoleManager,
		CreatedAt:   time.Now(),
	}
	_ = env.repRepo.Create(context.Background(), rep)

	_, _ = env.svc.StartConversation(context.Background(), clientID, bh.ID, nil)

	result, err := env.svc.ListConversations(context.Background(), repUserID, domain.RoleRepresentative, 1, 20)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.TotalCount != 1 {
		t.Errorf("total_count = %d, want 1", result.TotalCount)
	}
}

func TestChatService_ListMessages_Success(t *testing.T) {
	env := newChatTestEnv()
	ownerID := uuid.New()
	clientID := uuid.New()
	bh := createBathhouse(t, env.bhRepo, ownerID)

	conv, _ := env.svc.StartConversation(context.Background(), clientID, bh.ID, nil)
	_, _ = env.svc.SendMessage(context.Background(), clientID, domain.RoleClient, conv.ID, "msg1")
	_, _ = env.svc.SendMessage(context.Background(), clientID, domain.RoleClient, conv.ID, "msg2")

	result, err := env.svc.ListMessages(context.Background(), clientID, domain.RoleClient, conv.ID, 1, 20)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.TotalCount != 2 {
		t.Errorf("total_count = %d, want 2", result.TotalCount)
	}
}

func TestChatService_ListMessages_Forbidden(t *testing.T) {
	env := newChatTestEnv()
	ownerID := uuid.New()
	clientID := uuid.New()
	strangerID := uuid.New()
	bh := createBathhouse(t, env.bhRepo, ownerID)

	conv, _ := env.svc.StartConversation(context.Background(), clientID, bh.ID, nil)

	_, err := env.svc.ListMessages(context.Background(), strangerID, domain.RoleClient, conv.ID, 1, 20)
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("expected ErrForbidden, got: %v", err)
	}
}

func TestChatService_MarkAsRead_Success(t *testing.T) {
	env := newChatTestEnv()
	ownerID := uuid.New()
	clientID := uuid.New()
	bh := createBathhouse(t, env.bhRepo, ownerID)

	conv, _ := env.svc.StartConversation(context.Background(), clientID, bh.ID, nil)
	_, _ = env.svc.SendMessage(context.Background(), ownerID, domain.RoleOwner, conv.ID, "Hey client!")

	err := env.svc.MarkAsRead(context.Background(), clientID, domain.RoleClient, conv.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	count, err := env.svc.GetUnreadCount(context.Background(), clientID, domain.RoleClient)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 0 {
		t.Errorf("unread count = %d, want 0", count)
	}
}

func TestChatService_MarkAsRead_Forbidden(t *testing.T) {
	env := newChatTestEnv()
	ownerID := uuid.New()
	clientID := uuid.New()
	strangerID := uuid.New()
	bh := createBathhouse(t, env.bhRepo, ownerID)

	conv, _ := env.svc.StartConversation(context.Background(), clientID, bh.ID, nil)

	err := env.svc.MarkAsRead(context.Background(), strangerID, domain.RoleClient, conv.ID)
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("expected ErrForbidden, got: %v", err)
	}
}

func TestChatService_GetUnreadCount(t *testing.T) {
	env := newChatTestEnv()
	ownerID := uuid.New()
	clientID := uuid.New()
	bh := createBathhouse(t, env.bhRepo, ownerID)

	conv, _ := env.svc.StartConversation(context.Background(), clientID, bh.ID, nil)
	_, _ = env.svc.SendMessage(context.Background(), ownerID, domain.RoleOwner, conv.ID, "msg1")
	_, _ = env.svc.SendMessage(context.Background(), ownerID, domain.RoleOwner, conv.ID, "msg2")

	count, err := env.svc.GetUnreadCount(context.Background(), clientID, domain.RoleClient)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 2 {
		t.Errorf("unread count = %d, want 2", count)
	}
}

func TestChatService_GetUnreadCount_NoConversations(t *testing.T) {
	env := newChatTestEnv()
	clientID := uuid.New()

	count, err := env.svc.GetUnreadCount(context.Background(), clientID, domain.RoleClient)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 0 {
		t.Errorf("unread count = %d, want 0", count)
	}
}

func TestChatService_SendMessage_Admin(t *testing.T) {
	env := newChatTestEnv()
	ownerID := uuid.New()
	adminID := uuid.New()
	clientID := uuid.New()
	bh := createBathhouse(t, env.bhRepo, ownerID)

	conv, _ := env.svc.StartConversation(context.Background(), clientID, bh.ID, nil)

	msg, err := env.svc.SendMessage(context.Background(), adminID, domain.RoleAdmin, conv.ID, "Admin message")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if msg.SenderID != adminID {
		t.Errorf("sender_id = %v, want %v", msg.SenderID, adminID)
	}
}

func TestChatService_SendMessage_WhitespaceOnly(t *testing.T) {
	env := newChatTestEnv()
	ownerID := uuid.New()
	clientID := uuid.New()
	bh := createBathhouse(t, env.bhRepo, ownerID)

	conv, _ := env.svc.StartConversation(context.Background(), clientID, bh.ID, nil)

	_, err := env.svc.SendMessage(context.Background(), clientID, domain.RoleClient, conv.ID, "   \n\t  ")
	if !errors.Is(err, domain.ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput for whitespace-only text, got: %v", err)
	}
}

func TestChatService_ListConversations_Admin(t *testing.T) {
	env := newChatTestEnv()
	ownerID := uuid.New()
	adminID := uuid.New()
	client1 := uuid.New()
	client2 := uuid.New()
	bh := createBathhouse(t, env.bhRepo, ownerID)

	_, _ = env.svc.StartConversation(context.Background(), client1, bh.ID, nil)
	_, _ = env.svc.StartConversation(context.Background(), client2, bh.ID, nil)

	result, err := env.svc.ListConversations(context.Background(), adminID, domain.RoleAdmin, 1, 20)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.TotalCount != 2 {
		t.Errorf("total_count = %d, want 2", result.TotalCount)
	}
}

func TestChatService_GetUnreadCount_Admin(t *testing.T) {
	env := newChatTestEnv()
	adminID := uuid.New()

	count, err := env.svc.GetUnreadCount(context.Background(), adminID, domain.RoleAdmin)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 0 {
		t.Errorf("unread count = %d, want 0 for admin", count)
	}
}

func TestChatService_CanAccessConversation_AsClient(t *testing.T) {
	env := newChatTestEnv()
	ownerID := uuid.New()
	clientID := uuid.New()
	bh := createBathhouse(t, env.bhRepo, ownerID)

	conv, _ := env.svc.StartConversation(context.Background(), clientID, bh.ID, nil)

	if !env.svc.CanAccessConversation(context.Background(), clientID, domain.RoleClient, conv.ID) {
		t.Error("client should have access to own conversation")
	}
}

func TestChatService_CanAccessConversation_AsOwner(t *testing.T) {
	env := newChatTestEnv()
	ownerID := uuid.New()
	clientID := uuid.New()
	bh := createBathhouse(t, env.bhRepo, ownerID)

	conv, _ := env.svc.StartConversation(context.Background(), clientID, bh.ID, nil)

	if !env.svc.CanAccessConversation(context.Background(), ownerID, domain.RoleOwner, conv.ID) {
		t.Error("owner should have access to bathhouse conversation")
	}
}

func TestChatService_CanAccessConversation_Forbidden(t *testing.T) {
	env := newChatTestEnv()
	ownerID := uuid.New()
	clientID := uuid.New()
	strangerID := uuid.New()
	bh := createBathhouse(t, env.bhRepo, ownerID)

	conv, _ := env.svc.StartConversation(context.Background(), clientID, bh.ID, nil)

	if env.svc.CanAccessConversation(context.Background(), strangerID, domain.RoleClient, conv.ID) {
		t.Error("stranger should not have access to conversation")
	}
}

func TestChatService_CanAccessConversation_NotFound(t *testing.T) {
	env := newChatTestEnv()

	if env.svc.CanAccessConversation(context.Background(), uuid.New(), domain.RoleClient, uuid.New()) {
		t.Error("should return false for non-existent conversation")
	}
}

func TestChatService_SendMessage_ContactInfoFiltered(t *testing.T) {
	env := newChatTestEnv()
	ownerID := uuid.New()
	clientID := uuid.New()
	bh := createBathhouse(t, env.bhRepo, ownerID)

	conv, _ := env.svc.StartConversation(context.Background(), clientID, bh.ID, nil)

	// Message with phone number should be filtered.
	msg, err := env.svc.SendMessage(context.Background(), clientID, domain.RoleClient, conv.ID, "Звоните +7 999 123 45 67")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if msg.Text == "Звоните +7 999 123 45 67" {
		t.Error("phone number should have been filtered from message text")
	}
	if msg.Text != "Звоните [контактные данные скрыты]" {
		t.Errorf("filtered text = %q, want %q", msg.Text, "Звоните [контактные данные скрыты]")
	}
}

func TestChatService_SendMessage_EmailFiltered(t *testing.T) {
	env := newChatTestEnv()
	ownerID := uuid.New()
	clientID := uuid.New()
	bh := createBathhouse(t, env.bhRepo, ownerID)

	conv, _ := env.svc.StartConversation(context.Background(), clientID, bh.ID, nil)

	msg, err := env.svc.SendMessage(context.Background(), clientID, domain.RoleClient, conv.ID, "Пишите на user@mail.ru")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if msg.Text == "Пишите на user@mail.ru" {
		t.Error("email should have been filtered from message text")
	}
	if msg.Text != "Пишите на [контактные данные скрыты]" {
		t.Errorf("filtered text = %q, want %q", msg.Text, "Пишите на [контактные данные скрыты]")
	}
}

func TestChatService_SendMessage_TelegramFiltered(t *testing.T) {
	env := newChatTestEnv()
	ownerID := uuid.New()
	clientID := uuid.New()
	bh := createBathhouse(t, env.bhRepo, ownerID)

	conv, _ := env.svc.StartConversation(context.Background(), clientID, bh.ID, nil)

	msg, err := env.svc.SendMessage(context.Background(), clientID, domain.RoleClient, conv.ID, "Мой телеграм t.me/myuser")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if msg.Text == "Мой телеграм t.me/myuser" {
		t.Error("telegram link should have been filtered from message text")
	}
}

func TestChatService_SendMessage_CleanMessageNotFiltered(t *testing.T) {
	env := newChatTestEnv()
	ownerID := uuid.New()
	clientID := uuid.New()
	bh := createBathhouse(t, env.bhRepo, ownerID)

	conv, _ := env.svc.StartConversation(context.Background(), clientID, bh.ID, nil)

	msg, err := env.svc.SendMessage(context.Background(), clientID, domain.RoleClient, conv.ID, "Хочу забронировать на субботу в 14:00")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if msg.Text != "Хочу забронировать на субботу в 14:00" {
		t.Errorf("clean message was modified: %q", msg.Text)
	}
}
