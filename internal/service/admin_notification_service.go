package service

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/logger"
	"github.com/nikitaaldaev/bani/internal/notification"
	"github.com/nikitaaldaev/bani/internal/repository"
)

// AdminNotificationService manages admin notifications.
type AdminNotificationService interface {
	// Emit creates admin notifications for all target roles of the given type.
	Emit(ctx context.Context, notifType domain.AdminNotificationType, severity domain.AdminNotifSeverity, title, body string, data map[string]interface{}) error
	// List returns paginated admin notifications filtered by criteria.
	List(ctx context.Context, filter domain.AdminNotificationFilter) (*domain.PaginatedResult[domain.AdminNotification], error)
	// MarkAsRead marks a single notification as read.
	MarkAsRead(ctx context.Context, notifID, adminID uuid.UUID) error
	// MarkAllAsRead marks all notifications for a role as read.
	MarkAllAsRead(ctx context.Context, role domain.AdminSubRole, adminID uuid.UUID) error
	// CountUnread returns the count of unread notifications for a role.
	CountUnread(ctx context.Context, role domain.AdminSubRole) (int64, error)
	// ListUnreadCritical returns unread error/critical notifications for a role.
	ListUnreadCritical(ctx context.Context, role domain.AdminSubRole) ([]domain.AdminNotification, error)
	// SendDailyDigest sends email digest of unread critical alerts per admin role.
	SendDailyDigest(ctx context.Context) error
}

type adminNotificationService struct {
	repo        repository.AdminNotificationRepository
	userRepo    repository.UserRepository
	emailSender notification.EmailSender
	logger      *logger.Logger
}

func NewAdminNotificationService(
	repo repository.AdminNotificationRepository,
	userRepo repository.UserRepository,
	emailSender notification.EmailSender,
	log *logger.Logger,
) AdminNotificationService {
	return &adminNotificationService{
		repo:        repo,
		userRepo:    userRepo,
		emailSender: emailSender,
		logger:      log,
	}
}

func (s *adminNotificationService) Emit(ctx context.Context, notifType domain.AdminNotificationType, severity domain.AdminNotifSeverity, title, body string, data map[string]interface{}) error {
	if !notifType.IsValid() {
		return fmt.Errorf("invalid admin notification type: %w", domain.ErrInvalidInput)
	}
	if !severity.IsValid() {
		return fmt.Errorf("invalid admin notification severity: %w", domain.ErrInvalidInput)
	}

	var dataJSON json.RawMessage
	if data != nil {
		b, err := json.Marshal(data)
		if err != nil {
			return fmt.Errorf("marshal notification data: %w", err)
		}
		dataJSON = b
	} else {
		dataJSON = []byte("{}")
	}

	targetRoles := domain.AdminNotifTargetRoles(notifType)
	for _, role := range targetRoles {
		notif := &domain.AdminNotification{
			Role:     role,
			Severity: severity,
			Type:     notifType,
			Title:    title,
			Body:     body,
			Data:     dataJSON,
		}
		if err := s.repo.Create(ctx, notif); err != nil {
			s.logger.Error("failed to create admin notification",
				"role", role, "type", notifType, "error", err)
		}
	}

	s.logger.Info("admin notification emitted",
		"type", notifType, "severity", severity, "roles", len(targetRoles))
	return nil
}

func (s *adminNotificationService) List(ctx context.Context, filter domain.AdminNotificationFilter) (*domain.PaginatedResult[domain.AdminNotification], error) {
	return s.repo.List(ctx, filter)
}

func (s *adminNotificationService) MarkAsRead(ctx context.Context, notifID, adminID uuid.UUID) error {
	return s.repo.MarkAsRead(ctx, notifID, adminID)
}

func (s *adminNotificationService) MarkAllAsRead(ctx context.Context, role domain.AdminSubRole, adminID uuid.UUID) error {
	return s.repo.MarkAllAsReadByRole(ctx, role, adminID)
}

func (s *adminNotificationService) CountUnread(ctx context.Context, role domain.AdminSubRole) (int64, error) {
	return s.repo.CountUnreadByRole(ctx, role)
}

func (s *adminNotificationService) ListUnreadCritical(ctx context.Context, role domain.AdminSubRole) ([]domain.AdminNotification, error) {
	return s.repo.ListUnreadCriticalByRole(ctx, role)
}

func (s *adminNotificationService) SendDailyDigest(ctx context.Context) error {
	roles := domain.AllAdminSubRoles()
	for _, role := range roles {
		alerts, err := s.repo.ListUnreadCriticalByRole(ctx, role)
		if err != nil {
			s.logger.Error("failed to list critical alerts for digest", "role", role, "error", err)
			continue
		}
		if len(alerts) == 0 {
			continue
		}

		// Build digest email body
		subject := fmt.Sprintf("RelaxHub: %d непрочитанных критических оповещений (%s)", len(alerts), string(role))
		body := fmt.Sprintf("Непрочитанные критические оповещения для роли %s:\n\n", string(role))
		for i, a := range alerts {
			body += fmt.Sprintf("%d. [%s] %s\n   %s\n   Создано: %s\n\n",
				i+1, string(a.Severity), a.Title, a.Body, a.CreatedAt.Format("02.01.2006 15:04"))
		}

		// Find admin users with this role and send email
		users, err := s.userRepo.List(ctx, 1, 1000)
		if err != nil {
			s.logger.Error("failed to list users for digest", "error", err)
			continue
		}
		for _, u := range users.Items {
			if u.Role != domain.RoleAdmin || u.AdminSubRole != role {
				continue
			}
			if u.Email == "" {
				continue
			}
			if s.emailSender != nil {
				if err := s.emailSender.Send(ctx, u.Email, subject, body); err != nil {
					s.logger.Warn("failed to send digest email",
						"email", u.Email, "role", role, "error", err)
				}
			}
		}
	}
	return nil
}
