package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/repository"
)

type AmenityService interface {
	Create(ctx context.Context, amenity *domain.Amenity) error
	Update(ctx context.Context, amenity *domain.Amenity) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Amenity, error)
	ListAll(ctx context.Context) ([]domain.Amenity, error)
}

type amenityService struct {
	repo repository.AmenityRepository
}

func NewAmenityService(repo repository.AmenityRepository) AmenityService {
	return &amenityService{repo: repo}
}

func (s *amenityService) Create(ctx context.Context, amenity *domain.Amenity) error {
	if err := amenity.Validate(); err != nil {
		return err
	}
	return s.repo.Create(ctx, amenity)
}

func (s *amenityService) Update(ctx context.Context, amenity *domain.Amenity) error {
	if err := amenity.Validate(); err != nil {
		return err
	}
	return s.repo.Update(ctx, amenity)
}

func (s *amenityService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}

func (s *amenityService) GetByID(ctx context.Context, id uuid.UUID) (*domain.Amenity, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *amenityService) ListAll(ctx context.Context) ([]domain.Amenity, error) {
	return s.repo.ListAll(ctx)
}
