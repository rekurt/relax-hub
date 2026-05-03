package domain

import (
	"time"

	"github.com/google/uuid"
)

// ModificationRequestStatus represents the status of a booking modification request.
type ModificationRequestStatus string

const (
	ModReqPending  ModificationRequestStatus = "pending"
	ModReqApproved ModificationRequestStatus = "approved"
	ModReqRejected ModificationRequestStatus = "rejected"
	ModReqExpired  ModificationRequestStatus = "expired"
)

func (s ModificationRequestStatus) IsValid() bool {
	switch s {
	case ModReqPending, ModReqApproved, ModReqRejected, ModReqExpired:
		return true
	}
	return false
}

// ModificationRequestTimeout is how long the owner has to respond before auto-expiry.
const ModificationRequestTimeout = 24 * time.Hour

// BookingModificationRequest stores a pending modification request that needs owner approval.
type BookingModificationRequest struct {
	ID                 uuid.UUID
	BookingID          uuid.UUID
	UserID             uuid.UUID // client who requested the modification
	BathhouseID        uuid.UUID
	Status             ModificationRequestStatus
	OldStartTime       time.Time
	OldEndTime         time.Time
	OldGuestCount      int
	OldTotalPrice      int64
	ProposedStartTime  time.Time
	ProposedEndTime    time.Time
	ProposedGuestCount int
	ProposedTotalPrice int64
	RejectionReason    string
	CreatedAt          time.Time
	ExpiresAt          time.Time
	ResolvedAt         *time.Time
}
