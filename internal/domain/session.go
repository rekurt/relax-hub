package domain

import (
	"time"

	"github.com/google/uuid"
)

type Session struct {
	ID           uuid.UUID
	UserID       uuid.UUID
	DeviceInfo   string
	Browser      string
	IP           string
	LastActiveAt time.Time
	CreatedAt    time.Time
	ExpiresAt    time.Time
}

const SessionMaxAge = 30 * 24 * time.Hour // 30 days of inactivity
