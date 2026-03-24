package domain

import (
	"time"

	"github.com/google/uuid"
)

// ServiceFeeConfig defines a platform service fee percentage for a given region and optional category.
// Resolution priority: region+category > region > global default ("*").
type ServiceFeeConfig struct {
	ID         uuid.UUID
	Region     string  // e.g. "RU", "BY" or "*" for global default
	Category   *string // optional bathhouse category filter
	FeePercent float64 // range 0-25, step 0.5, default 10.0
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

func (c *ServiceFeeConfig) Validate() error {
	if c.Region == "" {
		return ErrInvalidInput
	}
	if c.FeePercent < 0 || c.FeePercent > 25 {
		return ErrInvalidInput
	}
	return nil
}
