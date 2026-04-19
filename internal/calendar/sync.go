package calendar

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/rekurt/relax-hub/internal/domain"
	"github.com/rekurt/relax-hub/internal/logger"
	"github.com/rekurt/relax-hub/internal/repository"
)

// HTTPClient is an interface for making HTTP requests, allowing test injection.
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// CalendarConflict represents a conflict between an external calendar event and an existing booking.
type CalendarConflict struct {
	SlotBlock domain.SlotBlock
	Booking   domain.Booking
}

// CalendarSyncService handles importing and syncing external calendar feeds.
type CalendarSyncService struct {
	extCalRepo    repository.ExternalCalendarRepository
	slotBlockRepo repository.SlotBlockRepository
	bookingRepo   repository.BookingRepository
	httpClient    HTTPClient
	log           *logger.Logger
}

// NewCalendarSyncService creates a new CalendarSyncService.
func NewCalendarSyncService(
	extCalRepo repository.ExternalCalendarRepository,
	slotBlockRepo repository.SlotBlockRepository,
	bookingRepo repository.BookingRepository,
	log *logger.Logger,
) *CalendarSyncService {
	return &CalendarSyncService{
		extCalRepo:    extCalRepo,
		slotBlockRepo: slotBlockRepo,
		bookingRepo:   bookingRepo,
		httpClient:    &http.Client{Timeout: 30 * time.Second},
		log:           log,
	}
}

// SetHTTPClient replaces the default HTTP client (used in tests).
func (s *CalendarSyncService) SetHTTPClient(c HTTPClient) {
	s.httpClient = c
}

// AddExternalCalendar subscribes a bathhouse to an external iCal feed.
func (s *CalendarSyncService) AddExternalCalendar(ctx context.Context, bathhouseID uuid.UUID, calendarURL string, source domain.SlotBlockSource) (*domain.ExternalCalendar, error) {
	cal := &domain.ExternalCalendar{
		BathhouseID: bathhouseID,
		URL:         calendarURL,
		Source:      source,
	}
	if err := cal.Validate(); err != nil {
		return nil, err
	}

	if err := s.extCalRepo.Create(ctx, cal); err != nil {
		return nil, fmt.Errorf("create external calendar: %w", err)
	}

	return cal, nil
}

// RemoveExternalCalendar removes a calendar subscription and its slot blocks.
func (s *CalendarSyncService) RemoveExternalCalendar(ctx context.Context, calendarID uuid.UUID) error {
	cal, err := s.extCalRepo.GetByID(ctx, calendarID)
	if err != nil {
		return err
	}

	// Delete all slot blocks imported from this calendar source for the bathhouse
	if err := s.slotBlockRepo.DeleteBySource(ctx, cal.BathhouseID, cal.Source); err != nil {
		return fmt.Errorf("delete slot blocks by source: %w", err)
	}

	return s.extCalRepo.Delete(ctx, calendarID)
}

// GetConflicts checks all external slot blocks against existing confirmed bookings for a bathhouse.
func (s *CalendarSyncService) GetConflicts(ctx context.Context, bathhouseID uuid.UUID) ([]CalendarConflict, error) {
	blocks, err := s.slotBlockRepo.ListByBathhouse(ctx, bathhouseID)
	if err != nil {
		return nil, fmt.Errorf("list slot blocks: %w", err)
	}

	var conflicts []CalendarConflict
	for _, block := range blocks {
		if block.Source == domain.SlotBlockSourceManual {
			continue
		}
		overlapping, err := s.bookingRepo.GetOverlapping(ctx, bathhouseID, block.StartTime, block.EndTime)
		if err != nil {
			return nil, fmt.Errorf("check booking overlap: %w", err)
		}
		for _, booking := range overlapping {
			if booking.Status == domain.BookingConfirmed || booking.Status == domain.BookingPendingOwner {
				conflicts = append(conflicts, CalendarConflict{
					SlotBlock: block,
					Booking:   booking,
				})
			}
		}
	}

	return conflicts, nil
}

// SyncCalendar fetches and syncs a single external calendar, returning any conflicts found.
func (s *CalendarSyncService) SyncCalendar(ctx context.Context, calendarID uuid.UUID) ([]CalendarConflict, error) {
	cal, err := s.extCalRepo.GetByID(ctx, calendarID)
	if err != nil {
		return nil, err
	}

	return s.syncOne(ctx, cal)
}

// SyncAllCalendars syncs all registered external calendars.
func (s *CalendarSyncService) SyncAllCalendars(ctx context.Context) {
	calendars, err := s.extCalRepo.ListAll(ctx)
	if err != nil {
		s.log.Error("Failed to list external calendars for sync", "error", err)
		return
	}

	for i := range calendars {
		conflicts, err := s.syncOne(ctx, &calendars[i])
		if err != nil {
			s.log.Error("Calendar sync failed",
				"calendar_id", calendars[i].ID,
				"bathhouse_id", calendars[i].BathhouseID,
				"error", err,
			)
		}
		if len(conflicts) > 0 {
			s.log.Warn("Calendar sync detected booking conflicts",
				"calendar_id", calendars[i].ID,
				"bathhouse_id", calendars[i].BathhouseID,
				"conflict_count", len(conflicts),
			)
		}
	}
}

func (s *CalendarSyncService) syncOne(ctx context.Context, cal *domain.ExternalCalendar) ([]CalendarConflict, error) {
	data, err := s.fetchFeed(ctx, cal.URL)
	now := time.Now()
	if err != nil {
		_ = s.extCalRepo.UpdateSyncStatus(ctx, cal.ID, now, err.Error())
		return nil, fmt.Errorf("fetch feed: %w", err)
	}

	events, err := ParseICal(data)
	if err != nil {
		_ = s.extCalRepo.UpdateSyncStatus(ctx, cal.ID, now, err.Error())
		return nil, fmt.Errorf("parse ical: %w", err)
	}

	// Delete old blocks from this source, then re-create from current feed
	if err := s.slotBlockRepo.DeleteBySource(ctx, cal.BathhouseID, cal.Source); err != nil {
		_ = s.extCalRepo.UpdateSyncStatus(ctx, cal.ID, now, err.Error())
		return nil, fmt.Errorf("delete old blocks: %w", err)
	}

	var conflicts []CalendarConflict
	cutoff := time.Now().Add(-24 * time.Hour)
	for _, ev := range events {
		// Skip past events (ended more than 24h ago)
		if ev.End.Before(cutoff) {
			continue
		}

		block := &domain.SlotBlock{
			BathhouseID: cal.BathhouseID,
			StartTime:   ev.Start,
			EndTime:     ev.End,
			Source:      cal.Source,
			ExternalID:  ev.UID,
			Description: ev.Summary,
		}
		if err := s.slotBlockRepo.Create(ctx, block); err != nil {
			s.log.Error("Failed to create slot block from calendar event",
				"calendar_id", cal.ID,
				"event_uid", ev.UID,
				"error", err,
			)
			continue
		}

		// Check for conflicts with existing confirmed bookings
		overlapping, err := s.bookingRepo.GetOverlapping(ctx, cal.BathhouseID, ev.Start, ev.End)
		if err != nil {
			s.log.Error("Failed to check booking conflicts",
				"calendar_id", cal.ID,
				"event_uid", ev.UID,
				"error", err,
			)
			continue
		}
		for _, booking := range overlapping {
			if booking.Status == domain.BookingConfirmed || booking.Status == domain.BookingPendingOwner {
				conflicts = append(conflicts, CalendarConflict{
					SlotBlock: *block,
					Booking:   booking,
				})
				s.log.Warn("External calendar event conflicts with existing booking",
					"calendar_id", cal.ID,
					"event_uid", ev.UID,
					"event_summary", ev.Summary,
					"booking_id", booking.ID,
					"booking_start", booking.StartTime,
					"booking_end", booking.EndTime,
				)
			}
		}
	}

	_ = s.extCalRepo.UpdateSyncStatus(ctx, cal.ID, now, "")
	return conflicts, nil
}

func (s *CalendarSyncService) fetchFeed(ctx context.Context, url string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 5*1024*1024)) // 5MB limit
	if err != nil {
		return "", fmt.Errorf("read body: %w", err)
	}

	return string(body), nil
}
