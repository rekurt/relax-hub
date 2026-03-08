package seo

import (
	"testing"
)

func TestGenerateMetaTags_FullData(t *testing.T) {
	input := MetaInput{
		Name:         "Баня на Липовой",
		CityName:     "Москве",
		CitySlug:     "moscow",
		Slug:         "banya-na-lipovoy",
		Description:  "Уютная баня с бассейном",
		PricePerHour: 200000, // 2000 rubles in kopecks
		Rating:       4.8,
		ReviewCount:  42,
		Images:       []string{"https://example.com/photo1.jpg", "https://example.com/photo2.jpg"},
		HasPool:      true,
		HasSauna:     true,
		HasSteamRoom: true,
		HasBBQ:       true,
		BaseURL:      "https://bani.ru",
	}

	meta := GenerateMetaTags(input)

	// Title should include city and site name
	expectedTitle := "Баня на Липовой в Москве — Bani.ru"
	if meta.Title != expectedTitle {
		t.Errorf("Title = %q, want %q", meta.Title, expectedTitle)
	}

	// Description should include amenities, price, rating
	if meta.Description == "" {
		t.Error("Description should not be empty")
	}
	if !containsSubstring(meta.Description, "русская баня") {
		t.Errorf("Description should contain 'русская баня', got %q", meta.Description)
	}
	if !containsSubstring(meta.Description, "бассейн") {
		t.Errorf("Description should contain 'бассейн', got %q", meta.Description)
	}
	if !containsSubstring(meta.Description, "2000 руб/ч") {
		t.Errorf("Description should contain '2000 руб/ч', got %q", meta.Description)
	}
	if !containsSubstring(meta.Description, "4.8") {
		t.Errorf("Description should contain rating '4.8', got %q", meta.Description)
	}
	if !containsSubstring(meta.Description, "Бронируйте онлайн") {
		t.Errorf("Description should contain 'Бронируйте онлайн', got %q", meta.Description)
	}

	// OGImage should be first image
	if meta.OGImage != "https://example.com/photo1.jpg" {
		t.Errorf("OGImage = %q, want first image", meta.OGImage)
	}

	// OGType
	if meta.OGType != "business.business" {
		t.Errorf("OGType = %q, want 'business.business'", meta.OGType)
	}

	// Canonical URL
	expectedCanonical := "https://bani.ru/moscow/banya-na-lipovoy"
	if meta.Canonical != expectedCanonical {
		t.Errorf("Canonical = %q, want %q", meta.Canonical, expectedCanonical)
	}
}

func TestGenerateMetaTags_WithoutCity(t *testing.T) {
	input := MetaInput{
		Name:         "Сауна Люкс",
		Slug:         "sauna-lyuks",
		PricePerHour: 300000,
		BaseURL:      "https://bani.ru",
	}

	meta := GenerateMetaTags(input)

	expectedTitle := "Сауна Люкс — Bani.ru"
	if meta.Title != expectedTitle {
		t.Errorf("Title = %q, want %q", meta.Title, expectedTitle)
	}

	// Without city, canonical should use /bathhouses/ path
	expectedCanonical := "https://bani.ru/bathhouses/sauna-lyuks"
	if meta.Canonical != expectedCanonical {
		t.Errorf("Canonical = %q, want %q", meta.Canonical, expectedCanonical)
	}
}

func TestGenerateMetaTags_NoImages(t *testing.T) {
	input := MetaInput{
		Name: "Тестовая баня",
		Slug: "testovaya-banya",
	}

	meta := GenerateMetaTags(input)

	if meta.OGImage != "" {
		t.Errorf("OGImage should be empty when no images, got %q", meta.OGImage)
	}
}

func TestGenerateMetaTags_NoRating(t *testing.T) {
	input := MetaInput{
		Name:         "Новая баня",
		Slug:         "novaya-banya",
		PricePerHour: 150000,
	}

	meta := GenerateMetaTags(input)

	// Description should not contain rating info
	if containsSubstring(meta.Description, "Рейтинг") {
		t.Errorf("Description should not contain rating when rating is 0, got %q", meta.Description)
	}
}

func TestGenerateMetaTags_AllAmenities(t *testing.T) {
	input := MetaInput{
		Name:         "Суперкомплекс",
		Slug:         "superkompleks",
		HasPool:      true,
		HasSauna:     true,
		HasSteamRoom: true,
		HasHotTub:    true,
		HasBBQ:       true,
		HasKaraoke:   true,
	}

	meta := GenerateMetaTags(input)

	for _, amenity := range []string{"русская баня", "сауна", "бассейн", "джакузи", "мангал", "караоке"} {
		if !containsSubstring(meta.Description, amenity) {
			t.Errorf("Description should contain %q, got %q", amenity, meta.Description)
		}
	}
}

func TestGenerateMetaTags_DefaultBaseURL(t *testing.T) {
	input := MetaInput{
		Name:     "Тест",
		CitySlug: "spb",
		Slug:     "test",
	}

	meta := GenerateMetaTags(input)

	expectedCanonical := "https://bani.ru/spb/test"
	if meta.Canonical != expectedCanonical {
		t.Errorf("Canonical = %q, want %q", meta.Canonical, expectedCanonical)
	}
}

func TestGenerateMetaTags_NoSlug(t *testing.T) {
	input := MetaInput{
		Name: "Баня без слага",
	}

	meta := GenerateMetaTags(input)

	if meta.Canonical != "" {
		t.Errorf("Canonical should be empty when no slug, got %q", meta.Canonical)
	}
}

func containsSubstring(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsCheck(s, substr))
}

func containsCheck(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
