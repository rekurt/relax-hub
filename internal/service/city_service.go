package service

import (
	"context"
	"sort"
	"strings"

	"github.com/rekurt/relax-hub/internal/domain"
	"github.com/rekurt/relax-hub/internal/repository"
)

type CreateCityInput struct {
	Name      string
	Slug      string
	Region    string
	Latitude  float64
	Longitude float64
}

type UpdateCityInput struct {
	Name      *string
	Slug      *string
	Region    *string
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
	cities, err := s.cityRepo.GetAll(ctx)
	if err != nil {
		return nil, err
	}
	return dedupeCities(cities), nil
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
		Region:    input.Region,
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
	if input.Region != nil {
		city.Region = *input.Region
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

func dedupeCities(cities []domain.City) []domain.City {
	if len(cities) <= 1 {
		return cities
	}

	grouped := make(map[string]domain.City, len(cities))
	for _, city := range cities {
		key := normalizeCityName(city.Name)
		if key == "" {
			key = strings.ToLower(strings.TrimSpace(city.Slug))
		}

		existing, exists := grouped[key]
		if !exists || cityPriority(city) > cityPriority(existing) {
			grouped[key] = city
		}
	}

	result := make([]domain.City, 0, len(grouped))
	for _, city := range grouped {
		result = append(result, city)
	}

	sort.Slice(result, func(left, right int) bool {
		leftName := strings.ToLower(result[left].Name)
		rightName := strings.ToLower(result[right].Name)
		if leftName == rightName {
			return result[left].ID < result[right].ID
		}
		return leftName < rightName
	})

	return result
}

func normalizeCityName(name string) string {
	return strings.ToLower(strings.Join(strings.Fields(strings.TrimSpace(name)), " "))
}

func cityPriority(city domain.City) int {
	score := 0
	if city.Latitude != 0 || city.Longitude != 0 {
		score += 100
	}
	if region := strings.TrimSpace(strings.ToLower(city.Region)); region != "" && region != "ru" {
		score += 20
	}
	if slug := strings.TrimSpace(strings.ToLower(city.Slug)); slug == "moskva" {
		score += 5
	}
	return score
}
