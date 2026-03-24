# Unified Plan: Booking, Pricing & Payment Enhancements (Subsystems 7-9)

## Overview

Combined implementation of booking enhancements (request-based mode, check-in/check-out, reminders, session extension, re-booking, buffer/lead times, owner penalties),
pricing enhancements (long session discounts, extra guest surcharges, holiday pricing, last-minute discounts, service fee), and payment/financial enhancements (SBP,
combo payments, payment holds, escrow, enhanced refunds, fiscalization).

## Context

- Booking service: `internal/service/booking_service.go` — BookingService interface with Create, Cancel, Confirm, Reject, Complete, ListByUser, ListByBathhouse, GetAvailableSlots
- Booking model: `internal/domain/booking.go` — Booking struct with ID, UserID, BathhouseID, StartTime, EndTime, GuestCount, TotalPrice, AddOnTotal, PointsSpent, ReferralBonusUsed, Status, Comment. Statuses: pending, confirmed, cancelled, rejected, completed
- Booking handler: `internal/handler/booking_handler.go` — BookingHandler with bookingService and paymentService deps, createBookingRequest struct, bookingResponse struct
- Bathhouse model: `internal/domain/bathhouse.go` — Bathhouse struct with PricePerHour (int64 kopecks), MinDuration (int), MaxGuests (int), WorkingHours, Rating, Status
- Pricing service: `internal/service/pricing_service.go` — PricingService interface with CalculatePrice (splits into hourly slots, applies multiplier from highest-priority rule), CreateRule, UpdateRule, DeleteRule, ListRules, GetActiveRules
- Pricing model: `internal/domain/pricing.go` — PricingRule with Type (weekday/weekend/holiday/time_range/season), Multiplier, Priority, DaysOfWeek, DateFrom/To, TimeFrom/To
- Payment service: `internal/service/payment_service.go` — PaymentService interface with InitiatePayment, HandleWebhook, RefundPayment, GetPaymentByBooking, ListUserPayments. Uses PaymentProvider interface. Refund tiers: 100% if >24h, 50% if 2-24h, 0% if <2h
- Payment provider: `internal/payment/provider.go` — PaymentProvider interface: CreatePayment, GetPaymentStatus, CreateRefund
- YooKassa: `internal/payment/yookassa.go` — YooKassaProvider implements PaymentProvider. Currently always sets Capture=true, confirmation type Redirect
- Payment model: `internal/domain/payment.go` — Payment struct with Amount, Currency, Status, Provider, ExternalID, RefundAmount, RefundedAt, Metadata
- Wallet: WalletRepository in interfaces.go has CreateHold, GetHoldByID, UpdateHoldStatus, GetActiveHolds, GetExpiredHolds — wallet holds are already implemented
- Repository interfaces: `internal/repository/interfaces.go` — all repo interfaces in single file
- Notification system: `internal/notification/` — multi-channel dispatcher
- Booking repo: BookingRepository has Create, GetByID, ListByUser, ListByBathhouse, UpdateStatus, CheckAvailability, GetOverlapping, CountActiveByBathhouse, GetUserStats

## Dependencies

- Requires Wallet System (Subsystem 1) to be implemented — WalletService.Hold/CaptureHold/ReleaseHold/Spend/Refund
- Requires Add-ons (Subsystem 5) to be implemented — add-on calculations in booking
- Tasks ordered to respect internal dependencies: pricing before booking payment changes, payment holds after request-based bookings

## Development Approach

- **Testing approach**: Regular (code first, then tests)
- Complete each task fully before moving to the next
- **CRITICAL: every task MUST include new/updated tests**
- **CRITICAL: all tests must pass before starting next task**

## Implementation Steps

### Task 1: Service Fee System (FR-097)

**Files:**
- Create: `internal/domain/service_fee.go`
- Create: `internal/service/service_fee_service.go`
- Create: `internal/repository/postgres/service_fee_repo.go`
- Create: `internal/repository/mock/service_fee_repo.go`
- Create: `internal/handler/service_fee_handler.go`
- Modify: `internal/repository/interfaces.go`
- Modify: `internal/service/booking_service.go`
- Modify: `internal/domain/booking.go`
- Create: `migrations/XXXXXX_service_fee.up.sql` / `.down.sql`

**Domain model (`internal/domain/service_fee.go`):**
- [x] Define `ServiceFeeConfig` struct:
  ```go
  type ServiceFeeConfig struct {
      ID         uuid.UUID
      Region     string     // e.g. "RU", "BY" or "*" for global default
      Category   *string    // optional bathhouse category filter
      FeePercent float64    // range 0-25, step 0.5, default 10.0
      CreatedAt  time.Time
      UpdatedAt  time.Time
  }
  ```
- [x] Validate method: FeePercent >= 0 && FeePercent <= 25, Region non-empty

**Migration (`migrations/XXXXXX_service_fee.up.sql`):**
- [x] Create `service_fee_configs` table:
  ```sql
  CREATE TABLE service_fee_configs (
      id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
      region VARCHAR(10) NOT NULL DEFAULT '*',
      category VARCHAR(100),
      fee_percent NUMERIC(5,2) NOT NULL DEFAULT 10.0 CHECK (fee_percent >= 0 AND fee_percent <= 25),
      created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
      updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
      UNIQUE (region, category)
  );
  INSERT INTO service_fee_configs (region, fee_percent) VALUES ('*', 10.0);
  ```
- [x] Add `service_fee_amount BIGINT NOT NULL DEFAULT 0` column to `bookings` table

**Repository (`internal/repository/interfaces.go`):**
- [x] Add `ServiceFeeRepository` interface:
  ```go
  type ServiceFeeRepository interface {
      GetByRegionAndCategory(ctx context.Context, region string, category *string) (*domain.ServiceFeeConfig, error)
      GetByRegion(ctx context.Context, region string) (*domain.ServiceFeeConfig, error)
      GetGlobalDefault(ctx context.Context) (*domain.ServiceFeeConfig, error)
      List(ctx context.Context) ([]domain.ServiceFeeConfig, error)
      Upsert(ctx context.Context, config *domain.ServiceFeeConfig) error
  }
  ```

**Postgres implementation (`internal/repository/postgres/service_fee_repo.go`):**
- [x] Implement all ServiceFeeRepository methods with pgx
- [x] `GetByRegionAndCategory`: SELECT WHERE region=$1 AND category=$2
- [x] `GetByRegion`: SELECT WHERE region=$1 AND category IS NULL
- [x] `GetGlobalDefault`: SELECT WHERE region='*' AND category IS NULL

**Mock implementation (`internal/repository/mock/service_fee_repo.go`):**
- [x] In-memory map implementation for testing

**Service (`internal/service/service_fee_service.go`):**
- [x] Define `ServiceFeeService` interface:
  ```go
  type ServiceFeeService interface {
      GetFeePercent(ctx context.Context, region string, category *string) (float64, error)
      CalculateFee(ctx context.Context, basePrice int64, region string, category *string) (int64, error)
  }
  ```
- [x] `GetFeePercent` priority resolution: region+category > region > global default ("*")
- [x] `CalculateFee`: `fee = int64(math.Round(float64(basePrice) * feePercent / 100))`

**Modify Booking model (`internal/domain/booking.go`):**
- [x] Add `ServiceFeeAmount int64` field to Booking struct

**Modify booking creation (`internal/service/booking_service.go`):**
- [x] Inject `serviceFeeSvc ServiceFeeService` into `bookingService` struct and `NewBookingService` constructor
- [x] After `CalculatePrice` (line ~167) and before add-on calculation, call `serviceFeeSvc.CalculateFee(ctx, totalPrice, bh.Region, bh.Category)` — note: service fee is on BASE price only, NOT add-ons per BRD
- [x] Store result in `booking.ServiceFeeAmount`
- [x] Add `ServiceFeeAmount` to `totalPrice` after calculation

**Handler (`internal/handler/service_fee_handler.go`):**
- [x] Admin endpoints (RequireRole: admin):
  - `GET /api/v1/admin/service-fee` — list all configs
  - `PUT /api/v1/admin/service-fee` — create/update config (upsert by region+category)
- [x] Add swagger annotations

**Modify BookingResult and response:**
- [x] Add `ServiceFeeAmount int64` to `BookingResult` struct
- [x] Add `service_fee_amount` to `bookingResponse` struct in handler
- [x] Include in booking API response

**Tests:**
- [x] `internal/service/service_fee_service_test.go`: test priority resolution (region+category > region > global), test CalculateFee with various percentages (0%, 10%, 25%), test rounding
- [x] `internal/service/booking_service_test.go`: update existing booking creation tests to include service fee mock, verify ServiceFeeAmount is set correctly on booking
- [x] Run `go test ./... -v` — must pass

### Task 2: Extended Pricing Rules — Long Session Discount & Extra Guest Surcharge (FR-084, FR-085)

**Files:**
- Modify: `internal/domain/bathhouse.go`
- Modify: `internal/service/pricing_service.go`
- Modify: `internal/service/booking_service.go`
- Modify: `internal/handler/booking_handler.go`
- Modify: `internal/handler/bathhouse_handler.go`
- Modify: `internal/repository/postgres/bathhouse_repo.go`
- Create: `migrations/XXXXXX_pricing_enhancements.up.sql` / `.down.sql`

**Bathhouse model changes (`internal/domain/bathhouse.go`):**
- [x] Add fields to Bathhouse struct:
  ```go
  LongSessionThresholdHours  int   // default 4, range 1-12
  LongSessionDiscountPercent int   // default 0, range 0-50
  BaseCapacity               int   // default equals MaxGuests, minimum 1
  ExtraGuestSurcharge        int64 // kopecks per extra guest per hour, default 0
  ```
- [x] Update Validate(): LongSessionDiscountPercent 0-50, LongSessionThresholdHours 1-12, BaseCapacity >= 1 && BaseCapacity <= MaxGuests, ExtraGuestSurcharge >= 0

**Migration:**
- [x] ALTER TABLE bathhouses ADD COLUMN:
  - `long_session_threshold_hours INT NOT NULL DEFAULT 4`
  - `long_session_discount_percent INT NOT NULL DEFAULT 0`
  - `base_capacity INT` (initially set to max_guests via UPDATE, then NOT NULL)
  - `extra_guest_surcharge BIGINT NOT NULL DEFAULT 0`
- [x] CHECK constraints: `long_session_discount_percent BETWEEN 0 AND 50`, `long_session_threshold_hours BETWEEN 1 AND 12`, `base_capacity >= 1`, `extra_guest_surcharge >= 0`

**Modify PricingService (`internal/service/pricing_service.go`):**
- [x] Extend `CalculatePrice` signature or create new `CalculateFullPrice` method that accepts additional params:
  ```go
  type PriceCalculationInput struct {
      BathhouseID   uuid.UUID
      BasePrice     int64
      StartTime     time.Time
      EndTime       time.Time
      GuestCount    int
      BaseCapacity  int
      ExtraGuestSurcharge        int64
      LongSessionThresholdHours  int
      LongSessionDiscountPercent int
  }
  type PriceBreakdown struct {
      BasePrice            int64
      LongSessionDiscount  int64
      ExtraGuestSurcharge  int64
  }
  ```
- [x] Long session discount logic:
  - `durationHours = EndTime.Sub(StartTime) / time.Hour`
  - If `durationHours >= LongSessionThresholdHours && LongSessionDiscountPercent > 0`:
    - `discountableHours = durationHours - LongSessionThresholdHours`
    - `discount = hourlyRate * discountableHours * LongSessionDiscountPercent / 100`
- [x] Extra guest surcharge logic:
  - If `GuestCount > BaseCapacity`:
    - `extraGuests = GuestCount - BaseCapacity`
    - `surcharge = ExtraGuestSurcharge * int64(extraGuests) * int64(durationHours)`

**Modify booking creation (`internal/service/booking_service.go`):**
- [x] After base price calculation (~line 167), calculate long session discount and extra guest surcharge
- [x] Apply: `totalPrice = basePrice - longSessionDiscount + extraGuestSurcharge`
- [x] Pass breakdown to BookingResult

**Modify BookingResult and response:**
- [x] Add `LongSessionDiscount int64`, `ExtraGuestSurcharge int64`, `BasePrice int64` fields to BookingResult
- [x] Add corresponding fields to `bookingResponse` in handler
- [x] Return full price breakdown in API response:
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

**Modify bathhouse update handler:**
- [x] Accept new fields in PUT /api/v1/my/bathhouses/{id} request
- [x] Validate ranges in handler before passing to service

**Postgres repo (`internal/repository/postgres/bathhouse_repo.go`):**
- [x] Add new columns to SELECT queries (GetByID, List, etc.)
- [x] Add new columns to INSERT/UPDATE queries

**Tests:**
- [x] `internal/service/pricing_service_test.go`: test long session discount (3h booking with 4h threshold = no discount; 6h booking with 4h threshold and 10% = discount on 2h; 8h booking); test extra guest surcharge (2 guests with base_capacity=2 = no surcharge; 5 guests with base_capacity=3 = 2 extra guests * surcharge * hours); test combined discount + surcharge
- [x] Update booking service tests for new price breakdown fields
- [x] Run `go test ./... -v` — must pass

### Task 3: Holiday Prices (FR-088)

**Files:**
- Create: `internal/domain/holiday.go`
- Create: `internal/repository/postgres/holiday_repo.go`
- Create: `internal/repository/mock/holiday_repo.go`
- Create: `internal/service/holiday_service.go`
- Create: `internal/handler/holiday_handler.go`
- Modify: `internal/repository/interfaces.go`
- Modify: `internal/service/pricing_service.go`
- Create: `migrations/XXXXXX_holidays.up.sql` / `.down.sql`

**Domain model (`internal/domain/holiday.go`):**
- [x] Define structs:
  ```go
  type Holiday struct {
      ID          uuid.UUID
      Name        string     // e.g. "Новый год", "23 февраля"
      Date        time.Time  // specific date (only year-month-day matters)
      Region      string     // "RU", "BY"
      IsRecurring bool       // if true, same month-day every year
      CreatedAt   time.Time
      UpdatedAt   time.Time
  }
  type BathhouseHolidayPrice struct {
      BathhouseID uuid.UUID
      Multiplier  float64   // range 1.0-2.0, step 0.1, platform default 1.5
  }
  ```
- [x] Validate: Name non-empty, Region in ("RU", "BY"), Multiplier 1.0-2.0

**Migration:**
- [x] Create `holidays` table:
  ```sql
  CREATE TABLE holidays (
      id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
      name VARCHAR(200) NOT NULL,
      date DATE NOT NULL,
      region VARCHAR(10) NOT NULL DEFAULT 'RU',
      is_recurring BOOLEAN NOT NULL DEFAULT false,
      created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
      updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
  );
  CREATE INDEX idx_holidays_date ON holidays (date);
  CREATE INDEX idx_holidays_region ON holidays (region);
  ```
- [x] Create `bathhouse_holiday_prices` table:
  ```sql
  CREATE TABLE bathhouse_holiday_prices (
      bathhouse_id UUID NOT NULL REFERENCES bathhouses(id) ON DELETE CASCADE,
      multiplier NUMERIC(3,1) NOT NULL DEFAULT 1.5 CHECK (multiplier >= 1.0 AND multiplier <= 2.0),
      PRIMARY KEY (bathhouse_id)
  );
  ```
- [x] Seed initial Russian holidays: New Year (Jan 1-8), Defender of Fatherland (Feb 23), International Women's (Mar 8), Spring/Labour (May 1), Victory Day (May 9), Russia Day (Jun 12), National Unity (Nov 4) — all with is_recurring=true

**Repository (`internal/repository/interfaces.go`):**
- [x] Add `HolidayRepository` interface:
  ```go
  type HolidayRepository interface {
      Create(ctx context.Context, holiday *domain.Holiday) error
      Update(ctx context.Context, holiday *domain.Holiday) error
      Delete(ctx context.Context, id uuid.UUID) error
      GetByID(ctx context.Context, id uuid.UUID) (*domain.Holiday, error)
      ListAll(ctx context.Context) ([]domain.Holiday, error)
      ListByRegion(ctx context.Context, region string) ([]domain.Holiday, error)
      IsHoliday(ctx context.Context, date time.Time, region string) (*domain.Holiday, error)
      GetBathhouseMultiplier(ctx context.Context, bathhouseID uuid.UUID) (float64, error)
      SetBathhouseMultiplier(ctx context.Context, bathhouseID uuid.UUID, multiplier float64) error
  }
  ```
- [x] `IsHoliday` query: match exact date OR (is_recurring=true AND EXTRACT(MONTH)=month AND EXTRACT(DAY)=day)

**Postgres implementation:**
- [x] Implement all HolidayRepository methods
- [x] `IsHoliday`: `SELECT * FROM holidays WHERE region=$1 AND (date = $2::date OR (is_recurring = true AND EXTRACT(MONTH FROM date) = EXTRACT(MONTH FROM $2::date) AND EXTRACT(DAY FROM date) = EXTRACT(DAY FROM $2::date))) LIMIT 1`
- [x] `GetBathhouseMultiplier`: SELECT multiplier FROM bathhouse_holiday_prices WHERE bathhouse_id=$1, return 1.5 default if not found

**Mock implementation:**
- [x] In-memory implementation for testing

**Service (`internal/service/holiday_service.go`):**
- [x] Define `HolidayService` interface and implementation
- [x] CRUD methods delegating to repo
- [x] `IsHolidayDate(ctx, date, region) (*domain.Holiday, float64, error)` — returns holiday info and applicable multiplier (bathhouse-specific or platform default 1.5)

**Integrate into PricingService (`internal/service/pricing_service.go`):**
- [x] Inject `HolidayService` into pricingService
- [x] In `CalculatePrice`, before applying dynamic rules, check if the booking date is a holiday via `holidaySvc.IsHolidayDate(ctx, startTime, region)`
- [x] If holiday: multiply the base price by multiplier BEFORE hourly slot iteration
- [x] This means holiday multiplier stacks with dynamic pricing rules

**Handler (`internal/handler/holiday_handler.go`):**
- [x] Admin CRUD endpoints (RequireRole: admin):
  - `GET /api/v1/admin/holidays` — list all holidays, optional ?region= filter
  - `POST /api/v1/admin/holidays` — create holiday (name, date, region, is_recurring)
  - `PUT /api/v1/admin/holidays/{id}` — update holiday
  - `DELETE /api/v1/admin/holidays/{id}` — delete holiday
- [x] Owner endpoint (RequireAuth + CanManageBathhouse):
  - `PUT /api/v1/my/bathhouses/{id}/holiday-multiplier` — set custom multiplier { "multiplier": 1.3 }
- [x] Add swagger annotations

**Modify price response:**
- [x] Add `is_holiday_price bool`, `holiday_name string`, `holiday_multiplier float64` to PriceBreakdown and booking response

**Tests:**
- [x] `internal/service/holiday_service_test.go`: test IsHolidayDate with recurring vs non-recurring, test region filter, test bathhouse multiplier override vs default
- [x] `internal/service/pricing_service_test.go`: test CalculatePrice on a holiday (base 1000 kopecks/h * 1.5 multiplier = 1500/h), test holiday + dynamic rule stacking, test non-holiday date returns normal price
- [x] Test seeded holidays exist after migration
- [x] Run `go test ./... -v` — must pass

### Task 4: Last-Minute Discounts (FR-090)

**Files:**
- Modify: `internal/domain/bathhouse.go`
- Modify: `internal/service/pricing_service.go`
- Modify: `internal/service/bathhouse_service.go`
- Modify: `internal/handler/bathhouse_handler.go`
- Modify: `internal/repository/postgres/bathhouse_repo.go`
- Modify: `internal/domain/filters.go` (or wherever BathhouseFilter is defined)
- Create: `migrations/XXXXXX_last_minute.up.sql` / `.down.sql`

**Bathhouse model changes:**
- [x] Add fields to Bathhouse struct:
  ```go
  LastMinuteEnabled         bool  // default false
  LastMinuteDiscountPercent int   // default 20, range 5-50
  LastMinuteHoursThreshold  int   // default 6, range 2-24
  ```
- [x] Update Validate(): if LastMinuteEnabled, DiscountPercent must be 5-50, HoursThreshold must be 2-24

**Migration:**
- [x] ALTER TABLE bathhouses ADD COLUMN:
  - `last_minute_enabled BOOLEAN NOT NULL DEFAULT false`
  - `last_minute_discount_percent INT NOT NULL DEFAULT 20`
  - `last_minute_hours_threshold INT NOT NULL DEFAULT 6`
- [x] CHECK constraints: `last_minute_discount_percent BETWEEN 5 AND 50`, `last_minute_hours_threshold BETWEEN 2 AND 24`

**Modify GetAvailableSlots (`internal/service/booking_service.go`):**
- [x] In the slot iteration loop (~line 616-654), after calculating slotPrice:
  - If `bh.LastMinuteEnabled && slot.StartTime.Sub(time.Now()).Hours() <= float64(bh.LastMinuteHoursThreshold)`:
    - `discountedPrice = slotPrice - (slotPrice * int64(bh.LastMinuteDiscountPercent) / 100)`
    - Set `slot.IsLastMinute = true`, `slot.OriginalPrice = slotPrice`, `slot.Price = discountedPrice`
- [x] Extend `TimeSlot` struct:
  ```go
  type TimeSlot struct {
      StartTime     time.Time `json:"start_time"`
      EndTime       time.Time `json:"end_time"`
      Available     bool      `json:"available"`
      Price         int64     `json:"price"`
      IsLastMinute  bool      `json:"is_last_minute,omitempty"`
      OriginalPrice int64     `json:"original_price,omitempty"`
  }
  ```

**Modify booking creation:**
- [x] When calculating total price for a booking, check if each hourly slot qualifies for last-minute discount
- [x] Apply discount per-slot, not flat on total (some slots may be within threshold, some not)
- [x] Add `LastMinuteDiscount int64` to BookingResult and response

**Search filter (`internal/domain/filters.go` or BathhouseFilter):**
- [x] Add `LastMinute *bool` field to BathhouseFilter
- [x] In bathhouse list query, if LastMinute=true: filter WHERE last_minute_enabled=true AND EXISTS (available slot within threshold hours)
- [x] Simpler approach: just filter `WHERE last_minute_enabled = true` and let client check available slots

**Bathhouse list response:**
- [x] Add `is_last_minute bool` and `last_minute_discount_percent int` to bathhouse response when applicable
- [x] Set `is_last_minute = true` when LastMinuteEnabled AND any slots are within threshold

**Owner settings:**
- [x] Accept LastMinuteEnabled, LastMinuteDiscountPercent, LastMinuteHoursThreshold in PUT /api/v1/my/bathhouses/{id}

**Tests:**
- [x] Test GetAvailableSlots with last-minute enabled: slot starting in 3h (threshold 6h) should have is_last_minute=true and discounted price
- [x] Test slot starting in 8h (threshold 6h) should have is_last_minute=false
- [x] Test booking creation applies last-minute discount correctly per-slot
- [x] Test BathhouseFilter with last_minute=true
- [x] Test disabled last-minute (LastMinuteEnabled=false) returns normal prices
- [x] Run `go test ./... -v` — must pass

### Task 5: Buffer Time, Lead Time & Booking Settings (FR-073, FR-077, FR-078)

**Files:**
- Modify: `internal/domain/bathhouse.go`
- Modify: `internal/service/booking_service.go`
- Modify: `internal/service/bathhouse_service.go`
- Modify: `internal/handler/bathhouse_handler.go`
- Modify: `internal/repository/postgres/bathhouse_repo.go`
- Create: `migrations/XXXXXX_booking_settings.up.sql` / `.down.sql`

**Bathhouse model changes:**
- [x] Add fields to Bathhouse struct:
  ```go
  BufferMinutes   int     // default 30, range 0-120, step 15 — cleanup time between bookings
  LeadTimeHours   int     // default 2, range 0-48 — minimum hours before booking start
  MaxAdvanceDays  int     // default 90, range 7-365 — max days ahead for booking
  ```
- [x] Note: MinDuration already exists as `MinDuration int` on Bathhouse, re-use it (rename to MinDurationHours if needed for clarity, but keep backward compatibility)
- [x] Update Validate(): BufferMinutes 0-120 and divisible by 15, LeadTimeHours 0-48, MaxAdvanceDays 7-365

**Migration:**
- [x] ALTER TABLE bathhouses ADD COLUMN:
  - `buffer_minutes INT NOT NULL DEFAULT 30`
  - `lead_time_hours INT NOT NULL DEFAULT 2`
  - `max_advance_days INT NOT NULL DEFAULT 90`
- [x] CHECK: `buffer_minutes BETWEEN 0 AND 120`, `lead_time_hours BETWEEN 0 AND 48`, `max_advance_days BETWEEN 7 AND 365`

**Modify GetAvailableSlots (`internal/service/booking_service.go`):**
- [x] After checking overlapping bookings (line ~626-631), also add buffer zone: mark slots as unavailable if they fall within BufferMinutes before/after a booked slot
  - For each booked slot [B.StartTime, B.EndTime], mark as unavailable: [B.EndTime, B.EndTime + BufferMinutes]
- [x] Only show slots where `slot.StartTime >= time.Now().Add(time.Duration(bh.LeadTimeHours) * time.Hour)` — enforce lead time
- [x] Only show slots where `slot.StartTime.Before(time.Now().AddDate(0, 0, bh.MaxAdvanceDays))` — enforce max advance days

**Modify booking validation in Create (`internal/service/booking_service.go`):**
- [x] Replace hardcoded `5 * time.Minute` lead time check (line 122) with `time.Duration(bh.LeadTimeHours) * time.Hour`
- [x] Add validation: `if input.StartTime.After(time.Now().AddDate(0, 0, bh.MaxAdvanceDays))` return ErrInvalidInput("exceeds max advance days")
- [x] Add buffer conflict validation: check that new booking doesn't overlap with buffer zone of adjacent bookings
  - Query bookings that end within BufferMinutes before the new start or start within BufferMinutes after the new end
  - Use `bookingRepo.GetOverlapping(ctx, bathhouseID, startTime.Add(-bufferDuration), endTime.Add(bufferDuration))` and check for conflicts

**Owner settings:**
- [x] Accept BufferMinutes, LeadTimeHours, MaxAdvanceDays in PUT /api/v1/my/bathhouses/{id}
- [x] Validate step: BufferMinutes must be multiple of 15

**Postgres repo:**
- [x] Add new columns to all bathhouse queries

**Tests:**
- [x] Test GetAvailableSlots with buffer: booking 10:00-12:00 with 30min buffer should block 12:00-12:30
- [x] Test lead time: booking starting in 1h with lead_time_hours=2 should be rejected
- [x] Test max advance: booking 100 days ahead with max_advance_days=90 should be rejected
- [x] Test buffer conflict: new booking 12:00-14:00 when existing is 10:00-12:00 with 30min buffer should be rejected (buffer zone 12:00-12:30 overlaps)
- [x] Test BookingSettings with 0 values (buffer=0 means no buffer, lead_time=0 means immediate booking OK)
- [x] Run `go test ./... -v` — must pass

### Task 6: Booking Mode & Request-based Booking (FR-058)

**Files:**
- Modify: `internal/domain/bathhouse.go`
- Modify: `internal/domain/booking.go`
- Modify: `internal/service/booking_service.go`
- Modify: `internal/handler/booking_handler.go`
- Create: `migrations/XXXXXX_booking_request_mode.up.sql` / `.down.sql`

**Bathhouse model changes:**
- [x] Add fields:
  ```go
  BookingMode    string // "instant" or "request", default "instant"
  RequestTimeout int    // hours, default 24, range 1-72
  ```
- [x] Add `BookingModeInstant = "instant"` and `BookingModeRequest = "request"` constants
- [x] Update Validate(): BookingMode must be "instant" or "request", RequestTimeout 1-72

**Booking model changes (`internal/domain/booking.go`):**
- [x] Add fields:
  ```go
  HoldID          *uuid.UUID // reference to wallet hold (for request-based bookings)
  RejectionReason string     // owner's reason for rejection
  ```
- [x] Add status constant: `BookingPendingOwner BookingStatus = "pending_owner"`
- [x] Update `IsValid()` to include `BookingPendingOwner`

**Migration:**
- [x] ALTER TABLE bathhouses:
  - `ADD COLUMN booking_mode VARCHAR(10) NOT NULL DEFAULT 'instant'`
  - `ADD COLUMN request_timeout INT NOT NULL DEFAULT 24`
  - CHECK: `booking_mode IN ('instant', 'request')`, `request_timeout BETWEEN 1 AND 72`
- [x] ALTER TABLE bookings:
  - `ADD COLUMN hold_id UUID`
  - `ADD COLUMN rejection_reason TEXT`

**Modify CreateBooking (`internal/service/booking_service.go`):**
- [x] After all price calculations and validation (around line ~258), add branching:
  ```go
  if bh.BookingMode == domain.BookingModeRequest {
      // 1. Create booking with status "pending_owner" instead of "pending"
      booking.Status = domain.BookingPendingOwner
      // 2. Create wallet hold for total amount
      hold, err := s.walletSvc.Hold(ctx, userID, totalPrice, booking.ID)
      if err != nil { return nil, err }
      booking.HoldID = &hold.ID
      // 3. Save booking
      if err := s.bookingRepo.Create(ctx, booking); err != nil { return nil, err }
      // 4. Notify owner about new booking request
      s.sendBookingRequestNotification(ctx, booking, bh)
      return result, nil
  }
  // existing instant flow continues...
  ```
- [x] Inject `walletSvc WalletService` into bookingService (add to struct and constructor)

**New owner endpoints (handler + service):**
- [x] Add to BookingService interface:
  ```go
  Approve(ctx context.Context, userID uuid.UUID, role domain.UserRole, bookingID uuid.UUID) error
  ```
- [x] `Approve` implementation:
  1. Get booking, verify status == "pending_owner"
  2. CanManageBathhouse check
  3. If booking.HoldID != nil: `walletSvc.CaptureHold(ctx, *booking.HoldID)`
  4. `bookingRepo.UpdateStatus(ctx, bookingID, domain.BookingConfirmed)`
  5. Notify client: "Ваша заявка одобрена!"
- [x] Modify existing `Reject` method:
  - If booking status is "pending_owner":
    1. If booking.HoldID != nil: `walletSvc.ReleaseHold(ctx, *booking.HoldID)`
    2. Accept optional rejection reason
    3. Update status to "rejected" with reason
    4. Notify client: "К сожалению, ваша заявка отклонена. Причина: {reason}"
- [x] Handler routes:
  - `PATCH /api/v1/bookings/{id}/approve` — calls Approve (RequireAuth + owner/rep)
  - Existing `PATCH /api/v1/bookings/{id}/reject` — updated to handle rejection_reason
  - Request body for reject: `{ "reason": "optional text" }`

**Cron — auto-reject timed-out requests:**
- [x] In `internal/cron/booking_jobs.go` (or new file):
  - Query bookings WHERE status='pending_owner' AND created_at + request_timeout < now()
  - For each: release wallet hold, set status='rejected', notify both parties
  - Add `ListTimedOutRequests(ctx context.Context, timeout time.Duration) ([]domain.Booking, error)` to BookingRepository
  - Run every 15 minutes

**Modify booking repo (`internal/repository/interfaces.go`):**
- [x] Add `Update(ctx context.Context, booking *domain.Booking) error` method (or use specific field updates)
- [x] Add `ListTimedOutRequests(ctx context.Context) ([]domain.Booking, error)` — returns pending_owner bookings past their timeout

**Notification helpers:**
- [x] `sendBookingRequestNotification(ctx, booking, bh)` — notify owner: "Новая заявка на бронирование от {date} {time}"
- [x] Add `domain.NotifBookingRequest` notification type

**Tests:**
- [x] Test instant mode: booking creation with BookingMode="instant" uses existing flow, status="pending"
- [x] Test request mode: booking creation with BookingMode="request" creates booking with status="pending_owner" and wallet hold
- [x] Test approve: status changes to confirmed, hold is captured, client notified
- [x] Test reject: status changes to rejected, hold released, client notified, rejection reason stored
- [x] Test auto-reject: request older than timeout is auto-rejected, hold released
- [x] Test approve with expired hold: should return error
- [x] Run `go test ./... -v` — must pass

### Task 7: SBP Payment Support (FR-092)

**Files:**
- Modify: `internal/payment/provider.go`
- Modify: `internal/payment/yookassa.go`
- Modify: `internal/domain/payment.go`
- Modify: `internal/handler/payment_handler.go`
- Modify: `internal/service/payment_service.go`

**Payment model changes (`internal/domain/payment.go`):**
- [x] Add PaymentMethod type and constants:
  ```go
  type PaymentMethod string
  const (
      PaymentMethodCard   PaymentMethod = "card"
      PaymentMethodSBP    PaymentMethod = "sbp"
      PaymentMethodWallet PaymentMethod = "wallet"
      PaymentMethodCombo  PaymentMethod = "combo"
  )
  ```
- [x] Add `PaymentMethod PaymentMethod` field to Payment struct (default "card")
- [x] Update Validate(): PaymentMethod must be valid

**Modify PaymentProvider interface (`internal/payment/provider.go`):**
- [x] Add `PaymentMethod` parameter to CreatePayment:
  ```go
  CreatePayment(ctx context.Context, req CreatePaymentRequest) (*PaymentResult, error)
  ```
  where CreatePaymentRequest:
  ```go
  type CreatePaymentRequest struct {
      Amount      int64
      Currency    string
      Description string
      ReturnURL   string
      Metadata    map[string]string
      Method      string  // "card" or "sbp"
      Capture     bool    // true for instant charge, false for hold
  }
  ```
- [x] Add `CapturePayment(ctx context.Context, externalID string, amount int64) error` for holds
- [x] Add `CancelPayment(ctx context.Context, externalID string) error` for releasing holds

**YooKassa implementation (`internal/payment/yookassa.go`):**
- [x] Update CreatePayment to use CreatePaymentRequest:
  - For "sbp": set `confirmation.type = "redirect"` (same as card, YooKassa handles SBP redirect)
  - For "card": existing flow
  - Set `Capture` from request (true = immediate charge, false = hold for request bookings)
- [x] Implement `CapturePayment`: call YooKassa capture API `POST /payments/{id}/capture`
- [x] Implement `CancelPayment`: call YooKassa cancel API `POST /payments/{id}/cancel`

**Modify PaymentService (`internal/service/payment_service.go`):**
- [x] Update `InitiatePayment` to accept payment method parameter
- [x] Pass method to provider.CreatePayment

**Handler (`internal/handler/payment_handler.go`):**
- [x] Add `payment_method` field to payment initiation request:
  ```go
  type initiatePaymentRequest struct {
      BookingID     string `json:"booking_id"`
      PaymentMethod string `json:"payment_method"` // "card" or "sbp", default "card"
  }
  ```
- [x] Validate payment method in handler

**Migration:**
- [x] ALTER TABLE payments ADD COLUMN `payment_method VARCHAR(10) NOT NULL DEFAULT 'card'`

**Tests:**
- [x] Test CreatePayment with method="card": existing behavior
- [x] Test CreatePayment with method="sbp": confirmation URL returned
- [x] Test CapturePayment: successful capture
- [x] Test CancelPayment: successful cancellation
- [x] Test webhook handling for SBP payments (status transitions are the same)
- [x] Run `go test ./... -v` — must pass

### Task 8: Combo Payment — Wallet + Card/SBP (FR-093)

**Files:**
- Modify: `internal/service/payment_service.go`
- Modify: `internal/handler/payment_handler.go`
- Modify: `internal/domain/payment.go`
- Create: `migrations/XXXXXX_combo_payments.up.sql` / `.down.sql`

**Payment model changes:**
- [x] Add fields to Payment struct:
  ```go
  WalletAmount int64  // kopecks paid from wallet
  CardAmount   int64  // kopecks paid by card/SBP
  ```

**Migration:**
- [x] ALTER TABLE payments:
  - `ADD COLUMN wallet_amount BIGINT NOT NULL DEFAULT 0`
  - `ADD COLUMN card_amount BIGINT NOT NULL DEFAULT 0`

**Modify PaymentService:**
- [x] New interface method:
  ```go
  InitiateComboPayment(ctx context.Context, userID uuid.UUID, bookingID uuid.UUID, req ComboPaymentRequest) (confirmationURL string, err error)
  ```
  ```go
  type ComboPaymentRequest struct {
      WalletAmount  int64         // amount to pay from wallet (0 = card only)
      CardAmount    int64         // amount to pay by card/SBP (0 = wallet only)
      PaymentMethod PaymentMethod // "card", "sbp" (for card portion)
  }
  ```
- [x] Inject `walletSvc WalletService` into paymentService
- [x] Combo payment flow implementation:
  1. Validate: `WalletAmount + CardAmount == booking.TotalPrice`
  2. Validate: `WalletAmount >= 0 && CardAmount >= 0`
  3. If `WalletAmount > 0`: check wallet balance, call `walletSvc.Spend(ctx, userID, WalletAmount, bookingID)`
  4. If `CardAmount > 0`: create card/SBP payment for CardAmount via provider
  5. If card payment creation fails: immediately `walletSvc.Refund(ctx, userID, WalletAmount, bookingID)` to rollback wallet debit
  6. If `CardAmount == 0` (full wallet): complete payment immediately, confirm booking (no card needed)
  7. Store WalletAmount, CardAmount, PaymentMethod on Payment record

**Handler changes:**
- [x] New request struct:
  ```go
  type initiatePaymentRequest struct {
      BookingID     string `json:"booking_id"`
      PaymentMethod string `json:"payment_method"`       // "card", "sbp", "wallet", "combo"
      WalletAmount  int64  `json:"wallet_amount,omitempty"` // for combo/wallet
      CardAmount    int64  `json:"card_amount,omitempty"`   // for combo
  }
  ```
- [x] Route to appropriate method based on payment_method:
  - "card"/"sbp": existing InitiatePayment flow
  - "wallet": call InitiateComboPayment with CardAmount=0
  - "combo": call InitiateComboPayment with both amounts

**Tests:**
- [x] Test full card payment: WalletAmount=0, CardAmount=total — existing flow
- [x] Test full wallet payment: WalletAmount=total, CardAmount=0 — wallet debited, booking confirmed immediately, no card charge
- [x] Test combo: WalletAmount=5000, CardAmount=10000 — wallet debited first, card payment created for remainder
- [x] Test combo with card failure: wallet debited, card fails — verify wallet is refunded
- [x] Test invalid amounts: WalletAmount + CardAmount != total — should return error
- [x] Test insufficient wallet balance: should return ErrInsufficientWalletBalance
- [x] Run `go test ./... -v` — must pass

### Task 9: Payment Hold for Request Bookings (FR-094)

**Files:**
- Modify: `internal/service/payment_service.go`
- Modify: `internal/service/booking_service.go`
- Modify: `internal/handler/booking_handler.go`

**Payment hold flow (integrates with Task 6 request-based booking):**
- [x] When booking is request-based AND payment is card/SBP:
  - Create payment with `Capture=false` in YooKassa (authorization hold)
  - Store external_id on payment record, status = "processing" (hold is active)
- [x] When booking is request-based AND payment is wallet:
  - Use existing `walletSvc.Hold()` (already implemented in Task 6)
- [x] When booking is request-based AND payment is combo:
  - Create wallet hold for wallet portion
  - Create card authorization (capture=false) for card portion
  - Both must succeed; if card auth fails, release wallet hold

**On owner approve:**
- [x] If card hold exists: call `provider.CapturePayment(ctx, externalID, amount)` to capture the authorized amount
- [x] If wallet hold exists: call `walletSvc.CaptureHold(ctx, holdID)`
- [x] If combo: capture both
- [x] Update payment status to "succeeded"
- [x] Confirm booking

**On owner reject / timeout:**
- [x] If card hold exists: call `provider.CancelPayment(ctx, externalID)` to release authorization
- [x] If wallet hold exists: call `walletSvc.ReleaseHold(ctx, holdID)`
- [x] If combo: release both
- [x] Update payment status to "failed"

**Max hold duration:**
- [x] YooKassa card authorization holds expire after 72 hours
- [x] In auto-reject cron (from Task 6): if `RequestTimeout > 72`, still auto-reject at 72h mark for card holds
- [x] Config: `BANI_MAX_CARD_HOLD_HOURS` (default 72)
- [x] If hold expires before owner responds: auto-reject booking

**Add to Payment model:**
- [x] `IsHold bool` field — indicates this payment is an authorization hold, not yet captured
- [x] `CapturedAt *time.Time` — when the hold was captured

**Tests:**
- [x] Test request booking with card: payment created with capture=false
- [x] Test approve with card hold: CapturePayment called, status becomes succeeded
- [x] Test reject with card hold: CancelPayment called, status becomes failed
- [x] Test request booking with combo: wallet hold + card auth created
- [x] Test approve combo: both captured
- [x] Test reject combo: both released
- [x] Test card auth failure in combo: wallet hold is released
- [x] Test timeout auto-reject: both holds released
- [x] Run `go test ./... -v` — must pass

### Task 10: Check-in/Check-out & No-show (FR-066, FR-070)

**Files:**
- Modify: `internal/domain/booking.go`
- Modify: `internal/service/booking_service.go`
- Modify: `internal/handler/booking_handler.go`
- Modify: `internal/repository/postgres/booking_repo.go`
- Create: `migrations/XXXXXX_booking_checkin.up.sql` / `.down.sql`

**Booking model changes:**
- [x] Add fields:
  ```go
  CheckedInAt  *time.Time
  CheckedOutAt *time.Time
  ```
- [x] Add status: `BookingNoShow BookingStatus = "no_show"`
- [x] Update `IsValid()` to include `BookingNoShow`

**Migration:**
- [x] ALTER TABLE bookings:
  - `ADD COLUMN checked_in_at TIMESTAMPTZ`
  - `ADD COLUMN checked_out_at TIMESTAMPTZ`

**New service methods (add to BookingService interface):**
- [x] `CheckIn(ctx context.Context, userID uuid.UUID, role domain.UserRole, bookingID uuid.UUID) error`:
  1. Get booking, verify status == "confirmed"
  2. CanManageBathhouse check (only owner/rep can check-in)
  3. Validate time window: `booking.StartTime.Add(-15*time.Minute) <= now <= booking.StartTime.Add(30*time.Minute)`
  4. Set `booking.CheckedInAt = &now`
  5. Save via `bookingRepo.Update(ctx, booking)` — need to add Update method to BookingRepo that can update specific fields
  6. Notify client: "Вы отмечены как прибывший"
- [x] `CheckOut(ctx context.Context, userID uuid.UUID, role domain.UserRole, bookingID uuid.UUID) error`:
  1. Get booking, verify CheckedInAt != nil (guest has checked in)
  2. CanManageBathhouse check
  3. Set `booking.CheckedOutAt = &now`, status = "completed"
  4. Save booking
  5. Trigger escrow creation (in Task 11)

**BookingRepository changes (`internal/repository/interfaces.go`):**
- [x] Add `UpdateCheckin(ctx context.Context, bookingID uuid.UUID, checkedInAt *time.Time) error`
- [x] Add `UpdateCheckout(ctx context.Context, bookingID uuid.UUID, checkedOutAt *time.Time, status domain.BookingStatus) error`
- [x] Add `ListConfirmedWithoutCheckin(ctx context.Context, noShowCutoff time.Time) ([]domain.Booking, error)` — for no-show cron

**Handler routes:**
- [x] `PATCH /api/v1/bookings/{id}/check-in` (RequireAuth + owner/rep)
- [x] `PATCH /api/v1/bookings/{id}/check-out` (RequireAuth + owner/rep)
- [x] `POST /api/v1/bookings/{id}/dispute-noshow` (RequireAuth, client only)
  - Request: `{ "gps_lat": 55.7, "gps_lon": 37.6, "comment": "optional" }`
  - Creates high-priority support ticket / complaint
- [x] Add swagger annotations

**No-show detection cron (in `internal/cron/booking_jobs.go`):**
- [x] Run every 15 minutes
- [x] Query: `ListConfirmedWithoutCheckin(ctx, time.Now().Add(-30*time.Minute))`
  - Find bookings WHERE status='confirmed' AND start_time + 30min < now AND checked_in_at IS NULL
- [x] For each no-show:
  1. Set status = "no_show"
  2. Do NOT refund (client is charged)
  3. Notify owner: "Гость не прибыл на бронирование {date} {time}, оплата сохранена"
  4. Notify client: "Вы не прибыли на бронирование. Если это ошибка, оспорьте в течение 2 часов"

**Owner reminder notification:**
- [x] In the reminder cron (Task 13), add: 5 min before session start, send push to owner: "Гость скоро прибудет в {bathhouse_name}!"

**No-show dispute:**
- [x] Client can dispute within 2 hours after no-show status set
- [x] Validate: booking.Status == "no_show" AND time.Since(booking.UpdatedAt) <= 2*time.Hour
- [x] Creates a complaint/support ticket with type "no_show_dispute"
- [x] Admin reviews and can reverse no-show to completed + refund if warranted

**Tests:**
- [x] Test check-in: valid window (start-15min to start+30min), owner role, status changes
- [x] Test check-in too early: 1 hour before — rejected
- [x] Test check-in too late: 45 minutes after start — rejected
- [x] Test check-out: sets CheckedOutAt, status=completed
- [x] Test check-out without check-in: rejected
- [x] Test no-show cron: confirmed booking 35 min past start without check-in -> marked no_show
- [x] Test no-show dispute: within 2h window succeeds, after 2h rejected
- [x] Test no-show for already checked-in booking: should not be affected
- [x] Run `go test ./... -v` — must pass

### Task 11: Escrow System (FR-124)

**Files:**
- Create: `internal/domain/escrow.go`
- Create: `internal/service/escrow_service.go`
- Create: `internal/repository/postgres/escrow_repo.go`
- Create: `internal/repository/mock/escrow_repo.go`
- Modify: `internal/repository/interfaces.go`
- Modify: `internal/service/booking_service.go` (check-out triggers escrow)
- Create: `migrations/XXXXXX_escrow.up.sql` / `.down.sql`

**Domain model (`internal/domain/escrow.go`):**
- [x] Define types:
  ```go
  type EscrowStatus string
  const (
      EscrowHeld     EscrowStatus = "held"
      EscrowReleased EscrowStatus = "released"
      EscrowDisputed EscrowStatus = "disputed"
      EscrowRefunded EscrowStatus = "refunded"
  )
  type Escrow struct {
      ID                uuid.UUID
      BookingID         uuid.UUID
      Amount            int64        // total payment amount in kopecks
      ServiceFee        int64        // platform service fee in kopecks
      Status            EscrowStatus
      ClaimPeriodEndsAt time.Time    // when funds can be released to owner
      ReleasedAt        *time.Time
      CreatedAt         time.Time
  }
  ```

**Migration:**
- [x] CREATE TABLE escrows:
  ```sql
  CREATE TABLE escrows (
      id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
      booking_id UUID NOT NULL REFERENCES bookings(id) UNIQUE,
      amount BIGINT NOT NULL,
      service_fee BIGINT NOT NULL DEFAULT 0,
      status VARCHAR(20) NOT NULL DEFAULT 'held',
      claim_period_ends_at TIMESTAMPTZ NOT NULL,
      released_at TIMESTAMPTZ,
      created_at TIMESTAMPTZ NOT NULL DEFAULT now()
  );
  CREATE INDEX idx_escrows_status ON escrows (status);
  CREATE INDEX idx_escrows_claim_period ON escrows (claim_period_ends_at) WHERE status = 'held';
  ```

**Repository (`internal/repository/interfaces.go`):**
- [x] Add `EscrowRepository` interface:
  ```go
  type EscrowRepository interface {
      Create(ctx context.Context, escrow *domain.Escrow) error
      GetByID(ctx context.Context, id uuid.UUID) (*domain.Escrow, error)
      GetByBookingID(ctx context.Context, bookingID uuid.UUID) (*domain.Escrow, error)
      UpdateStatus(ctx context.Context, id uuid.UUID, status domain.EscrowStatus, releasedAt *time.Time) error
      ListMatured(ctx context.Context) ([]domain.Escrow, error) // status=held AND claim_period_ends_at < now
  }
  ```

**Postgres & mock implementations:**
- [x] Implement EscrowRepository in postgres and mock

**Service (`internal/service/escrow_service.go`):**
- [x] Define `EscrowService` interface:
  ```go
  type EscrowService interface {
      CreateEscrow(ctx context.Context, bookingID uuid.UUID, amount int64, serviceFee int64) (*domain.Escrow, error)
      ReleaseToOwner(ctx context.Context, escrowID uuid.UUID) error
      MarkDisputed(ctx context.Context, escrowID uuid.UUID) error
      ProcessRefund(ctx context.Context, escrowID uuid.UUID, refundAmount int64) error
      ProcessMaturedEscrows(ctx context.Context) error // cron method
  }
  ```
- [x] `CreateEscrow`:
  - Create escrow with status "held"
  - ClaimPeriodEndsAt = now + BANI_ESCROW_CLAIM_HOURS (default 48h)
  - Config: `BANI_ESCROW_CLAIM_HOURS` (default 48, configurable 24-168)
- [x] `ReleaseToOwner`:
  1. Get escrow, verify status == "held" and ClaimPeriodEndsAt < now
  2. Credit owner wallet: `walletSvc.Credit(ctx, ownerID, escrow.Amount - escrow.ServiceFee)`
  3. Credit platform wallet: service fee (or just log it — platform doesn't have a wallet yet)
  4. Update status to "released", set ReleasedAt
- [x] `MarkDisputed`: update status to "disputed" (prevents auto-release)
- [x] `ProcessRefund`: refund client, update status to "refunded"
- [x] `ProcessMaturedEscrows`: query ListMatured(), call ReleaseToOwner for each

**Integrate with check-out (Task 10):**
- [x] In `CheckOut` method, after setting status=completed:
  - Get payment for booking to get amount and service fee
  - Call `escrowSvc.CreateEscrow(ctx, bookingID, payment.Amount, booking.ServiceFeeAmount)`

**Cron:**
- [x] Hourly job calling `escrowSvc.ProcessMaturedEscrows(ctx)`
- [x] Log each release: escrow_id, booking_id, owner_id, amount, service_fee

**Tests:**
- [x] Test CreateEscrow: correct claim period, status=held
- [x] Test ReleaseToOwner: after claim period, owner wallet credited with amount-fee
- [x] Test ReleaseToOwner before claim period: should be rejected
- [x] Test MarkDisputed: prevents release
- [x] Test ProcessMaturedEscrows: finds and releases all matured escrows
- [x] Test ProcessMaturedEscrows skips disputed escrows
- [x] Run `go test ./... -v` — must pass

### Task 12: Enhanced Refund Logic (FR-106-109)

**Files:**
- Modify: `internal/service/booking_service.go`
- Modify: `internal/service/payment_service.go`
- Modify: `internal/handler/booking_handler.go`
- Modify: `internal/handler/payment_handler.go`

**Refund destination choice:**
- [x] Modify cancel endpoint request to accept `refund_to`:
  ```go
  type cancelBookingRequest struct {
      RefundTo string `json:"refund_to"` // "wallet" or "card", default "card"
  }
  ```
- [x] Modify `Cancel` in BookingService to accept `refundTo string` parameter
- [x] Update handler to pass refund preference

**Wallet refund with bonus:**
- [x] When `refundTo == "wallet"`:
  - Calculate refund amount per existing tiers (100% if >24h, 50% if 2-24h)
  - Add bonus: `bonusAmount = refundAmount * BANI_WALLET_REFUND_BONUS_PERCENT / 100` (default 5%, range 0-15%)
  - Credit wallet: `walletSvc.Credit(ctx, userID, refundAmount + bonusAmount)`
  - Skip card refund via provider
  - Instant, no 3-10 day wait
- [x] When `refundTo == "card"`: existing flow via YooKassa CreateRefund

**Combo refund (proportional):**
- [x] When original payment was combo (WalletAmount > 0 AND CardAmount > 0):
  - Calculate total refund per tiers
  - `walletRefundRatio = originalPayment.WalletAmount / originalPayment.Amount`
  - `walletRefund = totalRefund * walletRefundRatio`
  - `cardRefund = totalRefund - walletRefund`
  - If `refundTo == "wallet"`: redirect card portion to wallet too (with bonus on wallet portion only)
  - If `refundTo == "card"`: wallet portion to wallet (with bonus), card portion to card
- [x] Actually simplify: wallet portion always refunds to wallet (with bonus), card portion follows refundTo preference

**Admin manual refund:**
- [x] Add to PaymentService:
  ```go
  AdminRefund(ctx context.Context, bookingID uuid.UUID, amount int64, reason string, refundTo string) error
  ```
- [x] Handler: `POST /api/v1/admin/bookings/{id}/refund` (RequireRole: admin)
  - Request: `{ "amount": 50000, "reason": "customer complaint", "refund_to": "wallet" }`
- [x] Create audit log entry with: refund amount, reason, initiator (admin user), destination

**Modify RefundPayment in PaymentService:**
- [x] Add `refundTo` parameter: `RefundPayment(ctx, bookingID, forceFullRefund, refundTo string) error`
- [x] Branch on refundTo for wallet vs card vs proportional combo refund

**Tests:**
- [x] Test wallet refund: 10000 kopecks refund with 5% bonus = 10500 credited to wallet
- [x] Test card refund: existing behavior via provider
- [x] Test combo refund proportional: paid 3000 wallet + 7000 card, refund 100% = 3150 to wallet (3000 + 5% bonus) + 7000 to card
- [x] Test combo refund all-to-wallet: 10000 total to wallet = 10500 (10000 + 5%)
- [x] Test admin manual refund: arbitrary amount, audit log created
- [x] Test refund with no refund tier (<2h): returns nil, no refund
- [x] Run `go test ./... -v` — must pass

### Task 13: Booking Reminders (FR-064)

**Files:**
- Create: `internal/cron/booking_jobs.go` (or extend if created in Task 10)

**Reminder cron job:**
- [x] Run every 15 minutes
- [x] Query upcoming confirmed bookings:
  - 24h reminder: bookings where `start_time BETWEEN now() + interval '23h 45m' AND now() + interval '24h 15m'` (15-min window)
  - 2h reminder: bookings where `start_time BETWEEN now() + interval '1h 45m' AND now() + interval '2h 15m'`

**Deduplication:**
- [x] Use Redis set key: `reminder_sent:{bookingID}:{type}` (type = "24h" or "2h")
- [x] SETNX with TTL 48h — if key exists, skip (already sent)
- [x] This prevents duplicate sends on cron overlap

**24h reminder content:**
- [x] Channel: push notification + email
- [x] Title: "Напоминание о бронировании"
- [x] Body: "Завтра в {time} — бронирование в {bathhouse_name}. Адрес: {address}"
- [x] Data includes: booking_id, bathhouse_id, maps_url (Yandex Maps link with lat/lon: `https://yandex.ru/maps/?pt={lon},{lat}&z=16`)

**2h reminder content:**
- [x] Channel: push notification only
- [x] Title: "Скоро бронирование"
- [x] Body: "Через 2 часа — {bathhouse_name}. Не забудьте!"
- [x] Data includes: booking_id, bathhouse_id

**Owner reminder (5 min before):**
- [x] Additional reminder to owner/rep: "Гость скоро прибудет в {bathhouse_name}!"
- [x] Push only, deduplicated with key `reminder_sent:{bookingID}:owner_5min`

**BookingRepository changes:**
- [x] Add `ListUpcoming(ctx context.Context, from, to time.Time) ([]domain.Booking, error)` — returns confirmed bookings with start_time in range
- [x] Need to join/fetch bathhouse data (name, address, lat, lon) for reminder content — may return enriched struct or fetch separately

**Tests:**
- [x] Test reminder scheduling: booking at 10:00 tomorrow, cron at 10:00 today -> 24h reminder triggered
- [x] Test deduplication: same booking, same type -> second call skipped
- [x] Test 2h reminder: booking at 14:00, cron at 12:00 -> 2h reminder triggered
- [x] Test owner 5-min reminder
- [x] Test no reminder for cancelled bookings
- [x] Run `go test ./... -v` — must pass

### Task 14: Session Extension (FR-065)

**Files:**
- Modify: `internal/service/booking_service.go`
- Modify: `internal/handler/booking_handler.go`

**New service method:**
- [x] Add to BookingService interface:
  ```go
  Extend(ctx context.Context, userID uuid.UUID, bookingID uuid.UUID, extraHours int) (*BookingResult, error)
  ```

**Implementation:**
- [x] Validate:
  1. Get booking, verify status is "confirmed" or checked_in (CheckedInAt != nil)
  2. Verify userID matches booking.UserID (client extends their own booking)
  3. `extraHours` must be 1 or 2
  4. New end time = booking.EndTime + extraHours hours
  5. Check new end time is within working hours: `validateWithinWorkingHours(bh, booking.StartTime, newEndTime)`
  6. Check no conflicting bookings in extended period: `bookingRepo.CheckAvailability(ctx, bathhouseID, booking.EndTime, newEndTime)`
  7. Check buffer time: no booking starts within buffer after newEndTime
  8. Check no slot blocks in extended period

**Price calculation:**
- [x] Calculate extension price: `pricingSvc.CalculatePrice(ctx, bathhouseID, bh.PricePerHour, booking.EndTime, newEndTime)`
- [x] Apply same pricing rules (holiday, last-minute if applicable, dynamic rules)
- [x] Do NOT re-apply service fee on extension (or apply based on business rule — check BRD)

**Update booking:**
- [x] Add `UpdateEndTime(ctx context.Context, bookingID uuid.UUID, newEndTime time.Time, newTotalPrice int64) error` to BookingRepository
- [x] Update booking.EndTime = newEndTime
- [x] Update booking.TotalPrice += extensionPrice

**Payment for extension:**
- [x] Create a new payment for the extension amount
- [x] Return confirmation URL for card payment, or debit wallet
- [x] Extension payment linked to same bookingID (need to support multiple payments per booking OR create separate extension payment)
- [x] Simpler: create new Payment record with metadata `{"type": "extension", "booking_id": "..."}`

**Handler:**
- [x] `POST /api/v1/bookings/{id}/extend` (RequireAuth, client)
  - Request: `{ "extra_hours": 1, "payment_method": "card" }`
  - Response: `{ "booking": {...updated}, "extension_price": 50000, "confirmation_url": "..." }`
- [x] Add swagger annotations

**Notifications:**
- [x] Notify owner: "Гость продлил сессию на {extra_hours}ч до {new_end_time}"
- [x] Notify client: "Сессия продлена до {new_end_time}"

**Tests:**
- [x] Test extend by 1h: next slot available -> success, EndTime updated, price recalculated
- [x] Test extend by 2h: both slots available -> success
- [x] Test extend with conflict: next slot booked -> rejected
- [x] Test extend beyond working hours: rejected
- [x] Test extend with status != confirmed/checked_in: rejected
- [x] Test extend by non-owner: rejected (client can only extend their own)
- [x] Run `go test ./... -v` — must pass

### Task 15: Re-booking from History (FR-059)

**Files:**
- Modify: `internal/service/booking_service.go`
- Modify: `internal/handler/booking_handler.go`

**New service method:**
- [x] Add to BookingService interface:
  ```go
  GetRebookData(ctx context.Context, userID uuid.UUID, bookingID uuid.UUID) (*RebookData, error)
  ```
  ```go
  type RebookData struct {
      BathhouseID   uuid.UUID
      DurationHours int
      TimeFrom      string // "10:00" — time-of-day from original booking
      TimeTo        string // "14:00"
      GuestCount    int
      AddOns        []RebookAddOn
  }
  type RebookAddOn struct {
      AddOnID  uuid.UUID
      Quantity int
  }
  ```

**Implementation:**
- [x] Get booking, verify:
  1. booking.UserID == userID
  2. booking.Status is "completed" or "cancelled" (only rebook from past bookings)
  3. bathhouse exists and is active
- [x] Calculate duration: `booking.EndTime.Sub(booking.StartTime) / time.Hour`
- [x] Extract time-of-day: `booking.StartTime.Format("15:04")`, `booking.EndTime.Format("15:04")`
- [x] Fetch add-ons from `addonRepo.ListByBooking(ctx, bookingID)`
- [x] Return data WITHOUT prices (recalculated at booking time with current rates)

**Handler:**
- [x] `GET /api/v1/bookings/{id}/rebook-data` (RequireAuth, client)
- [x] Response: RebookData as JSON
- [x] Add swagger annotations

**Tests:**
- [x] Test rebook from completed booking: returns correct data
- [x] Test rebook from cancelled booking: returns correct data
- [x] Test rebook from pending booking: rejected
- [x] Test rebook from other user's booking: forbidden
- [x] Test rebook with inactive bathhouse: appropriate error
- [x] Test rebook with add-ons: add-on IDs and quantities returned
- [x] Run `go test ./... -v` — must pass

### Task 16: Owner Cancellation Penalty & Response Rate (FR-069, FR-079)

**Files:**
- Modify: `internal/service/booking_service.go`
- Modify: `internal/domain/bathhouse.go`
- Modify: `internal/repository/postgres/bathhouse_repo.go`
- Modify: `internal/repository/interfaces.go`
- Create: `migrations/XXXXXX_owner_metrics.up.sql` / `.down.sql`

**Owner cancellation penalty:**
- [x] In `Cancel` method, when role is owner/representative:
  - After cancelling: credit 10% of booking.TotalPrice to client wallet as compensation
    - `compensationAmount = booking.TotalPrice / 10`
    - `walletSvc.Credit(ctx, booking.UserID, compensationAmount)`
  - Track cancellation: increment owner cancellation count
- [x] Add to BookingRepository:
  ```go
  CountOwnerCancellations(ctx context.Context, ownerID uuid.UUID, since time.Time) (int, error)
  ```
  — counts bookings cancelled by owner in the given time window
- [x] After cancellation, check 30-day count:
  - count > 3: send warning notification to owner
  - count > 5: deactivate ALL owner's bathhouses, notify admin
    - `bhRepo.ListByOwner` -> for each: `bhRepo.UpdateStatus(ctx, id, BathhouseStatusInactive)`
    - Notify admin about auto-deactivation

**Bathhouse model changes:**
- [x] Add fields:
  ```go
  ResponseRate          float64 // 0.0-1.0, percentage of requests responded to within timeout
  AvgResponseTimeMinutes int    // average response time in minutes for request-based bookings
  ```

**Migration:**
- [x] ALTER TABLE bathhouses:
  - `ADD COLUMN response_rate NUMERIC(5,4) DEFAULT 1.0`
  - `ADD COLUMN avg_response_time_minutes INT DEFAULT 0`
- [x] ALTER TABLE bookings:
  - `ADD COLUMN cancelled_by_owner BOOLEAN NOT NULL DEFAULT false` (to distinguish client vs owner cancellation)

**Response rate tracking (for request-based bookings):**
- [x] Add to BookingRepository:
  ```go
  GetResponseStats(ctx context.Context, bathhouseID uuid.UUID, since time.Time) (totalRequests int, respondedInTime int, avgResponseMinutes int, error)
  ```
  — queries bookings with status != pending_owner (responded) vs total request-based bookings
- [x] Daily cron job to recalculate response rate for all request-mode bathhouses:
  ```go
  responseRate = respondedInTime / totalRequests (over last 90 days)
  avgResponseMinutes = AVG(updated_at - created_at) for responded requests
  ```
  - Update bathhouse.ResponseRate and bathhouse.AvgResponseTimeMinutes
- [x] Enforcement:
  - ResponseRate < 0.5: send warning notification to owner "Ваш процент ответов {rate}%, рекомендуем отвечать быстрее"
  - ResponseRate < 0.3 for 60+ consecutive days: force booking_mode to "instant" OR deactivate listing
  - Track how long rate has been below 0.3 (check previous day's rate from bathhouse record)

**Modify Cancel to track owner cancellations:**
- [x] Set `booking.CancelledByOwner = true` when owner/rep cancels
- [x] Use this flag for counting owner cancellations

**Tests:**
- [x] Test owner cancellation: client receives 10% compensation in wallet
- [x] Test 4th cancellation in 30 days: warning notification sent to owner
- [x] Test 6th cancellation: all owner bathhouses deactivated, admin notified
- [x] Test response rate calculation: 8 out of 10 requests responded in time = 0.80
- [x] Test low response rate warning at 0.5
- [x] Test auto-deactivation at 0.3 for 60 days
- [x] Run `go test ./... -v -race` — must pass

### Task 17: Fiscalization Placeholder (FR-095)

**Files:**
- Create: `internal/fiscal/fiscal.go`
- Create: `internal/fiscal/atol.go`
- Create: `internal/fiscal/module.go`
- Modify: `internal/service/payment_service.go`

**Interface (`internal/fiscal/fiscal.go`):**
- [ ] Define:
  ```go
  type ReceiptType string
  const (
      ReceiptAdvance       ReceiptType = "advance"
      ReceiptAdvanceCredit ReceiptType = "advance_credit"
      ReceiptFullPayment   ReceiptType = "full_payment"
      ReceiptRefund        ReceiptType = "refund"
  )
  type ReceiptRequest struct {
      Type     ReceiptType
      Amount   int64    // kopecks
      Email    string
      Phone    string
      Items    []ReceiptItem
  }
  type ReceiptItem struct {
      Name     string
      Quantity int
      Price    int64  // kopecks
      VAT      string // "none", "vat0", "vat10", "vat20"
  }
  type Receipt struct {
      ID         string
      ExternalID string
      Status     string
  }
  type FiscalProvider interface {
      CreateReceipt(ctx context.Context, req ReceiptRequest) (*Receipt, error)
  }
  ```

**ATOL placeholder (`internal/fiscal/atol.go`):**
- [ ] Implement `ATOLProvider` struct:
  ```go
  type ATOLProvider struct {
      login     string
      password  string
      groupCode string
      logger    *logger.Logger
  }
  ```
- [ ] `CreateReceipt`: log the request, return mock Receipt with generated ID
  - `logger.Info("ATOL receipt created (placeholder)", "type", req.Type, "amount", req.Amount)`
  - Return `&Receipt{ID: uuid.New().String(), Status: "pending"}`
- [ ] Config vars: `BANI_FISCAL_PROVIDER` (default "none"), `BANI_FISCAL_ATOL_LOGIN`, `BANI_FISCAL_ATOL_PASSWORD`, `BANI_FISCAL_ATOL_GROUP_CODE`

**No-op provider:**
- [ ] When `BANI_FISCAL_PROVIDER == "none"`: use no-op provider that does nothing and returns nil

**fx module (`internal/fiscal/module.go`):**
- [ ] Register FiscalProvider based on config:
  ```go
  fx.Module("fiscal",
      fx.Provide(func(cfg *config.Config, log *logger.Logger) FiscalProvider {
          switch cfg.Fiscal.Provider {
          case "atol": return NewATOLProvider(cfg.Fiscal.ATOLLogin, ...)
          default: return NewNoOpProvider()
          }
      }),
  )
  ```

**Hook into payment flow (`internal/service/payment_service.go`):**
- [ ] Inject `fiscalProvider fiscal.FiscalProvider` into paymentService
- [ ] On successful payment (HandleWebhook, status "succeeded"):
  - Call `fiscalProvider.CreateReceipt(ctx, ReceiptRequest{Type: ReceiptAdvance, Amount: payment.Amount, ...})`
  - Log receipt ID on payment metadata
- [ ] On refund (RefundPayment):
  - Call `fiscalProvider.CreateReceipt(ctx, ReceiptRequest{Type: ReceiptRefund, Amount: refundAmount, ...})`
  - Best-effort: don't fail the refund if receipt creation fails (log error)

**Tests:**
- [ ] Test ATOLProvider.CreateReceipt: returns receipt with ID
- [ ] Test NoOpProvider.CreateReceipt: returns nil, no error
- [ ] Test payment webhook triggers receipt creation
- [ ] Test refund triggers refund receipt creation
- [ ] Test receipt failure doesn't block payment/refund flow
- [ ] Run `go test ./... -v -race` — must pass

### Task 18: Verify acceptance criteria

- [ ] Manual test: create booking with request mode, approve, check-in, check-out, verify escrow creation, wait for claim period, verify release to owner wallet
- [ ] Manual test: booking with holiday pricing + last-minute discount + service fee — verify price breakdown is correct
- [ ] Manual test: combo payment (wallet + card), then cancel with wallet refund + bonus — verify proportional refund
- [ ] Manual test: session extension during checked-in booking — verify slot availability check and price calculation
- [ ] Manual test: owner cancels 6 bookings in 30 days — verify auto-deactivation
- [ ] Run full test suite: `go test ./... -v -race`
- [ ] Run linter: `make lint`
- [ ] Verify test coverage meets 80%+

### Task 19: Update documentation

- [ ] Update CLAUDE.md:
  - Add new domain errors: ErrEscrowNotFound, ErrEscrowAlreadyReleased, etc.
  - Add new config vars: BANI_ESCROW_CLAIM_HOURS, BANI_WALLET_REFUND_BONUS_PERCENT, BANI_FISCAL_PROVIDER, BANI_MAX_CARD_HOLD_HOURS
  - Document new booking statuses: pending_owner, no_show
  - Document new payment methods: sbp, wallet, combo
  - Add cron jobs: reminder, no-show detection, escrow release, response rate calculation, auto-reject timed-out requests
- [ ] Run `make swagger` to regenerate OpenAPI spec
- [ ] Move this plan to `docs/plans/completed/`
