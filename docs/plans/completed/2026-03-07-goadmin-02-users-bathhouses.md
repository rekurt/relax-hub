# GoAdmin: User and Bathhouse Management

## Overview

GoAdmin table configurations for Users, Cities, Bathhouses, Representatives, Social Accounts, Telegram Links.

## Context

- Files involved: `internal/admin/tables/`, `internal/admin/engine.go`
- Related patterns: GoAdmin table generators, field types, filters, actions
- Dependencies: Plan 1 (infrastructure) must be completed first
- Domain models: User, City, Bathhouse, Representative, SocialAccount, TelegramLink

## Development Approach

- **Testing approach**: Regular (code first, then tests)
- Complete each task fully before moving to the next
- Each table file defines a GoAdmin table generator function
- **CRITICAL: every task MUST include new/updated tests**
- **CRITICAL: all tests must pass before starting next task**

## Implementation Steps

### Task 1: Users Table

**Files:**
- Create: `internal/admin/tables/users.go`

- [ ] Define GoAdmin table for users: all fields (id, email, name, phone, role, is_active, avatar_url, bio, city_id, created_at, updated_at)
- [ ] Role field as select dropdown: client/owner/representative/admin
- [ ] is_active as switch toggle
- [ ] city_id as select dropdown populated from cities table
- [ ] Password hash hidden from display, editable only via special form
- [ ] Filters: role, is_active, city_id, created_at range
- [ ] Actions: block/unblock (toggle is_active), change role
- [ ] Search by: email, name, phone

### Task 2: Cities Table

**Files:**
- Create: `internal/admin/tables/cities.go`

- [ ] Define GoAdmin table for cities: id, name, slug, latitude, longitude
- [ ] All fields editable
- [ ] Auto-generate slug from name on create
- [ ] Validation: unique slug, non-empty name

### Task 3: Bathhouses Table

**Files:**
- Create: `internal/admin/tables/bathhouses.go`

- [ ] Define GoAdmin table: all fields (id, owner_id, name, description, address, city_id, lat/lon, price_per_hour, min_duration, max_guests, amenity booleans, rating, review_count, images, status, api_key, created_at, updated_at)
- [ ] owner_id as select with user name display (filtered to role=owner)
- [ ] city_id as select dropdown
- [ ] status as select: active/inactive/pending/rejected
- [ ] price_per_hour displayed in rubles (value/100), stored in kopecks
- [ ] Amenity booleans as checkboxes
- [ ] Images as JSON array display
- [ ] Working hours as related sub-table
- [ ] Filters: status, city_id, owner_id, has_pool/sauna/etc, price range
- [ ] Actions: approve (set active), reject, deactivate
- [ ] Search by: name, address

### Task 4: Representatives Table

**Files:**
- Create: `internal/admin/tables/representatives.go`

- [ ] Define GoAdmin table: id, user_id, bathhouse_id, owner_id, created_at
- [ ] user_id, bathhouse_id, owner_id as selects with display names
- [ ] Filters: bathhouse_id, owner_id

### Task 5: Social Accounts Table

**Files:**
- Create: `internal/admin/tables/social_accounts.go`

- [ ] Define GoAdmin table: id, user_id, provider, provider_id, email, name, avatar_url, linked_at
- [ ] Read-only table (no create/edit, only view and delete)
- [ ] provider as select filter: vk/yandex/google
- [ ] user_id linked to users table

### Task 6: Telegram Links Table

**Files:**
- Create: `internal/admin/tables/telegram_links.go`

- [ ] Define GoAdmin table: id, user_id, telegram_id, telegram_username, linked_at
- [ ] Read-only view
- [ ] user_id linked to users table

### Task 7: Register Tables and Menu

**Files:**
- Modify: `internal/admin/engine.go`

- [ ] Register all 6 tables with GoAdmin engine
- [ ] Create menu structure: "Users Management" group with Users, Cities sub-items
- [ ] Create menu: "Bathhouses" group with Bathhouses, Representatives sub-items
- [ ] Create menu: "Integrations" group with Social Accounts, Telegram Links sub-items

### Task 8: Verify

- [ ] Build: go build ./...
- [ ] Start with --with-admin, verify all 6 tables render
- [ ] Test CRUD operations on users and cities
- [ ] Test filters and search
- [ ] Run tests and linter
