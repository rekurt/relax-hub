package seo

import (
	"context"
	"encoding/json"
	"fmt"
	"html"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	prerenderCachePrefix = "seo:prerender:"
	prerenderCacheTTL    = 1 * time.Hour
)

// PageType identifies the type of public page being pre-rendered.
type PageType string

const (
	PageTypeBathhouseList   PageType = "listing"
	PageTypeBathhouseDetail PageType = "detail"
	PageTypeReviews         PageType = "reviews"
	PageTypeCityListing     PageType = "city"
)

// PageData holds the data needed to render a pre-rendered HTML page.
type PageData struct {
	Title       string
	Description string
	Canonical   string
	OGImage     string
	OGType      string
	OGSiteName  string
	OGLocale    string
	TwitterCard string
	SchemaJSON  string // JSON-LD string
	BodyContent string // visible text content for crawlers
}

// Renderer generates pre-rendered HTML pages for search engine bots.
type Renderer struct {
	redis   *redis.Client
	baseURL string
}

// NewRenderer creates a new pre-render service.
func NewRenderer(redisClient *redis.Client, baseURL string) *Renderer {
	if baseURL == "" {
		baseURL = "https://bani.ru"
	}
	return &Renderer{
		redis:   redisClient,
		baseURL: strings.TrimRight(baseURL, "/"),
	}
}

// RenderPage generates a full HTML page from PageData.
func (r *Renderer) RenderPage(data PageData) string {
	var b strings.Builder

	b.WriteString("<!DOCTYPE html>\n<html lang=\"ru\">\n<head>\n")
	b.WriteString("<meta charset=\"utf-8\">\n")
	b.WriteString("<meta name=\"viewport\" content=\"width=device-width, initial-scale=1\">\n")

	// Title
	fmt.Fprintf(&b, "<title>%s</title>\n", html.EscapeString(data.Title))

	// Meta description
	if data.Description != "" {
		fmt.Fprintf(&b, "<meta name=\"description\" content=\"%s\">\n", html.EscapeString(data.Description))
	}

	// Canonical
	if data.Canonical != "" {
		fmt.Fprintf(&b, "<link rel=\"canonical\" href=\"%s\">\n", html.EscapeString(data.Canonical))
	}

	// Open Graph
	if data.OGType != "" {
		fmt.Fprintf(&b, "<meta property=\"og:type\" content=\"%s\">\n", html.EscapeString(data.OGType))
	}
	fmt.Fprintf(&b, "<meta property=\"og:title\" content=\"%s\">\n", html.EscapeString(data.Title))
	if data.Description != "" {
		fmt.Fprintf(&b, "<meta property=\"og:description\" content=\"%s\">\n", html.EscapeString(data.Description))
	}
	if data.Canonical != "" {
		fmt.Fprintf(&b, "<meta property=\"og:url\" content=\"%s\">\n", html.EscapeString(data.Canonical))
	}
	if data.OGImage != "" {
		fmt.Fprintf(&b, "<meta property=\"og:image\" content=\"%s\">\n", html.EscapeString(data.OGImage))
	}
	if data.OGSiteName != "" {
		fmt.Fprintf(&b, "<meta property=\"og:site_name\" content=\"%s\">\n", html.EscapeString(data.OGSiteName))
	}
	if data.OGLocale != "" {
		fmt.Fprintf(&b, "<meta property=\"og:locale\" content=\"%s\">\n", html.EscapeString(data.OGLocale))
	}

	// Twitter Card
	tc := data.TwitterCard
	if tc == "" {
		tc = "summary"
	}
	fmt.Fprintf(&b, "<meta name=\"twitter:card\" content=\"%s\">\n", html.EscapeString(tc))
	fmt.Fprintf(&b, "<meta name=\"twitter:title\" content=\"%s\">\n", html.EscapeString(data.Title))
	if data.Description != "" {
		fmt.Fprintf(&b, "<meta name=\"twitter:description\" content=\"%s\">\n", html.EscapeString(data.Description))
	}
	if data.OGImage != "" {
		fmt.Fprintf(&b, "<meta name=\"twitter:image\" content=\"%s\">\n", html.EscapeString(data.OGImage))
	}

	// Schema.org JSON-LD
	if data.SchemaJSON != "" {
		fmt.Fprintf(&b, "<script type=\"application/ld+json\">%s</script>\n", data.SchemaJSON)
	}

	b.WriteString("</head>\n<body>\n")

	// Body content for crawlers
	if data.BodyContent != "" {
		b.WriteString(data.BodyContent)
	}

	b.WriteString("\n</body>\n</html>")

	return b.String()
}

// GetCachedPage tries to retrieve a pre-rendered page from Redis cache.
func (r *Renderer) GetCachedPage(ctx context.Context, cacheKey string) (string, bool) {
	if r.redis == nil {
		return "", false
	}
	cached, err := r.redis.Get(ctx, prerenderCachePrefix+cacheKey).Result()
	if err != nil {
		return "", false
	}
	return cached, true
}

// CachePage stores a pre-rendered page in Redis.
func (r *Renderer) CachePage(ctx context.Context, cacheKey string, htmlContent string) {
	if r.redis == nil {
		return
	}
	_ = r.redis.Set(ctx, prerenderCachePrefix+cacheKey, htmlContent, prerenderCacheTTL).Err()
}

// InvalidateCache removes cached pre-rendered pages for a given bathhouse.
func (r *Renderer) InvalidateCache(ctx context.Context, bathhouseSlug string) {
	if r.redis == nil {
		return
	}
	// Invalidate detail page and reviews page
	keys := []string{
		prerenderCachePrefix + "detail:" + bathhouseSlug,
		prerenderCachePrefix + "reviews:" + bathhouseSlug,
	}
	_ = r.redis.Del(ctx, keys...).Err()
}

// InvalidateCityCache removes cached pre-rendered city listing page.
func (r *Renderer) InvalidateCityCache(ctx context.Context, citySlug string) {
	if r.redis == nil {
		return
	}
	_ = r.redis.Del(ctx, prerenderCachePrefix+"city:"+citySlug).Err()
}

// InvalidateListingCache removes cached pre-rendered listing page.
func (r *Renderer) InvalidateListingCache(ctx context.Context) {
	if r.redis == nil {
		return
	}
	_ = r.redis.Del(ctx, prerenderCachePrefix+"listing").Err()
}

// BuildBathhouseDetailPage builds PageData for a bathhouse detail page.
func (r *Renderer) BuildBathhouseDetailPage(input SchemaInput, meta MetaTags) PageData {
	schema := GenerateSchema(input)
	schemaBytes, _ := json.Marshal(schema)

	// Build body content with visible text for crawlers
	var body strings.Builder
	fmt.Fprintf(&body, "<h1>%s</h1>\n", html.EscapeString(input.Name))
	if input.Description != "" {
		fmt.Fprintf(&body, "<p>%s</p>\n", html.EscapeString(input.Description))
	}
	if input.CityName != "" {
		fmt.Fprintf(&body, "<p>Город: %s</p>\n", html.EscapeString(input.CityName))
	}
	if input.PricePerHour > 0 {
		fmt.Fprintf(&body, "<p>Цена от %d руб/ч</p>\n", input.PricePerHour/100)
	}
	if input.Rating > 0 && input.ReviewCount > 0 {
		fmt.Fprintf(&body, "<p>Рейтинг: %.1f (%d отзывов)</p>\n", input.Rating, input.ReviewCount)
	}

	// Amenities
	amenities := collectAmenities(MetaInput{
		HasPool:      input.HasPool,
		HasSauna:     input.HasSauna,
		HasSteamRoom: input.HasSteamRoom,
		HasHotTub:    input.HasHotTub,
		HasBBQ:       input.HasBBQ,
		HasKaraoke:   input.HasKaraoke,
	})
	if len(amenities) > 0 {
		fmt.Fprintf(&body, "<p>Удобства: %s</p>\n", html.EscapeString(strings.Join(amenities, ", ")))
	}

	return PageData{
		Title:       meta.Title,
		Description: meta.Description,
		Canonical:   meta.Canonical,
		OGImage:     meta.OGImage,
		OGType:      meta.OGType,
		OGSiteName:  meta.OGSiteName,
		OGLocale:    meta.OGLocale,
		TwitterCard: meta.TwitterCard,
		SchemaJSON:  string(schemaBytes),
		BodyContent: body.String(),
	}
}

// BuildCityListingPage builds PageData for a city catalog page.
func (r *Renderer) BuildCityListingPage(cityName, citySlug string, bathhouseCount int64) PageData {
	title := fmt.Sprintf("Бани и сауны в %s — %s", cityName, siteName)
	description := fmt.Sprintf("Найдено %d бань и саун в %s. Бронируйте онлайн на %s.", bathhouseCount, cityName, siteName)
	canonical := fmt.Sprintf("%s/%s", r.baseURL, citySlug)

	return PageData{
		Title:       title,
		Description: description,
		Canonical:   canonical,
		OGType:      "website",
		OGSiteName:  siteName,
		OGLocale:    "ru_RU",
		TwitterCard: "summary",
		BodyContent: fmt.Sprintf("<h1>%s</h1>\n<p>%s</p>\n", html.EscapeString(title), html.EscapeString(description)),
	}
}

// BuildMainListingPage builds PageData for the main catalog page.
func (r *Renderer) BuildMainListingPage(totalCount int64) PageData {
	title := fmt.Sprintf("Бани и сауны — %s", siteName)
	description := fmt.Sprintf("Каталог из %d бань и саун. Онлайн-бронирование, отзывы, цены. %s", totalCount, siteName)

	return PageData{
		Title:       title,
		Description: description,
		Canonical:   r.baseURL,
		OGType:      "website",
		OGSiteName:  siteName,
		OGLocale:    "ru_RU",
		TwitterCard: "summary",
		BodyContent: fmt.Sprintf("<h1>%s</h1>\n<p>%s</p>\n", html.EscapeString(title), html.EscapeString(description)),
	}
}

// BuildReviewsPage builds PageData for a bathhouse reviews page.
func (r *Renderer) BuildReviewsPage(bathhouseName, cityName, citySlug, bathhouseSlug string, rating float64, reviewCount int, reviewTexts []string) PageData {
	title := fmt.Sprintf("Отзывы о %s — %s", bathhouseName, siteName)
	description := fmt.Sprintf("Отзывы о %s", bathhouseName)
	if cityName != "" {
		description += fmt.Sprintf(" в %s", cityName)
	}
	if reviewCount > 0 {
		description += fmt.Sprintf(". Рейтинг %.1f, %d отзывов.", rating, reviewCount)
	}

	canonical := fmt.Sprintf("%s/bathhouses/%s", r.baseURL, bathhouseSlug)
	if citySlug != "" {
		canonical = fmt.Sprintf("%s/%s/%s/reviews", r.baseURL, citySlug, bathhouseSlug)
	}

	var body strings.Builder
	fmt.Fprintf(&body, "<h1>%s</h1>\n", html.EscapeString(title))
	if rating > 0 {
		fmt.Fprintf(&body, "<p>Средний рейтинг: %.1f из 5 (%d отзывов)</p>\n", rating, reviewCount)
	}
	for i, text := range reviewTexts {
		if i >= 10 { // limit to first 10 reviews in pre-rendered content
			break
		}
		fmt.Fprintf(&body, "<blockquote>%s</blockquote>\n", html.EscapeString(text))
	}

	return PageData{
		Title:       title,
		Description: description,
		Canonical:   canonical,
		OGType:      "website",
		OGSiteName:  siteName,
		OGLocale:    "ru_RU",
		TwitterCard: "summary",
		BodyContent: body.String(),
	}
}

// CacheKeyForDetail returns the cache key for a bathhouse detail page.
func CacheKeyForDetail(slug string) string {
	return "detail:" + slug
}

// CacheKeyForCity returns the cache key for a city listing page.
func CacheKeyForCity(slug string) string {
	return "city:" + slug
}

// CacheKeyForReviews returns the cache key for a bathhouse reviews page.
func CacheKeyForReviews(slug string) string {
	return "reviews:" + slug
}

// CacheKeyForListing returns the cache key for the main listing page.
func CacheKeyForListing() string {
	return "listing"
}
