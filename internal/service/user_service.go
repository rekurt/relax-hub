package service

import (
	"context"
	"fmt"
	"io"
	"path"
	"time"

	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/repository"
	"github.com/nikitaaldaev/bani/internal/storage"
)

type UpdateUserInput struct {
	Name   *string
	Phone  *string
	Bio    *string
	CityID **int64 // double pointer: nil = not provided, *nil = clear, *val = set
}

type UploadAvatarInput struct {
	Data        io.Reader
	ContentType string
	Ext         string // e.g. ".jpg", ".png"
}

type UserService interface {
	GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
	Update(ctx context.Context, id uuid.UUID, input UpdateUserInput) (*domain.User, error)
	UploadAvatar(ctx context.Context, userID uuid.UUID, input UploadAvatarInput) (*domain.User, error)
	DeleteAvatar(ctx context.Context, userID uuid.UUID) (*domain.User, error)
	GetPublicProfile(ctx context.Context, id uuid.UUID) (*domain.UserProfile, error)
	// Admin methods:
	List(ctx context.Context, page, pageSize int) (*domain.PaginatedResult[domain.User], error)
	Block(ctx context.Context, id uuid.UUID) error
	Unblock(ctx context.Context, id uuid.UUID) error
}

type userService struct {
	userRepo repository.UserRepository
	storage  storage.FileStorage
}

func NewUserService(userRepo repository.UserRepository, fileStorage storage.FileStorage) UserService {
	return &userService{userRepo: userRepo, storage: fileStorage}
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
		if *input.Name == "" {
			return nil, domain.ErrInvalidInput
		}
		user.Name = *input.Name
	}
	if input.Phone != nil {
		user.Phone = *input.Phone
	}
	if input.Bio != nil {
		user.Bio = *input.Bio
	}
	if input.CityID != nil {
		user.CityID = *input.CityID
	}
	user.UpdatedAt = time.Now()

	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *userService) UploadAvatar(ctx context.Context, userID uuid.UUID, input UploadAvatarInput) (*domain.User, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Delete old avatar if exists
	if user.AvatarURL != "" {
		oldFilename := path.Base(user.AvatarURL)
		_ = s.storage.Delete(ctx, oldFilename)
	}

	filename := fmt.Sprintf("avatars/%s%s", uuid.New().String(), input.Ext)
	url, err := s.storage.Upload(ctx, filename, input.Data, input.ContentType)
	if err != nil {
		return nil, fmt.Errorf("upload avatar: %w", err)
	}

	user.AvatarURL = url
	user.UpdatedAt = time.Now()

	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *userService) DeleteAvatar(ctx context.Context, userID uuid.UUID) (*domain.User, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if user.AvatarURL != "" {
		oldFilename := path.Base(user.AvatarURL)
		_ = s.storage.Delete(ctx, oldFilename)
	}

	user.AvatarURL = ""
	user.UpdatedAt = time.Now()

	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *userService) GetPublicProfile(ctx context.Context, id uuid.UUID) (*domain.UserProfile, error) {
	return s.userRepo.GetPublicProfile(ctx, id)
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
