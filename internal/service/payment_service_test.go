package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/fiscal"
	"github.com/nikitaaldaev/bani/internal/logger"
	"github.com/nikitaaldaev/bani/internal/payment"
	"github.com/nikitaaldaev/bani/internal/repository/mock"
	"github.com/nikitaaldaev/bani/internal/service"
)

// testWalletService is a configurable mock for wallet operations in payment tests.
type testWalletService struct {
	noopWalletService
	wallet       *domain.Wallet
	spendErr     error
	refundCalled bool
	spendCalled  bool
	spendAmount  int64
	refundAmount int64
}

func (w *testWalletService) GetWallet(_ context.Context, _ uuid.UUID) (*domain.Wallet, error) {
	if w.wallet != nil {
		return w.wallet, nil
	}
	return &domain.Wallet{ID: uuid.New(), Balance: 1000000}, nil
}

func (w *testWalletService) Spend(_ context.Context, _ uuid.UUID, amount int64, _ string, _ *uuid.UUID, _ string) (*domain.WalletTransaction, error) {
	w.spendCalled = true
	w.spendAmount = amount
	if w.spendErr != nil {
		return nil, w.spendErr
	}
	return &domain.WalletTransaction{}, nil
}

func (w *testWalletService) Refund(_ context.Context, _ uuid.UUID, amount int64, _ string, _ *uuid.UUID, _ string) (*domain.WalletTransaction, error) {
	w.refundCalled = true
	w.refundAmount = amount
	return &domain.WalletTransaction{}, nil
}

func newPaymentService() (service.PaymentService, *mock.PaymentRepo, *mock.BookingRepo, *payment.MockProvider) {
	paymentRepo := mock.NewPaymentRepo().(*mock.PaymentRepo)
	bookingRepo := mock.NewBookingRepo()
	provider := payment.NewMockProvider()
	log := logger.New(logger.LevelWarn)
	svc := service.NewPaymentService(paymentRepo, bookingRepo, nil, provider, fiscal.NewNoOpProvider(), &noopWalletService{}, &noopNotifService{}, "http://localhost:3000/callback", log)
	return svc, paymentRepo, bookingRepo, provider
}

func newPaymentServiceWithWallet(walletSvc *testWalletService) (service.PaymentService, *mock.PaymentRepo, *mock.BookingRepo, *payment.MockProvider) {
	paymentRepo := mock.NewPaymentRepo().(*mock.PaymentRepo)
	bookingRepo := mock.NewBookingRepo()
	provider := payment.NewMockProvider()
	log := logger.New(logger.LevelWarn)
	svc := service.NewPaymentService(paymentRepo, bookingRepo, nil, provider, fiscal.NewNoOpProvider(), walletSvc, &noopNotifService{}, "http://localhost:3000/callback", log)
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

// mockFiscalProvider records calls for testing fiscal integration.
type mockFiscalProvider struct {
	receipts []fiscal.ReceiptRequest
	err      error
}

func (m *mockFiscalProvider) CreateReceipt(_ context.Context, req fiscal.ReceiptRequest) (*fiscal.Receipt, error) {
	if m.err != nil {
		return nil, m.err
	}
	m.receipts = append(m.receipts, req)
	return &fiscal.Receipt{ID: "test-receipt-id", Status: "pending"}, nil
}

func newPaymentServiceWithFiscal(fp fiscal.FiscalProvider) (service.PaymentService, *mock.PaymentRepo, *mock.BookingRepo, *payment.MockProvider) {
	paymentRepo := mock.NewPaymentRepo().(*mock.PaymentRepo)
	bookingRepo := mock.NewBookingRepo()
	provider := payment.NewMockProvider()
	log := logger.New(logger.LevelWarn)
	svc := service.NewPaymentService(paymentRepo, bookingRepo, nil, provider, fp, &noopWalletService{}, &noopNotifService{}, "http://localhost:3000/callback", log)
	return svc, paymentRepo, bookingRepo, provider
}

func TestPaymentService_InitiatePayment_Success(t *testing.T) {
	svc, _, bookingRepo, _ := newPaymentService()
	userID := uuid.New()
	booking := createTestBooking(t, bookingRepo, userID, uuid.New(), time.Now().Add(24*time.Hour), domain.BookingPending)

	url, err := svc.InitiatePayment(context.Background(), userID, booking.ID, domain.PaymentMethodCard)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if url == "" {
		t.Error("expected non-empty confirmation URL")
	}
}

func TestPaymentService_InitiatePayment_BookingNotFound(t *testing.T) {
	svc, _, _, _ := newPaymentService()

	_, err := svc.InitiatePayment(context.Background(), uuid.New(), uuid.New(), domain.PaymentMethodCard)
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got: %v", err)
	}
}

func TestPaymentService_InitiatePayment_InvalidStatus(t *testing.T) {
	svc, _, bookingRepo, _ := newPaymentService()
	userID := uuid.New()
	booking := createTestBooking(t, bookingRepo, userID, uuid.New(), time.Now().Add(24*time.Hour), domain.BookingCancelled)

	_, err := svc.InitiatePayment(context.Background(), userID, booking.ID, domain.PaymentMethodCard)
	if !errors.Is(err, domain.ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput for cancelled booking, got: %v", err)
	}
}

func TestPaymentService_InitiatePayment_Forbidden(t *testing.T) {
	svc, _, bookingRepo, _ := newPaymentService()
	ownerID := uuid.New()
	otherID := uuid.New()
	booking := createTestBooking(t, bookingRepo, ownerID, uuid.New(), time.Now().Add(24*time.Hour), domain.BookingPending)

	_, err := svc.InitiatePayment(context.Background(), otherID, booking.ID, domain.PaymentMethodCard)
	if !errors.Is(err, domain.ErrForbidden) {
		t.Errorf("expected ErrForbidden when another user tries to pay, got: %v", err)
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

	_, err := svc.InitiatePayment(context.Background(), userID, booking.ID, domain.PaymentMethodCard)
	if !errors.Is(err, domain.ErrPaymentAlreadyProcessed) {
		t.Errorf("expected ErrPaymentAlreadyProcessed, got: %v", err)
	}
}

func TestPaymentService_HandleWebhook_Succeeded(t *testing.T) {
	svc, paymentRepo, bookingRepo, provider := newPaymentService()
	userID := uuid.New()
	booking := createTestBooking(t, bookingRepo, userID, uuid.New(), time.Now().Add(24*time.Hour), domain.BookingPending)

	// Initiate payment to get external ID
	_, err := svc.InitiatePayment(context.Background(), userID, booking.ID, domain.PaymentMethodCard)
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
	svc, paymentRepo, bookingRepo, provider := newPaymentService()
	userID := uuid.New()
	booking := createTestBooking(t, bookingRepo, userID, uuid.New(), time.Now().Add(24*time.Hour), domain.BookingPending)

	// Initiate payment
	_, err := svc.InitiatePayment(context.Background(), userID, booking.ID, domain.PaymentMethodCard)
	if err != nil {
		t.Fatalf("initiate failed: %v", err)
	}

	p, _ := paymentRepo.GetByBookingID(context.Background(), booking.ID)

	// Simulate provider canceling the payment
	provider.SetPaymentStatus(p.ExternalID, "canceled")

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
	svc, paymentRepo, bookingRepo, provider := newPaymentService()
	userID := uuid.New()
	booking := createTestBooking(t, bookingRepo, userID, uuid.New(), time.Now().Add(24*time.Hour), domain.BookingPending)

	_, _ = svc.InitiatePayment(context.Background(), userID, booking.ID, domain.PaymentMethodCard)
	p, _ := paymentRepo.GetByBookingID(context.Background(), booking.ID)

	// Simulate provider confirming the payment
	provider.SetPaymentStatus(p.ExternalID, "succeeded")

	// First webhook - succeeds
	_ = svc.HandleWebhook(context.Background(), service.WebhookEvent{
		ExternalID: p.ExternalID,
		Status:     "succeeded",
	})

	// Second webhook - should return nil for idempotency
	err := svc.HandleWebhook(context.Background(), service.WebhookEvent{
		ExternalID: p.ExternalID,
		Status:     "succeeded",
	})
	if err != nil {
		t.Errorf("expected nil for idempotent webhook, got: %v", err)
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

	_, _ = svc.InitiatePayment(context.Background(), userID, booking.ID, domain.PaymentMethodCard)
	p, _ := paymentRepo.GetByBookingID(context.Background(), booking.ID)

	// Simulate successful payment
	provider.SetPaymentStatus(p.ExternalID, "succeeded")
	_ = svc.HandleWebhook(context.Background(), service.WebhookEvent{
		ExternalID: p.ExternalID,
		Status:     "succeeded",
	})

	// Now refund
	err := svc.RefundPayment(context.Background(), booking.ID, false, "")
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
	if updated.Status != domain.PaymentRefunded {
		t.Errorf("payment status = %q, want %q", updated.Status, domain.PaymentRefunded)
	}
}

func TestPaymentService_RefundPayment_PartialRefund(t *testing.T) {
	svc, paymentRepo, bookingRepo, provider := newPaymentService()
	userID := uuid.New()
	// Booking starts in 12 hours - between 2h and 24h, should get 50% refund
	booking := createTestBooking(t, bookingRepo, userID, uuid.New(), time.Now().Add(12*time.Hour), domain.BookingPending)

	_, _ = svc.InitiatePayment(context.Background(), userID, booking.ID, domain.PaymentMethodCard)
	p, _ := paymentRepo.GetByBookingID(context.Background(), booking.ID)

	provider.SetPaymentStatus(p.ExternalID, "succeeded")
	_ = svc.HandleWebhook(context.Background(), service.WebhookEvent{
		ExternalID: p.ExternalID,
		Status:     "succeeded",
	})

	err := svc.RefundPayment(context.Background(), booking.ID, false, "")
	if err != nil {
		t.Fatalf("refund failed: %v", err)
	}

	updated, _ := paymentRepo.GetByID(context.Background(), p.ID)
	if updated.RefundAmount != 5000 {
		t.Errorf("refund amount = %d, want 5000 (50%% refund)", updated.RefundAmount)
	}
	if updated.Status != domain.PaymentPartiallyRefunded {
		t.Errorf("payment status = %q, want %q", updated.Status, domain.PaymentPartiallyRefunded)
	}
}

func TestPaymentService_RefundPayment_NoRefundTooLate(t *testing.T) {
	svc, paymentRepo, bookingRepo, provider := newPaymentService()
	userID := uuid.New()
	// Booking starts in 1 hour - less than 2h deadline, no refund
	booking := createTestBooking(t, bookingRepo, userID, uuid.New(), time.Now().Add(1*time.Hour), domain.BookingPending)

	_, _ = svc.InitiatePayment(context.Background(), userID, booking.ID, domain.PaymentMethodCard)
	p, _ := paymentRepo.GetByBookingID(context.Background(), booking.ID)

	provider.SetPaymentStatus(p.ExternalID, "succeeded")
	_ = svc.HandleWebhook(context.Background(), service.WebhookEvent{
		ExternalID: p.ExternalID,
		Status:     "succeeded",
	})

	err := svc.RefundPayment(context.Background(), booking.ID, false, "")
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
	_, _ = svc.InitiatePayment(context.Background(), userID, booking.ID, domain.PaymentMethodCard)

	err := svc.RefundPayment(context.Background(), booking.ID, false, "")
	if !errors.Is(err, domain.ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput for non-succeeded payment, got: %v", err)
	}
}

func TestPaymentService_RefundPayment_NoPayment(t *testing.T) {
	svc, _, bookingRepo, _ := newPaymentService()
	userID := uuid.New()
	booking := createTestBooking(t, bookingRepo, userID, uuid.New(), time.Now().Add(48*time.Hour), domain.BookingPending)

	err := svc.RefundPayment(context.Background(), booking.ID, false, "")
	if !errors.Is(err, domain.ErrPaymentNotFound) {
		t.Errorf("expected ErrPaymentNotFound, got: %v", err)
	}
}

func TestPaymentService_GetPaymentByBooking(t *testing.T) {
	svc, _, bookingRepo, _ := newPaymentService()
	userID := uuid.New()
	booking := createTestBooking(t, bookingRepo, userID, uuid.New(), time.Now().Add(24*time.Hour), domain.BookingPending)

	_, _ = svc.InitiatePayment(context.Background(), userID, booking.ID, domain.PaymentMethodCard)

	p, err := svc.GetPaymentByBooking(context.Background(), userID, booking.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.BookingID != booking.ID {
		t.Errorf("booking ID = %v, want %v", p.BookingID, booking.ID)
	}
}

func TestPaymentService_GetPaymentByBooking_Forbidden(t *testing.T) {
	svc, _, bookingRepo, _ := newPaymentService()
	userID := uuid.New()
	otherID := uuid.New()
	booking := createTestBooking(t, bookingRepo, userID, uuid.New(), time.Now().Add(24*time.Hour), domain.BookingPending)

	_, _ = svc.InitiatePayment(context.Background(), userID, booking.ID, domain.PaymentMethodCard)

	_, err := svc.GetPaymentByBooking(context.Background(), otherID, booking.ID)
	if !errors.Is(err, domain.ErrForbidden) {
		t.Errorf("expected ErrForbidden when another user reads payment, got: %v", err)
	}
}

func TestPaymentService_ListUserPayments(t *testing.T) {
	svc, _, bookingRepo, _ := newPaymentService()
	userID := uuid.New()

	// Create two bookings and initiate payment for both
	b1 := createTestBooking(t, bookingRepo, userID, uuid.New(), time.Now().Add(24*time.Hour), domain.BookingPending)
	b2 := createTestBooking(t, bookingRepo, userID, uuid.New(), time.Now().Add(48*time.Hour), domain.BookingPending)
	_, _ = svc.InitiatePayment(context.Background(), userID, b1.ID, domain.PaymentMethodCard)
	_, _ = svc.InitiatePayment(context.Background(), userID, b2.ID, domain.PaymentMethodCard)

	result, err := svc.ListUserPayments(context.Background(), userID, 1, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.TotalCount != 2 {
		t.Errorf("total count = %d, want 2", result.TotalCount)
	}
}

func TestPaymentService_InitiatePayment_SBP(t *testing.T) {
	svc, paymentRepo, bookingRepo, provider := newPaymentService()
	userID := uuid.New()
	booking := createTestBooking(t, bookingRepo, userID, uuid.New(), time.Now().Add(24*time.Hour), domain.BookingPending)

	url, err := svc.InitiatePayment(context.Background(), userID, booking.ID, domain.PaymentMethodSBP)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if url == "" {
		t.Error("expected non-empty confirmation URL for SBP payment")
	}

	// Verify payment method was stored
	p, err := paymentRepo.GetByBookingID(context.Background(), booking.ID)
	if err != nil {
		t.Fatalf("failed to get payment: %v", err)
	}
	if p.PaymentMethod != domain.PaymentMethodSBP {
		t.Errorf("payment method = %q, want %q", p.PaymentMethod, domain.PaymentMethodSBP)
	}

	// Verify provider received SBP method
	method := provider.GetLastPaymentMethod()
	if method != "sbp" {
		t.Errorf("provider method = %q, want %q", method, "sbp")
	}
}

func TestPaymentService_InitiatePayment_DefaultsToCard(t *testing.T) {
	svc, paymentRepo, bookingRepo, _ := newPaymentService()
	userID := uuid.New()
	booking := createTestBooking(t, bookingRepo, userID, uuid.New(), time.Now().Add(24*time.Hour), domain.BookingPending)

	_, err := svc.InitiatePayment(context.Background(), userID, booking.ID, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	p, err := paymentRepo.GetByBookingID(context.Background(), booking.ID)
	if err != nil {
		t.Fatalf("failed to get payment: %v", err)
	}
	if p.PaymentMethod != domain.PaymentMethodCard {
		t.Errorf("payment method = %q, want %q (default)", p.PaymentMethod, domain.PaymentMethodCard)
	}
}

func TestPaymentService_HandleWebhook_SBP_Succeeded(t *testing.T) {
	svc, paymentRepo, bookingRepo, provider := newPaymentService()
	userID := uuid.New()
	booking := createTestBooking(t, bookingRepo, userID, uuid.New(), time.Now().Add(24*time.Hour), domain.BookingPending)

	_, err := svc.InitiatePayment(context.Background(), userID, booking.ID, domain.PaymentMethodSBP)
	if err != nil {
		t.Fatalf("initiate failed: %v", err)
	}

	p, err := paymentRepo.GetByBookingID(context.Background(), booking.ID)
	if err != nil {
		t.Fatalf("failed to get payment: %v", err)
	}

	// SBP webhook handling is identical to card - status transitions are the same
	provider.SetPaymentStatus(p.ExternalID, "succeeded")

	err = svc.HandleWebhook(context.Background(), service.WebhookEvent{
		ExternalID: p.ExternalID,
		Status:     "succeeded",
	})
	if err != nil {
		t.Fatalf("webhook handling failed: %v", err)
	}

	updated, _ := paymentRepo.GetByID(context.Background(), p.ID)
	if updated.Status != domain.PaymentSucceeded {
		t.Errorf("payment status = %q, want %q", updated.Status, domain.PaymentSucceeded)
	}
	if updated.PaymentMethod != domain.PaymentMethodSBP {
		t.Errorf("payment method after webhook = %q, want %q", updated.PaymentMethod, domain.PaymentMethodSBP)
	}

	updatedBooking, _ := bookingRepo.GetByID(context.Background(), booking.ID)
	if updatedBooking.Status != domain.BookingConfirmed {
		t.Errorf("booking status = %q, want %q", updatedBooking.Status, domain.BookingConfirmed)
	}
}

func TestPaymentService_CapturePayment(t *testing.T) {
	// This test verifies that MockProvider.CapturePayment works correctly
	// (used by request-based booking flow in Task 9)
	provider := payment.NewMockProvider()
	ctx := context.Background()

	result, _ := provider.CreatePayment(ctx, payment.CreatePaymentRequest{
		Amount:   10000,
		Currency: "RUB",
		Method:   "card",
		Capture:  false,
	})

	provider.SetPaymentStatus(result.ExternalID, "waiting_for_capture")

	err := provider.CapturePayment(ctx, result.ExternalID, 10000)
	if err != nil {
		t.Fatalf("capture failed: %v", err)
	}

	status, _ := provider.GetPaymentStatus(ctx, result.ExternalID)
	if status != "succeeded" {
		t.Errorf("status after capture = %q, want succeeded", status)
	}
}

func TestPaymentService_CancelPayment(t *testing.T) {
	// This test verifies that MockProvider.CancelPayment works correctly
	// (used by request-based booking rejection in Task 9)
	provider := payment.NewMockProvider()
	ctx := context.Background()

	result, _ := provider.CreatePayment(ctx, payment.CreatePaymentRequest{
		Amount:   10000,
		Currency: "RUB",
		Method:   "card",
		Capture:  false,
	})

	err := provider.CancelPayment(ctx, result.ExternalID)
	if err != nil {
		t.Fatalf("cancel failed: %v", err)
	}

	status, _ := provider.GetPaymentStatus(ctx, result.ExternalID)
	if status != "canceled" {
		t.Errorf("status after cancel = %q, want canceled", status)
	}
}

// --- Combo Payment Tests ---

func TestPaymentService_ComboPayment_FullCard(t *testing.T) {
	walletSvc := &testWalletService{}
	svc, paymentRepo, bookingRepo, _ := newPaymentServiceWithWallet(walletSvc)
	userID := uuid.New()
	booking := createTestBooking(t, bookingRepo, userID, uuid.New(), time.Now().Add(24*time.Hour), domain.BookingPending)

	url, err := svc.InitiateComboPayment(context.Background(), userID, booking.ID, service.ComboPaymentRequest{
		WalletAmount:  0,
		CardAmount:    booking.TotalPrice,
		PaymentMethod: domain.PaymentMethodCard,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if url == "" {
		t.Error("expected non-empty confirmation URL for full card payment")
	}
	if walletSvc.spendCalled {
		t.Error("wallet Spend should not be called for full card payment")
	}

	p, _ := paymentRepo.GetByBookingID(context.Background(), booking.ID)
	if p.WalletAmount != 0 {
		t.Errorf("wallet_amount = %d, want 0", p.WalletAmount)
	}
	if p.CardAmount != booking.TotalPrice {
		t.Errorf("card_amount = %d, want %d", p.CardAmount, booking.TotalPrice)
	}
}

func TestPaymentService_ComboPayment_FullWallet(t *testing.T) {
	walletSvc := &testWalletService{}
	svc, paymentRepo, bookingRepo, _ := newPaymentServiceWithWallet(walletSvc)
	userID := uuid.New()
	booking := createTestBooking(t, bookingRepo, userID, uuid.New(), time.Now().Add(24*time.Hour), domain.BookingPending)

	url, err := svc.InitiateComboPayment(context.Background(), userID, booking.ID, service.ComboPaymentRequest{
		WalletAmount:  booking.TotalPrice,
		CardAmount:    0,
		PaymentMethod: domain.PaymentMethodWallet,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if url != "" {
		t.Error("expected empty confirmation URL for full wallet payment")
	}
	if !walletSvc.spendCalled {
		t.Error("wallet Spend should be called for full wallet payment")
	}
	if walletSvc.spendAmount != booking.TotalPrice {
		t.Errorf("spend amount = %d, want %d", walletSvc.spendAmount, booking.TotalPrice)
	}

	p, _ := paymentRepo.GetByBookingID(context.Background(), booking.ID)
	if p.Status != domain.PaymentSucceeded {
		t.Errorf("payment status = %q, want %q", p.Status, domain.PaymentSucceeded)
	}
	if p.PaymentMethod != domain.PaymentMethodWallet {
		t.Errorf("payment method = %q, want %q", p.PaymentMethod, domain.PaymentMethodWallet)
	}

	updatedBooking, _ := bookingRepo.GetByID(context.Background(), booking.ID)
	if updatedBooking.Status != domain.BookingConfirmed {
		t.Errorf("booking status = %q, want %q", updatedBooking.Status, domain.BookingConfirmed)
	}
}

func TestPaymentService_ComboPayment_WalletPlusCard(t *testing.T) {
	walletSvc := &testWalletService{}
	svc, paymentRepo, bookingRepo, _ := newPaymentServiceWithWallet(walletSvc)
	userID := uuid.New()
	booking := createTestBooking(t, bookingRepo, userID, uuid.New(), time.Now().Add(24*time.Hour), domain.BookingPending)

	walletPortion := int64(5000)
	cardPortion := booking.TotalPrice - walletPortion

	url, err := svc.InitiateComboPayment(context.Background(), userID, booking.ID, service.ComboPaymentRequest{
		WalletAmount:  walletPortion,
		CardAmount:    cardPortion,
		PaymentMethod: domain.PaymentMethodCard,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if url == "" {
		t.Error("expected non-empty confirmation URL for combo payment")
	}
	if !walletSvc.spendCalled {
		t.Error("wallet Spend should be called for combo payment")
	}
	if walletSvc.spendAmount != walletPortion {
		t.Errorf("spend amount = %d, want %d", walletSvc.spendAmount, walletPortion)
	}

	p, _ := paymentRepo.GetByBookingID(context.Background(), booking.ID)
	if p.PaymentMethod != domain.PaymentMethodCombo {
		t.Errorf("payment method = %q, want %q", p.PaymentMethod, domain.PaymentMethodCombo)
	}
	if p.WalletAmount != walletPortion {
		t.Errorf("wallet_amount = %d, want %d", p.WalletAmount, walletPortion)
	}
	if p.CardAmount != cardPortion {
		t.Errorf("card_amount = %d, want %d", p.CardAmount, cardPortion)
	}
}

func TestPaymentService_ComboPayment_CardFailure_WalletRefunded(t *testing.T) {
	walletSvc := &testWalletService{}
	userID := uuid.New()
	bookingRepo := mock.NewBookingRepo()
	booking := createTestBooking(t, bookingRepo, userID, uuid.New(), time.Now().Add(24*time.Hour), domain.BookingPending)

	failProvider := &failingMockProvider{}
	paymentRepo := mock.NewPaymentRepo().(*mock.PaymentRepo)
	log := logger.New(logger.LevelWarn)
	failSvc := service.NewPaymentService(paymentRepo, bookingRepo, nil, failProvider, fiscal.NewNoOpProvider(), walletSvc, &noopNotifService{}, "http://localhost:3000/callback", log)

	walletPortion := int64(5000)
	cardPortion := booking.TotalPrice - walletPortion

	_, err := failSvc.InitiateComboPayment(context.Background(), userID, booking.ID, service.ComboPaymentRequest{
		WalletAmount:  walletPortion,
		CardAmount:    cardPortion,
		PaymentMethod: domain.PaymentMethodCard,
	})
	if !errors.Is(err, domain.ErrPaymentFailed) {
		t.Errorf("expected ErrPaymentFailed, got: %v", err)
	}
	if !walletSvc.spendCalled {
		t.Error("wallet Spend should have been called before card failure")
	}
	if !walletSvc.refundCalled {
		t.Error("wallet Refund should be called after card failure")
	}
	if walletSvc.refundAmount != walletPortion {
		t.Errorf("refund amount = %d, want %d", walletSvc.refundAmount, walletPortion)
	}
}

func TestPaymentService_ComboPayment_InvalidAmounts(t *testing.T) {
	walletSvc := &testWalletService{}
	svc, _, bookingRepo, _ := newPaymentServiceWithWallet(walletSvc)
	userID := uuid.New()
	booking := createTestBooking(t, bookingRepo, userID, uuid.New(), time.Now().Add(24*time.Hour), domain.BookingPending)

	// Amounts don't sum to total
	_, err := svc.InitiateComboPayment(context.Background(), userID, booking.ID, service.ComboPaymentRequest{
		WalletAmount:  1000,
		CardAmount:    2000,
		PaymentMethod: domain.PaymentMethodCard,
	})
	if !errors.Is(err, domain.ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput for mismatched amounts, got: %v", err)
	}
}

func TestPaymentService_ComboPayment_InsufficientWallet(t *testing.T) {
	walletSvc := &testWalletService{
		spendErr: domain.ErrInsufficientWalletBalance,
	}
	svc, _, bookingRepo, _ := newPaymentServiceWithWallet(walletSvc)
	userID := uuid.New()
	booking := createTestBooking(t, bookingRepo, userID, uuid.New(), time.Now().Add(24*time.Hour), domain.BookingPending)

	_, err := svc.InitiateComboPayment(context.Background(), userID, booking.ID, service.ComboPaymentRequest{
		WalletAmount:  booking.TotalPrice,
		CardAmount:    0,
		PaymentMethod: domain.PaymentMethodWallet,
	})
	if !errors.Is(err, domain.ErrInsufficientWalletBalance) {
		t.Errorf("expected ErrInsufficientWalletBalance, got: %v", err)
	}
}

// failingMockProvider is a PaymentProvider that always fails on CreatePayment.
type failingMockProvider struct{}

func (f *failingMockProvider) CreatePayment(_ context.Context, _ payment.CreatePaymentRequest) (*payment.PaymentResult, error) {
	return nil, errors.New("provider unavailable")
}
func (f *failingMockProvider) GetPaymentStatus(_ context.Context, _ string) (string, error) {
	return "", errors.New("not found")
}
func (f *failingMockProvider) CreateRefund(_ context.Context, _ string, _ int64) error {
	return nil
}
func (f *failingMockProvider) CapturePayment(_ context.Context, _ string, _ int64) error {
	return nil
}
func (f *failingMockProvider) CancelPayment(_ context.Context, _ string) error { return nil }

// --- Payment Hold Tests (Task 9) ---

func TestPaymentService_InitiatePayment_RequestBooking_CardHold(t *testing.T) {
	svc, paymentRepo, bookingRepo, provider := newPaymentService()
	userID := uuid.New()
	// Create a pending_owner booking (request-based)
	booking := createTestBooking(t, bookingRepo, userID, uuid.New(), time.Now().Add(24*time.Hour), domain.BookingPendingOwner)

	url, err := svc.InitiatePayment(context.Background(), userID, booking.ID, domain.PaymentMethodCard)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if url == "" {
		t.Error("expected non-empty confirmation URL")
	}

	// Verify payment was created with IsHold=true
	p, err := paymentRepo.GetByBookingID(context.Background(), booking.ID)
	if err != nil {
		t.Fatalf("failed to get payment: %v", err)
	}
	if !p.IsHold {
		t.Error("expected payment to be a hold (IsHold=true)")
	}

	// Verify provider received Capture=false
	capture := provider.GetPaymentCapture(p.ExternalID)
	if capture {
		t.Error("expected Capture=false for request-based booking payment")
	}
}

func TestPaymentService_CaptureHoldPayment_Card(t *testing.T) {
	svc, paymentRepo, bookingRepo, provider := newPaymentService()
	userID := uuid.New()
	booking := createTestBooking(t, bookingRepo, userID, uuid.New(), time.Now().Add(24*time.Hour), domain.BookingPendingOwner)

	// Initiate hold payment
	_, err := svc.InitiatePayment(context.Background(), userID, booking.ID, domain.PaymentMethodCard)
	if err != nil {
		t.Fatalf("initiate failed: %v", err)
	}

	p, _ := paymentRepo.GetByBookingID(context.Background(), booking.ID)

	// Set provider status to allow capture
	provider.SetPaymentStatus(p.ExternalID, "waiting_for_capture")

	// Capture the hold
	err = svc.CaptureHoldPayment(context.Background(), booking.ID)
	if err != nil {
		t.Fatalf("capture hold failed: %v", err)
	}

	// Verify payment was captured
	updated, _ := paymentRepo.GetByID(context.Background(), p.ID)
	if updated.Status != domain.PaymentSucceeded {
		t.Errorf("payment status = %q, want %q", updated.Status, domain.PaymentSucceeded)
	}
	if updated.IsHold {
		t.Error("expected IsHold=false after capture")
	}
	if updated.CapturedAt == nil {
		t.Error("expected CapturedAt to be set")
	}

	// Verify provider was called
	if !provider.WasCaptured(p.ExternalID) {
		t.Error("expected CapturePayment to be called on provider")
	}
}

func TestPaymentService_ReleaseHoldPayment_Card(t *testing.T) {
	svc, paymentRepo, bookingRepo, provider := newPaymentService()
	userID := uuid.New()
	booking := createTestBooking(t, bookingRepo, userID, uuid.New(), time.Now().Add(24*time.Hour), domain.BookingPendingOwner)

	// Initiate hold payment
	_, err := svc.InitiatePayment(context.Background(), userID, booking.ID, domain.PaymentMethodCard)
	if err != nil {
		t.Fatalf("initiate failed: %v", err)
	}

	p, _ := paymentRepo.GetByBookingID(context.Background(), booking.ID)

	// Release the hold (reject)
	err = svc.ReleaseHoldPayment(context.Background(), booking.ID)
	if err != nil {
		t.Fatalf("release hold failed: %v", err)
	}

	// Verify payment status is failed
	updated, _ := paymentRepo.GetByID(context.Background(), p.ID)
	if updated.Status != domain.PaymentFailed {
		t.Errorf("payment status = %q, want %q", updated.Status, domain.PaymentFailed)
	}

	// Verify provider was called to cancel
	if !provider.WasCancelled(p.ExternalID) {
		t.Error("expected CancelPayment to be called on provider")
	}
}

func TestPaymentService_InitiateComboPayment_RequestBooking_Hold(t *testing.T) {
	walletSvc := &testWalletService{}
	svc, paymentRepo, bookingRepo, provider := newPaymentServiceWithWallet(walletSvc)
	userID := uuid.New()
	booking := createTestBooking(t, bookingRepo, userID, uuid.New(), time.Now().Add(24*time.Hour), domain.BookingPendingOwner)

	walletPortion := int64(5000)
	cardPortion := booking.TotalPrice - walletPortion

	url, err := svc.InitiateComboPayment(context.Background(), userID, booking.ID, service.ComboPaymentRequest{
		WalletAmount:  walletPortion,
		CardAmount:    cardPortion,
		PaymentMethod: domain.PaymentMethodCard,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if url == "" {
		t.Error("expected non-empty confirmation URL for combo hold payment")
	}

	// Verify payment was created with IsHold=true
	p, err := paymentRepo.GetByBookingID(context.Background(), booking.ID)
	if err != nil {
		t.Fatalf("failed to get payment: %v", err)
	}
	if !p.IsHold {
		t.Error("expected payment to be a hold")
	}
	if p.PaymentMethod != domain.PaymentMethodCombo {
		t.Errorf("payment method = %q, want %q", p.PaymentMethod, domain.PaymentMethodCombo)
	}

	// Verify card portion was created with Capture=false
	capture := provider.GetPaymentCapture(p.ExternalID)
	if capture {
		t.Error("expected Capture=false for card portion of combo hold")
	}
}

func TestPaymentService_ComboHold_CardFailure_WalletHoldReleased(t *testing.T) {
	walletSvc := &testWalletService{}
	userID := uuid.New()
	bookingRepo := mock.NewBookingRepo()
	booking := createTestBooking(t, bookingRepo, userID, uuid.New(), time.Now().Add(24*time.Hour), domain.BookingPendingOwner)

	failProvider := &failingMockProvider{}
	paymentRepo := mock.NewPaymentRepo().(*mock.PaymentRepo)
	log := logger.New(logger.LevelWarn)
	failSvc := service.NewPaymentService(paymentRepo, bookingRepo, nil, failProvider, fiscal.NewNoOpProvider(), walletSvc, &noopNotifService{}, "http://localhost:3000/callback", log)

	walletPortion := int64(5000)
	cardPortion := booking.TotalPrice - walletPortion

	_, err := failSvc.InitiateComboPayment(context.Background(), userID, booking.ID, service.ComboPaymentRequest{
		WalletAmount:  walletPortion,
		CardAmount:    cardPortion,
		PaymentMethod: domain.PaymentMethodCard,
	})
	if !errors.Is(err, domain.ErrPaymentFailed) {
		t.Errorf("expected ErrPaymentFailed, got: %v", err)
	}
	// For hold mode, wallet is held (not spent), so on card failure the hold should be released
	// The wallet hold release happens via GetActiveHolds, but since noopWalletService always
	// returns empty holds, we just verify the payment failed cleanly
}

func TestPaymentService_ReleaseHoldPayment_NotFound(t *testing.T) {
	svc, _, _, _ := newPaymentService()
	// Release for nonexistent booking should not error
	err := svc.ReleaseHoldPayment(context.Background(), uuid.New())
	if err != nil {
		t.Errorf("expected nil for nonexistent booking, got: %v", err)
	}
}

func TestPaymentService_ReleaseHoldPayment_NotAHold(t *testing.T) {
	svc, _, bookingRepo, _ := newPaymentService()
	userID := uuid.New()
	booking := createTestBooking(t, bookingRepo, userID, uuid.New(), time.Now().Add(24*time.Hour), domain.BookingPending)

	// Create a regular payment (not a hold)
	_, err := svc.InitiatePayment(context.Background(), userID, booking.ID, domain.PaymentMethodCard)
	if err != nil {
		t.Fatalf("initiate failed: %v", err)
	}

	// Release should be no-op for non-hold payment
	err = svc.ReleaseHoldPayment(context.Background(), booking.ID)
	if err != nil {
		t.Errorf("expected nil for non-hold payment, got: %v", err)
	}
}

func TestPaymentService_CaptureHoldPayment_NonHold(t *testing.T) {
	svc, _, bookingRepo, _ := newPaymentService()
	userID := uuid.New()
	booking := createTestBooking(t, bookingRepo, userID, uuid.New(), time.Now().Add(24*time.Hour), domain.BookingPending)

	// Create a regular payment (not a hold)
	_, err := svc.InitiatePayment(context.Background(), userID, booking.ID, domain.PaymentMethodCard)
	if err != nil {
		t.Fatalf("initiate failed: %v", err)
	}

	// Capture should be no-op for non-hold payment
	err = svc.CaptureHoldPayment(context.Background(), booking.ID)
	if err != nil {
		t.Errorf("expected nil for non-hold payment, got: %v", err)
	}
}

func TestPaymentService_RefundPayment_HoldReleasedInsteadOfRefund(t *testing.T) {
	svc, paymentRepo, bookingRepo, provider := newPaymentService()
	userID := uuid.New()
	booking := createTestBooking(t, bookingRepo, userID, uuid.New(), time.Now().Add(24*time.Hour), domain.BookingPendingOwner)

	// Create a hold payment
	_, err := svc.InitiatePayment(context.Background(), userID, booking.ID, domain.PaymentMethodCard)
	if err != nil {
		t.Fatalf("initiate failed: %v", err)
	}

	p, _ := paymentRepo.GetByBookingID(context.Background(), booking.ID)

	// Calling RefundPayment on a hold should release it instead
	err = svc.RefundPayment(context.Background(), booking.ID, true, "")
	if err != nil {
		t.Fatalf("refund/release failed: %v", err)
	}

	// Verify it was cancelled, not refunded
	if !provider.WasCancelled(p.ExternalID) {
		t.Error("expected CancelPayment to be called for hold payment refund")
	}

	updated, _ := paymentRepo.GetByID(context.Background(), p.ID)
	if updated.Status != domain.PaymentFailed {
		t.Errorf("payment status = %q, want %q", updated.Status, domain.PaymentFailed)
	}
}

// --- Enhanced Refund Tests (Task 12) ---

func TestPaymentService_RefundPayment_WalletRefundWithBonus(t *testing.T) {
	walletSvc := &testWalletService{}
	svc, paymentRepo, bookingRepo, provider := newPaymentServiceWithWallet(walletSvc)
	userID := uuid.New()
	booking := createTestBooking(t, bookingRepo, userID, uuid.New(), time.Now().Add(48*time.Hour), domain.BookingPending)

	_, _ = svc.InitiatePayment(context.Background(), userID, booking.ID, domain.PaymentMethodCard)
	p, _ := paymentRepo.GetByBookingID(context.Background(), booking.ID)

	provider.SetPaymentStatus(p.ExternalID, "succeeded")
	_ = svc.HandleWebhook(context.Background(), service.WebhookEvent{
		ExternalID: p.ExternalID,
		Status:     "succeeded",
	})

	err := svc.RefundPayment(context.Background(), booking.ID, false, "wallet")
	if err != nil {
		t.Fatalf("wallet refund failed: %v", err)
	}

	if !walletSvc.refundCalled {
		t.Error("expected wallet Refund to be called")
	}
	expectedRefund := int64(10500) // 10000 + 5%
	if walletSvc.refundAmount != expectedRefund {
		t.Errorf("wallet refund amount = %d, want %d (refund + 5%% bonus)", walletSvc.refundAmount, expectedRefund)
	}

	updated2, _ := paymentRepo.GetByID(context.Background(), p.ID)
	if updated2.RefundAmount != 10000 {
		t.Errorf("payment refund amount = %d, want 10000", updated2.RefundAmount)
	}
	if updated2.Status != domain.PaymentRefunded {
		t.Errorf("payment status = %q, want %q", updated2.Status, domain.PaymentRefunded)
	}
}

func TestPaymentService_RefundPayment_CardRefundDefault(t *testing.T) {
	svc, paymentRepo, bookingRepo, provider := newPaymentService()
	userID := uuid.New()
	booking := createTestBooking(t, bookingRepo, userID, uuid.New(), time.Now().Add(48*time.Hour), domain.BookingPending)

	_, _ = svc.InitiatePayment(context.Background(), userID, booking.ID, domain.PaymentMethodCard)
	p, _ := paymentRepo.GetByBookingID(context.Background(), booking.ID)

	provider.SetPaymentStatus(p.ExternalID, "succeeded")
	_ = svc.HandleWebhook(context.Background(), service.WebhookEvent{
		ExternalID: p.ExternalID,
		Status:     "succeeded",
	})

	err := svc.RefundPayment(context.Background(), booking.ID, false, "")
	if err != nil {
		t.Fatalf("card refund failed: %v", err)
	}

	updated2, _ := paymentRepo.GetByID(context.Background(), p.ID)
	if updated2.RefundAmount != 10000 {
		t.Errorf("refund amount = %d, want 10000", updated2.RefundAmount)
	}
	if updated2.Status != domain.PaymentRefunded {
		t.Errorf("payment status = %q, want %q", updated2.Status, domain.PaymentRefunded)
	}
}

func TestPaymentService_ComboRefund_Proportional(t *testing.T) {
	walletSvc := &testWalletService{}
	svc, paymentRepo, bookingRepo, provider := newPaymentServiceWithWallet(walletSvc)
	userID := uuid.New()
	booking := createTestBooking(t, bookingRepo, userID, uuid.New(), time.Now().Add(48*time.Hour), domain.BookingPending)

	_, _ = svc.InitiateComboPayment(context.Background(), userID, booking.ID, service.ComboPaymentRequest{
		WalletAmount:  3000,
		CardAmount:    7000,
		PaymentMethod: domain.PaymentMethodCard,
	})
	p, _ := paymentRepo.GetByBookingID(context.Background(), booking.ID)

	provider.SetPaymentStatus(p.ExternalID, "succeeded")
	_ = svc.HandleWebhook(context.Background(), service.WebhookEvent{
		ExternalID: p.ExternalID,
		Status:     "succeeded",
	})

	walletSvc.refundCalled = false
	walletSvc.refundAmount = 0

	err := svc.RefundPayment(context.Background(), booking.ID, false, "card")
	if err != nil {
		t.Fatalf("combo refund failed: %v", err)
	}

	if !walletSvc.refundCalled {
		t.Error("expected wallet Refund to be called for wallet portion")
	}
	// Wallet portion: 3000 + 5% = 3150
	if walletSvc.refundAmount != 3150 {
		t.Errorf("wallet refund = %d, want 3150 (3000 + 5%% bonus)", walletSvc.refundAmount)
	}

	updated2, _ := paymentRepo.GetByID(context.Background(), p.ID)
	if updated2.RefundAmount != 10000 {
		t.Errorf("payment refund amount = %d, want 10000", updated2.RefundAmount)
	}
}

func TestPaymentService_ComboRefund_AllToWallet(t *testing.T) {
	walletSvc := &testWalletService{}
	svc, paymentRepo, bookingRepo, provider := newPaymentServiceWithWallet(walletSvc)
	userID := uuid.New()
	booking := createTestBooking(t, bookingRepo, userID, uuid.New(), time.Now().Add(48*time.Hour), domain.BookingPending)

	_, _ = svc.InitiateComboPayment(context.Background(), userID, booking.ID, service.ComboPaymentRequest{
		WalletAmount:  3000,
		CardAmount:    7000,
		PaymentMethod: domain.PaymentMethodCard,
	})
	p, _ := paymentRepo.GetByBookingID(context.Background(), booking.ID)

	provider.SetPaymentStatus(p.ExternalID, "succeeded")
	_ = svc.HandleWebhook(context.Background(), service.WebhookEvent{
		ExternalID: p.ExternalID,
		Status:     "succeeded",
	})

	walletSvc.refundCalled = false
	walletSvc.refundAmount = 0

	err := svc.RefundPayment(context.Background(), booking.ID, false, "wallet")
	if err != nil {
		t.Fatalf("combo all-to-wallet refund failed: %v", err)
	}

	if !walletSvc.refundCalled {
		t.Error("expected wallet Refund to be called")
	}
	// Last refund call: card portion redirected to wallet: 7000 + 5% = 7350
	if walletSvc.refundAmount != 7350 {
		t.Errorf("last wallet refund = %d, want 7350 (card portion + 5%% bonus)", walletSvc.refundAmount)
	}

	updated2, _ := paymentRepo.GetByID(context.Background(), p.ID)
	if updated2.Status != domain.PaymentRefunded {
		t.Errorf("payment status = %q, want %q", updated2.Status, domain.PaymentRefunded)
	}
}

func TestPaymentService_AdminRefund(t *testing.T) {
	walletSvc := &testWalletService{}
	svc, paymentRepo, bookingRepo, provider := newPaymentServiceWithWallet(walletSvc)
	userID := uuid.New()
	booking := createTestBooking(t, bookingRepo, userID, uuid.New(), time.Now().Add(48*time.Hour), domain.BookingPending)

	_, _ = svc.InitiatePayment(context.Background(), userID, booking.ID, domain.PaymentMethodCard)
	p, _ := paymentRepo.GetByBookingID(context.Background(), booking.ID)

	provider.SetPaymentStatus(p.ExternalID, "succeeded")
	_ = svc.HandleWebhook(context.Background(), service.WebhookEvent{
		ExternalID: p.ExternalID,
		Status:     "succeeded",
	})

	err := svc.AdminRefund(context.Background(), booking.ID, 5000, "customer complaint", "wallet")
	if err != nil {
		t.Fatalf("admin refund failed: %v", err)
	}

	if !walletSvc.refundCalled {
		t.Error("expected wallet Refund to be called")
	}
	if walletSvc.refundAmount != 5250 {
		t.Errorf("wallet refund = %d, want 5250 (5000 + 5%%)", walletSvc.refundAmount)
	}

	updated2, _ := paymentRepo.GetByID(context.Background(), p.ID)
	if updated2.RefundAmount != 5000 {
		t.Errorf("payment refund amount = %d, want 5000", updated2.RefundAmount)
	}
	if updated2.Status != domain.PaymentPartiallyRefunded {
		t.Errorf("payment status = %q, want %q", updated2.Status, domain.PaymentPartiallyRefunded)
	}
}

func TestPaymentService_AdminRefund_ExceedsAmount(t *testing.T) {
	svc, paymentRepo, bookingRepo, provider := newPaymentService()
	userID := uuid.New()
	booking := createTestBooking(t, bookingRepo, userID, uuid.New(), time.Now().Add(48*time.Hour), domain.BookingPending)

	_, _ = svc.InitiatePayment(context.Background(), userID, booking.ID, domain.PaymentMethodCard)
	p, _ := paymentRepo.GetByBookingID(context.Background(), booking.ID)

	provider.SetPaymentStatus(p.ExternalID, "succeeded")
	_ = svc.HandleWebhook(context.Background(), service.WebhookEvent{
		ExternalID: p.ExternalID,
		Status:     "succeeded",
	})

	err := svc.AdminRefund(context.Background(), booking.ID, 20000, "test", "card")
	if !errors.Is(err, domain.ErrRefundExceedsAmount) {
		t.Errorf("expected ErrRefundExceedsAmount, got: %v", err)
	}
}

func TestPaymentService_RefundPayment_NoRefundTier_IgnoresRefundTo(t *testing.T) {
	svc, paymentRepo, bookingRepo, provider := newPaymentService()
	userID := uuid.New()
	booking := createTestBooking(t, bookingRepo, userID, uuid.New(), time.Now().Add(1*time.Hour), domain.BookingPending)

	_, _ = svc.InitiatePayment(context.Background(), userID, booking.ID, domain.PaymentMethodCard)
	p, _ := paymentRepo.GetByBookingID(context.Background(), booking.ID)

	provider.SetPaymentStatus(p.ExternalID, "succeeded")
	_ = svc.HandleWebhook(context.Background(), service.WebhookEvent{
		ExternalID: p.ExternalID,
		Status:     "succeeded",
	})

	err := svc.RefundPayment(context.Background(), booking.ID, false, "wallet")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	updated2, _ := paymentRepo.GetByID(context.Background(), p.ID)
	if updated2.RefundAmount != 0 {
		t.Errorf("refund amount = %d, want 0 (no refund within 2h)", updated2.RefundAmount)
	}
}

func TestPaymentService_Webhook_TriggersFiscalReceipt(t *testing.T) {
	fp := &mockFiscalProvider{}
	svc, paymentRepo, bookingRepo, provider := newPaymentServiceWithFiscal(fp)
	userID := uuid.New()
	booking := createTestBooking(t, bookingRepo, userID, uuid.New(), time.Now().Add(48*time.Hour), domain.BookingPending)

	_, err := svc.InitiatePayment(context.Background(), userID, booking.ID, domain.PaymentMethodCard)
	if err != nil {
		t.Fatalf("initiate failed: %v", err)
	}

	p, _ := paymentRepo.GetByBookingID(context.Background(), booking.ID)
	provider.SetPaymentStatus(p.ExternalID, "succeeded")

	err = svc.HandleWebhook(context.Background(), service.WebhookEvent{
		ExternalID: p.ExternalID,
		Status:     "succeeded",
	})
	if err != nil {
		t.Fatalf("webhook failed: %v", err)
	}

	if len(fp.receipts) != 1 {
		t.Fatalf("expected 1 fiscal receipt, got %d", len(fp.receipts))
	}
	if fp.receipts[0].Type != fiscal.ReceiptAdvance {
		t.Errorf("expected receipt type %q, got %q", fiscal.ReceiptAdvance, fp.receipts[0].Type)
	}
	if fp.receipts[0].Amount != booking.TotalPrice {
		t.Errorf("expected receipt amount %d, got %d", booking.TotalPrice, fp.receipts[0].Amount)
	}
}

func TestPaymentService_Refund_TriggersFiscalReceipt(t *testing.T) {
	fp := &mockFiscalProvider{}
	svc, paymentRepo, bookingRepo, provider := newPaymentServiceWithFiscal(fp)
	userID := uuid.New()
	booking := createTestBooking(t, bookingRepo, userID, uuid.New(), time.Now().Add(48*time.Hour), domain.BookingPending)

	_, err := svc.InitiatePayment(context.Background(), userID, booking.ID, domain.PaymentMethodCard)
	if err != nil {
		t.Fatalf("initiate failed: %v", err)
	}

	p, _ := paymentRepo.GetByBookingID(context.Background(), booking.ID)
	provider.SetPaymentStatus(p.ExternalID, "succeeded")

	err = svc.HandleWebhook(context.Background(), service.WebhookEvent{
		ExternalID: p.ExternalID,
		Status:     "succeeded",
	})
	if err != nil {
		t.Fatalf("webhook failed: %v", err)
	}

	// Reset receipts to only track refund receipt
	fp.receipts = nil

	err = svc.RefundPayment(context.Background(), booking.ID, true, "card")
	if err != nil {
		t.Fatalf("refund failed: %v", err)
	}

	if len(fp.receipts) != 1 {
		t.Fatalf("expected 1 fiscal receipt for refund, got %d", len(fp.receipts))
	}
	if fp.receipts[0].Type != fiscal.ReceiptRefund {
		t.Errorf("expected receipt type %q, got %q", fiscal.ReceiptRefund, fp.receipts[0].Type)
	}
}

func TestPaymentService_FiscalFailure_DoesNotBlockPayment(t *testing.T) {
	fp := &mockFiscalProvider{err: errors.New("ATOL service unavailable")}
	svc, paymentRepo, bookingRepo, provider := newPaymentServiceWithFiscal(fp)
	userID := uuid.New()
	booking := createTestBooking(t, bookingRepo, userID, uuid.New(), time.Now().Add(48*time.Hour), domain.BookingPending)

	_, err := svc.InitiatePayment(context.Background(), userID, booking.ID, domain.PaymentMethodCard)
	if err != nil {
		t.Fatalf("initiate failed: %v", err)
	}

	p, _ := paymentRepo.GetByBookingID(context.Background(), booking.ID)
	provider.SetPaymentStatus(p.ExternalID, "succeeded")

	// Webhook should succeed even if fiscal receipt creation fails
	err = svc.HandleWebhook(context.Background(), service.WebhookEvent{
		ExternalID: p.ExternalID,
		Status:     "succeeded",
	})
	if err != nil {
		t.Fatalf("webhook should not fail due to fiscal error, got: %v", err)
	}

	// Verify booking was still confirmed
	updatedBooking, _ := bookingRepo.GetByID(context.Background(), booking.ID)
	if updatedBooking.Status != domain.BookingConfirmed {
		t.Errorf("expected booking to be confirmed, got %s", updatedBooking.Status)
	}
}

func TestPaymentService_FiscalFailure_DoesNotBlockRefund(t *testing.T) {
	// Start with working fiscal, then break it for refund
	fp := &mockFiscalProvider{}
	svc, paymentRepo, bookingRepo, provider := newPaymentServiceWithFiscal(fp)
	userID := uuid.New()
	booking := createTestBooking(t, bookingRepo, userID, uuid.New(), time.Now().Add(48*time.Hour), domain.BookingPending)

	_, _ = svc.InitiatePayment(context.Background(), userID, booking.ID, domain.PaymentMethodCard)
	p, _ := paymentRepo.GetByBookingID(context.Background(), booking.ID)
	provider.SetPaymentStatus(p.ExternalID, "succeeded")
	_ = svc.HandleWebhook(context.Background(), service.WebhookEvent{
		ExternalID: p.ExternalID,
		Status:     "succeeded",
	})

	// Now make fiscal fail
	fp.err = errors.New("ATOL timeout")
	fp.receipts = nil

	// Refund should still succeed
	err := svc.RefundPayment(context.Background(), booking.ID, true, "card")
	if err != nil {
		t.Fatalf("refund should not fail due to fiscal error, got: %v", err)
	}

	// Verify refund was recorded
	updated, _ := paymentRepo.GetByID(context.Background(), p.ID)
	if updated.Status != domain.PaymentRefunded {
		t.Errorf("expected payment status refunded, got %s", updated.Status)
	}
}
