package payment

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClassifyError_Nil(t *testing.T) {
	info := ClassifyError(nil)
	assert.False(t, info.Retryable)
	assert.Empty(t, info.Code)
}

func TestClassifyError_Timeout(t *testing.T) {
	err := errors.New("context deadline exceeded")
	info := ClassifyError(err)
	assert.True(t, info.Retryable)
	assert.Equal(t, "timeout", info.Code)
	assert.NotEmpty(t, info.MessageRU)
	assert.NotEmpty(t, info.Suggestion)
}

func TestClassifyError_ConnectionRefused(t *testing.T) {
	err := errors.New("connection refused")
	info := ClassifyError(err)
	assert.True(t, info.Retryable)
	assert.Equal(t, "connection_error", info.Code)
}

func TestClassifyError_RateLimit(t *testing.T) {
	err := errors.New("too many requests")
	info := ClassifyError(err)
	assert.True(t, info.Retryable)
	assert.Equal(t, "rate_limited", info.Code)
}

func TestClassifyError_ServiceUnavailable(t *testing.T) {
	err := errors.New("503 service unavailable")
	info := ClassifyError(err)
	assert.True(t, info.Retryable)
	assert.Equal(t, "provider_error", info.Code)
}

func TestClassifyError_InsufficientFunds(t *testing.T) {
	err := errors.New("insufficient_funds")
	info := ClassifyError(err)
	assert.False(t, info.Retryable)
	assert.Equal(t, "insufficient_funds", info.Code)
	assert.Contains(t, info.Suggestion, "карту")
}

func TestClassifyError_CardDeclined(t *testing.T) {
	err := errors.New("card_declined")
	info := ClassifyError(err)
	assert.False(t, info.Retryable)
	assert.Equal(t, "card_declined", info.Code)
}

func TestClassifyError_ExpiredCard(t *testing.T) {
	err := errors.New("expired_card")
	info := ClassifyError(err)
	assert.False(t, info.Retryable)
	assert.Equal(t, "expired_card", info.Code)
}

func TestClassifyError_InvalidCard(t *testing.T) {
	err := errors.New("invalid_card")
	info := ClassifyError(err)
	assert.False(t, info.Retryable)
	assert.Equal(t, "invalid_card", info.Code)
}

func TestClassifyError_3DSFailed(t *testing.T) {
	err := errors.New("3d_secure failed")
	info := ClassifyError(err)
	assert.False(t, info.Retryable)
	assert.Equal(t, "3ds_failed", info.Code)
}

func TestClassifyError_Blocked(t *testing.T) {
	err := errors.New("card blocked by issuer")
	info := ClassifyError(err)
	assert.False(t, info.Retryable)
	assert.Equal(t, "card_blocked", info.Code)
}

func TestClassifyError_LimitExceeded(t *testing.T) {
	err := errors.New("limit exceeded for card")
	info := ClassifyError(err)
	assert.False(t, info.Retryable)
	assert.Equal(t, "limit_exceeded", info.Code)
}

func TestClassifyError_Unknown(t *testing.T) {
	err := errors.New("some unknown error from provider")
	info := ClassifyError(err)
	assert.False(t, info.Retryable)
	assert.Equal(t, "payment_error", info.Code)
	assert.NotEmpty(t, info.MessageRU)
}
