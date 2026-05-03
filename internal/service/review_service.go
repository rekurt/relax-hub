package service

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/rekurt/relax-hub/internal/domain"
	"github.com/rekurt/relax-hub/internal/logger"
	"github.com/rekurt/relax-hub/internal/moderation"
	"github.com/rekurt/relax-hub/internal/repository"
	"github.com/rekurt/relax-hub/internal/storage"
)

type CreateReviewInput struct {
	BookingID     uuid.UUID
	BathhouseID   uuid.UUID
	Rating        int
	Cleanliness   *float64
	Accuracy      *float64
	Communication *float64
	ValueForMoney *float64
	Text          string
}

type UpdateReviewInput struct {
	Rating        *int
	Cleanliness   *float64
	Accuracy      *float64
	Communication *float64
	ValueForMoney *float64
	Text          *string
	Images        []string
}

type ReviewService interface {
	Create(ctx context.Context, userID uuid.UUID, input CreateReviewInput) (*domain.Review, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Review, error)
	Update(ctx context.Context, userID uuid.UUID, reviewID uuid.UUID, input UpdateReviewInput) (*domain.Review, error)
	Delete(ctx context.Context, userID uuid.UUID, userRole domain.UserRole, reviewID uuid.UUID) error
	ListByBathhouse(ctx context.Context, bathhouseID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.Review], error)
	AddOwnerResponse(ctx context.Context, userID uuid.UUID, userRole domain.UserRole, reviewID uuid.UUID, response string) (*domain.Review, error)
	GetCriteriaAverages(ctx context.Context, bathhouseID uuid.UUID) (*domain.ReviewCriteriaAverages, error)
	RecalculateBayesianRating(ctx context.Context, bathhouseID uuid.UUID) error
	RefreshPlatformAverage(ctx context.Context) error
	// Auto review requests:
	SendReviewRequests(ctx context.Context, delayHours int) (int, error)
	// Admin moderation:
	ListAllReviews(ctx context.Context, filter domain.AdminReviewFilter) (*domain.PaginatedResult[domain.Review], error)
	CountPendingReviews(ctx context.Context) (int64, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status domain.ReviewStatus) error
	UpdateStatusWithReasons(ctx context.Context, id uuid.UUID, status domain.ReviewStatus, reasons []string) error
}

const (
	bayesianMinReviews        = 5 // m: minimum reviews threshold
	bayesianDisplayMinReviews = 3 // show Bayesian rating only when >= 3 reviews
	redisPlatformAvgKey       = "platform:avg_rating"
	redisPlatformAvgTTL       = 24 * time.Hour
	qualityMinReviews         = 10  // minimum reviews before quality thresholds apply
	qualityWarningThreshold   = 3.0 // rating below this triggers owner warning
	qualityDepublishThreshold = 2.0 // rating below this triggers auto-depublish
)

type reviewService struct {
	reviewRepo       repository.ReviewRepository
	clientReviewRepo repository.ClientReviewRepository
	bookingRepo      repository.BookingRepository
	bhRepo           repository.BathhouseRepository
	mediaRepo        repository.MediaRepository
	fileStorage      storage.FileStorage
	accessChecker    *AccessChecker
	notifSvc         NotificationService
	contentFilter    *moderation.ContentFilter
	textModerator    moderation.TextModerationService
	redisClient      *redis.Client
	logger           *logger.Logger
}

func NewReviewService(
	reviewRepo repository.ReviewRepository,
	clientReviewRepo repository.ClientReviewRepository,
	bookingRepo repository.BookingRepository,
	bhRepo repository.BathhouseRepository,
	mediaRepo repository.MediaRepository,
	fileStorage storage.FileStorage,
	accessChecker *AccessChecker,
	notifSvc NotificationService,
	contentFilter *moderation.ContentFilter,
	textModerator moderation.TextModerationService,
	redisClient *redis.Client,
	log *logger.Logger,
) ReviewService {
	return &reviewService{
		reviewRepo:       reviewRepo,
		clientReviewRepo: clientReviewRepo,
		bookingRepo:      bookingRepo,
		bhRepo:           bhRepo,
		mediaRepo:        mediaRepo,
		fileStorage:      fileStorage,
		accessChecker:    accessChecker,
		notifSvc:         notifSvc,
		contentFilter:    contentFilter,
		textModerator:    textModerator,
		redisClient:      redisClient,
		logger:           log,
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
	revealAt := now.Add(domain.ClientReviewBlindDays * 24 * time.Hour)
	review := &domain.Review{
		ID:            uuid.New(),
		UserID:        userID,
		BathhouseID:   booking.BathhouseID,
		BookingID:     input.BookingID,
		Rating:        input.Rating,
		Cleanliness:   input.Cleanliness,
		Accuracy:      input.Accuracy,
		Communication: input.Communication,
		ValueForMoney: input.ValueForMoney,
		Text:          input.Text,
		RevealAt:      &revealAt,
		IsRevealed:    false,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	// If criteria are provided, compute the overall rating from them
	review.ComputeOverallRating()

	// Apply text moderation (NLP auto-moderation)
	if s.textModerator != nil && s.contentFilter.IsEnabled() {
		modResult := s.textModerator.Analyze(ctx, input.Text)
		score := modResult.Score
		review.ModerationScore = &score
		review.ModerationFlags = modResult.Flags

		if modResult.Flagged {
			// Score >= 0.7: needs manual review
			review.Status = domain.ReviewStatusPending
		} else {
			review.Status = domain.ReviewStatusApproved
		}
	} else if s.contentFilter.IsEnabled() {
		// Fallback to legacy content filter (no text moderator)
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

	// Only update ratings for approved reviews; pending/rejected reviews should not affect scores
	if review.Status == domain.ReviewStatusApproved {
		if err := s.bhRepo.UpdateRating(ctx, booking.BathhouseID); err != nil {
			s.logger.Warn("Failed to update bathhouse rating", "bathhouse_id", booking.BathhouseID, "error", err)
		}
		if err := s.RecalculateBayesianRating(ctx, booking.BathhouseID); err != nil {
			s.logger.Warn("Failed to recalculate bayesian rating", "bathhouse_id", booking.BathhouseID, "error", err)
		}
		s.checkQualityThresholds(ctx, booking.BathhouseID)
	}

	// Check if owner already posted a client review — if so, reveal both
	if review.Status == domain.ReviewStatusApproved {
		s.tryMutualRevealFromGuest(ctx, review)
	}

	// Notify bathhouse owner about new review (only for approved reviews)
	if review.Status == domain.ReviewStatusApproved {
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
	if input.Cleanliness != nil {
		review.Cleanliness = input.Cleanliness
	}
	if input.Accuracy != nil {
		review.Accuracy = input.Accuracy
	}
	if input.Communication != nil {
		review.Communication = input.Communication
	}
	if input.ValueForMoney != nil {
		review.ValueForMoney = input.ValueForMoney
	}
	if input.Text != nil {
		review.Text = *input.Text

		// Re-run moderation on updated text — always require re-moderation on text change
		review.Status = domain.ReviewStatusPending
		if s.textModerator != nil && s.contentFilter.IsEnabled() {
			modResult := s.textModerator.Analyze(ctx, review.Text)
			score := modResult.Score
			review.ModerationScore = &score
			review.ModerationFlags = modResult.Flags
		}
	}
	if input.Images != nil {
		review.Images = input.Images
	}

	// Recompute overall rating from criteria if they are present
	review.ComputeOverallRating()

	if err := review.Validate(); err != nil {
		return nil, err
	}

	review.UpdatedAt = time.Now()
	if err := s.reviewRepo.Update(ctx, review); err != nil {
		return nil, err
	}

	criteriaChanged := input.Cleanliness != nil || input.Accuracy != nil || input.Communication != nil || input.ValueForMoney != nil
	if (input.Rating != nil || criteriaChanged) && review.Status == domain.ReviewStatusApproved {
		if err := s.bhRepo.UpdateRating(ctx, review.BathhouseID); err != nil {
			s.logger.Warn("Failed to update bathhouse rating", "bathhouse_id", review.BathhouseID, "error", err)
		}
		if err := s.RecalculateBayesianRating(ctx, review.BathhouseID); err != nil {
			s.logger.Warn("Failed to recalculate bayesian rating", "bathhouse_id", review.BathhouseID, "error", err)
		}
		s.checkQualityThresholds(ctx, review.BathhouseID)
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
	if err := s.RecalculateBayesianRating(ctx, review.BathhouseID); err != nil {
		s.logger.Warn("Failed to recalculate bayesian rating", "bathhouse_id", review.BathhouseID, "error", err)
	}
	s.checkQualityThresholds(ctx, review.BathhouseID)

	return nil
}

func (s *reviewService) ListByBathhouse(ctx context.Context, bathhouseID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.Review], error) {
	approvedStatus := domain.ReviewStatusApproved
	return s.reviewRepo.ListByBathhouseFiltered(ctx, domain.ReviewFilter{
		BathhouseID:  &bathhouseID,
		Status:       &approvedStatus,
		OnlyRevealed: true,
		Page:         page,
		PageSize:     pageSize,
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

// tryMutualRevealFromGuest checks if the owner has also reviewed the client. If both exist, reveal both.
func (s *reviewService) tryMutualRevealFromGuest(ctx context.Context, guestReview *domain.Review) {
	if s.clientReviewRepo == nil {
		return
	}

	clientReview, err := s.clientReviewRepo.GetByBookingID(ctx, guestReview.BookingID)
	if err != nil {
		// Owner hasn't reviewed yet, keep both in blind period
		return
	}

	// Both reviews exist — reveal both
	// Reveal the guest review
	guestReview.IsRevealed = true
	guestReview.UpdatedAt = time.Now()
	if err := s.reviewRepo.Update(ctx, guestReview); err != nil {
		s.logger.Warn("failed to reveal guest review on mutual", "id", guestReview.ID, "error", err)
	}

	// Reveal the client review
	if !clientReview.IsRevealed {
		if err := s.clientReviewRepo.RevealByID(ctx, clientReview.ID); err != nil {
			s.logger.Warn("failed to reveal client review on mutual", "id", clientReview.ID, "error", err)
		}
	}
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

func (s *reviewService) GetCriteriaAverages(ctx context.Context, bathhouseID uuid.UUID) (*domain.ReviewCriteriaAverages, error) {
	return s.reviewRepo.GetCriteriaAverages(ctx, bathhouseID)
}

func (s *reviewService) RecalculateBayesianRating(ctx context.Context, bathhouseID uuid.UUID) error {
	bh, err := s.bhRepo.GetByID(ctx, bathhouseID)
	if err != nil {
		return fmt.Errorf("get bathhouse for bayesian: %w", err)
	}

	n := float64(bh.ReviewCount)
	R := bh.Rating

	// Get platform average from Redis cache, fallback to DB
	C := s.getPlatformAverage(ctx)

	m := float64(bayesianMinReviews)
	bayesian := 0.0
	if n+m > 0 {
		bayesian = (n*R + m*C) / (n + m)
		bayesian = math.Round(bayesian*10) / 10 // round to 1 decimal
	}

	return s.bhRepo.UpdateBayesianRating(ctx, bathhouseID, bayesian)
}

func (s *reviewService) checkQualityThresholds(ctx context.Context, bathhouseID uuid.UUID) {
	bh, err := s.bhRepo.GetByID(ctx, bathhouseID)
	if err != nil {
		s.logger.Warn("failed to get bathhouse for quality check", "bathhouse_id", bathhouseID, "error", err)
		return
	}

	if bh.ReviewCount < qualityMinReviews {
		return
	}

	if bh.Status != domain.BathhouseStatusActive {
		return
	}

	data := map[string]string{
		"bathhouse_id":   bathhouseID.String(),
		"bathhouse_name": bh.Name,
		"rating":         fmt.Sprintf("%.1f", bh.Rating),
		"review_count":   strconv.Itoa(bh.ReviewCount),
	}

	if bh.Rating < qualityDepublishThreshold {
		// Auto-depublish: set status to inactive
		if err := s.bhRepo.UpdateStatus(ctx, bathhouseID, domain.BathhouseStatusInactive); err != nil {
			s.logger.Error("failed to auto-depublish bathhouse", "bathhouse_id", bathhouseID, "error", err)
			return
		}
		// Notify owner
		if err := s.notifSvc.Send(ctx, bh.OwnerID, domain.NotifBathhouseDepublished,
			"Баня снята с публикации",
			fmt.Sprintf("Ваша баня «%s» была автоматически снята с публикации из-за низкого рейтинга (%.1f)", bh.Name, bh.Rating),
			data,
		); err != nil {
			s.logger.Warn("failed to send depublish notification", "bathhouse_id", bathhouseID, "error", err)
		}
		s.logger.Info("bathhouse auto-depublished due to low rating",
			"bathhouse_id", bathhouseID, "rating", bh.Rating, "review_count", bh.ReviewCount)
	} else if bh.Rating < qualityWarningThreshold {
		// Warn owner
		if err := s.notifSvc.Send(ctx, bh.OwnerID, domain.NotifLowRatingWarning,
			"Низкий рейтинг",
			fmt.Sprintf("Рейтинг вашей бани «%s» упал до %.1f. Улучшите качество сервиса, чтобы избежать снятия с публикации.", bh.Name, bh.Rating),
			data,
		); err != nil {
			s.logger.Warn("failed to send low rating warning", "bathhouse_id", bathhouseID, "error", err)
		}
	}
}

func (s *reviewService) getPlatformAverage(ctx context.Context) float64 {
	if s.redisClient != nil {
		val, err := s.redisClient.Get(ctx, redisPlatformAvgKey).Result()
		if err == nil {
			if avg, parseErr := strconv.ParseFloat(val, 64); parseErr == nil {
				return avg
			}
		}
	}

	// Fallback: compute from DB
	avg, err := s.reviewRepo.GetPlatformAverageRating(ctx)
	if err != nil {
		s.logger.Warn("failed to get platform average rating from DB", "error", err)
		return 3.5 // sensible default
	}

	// Cache in Redis
	if s.redisClient != nil {
		if err := s.redisClient.Set(ctx, redisPlatformAvgKey, strconv.FormatFloat(avg, 'f', 2, 64), redisPlatformAvgTTL).Err(); err != nil {
			s.logger.Warn("failed to cache platform average in Redis", "error", err)
		}
	}

	return avg
}

func (s *reviewService) RefreshPlatformAverage(ctx context.Context) error {
	avg, err := s.reviewRepo.GetPlatformAverageRating(ctx)
	if err != nil {
		return fmt.Errorf("get platform average rating: %w", err)
	}

	if s.redisClient != nil {
		if err := s.redisClient.Set(ctx, redisPlatformAvgKey, strconv.FormatFloat(avg, 'f', 2, 64), redisPlatformAvgTTL).Err(); err != nil {
			return fmt.Errorf("cache platform average: %w", err)
		}
	}

	return nil
}

func (s *reviewService) ListAllReviews(ctx context.Context, filter domain.AdminReviewFilter) (*domain.PaginatedResult[domain.Review], error) {
	return s.reviewRepo.ListAllReviews(ctx, filter)
}

func (s *reviewService) CountPendingReviews(ctx context.Context) (int64, error) {
	return s.reviewRepo.CountPendingReviews(ctx)
}

func (s *reviewService) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.ReviewStatus) error {
	// Get the review first so we can recalculate ratings if approving
	review, err := s.reviewRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if err := s.reviewRepo.UpdateStatus(ctx, id, status); err != nil {
		return err
	}

	// When a pending review is approved, recalculate bathhouse ratings
	if status == domain.ReviewStatusApproved && review.Status != domain.ReviewStatusApproved {
		if err := s.bhRepo.UpdateRating(ctx, review.BathhouseID); err != nil {
			s.logger.Warn("Failed to update bathhouse rating after approval", "bathhouse_id", review.BathhouseID, "error", err)
		}
		if err := s.RecalculateBayesianRating(ctx, review.BathhouseID); err != nil {
			s.logger.Warn("Failed to recalculate bayesian rating after approval", "bathhouse_id", review.BathhouseID, "error", err)
		}
		s.checkQualityThresholds(ctx, review.BathhouseID)
	}

	return nil
}

func (s *reviewService) UpdateStatusWithReasons(ctx context.Context, id uuid.UUID, status domain.ReviewStatus, reasons []string) error {
	return s.reviewRepo.UpdateStatusWithReasons(ctx, id, status, reasons)
}

func (s *reviewService) SendReviewRequests(ctx context.Context, delayHours int) (int, error) {
	if delayHours <= 0 {
		delayHours = 2
	}
	checkedOutBefore := time.Now().Add(-time.Duration(delayHours) * time.Hour)

	bookings, err := s.bookingRepo.ListCompletedForReviewRequests(ctx, checkedOutBefore)
	if err != nil {
		return 0, fmt.Errorf("list completed bookings for review requests: %w", err)
	}

	sent := 0
	for _, b := range bookings {
		redisKey := fmt.Sprintf("review_request_sent:%s", b.ID.String())

		if s.redisClient != nil {
			wasSet, err := s.redisClient.SetArgs(ctx, redisKey, "1", redis.SetArgs{
				Mode: "NX",
				TTL:  30 * 24 * time.Hour, // 30 days TTL
			}).Result()
			if err != nil && err != redis.Nil {
				s.logger.Warn("Redis SET NX failed for review request dedup, sending anyway",
					"key", redisKey, "error", err)
			} else if wasSet != "OK" {
				continue // already sent
			}
		}

		bh, err := s.bhRepo.GetByID(ctx, b.BathhouseID)
		if err != nil {
			s.logger.Warn("failed to get bathhouse for review request",
				"booking_id", b.ID, "bathhouse_id", b.BathhouseID, "error", err)
			continue
		}

		body := fmt.Sprintf("Как вам визит в %s? Поделитесь впечатлениями — это поможет другим гостям!", bh.Name)
		data := map[string]string{
			"booking_id":     b.ID.String(),
			"bathhouse_id":   b.BathhouseID.String(),
			"bathhouse_name": bh.Name,
		}
		if err := s.notifSvc.Send(ctx, b.UserID, domain.NotifReviewRequest,
			"Оставьте отзыв", body, data); err != nil {
			s.logger.Warn("failed to send review request notification",
				"booking_id", b.ID, "user_id", b.UserID, "error", err)
			continue
		}
		sent++
	}
	return sent, nil
}
