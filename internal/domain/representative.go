package domain

import (
	"time"

	"github.com/google/uuid"
)

type Representative struct {
	ID          uuid.UUID
	UserID      uuid.UUID
	BathhouseID uuid.UUID
	OwnerID     uuid.UUID
	CreatedAt   time.Time
}
