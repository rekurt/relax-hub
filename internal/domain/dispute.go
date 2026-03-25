package domain

import (
	"time"

	"github.com/google/uuid"
)

type DisputeReason string

const (
	DisputeReasonServiceNotProvided DisputeReason = "service_not_provided"
	DisputeReasonPoorQuality        DisputeReason = "poor_quality"
	DisputeReasonDamage             DisputeReason = "damage"
	DisputeReasonSafetyIssue        DisputeReason = "safety_issue"
	DisputeReasonBillingError       DisputeReason = "billing_error"
	DisputeReasonOther              DisputeReason = "other"
)

func (r DisputeReason) IsValid() bool {
	switch r {
	case DisputeReasonServiceNotProvided, DisputeReasonPoorQuality, DisputeReasonDamage,
		DisputeReasonSafetyIssue, DisputeReasonBillingError, DisputeReasonOther:
		return true
	}
	return false
}

type DisputeStatus string

const (
	DisputeStatusOpen               DisputeStatus = "open"
	DisputeStatusEvidenceCollection DisputeStatus = "evidence_collection"
	DisputeStatusUnderReview        DisputeStatus = "under_review"
	DisputeStatusResolved           DisputeStatus = "resolved"
	DisputeStatusAppealed           DisputeStatus = "appealed"
	DisputeStatusClosed             DisputeStatus = "closed"
)

func (s DisputeStatus) IsValid() bool {
	switch s {
	case DisputeStatusOpen, DisputeStatusEvidenceCollection, DisputeStatusUnderReview,
		DisputeStatusResolved, DisputeStatusAppealed, DisputeStatusClosed:
		return true
	}
	return false
}

type DisputeResolution string

const (
	DisputeResolutionFullRefund    DisputeResolution = "full_refund"
	DisputeResolutionPartialRefund DisputeResolution = "partial_refund"
	DisputeResolutionNoRefund      DisputeResolution = "no_refund"
)

func (r DisputeResolution) IsValid() bool {
	switch r {
	case DisputeResolutionFullRefund, DisputeResolutionPartialRefund, DisputeResolutionNoRefund:
		return true
	}
	return false
}

type DisputeEvidenceType string

const (
	DisputeEvidencePhoto      DisputeEvidenceType = "photo"
	DisputeEvidenceScreenshot DisputeEvidenceType = "screenshot"
	DisputeEvidenceGPS        DisputeEvidenceType = "gps"
	DisputeEvidenceMessage    DisputeEvidenceType = "message"
	DisputeEvidenceReceipt    DisputeEvidenceType = "receipt"
)

func (t DisputeEvidenceType) IsValid() bool {
	switch t {
	case DisputeEvidencePhoto, DisputeEvidenceScreenshot, DisputeEvidenceGPS,
		DisputeEvidenceMessage, DisputeEvidenceReceipt:
		return true
	}
	return false
}

type Dispute struct {
	ID                 uuid.UUID
	BookingID          uuid.UUID
	InitiatorID        uuid.UUID
	RespondentID       uuid.UUID
	Reason             DisputeReason
	Description        string
	Status             DisputeStatus
	Resolution         *DisputeResolution
	RefundAmount       int64
	CompensationAmount int64
	MediatorID         *uuid.UUID
	MediatorNotes      string
	AppealDeadline     *time.Time
	EvidenceDeadline   *time.Time
	CreatedAt          time.Time
	UpdatedAt          time.Time
	ResolvedAt         *time.Time
}

func (d *Dispute) Validate() error {
	if d.BookingID == uuid.Nil {
		return ErrInvalidInput
	}
	if d.InitiatorID == uuid.Nil {
		return ErrInvalidInput
	}
	if d.RespondentID == uuid.Nil {
		return ErrInvalidInput
	}
	if !d.Reason.IsValid() {
		return ErrInvalidInput
	}
	if d.Description == "" || len(d.Description) > 5000 {
		return ErrInvalidInput
	}
	if d.Status != "" && !d.Status.IsValid() {
		return ErrInvalidInput
	}
	return nil
}

type DisputeEvidence struct {
	ID          uuid.UUID
	DisputeID   uuid.UUID
	UserID      uuid.UUID
	Type        DisputeEvidenceType
	URL         string
	Description string
	CreatedAt   time.Time
}

func (e *DisputeEvidence) Validate() error {
	if e.DisputeID == uuid.Nil || e.UserID == uuid.Nil {
		return ErrInvalidInput
	}
	if !e.Type.IsValid() {
		return ErrInvalidInput
	}
	if e.URL == "" || len(e.URL) > 2000 {
		return ErrInvalidInput
	}
	if len(e.Description) > 2000 {
		return ErrInvalidInput
	}
	return nil
}

type DisputeFilter struct {
	Status   *DisputeStatus
	UserID   *uuid.UUID
	Page     int
	PageSize int
}
