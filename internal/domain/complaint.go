package domain

import (
	"time"

	"github.com/google/uuid"
)

type ComplaintTargetType string

const (
	ComplaintTargetReview    ComplaintTargetType = "review"
	ComplaintTargetBathhouse ComplaintTargetType = "bathhouse"
	ComplaintTargetUser      ComplaintTargetType = "user"
)

func (t ComplaintTargetType) IsValid() bool {
	switch t {
	case ComplaintTargetReview, ComplaintTargetBathhouse, ComplaintTargetUser:
		return true
	}
	return false
}

type ComplaintReason string

const (
	ComplaintReasonSpam      ComplaintReason = "spam"
	ComplaintReasonOffensive ComplaintReason = "offensive"
	ComplaintReasonFake      ComplaintReason = "fake"
	ComplaintReasonFraud     ComplaintReason = "fraud"
	ComplaintReasonOther     ComplaintReason = "other"
)

func (r ComplaintReason) IsValid() bool {
	switch r {
	case ComplaintReasonSpam, ComplaintReasonOffensive, ComplaintReasonFake, ComplaintReasonFraud, ComplaintReasonOther:
		return true
	}
	return false
}

type ComplaintStatus string

const (
	ComplaintStatusPending   ComplaintStatus = "pending"
	ComplaintStatusResolved  ComplaintStatus = "resolved"
	ComplaintStatusDismissed ComplaintStatus = "dismissed"
)

func (s ComplaintStatus) IsValid() bool {
	switch s {
	case ComplaintStatusPending, ComplaintStatusResolved, ComplaintStatusDismissed:
		return true
	}
	return false
}

type Complaint struct {
	ID           uuid.UUID
	ReporterID   uuid.UUID
	TargetType   ComplaintTargetType
	TargetID     uuid.UUID
	Reason       ComplaintReason
	Description  string
	Status       ComplaintStatus
	ResolvedByID *uuid.UUID
	Resolution   string
	ResolvedAt   *time.Time
	CreatedAt    time.Time
}

func (c *Complaint) Validate() error {
	if c.ReporterID == uuid.Nil {
		return ErrInvalidInput
	}
	if !c.TargetType.IsValid() {
		return ErrInvalidInput
	}
	if c.TargetID == uuid.Nil {
		return ErrInvalidInput
	}
	if !c.Reason.IsValid() {
		return ErrInvalidInput
	}
	if c.Status != "" && !c.Status.IsValid() {
		return ErrInvalidInput
	}
	if len(c.Description) > 2000 {
		return ErrInvalidInput
	}
	return nil
}

type ComplaintFilter struct {
	Status     *ComplaintStatus
	TargetType *ComplaintTargetType
	Reason     *ComplaintReason
	FromDate   *time.Time
	ToDate     *time.Time
	Page       int
	PageSize   int
}
