package service

import (
	"context"

	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/repository"
)

type CreateCityInput struct {
	Name      string
	Slug      string
	Latitude  float64
	Longitude float64
}

type UpdateCityInput struct {
	Name      *string
	Slug      *string
	Latitude  *float64
	Longitude *float64
}

type CityService interface {
	GetAll(ctx context.Context) ([]domain.City, error)
	GetBySlug(ctx context.Context, slug string) (*domain.City, error)
	GetByID(ctx context.Context, id int64) (*domain.City, error)
	// Admin:
	Create(ctx context.Context, input CreateCityInput) (*domain.City, error)
	Update(ctx context.Context, id int64, input UpdateCityInput) (*domain.City, error)
	Delete(ctx context.Context, id int64) error
}

type cityService struct {
	cityRepo repository.CityRepository
}

func NewCityService(cityRepo repository.CityRepository) CityService {
	return &cityService{cityRepo: cityRepo}
}

func (s *cityService) GetAll(ctx context.Context) ([]domain.City, error) {
	return s.cityRepo.GetAll(ctx)
}

func (s *cityService) GetBySlug(ctx context.Context, slug string) (*domain.City, error) {
	return s.cityRepo.GetBySlug(ctx, slug)
}

func (s *cityService) GetByID(ctx context.Context, id int64) (*domain.City, error) {
	return s.cityRepo.GetByID(ctx, id)
}

func (s *cityService) Create(ctx context.Context, input CreateCityInput) (*domain.City, error) {
	city := &domain.City{
		Name:      input.Name,
		Slug:      input.Slug,
		Latitude:  input.Latitude,
		Longitude: input.Longitude,
	}

	if err := city.Validate(); err != nil {
		return nil, err
	}

	if err := s.cityRepo.Create(ctx, city); err != nil {
		return nil, err
	}

	return city, nil
}

func (s *cityService) Update(ctx context.Context, id int64, input UpdateCityInput) (*domain.City, error) {
	city, err := s.cityRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if input.Name != nil {
		city.Name = *input.Name
	}
	if input.Slug != nil {
		city.Slug = *input.Slug
	}
	if input.Latitude != nil {
		city.Latitude = *input.Latitude
	}
	if input.Longitude != nil {
		city.Longitude = *input.Longitude
	}

	if err := city.Validate(); err != nil {
		return nil, err
	}

	if err := s.cityRepo.Update(ctx, city); err != nil {
		return nil, err
	}

	return city, nil
}

func (s *cityService) Delete(ctx context.Context, id int64) error {
	return s.cityRepo.Delete(ctx, id)
}
