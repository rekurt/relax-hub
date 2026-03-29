# Plan: Full BRD Compliance - RelaxHub v6

## Overview

Comprehensive gap analysis and implementation plan to bring the RelaxHub codebase into full compliance with BRD v6. Covers 30+ missing or incomplete features across payments, location, calendar, CRM, admin, integrations, SEO, and support automation.

## Context

- BRD: docs/BRD_RelaxHub_v6.md (158 functional requirements)
- Architecture: handler -> service -> repository (Go, chi, pgx, fx DI)
- Frontend: React + Ant Design + TanStack Query + orval
- Current coverage: ~85% of BRD requirements implemented
- Remaining gaps: ~30 features/enhancements across 10 categories
- Files involved: internal/domain/, internal/service/, internal/handler/, internal/repository/postgres/, internal/payment/, internal/notification/, internal/geo/, internal/pms/, internal/seo/, internal/admin/, frontend/src/
- Related patterns: handler→service→repository, fx DI modules, domain error mapping
- Dependencies: github.com/xuri/excelize/v2 (xlsx), Yandex Maps API (isochrone), bePaid API, Yclients API, Restoplace API

## Development Approach

- **Testing approach**: Regular (code first, then tests)
- Complete each task fully before moving to the next
- Tasks ordered by dependency: backend domain/repo first, then service, then handler, then frontend
- **CRITICAL: every task MUST include new/updated tests**
- **CRITICAL: all tests must pass before starting next task**

---

## Implementation Steps

### Task 1: Belarus Payment Infrastructure (FR-092, FR-095)

**Why:** BRD requires dual-region payment: YooKassa for RU, bePaid/Assist for BY. Currently only YooKassa exists.

**Files:**
- Create: `internal/payment/bepaid_provider.go`
- Create: `internal/payment/provider_factory.go`
- Modify: `internal/domain/payment.go` (add Belkart, ERIP, MIR methods)
- Modify: `internal/service/booking_service.go` (route by region)
- Modify: `config/config.go` (add bePaid credentials)

- [x] Add payment methods: MIR, Belkart, ERIP to domain
- [x] Implement PaymentProvider interface for bePaid (create, capture, cancel, refund, webhook)
- [x] Create provider factory that selects provider by user region (RU -> YooKassa, BY -> bePaid)
- [x] Add bePaid webhook handler endpoint
- [x] Add bePaid config keys (BANI_PAYMENT_BEPAID_*)
- [x] Update fiscalization to support BY provider placeholder
- [x] Write tests for bePaid provider and provider factory
- [x] Run project test suite - must pass before task 2

### Task 2: Apple Pay & Google Pay (FR-057, FR-092)

**Why:** BRD requires Apple Pay and Google Pay as payment methods for RU region.

**Files:**
- Modify: `internal/domain/payment.go` (add ApplePay, GooglePay methods)
- Modify: `internal/payment/yookassa_provider.go` (token-based payments)
- Modify: `internal/handler/payment_handler.go` (new endpoints)
- Modify: `frontend/src/pages/client/BookingCreate.tsx`

- [x] Add ApplePay and GooglePay to PaymentMethod enum
- [x] Implement YooKassa token-based payment for Apple/Google Pay
- [x] Add frontend payment button components (Apple Pay JS, Google Pay API)
- [x] Integrate into booking checkout flow
- [x] Write tests for new payment methods
- [x] Run project test suite - must pass before task 3

### Task 3: Saved Card Tokens (FR-004)

**Why:** BRD requires clients to save cards for quick repeat payments via provider tokenization.

**Files:**
- Create: `internal/domain/saved_card.go`
- Create: `internal/repository/postgres/saved_card_repo.go`
- Create: `internal/service/saved_card_service.go`
- Create: `internal/handler/saved_card_handler.go`
- Create: `migrations/XXXX_saved_cards.up.sql`
- Modify: `frontend/src/pages/client/BookingCreate.tsx`

- [x] Create saved_cards table (id, user_id, provider_token, last4, brand, expires, is_default)
- [x] Implement repository and service (list, add, delete, set default)
- [x] Handler: GET/POST/DELETE /api/v1/my/cards
- [x] Save card token after successful payment (opt-in checkbox)
- [x] Show saved cards on checkout, allow one-click payment
- [x] Write tests for saved card CRUD
- [x] Run project test suite - must pass before task 4

### Task 4: Bank Statement Reconciliation (FR-128)

**Why:** BRD requires importing bank statements and auto-matching against internal transactions.

**Files:**
- Create: `internal/service/bank_reconciliation_service.go`
- Create: `internal/handler/bank_reconciliation_handler.go`
- Create: `migrations/XXXX_bank_reconciliation.up.sql`
- Modify: `internal/cron/scheduler.go` (optional daily import cron)

- [x] Create bank_statement_entries table (id, date, amount, description, counterparty, matched_tx_id, status)
- [x] Implement CSV/1C format parser for bank statements
- [x] Auto-matching algorithm: by amount + date + reference number
- [x] Admin endpoint: POST /api/v1/admin/finance/bank-statement (upload)
- [x] Admin endpoint: GET /api/v1/admin/finance/reconciliation (unmatched entries queue)
- [x] Admin endpoint: PUT /api/v1/admin/finance/reconciliation/{id}/match (manual match)
- [x] Write tests for parser and matching
- [x] Run project test suite - must pass before task 5

### Task 5: Admin Wallet Management (FR admin-panel)

**Why:** BRD requires admin manual credit/debit/freeze of user wallets.

**Files:**
- Modify: `internal/handler/wallet_handler.go` (admin endpoints)
- Modify: `internal/service/wallet_service.go` (admin operations)
- Create: `frontend/src/pages/admin/WalletManagement.tsx`

- [x] Add admin endpoint: POST /api/v1/admin/wallets/{id}/credit (amount, reason)
- [x] Add admin endpoint: POST /api/v1/admin/wallets/{id}/debit (amount, reason)
- [x] Add admin endpoint: POST /api/v1/admin/wallets/{id}/freeze
- [x] Add admin endpoint: POST /api/v1/admin/wallets/{id}/unfreeze
- [x] All operations logged to audit_log with admin_id and reason
- [x] Create admin wallet management page with search and actions
- [x] Write tests for admin wallet operations
- [x] Run project test suite - must pass before task 6

### Task 6: Admin Booking Management (FR admin-panel)

**Why:** BRD requires admin manual cancel, refund, status change for bookings.

**Files:**
- Modify: `internal/handler/booking_handler.go` (admin endpoints)
- Modify: `internal/service/booking_service.go` (admin override)
- Create: `frontend/src/pages/admin/BookingManagement.tsx`

- [x] Add admin endpoint: POST /api/v1/admin/bookings/{id}/cancel (reason)
- [x] Add admin endpoint: POST /api/v1/admin/bookings/{id}/change-status (status, reason)
- [x] Add admin endpoint: GET /api/v1/admin/bookings (search with filters)
- [x] Ensure manual refund endpoint exists (POST /api/v1/admin/bookings/{id}/refund already exists - verify)
- [x] All operations logged to audit_log
- [x] Create admin booking management page
- [x] Write tests for admin booking operations
- [x] Run project test suite - must pass before task 7

### Task 7: Admin Sub-Roles / RBAC (FR admin-panel)

**Why:** BRD requires multiple admin sub-roles: admin, moderator, support, finance. Currently only single "admin" role.

**Files:**
- Modify: `internal/domain/user.go` (add admin sub-roles or separate role field)
- Create: `internal/domain/admin_permission.go`
- Modify: `internal/middleware/auth.go` (permission-based checks)
- Create: `migrations/XXXX_admin_roles.up.sql`
- Create: `frontend/src/pages/admin/RoleManagement.tsx`

- [x] Add admin sub-roles: super_admin, moderator, support_l1, support_l2, support_l3, finance
- [x] Create admin_permissions table (role, permission, resource)
- [x] Define permission matrix: who can access which admin endpoints
- [x] Add middleware RequireAdminPermission(permission) alongside RequireRole
- [x] Enforce mandatory 2FA for all admin sub-roles
- [x] Create admin role management page (assign roles, view permissions)
- [x] Write tests for permission checks
- [x] Run project test suite - must pass before task 8

### Task 8: Admin Mass Operations (FR admin-panel)

**Why:** BRD requires batch operations up to 1000 records (approve/reject listings, block users, credit bonuses).

**Files:**
- Modify: `internal/handler/admin_handler.go` (batch endpoints)
- Modify: `internal/service/moderation_service.go` (batch logic)
- Modify: `frontend/src/pages/admin/BathhouseModeration.tsx`
- Modify: `frontend/src/pages/admin/UserManagement.tsx`

- [x] Add batch endpoint: POST /api/v1/admin/listings/batch (action: approve/reject, ids: [up to 1000])
- [x] Add batch endpoint: POST /api/v1/admin/users/batch (action: block/unblock, ids: [up to 1000])
- [x] Add batch endpoint: POST /api/v1/admin/wallets/batch-credit (ids, amount, reason)
- [x] Process in DB transactions with chunking (100 per batch)
- [x] Return results: succeeded[], failed[] with reasons
- [x] Update frontend moderation/user pages with bulk selection and actions
- [x] Write tests for batch operations
- [x] Run project test suite - must pass before task 9

### Task 9: Admin Notification Center (FR admin-panel)

**Why:** BRD requires critical alerts per admin role (antifraud, SLA violations, financial discrepancies), email digest.

**Files:**
- Create: `internal/domain/admin_notification.go`
- Create: `internal/repository/postgres/admin_notification_repo.go`
- Create: `internal/service/admin_notification_service.go`
- Create: `internal/handler/admin_notification_handler.go`
- Create: `frontend/src/pages/admin/AdminNotificationCenter.tsx`
- Modify: `internal/cron/scheduler.go` (daily digest cron)

- [x] Create admin_notifications table (id, role, severity, type, title, body, read, created_at)
- [x] Emit admin notifications from: antifraud flags, SLA violations, reconciliation discrepancies, float drift
- [x] Admin endpoint: GET /api/v1/admin/notifications (filtered by role)
- [x] Admin endpoint: PUT /api/v1/admin/notifications/{id}/read
- [x] Daily email digest cron: aggregate unread critical alerts and send per admin role
- [x] Create notification center page with severity filtering
- [x] Write tests for notification emission and digest
- [x] Run project test suite - must pass before task 10

### Task 10: Moderation Dashboard with SLA (FR-146)

**Why:** BRD requires moderation metrics: queue size, avg wait time, SLA compliance, per-moderator load.

**Files:**
- Modify: `internal/admin/pages/moderation.go` (add SLA metrics)
- Modify: `internal/admin/pages/templates/moderation.tmpl`
- Modify: `internal/repository/postgres/bathhouse_repo.go` (moderation queue stats)

- [x] Track moderation assignment (who moderates what, when assigned)
- [x] Calculate metrics: queue size, avg wait time, SLA compliance (48h), per-moderator throughput
- [x] Add metrics section to moderation admin page
- [x] Add alerts when SLA is being breached (items > 24h without review)
- [x] Write tests for SLA calculations
- [x] Run project test suite - must pass before task 11

### Task 11: Calendar Day & Month Views (FR-076)

**Why:** BRD requires day (hourly grid), week, and month views. Only weekly view exists.

**Files:**
- Modify: `frontend/src/pages/calendar/CalendarPage.tsx`

- [x] Add day view: hourly grid with bookings as colored blocks
- [x] Add month view: cells with booking count/status summary
- [x] Color coding: green (confirmed), yellow (pending), red (cancelled), gray (blocked)
- [x] Add view switcher (day/week/month) in calendar header
- [x] Multi-object consolidated view for owners with multiple bathhouses
- [x] Write frontend tests for new views
- [x] Run project test suite - must pass before task 12

### Task 12: CRM - RFM Analysis & Custom Segments (FR CRM)

**Why:** BRD requires RFM (Recency-Frequency-Monetary) analysis and custom segment constructor.

**Files:**
- Create: `internal/service/rfm_service.go`
- Modify: `internal/domain/guest_card.go` (add RFM scores)
- Modify: `internal/handler/guest_card_handler.go` (RFM endpoint)
- Create: `internal/domain/custom_segment.go`
- Create: `internal/repository/postgres/custom_segment_repo.go`
- Create: `frontend/src/pages/crm/RFMAnalysis.tsx`
- Create: `frontend/src/pages/crm/SegmentBuilder.tsx`

- [x] Calculate RFM scores (1-5 each) for each guest card based on last visit, visit count, total spent
- [x] Add endpoint: GET /api/v1/my/crm/rfm (returns guests with RFM scores and matrix)
- [x] Create custom_segments table (id, owner_id, bathhouse_id, name, conditions JSONB)
- [x] Conditions: visit_count (min/max), avg_check (min/max), last_visit_days_ago (min/max), tags (include/exclude), rfm_score ranges
- [x] CRUD endpoints: /api/v1/my/crm/segments/custom
- [x] Dynamic evaluation: resolve segment -> guest list on demand
- [x] Frontend RFM matrix visualization and segment builder UI
- [x] Write tests for RFM calculation and segment evaluation
- [x] Run project test suite - must pass before task 13

### Task 13: Broadcast Personalization & SMS Channel (FR CRM)

**Why:** BRD requires personalization tokens in broadcasts (name, last visit) and SMS channel.

**Files:**
- Modify: `internal/domain/broadcast.go` (add SMS channel)
- Modify: `internal/service/broadcast_service.go` (personalization, SMS)
- Modify: `internal/notification/dispatcher.go` (SMS for broadcasts)
- Modify: `frontend/src/pages/crm/BroadcastCreate.tsx`

- [x] Add SMS to broadcast channels enum
- [x] Implement template personalization: {{guest_name}}, {{last_visit_date}}, {{visit_count}}, {{promo_code}}
- [x] Replace placeholders with guest card data when sending
- [x] Route SMS broadcasts through existing SMSProvider
- [x] Add personalization token picker in frontend broadcast editor
- [x] Update broadcast statistics: add clicked field
- [x] Write tests for personalization and SMS delivery
- [x] Run project test suite - must pass before task 14

### Task 14: L1 Support Bot (FR-155)

**Why:** BRD requires automated L1 support with FAQ answers, SLA < 5 min.

**Files:**
- Create: `internal/domain/faq.go`
- Create: `internal/repository/postgres/faq_repo.go`
- Create: `internal/service/faq_bot_service.go`
- Create: `internal/handler/faq_handler.go`
- Modify: `internal/handler/ticket_handler.go` (auto-answer before creating ticket)
- Create: `frontend/src/components/SupportChatBot.tsx`

- [x] Create faq table (id, category, question, answer, keywords, sort_order, active)
- [x] Admin CRUD: /api/v1/admin/faq
- [x] FAQ matching: keyword/trigram search against incoming support message
- [x] When client opens support: show top 3 matching FAQ answers first
- [x] If client clicks "not helpful" -> escalate to L2 (create ticket)
- [x] Frontend: chat-like FAQ widget before ticket creation
- [x] Seed initial FAQ entries (booking, payment, cancellation, wallet common questions)
- [x] Write tests for FAQ matching logic
- [x] Run project test suite - must pass before task 15

### Task 15: Notification Preferences & SMS Fallback (FR-140, FR-141, FR-142)

**Why:** BRD requires per-event/per-channel notification preferences and push -> email -> SMS fallback chain.

**Files:**
- Create: `internal/domain/notification_preferences.go`
- Create: `internal/repository/postgres/notification_preferences_repo.go`
- Modify: `internal/notification/dispatcher.go` (preferences + fallback)
- Modify: `internal/handler/notification_handler.go` (preferences endpoints)
- Create: `frontend/src/pages/client/NotificationPreferences.tsx`

- [x] Create notification_preferences table (user_id, event_type, push_enabled, email_enabled, sms_enabled)
- [x] Define event types enum matching BRD table (booking_confirmed, reminder_24h, etc.)
- [x] Mark mandatory events (booking confirmation, dispute resolution) - cannot be disabled
- [x] Implement fallback chain: push first -> if not delivered in 5 min -> email -> if critical -> SMS
- [x] Track push delivery status via FCM delivery receipts
- [x] Endpoints: GET/PUT /api/v1/my/notification-preferences
- [x] Frontend preferences matrix (event types x channels toggle grid)
- [x] Trigger push permission request after first booking (FR-142)
- [x] Write tests for preference filtering and fallback chain
- [x] Run project test suite - must pass before task 16

### Task 16: Owner Outgoing Webhooks (FR section 2.16)

**Why:** BRD requires owners to configure webhooks for booking/cancellation/payment events to their CRM.

**Files:**
- Create: `internal/domain/webhook.go`
- Create: `internal/repository/postgres/webhook_repo.go`
- Create: `internal/service/webhook_service.go`
- Create: `internal/handler/webhook_handler.go`
- Modify: `internal/service/booking_service.go` (emit webhook events)
- Create: `frontend/src/pages/settings/WebhookSettings.tsx`

- [x] Create webhooks table (id, owner_id, url, secret, events[], active, created_at)
- [x] Events: booking.created, booking.confirmed, booking.cancelled, booking.completed, payment.received
- [x] HMAC-SHA256 signature in X-Webhook-Signature header
- [x] Async delivery with retry (3 attempts, exponential backoff)
- [x] Create webhook_deliveries table (webhook_id, event, payload, status, attempts, last_error)
- [x] Owner endpoints: CRUD /api/v1/my/webhooks + GET /api/v1/my/webhooks/{id}/deliveries
- [x] Frontend webhook configuration page with test button
- [x] Write tests for signature generation and delivery retry
- [x] Run project test suite - must pass before task 17

### Task 17: PMS API Integration Framework (FR section 2.16)

**Why:** BRD requires Yclients and Restoplace integration for large venues/chains.

**Files:**
- Create: `internal/pms/interface.go`
- Create: `internal/pms/yclients.go`
- Create: `internal/pms/restoplace.go`
- Create: `internal/domain/pms_connection.go`
- Create: `internal/service/pms_service.go`
- Create: `internal/handler/pms_handler.go`
- Create: `frontend/src/pages/settings/PMSIntegration.tsx`

- [x] Define PMSProvider interface: SyncBookings, SyncSchedule, PushBooking, PullBookings
- [x] Create pms_connections table (id, owner_id, bathhouse_id, provider, credentials_encrypted, sync_interval, last_sync)
- [x] Implement Yclients adapter (OAuth2 + REST API)
- [x] Implement Restoplace adapter (API key + REST)
- [x] Bidirectional sync: bookings from PMS -> RelaxHub slots blocked, RelaxHub bookings -> PMS events
- [x] Cron job: sync every 15 min for active connections
- [x] Owner endpoints: CRUD /api/v1/my/pms-connections + POST /api/v1/my/pms-connections/{id}/sync
- [x] Frontend integration settings page
- [x] Write tests with mock PMS responses
- [x] Run project test suite - must pass before task 18

### Task 18: Isochrone Search (FR-053)

**Why:** BRD requires "15 min by car" / "30 min by transit" zones instead of simple radius.

**Files:**
- Create: `internal/geo/isochrone.go`
- Modify: `internal/service/search_service.go` (isochrone filter)
- Modify: `internal/handler/search_handler.go` (isochrone params)
- Modify: `frontend/src/pages/client/BathhouseSearch.tsx`
- Modify: `frontend/src/components/BathhouseMap.tsx` (isochrone polygon)

- [x] Integrate with Yandex Maps isochrone API (or OpenRouteService)
- [x] API: GET /api/v1/isochrone?lat=&lon=&mode=car|transit&minutes=15
- [x] Cache isochrone polygons in Redis (key: lat_lon_mode_minutes, TTL 1h)
- [x] Filter search results: ST_Within(bathhouse.location, isochrone_polygon)
- [x] Display isochrone polygon on map
- [x] Add "travel time" filter: dropdown with 15/30/45/60 min, mode car/transit
- [x] Write tests for isochrone filtering
- [x] Run project test suite - must pass before task 19

### Task 19: Transport Accessibility (FR-054)

**Why:** BRD requires nearest metro/bus/parking info on bathhouse cards and map.

**Files:**
- Create: `internal/geo/transport.go`
- Modify: `internal/domain/bathhouse.go` (transport info cache)
- Modify: `internal/handler/bathhouse_handler.go` (transport endpoint)
- Modify: `frontend/src/pages/client/BathhouseDetail.tsx`

- [x] Integrate with Yandex Maps Geocoder/Search API for nearby POIs (metro, bus stops, parking)
- [x] Endpoint: GET /api/v1/bathhouses/{id}/transport (returns nearest transport with distances)
- [x] Cache in Redis (per bathhouse, TTL 7 days) - transport infrastructure changes rarely
- [x] Display on bathhouse detail page: "Metro Partizanskaya - 500m", "Bus stop - 200m"
- [x] Optional map layer with transport markers
- [x] Write tests for transport data parsing
- [x] Run project test suite - must pass before task 20

### Task 20: SSR for Public Pages (FR section 2.17)

**Why:** BRD requires server-side rendering for public pages (catalog, listings, reviews) for SEO indexing.

**Files:**
- Modify: `internal/handler/seo_handler.go` (serve pre-rendered HTML)
- Create: `internal/seo/renderer.go`

- [x] Evaluate approach: full SSR (Next.js migration) vs prerender service (Rendertron/Prerender.io)
- [x] Implement prerender middleware: detect bot user-agents, serve pre-rendered HTML
- [x] Pre-render public pages: bathhouse listing, bathhouse detail, reviews
- [x] Cache pre-rendered pages in Redis (TTL 1h, invalidate on update)
- [x] Ensure meta tags (OG, Schema.org) are in pre-rendered HTML
- [x] Configure prerender for Yandex, Google, social crawlers
- [x] Write tests for bot detection and cache invalidation
- [x] Run project test suite - must pass before task 21

### Task 21: Professional Photography (FR-035)

**Why:** BRD requires ordering professional photographers from owner cabinet.

**Files:**
- Create: `internal/domain/photo_order.go`
- Create: `internal/repository/postgres/photo_order_repo.go`
- Create: `internal/service/photo_order_service.go`
- Create: `internal/handler/photo_order_handler.go`
- Create: `frontend/src/pages/photos/PhotoOrderPage.tsx`

- [x] Create photo_orders table (id, owner_id, bathhouse_id, region, status, photographer_name, price, scheduled_at, notes)
- [x] Status flow: requested -> confirmed -> completed -> cancelled
- [x] Owner endpoint: POST /api/v1/my/bathhouses/{id}/photo-order (request)
- [x] Admin endpoint: GET /api/v1/admin/photo-orders (manage queue)
- [x] Admin endpoint: PUT /api/v1/admin/photo-orders/{id} (assign photographer, confirm, complete)
- [x] Payment from owner wallet on completion
- [x] Frontend order page with status tracking
- [x] Write tests for order flow
- [x] Run project test suite - must pass before task 22

### Task 22: Excel Import for Listings (FR-033)

**Why:** BRD requires both CSV and Excel format for mass listing import.

**Files:**
- Modify: `internal/service/listing_import_service.go` (add xlsx parsing)
- Modify: `internal/handler/listing_import_handler.go` (accept xlsx)
- Add dependency: `github.com/xuri/excelize/v2`

- [x] Add excelize dependency for .xlsx parsing
- [x] Detect file format by Content-Type or extension
- [x] Parse xlsx with same column mapping as CSV
- [x] Provide downloadable xlsx template (with column headers and sample data)
- [x] Endpoint: same POST /api/v1/my/listings/import (accept both csv and xlsx)
- [x] Write tests for xlsx parsing
- [x] Run project test suite - must pass before task 23

### Task 23: Search UI Enhancements (FR-043, FR-090, FR-044)

**Why:** BRD requires last-minute badges, promoted listing limit (max 3/page), average area price on card.

**Files:**
- Modify: `internal/handler/search_handler.go` (add fields to response)
- Modify: `internal/service/search_service.go` (limit promoted, calc area avg)
- Modify: `frontend/src/components/BathhouseCard.tsx` (badges)
- Modify: `frontend/src/pages/client/BathhouseDetail.tsx` (area avg price)

- [x] Add last_minute_active boolean to search results (true if discount currently applies)
- [x] Display "Last minute -XX%" badge on BathhouseCard when active
- [x] Limit promoted listings to max 3 per page in search results
- [x] Calculate average hourly price in area (same city + 5km radius) for bathhouse detail page
- [x] Display "Average in area: X rub/h" on price breakdown
- [x] Write tests for promoted limit and area avg calculation
- [x] Run project test suite - must pass before task 24

### Task 24: Promoted Listing Campaigns (FR-043)

**Why:** BRD requires auction-based promotion with min 50 rub/day bid, budget management.

**Files:**
- Modify: `internal/domain/promotion.go` (add bid, budget, dates)
- Modify: `internal/service/promotion_service.go` (auction logic, budget deduction)
- Modify: `internal/handler/promotion_handler.go` (campaign CRUD)
- Create: `frontend/src/pages/promotion/PromotionCampaign.tsx`

- [x] Verify promotion domain has: daily_bid (min 50 rub), total_budget, start_date, end_date, status
- [x] Implement daily budget deduction from owner wallet (cron)
- [x] Pause campaign when budget exhausted, notify owner
- [x] Auction ranking: higher bid = more impressions (weighted in search ranking)
- [x] Campaign statistics: impressions, clicks, CTR, cost
- [x] Frontend campaign creation and management page
- [x] Write tests for budget deduction and auction ranking
- [x] Run project test suite - must pass before task 25

### Task 25: Client Profile Completeness (FR-016, FR section 2.18)

**Why:** BRD requires profile completeness indicator with prompts to fill missing fields.

**Files:**
- Modify: `internal/handler/user_handler.go` (completeness endpoint)
- Modify: `frontend/src/components/OnboardingTour.tsx` (completeness widget)
- Modify: `frontend/src/pages/client/ClientProfile.tsx`

- [x] Endpoint returns completeness: name (required), photo, phone, email, preferences, notification_settings
- [x] Calculate percentage (0-100%) based on filled fields
- [x] Show progress bar in client profile with specific prompts for missing items
- [x] Show "complete your profile" nudge on client dashboard
- [x] Write tests for completeness calculation
- [x] Run project test suite - must pass before task 26

### Task 26: Chat Content Filtering Enhancement (FR-063)

**Why:** BRD requires blocking phone numbers, emails, URLs in chat to keep transactions on platform.

**Files:**
- Modify: `internal/service/chat_service.go` (filter outgoing messages)
- Modify: `internal/moderation/filter.go` (add contact detection patterns)

- [x] Verify existing chat filter covers: phone numbers (various RU/BY formats), email addresses, URLs
- [x] Add pattern: Telegram usernames (@username), WhatsApp links
- [x] Replace detected contacts with "[contact info hidden]" message + notification to both parties
- [x] Log filtered messages for anti-fraud review
- [x] Write tests for all contact detection patterns
- [x] Run project test suite - must pass before task 27

### Task 27: Empty States & UX Polish (FR section 2.15, 2.18)

**Why:** BRD requires empty states with explanations and CTAs on every screen without data.

**Files:**
- Modify: multiple frontend pages (BookingList, ReviewList, Favorites, etc.)

- [ ] Audit all list/grid pages for empty state handling
- [ ] Add empty state components with illustration, text, and CTA for:
  - Bookings: "You have no bookings yet. Find a bathhouse nearby?"
  - Reviews: "No reviews yet. Book a visit to leave your first review"
  - Favorites: "Your favorites list is empty. Start exploring!"
  - Owner bookings: "No bookings yet. Make sure your listing is active and complete"
  - Owner CRM: "No guests yet. They'll appear after the first completed booking"
  - Wallet: "Your wallet is empty. Top up to pay faster"
- [ ] Write snapshot tests for empty states
- [ ] Run project test suite - must pass before task 28

### Task 28: Bathhouse Card Enhancements (FR-044)

**Why:** BRD requires similar objects block (up to 6), owner profile section, wallet payment button on card.

**Files:**
- Modify: `internal/handler/bathhouse_handler.go` (similar objects endpoint)
- Modify: `internal/service/recommendation_service.go` (similar by type/location)
- Modify: `frontend/src/pages/client/BathhouseDetail.tsx`

- [ ] Add endpoint: GET /api/v1/bathhouses/{id}/similar (up to 6, same city + type, sorted by rating)
- [ ] Display similar objects section at bottom of bathhouse detail
- [ ] Add owner profile section: rating, number of objects, registration date
- [ ] Add "Pay from wallet" quick button on price section (if balance sufficient)
- [ ] Write tests for similar objects query
- [ ] Run project test suite - must pass before task 29

### Task 29: Listing Wizard Video Step (FR-034 step 1)

**Why:** BRD requires video instruction on step 1 of the 7-step listing wizard.

**Files:**
- Modify: `frontend/src/pages/bathhouses/BathhouseForm.tsx` (step 1)

- [ ] Add welcome/intro step with embedded video player (YouTube/Vimeo embed or self-hosted)
- [ ] Video placeholder with "How to create a listing" content
- [ ] "Skip" button to proceed directly to step 2
- [ ] Store video URL in platform_settings for admin configurability
- [ ] Run project test suite - must pass before task 30

### Task 30: Verify Acceptance Criteria

- [ ] Manual test: complete booking flow RU region (card + wallet combo)
- [ ] Manual test: Belarus region payment flow (placeholder for bePaid)
- [ ] Manual test: admin sub-role access (moderator can moderate but not manage finance)
- [ ] Manual test: CRM RFM analysis and custom segment creation
- [ ] Manual test: webhook delivery to external URL
- [ ] Manual test: FAQ bot answers before ticket creation
- [ ] Run full test suite: `go test ./... -race`
- [ ] Run linter: `make lint`
- [ ] Run frontend tests: `cd frontend && npx vitest run`
- [ ] Run frontend lint: `cd frontend && npm run lint`
- [ ] Verify test coverage meets 80%+

### Task 31: Update Documentation

- [ ] Update CLAUDE.md with new subsystems (saved cards, webhooks, PMS, FAQ bot, admin sub-roles, bank reconciliation)
- [ ] Update swagger annotations for all new endpoints, run `make swagger`
- [ ] Regenerate frontend API client: `make frontend-generate-api`
- [ ] Move this plan to `docs/plans/completed/`
