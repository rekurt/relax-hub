# Codebase Concerns

**Analysis Date:** 2026-04-17

## Tech Debt

**Missing Test Coverage:**
- Issue: Three service implementations lack test files
- Files: `internal/service/admin_notification_service.go`, `internal/service/device_token_service.go`
- Impact: Critical notification and device token flows are not unit tested; regressions in admin alerts and push notification delivery are undetectable
- Fix approach: Add comprehensive test suites with table-driven tests for both happy paths and error cases (follow pattern in `internal/service/booking_service_test.go` — 4600+ lines for similar service)

**Distributed Cron Lock Overhead:**
- Issue: Redis-based distributed locking for cron jobs uses SETNX with 10-minute TTL and manual release on every job
- Files: `internal/cron/scheduler.go` lines 65-79
- Impact: If a cron job crashes/panics, lock stays held for 10 minutes; multiple instances on slow jobs may experience lock contention; no reentrant lock support for dependent jobs
- Fix approach: Implement lease-based locking with automatic refresh during job execution, or use built-in Postgres advisory locks (consistent with booking creation serialization)

**N+1 Query Potential in Holiday/Seasonal Pricing:**
- Issue: Price calculation fetches bathhouse → city → region per booking creation call; cityRepo and bhRepo calls happen inside a loop during CalculateFullPrice
- Files: `internal/service/pricing_service.go` lines 142-149
- Impact: Under high booking volume, regional holiday checking multiplies DB queries; should cache region per bathhouse
- Fix approach: Pre-fetch bathhouse + city data before pricing loop; cache city region in bathhouse context struct

**Large Handler File - Booking Handler:**
- Issue: Main booking handler is 1068 lines with overlapping concerns
- Files: `internal/handler/booking_handler.go`
- Impact: Hard to navigate; increased risk of cross-cutting concerns (validation, authorization, response formatting) being duplicated or inconsistent
- Fix approach: Extract request/response types to dedicated file; move complex validation logic to service layer; split into sub-handlers (create, list, detail, modify, extend)

**Unbounded Slot Availability Check:**
- Issue: CheckAvailability queries count overlapping bookings without limit; on bathhouse with thousands of concurrent bookings, the COUNT(*) could be slow
- Files: `internal/repository/postgres/booking_repo.go` lines 235-240
- Impact: Availability checks slow down during high concurrency; no index optimization for range queries
- Fix approach: Ensure compound index on `(bathhouse_id, status, start_time, end_time)` exists in migrations; consider LIMIT 1 approach instead of COUNT

## Known Bugs

**Wallet Concurrent Update Retry Missing User Guidance:**
- Symptom: Client receives `ErrWalletConcurrentUpdate` on top-up/refund under race conditions; error message is "wallet was modified concurrently, please retry" but no automatic retry is implemented
- Files: `internal/domain/errors.go` line 55, `internal/service/wallet_service.go` line 111-112, `internal/handler/response.go` line 213
- Trigger: Two simultaneous wallet operations (e.g., booking payment + bonus expiry transaction) on same wallet in < 100ms window
- Workaround: Frontend should implement exponential backoff retry (2s, 4s, 8s) for 409 responses with `ErrWalletConcurrentUpdate` code; currently no retry logic in `frontend/src/api/axios-instance.ts`
- Fix approach: Implement client-side retry in axios instance for wallet/payment endpoints; add structured retry guidance in API documentation

**Admin Notification Service Missing Implementation Tests:**
- Symptom: Admin notifications for SLA violations and reconciliation mismatches may fail silently
- Files: `internal/service/admin_notification_service.go` (not tested)
- Trigger: Moderation queue > 24h wait time, or reconciliation float drift > threshold
- Workaround: None; notifications fire but correctness is unverified
- Fix approach: Add comprehensive tests covering all 8 notification types (antifraud_flag, sla_violation, reconciliation_mismatch, float_drift, ticket_escalation, dispute_opened, kyc_pending, system)

**Token Expiry Validation Gap:**
- Symptom: Partial JWT tokens issued during 2FA flow may not properly validate session expiry; malformed or expired tokens could pass validation in OptionalAuth middleware
- Files: `internal/middleware/auth.go` (no validation), `internal/service/auth_service.go` (JWT parsing logic needs review)
- Trigger: Client submits partial token from incomplete 2FA flow with >30 min delay
- Workaround: Sessions expire after 30 days; partial tokens for 2FA have no documented TTL
- Fix approach: Add explicit TTL validation to JWT parsing; document and enforce 5-minute TTL for partial 2FA tokens in claims

## Security Considerations

**localStorage Token Storage Risk:**
- Risk: JWT tokens stored in localStorage are vulnerable to XSS; no HttpOnly flag on cookies
- Files: `frontend/src/stores/auth.ts` line 15 (getItem/setItem)
- Current mitigation: Frontend does NOT use dangerouslySetInnerHTML; no eval; content security headers are not visible in provided code
- Recommendations: 
  1. Migrate token storage to HttpOnly cookies with SameSite=Strict (requires backend cookie-based sessions)
  2. If localStorage required, implement minimal CSP and sanitize all user-generated content renders
  3. Add subresource integrity (SRI) to external CDN scripts

**File Upload Path Traversal Potential (Low Risk):**
- Risk: Widget script/CSS loading via filepath.Join without canonicalization; client uploads via FormFile without explicit sanitization
- Files: `internal/handler/widget.go` lines 318-352 (filepath.Join)
- Current mitigation: Paths are hardcoded to `widget/dist/` directory; no user-controlled path component
- Recommendations: Explicitly verify file paths are within `widget/dist/` using `filepath.Clean` + prefix check before `os.ReadFile`; add test case for symlink attack

**Payment Provider Secret Exposure Risk:**
- Risk: YooKassa/bePaid secret keys passed to payment provider constructor; if error paths leak provider state, secrets could be logged
- Files: `internal/payment/provider_factory.go`, `internal/service/payment_service.go` line 94
- Current mitigation: Logger is configured with structured logging (zap); provider passed to service; no visible secrets logging
- Recommendations: Audit all error handling in payment service for stack traces or error details that could leak provider config; add secret redaction filter to logger output

**Device Token Forgery (Untested):**
- Risk: Device tokens (FCM/APNS) are user-supplied strings without validation; a client could register invalid/spoofed tokens
- Files: `internal/service/device_token_service.go` (no tests), handler likely at `internal/handler/device_token_handler.go`
- Current mitigation: Firebase validates token format before sending
- Recommendations: Add token format validation pre-store; implement token verification ping on registration; rate-limit token registration per user/session

## Performance Bottlenecks

**Admin Moderation Queue SQL Without Proper Indexes:**
- Problem: Moderation page queries pending reviews across bathhouses; no mention of compound index on `(status, created_at)` or per-bathhouse partition
- Files: `internal/admin/pages/moderation.go` (SQL generated dynamically based on filters)
- Cause: Full table scan on large reviews table when filtering by status + date range
- Improvement path:
  1. Ensure indexes: `CREATE INDEX idx_reviews_status_created ON reviews(status, created_at DESC)` and `CREATE INDEX idx_reviews_bathhouse_status ON reviews(bathhouse_id, status, created_at DESC)`
  2. Add pagination to moderation queue; currently may fetch all pending items
  3. Cache SLA metrics (recalculate hourly via cron) instead of on every page load

**Escrow Auto-Release Cron Every Hour (Scale Risk):**
- Problem: `internal/cron/escrow_jobs.go` runs hourly to find all matured escrows; no pagination or batch processing limit
- Cause: Linear scan of all escrows table; as escrow volume grows, job runtime increases
- Improvement path:
  1. Add `released_at IS NULL AND claim_deadline < NOW()` index
  2. Process in batches (100 escrows per batch) with sleep between batches
  3. Consider event-based release instead of scheduled scan (publish event on check-out, consume asynchronously)

**WebSocket Hub Memory Under High Connection Load:**
- Problem: `internal/notification/hub.go` uses in-memory map of user → clients; no connection limit or memory pressure handling
- Cause: Each client keeps a `Send` channel in memory; millions of concurrent clients could cause heap exhaustion
- Improvement path:
  1. Implement connection limits per user (e.g., max 5 simultaneous WS connections)
  2. Add memory pressure monitoring; disconnect idle clients if heap usage > 80%
  3. Consider Redis pub/sub for multi-instance broadcast instead of in-memory hub

**Session Cleanup Cron (Daily at Midnight):**
- Problem: Cron deletes expired sessions once per day; active but expired sessions remain in DB for up to 24 hours
- Cause: Lazy cleanup approach trades performance for stale session rows
- Improvement path:
  1. Run cleanup every 4 hours instead of daily
  2. Add background job that expires sessions in-app on logout (not just cron)
  3. Add session_ttl index to sessions table for faster cleanup queries

## Fragile Areas

**Payment Hold Release Race Condition:**
- Files: `internal/service/payment_service.go`, `internal/service/escrow_service.go`
- Why fragile: 
  1. Card hold captured on booking confirmation
  2. Escrow released 48h after check-out
  3. If payment webhook arrives after escrow released, fund movement could be inconsistent
  4. No explicit locking between payment status update and escrow state transition
- Safe modification: Always check booking status and payment status before releasing hold/escrow; use transaction-level locking for payment + escrow updates
- Test coverage: `internal/service/payment_service_test.go` has 1725 lines; check for test_hold_release_after_escrow_mature scenario

**Booking Modification Request Timeout Handling:**
- Files: `internal/domain/booking_modification.go`, `internal/cron/booking_jobs.go` (timeout cron)
- Why fragile:
  1. Modification request expires 24h after creation
  2. Cron auto-rejects expired requests
  3. If cron fails to run (e.g., lock held), request stays pending indefinitely
  4. Owner and client see inconsistent state (one thinks request expired, other sees pending)
- Safe modification: Add explicit validation in handlers to check expiry; never rely solely on cron cleanup; add is_expired flag to DB
- Test coverage: Check `internal/service/booking_modification_service_test.go` for timeout+cron failure scenarios

**Dynamic Pricing Rules Priority Tie-Breaking:**
- Files: `internal/service/pricing_service.go` lines 114-121
- Why fragile:
  1. Multiple rules with same priority are sorted by ID string (non-deterministic if IDs change)
  2. Pricing becomes unpredictable if rule ordering is ambiguous
  3. Rules created by owner via UI have no explicit priority UI; all new rules default to priority 0
- Safe modification: 
  1. Enforce unique priority per bathhouse; disallow multiple rules with same priority
  2. Add explicit priority UI slider in owner dashboard
  3. Add test with >5 equal-priority rules to ensure consistent ordering
- Test coverage: `internal/service/pricing_service_test.go` — check for "same_priority_rules_deterministic_order" test

**User Region Switch Blocking State Machine:**
- Files: `internal/service/region_service.go`
- Why fragile:
  1. Region switch blocked if wallet balance > 0
  2. But if client has pending booking refund, refund might land in old wallet after switch
  3. No atomic state transition between wallet freeze and switch
- Safe modification: 
  1. Freeze wallet before switch begins
  2. Ensure all pending payments/refunds are settled before allowing switch
  3. Add explicit status field to user: `region_switch_in_progress` with timeout
- Test coverage: Add test scenario: "client_with_pending_refund_cannot_switch_regions"

## Scaling Limits

**Redis In-Memory Limit (Recent Searches Cache):**
- Current capacity: No explicit limit on Redis memory; searches cached indefinitely per user session
- Limit: Redis default 2-4GB; with 100k daily users, searches + sessions cache could approach limit
- Scaling path:
  1. Add Redis memory eviction policy: `ALLKEYS-LRU` (default `noeviction` can cause OOM crashes)
  2. Implement Redis streams for pagination instead of sorted sets
  3. Consider time-series DB (InfluxDB) for aggregated search metrics instead of raw Redis cache

**PostgreSQL Connection Pool (Default 20 Max):**
- Current capacity: 20 active connections per app instance (configurable via `BANI_DATABASE_MAX_CONNS`)
- Limit: With 3-5 concurrent HTTP requests per connection + long-running cron jobs, effective capacity ~4-5 RPS on busy periods
- Scaling path:
  1. Increase to 50-100 max connections (monitor CPU/memory trade-off)
  2. Implement query timeout (currently no visible timeout config)
  3. Add PgBouncer or similar connection pooler between app and DB for connection multiplexing
  4. Benchmark: test at 100 RPS with current settings to find actual breaking point

**Escrow Table Growth (No Archive Policy):**
- Current capacity: All escrows kept in active tables indefinitely
- Limit: After 6 months with 10k daily bookings, escrow table ~2M rows; queries slow without proper indexes
- Scaling path:
  1. Add retention policy: archive released escrows older than 90 days to separate `escrows_archive` table
  2. Implement table partitioning by month (created_at)
  3. Add automated VACUUM/ANALYZE after bulk deletes

**Session Table Bloat:**
- Current capacity: Sessions expire after 30 days but only deleted by daily cron; table grows by ~100k rows/day with 100k users
- Limit: After 6 months, ~18M rows; cron cleanup takes hours
- Scaling path:
  1. Run cleanup every 4 hours instead of daily
  2. Add table partitioning by created_at (monthly)
  3. Consider Redis for short-lived sessions (< 30 days) instead of DB

## Dependencies at Risk

**go-blurhash (Unused or Minimal Usage):**
- Risk: Dependency for image blur-hash generation; if package abandoned, security patches may lag
- Impact: Media upload flow fails if blurhash generation fails; blocks new feature on missing dep update
- Migration plan: If needed, switch to open-source blurhash implementations or compute blur-hash client-side; currently seems well-maintained but not critical to core flow

**PostgreSQL + PostGIS Version Lock:**
- Risk: Migrations assume PostGIS 3.2+; no downgrade script if upgrading to PG 16 removes functions
- Impact: Schema migration fails on old PG versions; geo-queries break on incompatible PostGIS
- Migration plan: Add explicit version checks in `000001_init.up.sql`; document minimum PG 13, PostGIS 3.1; test on multiple versions

## Missing Critical Features

**Explicit Account Deletion Verification:**
- Problem: Account deletion grace period (30 days) lacks email verification or confirmation code
- Blocks: Mistaken deletion recovery is hard; no way for user to cancel deletion if email unattended for 27 days
- Fix approach: 
  1. Require confirmation code sent to email to initiate deletion
  2. Add deletion_initiated_token with 48h TTL
  3. Send reminder emails at day 14 and day 27 with cancellation link

**Booking Cancellation Reason Tracking:**
- Problem: Cancellations logged but reason (client vs owner) is stored in nullable text field, not structured enum
- Blocks: Analytics cannot easily segment cancellations by reason; owner incentive analysis is manual
- Fix approach: Add `cancellation_reason_code` enum field (user_initiated, owner_initiated, system_initiated, timeout); backfill existing data

**Password Reset Token Reuse Prevention:**
- Problem: No mechanism to invalidate old password reset tokens after successful reset
- Blocks: User resets password, then old token is used by attacker if intercepted in email
- Fix approach: 
  1. Store reset token + used_at timestamp in DB
  2. Invalidate all tokens for user after successful reset
  3. Add rate limiting on token validation attempts (3 per token, then expire)

## Test Coverage Gaps

**Untested Admin Notification Dispatcher:**
- What's not tested: All notification type routing, delivery to multiple admin sub-roles, daily digest cron
- Files: `internal/service/admin_notification_service.go` (no _test.go file)
- Risk: Admin alerts for reconciliation mismatches, SLA violations may silently fail to deliver
- Priority: High (admin rely on these alerts to catch issues)
- Fix approach: Add 20+ test cases covering each notification type, role filtering, email batching

**Device Token Lifecycle Not Tested:**
- What's not tested: Token registration, deletion, Firebase push delivery, invalid token cleanup
- Files: `internal/service/device_token_service.go` (no _test.go file)
- Risk: Push notifications fail silently for certain users; stale tokens accumulate
- Priority: Medium (feature works in prod but regressions undetected)
- Fix approach: Add tests for registration, duplicate token handling, token expiry, push delivery mocks

**Escrow Dispute Interaction Edge Cases:**
- What's not tested: Escrow released while dispute open, partial refund on disputed amount, hold release timing
- Files: `internal/service/escrow_service_test.go` — likely missing escrow+dispute scenarios
- Risk: Money trapped in escrow if dispute blocks release, or released early if dispute resolved after claim deadline
- Priority: Critical (financial impact)
- Fix approach: Add parametrized tests for all combinations: (escrow_released, dispute_open/resolved/appealed) × (refund_full/partial)

**Pricing Rule Conflict Resolution:**
- What's not tested: Multiple overlapping rules with different priorities, seasonal tariff + dynamic rule combination, holiday + seasonal conflict
- Files: `internal/service/pricing_service_test.go` — check for "multiple_rules_same_hour_deterministic" test
- Risk: Pricing inconsistent or unpredictable in edge cases; owner may dispute charges
- Priority: High (pricing is revenue-critical)
- Fix approach: Add 30+ test cases covering all rule combinations; add property-based test that prices are always >= base_price

**Booking Modification + Extension Simultaneous Requests:**
- What's not tested: Client requests modification AND extension on same booking in parallel
- Files: `internal/service/booking_service_test.go` — look for "concurrent_modification_and_extension" test
- Risk: Booking ends up in inconsistent state; double-charge or double-refund possible
- Priority: High (rare but financially critical)
- Fix approach: Add test with two parallel requests; ensure atomic state transition in service

---

*Concerns audit: 2026-04-17*
