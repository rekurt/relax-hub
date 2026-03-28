package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/logger"
	"github.com/nikitaaldaev/bani/internal/repository"
)

const maxSavedCardsPerUser = 10

type SavedCardService interface {
	CreateSavedCard(ctx context.Context, userID uuid.UUID, card *domain.SavedCard) (*domain.SavedCard, error)
	DeleteSavedCard(ctx context.Context, userID uuid.UUID, cardID uuid.UUID) error
	ListSavedCards(ctx context.Context, userID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.SavedCard], error)
	GetSavedCard(ctx context.Context, userID uuid.UUID, cardID uuid.UUID) (*domain.SavedCard, error)
	SetDefaultCard(ctx context.Context, userID uuid.UUID, cardID uuid.UUID) error
}

type savedCardService struct {
	cardRepo repository.SavedCardRepository
	logger   *logger.Logger
}

func NewSavedCardService(
	cardRepo repository.SavedCardRepository,
	log *logger.Logger,
) SavedCardService {
	return &savedCardService{
		cardRepo: cardRepo,
		logger:   log,
	}
}

func (s *savedCardService) CreateSavedCard(ctx context.Context, userID uuid.UUID, card *domain.SavedCard) (*domain.SavedCard, error) {
	card.UserID = userID

	if err := card.Validate(); err != nil {
		return nil, err
	}

	count, err := s.cardRepo.CountByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	if count >= maxSavedCardsPerUser {
		return nil, domain.ErrSavedCardLimitReached
	}

	card.ID = uuid.New()
	now := time.Now()
	card.CreatedAt = now
	card.UpdatedAt = now

	// If this is the first card, make it default
	if count == 0 {
		card.IsDefault = true
	}

	if err := s.cardRepo.Create(ctx, card); err != nil {
		return nil, err
	}

	s.logger.Info("saved card created", "card_id", card.ID, "user_id", userID)
	return card, nil
}

func (s *savedCardService) DeleteSavedCard(ctx context.Context, userID uuid.UUID, cardID uuid.UUID) error {
	card, err := s.cardRepo.GetByID(ctx, cardID)
	if err != nil {
		return err
	}

	if card.UserID != userID {
		return domain.ErrForbidden
	}

	if err := s.cardRepo.Delete(ctx, cardID); err != nil {
		return err
	}

	s.logger.Info("saved card deleted", "card_id", cardID, "user_id", userID)
	return nil
}

func (s *savedCardService) ListSavedCards(ctx context.Context, userID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.SavedCard], error) {
	return s.cardRepo.ListByUser(ctx, userID, page, pageSize)
}

func (s *savedCardService) GetSavedCard(ctx context.Context, userID uuid.UUID, cardID uuid.UUID) (*domain.SavedCard, error) {
	card, err := s.cardRepo.GetByID(ctx, cardID)
	if err != nil {
		return nil, err
	}

	if card.UserID != userID {
		return nil, domain.ErrForbidden
	}

	return card, nil
}

func (s *savedCardService) SetDefaultCard(ctx context.Context, userID uuid.UUID, cardID uuid.UUID) error {
	card, err := s.cardRepo.GetByID(ctx, cardID)
	if err != nil {
		return err
	}

	if card.UserID != userID {
		return domain.ErrForbidden
	}

	return s.cardRepo.SetDefault(ctx, userID, cardID)
}
