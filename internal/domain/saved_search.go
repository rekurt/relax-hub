package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// SavedSearch represents a user's saved search with filter criteria.
type SavedSearch struct {
	ID          uuid.UUID       `json:"id"`
	UserID      uuid.UUID       `json:"user_id"`
	Name        string          `json:"name"`
	Filters     json.RawMessage `json:"filters"`
	NotifyOnNew bool            `json:"notify_on_new"`
	CreatedAt   time.Time       `json:"created_at"`
}
