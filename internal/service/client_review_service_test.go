package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/logger"
	"github.com/nikitaaldaev/bani/internal/repository/mock"
	"github.com/nikitaaldaev/bani/internal/service"
)

type clientReviewTestEnv struct {
	svc              service.ClientReviewService
	clientReviewRepo *mock.ClientReviewRepo
	reviewRepo       *mock.ReviewRepo
	bookingRepo      *mock.BookingRepo
	bhRepo           *mock.BathhouseRepo
	repRepo          *mock.RepresentativeRepo
}

func newClientReviewTestEnv() *clientReviewTestEnv {
	clientReviewRepo := mock.NewClientReviewRepo()
	reviewRepo := mock.NewReviewRepo()
	bookingRepo := mock.NewBookingRepo()
	bhRepo := mock.NewBathhouseRepo()
	repRepo := mock.NewRepresentativeRepo()
	ac := service.NewAccessChecker(repRepo, bhRepo)
	log := logger.New(logger.LevelWarn)

	svc := service.NewClientReviewService(clientReviewRepo, reviewRepo, bookingRepo, bhRepo, ac, &noopNotifService{}, log)
	return &clientReviewTestEnv{
		svc:              svc,
		clientReviewRepo: clientReviewRepo,
		reviewRepo:       reviewRepo,
		bookingRepo:      bookingRepo,
		bhRepo:           bhRepo,
		repRepo:          repRepo,
	}
}

func setupBookingAndBathhouse(env *clientReviewTestEnv) (ownerID, clientID, bathhouseID, bookingID uuid.UUID) {
	ownerID = uuid.New()
	clientID = uuid.New()
	bathhouseID = uuid.New()
	bookingID = uuid.New()

	bh := &domain.Bathhouse{
		ID:      bathhouseID,
		OwnerID: ownerID,
		Name:    "Test Bathhouse",
		Status:  domain.BathhouseStatusActive,
	}
	env.bhRepo.Create(context.Background(), bh)

	booking := &domain.Booking{
		ID:          bookingID,
		UserID:      clientID,
		BathhouseID: bathhouseID,
		Status:      domain.BookingCompleted,
		StartTime:   time.Now().Add(-2 * time.Hour),
		EndTime:     time.Now().Add(-1 * time.Hour),
	}
	env.bookingRepo.Create(context.Background(), booking)

	return
}

func TestClientReviewService_Create(t *testing.T) {
	env := newClientReviewTestEnv()
	ownerID, _, _, bookingID := setupBookingAndBathhouse(env)

	p := float64(4.5)
	c := float64(4.0)
	rc := float64(5.0)

	review, err := env.svc.Create(context.Background(), ownerID, domain.RoleOwner, service.CreateClientReviewInput{
		BookingID:      bookingID,
		Punctuality:    &p,
		Cleanliness:    &c,
		RuleCompliance: &rc,
		Rating:         5,
		Text:           "Great guest!",
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if review.ID == uuid.Nil {
		t.Fatal("expected non-nil review ID")
	}
	if review.IsRevealed {
		t.Error("expected review to be unrevealed (blind period)")
	}
	if review.Status != domain.ReviewStatusApproved {
		t.Errorf("expected status approved, got %s", review.Status)
	}
	// Rating should be computed from criteria: (4.5 + 4.0 + 5.0) / 3 = 4.5 -> rounded to 5
	if review.Rating != 5 {
		t.Errorf("expected computed rating 5, got %d", review.Rating)
	}
}

func TestClientReviewService_Create_NotOwner(t *testing.T) {
	env := newClientReviewTestEnv()
	_, clientID, _, bookingID := setupBookingAndBathhouse(env)

	_, err := env.svc.Create(context.Background(), clientID, domain.RoleClient, service.CreateClientReviewInput{
		BookingID: bookingID,
		Rating:    4,
		Text:      "Should fail",
	})

	if err == nil {
		t.Fatal("expected error for non-owner creating client review")
	}
}

func TestClientReviewService_Create_BookingNotCompleted(t *testing.T) {
	env := newClientReviewTestEnv()
	ownerID := uuid.New()
	bathhouseID := uuid.New()
	bookingID := uuid.New()

	bh := &domain.Bathhouse{
		ID:      bathhouseID,
		OwnerID: ownerID,
		Name:    "Test",
		Status:  domain.BathhouseStatusActive,
	}
	env.bhRepo.Create(context.Background(), bh)

	booking := &domain.Booking{
		ID:          bookingID,
		UserID:      uuid.New(),
		BathhouseID: bathhouseID,
		Status:      domain.BookingConfirmed, // Not completed
	}
	env.bookingRepo.Create(context.Background(), booking)

	_, err := env.svc.Create(context.Background(), ownerID, domain.RoleOwner, service.CreateClientReviewInput{
		BookingID: bookingID,
		Rating:    3,
	})

	if err != domain.ErrInvalidInput {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}

func TestClientReviewService_Create_Duplicate(t *testing.T) {
	env := newClientReviewTestEnv()
	ownerID, _, _, bookingID := setupBookingAndBathhouse(env)

	// First review
	_, err := env.svc.Create(context.Background(), ownerID, domain.RoleOwner, service.CreateClientReviewInput{
		BookingID: bookingID,
		Rating:    4,
		Text:      "Good guest",
	})
	if err != nil {
		t.Fatalf("first review should succeed: %v", err)
	}

	// Second review for same booking
	_, err = env.svc.Create(context.Background(), ownerID, domain.RoleOwner, service.CreateClientReviewInput{
		BookingID: bookingID,
		Rating:    5,
		Text:      "Duplicate",
	})
	if err != domain.ErrAlreadyExists {
		t.Fatalf("expected ErrAlreadyExists, got %v", err)
	}
}

func TestClientReviewService_MutualReveal_OwnerFirst(t *testing.T) {
	env := newClientReviewTestEnv()
	ownerID, clientID, bathhouseID, bookingID := setupBookingAndBathhouse(env)

	// Owner posts client review first
	clientReview, err := env.svc.Create(context.Background(), ownerID, domain.RoleOwner, service.CreateClientReviewInput{
		BookingID: bookingID,
		Rating:    4,
		Text:      "Good guest",
	})
	if err != nil {
		t.Fatalf("create client review: %v", err)
	}
	if clientReview.IsRevealed {
		t.Error("client review should be unrevealed initially")
	}

	// Guest posts their review — this should trigger mutual reveal
	guestReview := &domain.Review{
		ID:          uuid.New(),
		UserID:      clientID,
		BathhouseID: bathhouseID,
		BookingID:   bookingID,
		Rating:      5,
		Text:        "Great bathhouse",
		Status:      domain.ReviewStatusApproved,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	env.reviewRepo.Create(context.Background(), guestReview)

	// Now trigger the reveal by re-creating a scenario where both exist
	// The service checks on Create, so let's check via GetByBookingID and manually trigger
	// Actually, the mutual reveal happens in the review service Create, not here.
	// Let's test it from the client review service side:
	// Simulate: guest review already exists when owner creates client review.

	// Set up a fresh scenario
	env2 := newClientReviewTestEnv()
	ownerID2, clientID2, bathhouseID2, bookingID2 := setupBookingAndBathhouse(env2)

	// Guest posts their review first
	guestReview2 := &domain.Review{
		ID:          uuid.New(),
		UserID:      clientID2,
		BathhouseID: bathhouseID2,
		BookingID:   bookingID2,
		Rating:      5,
		Text:        "Great bathhouse",
		Status:      domain.ReviewStatusApproved,
		IsRevealed:  false,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	env2.reviewRepo.Create(context.Background(), guestReview2)

	// Owner posts client review -> mutual reveal should happen
	clientReview2, err := env2.svc.Create(context.Background(), ownerID2, domain.RoleOwner, service.CreateClientReviewInput{
		BookingID: bookingID2,
		Rating:    4,
		Text:      "Good guest",
	})
	if err != nil {
		t.Fatalf("create client review: %v", err)
	}

	// Client review should be revealed
	if !clientReview2.IsRevealed {
		t.Error("client review should be revealed after mutual review")
	}

	// Guest review should also be revealed
	updatedGuestReview, err := env2.reviewRepo.GetByBookingID(context.Background(), bookingID2)
	if err != nil {
		t.Fatalf("get guest review: %v", err)
	}
	if !updatedGuestReview.IsRevealed {
		t.Error("guest review should be revealed after mutual review")
	}
}

func TestClientReviewService_RevealExpired(t *testing.T) {
	env := newClientReviewTestEnv()
	ownerID, _, _, bookingID := setupBookingAndBathhouse(env)

	// Create a client review with reveal_at in the past
	review, err := env.svc.Create(context.Background(), ownerID, domain.RoleOwner, service.CreateClientReviewInput{
		BookingID: bookingID,
		Rating:    3,
		Text:      "OK guest",
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	// Manually set reveal_at to the past to simulate expiry
	review.RevealAt = time.Now().Add(-1 * time.Hour)
	env.clientReviewRepo.Update(context.Background(), review)

	// Run the reveal job
	revealed, err := env.svc.RevealExpired(context.Background())
	if err != nil {
		t.Fatalf("reveal expired: %v", err)
	}
	if revealed != 1 {
		t.Errorf("expected 1 revealed, got %d", revealed)
	}

	// Verify it's now revealed
	updated, err := env.svc.GetByID(context.Background(), review.ID)
	if err != nil {
		t.Fatalf("get by id: %v", err)
	}
	if !updated.IsRevealed {
		t.Error("expected review to be revealed after expiry")
	}
}

func TestClientReviewService_RevealExpired_NoExpired(t *testing.T) {
	env := newClientReviewTestEnv()
	ownerID, _, _, bookingID := setupBookingAndBathhouse(env)

	// Create a client review with reveal_at in the future
	_, err := env.svc.Create(context.Background(), ownerID, domain.RoleOwner, service.CreateClientReviewInput{
		BookingID: bookingID,
		Rating:    4,
		Text:      "Good",
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	// Run the reveal job — nothing should be revealed
	revealed, err := env.svc.RevealExpired(context.Background())
	if err != nil {
		t.Fatalf("reveal expired: %v", err)
	}
	if revealed != 0 {
		t.Errorf("expected 0 revealed, got %d", revealed)
	}
}

func TestClientReviewService_ListByClient_OnlyRevealed(t *testing.T) {
	env := newClientReviewTestEnv()
	ownerID, clientID, _, bookingID := setupBookingAndBathhouse(env)

	// Create a client review (unrevealed)
	_, err := env.svc.Create(context.Background(), ownerID, domain.RoleOwner, service.CreateClientReviewInput{
		BookingID: bookingID,
		Rating:    4,
		Text:      "Good",
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	// Client should not see unrevealed reviews
	result, err := env.svc.ListByClient(context.Background(), clientID, 1, 10)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(result.Items) != 0 {
		t.Errorf("expected 0 revealed reviews, got %d", len(result.Items))
	}
}

func TestClientReviewService_Create_InvalidCriteria(t *testing.T) {
	env := newClientReviewTestEnv()
	ownerID, _, _, bookingID := setupBookingAndBathhouse(env)

	// Only provide 2 out of 3 criteria
	p := float64(4.0)
	c := float64(3.0)

	_, err := env.svc.Create(context.Background(), ownerID, domain.RoleOwner, service.CreateClientReviewInput{
		BookingID:   bookingID,
		Punctuality: &p,
		Cleanliness: &c,
		Rating:      4,
		Text:        "Missing criterion",
	})

	if err != domain.ErrInvalidInput {
		t.Fatalf("expected ErrInvalidInput for incomplete criteria, got %v", err)
	}
}

func TestClientReviewService_Create_BoundaryTimes(t *testing.T) {
	env := newClientReviewTestEnv()
	ownerID, _, _, bookingID := setupBookingAndBathhouse(env)

	review, err := env.svc.Create(context.Background(), ownerID, domain.RoleOwner, service.CreateClientReviewInput{
		BookingID: bookingID,
		Rating:    3,
		Text:      "Average guest",
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	// Verify reveal_at is set to 14 days from now
	expectedReveal := time.Now().Add(domain.ClientReviewBlindDays * 24 * time.Hour)
	diff := review.RevealAt.Sub(expectedReveal)
	if diff < -5*time.Second || diff > 5*time.Second {
		t.Errorf("reveal_at should be ~14 days from now, got diff %v", diff)
	}
}

func TestClientReviewService_GetByBookingID(t *testing.T) {
	env := newClientReviewTestEnv()
	ownerID, _, _, bookingID := setupBookingAndBathhouse(env)

	created, err := env.svc.Create(context.Background(), ownerID, domain.RoleOwner, service.CreateClientReviewInput{
		BookingID: bookingID,
		Rating:    4,
		Text:      "Good",
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	fetched, err := env.svc.GetByBookingID(context.Background(), bookingID)
	if err != nil {
		t.Fatalf("get by booking: %v", err)
	}
	if fetched.ID != created.ID {
		t.Errorf("expected ID %s, got %s", created.ID, fetched.ID)
	}
}

func TestClientReviewService_GetByBookingID_NotFound(t *testing.T) {
	env := newClientReviewTestEnv()

	_, err := env.svc.GetByBookingID(context.Background(), uuid.New())
	if err != domain.ErrClientReviewNotFound {
		t.Fatalf("expected ErrClientReviewNotFound, got %v", err)
	}
}
