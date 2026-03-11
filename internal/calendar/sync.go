package calendar

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/logger"
	"github.com/nikitaaldaev/bani/internal/repository"
)

// HTTPClient is an interface for making HTTP requests, allowing test injection.
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// CalendarSyncService handles importing and syncing external calendar feeds.
type CalendarSyncService struct {
	extCalRepo    repository.ExternalCalendarRepository
	slotBlockRepo repository.SlotBlockRepository
	httpClient    HTTPClient
	log           *logger.Logger
}

// NewCalendarSyncService creates a new CalendarSyncService.
func NewCalendarSyncService(
	extCalRepo repository.ExternalCalendarRepository,
	slotBlockRepo repository.SlotBlockRepository,
	log *logger.Logger,
) *CalendarSyncService {
	return &CalendarSyncService{
		extCalRepo:    extCalRepo,
		slotBlockRepo: slotBlockRepo,
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

// SyncCalendar fetches and syncs a single external calendar.
func (s *CalendarSyncService) SyncCalendar(ctx context.Context, calendarID uuid.UUID) error {
	cal, err := s.extCalRepo.GetByID(ctx, calendarID)
	if err != nil {
		return err
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
		if err := s.syncOne(ctx, &calendars[i]); err != nil {
			s.log.Error("Calendar sync failed",
				"calendar_id", calendars[i].ID,
				"bathhouse_id", calendars[i].BathhouseID,
				"error", err,
			)
		}
	}
}

func (s *CalendarSyncService) syncOne(ctx context.Context, cal *domain.ExternalCalendar) error {
	data, err := s.fetchFeed(ctx, cal.URL)
	now := time.Now()
	if err != nil {
		_ = s.extCalRepo.UpdateSyncStatus(ctx, cal.ID, now, err.Error())
		return fmt.Errorf("fetch feed: %w", err)
	}

	events, err := ParseICal(data)
	if err != nil {
		_ = s.extCalRepo.UpdateSyncStatus(ctx, cal.ID, now, err.Error())
		return fmt.Errorf("parse ical: %w", err)
	}

	// Delete old blocks from this source, then re-create from current feed
	if err := s.slotBlockRepo.DeleteBySource(ctx, cal.BathhouseID, cal.Source); err != nil {
		_ = s.extCalRepo.UpdateSyncStatus(ctx, cal.ID, now, err.Error())
		return fmt.Errorf("delete old blocks: %w", err)
	}

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
		}
	}

	_ = s.extCalRepo.UpdateSyncStatus(ctx, cal.ID, now, "")
	return nil
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
