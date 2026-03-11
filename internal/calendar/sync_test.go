package calendar

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/logger"
	"github.com/nikitaaldaev/bani/internal/repository/mock"
)

type mockHTTPClient struct {
	response *http.Response
	err      error
}

func (m *mockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	return m.response, m.err
}

func newTestSyncService() (*CalendarSyncService, *mock.ExternalCalendarRepo, *mock.SlotBlockRepo) {
	extRepo := mock.NewExternalCalendarRepo()
	blockRepo := mock.NewSlotBlockRepo()
	log := logger.New(logger.LevelDebug)

	svc := NewCalendarSyncService(extRepo, blockRepo, log)
	return svc, extRepo, blockRepo
}

func TestAddExternalCalendar(t *testing.T) {
	svc, extRepo, _ := newTestSyncService()
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
	svc, _, _ := newTestSyncService()
	ctx := context.Background()

	_, err := svc.AddExternalCalendar(ctx, uuid.New(), "https://example.com/cal.ics", domain.SlotBlockSourceManual)
	if err == nil {
		t.Error("expected error for manual source")
	}
}

func TestAddExternalCalendar_EmptyURL(t *testing.T) {
	svc, _, _ := newTestSyncService()
	ctx := context.Background()

	_, err := svc.AddExternalCalendar(ctx, uuid.New(), "", domain.SlotBlockSourceGoogleCalendar)
	if err == nil {
		t.Error("expected error for empty URL")
	}
}

func TestRemoveExternalCalendar(t *testing.T) {
	svc, extRepo, blockRepo := newTestSyncService()
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
	svc, extRepo, blockRepo := newTestSyncService()
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
	err := svc.SyncCalendar(ctx, cal.ID)
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
	svc, _, blockRepo := newTestSyncService()
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

	_ = svc.SyncCalendar(ctx, cal.ID)

	blocks, _ := blockRepo.ListByBathhouse(ctx, bathhouseID)
	if len(blocks) != 0 {
		t.Errorf("expected 0 blocks (past events skipped), got %d", len(blocks))
	}
}

func TestSyncCalendar_HTTPError(t *testing.T) {
	svc, extRepo, _ := newTestSyncService()
	ctx := context.Background()
	bathhouseID := uuid.New()

	cal, _ := svc.AddExternalCalendar(ctx, bathhouseID, "https://example.com/cal.ics", domain.SlotBlockSourceGoogleCalendar)

	svc.SetHTTPClient(&mockHTTPClient{
		response: &http.Response{
			StatusCode: 500,
			Body:       io.NopCloser(strings.NewReader("")),
		},
	})

	err := svc.SyncCalendar(ctx, cal.ID)
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
	svc, _, blockRepo := newTestSyncService()
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

	_ = svc.SyncCalendar(ctx, cal.ID)

	blocks, _ := blockRepo.ListByBathhouse(ctx, bathhouseID)
	if len(blocks) != 1 {
		t.Fatalf("expected 1 block after sync, got %d", len(blocks))
	}
	if blocks[0].ExternalID != "new-event@google.com" {
		t.Errorf("expected new event, got: %s", blocks[0].ExternalID)
	}
}
