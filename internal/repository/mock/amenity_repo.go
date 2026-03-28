package mock

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/repository"
)

type AmenityRepo struct {
	mu        sync.RWMutex
	amenities map[uuid.UUID]*domain.Amenity
}

func NewAmenityRepo() repository.AmenityRepository {
	return &AmenityRepo{
		amenities: make(map[uuid.UUID]*domain.Amenity),
	}
}

func (r *AmenityRepo) Create(_ context.Context, amenity *domain.Amenity) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if amenity.ID == uuid.Nil {
		amenity.ID = uuid.New()
	}
	now := time.Now()
	amenity.CreatedAt = now
	amenity.UpdatedAt = now

	cp := *amenity
	r.amenities[amenity.ID] = &cp
	return nil
}

func (r *AmenityRepo) Update(_ context.Context, amenity *domain.Amenity) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	existing, ok := r.amenities[amenity.ID]
	if !ok {
		return domain.ErrNotFound
	}

	amenity.UpdatedAt = time.Now()
	amenity.CreatedAt = existing.CreatedAt
	cp := *amenity
	r.amenities[amenity.ID] = &cp
	return nil
}

func (r *AmenityRepo) Delete(_ context.Context, id uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.amenities[id]; !ok {
		return domain.ErrNotFound
	}
	delete(r.amenities, id)
	return nil
}

func (r *AmenityRepo) GetByID(_ context.Context, id uuid.UUID) (*domain.Amenity, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	a, ok := r.amenities[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	cp := *a
	return &cp, nil
}

func (r *AmenityRepo) ListAll(_ context.Context) ([]domain.Amenity, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []domain.Amenity
	for _, a := range r.amenities {
		cp := *a
		result = append(result, cp)
	}
	return result, nil
}
