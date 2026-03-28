package domain

import "time"

type BathhouseFilter struct {
	CityID            *int64           `json:"city_id,omitempty"`
	CitySlug          *string          `json:"city_slug,omitempty"`
	PriceMin          *int64           `json:"price_min,omitempty"`
	PriceMax          *int64           `json:"price_max,omitempty"`
	MinGuests         *int             `json:"min_guests,omitempty"`
	GuestCount        *int             `json:"guest_count,omitempty"`
	HasPool           *bool            `json:"has_pool,omitempty"`
	HasSauna          *bool            `json:"has_sauna,omitempty"`
	HasSteamRoom      *bool            `json:"has_steam_room,omitempty"`
	HasHotTub         *bool            `json:"has_hot_tub,omitempty"`
	HasBBQ            *bool            `json:"has_bbq,omitempty"`
	HasKaraoke        *bool            `json:"has_karaoke,omitempty"`
	MinRating         *float64         `json:"min_rating,omitempty"`
	Latitude          *float64         `json:"latitude,omitempty"`
	Longitude         *float64         `json:"longitude,omitempty"`
	RadiusKm          *float64         `json:"radius_km,omitempty"`
	AvailableDate     *time.Time       `json:"available_date,omitempty"`
	AvailableTimeFrom *string          `json:"available_time_from,omitempty"`
	AvailableTimeTo   *string          `json:"available_time_to,omitempty"`
	OpenNow           *bool            `json:"open_now,omitempty"`
	SearchQuery       *string          `json:"search_query,omitempty"`
	LastMinute        *bool            `json:"last_minute,omitempty"`
	Status            *BathhouseStatus `json:"status,omitempty"`
	ShowAllStatuses   bool             `json:"show_all_statuses,omitempty"` // when true, don't filter by status even if Status is nil
	IsochroneWKT      *string          `json:"isochrone_wkt,omitempty"`     // WKT POLYGON for isochrone-based filtering
	SortBy            string           `json:"sort_by,omitempty"`           // "relevance", "price_asc", "price_desc", "rating", "distance", "newest"
	SortOrder         string           `json:"sort_order,omitempty"`        // "asc", "desc"
	Page              int              `json:"page,omitempty"`
	PageSize          int              `json:"page_size,omitempty"`
}

type PaginatedResult[T any] struct {
	Items      []T
	TotalCount int64
	Page       int
	PageSize   int
	TotalPages int
}
