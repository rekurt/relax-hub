---
# Subsystem 6: Search & Discovery Enhancements (FR-038-052)

## Overview
Full-text search with Russian language support, search suggestions, advanced multi-factor ranking, bathhouse comparison, recently viewed history, and saved searches with notifications.

## Context
- Existing search: `internal/repository/postgres/bathhouse_repo.go` — filter by city, type, price range, amenities, geo-radius
- BathhouseFilter: `internal/domain/filters.go`
- Existing favorites: `internal/service/favorite_service.go`
- Existing recommendations: `internal/service/recommendation_service.go`
- PostgreSQL with PostGIS already in use
- Redis available for caching

## Dependencies
- No dependencies on other new subsystems
- Uses existing bathhouse, subscription (for promotion boost), and notification subsystems

## Development Approach
- **Testing approach**: Regular (code first, then tests)
- **CRITICAL: every task MUST include new/updated tests**
- **CRITICAL: all tests must pass before starting next task**

## Implementation Steps

### Task 6.1: Full-text Search (FR-038)

**Files:**
- Modify: `internal/repository/postgres/bathhouse_repo.go`
- Modify: `internal/domain/filters.go`
- Create: `migrations/XXXXXX_fulltext_search.up.sql`

- [ ] Migration:
  ```sql
  CREATE EXTENSION IF NOT EXISTS pg_trgm;

  ALTER TABLE bathhouses ADD COLUMN search_vector tsvector;

  CREATE OR REPLACE FUNCTION bathhouses_search_vector_update() RETURNS trigger AS $$
  BEGIN
      NEW.search_vector :=
          setweight(to_tsvector('russian', coalesce(NEW.name, '')), 'A') ||
          setweight(to_tsvector('russian', coalesce(NEW.description, '')), 'B') ||
          setweight(to_tsvector('russian', coalesce(NEW.type, '')), 'C');
      RETURN NEW;
  END;
  $$ LANGUAGE plpgsql;

  CREATE TRIGGER bathhouses_search_vector_trigger
      BEFORE INSERT OR UPDATE ON bathhouses
      FOR EACH ROW EXECUTE FUNCTION bathhouses_search_vector_update();

  UPDATE bathhouses SET search_vector =
      setweight(to_tsvector('russian', coalesce(name, '')), 'A') ||
      setweight(to_tsvector('russian', coalesce(description, '')), 'B') ||
      setweight(to_tsvector('russian', coalesce(type, '')), 'C');

  CREATE INDEX idx_bathhouses_search ON bathhouses USING GIN(search_vector);
  CREATE INDEX idx_bathhouses_name_trgm ON bathhouses USING GIN(name gin_trgm_ops);
  ```
- [ ] Add Query (string) field to BathhouseFilter
- [ ] Modify bathhouse list query in repo:
  - When Query is set, use `plainto_tsquery('russian', query)` for matching
  - Use `ts_rank(search_vector, query)` for relevance scoring
  - Fallback to trigram similarity when ts_rank returns 0 (fuzzy matching)
  - Combine with existing filters (city, type, price, geo)
- [ ] Support partial matching via prefix search: `to_tsquery('russian', query || ':*')`
- [ ] Write tests
- [ ] Run `go test ./... -v` — must pass

### Task 6.2: Search Suggestions (FR-039)

**Files:**
- Create: `internal/service/search_suggestion_service.go`
- Create: `internal/handler/search_handler.go`
- Modify: `internal/server/router.go`

- [ ] SearchSuggestionService:
  - GetSuggestions(ctx, query, limit) — return matching suggestions
  - Sources:
    1. Bathhouse names matching prefix (trigram similarity > 0.3)
    2. Bathhouse type names matching prefix
    3. City names matching prefix
    4. Popular recent queries (stored in Redis sorted set by frequency)
  - RecordQuery(ctx, query) — increment query frequency in Redis
- [ ] GET /api/v1/search/suggestions?q=баня&limit=10 — return suggestions
  - Response: `{ suggestions: [{ text, type: "bathhouse"|"type"|"city"|"popular", id? }] }`
- [ ] Cache results in Redis with 5 min TTL for repeated queries
- [ ] Add swagger annotations
- [ ] Write tests
- [ ] Run `go test ./... -v` — must pass

### Task 6.3: Advanced Ranking (FR-041)

**Files:**
- Modify: `internal/repository/postgres/bathhouse_repo.go`
- Modify: `internal/domain/filters.go`
- Modify: `internal/domain/bathhouse.go`
- Create: `migrations/XXXXXX_ranking_fields.up.sql`

- [ ] Add computed fields to bathhouse (updated periodically by cron):
  - ConversionRate (float64) — bookings / views
  - OccupancyRate (float64) — booked hours / available hours
  - ViewCount (int) — total views
- [ ] Migration: add `conversion_rate DECIMAL(5,4) DEFAULT 0`, `occupancy_rate DECIMAL(5,4) DEFAULT 0`, `view_count INT DEFAULT 0` to bathhouses
- [ ] Implement composite ranking score in SQL:
  ```
  score = (relevance * 0.30) + (bayesian_rating * 0.25) + (conversion_rate * 0.20) + (occupancy_rate * 0.15) + (promotion_boost * 0.10)
  ```
  - relevance: ts_rank from full-text search (or 1.0 if no query)
  - bayesian_rating: from review system (or platform average if < 3 reviews)
  - promotion_boost: +10 for premium subscription, +20 for promoted (from existing subscription system)
- [ ] Add SortBy field to BathhouseFilter: "relevance" (default), "price_asc", "price_desc", "rating", "distance", "newest"
- [ ] Write tests
- [ ] Run `go test ./... -v` — must pass

### Task 6.4: Comparison (FR-046)

**Files:**
- Create: `internal/handler/comparison_handler.go`
- Modify: `internal/server/router.go`

- [ ] POST /api/v1/bathhouses/compare — accept 2-3 bathhouse IDs
  - Request: `{ ids: [uuid, uuid, uuid?] }`
  - Response: comparison table data:
    ```json
    {
      "bathhouses": [
        {
          "id": "...",
          "name": "...",
          "price_per_hour": 500000,
          "rating": 4.5,
          "review_count": 42,
          "capacity": 10,
          "amenities": ["sauna", "pool", ...],
          "distance_km": 5.2,
          "cancellation_policy": "flexible",
          "type": "russian",
          "photos": ["url1", "url2"]
        }
      ]
    }
    ```
- [ ] Validate: 2-3 IDs required, all must exist and be active
- [ ] Optionally accept user location for distance calculation
- [ ] Write tests
- [ ] Run `go test ./... -v` — must pass

### Task 6.5: Recently Viewed & Saved Searches (FR-051, FR-052)

**Files:**
- Create: `internal/domain/saved_search.go`
- Create: `internal/service/saved_search_service.go`
- Create: `internal/handler/saved_search_handler.go`
- Create: `internal/repository/postgres/saved_search_repo.go`
- Modify: `internal/repository/interfaces.go`
- Modify: `internal/handler/bathhouse_handler.go` (record views)
- Create: `migrations/XXXXXX_saved_searches.up.sql`

- [ ] Recently viewed: store in Redis sorted set per user
  - Key: `recently_viewed:{userID}`, member: bathhouseID, score: unix timestamp
  - Keep last 20 entries (ZREMRANGEBYRANK to trim)
  - Record view on GET /bathhouses/{id} detail endpoint
- [ ] GET /api/v1/my/recently-viewed — return list of recently viewed bathhouses (RequireAuth)
- [ ] SavedSearch model: ID, UserID, Name (string), Filters (JSONB — serialized BathhouseFilter), NotifyOnNew (bool), CreatedAt
- [ ] SavedSearchRepository: Create, ListByUser, Delete, ListWithNotifications
- [ ] SavedSearchService:
  - Save(ctx, userID, name, filters, notifyOnNew)
  - List(ctx, userID) — return saved searches
  - Delete(ctx, userID, searchID)
  - CheckNewMatches(ctx) — cron: for each saved search with NotifyOnNew=true, check new bathhouses since last check, send notification
- [ ] POST /api/v1/my/saved-searches — save search params
- [ ] GET /api/v1/my/saved-searches — list saved searches
- [ ] DELETE /api/v1/my/saved-searches/{id} — delete
- [ ] Cron: daily check for new bathhouses matching saved searches
- [ ] Write tests
- [ ] Run `go test ./... -v -race` — must pass
- [ ] Run linter: `make lint`
