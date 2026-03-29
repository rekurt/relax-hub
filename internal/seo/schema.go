package seo

import (
	"fmt"
	"strings"
)

// SchemaLocalBusiness represents a Schema.org LocalBusiness JSON-LD object.
type SchemaLocalBusiness struct {
	Context         string                 `json:"@context"`
	Type            string                 `json:"@type"`
	Name            string                 `json:"name"`
	Description     string                 `json:"description,omitempty"`
	URL             string                 `json:"url,omitempty"`
	Image           []string               `json:"image,omitempty"`
	Address         *SchemaPostalAddress   `json:"address,omitempty"`
	Geo             *SchemaGeoCoordinates  `json:"geo,omitempty"`
	AggregateRating *SchemaAggregateRating `json:"aggregateRating,omitempty"`
	PriceRange      string                 `json:"priceRange,omitempty"`
	OpeningHours    []SchemaOpeningHours   `json:"openingHoursSpecification,omitempty"`
	AmenityFeature  []SchemaAmenity        `json:"amenityFeature,omitempty"`
}

// SchemaPostalAddress represents a Schema.org PostalAddress.
type SchemaPostalAddress struct {
	Type            string `json:"@type"`
	StreetAddress   string `json:"streetAddress,omitempty"`
	AddressLocality string `json:"addressLocality,omitempty"`
}

// SchemaGeoCoordinates represents Schema.org GeoCoordinates.
type SchemaGeoCoordinates struct {
	Type      string  `json:"@type"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

// SchemaAggregateRating represents Schema.org AggregateRating.
type SchemaAggregateRating struct {
	Type        string `json:"@type"`
	RatingValue string `json:"ratingValue"`
	ReviewCount string `json:"reviewCount"`
}

// SchemaOpeningHours represents Schema.org OpeningHoursSpecification.
type SchemaOpeningHours struct {
	Type      string `json:"@type"`
	DayOfWeek string `json:"dayOfWeek"`
	Opens     string `json:"opens"`
	Closes    string `json:"closes"`
}

// SchemaAmenity represents Schema.org LocationFeatureSpecification.
type SchemaAmenity struct {
	Type  string `json:"@type"`
	Name  string `json:"name"`
	Value bool   `json:"value"`
}

// SchemaInput holds the data needed to generate Schema.org JSON-LD.
type SchemaInput struct {
	Name         string
	Description  string
	Slug         string
	CityName     string
	CitySlug     string
	Address      string
	Latitude     float64
	Longitude    float64
	PricePerHour int64 // kopecks
	Rating       float64
	ReviewCount  int
	Images       []string
	HasPool      bool
	HasSauna     bool
	HasSteamRoom bool
	HasHotTub    bool
	HasBBQ       bool
	HasKaraoke   bool
	WorkingHours []WorkingHoursInput
	BaseURL      string
}

// WorkingHoursInput represents working hours for schema generation.
type WorkingHoursInput struct {
	DayOfWeek int    // 0=Mon, 6=Sun
	OpenTime  string // "09:00"
	CloseTime string // "23:00"
}

// dayOfWeekMap maps 0=Mon..6=Sun to Schema.org day names.
var dayOfWeekMap = map[int]string{
	0: "Monday",
	1: "Tuesday",
	2: "Wednesday",
	3: "Thursday",
	4: "Friday",
	5: "Saturday",
	6: "Sunday",
}

// GenerateSchema creates a Schema.org LocalBusiness JSON-LD object.
func GenerateSchema(input SchemaInput) SchemaLocalBusiness {
	schema := SchemaLocalBusiness{
		Context:     "https://schema.org",
		Type:        "LocalBusiness",
		Name:        input.Name,
		Description: input.Description,
	}

	// URL
	schema.URL = generateSchemaURL(input)

	// Images
	if len(input.Images) > 0 {
		schema.Image = input.Images
	}

	// Address
	if input.Address != "" || input.CityName != "" {
		schema.Address = &SchemaPostalAddress{
			Type:            "PostalAddress",
			StreetAddress:   input.Address,
			AddressLocality: input.CityName,
		}
	}

	// Geo
	if input.Latitude != 0 || input.Longitude != 0 {
		schema.Geo = &SchemaGeoCoordinates{
			Type:      "GeoCoordinates",
			Latitude:  input.Latitude,
			Longitude: input.Longitude,
		}
	}

	// AggregateRating
	if input.Rating > 0 && input.ReviewCount > 0 {
		schema.AggregateRating = &SchemaAggregateRating{
			Type:        "AggregateRating",
			RatingValue: fmt.Sprintf("%.1f", input.Rating),
			ReviewCount: fmt.Sprintf("%d", input.ReviewCount),
		}
	}

	// PriceRange
	if input.PricePerHour > 0 {
		rubles := input.PricePerHour / 100
		switch {
		case rubles < 1500:
			schema.PriceRange = "$"
		case rubles < 3000:
			schema.PriceRange = "$$"
		case rubles < 5000:
			schema.PriceRange = "$$$"
		default:
			schema.PriceRange = "$$$$"
		}
	}

	// Opening hours
	for _, wh := range input.WorkingHours {
		dayName, ok := dayOfWeekMap[wh.DayOfWeek]
		if !ok {
			continue
		}
		schema.OpeningHours = append(schema.OpeningHours, SchemaOpeningHours{
			Type:      "OpeningHoursSpecification",
			DayOfWeek: dayName,
			Opens:     wh.OpenTime,
			Closes:    wh.CloseTime,
		})
	}

	// Amenities
	amenities := []struct {
		name string
		has  bool
	}{
		{"Русская баня", input.HasSteamRoom},
		{"Сауна", input.HasSauna},
		{"Бассейн", input.HasPool},
		{"Джакузи", input.HasHotTub},
		{"Мангал", input.HasBBQ},
		{"Караоке", input.HasKaraoke},
	}
	for _, a := range amenities {
		if a.has {
			schema.AmenityFeature = append(schema.AmenityFeature, SchemaAmenity{
				Type:  "LocationFeatureSpecification",
				Name:  a.name,
				Value: true,
			})
		}
	}

	return schema
}

func generateSchemaURL(input SchemaInput) string {
	base := input.BaseURL
	if base == "" {
		base = "https://bani.ru"
	}
	base = strings.TrimRight(base, "/")
	if input.CitySlug != "" && input.Slug != "" {
		return fmt.Sprintf("%s/%s/%s", base, input.CitySlug, input.Slug)
	}
	if input.Slug != "" {
		return fmt.Sprintf("%s/bathhouses/%s", base, input.Slug)
	}
	return base
}
