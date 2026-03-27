package domain

import "time"

// FeatureFlag represents a toggleable platform feature, optionally scoped to a region.
type FeatureFlag struct {
	Key         string    `json:"key"`
	Enabled     bool      `json:"enabled"`
	Description string    `json:"description"`
	Region      *string   `json:"region,omitempty"`
	UpdatedAt   time.Time `json:"updated_at"`
	UpdatedBy   *string   `json:"updated_by,omitempty"`
}

func (f *FeatureFlag) Validate() error {
	if f.Key == "" {
		return ErrInvalidInput
	}
	return nil
}
