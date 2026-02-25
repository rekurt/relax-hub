package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/repository/mock"
	"github.com/nikitaaldaev/bani/internal/service"
)

func newReviewService() (service.ReviewService, *mock.BathhouseRepo, *mock.BookingRepo, *mock.ReviewRepo) {
	bhRepo := mock.NewBathhouseRepo()
	bookingRepo := mock.NewBookingRepo()
	reviewRepo := mock.NewReviewRepo()
	svc := service.NewReviewService(reviewRepo, bookingRepo, bhRepo)
	return svc, bhRepo, bookingRepo, reviewRepo
}

func TestReviewService_Create_Success(t *testing.T) {
	svc, bhRepo, bookingRepo, _ := newReviewService()
	ownerID := uuid.New()
	clientID := uuid.New()
	bh := createBathhouse(t, bhRepo, ownerID)

	booking := &domain.Booking{
		ID: uuid.New(), UserID: clientID, BathhouseID: bh.ID,
		StartTime: time.Now().Add(-24 * time.Hour), EndTime: time.Now().Add(-22 * time.Hour),
		GuestCount: 2, TotalPrice: 10000, Status: domain.BookingCompleted,
	}
	_ = bookingRepo.Create(context.Background(), booking)

	review, err := svc.Create(context.Background(), clientID, service.CreateReviewInput{
		BookingID: booking.ID,
		Rating:    5,
		Text:      "Great bathhouse!",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if review.Rating != 5 {
		t.Errorf("rating = %d, want 5", review.Rating)
	}
}

func TestReviewService_Create_NotOwnBooking(t *testing.T) {
	svc, bhRepo, bookingRepo, _ := newReviewService()
	bh := createBathhouse(t, bhRepo, uuid.New())

	booking := &domain.Booking{
		ID: uuid.New(), UserID: uuid.New(), BathhouseID: bh.ID,
		StartTime: time.Now().Add(-24 * time.Hour), EndTime: time.Now().Add(-22 * time.Hour),
		GuestCount: 2, TotalPrice: 10000, Status: domain.BookingCompleted,
	}
	_ = bookingRepo.Create(context.Background(), booking)

	otherUserID := uuid.New()
	_, err := svc.Create(context.Background(), otherUserID, service.CreateReviewInput{
		BookingID: booking.ID,
		Rating:    5,
		Text:      "Fake review",
	})

	if !errors.Is(err, domain.ErrForbidden) {
		t.Errorf("should be forbidden for other user's booking, got: %v", err)
	}
}

func TestReviewService_Create_NotCompletedBooking(t *testing.T) {
	svc, bhRepo, bookingRepo, _ := newReviewService()
	clientID := uuid.New()
	bh := createBathhouse(t, bhRepo, uuid.New())

	booking := &domain.Booking{
		ID: uuid.New(), UserID: clientID, BathhouseID: bh.ID,
		StartTime: time.Now().Add(24 * time.Hour), EndTime: time.Now().Add(26 * time.Hour),
		GuestCount: 2, TotalPrice: 10000, Status: domain.BookingPending,
	}
	_ = bookingRepo.Create(context.Background(), booking)

	_, err := svc.Create(context.Background(), clientID, service.CreateReviewInput{
		BookingID: booking.ID,
		Rating:    5,
		Text:      "Review pending booking",
	})

	if !errors.Is(err, domain.ErrInvalidInput) {
		t.Errorf("should fail for non-completed booking, got: %v", err)
	}
}

func TestReviewService_Create_DuplicateReview(t *testing.T) {
	svc, bhRepo, bookingRepo, _ := newReviewService()
	clientID := uuid.New()
	bh := createBathhouse(t, bhRepo, uuid.New())

	booking := &domain.Booking{
		ID: uuid.New(), UserID: clientID, BathhouseID: bh.ID,
		StartTime: time.Now().Add(-24 * time.Hour), EndTime: time.Now().Add(-22 * time.Hour),
		GuestCount: 2, TotalPrice: 10000, Status: domain.BookingCompleted,
	}
	_ = bookingRepo.Create(context.Background(), booking)

	_, _ = svc.Create(context.Background(), clientID, service.CreateReviewInput{
		BookingID: booking.ID, Rating: 5, Text: "First review",
	})

	_, err := svc.Create(context.Background(), clientID, service.CreateReviewInput{
		BookingID: booking.ID, Rating: 4, Text: "Second review",
	})

	if !errors.Is(err, domain.ErrAlreadyExists) {
		t.Errorf("should fail for duplicate review, got: %v", err)
	}
}

func TestReviewService_Create_InvalidRating(t *testing.T) {
	svc, bhRepo, bookingRepo, _ := newReviewService()
	clientID := uuid.New()
	bh := createBathhouse(t, bhRepo, uuid.New())

	booking := &domain.Booking{
		ID: uuid.New(), UserID: clientID, BathhouseID: bh.ID,
		StartTime: time.Now().Add(-24 * time.Hour), EndTime: time.Now().Add(-22 * time.Hour),
		GuestCount: 2, TotalPrice: 10000, Status: domain.BookingCompleted,
	}
	_ = bookingRepo.Create(context.Background(), booking)

	_, err := svc.Create(context.Background(), clientID, service.CreateReviewInput{
		BookingID: booking.ID, Rating: 0, Text: "Bad rating",
	})

	if !errors.Is(err, domain.ErrInvalidInput) {
		t.Errorf("should fail for invalid rating, got: %v", err)
	}
}

func TestReviewService_ListByBathhouse(t *testing.T) {
	svc, bhRepo, bookingRepo, _ := newReviewService()
	clientID := uuid.New()
	bh := createBathhouse(t, bhRepo, uuid.New())

	booking := &domain.Booking{
		ID: uuid.New(), UserID: clientID, BathhouseID: bh.ID,
		StartTime: time.Now().Add(-24 * time.Hour), EndTime: time.Now().Add(-22 * time.Hour),
		GuestCount: 2, TotalPrice: 10000, Status: domain.BookingCompleted,
	}
	_ = bookingRepo.Create(context.Background(), booking)

	_, _ = svc.Create(context.Background(), clientID, service.CreateReviewInput{
		BookingID: booking.ID, Rating: 5, Text: "Great!",
	})

	result, err := svc.ListByBathhouse(context.Background(), bh.ID, 1, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.TotalCount != 1 {
		t.Errorf("totalCount = %d, want 1", result.TotalCount)
	}
}
