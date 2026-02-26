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

type reviewTestEnv struct {
	svc         service.ReviewService
	bhRepo      *mock.BathhouseRepo
	bookingRepo *mock.BookingRepo
	reviewRepo  *mock.ReviewRepo
	repRepo     *mock.RepresentativeRepo
}

func newReviewTestEnv() *reviewTestEnv {
	bhRepo := mock.NewBathhouseRepo()
	bookingRepo := mock.NewBookingRepo()
	reviewRepo := mock.NewReviewRepo()
	repRepo := mock.NewRepresentativeRepo()
	ac := service.NewAccessChecker(repRepo, bhRepo)
	svc := service.NewReviewService(reviewRepo, bookingRepo, bhRepo, ac)
	return &reviewTestEnv{
		svc:         svc,
		bhRepo:      bhRepo,
		bookingRepo: bookingRepo,
		reviewRepo:  reviewRepo,
		repRepo:     repRepo,
	}
}

// Keep backwards-compatible helper for existing tests.
func newReviewService() (service.ReviewService, *mock.BathhouseRepo, *mock.BookingRepo, *mock.ReviewRepo) {
	env := newReviewTestEnv()
	return env.svc, env.bhRepo, env.bookingRepo, env.reviewRepo
}

func createCompletedBooking(t *testing.T, bookingRepo *mock.BookingRepo, clientID, bathhouseID uuid.UUID) *domain.Booking {
	t.Helper()
	booking := &domain.Booking{
		ID: uuid.New(), UserID: clientID, BathhouseID: bathhouseID,
		StartTime: time.Now().Add(-24 * time.Hour), EndTime: time.Now().Add(-22 * time.Hour),
		GuestCount: 2, TotalPrice: 10000, Status: domain.BookingCompleted,
	}
	if err := bookingRepo.Create(context.Background(), booking); err != nil {
		t.Fatalf("create booking: %v", err)
	}
	return booking
}

func createReview(t *testing.T, svc service.ReviewService, clientID uuid.UUID, bookingID uuid.UUID) *domain.Review {
	t.Helper()
	review, err := svc.Create(context.Background(), clientID, service.CreateReviewInput{
		BookingID: bookingID, Rating: 5, Text: "Great bathhouse!",
	})
	if err != nil {
		t.Fatalf("create review: %v", err)
	}
	return review
}

func TestReviewService_Create_Success(t *testing.T) {
	svc, bhRepo, bookingRepo, _ := newReviewService()
	ownerID := uuid.New()
	clientID := uuid.New()
	bh := createBathhouse(t, bhRepo, ownerID)

	booking := createCompletedBooking(t, bookingRepo, clientID, bh.ID)

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
	if review.Status != domain.ReviewStatusPending {
		t.Errorf("status = %s, want pending", review.Status)
	}
}

func TestReviewService_Create_NotOwnBooking(t *testing.T) {
	svc, bhRepo, bookingRepo, _ := newReviewService()
	bh := createBathhouse(t, bhRepo, uuid.New())

	booking := createCompletedBooking(t, bookingRepo, uuid.New(), bh.ID)

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

	booking := createCompletedBooking(t, bookingRepo, clientID, bh.ID)

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

	booking := createCompletedBooking(t, bookingRepo, clientID, bh.ID)

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

	booking := createCompletedBooking(t, bookingRepo, clientID, bh.ID)

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

func TestReviewService_GetByID(t *testing.T) {
	env := newReviewTestEnv()
	clientID := uuid.New()
	bh := createBathhouse(t, env.bhRepo, uuid.New())
	booking := createCompletedBooking(t, env.bookingRepo, clientID, bh.ID)
	review := createReview(t, env.svc, clientID, booking.ID)

	got, err := env.svc.GetByID(context.Background(), review.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID != review.ID {
		t.Errorf("got ID = %s, want %s", got.ID, review.ID)
	}
}

func TestReviewService_GetByID_NotFound(t *testing.T) {
	env := newReviewTestEnv()

	_, err := env.svc.GetByID(context.Background(), uuid.New())
	if !errors.Is(err, domain.ErrReviewNotFound) {
		t.Errorf("should be ErrReviewNotFound, got: %v", err)
	}
}

func TestReviewService_Update_Success(t *testing.T) {
	env := newReviewTestEnv()
	clientID := uuid.New()
	bh := createBathhouse(t, env.bhRepo, uuid.New())
	booking := createCompletedBooking(t, env.bookingRepo, clientID, bh.ID)
	review := createReview(t, env.svc, clientID, booking.ID)

	newRating := 3
	newText := "Updated text"
	updated, err := env.svc.Update(context.Background(), clientID, review.ID, service.UpdateReviewInput{
		Rating: &newRating,
		Text:   &newText,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated.Rating != 3 {
		t.Errorf("rating = %d, want 3", updated.Rating)
	}
	if updated.Text != "Updated text" {
		t.Errorf("text = %s, want 'Updated text'", updated.Text)
	}
}

func TestReviewService_Update_NotAuthor(t *testing.T) {
	env := newReviewTestEnv()
	clientID := uuid.New()
	bh := createBathhouse(t, env.bhRepo, uuid.New())
	booking := createCompletedBooking(t, env.bookingRepo, clientID, bh.ID)
	review := createReview(t, env.svc, clientID, booking.ID)

	otherUser := uuid.New()
	newText := "Hacked"
	_, err := env.svc.Update(context.Background(), otherUser, review.ID, service.UpdateReviewInput{
		Text: &newText,
	})
	if !errors.Is(err, domain.ErrForbidden) {
		t.Errorf("should be forbidden for non-author, got: %v", err)
	}
}

func TestReviewService_Update_After24h(t *testing.T) {
	env := newReviewTestEnv()
	clientID := uuid.New()
	bh := createBathhouse(t, env.bhRepo, uuid.New())
	booking := createCompletedBooking(t, env.bookingRepo, clientID, bh.ID)
	review := createReview(t, env.svc, clientID, booking.ID)

	// Manually set CreatedAt to 25 hours ago
	old, _ := env.reviewRepo.GetByID(context.Background(), review.ID)
	old.CreatedAt = time.Now().Add(-25 * time.Hour)
	_ = env.reviewRepo.Update(context.Background(), old)

	newText := "Too late"
	_, err := env.svc.Update(context.Background(), clientID, review.ID, service.UpdateReviewInput{
		Text: &newText,
	})
	if !errors.Is(err, domain.ErrForbidden) {
		t.Errorf("should be forbidden after 24h, got: %v", err)
	}
}

func TestReviewService_Delete_ByAuthor(t *testing.T) {
	env := newReviewTestEnv()
	clientID := uuid.New()
	bh := createBathhouse(t, env.bhRepo, uuid.New())
	booking := createCompletedBooking(t, env.bookingRepo, clientID, bh.ID)
	review := createReview(t, env.svc, clientID, booking.ID)

	err := env.svc.Delete(context.Background(), clientID, domain.RoleClient, review.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = env.svc.GetByID(context.Background(), review.ID)
	if !errors.Is(err, domain.ErrReviewNotFound) {
		t.Errorf("review should be deleted, got: %v", err)
	}
}

func TestReviewService_Delete_ByAdmin(t *testing.T) {
	env := newReviewTestEnv()
	clientID := uuid.New()
	adminID := uuid.New()
	bh := createBathhouse(t, env.bhRepo, uuid.New())
	booking := createCompletedBooking(t, env.bookingRepo, clientID, bh.ID)
	review := createReview(t, env.svc, clientID, booking.ID)

	err := env.svc.Delete(context.Background(), adminID, domain.RoleAdmin, review.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = env.svc.GetByID(context.Background(), review.ID)
	if !errors.Is(err, domain.ErrReviewNotFound) {
		t.Errorf("review should be deleted, got: %v", err)
	}
}

func TestReviewService_Delete_Forbidden(t *testing.T) {
	env := newReviewTestEnv()
	clientID := uuid.New()
	bh := createBathhouse(t, env.bhRepo, uuid.New())
	booking := createCompletedBooking(t, env.bookingRepo, clientID, bh.ID)
	review := createReview(t, env.svc, clientID, booking.ID)

	otherUser := uuid.New()
	err := env.svc.Delete(context.Background(), otherUser, domain.RoleClient, review.ID)
	if !errors.Is(err, domain.ErrForbidden) {
		t.Errorf("should be forbidden for non-author/non-admin, got: %v", err)
	}
}

func TestReviewService_AddOwnerResponse_Success(t *testing.T) {
	env := newReviewTestEnv()
	ownerID := uuid.New()
	clientID := uuid.New()
	bh := createBathhouse(t, env.bhRepo, ownerID)
	booking := createCompletedBooking(t, env.bookingRepo, clientID, bh.ID)
	review := createReview(t, env.svc, clientID, booking.ID)

	updated, err := env.svc.AddOwnerResponse(context.Background(), ownerID, domain.RoleOwner, review.ID, "Thank you!")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated.OwnerResponse != "Thank you!" {
		t.Errorf("owner_response = %s, want 'Thank you!'", updated.OwnerResponse)
	}
	if updated.OwnerResponseAt == nil {
		t.Error("owner_response_at should not be nil")
	}
}

func TestReviewService_AddOwnerResponse_AlreadyResponded(t *testing.T) {
	env := newReviewTestEnv()
	ownerID := uuid.New()
	clientID := uuid.New()
	bh := createBathhouse(t, env.bhRepo, ownerID)
	booking := createCompletedBooking(t, env.bookingRepo, clientID, bh.ID)
	review := createReview(t, env.svc, clientID, booking.ID)

	_, _ = env.svc.AddOwnerResponse(context.Background(), ownerID, domain.RoleOwner, review.ID, "First response")

	_, err := env.svc.AddOwnerResponse(context.Background(), ownerID, domain.RoleOwner, review.ID, "Second response")
	if !errors.Is(err, domain.ErrReviewAlreadyResponded) {
		t.Errorf("should fail for already responded, got: %v", err)
	}
}

func TestReviewService_AddOwnerResponse_Forbidden(t *testing.T) {
	env := newReviewTestEnv()
	ownerID := uuid.New()
	clientID := uuid.New()
	bh := createBathhouse(t, env.bhRepo, ownerID)
	booking := createCompletedBooking(t, env.bookingRepo, clientID, bh.ID)
	review := createReview(t, env.svc, clientID, booking.ID)

	otherOwner := uuid.New()
	_, err := env.svc.AddOwnerResponse(context.Background(), otherOwner, domain.RoleOwner, review.ID, "Not my bathhouse")
	if !errors.Is(err, domain.ErrForbidden) {
		t.Errorf("should be forbidden for non-owner, got: %v", err)
	}
}

func TestReviewService_AddOwnerResponse_EmptyText(t *testing.T) {
	env := newReviewTestEnv()
	ownerID := uuid.New()
	clientID := uuid.New()
	bh := createBathhouse(t, env.bhRepo, ownerID)
	booking := createCompletedBooking(t, env.bookingRepo, clientID, bh.ID)
	review := createReview(t, env.svc, clientID, booking.ID)

	_, err := env.svc.AddOwnerResponse(context.Background(), ownerID, domain.RoleOwner, review.ID, "")
	if !errors.Is(err, domain.ErrInvalidInput) {
		t.Errorf("should fail for empty response, got: %v", err)
	}
}

func TestReviewService_AddOwnerResponse_ByRepresentative(t *testing.T) {
	env := newReviewTestEnv()
	ownerID := uuid.New()
	clientID := uuid.New()
	repUserID := uuid.New()
	bh := createBathhouse(t, env.bhRepo, ownerID)
	booking := createCompletedBooking(t, env.bookingRepo, clientID, bh.ID)
	review := createReview(t, env.svc, clientID, booking.ID)

	// Assign representative to this bathhouse
	rep := &domain.Representative{
		ID:          uuid.New(),
		UserID:      repUserID,
		BathhouseID: bh.ID,
	}
	_ = env.repRepo.Create(context.Background(), rep)

	updated, err := env.svc.AddOwnerResponse(context.Background(), repUserID, domain.RoleRepresentative, review.ID, "Thanks from rep!")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated.OwnerResponse != "Thanks from rep!" {
		t.Errorf("owner_response = %s, want 'Thanks from rep!'", updated.OwnerResponse)
	}
}
