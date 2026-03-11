package calendar

import (
	"fmt"
	"strings"
	"time"

	"github.com/nikitaaldaev/bani/internal/domain"
)

// GenerateICal generates an iCalendar (RFC 5545) feed from a list of bookings.
func GenerateICal(bathhouse *domain.Bathhouse, bookings []domain.Booking) string {
	var b strings.Builder

	b.WriteString("BEGIN:VCALENDAR\r\n")
	b.WriteString("VERSION:2.0\r\n")
	b.WriteString("PRODID:-//Bani//Booking Calendar//RU\r\n")
	b.WriteString("CALSCALE:GREGORIAN\r\n")
	b.WriteString("METHOD:PUBLISH\r\n")
	b.WriteString(fmt.Sprintf("X-WR-CALNAME:%s\r\n", escapeICalText(bathhouse.Name)))

	for _, booking := range bookings {
		b.WriteString("BEGIN:VEVENT\r\n")
		b.WriteString(fmt.Sprintf("UID:%s@bani.app\r\n", booking.ID.String()))
		b.WriteString(fmt.Sprintf("DTSTART:%s\r\n", formatICalTime(booking.StartTime)))
		b.WriteString(fmt.Sprintf("DTEND:%s\r\n", formatICalTime(booking.EndTime)))
		b.WriteString(fmt.Sprintf("DTSTAMP:%s\r\n", formatICalTime(booking.CreatedAt)))

		summary := fmt.Sprintf("Бронирование - %s", bathhouse.Name)
		b.WriteString(fmt.Sprintf("SUMMARY:%s\r\n", escapeICalText(summary)))

		description := fmt.Sprintf("Гостей: %d, Статус: %s", booking.GuestCount, booking.Status)
		if booking.Comment != "" {
			description += fmt.Sprintf(", Комментарий: %s", booking.Comment)
		}
		b.WriteString(fmt.Sprintf("DESCRIPTION:%s\r\n", escapeICalText(description)))

		b.WriteString(fmt.Sprintf("STATUS:%s\r\n", bookingStatusToICalStatus(booking.Status)))
		b.WriteString("END:VEVENT\r\n")
	}

	b.WriteString("END:VCALENDAR\r\n")
	return b.String()
}

func formatICalTime(t time.Time) string {
	return t.UTC().Format("20060102T150405Z")
}

func escapeICalText(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, ";", "\\;")
	s = strings.ReplaceAll(s, ",", "\\,")
	s = strings.ReplaceAll(s, "\n", "\\n")
	return s
}

func bookingStatusToICalStatus(status domain.BookingStatus) string {
	switch status {
	case domain.BookingConfirmed, domain.BookingCompleted:
		return "CONFIRMED"
	case domain.BookingCancelled, domain.BookingRejected:
		return "CANCELLED"
	default:
		return "TENTATIVE"
	}
}
