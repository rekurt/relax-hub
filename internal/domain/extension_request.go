package domain

import (
	"time"

	"github.com/google/uuid"
)

// ExtensionRequestStatus represents the status of a booking extension request.
type ExtensionRequestStatus string

const (
	ExtReqPending  ExtensionRequestStatus = "pending"
	ExtReqApproved ExtensionRequestStatus = "approved"
	ExtReqRejected ExtensionRequestStatus = "rejected"
	ExtReqExpired  ExtensionRequestStatus = "expired"
)

func (s ExtensionRequestStatus) IsValid() bool {
	switch s {
	case ExtReqPending, ExtReqApproved, ExtReqRejected, ExtReqExpired:
		return true
	}
	return false
}

// ExtensionRequestTimeout is how long the owner has to respond before auto-expiry.
const ExtensionRequestTimeout = 30 * time.Minute

// BookingExtensionRequest stores a pending extension request that needs owner approval.
type BookingExtensionRequest struct {
	ID              uuid.UUID
	BookingID       uuid.UUID
	UserID          uuid.UUID // client who requested the extension
	BathhouseID     uuid.UUID
	Status          ExtensionRequestStatus
	ExtraHours      int
	ExtensionPrice  int64 // price for the extension in kopecks
	NewEndTime      time.Time
	HoldID          *uuid.UUID // wallet hold ID for the extension payment
	RejectionReason string
	CreatedAt       time.Time
	ExpiresAt       time.Time
	ResolvedAt      *time.Time
}
