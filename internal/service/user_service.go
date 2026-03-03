package service

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"strings"

	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/logger"
	"github.com/nikitaaldaev/bani/internal/repository"
	"github.com/nikitaaldaev/bani/internal/storage"
)

const maxBioLength = 1000

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
	GetMyStats(ctx context.Context, userID uuid.UUID) (*MyStatsOutput, error)
	// Admin methods:
	List(ctx context.Context, page, pageSize int) (*domain.PaginatedResult[domain.User], error)
	Block(ctx context.Context, id uuid.UUID) error
	Unblock(ctx context.Context, id uuid.UUID) error
}

type MyStatsOutput struct {
	TotalVisits int     `json:"total_visits"`
	TotalSpent  int64   `json:"total_spent"`
	AvgCheck    int64   `json:"avg_check"`
	ReviewCount int     `json:"review_count"`
	AvgRating   float64 `json:"avg_rating"`
}

type userService struct {
	userRepo    repository.UserRepository
	bookingRepo repository.BookingRepository
	reviewRepo  repository.ReviewRepository
	storage     storage.FileStorage
	log         *logger.Logger
}

func NewUserService(userRepo repository.UserRepository, bookingRepo repository.BookingRepository, reviewRepo repository.ReviewRepository, fileStorage storage.FileStorage, log *logger.Logger) UserService {
	return &userService{userRepo: userRepo, bookingRepo: bookingRepo, reviewRepo: reviewRepo, storage: fileStorage, log: log}
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
		if len(*input.Bio) > maxBioLength {
			return nil, domain.ErrInvalidInput
		}
		user.Bio = *input.Bio
	}
	if input.CityID != nil {
		user.CityID = *input.CityID
	}

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
		oldKey := extractS3Key(user.AvatarURL)
		if err := s.storage.Delete(ctx, oldKey); err != nil {
			s.log.Warn("failed to delete old avatar", "key", oldKey, "error", err)
		}
	}

	filename := fmt.Sprintf("avatars/%s%s", uuid.New().String(), input.Ext)
	avatarURL, err := s.storage.Upload(ctx, filename, input.Data, input.ContentType)
	if err != nil {
		return nil, fmt.Errorf("upload avatar: %w", err)
	}

	user.AvatarURL = avatarURL

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
		oldKey := extractS3Key(user.AvatarURL)
		if err := s.storage.Delete(ctx, oldKey); err != nil {
			s.log.Warn("failed to delete avatar", "key", oldKey, "error", err)
		}
	}

	user.AvatarURL = ""

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

func (s *userService) GetMyStats(ctx context.Context, userID uuid.UUID) (*MyStatsOutput, error) {
	bookingStats, err := s.bookingRepo.GetUserStats(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get booking stats: %w", err)
	}

	reviewStats, err := s.reviewRepo.GetUserReviewStats(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get review stats: %w", err)
	}

	return &MyStatsOutput{
		TotalVisits: bookingStats.TotalVisits,
		TotalSpent:  bookingStats.TotalSpent,
		AvgCheck:    bookingStats.AvgCheck,
		ReviewCount: reviewStats.ReviewCount,
		AvgRating:   reviewStats.AvgRating,
	}, nil
}

// extractS3Key extracts the S3 object key from a full URL.
// URL format: scheme://host/bucket/key -> returns "key"
func extractS3Key(avatarURL string) string {
	u, err := url.Parse(avatarURL)
	if err != nil {
		return avatarURL
	}
	// Path is /bucket/key, trim leading / and split on first /
	parts := strings.SplitN(strings.TrimPrefix(u.Path, "/"), "/", 2)
	if len(parts) == 2 {
		return parts[1]
	}
	return avatarURL
}
