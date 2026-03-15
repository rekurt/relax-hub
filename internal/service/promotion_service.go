package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/logger"
	"github.com/nikitaaldaev/bani/internal/repository"
)

type PromotionService interface {
	RecordImpression(ctx context.Context, bathhouseID uuid.UUID) error
	RecordClick(ctx context.Context, bathhouseID uuid.UUID) error
	GetActiveBybathhouse(ctx context.Context, bathhouseID uuid.UUID) (*domain.Promotion, error)
	Create(ctx context.Context, promo *domain.Promotion) error
}

type promotionService struct {
	promoRepo repository.PromotionRepository
	log       *logger.Logger
}

func NewPromotionService(promoRepo repository.PromotionRepository, log *logger.Logger) PromotionService {
	return &promotionService{promoRepo: promoRepo, log: log}
}

func (s *promotionService) RecordImpression(ctx context.Context, bathhouseID uuid.UUID) error {
	promo, err := s.promoRepo.GetActiveBybathhouse(ctx, bathhouseID)
	if err != nil || promo == nil {
		return nil
	}
	if err := s.promoRepo.RecordImpression(ctx, promo.ID); err != nil {
		s.log.Warn("failed to record promotion impression", "promotion_id", promo.ID, "error", err)
	}
	return nil
}

func (s *promotionService) RecordClick(ctx context.Context, bathhouseID uuid.UUID) error {
	promo, err := s.promoRepo.GetActiveBybathhouse(ctx, bathhouseID)
	if err != nil || promo == nil {
		return nil
	}
	if err := s.promoRepo.RecordClick(ctx, promo.ID); err != nil {
		s.log.Warn("failed to record promotion click", "promotion_id", promo.ID, "error", err)
	}
	return nil
}

func (s *promotionService) GetActiveBybathhouse(ctx context.Context, bathhouseID uuid.UUID) (*domain.Promotion, error) {
	return s.promoRepo.GetActiveBybathhouse(ctx, bathhouseID)
}

func (s *promotionService) Create(ctx context.Context, promo *domain.Promotion) error {
	return s.promoRepo.Create(ctx, promo)
}
