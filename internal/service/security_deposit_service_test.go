package service

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/logger"
	"github.com/nikitaaldaev/bani/internal/payment"
	"github.com/nikitaaldaev/bani/internal/repository/mock"
)

func newDepositService() (SecurityDepositService, *mock.BookingRepo, *payment.MockProvider) {
	bookingRepo := mock.NewBookingRepo()
	provider := payment.NewMockProvider()
	log := logger.New(logger.LevelWarn)
	svc := NewSecurityDepositService(bookingRepo, provider, log, 48, "https://example.com/return")
	return svc, bookingRepo, provider
}

func TestCalculateDepositAmount(t *testing.T) {
	tests := []struct {
		name           string
		basePrice      int64
		depositPercent int
		want           int64
	}{
		{"zero percent", 10000, 0, 0},
		{"negative percent", 10000, -5, 0},
		{"over 50 percent", 10000, 51, 0},
		{"10 percent", 10000, 10, 1000},
		{"50 percent", 10000, 50, 5000},
		{"25 percent of 20000", 20000, 25, 5000},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalculateDepositAmount(tt.basePrice, tt.depositPercent)
			if got != tt.want {
				t.Errorf("CalculateDepositAmount(%d, %d) = %d, want %d", tt.basePrice, tt.depositPercent, got, tt.want)
			}
		})
	}
}

func TestHoldDeposit_Success(t *testing.T) {
	svc, bookingRepo, provider := newDepositService()
	bookingID := uuid.New()
	booking := &domain.Booking{
		ID: bookingID, UserID: uuid.New(), BathhouseID: uuid.New(),
		StartTime: time.Now().Add(24 * time.Hour), EndTime: time.Now().Add(26 * time.Hour),
		GuestCount: 2, TotalPrice: 10000, Status: domain.BookingConfirmed,
	}
	_ = bookingRepo.Create(context.Background(), booking)

	err := svc.HoldDeposit(context.Background(), bookingID, 2000, "card")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	updated, _ := bookingRepo.GetByID(context.Background(), bookingID)
	if updated.DepositAmount != 2000 {
		t.Errorf("DepositAmount = %d, want 2000", updated.DepositAmount)
	}
	if updated.DepositStatus != domain.DepositHeld {
		t.Errorf("DepositStatus = %s, want %s", updated.DepositStatus, domain.DepositHeld)
	}
	if updated.DepositExternalID == "" {
		t.Error("DepositExternalID should be set")
	}

	// Verify payment was created with Capture=false (authorization hold)
	if provider.GetPaymentCapture(updated.DepositExternalID) {
		t.Error("expected Capture=false for deposit hold")
	}
}

func TestHoldDeposit_ZeroAmount(t *testing.T) {
	svc, _, _ := newDepositService()
	err := svc.HoldDeposit(context.Background(), uuid.New(), 0, "card")
	if err != nil {
		t.Fatalf("expected no error for zero amount, got: %v", err)
	}
}

func TestReleaseDeposit_Success(t *testing.T) {
	svc, bookingRepo, provider := newDepositService()
	bookingID := uuid.New()
	booking := &domain.Booking{
		ID: bookingID, UserID: uuid.New(), BathhouseID: uuid.New(),
		StartTime: time.Now().Add(-2 * time.Hour), EndTime: time.Now().Add(-1 * time.Hour),
		GuestCount: 2, TotalPrice: 10000, Status: domain.BookingCompleted,
	}
	_ = bookingRepo.Create(context.Background(), booking)

	// Hold the deposit first
	err := svc.HoldDeposit(context.Background(), bookingID, 2000, "card")
	if err != nil {
		t.Fatalf("hold failed: %v", err)
	}

	held, _ := bookingRepo.GetByID(context.Background(), bookingID)
	extID := held.DepositExternalID

	// Now release
	err = svc.ReleaseDeposit(context.Background(), bookingID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	updated, _ := bookingRepo.GetByID(context.Background(), bookingID)
	if updated.DepositStatus != domain.DepositReleased {
		t.Errorf("DepositStatus = %s, want %s", updated.DepositStatus, domain.DepositReleased)
	}
	if updated.DepositReleasedAt == nil {
		t.Error("DepositReleasedAt should be set")
	}
	if !provider.WasCancelled(extID) {
		t.Error("expected CancelPayment to be called on release")
	}
}

func TestReleaseDeposit_AlreadyReleased(t *testing.T) {
	svc, bookingRepo, _ := newDepositService()
	bookingID := uuid.New()
	booking := &domain.Booking{
		ID: bookingID, UserID: uuid.New(), BathhouseID: uuid.New(),
		StartTime: time.Now(), EndTime: time.Now().Add(time.Hour),
		GuestCount: 1, TotalPrice: 5000, Status: domain.BookingCompleted,
	}
	_ = bookingRepo.Create(context.Background(), booking)
	_ = svc.HoldDeposit(context.Background(), bookingID, 1000, "card")
	_ = svc.ReleaseDeposit(context.Background(), bookingID)

	err := svc.ReleaseDeposit(context.Background(), bookingID)
	if err != domain.ErrDepositAlreadyReleased {
		t.Errorf("expected ErrDepositAlreadyReleased, got: %v", err)
	}
}

func TestClaimDeposit_Success(t *testing.T) {
	svc, bookingRepo, provider := newDepositService()
	bookingID := uuid.New()
	booking := &domain.Booking{
		ID: bookingID, UserID: uuid.New(), BathhouseID: uuid.New(),
		StartTime: time.Now(), EndTime: time.Now().Add(time.Hour),
		GuestCount: 1, TotalPrice: 5000, Status: domain.BookingCompleted,
	}
	_ = bookingRepo.Create(context.Background(), booking)
	_ = svc.HoldDeposit(context.Background(), bookingID, 1000, "card")

	held, _ := bookingRepo.GetByID(context.Background(), bookingID)
	extID := held.DepositExternalID

	err := svc.ClaimDeposit(context.Background(), bookingID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	updated, _ := bookingRepo.GetByID(context.Background(), bookingID)
	if updated.DepositStatus != domain.DepositClaimed {
		t.Errorf("DepositStatus = %s, want %s", updated.DepositStatus, domain.DepositClaimed)
	}
	if !provider.WasCaptured(extID) {
		t.Error("expected CapturePayment to be called on claim")
	}
}

func TestClaimDeposit_AlreadyClaimed(t *testing.T) {
	svc, bookingRepo, _ := newDepositService()
	bookingID := uuid.New()
	booking := &domain.Booking{
		ID: bookingID, UserID: uuid.New(), BathhouseID: uuid.New(),
		StartTime: time.Now(), EndTime: time.Now().Add(time.Hour),
		GuestCount: 1, TotalPrice: 5000, Status: domain.BookingCompleted,
	}
	_ = bookingRepo.Create(context.Background(), booking)
	_ = svc.HoldDeposit(context.Background(), bookingID, 1000, "card")
	_ = svc.ClaimDeposit(context.Background(), bookingID)

	err := svc.ClaimDeposit(context.Background(), bookingID)
	if err != domain.ErrDepositAlreadyClaimed {
		t.Errorf("expected ErrDepositAlreadyClaimed, got: %v", err)
	}
}

func TestFreezeDeposit_Success(t *testing.T) {
	svc, bookingRepo, _ := newDepositService()
	bookingID := uuid.New()
	booking := &domain.Booking{
		ID: bookingID, UserID: uuid.New(), BathhouseID: uuid.New(),
		StartTime: time.Now(), EndTime: time.Now().Add(time.Hour),
		GuestCount: 1, TotalPrice: 5000, Status: domain.BookingCompleted,
	}
	_ = bookingRepo.Create(context.Background(), booking)
	_ = svc.HoldDeposit(context.Background(), bookingID, 1000, "card")

	err := svc.FreezeDeposit(context.Background(), bookingID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	updated, _ := bookingRepo.GetByID(context.Background(), bookingID)
	if updated.DepositStatus != domain.DepositDisputed {
		t.Errorf("DepositStatus = %s, want %s", updated.DepositStatus, domain.DepositDisputed)
	}
}

func TestFreezeDeposit_NotHeld_Noop(t *testing.T) {
	svc, bookingRepo, _ := newDepositService()
	bookingID := uuid.New()
	booking := &domain.Booking{
		ID: bookingID, UserID: uuid.New(), BathhouseID: uuid.New(),
		StartTime: time.Now(), EndTime: time.Now().Add(time.Hour),
		GuestCount: 1, TotalPrice: 5000, Status: domain.BookingCompleted,
		DepositStatus: domain.DepositNone,
	}
	_ = bookingRepo.Create(context.Background(), booking)

	err := svc.FreezeDeposit(context.Background(), bookingID)
	if err != nil {
		t.Fatalf("expected no error for non-held deposit, got: %v", err)
	}
}

func TestReleaseDeposit_AfterFreeze(t *testing.T) {
	svc, bookingRepo, _ := newDepositService()
	bookingID := uuid.New()
	booking := &domain.Booking{
		ID: bookingID, UserID: uuid.New(), BathhouseID: uuid.New(),
		StartTime: time.Now(), EndTime: time.Now().Add(time.Hour),
		GuestCount: 1, TotalPrice: 5000, Status: domain.BookingCompleted,
	}
	_ = bookingRepo.Create(context.Background(), booking)
	_ = svc.HoldDeposit(context.Background(), bookingID, 1000, "card")
	_ = svc.FreezeDeposit(context.Background(), bookingID)

	// Release should still work on disputed deposits (admin resolution)
	err := svc.ReleaseDeposit(context.Background(), bookingID)
	if err != nil {
		t.Fatalf("expected release to work on disputed deposit, got: %v", err)
	}

	updated, _ := bookingRepo.GetByID(context.Background(), bookingID)
	if updated.DepositStatus != domain.DepositReleased {
		t.Errorf("DepositStatus = %s, want %s", updated.DepositStatus, domain.DepositReleased)
	}
}

func TestProcessMaturedDeposits(t *testing.T) {
	svc, bookingRepo, _ := newDepositService()

	// Create two bookings with held deposits, checked out long ago
	checkedOut := time.Now().Add(-72 * time.Hour)
	for i := 0; i < 2; i++ {
		bookingID := uuid.New()
		booking := &domain.Booking{
			ID: bookingID, UserID: uuid.New(), BathhouseID: uuid.New(),
			StartTime: checkedOut.Add(-2 * time.Hour), EndTime: checkedOut.Add(-time.Hour),
			GuestCount: 1, TotalPrice: 5000, Status: domain.BookingCompleted,
			CheckedOutAt: &checkedOut,
		}
		_ = bookingRepo.Create(context.Background(), booking)
		_ = svc.HoldDeposit(context.Background(), bookingID, 1000, "card")
	}

	released, err := svc.ProcessMaturedDeposits(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if released != 2 {
		t.Errorf("released = %d, want 2", released)
	}
}

func TestNewSecurityDepositService_ClaimHoursDefault(t *testing.T) {
	bookingRepo := mock.NewBookingRepo()
	provider := payment.NewMockProvider()
	log := logger.New(logger.LevelWarn)

	// Out of range claimHours should default to 48
	svc := NewSecurityDepositService(bookingRepo, provider, log, 10, "https://example.com")
	s := svc.(*securityDepositService)
	if s.claimHours != 48 {
		t.Errorf("claimHours = %d, want 48 (default)", s.claimHours)
	}

	svc2 := NewSecurityDepositService(bookingRepo, provider, log, 200, "https://example.com")
	s2 := svc2.(*securityDepositService)
	if s2.claimHours != 48 {
		t.Errorf("claimHours = %d, want 48 (default)", s2.claimHours)
	}

	// Valid value should be kept
	svc3 := NewSecurityDepositService(bookingRepo, provider, log, 72, "https://example.com")
	s3 := svc3.(*securityDepositService)
	if s3.claimHours != 72 {
		t.Errorf("claimHours = %d, want 72", s3.claimHours)
	}
}
