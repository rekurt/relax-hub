package calendar

import (
	"testing"
	"time"
)

func TestParseICal_BasicEvents(t *testing.T) {
	data := "BEGIN:VCALENDAR\r\n" +
		"VERSION:2.0\r\n" +
		"PRODID:-//Google Inc//Google Calendar//EN\r\n" +
		"BEGIN:VEVENT\r\n" +
		"UID:event1@google.com\r\n" +
		"DTSTART:20260304T100000Z\r\n" +
		"DTEND:20260304T120000Z\r\n" +
		"SUMMARY:Team meeting\r\n" +
		"DESCRIPTION:Weekly sync\r\n" +
		"END:VEVENT\r\n" +
		"BEGIN:VEVENT\r\n" +
		"UID:event2@google.com\r\n" +
		"DTSTART:20260305T140000Z\r\n" +
		"DTEND:20260305T160000Z\r\n" +
		"SUMMARY:Client call\r\n" +
		"END:VEVENT\r\n" +
		"END:VCALENDAR\r\n"

	events, err := ParseICal(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(events) != 2 {
		t.Fatalf("expected 2 events, got %d", len(events))
	}

	ev := events[0]
	if ev.UID != "event1@google.com" {
		t.Errorf("expected UID event1@google.com, got %s", ev.UID)
	}
	if ev.Summary != "Team meeting" {
		t.Errorf("expected summary 'Team meeting', got %s", ev.Summary)
	}
	if ev.Description != "Weekly sync" {
		t.Errorf("expected description 'Weekly sync', got %s", ev.Description)
	}

	expectedStart := time.Date(2026, 3, 4, 10, 0, 0, 0, time.UTC)
	if !ev.Start.Equal(expectedStart) {
		t.Errorf("expected start %v, got %v", expectedStart, ev.Start)
	}

	expectedEnd := time.Date(2026, 3, 4, 12, 0, 0, 0, time.UTC)
	if !ev.End.Equal(expectedEnd) {
		t.Errorf("expected end %v, got %v", expectedEnd, ev.End)
	}
}

func TestParseICal_AllDayEvent(t *testing.T) {
	data := "BEGIN:VCALENDAR\r\n" +
		"BEGIN:VEVENT\r\n" +
		"UID:allday@test.com\r\n" +
		"DTSTART;VALUE=DATE:20260310\r\n" +
		"DTEND;VALUE=DATE:20260311\r\n" +
		"SUMMARY:All day event\r\n" +
		"END:VEVENT\r\n" +
		"END:VCALENDAR\r\n"

	events, err := ParseICal(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}

	ev := events[0]
	expectedStart := time.Date(2026, 3, 10, 0, 0, 0, 0, time.UTC)
	if !ev.Start.Equal(expectedStart) {
		t.Errorf("expected start %v, got %v", expectedStart, ev.Start)
	}
}

func TestParseICal_EscapedText(t *testing.T) {
	data := "BEGIN:VCALENDAR\r\n" +
		"BEGIN:VEVENT\r\n" +
		"UID:esc@test.com\r\n" +
		"DTSTART:20260304T100000Z\r\n" +
		"DTEND:20260304T120000Z\r\n" +
		"SUMMARY:Meeting\\, important\\; urgent\r\n" +
		"DESCRIPTION:Line1\\nLine2\r\n" +
		"END:VEVENT\r\n" +
		"END:VCALENDAR\r\n"

	events, err := ParseICal(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}

	if events[0].Summary != "Meeting, important; urgent" {
		t.Errorf("unexpected summary: %s", events[0].Summary)
	}
	if events[0].Description != "Line1\nLine2" {
		t.Errorf("unexpected description: %s", events[0].Description)
	}
}

func TestParseICal_SkipIncompleteEvents(t *testing.T) {
	data := "BEGIN:VCALENDAR\r\n" +
		"BEGIN:VEVENT\r\n" +
		"UID:no-end@test.com\r\n" +
		"DTSTART:20260304T100000Z\r\n" +
		"SUMMARY:No end time\r\n" +
		"END:VEVENT\r\n" +
		"END:VCALENDAR\r\n"

	events, err := ParseICal(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(events) != 0 {
		t.Fatalf("expected 0 events (incomplete), got %d", len(events))
	}
}

func TestParseICal_EmptyInput(t *testing.T) {
	events, err := ParseICal("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(events) != 0 {
		t.Fatalf("expected 0 events, got %d", len(events))
	}
}

func TestParseICal_LocalTime(t *testing.T) {
	data := "BEGIN:VCALENDAR\r\n" +
		"BEGIN:VEVENT\r\n" +
		"UID:local@test.com\r\n" +
		"DTSTART:20260304T100000\r\n" +
		"DTEND:20260304T120000\r\n" +
		"SUMMARY:Local time event\r\n" +
		"END:VEVENT\r\n" +
		"END:VCALENDAR\r\n"

	events, err := ParseICal(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if events[0].Start.Hour() != 10 {
		t.Errorf("expected hour 10, got %d", events[0].Start.Hour())
	}
}

func TestParseICalDateTime(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"UTC", "20260304T100000Z", false},
		{"Local", "20260304T100000", false},
		{"Date only", "20260304", false},
		{"Invalid", "invalid", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parseICalDateTime(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseICalDateTime(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}
