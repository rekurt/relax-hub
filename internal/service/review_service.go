package service

import (
	"context"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/repository"
)

type CreateReviewInput struct {
	BookingID uuid.UUID
	Rating    int
	Text      string
}

type ReviewService interface {
	Create(ctx context.Context, userID uuid.UUID, input CreateReviewInput) (*domain.Review, error)
	ListByBathhouse(ctx context.Context, bathhouseID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.Review], error)
}

type reviewService struct {
	reviewRepo  repository.ReviewRepository
	bookingRepo repository.BookingRepository
	bhRepo      repository.BathhouseRepository
}

func NewReviewService(
	reviewRepo repository.ReviewRepository,
	bookingRepo repository.BookingRepository,
	bhRepo repository.BathhouseRepository,
) ReviewService {
	return &reviewService{
		reviewRepo:  reviewRepo,
		bookingRepo: bookingRepo,
		bhRepo:      bhRepo,
	}
}

func (s *reviewService) Create(ctx context.Context, userID uuid.UUID, input CreateReviewInput) (*domain.Review, error) {
	// Verify booking exists and belongs to user
	booking, err := s.bookingRepo.GetByID(ctx, input.BookingID)
	if err != nil {
		return nil, err
	}

	if booking.UserID != userID {
		return nil, domain.ErrForbidden
	}

	// Only completed bookings can be reviewed
	if booking.Status != domain.BookingCompleted {
		return nil, domain.ErrInvalidInput
	}

	// Check if review already exists for this booking
	existing, _ := s.reviewRepo.GetByBookingID(ctx, input.BookingID)
	if existing != nil {
		return nil, domain.ErrAlreadyExists
	}

	review := &domain.Review{
		ID:          uuid.New(),
		UserID:      userID,
		BathhouseID: booking.BathhouseID,
		BookingID:   input.BookingID,
		Rating:      input.Rating,
		Text:        input.Text,
		CreatedAt:   time.Now(),
	}

	if err := review.Validate(); err != nil {
		return nil, err
	}

	if err := s.reviewRepo.Create(ctx, review); err != nil {
		return nil, err
	}

	// Update bathhouse rating (non-fatal: review is already saved)
	if err := s.bhRepo.UpdateRating(ctx, booking.BathhouseID); err != nil {
		log.Printf("WARNING: failed to update bathhouse rating for %s: %v", booking.BathhouseID, err)
	}

	return review, nil
}

func (s *reviewService) ListByBathhouse(ctx context.Context, bathhouseID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.Review], error) {
	return s.reviewRepo.ListByBathhouse(ctx, bathhouseID, page, pageSize)
}
