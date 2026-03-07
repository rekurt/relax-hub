# GoAdmin: Business Features (Subscriptions, Loyalty, Pricing, Analytics, Promos, Chat)

## Overview

GoAdmin tables for Subscriptions, Promotions, Pricing Rules, Loyalty, Promo Codes, Chat, Analytics, Recommendations.

## Context

- Files involved: `internal/admin/tables/`, `internal/admin/engine.go`
- Related patterns: GoAdmin table generators, kopecks-to-rubles conversion, colored badges
- Dependencies: Plans 1-3 must be completed first
- Domain models: Subscription, Promotion, PricingRule, LoyaltyAccount, LoyaltyTransaction, PromoCode, PromoUsage, Conversation, Message, BathhouseView, AnalyticsSnapshot, UserPreference, UserActivity

## Development Approach

- **Testing approach**: Regular (code first, then tests)
- Complete each task fully before moving to the next
- Many tables are read-only analytics views
- **CRITICAL: every task MUST include new/updated tests**
- **CRITICAL: all tests must pass before starting next task**

## Implementation Steps

### Task 1: Subscriptions Table

**Files:**
- Create: `internal/admin/tables/subscriptions.go`

- [ ] Define GoAdmin table: id, bathhouse_id, owner_id, plan, status, start_date, end_date, auto_renew, price_kopecks, created_at, updated_at
- [ ] plan as select: free/premium/promoted
- [ ] status as colored badges: active=green, expired=gray, cancelled=red
- [ ] price_kopecks displayed in rubles
- [ ] Filters: plan, status, owner_id, expiring soon
- [ ] Actions: activate, cancel, extend

### Task 2: Promotions Table

**Files:**
- Create: `internal/admin/tables/promotions.go`

- [ ] Define GoAdmin table: id, bathhouse_id, budget_kopecks, spent_kopecks, start_date, end_date, target_city_id, status, impression_count, click_count, created_at, updated_at
- [ ] Budget and spent in rubles display
- [ ] Status as colored badges: active/paused/exhausted/expired
- [ ] Calculated field: budget utilization percentage
- [ ] Filters: status, city_id, date range

### Task 3: Pricing Rules Table

**Files:**
- Create: `internal/admin/tables/pricing_rules.go`

- [ ] Define GoAdmin table: id, bathhouse_id, name, type, multiplier, days_of_week, time_from, time_to, date_from, date_to, priority, is_active, created_at
- [ ] type as select: weekday/weekend/holiday/time_range/season
- [ ] days_of_week as multi-select checkboxes (Mon-Sun)
- [ ] multiplier as decimal input with preview (e.g., 1.5 = "+50%")
- [ ] is_active as switch toggle
- [ ] Filters: bathhouse_id, type, is_active

### Task 4: Loyalty Accounts Table

**Files:**
- Create: `internal/admin/tables/loyalty_accounts.go`

- [ ] Define GoAdmin table: user_id, level, points, total_earned, total_spent, visit_count, updated_at, created_at
- [ ] level as colored badges: bronze/silver/gold/platinum
- [ ] Points and totals as formatted numbers
- [ ] Filters: level, points range, visit_count range
- [ ] Actions: manually adjust points, change level

### Task 5: Loyalty Transactions Table

**Files:**
- Create: `internal/admin/tables/loyalty_transactions.go`

- [ ] Define GoAdmin table: id, user_id, type, amount, booking_id, description, created_at
- [ ] type as colored badge: earn=green, spend=red, refund=blue
- [ ] Read-only table
- [ ] Filters: type, user_id, date range

### Task 6: Promo Codes Table

**Files:**
- Create: `internal/admin/tables/promo_codes.go`

- [ ] Define GoAdmin table: id, code, type, value, bathhouse_id, creator_id, max_uses, current_uses, min_amount, valid_from, valid_until, is_active, created_at
- [ ] type as select: percentage/fixed_amount/free_hour
- [ ] value context-dependent display (% or rubles)
- [ ] Usage progress bar: current_uses/max_uses
- [ ] is_active as switch toggle
- [ ] Filters: type, is_active, bathhouse_id, validity period
- [ ] Note: PromoCode repo not yet implemented - table will be read/write directly to DB

### Task 7: Promo Usages Table

**Files:**
- Create: `internal/admin/tables/promo_usages.go`

- [ ] Define GoAdmin table: id, promo_code_id, user_id, booking_id, discount_amount, used_at
- [ ] Read-only table
- [ ] discount_amount in rubles
- [ ] Filters: promo_code_id, user_id, date range

### Task 8: Conversations Table

**Files:**
- Create: `internal/admin/tables/conversations.go`

- [ ] Define GoAdmin table: id, bathhouse_id, client_id, booking_id, last_message_at, created_at
- [ ] Display names for bathhouse and client
- [ ] Filters: bathhouse_id, client_id
- [ ] Link to messages sub-table

### Task 9: Messages Table

**Files:**
- Create: `internal/admin/tables/messages.go`

- [ ] Define GoAdmin table: id, conversation_id, sender_id, text, is_read, read_at, created_at
- [ ] Read-only for admin (view only, no editing messages)
- [ ] Text truncated in list view, full in detail view
- [ ] Filters: conversation_id, is_read

### Task 10: Analytics Tables

**Files:**
- Create: `internal/admin/tables/bathhouse_views.go`
- Create: `internal/admin/tables/analytics_snapshots.go`

- [ ] Bathhouse Views: id, bathhouse_id, viewer_id, source, ip_hash, viewed_at - read-only
- [ ] Analytics Snapshots: bathhouse_id, date, views, unique_views, bookings, revenue, review_count, avg_rating - read-only
- [ ] Revenue in rubles display
- [ ] Filters: bathhouse_id, date range, source

### Task 11: User Preferences and Activities

**Files:**
- Create: `internal/admin/tables/user_preferences.go`
- Create: `internal/admin/tables/user_activities.go`

- [ ] User Preferences: user_id, preferred_city_id, price_range_min/max, prefer_* booleans - read-only
- [ ] User Activities: id, user_id, bathhouse_id, type (view/booking/favorite), created_at - read-only
- [ ] Filters: user_id, type, date range

### Task 12: Register Tables and Menu

**Files:**
- Modify: `internal/admin/engine.go`

- [ ] Register all tables from this plan
- [ ] Menu: "Subscriptions" group with Subscriptions, Promotions
- [ ] Menu: "Pricing" group with Pricing Rules, Promo Codes, Promo Usages
- [ ] Menu: "Loyalty" group with Loyalty Accounts, Loyalty Transactions
- [ ] Menu: "Chat" group with Conversations, Messages
- [ ] Menu: "Analytics" group with Views, Snapshots, User Activities, User Preferences

### Task 13: Verify

- [ ] Build, start with --with-admin
- [ ] Verify all tables render with correct field types
- [ ] Test subscription and loyalty management actions
- [ ] Run tests and linter
