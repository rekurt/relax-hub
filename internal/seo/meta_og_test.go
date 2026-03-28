package seo

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateMetaTags_OGFields(t *testing.T) {
	input := MetaInput{
		Name:         "Баня Люкс",
		CityName:     "Москве",
		CitySlug:     "moscow",
		Slug:         "banya-lyuks",
		PricePerHour: 200000,
		Rating:       4.5,
		ReviewCount:  10,
		Images:       []string{"https://example.com/photo.jpg"},
		BaseURL:      "https://bani.ru",
	}

	meta := GenerateMetaTags(input)

	assert.Equal(t, "https://bani.ru/moscow/banya-lyuks", meta.OGUrl)
	assert.Equal(t, "Bani.ru", meta.OGSiteName)
	assert.Equal(t, "ru_RU", meta.OGLocale)
	assert.Equal(t, "summary_large_image", meta.TwitterCard)
}

func TestGenerateMetaTags_TwitterCardSummary_WhenNoImage(t *testing.T) {
	input := MetaInput{
		Name: "Тест",
		Slug: "test",
	}

	meta := GenerateMetaTags(input)

	assert.Equal(t, "summary", meta.TwitterCard)
	assert.Empty(t, meta.OGImage)
}

func TestGenerateMetaTags_OGUrl_EqualsCanonical(t *testing.T) {
	input := MetaInput{
		Name:     "Сауна",
		CitySlug: "spb",
		Slug:     "sauna",
		BaseURL:  "https://bani.ru",
	}

	meta := GenerateMetaTags(input)

	assert.Equal(t, meta.Canonical, meta.OGUrl)
	assert.Equal(t, "https://bani.ru/spb/sauna", meta.OGUrl)
}
