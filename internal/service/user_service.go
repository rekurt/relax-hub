package service

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/rekurt/relax-hub/internal/domain"
	"github.com/rekurt/relax-hub/internal/logger"
	"github.com/rekurt/relax-hub/internal/repository"
	"github.com/rekurt/relax-hub/internal/storage"
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

type ProfileCompletenessOutput struct {
	Percentage int                              `json:"percentage"`
	Items      []domain.ProfileCompletenessItem `json:"items"`
}

type UserService interface {
	GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
	Update(ctx context.Context, id uuid.UUID, input UpdateUserInput) (*domain.User, error)
	UploadAvatar(ctx context.Context, userID uuid.UUID, input UploadAvatarInput) (*domain.User, error)
	DeleteAvatar(ctx context.Context, userID uuid.UUID) (*domain.User, error)
	GetPublicProfile(ctx context.Context, id uuid.UUID) (*domain.UserProfile, error)
	GetMyStats(ctx context.Context, userID uuid.UUID) (*MyStatsOutput, error)
	GetProfileCompleteness(ctx context.Context, userID uuid.UUID) (*ProfileCompletenessOutput, error)
	CompleteOnboarding(ctx context.Context, userID uuid.UUID) error
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
	recRepo     repository.RecommendationRepository
	notifRepo   repository.NotificationRepository
	storage     storage.FileStorage
	log         *logger.Logger
}

func NewUserService(userRepo repository.UserRepository, bookingRepo repository.BookingRepository, reviewRepo repository.ReviewRepository, recRepo repository.RecommendationRepository, notifRepo repository.NotificationRepository, fileStorage storage.FileStorage, log *logger.Logger) UserService {
	return &userService{userRepo: userRepo, bookingRepo: bookingRepo, reviewRepo: reviewRepo, recRepo: recRepo, notifRepo: notifRepo, storage: fileStorage, log: log}
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
		if utf8.RuneCountInString(*input.Bio) > maxBioLength {
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

	// Resize avatar to required dimensions (200x200 and 50x50)
	resized, err := storage.ResizeAvatar(input.Data)
	if err != nil {
		return nil, fmt.Errorf("resize avatar: %w", err)
	}

	// Upload full-size avatar (200x200)
	baseID := uuid.New().String()
	fullFilename := fmt.Sprintf("avatars/%s_full%s", baseID, input.Ext)
	avatarURL, err := s.storage.Upload(ctx, fullFilename, resized[0].Data, input.ContentType)
	if err != nil {
		return nil, fmt.Errorf("upload avatar: %w", err)
	}

	// Upload thumbnail (50x50)
	thumbFilename := fmt.Sprintf("avatars/%s_thumb%s", baseID, input.Ext)
	_, err = s.storage.Upload(ctx, thumbFilename, resized[1].Data, input.ContentType)
	if err != nil {
		// Log warning but continue - full avatar uploaded successfully
		s.log.Warn("failed to upload avatar thumbnail", "filename", thumbFilename, "error", err)
	}

	oldAvatarURL := user.AvatarURL
	user.AvatarURL = avatarURL

	if err := s.userRepo.Update(ctx, user); err != nil {
		// Clean up newly uploaded file since DB update failed
		if delErr := s.storage.Delete(ctx, fullFilename); delErr != nil {
			s.log.Warn("failed to clean up avatar after db error", "key", fullFilename, "error", delErr)
		}
		return nil, err
	}

	// Delete old avatar only after successful DB update
	if oldAvatarURL != "" {
		oldKey := extractS3Key(oldAvatarURL)
		if err := s.storage.Delete(ctx, oldKey); err != nil {
			s.log.Warn("failed to delete old avatar", "key", oldKey, "error", err)
		}
	}

	return user, nil
}

func (s *userService) DeleteAvatar(ctx context.Context, userID uuid.UUID) (*domain.User, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	oldAvatarURL := user.AvatarURL
	user.AvatarURL = ""

	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, err
	}

	// Delete from storage only after successful DB update
	if oldAvatarURL != "" {
		oldKey := extractS3Key(oldAvatarURL)
		if err := s.storage.Delete(ctx, oldKey); err != nil {
			s.log.Warn("failed to delete avatar from storage", "key", oldKey, "error", err)
		}
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

func (s *userService) GetProfileCompleteness(ctx context.Context, userID uuid.UUID) (*ProfileCompletenessOutput, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	hasPreferences := false
	prefs, err := s.recRepo.GetUserPreferences(ctx, userID)
	if err == nil && prefs != nil {
		hasPreferences = prefs.PreferPool || prefs.PreferSauna || prefs.PreferSteamRoom ||
			prefs.PreferHotTub || prefs.PreferBBQ || prefs.PreferKaraoke ||
			prefs.PreferredCityID != nil || prefs.PriceRangeMin != nil || prefs.PriceRangeMax != nil
	}

	hasNotificationSettings := false
	eventPrefs, err := s.notifRepo.GetEventPreferences(ctx, userID)
	if err == nil && len(eventPrefs) > 0 {
		hasNotificationSettings = true
	}

	pct, items := user.ProfileCompleteness(hasPreferences, hasNotificationSettings)
	return &ProfileCompletenessOutput{Percentage: pct, Items: items}, nil
}

func (s *userService) CompleteOnboarding(ctx context.Context, userID uuid.UUID) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}
	user.OnboardingCompleted = true
	return s.userRepo.Update(ctx, user)
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
