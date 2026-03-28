package domain

import (
	"math"
	"time"

	"github.com/google/uuid"
)

type ReviewStatus string

const (
	ReviewStatusPending  ReviewStatus = "pending"
	ReviewStatusApproved ReviewStatus = "approved"
	ReviewStatusRejected ReviewStatus = "rejected"
	ReviewStatusHidden   ReviewStatus = "hidden"
)

func (s ReviewStatus) IsValid() bool {
	switch s {
	case ReviewStatusPending, ReviewStatusApproved, ReviewStatusRejected, ReviewStatusHidden:
		return true
	}
	return false
}

type Review struct {
	ID               uuid.UUID
	UserID           uuid.UUID
	BathhouseID      uuid.UUID
	BookingID        uuid.UUID
	Rating           int
	Cleanliness      *float64
	Accuracy         *float64
	Communication    *float64
	ValueForMoney    *float64
	Text             string
	Status           ReviewStatus
	RejectionReasons []string
	OwnerResponse    string
	OwnerResponseAt  *time.Time
	ModerationScore  *float64
	ModerationFlags  []string
	Images           []string
	RevealAt         *time.Time
	IsRevealed       bool
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// ComputeOverallRating calculates the overall rating as the average of the 4 criteria,
// rounded to the nearest integer for the Rating field.
func (r *Review) ComputeOverallRating() {
	if r.Cleanliness != nil && r.Accuracy != nil && r.Communication != nil && r.ValueForMoney != nil {
		avg := (*r.Cleanliness + *r.Accuracy + *r.Communication + *r.ValueForMoney) / 4.0
		r.Rating = int(math.Round(avg))
	}
}

// HasCriteria returns true if all 4 criteria are set.
func (r *Review) HasCriteria() bool {
	return r.Cleanliness != nil && r.Accuracy != nil && r.Communication != nil && r.ValueForMoney != nil
}

// ValidateCriterion checks if a single criterion value is valid (1.0-5.0, step 0.5).
func ValidateCriterion(v float64) bool {
	if v < 1.0 || v > 5.0 {
		return false
	}
	// Check step 0.5: value * 2 should be an integer
	return math.Mod(v*2, 1.0) == 0
}

// ReviewCriteriaAverages holds per-criteria averages for a bathhouse.
type ReviewCriteriaAverages struct {
	AvgCleanliness   float64 `json:"avg_cleanliness"`
	AvgAccuracy      float64 `json:"avg_accuracy"`
	AvgCommunication float64 `json:"avg_communication"`
	AvgValueForMoney float64 `json:"avg_value_for_money"`
}

func (r *Review) Validate() error {
	if r.BookingID == uuid.Nil {
		return ErrInvalidInput
	}
	if r.Rating < 1 || r.Rating > 5 {
		return ErrInvalidInput
	}
	if r.Status != "" && !r.Status.IsValid() {
		return ErrInvalidInput
	}
	// Validate criteria if any are provided — all 4 must be present together
	hasCriteria := r.Cleanliness != nil || r.Accuracy != nil || r.Communication != nil || r.ValueForMoney != nil
	if hasCriteria {
		if !r.HasCriteria() {
			return ErrInvalidInput
		}
		if !ValidateCriterion(*r.Cleanliness) || !ValidateCriterion(*r.Accuracy) ||
			!ValidateCriterion(*r.Communication) || !ValidateCriterion(*r.ValueForMoney) {
			return ErrInvalidInput
		}
	}
	return nil
}

type ReviewFilter struct {
	BathhouseID  *uuid.UUID
	Status       *ReviewStatus
	MinRating    *int
	OnlyRevealed bool
	Page         int
	PageSize     int
}

type AdminReviewFilter struct {
	BathhouseID *uuid.UUID
	Status      *ReviewStatus
	MinRating   *int
	MaxRating   *int
	FromDate    *time.Time
	ToDate      *time.Time
	Page        int
	PageSize    int
}
