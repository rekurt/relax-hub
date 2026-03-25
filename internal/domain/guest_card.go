package domain

import (
	"time"

	"github.com/google/uuid"
)

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
