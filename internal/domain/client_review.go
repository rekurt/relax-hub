package domain

import (
	"math"
	"time"

	"github.com/google/uuid"
)

// ClientReview represents an owner's review of a client after a booking.
type ClientReview struct {
	ID              uuid.UUID
	OwnerID         uuid.UUID
	ClientID        uuid.UUID
	BookingID       uuid.UUID
	BathhouseID     uuid.UUID
	Punctuality     *float64 // 1.0-5.0 step 0.5
	Cleanliness     *float64 // 1.0-5.0 step 0.5
	RuleCompliance  *float64 // 1.0-5.0 step 0.5
	Rating          int      // 1-5 (computed from criteria)
	Text            string
	RevealAt        time.Time // when the review becomes visible
	IsRevealed      bool
	ModerationScore *float64
	ModerationFlags []string
	Status          ReviewStatus
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

const (
	// ClientReviewBlindDays is the number of days before auto-reveal if only one side has reviewed.
	ClientReviewBlindDays = 14
)

// ComputeOverallRating calculates the overall rating from the 3 criteria.
func (r *ClientReview) ComputeOverallRating() {
	if r.Punctuality != nil && r.Cleanliness != nil && r.RuleCompliance != nil {
		avg := (*r.Punctuality + *r.Cleanliness + *r.RuleCompliance) / 3.0
		r.Rating = int(math.Round(avg))
	}
}

// HasCriteria returns true if all 3 criteria are set.
func (r *ClientReview) HasCriteria() bool {
	return r.Punctuality != nil && r.Cleanliness != nil && r.RuleCompliance != nil
}

func (r *ClientReview) Validate() error {
	if r.BookingID == uuid.Nil {
		return ErrInvalidInput
	}
	if r.OwnerID == uuid.Nil {
		return ErrInvalidInput
	}
	if r.ClientID == uuid.Nil {
		return ErrInvalidInput
	}
	if r.Rating < 1 || r.Rating > 5 {
		return ErrInvalidInput
	}
	if r.Status != "" && !r.Status.IsValid() {
		return ErrInvalidInput
	}
	// Validate criteria if any are provided — all 3 must be present together
	hasCriteria := r.Punctuality != nil || r.Cleanliness != nil || r.RuleCompliance != nil
	if hasCriteria {
		if !r.HasCriteria() {
			return ErrInvalidInput
		}
		if !ValidateCriterion(*r.Punctuality) || !ValidateCriterion(*r.Cleanliness) ||
			!ValidateCriterion(*r.RuleCompliance) {
			return ErrInvalidInput
		}
	}
	return nil
}

// ClientReviewFilter for listing client reviews.
type ClientReviewFilter struct {
	ClientID *uuid.UUID
	OwnerID  *uuid.UUID
	Page     int
	PageSize int
}
