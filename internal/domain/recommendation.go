package domain

import (
	"time"

	"github.com/google/uuid"
)

// BookedBathhouseWithDate represents a booked bathhouse with booking date for recency calculation
type BookedBathhouseWithDate struct {
	BathhouseID uuid.UUID
	BookedAt    time.Time
}

// UserActivityType represents the type of user activity
type UserActivityType string

const (
	ActivityTypeView     UserActivityType = "view"
	ActivityTypeBooking  UserActivityType = "booking"
	ActivityTypeFavorite UserActivityType = "favorite"
)

func (t UserActivityType) IsValid() bool {
	switch t {
	case ActivityTypeView, ActivityTypeBooking, ActivityTypeFavorite:
		return true
	}
	return false
}

// UserPreferences stores explicit user preferences for bathhouse recommendations
type UserPreferences struct {
	UserID          uuid.UUID
	PreferredCityID *int64
	PriceRangeMin   *int64 // in kopecks
	PriceRangeMax   *int64 // in kopecks
	PreferPool      bool
	PreferSauna     bool
	PreferSteamRoom bool
	PreferHotTub    bool
	PreferBBQ       bool
	PreferKaraoke   bool
	UpdatedAt       time.Time
}

// Validate checks if UserPreferences are valid
func (p *UserPreferences) Validate() error {
	if p.UserID == uuid.Nil {
		return ErrInvalidInput
	}
	// Price range validation (if set)
	if p.PriceRangeMin != nil && p.PriceRangeMax != nil {
		if *p.PriceRangeMin < 0 || *p.PriceRangeMax < 0 {
			return ErrInvalidInput
		}
		if *p.PriceRangeMin > *p.PriceRangeMax {
			return ErrInvalidInput
		}
	}
	if p.PreferredCityID != nil && *p.PreferredCityID <= 0 {
		return ErrInvalidInput
	}
	return nil
}

// UserActivity tracks user interactions with bathhouses
type UserActivity struct {
	ID         uuid.UUID
	UserID     uuid.UUID
	BathhouseID uuid.UUID
	Type       UserActivityType
	CreatedAt  time.Time
}

// Validate checks if UserActivity is valid
func (a *UserActivity) Validate() error {
	if a.UserID == uuid.Nil {
		return ErrInvalidInput
	}
	if a.BathhouseID == uuid.Nil {
		return ErrInvalidInput
	}
	if !a.Type.IsValid() {
		return ErrInvalidInput
	}
	return nil
}
