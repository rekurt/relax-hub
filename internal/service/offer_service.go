package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/rekurt/relax-hub/internal/domain"
	"github.com/rekurt/relax-hub/internal/logger"
	"github.com/rekurt/relax-hub/internal/repository"
	"github.com/spf13/viper"
)

type AcceptOfferInput struct {
	IPAddress string
	UserAgent string
}

type OfferStatusResponse struct {
	Accepted       bool                    `json:"accepted"`
	CurrentVersion string                  `json:"current_version"`
	Acceptance     *domain.OfferAcceptance `json:"acceptance,omitempty"`
}

type OfferService interface {
	Accept(ctx context.Context, userID uuid.UUID, input AcceptOfferInput) (*domain.OfferAcceptance, error)
	GetStatus(ctx context.Context, userID uuid.UUID) (*OfferStatusResponse, error)
	IsAccepted(ctx context.Context, userID uuid.UUID) (bool, error)
}

type offerService struct {
	offerRepo repository.OfferRepository
	logger    *logger.Logger
}

func NewOfferService(
	offerRepo repository.OfferRepository,
	log *logger.Logger,
) OfferService {
	return &offerService{
		offerRepo: offerRepo,
		logger:    log,
	}
}

func (s *offerService) currentVersion() string {
	v := viper.GetString("offer.current_version")
	if v == "" {
		return "1.0"
	}
	return v
}

func (s *offerService) Accept(ctx context.Context, userID uuid.UUID, input AcceptOfferInput) (*domain.OfferAcceptance, error) {
	version := s.currentVersion()

	// Check if already accepted this version
	_, err := s.offerRepo.GetByUserAndVersion(ctx, userID, version)
	if err == nil {
		return nil, domain.ErrOfferAlreadyAccepted
	}

	now := time.Now()
	acceptance := &domain.OfferAcceptance{
		ID:           uuid.New(),
		UserID:       userID,
		OfferVersion: version,
		AcceptedAt:   now,
		IPAddress:    input.IPAddress,
		UserAgent:    input.UserAgent,
		CreatedAt:    now,
	}

	if err := s.offerRepo.Create(ctx, acceptance); err != nil {
		return nil, err
	}

	s.logger.Info("offer accepted", "user_id", userID, "version", version)
	return acceptance, nil
}

func (s *offerService) GetStatus(ctx context.Context, userID uuid.UUID) (*OfferStatusResponse, error) {
	version := s.currentVersion()

	acceptance, err := s.offerRepo.GetByUserAndVersion(ctx, userID, version)
	if err != nil {
		if errors.Is(err, domain.ErrOfferNotFound) {
			return &OfferStatusResponse{
				Accepted:       false,
				CurrentVersion: version,
			}, nil
		}
		return nil, err
	}

	return &OfferStatusResponse{
		Accepted:       true,
		CurrentVersion: version,
		Acceptance:     acceptance,
	}, nil
}

func (s *offerService) IsAccepted(ctx context.Context, userID uuid.UUID) (bool, error) {
	version := s.currentVersion()
	_, err := s.offerRepo.GetByUserAndVersion(ctx, userID, version)
	if err != nil {
		if errors.Is(err, domain.ErrOfferNotFound) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
