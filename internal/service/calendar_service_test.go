package service_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/logger"
	"github.com/nikitaaldaev/bani/internal/repository/mock"
	"github.com/nikitaaldaev/bani/internal/service"
)

func newCalendarService() (service.CalendarService, *mock.BathhouseRepo, *mock.BookingRepo, *mock.RepresentativeRepo) {
	bhRepo := mock.NewBathhouseRepo()
	bookingRepo := mock.NewBookingRepo()
	repRepo := mock.NewRepresentativeRepo()
	access := service.NewAccessChecker(repRepo, bhRepo)
	log := logger.New(logger.LevelWarn)
	svc := service.NewCalendarService(bhRepo, bookingRepo, access, log)
	return svc, bhRepo, bookingRepo, repRepo
}

func TestCalendarService_ExportICal_Success(t *testing.T) {
	svc, bhRepo, bookingRepo, _ := newCalendarService()
	ownerID := uuid.New()

	bh := &domain.Bathhouse{
		ID: uuid.New(), OwnerID: ownerID, Name: "Баня Тест",
		Address: "Test St", CityID: 1, PricePerHour: 5000,
		MinDuration: 1, MaxGuests: 10, Status: domain.BathhouseStatusActive,
		WorkingHours: []domain.WorkingHours{
			{DayOfWeek: 0, OpenTime: "08:00", CloseTime: "22:00"},
			{DayOfWeek: 1, OpenTime: "08:00", CloseTime: "22:00"},
			{DayOfWeek: 2, OpenTime: "08:00", CloseTime: "22:00"},
			{DayOfWeek: 3, OpenTime: "08:00", CloseTime: "22:00"},
			{DayOfWeek: 4, OpenTime: "08:00", CloseTime: "22:00"},
			{DayOfWeek: 5, OpenTime: "08:00", CloseTime: "22:00"},
			{DayOfWeek: 6, OpenTime: "08:00", CloseTime: "22:00"},
		},
	}
	_ = bhRepo.Create(context.Background(), bh)

	start := time.Now().Add(24 * time.Hour).Truncate(time.Hour)
	booking := &domain.Booking{
		ID: uuid.New(), UserID: uuid.New(), BathhouseID: bh.ID,
		StartTime: start, EndTime: start.Add(2 * time.Hour),
		GuestCount: 4, TotalPrice: 10000, Status: domain.BookingConfirmed,
	}
	_ = bookingRepo.Create(context.Background(), booking)

	result, err := svc.ExportICal(context.Background(), ownerID, domain.RoleOwner, bh.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result, "BEGIN:VCALENDAR") {
		t.Error("expected iCal format")
	}
	if !strings.Contains(result, "BEGIN:VEVENT") {
		t.Error("expected VEVENT for booking")
	}
	if !strings.Contains(result, booking.ID.String()) {
		t.Error("expected booking ID in UID")
	}
}

func TestCalendarService_ExportICal_Forbidden(t *testing.T) {
	svc, bhRepo, _, _ := newCalendarService()
	ownerID := uuid.New()
	otherUserID := uuid.New()

	bh := &domain.Bathhouse{
		ID: uuid.New(), OwnerID: ownerID, Name: "Баня",
		Address: "St", CityID: 1, PricePerHour: 5000,
		MinDuration: 1, MaxGuests: 10, Status: domain.BathhouseStatusActive,
		WorkingHours: []domain.WorkingHours{
			{DayOfWeek: 0, OpenTime: "08:00", CloseTime: "22:00"},
		},
	}
	_ = bhRepo.Create(context.Background(), bh)

	_, err := svc.ExportICal(context.Background(), otherUserID, domain.RoleOwner, bh.ID)
	if err == nil {
		t.Error("expected forbidden error for non-owner")
	}
}

func TestCalendarService_ExportICalByToken(t *testing.T) {
	svc, bhRepo, bookingRepo, _ := newCalendarService()
	ownerID := uuid.New()

	bh := &domain.Bathhouse{
		ID: uuid.New(), OwnerID: ownerID, Name: "Токен Баня",
		Address: "St", CityID: 1, PricePerHour: 5000,
		MinDuration: 1, MaxGuests: 10, Status: domain.BathhouseStatusActive,
		CalendarToken: "test-secret-token-123",
		WorkingHours: []domain.WorkingHours{
			{DayOfWeek: 0, OpenTime: "08:00", CloseTime: "22:00"},
			{DayOfWeek: 1, OpenTime: "08:00", CloseTime: "22:00"},
			{DayOfWeek: 2, OpenTime: "08:00", CloseTime: "22:00"},
			{DayOfWeek: 3, OpenTime: "08:00", CloseTime: "22:00"},
			{DayOfWeek: 4, OpenTime: "08:00", CloseTime: "22:00"},
			{DayOfWeek: 5, OpenTime: "08:00", CloseTime: "22:00"},
			{DayOfWeek: 6, OpenTime: "08:00", CloseTime: "22:00"},
		},
	}
	_ = bhRepo.Create(context.Background(), bh)

	start := time.Now().Add(24 * time.Hour).Truncate(time.Hour)
	booking := &domain.Booking{
		ID: uuid.New(), UserID: uuid.New(), BathhouseID: bh.ID,
		StartTime: start, EndTime: start.Add(2 * time.Hour),
		GuestCount: 2, TotalPrice: 10000, Status: domain.BookingPending,
	}
	_ = bookingRepo.Create(context.Background(), booking)

	result, err := svc.ExportICalByToken(context.Background(), "test-secret-token-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(result, "BEGIN:VCALENDAR") {
		t.Error("expected iCal format")
	}
	if !strings.Contains(result, "TENTATIVE") {
		t.Error("expected TENTATIVE status for pending booking")
	}
}

func TestCalendarService_ExportICalByToken_NotFound(t *testing.T) {
	svc, _, _, _ := newCalendarService()

	_, err := svc.ExportICalByToken(context.Background(), "nonexistent-token")
	if err == nil {
		t.Error("expected error for nonexistent token")
	}
}

func TestCalendarService_GetOrCreateCalendarToken(t *testing.T) {
	svc, bhRepo, _, _ := newCalendarService()
	ownerID := uuid.New()

	bh := &domain.Bathhouse{
		ID: uuid.New(), OwnerID: ownerID, Name: "Баня",
		Address: "St", CityID: 1, PricePerHour: 5000,
		MinDuration: 1, MaxGuests: 10, Status: domain.BathhouseStatusActive,
		WorkingHours: []domain.WorkingHours{
			{DayOfWeek: 0, OpenTime: "08:00", CloseTime: "22:00"},
		},
	}
	_ = bhRepo.Create(context.Background(), bh)

	token1, err := svc.GetOrCreateCalendarToken(context.Background(), ownerID, domain.RoleOwner, bh.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token1 == "" {
		t.Error("expected non-empty token")
	}

	// Second call should return the same token
	token2, err := svc.GetOrCreateCalendarToken(context.Background(), ownerID, domain.RoleOwner, bh.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token1 != token2 {
		t.Errorf("expected same token, got %q and %q", token1, token2)
	}
}

func TestCalendarService_ExportICal_FiltersCancelled(t *testing.T) {
	svc, bhRepo, bookingRepo, _ := newCalendarService()
	ownerID := uuid.New()

	bh := &domain.Bathhouse{
		ID: uuid.New(), OwnerID: ownerID, Name: "Фильтр",
		Address: "St", CityID: 1, PricePerHour: 5000,
		MinDuration: 1, MaxGuests: 10, Status: domain.BathhouseStatusActive,
		WorkingHours: []domain.WorkingHours{
			{DayOfWeek: 0, OpenTime: "08:00", CloseTime: "22:00"},
			{DayOfWeek: 1, OpenTime: "08:00", CloseTime: "22:00"},
			{DayOfWeek: 2, OpenTime: "08:00", CloseTime: "22:00"},
			{DayOfWeek: 3, OpenTime: "08:00", CloseTime: "22:00"},
			{DayOfWeek: 4, OpenTime: "08:00", CloseTime: "22:00"},
			{DayOfWeek: 5, OpenTime: "08:00", CloseTime: "22:00"},
			{DayOfWeek: 6, OpenTime: "08:00", CloseTime: "22:00"},
		},
	}
	_ = bhRepo.Create(context.Background(), bh)

	start := time.Now().Add(24 * time.Hour).Truncate(time.Hour)

	// Create confirmed booking
	confirmedID := uuid.New()
	_ = bookingRepo.Create(context.Background(), &domain.Booking{
		ID: confirmedID, UserID: uuid.New(), BathhouseID: bh.ID,
		StartTime: start, EndTime: start.Add(time.Hour),
		GuestCount: 2, TotalPrice: 5000, Status: domain.BookingConfirmed,
	})

	// Create cancelled booking
	cancelledID := uuid.New()
	cancelled := &domain.Booking{
		ID: cancelledID, UserID: uuid.New(), BathhouseID: bh.ID,
		StartTime: start.Add(3 * time.Hour), EndTime: start.Add(4 * time.Hour),
		GuestCount: 2, TotalPrice: 5000, Status: domain.BookingPending,
	}
	_ = bookingRepo.Create(context.Background(), cancelled)
	_ = bookingRepo.UpdateStatus(context.Background(), cancelledID, domain.BookingCancelled)

	result, err := svc.ExportICal(context.Background(), ownerID, domain.RoleOwner, bh.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result, confirmedID.String()) {
		t.Error("expected confirmed booking in output")
	}
	if strings.Contains(result, cancelledID.String()) {
		t.Error("expected cancelled booking to be filtered out")
	}
}
