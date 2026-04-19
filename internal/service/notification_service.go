package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/rekurt/relax-hub/internal/domain"
	"github.com/rekurt/relax-hub/internal/logger"
	"github.com/rekurt/relax-hub/internal/notification"
	"github.com/rekurt/relax-hub/internal/repository"
)

type NotificationService interface {
	Send(ctx context.Context, userID uuid.UUID, notifType domain.NotificationType, title, body string, data map[string]string) error
	List(ctx context.Context, userID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.Notification], error)
	MarkAsRead(ctx context.Context, userID uuid.UUID, notificationID uuid.UUID) error
	MarkAllAsRead(ctx context.Context, userID uuid.UUID) error
	GetUnreadCount(ctx context.Context, userID uuid.UUID) (int64, error)
	GetPreferences(ctx context.Context, userID uuid.UUID) (*domain.NotificationPreferences, error)
	UpdatePreferences(ctx context.Context, userID uuid.UUID, prefs *domain.NotificationPreferences) error
	HasRecentByType(ctx context.Context, userID uuid.UUID, notifType domain.NotificationType, since time.Time) (bool, error)
	GetEventPreferences(ctx context.Context, userID uuid.UUID) ([]domain.NotificationEventPreference, error)
	UpdateEventPreferences(ctx context.Context, userID uuid.UUID, prefs []domain.NotificationEventPreference) error
}

type notificationService struct {
	notifRepo  repository.NotificationRepository
	userRepo   repository.UserRepository
	dispatcher *notification.Dispatcher
	logger     *logger.Logger
}

func NewNotificationService(
	notifRepo repository.NotificationRepository,
	userRepo repository.UserRepository,
	dispatcher *notification.Dispatcher,
	log *logger.Logger,
) NotificationService {
	return &notificationService{
		notifRepo:  notifRepo,
		userRepo:   userRepo,
		dispatcher: dispatcher,
		logger:     log,
	}
}

func (s *notificationService) Send(ctx context.Context, userID uuid.UUID, notifType domain.NotificationType, title, body string, data map[string]string) error {
	notif := &domain.Notification{
		ID:        uuid.New(),
		UserID:    userID,
		Type:      notifType,
		Title:     title,
		Body:      body,
		Data:      data,
		CreatedAt: time.Now(),
	}

	if err := notif.Validate(); err != nil {
		return err
	}

	prefs, err := s.notifRepo.GetPreferences(ctx, userID)
	if err != nil {
		return err
	}

	var email string
	if prefs.Email {
		user, err := s.userRepo.GetByID(ctx, userID)
		if err != nil {
			s.logger.Warn("failed to get user for email notification", "user_id", userID, "error", err)
		} else {
			email = user.Email
		}
	}

	s.dispatcher.Dispatch(ctx, notif, prefs, email)
	return nil
}

func (s *notificationService) List(ctx context.Context, userID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.Notification], error) {
	return s.notifRepo.ListByUser(ctx, userID, page, pageSize)
}

func (s *notificationService) MarkAsRead(ctx context.Context, userID uuid.UUID, notificationID uuid.UUID) error {
	notif, err := s.notifRepo.GetByID(ctx, notificationID)
	if err != nil {
		return err
	}
	if notif.UserID != userID {
		return domain.ErrForbidden
	}
	return s.notifRepo.MarkAsRead(ctx, notificationID)
}

func (s *notificationService) MarkAllAsRead(ctx context.Context, userID uuid.UUID) error {
	return s.notifRepo.MarkAllAsRead(ctx, userID)
}

func (s *notificationService) GetUnreadCount(ctx context.Context, userID uuid.UUID) (int64, error) {
	return s.notifRepo.CountUnread(ctx, userID)
}

func (s *notificationService) GetPreferences(ctx context.Context, userID uuid.UUID) (*domain.NotificationPreferences, error) {
	return s.notifRepo.GetPreferences(ctx, userID)
}

func (s *notificationService) UpdatePreferences(ctx context.Context, userID uuid.UUID, prefs *domain.NotificationPreferences) error {
	prefs.UserID = userID
	if err := prefs.Validate(); err != nil {
		return err
	}
	return s.notifRepo.UpdatePreferences(ctx, prefs)
}

func (s *notificationService) HasRecentByType(ctx context.Context, userID uuid.UUID, notifType domain.NotificationType, since time.Time) (bool, error) {
	return s.notifRepo.HasRecentByType(ctx, userID, notifType, since)
}

func (s *notificationService) GetEventPreferences(ctx context.Context, userID uuid.UUID) ([]domain.NotificationEventPreference, error) {
	stored, err := s.notifRepo.GetEventPreferences(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Build a map of stored preferences
	storedMap := make(map[domain.NotificationEventType]domain.NotificationEventPreference, len(stored))
	for _, p := range stored {
		storedMap[p.EventType] = p
	}

	// Return full list with defaults for missing event types
	result := make([]domain.NotificationEventPreference, 0, len(domain.AllNotificationEventTypes))
	for _, et := range domain.AllNotificationEventTypes {
		if p, ok := storedMap[et]; ok {
			result = append(result, p)
		} else {
			result = append(result, domain.DefaultEventPreference(userID, et))
		}
	}
	return result, nil
}

func (s *notificationService) UpdateEventPreferences(ctx context.Context, userID uuid.UUID, prefs []domain.NotificationEventPreference) error {
	// Enforce mandatory events: cannot disable channels for mandatory events
	for i := range prefs {
		prefs[i].UserID = userID
		if !prefs[i].EventType.IsValid() {
			return domain.ErrInvalidInput
		}
		if prefs[i].EventType.IsMandatory() {
			prefs[i].PushEnabled = true
			prefs[i].EmailEnabled = true
		}
	}
	return s.notifRepo.UpsertEventPreferences(ctx, prefs)
}
