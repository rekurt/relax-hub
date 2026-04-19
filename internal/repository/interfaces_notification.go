package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/rekurt/relax-hub/internal/domain"
)

type NotificationRepository interface {
	Create(ctx context.Context, notification *domain.Notification) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Notification, error)
	ListByUser(ctx context.Context, userID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.Notification], error)
	MarkAsRead(ctx context.Context, id uuid.UUID) error
	MarkAllAsRead(ctx context.Context, userID uuid.UUID) error
	CountUnread(ctx context.Context, userID uuid.UUID) (int64, error)
	GetPreferences(ctx context.Context, userID uuid.UUID) (*domain.NotificationPreferences, error)
	UpdatePreferences(ctx context.Context, prefs *domain.NotificationPreferences) error
	HasRecentByType(ctx context.Context, userID uuid.UUID, notifType domain.NotificationType, since time.Time) (bool, error)
	GetEventPreferences(ctx context.Context, userID uuid.UUID) ([]domain.NotificationEventPreference, error)
	GetEventPreference(ctx context.Context, userID uuid.UUID, eventType domain.NotificationEventType) (*domain.NotificationEventPreference, error)
	UpsertEventPreference(ctx context.Context, pref *domain.NotificationEventPreference) error
	UpsertEventPreferences(ctx context.Context, prefs []domain.NotificationEventPreference) error
	CreatePushDeliveryLog(ctx context.Context, log *domain.PushDeliveryLog) error
	GetPendingPushDeliveries(ctx context.Context, olderThan time.Time) ([]domain.PushDeliveryLog, error)
	UpdatePushDeliveryStatus(ctx context.Context, id uuid.UUID, status domain.PushDeliveryStatus) error
	MarkPushFallbackSent(ctx context.Context, id uuid.UUID) error
}


type ConversationRepository interface {
	Create(ctx context.Context, conv *domain.Conversation) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Conversation, error)
	GetByParticipants(ctx context.Context, bathhouseID, clientID uuid.UUID) (*domain.Conversation, error)
	ListByUser(ctx context.Context, userID uuid.UUID, bathhouseIDs []uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.Conversation], error)
	ListAll(ctx context.Context, page, pageSize int) (*domain.PaginatedResult[domain.Conversation], error)
	GetOrCreate(ctx context.Context, conv *domain.Conversation) (*domain.Conversation, error)
	UpdateLastMessageAt(ctx context.Context, id uuid.UUID, t time.Time) error
}


type MessageRepository interface {
	Create(ctx context.Context, msg *domain.Message) error
	ListByConversation(ctx context.Context, conversationID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.Message], error)
	MarkAsRead(ctx context.Context, conversationID, userID uuid.UUID) error
	CountUnread(ctx context.Context, userID uuid.UUID, conversationIDs []uuid.UUID) (int64, error)
	CountUnreadByUser(ctx context.Context, userID uuid.UUID, bathhouseIDs []uuid.UUID) (int64, error)
}


type TelegramLinkRepository interface {
	Create(ctx context.Context, link *domain.TelegramLink) error
	GetByTelegramID(ctx context.Context, telegramID int64) (*domain.TelegramLink, error)
	GetByUserID(ctx context.Context, userID uuid.UUID) (*domain.TelegramLink, error)
	Delete(ctx context.Context, userID uuid.UUID) error
}


// WebhookRepository manages owner outgoing webhooks.
type WebhookRepository interface {
	Create(ctx context.Context, webhook *domain.Webhook) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Webhook, error)
	Update(ctx context.Context, webhook *domain.Webhook) error
	Delete(ctx context.Context, id uuid.UUID) error
	ListByOwner(ctx context.Context, ownerID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.Webhook], error)
	CountByOwner(ctx context.Context, ownerID uuid.UUID) (int64, error)
	ListActiveByEvent(ctx context.Context, bathhouseOwnerID uuid.UUID, event domain.WebhookEventType) ([]domain.Webhook, error)
}


// WebhookDeliveryRepository manages webhook delivery attempts.
type WebhookDeliveryRepository interface {
	Create(ctx context.Context, delivery *domain.WebhookDelivery) error
	Update(ctx context.Context, delivery *domain.WebhookDelivery) error
	ListByWebhook(ctx context.Context, webhookID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.WebhookDelivery], error)
	ListPendingRetries(ctx context.Context, before time.Time) ([]domain.WebhookDelivery, error)
}

