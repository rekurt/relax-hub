package domain

import (
	"time"

	"github.com/google/uuid"
)

type BathhouseStatus string

const (
	BathhouseStatusActive   BathhouseStatus = "active"
	BathhouseStatusInactive BathhouseStatus = "inactive"
	BathhouseStatusPending  BathhouseStatus = "pending"
	BathhouseStatusRejected BathhouseStatus = "rejected"
)

func (s BathhouseStatus) IsValid() bool {
	switch s {
	case BathhouseStatusActive, BathhouseStatusInactive, BathhouseStatusPending, BathhouseStatusRejected:
		return true
	}
	return false
}

type WorkingHours struct {
	DayOfWeek int    // 0=Mon, 6=Sun
	OpenTime  string // "09:00"
	CloseTime string // "23:00"
}

type Bathhouse struct {
	ID           uuid.UUID
	OwnerID      uuid.UUID
	Name         string
	Description  string
	Address      string
	CityID       int64
	Latitude     float64
	Longitude    float64
	PricePerHour int64 // in kopecks
	MinDuration  int   // minimum duration in hours
	MaxGuests    int
	HasPool      bool
	HasSauna     bool
	HasSteamRoom bool
	HasHotTub    bool
	HasBBQ       bool
	HasKaraoke   bool
	Rating       float64
	ReviewCount  int
	Images       []string
	WorkingHours []WorkingHours
	Status       BathhouseStatus
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (b *Bathhouse) Validate() error {
	if b.Name == "" {
		return ErrInvalidInput
	}
	if b.Address == "" {
		return ErrInvalidInput
	}
	if b.CityID <= 0 {
		return ErrInvalidInput
	}
	if b.PricePerHour <= 0 {
		return ErrInvalidInput
	}
	if b.MaxGuests <= 0 {
		return ErrInvalidInput
	}
	if b.MinDuration <= 0 {
		return ErrInvalidInput
	}
	return nil
}
