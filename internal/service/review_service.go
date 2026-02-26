package service

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/repository"
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
}

type reviewService struct {
	reviewRepo    repository.ReviewRepository
	bookingRepo   repository.BookingRepository
	bhRepo        repository.BathhouseRepository
	accessChecker *AccessChecker
}

func NewReviewService(
	reviewRepo repository.ReviewRepository,
	bookingRepo repository.BookingRepository,
	bhRepo repository.BathhouseRepository,
	accessChecker *AccessChecker,
) ReviewService {
	return &reviewService{
		reviewRepo:    reviewRepo,
		bookingRepo:   bookingRepo,
		bhRepo:        bhRepo,
		accessChecker: accessChecker,
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
		Status:      domain.ReviewStatusPending,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := review.Validate(); err != nil {
		return nil, err
	}

	if err := s.reviewRepo.Create(ctx, review); err != nil {
		return nil, err
	}

	if err := s.bhRepo.UpdateRating(ctx, booking.BathhouseID); err != nil {
		log.Printf("WARNING: failed to update bathhouse rating for %s: %v", booking.BathhouseID, err)
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
			log.Printf("WARNING: failed to update bathhouse rating for %s: %v", review.BathhouseID, err)
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

	if err := s.bhRepo.UpdateRating(ctx, review.BathhouseID); err != nil {
		log.Printf("WARNING: failed to update bathhouse rating for %s: %v", review.BathhouseID, err)
	}

	return nil
}

func (s *reviewService) ListByBathhouse(ctx context.Context, bathhouseID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.Review], error) {
	return s.reviewRepo.ListByBathhouse(ctx, bathhouseID, page, pageSize)
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

	return review, nil
}
