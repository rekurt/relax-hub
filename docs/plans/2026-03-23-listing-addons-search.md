# Execution Plan: Subsystems 4, 5, 6 — Listing Management, Add-ons, Search & Discovery

## Overview

Sequential execution of three independent subsystems: listing management enhancements (completeness check, audit log, duplication, deactivation/archival), configurable add-ons with booking integration, and search/discovery improvements (full-text search, suggestions, ranking, comparison, saved searches).

## Context

- Files involved: 30+ files across domain, repository, service, handler, and migrations
- Related patterns: existing handler->service->repository clean architecture, fx DI modules
- Dependencies: no cross-subsystem dependencies; subsystem 5 (addons) will later integrate with subsystem 7 (booking enhancements)
- BathhouseFilter already has SearchQuery (*string) and SortBy (string) fields — extend, not duplicate
- Bathhouse domain already has "inactive" status — extend with "archived"
- Next migration: 000049

## Development Approach

- **Testing approach**: Regular (code first, then tests)
- Complete each task fully before moving to the next
- Follow existing patterns: domain model -> repository interface -> postgres impl -> mock impl -> service -> handler -> router registration -> fx module registration
- **CRITICAL: every task MUST include new/updated tests**
- **CRITICAL: all tests must pass before starting next task**

## Implementation Steps

### Task 1: Listing Completeness Check (FR-025)

**Files:**
- Modify: `internal/service/bathhouse_service.go`
- Modify: `internal/handler/bathhouse_handler.go`
- Modify: `internal/server/router.go`

- [x] Add CompletenessItem, CompletenessResult types to domain or service layer
- [x] Add CheckCompleteness method to BathhouseService — checks required fields (name, description 50+ chars, type, geocoded address, city, 3+ verified photos, price > 0, schedule, capacity) and optional fields (3+ amenities, contact phone, cancellation policy)
- [x] Add GET /api/v1/my/bathhouses/{id}/completeness endpoint (RequireAuth, owner/rep)
- [x] Block moderation submission if required items incomplete
- [x] Write tests for completeness logic
- [x] Run `go test ./... -v` — must pass

### Task 2: Audit Log for Edits (FR-029)

**Files:**
- Create: `internal/domain/audit_log.go`
- Modify: `internal/repository/interfaces.go`
- Create: `internal/repository/postgres/audit_log_repo.go`
- Create: `internal/repository/mock/audit_log_repo.go`
- Create: `internal/service/audit_log_service.go`
- Modify: `internal/service/bathhouse_service.go`
- Modify: `internal/handler/bathhouse_handler.go`
- Modify: `internal/server/router.go`
- Modify: `internal/service/module.go`, `internal/handler/module.go`, `internal/repository/postgres/module.go`
- Create: `migrations/000049_audit_log.up.sql`, `migrations/000049_audit_log.down.sql`

- [x] AuditLog model: ID, EntityType, EntityID, UserID, Action, ChangedFields (JSONB), CreatedAt
- [x] AuditLogRepository interface: Create, ListByEntity, ListByUser
- [x] Postgres and mock implementations
- [x] AuditLogService: LogChange, GetHistory, IsSubstantialChange (address/type change or >50% photos replaced)
- [x] Hook into bathhouse Update: compare old vs new, log diff, auto-set status to pending on substantial change
- [x] GET /api/v1/admin/audit-log — admin view with filters
- [x] GET /api/v1/my/bathhouses/{id}/history — owner view
- [x] Register in fx modules
- [x] Write tests
- [x] Run `go test ./... -v` — must pass

### Task 3: Listing Duplication (FR-032)

**Files:**
- Modify: `internal/service/bathhouse_service.go`
- Modify: `internal/handler/bathhouse_handler.go`
- Modify: `internal/server/router.go`

- [x] POST /api/v1/my/bathhouses/{id}/duplicate (RequireAuth, owner/rep)
- [x] DuplicateBathhouse in service: copy all fields, name + " (копия)", new slug, status=draft, same photo URLs, new UUID
- [x] Write tests
- [x] Run `go test ./... -v` — must pass

### Task 4: Temporary Deactivation & Full Archival (FR-030, FR-031)

**Files:**
- Modify: `internal/domain/bathhouse.go`
- Modify: `internal/service/bathhouse_service.go`
- Modify: `internal/handler/bathhouse_handler.go`
- Modify: `internal/server/router.go`
- Modify: `internal/repository/postgres/bathhouse_repo.go`

- [x] Add "archived" status constant (inactive already exists)
- [x] POST /api/v1/my/bathhouses/{id}/deactivate — set status to inactive, keep bookings
- [x] POST /api/v1/my/bathhouses/{id}/activate — restore to active (only from inactive)
- [x] DELETE /api/v1/my/bathhouses/{id}/archive — archive if no active bookings, not reversible via API
- [x] Modify search queries to exclude inactive and archived listings
- [x] Write tests
- [x] Run `go test ./... -v -race` — must pass

### Task 5: Add-ons Domain & Repository (FR-024)

**Files:**
- Create: `internal/domain/addon.go`
- Modify: `internal/repository/interfaces.go`
- Create: `internal/repository/postgres/addon_repo.go`
- Create: `internal/repository/mock/addon_repo.go`
- Create: `migrations/000050_addons.up.sql`, `migrations/000050_addons.down.sql`

- [x] AddOn model: ID, BathhouseID, Name, Description, Price (kopecks), Unit (per_item/per_hour/per_person), IsActive, SortOrder, CreatedAt, UpdatedAt
- [x] BookingAddOn model: ID, BookingID, AddOnID, Name (denormalized), Quantity, UnitPrice, TotalPrice
- [x] Migration: addons table + booking_addons table with indexes
- [x] AddOnRepository interface: Create, Update, Delete, GetByID, ListByBathhouse, CountByBathhouse, CreateBookingAddOn, ListByBooking
- [x] Postgres and mock implementations
- [x] Register in fx modules
- [x] Write tests
- [x] Run `go test ./... -v` — must pass

### Task 6: Add-ons Service & Handler

**Files:**
- Create: `internal/service/addon_service.go`
- Create: `internal/handler/addon_handler.go`
- Modify: `internal/server/router.go`
- Modify: `internal/service/module.go`, `internal/handler/module.go`

- [x] AddOnService: CreateAddOn (check ownership + max 20 limit), UpdateAddOn, DeleteAddOn (soft), ListAddOns, CalculateAddOnTotal (per_item/per_hour/per_person pricing)
- [x] Owner endpoints: POST/PUT/DELETE /api/v1/my/bathhouses/{id}/addons, /api/v1/my/addons/{id}
- [x] Public endpoint: GET /api/v1/bathhouses/{id}/addons
- [x] Add swagger annotations
- [x] Register in fx modules
- [x] Write tests
- [x] Run `go test ./... -v` — must pass

### Task 7: Add-ons Booking Integration

**Files:**
- Modify: `internal/service/booking_service.go`
- Modify: `internal/handler/booking_handler.go`
- Modify: `internal/domain/booking.go`

- [x] Add AddOnTotal (int64) field to Booking model
- [x] Extend CreateBookingRequest with AddOns []AddOnSelection (addon_id + quantity)
- [x] In booking creation: validate add-on IDs belong to bathhouse, validate active, calculate totals by unit type, store booking_addons, update total price
- [x] Include add-ons in booking detail response
- [x] Write tests for booking with add-ons
- [x] Run `go test ./... -v -race` — must pass

### Task 8: Full-text Search (FR-038)

**Files:**
- Modify: `internal/repository/postgres/bathhouse_repo.go`
- Modify: `internal/domain/filters.go` (extend existing SearchQuery handling)
- Create: `migrations/000051_fulltext_search.up.sql`, `migrations/000051_fulltext_search.down.sql`

- [x] Migration: pg_trgm extension, search_vector tsvector column, trigger for auto-update, GIN indexes
- [x] Modify bathhouse list query: when SearchQuery is set, use plainto_tsquery('russian', query) + ts_rank for relevance, fallback to trigram similarity for fuzzy matching
- [x] Support prefix search with to_tsquery('russian', query || ':*')
- [x] Write tests
- [x] Run `go test ./... -v` — must pass

### Task 9: Search Suggestions (FR-039)

**Files:**
- Create: `internal/service/search_suggestion_service.go`
- Create: `internal/handler/search_handler.go`
- Modify: `internal/server/router.go`
- Modify: `internal/service/module.go`, `internal/handler/module.go`

- [x] SearchSuggestionService: GetSuggestions (bathhouse names via trigram, type names, city names, popular queries from Redis sorted set), RecordQuery (increment frequency in Redis)
- [x] GET /api/v1/search/suggestions?q=...&limit=10 — return typed suggestions
- [x] Cache in Redis with 5 min TTL
- [x] Register in fx modules
- [x] Write tests
- [x] Run `go test ./... -v` — must pass

### Task 10: Advanced Ranking (FR-041)

**Files:**
- Modify: `internal/repository/postgres/bathhouse_repo.go`
- Modify: `internal/domain/bathhouse.go`
- Create: `migrations/000052_ranking_fields.up.sql`, `migrations/000052_ranking_fields.down.sql`

- [x] Migration: add conversion_rate, occupancy_rate, view_count columns to bathhouses
- [x] Composite ranking score in SQL: relevance*0.30 + bayesian_rating*0.25 + conversion_rate*0.20 + occupancy_rate*0.15 + promotion_boost*0.10
- [x] Extend SortBy options: "relevance" (default), "price_asc", "price_desc", "rating", "distance", "newest"
- [x] Write tests
- [x] Run `go test ./... -v` — must pass

### Task 11: Comparison (FR-046)

**Files:**
- Create: `internal/handler/comparison_handler.go`
- Modify: `internal/server/router.go`
- Modify: `internal/handler/module.go`

- [ ] POST /api/v1/bathhouses/compare — accept 2-3 bathhouse IDs
- [ ] Return comparison table: name, price, rating, review_count, capacity, amenities, distance, cancellation_policy, type, photos
- [ ] Validate 2-3 IDs, all must exist and be active
- [ ] Optional user location for distance calculation
- [ ] Register in fx modules
- [ ] Write tests
- [ ] Run `go test ./... -v` — must pass

### Task 12: Recently Viewed & Saved Searches (FR-051, FR-052)

**Files:**
- Create: `internal/domain/saved_search.go`
- Modify: `internal/repository/interfaces.go`
- Create: `internal/repository/postgres/saved_search_repo.go`
- Create: `internal/repository/mock/saved_search_repo.go`
- Create: `internal/service/saved_search_service.go`
- Create: `internal/handler/saved_search_handler.go`
- Modify: `internal/handler/bathhouse_handler.go` (record views)
- Modify: `internal/server/router.go`
- Modify: `internal/service/module.go`, `internal/handler/module.go`, `internal/repository/postgres/module.go`
- Create: `migrations/000053_saved_searches.up.sql`, `migrations/000053_saved_searches.down.sql`

- [ ] Recently viewed: Redis sorted set per user (recently_viewed:{userID}), record on GET /bathhouses/{id}, keep last 20
- [ ] GET /api/v1/my/recently-viewed — list recently viewed bathhouses
- [ ] SavedSearch model: ID, UserID, Name, Filters (JSONB), NotifyOnNew, CreatedAt
- [ ] SavedSearchRepository: Create, ListByUser, Delete, ListWithNotifications
- [ ] SavedSearchService: Save, List, Delete, CheckNewMatches (cron)
- [ ] POST/GET/DELETE /api/v1/my/saved-searches
- [ ] Cron: daily check for new bathhouses matching saved searches
- [ ] Register in fx modules
- [ ] Write tests
- [ ] Run `go test ./... -v -race` — must pass

### Task 13: Verify acceptance criteria

- [ ] Run full test suite: `go test ./... -v -race`
- [ ] Run linter: `make lint`
- [ ] Manual verification:
  - Listing completeness endpoint returns correct checklist
  - Audit log records changes and triggers re-moderation
  - Listing duplication creates proper copy
  - Deactivation/archival works with booking checks
  - Add-ons CRUD and booking integration work end-to-end
  - Full-text search returns relevant results with Russian language
  - Search suggestions return typed results
  - Ranking sorts by composite score
  - Comparison returns structured data for 2-3 bathhouses
  - Recently viewed and saved searches persist correctly

### Task 14: Update documentation

- [ ] Update CLAUDE.md if internal patterns changed (new subsystems, new domain errors, new config vars)
- [ ] Run `make swagger` to regenerate OpenAPI spec
- [ ] Move completed plans to `docs/plans/completed/`
