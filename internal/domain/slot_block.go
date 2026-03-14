package domain

import (
	"time"

	"github.com/google/uuid"
)

// SlotBlockSource represents the source of a slot block.
type SlotBlockSource string

const (
	SlotBlockSourceManual         SlotBlockSource = "manual"
	SlotBlockSourceGoogleCalendar SlotBlockSource = "google_calendar"
	SlotBlockSourceYandexCalendar SlotBlockSource = "yandex_calendar"
)

// SlotBlock represents a blocked time slot for a bathhouse.
// Blocks can come from manual creation or external calendar sync.
type SlotBlock struct {
	ID                 uuid.UUID
	BathhouseID        uuid.UUID
	StartTime          time.Time
	EndTime            time.Time
	Source             SlotBlockSource
	ExternalID         string
	ExternalCalendarID *uuid.UUID
	Description        string
	CreatedAt          time.Time
}

func (sb *SlotBlock) Validate() error {
	if sb.BathhouseID == uuid.Nil {
		return ErrInvalidInput
	}
	if !sb.EndTime.After(sb.StartTime) {
		return ErrInvalidInput
	}
	if sb.Source == "" {
		return ErrInvalidInput
	}
	return nil
}
