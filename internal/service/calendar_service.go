package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"

	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/calendar"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/logger"
	"github.com/nikitaaldaev/bani/internal/repository"
)

type CalendarService interface {
	ExportICal(ctx context.Context, userID uuid.UUID, userRole domain.UserRole, bathhouseID uuid.UUID) (string, error)
	ExportICalByToken(ctx context.Context, token string) (string, error)
	GetOrCreateCalendarToken(ctx context.Context, userID uuid.UUID, userRole domain.UserRole, bathhouseID uuid.UUID) (string, error)
}

type calendarService struct {
	bhRepo      repository.BathhouseRepository
	bookingRepo repository.BookingRepository
	access      *AccessChecker
	log         *logger.Logger
}

func NewCalendarService(
	bhRepo repository.BathhouseRepository,
	bookingRepo repository.BookingRepository,
	access *AccessChecker,
	log *logger.Logger,
) CalendarService {
	return &calendarService{
		bhRepo:      bhRepo,
		bookingRepo: bookingRepo,
		access:      access,
		log:         log,
	}
}

func (s *calendarService) ExportICal(ctx context.Context, userID uuid.UUID, userRole domain.UserRole, bathhouseID uuid.UUID) (string, error) {
	if err := s.access.CanManageBathhouse(ctx, userID, userRole, bathhouseID); err != nil {
		return "", err
	}

	bh, err := s.bhRepo.GetByID(ctx, bathhouseID)
	if err != nil {
		return "", err
	}

	return s.generateICalForBathhouse(ctx, bh)
}

func (s *calendarService) ExportICalByToken(ctx context.Context, token string) (string, error) {
	if token == "" {
		return "", domain.ErrNotFound
	}

	bh, err := s.bhRepo.GetByCalendarToken(ctx, token)
	if err != nil {
		return "", err
	}

	return s.generateICalForBathhouse(ctx, bh)
}

func (s *calendarService) GetOrCreateCalendarToken(ctx context.Context, userID uuid.UUID, userRole domain.UserRole, bathhouseID uuid.UUID) (string, error) {
	if err := s.access.CanManageBathhouse(ctx, userID, userRole, bathhouseID); err != nil {
		return "", err
	}

	token, err := s.bhRepo.GetCalendarToken(ctx, bathhouseID)
	if err != nil {
		return "", err
	}

	if token != "" {
		return token, nil
	}

	token, err = generateSecureToken()
	if err != nil {
		return "", fmt.Errorf("generate calendar token: %w", err)
	}

	if err := s.bhRepo.SetCalendarToken(ctx, bathhouseID, token); err != nil {
		return "", err
	}

	return token, nil
}

func (s *calendarService) generateICalForBathhouse(ctx context.Context, bh *domain.Bathhouse) (string, error) {
	// Fetch all non-cancelled bookings (get a large page to include all)
	result, err := s.bookingRepo.ListByBathhouse(ctx, bh.ID, 1, 1000)
	if err != nil {
		return "", fmt.Errorf("list bookings: %w", err)
	}

	// Filter out cancelled/rejected bookings
	var activeBookings []domain.Booking
	for _, b := range result.Items {
		if b.Status != domain.BookingCancelled && b.Status != domain.BookingRejected {
			activeBookings = append(activeBookings, b)
		}
	}

	return calendar.GenerateICal(bh, activeBookings), nil
}

func generateSecureToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
