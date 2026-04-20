package seo

import (
	"testing"
)

func TestGenerateSchema_FullData(t *testing.T) {
	input := SchemaInput{
		Name:         "Баня на Липовой",
		Description:  "Лучшая баня в городе",
		Slug:         "banya-na-lipovoy",
		CityName:     "Москва",
		CitySlug:     "moscow",
		Address:      "ул. Липовая, 10",
		Latitude:     55.7558,
		Longitude:    37.6173,
		PricePerHour: 250000, // 2500 rubles
		Rating:       4.8,
		ReviewCount:  42,
		Images:       []string{"https://example.com/img1.jpg", "https://example.com/img2.jpg"},
		HasPool:      true,
		HasSauna:     true,
		HasSteamRoom: true,
		HasHotTub:    false,
		HasBBQ:       true,
		HasKaraoke:   false,
		WorkingHours: []WorkingHoursInput{
			{DayOfWeek: 0, OpenTime: "09:00", CloseTime: "23:00"},
			{DayOfWeek: 5, OpenTime: "10:00", CloseTime: "22:00"},
		},
		BaseURL: "https://bani.ru",
	}

	schema := GenerateSchema(input)

	if schema.Context != "https://schema.org" {
		t.Errorf("expected context https://schema.org, got %s", schema.Context)
	}
	if schema.Type != "LocalBusiness" {
		t.Errorf("expected type LocalBusiness, got %s", schema.Type)
	}
	if schema.Name != "Баня на Липовой" {
		t.Errorf("expected name 'Баня на Липовой', got %s", schema.Name)
	}
	if schema.URL != "https://bani.ru/moscow/banya-na-lipovoy" {
		t.Errorf("expected URL with city slug, got %s", schema.URL)
	}
	if len(schema.Image) != 2 {
		t.Errorf("expected 2 images, got %d", len(schema.Image))
	}
	if schema.Address == nil {
		t.Fatal("expected non-nil address")
	}
	if schema.Address.AddressLocality != "Москва" {
		t.Errorf("expected addressLocality 'Москва', got %s", schema.Address.AddressLocality)
	}
	if schema.Address.StreetAddress != "ул. Липовая, 10" {
		t.Errorf("expected streetAddress, got %s", schema.Address.StreetAddress)
	}
	if schema.Geo == nil {
		t.Fatal("expected non-nil geo")
	}
	if schema.Geo.Latitude != 55.7558 {
		t.Errorf("expected lat 55.7558, got %f", schema.Geo.Latitude)
	}
	if schema.AggregateRating == nil {
		t.Fatal("expected non-nil aggregateRating")
	}
	if schema.AggregateRating.RatingValue != "4.8" {
		t.Errorf("expected ratingValue 4.8, got %s", schema.AggregateRating.RatingValue)
	}
	if schema.AggregateRating.ReviewCount != "42" {
		t.Errorf("expected reviewCount 42, got %s", schema.AggregateRating.ReviewCount)
	}
	if schema.PriceRange != "$$" {
		t.Errorf("expected priceRange $$, got %s", schema.PriceRange)
	}
	if len(schema.OpeningHours) != 2 {
		t.Errorf("expected 2 opening hours, got %d", len(schema.OpeningHours))
	}
	if schema.OpeningHours[0].DayOfWeek != "Monday" {
		t.Errorf("expected Monday, got %s", schema.OpeningHours[0].DayOfWeek)
	}
	if schema.OpeningHours[1].DayOfWeek != "Saturday" {
		t.Errorf("expected Saturday, got %s", schema.OpeningHours[1].DayOfWeek)
	}
	// 3 amenities: steam room, sauna, pool, bbq = 4
	if len(schema.AmenityFeature) != 4 {
		t.Errorf("expected 4 amenities, got %d", len(schema.AmenityFeature))
	}
}

func TestGenerateSchema_NoRating(t *testing.T) {
	input := SchemaInput{
		Name:    "Тест",
		Rating:  0,
		BaseURL: "https://bani.ru",
	}

	schema := GenerateSchema(input)

	if schema.AggregateRating != nil {
		t.Error("expected nil aggregateRating when no rating")
	}
}

func TestGenerateSchema_NoGeo(t *testing.T) {
	input := SchemaInput{
		Name: "Тест",
	}

	schema := GenerateSchema(input)

	if schema.Geo != nil {
		t.Error("expected nil geo when coordinates are zero")
	}
}

func TestGenerateSchema_PriceRanges(t *testing.T) {
	tests := []struct {
		name         string
		pricePerHour int64
		expected     string
	}{
		{"cheap", 100000, "$"},       // 1000 rub
		{"mid", 200000, "$$"},        // 2000 rub
		{"expensive", 400000, "$$$"}, // 4000 rub
		{"luxury", 600000, "$$$$"},   // 6000 rub
		{"zero", 0, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schema := GenerateSchema(SchemaInput{
				Name:         "Тест",
				PricePerHour: tt.pricePerHour,
			})
			if schema.PriceRange != tt.expected {
				t.Errorf("expected priceRange %q, got %q", tt.expected, schema.PriceRange)
			}
		})
	}
}

func TestGenerateSchema_URLWithoutCity(t *testing.T) {
	input := SchemaInput{
		Name:    "Тест",
		Slug:    "test-banya",
		BaseURL: "https://bani.ru",
	}

	schema := GenerateSchema(input)

	if schema.URL != "https://bani.ru/bathhouses/test-banya" {
		t.Errorf("expected URL without city, got %s", schema.URL)
	}
}

func TestGenerateSchema_NoSlug(t *testing.T) {
	input := SchemaInput{
		Name:    "Тест",
		BaseURL: "https://bani.ru",
	}

	schema := GenerateSchema(input)

	if schema.URL != "https://bani.ru" {
		t.Errorf("expected base URL when no slug, got %s", schema.URL)
	}
}

func TestGenerateSchema_AllAmenities(t *testing.T) {
	input := SchemaInput{
		Name:         "Тест",
		HasPool:      true,
		HasSauna:     true,
		HasSteamRoom: true,
		HasHotTub:    true,
		HasBBQ:       true,
		HasKaraoke:   true,
	}

	schema := GenerateSchema(input)

	if len(schema.AmenityFeature) != 6 {
		t.Errorf("expected 6 amenities, got %d", len(schema.AmenityFeature))
	}
}

func TestGenerateSchema_NoAmenities(t *testing.T) {
	input := SchemaInput{
		Name: "Тест",
	}

	schema := GenerateSchema(input)

	if len(schema.AmenityFeature) != 0 {
		t.Errorf("expected 0 amenities, got %d", len(schema.AmenityFeature))
	}
}

func TestGenerateSchema_InvalidDayOfWeek(t *testing.T) {
	input := SchemaInput{
		Name: "Тест",
		WorkingHours: []WorkingHoursInput{
			{DayOfWeek: 99, OpenTime: "09:00", CloseTime: "23:00"},
		},
	}

	schema := GenerateSchema(input)

	if len(schema.OpeningHours) != 0 {
		t.Errorf("expected 0 opening hours for invalid day, got %d", len(schema.OpeningHours))
	}
}

func TestGenerateSchema_RelativeImagesUseBaseURL(t *testing.T) {
	input := SchemaInput{
		Name:    "Тест",
		Slug:    "test-banya",
		BaseURL: "https://bani.ru",
		Images:  []string{"/demo/bathhouses/steam-room-birch.png"},
	}

	schema := GenerateSchema(input)

	if len(schema.Image) != 1 {
		t.Fatalf("expected 1 schema image, got %d", len(schema.Image))
	}
	if schema.Image[0] != "https://bani.ru/demo/bathhouses/steam-room-birch.png" {
		t.Fatalf("schema image = %q, want absolute base-url image", schema.Image[0])
	}
}
