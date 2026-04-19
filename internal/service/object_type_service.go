package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/rekurt/relax-hub/internal/domain"
	"github.com/rekurt/relax-hub/internal/repository"
)

type ObjectTypeService interface {
	Create(ctx context.Context, objType *domain.ObjectType) error
	Update(ctx context.Context, objType *domain.ObjectType) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.ObjectType, error)
	ListAll(ctx context.Context) ([]domain.ObjectType, error)
}

type objectTypeService struct {
	repo repository.ObjectTypeRepository
}

func NewObjectTypeService(repo repository.ObjectTypeRepository) ObjectTypeService {
	return &objectTypeService{repo: repo}
}

func (s *objectTypeService) Create(ctx context.Context, objType *domain.ObjectType) error {
	if err := objType.Validate(); err != nil {
		return err
	}
	return s.repo.Create(ctx, objType)
}

func (s *objectTypeService) Update(ctx context.Context, objType *domain.ObjectType) error {
	if err := objType.Validate(); err != nil {
		return err
	}
	return s.repo.Update(ctx, objType)
}

func (s *objectTypeService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}

func (s *objectTypeService) GetByID(ctx context.Context, id uuid.UUID) (*domain.ObjectType, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *objectTypeService) ListAll(ctx context.Context) ([]domain.ObjectType, error) {
	return s.repo.ListAll(ctx)
}
