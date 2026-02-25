package domain

type BathhouseFilter struct {
	CityID       *int64
	CitySlug     *string
	PriceMin     *int64
	PriceMax     *int64
	MinGuests    *int
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
	Status       *BathhouseStatus
	SortBy       string // "price", "rating", "distance"
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
