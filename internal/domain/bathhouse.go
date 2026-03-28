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
	BathhouseStatusArchived BathhouseStatus = "archived"
)

const (
	BookingModeInstant = "instant"
	BookingModeRequest = "request"
)

func (s BathhouseStatus) IsValid() bool {
	switch s {
	case BathhouseStatusActive, BathhouseStatusInactive, BathhouseStatusPending, BathhouseStatusRejected, BathhouseStatusArchived:
		return true
	}
	return false
}

type WorkingHours struct {
	DayOfWeek int    `json:"day_of_week"` // 0=Mon, 6=Sun
	OpenTime  string `json:"open_time"`   // "09:00"
	CloseTime string `json:"close_time"`  // "23:00"
}

type Bathhouse struct {
	ID                         uuid.UUID
	OwnerID                    uuid.UUID
	Name                       string
	Slug                       string
	Description                string
	Address                    string
	CityID                     int64
	Latitude                   float64
	Longitude                  float64
	PricePerHour               int64 // in kopecks
	MinDuration                int   // minimum duration in hours
	MaxGuests                  int
	HasPool                    bool
	HasSauna                   bool
	HasSteamRoom               bool
	HasHotTub                  bool
	HasBBQ                     bool
	HasKaraoke                 bool
	Rating                     float64
	BayesianRating             float64
	ReviewCount                int
	ConversionRate             float64
	OccupancyRate              float64
	ViewCount                  int64
	Images                     []string
	WorkingHours               []WorkingHours
	Status                     BathhouseStatus
	IsPromoted                 bool    // transient field, set during List queries
	LongSessionThresholdHours  int     // default 4, minimum hours before discount kicks in
	LongSessionDiscountPercent int     // 0-50, discount on hours beyond threshold
	BaseCapacity               int     // default equals MaxGuests, guests included in base price
	ExtraGuestSurcharge        int64   // kopecks per extra guest per hour
	LastMinuteEnabled          bool    // default false
	LastMinuteDiscountPercent  int     // 5-50, discount for slots starting soon
	LastMinuteHoursThreshold   int     // 2-24, hours before start to apply discount
	BufferMinutes              int     // 0-120 step 15, cleanup time between bookings
	LeadTimeHours              int     // 0-48, minimum hours before booking start
	MaxAdvanceDays             int     // 7-365, max days ahead for booking
	BookingMode                string  // "instant" or "request", default "instant"
	RequestTimeout             int     // hours, default 24, range 1-72
	ResponseRate               float64 // 0.0-1.0, percentage of requests responded to within timeout
	AvgResponseTimeMinutes     int     // average response time in minutes for request-based bookings
	IsPhotoVerified            bool
	ApiKey                     string
	CancellationPolicy         CancellationPolicy // flexible, moderate, strict
	CalendarToken              string
	CreatedAt                  time.Time
	UpdatedAt                  time.Time
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
	if b.Latitude < -90 || b.Latitude > 90 {
		return ErrInvalidInput
	}
	if b.Longitude < -180 || b.Longitude > 180 {
		return ErrInvalidInput
	}
	// Apply defaults for pricing fields when unset (zero value)
	if b.LongSessionThresholdHours == 0 {
		b.LongSessionThresholdHours = 4
	}
	if b.BaseCapacity == 0 {
		b.BaseCapacity = b.MaxGuests
	}
	if b.LongSessionDiscountPercent < 0 || b.LongSessionDiscountPercent > 50 {
		return ErrInvalidInput
	}
	if b.LongSessionThresholdHours < 1 || b.LongSessionThresholdHours > 12 {
		return ErrInvalidInput
	}
	if b.BaseCapacity < 1 || b.BaseCapacity > b.MaxGuests {
		return ErrInvalidInput
	}
	if b.ExtraGuestSurcharge < 0 {
		return ErrInvalidInput
	}
	// Apply defaults for booking settings when unset (BufferMinutes=0 is valid, means no buffer)
	if b.MaxAdvanceDays == 0 {
		b.MaxAdvanceDays = 90
	}
	// LeadTimeHours defaults to 2 but 0 is valid (immediate booking), so default only on creation
	if b.BufferMinutes < 0 || b.BufferMinutes > 120 || b.BufferMinutes%15 != 0 {
		return ErrInvalidInput
	}
	if b.LeadTimeHours < 0 || b.LeadTimeHours > 48 {
		return ErrInvalidInput
	}
	if b.MaxAdvanceDays < 7 || b.MaxAdvanceDays > 365 {
		return ErrInvalidInput
	}
	// Apply defaults for cancellation policy
	if b.CancellationPolicy == "" {
		b.CancellationPolicy = CancellationPolicyFlexible
	}
	if !b.CancellationPolicy.IsValid() {
		return ErrInvalidInput
	}
	// Apply defaults for booking mode
	if b.BookingMode == "" {
		b.BookingMode = BookingModeInstant
	}
	if b.BookingMode != BookingModeInstant && b.BookingMode != BookingModeRequest {
		return ErrInvalidInput
	}
	if b.RequestTimeout == 0 {
		b.RequestTimeout = 24
	}
	if b.RequestTimeout < 1 || b.RequestTimeout > 72 {
		return ErrInvalidInput
	}
	// Apply defaults for last-minute fields when unset
	if b.LastMinuteEnabled {
		if b.LastMinuteDiscountPercent == 0 {
			b.LastMinuteDiscountPercent = 20
		}
		if b.LastMinuteHoursThreshold == 0 {
			b.LastMinuteHoursThreshold = 6
		}
		if b.LastMinuteDiscountPercent < 5 || b.LastMinuteDiscountPercent > 50 {
			return ErrInvalidInput
		}
		if b.LastMinuteHoursThreshold < 2 || b.LastMinuteHoursThreshold > 24 {
			return ErrInvalidInput
		}
	}
	for _, wh := range b.WorkingHours {
		if err := wh.Validate(); err != nil {
			return err
		}
	}
	return nil
}

func (wh WorkingHours) Validate() error {
	if wh.DayOfWeek < 0 || wh.DayOfWeek > 6 {
		return ErrInvalidInput
	}
	if !IsValidTimeFormat(wh.OpenTime) {
		return ErrInvalidInput
	}
	if !IsValidTimeFormat(wh.CloseTime) {
		return ErrInvalidInput
	}
	if wh.OpenTime == wh.CloseTime {
		return ErrInvalidInput
	}
	return nil
}

func IsValidTimeFormat(s string) bool {
	if len(s) != 5 || s[2] != ':' {
		return false
	}
	for _, i := range []int{0, 1, 3, 4} {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	h := (int(s[0]-'0') * 10) + int(s[1]-'0')
	m := (int(s[3]-'0') * 10) + int(s[4]-'0')
	return h >= 0 && h <= 23 && m >= 0 && m <= 59
}
