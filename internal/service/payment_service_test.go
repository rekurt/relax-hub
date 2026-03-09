package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/logger"
	"github.com/nikitaaldaev/bani/internal/payment"
	"github.com/nikitaaldaev/bani/internal/repository/mock"
	"github.com/nikitaaldaev/bani/internal/service"
)

func newPaymentService() (service.PaymentService, *mock.PaymentRepo, *mock.BookingRepo, *payment.MockProvider) {
	paymentRepo := mock.NewPaymentRepo().(*mock.PaymentRepo)
	bookingRepo := mock.NewBookingRepo()
	provider := payment.NewMockProvider()
	log := logger.New(logger.LevelWarn)
	svc := service.NewPaymentService(paymentRepo, bookingRepo, provider, "http://localhost:3000/callback", log)
	return svc, paymentRepo, bookingRepo, provider
}

func createTestBooking(t *testing.T, bookingRepo *mock.BookingRepo, userID, bathhouseID uuid.UUID, start time.Time, status domain.BookingStatus) *domain.Booking {
	t.Helper()
	booking := &domain.Booking{
		ID:          uuid.New(),
		UserID:      userID,
		BathhouseID: bathhouseID,
		StartTime:   start,
		EndTime:     start.Add(2 * time.Hour),
		GuestCount:  2,
		TotalPrice:  10000,
		Status:      status,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	if err := bookingRepo.Create(context.Background(), booking); err != nil {
		t.Fatal(err)
	}
	return booking
}

func TestPaymentService_InitiatePayment_Success(t *testing.T) {
	svc, _, bookingRepo, _ := newPaymentService()
	userID := uuid.New()
	booking := createTestBooking(t, bookingRepo, userID, uuid.New(), time.Now().Add(24*time.Hour), domain.BookingPending)

	url, err := svc.InitiatePayment(context.Background(), booking.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if url == "" {
		t.Error("expected non-empty confirmation URL")
	}
}

func TestPaymentService_InitiatePayment_BookingNotFound(t *testing.T) {
	svc, _, _, _ := newPaymentService()

	_, err := svc.InitiatePayment(context.Background(), uuid.New())
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got: %v", err)
	}
}

func TestPaymentService_InitiatePayment_InvalidStatus(t *testing.T) {
	svc, _, bookingRepo, _ := newPaymentService()
	userID := uuid.New()
	booking := createTestBooking(t, bookingRepo, userID, uuid.New(), time.Now().Add(24*time.Hour), domain.BookingCancelled)

	_, err := svc.InitiatePayment(context.Background(), booking.ID)
	if !errors.Is(err, domain.ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput for cancelled booking, got: %v", err)
	}
}

func TestPaymentService_InitiatePayment_AlreadyPaid(t *testing.T) {
	svc, paymentRepo, bookingRepo, _ := newPaymentService()
	userID := uuid.New()
	booking := createTestBooking(t, bookingRepo, userID, uuid.New(), time.Now().Add(24*time.Hour), domain.BookingPending)

	// Create an existing succeeded payment
	now := time.Now()
	existingPayment := &domain.Payment{
		ID:        uuid.New(),
		BookingID: booking.ID,
		UserID:    userID,
		Amount:    10000,
		Currency:  "RUB",
		Status:    domain.PaymentSucceeded,
		Provider:  "yookassa",
		CreatedAt: now,
		UpdatedAt: now,
	}
	_ = paymentRepo.Create(context.Background(), existingPayment)

	_, err := svc.InitiatePayment(context.Background(), booking.ID)
	if !errors.Is(err, domain.ErrPaymentAlreadyProcessed) {
		t.Errorf("expected ErrPaymentAlreadyProcessed, got: %v", err)
	}
}

func TestPaymentService_HandleWebhook_Succeeded(t *testing.T) {
	svc, paymentRepo, bookingRepo, provider := newPaymentService()
	userID := uuid.New()
	booking := createTestBooking(t, bookingRepo, userID, uuid.New(), time.Now().Add(24*time.Hour), domain.BookingPending)

	// Initiate payment to get external ID
	_, err := svc.InitiatePayment(context.Background(), booking.ID)
	if err != nil {
		t.Fatalf("initiate failed: %v", err)
	}

	// Find the payment to get the external ID
	p, err := paymentRepo.GetByBookingID(context.Background(), booking.ID)
	if err != nil {
		t.Fatalf("failed to get payment: %v", err)
	}

	// Simulate provider confirming the payment
	provider.SetPaymentStatus(p.ExternalID, "succeeded")

	err = svc.HandleWebhook(context.Background(), service.WebhookEvent{
		ExternalID: p.ExternalID,
		Status:     "succeeded",
	})
	if err != nil {
		t.Fatalf("webhook handling failed: %v", err)
	}

	// Verify payment status updated
	updated, _ := paymentRepo.GetByID(context.Background(), p.ID)
	if updated.Status != domain.PaymentSucceeded {
		t.Errorf("payment status = %q, want %q", updated.Status, domain.PaymentSucceeded)
	}

	// Verify booking confirmed
	updatedBooking, _ := bookingRepo.GetByID(context.Background(), booking.ID)
	if updatedBooking.Status != domain.BookingConfirmed {
		t.Errorf("booking status = %q, want %q", updatedBooking.Status, domain.BookingConfirmed)
	}
}

func TestPaymentService_HandleWebhook_Canceled(t *testing.T) {
	svc, paymentRepo, bookingRepo, _ := newPaymentService()
	userID := uuid.New()
	booking := createTestBooking(t, bookingRepo, userID, uuid.New(), time.Now().Add(24*time.Hour), domain.BookingPending)

	// Initiate payment
	_, err := svc.InitiatePayment(context.Background(), booking.ID)
	if err != nil {
		t.Fatalf("initiate failed: %v", err)
	}

	p, _ := paymentRepo.GetByBookingID(context.Background(), booking.ID)

	err = svc.HandleWebhook(context.Background(), service.WebhookEvent{
		ExternalID: p.ExternalID,
		Status:     "canceled",
	})
	if err != nil {
		t.Fatalf("webhook handling failed: %v", err)
	}

	updated, _ := paymentRepo.GetByID(context.Background(), p.ID)
	if updated.Status != domain.PaymentFailed {
		t.Errorf("payment status = %q, want %q", updated.Status, domain.PaymentFailed)
	}
}

func TestPaymentService_HandleWebhook_AlreadyProcessed(t *testing.T) {
	svc, paymentRepo, bookingRepo, _ := newPaymentService()
	userID := uuid.New()
	booking := createTestBooking(t, bookingRepo, userID, uuid.New(), time.Now().Add(24*time.Hour), domain.BookingPending)

	_, _ = svc.InitiatePayment(context.Background(), booking.ID)
	p, _ := paymentRepo.GetByBookingID(context.Background(), booking.ID)

	// First webhook - succeeds
	_ = svc.HandleWebhook(context.Background(), service.WebhookEvent{
		ExternalID: p.ExternalID,
		Status:     "succeeded",
	})

	// Second webhook - should return already processed
	err := svc.HandleWebhook(context.Background(), service.WebhookEvent{
		ExternalID: p.ExternalID,
		Status:     "succeeded",
	})
	if !errors.Is(err, domain.ErrPaymentAlreadyProcessed) {
		t.Errorf("expected ErrPaymentAlreadyProcessed, got: %v", err)
	}
}

func TestPaymentService_HandleWebhook_NotFound(t *testing.T) {
	svc, _, _, _ := newPaymentService()

	err := svc.HandleWebhook(context.Background(), service.WebhookEvent{
		ExternalID: "nonexistent-id",
		Status:     "succeeded",
	})
	if !errors.Is(err, domain.ErrPaymentNotFound) {
		t.Errorf("expected ErrPaymentNotFound, got: %v", err)
	}
}

func TestPaymentService_RefundPayment_FullRefund(t *testing.T) {
	svc, paymentRepo, bookingRepo, provider := newPaymentService()
	userID := uuid.New()
	// Booking starts in 48 hours - should get full refund
	booking := createTestBooking(t, bookingRepo, userID, uuid.New(), time.Now().Add(48*time.Hour), domain.BookingPending)

	_, _ = svc.InitiatePayment(context.Background(), booking.ID)
	p, _ := paymentRepo.GetByBookingID(context.Background(), booking.ID)

	// Simulate successful payment
	provider.SetPaymentStatus(p.ExternalID, "succeeded")
	_ = svc.HandleWebhook(context.Background(), service.WebhookEvent{
		ExternalID: p.ExternalID,
		Status:     "succeeded",
	})

	// Now refund
	err := svc.RefundPayment(context.Background(), booking.ID)
	if err != nil {
		t.Fatalf("refund failed: %v", err)
	}

	updated, _ := paymentRepo.GetByID(context.Background(), p.ID)
	if updated.RefundAmount != 10000 {
		t.Errorf("refund amount = %d, want 10000 (full refund)", updated.RefundAmount)
	}
	if updated.RefundedAt == nil {
		t.Error("refunded_at should be set")
	}
}

func TestPaymentService_RefundPayment_PartialRefund(t *testing.T) {
	svc, paymentRepo, bookingRepo, provider := newPaymentService()
	userID := uuid.New()
	// Booking starts in 12 hours - between 2h and 24h, should get 50% refund
	booking := createTestBooking(t, bookingRepo, userID, uuid.New(), time.Now().Add(12*time.Hour), domain.BookingPending)

	_, _ = svc.InitiatePayment(context.Background(), booking.ID)
	p, _ := paymentRepo.GetByBookingID(context.Background(), booking.ID)

	provider.SetPaymentStatus(p.ExternalID, "succeeded")
	_ = svc.HandleWebhook(context.Background(), service.WebhookEvent{
		ExternalID: p.ExternalID,
		Status:     "succeeded",
	})

	err := svc.RefundPayment(context.Background(), booking.ID)
	if err != nil {
		t.Fatalf("refund failed: %v", err)
	}

	updated, _ := paymentRepo.GetByID(context.Background(), p.ID)
	if updated.RefundAmount != 5000 {
		t.Errorf("refund amount = %d, want 5000 (50%% refund)", updated.RefundAmount)
	}
}

func TestPaymentService_RefundPayment_NoRefundTooLate(t *testing.T) {
	svc, paymentRepo, bookingRepo, provider := newPaymentService()
	userID := uuid.New()
	// Booking starts in 1 hour - less than 2h deadline, no refund
	booking := createTestBooking(t, bookingRepo, userID, uuid.New(), time.Now().Add(1*time.Hour), domain.BookingPending)

	_, _ = svc.InitiatePayment(context.Background(), booking.ID)
	p, _ := paymentRepo.GetByBookingID(context.Background(), booking.ID)

	provider.SetPaymentStatus(p.ExternalID, "succeeded")
	_ = svc.HandleWebhook(context.Background(), service.WebhookEvent{
		ExternalID: p.ExternalID,
		Status:     "succeeded",
	})

	err := svc.RefundPayment(context.Background(), booking.ID)
	if err != nil {
		t.Fatalf("expected no error (just no refund), got: %v", err)
	}

	updated, _ := paymentRepo.GetByID(context.Background(), p.ID)
	if updated.RefundAmount != 0 {
		t.Errorf("refund amount = %d, want 0 (no refund)", updated.RefundAmount)
	}
}

func TestPaymentService_RefundPayment_NotSucceeded(t *testing.T) {
	svc, _, bookingRepo, _ := newPaymentService()
	userID := uuid.New()
	booking := createTestBooking(t, bookingRepo, userID, uuid.New(), time.Now().Add(48*time.Hour), domain.BookingPending)

	// Initiate but don't complete payment
	_, _ = svc.InitiatePayment(context.Background(), booking.ID)

	err := svc.RefundPayment(context.Background(), booking.ID)
	if !errors.Is(err, domain.ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput for non-succeeded payment, got: %v", err)
	}
}

func TestPaymentService_RefundPayment_NoPayment(t *testing.T) {
	svc, _, bookingRepo, _ := newPaymentService()
	userID := uuid.New()
	booking := createTestBooking(t, bookingRepo, userID, uuid.New(), time.Now().Add(48*time.Hour), domain.BookingPending)

	err := svc.RefundPayment(context.Background(), booking.ID)
	if !errors.Is(err, domain.ErrPaymentNotFound) {
		t.Errorf("expected ErrPaymentNotFound, got: %v", err)
	}
}

func TestPaymentService_GetPaymentByBooking(t *testing.T) {
	svc, _, bookingRepo, _ := newPaymentService()
	userID := uuid.New()
	booking := createTestBooking(t, bookingRepo, userID, uuid.New(), time.Now().Add(24*time.Hour), domain.BookingPending)

	_, _ = svc.InitiatePayment(context.Background(), booking.ID)

	p, err := svc.GetPaymentByBooking(context.Background(), booking.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.BookingID != booking.ID {
		t.Errorf("booking ID = %v, want %v", p.BookingID, booking.ID)
	}
}

func TestPaymentService_ListUserPayments(t *testing.T) {
	svc, _, bookingRepo, _ := newPaymentService()
	userID := uuid.New()

	// Create two bookings and initiate payment for both
	b1 := createTestBooking(t, bookingRepo, userID, uuid.New(), time.Now().Add(24*time.Hour), domain.BookingPending)
	b2 := createTestBooking(t, bookingRepo, userID, uuid.New(), time.Now().Add(48*time.Hour), domain.BookingPending)
	_, _ = svc.InitiatePayment(context.Background(), b1.ID)
	_, _ = svc.InitiatePayment(context.Background(), b2.ID)

	result, err := svc.ListUserPayments(context.Background(), userID, 1, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.TotalCount != 2 {
		t.Errorf("total count = %d, want 2", result.TotalCount)
	}
}
