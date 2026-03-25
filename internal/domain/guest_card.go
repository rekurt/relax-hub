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
	OwnerID     uuid.UUID
	BathhouseID *uuid.UUID
	Search      *string
	Tag         *string
	Segment     *GuestSegmentSlug
	DateFrom    *time.Time
	DateTo      *time.Time
	SortBy      string // "last_visit", "total_spent", "visit_count", "avg_check"
	Page        int
	PageSize    int
}

type GuestCardStats struct {
	TotalGuests   int64 `json:"total_guests"`
	NewThisMonth  int64 `json:"new_this_month"`
	AvgVisitCount int64 `json:"avg_visit_count"`
	AvgSpent      int64 `json:"avg_spent"`
}
