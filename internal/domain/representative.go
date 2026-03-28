package domain

import (
	"time"

	"github.com/google/uuid"
)

// RepresentativeRole defines sub-roles for representatives.
type RepresentativeRole string

const (
	RepRoleManager  RepresentativeRole = "manager"
	RepRoleObserver RepresentativeRole = "observer"
)

func (r RepresentativeRole) IsValid() bool {
	return r == RepRoleManager || r == RepRoleObserver
}

// CanWrite returns true if the role allows write operations (create, update, delete).
func (r RepresentativeRole) CanWrite() bool {
	return r == RepRoleManager
}

type Representative struct {
	ID          uuid.UUID
	UserID      uuid.UUID
	BathhouseID uuid.UUID
	OwnerID     uuid.UUID
	Role        RepresentativeRole
	CreatedAt   time.Time
}
