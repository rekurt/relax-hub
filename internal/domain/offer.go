package domain

import (
	"time"

	"github.com/google/uuid"
)

// OfferAcceptance - запись о принятии оферты пользователем
type OfferAcceptance struct {
	ID           uuid.UUID `json:"id"`
	UserID       uuid.UUID `json:"user_id"`
	OfferVersion string    `json:"offer_version"`
	AcceptedAt   time.Time `json:"accepted_at"`
	IPAddress    string    `json:"ip_address"`
	UserAgent    string    `json:"user_agent"`
	CreatedAt    time.Time `json:"created_at"`
}
