package domain

import (
	"time"

	"github.com/google/uuid"
)

type Favorite struct {
	ID          uuid.UUID
	UserID      uuid.UUID
	BathhouseID uuid.UUID
	CreatedAt   time.Time
}
