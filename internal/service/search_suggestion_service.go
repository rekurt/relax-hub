package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/rekurt/relax-hub/internal/logger"
	"github.com/rekurt/relax-hub/internal/repository"
	"github.com/redis/go-redis/v9"
)

const (
	popularQueriesKey      = "search:popular_queries"
	suggestionCachePrefix  = "search:suggestions:"
	suggestionCacheTTL     = 5 * time.Minute
	minQueryLength         = 2
	maxSuggestionLimit     = 20
	defaultSuggestionLimit = 10
)

// SuggestionType identifies the source of a suggestion.
type SuggestionType string

const (
	SuggestionTypeBathhouse SuggestionType = "bathhouse"
	SuggestionTypeCity      SuggestionType = "city"
	SuggestionTypePopular   SuggestionType = "popular"
)

// Suggestion represents a single search suggestion.
type Suggestion struct {
	Text string         `json:"text"`
	Type SuggestionType `json:"type"`
}

// SearchSuggestionService provides search autocomplete suggestions.
type SearchSuggestionService interface {
	GetSuggestions(ctx context.Context, query string, limit int) ([]Suggestion, error)
	RecordQuery(ctx context.Context, query string) error
}

type searchSuggestionService struct {
	bathhouseRepo repository.BathhouseRepository
	cityRepo      repository.CityRepository
	redis         *redis.Client
	logger        *logger.Logger
}

// NewSearchSuggestionService creates a new SearchSuggestionService.
func NewSearchSuggestionService(
	bathhouseRepo repository.BathhouseRepository,
	cityRepo repository.CityRepository,
	redisClient *redis.Client,
	log *logger.Logger,
) SearchSuggestionService {
	return &searchSuggestionService{
		bathhouseRepo: bathhouseRepo,
		cityRepo:      cityRepo,
		redis:         redisClient,
		logger:        log,
	}
}

func (s *searchSuggestionService) GetSuggestions(ctx context.Context, query string, limit int) ([]Suggestion, error) {
	query = strings.TrimSpace(query)
	if len([]rune(query)) < minQueryLength {
		return []Suggestion{}, nil
	}

	if limit <= 0 || limit > maxSuggestionLimit {
		limit = defaultSuggestionLimit
	}

	// Try cache first
	cacheKey := suggestionCachePrefix + strings.ToLower(query)
	cached, err := s.redis.Get(ctx, cacheKey).Result()
	if err == nil && cached != "" {
		suggestions, parseErr := parseCachedSuggestions(cached)
		if parseErr == nil {
			if len(suggestions) > limit {
				suggestions = suggestions[:limit]
			}
			return suggestions, nil
		}
	}

	var suggestions []Suggestion

	// 1. Bathhouse names via trigram similarity
	bathSuggestions, err := s.getBathhouseNameSuggestions(ctx, query, limit)
	if err != nil {
		s.logger.Error("failed to get bathhouse suggestions", "error", err)
	} else {
		suggestions = append(suggestions, bathSuggestions...)
	}

	// 2. City names
	citySuggestions, err := s.getCityNameSuggestions(ctx, query)
	if err != nil {
		s.logger.Error("failed to get city suggestions", "error", err)
	} else {
		suggestions = append(suggestions, citySuggestions...)
	}

	// 3. Popular queries from Redis
	popularSuggestions, err := s.getPopularQuerySuggestions(ctx, query, limit)
	if err != nil {
		s.logger.Error("failed to get popular suggestions", "error", err)
	} else {
		suggestions = append(suggestions, popularSuggestions...)
	}

	// Deduplicate by text (case-insensitive)
	suggestions = deduplicateSuggestions(suggestions)

	// Trim to limit
	if len(suggestions) > limit {
		suggestions = suggestions[:limit]
	}

	// Cache result
	if cacheData, err := serializeSuggestions(suggestions); err == nil {
		_ = s.redis.Set(ctx, cacheKey, cacheData, suggestionCacheTTL).Err()
	}

	return suggestions, nil
}

func (s *searchSuggestionService) RecordQuery(ctx context.Context, query string) error {
	query = strings.TrimSpace(strings.ToLower(query))
	if len([]rune(query)) < minQueryLength {
		return nil
	}

	if err := s.redis.ZIncrBy(ctx, popularQueriesKey, 1, query).Err(); err != nil {
		return fmt.Errorf("record query: %w", err)
	}

	// Trim to top 10,000 entries to prevent unbounded growth
	_ = s.redis.ZRemRangeByRank(ctx, popularQueriesKey, 0, -10001).Err()

	return nil
}

func (s *searchSuggestionService) getBathhouseNameSuggestions(ctx context.Context, query string, limit int) ([]Suggestion, error) {
	q := strings.ToLower(query)

	// Use the bathhouse repository List with SearchQuery to find matching names
	searchQuery := q
	filter := repository.SuggestionFilter{
		Query: searchQuery,
		Limit: limit,
	}

	names, err := s.bathhouseRepo.SuggestNames(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("suggest bathhouse names: %w", err)
	}

	suggestions := make([]Suggestion, 0, len(names))
	for _, name := range names {
		suggestions = append(suggestions, Suggestion{
			Text: name,
			Type: SuggestionTypeBathhouse,
		})
	}

	return suggestions, nil
}

func (s *searchSuggestionService) getCityNameSuggestions(ctx context.Context, query string) ([]Suggestion, error) {
	cities, err := s.cityRepo.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("get cities: %w", err)
	}

	q := strings.ToLower(query)
	var suggestions []Suggestion
	for _, city := range cities {
		if strings.Contains(strings.ToLower(city.Name), q) {
			suggestions = append(suggestions, Suggestion{
				Text: city.Name,
				Type: SuggestionTypeCity,
			})
		}
	}

	return suggestions, nil
}

func (s *searchSuggestionService) getPopularQuerySuggestions(ctx context.Context, query string, limit int) ([]Suggestion, error) {
	// Get top popular queries from sorted set
	members, err := s.redis.ZRangeArgs(ctx, redis.ZRangeArgs{
		Key:     popularQueriesKey,
		Start:   "2", // at least 2 searches
		Stop:    "+inf",
		ByScore: true,
		Rev:     true,
		Count:   int64(limit * 3), // get more to filter
	}).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, fmt.Errorf("get popular queries: %w", err)
	}

	q := strings.ToLower(query)
	var suggestions []Suggestion
	for _, member := range members {
		if strings.Contains(member, q) && member != q {
			suggestions = append(suggestions, Suggestion{
				Text: member,
				Type: SuggestionTypePopular,
			})
			if len(suggestions) >= limit {
				break
			}
		}
	}

	return suggestions, nil
}

func deduplicateSuggestions(suggestions []Suggestion) []Suggestion {
	seen := make(map[string]bool)
	result := make([]Suggestion, 0, len(suggestions))
	for _, s := range suggestions {
		key := strings.ToLower(s.Text)
		if !seen[key] {
			seen[key] = true
			result = append(result, s)
		}
	}
	return result
}

func serializeSuggestions(suggestions []Suggestion) (string, error) {
	data, err := json.Marshal(suggestions)
	if err != nil {
		return "", fmt.Errorf("marshal suggestions: %w", err)
	}
	return string(data), nil
}

func parseCachedSuggestions(data string) ([]Suggestion, error) {
	if data == "" {
		return []Suggestion{}, nil
	}
	var suggestions []Suggestion
	if err := json.Unmarshal([]byte(data), &suggestions); err != nil {
		return nil, fmt.Errorf("unmarshal suggestions: %w", err)
	}
	return suggestions, nil
}
