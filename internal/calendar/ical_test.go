package calendar

import (
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
)

func TestGenerateICal_Empty(t *testing.T) {
	bh := &domain.Bathhouse{
		ID:   uuid.New(),
		Name: "Тестовая баня",
	}

	result := GenerateICal(bh, nil)

	if !strings.Contains(result, "BEGIN:VCALENDAR") {
		t.Error("expected BEGIN:VCALENDAR")
	}
	if !strings.Contains(result, "END:VCALENDAR") {
		t.Error("expected END:VCALENDAR")
	}
	if !strings.Contains(result, "PRODID:-//Bani//Booking Calendar//RU") {
		t.Error("expected PRODID")
	}
	if !strings.Contains(result, "X-WR-CALNAME:Тестовая баня") {
		t.Error("expected calendar name")
	}
	if strings.Contains(result, "BEGIN:VEVENT") {
		t.Error("expected no VEVENT for empty bookings")
	}
}

func TestGenerateICal_WithBookings(t *testing.T) {
	bhID := uuid.New()
	bh := &domain.Bathhouse{
		ID:   bhID,
		Name: "Баня у реки",
	}

	bookingID := uuid.New()
	startTime := time.Date(2026, 3, 4, 10, 0, 0, 0, time.UTC)
	endTime := time.Date(2026, 3, 4, 12, 0, 0, 0, time.UTC)
	createdAt := time.Date(2026, 3, 1, 8, 0, 0, 0, time.UTC)

	bookings := []domain.Booking{
		{
			ID:          bookingID,
			BathhouseID: bhID,
			StartTime:   startTime,
			EndTime:     endTime,
			GuestCount:  4,
			Status:      domain.BookingConfirmed,
			Comment:     "С днем рождения!",
			CreatedAt:   createdAt,
		},
	}

	result := GenerateICal(bh, bookings)

	checks := []struct {
		name     string
		contains string
	}{
		{"VEVENT start", "BEGIN:VEVENT"},
		{"VEVENT end", "END:VEVENT"},
		{"UID", "UID:" + bookingID.String() + "@bani.app"},
		{"DTSTART", "DTSTART:20260304T100000Z"},
		{"DTEND", "DTEND:20260304T120000Z"},
		{"SUMMARY", "SUMMARY:Бронирование - Баня у реки"},
		{"DESCRIPTION guests", "Гостей: 4"},
		{"DESCRIPTION comment", "Комментарий: С днем рождения!"},
		{"STATUS", "STATUS:CONFIRMED"},
	}

	for _, c := range checks {
		if !strings.Contains(result, c.contains) {
			t.Errorf("%s: expected %q in output", c.name, c.contains)
		}
	}
}

func TestGenerateICal_StatusMapping(t *testing.T) {
	bh := &domain.Bathhouse{ID: uuid.New(), Name: "Test"}

	tests := []struct {
		status   domain.BookingStatus
		expected string
	}{
		{domain.BookingPending, "TENTATIVE"},
		{domain.BookingConfirmed, "CONFIRMED"},
		{domain.BookingCompleted, "CONFIRMED"},
		{domain.BookingCancelled, "CANCELLED"},
		{domain.BookingRejected, "CANCELLED"},
	}

	for _, tt := range tests {
		t.Run(string(tt.status), func(t *testing.T) {
			bookings := []domain.Booking{
				{
					ID:         uuid.New(),
					StartTime:  time.Now(),
					EndTime:    time.Now().Add(time.Hour),
					GuestCount: 2,
					Status:     tt.status,
					CreatedAt:  time.Now(),
				},
			}
			result := GenerateICal(bh, bookings)
			if !strings.Contains(result, "STATUS:"+tt.expected) {
				t.Errorf("expected STATUS:%s for booking status %s", tt.expected, tt.status)
			}
		})
	}
}

func TestEscapeICalText(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"simple text", "simple text"},
		{"with, comma", "with\\, comma"},
		{"with; semicolon", "with\\; semicolon"},
		{"with\nnewline", "with\\nnewline"},
		{"back\\slash", "back\\\\slash"},
	}

	for _, tt := range tests {
		result := escapeICalText(tt.input)
		if result != tt.expected {
			t.Errorf("escapeICalText(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestFormatICalTime(t *testing.T) {
	tm := time.Date(2026, 3, 4, 10, 30, 0, 0, time.UTC)
	result := formatICalTime(tm)
	if result != "20260304T103000Z" {
		t.Errorf("formatICalTime = %q, want 20260304T103000Z", result)
	}
}
