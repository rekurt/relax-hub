---
# Subsystem 8: Pricing Enhancements (FR-082-090, FR-097)

## Overview
Extended pricing rules (long session discounts, extra guest surcharges), holiday pricing, last-minute discounts, and service fee configuration.

## Context
- Existing pricing: `internal/service/pricing_service.go` — dynamic pricing with rules (priority, multipliers per hourly slot)
- Existing pricing model: `internal/domain/pricing.go`
- Existing pricing repo: `internal/repository/postgres/pricing_repo.go`
- Price calculation happens during booking creation in `internal/service/booking_service.go`
- Prices in kopecks (int64)

## Dependencies
- Depends on: Add-ons (Subsystem 5) — service fee may or may not apply to add-ons
- Uses existing pricing infrastructure

## Development Approach
- **Testing approach**: Regular (code first, then tests)
- **CRITICAL: every task MUST include new/updated tests**
- **CRITICAL: all tests must pass before starting next task**

## Implementation Steps

### Task 8.1: Extended Pricing Rules (FR-084, FR-085)

**Files:**
- Modify: `internal/domain/bathhouse.go` or `internal/domain/pricing.go`
- Modify: `internal/service/pricing_service.go`
- Modify: `internal/service/booking_service.go`
- Create: `migrations/XXXXXX_pricing_enhancements.up.sql`

- [ ] Long session discount (FR-084):
  - Add to Bathhouse: LongSessionThresholdHours (int, default 4), LongSessionDiscountPercent (int, default 10, range 0-50)
  - When booking duration >= threshold, apply discount to hours beyond threshold
  - Example: 6h booking with 4h threshold and 10% discount: 4h at full price + 2h at 90%
- [ ] Extra guest surcharge (FR-085):
  - Add to Bathhouse: BaseCapacity (int), ExtraGuestSurcharge (int64, kopecks per extra guest per hour)
  - When guest_count > BaseCapacity, add surcharge for extra guests
  - Surcharge = (guest_count - base_capacity) * surcharge * duration_hours
- [ ] Migration: add `long_session_threshold_hours INT DEFAULT 4`, `long_session_discount_percent INT DEFAULT 0`, `base_capacity INT`, `extra_guest_surcharge BIGINT DEFAULT 0` to bathhouses
- [ ] Modify price calculation flow to include these rules (after base price, before discounts)
- [ ] Include breakdown in price response:
  ```json
  {
    "base_price": 1000000,
    "long_session_discount": -100000,
    "extra_guest_surcharge": 50000,
    "addons_total": 30000,
    "service_fee": 98000,
    "promo_discount": -50000,
    "total": 1028000
  }
  ```
- [ ] Write tests with various combinations
- [ ] Run `go test ./... -v` — must pass

### Task 8.2: Holiday Prices (FR-088)

**Files:**
- Create: `internal/domain/holiday.go`
- Create: `internal/repository/postgres/holiday_repo.go`
- Create: `internal/repository/mock/holiday_repo.go`
- Create: `internal/service/holiday_service.go`
- Create: `internal/handler/holiday_handler.go`
- Modify: `internal/repository/interfaces.go`
- Modify: `internal/service/pricing_service.go`
- Create: `migrations/XXXXXX_holidays.up.sql`

- [ ] Holiday model: ID, Name (string), Date (time.Time), Region (string: "RU"/"BY"), IsRecurring (bool — same date every year)
- [ ] BathhouseHolidayPrice model: BathhouseID, Multiplier (float64, range 1.0-2.0, step 0.1)
  - Per-bathhouse multiplier applied on holidays
  - If not set, use platform default (1.5x)
- [ ] HolidayRepository: Create, Update, Delete, ListByRegion, GetByDate, ListAll
- [ ] Admin endpoints:
  - GET /api/v1/admin/holidays — list all holidays
  - POST /api/v1/admin/holidays — create holiday (RequireRole: admin)
  - PUT /api/v1/admin/holidays/{id} — update holiday
  - DELETE /api/v1/admin/holidays/{id} — delete holiday
- [ ] Owner endpoint:
  - PUT /api/v1/my/bathhouses/{id}/holiday-multiplier — set holiday price multiplier
- [ ] Integrate into price calculation:
  - Check if booking date is a holiday for the region
  - Apply multiplier to base price
  - Add "is_holiday_price: true" and "holiday_name" to price response
- [ ] Seed initial holidays: New Year (1-8 Jan), 23 Feb, 8 Mar, 1 May, 9 May, 12 Jun, 4 Nov
- [ ] Write tests
- [ ] Run `go test ./... -v` — must pass

### Task 8.3: Last-Minute Discounts (FR-090)

**Files:**
- Modify: `internal/domain/bathhouse.go`
- Modify: `internal/service/pricing_service.go`
- Modify: `internal/repository/postgres/bathhouse_repo.go`
- Create: `migrations/XXXXXX_last_minute.up.sql`

- [ ] Add to Bathhouse:
  - LastMinuteEnabled (bool, default false)
  - LastMinuteDiscountPercent (int, default 20, range 5-50)
  - LastMinuteHoursThreshold (int, default 6, range 2-24)
- [ ] Migration: add these columns to bathhouses
- [ ] Logic: when an unbooked slot starts within LastMinuteHoursThreshold hours from now:
  - Apply LastMinuteDiscountPercent discount to base price
  - Mark in available slots response: `is_last_minute: true, original_price: X, discounted_price: Y`
- [ ] Add last_minute filter to search:
  - GET /api/v1/bathhouses?last_minute=true — show only bathhouses with active last-minute deals
- [ ] Include "last_minute" badge in bathhouse list response when applicable
- [ ] Write tests
- [ ] Run `go test ./... -v` — must pass

### Task 8.4: Service Fee (FR-097)

**Files:**
- Create: `internal/domain/service_fee.go`
- Create: `internal/service/service_fee_service.go`
- Modify: `internal/service/booking_service.go`
- Modify: `internal/domain/booking.go`
- Create: `migrations/XXXXXX_service_fee.up.sql`

- [ ] ServiceFeeConfig model:
  - ID, Region (string), Category (*string — optional bathhouse type), FeePercent (float64, default 10.0, range 0-25, step 0.5)
  - Priority: region+category > region > global default
- [ ] Migration: service_fee_configs table
- [ ] ServiceFeeService:
  - GetFeePercent(ctx, region, category) — resolve fee with priority rules
  - CalculateFee(ctx, basePrice, region, category) — return fee amount in kopecks
- [ ] Admin endpoint:
  - GET /api/v1/admin/service-fee — list all configs
  - PUT /api/v1/admin/service-fee — create/update config
- [ ] Modify booking creation:
  - Calculate service fee on base price (NOT on add-ons per BRD specification)
  - ServiceFeeAmount = round(basePrice * feePercent / 100)
  - Show as separate line in price breakdown
- [ ] Add ServiceFeeAmount (int64) to Booking model
- [ ] Migration: add `service_fee_amount BIGINT DEFAULT 0` to bookings
- [ ] Write tests
- [ ] Run `go test ./... -v -race` — must pass
- [ ] Run linter: `make lint`
