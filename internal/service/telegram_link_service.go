package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/rekurt/relax-hub/internal/domain"
	"github.com/rekurt/relax-hub/internal/repository"
)

// TelegramLinkService manages Telegram account linking
type TelegramLinkService interface {
	LinkAccount(ctx context.Context, userID uuid.UUID, telegramID int64, telegramUsername string) (*domain.TelegramLink, error)
	GetByTelegramID(ctx context.Context, telegramID int64) (*domain.TelegramLink, error)
	GetByUserID(ctx context.Context, userID uuid.UUID) (*domain.TelegramLink, error)
	UnlinkAccount(ctx context.Context, userID uuid.UUID) error
}

type telegramLinkService struct {
	telegramLinkRepo repository.TelegramLinkRepository
}

func NewTelegramLinkService(telegramLinkRepo repository.TelegramLinkRepository) TelegramLinkService {
	return &telegramLinkService{
		telegramLinkRepo: telegramLinkRepo,
	}
}

func (s *telegramLinkService) LinkAccount(ctx context.Context, userID uuid.UUID, telegramID int64, telegramUsername string) (*domain.TelegramLink, error) {
	link := &domain.TelegramLink{
		ID:               uuid.New(),
		UserID:           userID,
		TelegramID:       telegramID,
		TelegramUsername: telegramUsername,
		LinkedAt:         time.Now(),
	}

	if err := link.Validate(); err != nil {
		return nil, fmt.Errorf("%w: invalid telegram link data", err)
	}

	if err := s.telegramLinkRepo.Create(ctx, link); err != nil {
		return nil, err
	}

	return link, nil
}

func (s *telegramLinkService) GetByTelegramID(ctx context.Context, telegramID int64) (*domain.TelegramLink, error) {
	return s.telegramLinkRepo.GetByTelegramID(ctx, telegramID)
}

func (s *telegramLinkService) GetByUserID(ctx context.Context, userID uuid.UUID) (*domain.TelegramLink, error) {
	return s.telegramLinkRepo.GetByUserID(ctx, userID)
}

func (s *telegramLinkService) UnlinkAccount(ctx context.Context, userID uuid.UUID) error {
	return s.telegramLinkRepo.Delete(ctx, userID)
}
