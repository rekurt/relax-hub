package domain

import (
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
	ID              uuid.UUID
	UserID          uuid.UUID
	BathhouseID     uuid.UUID
	BookingID       uuid.UUID
	Rating          int
	Text            string
	Status          ReviewStatus
	OwnerResponse   string
	OwnerResponseAt *time.Time
	Images          []string
	CreatedAt       time.Time
	UpdatedAt       time.Time
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
	return nil
}

type ReviewFilter struct {
	BathhouseID *uuid.UUID
	Status      *ReviewStatus
	MinRating   *int
	Page        int
	PageSize    int
}
