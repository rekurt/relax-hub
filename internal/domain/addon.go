package domain

import (
	"time"

	"github.com/google/uuid"
)

type AddOnUnit string

const (
	AddOnUnitPerItem   AddOnUnit = "per_item"
	AddOnUnitPerHour   AddOnUnit = "per_hour"
	AddOnUnitPerPerson AddOnUnit = "per_person"
)

func (u AddOnUnit) IsValid() bool {
	switch u {
	case AddOnUnitPerItem, AddOnUnitPerHour, AddOnUnitPerPerson:
		return true
	}
	return false
}

type AddOn struct {
	ID          uuid.UUID
	BathhouseID uuid.UUID
	Name        string
	Description string
	Price       int64 // kopecks
	Unit        AddOnUnit
	IsActive    bool
	SortOrder   int
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (a *AddOn) Validate() error {
	if a.BathhouseID == uuid.Nil {
		return ErrInvalidInput
	}
	if a.Name == "" {
		return ErrInvalidInput
	}
	if a.Price < 0 {
		return ErrInvalidInput
	}
	if !a.Unit.IsValid() {
		return ErrInvalidInput
	}
	return nil
}

type BookingAddOn struct {
	ID         uuid.UUID
	BookingID  uuid.UUID
	AddOnID    uuid.UUID
	Name       string // denormalized
	Quantity   int
	UnitPrice  int64 // kopecks, denormalized
	TotalPrice int64 // kopecks
	CreatedAt  time.Time
}
