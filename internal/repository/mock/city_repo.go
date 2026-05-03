package mock

import (
	"context"
	"sync"

	"github.com/rekurt/relax-hub/internal/domain"
)

// CityRepo is an in-memory mock implementation of repository.CityRepository.
type CityRepo struct {
	mu     sync.RWMutex
	cities map[int64]*domain.City
	nextID int64
}

func NewCityRepo() *CityRepo {
	return &CityRepo{cities: make(map[int64]*domain.City), nextID: 1}
}

func (r *CityRepo) Create(_ context.Context, city *domain.City) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, c := range r.cities {
		if c.Slug == city.Slug {
			return domain.ErrAlreadyExists
		}
	}
	city.ID = r.nextID
	r.nextID++
	cp := *city
	r.cities[city.ID] = &cp
	return nil
}

func (r *CityRepo) GetAll(_ context.Context) ([]domain.City, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]domain.City, 0, len(r.cities))
	for _, c := range r.cities {
		result = append(result, *c)
	}
	return result, nil
}

func (r *CityRepo) GetBySlug(_ context.Context, slug string) (*domain.City, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, c := range r.cities {
		if c.Slug == slug {
			cp := *c
			return &cp, nil
		}
	}
	return nil, domain.ErrNotFound
}

func (r *CityRepo) GetByID(_ context.Context, id int64) (*domain.City, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	c, ok := r.cities[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	cp := *c
	return &cp, nil
}

func (r *CityRepo) Update(_ context.Context, city *domain.City) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.cities[city.ID]; !ok {
		return domain.ErrNotFound
	}
	cp := *city
	r.cities[city.ID] = &cp
	return nil
}

func (r *CityRepo) Delete(_ context.Context, id int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.cities[id]; !ok {
		return domain.ErrNotFound
	}
	delete(r.cities, id)
	return nil
}
