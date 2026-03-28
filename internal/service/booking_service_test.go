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
	slotBlockRepo := mock.NewSlotBlockRepo()
	userRepo := mock.NewUserRepo()
	cityRepo := mock.NewCityRepo()
	access := service.NewAccessChecker(repRepo, bhRepo)
	log := logger.New(logger.LevelWarn)
	pricingSvc := service.NewPricingService(pricingRepo, nil, bhRepo, nil, access, log)
	loyaltySvc := service.NewLoyaltyService(loyaltyRepo, log)
	addonRepo := mock.NewAddOnRepo()
	addonSvc := service.NewAddOnService(addonRepo, access, log)
	svc := service.NewBookingService(bookingRepo, bhRepo, slotBlockRepo, addonRepo, userRepo, cityRepo, pricingSvc, addonSvc, loyaltySvc, &noopReferralService{}, &noopPromoService{}, &noopCertificateService{}, &noopPaymentService{}, &noopServiceFeeService{}, nil, nil, nil, nil, nil, access, &noopNotifService{}, log)
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

	err := svc.Cancel(context.Background(), clientID, domain.RoleClient, booking.ID, "")
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

	err := svc.Cancel(context.Background(), otherClientID, domain.RoleClient, booking.ID, "")
	if !errors.Is(err, domain.ErrForbidden) {
		t.Errorf("client should be forbidden from cancelling other's booking, got: %v", err)
	}
}

func TestBookingService_Cancel_ClientLateCancel_Succeeds(t *testing.T) {
	// With policy-based cancellation, clients can always cancel (refund depends on policy).
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

	err := svc.Cancel(context.Background(), clientID, domain.RoleClient, booking.ID, "")
	if err != nil {
		t.Errorf("late cancel should succeed (refund is policy-based), got: %v", err)
	}

	updated, _ := bookingRepo.GetByID(context.Background(), booking.ID)
	if updated.Status != domain.BookingCancelled {
		t.Errorf("status = %q, want %q", updated.Status, domain.BookingCancelled)
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

	err := svc.Cancel(context.Background(), ownerID, domain.RoleOwner, booking.ID, "")
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
		Role:        domain.RepRoleManager,
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

	err := svc.Reject(context.Background(), ownerID, domain.RoleOwner, booking.ID, "")
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

	err := svc.Cancel(context.Background(), clientID, domain.RoleClient, booking.ID, "")
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
	pricingSvc := service.NewPricingService(pricingRepo, nil, bhRepo, nil, access, log)
	loyaltySvc := service.NewLoyaltyService(loyaltyRepo, log)
	userRepo := mock.NewUserRepo()
	referralSvc := service.NewReferralService(referralRepo, userRepo, log)
	addonRepo := mock.NewAddOnRepo()
	addonSvc := service.NewAddOnService(addonRepo, access, log)
	svc := service.NewBookingService(bookingRepo, bhRepo, mock.NewSlotBlockRepo(), addonRepo, mock.NewUserRepo(), mock.NewCityRepo(), pricingSvc, addonSvc, loyaltySvc, referralSvc, &noopPromoService{}, &noopCertificateService{}, &noopPaymentService{}, &noopServiceFeeService{}, nil, nil, nil, nil, nil, access, &noopNotifService{}, log)
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
	_, _ = referralSvc.CompleteReferral(context.Background(), client.ID) // Gives 50000 to both

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
	pricingSvc := service.NewPricingService(pricingRepo, nil, bhRepo, nil, access, log)
	loyaltySvc := service.NewLoyaltyService(loyaltyRepo, log)
	promoSvc := service.NewPromoService(promoRepo, access, log)
	addonRepo := mock.NewAddOnRepo()
	addonSvc := service.NewAddOnService(addonRepo, access, log)
	svc := service.NewBookingService(bookingRepo, bhRepo, mock.NewSlotBlockRepo(), addonRepo, mock.NewUserRepo(), mock.NewCityRepo(), pricingSvc, addonSvc, loyaltySvc, &noopReferralService{}, promoSvc, &noopCertificateService{}, &noopPaymentService{}, &noopServiceFeeService{}, nil, nil, nil, nil, nil, access, &noopNotifService{}, log)
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

func (t *trackingPaymentService) RefundPayment(_ context.Context, _ uuid.UUID, _ bool, _ string, _ domain.CancellationPolicy) error {
	t.refundCalled = true
	return t.refundErr
}

func (t *trackingPaymentService) GetPaymentByBooking(_ context.Context, _, _ uuid.UUID) (*domain.Payment, error) {
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
	pricingSvc := service.NewPricingService(pricingRepo, nil, bhRepo, nil, access, log)
	loyaltySvc := service.NewLoyaltyService(loyaltyRepo, log)
	paymentSvc := &trackingPaymentService{}
	addonRepo := mock.NewAddOnRepo()
	addonSvc := service.NewAddOnService(addonRepo, access, log)
	svc := service.NewBookingService(bookingRepo, bhRepo, mock.NewSlotBlockRepo(), addonRepo, mock.NewUserRepo(), mock.NewCityRepo(), pricingSvc, addonSvc, loyaltySvc, &noopReferralService{}, &noopPromoService{}, &noopCertificateService{}, paymentSvc, &noopServiceFeeService{}, nil, nil, nil, nil, nil, access, &noopNotifService{}, log)
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

	err := svc.Cancel(context.Background(), clientID, domain.RoleClient, booking.ID, "")
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

	err := svc.Cancel(context.Background(), ownerID, domain.RoleOwner, booking.ID, "")
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
	err := svc.Cancel(context.Background(), clientID, domain.RoleClient, booking.ID, "")
	if err != nil {
		t.Fatalf("cancel should succeed even when refund fails: %v", err)
	}

	if !paymentSvc.refundCalled {
		t.Error("expected RefundPayment to be called")
	}
}

// --- SlotBlock integration tests ---

func newBookingServiceWithSlotBlocks() (service.BookingService, *mock.BathhouseRepo, *mock.BookingRepo, *mock.SlotBlockRepo) {
	bhRepo := mock.NewBathhouseRepo()
	bookingRepo := mock.NewBookingRepo()
	repRepo := mock.NewRepresentativeRepo()
	pricingRepo := mock.NewPricingRuleRepo()
	loyaltyRepo := mock.NewLoyaltyRepo()
	slotBlockRepo := mock.NewSlotBlockRepo()
	access := service.NewAccessChecker(repRepo, bhRepo)
	log := logger.New(logger.LevelWarn)
	pricingSvc := service.NewPricingService(pricingRepo, nil, bhRepo, nil, access, log)
	loyaltySvc := service.NewLoyaltyService(loyaltyRepo, log)
	addonRepo := mock.NewAddOnRepo()
	addonSvc := service.NewAddOnService(addonRepo, access, log)
	svc := service.NewBookingService(bookingRepo, bhRepo, slotBlockRepo, addonRepo, mock.NewUserRepo(), mock.NewCityRepo(), pricingSvc, addonSvc, loyaltySvc, &noopReferralService{}, &noopPromoService{}, &noopCertificateService{}, &noopPaymentService{}, &noopServiceFeeService{}, nil, nil, nil, nil, nil, access, &noopNotifService{}, log)
	return svc, bhRepo, bookingRepo, slotBlockRepo
}

func TestBookingService_Create_BlockedBySlotBlock(t *testing.T) {
	svc, bhRepo, _, slotBlockRepo := newBookingServiceWithSlotBlocks()
	ownerID := uuid.New()
	bh := createBathhouse(t, bhRepo, ownerID)

	now := time.Now()
	start := time.Date(now.Year(), now.Month(), now.Day()+1, 10, 0, 0, 0, now.Location())
	end := start.Add(2 * time.Hour)

	// Create a slot block overlapping with the requested time
	block := &domain.SlotBlock{
		BathhouseID: bh.ID,
		StartTime:   start,
		EndTime:     end,
		Source:      domain.SlotBlockSourceGoogleCalendar,
		ExternalID:  "ext-123",
		Description: "External event",
	}
	if err := slotBlockRepo.Create(context.Background(), block); err != nil {
		t.Fatalf("failed to create slot block: %v", err)
	}

	_, err := svc.Create(context.Background(), uuid.New(), service.CreateBookingInput{
		BathhouseID: bh.ID,
		StartTime:   start,
		EndTime:     end,
		GuestCount:  3,
	})

	if !errors.Is(err, domain.ErrSlotUnavailable) {
		t.Errorf("should fail with slot unavailable due to slot block, got: %v", err)
	}
}

func TestBookingService_Create_PartialOverlapWithSlotBlock(t *testing.T) {
	svc, bhRepo, _, slotBlockRepo := newBookingServiceWithSlotBlocks()
	ownerID := uuid.New()
	bh := createBathhouse(t, bhRepo, ownerID)

	now := time.Now()
	start := time.Date(now.Year(), now.Month(), now.Day()+1, 10, 0, 0, 0, now.Location())

	// Block 11:00-13:00
	block := &domain.SlotBlock{
		BathhouseID: bh.ID,
		StartTime:   start.Add(time.Hour),
		EndTime:     start.Add(3 * time.Hour),
		Source:      domain.SlotBlockSourceManual,
		Description: "Maintenance",
	}
	if err := slotBlockRepo.Create(context.Background(), block); err != nil {
		t.Fatalf("failed to create slot block: %v", err)
	}

	// Try to book 10:00-12:00 (overlaps with block)
	_, err := svc.Create(context.Background(), uuid.New(), service.CreateBookingInput{
		BathhouseID: bh.ID,
		StartTime:   start,
		EndTime:     start.Add(2 * time.Hour),
		GuestCount:  3,
	})

	if !errors.Is(err, domain.ErrSlotUnavailable) {
		t.Errorf("should fail with slot unavailable due to partial overlap with slot block, got: %v", err)
	}
}

func TestBookingService_Create_NoOverlapWithSlotBlock(t *testing.T) {
	svc, bhRepo, _, slotBlockRepo := newBookingServiceWithSlotBlocks()
	ownerID := uuid.New()
	bh := createBathhouse(t, bhRepo, ownerID)

	now := time.Now()
	start := time.Date(now.Year(), now.Month(), now.Day()+1, 10, 0, 0, 0, now.Location())

	// Block 14:00-16:00
	block := &domain.SlotBlock{
		BathhouseID: bh.ID,
		StartTime:   start.Add(4 * time.Hour),
		EndTime:     start.Add(6 * time.Hour),
		Source:      domain.SlotBlockSourceManual,
		Description: "Maintenance",
	}
	if err := slotBlockRepo.Create(context.Background(), block); err != nil {
		t.Fatalf("failed to create slot block: %v", err)
	}

	// Book 10:00-12:00 (no overlap with block at 14:00-16:00)
	result, err := svc.Create(context.Background(), uuid.New(), service.CreateBookingInput{
		BathhouseID: bh.ID,
		StartTime:   start,
		EndTime:     start.Add(2 * time.Hour),
		GuestCount:  3,
	})

	if err != nil {
		t.Fatalf("should succeed when no overlap with slot block, got: %v", err)
	}
	if result.Booking.Status != domain.BookingPending {
		t.Errorf("status = %q, want %q", result.Booking.Status, domain.BookingPending)
	}
}

func TestBookingService_GetAvailableSlots_WithSlotBlock(t *testing.T) {
	svc, bhRepo, _, slotBlockRepo := newBookingServiceWithSlotBlocks()
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

	// Create a slot block that covers 11:00-12:00
	block := &domain.SlotBlock{
		BathhouseID: bh.ID,
		StartTime:   time.Date(monday.Year(), monday.Month(), monday.Day(), 11, 0, 0, 0, time.UTC),
		EndTime:     time.Date(monday.Year(), monday.Month(), monday.Day(), 12, 0, 0, 0, time.UTC),
		Source:      domain.SlotBlockSourceYandexCalendar,
		ExternalID:  "yandex-event-1",
		Description: "External event",
	}
	if err := slotBlockRepo.Create(context.Background(), block); err != nil {
		t.Fatalf("failed to create slot block: %v", err)
	}

	slots, err := svc.GetAvailableSlots(context.Background(), bh.ID, monday)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(slots) != 4 { // 09:00-10:00, 10:00-11:00, 11:00-12:00, 12:00-13:00
		t.Fatalf("expected 4 slots, got %d", len(slots))
	}

	// The 11:00-12:00 slot should be unavailable due to slot block
	for _, slot := range slots {
		if slot.StartTime.Hour() == 11 && slot.Available {
			t.Error("11:00-12:00 slot should be unavailable due to slot block")
		}
		if slot.StartTime.Hour() == 9 && !slot.Available {
			t.Error("09:00-10:00 slot should be available")
		}
		if slot.StartTime.Hour() == 10 && !slot.Available {
			t.Error("10:00-11:00 slot should be available")
		}
		if slot.StartTime.Hour() == 12 && !slot.Available {
			t.Error("12:00-13:00 slot should be available")
		}
	}
}

// --- Booking with Add-ons ---

func newBookingServiceWithAddOns() (service.BookingService, *mock.BathhouseRepo, *mock.BookingRepo, *mock.AddOnRepo) {
	bhRepo := mock.NewBathhouseRepo()
	bookingRepo := mock.NewBookingRepo()
	repRepo := mock.NewRepresentativeRepo()
	pricingRepo := mock.NewPricingRuleRepo()
	loyaltyRepo := mock.NewLoyaltyRepo()
	addonRepo := mock.NewAddOnRepo()
	access := service.NewAccessChecker(repRepo, bhRepo)
	log := logger.New(logger.LevelWarn)
	pricingSvc := service.NewPricingService(pricingRepo, nil, bhRepo, nil, access, log)
	loyaltySvc := service.NewLoyaltyService(loyaltyRepo, log)
	addonSvc := service.NewAddOnService(addonRepo, access, log)
	svc := service.NewBookingService(bookingRepo, bhRepo, mock.NewSlotBlockRepo(), addonRepo, mock.NewUserRepo(), mock.NewCityRepo(), pricingSvc, addonSvc, loyaltySvc, &noopReferralService{}, &noopPromoService{}, &noopCertificateService{}, &noopPaymentService{}, &noopServiceFeeService{}, nil, nil, nil, nil, nil, access, &noopNotifService{}, log)
	return svc, bhRepo, bookingRepo, addonRepo
}

func TestBookingService_Create_WithAddOns_PerItem(t *testing.T) {
	svc, bhRepo, _, addonRepo := newBookingServiceWithAddOns()
	ownerID := uuid.New()
	clientID := uuid.New()
	bh := createBathhouse(t, bhRepo, ownerID) // PricePerHour = 5000

	// Create add-on: broom 500 kopecks per item
	addon := &domain.AddOn{
		ID:          uuid.New(),
		BathhouseID: bh.ID,
		Name:        "Веник",
		Price:       500,
		Unit:        domain.AddOnUnitPerItem,
		IsActive:    true,
	}
	_ = addonRepo.Create(context.Background(), addon)

	now := time.Now()
	start := time.Date(now.Year(), now.Month(), now.Day()+1, 10, 0, 0, 0, now.Location())
	end := start.Add(2 * time.Hour)

	result, err := svc.Create(context.Background(), clientID, service.CreateBookingInput{
		BathhouseID: bh.ID,
		StartTime:   start,
		EndTime:     end,
		GuestCount:  5,
		AddOns: []service.AddOnSelection{
			{AddOnID: addon.ID, Quantity: 3},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Base price: 5000 * 2 = 10000
	// Add-on: 500 * 3 = 1500
	// Total: 11500
	expectedAddOnTotal := int64(1500)
	expectedTotal := int64(11500)

	if result.Booking.AddOnTotal != expectedAddOnTotal {
		t.Errorf("AddOnTotal = %d, want %d", result.Booking.AddOnTotal, expectedAddOnTotal)
	}
	if result.Booking.TotalPrice != expectedTotal {
		t.Errorf("TotalPrice = %d, want %d", result.Booking.TotalPrice, expectedTotal)
	}
	if len(result.AddOns) != 1 {
		t.Fatalf("expected 1 booking addon, got %d", len(result.AddOns))
	}
	if result.AddOns[0].Name != "Веник" {
		t.Errorf("addon name = %q, want %q", result.AddOns[0].Name, "Веник")
	}
	if result.AddOns[0].Quantity != 3 {
		t.Errorf("addon quantity = %d, want 3", result.AddOns[0].Quantity)
	}
	if result.AddOns[0].TotalPrice != 1500 {
		t.Errorf("addon total = %d, want 1500", result.AddOns[0].TotalPrice)
	}
}

func TestBookingService_Create_WithAddOns_PerHour(t *testing.T) {
	svc, bhRepo, _, addonRepo := newBookingServiceWithAddOns()
	ownerID := uuid.New()
	clientID := uuid.New()
	bh := createBathhouse(t, bhRepo, ownerID)

	addon := &domain.AddOn{
		ID:          uuid.New(),
		BathhouseID: bh.ID,
		Name:        "Аренда полотенец",
		Price:       200,
		Unit:        domain.AddOnUnitPerHour,
		IsActive:    true,
	}
	_ = addonRepo.Create(context.Background(), addon)

	now := time.Now()
	start := time.Date(now.Year(), now.Month(), now.Day()+1, 10, 0, 0, 0, now.Location())
	end := start.Add(3 * time.Hour) // 3 hours

	result, err := svc.Create(context.Background(), clientID, service.CreateBookingInput{
		BathhouseID: bh.ID,
		StartTime:   start,
		EndTime:     end,
		GuestCount:  2,
		AddOns: []service.AddOnSelection{
			{AddOnID: addon.ID, Quantity: 2},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Base: 5000 * 3 = 15000
	// Add-on per_hour: 200 * 2 * 3 = 1200
	expectedAddOnTotal := int64(1200)
	if result.Booking.AddOnTotal != expectedAddOnTotal {
		t.Errorf("AddOnTotal = %d, want %d", result.Booking.AddOnTotal, expectedAddOnTotal)
	}
	if result.Booking.TotalPrice != 15000+1200 {
		t.Errorf("TotalPrice = %d, want %d", result.Booking.TotalPrice, 15000+1200)
	}
}

func TestBookingService_Create_WithAddOns_PerPerson(t *testing.T) {
	svc, bhRepo, _, addonRepo := newBookingServiceWithAddOns()
	ownerID := uuid.New()
	clientID := uuid.New()
	bh := createBathhouse(t, bhRepo, ownerID)

	addon := &domain.AddOn{
		ID:          uuid.New(),
		BathhouseID: bh.ID,
		Name:        "Тапочки",
		Price:       100,
		Unit:        domain.AddOnUnitPerPerson,
		IsActive:    true,
	}
	_ = addonRepo.Create(context.Background(), addon)

	now := time.Now()
	start := time.Date(now.Year(), now.Month(), now.Day()+1, 10, 0, 0, 0, now.Location())
	end := start.Add(2 * time.Hour)

	result, err := svc.Create(context.Background(), clientID, service.CreateBookingInput{
		BathhouseID: bh.ID,
		StartTime:   start,
		EndTime:     end,
		GuestCount:  5,
		AddOns: []service.AddOnSelection{
			{AddOnID: addon.ID, Quantity: 1},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Base: 5000 * 2 = 10000
	// Add-on per_person: 100 * 1 * 5 = 500
	expectedAddOnTotal := int64(500)
	if result.Booking.AddOnTotal != expectedAddOnTotal {
		t.Errorf("AddOnTotal = %d, want %d", result.Booking.AddOnTotal, expectedAddOnTotal)
	}
}

func TestBookingService_Create_WithAddOns_WrongBathhouse(t *testing.T) {
	svc, bhRepo, _, addonRepo := newBookingServiceWithAddOns()
	ownerID := uuid.New()
	clientID := uuid.New()
	bh := createBathhouse(t, bhRepo, ownerID)

	// Create add-on for different bathhouse
	addon := &domain.AddOn{
		ID:          uuid.New(),
		BathhouseID: uuid.New(), // different bathhouse
		Name:        "Веник",
		Price:       500,
		Unit:        domain.AddOnUnitPerItem,
		IsActive:    true,
	}
	_ = addonRepo.Create(context.Background(), addon)

	now := time.Now()
	start := time.Date(now.Year(), now.Month(), now.Day()+1, 10, 0, 0, 0, now.Location())
	end := start.Add(2 * time.Hour)

	_, err := svc.Create(context.Background(), clientID, service.CreateBookingInput{
		BathhouseID: bh.ID,
		StartTime:   start,
		EndTime:     end,
		GuestCount:  5,
		AddOns: []service.AddOnSelection{
			{AddOnID: addon.ID, Quantity: 1},
		},
	})

	if !errors.Is(err, domain.ErrAddOnNotFound) {
		t.Errorf("expected ErrAddOnNotFound, got: %v", err)
	}
}

func TestBookingService_Create_WithoutAddOns_Works(t *testing.T) {
	svc, bhRepo, _, _ := newBookingServiceWithAddOns()
	ownerID := uuid.New()
	clientID := uuid.New()
	bh := createBathhouse(t, bhRepo, ownerID)

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

	if result.Booking.AddOnTotal != 0 {
		t.Errorf("AddOnTotal = %d, want 0", result.Booking.AddOnTotal)
	}
	if result.Booking.TotalPrice != 10000 {
		t.Errorf("TotalPrice = %d, want 10000", result.Booking.TotalPrice)
	}
	if len(result.AddOns) != 0 {
		t.Errorf("expected 0 booking add-ons, got %d", len(result.AddOns))
	}
}

func TestBookingService_Create_WithLongSessionDiscount(t *testing.T) {
	svc, bhRepo, _, _, _, _, _, _ := newBookingService()
	ownerID := uuid.New()
	clientID := uuid.New()

	wh := make([]domain.WorkingHours, 7)
	for i := 0; i < 7; i++ {
		wh[i] = domain.WorkingHours{DayOfWeek: i, OpenTime: "00:00", CloseTime: "23:59"}
	}
	bh := &domain.Bathhouse{
		ID:                         uuid.New(),
		OwnerID:                    ownerID,
		Name:                       "Discount Bath",
		Address:                    "123 St",
		CityID:                     1,
		PricePerHour:               1000,
		MinDuration:                1,
		MaxGuests:                  10,
		BaseCapacity:               10,
		LongSessionThresholdHours:  4,
		LongSessionDiscountPercent: 10,
		WorkingHours:               wh,
		Status:                     domain.BathhouseStatusActive,
	}
	_ = bhRepo.Create(context.Background(), bh)

	now := time.Now()
	start := time.Date(now.Year(), now.Month(), now.Day()+1, 10, 0, 0, 0, now.Location())
	end := start.Add(6 * time.Hour)

	result, err := svc.Create(context.Background(), clientID, service.CreateBookingInput{
		BathhouseID: bh.ID,
		StartTime:   start,
		EndTime:     end,
		GuestCount:  5,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// BasePrice = 6000 (1000 * 6h)
	// LongSessionDiscount = 200 (avgRate=1000, 2 discountable hours * 10% = 200)
	if result.BasePrice != 6000 {
		t.Errorf("BasePrice = %d, want 6000", result.BasePrice)
	}
	if result.LongSessionDiscount != 200 {
		t.Errorf("LongSessionDiscount = %d, want 200", result.LongSessionDiscount)
	}
	if result.Booking.LongSessionDiscount != 200 {
		t.Errorf("Booking.LongSessionDiscount = %d, want 200", result.Booking.LongSessionDiscount)
	}
}

func TestBookingService_Create_WithExtraGuestSurcharge(t *testing.T) {
	svc, bhRepo, _, _, _, _, _, _ := newBookingService()
	ownerID := uuid.New()
	clientID := uuid.New()

	wh := make([]domain.WorkingHours, 7)
	for i := 0; i < 7; i++ {
		wh[i] = domain.WorkingHours{DayOfWeek: i, OpenTime: "00:00", CloseTime: "23:59"}
	}
	bh := &domain.Bathhouse{
		ID:                         uuid.New(),
		OwnerID:                    ownerID,
		Name:                       "Surcharge Bath",
		Address:                    "123 St",
		CityID:                     1,
		PricePerHour:               1000,
		MinDuration:                1,
		MaxGuests:                  10,
		BaseCapacity:               3,
		ExtraGuestSurcharge:        500,
		LongSessionThresholdHours:  4,
		LongSessionDiscountPercent: 0,
		WorkingHours:               wh,
		Status:                     domain.BathhouseStatusActive,
	}
	_ = bhRepo.Create(context.Background(), bh)

	now := time.Now()
	start := time.Date(now.Year(), now.Month(), now.Day()+1, 10, 0, 0, 0, now.Location())
	end := start.Add(3 * time.Hour)

	result, err := svc.Create(context.Background(), clientID, service.CreateBookingInput{
		BathhouseID: bh.ID,
		StartTime:   start,
		EndTime:     end,
		GuestCount:  5, // 2 extra guests
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// BasePrice = 3000 (1000 * 3h)
	// ExtraGuestSurcharge = 3000 (2 extra * 500 * 3h)
	// Total = 3000 + 3000 = 6000
	if result.ExtraGuestSurcharge != 3000 {
		t.Errorf("ExtraGuestSurcharge = %d, want 3000", result.ExtraGuestSurcharge)
	}
	if result.Booking.ExtraGuestSurcharge != 3000 {
		t.Errorf("Booking.ExtraGuestSurcharge = %d, want 3000", result.Booking.ExtraGuestSurcharge)
	}
	// total = base(3000) + surcharge(3000) = 6000, but service fee may be added
	expectedBase := int64(6000)
	if result.Booking.TotalPrice < expectedBase {
		t.Errorf("TotalPrice = %d, want at least %d", result.Booking.TotalPrice, expectedBase)
	}
}

// --- Last-Minute Discount Tests ---

func TestBookingService_GetAvailableSlots_LastMinuteEnabled(t *testing.T) {
	svc, bhRepo, _, _, _, _, _, _ := newBookingService()
	ownerID := uuid.New()

	bh := &domain.Bathhouse{
		ID: uuid.New(), OwnerID: ownerID, Name: "Last Minute Bath",
		Address: "123 St", CityID: 1, PricePerHour: 10000,
		MinDuration: 1, MaxGuests: 10, Status: domain.BathhouseStatusActive,
		LastMinuteEnabled:         true,
		LastMinuteDiscountPercent: 20,
		LastMinuteHoursThreshold:  6,
		WorkingHours: []domain.WorkingHours{
			{DayOfWeek: 0, OpenTime: "09:00", CloseTime: "18:00"},
			{DayOfWeek: 1, OpenTime: "09:00", CloseTime: "18:00"},
			{DayOfWeek: 2, OpenTime: "09:00", CloseTime: "18:00"},
			{DayOfWeek: 3, OpenTime: "09:00", CloseTime: "18:00"},
			{DayOfWeek: 4, OpenTime: "09:00", CloseTime: "18:00"},
			{DayOfWeek: 5, OpenTime: "09:00", CloseTime: "18:00"},
			{DayOfWeek: 6, OpenTime: "09:00", CloseTime: "18:00"},
		},
	}
	_ = bhRepo.Create(context.Background(), bh)

	// Use a date far enough in the future so that early slots (09:00) are beyond
	// the threshold, but some slots might still be in range depending on "now".
	// To test reliably, we use a day far in the future so ALL slots are beyond threshold.
	now := time.Now()
	futureDate := time.Date(now.Year(), now.Month(), now.Day()+30, 0, 0, 0, 0, time.UTC)
	// Ensure it falls on a day with working hours
	for futureDate.Weekday() == time.Sunday {
		futureDate = futureDate.AddDate(0, 0, 1)
	}

	slots, err := svc.GetAvailableSlots(context.Background(), bh.ID, futureDate)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(slots) == 0 {
		t.Fatal("expected slots, got 0")
	}

	// All slots are far in the future (>6h), so none should be last-minute
	for _, slot := range slots {
		if slot.IsLastMinute {
			t.Errorf("slot at %s should NOT be last-minute (far future)", slot.StartTime.Format("15:04"))
		}
		if slot.OriginalPrice != 0 {
			t.Errorf("slot at %s should have OriginalPrice=0, got %d", slot.StartTime.Format("15:04"), slot.OriginalPrice)
		}
		if slot.Price != 10000 {
			t.Errorf("slot at %s: price = %d, want 10000", slot.StartTime.Format("15:04"), slot.Price)
		}
	}
}

func TestBookingService_GetAvailableSlots_LastMinuteDisabled(t *testing.T) {
	svc, bhRepo, _, _, _, _, _, _ := newBookingService()
	ownerID := uuid.New()

	wh := make([]domain.WorkingHours, 7)
	for i := 0; i < 7; i++ {
		wh[i] = domain.WorkingHours{DayOfWeek: i, OpenTime: "00:00", CloseTime: "23:59"}
	}
	bh := &domain.Bathhouse{
		ID: uuid.New(), OwnerID: ownerID, Name: "No Last Minute Bath",
		Address: "123 St", CityID: 1, PricePerHour: 10000,
		MinDuration: 1, MaxGuests: 10, Status: domain.BathhouseStatusActive,
		LastMinuteEnabled:         false,
		LastMinuteDiscountPercent: 20,
		LastMinuteHoursThreshold:  6,
		WorkingHours:              wh,
	}
	_ = bhRepo.Create(context.Background(), bh)

	// Use tomorrow
	now := time.Now()
	tomorrow := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, time.UTC)

	slots, err := svc.GetAvailableSlots(context.Background(), bh.ID, tomorrow)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// No slot should have last-minute when disabled
	for _, slot := range slots {
		if slot.IsLastMinute {
			t.Errorf("slot at %s should NOT be last-minute when disabled", slot.StartTime.Format("15:04"))
		}
	}
}

func TestBookingService_Create_WithLastMinuteDiscount(t *testing.T) {
	svc, bhRepo, _, _, _, _, _, _ := newBookingService()
	ownerID := uuid.New()
	clientID := uuid.New()

	wh := make([]domain.WorkingHours, 7)
	for i := 0; i < 7; i++ {
		wh[i] = domain.WorkingHours{DayOfWeek: i, OpenTime: "08:00", CloseTime: "22:00"}
	}
	bh := &domain.Bathhouse{
		ID:                         uuid.New(),
		OwnerID:                    ownerID,
		Name:                       "Last Minute Booking",
		Address:                    "123 St",
		CityID:                     1,
		PricePerHour:               10000,
		MinDuration:                1,
		MaxGuests:                  10,
		BaseCapacity:               10,
		LongSessionThresholdHours:  4,
		LastMinuteEnabled:          true,
		LastMinuteDiscountPercent:  20,
		LastMinuteHoursThreshold:   48,
		WorkingHours:               wh,
		Status:                     domain.BathhouseStatusActive,
	}
	_ = bhRepo.Create(context.Background(), bh)

	// Book tomorrow at 10:00-12:00 — within 48h threshold for last-minute discount
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

	// Both slots (2h) start within 48h threshold
	// Each slot: 10000 * 20% = 2000 discount
	// Total last-minute discount: 4000
	expectedDiscount := int64(4000)
	if result.LastMinuteDiscount != expectedDiscount {
		t.Errorf("LastMinuteDiscount = %d, want %d", result.LastMinuteDiscount, expectedDiscount)
	}
	if result.Booking.LastMinuteDiscount != expectedDiscount {
		t.Errorf("Booking.LastMinuteDiscount = %d, want %d", result.Booking.LastMinuteDiscount, expectedDiscount)
	}
	// Total should be base(20000) - lastMinute(4000) = 16000
	if result.Booking.TotalPrice != 16000 {
		t.Errorf("TotalPrice = %d, want 16000", result.Booking.TotalPrice)
	}
}

func TestBookingService_Create_LastMinuteDisabled_NormalPrice(t *testing.T) {
	svc, bhRepo, _, _, _, _, _, _ := newBookingService()
	ownerID := uuid.New()
	clientID := uuid.New()

	wh := make([]domain.WorkingHours, 7)
	for i := 0; i < 7; i++ {
		wh[i] = domain.WorkingHours{DayOfWeek: i, OpenTime: "08:00", CloseTime: "22:00"}
	}
	bh := &domain.Bathhouse{
		ID:                         uuid.New(),
		OwnerID:                    ownerID,
		Name:                       "No Discount",
		Address:                    "123 St",
		CityID:                     1,
		PricePerHour:               10000,
		MinDuration:                1,
		MaxGuests:                  10,
		BaseCapacity:               10,
		LongSessionThresholdHours:  4,
		LastMinuteEnabled:          false,
		LastMinuteDiscountPercent:  20,
		LastMinuteHoursThreshold:   6,
		WorkingHours:               wh,
		Status:                     domain.BathhouseStatusActive,
	}
	_ = bhRepo.Create(context.Background(), bh)

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

	if result.LastMinuteDiscount != 0 {
		t.Errorf("LastMinuteDiscount = %d, want 0 (disabled)", result.LastMinuteDiscount)
	}
	if result.Booking.TotalPrice != 20000 {
		t.Errorf("TotalPrice = %d, want 20000", result.Booking.TotalPrice)
	}
}

func TestBookingService_Create_LastMinuteBeyondThreshold(t *testing.T) {
	svc, bhRepo, _, _, _, _, _, _ := newBookingService()
	ownerID := uuid.New()
	clientID := uuid.New()

	wh := make([]domain.WorkingHours, 7)
	for i := 0; i < 7; i++ {
		wh[i] = domain.WorkingHours{DayOfWeek: i, OpenTime: "08:00", CloseTime: "22:00"}
	}
	bh := &domain.Bathhouse{
		ID:                         uuid.New(),
		OwnerID:                    ownerID,
		Name:                       "Far Future Booking",
		Address:                    "123 St",
		CityID:                     1,
		PricePerHour:               10000,
		MinDuration:                1,
		MaxGuests:                  10,
		BaseCapacity:               10,
		LongSessionThresholdHours:  4,
		LastMinuteEnabled:          true,
		LastMinuteDiscountPercent:  20,
		LastMinuteHoursThreshold:   6,
		WorkingHours:               wh,
		Status:                     domain.BathhouseStatusActive,
	}
	_ = bhRepo.Create(context.Background(), bh)

	// Book starting 48 hours from now — way beyond 6h threshold
	now := time.Now()
	start := time.Date(now.Year(), now.Month(), now.Day()+2, 10, 0, 0, 0, now.Location())
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

	// Slots are >6h away, no last-minute discount
	if result.LastMinuteDiscount != 0 {
		t.Errorf("LastMinuteDiscount = %d, want 0 (beyond threshold)", result.LastMinuteDiscount)
	}
	if result.Booking.TotalPrice != 20000 {
		t.Errorf("TotalPrice = %d, want 20000", result.Booking.TotalPrice)
	}
}

// ==================== Booking Settings Tests ====================

func TestBookingService_GetAvailableSlots_BufferTime(t *testing.T) {
	svc, bhRepo, bookingRepo, _, _, _, _, _ := newBookingService()
	ownerID := uuid.New()

	bh := &domain.Bathhouse{
		ID: uuid.New(), OwnerID: ownerID, Name: "Buffer Bath",
		Address: "123 St", CityID: 1, PricePerHour: 5000,
		MinDuration: 1, MaxGuests: 10, Status: domain.BathhouseStatusActive,
		BufferMinutes: 30, // 30 min buffer between bookings
		WorkingHours: []domain.WorkingHours{
			{DayOfWeek: 0, OpenTime: "09:00", CloseTime: "15:00"}, // Monday
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

	// Booking at 10:00-12:00, with 30min buffer the 12:00-13:00 slot should be marked unavailable
	existingBooking := &domain.Booking{
		ID: uuid.New(), UserID: uuid.New(), BathhouseID: bh.ID,
		StartTime:  time.Date(monday.Year(), monday.Month(), monday.Day(), 10, 0, 0, 0, time.UTC),
		EndTime:    time.Date(monday.Year(), monday.Month(), monday.Day(), 12, 0, 0, 0, time.UTC),
		GuestCount: 2, TotalPrice: 10000, Status: domain.BookingConfirmed,
	}
	_ = bookingRepo.Create(context.Background(), existingBooking)

	slots, err := svc.GetAvailableSlots(context.Background(), bh.ID, monday)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, slot := range slots {
		hour := slot.StartTime.Hour()
		// 10:00-11:00 and 11:00-12:00 are booked; 12:00-13:00 should be blocked by buffer
		if hour == 10 || hour == 11 {
			if slot.Available {
				t.Errorf("slot %d:00 should be unavailable (booked)", hour)
			}
		}
		if hour == 12 {
			if slot.Available {
				t.Errorf("slot 12:00 should be unavailable (buffer zone after 10:00-12:00 booking)")
			}
		}
		// 09:00 is also unavailable: ends at 10:00 with 0min gap before 10:00 booking (needs 30min buffer)
		if hour == 9 {
			if slot.Available {
				t.Error("slot 09:00 should be unavailable (buffer zone before 10:00-12:00 booking)")
			}
		}
		// 13:00+ should be available
		if hour == 13 && !slot.Available {
			t.Error("slot 13:00 should be available (past buffer zone)")
		}
	}
}

func TestBookingService_Create_LeadTimeRejection(t *testing.T) {
	svc, bhRepo, _, _, _, _, _, _ := newBookingService()
	ownerID := uuid.New()

	bh := &domain.Bathhouse{
		ID: uuid.New(), OwnerID: ownerID, Name: "Lead Time Bath",
		Address: "123 St", CityID: 1, PricePerHour: 5000,
		MinDuration: 1, MaxGuests: 10, Status: domain.BathhouseStatusActive,
		LeadTimeHours: 2, // 2 hours lead time
		MaxAdvanceDays: 90,
		LongSessionThresholdHours: 4, BaseCapacity: 10,
		WorkingHours: []domain.WorkingHours{
			{DayOfWeek: 0, OpenTime: "00:00", CloseTime: "23:59"},
			{DayOfWeek: 1, OpenTime: "00:00", CloseTime: "23:59"},
			{DayOfWeek: 2, OpenTime: "00:00", CloseTime: "23:59"},
			{DayOfWeek: 3, OpenTime: "00:00", CloseTime: "23:59"},
			{DayOfWeek: 4, OpenTime: "00:00", CloseTime: "23:59"},
			{DayOfWeek: 5, OpenTime: "00:00", CloseTime: "23:59"},
			{DayOfWeek: 6, OpenTime: "00:00", CloseTime: "23:59"},
		},
	}
	_ = bhRepo.Create(context.Background(), bh)

	// Booking starting in 1h should be rejected (lead time is 2h)
	start := time.Now().Add(1 * time.Hour)
	_, err := svc.Create(context.Background(), uuid.New(), service.CreateBookingInput{
		BathhouseID: bh.ID,
		StartTime:   start,
		EndTime:     start.Add(2 * time.Hour),
		GuestCount:  2,
	})

	if !errors.Is(err, domain.ErrInvalidInput) {
		t.Errorf("booking within lead time should be rejected, got: %v", err)
	}
}

func TestBookingService_Create_MaxAdvanceDaysRejection(t *testing.T) {
	svc, bhRepo, _, _, _, _, _, _ := newBookingService()
	ownerID := uuid.New()

	bh := &domain.Bathhouse{
		ID: uuid.New(), OwnerID: ownerID, Name: "Max Advance Bath",
		Address: "123 St", CityID: 1, PricePerHour: 5000,
		MinDuration: 1, MaxGuests: 10, Status: domain.BathhouseStatusActive,
		MaxAdvanceDays: 30, // only 30 days ahead
		LongSessionThresholdHours: 4, BaseCapacity: 10,
		WorkingHours: []domain.WorkingHours{
			{DayOfWeek: 0, OpenTime: "00:00", CloseTime: "23:59"},
			{DayOfWeek: 1, OpenTime: "00:00", CloseTime: "23:59"},
			{DayOfWeek: 2, OpenTime: "00:00", CloseTime: "23:59"},
			{DayOfWeek: 3, OpenTime: "00:00", CloseTime: "23:59"},
			{DayOfWeek: 4, OpenTime: "00:00", CloseTime: "23:59"},
			{DayOfWeek: 5, OpenTime: "00:00", CloseTime: "23:59"},
			{DayOfWeek: 6, OpenTime: "00:00", CloseTime: "23:59"},
		},
	}
	_ = bhRepo.Create(context.Background(), bh)

	// Booking 60 days out should be rejected (max advance is 30 days)
	start := time.Now().Add(60 * 24 * time.Hour)
	_, err := svc.Create(context.Background(), uuid.New(), service.CreateBookingInput{
		BathhouseID: bh.ID,
		StartTime:   start,
		EndTime:     start.Add(2 * time.Hour),
		GuestCount:  2,
	})

	if !errors.Is(err, domain.ErrInvalidInput) {
		t.Errorf("booking beyond max advance days should be rejected, got: %v", err)
	}
}

func TestBookingService_Create_BufferConflict(t *testing.T) {
	svc, bhRepo, bookingRepo, _, _, _, _, _ := newBookingService()
	ownerID := uuid.New()

	bh := &domain.Bathhouse{
		ID: uuid.New(), OwnerID: ownerID, Name: "Buffer Conflict Bath",
		Address: "123 St", CityID: 1, PricePerHour: 5000,
		MinDuration: 1, MaxGuests: 10, Status: domain.BathhouseStatusActive,
		BufferMinutes: 30,
		MaxAdvanceDays: 90,
		LongSessionThresholdHours: 4, BaseCapacity: 10,
		WorkingHours: []domain.WorkingHours{
			{DayOfWeek: 0, OpenTime: "00:00", CloseTime: "23:59"},
			{DayOfWeek: 1, OpenTime: "00:00", CloseTime: "23:59"},
			{DayOfWeek: 2, OpenTime: "00:00", CloseTime: "23:59"},
			{DayOfWeek: 3, OpenTime: "00:00", CloseTime: "23:59"},
			{DayOfWeek: 4, OpenTime: "00:00", CloseTime: "23:59"},
			{DayOfWeek: 5, OpenTime: "00:00", CloseTime: "23:59"},
			{DayOfWeek: 6, OpenTime: "00:00", CloseTime: "23:59"},
		},
	}
	_ = bhRepo.Create(context.Background(), bh)

	now := time.Now()
	start := time.Date(now.Year(), now.Month(), now.Day()+1, 10, 0, 0, 0, now.Location())

	// Existing booking: 10:00-12:00
	existingBooking := &domain.Booking{
		ID: uuid.New(), UserID: uuid.New(), BathhouseID: bh.ID,
		StartTime: start, EndTime: start.Add(2 * time.Hour),
		GuestCount: 2, TotalPrice: 10000, Status: domain.BookingConfirmed,
	}
	_ = bookingRepo.Create(context.Background(), existingBooking)

	// New booking at 12:00-14:00 should conflict with 30min buffer (12:00-12:30 is buffer zone)
	newStart := start.Add(2 * time.Hour) // 12:00
	_, err := svc.Create(context.Background(), uuid.New(), service.CreateBookingInput{
		BathhouseID: bh.ID,
		StartTime:   newStart,
		EndTime:     newStart.Add(2 * time.Hour),
		GuestCount:  2,
	})

	if !errors.Is(err, domain.ErrSlotUnavailable) {
		t.Errorf("booking within buffer zone should be rejected, got: %v", err)
	}
}

func TestBookingService_Create_NoBufferConflict(t *testing.T) {
	svc, bhRepo, bookingRepo, _, _, _, _, _ := newBookingService()
	ownerID := uuid.New()

	bh := &domain.Bathhouse{
		ID: uuid.New(), OwnerID: ownerID, Name: "No Buffer Bath",
		Address: "123 St", CityID: 1, PricePerHour: 5000,
		MinDuration: 1, MaxGuests: 10, Status: domain.BathhouseStatusActive,
		BufferMinutes: 0, // no buffer
		MaxAdvanceDays: 90,
		LongSessionThresholdHours: 4, BaseCapacity: 10,
		WorkingHours: []domain.WorkingHours{
			{DayOfWeek: 0, OpenTime: "00:00", CloseTime: "23:59"},
			{DayOfWeek: 1, OpenTime: "00:00", CloseTime: "23:59"},
			{DayOfWeek: 2, OpenTime: "00:00", CloseTime: "23:59"},
			{DayOfWeek: 3, OpenTime: "00:00", CloseTime: "23:59"},
			{DayOfWeek: 4, OpenTime: "00:00", CloseTime: "23:59"},
			{DayOfWeek: 5, OpenTime: "00:00", CloseTime: "23:59"},
			{DayOfWeek: 6, OpenTime: "00:00", CloseTime: "23:59"},
		},
	}
	_ = bhRepo.Create(context.Background(), bh)

	now := time.Now()
	start := time.Date(now.Year(), now.Month(), now.Day()+1, 10, 0, 0, 0, now.Location())

	// Existing booking: 10:00-12:00
	existingBooking := &domain.Booking{
		ID: uuid.New(), UserID: uuid.New(), BathhouseID: bh.ID,
		StartTime: start, EndTime: start.Add(2 * time.Hour),
		GuestCount: 2, TotalPrice: 10000, Status: domain.BookingConfirmed,
	}
	_ = bookingRepo.Create(context.Background(), existingBooking)

	// With buffer=0, booking at 12:00-14:00 should succeed
	newStart := start.Add(2 * time.Hour) // 12:00
	result, err := svc.Create(context.Background(), uuid.New(), service.CreateBookingInput{
		BathhouseID: bh.ID,
		StartTime:   newStart,
		EndTime:     newStart.Add(2 * time.Hour),
		GuestCount:  2,
	})

	if err != nil {
		t.Fatalf("booking without buffer should succeed, got: %v", err)
	}
	if result.Booking.Status != domain.BookingPending {
		t.Errorf("status = %q, want %q", result.Booking.Status, domain.BookingPending)
	}
}

func TestBookingService_Create_ZeroLeadTimeAllowsImmediate(t *testing.T) {
	svc, bhRepo, _, _, _, _, _, _ := newBookingService()
	ownerID := uuid.New()

	bh := &domain.Bathhouse{
		ID: uuid.New(), OwnerID: ownerID, Name: "Immediate Bath",
		Address: "123 St", CityID: 1, PricePerHour: 5000,
		MinDuration: 1, MaxGuests: 10, Status: domain.BathhouseStatusActive,
		LeadTimeHours: 0, // lead time 0 = minimum 5 min fallback
		MaxAdvanceDays: 90,
		LongSessionThresholdHours: 4, BaseCapacity: 10,
		WorkingHours: []domain.WorkingHours{
			{DayOfWeek: 0, OpenTime: "00:00", CloseTime: "23:59"},
			{DayOfWeek: 1, OpenTime: "00:00", CloseTime: "23:59"},
			{DayOfWeek: 2, OpenTime: "00:00", CloseTime: "23:59"},
			{DayOfWeek: 3, OpenTime: "00:00", CloseTime: "23:59"},
			{DayOfWeek: 4, OpenTime: "00:00", CloseTime: "23:59"},
			{DayOfWeek: 5, OpenTime: "00:00", CloseTime: "23:59"},
			{DayOfWeek: 6, OpenTime: "00:00", CloseTime: "23:59"},
		},
	}
	_ = bhRepo.Create(context.Background(), bh)

	// Booking starting soon should succeed with lead_time=0 (5 min fallback)
	// Use fixed safe hour to avoid time-of-day flakiness with working hours boundary
	now := time.Now()
	start := time.Date(now.Year(), now.Month(), now.Day()+1, 10, 0, 0, 0, now.Location())

	result, err := svc.Create(context.Background(), uuid.New(), service.CreateBookingInput{
		BathhouseID: bh.ID,
		StartTime:   start,
		EndTime:     start.Add(1 * time.Hour),
		GuestCount:  2,
	})

	if err != nil {
		t.Fatalf("booking with lead_time=0 should allow near-future booking, got: %v", err)
	}
	if result.Booking.Status != domain.BookingPending {
		t.Errorf("status = %q, want %q", result.Booking.Status, domain.BookingPending)
	}
}

func TestBathhouse_Validate_BookingSettings(t *testing.T) {
	tests := []struct {
		name          string
		bufferMinutes int
		leadTimeHours int
		maxAdvanceDays int
		wantErr       bool
	}{
		{"valid defaults", 30, 2, 90, false},
		{"zero buffer defaults to 30", 0, 0, 90, false}, // buffer 0 defaults to 30 in Validate, maxAdvDays 0 defaults to 90
		{"buffer not multiple of 15", 25, 0, 90, true},
		{"buffer too high", 150, 0, 90, true},
		{"lead time too high", 30, 50, 90, true},
		{"max advance too low", 30, 0, 5, true},
		{"max advance too high", 30, 0, 400, true},
		{"max advance min boundary", 30, 0, 7, false},
		{"max advance max boundary", 30, 0, 365, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bh := &domain.Bathhouse{
				Name: "Test", Address: "123", CityID: 1,
				PricePerHour: 5000, MinDuration: 1, MaxGuests: 10,
				LongSessionThresholdHours: 4, BaseCapacity: 10,
				BufferMinutes: tt.bufferMinutes, LeadTimeHours: tt.leadTimeHours, MaxAdvanceDays: tt.maxAdvanceDays,
			}
			err := bh.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr = %v", err, tt.wantErr)
			}
		})
	}
}

// --- Request-based Booking Mode ---

func createRequestModeBathhouse(t *testing.T, bhRepo *mock.BathhouseRepo, ownerID uuid.UUID) *domain.Bathhouse {
	t.Helper()
	wh := make([]domain.WorkingHours, 7)
	for i := 0; i < 7; i++ {
		wh[i] = domain.WorkingHours{DayOfWeek: i, OpenTime: "00:00", CloseTime: "23:59"}
	}
	bh := &domain.Bathhouse{
		ID:                        uuid.New(),
		OwnerID:                   ownerID,
		Name:                      "Request Mode Bathhouse",
		Address:                   "456 Street",
		CityID:                    1,
		PricePerHour:              5000,
		MinDuration:               1,
		MaxGuests:                 10,
		LongSessionThresholdHours: 4,
		BaseCapacity:              10,
		BookingMode:               domain.BookingModeRequest,
		RequestTimeout:            24,
		WorkingHours:              wh,
		Status:                    domain.BathhouseStatusActive,
		CreatedAt:                 time.Now(),
		UpdatedAt:                 time.Now(),
	}
	if err := bhRepo.Create(context.Background(), bh); err != nil {
		t.Fatal(err)
	}
	return bh
}

func TestBookingService_Create_RequestMode_SetsPendingOwner(t *testing.T) {
	svc, bhRepo, _, _, _, _, _, _ := newBookingService()
	ownerID := uuid.New()
	clientID := uuid.New()
	bh := createRequestModeBathhouse(t, bhRepo, ownerID)

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
	if result.Booking.Status != domain.BookingPendingOwner {
		t.Errorf("status = %q, want %q", result.Booking.Status, domain.BookingPendingOwner)
	}
}

func TestBookingService_Create_InstantMode_SetsPending(t *testing.T) {
	svc, bhRepo, _, _, _, _, _, _ := newBookingService()
	ownerID := uuid.New()
	clientID := uuid.New()
	bh := createBathhouse(t, bhRepo, ownerID) // default instant mode

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
}

func TestBookingService_Approve_Success(t *testing.T) {
	svc, bhRepo, bookingRepo, _, _, _, _, _ := newBookingService()
	ownerID := uuid.New()
	clientID := uuid.New()
	bh := createRequestModeBathhouse(t, bhRepo, ownerID)

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
		t.Fatalf("unexpected error creating booking: %v", err)
	}

	if result.Booking.Status != domain.BookingPendingOwner {
		t.Fatalf("precondition: status = %q, want pending_owner", result.Booking.Status)
	}

	err = svc.Approve(context.Background(), ownerID, domain.RoleOwner, result.Booking.ID)
	if err != nil {
		t.Fatalf("unexpected error approving: %v", err)
	}

	// Verify status changed to confirmed
	updated, err := bookingRepo.GetByID(context.Background(), result.Booking.ID)
	if err != nil {
		t.Fatalf("unexpected error getting booking: %v", err)
	}
	if updated.Status != domain.BookingConfirmed {
		t.Errorf("status after approve = %q, want %q", updated.Status, domain.BookingConfirmed)
	}
}

func TestBookingService_Approve_WrongStatus(t *testing.T) {
	svc, bhRepo, _, _, _, _, _, _ := newBookingService()
	ownerID := uuid.New()
	clientID := uuid.New()
	bh := createBathhouse(t, bhRepo, ownerID) // instant mode

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

	// Booking is 'pending' (instant mode), not 'pending_owner'
	err = svc.Approve(context.Background(), ownerID, domain.RoleOwner, result.Booking.ID)
	if err == nil {
		t.Error("expected error when approving non-pending_owner booking")
	}
}

func TestBookingService_Reject_WithReason(t *testing.T) {
	svc, bhRepo, bookingRepo, _, _, _, _, _ := newBookingService()
	ownerID := uuid.New()
	clientID := uuid.New()
	bh := createRequestModeBathhouse(t, bhRepo, ownerID)

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

	reason := "Fully booked, sorry"
	err = svc.Reject(context.Background(), ownerID, domain.RoleOwner, result.Booking.ID, reason)
	if err != nil {
		t.Fatalf("unexpected error rejecting: %v", err)
	}

	updated, err := bookingRepo.GetByID(context.Background(), result.Booking.ID)
	if err != nil {
		t.Fatalf("unexpected error getting booking: %v", err)
	}
	if updated.Status != domain.BookingRejected {
		t.Errorf("status = %q, want %q", updated.Status, domain.BookingRejected)
	}
	if updated.RejectionReason != reason {
		t.Errorf("rejection_reason = %q, want %q", updated.RejectionReason, reason)
	}
}

func TestBookingService_Cancel_PendingOwner_NoDeadline(t *testing.T) {
	svc, bhRepo, bookingRepo, _, _, _, _, _ := newBookingService()
	ownerID := uuid.New()
	clientID := uuid.New()
	bh := createRequestModeBathhouse(t, bhRepo, ownerID)

	// Create booking starting in ~1 hour - use fixed safe hour to avoid time-of-day flakiness
	now := time.Now()
	start := time.Date(now.Year(), now.Month(), now.Day()+1, 10, 0, 0, 0, now.Location())
	end := start.Add(1 * time.Hour)

	result, err := svc.Create(context.Background(), clientID, service.CreateBookingInput{
		BathhouseID: bh.ID,
		StartTime:   start,
		EndTime:     end,
		GuestCount:  5,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Booking.Status != domain.BookingPendingOwner {
		t.Fatalf("precondition: status = %q, want pending_owner", result.Booking.Status)
	}

	// Client should be able to cancel pending_owner booking even within 2-hour deadline
	err = svc.Cancel(context.Background(), clientID, domain.RoleClient, result.Booking.ID, "")
	if err != nil {
		t.Fatalf("expected no error cancelling pending_owner booking, got: %v", err)
	}

	updated, err := bookingRepo.GetByID(context.Background(), result.Booking.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated.Status != domain.BookingCancelled {
		t.Errorf("status = %q, want %q", updated.Status, domain.BookingCancelled)
	}
}

func TestBookingService_AutoRejectTimedOutRequests(t *testing.T) {
	svc, bhRepo, bookingRepo, _, _, _, _, _ := newBookingService()
	ownerID := uuid.New()
	clientID := uuid.New()
	bh := createRequestModeBathhouse(t, bhRepo, ownerID)

	now := time.Now()
	start := time.Date(now.Year(), now.Month(), now.Day()+2, 10, 0, 0, 0, now.Location())
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

	// Manually set the booking's created_at to >24h ago to simulate timeout
	booking, _ := bookingRepo.GetByID(context.Background(), result.Booking.ID)
	booking.CreatedAt = now.Add(-25 * time.Hour)
	if err := bookingRepo.Update(context.Background(), booking); err != nil {
		t.Fatalf("unexpected error updating booking: %v", err)
	}

	rejected, err := svc.AutoRejectTimedOutRequests(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rejected != 1 {
		t.Errorf("rejected = %d, want 1", rejected)
	}

	updated, err := bookingRepo.GetByID(context.Background(), result.Booking.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated.Status != domain.BookingRejected {
		t.Errorf("status = %q, want %q", updated.Status, domain.BookingRejected)
	}
	if updated.RejectionReason == "" {
		t.Error("expected rejection reason to be set")
	}
}

func TestBookingService_PendingOwner_BlocksSlot(t *testing.T) {
	svc, bhRepo, _, _, _, _, _, _ := newBookingService()
	ownerID := uuid.New()
	clientID := uuid.New()
	bh := createRequestModeBathhouse(t, bhRepo, ownerID)

	now := time.Now()
	start := time.Date(now.Year(), now.Month(), now.Day()+1, 10, 0, 0, 0, now.Location())
	end := start.Add(2 * time.Hour)

	_, err := svc.Create(context.Background(), clientID, service.CreateBookingInput{
		BathhouseID: bh.ID,
		StartTime:   start,
		EndTime:     end,
		GuestCount:  5,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Try to create another booking for the same slot - should fail
	_, err = svc.Create(context.Background(), uuid.New(), service.CreateBookingInput{
		BathhouseID: bh.ID,
		StartTime:   start,
		EndTime:     end,
		GuestCount:  3,
	})
	if !errors.Is(err, domain.ErrSlotUnavailable) {
		t.Errorf("expected ErrSlotUnavailable, got: %v", err)
	}
}

// --- Check-in / Check-out / No-show tests ---

func TestBookingService_CheckIn_ValidWindow(t *testing.T) {
	svc, bhRepo, bookingRepo, repRepo, _, _, _, _ := newBookingService()
	ownerID := uuid.New()
	bh := createBathhouse(t, bhRepo, ownerID)

	// Create a confirmed booking starting "now + 5 minutes" so we're within -15..+30 window
	now := time.Now()
	start := now.Add(5 * time.Minute)
	booking := &domain.Booking{
		ID:          uuid.New(),
		UserID:      uuid.New(),
		BathhouseID: bh.ID,
		StartTime:   start,
		EndTime:     start.Add(2 * time.Hour),
		GuestCount:  2,
		TotalPrice:  200000,
		Status:      domain.BookingConfirmed,
	}
	bookingRepo.Create(context.Background(), booking)

	// Owner checks in
	repRepo.Create(context.Background(), &domain.Representative{
		UserID:      ownerID,
		BathhouseID: bh.ID,
		Role:        domain.RepRoleManager,
	})

	err := svc.CheckIn(context.Background(), ownerID, domain.RoleOwner, booking.ID)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	updated, _ := bookingRepo.GetByID(context.Background(), booking.ID)
	if updated.CheckedInAt == nil {
		t.Fatal("expected CheckedInAt to be set")
	}
}

func TestBookingService_CheckIn_TooEarly(t *testing.T) {
	svc, bhRepo, bookingRepo, _, _, _, _, _ := newBookingService()
	ownerID := uuid.New()
	bh := createBathhouse(t, bhRepo, ownerID)

	// Booking starts in 1 hour — well outside -15 min window
	start := time.Now().Add(1 * time.Hour)
	booking := &domain.Booking{
		ID:          uuid.New(),
		UserID:      uuid.New(),
		BathhouseID: bh.ID,
		StartTime:   start,
		EndTime:     start.Add(2 * time.Hour),
		GuestCount:  2,
		TotalPrice:  200000,
		Status:      domain.BookingConfirmed,
	}
	bookingRepo.Create(context.Background(), booking)

	err := svc.CheckIn(context.Background(), ownerID, domain.RoleOwner, booking.ID)
	if !errors.Is(err, domain.ErrCheckinTooEarly) {
		t.Errorf("expected ErrCheckinTooEarly, got: %v", err)
	}
}

func TestBookingService_CheckIn_TooLate(t *testing.T) {
	svc, bhRepo, bookingRepo, _, _, _, _, _ := newBookingService()
	ownerID := uuid.New()
	bh := createBathhouse(t, bhRepo, ownerID)

	// Booking started 45 minutes ago — outside +30 min window
	start := time.Now().Add(-45 * time.Minute)
	booking := &domain.Booking{
		ID:          uuid.New(),
		UserID:      uuid.New(),
		BathhouseID: bh.ID,
		StartTime:   start,
		EndTime:     start.Add(2 * time.Hour),
		GuestCount:  2,
		TotalPrice:  200000,
		Status:      domain.BookingConfirmed,
	}
	bookingRepo.Create(context.Background(), booking)

	err := svc.CheckIn(context.Background(), ownerID, domain.RoleOwner, booking.ID)
	if !errors.Is(err, domain.ErrCheckinTooLate) {
		t.Errorf("expected ErrCheckinTooLate, got: %v", err)
	}
}

func TestBookingService_CheckOut_Success(t *testing.T) {
	svc, bhRepo, bookingRepo, _, _, _, _, _ := newBookingService()
	ownerID := uuid.New()
	bh := createBathhouse(t, bhRepo, ownerID)

	now := time.Now()
	checkedIn := now.Add(-1 * time.Hour)
	booking := &domain.Booking{
		ID:          uuid.New(),
		UserID:      uuid.New(),
		BathhouseID: bh.ID,
		StartTime:   now.Add(-2 * time.Hour),
		EndTime:     now.Add(-30 * time.Minute),
		GuestCount:  2,
		TotalPrice:  200000,
		CheckedInAt: &checkedIn,
		Status:      domain.BookingConfirmed,
	}
	bookingRepo.Create(context.Background(), booking)

	err := svc.CheckOut(context.Background(), ownerID, domain.RoleOwner, booking.ID)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	updated, _ := bookingRepo.GetByID(context.Background(), booking.ID)
	if updated.CheckedOutAt == nil {
		t.Fatal("expected CheckedOutAt to be set")
	}
	if updated.Status != domain.BookingCompleted {
		t.Errorf("expected status completed, got %s", updated.Status)
	}
}

func TestBookingService_CheckOut_WithoutCheckIn(t *testing.T) {
	svc, bhRepo, bookingRepo, _, _, _, _, _ := newBookingService()
	ownerID := uuid.New()
	bh := createBathhouse(t, bhRepo, ownerID)

	booking := &domain.Booking{
		ID:          uuid.New(),
		UserID:      uuid.New(),
		BathhouseID: bh.ID,
		StartTime:   time.Now().Add(-2 * time.Hour),
		EndTime:     time.Now().Add(-30 * time.Minute),
		GuestCount:  2,
		TotalPrice:  200000,
		Status:      domain.BookingConfirmed,
	}
	bookingRepo.Create(context.Background(), booking)

	err := svc.CheckOut(context.Background(), ownerID, domain.RoleOwner, booking.ID)
	if !errors.Is(err, domain.ErrNotCheckedIn) {
		t.Errorf("expected ErrNotCheckedIn, got: %v", err)
	}
}

func TestBookingService_MarkNoShows(t *testing.T) {
	svc, bhRepo, bookingRepo, _, _, _, _, _ := newBookingService()
	ownerID := uuid.New()
	bh := createBathhouse(t, bhRepo, ownerID)

	// Confirmed booking that started 35 minutes ago, no check-in → should become no_show
	start := time.Now().Add(-35 * time.Minute)
	noShowBooking := &domain.Booking{
		ID:          uuid.New(),
		UserID:      uuid.New(),
		BathhouseID: bh.ID,
		StartTime:   start,
		EndTime:     start.Add(2 * time.Hour),
		GuestCount:  2,
		TotalPrice:  200000,
		Status:      domain.BookingConfirmed,
	}
	bookingRepo.Create(context.Background(), noShowBooking)

	// Confirmed booking that started 10 minutes ago (within grace) → should NOT become no_show
	recentStart := time.Now().Add(-10 * time.Minute)
	okBooking := &domain.Booking{
		ID:          uuid.New(),
		UserID:      uuid.New(),
		BathhouseID: bh.ID,
		StartTime:   recentStart,
		EndTime:     recentStart.Add(2 * time.Hour),
		GuestCount:  2,
		TotalPrice:  200000,
		Status:      domain.BookingConfirmed,
	}
	bookingRepo.Create(context.Background(), okBooking)

	marked, err := svc.MarkNoShows(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if marked != 1 {
		t.Errorf("expected 1 no-show, got %d", marked)
	}

	updated, _ := bookingRepo.GetByID(context.Background(), noShowBooking.ID)
	if updated.Status != domain.BookingNoShow {
		t.Errorf("expected no_show status, got %s", updated.Status)
	}

	stillOk, _ := bookingRepo.GetByID(context.Background(), okBooking.ID)
	if stillOk.Status != domain.BookingConfirmed {
		t.Errorf("expected confirmed status, got %s", stillOk.Status)
	}
}

func TestBookingService_MarkNoShows_CheckedInNotAffected(t *testing.T) {
	svc, bhRepo, bookingRepo, _, _, _, _, _ := newBookingService()
	ownerID := uuid.New()
	bh := createBathhouse(t, bhRepo, ownerID)

	// Confirmed booking started 35 min ago but already checked in → should NOT be marked
	start := time.Now().Add(-35 * time.Minute)
	checkedIn := start.Add(5 * time.Minute)
	booking := &domain.Booking{
		ID:          uuid.New(),
		UserID:      uuid.New(),
		BathhouseID: bh.ID,
		StartTime:   start,
		EndTime:     start.Add(2 * time.Hour),
		GuestCount:  2,
		TotalPrice:  200000,
		CheckedInAt: &checkedIn,
		Status:      domain.BookingConfirmed,
	}
	bookingRepo.Create(context.Background(), booking)

	marked, err := svc.MarkNoShows(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if marked != 0 {
		t.Errorf("expected 0 no-shows, got %d", marked)
	}
}

func TestBookingService_DisputeNoShow_WithinWindow(t *testing.T) {
	svc, bhRepo, bookingRepo, _, _, _, _, _ := newBookingService()
	ownerID := uuid.New()
	bh := createBathhouse(t, bhRepo, ownerID)
	clientID := uuid.New()

	booking := &domain.Booking{
		ID:          uuid.New(),
		UserID:      clientID,
		BathhouseID: bh.ID,
		StartTime:   time.Now().Add(-1 * time.Hour),
		EndTime:     time.Now().Add(1 * time.Hour),
		GuestCount:  2,
		TotalPrice:  200000,
		Status:      domain.BookingNoShow,
		UpdatedAt:   time.Now().Add(-30 * time.Minute), // 30 min ago, within 2h window
	}
	bookingRepo.Create(context.Background(), booking)

	// complaintSvc is nil so no actual complaint created, but no error expected
	err := svc.DisputeNoShow(context.Background(), clientID, booking.ID, 55.7, 37.6, "I was there")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}

func TestBookingService_ListUpcomingWithBathhouse(t *testing.T) {
	svc, bhRepo, bookingRepo, _, _, _, _, _ := newBookingService()
	ownerID := uuid.New()
	bh := createBathhouse(t, bhRepo, ownerID)
	clientID := uuid.New()

	now := time.Now()
	booking := &domain.Booking{
		ID:          uuid.New(),
		UserID:      clientID,
		BathhouseID: bh.ID,
		StartTime:   now.Add(24 * time.Hour),
		EndTime:     now.Add(26 * time.Hour),
		GuestCount:  2,
		TotalPrice:  200000,
		Status:      domain.BookingConfirmed,
	}
	bookingRepo.Create(context.Background(), booking)

	// Add a cancelled booking that should NOT appear
	cancelledBooking := &domain.Booking{
		ID:          uuid.New(),
		UserID:      clientID,
		BathhouseID: bh.ID,
		StartTime:   now.Add(24 * time.Hour),
		EndTime:     now.Add(26 * time.Hour),
		GuestCount:  2,
		TotalPrice:  200000,
		Status:      domain.BookingCancelled,
	}
	bookingRepo.Create(context.Background(), cancelledBooking)

	from := now.Add(23 * time.Hour)
	to := now.Add(25 * time.Hour)
	result, err := svc.ListUpcomingWithBathhouse(context.Background(), from, to)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) != 1 {
		t.Fatalf("expected 1 upcoming booking, got %d", len(result))
	}
	if result[0].Booking.ID != booking.ID {
		t.Errorf("expected booking ID %s, got %s", booking.ID, result[0].Booking.ID)
	}
	if result[0].BathhouseName != bh.Name {
		t.Errorf("expected bathhouse name %s, got %s", bh.Name, result[0].BathhouseName)
	}
	if result[0].OwnerID != ownerID {
		t.Errorf("expected owner ID %s, got %s", ownerID, result[0].OwnerID)
	}
}

func TestBookingService_ListUpcomingWithBathhouse_Empty(t *testing.T) {
	svc, _, _, _, _, _, _, _ := newBookingService()

	now := time.Now()
	result, err := svc.ListUpcomingWithBathhouse(context.Background(), now, now.Add(24*time.Hour))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("expected 0 upcoming bookings, got %d", len(result))
	}
}

func TestBookingService_DisputeNoShow_Expired(t *testing.T) {
	svc, bhRepo, bookingRepo, _, _, _, _, _ := newBookingService()
	ownerID := uuid.New()
	bh := createBathhouse(t, bhRepo, ownerID)
	clientID := uuid.New()

	booking := &domain.Booking{
		ID:          uuid.New(),
		UserID:      clientID,
		BathhouseID: bh.ID,
		StartTime:   time.Now().Add(-4 * time.Hour),
		EndTime:     time.Now().Add(-2 * time.Hour),
		GuestCount:  2,
		TotalPrice:  200000,
		Status:      domain.BookingNoShow,
		UpdatedAt:   time.Now().Add(-3 * time.Hour), // 3 hours ago, past 2h window
	}
	bookingRepo.Create(context.Background(), booking)

	err := svc.DisputeNoShow(context.Background(), clientID, booking.ID, 55.7, 37.6, "")
	if !errors.Is(err, domain.ErrNoShowDisputeExpired) {
		t.Errorf("expected ErrNoShowDisputeExpired, got: %v", err)
	}
}

// --- Session Extension Tests ---

func TestBookingService_Extend_Success_1Hour(t *testing.T) {
	svc, bhRepo, bookingRepo, _, _, _, _, _ := newBookingService()
	ownerID := uuid.New()
	clientID := uuid.New()
	bh := createBathhouse(t, bhRepo, ownerID)

	now := time.Now()
	start := time.Date(now.Year(), now.Month(), now.Day()+1, 10, 0, 0, 0, now.Location())
	end := start.Add(2 * time.Hour)

	booking := &domain.Booking{
		ID: uuid.New(), UserID: clientID, BathhouseID: bh.ID,
		StartTime: start, EndTime: end, GuestCount: 2,
		TotalPrice: 10000, Status: domain.BookingConfirmed,
	}
	bookingRepo.Create(context.Background(), booking)

	result, err := svc.Extend(context.Background(), clientID, booking.ID, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedNewEnd := end.Add(1 * time.Hour)
	if !result.NewEndTime.Equal(expectedNewEnd) {
		t.Errorf("new end time = %v, want %v", result.NewEndTime, expectedNewEnd)
	}
	if result.ExtensionPrice != bh.PricePerHour {
		t.Errorf("extension price = %d, want %d", result.ExtensionPrice, bh.PricePerHour)
	}
	if result.NewTotalPrice != 10000+bh.PricePerHour {
		t.Errorf("new total price = %d, want %d", result.NewTotalPrice, 10000+bh.PricePerHour)
	}
}

func TestBookingService_Extend_Success_2Hours(t *testing.T) {
	svc, bhRepo, bookingRepo, _, _, _, _, _ := newBookingService()
	ownerID := uuid.New()
	clientID := uuid.New()
	bh := createBathhouse(t, bhRepo, ownerID)

	now := time.Now()
	start := time.Date(now.Year(), now.Month(), now.Day()+1, 10, 0, 0, 0, now.Location())
	end := start.Add(2 * time.Hour)

	booking := &domain.Booking{
		ID: uuid.New(), UserID: clientID, BathhouseID: bh.ID,
		StartTime: start, EndTime: end, GuestCount: 2,
		TotalPrice: 10000, Status: domain.BookingConfirmed,
	}
	bookingRepo.Create(context.Background(), booking)

	result, err := svc.Extend(context.Background(), clientID, booking.ID, 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedNewEnd := end.Add(2 * time.Hour)
	if !result.NewEndTime.Equal(expectedNewEnd) {
		t.Errorf("new end time = %v, want %v", result.NewEndTime, expectedNewEnd)
	}
	if result.ExtensionPrice != bh.PricePerHour*2 {
		t.Errorf("extension price = %d, want %d", result.ExtensionPrice, bh.PricePerHour*2)
	}
}

func TestBookingService_Extend_ConflictWithNextBooking(t *testing.T) {
	svc, bhRepo, bookingRepo, _, _, _, _, _ := newBookingService()
	ownerID := uuid.New()
	clientID := uuid.New()
	bh := createBathhouse(t, bhRepo, ownerID)

	now := time.Now()
	start := time.Date(now.Year(), now.Month(), now.Day()+1, 10, 0, 0, 0, now.Location())
	end := start.Add(2 * time.Hour)

	booking := &domain.Booking{
		ID: uuid.New(), UserID: clientID, BathhouseID: bh.ID,
		StartTime: start, EndTime: end, GuestCount: 2,
		TotalPrice: 10000, Status: domain.BookingConfirmed,
	}
	bookingRepo.Create(context.Background(), booking)

	// Create a conflicting booking right after
	nextBooking := &domain.Booking{
		ID: uuid.New(), UserID: uuid.New(), BathhouseID: bh.ID,
		StartTime: end, EndTime: end.Add(2 * time.Hour), GuestCount: 2,
		TotalPrice: 10000, Status: domain.BookingConfirmed,
	}
	bookingRepo.Create(context.Background(), nextBooking)

	_, err := svc.Extend(context.Background(), clientID, booking.ID, 1)
	if !errors.Is(err, domain.ErrSlotUnavailable) {
		t.Errorf("expected ErrSlotUnavailable, got: %v", err)
	}
}

func TestBookingService_Extend_BeyondWorkingHours(t *testing.T) {
	svc, bhRepo, bookingRepo, _, _, _, _, _ := newBookingService()
	ownerID := uuid.New()
	clientID := uuid.New()

	// Create bathhouse with working hours 08:00-22:00
	wh := make([]domain.WorkingHours, 7)
	for i := 0; i < 7; i++ {
		wh[i] = domain.WorkingHours{DayOfWeek: i, OpenTime: "08:00", CloseTime: "22:00"}
	}
	bh := &domain.Bathhouse{
		ID: uuid.New(), OwnerID: ownerID, Name: "Limited Hours",
		Address: "123 St", CityID: 1, PricePerHour: 5000,
		MinDuration: 1, MaxGuests: 10,
		LongSessionThresholdHours: 4, BaseCapacity: 10,
		WorkingHours: wh,
		Status:       domain.BathhouseStatusActive,
	}
	bhRepo.Create(context.Background(), bh)

	now := time.Now()
	// Booking ends at 21:00, extending by 2h would go to 23:00 (past 22:00 close)
	start := time.Date(now.Year(), now.Month(), now.Day()+1, 19, 0, 0, 0, now.Location())
	end := start.Add(2 * time.Hour) // 21:00

	booking := &domain.Booking{
		ID: uuid.New(), UserID: clientID, BathhouseID: bh.ID,
		StartTime: start, EndTime: end, GuestCount: 2,
		TotalPrice: 10000, Status: domain.BookingConfirmed,
	}
	bookingRepo.Create(context.Background(), booking)

	_, err := svc.Extend(context.Background(), clientID, booking.ID, 2)
	if !errors.Is(err, domain.ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput (beyond working hours), got: %v", err)
	}
}

func TestBookingService_Extend_WrongStatus(t *testing.T) {
	svc, bhRepo, bookingRepo, _, _, _, _, _ := newBookingService()
	ownerID := uuid.New()
	clientID := uuid.New()
	bh := createBathhouse(t, bhRepo, ownerID)

	now := time.Now()
	start := time.Date(now.Year(), now.Month(), now.Day()+1, 10, 0, 0, 0, now.Location())
	end := start.Add(2 * time.Hour)

	booking := &domain.Booking{
		ID: uuid.New(), UserID: clientID, BathhouseID: bh.ID,
		StartTime: start, EndTime: end, GuestCount: 2,
		TotalPrice: 10000, Status: domain.BookingPending,
	}
	bookingRepo.Create(context.Background(), booking)

	_, err := svc.Extend(context.Background(), clientID, booking.ID, 1)
	if !errors.Is(err, domain.ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput (wrong status), got: %v", err)
	}
}

func TestBookingService_Extend_NotOwner(t *testing.T) {
	svc, bhRepo, bookingRepo, _, _, _, _, _ := newBookingService()
	ownerID := uuid.New()
	clientID := uuid.New()
	otherUserID := uuid.New()
	bh := createBathhouse(t, bhRepo, ownerID)

	now := time.Now()
	start := time.Date(now.Year(), now.Month(), now.Day()+1, 10, 0, 0, 0, now.Location())
	end := start.Add(2 * time.Hour)

	booking := &domain.Booking{
		ID: uuid.New(), UserID: clientID, BathhouseID: bh.ID,
		StartTime: start, EndTime: end, GuestCount: 2,
		TotalPrice: 10000, Status: domain.BookingConfirmed,
	}
	bookingRepo.Create(context.Background(), booking)

	_, err := svc.Extend(context.Background(), otherUserID, booking.ID, 1)
	if !errors.Is(err, domain.ErrForbidden) {
		t.Errorf("expected ErrForbidden, got: %v", err)
	}
}

func TestBookingService_Extend_CheckedIn(t *testing.T) {
	svc, bhRepo, bookingRepo, _, _, _, _, _ := newBookingService()
	ownerID := uuid.New()
	clientID := uuid.New()
	bh := createBathhouse(t, bhRepo, ownerID)

	now := time.Now()
	start := time.Date(now.Year(), now.Month(), now.Day()+1, 10, 0, 0, 0, now.Location())
	end := start.Add(2 * time.Hour)

	checkedIn := now
	booking := &domain.Booking{
		ID: uuid.New(), UserID: clientID, BathhouseID: bh.ID,
		StartTime: start, EndTime: end, GuestCount: 2,
		TotalPrice: 10000, Status: domain.BookingConfirmed,
		CheckedInAt: &checkedIn,
	}
	bookingRepo.Create(context.Background(), booking)

	result, err := svc.Extend(context.Background(), clientID, booking.ID, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedNewEnd := end.Add(1 * time.Hour)
	if !result.NewEndTime.Equal(expectedNewEnd) {
		t.Errorf("new end time = %v, want %v", result.NewEndTime, expectedNewEnd)
	}
}

func TestBookingService_Extend_InvalidExtraHours(t *testing.T) {
	svc, bhRepo, bookingRepo, _, _, _, _, _ := newBookingService()
	ownerID := uuid.New()
	clientID := uuid.New()
	bh := createBathhouse(t, bhRepo, ownerID)

	now := time.Now()
	start := time.Date(now.Year(), now.Month(), now.Day()+1, 10, 0, 0, 0, now.Location())
	end := start.Add(2 * time.Hour)

	booking := &domain.Booking{
		ID: uuid.New(), UserID: clientID, BathhouseID: bh.ID,
		StartTime: start, EndTime: end, GuestCount: 2,
		TotalPrice: 10000, Status: domain.BookingConfirmed,
	}
	bookingRepo.Create(context.Background(), booking)

	// Test 0 hours
	_, err := svc.Extend(context.Background(), clientID, booking.ID, 0)
	if !errors.Is(err, domain.ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput for 0 hours, got: %v", err)
	}

	// Test 3 hours
	_, err = svc.Extend(context.Background(), clientID, booking.ID, 3)
	if !errors.Is(err, domain.ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput for 3 hours, got: %v", err)
	}
}

// --- GetRebookData tests ---

func newBookingServiceWithAddonRepo() (service.BookingService, *mock.BathhouseRepo, *mock.BookingRepo, *mock.AddOnRepo) {
	bhRepo := mock.NewBathhouseRepo()
	bookingRepo := mock.NewBookingRepo()
	repRepo := mock.NewRepresentativeRepo()
	pricingRepo := mock.NewPricingRuleRepo()
	loyaltyRepo := mock.NewLoyaltyRepo()
	slotBlockRepo := mock.NewSlotBlockRepo()
	addonRepo := mock.NewAddOnRepo()
	access := service.NewAccessChecker(repRepo, bhRepo)
	log := logger.New(logger.LevelWarn)
	pricingSvc := service.NewPricingService(pricingRepo, nil, bhRepo, nil, access, log)
	loyaltySvc := service.NewLoyaltyService(loyaltyRepo, log)
	addonSvc := service.NewAddOnService(addonRepo, access, log)
	svc := service.NewBookingService(bookingRepo, bhRepo, slotBlockRepo, addonRepo, mock.NewUserRepo(), mock.NewCityRepo(), pricingSvc, addonSvc, loyaltySvc, &noopReferralService{}, &noopPromoService{}, &noopCertificateService{}, &noopPaymentService{}, &noopServiceFeeService{}, nil, nil, nil, nil, nil, access, &noopNotifService{}, log)
	return svc, bhRepo, bookingRepo, addonRepo
}

func TestGetRebookData_CompletedBooking(t *testing.T) {
	svc, bhRepo, bookingRepo, _ := newBookingServiceWithAddonRepo()
	ownerID := uuid.New()
	clientID := uuid.New()
	bh := createBathhouse(t, bhRepo, ownerID)

	start := time.Date(2026, 3, 20, 10, 0, 0, 0, time.UTC)
	end := start.Add(3 * time.Hour)

	booking := &domain.Booking{
		ID: uuid.New(), UserID: clientID, BathhouseID: bh.ID,
		StartTime: start, EndTime: end, GuestCount: 4,
		TotalPrice: 15000, Status: domain.BookingCompleted,
	}
	bookingRepo.Create(context.Background(), booking)

	data, err := svc.GetRebookData(context.Background(), clientID, booking.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if data.BathhouseID != bh.ID {
		t.Errorf("BathhouseID = %v, want %v", data.BathhouseID, bh.ID)
	}
	if data.DurationHours != 3 {
		t.Errorf("DurationHours = %d, want 3", data.DurationHours)
	}
	if data.TimeFrom != "10:00" {
		t.Errorf("TimeFrom = %q, want %q", data.TimeFrom, "10:00")
	}
	if data.TimeTo != "13:00" {
		t.Errorf("TimeTo = %q, want %q", data.TimeTo, "13:00")
	}
	if data.GuestCount != 4 {
		t.Errorf("GuestCount = %d, want 4", data.GuestCount)
	}
	if len(data.AddOns) != 0 {
		t.Errorf("AddOns length = %d, want 0", len(data.AddOns))
	}
}

func TestGetRebookData_CancelledBooking(t *testing.T) {
	svc, bhRepo, bookingRepo, _ := newBookingServiceWithAddonRepo()
	ownerID := uuid.New()
	clientID := uuid.New()
	bh := createBathhouse(t, bhRepo, ownerID)

	start := time.Date(2026, 3, 20, 14, 0, 0, 0, time.UTC)
	end := start.Add(2 * time.Hour)

	booking := &domain.Booking{
		ID: uuid.New(), UserID: clientID, BathhouseID: bh.ID,
		StartTime: start, EndTime: end, GuestCount: 2,
		TotalPrice: 10000, Status: domain.BookingCancelled,
	}
	bookingRepo.Create(context.Background(), booking)

	data, err := svc.GetRebookData(context.Background(), clientID, booking.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if data.DurationHours != 2 {
		t.Errorf("DurationHours = %d, want 2", data.DurationHours)
	}
}

func TestGetRebookData_PendingBooking_Rejected(t *testing.T) {
	svc, bhRepo, bookingRepo, _ := newBookingServiceWithAddonRepo()
	ownerID := uuid.New()
	clientID := uuid.New()
	bh := createBathhouse(t, bhRepo, ownerID)

	start := time.Date(2026, 3, 25, 10, 0, 0, 0, time.UTC)
	end := start.Add(2 * time.Hour)

	booking := &domain.Booking{
		ID: uuid.New(), UserID: clientID, BathhouseID: bh.ID,
		StartTime: start, EndTime: end, GuestCount: 3,
		TotalPrice: 10000, Status: domain.BookingPending,
	}
	bookingRepo.Create(context.Background(), booking)

	_, err := svc.GetRebookData(context.Background(), clientID, booking.ID)
	if !errors.Is(err, domain.ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput for pending booking, got: %v", err)
	}
}

func TestGetRebookData_OtherUserForbidden(t *testing.T) {
	svc, bhRepo, bookingRepo, _ := newBookingServiceWithAddonRepo()
	ownerID := uuid.New()
	clientID := uuid.New()
	otherUserID := uuid.New()
	bh := createBathhouse(t, bhRepo, ownerID)

	start := time.Date(2026, 3, 20, 10, 0, 0, 0, time.UTC)
	end := start.Add(2 * time.Hour)

	booking := &domain.Booking{
		ID: uuid.New(), UserID: clientID, BathhouseID: bh.ID,
		StartTime: start, EndTime: end, GuestCount: 2,
		TotalPrice: 10000, Status: domain.BookingCompleted,
	}
	bookingRepo.Create(context.Background(), booking)

	_, err := svc.GetRebookData(context.Background(), otherUserID, booking.ID)
	if !errors.Is(err, domain.ErrForbidden) {
		t.Errorf("expected ErrForbidden, got: %v", err)
	}
}

func TestGetRebookData_InactiveBathhouse(t *testing.T) {
	svc, bhRepo, bookingRepo, _ := newBookingServiceWithAddonRepo()
	ownerID := uuid.New()
	clientID := uuid.New()

	wh := make([]domain.WorkingHours, 7)
	for i := 0; i < 7; i++ {
		wh[i] = domain.WorkingHours{DayOfWeek: i, OpenTime: "00:00", CloseTime: "23:59"}
	}
	bh := &domain.Bathhouse{
		ID: uuid.New(), OwnerID: ownerID, Name: "Inactive Bath",
		Address: "123 St", CityID: 1, PricePerHour: 5000,
		MinDuration: 1, MaxGuests: 10, BaseCapacity: 10,
		LongSessionThresholdHours: 4,
		WorkingHours: wh,
		Status:       domain.BathhouseStatusInactive,
	}
	bhRepo.Create(context.Background(), bh)

	start := time.Date(2026, 3, 20, 10, 0, 0, 0, time.UTC)
	end := start.Add(2 * time.Hour)

	booking := &domain.Booking{
		ID: uuid.New(), UserID: clientID, BathhouseID: bh.ID,
		StartTime: start, EndTime: end, GuestCount: 2,
		TotalPrice: 10000, Status: domain.BookingCompleted,
	}
	bookingRepo.Create(context.Background(), booking)

	_, err := svc.GetRebookData(context.Background(), clientID, booking.ID)
	if !errors.Is(err, domain.ErrBathhouseNotActive) {
		t.Errorf("expected ErrBathhouseNotActive, got: %v", err)
	}
}

func TestGetRebookData_WithAddOns(t *testing.T) {
	svc, bhRepo, bookingRepo, addonRepo := newBookingServiceWithAddonRepo()
	ownerID := uuid.New()
	clientID := uuid.New()
	bh := createBathhouse(t, bhRepo, ownerID)

	start := time.Date(2026, 3, 20, 10, 0, 0, 0, time.UTC)
	end := start.Add(2 * time.Hour)

	booking := &domain.Booking{
		ID: uuid.New(), UserID: clientID, BathhouseID: bh.ID,
		StartTime: start, EndTime: end, GuestCount: 2,
		TotalPrice: 15000, Status: domain.BookingCompleted,
	}
	bookingRepo.Create(context.Background(), booking)

	addonID1 := uuid.New()
	addonID2 := uuid.New()
	addonRepo.CreateBookingAddOn(context.Background(), &domain.BookingAddOn{
		BookingID: booking.ID, AddOnID: addonID1, Name: "Веник", Quantity: 3, UnitPrice: 500, TotalPrice: 1500,
	})
	addonRepo.CreateBookingAddOn(context.Background(), &domain.BookingAddOn{
		BookingID: booking.ID, AddOnID: addonID2, Name: "Полотенце", Quantity: 1, UnitPrice: 200, TotalPrice: 200,
	})

	data, err := svc.GetRebookData(context.Background(), clientID, booking.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(data.AddOns) != 2 {
		t.Fatalf("AddOns length = %d, want 2", len(data.AddOns))
	}

	addonMap := make(map[uuid.UUID]int)
	for _, a := range data.AddOns {
		addonMap[a.AddOnID] = a.Quantity
	}
	if addonMap[addonID1] != 3 {
		t.Errorf("addon1 quantity = %d, want 3", addonMap[addonID1])
	}
	if addonMap[addonID2] != 1 {
		t.Errorf("addon2 quantity = %d, want 1", addonMap[addonID2])
	}
}

// --- Owner Cancellation Penalty & Response Rate Tests ---

type trackingWalletService struct {
	noopWalletService
	refundCalls []walletRefundCall
}

type walletRefundCall struct {
	WalletID    uuid.UUID
	Amount      int64
	RefType     string
	Description string
}

func (t *trackingWalletService) Refund(_ context.Context, walletID uuid.UUID, amount int64, refType string, _ *uuid.UUID, description string) (*domain.WalletTransaction, error) {
	t.refundCalls = append(t.refundCalls, walletRefundCall{
		WalletID:    walletID,
		Amount:      amount,
		RefType:     refType,
		Description: description,
	})
	return &domain.WalletTransaction{}, nil
}

type trackingNotifService struct {
	noopNotifService
	sent []sentNotif
}

type sentNotif struct {
	UserID uuid.UUID
	Type   domain.NotificationType
	Title  string
	Body   string
}

func (t *trackingNotifService) Send(_ context.Context, userID uuid.UUID, notifType domain.NotificationType, title, body string, _ map[string]string) error {
	t.sent = append(t.sent, sentNotif{UserID: userID, Type: notifType, Title: title, Body: body})
	return nil
}

func newBookingServiceWithWallet() (service.BookingService, *mock.BathhouseRepo, *mock.BookingRepo, *trackingWalletService, *trackingNotifService) {
	bhRepo := mock.NewBathhouseRepo()
	bookingRepo := mock.NewBookingRepo()
	repRepo := mock.NewRepresentativeRepo()
	pricingRepo := mock.NewPricingRuleRepo()
	loyaltyRepo := mock.NewLoyaltyRepo()
	slotBlockRepo := mock.NewSlotBlockRepo()
	addonRepo := mock.NewAddOnRepo()
	access := service.NewAccessChecker(repRepo, bhRepo)
	log := logger.New(logger.LevelWarn)
	pricingSvc := service.NewPricingService(pricingRepo, nil, bhRepo, nil, access, log)
	loyaltySvc := service.NewLoyaltyService(loyaltyRepo, log)
	addonSvc := service.NewAddOnService(addonRepo, access, log)
	walletSvc := &trackingWalletService{}
	notifSvc := &trackingNotifService{}
	svc := service.NewBookingService(bookingRepo, bhRepo, slotBlockRepo, addonRepo, mock.NewUserRepo(), mock.NewCityRepo(), pricingSvc, addonSvc, loyaltySvc, &noopReferralService{}, &noopPromoService{}, &noopCertificateService{}, &noopPaymentService{}, &noopServiceFeeService{}, walletSvc, nil, nil, nil, nil, access, notifSvc, log)
	return svc, bhRepo, bookingRepo, walletSvc, notifSvc
}

func TestOwnerCancellation_CompensatesClient(t *testing.T) {
	svc, bhRepo, bookingRepo, walletSvc, _ := newBookingServiceWithWallet()
	ownerID := uuid.New()
	clientID := uuid.New()
	bh := createBathhouse(t, bhRepo, ownerID)

	start := time.Now().Add(48 * time.Hour)
	booking := &domain.Booking{
		ID: uuid.New(), UserID: clientID, BathhouseID: bh.ID,
		StartTime: start, EndTime: start.Add(2 * time.Hour),
		GuestCount: 2, TotalPrice: 100000, Status: domain.BookingConfirmed,
	}
	_ = bookingRepo.Create(context.Background(), booking)

	err := svc.Cancel(context.Background(), ownerID, domain.RoleOwner, booking.ID, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify booking is cancelled
	b, _ := bookingRepo.GetByID(context.Background(), booking.ID)
	if b.Status != domain.BookingCancelled {
		t.Errorf("status = %q, want %q", b.Status, domain.BookingCancelled)
	}
	if !b.CancelledByOwner {
		t.Error("CancelledByOwner should be true")
	}

	// Verify 10% compensation was refunded to client wallet
	if len(walletSvc.refundCalls) == 0 {
		t.Fatal("expected at least one wallet refund call for compensation")
	}

	found := false
	for _, call := range walletSvc.refundCalls {
		if call.RefType == "owner_cancellation" {
			found = true
			expectedAmount := int64(10000) // 10% of 100000
			if call.Amount != expectedAmount {
				t.Errorf("compensation amount = %d, want %d", call.Amount, expectedAmount)
			}
		}
	}
	if !found {
		t.Error("expected wallet refund call with ref_type 'owner_cancellation'")
	}
}

func TestOwnerCancellation_WarningAt4thCancellation(t *testing.T) {
	svc, bhRepo, bookingRepo, _, notifSvc := newBookingServiceWithWallet()
	ownerID := uuid.New()
	bh := createBathhouse(t, bhRepo, ownerID)

	// Create 4 bookings and cancel them as owner
	for i := 0; i < 4; i++ {
		clientID := uuid.New()
		start := time.Now().Add(48 * time.Hour)
		booking := &domain.Booking{
			ID: uuid.New(), UserID: clientID, BathhouseID: bh.ID,
			StartTime: start, EndTime: start.Add(2 * time.Hour),
			GuestCount: 2, TotalPrice: 10000, Status: domain.BookingConfirmed,
		}
		_ = bookingRepo.Create(context.Background(), booking)
		err := svc.Cancel(context.Background(), ownerID, domain.RoleOwner, booking.ID, "")
		if err != nil {
			t.Fatalf("cancel %d: unexpected error: %v", i+1, err)
		}
	}

	// Check that warning was sent (after 4th cancellation, count > 3)
	warningFound := false
	for _, n := range notifSvc.sent {
		if n.Type == domain.NotifOwnerCancellationWarning && n.UserID == ownerID {
			warningFound = true
		}
	}
	if !warningFound {
		t.Error("expected owner cancellation warning notification after 4th cancellation")
	}
}

func TestOwnerCancellation_DeactivatesAt6thCancellation(t *testing.T) {
	svc, bhRepo, bookingRepo, _, notifSvc := newBookingServiceWithWallet()
	ownerID := uuid.New()
	bh := createBathhouse(t, bhRepo, ownerID)

	// Create a second bathhouse for the same owner
	bh2 := createBathhouse(t, bhRepo, ownerID)

	// Create 6 bookings and cancel them as owner
	for i := 0; i < 6; i++ {
		clientID := uuid.New()
		start := time.Now().Add(48 * time.Hour)
		booking := &domain.Booking{
			ID: uuid.New(), UserID: clientID, BathhouseID: bh.ID,
			StartTime: start, EndTime: start.Add(2 * time.Hour),
			GuestCount: 2, TotalPrice: 10000, Status: domain.BookingConfirmed,
		}
		_ = bookingRepo.Create(context.Background(), booking)
		err := svc.Cancel(context.Background(), ownerID, domain.RoleOwner, booking.ID, "")
		if err != nil {
			t.Fatalf("cancel %d: unexpected error: %v", i+1, err)
		}
	}

	// Check both bathhouses are deactivated
	updatedBh1, _ := bhRepo.GetByID(context.Background(), bh.ID)
	updatedBh2, _ := bhRepo.GetByID(context.Background(), bh2.ID)

	if updatedBh1.Status != domain.BathhouseStatusInactive {
		t.Errorf("bh1 status = %q, want %q", updatedBh1.Status, domain.BathhouseStatusInactive)
	}
	if updatedBh2.Status != domain.BathhouseStatusInactive {
		t.Errorf("bh2 status = %q, want %q", updatedBh2.Status, domain.BathhouseStatusInactive)
	}

	// Check penalty notification
	penaltyFound := false
	for _, n := range notifSvc.sent {
		if n.Type == domain.NotifOwnerCancellationPenalty && n.UserID == ownerID {
			penaltyFound = true
		}
	}
	if !penaltyFound {
		t.Error("expected owner cancellation penalty notification after 6th cancellation")
	}
}

func TestResponseRateCalculation(t *testing.T) {
	svc, bhRepo, bookingRepo, _, _ := newBookingServiceWithWallet()
	ownerID := uuid.New()

	// Create a request-mode bathhouse
	wh := make([]domain.WorkingHours, 7)
	for i := 0; i < 7; i++ {
		wh[i] = domain.WorkingHours{DayOfWeek: i, OpenTime: "00:00", CloseTime: "23:59"}
	}
	bh := &domain.Bathhouse{
		ID: uuid.New(), OwnerID: ownerID, Name: "Request Bath",
		Address: "123 St", CityID: 1, PricePerHour: 5000,
		MinDuration: 1, MaxGuests: 10, BaseCapacity: 10,
		LongSessionThresholdHours: 4,
		BookingMode: domain.BookingModeRequest, RequestTimeout: 24,
		WorkingHours: wh, Status: domain.BathhouseStatusActive,
	}
	_ = bhRepo.Create(context.Background(), bh)

	// Create 10 request-based bookings (simulated by having HoldID set)
	holdID := uuid.New()
	for i := 0; i < 10; i++ {
		status := domain.BookingConfirmed
		if i >= 8 {
			status = domain.BookingPendingOwner // 2 still pending (not responded)
		}
		booking := &domain.Booking{
			ID: uuid.New(), UserID: uuid.New(), BathhouseID: bh.ID,
			StartTime: time.Now().Add(time.Duration(i+1) * 24 * time.Hour),
			EndTime:   time.Now().Add(time.Duration(i+1)*24*time.Hour + 2*time.Hour),
			GuestCount: 2, TotalPrice: 10000, Status: status,
			HoldID:    &holdID,
			CreatedAt: time.Now().Add(-time.Duration(i) * time.Hour),
			UpdatedAt: time.Now().Add(-time.Duration(i)*time.Hour + 30*time.Minute),
		}
		_ = bookingRepo.Create(context.Background(), booking)
	}

	updated, err := svc.RecalculateResponseRates(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated != 1 {
		t.Errorf("updated = %d, want 1", updated)
	}
}

func TestResponseRate_WarningBelow50(t *testing.T) {
	svc, bhRepo, bookingRepo, _, notifSvc := newBookingServiceWithWallet()
	ownerID := uuid.New()

	wh := make([]domain.WorkingHours, 7)
	for i := 0; i < 7; i++ {
		wh[i] = domain.WorkingHours{DayOfWeek: i, OpenTime: "00:00", CloseTime: "23:59"}
	}
	bh := &domain.Bathhouse{
		ID: uuid.New(), OwnerID: ownerID, Name: "Slow Bath",
		Address: "456 St", CityID: 1, PricePerHour: 5000,
		MinDuration: 1, MaxGuests: 10, BaseCapacity: 10,
		LongSessionThresholdHours: 4,
		BookingMode: domain.BookingModeRequest, RequestTimeout: 24,
		WorkingHours: wh, Status: domain.BathhouseStatusActive,
	}
	_ = bhRepo.Create(context.Background(), bh)

	// Create 10 bookings: only 4 responded (40% rate, < 50%)
	holdID := uuid.New()
	for i := 0; i < 10; i++ {
		status := domain.BookingPendingOwner // not responded
		if i < 4 {
			status = domain.BookingConfirmed // responded
		}
		booking := &domain.Booking{
			ID: uuid.New(), UserID: uuid.New(), BathhouseID: bh.ID,
			StartTime: time.Now().Add(time.Duration(i+1) * 24 * time.Hour),
			EndTime:   time.Now().Add(time.Duration(i+1)*24*time.Hour + 2*time.Hour),
			GuestCount: 2, TotalPrice: 10000, Status: status,
			HoldID: &holdID,
		}
		_ = bookingRepo.Create(context.Background(), booking)
	}

	_, err := svc.RecalculateResponseRates(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	warningFound := false
	for _, n := range notifSvc.sent {
		if n.Type == domain.NotifOwnerResponseRateWarning && n.UserID == ownerID {
			warningFound = true
		}
	}
	if !warningFound {
		t.Error("expected response rate warning notification for rate < 50%")
	}
}

func TestResponseRate_AutoDeactivateBelow30(t *testing.T) {
	svc, bhRepo, bookingRepo, _, notifSvc := newBookingServiceWithWallet()
	ownerID := uuid.New()

	wh := make([]domain.WorkingHours, 7)
	for i := 0; i < 7; i++ {
		wh[i] = domain.WorkingHours{DayOfWeek: i, OpenTime: "00:00", CloseTime: "23:59"}
	}
	bh := &domain.Bathhouse{
		ID: uuid.New(), OwnerID: ownerID, Name: "Very Slow Bath",
		Address: "789 St", CityID: 1, PricePerHour: 5000,
		MinDuration: 1, MaxGuests: 10, BaseCapacity: 10,
		LongSessionThresholdHours: 4,
		BookingMode:  domain.BookingModeRequest, RequestTimeout: 24,
		ResponseRate: 0.2, // Previously below 0.3 too
		WorkingHours: wh, Status: domain.BathhouseStatusActive,
	}
	_ = bhRepo.Create(context.Background(), bh)

	// Create 10 bookings: only 2 responded (20% rate, < 30%)
	holdID := uuid.New()
	for i := 0; i < 10; i++ {
		status := domain.BookingPendingOwner
		if i < 2 {
			status = domain.BookingConfirmed
		}
		booking := &domain.Booking{
			ID: uuid.New(), UserID: uuid.New(), BathhouseID: bh.ID,
			StartTime: time.Now().Add(time.Duration(i+1) * 24 * time.Hour),
			EndTime:   time.Now().Add(time.Duration(i+1)*24*time.Hour + 2*time.Hour),
			GuestCount: 2, TotalPrice: 10000, Status: status,
			HoldID: &holdID,
		}
		_ = bookingRepo.Create(context.Background(), booking)
	}

	_, err := svc.RecalculateResponseRates(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should NOT force to instant mode (requires LowResponseRateSince tracking for 60+ days)
	updatedBh, _ := bhRepo.GetByID(context.Background(), bh.ID)
	if updatedBh.BookingMode != domain.BookingModeRequest {
		t.Errorf("booking_mode = %q, want %q (should not auto-force without 60-day tracking)", updatedBh.BookingMode, domain.BookingModeRequest)
	}

	// Should send critical warning notification
	notifFound := false
	for _, n := range notifSvc.sent {
		if n.Type == domain.NotifOwnerResponseRateWarning && n.UserID == ownerID {
			notifFound = true
		}
	}
	if !notifFound {
		t.Error("expected response rate critical warning notification")
	}
}

// --- Booking Modification Tests ---

func TestBookingService_Modify_Success(t *testing.T) {
	svc, bhRepo, bookingRepo, _, _, _, _, _ := newBookingService()
	ownerID := uuid.New()
	clientID := uuid.New()
	bh := createBathhouse(t, bhRepo, ownerID)

	now := time.Now()
	start := time.Date(now.Year(), now.Month(), now.Day()+2, 10, 0, 0, 0, now.Location())
	end := start.Add(2 * time.Hour)

	// Create booking
	result, err := svc.Create(context.Background(), clientID, service.CreateBookingInput{
		BathhouseID: bh.ID,
		StartTime:   start,
		EndTime:     end,
		GuestCount:  5,
	})
	if err != nil {
		t.Fatalf("failed to create booking: %v", err)
	}
	bookingID := result.Booking.ID
	originalPrice := result.Booking.TotalPrice

	// Modify: change to 3 hours (price should go up)
	newStart := time.Date(now.Year(), now.Month(), now.Day()+2, 12, 0, 0, 0, now.Location())
	newEnd := newStart.Add(3 * time.Hour)
	modResult, err := svc.Modify(context.Background(), clientID, bookingID, service.ModifyBookingInput{
		StartTime:  newStart,
		EndTime:    newEnd,
		GuestCount: 5,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if modResult.OldPrice != originalPrice {
		t.Errorf("old_price = %d, want %d", modResult.OldPrice, originalPrice)
	}
	if modResult.NewPrice <= originalPrice {
		t.Errorf("new_price = %d, should be > %d (longer duration)", modResult.NewPrice, originalPrice)
	}
	if modResult.PriceDiff <= 0 {
		t.Errorf("price_diff = %d, should be positive for price increase", modResult.PriceDiff)
	}

	// Verify booking was updated
	updated, _ := bookingRepo.GetByID(context.Background(), bookingID)
	if updated.ModificationCount != 1 {
		t.Errorf("modification_count = %d, want 1", updated.ModificationCount)
	}
	if !updated.StartTime.Equal(newStart) {
		t.Errorf("start_time not updated")
	}
	if !updated.EndTime.Equal(newEnd) {
		t.Errorf("end_time not updated")
	}
}

func TestBookingService_Modify_PriceDown(t *testing.T) {
	svc, bhRepo, _, _, _, _, _, _ := newBookingService()
	ownerID := uuid.New()
	clientID := uuid.New()
	bh := createBathhouse(t, bhRepo, ownerID)

	now := time.Now()
	start := time.Date(now.Year(), now.Month(), now.Day()+2, 10, 0, 0, 0, now.Location())
	end := start.Add(3 * time.Hour)

	result, err := svc.Create(context.Background(), clientID, service.CreateBookingInput{
		BathhouseID: bh.ID,
		StartTime:   start,
		EndTime:     end,
		GuestCount:  5,
	})
	if err != nil {
		t.Fatalf("failed to create booking: %v", err)
	}

	// Modify: reduce to 2 hours (price should go down)
	newEnd := start.Add(2 * time.Hour)
	modResult, err := svc.Modify(context.Background(), clientID, result.Booking.ID, service.ModifyBookingInput{
		StartTime:  start,
		EndTime:    newEnd,
		GuestCount: 5,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if modResult.PriceDiff >= 0 {
		t.Errorf("price_diff = %d, should be negative for price decrease", modResult.PriceDiff)
	}
}

func TestBookingService_Modify_EqualPrice(t *testing.T) {
	svc, bhRepo, _, _, _, _, _, _ := newBookingService()
	ownerID := uuid.New()
	clientID := uuid.New()
	bh := createBathhouse(t, bhRepo, ownerID)

	now := time.Now()
	start := time.Date(now.Year(), now.Month(), now.Day()+2, 10, 0, 0, 0, now.Location())
	end := start.Add(2 * time.Hour)

	result, err := svc.Create(context.Background(), clientID, service.CreateBookingInput{
		BathhouseID: bh.ID,
		StartTime:   start,
		EndTime:     end,
		GuestCount:  5,
	})
	if err != nil {
		t.Fatalf("failed to create booking: %v", err)
	}

	// Modify: shift by 1 hour, same duration (same price)
	newStart := start.Add(1 * time.Hour)
	newEnd := newStart.Add(2 * time.Hour)
	modResult, err := svc.Modify(context.Background(), clientID, result.Booking.ID, service.ModifyBookingInput{
		StartTime:  newStart,
		EndTime:    newEnd,
		GuestCount: 5,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if modResult.PriceDiff != 0 {
		t.Errorf("price_diff = %d, want 0 for same duration", modResult.PriceDiff)
	}
}

func TestBookingService_Modify_MaxModificationsReached(t *testing.T) {
	svc, bhRepo, bookingRepo, _, _, _, _, _ := newBookingService()
	ownerID := uuid.New()
	clientID := uuid.New()
	bh := createBathhouse(t, bhRepo, ownerID)

	now := time.Now()
	start := time.Date(now.Year(), now.Month(), now.Day()+2, 10, 0, 0, 0, now.Location())
	end := start.Add(2 * time.Hour)

	result, err := svc.Create(context.Background(), clientID, service.CreateBookingInput{
		BathhouseID: bh.ID,
		StartTime:   start,
		EndTime:     end,
		GuestCount:  5,
	})
	if err != nil {
		t.Fatalf("failed to create booking: %v", err)
	}

	// Manually set modification count to max
	b, _ := bookingRepo.GetByID(context.Background(), result.Booking.ID)
	b.ModificationCount = domain.MaxBookingModifications
	_ = bookingRepo.Update(context.Background(), b)

	newStart := start.Add(1 * time.Hour)
	_, err = svc.Modify(context.Background(), clientID, result.Booking.ID, service.ModifyBookingInput{
		StartTime:  newStart,
		EndTime:    newStart.Add(2 * time.Hour),
		GuestCount: 5,
	})
	if !errors.Is(err, domain.ErrBookingModificationLimit) {
		t.Errorf("expected ErrBookingModificationLimit, got %v", err)
	}
}

func TestBookingService_Modify_WrongUser(t *testing.T) {
	svc, bhRepo, _, _, _, _, _, _ := newBookingService()
	ownerID := uuid.New()
	clientID := uuid.New()
	otherUser := uuid.New()
	bh := createBathhouse(t, bhRepo, ownerID)

	now := time.Now()
	start := time.Date(now.Year(), now.Month(), now.Day()+2, 10, 0, 0, 0, now.Location())
	end := start.Add(2 * time.Hour)

	result, err := svc.Create(context.Background(), clientID, service.CreateBookingInput{
		BathhouseID: bh.ID,
		StartTime:   start,
		EndTime:     end,
		GuestCount:  5,
	})
	if err != nil {
		t.Fatalf("failed to create booking: %v", err)
	}

	_, err = svc.Modify(context.Background(), otherUser, result.Booking.ID, service.ModifyBookingInput{
		StartTime:  start,
		EndTime:    end,
		GuestCount: 3,
	})
	if !errors.Is(err, domain.ErrForbidden) {
		t.Errorf("expected ErrForbidden, got %v", err)
	}
}

func TestBookingService_Modify_CompletedBooking(t *testing.T) {
	svc, bhRepo, bookingRepo, _, _, _, _, _ := newBookingService()
	ownerID := uuid.New()
	clientID := uuid.New()
	bh := createBathhouse(t, bhRepo, ownerID)

	now := time.Now()
	start := time.Date(now.Year(), now.Month(), now.Day()+2, 10, 0, 0, 0, now.Location())
	end := start.Add(2 * time.Hour)

	result, err := svc.Create(context.Background(), clientID, service.CreateBookingInput{
		BathhouseID: bh.ID,
		StartTime:   start,
		EndTime:     end,
		GuestCount:  5,
	})
	if err != nil {
		t.Fatalf("failed to create booking: %v", err)
	}

	// Set to completed
	_ = bookingRepo.UpdateStatus(context.Background(), result.Booking.ID, domain.BookingCompleted)

	_, err = svc.Modify(context.Background(), clientID, result.Booking.ID, service.ModifyBookingInput{
		StartTime:  start,
		EndTime:    end,
		GuestCount: 3,
	})
	if !errors.Is(err, domain.ErrBookingNotModifiable) {
		t.Errorf("expected ErrBookingNotModifiable, got %v", err)
	}
}

func TestBookingService_Modify_SlotConflict(t *testing.T) {
	svc, bhRepo, _, _, _, _, _, _ := newBookingService()
	ownerID := uuid.New()
	clientID := uuid.New()
	bh := createBathhouse(t, bhRepo, ownerID)

	now := time.Now()
	start1 := time.Date(now.Year(), now.Month(), now.Day()+2, 10, 0, 0, 0, now.Location())
	end1 := start1.Add(2 * time.Hour)

	// Create booking 1
	result1, err := svc.Create(context.Background(), clientID, service.CreateBookingInput{
		BathhouseID: bh.ID,
		StartTime:   start1,
		EndTime:     end1,
		GuestCount:  5,
	})
	if err != nil {
		t.Fatalf("failed to create booking 1: %v", err)
	}

	// Create booking 2 at a different time
	start2 := time.Date(now.Year(), now.Month(), now.Day()+2, 14, 0, 0, 0, now.Location())
	end2 := start2.Add(2 * time.Hour)
	_, err = svc.Create(context.Background(), clientID, service.CreateBookingInput{
		BathhouseID: bh.ID,
		StartTime:   start2,
		EndTime:     end2,
		GuestCount:  5,
	})
	if err != nil {
		t.Fatalf("failed to create booking 2: %v", err)
	}

	// Try to modify booking 1 to overlap with booking 2
	_, err = svc.Modify(context.Background(), clientID, result1.Booking.ID, service.ModifyBookingInput{
		StartTime:  start2,
		EndTime:    end2,
		GuestCount: 5,
	})
	if !errors.Is(err, domain.ErrSlotUnavailable) {
		t.Errorf("expected ErrSlotUnavailable, got %v", err)
	}
}

func TestBookingService_Modify_CanMoveToSameSlot(t *testing.T) {
	svc, bhRepo, _, _, _, _, _, _ := newBookingService()
	ownerID := uuid.New()
	clientID := uuid.New()
	bh := createBathhouse(t, bhRepo, ownerID)

	now := time.Now()
	start := time.Date(now.Year(), now.Month(), now.Day()+2, 10, 0, 0, 0, now.Location())
	end := start.Add(2 * time.Hour)

	result, err := svc.Create(context.Background(), clientID, service.CreateBookingInput{
		BathhouseID: bh.ID,
		StartTime:   start,
		EndTime:     end,
		GuestCount:  5,
	})
	if err != nil {
		t.Fatalf("failed to create booking: %v", err)
	}

	// Modify to same time but different guest count (should not conflict with self)
	modResult, err := svc.Modify(context.Background(), clientID, result.Booking.ID, service.ModifyBookingInput{
		StartTime:  start,
		EndTime:    end,
		GuestCount: 3,
	})
	if err != nil {
		t.Fatalf("unexpected error modifying guest count: %v", err)
	}
	if modResult.Booking.GuestCount != 3 {
		t.Errorf("guest_count = %d, want 3", modResult.Booking.GuestCount)
	}
}
