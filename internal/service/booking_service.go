package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/repository"
)

const cancelDeadline = 2 * time.Hour

type CreateBookingInput struct {
	BathhouseID uuid.UUID
	StartTime   time.Time
	EndTime     time.Time
	GuestCount  int
	Comment     string
}

type TimeSlot struct {
	StartTime time.Time
	EndTime   time.Time
	Available bool
}

type BookingService interface {
	Create(ctx context.Context, userID uuid.UUID, input CreateBookingInput) (*domain.Booking, error)
	Cancel(ctx context.Context, userID uuid.UUID, role domain.UserRole, bookingID uuid.UUID) error
	Confirm(ctx context.Context, userID uuid.UUID, role domain.UserRole, bookingID uuid.UUID) error
	Reject(ctx context.Context, userID uuid.UUID, role domain.UserRole, bookingID uuid.UUID) error
	Complete(ctx context.Context, userID uuid.UUID, role domain.UserRole, bookingID uuid.UUID) error
	ListByUser(ctx context.Context, userID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.Booking], error)
	ListByBathhouse(ctx context.Context, userID uuid.UUID, role domain.UserRole, bathhouseID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.Booking], error)
	GetAvailableSlots(ctx context.Context, bathhouseID uuid.UUID, date time.Time) ([]TimeSlot, error)
}

type bookingService struct {
	bookingRepo repository.BookingRepository
	bhRepo      repository.BathhouseRepository
	access      *AccessChecker
}

func NewBookingService(
	bookingRepo repository.BookingRepository,
	bhRepo repository.BathhouseRepository,
	access *AccessChecker,
) BookingService {
	return &bookingService{
		bookingRepo: bookingRepo,
		bhRepo:      bhRepo,
		access:      access,
	}
}

func (s *bookingService) Create(ctx context.Context, userID uuid.UUID, input CreateBookingInput) (*domain.Booking, error) {
	bh, err := s.bhRepo.GetByID(ctx, input.BathhouseID)
	if err != nil {
		return nil, err
	}

	if bh.Status != domain.BathhouseStatusActive {
		return nil, domain.ErrBathhouseNotActive
	}

	if input.StartTime.Before(time.Now()) {
		return nil, fmt.Errorf("%w: start time must be in the future", domain.ErrInvalidInput)
	}

	if input.GuestCount <= 0 || input.GuestCount > bh.MaxGuests {
		return nil, domain.ErrInvalidInput
	}

	duration := input.EndTime.Sub(input.StartTime)
	durationHours := int(duration / time.Hour)
	if duration%time.Hour != 0 || durationHours < bh.MinDuration {
		return nil, domain.ErrInvalidInput
	}

	// Validate booking falls within working hours
	if err := validateWithinWorkingHours(bh, input.StartTime, input.EndTime); err != nil {
		return nil, err
	}

	available, err := s.bookingRepo.CheckAvailability(ctx, input.BathhouseID, input.StartTime, input.EndTime)
	if err != nil {
		return nil, err
	}
	if !available {
		return nil, domain.ErrSlotUnavailable
	}

	totalPrice := bh.PricePerHour * int64(durationHours)

	now := time.Now()
	booking := &domain.Booking{
		ID:          uuid.New(),
		UserID:      userID,
		BathhouseID: input.BathhouseID,
		StartTime:   input.StartTime,
		EndTime:     input.EndTime,
		GuestCount:  input.GuestCount,
		TotalPrice:  totalPrice,
		Status:      domain.BookingPending,
		Comment:     input.Comment,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := booking.Validate(); err != nil {
		return nil, err
	}

	if err := s.bookingRepo.Create(ctx, booking); err != nil {
		return nil, err
	}

	return booking, nil
}

func (s *bookingService) Cancel(ctx context.Context, userID uuid.UUID, role domain.UserRole, bookingID uuid.UUID) error {
	booking, err := s.bookingRepo.GetByID(ctx, bookingID)
	if err != nil {
		return err
	}

	if booking.Status != domain.BookingPending && booking.Status != domain.BookingConfirmed {
		return domain.ErrInvalidInput
	}

	// Client cancels their own booking (at least 2 hours before start)
	if role == domain.RoleClient {
		if booking.UserID != userID {
			return domain.ErrForbidden
		}
		if time.Until(booking.StartTime) < cancelDeadline {
			return domain.ErrBookingCancelLate
		}
		return s.bookingRepo.UpdateStatus(ctx, bookingID, domain.BookingCancelled)
	}

	// Owner/representative cancels booking for their bathhouse
	if err := s.access.CanManageBathhouse(ctx, userID, role, booking.BathhouseID); err != nil {
		return err
	}

	return s.bookingRepo.UpdateStatus(ctx, bookingID, domain.BookingCancelled)
}

func (s *bookingService) Confirm(ctx context.Context, userID uuid.UUID, role domain.UserRole, bookingID uuid.UUID) error {
	booking, err := s.bookingRepo.GetByID(ctx, bookingID)
	if err != nil {
		return err
	}

	if booking.Status != domain.BookingPending {
		return domain.ErrInvalidInput
	}

	if err := s.access.CanManageBathhouse(ctx, userID, role, booking.BathhouseID); err != nil {
		return err
	}

	return s.bookingRepo.UpdateStatus(ctx, bookingID, domain.BookingConfirmed)
}

func (s *bookingService) Reject(ctx context.Context, userID uuid.UUID, role domain.UserRole, bookingID uuid.UUID) error {
	booking, err := s.bookingRepo.GetByID(ctx, bookingID)
	if err != nil {
		return err
	}

	if booking.Status != domain.BookingPending {
		return domain.ErrInvalidInput
	}

	if err := s.access.CanManageBathhouse(ctx, userID, role, booking.BathhouseID); err != nil {
		return err
	}

	return s.bookingRepo.UpdateStatus(ctx, bookingID, domain.BookingRejected)
}

func (s *bookingService) Complete(ctx context.Context, userID uuid.UUID, role domain.UserRole, bookingID uuid.UUID) error {
	booking, err := s.bookingRepo.GetByID(ctx, bookingID)
	if err != nil {
		return err
	}

	if booking.Status != domain.BookingConfirmed {
		return fmt.Errorf("%w: only confirmed bookings can be completed", domain.ErrInvalidInput)
	}

	if time.Now().Before(booking.EndTime) {
		return fmt.Errorf("%w: booking can only be completed after end time", domain.ErrInvalidInput)
	}

	if err := s.access.CanManageBathhouse(ctx, userID, role, booking.BathhouseID); err != nil {
		return err
	}

	return s.bookingRepo.UpdateStatus(ctx, bookingID, domain.BookingCompleted)
}

func (s *bookingService) ListByUser(ctx context.Context, userID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.Booking], error) {
	return s.bookingRepo.ListByUser(ctx, userID, page, pageSize)
}

func (s *bookingService) ListByBathhouse(ctx context.Context, userID uuid.UUID, role domain.UserRole, bathhouseID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.Booking], error) {
	if err := s.access.CanViewBathhouseBookings(ctx, userID, role, bathhouseID); err != nil {
		return nil, err
	}
	return s.bookingRepo.ListByBathhouse(ctx, bathhouseID, page, pageSize)
}

func (s *bookingService) GetAvailableSlots(ctx context.Context, bathhouseID uuid.UUID, date time.Time) ([]TimeSlot, error) {
	bh, err := s.bhRepo.GetByID(ctx, bathhouseID)
	if err != nil {
		return nil, err
	}

	dayOfWeek := toDayOfWeek(date.Weekday())

	var wh *domain.WorkingHours
	for i := range bh.WorkingHours {
		if bh.WorkingHours[i].DayOfWeek == dayOfWeek {
			wh = &bh.WorkingHours[i]
			break
		}
	}

	if wh == nil {
		return nil, nil // closed on this day
	}

	openHour, openMin, err := parseTime(wh.OpenTime)
	if err != nil {
		return nil, fmt.Errorf("invalid open time: %w", err)
	}
	closeHour, closeMin, err := parseTime(wh.CloseTime)
	if err != nil {
		return nil, fmt.Errorf("invalid close time: %w", err)
	}

	dayStart := time.Date(date.Year(), date.Month(), date.Day(), openHour, openMin, 0, 0, date.Location())
	dayEnd := time.Date(date.Year(), date.Month(), date.Day(), closeHour, closeMin, 0, 0, date.Location())

	// Handle overnight working hours (e.g., 18:00 - 06:00)
	if !dayEnd.After(dayStart) {
		dayEnd = dayEnd.Add(24 * time.Hour)
	}

	overlapping, err := s.bookingRepo.GetOverlapping(ctx, bathhouseID, dayStart, dayEnd)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	var slots []TimeSlot
	for t := dayStart; t.Before(dayEnd); t = t.Add(time.Hour) {
		slotEnd := t.Add(time.Hour)
		if slotEnd.After(dayEnd) {
			break
		}
		// Skip slots that have already passed
		if slotEnd.Before(now) {
			continue
		}
		avail := true
		for _, b := range overlapping {
			if t.Before(b.EndTime) && slotEnd.After(b.StartTime) {
				avail = false
				break
			}
		}
		slots = append(slots, TimeSlot{
			StartTime: t,
			EndTime:   slotEnd,
			Available: avail,
		})
	}

	return slots, nil
}

func parseTime(s string) (int, int, error) {
	var h, m int
	n, _ := fmt.Sscanf(s, "%d:%d", &h, &m)
	if n != 2 {
		return 0, 0, fmt.Errorf("invalid time format: %q", s)
	}
	return h, m, nil
}

func toDayOfWeek(wd time.Weekday) int {
	if wd == time.Sunday {
		return 6
	}
	return int(wd) - 1
}

func validateWithinWorkingHours(bh *domain.Bathhouse, startTime, endTime time.Time) error {
	dayOfWeek := toDayOfWeek(startTime.Weekday())

	var wh *domain.WorkingHours
	for i := range bh.WorkingHours {
		if bh.WorkingHours[i].DayOfWeek == dayOfWeek {
			wh = &bh.WorkingHours[i]
			break
		}
	}

	if wh == nil {
		return fmt.Errorf("%w: bathhouse is closed on this day", domain.ErrInvalidInput)
	}

	openH, openM, err := parseTime(wh.OpenTime)
	if err != nil {
		return fmt.Errorf("invalid open time: %w", err)
	}
	closeH, closeM, err := parseTime(wh.CloseTime)
	if err != nil {
		return fmt.Errorf("invalid close time: %w", err)
	}

	loc := startTime.Location()
	dayOpen := time.Date(startTime.Year(), startTime.Month(), startTime.Day(), openH, openM, 0, 0, loc)
	dayClose := time.Date(startTime.Year(), startTime.Month(), startTime.Day(), closeH, closeM, 0, 0, loc)

	// Handle overnight working hours (e.g., 18:00 - 06:00)
	if !dayClose.After(dayOpen) {
		dayClose = dayClose.Add(24 * time.Hour)
	}

	if startTime.Before(dayOpen) || endTime.After(dayClose) {
		return fmt.Errorf("%w: booking must be within working hours (%s-%s)", domain.ErrInvalidInput, wh.OpenTime, wh.CloseTime)
	}

	return nil
}
