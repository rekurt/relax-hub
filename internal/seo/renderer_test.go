package seo

import (
	"context"
	"strings"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRenderer_RenderPage(t *testing.T) {
	r := NewRenderer(nil, "https://bani.ru")

	data := PageData{
		Title:       "Баня Люкс в Москве — Bani.ru",
		Description: "Баня Люкс: сауна, бассейн. Цена от 2000 руб/ч.",
		Canonical:   "https://bani.ru/moscow/banya-lyuks",
		OGImage:     "https://bani.ru/images/photo.jpg",
		OGType:      "business.business",
		OGSiteName:  "Bani.ru",
		OGLocale:    "ru_RU",
		TwitterCard: "summary_large_image",
		SchemaJSON:  `{"@context":"https://schema.org","@type":"LocalBusiness","name":"Баня Люкс"}`,
		BodyContent: "<h1>Баня Люкс</h1>\n<p>Сауна и бассейн</p>",
	}

	html := r.RenderPage(data)

	// Check basic structure
	assert.True(t, strings.HasPrefix(html, "<!DOCTYPE html>"))
	assert.Contains(t, html, `<html lang="ru">`)
	assert.Contains(t, html, `<meta charset="utf-8">`)

	// Check title
	assert.Contains(t, html, "<title>Баня Люкс в Москве — Bani.ru</title>")

	// Check meta description
	assert.Contains(t, html, `<meta name="description" content="Баня Люкс: сауна, бассейн. Цена от 2000 руб/ч.">`)

	// Check canonical
	assert.Contains(t, html, `<link rel="canonical" href="https://bani.ru/moscow/banya-lyuks">`)

	// Check OG tags
	assert.Contains(t, html, `<meta property="og:type" content="business.business">`)
	assert.Contains(t, html, `<meta property="og:title" content="Баня Люкс в Москве — Bani.ru">`)
	assert.Contains(t, html, `<meta property="og:url" content="https://bani.ru/moscow/banya-lyuks">`)
	assert.Contains(t, html, `<meta property="og:image" content="https://bani.ru/images/photo.jpg">`)
	assert.Contains(t, html, `<meta property="og:site_name" content="Bani.ru">`)
	assert.Contains(t, html, `<meta property="og:locale" content="ru_RU">`)

	// Check Twitter Card
	assert.Contains(t, html, `<meta name="twitter:card" content="summary_large_image">`)
	assert.Contains(t, html, `<meta name="twitter:title" content="Баня Люкс в Москве — Bani.ru">`)

	// Check Schema.org JSON-LD
	assert.Contains(t, html, `<script type="application/ld+json">`)
	assert.Contains(t, html, `"@context":"https://schema.org"`)

	// Check body content
	assert.Contains(t, html, "<h1>Баня Люкс</h1>")
	assert.Contains(t, html, "<p>Сауна и бассейн</p>")
}

func TestRenderer_RenderPage_DefaultTwitterCard(t *testing.T) {
	r := NewRenderer(nil, "")

	data := PageData{
		Title: "Test",
	}

	html := r.RenderPage(data)
	assert.Contains(t, html, `<meta name="twitter:card" content="summary">`)
}

func TestRenderer_Cache(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	ctx := context.Background()

	r := NewRenderer(rdb, "https://bani.ru")

	// Cache miss
	_, ok := r.GetCachedPage(ctx, "detail:test-slug")
	assert.False(t, ok)

	// Cache set
	r.CachePage(ctx, "detail:test-slug", "<html>cached</html>")

	// Cache hit
	cached, ok := r.GetCachedPage(ctx, "detail:test-slug")
	assert.True(t, ok)
	assert.Equal(t, "<html>cached</html>", cached)

	// Invalidate
	r.InvalidateCache(ctx, "test-slug")

	// Cache miss after invalidation
	_, ok = r.GetCachedPage(ctx, "detail:test-slug")
	assert.False(t, ok)
}

func TestRenderer_CacheWithNilRedis(t *testing.T) {
	r := NewRenderer(nil, "https://bani.ru")
	ctx := context.Background()

	// Should not panic with nil redis
	_, ok := r.GetCachedPage(ctx, "key")
	assert.False(t, ok)

	r.CachePage(ctx, "key", "content")
	r.InvalidateCache(ctx, "slug")
	r.InvalidateCityCache(ctx, "city")
	r.InvalidateListingCache(ctx)
}

func TestRenderer_InvalidateCityCache(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	ctx := context.Background()

	r := NewRenderer(rdb, "https://bani.ru")

	r.CachePage(ctx, "city:moscow", "<html>moscow</html>")
	cached, ok := r.GetCachedPage(ctx, "city:moscow")
	assert.True(t, ok)
	assert.Equal(t, "<html>moscow</html>", cached)

	r.InvalidateCityCache(ctx, "moscow")
	_, ok = r.GetCachedPage(ctx, "city:moscow")
	assert.False(t, ok)
}

func TestRenderer_BuildBathhouseDetailPage(t *testing.T) {
	r := NewRenderer(nil, "https://bani.ru")

	schemaInput := SchemaInput{
		Name:         "Баня Люкс",
		Description:  "Лучшая баня",
		CityName:     "Москве",
		CitySlug:     "moscow",
		Slug:         "banya-lyuks",
		PricePerHour: 200000,
		Rating:       4.5,
		ReviewCount:  10,
		HasSauna:     true,
		HasPool:      true,
		BaseURL:      "https://bani.ru",
	}

	meta := MetaTags{
		Title:       "Баня Люкс в Москве — Bani.ru",
		Description: "Баня Люкс: сауна, бассейн. Цена от 2000 руб/ч.",
		Canonical:   "https://bani.ru/moscow/banya-lyuks",
		OGType:      "business.business",
		OGSiteName:  "Bani.ru",
		OGLocale:    "ru_RU",
		TwitterCard: "summary",
	}

	page := r.BuildBathhouseDetailPage(schemaInput, meta)

	assert.Equal(t, meta.Title, page.Title)
	assert.Equal(t, meta.Description, page.Description)
	assert.Contains(t, page.SchemaJSON, `"@context":"https://schema.org"`)
	assert.Contains(t, page.BodyContent, "Баня Люкс")
	assert.Contains(t, page.BodyContent, "Москве")
	assert.Contains(t, page.BodyContent, "2000 руб/ч")
	assert.Contains(t, page.BodyContent, "4.5")
	assert.Contains(t, page.BodyContent, "сауна")
	assert.Contains(t, page.BodyContent, "бассейн")
}

func TestRenderer_BuildCityListingPage(t *testing.T) {
	r := NewRenderer(nil, "https://bani.ru")

	page := r.BuildCityListingPage("Москве", "moscow", 42)

	assert.Contains(t, page.Title, "Москве")
	assert.Contains(t, page.Description, "42")
	assert.Equal(t, "https://bani.ru/moscow", page.Canonical)
	assert.Equal(t, "website", page.OGType)
}

func TestRenderer_BuildMainListingPage(t *testing.T) {
	r := NewRenderer(nil, "https://bani.ru")

	page := r.BuildMainListingPage(150)

	assert.Contains(t, page.Title, "Бани и сауны")
	assert.Contains(t, page.Description, "150")
	assert.Equal(t, "https://bani.ru", page.Canonical)
}

func TestRenderer_BuildReviewsPage(t *testing.T) {
	r := NewRenderer(nil, "https://bani.ru")

	reviews := []string{"Отличная баня!", "Рекомендую всем", "Супер место"}
	page := r.BuildReviewsPage("Баня Люкс", "Москве", "moscow", "banya-lyuks", 4.5, 10, reviews)

	assert.Contains(t, page.Title, "Отзывы")
	assert.Contains(t, page.Title, "Баня Люкс")
	assert.Contains(t, page.Description, "4.5")
	assert.Contains(t, page.Description, "10 отзывов")
	assert.Contains(t, page.BodyContent, "Отличная баня!")
	assert.Contains(t, page.BodyContent, "Рекомендую всем")
}

func TestRenderer_BuildReviewsPage_LimitsTo10(t *testing.T) {
	r := NewRenderer(nil, "https://bani.ru")

	reviews := make([]string, 15)
	for i := range reviews {
		reviews[i] = "Review text"
	}
	page := r.BuildReviewsPage("Test", "", "", "test", 4.0, 15, reviews)

	// Count blockquote tags - should be max 10
	count := strings.Count(page.BodyContent, "<blockquote>")
	assert.Equal(t, 10, count)
}

func TestCacheKeys(t *testing.T) {
	assert.Equal(t, "detail:test-slug", CacheKeyForDetail("test-slug"))
	assert.Equal(t, "city:moscow", CacheKeyForCity("moscow"))
	assert.Equal(t, "reviews:test-slug", CacheKeyForReviews("test-slug"))
	assert.Equal(t, "listing", CacheKeyForListing())
}

func TestRenderer_DefaultBaseURL(t *testing.T) {
	r := NewRenderer(nil, "")
	assert.Equal(t, "https://bani.ru", r.baseURL)
}
