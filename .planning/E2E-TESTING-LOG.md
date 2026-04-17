# E2E Testing Session Log — 2026-04-18

## Setup
- Docker deps isolated: postgres:5433, redis:6380, minio:9010 (docker-compose.deps.yml)
- Backend: `./bin/bani-server serve` on :8081
- Frontend: `vite dev --port 5174`
- Test users: admin@test.local / client@test.local / owner@test.local (pw: Test1234!)

## Bugs Found & Status

### FIXED

#### Bug #1: Migration 113 FK references non-existent table `add_ons`
- File: `migrations/000113_promo_free_addon.up.sql`
- Root cause: `REFERENCES add_ons(id)` but actual table is `addons` (no underscore)
- Impact: Fresh installs fail at migration 113
- Fix: Changed to `REFERENCES addons(id)`

#### Bug #2: Seed-admin fails on fresh DB
- File: `cmd/server/seed.go`
- Root cause: `ON CONFLICT (email)` doesn't match partial unique index
- Impact: `make seed-admin` fails with SQLSTATE 42P10
- Fix: Added `WHERE email::text <> ''::text` predicate to ON CONFLICT

#### Bug #3: Recommendations 500 — SQL DISTINCT bug 1
- File: `internal/repository/postgres/recommendation.go:109 GetUserBookedBathhouses`
- Root cause: `SELECT DISTINCT x ... ORDER BY y` — PG requires ORDER BY cols in SELECT with DISTINCT
- Fix: Changed to `SELECT DISTINCT ON (bathhouse_id) ... ORDER BY bathhouse_id, created_at DESC`

#### Bug #4: Recommendations 500 — SQL DISTINCT bug 2
- File: `internal/repository/postgres/recommendation.go:186 GetSimilarUsers`
- Root cause: Same DISTINCT+ORDER BY issue
- Fix: Removed redundant DISTINCT (GROUP BY already ensures uniqueness)

#### Bug #5: 500 errors are unlogged (observability)
- File: `internal/handler/response.go:429`
- Fix: Added stderr log of error path+detail before writing 500

#### Bug #6: Search bathhouses 500 — SQL alias in ORDER BY CASE
- File: `internal/repository/postgres/bathhouse_repo.go:424`
- Root cause: PG doesn't resolve SELECT aliases (`is_promoted`, `promo_rank`) inside CASE in ORDER BY
- Fix: Inlined `promotionExists` EXISTS subquery directly in `promotedCap`. Top-3 cap (BRD FR-043) deferred — needs subquery wrapper

#### Bug #7: POST /my/recently-viewed returns 405
- Files: `internal/handler/saved_search_handler.go` (added handler), `internal/server/router.go` (added route)
- Root cause: Frontend calls POST to record view, but only GET was registered
- Fix: Added `RecordRecentlyViewed` handler + POST route binding. Service already had `RecordView` method.

### DEFERRED

#### Bug #8: --with-admin flag breaks startup
- Error: "mount goadmin engine: chi adapter SetApp: wrong parameter"
- Status: Deferred — backend works fine without `--with-admin`

#### Bug #9: GET /bookings/{id} missing
- Files: `internal/service/booking_service.go`, `internal/handler/booking_handler.go`, `internal/server/router.go`
- Root cause: frontend `ClientBookingDetail.tsx:61` calls `GET /bookings/{id}` but only `/bookings/{id}/cancel` etc. were registered
- Fix: added `BookingService.GetByID` (with role-based access), handler, and route

#### Bug #10: Frontend generated URLs have duplicated /api/v1/ prefix
- Files: `frontend/src/api/generated/{admin,share,bathhouses}.ts` (generated from swagger with basePath=/api/v1, but axios baseURL is also /api/v1)
- Root cause: orval + swagger basePath='/api/v1' + axios baseURL='/api/v1' → URL collision `/api/v1/api/v1/...`
- Fix: added post-process sed step in `frontend/package.json` generate:api script to strip `/api/v1/` prefix from generated URL strings. Fix is persistent through future regeneration.

### DEFERRED

#### Bug #11: --with-admin flag breaks startup
- Error: "mount goadmin engine: chi adapter SetApp: wrong parameter"
- Status: Deferred — backend works fine without `--with-admin`

### GAPS (contract mismatches)

#### Gap #1: /api/v1/promotions/banners missing
- Frontend: `frontend/src/pages/client/ClientHome.tsx:78`
- Backend: no route registered
- Impact: 404 silent (.catch swallows)
- TODO: add stub endpoint returning empty array

### INFRASTRUCTURE NOTES (for future testers)

- Admin user needs `admin_sub_role` set in DB + `two_fa_method != 'none'` — otherwise gets 403 on all /admin/* endpoints.
  ```sql
  UPDATE users SET admin_sub_role='super_admin', two_fa_method='totp', totp_secret='TESTSECRETXXXXXXXX' WHERE email='admin@test.local';
  ```
- Bathhouses need `working_hours` populated (JSONB array with day_of_week 0-6, open_time HH:MM, close_time HH:MM) for bookings to succeed.
- Test seed available in this session's SQL outputs; can be made into a proper fixture script.

## Flows Tested — Cycle 1

### Client Flow (full)
- [x] Login via email/password (note: antd Form requires React native-setValue + dispatchEvent)
- [x] Client home (recommendations, recently-viewed, profile completeness)
- [x] Search bathhouses (3 cities, 3 bathhouses seeded)
- [x] Bathhouse detail (/bathhouses/:slug)
- [x] Booking create (via API — UI wizard not clicked through)
- [x] Wallet dashboard (balance/transactions/holds)
- [x] Profile settings (profile-completeness, stats, region, notification-prefs, social-accounts)
- [x] Bookings list
- [x] GET booking by ID (after fix)

### Owner Flow (partial)
- [x] Login as owner
- [x] Home (owner dashboard, analytics)
- [x] Bookings mgmt (`/bookings` via bathhouse_id)
- [x] Calendar (pricing + external calendars + calendar-token)
- [x] Pricing rules + seasonal tariffs + price recommendation
- [x] Finance (wallet + transactions)
- [x] Reviews
- [ ] Listing wizard deep-click (not tested)
- [ ] Chat
- [ ] Photos upload
- [ ] Promotion auction
- [ ] Representatives
- [ ] Subscriptions

### Admin Flow
- [ ] Login as admin
- [ ] Moderation queue
- [ ] User management
- [ ] Analytics
- [ ] Finance dashboard
- [ ] Dispute management

## Tests to Add
- TODO: regression test for recommendation SQL fix (table-driven for empty user)
- TODO: regression test for seed.go ON CONFLICT fix
- TODO: test for /promotions/banners endpoint once added
