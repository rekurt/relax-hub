package domain

import (
	"time"

	"github.com/google/uuid"
)

// GuestSegmentSlug identifies a predefined CRM segment.
type GuestSegmentSlug string

const (
	SegmentNew          GuestSegmentSlug = "new"           // 1 visit
	SegmentRegular      GuestSegmentSlug = "regular"       // >= 3 visits
	SegmentLost         GuestSegmentSlug = "lost"          // > 90 days since last visit
	SegmentVIP          GuestSegmentSlug = "vip"           // > 50,000 RUB total spent
	SegmentBirthdaySoon GuestSegmentSlug = "birthday_soon" // birthday within 7 days
)

// AllSegments returns the list of all predefined segment slugs.
func AllSegments() []GuestSegmentSlug {
	return []GuestSegmentSlug{
		SegmentNew,
		SegmentRegular,
		SegmentLost,
		SegmentVIP,
		SegmentBirthdaySoon,
	}
}

// GuestSegment represents a predefined CRM segment with its guest count.
type GuestSegment struct {
	Slug        GuestSegmentSlug `json:"slug"`
	Name        string           `json:"name"`
	Description string           `json:"description"`
	Count       int64            `json:"count"`
}

// SegmentMeta returns display name and description for a segment slug.
func SegmentMeta(slug GuestSegmentSlug) (name, description string) {
	switch slug {
	case SegmentNew:
		return "Новые", "Гости с 1 визитом"
	case SegmentRegular:
		return "Постоянные", "Гости с 3+ визитами"
	case SegmentLost:
		return "Потерянные", "Гости без визита более 90 дней"
	case SegmentVIP:
		return "VIP", "Гости с общей суммой > 50 000 ₽"
	case SegmentBirthdaySoon:
		return "День рождения скоро", "Гости с днём рождения в ближайшие 7 дней"
	default:
		return string(slug), ""
	}
}

type GuestCard struct {
	ID           uuid.UUID
	OwnerID      uuid.UUID
	ClientID     uuid.UUID
	BathhouseID  uuid.UUID
	FirstVisitAt time.Time
	LastVisitAt  time.Time
	VisitCount   int
	TotalSpent   int64 // kopecks
	AvgCheck     int64 // kopecks
	Notes        string
	Tags         []string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type GuestCardFilter struct {
	OwnerID       uuid.UUID
	BathhouseID   *uuid.UUID
	BathhouseIDs  []uuid.UUID // for representatives: filter by managed bathhouse IDs
	NoOwnerFilter bool        // for admins: skip owner/bathhouse filtering
	Search        *string
	Tag           *string
	Segment       *GuestSegmentSlug
	DateFrom      *time.Time
	DateTo        *time.Time
	SortBy        string // "last_visit", "total_spent", "visit_count", "avg_check"
	Page          int
	PageSize      int
}

type GuestCardStats struct {
	TotalGuests   int64 `json:"total_guests"`
	NewThisMonth  int64 `json:"new_this_month"`
	AvgVisitCount int64 `json:"avg_visit_count"`
	AvgSpent      int64 `json:"avg_spent"`
}

// RFMScore represents the RFM (Recency-Frequency-Monetary) scores for a guest.
// Each score ranges from 1 (lowest) to 5 (highest).
type RFMScore struct {
	Recency   int `json:"recency"`   // 5=most recent, 1=least recent
	Frequency int `json:"frequency"` // 5=most visits, 1=fewest visits
	Monetary  int `json:"monetary"`  // 5=highest spend, 1=lowest spend
}

// GuestRFM combines a guest card with its computed RFM scores.
type GuestRFM struct {
	GuestCard
	RFM RFMScore `json:"rfm"`
}

// RFMMatrix holds the count of guests in each RFM cell (R x F).
type RFMMatrixCell struct {
	Recency   int   `json:"recency"`
	Frequency int   `json:"frequency"`
	Count     int64 `json:"count"`
}

// RFMResult holds all RFM data for an owner.
type RFMResult struct {
	Guests []GuestRFM      `json:"guests"`
	Matrix []RFMMatrixCell `json:"matrix"`
}
