package domain

import (
	"time"

	"github.com/google/uuid"
)

// PhotoOrderStatus represents the current state of a photo order.
type PhotoOrderStatus string

const (
	PhotoOrderStatusRequested PhotoOrderStatus = "requested"
	PhotoOrderStatusConfirmed PhotoOrderStatus = "confirmed"
	PhotoOrderStatusCompleted PhotoOrderStatus = "completed"
	PhotoOrderStatusCancelled PhotoOrderStatus = "cancelled"
)

func (s PhotoOrderStatus) IsValid() bool {
	switch s {
	case PhotoOrderStatusRequested, PhotoOrderStatusConfirmed,
		PhotoOrderStatusCompleted, PhotoOrderStatusCancelled:
		return true
	}
	return false
}

// PhotoOrder represents a professional photography order for a bathhouse.
type PhotoOrder struct {
	ID               uuid.UUID        `json:"id"`
	OwnerID          uuid.UUID        `json:"owner_id"`
	BathhouseID      uuid.UUID        `json:"bathhouse_id"`
	Region           string           `json:"region"`
	Status           PhotoOrderStatus `json:"status"`
	PhotographerName string           `json:"photographer_name,omitempty"`
	Price            int64            `json:"price"`
	ScheduledAt      *time.Time       `json:"scheduled_at,omitempty"`
	Notes            string           `json:"notes,omitempty"`
	AdminNotes       string           `json:"admin_notes,omitempty"`
	CreatedAt        time.Time        `json:"created_at"`
	UpdatedAt        time.Time        `json:"updated_at"`
}

func (o *PhotoOrder) Validate() error {
	if o.OwnerID == uuid.Nil {
		return ErrInvalidInput
	}
	if o.BathhouseID == uuid.Nil {
		return ErrInvalidInput
	}
	if o.Region == "" {
		return ErrInvalidInput
	}
	return nil
}

// PhotoOrderFilter for listing photo orders with filters.
type PhotoOrderFilter struct {
	OwnerID     *uuid.UUID
	BathhouseID *uuid.UUID
	Status      *PhotoOrderStatus
	Region      *string
	Page        int
	PageSize    int
}
