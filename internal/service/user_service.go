package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/repository"
)

type UpdateUserInput struct {
	Name  *string
	Phone *string
}

type UserService interface {
	GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
	Update(ctx context.Context, id uuid.UUID, input UpdateUserInput) (*domain.User, error)
	// Admin methods:
	List(ctx context.Context, page, pageSize int) (*domain.PaginatedResult[domain.User], error)
	Block(ctx context.Context, id uuid.UUID) error
	Unblock(ctx context.Context, id uuid.UUID) error
}

type userService struct {
	userRepo repository.UserRepository
}

func NewUserService(userRepo repository.UserRepository) UserService {
	return &userService{userRepo: userRepo}
}

func (s *userService) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	return s.userRepo.GetByID(ctx, id)
}

func (s *userService) Update(ctx context.Context, id uuid.UUID, input UpdateUserInput) (*domain.User, error) {
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if input.Name != nil {
		user.Name = *input.Name
	}
	if input.Phone != nil {
		user.Phone = *input.Phone
	}
	user.UpdatedAt = time.Now()

	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *userService) List(ctx context.Context, page, pageSize int) (*domain.PaginatedResult[domain.User], error) {
	return s.userRepo.List(ctx, page, pageSize)
}

func (s *userService) Block(ctx context.Context, id uuid.UUID) error {
	return s.userRepo.SetActive(ctx, id, false)
}

func (s *userService) Unblock(ctx context.Context, id uuid.UUID) error {
	return s.userRepo.SetActive(ctx, id, true)
}
