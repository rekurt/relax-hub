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

func newBookingService() (service.BookingService, *mock.BathhouseRepo, *mock.BookingRepo, *mock.RepresentativeRepo) {
	bhRepo := mock.NewBathhouseRepo()
	bookingRepo := mock.NewBookingRepo()
	repRepo := mock.NewRepresentativeRepo()
	access := service.NewAccessChecker(repRepo, bhRepo)
	svc := service.NewBookingService(bookingRepo, bhRepo, access)
	return svc, bhRepo, bookingRepo, repRepo
}

func TestBookingService_Create_Success(t *testing.T) {
	svc, bhRepo, _, _ := newBookingService()
	ownerID := uuid.New()
	clientID := uuid.New()
	bh := createBathhouse(t, bhRepo, ownerID)

	start := time.Now().Add(24 * time.Hour)
	end := start.Add(2 * time.Hour)

	booking, err := svc.Create(context.Background(), clientID, service.CreateBookingInput{
		BathhouseID: bh.ID,
		StartTime:   start,
		EndTime:     end,
		GuestCount:  5,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if booking.Status != domain.BookingPending {
		t.Errorf("status = %q, want %q", booking.Status, domain.BookingPending)
	}
	if booking.TotalPrice != bh.PricePerHour*2 {
		t.Errorf("totalPrice = %d, want %d", booking.TotalPrice, bh.PricePerHour*2)
	}
}

func TestBookingService_Create_InactiveBathhouse(t *testing.T) {
	svc, bhRepo, _, _ := newBookingService()
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
	svc, bhRepo, _, _ := newBookingService()
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
	svc, bhRepo, bookingRepo, _ := newBookingService()
	ownerID := uuid.New()
	bh := createBathhouse(t, bhRepo, ownerID)

	start := time.Now().Add(24 * time.Hour)
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
	svc, bhRepo, bookingRepo, _ := newBookingService()
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
	svc, bhRepo, bookingRepo, _ := newBookingService()
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
	svc, bhRepo, bookingRepo, _ := newBookingService()
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
	svc, bhRepo, bookingRepo, _ := newBookingService()
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
	svc, bhRepo, bookingRepo, _ := newBookingService()
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
	svc, bhRepo, bookingRepo, repRepo := newBookingService()
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
	svc, bhRepo, bookingRepo, _ := newBookingService()
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
	svc, bhRepo, bookingRepo, _ := newBookingService()
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
	svc, bhRepo, _, _ := newBookingService()
	ownerID := uuid.New()
	bh := createBathhouse(t, bhRepo, ownerID)

	_, err := svc.ListByBathhouse(context.Background(), ownerID, domain.RoleOwner, bh.ID, 1, 10)
	if err != nil {
		t.Errorf("owner should list bookings for own bathhouse: %v", err)
	}
}

func TestBookingService_ListByBathhouse_ClientForbidden(t *testing.T) {
	svc, bhRepo, _, _ := newBookingService()
	bh := createBathhouse(t, bhRepo, uuid.New())
	clientID := uuid.New()

	_, err := svc.ListByBathhouse(context.Background(), clientID, domain.RoleClient, bh.ID, 1, 10)
	if !errors.Is(err, domain.ErrForbidden) {
		t.Errorf("client should be forbidden from listing bathhouse bookings, got: %v", err)
	}
}

func TestBookingService_ListByUser(t *testing.T) {
	svc, bhRepo, bookingRepo, _ := newBookingService()
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
	svc, bhRepo, bookingRepo, _ := newBookingService()
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
	svc, bhRepo, _, _ := newBookingService()
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
	svc, bhRepo, bookingRepo, _ := newBookingService()
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
	svc, bhRepo, bookingRepo, _ := newBookingService()
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
	svc, bhRepo, _, _ := newBookingService()
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
