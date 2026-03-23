package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/logger"
	"github.com/nikitaaldaev/bani/internal/repository"
)

type SessionService interface {
	CreateSession(ctx context.Context, userID uuid.UUID, deviceInfo, browser, ip string) (*domain.Session, error)
	ListSessions(ctx context.Context, userID uuid.UUID) ([]domain.Session, error)
	TerminateSession(ctx context.Context, userID uuid.UUID, sessionID uuid.UUID) error
	TerminateAllExceptCurrent(ctx context.Context, userID uuid.UUID, currentSessionID uuid.UUID) error
	ValidateSession(ctx context.Context, sessionID uuid.UUID) (*domain.Session, error)
	ValidateAndTouch(ctx context.Context, sessionID uuid.UUID) error
	CleanExpired(ctx context.Context) (int64, error)
}

type sessionService struct {
	sessionRepo repository.SessionRepository
	logger      *logger.Logger
}

func NewSessionService(sessionRepo repository.SessionRepository, log *logger.Logger) SessionService {
	return &sessionService{
		sessionRepo: sessionRepo,
		logger:      log,
	}
}

func (s *sessionService) CreateSession(ctx context.Context, userID uuid.UUID, deviceInfo, browser, ip string) (*domain.Session, error) {
	now := time.Now()
	session := &domain.Session{
		ID:           uuid.New(),
		UserID:       userID,
		DeviceInfo:   deviceInfo,
		Browser:      browser,
		IP:           ip,
		LastActiveAt: now,
		CreatedAt:    now,
		ExpiresAt:    now.Add(domain.SessionMaxAge),
	}

	if err := s.sessionRepo.Create(ctx, session); err != nil {
		return nil, err
	}

	return session, nil
}

func (s *sessionService) ListSessions(ctx context.Context, userID uuid.UUID) ([]domain.Session, error) {
	return s.sessionRepo.ListByUser(ctx, userID)
}

func (s *sessionService) TerminateSession(ctx context.Context, userID uuid.UUID, sessionID uuid.UUID) error {
	session, err := s.sessionRepo.GetByID(ctx, sessionID)
	if err != nil {
		return err
	}
	if session.UserID != userID {
		return domain.ErrForbidden
	}
	return s.sessionRepo.Delete(ctx, sessionID)
}

func (s *sessionService) TerminateAllExceptCurrent(ctx context.Context, userID uuid.UUID, currentSessionID uuid.UUID) error {
	return s.sessionRepo.DeleteAllExcept(ctx, userID, currentSessionID)
}

// ValidateAndTouch implements middleware.SessionValidator.
func (s *sessionService) ValidateAndTouch(ctx context.Context, sessionID uuid.UUID) error {
	session, err := s.sessionRepo.GetByID(ctx, sessionID)
	if err != nil {
		return domain.ErrSessionExpired
	}

	if time.Now().After(session.ExpiresAt) {
		_ = s.sessionRepo.Delete(ctx, sessionID)
		return domain.ErrSessionExpired
	}

	// Update last active (best-effort, don't fail the request)
	_ = s.sessionRepo.UpdateLastActive(ctx, sessionID, time.Now())
	return nil
}

func (s *sessionService) ValidateSession(ctx context.Context, sessionID uuid.UUID) (*domain.Session, error) {
	session, err := s.sessionRepo.GetByID(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	if time.Now().After(session.ExpiresAt) {
		_ = s.sessionRepo.Delete(ctx, sessionID)
		return nil, domain.ErrSessionExpired
	}

	return session, nil
}

func (s *sessionService) CleanExpired(ctx context.Context) (int64, error) {
	count, err := s.sessionRepo.DeleteExpired(ctx)
	if err != nil {
		s.logger.Error("Failed to clean expired sessions", "error", err)
		return 0, err
	}
	if count > 0 {
		s.logger.Info("Cleaned expired sessions", "count", count)
	}
	return count, nil
}
