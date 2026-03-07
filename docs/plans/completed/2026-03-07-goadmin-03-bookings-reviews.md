# GoAdmin: Bookings, Reviews, Moderation

## Overview

GoAdmin tables for Bookings, Reviews (with moderation workflow), Favorites, Notifications, Notification Preferences.

## Context

- Files involved: `internal/admin/tables/`, `internal/admin/engine.go`
- Related patterns: GoAdmin table generators, colored badges, batch actions
- Dependencies: Plans 1-2 must be completed first
- Domain models: Booking, Review, Favorite, Notification, NotificationPreference

## Development Approach

- **Testing approach**: Regular (code first, then tests)
- Complete each task fully before moving to the next
- Review moderation is the key workflow in this stage
- **CRITICAL: every task MUST include new/updated tests**
- **CRITICAL: all tests must pass before starting next task**

## Implementation Steps

### Task 1: Bookings Table

**Files:**
- Create: `internal/admin/tables/bookings.go`

- [ ] Define GoAdmin table: id, user_id, bathhouse_id, start_time, end_time, guest_count, total_price, points_spent, status, comment, created_at, updated_at
- [ ] user_id and bathhouse_id as selects with display names
- [ ] total_price in rubles (value/100)
- [ ] status as colored badges: pending=yellow, confirmed=blue, cancelled=red, rejected=gray, completed=green
- [ ] Filters: status, bathhouse_id, user_id, date range, price range
- [ ] Actions: confirm, reject, cancel, complete (status transitions)
- [ ] Read-only for most fields except status and comment

### Task 2: Reviews Table with Moderation

**Files:**
- Create: `internal/admin/tables/reviews.go`

- [ ] Define GoAdmin table: id, user_id, bathhouse_id, booking_id, rating, text, status, rejection_reasons, owner_response, owner_response_at, images, created_at, updated_at
- [ ] Rating as star display (1-5)
- [ ] status as colored badges: pending=yellow, approved=green, rejected=red, hidden=gray
- [ ] Filters: status, rating, bathhouse_id, date range
- [ ] Actions: approve, reject (with rejection reasons multi-select), hide
- [ ] Batch actions: batch-approve, batch-reject selected reviews
- [ ] Quick filter button: "Pending reviews" (pre-filtered to status=pending)
- [ ] Pending count badge on menu item

### Task 3: Favorites Table

**Files:**
- Create: `internal/admin/tables/favorites.go`

- [ ] Define GoAdmin table: id, user_id, bathhouse_id, created_at
- [ ] Read-only analytics view
- [ ] Display user name and bathhouse name
- [ ] Filters: bathhouse_id, user_id

### Task 4: Notifications Table

**Files:**
- Create: `internal/admin/tables/notifications.go`

- [ ] Define GoAdmin table: id, user_id, type, title, body, data, is_read, read_at, created_at
- [ ] type as select filter with all notification types
- [ ] is_read as boolean filter
- [ ] data as JSON display
- [ ] Read-only table (admin can only view and delete notifications)
- [ ] Action: bulk delete old notifications

### Task 5: Notification Preferences Table

**Files:**
- Create: `internal/admin/tables/notification_preferences.go`

- [ ] Define GoAdmin table: user_id, in_app, email, push, telegram, booking_events, review_events, promo_events, reminders
- [ ] All boolean fields as switch toggles
- [ ] Linked to users table

### Task 6: Register Tables and Menu

**Files:**
- Modify: `internal/admin/engine.go`

- [ ] Register all 5 tables
- [ ] Menu: "Bookings" group with Bookings sub-item
- [ ] Menu: "Reviews & Moderation" group with Reviews sub-item (with pending count badge)
- [ ] Menu: "Engagement" group with Favorites, Notifications, Notification Preferences sub-items

### Task 7: Verify

- [ ] Build, start with --with-admin
- [ ] Test booking status transitions
- [ ] Test review moderation workflow (approve/reject single and batch)
- [ ] Verify filters and search
- [ ] Run tests and linter
