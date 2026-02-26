package domain

import "time"

type BathhouseFilter struct {
	CityID       *int64
	CitySlug     *string
	PriceMin     *int64
	PriceMax     *int64
	MinGuests    *int
	GuestCount   *int
	HasPool      *bool
	HasSauna     *bool
	HasSteamRoom *bool
	HasHotTub    *bool
	HasBBQ       *bool
	HasKaraoke   *bool
	MinRating    *float64
	Latitude     *float64
	Longitude    *float64
	RadiusKm     *float64
	AvailableDate     *time.Time
	AvailableTimeFrom *string
	AvailableTimeTo   *string
	OpenNow           *bool
	SearchQuery       *string
	Status          *BathhouseStatus
	ShowAllStatuses bool   // when true, don't filter by status even if Status is nil
	SortBy          string // "price", "rating", "distance"
	SortOrder    string // "asc", "desc"
	Page         int
	PageSize     int
}

type PaginatedResult[T any] struct {
	Items      []T
	TotalCount int64
	Page       int
	PageSize   int
	TotalPages int
}
