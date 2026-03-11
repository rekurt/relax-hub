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
	AddExternalCalendar(ctx context.Context, userID uuid.UUID, userRole domain.UserRole, bathhouseID uuid.UUID, calendarURL string, source domain.SlotBlockSource) (*domain.ExternalCalendar, error)
	ListExternalCalendars(ctx context.Context, userID uuid.UUID, userRole domain.UserRole, bathhouseID uuid.UUID) ([]domain.ExternalCalendar, error)
	RemoveExternalCalendar(ctx context.Context, userID uuid.UUID, userRole domain.UserRole, calendarID uuid.UUID) error
	SyncExternalCalendars(ctx context.Context, userID uuid.UUID, userRole domain.UserRole, bathhouseID uuid.UUID) error
	CreateSlotBlock(ctx context.Context, userID uuid.UUID, userRole domain.UserRole, bathhouseID uuid.UUID, block *domain.SlotBlock) (*domain.SlotBlock, error)
	DeleteSlotBlock(ctx context.Context, userID uuid.UUID, userRole domain.UserRole, blockID uuid.UUID) error
}

type calendarService struct {
	bhRepo        repository.BathhouseRepository
	bookingRepo   repository.BookingRepository
	extCalRepo    repository.ExternalCalendarRepository
	slotBlockRepo repository.SlotBlockRepository
	syncSvc       *calendar.CalendarSyncService
	access        *AccessChecker
	log           *logger.Logger
}

func NewCalendarService(
	bhRepo repository.BathhouseRepository,
	bookingRepo repository.BookingRepository,
	extCalRepo repository.ExternalCalendarRepository,
	slotBlockRepo repository.SlotBlockRepository,
	syncSvc *calendar.CalendarSyncService,
	access *AccessChecker,
	log *logger.Logger,
) CalendarService {
	return &calendarService{
		bhRepo:        bhRepo,
		bookingRepo:   bookingRepo,
		extCalRepo:    extCalRepo,
		slotBlockRepo: slotBlockRepo,
		syncSvc:       syncSvc,
		access:        access,
		log:           log,
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

func (s *calendarService) AddExternalCalendar(ctx context.Context, userID uuid.UUID, userRole domain.UserRole, bathhouseID uuid.UUID, calendarURL string, source domain.SlotBlockSource) (*domain.ExternalCalendar, error) {
	if err := s.access.CanManageBathhouse(ctx, userID, userRole, bathhouseID); err != nil {
		return nil, err
	}
	return s.syncSvc.AddExternalCalendar(ctx, bathhouseID, calendarURL, source)
}

func (s *calendarService) ListExternalCalendars(ctx context.Context, userID uuid.UUID, userRole domain.UserRole, bathhouseID uuid.UUID) ([]domain.ExternalCalendar, error) {
	if err := s.access.CanManageBathhouse(ctx, userID, userRole, bathhouseID); err != nil {
		return nil, err
	}
	return s.extCalRepo.ListByBathhouse(ctx, bathhouseID)
}

func (s *calendarService) RemoveExternalCalendar(ctx context.Context, userID uuid.UUID, userRole domain.UserRole, calendarID uuid.UUID) error {
	cal, err := s.extCalRepo.GetByID(ctx, calendarID)
	if err != nil {
		return err
	}
	if err := s.access.CanManageBathhouse(ctx, userID, userRole, cal.BathhouseID); err != nil {
		return err
	}
	return s.syncSvc.RemoveExternalCalendar(ctx, calendarID)
}

func (s *calendarService) SyncExternalCalendars(ctx context.Context, userID uuid.UUID, userRole domain.UserRole, bathhouseID uuid.UUID) error {
	if err := s.access.CanManageBathhouse(ctx, userID, userRole, bathhouseID); err != nil {
		return err
	}
	calendars, err := s.extCalRepo.ListByBathhouse(ctx, bathhouseID)
	if err != nil {
		return fmt.Errorf("list calendars: %w", err)
	}
	for _, cal := range calendars {
		if err := s.syncSvc.SyncCalendar(ctx, cal.ID); err != nil {
			s.log.Error("sync calendar failed", "calendar_id", cal.ID, "error", err)
		}
	}
	return nil
}

func (s *calendarService) CreateSlotBlock(ctx context.Context, userID uuid.UUID, userRole domain.UserRole, bathhouseID uuid.UUID, block *domain.SlotBlock) (*domain.SlotBlock, error) {
	if err := s.access.CanManageBathhouse(ctx, userID, userRole, bathhouseID); err != nil {
		return nil, err
	}
	block.BathhouseID = bathhouseID
	block.Source = domain.SlotBlockSourceManual
	if err := block.Validate(); err != nil {
		return nil, err
	}
	if err := s.slotBlockRepo.Create(ctx, block); err != nil {
		return nil, fmt.Errorf("create slot block: %w", err)
	}
	return block, nil
}

func (s *calendarService) DeleteSlotBlock(ctx context.Context, userID uuid.UUID, userRole domain.UserRole, blockID uuid.UUID) error {
	block, err := s.slotBlockRepo.GetByID(ctx, blockID)
	if err != nil {
		return err
	}
	if block.Source != domain.SlotBlockSourceManual {
		return domain.ErrForbidden
	}
	if err := s.access.CanManageBathhouse(ctx, userID, userRole, block.BathhouseID); err != nil {
		return err
	}
	return s.slotBlockRepo.Delete(ctx, blockID)
}

func generateSecureToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
