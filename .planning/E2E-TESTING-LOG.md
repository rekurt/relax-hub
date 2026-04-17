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

### GAPS (contract mismatches)

#### Gap #1: /api/v1/promotions/banners missing
- Frontend: `frontend/src/pages/client/ClientHome.tsx:78`
- Backend: no route registered
- Impact: 404 silent (.catch swallows)
- TODO: add stub endpoint returning empty array

## Flows Tested — Cycle 1

### Client Flow
- [x] Login via email/password
- [x] Client home page renders
- [ ] Search bathhouses
- [ ] Bathhouse detail
- [ ] Booking flow
- [ ] Payment
- [ ] Review submission
- [ ] Profile settings
- [ ] Wallet dashboard
- [ ] Certificates
- [ ] Loyalty program
- [ ] Support tickets

### Owner Flow
- [ ] Login as owner
- [ ] KYC onboarding
- [ ] Listing wizard (7 steps)
- [ ] Pricing rules
- [ ] Calendar management
- [ ] Bookings management
- [ ] Chat with clients
- [ ] Payouts

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
