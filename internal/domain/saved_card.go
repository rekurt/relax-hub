package domain

import (
	"time"

	"github.com/google/uuid"
)

// SavedCard represents a tokenized payment card saved for repeat payments.
type SavedCard struct {
	ID            uuid.UUID `json:"id"`
	UserID        uuid.UUID `json:"user_id"`
	ProviderToken string    `json:"provider_token"`
	Last4         string    `json:"last4"`
	Brand         string    `json:"brand"`
	ExpiryMonth   int       `json:"expiry_month"`
	ExpiryYear    int       `json:"expiry_year"`
	IsDefault     bool      `json:"is_default"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

func (c *SavedCard) Validate() error {
	if c.UserID == uuid.Nil {
		return ErrInvalidInput
	}
	if c.ProviderToken == "" {
		return ErrInvalidInput
	}
	if c.Last4 == "" || len(c.Last4) != 4 {
		return ErrInvalidInput
	}
	if c.Brand == "" {
		return ErrInvalidInput
	}
	if c.ExpiryMonth < 1 || c.ExpiryMonth > 12 {
		return ErrInvalidInput
	}
	if c.ExpiryYear < 2024 {
		return ErrInvalidInput
	}
	return nil
}
