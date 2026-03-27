---
# Subsystem 15: Configuration & Admin Enhancements

## Overview
Admin-configurable platform settings (key-value store with Redis cache), feature flags for gradual rollout (with region support via cities.region), and force majeure handling for mass booking cancellations by region.

## Context
- Files involved: internal/domain/, internal/service/, internal/handler/, internal/repository/postgres/, internal/repository/interfaces.go, internal/server/router.go, migrations/
- Related patterns: ServiceFeeService (simple CRUD + admin handler), SearchSuggestionService (Redis caching), CityService (simple entity)
- Dependencies: redis/go-redis v9, pgxpool, existing domain/booking/wallet/notification infrastructure
- Latest migration: 000077. Next: 000078
- fx DI via module.go files in each layer
- Region: new field in cities table (user chose this approach)

## Development Approach
- **Testing approach**: Regular (code first, then tests)
- Follow existing patterns: handler -> service -> repository with fx DI
- Redis caching: best-effort (ignore cache errors), JSON for complex values, plain strings for scalars
- **CRITICAL: every task MUST include new/updated tests**
- **CRITICAL: all tests must pass before starting next task**

## Implementation Steps

### Task 1: Add region field to cities

**Files:**
- Modify: `internal/domain/city.go`
- Modify: `internal/repository/postgres/city_repo.go`
- Modify: `internal/repository/mock/mock_repos.go` (city mock)
- Modify: `internal/handler/city_handler.go`
- Modify: `internal/service/city_service.go`
- Create: `migrations/000078_city_region.up.sql`
- Create: `migrations/000078_city_region.down.sql`

- [x] Add `Region string` field to `domain.City`
- [x] Migration: `ALTER TABLE cities ADD COLUMN region VARCHAR(100) NOT NULL DEFAULT ''`
- [x] Update city repo CRUD to include region in queries
- [x] Update city handler create/update to accept region
- [x] Update city service CreateCityInput/UpdateCityInput
- [x] Update mock repo if needed
- [x] Write tests for city service with region field
- [x] Run `go test ./... -v` - must pass

### Task 2: Platform Settings (key-value store with Redis cache)

**Files:**
- Create: `internal/domain/platform_settings.go`
- Create: `internal/repository/postgres/platform_settings_repo.go`
- Create: `internal/repository/mock/platform_settings_repo.go`
- Create: `internal/service/platform_settings_service.go`
- Create: `internal/handler/platform_settings_handler.go`
- Modify: `internal/repository/interfaces.go`
- Modify: `internal/repository/postgres/module.go`
- Modify: `internal/service/module.go`
- Modify: `internal/handler/module.go`
- Modify: `internal/server/router.go`
- Create: `migrations/000079_platform_settings.up.sql`
- Create: `migrations/000079_platform_settings.down.sql`

- [x] PlatformSetting domain model: Key (string PK), Value (string/JSON), Description (string), Type (enum: int/float/string/bool/json), UpdatedAt, UpdatedBy (*uuid)
- [x] Migration: create platform_settings table + seed 12 initial settings (service_fee_percent, welcome_bonus_amount, welcome_bonus_expiry_days, wallet_bonus_expiry_days, wallet_refund_bonus_percent, max_wallet_balance, escrow_claim_hours, min_payout_amount, bayesian_min_reviews, review_request_delay_hours, noshow_grace_minutes, offer_version)
- [x] PlatformSettingsRepository interface: Get(key), GetAll(), Set(key, value, updatedBy)
- [x] Postgres repo implementation following city_repo.go pattern
- [x] Mock repo for testing
- [x] PlatformSettingsService with typed getters (GetString, GetInt, GetFloat, GetBool) + Redis cache (5 min TTL, key prefix "platform:settings:")
- [x] Set method: update DB + invalidate Redis cache
- [x] GetAll method: return all settings (no cache, admin-only)
- [x] Admin handler: GET /api/v1/admin/settings, PUT /api/v1/admin/settings/{key}
- [x] Register in fx modules (repo, service, handler) and router
- [x] Swagger annotations on endpoints
- [x] Write tests for service (mock repo, typed getters, cache invalidation)
- [x] Run `go test ./... -v` - must pass

### Task 3: Feature Flags (with region support)

**Files:**
- Create: `internal/domain/feature_flag.go`
- Create: `internal/repository/postgres/feature_flag_repo.go`
- Create: `internal/repository/mock/feature_flag_repo.go`
- Create: `internal/service/feature_flag_service.go`
- Create: `internal/handler/feature_flag_handler.go`
- Modify: `internal/repository/interfaces.go`
- Modify: `internal/repository/postgres/module.go`
- Modify: `internal/service/module.go`
- Modify: `internal/handler/module.go`
- Modify: `internal/server/router.go`
- Create: `migrations/000080_feature_flags.up.sql`
- Create: `migrations/000080_feature_flags.down.sql`

- [x] FeatureFlag domain model: Key (string PK), Enabled (bool), Description (string), Region (*string - optional), UpdatedAt, UpdatedBy (*uuid)
- [x] Migration: create feature_flags table + seed 13 flags (wallet_enabled, phone_auth_enabled, two_fa_enabled, kyc_required, addons_enabled, request_booking_enabled, escrow_enabled, crm_enabled, disputes_enabled, antifraud_enabled, sbp_payments_enabled, last_minute_enabled, fulltext_search_enabled) - all false by default
- [x] FeatureFlagRepository interface: Get(key), GetAll(), Set(key, enabled, region, updatedBy)
- [x] Postgres repo implementation
- [x] Mock repo for testing
- [x] FeatureFlagService: IsEnabled(ctx, key) bool with Redis cache (1 min TTL), IsEnabledForRegion(ctx, key, region) bool, SetFlag(ctx, key, enabled, region, adminID), GetAll(ctx)
- [x] Admin handler: GET /api/v1/admin/feature-flags, PUT /api/v1/admin/feature-flags/{key}
- [x] Register in fx modules and router
- [x] Swagger annotations
- [x] Write tests for service (mock repo, region matching, cache behavior)
- [x] Run `go test ./... -v` - must pass

### Task 4: Force Majeure

**Files:**
- Create: `internal/domain/force_majeure.go`
- Create: `internal/service/force_majeure_service.go`
- Create: `internal/handler/force_majeure_handler.go`
- Modify: `internal/domain/booking.go` (add BookingForceMajeure status)
- Modify: `internal/repository/interfaces.go` (add ListConfirmedByRegionAndDateRange to BookingRepository)
- Modify: `internal/repository/postgres/booking_repo.go`
- Modify: `internal/repository/mock/mock_repos.go`
- Modify: `internal/service/module.go`
- Modify: `internal/handler/module.go`
- Modify: `internal/server/router.go`
- Create: `migrations/000081_force_majeure.up.sql`
- Create: `migrations/000081_force_majeure.down.sql`

- [ ] Add `BookingForceMajeure BookingStatus = "force_majeure_cancelled"` to domain/booking.go, update IsValid()
- [ ] ForceMajeureEvent domain model: ID (uuid), AdminID (uuid), Region (string), DateFrom/DateTo (time.Time), Reason (string), AffectedCount (int), TotalRefund (int64), CreatedAt
- [ ] Migration: create force_majeure_events table, add 'force_majeure_cancelled' to any CHECK constraints if they exist
- [ ] Add BookingRepository.ListConfirmedByRegionAndDateRange(ctx, region, dateFrom, dateTo) - joins bookings -> bathhouses -> cities to filter by cities.region
- [ ] ForceMajeureService.Activate(ctx, adminID, region, dateFrom, dateTo, reason):
  - Find confirmed bookings in region for date range
  - Cancel each with force_majeure_cancelled status
  - 100% refund to wallet for each client
  - Notify affected clients and owners
  - Create audit log entry
  - Save ForceMajeureEvent record
  - Return affected count and total refund amount
- [ ] ForceMajeureService.List(ctx) - return all events ordered by created_at desc
- [ ] POST /api/v1/admin/force-majeure - activate (RequireRole: admin)
- [ ] GET /api/v1/admin/force-majeure - list events (RequireRole: admin)
- [ ] Register in fx modules and router
- [ ] Swagger annotations
- [ ] Write tests for service (mock repos, verify cancellation + refund + notification logic)
- [ ] Run `go test ./... -v` - must pass

### Task 5: Verify acceptance criteria

- [ ] Run full test suite: `go test ./... -v -race`
- [ ] Run linter: `make lint`
- [ ] Verify build: `go build ./...`
- [ ] Regenerate swagger: `make swagger`
- [ ] Manual verification: all admin endpoints return correct JSON format { success, data, error, meta }

### Task 6: Update documentation

- [ ] Update CLAUDE.md: add platform settings, feature flags, force majeure to feature subsystems section
- [ ] Move this plan to `docs/plans/completed/`
