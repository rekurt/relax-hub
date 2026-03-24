---
# Subsystem 7: Booking Enhancements (FR-058-081)

## Overview
Request-based booking mode, check-in/check-out with no-show detection, booking reminders, session extension, re-booking from history, buffer/lead times, owner cancellation penalties, and response rate monitoring.

## Context
- Existing booking: `internal/service/booking_service.go`, `internal/handler/booking_handler.go`
- Booking model: `internal/domain/booking.go` — has Status (pending/confirmed/cancelled/completed), Date, TimeFrom, TimeTo, TotalPrice
- Available slots: calculated in `internal/service/bathhouse_service.go` (GetAvailableSlots)
- Cancellation: existing refund logic with time-based tiers (100% if >24h, 50% if 2-24h, 0% if <2h)
- Existing notification system: `internal/notification/`
- Existing cron: needs centralized scheduler

## Dependencies
- Depends on: Wallet System (Subsystem 1) — for holds and wallet payments
- Depends on: Add-ons (Subsystem 5) — for add-on integration in booking
- Uses existing payment, notification subsystems

## Development Approach
- **Testing approach**: Regular (code first, then tests)
- **CRITICAL: every task MUST include new/updated tests**
- **CRITICAL: all tests must pass before starting next task**

## Implementation Steps

### Task 7.1: Booking Mode & Request-based Booking (FR-058)

**Files:**
- Modify: `internal/domain/bathhouse.go`
- Modify: `internal/domain/booking.go`
- Modify: `internal/service/booking_service.go`
- Modify: `internal/handler/booking_handler.go`
- Create: `migrations/XXXXXX_booking_request_mode.up.sql`

- [x] Add to Bathhouse model:
  - BookingMode (string: "instant" / "request", default "instant")
  - RequestTimeout (int, hours, default 24)
- [x] Add to Booking model:
  - HoldID (*uuid.UUID — reference to wallet hold)
  - New statuses: "pending_owner" (awaiting owner approval), "rejected" (owner rejected)
- [x] Migration: add `booking_mode VARCHAR(10) DEFAULT 'instant'`, `request_timeout INT DEFAULT 24` to bathhouses; add `hold_id UUID` to bookings
- [x] Modify CreateBooking in service:
  - If bathhouse.BookingMode == "instant": existing flow (confirm immediately)
  - If bathhouse.BookingMode == "request":
    1. Create booking with status "pending_owner"
    2. Create wallet hold for total amount (WalletService.Hold)
    3. Notify owner about new booking request
    4. Return booking with pending_owner status
- [x] New owner endpoints:
  - PATCH /api/v1/bookings/{id}/approve — owner approves request
    - Capture wallet hold, set status to confirmed
    - Notify client: "Ваша заявка одобрена!"
  - PATCH /api/v1/bookings/{id}/reject — owner rejects request
    - Release wallet hold, set status to rejected
    - Notify client: "К сожалению, ваша заявка отклонена"
    - Accept optional rejection reason
- [x] Cron: auto-reject requests older than RequestTimeout hours
  - Release hold, set status rejected
  - Notify both parties
- [x] Write tests for both instant and request flows
- [x] Run `go test ./... -v` — must pass

### Task 7.2: Check-in/Check-out & No-show (FR-066, FR-070)

**Files:**
- Modify: `internal/domain/booking.go`
- Modify: `internal/service/booking_service.go`
- Modify: `internal/handler/booking_handler.go`
- Create: `migrations/XXXXXX_booking_checkin.up.sql`

- [x] Add to Booking model: CheckedInAt (*time.Time), CheckedOutAt (*time.Time)
- [x] New booking status: "no_show"
- [x] Migration: add `checked_in_at TIMESTAMPTZ`, `checked_out_at TIMESTAMPTZ` to bookings
- [x] PATCH /api/v1/bookings/{id}/check-in — owner marks guest arrival (RequireAuth, owner/rep)
  - Validate: booking is confirmed, current time is within [start - 15min, start + 30min]
  - Set CheckedInAt = now
- [x] PATCH /api/v1/bookings/{id}/check-out — owner marks guest departure (RequireAuth, owner/rep)
  - Validate: booking is checked in
  - Set CheckedOutAt = now, status = completed
- [x] Owner reminder: push notification 5 min before session start ("Гость скоро прибудет!")
- [x] No-show detection cron (every 15 min):
  - Find confirmed bookings where start_time + 30min < now AND checked_in_at IS NULL
  - Mark as "no_show"
  - Charge client (no refund)
  - Notify owner: "Гость не прибыл, оплата сохранена"
- [x] No-show dispute: client can dispute within 2 hours
  - POST /api/v1/bookings/{id}/dispute-noshow — with optional GPS proof
  - Creates support ticket with high priority
- [x] Write tests
- [x] Run `go test ./... -v` — must pass

### Task 7.3: Booking Reminders (FR-064)

**Files:**
- Create: `internal/cron/booking_jobs.go`

- [x] Reminder cron job (every 15 min):
  - 24h before booking: send push + email
    - "Напоминаем: завтра в {time} бронирование в {bathhouse_name}. Адрес: {address}"
  - 2h before booking: send push only
    - "Через 2 часа — {bathhouse_name}. Не забудьте!"
  - Track sent reminders to avoid duplicates (Redis set `reminder_sent:{bookingID}:{type}`)
- [x] Include in reminder:
  - Bathhouse name, address
  - Date, time
  - Route link (Yandex Maps URL with coordinates)
- [x] Write tests
- [x] Run `go test ./... -v` — must pass

### Task 7.4: Session Extension (FR-065)

**Files:**
- Modify: `internal/service/booking_service.go`
- Modify: `internal/handler/booking_handler.go`

- [x] POST /api/v1/bookings/{id}/extend — extend booking by N hours (RequireAuth)
  - Request: { extra_hours: 1 or 2 }
  - Validate:
    - Booking is confirmed or checked_in
    - Next time slot is available (no conflicting bookings)
    - Extra hours within bathhouse operating hours
  - Calculate extension price at current hourly rate
  - Create payment for extension (wallet or card)
  - Update booking TimeTo
  - Notify both parties
- [x] Return updated booking with new end time and extension payment details
- [x] Write tests
- [x] Run `go test ./... -v` — must pass

### Task 7.5: Re-booking (FR-059)

**Files:**
- Modify: `internal/handler/booking_handler.go`
- Modify: `internal/service/booking_service.go`

- [x] GET /api/v1/bookings/{id}/rebook-data — return pre-filled booking data from past booking (RequireAuth)
  - Source: completed or cancelled booking owned by requesting user
  - Return: { bathhouse_id, duration_hours, time_from, time_to, guest_count, addons: [{addon_id, quantity}] }
  - Prices NOT included (recalculated at booking time with current rates)
- [x] Validate bathhouse is still active
- [x] Write tests
- [x] Run `go test ./... -v` — must pass

### Task 7.6: Buffer Time, Lead Time & Booking Settings (FR-073, FR-077, FR-078)

**Files:**
- Modify: `internal/domain/bathhouse.go`
- Modify: `internal/service/bathhouse_service.go`
- Modify: `internal/handler/bathhouse_handler.go`
- Create: `migrations/XXXXXX_booking_settings.up.sql`

- [x] Add to Bathhouse model:
  - BufferMinutes (int, default 30, range 0-120, step 15) — cleanup time between bookings
  - LeadTimeHours (int, default 2, range 0-48) — minimum hours before booking start
  - MaxAdvanceDays (int, default 90, range 7-365) — max days ahead for booking
  - MinDurationHours (float64, default 2.0, range 1.0-8.0, step 0.5) — minimum booking duration
- [x] Migration: add these columns to bathhouses with defaults
- [x] Modify GetAvailableSlots logic:
  - After each booked slot, add BufferMinutes as unavailable
  - Only show slots where start_time >= now + LeadTimeHours
  - Only show slots where date <= today + MaxAdvanceDays
- [x] Modify booking validation:
  - Reject if duration < MinDurationHours
  - Reject if start_time < now + LeadTimeHours
  - Reject if date > today + MaxAdvanceDays
  - Reject if conflicts with buffer time of adjacent bookings
- [x] Owner settings endpoints (PUT /api/v1/my/bathhouses/{id}) — add these fields to update
- [x] Write tests for all edge cases
- [x] Run `go test ./... -v` — must pass

### Task 7.7: Owner Cancellation Penalty (FR-069) & Response Rate (FR-079)

**Files:**
- Modify: `internal/service/booking_service.go`
- Modify: `internal/domain/bathhouse.go`
- Create: `migrations/XXXXXX_owner_metrics.up.sql`

- [x] Owner cancellation penalty:
  - When owner cancels a confirmed booking: credit 10% of booking price to client wallet as compensation
  - Track owner cancellations in 30-day rolling window
  - > 3 cancellations in 30 days: send warning notification to owner
  - > 5 cancellations in 30 days: auto-deactivate ALL owner's listings, notify admin
- [x] Response rate tracking (for request-based bookings):
  - Add to Bathhouse (or owner stats): ResponseRate (float64), AvgResponseTime (duration)
  - Calculate: approved_or_rejected_within_timeout / total_requests
  - Migration: add response_rate, avg_response_time_minutes columns
  - Recalculate daily via cron
  - ResponseRate < 50%: send warning notification
  - ResponseRate < 30% for 60+ days: force booking_mode to "instant" OR deactivate listing
- [x] Write tests
- [x] Run `go test ./... -v -race` — must pass
- [x] Run linter: `make lint`
