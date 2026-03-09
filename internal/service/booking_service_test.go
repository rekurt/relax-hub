package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/logger"
	"github.com/nikitaaldaev/bani/internal/repository/mock"
	"github.com/nikitaaldaev/bani/internal/service"
)

func newBookingService() (service.BookingService, *mock.BathhouseRepo, *mock.BookingRepo, *mock.RepresentativeRepo, service.PricingService, *mock.PricingRuleRepo, service.LoyaltyService, *mock.LoyaltyRepo) {
	bhRepo := mock.NewBathhouseRepo()
	bookingRepo := mock.NewBookingRepo()
	repRepo := mock.NewRepresentativeRepo()
	pricingRepo := mock.NewPricingRuleRepo()
	loyaltyRepo := mock.NewLoyaltyRepo()
	access := service.NewAccessChecker(repRepo, bhRepo)
	log := logger.New(logger.LevelWarn)
	pricingSvc := service.NewPricingService(pricingRepo, bhRepo, access, log)
	loyaltySvc := service.NewLoyaltyService(loyaltyRepo, log)
	svc := service.NewBookingService(bookingRepo, bhRepo, pricingSvc, loyaltySvc, &noopReferralService{}, &noopPromoService{}, &noopPaymentService{}, access, &noopNotifService{}, log)
	return svc, bhRepo, bookingRepo, repRepo, pricingSvc, pricingRepo, loyaltySvc, loyaltyRepo
}

func TestBookingService_Create_Success(t *testing.T) {
	svc, bhRepo, _, _, _, _, _, _ := newBookingService()
	ownerID := uuid.New()
	clientID := uuid.New()
	bh := createBathhouse(t, bhRepo, ownerID)

	// Set start time to tomorrow at a specific hour (e.g., 10:00) to avoid crossing midnight
	now := time.Now()
	start := time.Date(now.Year(), now.Month(), now.Day()+1, 10, 0, 0, 0, now.Location())
	end := start.Add(2 * time.Hour)

	result, err := svc.Create(context.Background(), clientID, service.CreateBookingInput{
		BathhouseID: bh.ID,
		StartTime:   start,
		EndTime:     end,
		GuestCount:  5,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Booking.Status != domain.BookingPending {
		t.Errorf("status = %q, want %q", result.Booking.Status, domain.BookingPending)
	}
	if result.Booking.TotalPrice != bh.PricePerHour*2 {
		t.Errorf("totalPrice = %d, want %d", result.Booking.TotalPrice, bh.PricePerHour*2)
	}
}

func TestBookingService_Create_InactiveBathhouse(t *testing.T) {
	svc, bhRepo, _, _, _, _, _, _ := newBookingService()
	ownerID := uuid.New()

	bh := &domain.Bathhouse{
		ID: uuid.New(), OwnerID: ownerID, Name: "Pending Bath",
		Address: "123 St", CityID: 1, PricePerHour: 5000,
		MinDuration: 1, MaxGuests: 10, Status: domain.BathhouseStatusPending,
	}
	_ = bhRepo.Create(context.Background(), bh)

	start := time.Now().Add(24 * time.Hour)
	_, err := svc.Create(context.Background(), uuid.New(), service.CreateBookingInput{
		BathhouseID: bh.ID,
		StartTime:   start,
		EndTime:     start.Add(2 * time.Hour),
		GuestCount:  5,
	})

	if !errors.Is(err, domain.ErrBathhouseNotActive) {
		t.Errorf("should fail for inactive bathhouse, got: %v", err)
	}
}

func TestBookingService_Create_TooManyGuests(t *testing.T) {
	svc, bhRepo, _, _, _, _, _, _ := newBookingService()
	bh := createBathhouse(t, bhRepo, uuid.New()) // MaxGuests = 10

	start := time.Now().Add(24 * time.Hour)
	_, err := svc.Create(context.Background(), uuid.New(), service.CreateBookingInput{
		BathhouseID: bh.ID,
		StartTime:   start,
		EndTime:     start.Add(2 * time.Hour),
		GuestCount:  15,
	})

	if !errors.Is(err, domain.ErrInvalidInput) {
		t.Errorf("should fail for too many guests, got: %v", err)
	}
}

func TestBookingService_Create_SlotUnavailable(t *testing.T) {
	svc, bhRepo, bookingRepo, _, _, _, _, _ := newBookingService()
	ownerID := uuid.New()
	bh := createBathhouse(t, bhRepo, ownerID)

	// Set start time to tomorrow at a specific hour (e.g., 10:00) to avoid crossing midnight
	now := time.Now()
	start := time.Date(now.Year(), now.Month(), now.Day()+1, 10, 0, 0, 0, now.Location())
	end := start.Add(2 * time.Hour)

	// Create first booking
	existingBooking := &domain.Booking{
		ID: uuid.New(), UserID: uuid.New(), BathhouseID: bh.ID,
		StartTime: start, EndTime: end, GuestCount: 2,
		TotalPrice: 10000, Status: domain.BookingConfirmed,
	}
	_ = bookingRepo.Create(context.Background(), existingBooking)

	// Try to create overlapping booking
	_, err := svc.Create(context.Background(), uuid.New(), service.CreateBookingInput{
		BathhouseID: bh.ID,
		StartTime:   start.Add(time.Hour),
		EndTime:     end.Add(time.Hour),
		GuestCount:  3,
	})

	if !errors.Is(err, domain.ErrSlotUnavailable) {
		t.Errorf("should fail for unavailable slot, got: %v", err)
	}
}

func TestBookingService_Cancel_ClientOwnBooking(t *testing.T) {
	svc, bhRepo, bookingRepo, _, _, _, _, _ := newBookingService()
	bh := createBathhouse(t, bhRepo, uuid.New())
	clientID := uuid.New()

	start := time.Now().Add(24 * time.Hour) // well in advance
	booking := &domain.Booking{
		ID: uuid.New(), UserID: clientID, BathhouseID: bh.ID,
		StartTime: start, EndTime: start.Add(2 * time.Hour),
		GuestCount: 2, TotalPrice: 10000, Status: domain.BookingPending,
	}
	_ = bookingRepo.Create(context.Background(), booking)

	err := svc.Cancel(context.Background(), clientID, domain.RoleClient, booking.ID)
	if err != nil {
		t.Fatalf("client should cancel own booking: %v", err)
	}
}

func TestBookingService_Cancel_ClientOtherBookingForbidden(t *testing.T) {
	svc, bhRepo, bookingRepo, _, _, _, _, _ := newBookingService()
	bh := createBathhouse(t, bhRepo, uuid.New())
	otherClientID := uuid.New()

	start := time.Now().Add(24 * time.Hour)
	booking := &domain.Booking{
		ID: uuid.New(), UserID: uuid.New(), BathhouseID: bh.ID,
		StartTime: start, EndTime: start.Add(2 * time.Hour),
		GuestCount: 2, TotalPrice: 10000, Status: domain.BookingPending,
	}
	_ = bookingRepo.Create(context.Background(), booking)

	err := svc.Cancel(context.Background(), otherClientID, domain.RoleClient, booking.ID)
	if !errors.Is(err, domain.ErrForbidden) {
		t.Errorf("client should be forbidden from cancelling other's booking, got: %v", err)
	}
}

func TestBookingService_Cancel_ClientTooLate(t *testing.T) {
	svc, bhRepo, bookingRepo, _, _, _, _, _ := newBookingService()
	bh := createBathhouse(t, bhRepo, uuid.New())
	clientID := uuid.New()

	start := time.Now().Add(30 * time.Minute) // less than 2 hours
	booking := &domain.Booking{
		ID: uuid.New(), UserID: clientID, BathhouseID: bh.ID,
		StartTime: start, EndTime: start.Add(2 * time.Hour),
		GuestCount: 2, TotalPrice: 10000, Status: domain.BookingPending,
	}
	_ = bookingRepo.Create(context.Background(), booking)

	err := svc.Cancel(context.Background(), clientID, domain.RoleClient, booking.ID)
	if !errors.Is(err, domain.ErrBookingCancelLate) {
		t.Errorf("should fail for too late cancel, got: %v", err)
	}
}

func TestBookingService_Cancel_OwnerCanCancelAnytime(t *testing.T) {
	svc, bhRepo, bookingRepo, _, _, _, _, _ := newBookingService()
	ownerID := uuid.New()
	bh := createBathhouse(t, bhRepo, ownerID)

	start := time.Now().Add(30 * time.Minute) // less than 2 hours - but owner can cancel
	booking := &domain.Booking{
		ID: uuid.New(), UserID: uuid.New(), BathhouseID: bh.ID,
		StartTime: start, EndTime: start.Add(2 * time.Hour),
		GuestCount: 2, TotalPrice: 10000, Status: domain.BookingPending,
	}
	_ = bookingRepo.Create(context.Background(), booking)

	err := svc.Cancel(context.Background(), ownerID, domain.RoleOwner, booking.ID)
	if err != nil {
		t.Errorf("owner should cancel booking for own bathhouse anytime, got: %v", err)
	}
}

func TestBookingService_Confirm_OwnerAllowed(t *testing.T) {
	svc, bhRepo, bookingRepo, _, _, _, _, _ := newBookingService()
	ownerID := uuid.New()
	bh := createBathhouse(t, bhRepo, ownerID)

	start := time.Now().Add(24 * time.Hour)
	booking := &domain.Booking{
		ID: uuid.New(), UserID: uuid.New(), BathhouseID: bh.ID,
		StartTime: start, EndTime: start.Add(2 * time.Hour),
		GuestCount: 2, TotalPrice: 10000, Status: domain.BookingPending,
	}
	_ = bookingRepo.Create(context.Background(), booking)

	err := svc.Confirm(context.Background(), ownerID, domain.RoleOwner, booking.ID)
	if err != nil {
		t.Fatalf("owner should confirm: %v", err)
	}
}

func TestBookingService_Confirm_RepresentativeAllowed(t *testing.T) {
	svc, bhRepo, bookingRepo, repRepo, _, _, _, _ := newBookingService()
	ownerID := uuid.New()
	repUserID := uuid.New()
	bh := createBathhouse(t, bhRepo, ownerID)

	rep := &domain.Representative{
		ID: uuid.New(), UserID: repUserID, BathhouseID: bh.ID, OwnerID: ownerID,
	}
	_ = repRepo.Create(context.Background(), rep)

	start := time.Now().Add(24 * time.Hour)
	booking := &domain.Booking{
		ID: uuid.New(), UserID: uuid.New(), BathhouseID: bh.ID,
		StartTime: start, EndTime: start.Add(2 * time.Hour),
		GuestCount: 2, TotalPrice: 10000, Status: domain.BookingPending,
	}
	_ = bookingRepo.Create(context.Background(), booking)

	err := svc.Confirm(context.Background(), repUserID, domain.RoleRepresentative, booking.ID)
	if err != nil {
		t.Errorf("representative should confirm for assigned bathhouse, got: %v", err)
	}
}

func TestBookingService_Confirm_ClientForbidden(t *testing.T) {
	svc, bhRepo, bookingRepo, _, _, _, _, _ := newBookingService()
	bh := createBathhouse(t, bhRepo, uuid.New())
	clientID := uuid.New()

	start := time.Now().Add(24 * time.Hour)
	booking := &domain.Booking{
		ID: uuid.New(), UserID: clientID, BathhouseID: bh.ID,
		StartTime: start, EndTime: start.Add(2 * time.Hour),
		GuestCount: 2, TotalPrice: 10000, Status: domain.BookingPending,
	}
	_ = bookingRepo.Create(context.Background(), booking)

	err := svc.Confirm(context.Background(), clientID, domain.RoleClient, booking.ID)
	if !errors.Is(err, domain.ErrForbidden) {
		t.Errorf("client should be forbidden from confirming, got: %v", err)
	}
}

func TestBookingService_Reject_OwnerAllowed(t *testing.T) {
	svc, bhRepo, bookingRepo, _, _, _, _, _ := newBookingService()
	ownerID := uuid.New()
	bh := createBathhouse(t, bhRepo, ownerID)

	start := time.Now().Add(24 * time.Hour)
	booking := &domain.Booking{
		ID: uuid.New(), UserID: uuid.New(), BathhouseID: bh.ID,
		StartTime: start, EndTime: start.Add(2 * time.Hour),
		GuestCount: 2, TotalPrice: 10000, Status: domain.BookingPending,
	}
	_ = bookingRepo.Create(context.Background(), booking)

	err := svc.Reject(context.Background(), ownerID, domain.RoleOwner, booking.ID)
	if err != nil {
		t.Fatalf("owner should reject: %v", err)
	}
}

func TestBookingService_ListByBathhouse_OwnerAllowed(t *testing.T) {
	svc, bhRepo, _, _, _, _, _, _ := newBookingService()
	ownerID := uuid.New()
	bh := createBathhouse(t, bhRepo, ownerID)

	_, err := svc.ListByBathhouse(context.Background(), ownerID, domain.RoleOwner, bh.ID, 1, 10)
	if err != nil {
		t.Errorf("owner should list bookings for own bathhouse: %v", err)
	}
}

func TestBookingService_ListByBathhouse_ClientForbidden(t *testing.T) {
	svc, bhRepo, _, _, _, _, _, _ := newBookingService()
	bh := createBathhouse(t, bhRepo, uuid.New())
	clientID := uuid.New()

	_, err := svc.ListByBathhouse(context.Background(), clientID, domain.RoleClient, bh.ID, 1, 10)
	if !errors.Is(err, domain.ErrForbidden) {
		t.Errorf("client should be forbidden from listing bathhouse bookings, got: %v", err)
	}
}

func TestBookingService_ListByUser(t *testing.T) {
	svc, bhRepo, bookingRepo, _, _, _, _, _ := newBookingService()
	bh := createBathhouse(t, bhRepo, uuid.New())
	clientID := uuid.New()

	start := time.Now().Add(24 * time.Hour)
	booking := &domain.Booking{
		ID: uuid.New(), UserID: clientID, BathhouseID: bh.ID,
		StartTime: start, EndTime: start.Add(2 * time.Hour),
		GuestCount: 2, TotalPrice: 10000, Status: domain.BookingPending,
	}
	_ = bookingRepo.Create(context.Background(), booking)

	result, err := svc.ListByUser(context.Background(), clientID, 1, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.TotalCount != 1 {
		t.Errorf("totalCount = %d, want 1", result.TotalCount)
	}
}

func TestBookingService_GetAvailableSlots(t *testing.T) {
	svc, bhRepo, bookingRepo, _, _, _, _, _ := newBookingService()
	ownerID := uuid.New()

	bh := &domain.Bathhouse{
		ID: uuid.New(), OwnerID: ownerID, Name: "Test Bath",
		Address: "123 St", CityID: 1, PricePerHour: 5000,
		MinDuration: 1, MaxGuests: 10, Status: domain.BathhouseStatusActive,
		WorkingHours: []domain.WorkingHours{
			{DayOfWeek: 0, OpenTime: "09:00", CloseTime: "13:00"}, // Monday
		},
	}
	_ = bhRepo.Create(context.Background(), bh)

	// Find next Monday
	now := time.Now()
	daysUntilMonday := (8 - int(now.Weekday())) % 7
	if daysUntilMonday == 0 {
		daysUntilMonday = 7
	}
	monday := time.Date(now.Year(), now.Month(), now.Day()+daysUntilMonday, 0, 0, 0, 0, time.UTC)

	// Create a booking that occupies 10:00-11:00
	existingBooking := &domain.Booking{
		ID: uuid.New(), UserID: uuid.New(), BathhouseID: bh.ID,
		StartTime:  time.Date(monday.Year(), monday.Month(), monday.Day(), 10, 0, 0, 0, time.UTC),
		EndTime:    time.Date(monday.Year(), monday.Month(), monday.Day(), 11, 0, 0, 0, time.UTC),
		GuestCount: 2, TotalPrice: 5000, Status: domain.BookingConfirmed,
	}
	_ = bookingRepo.Create(context.Background(), existingBooking)

	slots, err := svc.GetAvailableSlots(context.Background(), bh.ID, monday)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(slots) != 4 { // 09:00-10:00, 10:00-11:00, 11:00-12:00, 12:00-13:00
		t.Errorf("expected 4 slots, got %d", len(slots))
	}

	// The 10:00-11:00 slot should be unavailable
	for _, slot := range slots {
		if slot.StartTime.Hour() == 10 && slot.Available {
			t.Error("10:00-11:00 slot should be unavailable")
		}
		if slot.StartTime.Hour() == 9 && !slot.Available {
			t.Error("09:00-10:00 slot should be available")
		}
	}
}

func TestBookingService_GetAvailableSlots_ClosedDay(t *testing.T) {
	svc, bhRepo, _, _, _, _, _, _ := newBookingService()
	ownerID := uuid.New()

	bh := &domain.Bathhouse{
		ID: uuid.New(), OwnerID: ownerID, Name: "Test Bath",
		Address: "123 St", CityID: 1, PricePerHour: 5000,
		MinDuration: 1, MaxGuests: 10, Status: domain.BathhouseStatusActive,
		WorkingHours: []domain.WorkingHours{
			{DayOfWeek: 0, OpenTime: "09:00", CloseTime: "18:00"}, // Monday only
		},
	}
	_ = bhRepo.Create(context.Background(), bh)

	// Find next Tuesday
	now := time.Now()
	daysUntilTuesday := (9 - int(now.Weekday())) % 7
	if daysUntilTuesday == 0 {
		daysUntilTuesday = 7
	}
	tuesday := time.Date(now.Year(), now.Month(), now.Day()+daysUntilTuesday, 0, 0, 0, 0, time.UTC)

	slots, err := svc.GetAvailableSlots(context.Background(), bh.ID, tuesday)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if slots != nil {
		t.Errorf("expected nil slots for closed day, got %d", len(slots))
	}
}

func TestBookingService_Cancel_AlreadyCancelled(t *testing.T) {
	svc, bhRepo, bookingRepo, _, _, _, _, _ := newBookingService()
	bh := createBathhouse(t, bhRepo, uuid.New())
	clientID := uuid.New()

	start := time.Now().Add(24 * time.Hour)
	booking := &domain.Booking{
		ID: uuid.New(), UserID: clientID, BathhouseID: bh.ID,
		StartTime: start, EndTime: start.Add(2 * time.Hour),
		GuestCount: 2, TotalPrice: 10000, Status: domain.BookingCancelled,
	}
	_ = bookingRepo.Create(context.Background(), booking)

	err := svc.Cancel(context.Background(), clientID, domain.RoleClient, booking.ID)
	if !errors.Is(err, domain.ErrInvalidInput) {
		t.Errorf("should fail for already cancelled booking, got: %v", err)
	}
}

func TestBookingService_Confirm_AlreadyConfirmed(t *testing.T) {
	svc, bhRepo, bookingRepo, _, _, _, _, _ := newBookingService()
	ownerID := uuid.New()
	bh := createBathhouse(t, bhRepo, ownerID)

	start := time.Now().Add(24 * time.Hour)
	booking := &domain.Booking{
		ID: uuid.New(), UserID: uuid.New(), BathhouseID: bh.ID,
		StartTime: start, EndTime: start.Add(2 * time.Hour),
		GuestCount: 2, TotalPrice: 10000, Status: domain.BookingConfirmed,
	}
	_ = bookingRepo.Create(context.Background(), booking)

	err := svc.Confirm(context.Background(), ownerID, domain.RoleOwner, booking.ID)
	if !errors.Is(err, domain.ErrInvalidInput) {
		t.Errorf("should fail for already confirmed booking, got: %v", err)
	}
}

func TestBookingService_Create_ShortDuration(t *testing.T) {
	svc, bhRepo, _, _, _, _, _, _ := newBookingService()
	bh := createBathhouse(t, bhRepo, uuid.New()) // MinDuration = 1

	start := time.Now().Add(24 * time.Hour)
	_, err := svc.Create(context.Background(), uuid.New(), service.CreateBookingInput{
		BathhouseID: bh.ID,
		StartTime:   start,
		EndTime:     start.Add(30 * time.Minute), // less than 1 hour min
		GuestCount:  2,
	})

	if !errors.Is(err, domain.ErrInvalidInput) {
		t.Errorf("should fail for too short duration, got: %v", err)
	}
}

func TestBookingService_Create_WithDynamicPricing(t *testing.T) {
	svc, bhRepo, _, _, _, pricingRepo, _, _ := newBookingService()
	ownerID := uuid.New()
	clientID := uuid.New()
	bh := createBathhouse(t, bhRepo, ownerID)

	// Create a pricing rule: 1.5x multiplier for the booking time
	basePrice := bh.PricePerHour
	now := time.Now()
	start := time.Date(now.Year(), now.Month(), now.Day()+1, 18, 0, 0, 0, now.Location())
	end := start.Add(2 * time.Hour)

	// Create a pricing rule that applies to 18:00-20:00
	rule := &domain.PricingRule{
		ID:          uuid.New(),
		BathhouseID: bh.ID,
		Name:        "Evening Boost",
		Type:        domain.RuleTypeTimeRange,
		Multiplier:  1.5,
		TimeFrom:    toPtr("18:00"),
		TimeTo:      toPtr("20:00"),
		Priority:    10,
		IsActive:    true,
	}

	_ = pricingRepo.Create(context.Background(), rule)

	result, err := svc.Create(context.Background(), clientID, service.CreateBookingInput{
		BathhouseID: bh.ID,
		StartTime:   start,
		EndTime:     end,
		GuestCount:  5,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Expected: basePrice * 2 hours * 1.5 multiplier = basePrice * 3
	expectedPrice := int64(float64(basePrice*2) * 1.5)
	if result.Booking.TotalPrice != expectedPrice {
		t.Errorf("totalPrice = %d, want %d (basePrice=%d, multiplier=1.5)", result.Booking.TotalPrice, expectedPrice, basePrice)
	}
}

func TestBookingService_GetAvailableSlots_WithPricing(t *testing.T) {
	svc, bhRepo, _, _, _, _, _, _ := newBookingService()
	ownerID := uuid.New()

	bh := &domain.Bathhouse{
		ID: uuid.New(), OwnerID: ownerID, Name: "Test Bath",
		Address: "123 St", CityID: 1, PricePerHour: 5000,
		MinDuration: 1, MaxGuests: 10, Status: domain.BathhouseStatusActive,
		WorkingHours: []domain.WorkingHours{
			{DayOfWeek: 0, OpenTime: "09:00", CloseTime: "13:00"}, // Monday
		},
	}
	_ = bhRepo.Create(context.Background(), bh)

	// Find next Monday
	now := time.Now()
	daysUntilMonday := (8 - int(now.Weekday())) % 7
	if daysUntilMonday == 0 {
		daysUntilMonday = 7
	}
	monday := time.Date(now.Year(), now.Month(), now.Day()+daysUntilMonday, 0, 0, 0, 0, time.UTC)

	slots, err := svc.GetAvailableSlots(context.Background(), bh.ID, monday)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(slots) == 0 {
		t.Errorf("expected at least one slot, got 0")
	}

	// Verify each slot has a price
	for i, slot := range slots {
		if slot.Price <= 0 {
			t.Errorf("slot %d: price should be positive, got %d", i, slot.Price)
		}
		// Without pricing rules, price should equal base price
		if slot.Price != bh.PricePerHour {
			t.Errorf("slot %d: price = %d, want %d (base price)", i, slot.Price, bh.PricePerHour)
		}
	}
}

func toPtr(s string) *string {
	return &s
}

func TestBookingService_Complete_EarnsLoyaltyPoints(t *testing.T) {
	svc, bhRepo, bookingRepo, _, _, _, loyaltySvc, _ := newBookingService()
	ownerID := uuid.New()
	clientID := uuid.New()
	bh := createBathhouse(t, bhRepo, ownerID)

	// Create and confirm a booking that ended in the past
	start := time.Now().Add(-3 * time.Hour)
	end := start.Add(2 * time.Hour)
	booking := &domain.Booking{
		ID: uuid.New(), UserID: clientID, BathhouseID: bh.ID,
		StartTime: start, EndTime: end,
		GuestCount: 2, TotalPrice: 10000, Status: domain.BookingConfirmed,
	}
	_ = bookingRepo.Create(context.Background(), booking)

	result, err := svc.Complete(context.Background(), ownerID, domain.RoleOwner, booking.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Bronze level: multiplier 1.0, so 10000/100 * 1.0 = 100 points
	if result.EarnedPoints != 100 {
		t.Errorf("earnedPoints = %d, want 100", result.EarnedPoints)
	}

	// Verify loyalty account was updated
	account, err := loyaltySvc.GetAccount(context.Background(), clientID)
	if err != nil {
		t.Fatalf("failed to get loyalty account: %v", err)
	}
	if account.Points != 100 {
		t.Errorf("account points = %d, want 100", account.Points)
	}
}

func TestBookingService_Create_WithLoyaltyDiscount(t *testing.T) {
	svc, bhRepo, _, _, _, _, _, loyaltyRepo := newBookingService()
	ownerID := uuid.New()
	clientID := uuid.New()
	bh := createBathhouse(t, bhRepo, ownerID)

	// Set up a Silver level account (3% discount)
	if err := loyaltyRepo.CreateAccount(context.Background(), &domain.LoyaltyAccount{
		UserID:     clientID,
		Level:      domain.LoyaltySilver,
		Points:     500,
		VisitCount: 5,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}); err != nil {
		t.Fatalf("failed to create loyalty account: %v", err)
	}

	now := time.Now()
	start := time.Date(now.Year(), now.Month(), now.Day()+1, 10, 0, 0, 0, now.Location())
	end := start.Add(2 * time.Hour)

	result, err := svc.Create(context.Background(), clientID, service.CreateBookingInput{
		BathhouseID: bh.ID,
		StartTime:   start,
		EndTime:     end,
		GuestCount:  5,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Base price: 5000 * 2 = 10000. Silver discount 3% = 300
	expectedDiscount := int64(300)
	expectedPrice := int64(10000 - 300)
	if result.LoyaltyDiscount != expectedDiscount {
		t.Errorf("loyaltyDiscount = %d, want %d", result.LoyaltyDiscount, expectedDiscount)
	}
	if result.Booking.TotalPrice != expectedPrice {
		t.Errorf("totalPrice = %d, want %d", result.Booking.TotalPrice, expectedPrice)
	}
}

func TestBookingService_Create_WithPointsSpending(t *testing.T) {
	svc, bhRepo, _, _, _, _, _, loyaltyRepo := newBookingService()
	ownerID := uuid.New()
	clientID := uuid.New()
	bh := createBathhouse(t, bhRepo, ownerID)

	// Set up account with some points
	if err := loyaltyRepo.CreateAccount(context.Background(), &domain.LoyaltyAccount{
		UserID:     clientID,
		Level:      domain.LoyaltyBronze,
		Points:     2000,
		VisitCount: 2,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}); err != nil {
		t.Fatalf("failed to create loyalty account: %v", err)
	}

	now := time.Now()
	start := time.Date(now.Year(), now.Month(), now.Day()+1, 10, 0, 0, 0, now.Location())
	end := start.Add(2 * time.Hour)

	result, err := svc.Create(context.Background(), clientID, service.CreateBookingInput{
		BathhouseID: bh.ID,
		StartTime:   start,
		EndTime:     end,
		GuestCount:  5,
		UsePoints:   1000,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Base price: 5000 * 2 = 10000. Bronze: no discount. Spend 1000 points -> price = 9000
	if result.PointsSpent != 1000 {
		t.Errorf("pointsSpent = %d, want 1000", result.PointsSpent)
	}
	if result.Booking.TotalPrice != 9000 {
		t.Errorf("totalPrice = %d, want 9000", result.Booking.TotalPrice)
	}
}

func TestBookingService_Create_InsufficientPoints(t *testing.T) {
	svc, bhRepo, _, _, _, _, _, loyaltyRepo := newBookingService()
	ownerID := uuid.New()
	clientID := uuid.New()
	bh := createBathhouse(t, bhRepo, ownerID)

	// Set up account with few points
	if err := loyaltyRepo.CreateAccount(context.Background(), &domain.LoyaltyAccount{
		UserID:     clientID,
		Level:      domain.LoyaltyBronze,
		Points:     50,
		VisitCount: 1,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}); err != nil {
		t.Fatalf("failed to create loyalty account: %v", err)
	}

	now := time.Now()
	start := time.Date(now.Year(), now.Month(), now.Day()+1, 10, 0, 0, 0, now.Location())
	end := start.Add(2 * time.Hour)

	_, err := svc.Create(context.Background(), clientID, service.CreateBookingInput{
		BathhouseID: bh.ID,
		StartTime:   start,
		EndTime:     end,
		GuestCount:  5,
		UsePoints:   1000,
	})

	if !errors.Is(err, domain.ErrInsufficientPoints) {
		t.Errorf("should fail with insufficient points, got: %v", err)
	}
}

func TestBookingService_Create_PointsExceedPrice(t *testing.T) {
	svc, bhRepo, _, _, _, _, _, loyaltyRepo := newBookingService()
	ownerID := uuid.New()
	clientID := uuid.New()
	bh := createBathhouse(t, bhRepo, ownerID)

	// Set up account with a lot of points
	if err := loyaltyRepo.CreateAccount(context.Background(), &domain.LoyaltyAccount{
		UserID:     clientID,
		Level:      domain.LoyaltyBronze,
		Points:     999999,
		VisitCount: 1,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}); err != nil {
		t.Fatalf("failed to create loyalty account: %v", err)
	}

	now := time.Now()
	start := time.Date(now.Year(), now.Month(), now.Day()+1, 10, 0, 0, 0, now.Location())
	end := start.Add(2 * time.Hour)

	_, err := svc.Create(context.Background(), clientID, service.CreateBookingInput{
		BathhouseID: bh.ID,
		StartTime:   start,
		EndTime:     end,
		GuestCount:  5,
		UsePoints:   20000, // More than 10000 price
	})

	if !errors.Is(err, domain.ErrInvalidInput) {
		t.Errorf("should fail when points exceed price, got: %v", err)
	}
}

func newBookingServiceWithReferral() (service.BookingService, *mock.BathhouseRepo, *mock.BookingRepo, service.ReferralService, *mock.UserRepo) {
	bhRepo := mock.NewBathhouseRepo()
	bookingRepo := mock.NewBookingRepo()
	repRepo := mock.NewRepresentativeRepo()
	pricingRepo := mock.NewPricingRuleRepo()
	loyaltyRepo := mock.NewLoyaltyRepo()
	referralRepo := mock.NewReferralRepo()
	access := service.NewAccessChecker(repRepo, bhRepo)
	log := logger.New(logger.LevelWarn)
	pricingSvc := service.NewPricingService(pricingRepo, bhRepo, access, log)
	loyaltySvc := service.NewLoyaltyService(loyaltyRepo, log)
	userRepo := mock.NewUserRepo()
	referralSvc := service.NewReferralService(referralRepo, userRepo, log)
	svc := service.NewBookingService(bookingRepo, bhRepo, pricingSvc, loyaltySvc, referralSvc, &noopPromoService{}, &noopPaymentService{}, access, &noopNotifService{}, log)
	return svc, bhRepo, bookingRepo, referralSvc, userRepo
}

func TestBookingService_Complete_CompletesReferral(t *testing.T) {
	svc, bhRepo, bookingRepo, referralSvc, userRepo := newBookingServiceWithReferral()
	ownerID := uuid.New()

	// Create referrer and referee users
	referrer := &domain.User{ID: uuid.New(), Email: "referrer@test.com", Name: "Referrer", Role: domain.RoleClient, IsActive: true}
	referee := &domain.User{ID: uuid.New(), Email: "referee@test.com", Name: "Referee", Role: domain.RoleClient, IsActive: true}
	_ = userRepo.Create(context.Background(), referrer)
	_ = userRepo.Create(context.Background(), referee)

	// Generate referral code and register referral
	code, err := referralSvc.GenerateCode(context.Background(), referrer.ID)
	if err != nil {
		t.Fatalf("failed to generate code: %v", err)
	}
	if err := referralSvc.RegisterReferral(context.Background(), code, referee.ID); err != nil {
		t.Fatalf("failed to register referral: %v", err)
	}

	bh := createBathhouse(t, bhRepo, ownerID)

	// Create and confirm a booking that ended in the past
	start := time.Now().Add(-3 * time.Hour)
	end := start.Add(2 * time.Hour)
	booking := &domain.Booking{
		ID: uuid.New(), UserID: referee.ID, BathhouseID: bh.ID,
		StartTime: start, EndTime: end,
		GuestCount: 2, TotalPrice: 10000, Status: domain.BookingConfirmed,
	}
	_ = bookingRepo.Create(context.Background(), booking)

	// Complete the booking - should trigger CompleteReferral and credit bonuses
	result, err := svc.Complete(context.Background(), ownerID, domain.RoleOwner, booking.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Booking == nil {
		t.Fatal("booking should not be nil")
	}

	// Verify referrer got bonus
	referrerBalance, err := referralSvc.GetBalance(context.Background(), referrer.ID)
	if err != nil {
		t.Fatalf("failed to get referrer balance: %v", err)
	}
	if referrerBalance.Balance != 50000 {
		t.Errorf("referrer balance = %d, want 50000", referrerBalance.Balance)
	}

	// Verify referee got bonus
	refereeBalance, err := referralSvc.GetBalance(context.Background(), referee.ID)
	if err != nil {
		t.Fatalf("failed to get referee balance: %v", err)
	}
	if refereeBalance.Balance != 50000 {
		t.Errorf("referee balance = %d, want 50000", refereeBalance.Balance)
	}
}

func TestBookingService_Create_WithReferralBonus(t *testing.T) {
	svc, bhRepo, _, referralSvc, userRepo := newBookingServiceWithReferral()
	ownerID := uuid.New()
	bh := createBathhouse(t, bhRepo, ownerID)

	// Create referrer and client users, set up referral with balance
	referrer := &domain.User{ID: uuid.New(), Email: "ref-r@test.com", Name: "Referrer", Role: domain.RoleClient, IsActive: true}
	client := &domain.User{ID: uuid.New(), Email: "ref-e@test.com", Name: "Client", Role: domain.RoleClient, IsActive: true}
	_ = userRepo.Create(context.Background(), referrer)
	_ = userRepo.Create(context.Background(), client)

	code, _ := referralSvc.GenerateCode(context.Background(), referrer.ID)
	_ = referralSvc.RegisterReferral(context.Background(), code, client.ID)
	_ = referralSvc.CompleteReferral(context.Background(), client.ID) // Gives 50000 to both

	now := time.Now()
	start := time.Date(now.Year(), now.Month(), now.Day()+1, 10, 0, 0, 0, now.Location())
	end := start.Add(2 * time.Hour)

	// Test that referral bonus exceeding price is rejected
	_, err := svc.Create(context.Background(), client.ID, service.CreateBookingInput{
		BathhouseID:      bh.ID,
		StartTime:        start,
		EndTime:          end,
		GuestCount:       5,
		UseReferralBonus: 20000, // More than base price (10000)
	})

	if !errors.Is(err, domain.ErrInvalidInput) {
		t.Errorf("should fail when referral bonus exceeds price, got: %v", err)
	}

	// Test successful referral bonus usage
	result, err := svc.Create(context.Background(), client.ID, service.CreateBookingInput{
		BathhouseID:      bh.ID,
		StartTime:        start,
		EndTime:          end,
		GuestCount:       5,
		UseReferralBonus: 5000,
	})
	if err != nil {
		t.Fatalf("unexpected error creating booking with referral bonus: %v", err)
	}
	if result.ReferralBonusUsed != 5000 {
		t.Errorf("referral bonus used = %d, want 5000", result.ReferralBonusUsed)
	}

	// Verify balance was deducted
	balance, _ := referralSvc.GetBalance(context.Background(), client.ID)
	if balance.Balance != 45000 {
		t.Errorf("client referral balance = %d, want 45000", balance.Balance)
	}
}

func newBookingServiceWithPromo() (service.BookingService, *mock.BathhouseRepo, *mock.BookingRepo, service.PromoService) {
	bhRepo := mock.NewBathhouseRepo()
	bookingRepo := mock.NewBookingRepo()
	repRepo := mock.NewRepresentativeRepo()
	pricingRepo := mock.NewPricingRuleRepo()
	loyaltyRepo := mock.NewLoyaltyRepo()
	promoRepo := mock.NewPromoCodeRepo()
	access := service.NewAccessChecker(repRepo, bhRepo)
	log := logger.New(logger.LevelWarn)
	pricingSvc := service.NewPricingService(pricingRepo, bhRepo, access, log)
	loyaltySvc := service.NewLoyaltyService(loyaltyRepo, log)
	promoSvc := service.NewPromoService(promoRepo, access, log)
	svc := service.NewBookingService(bookingRepo, bhRepo, pricingSvc, loyaltySvc, &noopReferralService{}, promoSvc, &noopPaymentService{}, access, &noopNotifService{}, log)
	return svc, bhRepo, bookingRepo, promoSvc
}

func TestBookingService_Create_WithPromoCodePercentage(t *testing.T) {
	svc, bhRepo, _, promoSvc := newBookingServiceWithPromo()
	ownerID := uuid.New()
	clientID := uuid.New()
	bh := createBathhouse(t, bhRepo, ownerID)

	// Create a 10% promo code for this bathhouse
	bhID := bh.ID
	_, err := promoSvc.Create(context.Background(), ownerID, domain.RoleOwner, &domain.PromoCode{
		Code:        "SAVE10",
		Type:        domain.PromoTypePercentage,
		Value:       10,
		BathhouseID: &bhID,
		ValidFrom:   time.Now().Add(-time.Hour),
		ValidUntil:  time.Now().Add(24 * time.Hour),
	})
	if err != nil {
		t.Fatalf("failed to create promo: %v", err)
	}

	now := time.Now()
	start := time.Date(now.Year(), now.Month(), now.Day()+1, 10, 0, 0, 0, now.Location())
	end := start.Add(2 * time.Hour)

	result, err := svc.Create(context.Background(), clientID, service.CreateBookingInput{
		BathhouseID: bh.ID,
		StartTime:   start,
		EndTime:     end,
		GuestCount:  5,
		PromoCode:   "SAVE10",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Base price: 5000 * 2 = 10000. 10% discount = 1000. Final = 9000
	if result.PromoDiscount != 1000 {
		t.Errorf("promoDiscount = %d, want 1000", result.PromoDiscount)
	}
	if result.OriginalPrice != 10000 {
		t.Errorf("originalPrice = %d, want 10000", result.OriginalPrice)
	}
	if result.Booking.TotalPrice != 9000 {
		t.Errorf("totalPrice = %d, want 9000", result.Booking.TotalPrice)
	}
}

func TestBookingService_Create_WithPromoCodeFixedAmount(t *testing.T) {
	svc, bhRepo, _, promoSvc := newBookingServiceWithPromo()
	ownerID := uuid.New()
	clientID := uuid.New()
	bh := createBathhouse(t, bhRepo, ownerID)

	// Create a 2000 kopeck (20 rub) fixed discount promo code
	bhID := bh.ID
	_, err := promoSvc.Create(context.Background(), ownerID, domain.RoleOwner, &domain.PromoCode{
		Code:        "FLAT20",
		Type:        domain.PromoTypeFixedAmount,
		Value:       2000,
		BathhouseID: &bhID,
		ValidFrom:   time.Now().Add(-time.Hour),
		ValidUntil:  time.Now().Add(24 * time.Hour),
	})
	if err != nil {
		t.Fatalf("failed to create promo: %v", err)
	}

	now := time.Now()
	start := time.Date(now.Year(), now.Month(), now.Day()+1, 10, 0, 0, 0, now.Location())
	end := start.Add(2 * time.Hour)

	result, err := svc.Create(context.Background(), clientID, service.CreateBookingInput{
		BathhouseID: bh.ID,
		StartTime:   start,
		EndTime:     end,
		GuestCount:  5,
		PromoCode:   "FLAT20",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Base price: 5000 * 2 = 10000. Fixed discount 2000. Final = 8000
	if result.PromoDiscount != 2000 {
		t.Errorf("promoDiscount = %d, want 2000", result.PromoDiscount)
	}
	if result.Booking.TotalPrice != 8000 {
		t.Errorf("totalPrice = %d, want 8000", result.Booking.TotalPrice)
	}
}

func TestBookingService_Create_WithInvalidPromoCode(t *testing.T) {
	svc, bhRepo, _, _ := newBookingServiceWithPromo()
	ownerID := uuid.New()
	clientID := uuid.New()
	bh := createBathhouse(t, bhRepo, ownerID)

	now := time.Now()
	start := time.Date(now.Year(), now.Month(), now.Day()+1, 10, 0, 0, 0, now.Location())
	end := start.Add(2 * time.Hour)

	_, err := svc.Create(context.Background(), clientID, service.CreateBookingInput{
		BathhouseID: bh.ID,
		StartTime:   start,
		EndTime:     end,
		GuestCount:  5,
		PromoCode:   "NONEXISTENT",
	})

	if !errors.Is(err, domain.ErrPromoNotFound) {
		t.Errorf("should fail with promo not found, got: %v", err)
	}
}

func TestBookingService_Create_WithExpiredPromoCode(t *testing.T) {
	svc, bhRepo, _, promoSvc := newBookingServiceWithPromo()
	ownerID := uuid.New()
	clientID := uuid.New()
	bh := createBathhouse(t, bhRepo, ownerID)

	// Create an expired promo code
	bhID := bh.ID
	_, err := promoSvc.Create(context.Background(), ownerID, domain.RoleOwner, &domain.PromoCode{
		Code:        "EXPIRED",
		Type:        domain.PromoTypePercentage,
		Value:       10,
		BathhouseID: &bhID,
		ValidFrom:   time.Now().Add(-48 * time.Hour),
		ValidUntil:  time.Now().Add(-24 * time.Hour),
	})
	if err != nil {
		t.Fatalf("failed to create promo: %v", err)
	}

	now := time.Now()
	start := time.Date(now.Year(), now.Month(), now.Day()+1, 10, 0, 0, 0, now.Location())
	end := start.Add(2 * time.Hour)

	_, err = svc.Create(context.Background(), clientID, service.CreateBookingInput{
		BathhouseID: bh.ID,
		StartTime:   start,
		EndTime:     end,
		GuestCount:  5,
		PromoCode:   "EXPIRED",
	})

	if !errors.Is(err, domain.ErrPromoExpired) {
		t.Errorf("should fail with promo expired, got: %v", err)
	}
}

func TestBookingService_Create_WithPromoCodeNoDiscount(t *testing.T) {
	svc, bhRepo, _, _ := newBookingServiceWithPromo()
	ownerID := uuid.New()
	clientID := uuid.New()
	bh := createBathhouse(t, bhRepo, ownerID)

	// Create booking without promo code - should work as before
	now := time.Now()
	start := time.Date(now.Year(), now.Month(), now.Day()+1, 10, 0, 0, 0, now.Location())
	end := start.Add(2 * time.Hour)

	result, err := svc.Create(context.Background(), clientID, service.CreateBookingInput{
		BathhouseID: bh.ID,
		StartTime:   start,
		EndTime:     end,
		GuestCount:  5,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// No promo - originalPrice and promoDiscount should be 0 (omitted)
	if result.PromoDiscount != 0 {
		t.Errorf("promoDiscount = %d, want 0", result.PromoDiscount)
	}
	if result.Booking.TotalPrice != 10000 {
		t.Errorf("totalPrice = %d, want 10000", result.Booking.TotalPrice)
	}
}

// --- Booking + Payment integration tests ---

type trackingPaymentService struct {
	noopPaymentService
	refundCalled bool
	refundErr    error
	payment      *domain.Payment
}

func (t *trackingPaymentService) RefundPayment(_ context.Context, _ uuid.UUID) error {
	t.refundCalled = true
	return t.refundErr
}

func (t *trackingPaymentService) GetPaymentByBooking(_ context.Context, _ uuid.UUID) (*domain.Payment, error) {
	if t.payment != nil {
		return t.payment, nil
	}
	return nil, domain.ErrPaymentNotFound
}

func newBookingServiceWithPayment() (service.BookingService, *mock.BathhouseRepo, *mock.BookingRepo, *trackingPaymentService) {
	bhRepo := mock.NewBathhouseRepo()
	bookingRepo := mock.NewBookingRepo()
	repRepo := mock.NewRepresentativeRepo()
	pricingRepo := mock.NewPricingRuleRepo()
	loyaltyRepo := mock.NewLoyaltyRepo()
	access := service.NewAccessChecker(repRepo, bhRepo)
	log := logger.New(logger.LevelWarn)
	pricingSvc := service.NewPricingService(pricingRepo, bhRepo, access, log)
	loyaltySvc := service.NewLoyaltyService(loyaltyRepo, log)
	paymentSvc := &trackingPaymentService{}
	svc := service.NewBookingService(bookingRepo, bhRepo, pricingSvc, loyaltySvc, &noopReferralService{}, &noopPromoService{}, paymentSvc, access, &noopNotifService{}, log)
	return svc, bhRepo, bookingRepo, paymentSvc
}

func TestBookingService_Cancel_RefundsPayment(t *testing.T) {
	svc, bhRepo, bookingRepo, paymentSvc := newBookingServiceWithPayment()
	clientID := uuid.New()
	bh := createBathhouse(t, bhRepo, uuid.New())

	start := time.Now().Add(48 * time.Hour) // well in advance
	booking := &domain.Booking{
		ID: uuid.New(), UserID: clientID, BathhouseID: bh.ID,
		StartTime: start, EndTime: start.Add(2 * time.Hour),
		GuestCount: 2, TotalPrice: 10000, Status: domain.BookingConfirmed,
	}
	_ = bookingRepo.Create(context.Background(), booking)

	err := svc.Cancel(context.Background(), clientID, domain.RoleClient, booking.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !paymentSvc.refundCalled {
		t.Error("expected RefundPayment to be called on booking cancellation")
	}
}

func TestBookingService_Cancel_OwnerRefundsPayment(t *testing.T) {
	svc, bhRepo, bookingRepo, paymentSvc := newBookingServiceWithPayment()
	ownerID := uuid.New()
	bh := createBathhouse(t, bhRepo, ownerID)

	start := time.Now().Add(30 * time.Minute) // owner can cancel anytime
	booking := &domain.Booking{
		ID: uuid.New(), UserID: uuid.New(), BathhouseID: bh.ID,
		StartTime: start, EndTime: start.Add(2 * time.Hour),
		GuestCount: 2, TotalPrice: 10000, Status: domain.BookingPending,
	}
	_ = bookingRepo.Create(context.Background(), booking)

	err := svc.Cancel(context.Background(), ownerID, domain.RoleOwner, booking.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !paymentSvc.refundCalled {
		t.Error("expected RefundPayment to be called when owner cancels booking")
	}
}

func TestBookingService_Cancel_RefundErrorDoesNotBlockCancel(t *testing.T) {
	svc, bhRepo, bookingRepo, paymentSvc := newBookingServiceWithPayment()
	clientID := uuid.New()
	bh := createBathhouse(t, bhRepo, uuid.New())

	paymentSvc.refundErr = errors.New("payment provider error")

	start := time.Now().Add(48 * time.Hour)
	booking := &domain.Booking{
		ID: uuid.New(), UserID: clientID, BathhouseID: bh.ID,
		StartTime: start, EndTime: start.Add(2 * time.Hour),
		GuestCount: 2, TotalPrice: 10000, Status: domain.BookingConfirmed,
	}
	_ = bookingRepo.Create(context.Background(), booking)

	// Cancel should succeed even if refund fails
	err := svc.Cancel(context.Background(), clientID, domain.RoleClient, booking.ID)
	if err != nil {
		t.Fatalf("cancel should succeed even when refund fails: %v", err)
	}

	if !paymentSvc.refundCalled {
		t.Error("expected RefundPayment to be called")
	}
}
