package domain

import (
	"net"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
)

// ExternalCalendar represents a subscription to an external iCal calendar feed.
type ExternalCalendar struct {
	ID          uuid.UUID
	BathhouseID uuid.UUID
	URL         string
	Source      SlotBlockSource // google_calendar, yandex_calendar
	LastSyncAt  *time.Time
	LastError   string
	CreatedAt   time.Time
}

func (ec *ExternalCalendar) Validate() error {
	if ec.BathhouseID == uuid.Nil {
		return ErrInvalidInput
	}
	if ec.URL == "" {
		return ErrInvalidInput
	}
	if err := validateCalendarURL(ec.URL); err != nil {
		return ErrInvalidInput
	}
	if ec.Source != SlotBlockSourceGoogleCalendar && ec.Source != SlotBlockSourceYandexCalendar {
		return ErrInvalidInput
	}
	return nil
}

// validateCalendarURL checks that the URL uses HTTPS and does not point to private/loopback addresses.
func validateCalendarURL(rawURL string) error {
	u, err := url.Parse(rawURL)
	if err != nil {
		return ErrInvalidInput
	}
	if !strings.EqualFold(u.Scheme, "https") {
		return ErrInvalidInput
	}
	host := u.Hostname()
	if host == "" {
		return ErrInvalidInput
	}
	ip := net.ParseIP(host)
	if ip != nil && (ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast()) {
		return ErrInvalidInput
	}
	return nil
}
