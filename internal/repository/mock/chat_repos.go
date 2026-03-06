package mock

import (
	"context"
	"math"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
)

// ConversationRepo is an in-memory mock implementation of repository.ConversationRepository.
type ConversationRepo struct {
	mu            sync.RWMutex
	conversations map[uuid.UUID]*domain.Conversation
}

func NewConversationRepo() *ConversationRepo {
	return &ConversationRepo{conversations: make(map[uuid.UUID]*domain.Conversation)}
}

func (r *ConversationRepo) Create(_ context.Context, conv *domain.Conversation) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if conv.ID == uuid.Nil {
		conv.ID = uuid.New()
	}
	if conv.CreatedAt.IsZero() {
		conv.CreatedAt = time.Now()
	}

	for _, c := range r.conversations {
		if c.BathhouseID == conv.BathhouseID && c.ClientID == conv.ClientID {
			return domain.ErrAlreadyExists
		}
	}

	cp := *conv
	r.conversations[conv.ID] = &cp
	return nil
}

func (r *ConversationRepo) GetByID(_ context.Context, id uuid.UUID) (*domain.Conversation, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	c, ok := r.conversations[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	cp := *c
	return &cp, nil
}

func (r *ConversationRepo) GetByParticipants(_ context.Context, bathhouseID, clientID uuid.UUID) (*domain.Conversation, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, c := range r.conversations {
		if c.BathhouseID == bathhouseID && c.ClientID == clientID {
			cp := *c
			return &cp, nil
		}
	}
	return nil, domain.ErrNotFound
}

func (r *ConversationRepo) ListByUser(_ context.Context, userID uuid.UUID, bathhouseIDs []uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.Conversation], error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}

	bhSet := make(map[uuid.UUID]bool, len(bathhouseIDs))
	for _, id := range bathhouseIDs {
		bhSet[id] = true
	}

	var filtered []domain.Conversation
	for _, c := range r.conversations {
		if len(bathhouseIDs) > 0 {
			if bhSet[c.BathhouseID] {
				filtered = append(filtered, *c)
			}
		} else {
			if c.ClientID == userID {
				filtered = append(filtered, *c)
			}
		}
	}

	totalCount := int64(len(filtered))
	offset := (page - 1) * pageSize
	end := offset + pageSize
	if offset > len(filtered) {
		offset = len(filtered)
	}
	if end > len(filtered) {
		end = len(filtered)
	}

	return &domain.PaginatedResult[domain.Conversation]{
		Items:      filtered[offset:end],
		TotalCount: totalCount,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: int(math.Ceil(float64(totalCount) / float64(pageSize))),
	}, nil
}

func (r *ConversationRepo) GetOrCreate(ctx context.Context, conv *domain.Conversation) (*domain.Conversation, error) {
	existing, err := r.GetByParticipants(ctx, conv.BathhouseID, conv.ClientID)
	if err == nil {
		return existing, nil
	}
	if err != domain.ErrNotFound {
		return nil, err
	}
	if err := r.Create(ctx, conv); err != nil {
		return nil, err
	}
	cp := *conv
	return &cp, nil
}

func (r *ConversationRepo) UpdateLastMessageAt(_ context.Context, id uuid.UUID, t time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	c, ok := r.conversations[id]
	if !ok {
		return domain.ErrNotFound
	}
	c.LastMessageAt = &t
	return nil
}

// MessageRepo is an in-memory mock implementation of repository.MessageRepository.
type MessageRepo struct {
	mu       sync.RWMutex
	messages map[uuid.UUID]*domain.Message
}

func NewMessageRepo() *MessageRepo {
	return &MessageRepo{messages: make(map[uuid.UUID]*domain.Message)}
}

func (r *MessageRepo) Create(_ context.Context, msg *domain.Message) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if msg.ID == uuid.Nil {
		msg.ID = uuid.New()
	}
	if msg.CreatedAt.IsZero() {
		msg.CreatedAt = time.Now()
	}

	cp := *msg
	r.messages[msg.ID] = &cp
	return nil
}

func (r *MessageRepo) ListByConversation(_ context.Context, conversationID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.Message], error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}

	var filtered []domain.Message
	for _, m := range r.messages {
		if m.ConversationID == conversationID {
			filtered = append(filtered, *m)
		}
	}

	totalCount := int64(len(filtered))
	offset := (page - 1) * pageSize
	end := offset + pageSize
	if offset > len(filtered) {
		offset = len(filtered)
	}
	if end > len(filtered) {
		end = len(filtered)
	}

	return &domain.PaginatedResult[domain.Message]{
		Items:      filtered[offset:end],
		TotalCount: totalCount,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: int(math.Ceil(float64(totalCount) / float64(pageSize))),
	}, nil
}

func (r *MessageRepo) MarkAsRead(_ context.Context, conversationID, userID uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	for _, m := range r.messages {
		if m.ConversationID == conversationID && m.SenderID != userID && !m.IsRead {
			m.IsRead = true
			m.ReadAt = &now
		}
	}
	return nil
}

func (r *MessageRepo) CountUnread(_ context.Context, userID uuid.UUID, conversationIDs []uuid.UUID) (int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if len(conversationIDs) == 0 {
		return 0, nil
	}

	convSet := make(map[uuid.UUID]bool, len(conversationIDs))
	for _, id := range conversationIDs {
		convSet[id] = true
	}

	var count int64
	for _, m := range r.messages {
		if convSet[m.ConversationID] && m.SenderID != userID && !m.IsRead {
			count++
		}
	}
	return count, nil
}
