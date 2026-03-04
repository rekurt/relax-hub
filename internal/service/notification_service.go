package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/logger"
	"github.com/nikitaaldaev/bani/internal/notification"
	"github.com/nikitaaldaev/bani/internal/repository"
)

type NotificationService interface {
	Send(ctx context.Context, userID uuid.UUID, notifType domain.NotificationType, title, body string, data map[string]string) error
	List(ctx context.Context, userID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.Notification], error)
	MarkAsRead(ctx context.Context, userID uuid.UUID, notificationID uuid.UUID) error
	MarkAllAsRead(ctx context.Context, userID uuid.UUID) error
	GetUnreadCount(ctx context.Context, userID uuid.UUID) (int64, error)
	GetPreferences(ctx context.Context, userID uuid.UUID) (*domain.NotificationPreferences, error)
	UpdatePreferences(ctx context.Context, userID uuid.UUID, prefs *domain.NotificationPreferences) error
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
