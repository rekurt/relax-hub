package domain

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestSavedCard_Validate(t *testing.T) {
	validCard := SavedCard{
		UserID:        uuid.New(),
		ProviderToken: "tok_123",
		Last4:         "4242",
		Brand:         "visa",
		ExpiryMonth:   12,
		ExpiryYear:    2027,
	}

	t.Run("valid card", func(t *testing.T) {
		assert.NoError(t, validCard.Validate())
	})

	t.Run("missing user ID", func(t *testing.T) {
		c := validCard
		c.UserID = uuid.Nil
		assert.ErrorIs(t, c.Validate(), ErrInvalidInput)
	})

	t.Run("empty provider token", func(t *testing.T) {
		c := validCard
		c.ProviderToken = ""
		assert.ErrorIs(t, c.Validate(), ErrInvalidInput)
	})

	t.Run("invalid last4 length", func(t *testing.T) {
		c := validCard
		c.Last4 = "42"
		assert.ErrorIs(t, c.Validate(), ErrInvalidInput)
	})

	t.Run("empty last4", func(t *testing.T) {
		c := validCard
		c.Last4 = ""
		assert.ErrorIs(t, c.Validate(), ErrInvalidInput)
	})

	t.Run("empty brand", func(t *testing.T) {
		c := validCard
		c.Brand = ""
		assert.ErrorIs(t, c.Validate(), ErrInvalidInput)
	})

	t.Run("invalid expiry month low", func(t *testing.T) {
		c := validCard
		c.ExpiryMonth = 0
		assert.ErrorIs(t, c.Validate(), ErrInvalidInput)
	})

	t.Run("invalid expiry month high", func(t *testing.T) {
		c := validCard
		c.ExpiryMonth = 13
		assert.ErrorIs(t, c.Validate(), ErrInvalidInput)
	})

	t.Run("invalid expiry year", func(t *testing.T) {
		c := validCard
		c.ExpiryYear = 2023
		assert.ErrorIs(t, c.Validate(), ErrInvalidInput)
	})
}
