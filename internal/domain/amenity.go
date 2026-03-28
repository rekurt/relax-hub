package domain

import (
	"time"

	"github.com/google/uuid"
)

type Amenity struct {
	ID        uuid.UUID
	Name      string
	Icon      string // icon identifier (e.g. "pool", "sauna", "bbq")
	SortOrder int
	IsActive  bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (a *Amenity) Validate() error {
	if a.Name == "" {
		return ErrInvalidInput
	}
	if a.SortOrder < 0 {
		return ErrInvalidInput
	}
	return nil
}
