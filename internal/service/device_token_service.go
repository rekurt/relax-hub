package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/rekurt/relax-hub/internal/domain"
	"github.com/rekurt/relax-hub/internal/repository"
)

type DeviceTokenService interface {
	Register(ctx context.Context, token *domain.DeviceToken) error
	Delete(ctx context.Context, id uuid.UUID, userID uuid.UUID) error
}

type deviceTokenService struct {
	repo repository.DeviceTokenRepository
}

func NewDeviceTokenService(repo repository.DeviceTokenRepository) DeviceTokenService {
	return &deviceTokenService{repo: repo}
}

func (s *deviceTokenService) Register(ctx context.Context, token *domain.DeviceToken) error {
	return s.repo.Create(ctx, token)
}

func (s *deviceTokenService) Delete(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	return s.repo.Delete(ctx, id, userID)
}
