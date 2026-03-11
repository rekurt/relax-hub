package domain

import (
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
	if ec.Source != SlotBlockSourceGoogleCalendar && ec.Source != SlotBlockSourceYandexCalendar {
		return ErrInvalidInput
	}
	return nil
}
