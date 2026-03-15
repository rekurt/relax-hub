package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/logger"
	"github.com/nikitaaldaev/bani/internal/moderation"
	"github.com/nikitaaldaev/bani/internal/repository"
	"github.com/nikitaaldaev/bani/internal/storage"
)

type CreateReviewInput struct {
	BookingID   uuid.UUID
	BathhouseID uuid.UUID
	Rating      int
	Text        string
}

type UpdateReviewInput struct {
	Rating *int
	Text   *string
	Images []string
}

type ReviewService interface {
	Create(ctx context.Context, userID uuid.UUID, input CreateReviewInput) (*domain.Review, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Review, error)
	Update(ctx context.Context, userID uuid.UUID, reviewID uuid.UUID, input UpdateReviewInput) (*domain.Review, error)
	Delete(ctx context.Context, userID uuid.UUID, userRole domain.UserRole, reviewID uuid.UUID) error
	ListByBathhouse(ctx context.Context, bathhouseID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.Review], error)
	AddOwnerResponse(ctx context.Context, userID uuid.UUID, userRole domain.UserRole, reviewID uuid.UUID, response string) (*domain.Review, error)
	// Admin moderation:
	ListAllReviews(ctx context.Context, filter domain.AdminReviewFilter) (*domain.PaginatedResult[domain.Review], error)
	CountPendingReviews(ctx context.Context) (int64, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status domain.ReviewStatus) error
	UpdateStatusWithReasons(ctx context.Context, id uuid.UUID, status domain.ReviewStatus, reasons []string) error
}

type reviewService struct {
	reviewRepo     repository.ReviewRepository
	bookingRepo    repository.BookingRepository
	bhRepo         repository.BathhouseRepository
	mediaRepo      repository.MediaRepository
	fileStorage    storage.FileStorage
	accessChecker  *AccessChecker
	notifSvc       NotificationService
	contentFilter  *moderation.ContentFilter
	logger         *logger.Logger
}

func NewReviewService(
	reviewRepo repository.ReviewRepository,
	bookingRepo repository.BookingRepository,
	bhRepo repository.BathhouseRepository,
	mediaRepo repository.MediaRepository,
	fileStorage storage.FileStorage,
	accessChecker *AccessChecker,
	notifSvc NotificationService,
	contentFilter *moderation.ContentFilter,
	log *logger.Logger,
) ReviewService {
	return &reviewService{
		reviewRepo:    reviewRepo,
		bookingRepo:   bookingRepo,
		bhRepo:        bhRepo,
		mediaRepo:     mediaRepo,
		fileStorage:   fileStorage,
		accessChecker: accessChecker,
		notifSvc:      notifSvc,
		contentFilter: contentFilter,
		logger:        log,
	}
}

func (s *reviewService) Create(ctx context.Context, userID uuid.UUID, input CreateReviewInput) (*domain.Review, error) {
	booking, err := s.bookingRepo.GetByID(ctx, input.BookingID)
	if err != nil {
		return nil, err
	}

	if booking.UserID != userID {
		return nil, domain.ErrForbidden
	}

	if input.BathhouseID != uuid.Nil && booking.BathhouseID != input.BathhouseID {
		return nil, domain.ErrInvalidInput
	}

	if booking.Status != domain.BookingCompleted {
		return nil, domain.ErrInvalidInput
	}

	existing, err := s.reviewRepo.GetByBookingID(ctx, input.BookingID)
	if err != nil && !errors.Is(err, domain.ErrNotFound) {
		return nil, err
	}
	if existing != nil {
		return nil, domain.ErrAlreadyExists
	}

	now := time.Now()
	review := &domain.Review{
		ID:          uuid.New(),
		UserID:      userID,
		BathhouseID: booking.BathhouseID,
		BookingID:   input.BookingID,
		Rating:      input.Rating,
		Text:        input.Text,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	// Apply content filtering if moderation is enabled
	if s.contentFilter.IsEnabled() {
		filterResult := s.contentFilter.CheckText(input.Text)
		if !filterResult.IsClean {
			review.Status = domain.ReviewStatusRejected
			review.RejectionReasons = filterResult.Reasons
		} else if s.contentFilter.ShouldAutoApprove() {
			review.Status = domain.ReviewStatusApproved
		} else {
			review.Status = domain.ReviewStatusPending
		}
	} else {
		review.Status = domain.ReviewStatusApproved
	}

	if err := review.Validate(); err != nil {
		return nil, err
	}

	if err := s.reviewRepo.Create(ctx, review); err != nil {
		return nil, err
	}

	if err := s.bhRepo.UpdateRating(ctx, booking.BathhouseID); err != nil {
		s.logger.Warn("Failed to update bathhouse rating", "bathhouse_id", booking.BathhouseID, "error", err)
	}

	// Notify bathhouse owner about new review (regardless of status)
	bh, err := s.bhRepo.GetByID(ctx, booking.BathhouseID)
	if err != nil {
		s.logger.Warn("failed to get bathhouse for review notification", "bathhouse_id", booking.BathhouseID, "error", err)
	} else {
		data := map[string]string{
			"review_id":    review.ID.String(),
			"bathhouse_id": booking.BathhouseID.String(),
		}
		notifErr := s.notifSvc.Send(ctx, bh.OwnerID, domain.NotifNewReview,
			"Новый отзыв",
			fmt.Sprintf("Получен новый отзыв с оценкой %d для %s", review.Rating, bh.Name),
			data,
		)
		if notifErr != nil {
			s.logger.Warn("failed to send review notification", "review_id", review.ID, "error", notifErr)
		}
	}

	return review, nil
}

func (s *reviewService) GetByID(ctx context.Context, id uuid.UUID) (*domain.Review, error) {
	return s.reviewRepo.GetByID(ctx, id)
}

func (s *reviewService) Update(ctx context.Context, userID uuid.UUID, reviewID uuid.UUID, input UpdateReviewInput) (*domain.Review, error) {
	review, err := s.reviewRepo.GetByID(ctx, reviewID)
	if err != nil {
		return nil, err
	}

	if review.UserID != userID {
		return nil, domain.ErrForbidden
	}

	if time.Since(review.CreatedAt) > 24*time.Hour {
		return nil, domain.ErrForbidden
	}

	if input.Rating != nil {
		review.Rating = *input.Rating
	}
	if input.Text != nil {
		review.Text = *input.Text
	}
	if input.Images != nil {
		review.Images = input.Images
	}

	if err := review.Validate(); err != nil {
		return nil, err
	}

	review.UpdatedAt = time.Now()
	if err := s.reviewRepo.Update(ctx, review); err != nil {
		return nil, err
	}

	if input.Rating != nil {
		if err := s.bhRepo.UpdateRating(ctx, review.BathhouseID); err != nil {
			s.logger.Warn("Failed to update bathhouse rating", "bathhouse_id", review.BathhouseID, "error", err)
		}
	}

	return review, nil
}

func (s *reviewService) Delete(ctx context.Context, userID uuid.UUID, userRole domain.UserRole, reviewID uuid.UUID) error {
	review, err := s.reviewRepo.GetByID(ctx, reviewID)
	if err != nil {
		return err
	}

	if userRole == domain.RoleAdmin {
		// Admin can delete any review
	} else if review.UserID == userID {
		// Author can delete own review
	} else {
		return domain.ErrForbidden
	}

	if err := s.reviewRepo.Delete(ctx, reviewID); err != nil {
		return err
	}

	// Clean up associated media files
	s.cleanupReviewMedia(ctx, reviewID)

	if err := s.bhRepo.UpdateRating(ctx, review.BathhouseID); err != nil {
		s.logger.Warn("Failed to update bathhouse rating", "bathhouse_id", review.BathhouseID, "error", err)
	}

	return nil
}

func (s *reviewService) ListByBathhouse(ctx context.Context, bathhouseID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.Review], error) {
	approvedStatus := domain.ReviewStatusApproved
	return s.reviewRepo.ListByBathhouseFiltered(ctx, domain.ReviewFilter{
		BathhouseID: &bathhouseID,
		Status:      &approvedStatus,
		Page:        page,
		PageSize:    pageSize,
	})
}

func (s *reviewService) AddOwnerResponse(ctx context.Context, userID uuid.UUID, userRole domain.UserRole, reviewID uuid.UUID, response string) (*domain.Review, error) {
	if response == "" {
		return nil, domain.ErrInvalidInput
	}

	review, err := s.reviewRepo.GetByID(ctx, reviewID)
	if err != nil {
		return nil, err
	}

	if review.OwnerResponse != "" {
		return nil, domain.ErrReviewAlreadyResponded
	}

	if err := s.accessChecker.CanManageBathhouse(ctx, userID, userRole, review.BathhouseID); err != nil {
		return nil, err
	}

	now := time.Now()
	if err := s.reviewRepo.AddOwnerResponse(ctx, reviewID, response, now); err != nil {
		return nil, err
	}

	review.OwnerResponse = response
	review.OwnerResponseAt = &now
	review.UpdatedAt = now

	// Notify review author about owner response
	data := map[string]string{
		"review_id":    review.ID.String(),
		"bathhouse_id": review.BathhouseID.String(),
	}
	if notifErr := s.notifSvc.Send(ctx, review.UserID, domain.NotifReviewResponse,
		"Ответ на ваш отзыв",
		"Владелец бани ответил на ваш отзыв",
		data,
	); notifErr != nil {
		s.logger.Warn("failed to send review response notification", "review_id", review.ID, "error", notifErr)
	}

	return review, nil
}

func (s *reviewService) cleanupReviewMedia(ctx context.Context, reviewID uuid.UUID) {
	result, err := s.mediaRepo.ListByOwner(ctx, domain.MediaOwnerReview, reviewID, 1, domain.MaxImagesPerReview+domain.MaxVideosPerReview)
	if err != nil {
		s.logger.Warn("failed to list review media for cleanup", "review_id", reviewID, "error", err)
		return
	}
	for _, m := range result.Items {
		if err := s.mediaRepo.Delete(ctx, m.ID); err != nil {
			s.logger.Warn("failed to delete media record", "media_id", m.ID, "error", err)
			continue
		}
		if err := s.fileStorage.Delete(ctx, extractStorageKey(m.URL)); err != nil {
			s.logger.Warn("failed to delete media file", "url", m.URL, "error", err)
		}
		if m.ThumbnailURL != "" {
			if err := s.fileStorage.Delete(ctx, extractStorageKey(m.ThumbnailURL)); err != nil {
				s.logger.Warn("failed to delete thumbnail file", "url", m.ThumbnailURL, "error", err)
			}
		}
	}
}

func (s *reviewService) ListAllReviews(ctx context.Context, filter domain.AdminReviewFilter) (*domain.PaginatedResult[domain.Review], error) {
	return s.reviewRepo.ListAllReviews(ctx, filter)
}

func (s *reviewService) CountPendingReviews(ctx context.Context) (int64, error) {
	return s.reviewRepo.CountPendingReviews(ctx)
}

func (s *reviewService) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.ReviewStatus) error {
	return s.reviewRepo.UpdateStatus(ctx, id, status)
}

func (s *reviewService) UpdateStatusWithReasons(ctx context.Context, id uuid.UUID, status domain.ReviewStatus, reasons []string) error {
	return s.reviewRepo.UpdateStatusWithReasons(ctx, id, status, reasons)
}
