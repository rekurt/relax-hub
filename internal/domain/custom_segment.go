package domain

import (
	"time"

	"github.com/google/uuid"
)

// CustomSegment is a user-defined CRM segment with dynamic filter conditions.
type CustomSegment struct {
	ID          uuid.UUID              `json:"id"`
	OwnerID     uuid.UUID              `json:"owner_id"`
	BathhouseID *uuid.UUID             `json:"bathhouse_id,omitempty"`
	Name        string                 `json:"name"`
	Conditions  CustomSegmentCondition `json:"conditions"`
	GuestCount  int64                  `json:"guest_count"` // computed on demand, not stored
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// CustomSegmentCondition defines the filter rules for a custom segment.
// All non-nil fields are ANDed together.
type CustomSegmentCondition struct {
	VisitCountMin      *int     `json:"visit_count_min,omitempty"`
	VisitCountMax      *int     `json:"visit_count_max,omitempty"`
	AvgCheckMin        *int64   `json:"avg_check_min,omitempty"`  // kopecks
	AvgCheckMax        *int64   `json:"avg_check_max,omitempty"`  // kopecks
	TotalSpentMin      *int64   `json:"total_spent_min,omitempty"` // kopecks
	TotalSpentMax      *int64   `json:"total_spent_max,omitempty"` // kopecks
	LastVisitDaysMin   *int     `json:"last_visit_days_min,omitempty"`
	LastVisitDaysMax   *int     `json:"last_visit_days_max,omitempty"`
	TagsInclude        []string `json:"tags_include,omitempty"`
	TagsExclude        []string `json:"tags_exclude,omitempty"`
	RFMRecencyMin      *int     `json:"rfm_recency_min,omitempty"`
	RFMRecencyMax      *int     `json:"rfm_recency_max,omitempty"`
	RFMFrequencyMin    *int     `json:"rfm_frequency_min,omitempty"`
	RFMFrequencyMax    *int     `json:"rfm_frequency_max,omitempty"`
	RFMMonetaryMin     *int     `json:"rfm_monetary_min,omitempty"`
	RFMMonetaryMax     *int     `json:"rfm_monetary_max,omitempty"`
}

func (s *CustomSegment) Validate() error {
	if s.Name == "" {
		return ErrInvalidInput
	}
	if s.OwnerID == uuid.Nil {
		return ErrInvalidInput
	}
	c := s.Conditions
	if c.VisitCountMin != nil && *c.VisitCountMin < 0 {
		return ErrInvalidInput
	}
	if c.VisitCountMax != nil && *c.VisitCountMax < 0 {
		return ErrInvalidInput
	}
	if c.VisitCountMin != nil && c.VisitCountMax != nil && *c.VisitCountMin > *c.VisitCountMax {
		return ErrInvalidInput
	}
	if c.AvgCheckMin != nil && c.AvgCheckMax != nil && *c.AvgCheckMin > *c.AvgCheckMax {
		return ErrInvalidInput
	}
	if c.TotalSpentMin != nil && c.TotalSpentMax != nil && *c.TotalSpentMin > *c.TotalSpentMax {
		return ErrInvalidInput
	}
	if c.LastVisitDaysMin != nil && *c.LastVisitDaysMin < 0 {
		return ErrInvalidInput
	}
	if c.LastVisitDaysMax != nil && *c.LastVisitDaysMax < 0 {
		return ErrInvalidInput
	}
	if c.LastVisitDaysMin != nil && c.LastVisitDaysMax != nil && *c.LastVisitDaysMin > *c.LastVisitDaysMax {
		return ErrInvalidInput
	}
	// RFM scores must be 1-5
	for _, v := range []*int{c.RFMRecencyMin, c.RFMRecencyMax, c.RFMFrequencyMin, c.RFMFrequencyMax, c.RFMMonetaryMin, c.RFMMonetaryMax} {
		if v != nil && (*v < 1 || *v > 5) {
			return ErrInvalidInput
		}
	}
	return nil
}

type CustomSegmentFilter struct {
	OwnerID     uuid.UUID
	BathhouseID *uuid.UUID
}
