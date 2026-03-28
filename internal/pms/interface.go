package pms

import (
	"context"
	"time"

	"github.com/nikitaaldaev/bani/internal/domain"
)

// PMSProvider defines the interface for external Property Management System integrations.
// Implementations handle communication with specific PMS platforms (Yclients, Restoplace).
type PMSProvider interface {
	// Name returns the provider identifier.
	Name() domain.PMSProvider

	// TestConnection verifies that the credentials are valid.
	TestConnection(ctx context.Context, credentials string) error

	// PullBookings retrieves bookings from the external PMS for a given date range.
	PullBookings(ctx context.Context, credentials string, externalID string, from, to time.Time) ([]domain.PMSBooking, error)

	// PushBooking sends a booking to the external PMS.
	PushBooking(ctx context.Context, credentials string, externalID string, booking domain.PMSBooking) error

	// SyncSchedule retrieves available schedule slots from the PMS.
	SyncSchedule(ctx context.Context, credentials string, externalID string, from, to time.Time) ([]domain.PMSScheduleSlot, error)

	// CancelBooking cancels a booking in the external PMS.
	CancelBooking(ctx context.Context, credentials string, externalID string, bookingExternalID string) error
}

// ProviderRegistry holds all available PMS providers for lookup by name.
type ProviderRegistry struct {
	providers map[domain.PMSProvider]PMSProvider
}

// NewProviderRegistry creates a registry from a slice of providers.
func NewProviderRegistry(providers []PMSProvider) *ProviderRegistry {
	m := make(map[domain.PMSProvider]PMSProvider, len(providers))
	for _, p := range providers {
		m[p.Name()] = p
	}
	return &ProviderRegistry{providers: m}
}

// Get returns the provider for the given name, or nil if not found.
func (r *ProviderRegistry) Get(name domain.PMSProvider) PMSProvider {
	return r.providers[name]
}

// Available returns the list of registered provider names.
func (r *ProviderRegistry) Available() []domain.PMSProvider {
	result := make([]domain.PMSProvider, 0, len(r.providers))
	for name := range r.providers {
		result = append(result, name)
	}
	return result
}
