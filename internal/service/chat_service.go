package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/logger"
	"github.com/nikitaaldaev/bani/internal/repository"
)

type ChatService interface {
	StartConversation(ctx context.Context, clientID uuid.UUID, bathhouseID uuid.UUID, bookingID *uuid.UUID) (*domain.Conversation, error)
	SendMessage(ctx context.Context, senderID uuid.UUID, role domain.UserRole, conversationID uuid.UUID, text string) (*domain.Message, error)
	ListConversations(ctx context.Context, userID uuid.UUID, role domain.UserRole, page, pageSize int) (*domain.PaginatedResult[domain.Conversation], error)
	ListMessages(ctx context.Context, userID uuid.UUID, role domain.UserRole, conversationID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.Message], error)
	MarkAsRead(ctx context.Context, userID uuid.UUID, role domain.UserRole, conversationID uuid.UUID) error
	GetUnreadCount(ctx context.Context, userID uuid.UUID, role domain.UserRole) (int64, error)
}

type chatService struct {
	convRepo repository.ConversationRepository
	msgRepo  repository.MessageRepository
	bhRepo   repository.BathhouseRepository
	repRepo  repository.RepresentativeRepository
	access   *AccessChecker
	notifSvc NotificationService
	logger   *logger.Logger
}

func NewChatService(
	convRepo repository.ConversationRepository,
	msgRepo repository.MessageRepository,
	bhRepo repository.BathhouseRepository,
	repRepo repository.RepresentativeRepository,
	access *AccessChecker,
	notifSvc NotificationService,
	log *logger.Logger,
) ChatService {
	return &chatService{
		convRepo: convRepo,
		msgRepo:  msgRepo,
		bhRepo:   bhRepo,
		repRepo:  repRepo,
		access:   access,
		notifSvc: notifSvc,
		logger:   log,
	}
}

func (s *chatService) StartConversation(ctx context.Context, clientID uuid.UUID, bathhouseID uuid.UUID, bookingID *uuid.UUID) (*domain.Conversation, error) {
	if _, err := s.bhRepo.GetByID(ctx, bathhouseID); err != nil {
		return nil, err
	}

	conv := &domain.Conversation{
		ID:          uuid.New(),
		BathhouseID: bathhouseID,
		ClientID:    clientID,
		BookingID:   bookingID,
		CreatedAt:   time.Now(),
	}

	if err := conv.Validate(); err != nil {
		return nil, err
	}

	return s.convRepo.GetOrCreate(ctx, conv)
}

func (s *chatService) SendMessage(ctx context.Context, senderID uuid.UUID, role domain.UserRole, conversationID uuid.UUID, text string) (*domain.Message, error) {
	if text == "" {
		return nil, domain.ErrInvalidInput
	}

	conv, err := s.convRepo.GetByID(ctx, conversationID)
	if err != nil {
		return nil, err
	}

	if err := s.checkConversationAccess(ctx, senderID, role, conv); err != nil {
		return nil, err
	}

	msg := &domain.Message{
		ID:             uuid.New(),
		ConversationID: conversationID,
		SenderID:       senderID,
		Text:           text,
		CreatedAt:      time.Now(),
	}

	if err := msg.Validate(); err != nil {
		return nil, err
	}

	if err := s.msgRepo.Create(ctx, msg); err != nil {
		return nil, err
	}

	now := time.Now()
	if err := s.convRepo.UpdateLastMessageAt(ctx, conversationID, now); err != nil {
		s.logger.Warn("failed to update conversation last_message_at", "conversation_id", conversationID, "error", err)
	}

	s.sendMessageNotification(ctx, conv, senderID, text)

	return msg, nil
}

func (s *chatService) ListConversations(ctx context.Context, userID uuid.UUID, role domain.UserRole, page, pageSize int) (*domain.PaginatedResult[domain.Conversation], error) {
	bathhouseIDs, err := s.getUserBathhouseIDs(ctx, userID, role)
	if err != nil {
		return nil, err
	}

	return s.convRepo.ListByUser(ctx, userID, bathhouseIDs, page, pageSize)
}

func (s *chatService) ListMessages(ctx context.Context, userID uuid.UUID, role domain.UserRole, conversationID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.Message], error) {
	conv, err := s.convRepo.GetByID(ctx, conversationID)
	if err != nil {
		return nil, err
	}

	if err := s.checkConversationAccess(ctx, userID, role, conv); err != nil {
		return nil, err
	}

	return s.msgRepo.ListByConversation(ctx, conversationID, page, pageSize)
}

func (s *chatService) MarkAsRead(ctx context.Context, userID uuid.UUID, role domain.UserRole, conversationID uuid.UUID) error {
	conv, err := s.convRepo.GetByID(ctx, conversationID)
	if err != nil {
		return err
	}

	if err := s.checkConversationAccess(ctx, userID, role, conv); err != nil {
		return err
	}

	return s.msgRepo.MarkAsRead(ctx, conversationID, userID)
}

func (s *chatService) GetUnreadCount(ctx context.Context, userID uuid.UUID, role domain.UserRole) (int64, error) {
	bathhouseIDs, err := s.getUserBathhouseIDs(ctx, userID, role)
	if err != nil {
		return 0, err
	}

	result, err := s.convRepo.ListByUser(ctx, userID, bathhouseIDs, 1, 10000)
	if err != nil {
		return 0, err
	}

	if len(result.Items) == 0 {
		return 0, nil
	}

	convIDs := make([]uuid.UUID, len(result.Items))
	for i, c := range result.Items {
		convIDs[i] = c.ID
	}

	return s.msgRepo.CountUnread(ctx, userID, convIDs)
}

// checkConversationAccess verifies the user is a participant of the conversation.
// Clients can access conversations where they are the client.
// Owners/representatives can access conversations for bathhouses they manage.
// Admins can access all conversations.
func (s *chatService) checkConversationAccess(ctx context.Context, userID uuid.UUID, role domain.UserRole, conv *domain.Conversation) error {
	if role == domain.RoleAdmin {
		return nil
	}

	if conv.ClientID == userID {
		return nil
	}

	return s.access.CanManageBathhouse(ctx, userID, role, conv.BathhouseID)
}

// getUserBathhouseIDs returns bathhouse IDs the user manages (for owners/reps).
// For clients, returns nil so ListByUser filters by clientID instead.
func (s *chatService) getUserBathhouseIDs(ctx context.Context, userID uuid.UUID, role domain.UserRole) ([]uuid.UUID, error) {
	switch role {
	case domain.RoleOwner:
		result, err := s.bhRepo.ListByOwner(ctx, userID, 1, 10000)
		if err != nil {
			return nil, err
		}
		ids := make([]uuid.UUID, len(result.Items))
		for i, bh := range result.Items {
			ids[i] = bh.ID
		}
		return ids, nil

	case domain.RoleRepresentative:
		reps, err := s.repRepo.ListByUser(ctx, userID)
		if err != nil {
			return nil, err
		}
		ids := make([]uuid.UUID, len(reps))
		for i, r := range reps {
			ids[i] = r.BathhouseID
		}
		return ids, nil

	case domain.RoleAdmin:
		// Admin sees all — pass nil to get all conversations
		// ListByUser with nil bathhouseIDs will filter by clientID,
		// which won't work for admin. For now, return empty slice
		// meaning admin would need a separate admin endpoint.
		return nil, nil

	default:
		// Client role — nil means filter by clientID
		return nil, nil
	}
}

func (s *chatService) sendMessageNotification(ctx context.Context, conv *domain.Conversation, senderID uuid.UUID, text string) {
	recipientID := conv.ClientID
	if senderID == conv.ClientID {
		bh, err := s.bhRepo.GetByID(ctx, conv.BathhouseID)
		if err != nil {
			s.logger.Warn("failed to get bathhouse for notification", "bathhouse_id", conv.BathhouseID, "error", err)
			return
		}
		recipientID = bh.OwnerID
	}

	preview := text
	if len(preview) > 100 {
		preview = preview[:100] + "..."
	}

	data := map[string]string{
		"conversation_id": conv.ID.String(),
		"sender_id":       senderID.String(),
		"bathhouse_id":    conv.BathhouseID.String(),
	}

	if err := s.notifSvc.Send(ctx, recipientID, domain.NotifNewMessage,
		"Новое сообщение",
		fmt.Sprintf("У вас новое сообщение: %s", preview),
		data,
	); err != nil {
		s.logger.Warn("failed to send message notification", "conversation_id", conv.ID, "error", err)
	}
}
