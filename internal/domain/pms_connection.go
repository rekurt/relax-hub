package domain

import (
	"time"

	"github.com/google/uuid"
)

type PMSProvider string

const (
	PMSProviderYclients   PMSProvider = "yclients"
	PMSProviderRestoplace PMSProvider = "restoplace"
)

func AllPMSProviders() []PMSProvider {
	return []PMSProvider{PMSProviderYclients, PMSProviderRestoplace}
}

func (p PMSProvider) IsValid() bool {
	for _, v := range AllPMSProviders() {
		if v == p {
			return true
		}
	}
	return false
}

type PMSConnectionStatus string

const (
	PMSConnectionActive   PMSConnectionStatus = "active"
	PMSConnectionInactive PMSConnectionStatus = "inactive"
	PMSConnectionError    PMSConnectionStatus = "error"
)

type PMSSyncDirection string

const (
	PMSSyncDirectionInbound  PMSSyncDirection = "inbound"  // PMS -> RelaxHub
	PMSSyncDirectionOutbound PMSSyncDirection = "outbound" // RelaxHub -> PMS
	PMSSyncDirectionBoth     PMSSyncDirection = "both"
)

const (
	MaxPMSConnectionsPerBathhouse = 1
	DefaultPMSSyncIntervalMin     = 15
)

type PMSConnection struct {
	ID                   uuid.UUID
	OwnerID              uuid.UUID
	BathhouseID          uuid.UUID
	Provider             PMSProvider
	CredentialsEncrypted string
	SyncDirection        PMSSyncDirection
	SyncIntervalMinutes  int
	Status               PMSConnectionStatus
	LastSyncAt           *time.Time
	LastSyncError        string
	ExternalID           string // Bathhouse ID in the external PMS
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

func (c *PMSConnection) Validate() error {
	if c.OwnerID == uuid.Nil {
		return ErrInvalidInput
	}
	if c.BathhouseID == uuid.Nil {
		return ErrInvalidInput
	}
	if !c.Provider.IsValid() {
		return ErrInvalidInput
	}
	if c.CredentialsEncrypted == "" {
		return ErrInvalidInput
	}
	if c.SyncIntervalMinutes < 5 || c.SyncIntervalMinutes > 1440 {
		return ErrInvalidInput
	}
	return nil
}

type PMSSyncLog struct {
	ID           uuid.UUID
	ConnectionID uuid.UUID
	Direction    PMSSyncDirection
	Status       string // success, error
	ItemsSynced  int
	ErrorMessage string
	StartedAt    time.Time
	CompletedAt  time.Time
}

type PMSBooking struct {
	ExternalID string
	StartTime  time.Time
	EndTime    time.Time
	GuestName  string
	GuestPhone string
	Status     string
	Notes      string
}

type PMSScheduleSlot struct {
	Date      string // YYYY-MM-DD
	TimeFrom  string // HH:MM
	TimeTo    string // HH:MM
	Available bool
}
