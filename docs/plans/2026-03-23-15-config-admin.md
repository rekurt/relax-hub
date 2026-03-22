---
# Subsystem 15: Configuration & Admin Enhancements

## Overview
Admin-configurable platform settings (key-value store), feature flags for gradual rollout, and force majeure handling for mass cancellations.

## Context
- Existing config: Viper with BANI_ env prefix
- Existing admin panel: GoAdmin-based at `internal/admin/`
- Existing admin routes in `internal/server/router.go`
- Many new features need admin-configurable parameters (service fee, bonus expiry, claim period, etc.)
- Currently all config is via env vars — no runtime admin changes possible

## Dependencies
- No dependencies on other new subsystems (but many subsystems read from platform settings)
- Should be implemented EARLY so other subsystems can use it

## Development Approach
- **Testing approach**: Regular (code first, then tests)
- **CRITICAL: every task MUST include new/updated tests**
- **CRITICAL: all tests must pass before starting next task**

## Implementation Steps

### Task 15.1: Platform Settings (Admin-configurable)

**Files:**
- Create: `internal/domain/platform_settings.go`
- Create: `internal/service/platform_settings_service.go`
- Create: `internal/handler/platform_settings_handler.go`
- Create: `internal/repository/postgres/platform_settings_repo.go`
- Create: `internal/repository/mock/platform_settings_repo.go`
- Modify: `internal/repository/interfaces.go`
- Modify: `internal/server/router.go`
- Create: `migrations/XXXXXX_platform_settings.up.sql`

- [ ] PlatformSetting model:
  - Key (string, primary key), Value (string — JSON encoded), Description (string)
  - Type (enum: int/float/string/bool/json), UpdatedAt, UpdatedBy (*uuid)
- [ ] Migration:
  ```sql
  CREATE TABLE platform_settings (
      key VARCHAR(100) PRIMARY KEY,
      value TEXT NOT NULL,
      description TEXT,
      type VARCHAR(10) NOT NULL DEFAULT 'string',
      updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
      updated_by UUID REFERENCES users(id)
  );
  ```
- [ ] Seed initial settings:
  - service_fee_percent: 10.0
  - welcome_bonus_amount: 50000 (kopecks)
  - welcome_bonus_expiry_days: 30
  - wallet_bonus_expiry_days: 180
  - wallet_refund_bonus_percent: 5
  - max_wallet_balance_rub: 10000000 (kopecks)
  - escrow_claim_hours: 48
  - min_payout_amount: 50000 (kopecks)
  - bayesian_min_reviews: 5
  - review_request_delay_hours: 2
  - noshow_grace_minutes: 30
  - offer_version: "v1.0"
- [ ] PlatformSettingsRepository: Get(key), GetAll(), Set(key, value, updatedBy)
- [ ] PlatformSettingsService:
  - GetString(ctx, key) (string, error) — with Redis cache (5 min TTL)
  - GetInt(ctx, key) (int64, error)
  - GetFloat(ctx, key) (float64, error)
  - GetBool(ctx, key) (bool, error)
  - Set(ctx, key, value, adminID) — update DB + invalidate Redis cache
  - GetAll(ctx) — return all settings
- [ ] Endpoints:
  - GET /api/v1/admin/settings — list all settings (RequireRole: admin)
  - PUT /api/v1/admin/settings/{key} — update setting value (RequireRole: admin)
- [ ] Integrate: other services should use PlatformSettingsService instead of hardcoded values
- [ ] Write tests
- [ ] Run `go test ./... -v` — must pass

### Task 15.2: Feature Flags

**Files:**
- Create: `internal/domain/feature_flag.go`
- Create: `internal/service/feature_flag_service.go`
- Create: `internal/handler/feature_flag_handler.go`
- Create: `internal/repository/postgres/feature_flag_repo.go`
- Modify: `internal/repository/interfaces.go`
- Create: `migrations/XXXXXX_feature_flags.up.sql`

- [ ] FeatureFlag model:
  - Key (string, primary key), Enabled (bool), Description (string)
  - Region (*string — optional, for region-specific flags)
  - UpdatedAt, UpdatedBy (*uuid)
- [ ] Migration + seed flags for all new subsystems:
  - wallet_enabled: false
  - phone_auth_enabled: false
  - two_fa_enabled: false
  - kyc_required: false
  - addons_enabled: false
  - request_booking_enabled: false
  - escrow_enabled: false
  - crm_enabled: false
  - disputes_enabled: false
  - antifraud_enabled: false
  - sbp_payments_enabled: false
  - last_minute_enabled: false
  - fulltext_search_enabled: false
- [ ] FeatureFlagRepository: Get, GetAll, Set
- [ ] FeatureFlagService:
  - IsEnabled(ctx, key) (bool) — with Redis cache (1 min TTL)
  - IsEnabledForRegion(ctx, key, region) (bool)
  - SetFlag(ctx, key, enabled, adminID)
  - GetAll(ctx) — return all flags
- [ ] Endpoints:
  - GET /api/v1/admin/feature-flags — list all (RequireRole: admin)
  - PUT /api/v1/admin/feature-flags/{key} — toggle (RequireRole: admin)
- [ ] Integrate: check feature flags before enabling new functionality
  - Example: in booking service, check "request_booking_enabled" before allowing request mode
- [ ] Write tests
- [ ] Run `go test ./... -v` — must pass

### Task 15.3: Force Majeure (FR-080)

**Files:**
- Create: `internal/service/force_majeure_service.go`
- Create: `internal/handler/force_majeure_handler.go`
- Modify: `internal/server/router.go`

- [ ] ForceMajeureService:
  - Activate(ctx, adminID, region, dateFrom, dateTo, reason)
    1. Find all confirmed bookings in region for date range
    2. Cancel all bookings with status "force_majeure_cancelled"
    3. 100% refund to all clients (to wallet, instant)
    4. Notify all affected clients: "Бронирование отменено по форс-мажору: {reason}"
    5. Notify all affected owners: "Все бронирования на {dates} отменены по форс-мажору"
    6. Log in audit log
  - GetActive(ctx) — return currently active force majeure events
- [ ] POST /api/v1/admin/force-majeure — activate (RequireRole: admin)
  - Request: { region, date_from, date_to, reason }
  - Response: { affected_bookings_count, total_refund_amount }
- [ ] GET /api/v1/admin/force-majeure — list active/past events
- [ ] Add "force_majeure_cancelled" to booking status enum
- [ ] Write tests
- [ ] Run `go test ./... -v -race` — must pass
- [ ] Run linter: `make lint`
- [ ] Verify build: `go build ./...`
- [ ] Regenerate swagger: `make swagger`
