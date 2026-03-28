package mock

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/repository"
)

type ObjectTypeRepo struct {
	mu    sync.RWMutex
	types map[uuid.UUID]*domain.ObjectType
}

func NewObjectTypeRepo() repository.ObjectTypeRepository {
	return &ObjectTypeRepo{
		types: make(map[uuid.UUID]*domain.ObjectType),
	}
}

func (r *ObjectTypeRepo) Create(_ context.Context, objType *domain.ObjectType) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if objType.ID == uuid.Nil {
		objType.ID = uuid.New()
	}
	now := time.Now()
	objType.CreatedAt = now
	objType.UpdatedAt = now

	cp := *objType
	r.types[objType.ID] = &cp
	return nil
}

func (r *ObjectTypeRepo) Update(_ context.Context, objType *domain.ObjectType) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	existing, ok := r.types[objType.ID]
	if !ok {
		return domain.ErrNotFound
	}

	objType.UpdatedAt = time.Now()
	objType.CreatedAt = existing.CreatedAt
	cp := *objType
	r.types[objType.ID] = &cp
	return nil
}

func (r *ObjectTypeRepo) Delete(_ context.Context, id uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.types[id]; !ok {
		return domain.ErrNotFound
	}
	delete(r.types, id)
	return nil
}

func (r *ObjectTypeRepo) GetByID(_ context.Context, id uuid.UUID) (*domain.ObjectType, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	o, ok := r.types[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	cp := *o
	return &cp, nil
}

func (r *ObjectTypeRepo) ListAll(_ context.Context) ([]domain.ObjectType, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []domain.ObjectType
	for _, o := range r.types {
		cp := *o
		result = append(result, cp)
	}
	return result, nil
}
