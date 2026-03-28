package seo

import (
	"fmt"
	"strings"
)

// MetaTags contains SEO meta-tags for a bathhouse page.
type MetaTags struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	OGImage     string `json:"og_image"`
	OGType      string `json:"og_type"`
	OGUrl       string `json:"og_url,omitempty"`
	OGSiteName  string `json:"og_site_name,omitempty"`
	OGLocale    string `json:"og_locale,omitempty"`
	TwitterCard string `json:"twitter_card,omitempty"`
	Canonical   string `json:"canonical"`
}

// MetaInput holds the data needed to generate meta-tags.
type MetaInput struct {
	Name         string
	CityName     string
	CitySlug     string
	Slug         string
	Description  string
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
	BaseURL      string // e.g. "https://bani.ru"
}

const siteName = "Bani.ru"

// GenerateMetaTags creates SEO meta-tags from bathhouse data.
func GenerateMetaTags(input MetaInput) MetaTags {
	title := generateTitle(input)
	description := generateDescription(input)

	var ogImage string
	if len(input.Images) > 0 {
		ogImage = input.Images[0]
	}

	canonical := generateCanonical(input)

	twitterCard := "summary"
	if ogImage != "" {
		twitterCard = "summary_large_image"
	}

	return MetaTags{
		Title:       title,
		Description: description,
		OGImage:     ogImage,
		OGType:      "business.business",
		OGUrl:       canonical,
		OGSiteName:  siteName,
		OGLocale:    "ru_RU",
		TwitterCard: twitterCard,
		Canonical:   canonical,
	}
}

func generateTitle(input MetaInput) string {
	if input.CityName != "" {
		return fmt.Sprintf("%s в %s — %s", input.Name, input.CityName, siteName)
	}
	return fmt.Sprintf("%s — %s", input.Name, siteName)
}

func generateDescription(input MetaInput) string {
	var parts []string

	// Name with amenities
	amenities := collectAmenities(input)
	if len(amenities) > 0 {
		parts = append(parts, fmt.Sprintf("%s: %s.", input.Name, strings.Join(amenities, ", ")))
	} else {
		parts = append(parts, fmt.Sprintf("%s.", input.Name))
	}

	// Price
	if input.PricePerHour > 0 {
		rubles := input.PricePerHour / 100
		parts = append(parts, fmt.Sprintf("Цена от %d руб/ч.", rubles))
	}

	// Rating
	if input.Rating > 0 && input.ReviewCount > 0 {
		parts = append(parts, fmt.Sprintf("Рейтинг %.1f.", input.Rating))
	}

	parts = append(parts, "Бронируйте онлайн.")

	return strings.Join(parts, " ")
}

func collectAmenities(input MetaInput) []string {
	var amenities []string
	if input.HasSteamRoom {
		amenities = append(amenities, "русская баня")
	}
	if input.HasSauna {
		amenities = append(amenities, "сауна")
	}
	if input.HasPool {
		amenities = append(amenities, "бассейн")
	}
	if input.HasHotTub {
		amenities = append(amenities, "джакузи")
	}
	if input.HasBBQ {
		amenities = append(amenities, "мангал")
	}
	if input.HasKaraoke {
		amenities = append(amenities, "караоке")
	}
	return amenities
}

func generateCanonical(input MetaInput) string {
	base := input.BaseURL
	if base == "" {
		base = "https://bani.ru"
	}
	// Remove trailing slash
	base = strings.TrimRight(base, "/")

	if input.CitySlug != "" && input.Slug != "" {
		return fmt.Sprintf("%s/%s/%s", base, input.CitySlug, input.Slug)
	}
	if input.Slug != "" {
		return fmt.Sprintf("%s/bathhouses/%s", base, input.Slug)
	}
	return ""
}
