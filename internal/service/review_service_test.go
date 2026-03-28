package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/logger"
	"github.com/nikitaaldaev/bani/internal/moderation"
	"github.com/nikitaaldaev/bani/internal/repository/mock"
	"github.com/nikitaaldaev/bani/internal/service"
	"github.com/nikitaaldaev/bani/internal/storage"
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
	log := logger.New(logger.LevelWarn) // Use Warn level to suppress debug output during tests
	// ContentFilter with moderation disabled (so reviews default to pending)
	contentFilter := moderation.NewContentFilter(false, false)
	mediaRepo := mock.NewMediaRepo()
	noopStore := storage.NewMockStorage()
	svc := service.NewReviewService(reviewRepo, mock.NewClientReviewRepo(), bookingRepo, bhRepo, mediaRepo, noopStore, ac, &noopNotifService{}, contentFilter, nil, nil, log)
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

// Note: logger is imported but only used indirectly in newReviewTestEnv

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
	// When moderation is disabled, reviews are auto-approved
	if review.Status != domain.ReviewStatusApproved {
		t.Errorf("status = %s, want approved", review.Status)
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
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("should be ErrNotFound, got: %v", err)
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
	if !errors.Is(err, domain.ErrNotFound) {
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
	if !errors.Is(err, domain.ErrNotFound) {
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

// Test moderation integration
func TestReviewService_Create_WithModeration_CleanText(t *testing.T) {
	bhRepo := mock.NewBathhouseRepo()
	bookingRepo := mock.NewBookingRepo()
	reviewRepo := mock.NewReviewRepo()
	repRepo := mock.NewRepresentativeRepo()
	ac := service.NewAccessChecker(repRepo, bhRepo)
	log := logger.New(logger.LevelWarn)
	// ContentFilter with moderation enabled and auto-approve enabled
	contentFilter := moderation.NewContentFilter(true, true)
	mediaRepo := mock.NewMediaRepo()
	noopStore := storage.NewMockStorage()
	svc := service.NewReviewService(reviewRepo, mock.NewClientReviewRepo(), bookingRepo, bhRepo, mediaRepo, noopStore, ac, &noopNotifService{}, contentFilter, nil, nil, log)

	clientID := uuid.New()
	bh := createBathhouse(t, bhRepo, uuid.New())
	booking := createCompletedBooking(t, bookingRepo, clientID, bh.ID)

	review, err := svc.Create(context.Background(), clientID, service.CreateReviewInput{
		BookingID: booking.ID,
		Rating:    5,
		Text:      "This is a great bathhouse experience, highly recommended!",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if review.Status != domain.ReviewStatusApproved {
		t.Errorf("status = %s, want approved", review.Status)
	}
}

func TestReviewService_Create_WithModeration_AutoRejectProfanity(t *testing.T) {
	bhRepo := mock.NewBathhouseRepo()
	bookingRepo := mock.NewBookingRepo()
	reviewRepo := mock.NewReviewRepo()
	repRepo := mock.NewRepresentativeRepo()
	ac := service.NewAccessChecker(repRepo, bhRepo)
	log := logger.New(logger.LevelWarn)
	// ContentFilter with moderation enabled but auto-approve disabled
	contentFilter := moderation.NewContentFilter(true, false)
	mediaRepo := mock.NewMediaRepo()
	noopStore := storage.NewMockStorage()
	svc := service.NewReviewService(reviewRepo, mock.NewClientReviewRepo(), bookingRepo, bhRepo, mediaRepo, noopStore, ac, &noopNotifService{}, contentFilter, nil, nil, log)

	clientID := uuid.New()
	bh := createBathhouse(t, bhRepo, uuid.New())
	booking := createCompletedBooking(t, bookingRepo, clientID, bh.ID)

	review, err := svc.Create(context.Background(), clientID, service.CreateReviewInput{
		BookingID: booking.ID,
		Rating:    1,
		Text:      "This place is хуйня and I hate it",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if review.Status != domain.ReviewStatusRejected {
		t.Errorf("status = %s, want rejected", review.Status)
	}
	if len(review.RejectionReasons) == 0 {
		t.Error("rejection_reasons should not be empty")
	}
}

func TestReviewService_Create_WithModeration_PendingCleanText(t *testing.T) {
	bhRepo := mock.NewBathhouseRepo()
	bookingRepo := mock.NewBookingRepo()
	reviewRepo := mock.NewReviewRepo()
	repRepo := mock.NewRepresentativeRepo()
	ac := service.NewAccessChecker(repRepo, bhRepo)
	log := logger.New(logger.LevelWarn)
	// ContentFilter with moderation enabled but auto-approve disabled
	contentFilter := moderation.NewContentFilter(true, false)
	mediaRepo := mock.NewMediaRepo()
	noopStore := storage.NewMockStorage()
	svc := service.NewReviewService(reviewRepo, mock.NewClientReviewRepo(), bookingRepo, bhRepo, mediaRepo, noopStore, ac, &noopNotifService{}, contentFilter, nil, nil, log)

	clientID := uuid.New()
	bh := createBathhouse(t, bhRepo, uuid.New())
	booking := createCompletedBooking(t, bookingRepo, clientID, bh.ID)

	review, err := svc.Create(context.Background(), clientID, service.CreateReviewInput{
		BookingID: booking.ID,
		Rating:    4,
		Text:      "Good bathhouse with nice facilities and friendly staff",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if review.Status != domain.ReviewStatusPending {
		t.Errorf("status = %s, want pending", review.Status)
	}
	if len(review.RejectionReasons) > 0 {
		t.Errorf("rejection_reasons should be empty, got %v", review.RejectionReasons)
	}
}

func TestReviewService_ListByBathhouse_FiltersNonApprovedReviews(t *testing.T) {
	env := newReviewTestEnv()
	clientID := uuid.New()
	bh := createBathhouse(t, env.bhRepo, uuid.New())

	// Create multiple reviews
	booking1 := createCompletedBooking(t, env.bookingRepo, clientID, bh.ID)
	review1 := createReview(t, env.svc, clientID, booking1.ID)

	// Manually set one review to pending (simulating a review that hasn't been auto-approved)
	review1.Status = domain.ReviewStatusPending
	_ = env.reviewRepo.Update(context.Background(), review1)

	// Create another booking and review
	user2 := uuid.New()
	booking2 := createCompletedBooking(t, env.bookingRepo, user2, bh.ID)
	review2 := createReview(t, env.svc, user2, booking2.ID)

	result, err := env.svc.ListByBathhouse(context.Background(), bh.ID, 1, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should only include approved reviews
	if result.TotalCount != 1 {
		t.Errorf("totalCount = %d, want 1 (only approved reviews)", result.TotalCount)
	}
	if len(result.Items) > 0 && result.Items[0].ID == review1.ID {
		t.Error("pending review should not be in results")
	}
	if len(result.Items) > 0 && result.Items[0].ID != review2.ID {
		t.Errorf("expected review2 in results, got %v", result.Items[0].ID)
	}
}

// --- Multi-criteria rating tests ---

func ptrFloat(v float64) *float64 { return &v }

func TestReviewService_Create_WithCriteria(t *testing.T) {
	svc, bhRepo, bookingRepo, _ := newReviewService()
	ownerID := uuid.New()
	clientID := uuid.New()
	bh := createBathhouse(t, bhRepo, ownerID)
	booking := createCompletedBooking(t, bookingRepo, clientID, bh.ID)

	review, err := svc.Create(context.Background(), clientID, service.CreateReviewInput{
		BookingID:     booking.ID,
		Rating:        3, // will be overridden by criteria average
		Cleanliness:   ptrFloat(5.0),
		Accuracy:      ptrFloat(4.0),
		Communication: ptrFloat(3.5),
		ValueForMoney: ptrFloat(4.5),
		Text:          "Good with criteria",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Overall rating should be computed as average of criteria: (5+4+3.5+4.5)/4 = 4.25 -> 4
	if review.Rating != 4 {
		t.Errorf("rating = %d, want 4 (computed from criteria)", review.Rating)
	}
	if review.Cleanliness == nil || *review.Cleanliness != 5.0 {
		t.Errorf("cleanliness = %v, want 5.0", review.Cleanliness)
	}
	if review.ValueForMoney == nil || *review.ValueForMoney != 4.5 {
		t.Errorf("value_for_money = %v, want 4.5", review.ValueForMoney)
	}
}

func TestReviewService_Create_WithPartialCriteria_Fails(t *testing.T) {
	svc, bhRepo, bookingRepo, _ := newReviewService()
	clientID := uuid.New()
	bh := createBathhouse(t, bhRepo, uuid.New())
	booking := createCompletedBooking(t, bookingRepo, clientID, bh.ID)

	_, err := svc.Create(context.Background(), clientID, service.CreateReviewInput{
		BookingID:   booking.ID,
		Rating:      4,
		Cleanliness: ptrFloat(5.0),
		// Missing other criteria
		Text: "Partial criteria",
	})
	if !errors.Is(err, domain.ErrInvalidInput) {
		t.Errorf("should fail with partial criteria, got: %v", err)
	}
}

func TestReviewService_Create_WithInvalidCriterion_Fails(t *testing.T) {
	svc, bhRepo, bookingRepo, _ := newReviewService()
	clientID := uuid.New()
	bh := createBathhouse(t, bhRepo, uuid.New())
	booking := createCompletedBooking(t, bookingRepo, clientID, bh.ID)

	_, err := svc.Create(context.Background(), clientID, service.CreateReviewInput{
		BookingID:     booking.ID,
		Rating:        4,
		Cleanliness:   ptrFloat(5.5), // invalid: max 5.0
		Accuracy:      ptrFloat(4.0),
		Communication: ptrFloat(3.5),
		ValueForMoney: ptrFloat(4.5),
		Text:          "Invalid criterion",
	})
	if !errors.Is(err, domain.ErrInvalidInput) {
		t.Errorf("should fail with invalid criterion value, got: %v", err)
	}
}

func TestReviewService_Create_WithInvalidStep_Fails(t *testing.T) {
	svc, bhRepo, bookingRepo, _ := newReviewService()
	clientID := uuid.New()
	bh := createBathhouse(t, bhRepo, uuid.New())
	booking := createCompletedBooking(t, bookingRepo, clientID, bh.ID)

	_, err := svc.Create(context.Background(), clientID, service.CreateReviewInput{
		BookingID:     booking.ID,
		Rating:        4,
		Cleanliness:   ptrFloat(4.3), // invalid: must be step 0.5
		Accuracy:      ptrFloat(4.0),
		Communication: ptrFloat(3.5),
		ValueForMoney: ptrFloat(4.5),
		Text:          "Invalid step",
	})
	if !errors.Is(err, domain.ErrInvalidInput) {
		t.Errorf("should fail with non-0.5 step, got: %v", err)
	}
}

func TestReviewService_Create_WithoutCriteria_BackwardsCompat(t *testing.T) {
	svc, bhRepo, bookingRepo, _ := newReviewService()
	clientID := uuid.New()
	bh := createBathhouse(t, bhRepo, uuid.New())
	booking := createCompletedBooking(t, bookingRepo, clientID, bh.ID)

	review, err := svc.Create(context.Background(), clientID, service.CreateReviewInput{
		BookingID: booking.ID,
		Rating:    5,
		Text:      "No criteria, old style",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if review.Rating != 5 {
		t.Errorf("rating = %d, want 5", review.Rating)
	}
	if review.Cleanliness != nil {
		t.Error("cleanliness should be nil for reviews without criteria")
	}
}

func TestReviewService_Update_WithCriteria(t *testing.T) {
	env := newReviewTestEnv()
	clientID := uuid.New()
	bh := createBathhouse(t, env.bhRepo, uuid.New())
	booking := createCompletedBooking(t, env.bookingRepo, clientID, bh.ID)
	review := createReview(t, env.svc, clientID, booking.ID)

	updated, err := env.svc.Update(context.Background(), clientID, review.ID, service.UpdateReviewInput{
		Cleanliness:   ptrFloat(4.0),
		Accuracy:      ptrFloat(3.5),
		Communication: ptrFloat(4.5),
		ValueForMoney: ptrFloat(5.0),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// (4+3.5+4.5+5)/4 = 4.25 -> 4
	if updated.Rating != 4 {
		t.Errorf("rating = %d, want 4 (recomputed from criteria)", updated.Rating)
	}
	if updated.Cleanliness == nil || *updated.Cleanliness != 4.0 {
		t.Errorf("cleanliness = %v, want 4.0", updated.Cleanliness)
	}
}

func TestReviewService_GetCriteriaAverages(t *testing.T) {
	svc, bhRepo, bookingRepo, _ := newReviewService()
	ownerID := uuid.New()
	bh := createBathhouse(t, bhRepo, ownerID)

	// Create 2 reviews with criteria
	client1 := uuid.New()
	booking1 := createCompletedBooking(t, bookingRepo, client1, bh.ID)
	_, err := svc.Create(context.Background(), client1, service.CreateReviewInput{
		BookingID:     booking1.ID,
		Rating:        4,
		Cleanliness:   ptrFloat(5.0),
		Accuracy:      ptrFloat(4.0),
		Communication: ptrFloat(3.0),
		ValueForMoney: ptrFloat(4.0),
		Text:          "Review 1",
	})
	if err != nil {
		t.Fatalf("create review 1: %v", err)
	}

	client2 := uuid.New()
	booking2 := createCompletedBooking(t, bookingRepo, client2, bh.ID)
	_, err = svc.Create(context.Background(), client2, service.CreateReviewInput{
		BookingID:     booking2.ID,
		Rating:        3,
		Cleanliness:   ptrFloat(3.0),
		Accuracy:      ptrFloat(4.0),
		Communication: ptrFloat(5.0),
		ValueForMoney: ptrFloat(2.0),
		Text:          "Review 2",
	})
	if err != nil {
		t.Fatalf("create review 2: %v", err)
	}

	avgs, err := svc.GetCriteriaAverages(context.Background(), bh.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Expected averages: cleanliness (5+3)/2=4.0, accuracy (4+4)/2=4.0, communication (3+5)/2=4.0, value_for_money (4+2)/2=3.0
	if avgs.AvgCleanliness != 4.0 {
		t.Errorf("avg_cleanliness = %f, want 4.0", avgs.AvgCleanliness)
	}
	if avgs.AvgAccuracy != 4.0 {
		t.Errorf("avg_accuracy = %f, want 4.0", avgs.AvgAccuracy)
	}
	if avgs.AvgCommunication != 4.0 {
		t.Errorf("avg_communication = %f, want 4.0", avgs.AvgCommunication)
	}
	if avgs.AvgValueForMoney != 3.0 {
		t.Errorf("avg_value_for_money = %f, want 3.0", avgs.AvgValueForMoney)
	}
}

func TestReviewService_GetCriteriaAverages_Empty(t *testing.T) {
	svc, bhRepo, _, _ := newReviewService()
	bh := createBathhouse(t, bhRepo, uuid.New())

	avgs, err := svc.GetCriteriaAverages(context.Background(), bh.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if avgs.AvgCleanliness != 0 || avgs.AvgAccuracy != 0 || avgs.AvgCommunication != 0 || avgs.AvgValueForMoney != 0 {
		t.Errorf("averages should all be 0 for bathhouse with no criteria reviews, got %+v", avgs)
	}
}

// --- Bayesian rating tests ---

func TestReviewService_BayesianRating_ZeroReviews(t *testing.T) {
	env := newReviewTestEnv()
	bh := createBathhouse(t, env.bhRepo, uuid.New())

	err := env.svc.RecalculateBayesianRating(context.Background(), bh.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	updated, _ := env.bhRepo.GetByID(context.Background(), bh.ID)
	// 0 reviews: (0*0 + 5*0) / (0+5) = 0
	if updated.BayesianRating != 0 {
		t.Errorf("bayesian_rating = %f, want 0 for bathhouse with no reviews", updated.BayesianRating)
	}
}

func TestReviewService_BayesianRating_OneReview(t *testing.T) {
	env := newReviewTestEnv()
	ownerID := uuid.New()
	bh := createBathhouse(t, env.bhRepo, ownerID)
	clientID := uuid.New()
	booking := createCompletedBooking(t, env.bookingRepo, clientID, bh.ID)

	_, err := env.svc.Create(context.Background(), clientID, service.CreateReviewInput{
		BookingID: booking.ID,
		Rating:    5,
		Text:      "Perfect!",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	updated, _ := env.bhRepo.GetByID(context.Background(), bh.ID)
	// 1 review (rating 5), platform avg=5 (only review), m=5
	// bayesian = (1*5 + 5*5) / (1+5) = 30/6 = 5.0
	if updated.BayesianRating != 5.0 {
		t.Errorf("bayesian_rating = %f, want 5.0 for single 5-star review with platform avg=5", updated.BayesianRating)
	}
}

func TestReviewService_BayesianRating_ManyReviews(t *testing.T) {
	env := newReviewTestEnv()
	ownerID := uuid.New()
	bh := createBathhouse(t, env.bhRepo, ownerID)

	// Create 10 reviews with rating 4
	for i := 0; i < 10; i++ {
		clientID := uuid.New()
		booking := createCompletedBooking(t, env.bookingRepo, clientID, bh.ID)
		_, err := env.svc.Create(context.Background(), clientID, service.CreateReviewInput{
			BookingID: booking.ID,
			Rating:    4,
			Text:      "Good bathhouse",
		})
		if err != nil {
			t.Fatalf("create review %d: %v", i, err)
		}
	}

	updated, _ := env.bhRepo.GetByID(context.Background(), bh.ID)
	// 10 reviews all rating 4, platform avg = 4.0, m = 5
	// bayesian = (10*4 + 5*4) / (10+5) = 60/15 = 4.0
	if updated.BayesianRating != 4.0 {
		t.Errorf("bayesian_rating = %f, want 4.0", updated.BayesianRating)
	}
}

func TestReviewService_BayesianRating_PullsTowardPlatformAvg(t *testing.T) {
	env := newReviewTestEnv()

	// First create a bathhouse with many reviews to establish platform avg
	ownerA := uuid.New()
	bhA := createBathhouse(t, env.bhRepo, ownerA)
	for i := 0; i < 10; i++ {
		clientID := uuid.New()
		booking := createCompletedBooking(t, env.bookingRepo, clientID, bhA.ID)
		_, _ = env.svc.Create(context.Background(), clientID, service.CreateReviewInput{
			BookingID: booking.ID,
			Rating:    3,
			Text:      "Average",
		})
	}

	// Now create a new bathhouse with 1 five-star review
	ownerB := uuid.New()
	bhB := createBathhouse(t, env.bhRepo, ownerB)
	clientB := uuid.New()
	bookingB := createCompletedBooking(t, env.bookingRepo, clientB, bhB.ID)
	_, err := env.svc.Create(context.Background(), clientB, service.CreateReviewInput{
		BookingID: bookingB.ID,
		Rating:    5,
		Text:      "Amazing!",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	updatedB, _ := env.bhRepo.GetByID(context.Background(), bhB.ID)
	// Platform avg is about 3.0 (10 reviews of 3 + 1 review of 5 = 35/11 ≈ 3.18)
	// For bhB: n=1, R=5, C≈3.18, m=5
	// bayesian = (1*5 + 5*3.18) / (1+5) = (5+15.9)/6 ≈ 3.5
	// The Bayesian should be much lower than the raw 5.0, pulled toward platform avg
	if updatedB.BayesianRating >= 5.0 {
		t.Errorf("bayesian_rating = %f, should be pulled below 5.0 toward platform avg", updatedB.BayesianRating)
	}
	if updatedB.BayesianRating <= 3.0 {
		t.Errorf("bayesian_rating = %f, should be above platform avg (3.0) since review is 5", updatedB.BayesianRating)
	}
}

func TestReviewService_BayesianRating_UpdateOnDelete(t *testing.T) {
	env := newReviewTestEnv()
	ownerID := uuid.New()
	bh := createBathhouse(t, env.bhRepo, ownerID)
	clientID := uuid.New()
	booking := createCompletedBooking(t, env.bookingRepo, clientID, bh.ID)

	review, err := env.svc.Create(context.Background(), clientID, service.CreateReviewInput{
		BookingID: booking.ID,
		Rating:    5,
		Text:      "Will be deleted",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Delete the review
	err = env.svc.Delete(context.Background(), clientID, domain.RoleClient, review.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	updated, _ := env.bhRepo.GetByID(context.Background(), bh.ID)
	// After deleting all reviews, bayesian should be 0
	if updated.BayesianRating != 0 {
		t.Errorf("bayesian_rating = %f, want 0 after deleting all reviews", updated.BayesianRating)
	}
}

func TestReviewService_RefreshPlatformAverage(t *testing.T) {
	env := newReviewTestEnv()

	// No reviews: should not error
	err := env.svc.RefreshPlatformAverage(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestReviewService_SendReviewRequests_NoBookings(t *testing.T) {
	env := newReviewTestEnv()

	sent, err := env.svc.SendReviewRequests(context.Background(), 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sent != 0 {
		t.Errorf("sent = %d, want 0", sent)
	}
}

func TestReviewService_SendReviewRequests_SendsForEligibleBooking(t *testing.T) {
	env := newReviewTestEnv()

	ownerID := uuid.New()
	clientID := uuid.New()
	bathhouseID := uuid.New()

	bh := &domain.Bathhouse{
		ID: bathhouseID, OwnerID: ownerID, Name: "Тестовая баня",
		Status: domain.BathhouseStatusActive, PricePerHour: 5000,
	}
	if err := env.bhRepo.Create(context.Background(), bh); err != nil {
		t.Fatalf("create bathhouse: %v", err)
	}

	// Booking completed and checked out 3 hours ago
	checkedOutAt := time.Now().Add(-3 * time.Hour)
	booking := &domain.Booking{
		ID: uuid.New(), UserID: clientID, BathhouseID: bathhouseID,
		StartTime: time.Now().Add(-6 * time.Hour), EndTime: time.Now().Add(-4 * time.Hour),
		GuestCount: 2, TotalPrice: 10000, Status: domain.BookingCompleted,
		CheckedOutAt: &checkedOutAt,
	}
	if err := env.bookingRepo.Create(context.Background(), booking); err != nil {
		t.Fatalf("create booking: %v", err)
	}

	sent, err := env.svc.SendReviewRequests(context.Background(), 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sent != 1 {
		t.Errorf("sent = %d, want 1", sent)
	}
}

func TestReviewService_SendReviewRequests_SkipsTooRecent(t *testing.T) {
	env := newReviewTestEnv()

	ownerID := uuid.New()
	clientID := uuid.New()
	bathhouseID := uuid.New()

	bh := &domain.Bathhouse{
		ID: bathhouseID, OwnerID: ownerID, Name: "Тестовая баня",
		Status: domain.BathhouseStatusActive, PricePerHour: 5000,
	}
	if err := env.bhRepo.Create(context.Background(), bh); err != nil {
		t.Fatalf("create bathhouse: %v", err)
	}

	// Booking checked out only 30 minutes ago (less than 2h delay)
	checkedOutAt := time.Now().Add(-30 * time.Minute)
	booking := &domain.Booking{
		ID: uuid.New(), UserID: clientID, BathhouseID: bathhouseID,
		StartTime: time.Now().Add(-3 * time.Hour), EndTime: time.Now().Add(-1 * time.Hour),
		GuestCount: 2, TotalPrice: 10000, Status: domain.BookingCompleted,
		CheckedOutAt: &checkedOutAt,
	}
	if err := env.bookingRepo.Create(context.Background(), booking); err != nil {
		t.Fatalf("create booking: %v", err)
	}

	sent, err := env.svc.SendReviewRequests(context.Background(), 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sent != 0 {
		t.Errorf("sent = %d, want 0 (booking too recent)", sent)
	}
}

func TestReviewService_SendReviewRequests_SkipsNonCompleted(t *testing.T) {
	env := newReviewTestEnv()

	clientID := uuid.New()
	bathhouseID := uuid.New()

	// Booking is confirmed, not completed
	booking := &domain.Booking{
		ID: uuid.New(), UserID: clientID, BathhouseID: bathhouseID,
		StartTime: time.Now().Add(-6 * time.Hour), EndTime: time.Now().Add(-4 * time.Hour),
		GuestCount: 2, TotalPrice: 10000, Status: domain.BookingConfirmed,
	}
	if err := env.bookingRepo.Create(context.Background(), booking); err != nil {
		t.Fatalf("create booking: %v", err)
	}

	sent, err := env.svc.SendReviewRequests(context.Background(), 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sent != 0 {
		t.Errorf("sent = %d, want 0 (booking not completed)", sent)
	}
}

func TestReviewService_SendReviewRequests_DefaultDelay(t *testing.T) {
	env := newReviewTestEnv()

	ownerID := uuid.New()
	clientID := uuid.New()
	bathhouseID := uuid.New()

	bh := &domain.Bathhouse{
		ID: bathhouseID, OwnerID: ownerID, Name: "Тестовая баня",
		Status: domain.BathhouseStatusActive, PricePerHour: 5000,
	}
	if err := env.bhRepo.Create(context.Background(), bh); err != nil {
		t.Fatalf("create bathhouse: %v", err)
	}

	checkedOutAt := time.Now().Add(-3 * time.Hour)
	booking := &domain.Booking{
		ID: uuid.New(), UserID: clientID, BathhouseID: bathhouseID,
		StartTime: time.Now().Add(-6 * time.Hour), EndTime: time.Now().Add(-4 * time.Hour),
		GuestCount: 2, TotalPrice: 10000, Status: domain.BookingCompleted,
		CheckedOutAt: &checkedOutAt,
	}
	if err := env.bookingRepo.Create(context.Background(), booking); err != nil {
		t.Fatalf("create booking: %v", err)
	}

	// Pass 0 to trigger default delay (2h)
	sent, err := env.svc.SendReviewRequests(context.Background(), 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sent != 1 {
		t.Errorf("sent = %d, want 1 (default delay should be 2h)", sent)
	}
}

func TestReviewService_SendReviewRequests_MultipleBookings(t *testing.T) {
	env := newReviewTestEnv()

	ownerID := uuid.New()
	bathhouseID := uuid.New()

	bh := &domain.Bathhouse{
		ID: bathhouseID, OwnerID: ownerID, Name: "Тестовая баня",
		Status: domain.BathhouseStatusActive, PricePerHour: 5000,
	}
	if err := env.bhRepo.Create(context.Background(), bh); err != nil {
		t.Fatalf("create bathhouse: %v", err)
	}

	// 3 bookings: 2 eligible, 1 too recent
	for i := 0; i < 2; i++ {
		checkedOutAt := time.Now().Add(-time.Duration(3+i) * time.Hour)
		booking := &domain.Booking{
			ID: uuid.New(), UserID: uuid.New(), BathhouseID: bathhouseID,
			StartTime: time.Now().Add(-6 * time.Hour), EndTime: time.Now().Add(-4 * time.Hour),
			GuestCount: 2, TotalPrice: 10000, Status: domain.BookingCompleted,
			CheckedOutAt: &checkedOutAt,
		}
		if err := env.bookingRepo.Create(context.Background(), booking); err != nil {
			t.Fatalf("create booking %d: %v", i, err)
		}
	}
	// Too recent
	recentCheckout := time.Now().Add(-30 * time.Minute)
	recentBooking := &domain.Booking{
		ID: uuid.New(), UserID: uuid.New(), BathhouseID: bathhouseID,
		StartTime: time.Now().Add(-3 * time.Hour), EndTime: time.Now().Add(-1 * time.Hour),
		GuestCount: 1, TotalPrice: 5000, Status: domain.BookingCompleted,
		CheckedOutAt: &recentCheckout,
	}
	if err := env.bookingRepo.Create(context.Background(), recentBooking); err != nil {
		t.Fatalf("create recent booking: %v", err)
	}

	sent, err := env.svc.SendReviewRequests(context.Background(), 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sent != 2 {
		t.Errorf("sent = %d, want 2", sent)
	}
}

// --- Quality Monitoring Tests ---

type reviewTestEnvWithTracking struct {
	svc         service.ReviewService
	bhRepo      *mock.BathhouseRepo
	bookingRepo *mock.BookingRepo
	reviewRepo  *mock.ReviewRepo
	notifSvc    *trackingNotifService
}

func newReviewTestEnvWithTracking() *reviewTestEnvWithTracking {
	bhRepo := mock.NewBathhouseRepo()
	bookingRepo := mock.NewBookingRepo()
	reviewRepo := mock.NewReviewRepo()
	repRepo := mock.NewRepresentativeRepo()
	ac := service.NewAccessChecker(repRepo, bhRepo)
	log := logger.New(logger.LevelWarn)
	contentFilter := moderation.NewContentFilter(false, false)
	mediaRepo := mock.NewMediaRepo()
	noopStore := storage.NewMockStorage()
	notifSvc := &trackingNotifService{}
	svc := service.NewReviewService(reviewRepo, mock.NewClientReviewRepo(), bookingRepo, bhRepo, mediaRepo, noopStore, ac, notifSvc, contentFilter, nil, nil, log)
	return &reviewTestEnvWithTracking{
		svc: svc, bhRepo: bhRepo, bookingRepo: bookingRepo, reviewRepo: reviewRepo, notifSvc: notifSvc,
	}
}

func TestQualityMonitoring_LowRatingWarning(t *testing.T) {
	env := newReviewTestEnvWithTracking()
	ctx := context.Background()
	ownerID := uuid.New()
	bathhouseID := uuid.New()
	clientID := uuid.New()

	// Pre-set bathhouse with 10 reviews and rating 2.5 (< 3.0 but >= 2.0)
	// Mock UpdateRating is a no-op, so we set rating/review_count directly
	bh := &domain.Bathhouse{
		ID: bathhouseID, OwnerID: ownerID, Name: "Тест Баня", Slug: "test-banya",
		Address: "ул. Тестовая 1", CityID: 1, PricePerHour: 5000,
		MaxGuests: 10, MinDuration: 1, Status: domain.BathhouseStatusActive,
		Rating: 2.5, ReviewCount: 10,
	}
	_ = env.bhRepo.Create(ctx, bh)

	// Create one review to trigger quality check
	booking := &domain.Booking{
		ID: uuid.New(), UserID: clientID, BathhouseID: bathhouseID,
		StartTime: time.Now().Add(-48 * time.Hour), EndTime: time.Now().Add(-46 * time.Hour),
		GuestCount: 2, TotalPrice: 10000, Status: domain.BookingCompleted,
	}
	_ = env.bookingRepo.Create(ctx, booking)
	_, err := env.svc.Create(ctx, clientID, service.CreateReviewInput{
		BookingID: booking.ID, Rating: 2, Text: "Не понравилось",
	})
	if err != nil {
		t.Fatalf("create review: %v", err)
	}

	// Check that a low rating warning was sent
	found := false
	for _, n := range env.notifSvc.sent {
		if n.Type == domain.NotifLowRatingWarning {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected low rating warning notification, got none")
	}

	// Bathhouse should still be active (rating 2.5 is not < 2.0)
	updatedBh, _ := env.bhRepo.GetByID(ctx, bathhouseID)
	if updatedBh.Status != domain.BathhouseStatusActive {
		t.Errorf("expected status active, got %s", updatedBh.Status)
	}
}

func TestQualityMonitoring_AutoDepublish(t *testing.T) {
	env := newReviewTestEnvWithTracking()
	ctx := context.Background()
	ownerID := uuid.New()
	bathhouseID := uuid.New()
	clientID := uuid.New()

	// Pre-set bathhouse with 10 reviews and rating 1.5 (< 2.0) to trigger depublish
	bh := &domain.Bathhouse{
		ID: bathhouseID, OwnerID: ownerID, Name: "Тест Баня 2", Slug: "test-banya-2",
		Address: "ул. Тестовая 2", CityID: 1, PricePerHour: 5000,
		MaxGuests: 10, MinDuration: 1, Status: domain.BathhouseStatusActive,
		Rating: 1.5, ReviewCount: 10,
	}
	_ = env.bhRepo.Create(ctx, bh)

	// Create one review to trigger quality check
	booking := &domain.Booking{
		ID: uuid.New(), UserID: clientID, BathhouseID: bathhouseID,
		StartTime: time.Now().Add(-48 * time.Hour), EndTime: time.Now().Add(-46 * time.Hour),
		GuestCount: 2, TotalPrice: 10000, Status: domain.BookingCompleted,
	}
	_ = env.bookingRepo.Create(ctx, booking)
	_, err := env.svc.Create(ctx, clientID, service.CreateReviewInput{
		BookingID: booking.ID, Rating: 1, Text: "Ужасно",
	})
	if err != nil {
		t.Fatalf("create review: %v", err)
	}

	// Check bathhouse was depublished
	updatedBh, _ := env.bhRepo.GetByID(ctx, bathhouseID)
	if updatedBh.Status != domain.BathhouseStatusInactive {
		t.Errorf("expected status inactive (depublished), got %s", updatedBh.Status)
	}

	// Check depublish notification was sent
	found := false
	for _, n := range env.notifSvc.sent {
		if n.Type == domain.NotifBathhouseDepublished {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected bathhouse depublished notification, got none")
	}
}

func TestQualityMonitoring_NoActionUnderThreshold(t *testing.T) {
	env := newReviewTestEnvWithTracking()
	ctx := context.Background()
	ownerID := uuid.New()
	bathhouseID := uuid.New()
	clientID := uuid.New()

	// Pre-set bathhouse with only 5 reviews (below threshold of 10)
	bh := &domain.Bathhouse{
		ID: bathhouseID, OwnerID: ownerID, Name: "Тест Баня 3", Slug: "test-banya-3",
		Address: "ул. Тестовая 3", CityID: 1, PricePerHour: 5000,
		MaxGuests: 10, MinDuration: 1, Status: domain.BathhouseStatusActive,
		Rating: 1.0, ReviewCount: 5,
	}
	_ = env.bhRepo.Create(ctx, bh)

	// Create one review - quality check should not trigger because review count < 10
	booking := &domain.Booking{
		ID: uuid.New(), UserID: clientID, BathhouseID: bathhouseID,
		StartTime: time.Now().Add(-48 * time.Hour), EndTime: time.Now().Add(-46 * time.Hour),
		GuestCount: 2, TotalPrice: 10000, Status: domain.BookingCompleted,
	}
	_ = env.bookingRepo.Create(ctx, booking)
	_, err := env.svc.Create(ctx, clientID, service.CreateReviewInput{
		BookingID: booking.ID, Rating: 1, Text: "Плохо",
	})
	if err != nil {
		t.Fatalf("create review: %v", err)
	}

	// No quality-related notifications should be sent
	for _, n := range env.notifSvc.sent {
		if n.Type == domain.NotifLowRatingWarning || n.Type == domain.NotifBathhouseDepublished {
			t.Errorf("unexpected quality notification with < 10 reviews: type=%s", n.Type)
		}
	}

	// Bathhouse should still be active
	updatedBh, _ := env.bhRepo.GetByID(ctx, bathhouseID)
	if updatedBh.Status != domain.BathhouseStatusActive {
		t.Errorf("expected status active, got %s", updatedBh.Status)
	}
}

// --- Text Moderation Integration Tests ---

func newReviewTestEnvWithModeration() *reviewTestEnv {
	bhRepo := mock.NewBathhouseRepo()
	bookingRepo := mock.NewBookingRepo()
	reviewRepo := mock.NewReviewRepo()
	repRepo := mock.NewRepresentativeRepo()
	ac := service.NewAccessChecker(repRepo, bhRepo)
	log := logger.New(logger.LevelWarn)
	contentFilter := moderation.NewContentFilter(true, false) // moderation enabled
	textModerator := moderation.NewRegexTextModerator()
	mediaRepo := mock.NewMediaRepo()
	noopStore := storage.NewMockStorage()
	svc := service.NewReviewService(reviewRepo, mock.NewClientReviewRepo(), bookingRepo, bhRepo, mediaRepo, noopStore, ac, &noopNotifService{}, contentFilter, textModerator, nil, log)
	return &reviewTestEnv{
		svc:         svc,
		bhRepo:      bhRepo,
		bookingRepo: bookingRepo,
		reviewRepo:  reviewRepo,
		repRepo:     repRepo,
	}
}

func TestTextModeration_CleanReviewAutoApproved(t *testing.T) {
	env := newReviewTestEnvWithModeration()
	ctx := context.Background()
	clientID := uuid.New()
	bathhouseID := uuid.New()

	bh := &domain.Bathhouse{
		ID: bathhouseID, OwnerID: uuid.New(), Name: "Тест Баня", Slug: "test-mod-1",
		Address: "ул. Тестовая 1", CityID: 1, PricePerHour: 5000,
		MaxGuests: 10, MinDuration: 1, Status: domain.BathhouseStatusActive,
	}
	_ = env.bhRepo.Create(ctx, bh)

	booking := createCompletedBooking(t, env.bookingRepo, clientID, bathhouseID)

	review, err := env.svc.Create(ctx, clientID, service.CreateReviewInput{
		BookingID: booking.ID, Rating: 5, Text: "Отличное место для отдыха! Рекомендую всем.",
	})
	if err != nil {
		t.Fatalf("create review: %v", err)
	}

	if review.Status != domain.ReviewStatusApproved {
		t.Errorf("expected approved, got %s", review.Status)
	}
	if review.ModerationScore == nil {
		t.Fatal("expected moderation score to be set")
	}
	if *review.ModerationScore >= 0.7 {
		t.Errorf("expected score < 0.7, got %f", *review.ModerationScore)
	}
}

func TestTextModeration_ProfanitySetsPending(t *testing.T) {
	env := newReviewTestEnvWithModeration()
	ctx := context.Background()
	clientID := uuid.New()
	bathhouseID := uuid.New()

	bh := &domain.Bathhouse{
		ID: bathhouseID, OwnerID: uuid.New(), Name: "Тест Баня", Slug: "test-mod-2",
		Address: "ул. Тестовая 2", CityID: 1, PricePerHour: 5000,
		MaxGuests: 10, MinDuration: 1, Status: domain.BathhouseStatusActive,
	}
	_ = env.bhRepo.Create(ctx, bh)

	booking := createCompletedBooking(t, env.bookingRepo, clientID, bathhouseID)

	review, err := env.svc.Create(ctx, clientID, service.CreateReviewInput{
		BookingID: booking.ID, Rating: 1,
		Text: "Это просто пиздец, ужасное обслуживание в этом заведении https://example.com",
	})
	if err != nil {
		t.Fatalf("create review: %v", err)
	}

	if review.Status != domain.ReviewStatusPending {
		t.Errorf("expected pending (moderation flagged), got %s", review.Status)
	}
	if review.ModerationScore == nil {
		t.Fatal("expected moderation score to be set")
	}
	if *review.ModerationScore < 0.7 {
		t.Errorf("expected score >= 0.7, got %f", *review.ModerationScore)
	}
	if len(review.ModerationFlags) == 0 {
		t.Error("expected moderation flags to be set")
	}
}

func TestTextModeration_StoredForAdminReference(t *testing.T) {
	env := newReviewTestEnvWithModeration()
	ctx := context.Background()
	clientID := uuid.New()
	bathhouseID := uuid.New()

	bh := &domain.Bathhouse{
		ID: bathhouseID, OwnerID: uuid.New(), Name: "Тест Баня", Slug: "test-mod-3",
		Address: "ул. Тестовая 3", CityID: 1, PricePerHour: 5000,
		MaxGuests: 10, MinDuration: 1, Status: domain.BathhouseStatusActive,
	}
	_ = env.bhRepo.Create(ctx, bh)

	booking := createCompletedBooking(t, env.bookingRepo, clientID, bathhouseID)

	review, err := env.svc.Create(ctx, clientID, service.CreateReviewInput{
		BookingID: booking.ID, Rating: 3,
		Text: "Позвоните на +7 999 123-45-67 для записи в другое место",
	})
	if err != nil {
		t.Fatalf("create review: %v", err)
	}

	// Score should be stored even if not flagged (phone = 0.3, < 0.7)
	if review.ModerationScore == nil {
		t.Fatal("expected moderation score to be stored")
	}

	// Fetch from repo to verify persistence
	stored, err := env.reviewRepo.GetByID(ctx, review.ID)
	if err != nil {
		t.Fatalf("get review: %v", err)
	}
	if stored.ModerationScore == nil {
		t.Error("expected moderation score persisted in repo")
	}
	found := false
	for _, f := range stored.ModerationFlags {
		if f == "contains_phone_numbers" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected phone flag in stored review, got %v", stored.ModerationFlags)
	}
}

func TestTextModeration_DisabledFallsBackToContentFilter(t *testing.T) {
	// When textModerator is nil but contentFilter is enabled, legacy path is used
	bhRepo := mock.NewBathhouseRepo()
	bookingRepo := mock.NewBookingRepo()
	reviewRepo := mock.NewReviewRepo()
	repRepo := mock.NewRepresentativeRepo()
	ac := service.NewAccessChecker(repRepo, bhRepo)
	log := logger.New(logger.LevelWarn)
	contentFilter := moderation.NewContentFilter(true, true) // enabled + auto-approve
	mediaRepo := mock.NewMediaRepo()
	noopStore := storage.NewMockStorage()
	// textModerator is nil - legacy path
	svc := service.NewReviewService(reviewRepo, mock.NewClientReviewRepo(), bookingRepo, bhRepo, mediaRepo, noopStore, ac, &noopNotifService{}, contentFilter, nil, nil, log)

	ctx := context.Background()
	clientID := uuid.New()
	bathhouseID := uuid.New()

	bh := &domain.Bathhouse{
		ID: bathhouseID, OwnerID: uuid.New(), Name: "Тест Баня", Slug: "test-mod-4",
		Address: "ул. Тестовая 4", CityID: 1, PricePerHour: 5000,
		MaxGuests: 10, MinDuration: 1, Status: domain.BathhouseStatusActive,
	}
	_ = bhRepo.Create(ctx, bh)

	booking := createCompletedBooking(t, bookingRepo, clientID, bathhouseID)

	review, err := svc.Create(ctx, clientID, service.CreateReviewInput{
		BookingID: booking.ID, Rating: 4, Text: "Хороший сервис, всё понравилось!",
	})
	if err != nil {
		t.Fatalf("create review: %v", err)
	}

	// Should use legacy content filter auto-approve path
	if review.Status != domain.ReviewStatusApproved {
		t.Errorf("expected approved via legacy path, got %s", review.Status)
	}
	// No moderation score since textModerator was nil
	if review.ModerationScore != nil {
		t.Error("expected nil moderation score when textModerator is nil")
	}
}
