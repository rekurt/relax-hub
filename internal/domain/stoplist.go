package domain

import (
	"time"

	"github.com/google/uuid"
)

// StoplistEntry represents a blocked phone/email/payment detail in the antifraud stoplist.
type StoplistEntry struct {
	ID             uuid.UUID  `json:"id"`
	Phone          string     `json:"phone,omitempty"`
	Email          string     `json:"email,omitempty"`
	INN            string     `json:"inn,omitempty"`
	BankCardNumber string     `json:"bank_card_number,omitempty"`
	Reason         string     `json:"reason"`
	BlockedAt      time.Time  `json:"blocked_at"`
	CreatedBy      *uuid.UUID `json:"created_by,omitempty"`
}

// StoplistFilter is used for listing stoplist entries.
type StoplistFilter struct {
	Phone string
	Email string
	Page  int
	PageSize int
}
