package calendar

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/rekurt/relax-hub/internal/domain"
	"github.com/rekurt/relax-hub/internal/logger"
	"github.com/rekurt/relax-hub/internal/repository/mock"
)

type mockHTTPClient struct {
	response *http.Response
	err      error
}

func (m *mockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	return m.response, m.err
}

func newTestSyncService() (*CalendarSyncService, *mock.ExternalCalendarRepo, *mock.SlotBlockRepo, *mock.BookingRepo) {
	extRepo := mock.NewExternalCalendarRepo()
	blockRepo := mock.NewSlotBlockRepo()
	bookingRepo := mock.NewBookingRepo()
	log := logger.New(logger.LevelDebug)

	svc := NewCalendarSyncService(extRepo, blockRepo, bookingRepo, log)
	return svc, extRepo, blockRepo, bookingRepo
}

func TestAddExternalCalendar(t *testing.T) {
	svc, extRepo, _, _ := newTestSyncService()
	ctx := context.Background()
	bathhouseID := uuid.New()

	cal, err := svc.AddExternalCalendar(ctx, bathhouseID, "https://calendar.google.com/test.ics", domain.SlotBlockSourceGoogleCalendar)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cal.ID == uuid.Nil {
		t.Error("expected calendar ID to be set")
	}
	if cal.BathhouseID != bathhouseID {
		t.Error("bathhouse ID mismatch")
	}

	// Verify it was saved
	saved, err := extRepo.GetByID(ctx, cal.ID)
	if err != nil {
		t.Fatalf("failed to get saved calendar: %v", err)
	}
	if saved.URL != "https://calendar.google.com/test.ics" {
		t.Errorf("URL mismatch: %s", saved.URL)
	}
}

func TestAddExternalCalendar_InvalidSource(t *testing.T) {
	svc, _, _, _ := newTestSyncService()
	ctx := context.Background()

	_, err := svc.AddExternalCalendar(ctx, uuid.New(), "https://example.com/cal.ics", domain.SlotBlockSourceManual)
	if err == nil {
		t.Error("expected error for manual source")
	}
}

func TestAddExternalCalendar_EmptyURL(t *testing.T) {
	svc, _, _, _ := newTestSyncService()
	ctx := context.Background()

	_, err := svc.AddExternalCalendar(ctx, uuid.New(), "", domain.SlotBlockSourceGoogleCalendar)
	if err == nil {
		t.Error("expected error for empty URL")
	}
}

func TestRemoveExternalCalendar(t *testing.T) {
	svc, extRepo, blockRepo, _ := newTestSyncService()
	ctx := context.Background()
	bathhouseID := uuid.New()

	// Add a calendar
	cal, _ := svc.AddExternalCalendar(ctx, bathhouseID, "https://example.com/cal.ics", domain.SlotBlockSourceGoogleCalendar)

	// Add some blocks from that source
	_ = blockRepo.Create(ctx, &domain.SlotBlock{
		BathhouseID: bathhouseID,
		StartTime:   time.Now(),
		EndTime:     time.Now().Add(2 * time.Hour),
		Source:      domain.SlotBlockSourceGoogleCalendar,
		ExternalID:  "ext1",
	})

	// Remove the calendar
	err := svc.RemoveExternalCalendar(ctx, cal.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify calendar was deleted
	_, err = extRepo.GetByID(ctx, cal.ID)
	if err != domain.ErrNotFound {
		t.Error("expected calendar to be deleted")
	}

	// Verify blocks from that source were deleted
	blocks, _ := blockRepo.ListByBathhouse(ctx, bathhouseID)
	for _, b := range blocks {
		if b.Source == domain.SlotBlockSourceGoogleCalendar {
			t.Error("expected google calendar blocks to be deleted")
		}
	}
}

func TestSyncCalendar(t *testing.T) {
	svc, extRepo, blockRepo, _ := newTestSyncService()
	ctx := context.Background()
	bathhouseID := uuid.New()

	// Add calendar
	cal, _ := svc.AddExternalCalendar(ctx, bathhouseID, "https://example.com/cal.ics", domain.SlotBlockSourceGoogleCalendar)

	// Set up mock HTTP client with iCal response
	futureStart := time.Now().Add(24 * time.Hour)
	futureEnd := futureStart.Add(2 * time.Hour)
	icalData := "BEGIN:VCALENDAR\r\n" +
		"VERSION:2.0\r\n" +
		"BEGIN:VEVENT\r\n" +
		"UID:sync-event-1@google.com\r\n" +
		"DTSTART:" + futureStart.UTC().Format("20060102T150405Z") + "\r\n" +
		"DTEND:" + futureEnd.UTC().Format("20060102T150405Z") + "\r\n" +
		"SUMMARY:External event\r\n" +
		"END:VEVENT\r\n" +
		"END:VCALENDAR\r\n"

	svc.SetHTTPClient(&mockHTTPClient{
		response: &http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(strings.NewReader(icalData)),
		},
	})

	// Sync
	_, err := svc.SyncCalendar(ctx, cal.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify blocks were created
	blocks, _ := blockRepo.ListByBathhouse(ctx, bathhouseID)
	if len(blocks) != 1 {
		t.Fatalf("expected 1 block, got %d", len(blocks))
	}
	if blocks[0].ExternalID != "sync-event-1@google.com" {
		t.Errorf("unexpected external ID: %s", blocks[0].ExternalID)
	}
	if blocks[0].Description != "External event" {
		t.Errorf("unexpected description: %s", blocks[0].Description)
	}

	// Verify sync status was updated
	updated, _ := extRepo.GetByID(ctx, cal.ID)
	if updated.LastSyncAt == nil {
		t.Error("expected LastSyncAt to be set")
	}
	if updated.LastError != "" {
		t.Errorf("expected no error, got: %s", updated.LastError)
	}
}

func TestSyncCalendar_SkipsPastEvents(t *testing.T) {
	svc, _, blockRepo, _ := newTestSyncService()
	ctx := context.Background()
	bathhouseID := uuid.New()

	cal, _ := svc.AddExternalCalendar(ctx, bathhouseID, "https://example.com/cal.ics", domain.SlotBlockSourceGoogleCalendar)

	// iCal with past event (more than 24h ago)
	pastStart := time.Now().Add(-48 * time.Hour)
	pastEnd := pastStart.Add(2 * time.Hour)
	icalData := "BEGIN:VCALENDAR\r\n" +
		"BEGIN:VEVENT\r\n" +
		"UID:old-event@google.com\r\n" +
		"DTSTART:" + pastStart.UTC().Format("20060102T150405Z") + "\r\n" +
		"DTEND:" + pastEnd.UTC().Format("20060102T150405Z") + "\r\n" +
		"SUMMARY:Old event\r\n" +
		"END:VEVENT\r\n" +
		"END:VCALENDAR\r\n"

	svc.SetHTTPClient(&mockHTTPClient{
		response: &http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(strings.NewReader(icalData)),
		},
	})

	_, _ = svc.SyncCalendar(ctx, cal.ID)

	blocks, _ := blockRepo.ListByBathhouse(ctx, bathhouseID)
	if len(blocks) != 0 {
		t.Errorf("expected 0 blocks (past events skipped), got %d", len(blocks))
	}
}

func TestSyncCalendar_HTTPError(t *testing.T) {
	svc, extRepo, _, _ := newTestSyncService()
	ctx := context.Background()
	bathhouseID := uuid.New()

	cal, _ := svc.AddExternalCalendar(ctx, bathhouseID, "https://example.com/cal.ics", domain.SlotBlockSourceGoogleCalendar)

	svc.SetHTTPClient(&mockHTTPClient{
		response: &http.Response{
			StatusCode: 500,
			Body:       io.NopCloser(strings.NewReader("")),
		},
	})

	_, err := svc.SyncCalendar(ctx, cal.ID)
	if err == nil {
		t.Error("expected error for HTTP 500")
	}

	// Check error was recorded
	updated, _ := extRepo.GetByID(ctx, cal.ID)
	if updated.LastError == "" {
		t.Error("expected LastError to be set")
	}
}

func TestSyncCalendar_ReplacesOldBlocks(t *testing.T) {
	svc, _, blockRepo, _ := newTestSyncService()
	ctx := context.Background()
	bathhouseID := uuid.New()

	cal, _ := svc.AddExternalCalendar(ctx, bathhouseID, "https://example.com/cal.ics", domain.SlotBlockSourceGoogleCalendar)

	// Pre-create a block from google_calendar
	_ = blockRepo.Create(ctx, &domain.SlotBlock{
		BathhouseID: bathhouseID,
		StartTime:   time.Now().Add(1 * time.Hour),
		EndTime:     time.Now().Add(3 * time.Hour),
		Source:      domain.SlotBlockSourceGoogleCalendar,
		ExternalID:  "old-event@google.com",
	})

	futureStart := time.Now().Add(24 * time.Hour)
	futureEnd := futureStart.Add(2 * time.Hour)
	icalData := "BEGIN:VCALENDAR\r\n" +
		"BEGIN:VEVENT\r\n" +
		"UID:new-event@google.com\r\n" +
		"DTSTART:" + futureStart.UTC().Format("20060102T150405Z") + "\r\n" +
		"DTEND:" + futureEnd.UTC().Format("20060102T150405Z") + "\r\n" +
		"SUMMARY:New event\r\n" +
		"END:VEVENT\r\n" +
		"END:VCALENDAR\r\n"

	svc.SetHTTPClient(&mockHTTPClient{
		response: &http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(strings.NewReader(icalData)),
		},
	})

	_, _ = svc.SyncCalendar(ctx, cal.ID)

	blocks, _ := blockRepo.ListByBathhouse(ctx, bathhouseID)
	if len(blocks) != 1 {
		t.Fatalf("expected 1 block after sync, got %d", len(blocks))
	}
	if blocks[0].ExternalID != "new-event@google.com" {
		t.Errorf("expected new event, got: %s", blocks[0].ExternalID)
	}
}

func TestSyncCalendar_DetectsConflicts(t *testing.T) {
	svc, _, _, bookingRepo := newTestSyncService()
	ctx := context.Background()
	bathhouseID := uuid.New()

	cal, _ := svc.AddExternalCalendar(ctx, bathhouseID, "https://example.com/cal.ics", domain.SlotBlockSourceGoogleCalendar)

	// Create a confirmed booking that will overlap with the external event
	futureStart := time.Now().Add(24 * time.Hour)
	futureEnd := futureStart.Add(2 * time.Hour)
	booking := &domain.Booking{
		ID:          uuid.New(),
		UserID:      uuid.New(),
		BathhouseID: bathhouseID,
		StartTime:   futureStart,
		EndTime:     futureEnd,
		GuestCount:  2,
		TotalPrice:  10000,
		Status:      domain.BookingConfirmed,
	}
	_ = bookingRepo.Create(ctx, booking)

	// External event that overlaps with the booking
	icalData := "BEGIN:VCALENDAR\r\n" +
		"BEGIN:VEVENT\r\n" +
		"UID:conflict-event@google.com\r\n" +
		"DTSTART:" + futureStart.Add(-30*time.Minute).UTC().Format("20060102T150405Z") + "\r\n" +
		"DTEND:" + futureEnd.Add(30*time.Minute).UTC().Format("20060102T150405Z") + "\r\n" +
		"SUMMARY:Personal event\r\n" +
		"END:VEVENT\r\n" +
		"END:VCALENDAR\r\n"

	svc.SetHTTPClient(&mockHTTPClient{
		response: &http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(strings.NewReader(icalData)),
		},
	})

	conflicts, err := svc.SyncCalendar(ctx, cal.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(conflicts) != 1 {
		t.Fatalf("expected 1 conflict, got %d", len(conflicts))
	}
	if conflicts[0].Booking.ID != booking.ID {
		t.Errorf("expected booking ID %s, got %s", booking.ID, conflicts[0].Booking.ID)
	}
	if conflicts[0].SlotBlock.ExternalID != "conflict-event@google.com" {
		t.Errorf("expected external ID conflict-event@google.com, got %s", conflicts[0].SlotBlock.ExternalID)
	}
}

func TestSyncCalendar_NoConflictsWhenNoOverlap(t *testing.T) {
	svc, _, _, bookingRepo := newTestSyncService()
	ctx := context.Background()
	bathhouseID := uuid.New()

	cal, _ := svc.AddExternalCalendar(ctx, bathhouseID, "https://example.com/cal.ics", domain.SlotBlockSourceGoogleCalendar)

	// Booking at 10-12
	bookingStart := time.Now().Add(24 * time.Hour).Truncate(time.Hour)
	bookingEnd := bookingStart.Add(2 * time.Hour)
	_ = bookingRepo.Create(ctx, &domain.Booking{
		ID: uuid.New(), UserID: uuid.New(), BathhouseID: bathhouseID,
		StartTime: bookingStart, EndTime: bookingEnd,
		GuestCount: 2, TotalPrice: 10000, Status: domain.BookingConfirmed,
	})

	// External event at 14-16 (no overlap)
	eventStart := bookingEnd.Add(2 * time.Hour)
	eventEnd := eventStart.Add(2 * time.Hour)
	icalData := "BEGIN:VCALENDAR\r\n" +
		"BEGIN:VEVENT\r\n" +
		"UID:no-conflict@google.com\r\n" +
		"DTSTART:" + eventStart.UTC().Format("20060102T150405Z") + "\r\n" +
		"DTEND:" + eventEnd.UTC().Format("20060102T150405Z") + "\r\n" +
		"SUMMARY:No conflict\r\n" +
		"END:VEVENT\r\n" +
		"END:VCALENDAR\r\n"

	svc.SetHTTPClient(&mockHTTPClient{
		response: &http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(strings.NewReader(icalData)),
		},
	})

	conflicts, err := svc.SyncCalendar(ctx, cal.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(conflicts) != 0 {
		t.Errorf("expected 0 conflicts, got %d", len(conflicts))
	}
}

func TestSyncCalendar_IgnoresCancelledBookings(t *testing.T) {
	svc, _, _, bookingRepo := newTestSyncService()
	ctx := context.Background()
	bathhouseID := uuid.New()

	cal, _ := svc.AddExternalCalendar(ctx, bathhouseID, "https://example.com/cal.ics", domain.SlotBlockSourceGoogleCalendar)

	// Cancelled booking that overlaps
	futureStart := time.Now().Add(24 * time.Hour)
	futureEnd := futureStart.Add(2 * time.Hour)
	bookingID := uuid.New()
	_ = bookingRepo.Create(ctx, &domain.Booking{
		ID: bookingID, UserID: uuid.New(), BathhouseID: bathhouseID,
		StartTime: futureStart, EndTime: futureEnd,
		GuestCount: 2, TotalPrice: 10000, Status: domain.BookingPending,
	})
	_ = bookingRepo.UpdateStatus(ctx, bookingID, domain.BookingCancelled)

	icalData := "BEGIN:VCALENDAR\r\n" +
		"BEGIN:VEVENT\r\n" +
		"UID:overlap@google.com\r\n" +
		"DTSTART:" + futureStart.UTC().Format("20060102T150405Z") + "\r\n" +
		"DTEND:" + futureEnd.UTC().Format("20060102T150405Z") + "\r\n" +
		"SUMMARY:Overlap\r\n" +
		"END:VEVENT\r\n" +
		"END:VCALENDAR\r\n"

	svc.SetHTTPClient(&mockHTTPClient{
		response: &http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(strings.NewReader(icalData)),
		},
	})

	conflicts, err := svc.SyncCalendar(ctx, cal.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(conflicts) != 0 {
		t.Errorf("expected 0 conflicts (cancelled booking), got %d", len(conflicts))
	}
}

func TestGetConflicts(t *testing.T) {
	svc, _, blockRepo, bookingRepo := newTestSyncService()
	ctx := context.Background()
	bathhouseID := uuid.New()

	// Create an external slot block
	futureStart := time.Now().Add(24 * time.Hour)
	futureEnd := futureStart.Add(2 * time.Hour)
	_ = blockRepo.Create(ctx, &domain.SlotBlock{
		BathhouseID: bathhouseID,
		StartTime:   futureStart,
		EndTime:     futureEnd,
		Source:      domain.SlotBlockSourceGoogleCalendar,
		ExternalID:  "ext-1@google.com",
		Description: "External event",
	})

	// Create an overlapping confirmed booking
	_ = bookingRepo.Create(ctx, &domain.Booking{
		ID: uuid.New(), UserID: uuid.New(), BathhouseID: bathhouseID,
		StartTime: futureStart.Add(30 * time.Minute), EndTime: futureEnd.Add(-30 * time.Minute),
		GuestCount: 2, TotalPrice: 10000, Status: domain.BookingConfirmed,
	})

	// Create a manual block (should be ignored)
	_ = blockRepo.Create(ctx, &domain.SlotBlock{
		BathhouseID: bathhouseID,
		StartTime:   futureEnd.Add(time.Hour),
		EndTime:     futureEnd.Add(3 * time.Hour),
		Source:      domain.SlotBlockSourceManual,
		Description: "Manual block",
	})

	conflicts, err := svc.GetConflicts(ctx, bathhouseID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(conflicts) != 1 {
		t.Fatalf("expected 1 conflict, got %d", len(conflicts))
	}
	if conflicts[0].SlotBlock.ExternalID != "ext-1@google.com" {
		t.Errorf("expected external block, got: %s", conflicts[0].SlotBlock.ExternalID)
	}
}
