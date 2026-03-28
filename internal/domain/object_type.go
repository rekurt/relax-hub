package domain

import (
	"time"

	"github.com/google/uuid"
)

type ObjectType struct {
	ID          uuid.UUID
	Name        string
	Description string
	SortOrder   int
	IsActive    bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (o *ObjectType) Validate() error {
	if o.Name == "" {
		return ErrInvalidInput
	}
	if o.SortOrder < 0 {
		return ErrInvalidInput
	}
	return nil
}
