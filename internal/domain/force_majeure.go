package domain

import (
	"time"

	"github.com/google/uuid"
)

type ForceMajeureEvent struct {
	ID            uuid.UUID
	AdminID       uuid.UUID
	Region        string
	DateFrom      time.Time
	DateTo        time.Time
	Reason        string
	AffectedCount int
	TotalRefund   int64
	CreatedAt     time.Time
}
