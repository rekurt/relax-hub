package service

import (
	"context"
	"crypto/sha256"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/rekurt/relax-hub/internal/domain"
	"github.com/rekurt/relax-hub/internal/logger"
	"github.com/rekurt/relax-hub/internal/repository"
)

const (
	deletionGracePeriodDays = 30
)

type AccountDeletionService interface {
	RequestDeletion(ctx context.Context, userID uuid.UUID) error
	RestoreAccount(ctx context.Context, userID uuid.UUID) error
	ExecuteDeletion(ctx context.Context, userID uuid.UUID) error
	CheckPendingDeletions(ctx context.Context) (int, error)
	SendDeletionReminders(ctx context.Context) error
}

type accountDeletionService struct {
	userRepo   repository.UserRepository
	sessionSvc SessionService
	walletSvc  WalletService
	notifSvc   NotificationService
	logger     *logger.Logger
}

func NewAccountDeletionService(
	userRepo repository.UserRepository,
	sessionSvc SessionService,
	walletSvc WalletService,
	notifSvc NotificationService,
	log *logger.Logger,
) AccountDeletionService {
	return &accountDeletionService{
		userRepo:   userRepo,
		sessionSvc: sessionSvc,
		walletSvc:  walletSvc,
		notifSvc:   notifSvc,
		logger:     log,
	}
}

func (s *accountDeletionService) RequestDeletion(ctx context.Context, userID uuid.UUID) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	if user.DeletionScheduledAt != nil {
		return domain.ErrAccountDeletionPending
	}

	now := time.Now()
	scheduledAt := now.AddDate(0, 0, deletionGracePeriodDays)

	if err := s.userRepo.SetDeletionSchedule(ctx, userID, &now, &scheduledAt); err != nil {
		return err
	}

	_ = s.notifSvc.Send(ctx, userID, domain.NotifAccountDeletionRequested,
		"Запрос на удаление аккаунта",
		fmt.Sprintf("Ваш аккаунт будет удалён %s. Вы можете отменить удаление в настройках профиля.",
			scheduledAt.Format("02.01.2006")),
		map[string]string{
			"scheduled_at": scheduledAt.Format(time.RFC3339),
		},
	)

	return nil
}

func (s *accountDeletionService) RestoreAccount(ctx context.Context, userID uuid.UUID) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	if user.DeletionScheduledAt == nil {
		return domain.ErrAccountDeletionNotPending
	}

	if err := s.userRepo.SetDeletionSchedule(ctx, userID, nil, nil); err != nil {
		return err
	}

	_ = s.notifSvc.Send(ctx, userID, domain.NotifSystem,
		"Удаление аккаунта отменено",
		"Запрос на удаление вашего аккаунта был отменён. Ваш аккаунт восстановлен.",
		nil,
	)

	return nil
}

func (s *accountDeletionService) ExecuteDeletion(ctx context.Context, userID uuid.UUID) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	if user.DeletionScheduledAt == nil {
		return domain.ErrAccountDeletionNotPending
	}

	// Freeze wallet and zero balance (best-effort)
	if s.walletSvc != nil {
		if err := s.walletSvc.FreezeAndZeroBalance(ctx, userID); err != nil {
			s.logger.Warn("Failed to freeze/zero wallet on account deletion",
				"user_id", userID,
				"error", err,
			)
		}
	}

	// Anonymize: hash email for uniqueness, clear personal data
	anonEmail := fmt.Sprintf("deleted_%s@deleted.local", hashString(user.ID.String()))

	if err := s.userRepo.AnonymizeUser(ctx, userID, anonEmail); err != nil {
		return fmt.Errorf("anonymize user: %w", err)
	}

	// Terminate all sessions
	_ = s.sessionSvc.TerminateAllSessions(ctx, userID)

	s.logger.Info("Account deleted",
		"user_id", userID,
		"original_email_hash", hashString(user.Email),
	)

	return nil
}

func (s *accountDeletionService) CheckPendingDeletions(ctx context.Context) (int, error) {
	users, err := s.userRepo.ListPendingDeletions(ctx, time.Now())
	if err != nil {
		return 0, fmt.Errorf("list pending deletions: %w", err)
	}

	executed := 0
	for _, u := range users {
		if err := s.ExecuteDeletion(ctx, u.ID); err != nil {
			s.logger.Error("Failed to execute account deletion",
				"user_id", u.ID,
				"error", err,
			)
			continue
		}
		executed++
	}

	return executed, nil
}

func (s *accountDeletionService) SendDeletionReminders(ctx context.Context) error {
	// Look for users scheduled for deletion in the future and send reminders
	// at day 14 (halfway) and day 27 (3 days before)
	now := time.Now()

	// Day 14 reminder: users requested ~14 days ago (scheduled ~16 days from now)
	reminderWindows := []struct {
		daysFromRequest int
		notifType       domain.NotificationType
		title           string
		bodyFmt         string
	}{
		{
			daysFromRequest: 14,
			notifType:       domain.NotifAccountDeletionReminder,
			title:           "Напоминание об удалении аккаунта",
			bodyFmt:         "Ваш аккаунт будет удалён через %d дней. Отмените удаление, если передумали.",
		},
		{
			daysFromRequest: 27,
			notifType:       domain.NotifAccountDeletionFinal,
			title:           "Последнее предупреждение об удалении",
			bodyFmt:         "Ваш аккаунт будет удалён через %d дня. Это последний шанс отменить удаление.",
		},
	}

	for _, w := range reminderWindows {
		daysLeft := deletionGracePeriodDays - w.daysFromRequest

		// scheduled_at = requested_at + 30 days, so users in this window have
		// scheduled_at between daysLeft-1 and daysLeft+1 days from now
		scheduledStart := now.Add(time.Duration(daysLeft-1) * 24 * time.Hour)
		scheduledEnd := now.Add(time.Duration(daysLeft+1) * 24 * time.Hour)

		users, err := s.userRepo.ListPendingDeletions(ctx, scheduledEnd)
		if err != nil {
			s.logger.Error("Failed to list users for deletion reminder", "error", err)
			continue
		}

		for _, u := range users {
			if u.DeletionScheduledAt == nil {
				continue
			}
			if u.DeletionScheduledAt.Before(scheduledStart) || u.DeletionScheduledAt.After(scheduledEnd) {
				continue
			}
			if u.DeletionScheduledAt.Before(now) {
				continue
			}

			_ = s.notifSvc.Send(ctx, u.ID, w.notifType,
				w.title,
				fmt.Sprintf(w.bodyFmt, daysLeft),
				map[string]string{
					"scheduled_at": u.DeletionScheduledAt.Format(time.RFC3339),
					"days_left":    fmt.Sprintf("%d", daysLeft),
				},
			)
		}
	}

	return nil
}

func hashString(s string) string {
	h := sha256.Sum256([]byte(s))
	return fmt.Sprintf("%x", h[:8])
}
