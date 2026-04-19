package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/rekurt/relax-hub/internal/domain"
	"github.com/rekurt/relax-hub/internal/logger"
	"github.com/rekurt/relax-hub/internal/repository"
)

type CreateClientReviewInput struct {
	BookingID      uuid.UUID
	Punctuality    *float64
	Cleanliness    *float64
	RuleCompliance *float64
	Rating         int
	Text           string
}

type ClientReviewService interface {
	Create(ctx context.Context, ownerID uuid.UUID, ownerRole domain.UserRole, input CreateClientReviewInput) (*domain.ClientReview, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.ClientReview, error)
	GetByBookingID(ctx context.Context, bookingID uuid.UUID) (*domain.ClientReview, error)
	ListByClient(ctx context.Context, clientID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.ClientReview], error)
	RevealExpired(ctx context.Context) (int, error)
}

type clientReviewService struct {
	clientReviewRepo repository.ClientReviewRepository
	reviewRepo       repository.ReviewRepository
	bookingRepo      repository.BookingRepository
	bhRepo           repository.BathhouseRepository
	accessChecker    *AccessChecker
	notifSvc         NotificationService
	logger           *logger.Logger
}

func NewClientReviewService(
	clientReviewRepo repository.ClientReviewRepository,
	reviewRepo repository.ReviewRepository,
	bookingRepo repository.BookingRepository,
	bhRepo repository.BathhouseRepository,
	accessChecker *AccessChecker,
	notifSvc NotificationService,
	log *logger.Logger,
) ClientReviewService {
	return &clientReviewService{
		clientReviewRepo: clientReviewRepo,
		reviewRepo:       reviewRepo,
		bookingRepo:      bookingRepo,
		bhRepo:           bhRepo,
		accessChecker:    accessChecker,
		notifSvc:         notifSvc,
		logger:           log,
	}
}

func (s *clientReviewService) Create(ctx context.Context, ownerID uuid.UUID, ownerRole domain.UserRole, input CreateClientReviewInput) (*domain.ClientReview, error) {
	booking, err := s.bookingRepo.GetByID(ctx, input.BookingID)
	if err != nil {
		return nil, err
	}

	if booking.Status != domain.BookingCompleted {
		return nil, domain.ErrInvalidInput
	}

	// Verify the caller is the owner or representative of the bathhouse
	if err := s.accessChecker.CanManageBathhouse(ctx, ownerID, ownerRole, booking.BathhouseID); err != nil {
		return nil, err
	}

	// Check if client review already exists for this booking
	existing, err := s.clientReviewRepo.GetByBookingID(ctx, input.BookingID)
	if err != nil && !errors.Is(err, domain.ErrClientReviewNotFound) {
		return nil, err
	}
	if existing != nil {
		return nil, domain.ErrAlreadyExists
	}

	now := time.Now()
	revealAt := now.Add(domain.ClientReviewBlindDays * 24 * time.Hour)

	review := &domain.ClientReview{
		ID:             uuid.New(),
		OwnerID:        ownerID,
		ClientID:       booking.UserID,
		BookingID:      input.BookingID,
		BathhouseID:    booking.BathhouseID,
		Punctuality:    input.Punctuality,
		Cleanliness:    input.Cleanliness,
		RuleCompliance: input.RuleCompliance,
		Rating:         input.Rating,
		Text:           input.Text,
		RevealAt:       revealAt,
		IsRevealed:     false,
		Status:         domain.ReviewStatusApproved,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	review.ComputeOverallRating()

	if err := review.Validate(); err != nil {
		return nil, err
	}

	if err := s.clientReviewRepo.Create(ctx, review); err != nil {
		return nil, err
	}

	// Check if the client has also posted their review for this booking.
	// If so, both sides have reviewed -> reveal both immediately.
	s.tryMutualReveal(ctx, review)

	return review, nil
}

// tryMutualReveal checks if both parties have reviewed and reveals both if so.
func (s *clientReviewService) tryMutualReveal(ctx context.Context, clientReview *domain.ClientReview) {
	// Check if the client (guest) posted their review
	guestReview, err := s.reviewRepo.GetByBookingID(ctx, clientReview.BookingID)
	if err != nil {
		// Guest hasn't reviewed yet, keep both in blind period
		return
	}

	// Both reviews exist - reveal both immediately
	// Reveal the client review (owner's review of client)
	if err := s.clientReviewRepo.RevealByID(ctx, clientReview.ID); err != nil {
		s.logger.Warn("failed to reveal client review", "id", clientReview.ID, "error", err)
		return
	}
	clientReview.IsRevealed = true

	// Reveal the guest review (client's review of bathhouse)
	if !guestReview.IsRevealed {
		guestReview.IsRevealed = true
		guestReview.UpdatedAt = time.Now()
		if err := s.reviewRepo.Update(ctx, guestReview); err != nil {
			s.logger.Warn("failed to reveal guest review", "id", guestReview.ID, "error", err)
		}
	}

	// Notify both parties that reviews are revealed
	bh, _ := s.bhRepo.GetByID(ctx, clientReview.BathhouseID)
	bhName := ""
	if bh != nil {
		bhName = bh.Name
	}

	data := map[string]string{
		"booking_id":   clientReview.BookingID.String(),
		"bathhouse_id": clientReview.BathhouseID.String(),
	}

	// Notify the client
	_ = s.notifSvc.Send(ctx, clientReview.ClientID, domain.NotifReviewRevealed,
		"Отзывы раскрыты",
		fmt.Sprintf("Оба отзыва по бронированию в %s теперь видны", bhName),
		data,
	)

	// Notify the owner
	_ = s.notifSvc.Send(ctx, clientReview.OwnerID, domain.NotifReviewRevealed,
		"Отзывы раскрыты",
		fmt.Sprintf("Оба отзыва по бронированию в %s теперь видны", bhName),
		data,
	)
}

func (s *clientReviewService) GetByID(ctx context.Context, id uuid.UUID) (*domain.ClientReview, error) {
	return s.clientReviewRepo.GetByID(ctx, id)
}

func (s *clientReviewService) GetByBookingID(ctx context.Context, bookingID uuid.UUID) (*domain.ClientReview, error) {
	return s.clientReviewRepo.GetByBookingID(ctx, bookingID)
}

func (s *clientReviewService) ListByClient(ctx context.Context, clientID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.ClientReview], error) {
	return s.clientReviewRepo.ListByClient(ctx, clientID, true, page, pageSize)
}

// RevealExpired reveals client reviews whose blind period has elapsed (14 days).
// Also reveals corresponding guest reviews. Called by cron job.
func (s *clientReviewService) RevealExpired(ctx context.Context) (int, error) {
	now := time.Now()

	// Find unrevealed client reviews past deadline
	clientReviews, err := s.clientReviewRepo.ListUnrevealedPastDeadline(ctx, now)
	if err != nil {
		return 0, fmt.Errorf("list unrevealed client reviews: %w", err)
	}

	revealed := 0
	for _, cr := range clientReviews {
		if err := s.clientReviewRepo.RevealByID(ctx, cr.ID); err != nil {
			s.logger.Warn("failed to reveal client review by deadline", "id", cr.ID, "error", err)
			continue
		}

		// Also reveal the corresponding guest review if it exists and is unrevealed
		guestReview, err := s.reviewRepo.GetByBookingID(ctx, cr.BookingID)
		if err == nil && guestReview.RevealAt != nil && !guestReview.IsRevealed {
			guestReview.IsRevealed = true
			guestReview.UpdatedAt = now
			if err := s.reviewRepo.Update(ctx, guestReview); err != nil {
				s.logger.Warn("failed to reveal guest review by deadline", "id", guestReview.ID, "error", err)
			}
		}

		// Notify parties about reveal
		data := map[string]string{
			"booking_id":   cr.BookingID.String(),
			"bathhouse_id": cr.BathhouseID.String(),
		}
		_ = s.notifSvc.Send(ctx, cr.ClientID, domain.NotifReviewRevealed,
			"Отзыв раскрыт",
			"Период слепого обзора истёк — отзывы теперь видны",
			data,
		)
		_ = s.notifSvc.Send(ctx, cr.OwnerID, domain.NotifReviewRevealed,
			"Отзыв раскрыт",
			"Период слепого обзора истёк — отзывы теперь видны",
			data,
		)

		revealed++
	}

	// Also reveal guest reviews that have unrevealed status past deadline
	// (case where guest reviewed but owner never did — still reveal after 14 days)
	s.revealExpiredGuestReviews(ctx, now)

	return revealed, nil
}

// revealExpiredGuestReviews reveals guest reviews that are past their reveal_at deadline.
func (s *clientReviewService) revealExpiredGuestReviews(ctx context.Context, now time.Time) {
	reviews, err := s.reviewRepo.ListUnrevealedPastDeadline(ctx, now)
	if err != nil {
		s.logger.Warn("failed to list unrevealed guest reviews", "error", err)
		return
	}
	for _, rev := range reviews {
		rev.IsRevealed = true
		rev.UpdatedAt = now
		if err := s.reviewRepo.Update(ctx, &rev); err != nil {
			s.logger.Warn("failed to reveal guest review by deadline", "id", rev.ID, "error", err)
		}
	}
}
