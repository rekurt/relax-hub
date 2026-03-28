# Gap Analysis & Implementation Plan: BRD RelaxHub v6 Full Compliance

## Overview

Comprehensive plan to bring the platform into full compliance with BRD RelaxHub v6. Based on deep analysis of all 170+ FR requirements vs current implementation across backend (81 domain models, 91+ handlers, 162 migrations) and frontend (48 pages). Plan covers missing features, incomplete implementations, and absent frontend pages for existing backend APIs.

## Context

- BRD document: `docs/BRD_RelaxHub_v6.md` (170+ functional requirements)
- Current state: ~80% backend coverage, ~65% frontend coverage
- Key gaps: per-bathhouse cancellation policies, booking modifications, region switching, insurance deposit, bidirectional reviews, financial reporting, missing frontend pages for CRM/tickets/disputes/comparison
- Files involved: `internal/domain/`, `internal/handler/`, `internal/service/`, `internal/repository/`, `frontend/src/pages/`, `migrations/`
- Related patterns: clean architecture (handler -> service -> repository), Uber fx DI, chi router
- Dependencies: excelize (Excel import), PDF generation library (e.g. go-pdf or wkhtmltopdf)

## Gap Summary

**FULLY MISSING (neither backend nor frontend):**
- FR-017: Region switching (BY <-> RU)
- FR-022: Per-bathhouse cancellation policies (Flexible/Moderate/Strict)
- FR-033: Mass CSV/Excel import for listings
- FR-047: Share listing (Open Graph, deep links)
- FR-048/049/050: Map search with clustering, split-view, isochrones
- FR-053/054: Isochrones and transport accessibility
- FR-067: Booking modifications (change date/time/guests/addons)
- FR-081: Share booking
- FR-087: Seasonal tariffs (date range pricing)
- FR-089: Smart pricing recommendations
- FR-091: Insurance deposit (security deposit)
- FR-096: Payment error retry with exponential backoff
- FR-104: Financial reports/acts generation (PDF, 1C XML)
- FR-115/120: Wallet history export (CSV/PDF)
- FR-122: Bonus expiry notifications (14d/3d warnings)
- FR-125: Transaction reconciliation with payment provider
- FR-127/128: Float monitoring dashboard, bank reconciliation
- FR-130/131: Bidirectional reviews (owner rates client, double-blind reveal)
- FR-147-154: Advanced analytics (conversion funnels, cohort analysis, demand/supply heatmap)
- FR-155 L1: Support bot (automated answers)
- Representative roles expansion (manager/observer with per-object access)

**BACKEND EXISTS, FRONTEND MISSING:**
- CRM pages (guest cards, segments, broadcasts, auto-scenarios, templates)
- Support ticket pages (create, list, thread, CSAT)
- Dispute pages (create, evidence upload, appeal, admin mediation)
- Comparison page (2-3 bathhouses side-by-side)
- Saved searches management UI
- Anti-fraud admin dashboard

**PARTIALLY IMPLEMENTED (needs completion):**
- FR-074: Calendar sync (iCal export exists, bidirectional sync incomplete)
- FR-079: Response rate monitoring (cron exists, thresholds need BRD alignment)
- FR-092: Payment methods (YooKassa exists, Apple Pay/Google Pay/BELKART/ERIP missing)
- Chat content filtering (regex exists, phone/email/URL blocking per FR-063 needs verification)
- FR-016: Onboarding tour and profile completeness indicator

## Development Approach

- **Testing approach**: Regular (code first, then tests)
- Complete each task fully before moving to the next
- Backend tasks use clean architecture pattern: domain -> repository -> service -> handler
- Frontend tasks use existing patterns: Ant Design + React Query + orval generation
- **CRITICAL: every task MUST include new/updated tests**
- **CRITICAL: all tests must pass before starting next task**

## Implementation Steps

### Task 1: Per-bathhouse cancellation policies (FR-022)

**Files:**
- Modify: `internal/domain/bathhouse.go` (add CancellationPolicy field: flexible/moderate/strict)
- Modify: `internal/domain/booking.go` (refund calculation based on policy)
- Create: `internal/domain/cancellation_policy.go` (policy definitions with time windows and refund percentages)
- Modify: `internal/service/booking_service.go` (calculate refund by policy)
- Modify: `internal/service/payment_service.go` (apply policy-based refund)
- Modify: `internal/handler/bathhouse_handler.go` (expose policy in API)
- Modify: `internal/handler/booking_handler.go` (show refund amount per policy)
- Create: migration for cancellation_policy column on bathhouses
- Modify: `frontend/src/pages/client/BookingCreate.tsx` (display policy)
- Modify: `frontend/src/pages/client/BathhouseDetail.tsx` (display policy)
- Modify: `frontend/src/pages/bathhouses/BathhouseForm.tsx` (policy selector)

- [x] Define CancellationPolicy type with 3 presets (flexible: 100%/24h+50%/<24h, moderate: 100%/72h+50%/24-72h+0%/<24h, strict: 100%/7d+50%/3-7d+0%/<3d)
- [x] Add cancellation_policy field to bathhouse domain model and migration
- [x] Update refund calculation in booking service to use per-bathhouse policy instead of flat rules
- [x] Update payment service refund logic
- [x] Expose policy in bathhouse API responses and listing creation/edit
- [x] Frontend: add policy selector in BathhouseForm, display on BathhouseDetail and BookingCreate
- [x] Write tests for all 3 policies with edge cases (boundary times)
- [x] Run project test suite - must pass before next task

### Task 2: Booking modifications (FR-067)

**Files:**
- Modify: `internal/domain/booking.go` (add ModificationCount, MaxModifications=3)
- Create: `internal/service/booking_modification_service.go`
- Modify: `internal/handler/booking_handler.go` (PUT /api/v1/bookings/{id}/modify)
- Create: migration for modification_count column
- Modify: `frontend/src/pages/client/BookingDetail.tsx` (modify button + form)

- [x] Add modification_count field to booking model, max 3 modifications
- [x] Create modification service: validate new slot availability, recalculate price, handle price difference (charge or refund proportionally)
- [x] Handle combo payment modifications (wallet + card proportional adjustments)
- [x] Add API endpoint for booking modification
- [x] Frontend: add modify button on BookingDetail with date/time/duration/addons change form
- [x] Write tests for modification scenarios (price up, price down, equal, max modifications reached, slot conflicts)
- [x] Run project test suite - must pass before next task

### Task 3: Insurance deposit / security deposit (FR-091, FR-103)

**Files:**
- Modify: `internal/domain/bathhouse.go` (add SecurityDeposit field, max 50% of base price)
- Modify: `internal/domain/booking.go` (add DepositAmount, DepositStatus)
- Modify: `internal/service/booking_service.go` (hold deposit on booking)
- Modify: `internal/service/payment_service.go` (deposit hold/release/claim)
- Create: migration for deposit fields
- Modify: `frontend/src/pages/client/BookingCreate.tsx` (show deposit info)
- Modify: `frontend/src/pages/bathhouses/BathhouseForm.tsx` (deposit config)

- [x] Add security_deposit_percent field to bathhouse (0-50%, default 0)
- [x] Add deposit tracking fields to booking (amount, status: none/held/released/claimed)
- [x] Implement deposit hold on booking creation (card hold via YooKassa)
- [x] Auto-release deposit 48h after check-out if no dispute
- [x] Deposit freeze on dispute opening
- [x] Frontend: display deposit info in booking flow and bathhouse settings
- [x] Write tests for deposit lifecycle
- [x] Run project test suite - must pass before next task

### Task 4: Seasonal tariffs (FR-087)

**Files:**
- Modify: `internal/domain/bathhouse.go` or create `internal/domain/seasonal_tariff.go`
- Create: `internal/repository/postgres/seasonal_tariff_repo.go`
- Modify: `internal/service/pricing_service.go` (apply seasonal multipliers)
- Modify: `internal/handler/pricing.go` (CRUD for seasonal tariffs)
- Create: migration for seasonal_tariffs table
- Modify: `frontend/src/pages/pricing/PricingRules.tsx` (seasonal tariff UI)

- [x] Create seasonal_tariffs model (bathhouse_id, date_from, date_to, multiplier, name)
- [x] Add repository and service CRUD
- [x] Integrate into price calculation pipeline (after base price, before other adjustments)
- [x] Add API endpoints for managing seasonal tariffs
- [x] Frontend: add seasonal tariff management in PricingRules page
- [x] Write tests for seasonal pricing with overlapping ranges
- [x] Run project test suite - must pass before next task

### Task 5: Bidirectional reviews - owner rates client (FR-130, FR-131)

**Files:**
- Create: `internal/domain/client_review.go` (owner review of client)
- Create: `internal/repository/postgres/client_review_repo.go`
- Create: `internal/service/client_review_service.go`
- Modify: `internal/handler/review_handler.go` (add owner-to-client review endpoints)
- Modify: `internal/domain/review.go` (add reveal logic: both posted or 14 days)
- Create: migration for client_reviews table and reveal_at field on reviews

- [x] Create client_review domain model (punctuality, cleanliness, rule_compliance ratings)
- [x] Implement double-blind reveal: reviews hidden until both posted OR 14 days elapsed
- [x] Add reveal_at computed field, cron job for auto-reveal after 14 days
- [x] Add API endpoints for owner to rate client
- [x] Update existing review endpoints to respect blind period
- [x] Write tests for reveal logic (both post, timeout, edit before reveal)
- [x] Run project test suite - must pass before next task

### Task 6: Region switching (FR-017)

**Files:**
- Modify: `internal/domain/user.go` (add Region field: RU/BY)
- Create: `internal/service/region_service.go` (switch region with validation)
- Modify: `internal/handler/auth_handler.go` or create region handler
- Create: migration for user region field and wallet archival
- Modify: `internal/domain/wallet.go` (add Currency, archived status)

- [x] Add region field to user model (default RU)
- [x] Implement region switch validation: block if non-zero wallet balance, active bookings, open disputes, or unactivated certificates
- [x] On switch: archive old wallet, create new wallet in new currency, reset loyalty status
- [x] Block cross-regional bookings (client region must match bathhouse region)
- [x] Write tests for region switching with all blocking conditions
- [x] Run project test suite - must pass before next task

### Task 7: Financial reports and acts (FR-104, FR-115, FR-120)

**Files:**
- Create: `internal/service/financial_report_service.go`
- Create: `internal/handler/financial_report_handler.go`
- Modify: `internal/handler/wallet_handler.go` (add export endpoints)
- Create: `internal/service/pdf_generator.go` (PDF report generation)

- [x] Implement wallet history export to CSV and PDF for clients and owners
- [x] Implement act generation for owners (PDF format: booking details, amounts, dates)
- [x] Implement XML export for legal entities (1C-compatible format)
- [x] Add API endpoints: GET /api/v1/my/wallet/export?format=csv|pdf, GET /api/v1/my/finance/acts
- [x] Write tests for export formats and data correctness
- [x] Run project test suite - must pass before next task

### Task 8: Float monitoring and transaction reconciliation (FR-125, FR-127, FR-128)

**Files:**
- Create: `internal/service/reconciliation_service.go`
- Create: `internal/handler/reconciliation_handler.go`
- Modify: `internal/admin/pages/` (add finance dashboard)
- Create: `internal/cron/reconciliation_job.go`

- [x] Implement daily float snapshot: sum(client wallets) + sum(owner wallets) + sum(escrow) = expected total
- [x] Implement transaction reconciliation with YooKassa (compare internal transactions with provider)
- [x] Alert on any discrepancy (zero tolerance per FR-125)
- [x] Add admin dashboard widget showing float status
- [x] Add cron job for daily reconciliation
- [x] Write tests for reconciliation logic
- [x] Run project test suite - must pass before next task

### Task 9: Bonus expiry notifications (FR-122)

**Files:**
- Create: `internal/cron/bonus_expiry_notification_job.go`
- Modify: `internal/service/notification_service.go`
- Modify: `internal/domain/notification.go` (add bonus_expiry_warning type)

- [x] Create cron job to find bonuses expiring in 14 days and 3 days
- [x] Send push + email notifications with expiring amount and suggestion to use
- [x] Deduplicate notifications (don't re-send if already notified)
- [x] Write tests for notification triggers and deduplication
- [x] Run project test suite - must pass before next task

### Task 10: Payment error retry logic (FR-096)

**Files:**
- Modify: `internal/payment/` (add retry logic with exponential backoff)
- Modify: `internal/service/payment_service.go` (handle retryable vs permanent errors)
- Modify: `internal/handler/payment_handler.go` (user-friendly error messages in Russian)

- [x] Classify payment errors: retryable (timeout, network) vs permanent (insufficient funds, blocked card)
- [x] Implement retry for retryable errors: up to 3 attempts with 2s, 4s, 8s delays
- [x] Map provider error codes to Russian user-friendly messages
- [x] Return helpful suggestions (e.g., "Попробуйте другую карту")
- [x] Write tests for retry logic and error classification
- [x] Run project test suite - must pass before next task

### Task 11: Mass CSV/Excel import for listings (FR-033)

**Files:**
- Create: `internal/service/listing_import_service.go`
- Create: `internal/handler/listing_import_handler.go`
- Create: CSV/Excel template file

- [x] Create import template (CSV format with all required listing fields)
- [x] Implement file upload and parsing (CSV, optionally Excel via excelize)
- [x] Validate each row against listing requirements (required fields, photo URLs, price ranges)
- [x] Create listings as drafts (pending moderation)
- [x] Return import report (success count, error details per row)
- [x] Add API endpoint: POST /api/v1/my/listings/import
- [x] Write tests for import validation and error handling
- [x] Run project test suite - must pass before next task

### Task 12: Share listing and share booking (FR-047, FR-081)

**Files:**
- Modify: `internal/handler/bathhouse_handler.go` (Open Graph meta tags)
- Modify: `internal/seo/schema.go` (enhance OG tags)
- Create: `internal/handler/share_handler.go` (share booking link generation)
- Modify: `internal/handler/booking_handler.go` (share booking endpoint)

- [x] Enhance Open Graph meta tags on bathhouse pages (photo, title, description, price)
- [x] Generate shareable booking links with pre-filled date/bathhouse params
- [x] Add API endpoint: POST /api/v1/bookings/{id}/share -> returns share URL
- [x] Handle deep link resolution: /share/booking/{token} -> booking page with pre-filled data
- [x] Write tests for share URL generation and resolution
- [x] Run project test suite - must pass before next task

### Task 13: Representative roles expansion (FR - Owner management section)

**Files:**
- Modify: `internal/domain/user.go` (add representative sub-roles: manager/observer)
- Modify: `internal/middleware/` (add per-object access control)
- Modify: `internal/handler/representative_handler.go`
- Create: migration for representative_role and object_access

- [x] Add sub-roles for representatives: manager (manage bookings, reply messages) and observer (read-only)
- [x] Implement per-bathhouse access control (representative can be assigned to specific objects)
- [x] Audit all representative actions in audit log
- [x] Add invite-by-email flow for representatives
- [x] Frontend: update RepresentativeList page with role and object assignment
- [x] Write tests for role-based access with object-level restrictions
- [x] Run project test suite - must pass before next task

### Task 14: Frontend - CRM pages for owners

**Files:**
- Create: `frontend/src/pages/crm/GuestCardList.tsx`
- Create: `frontend/src/pages/crm/GuestCardDetail.tsx`
- Create: `frontend/src/pages/crm/SegmentList.tsx`
- Create: `frontend/src/pages/crm/BroadcastList.tsx`
- Create: `frontend/src/pages/crm/BroadcastCreate.tsx`
- Create: `frontend/src/pages/crm/AutoScenarios.tsx`
- Create: `frontend/src/pages/crm/ResponseTemplates.tsx`
- Modify: `frontend/src/router.tsx` (add CRM routes)

- [x] Regenerate API client (npm run generate:api) to include CRM endpoints
- [x] Guest card list page: search, filter by tags/segment, CSV export
- [x] Guest card detail page: visit history, LTV, avg check, notes, tags editor
- [x] Segment list page: view predefined segments (new/regular/lost/VIP/birthday), segment member counts
- [x] Broadcast management: create broadcast with text+image, select segment, attach promo, view send statistics
- [x] Auto-scenarios page: toggle on/off, customize text/timing/channel per scenario
- [x] Response templates page: CRUD for quick reply templates
- [x] Write frontend tests for key CRM flows
- [x] Run project test suite - must pass before next task

### Task 15: Frontend - Support tickets pages

**Files:**
- Create: `frontend/src/pages/client/SupportTickets.tsx`
- Create: `frontend/src/pages/client/TicketDetail.tsx`
- Create: `frontend/src/pages/admin/TicketManagement.tsx`
- Create: `frontend/src/pages/admin/TicketDetail.tsx`
- Modify: `frontend/src/router.tsx`

- [x] Regenerate API client for ticket endpoints
- [x] Client ticket list: create ticket (category selector), view open/resolved tickets
- [x] Client ticket detail: message thread with attachments, CSAT survey after resolution
- [x] Admin ticket list: filterable queue by status/priority/escalation level, assign to agent
- [x] Admin ticket detail: respond, escalate, resolve, close with full context
- [x] Write frontend tests
- [x] Run project test suite - must pass before next task

### Task 16: Frontend - Dispute pages

**Files:**
- Create: `frontend/src/pages/client/DisputeCreate.tsx`
- Create: `frontend/src/pages/client/DisputeDetail.tsx`
- Create: `frontend/src/pages/client/DisputeList.tsx`
- Create: `frontend/src/pages/admin/DisputeManagement.tsx`
- Create: `frontend/src/pages/admin/DisputeDetail.tsx`
- Modify: `frontend/src/router.tsx`

- [x] Regenerate API client for dispute endpoints
- [x] Client: open dispute from booking detail, upload evidence (photos, screenshots, GPS, messages, receipts)
- [x] Client: view dispute status, submit appeal within 7 days of resolution
- [x] Admin: dispute queue with assignment, view evidence from both sides, render resolution (full/partial/no refund + compensation)
- [x] Write frontend tests
- [x] Run project test suite - must pass before next task

### Task 17: Frontend - Comparison page and map search

**Files:**
- Create: `frontend/src/pages/client/ComparisonPage.tsx`
- Modify: `frontend/src/pages/client/BathhouseSearch.tsx` (add map mode, split-view)
- Create: `frontend/src/components/BathhouseMap.tsx`

- [x] Regenerate API client for comparison endpoints
- [x] Comparison page: side-by-side table for 2-3 bathhouses (price, rating, capacity, amenities, distance, cancellation policy)
- [x] Add "Compare" button on bathhouse cards in search results
- [x] Map view in search: Yandex Maps with price markers, clustering for 50+ results
- [x] Split-view mode: list on left, map on right, synchronized highlights
- [x] "Search in this area" button on map drag
- [x] Write frontend tests
- [x] Run project test suite - must pass before next task

### Task 18: Frontend - Saved searches and anti-fraud admin

**Files:**
- Create: `frontend/src/pages/client/SavedSearches.tsx`
- Create: `frontend/src/pages/admin/AntiFraudDashboard.tsx`
- Modify: `frontend/src/router.tsx`

- [x] Saved searches page: list saved searches, delete, view notification history
- [x] Anti-fraud admin dashboard: pending fraud flags, review actions (approve/dismiss), filter by rule type
- [x] Write frontend tests
- [x] Run project test suite - must pass before next task

### Task 19: Advanced analytics (FR-147-154)

**Files:**
- Modify: `internal/service/analytics_service.go` (add conversion funnels, cohort, demand/supply metrics)
- Modify: `internal/handler/analytics.go`
- Create: `internal/admin/pages/advanced_analytics.go`
- Modify: `frontend/src/pages/admin/AdminDashboard.tsx` (add funnel/cohort charts)

- [x] Conversion funnel: visit -> search -> view card -> start booking -> pay -> complete visit (with counts and percentages per step)
- [x] Cohort analysis: group users by registration month, track retention and spending over time
- [x] Geographic demand/supply map data: count of searches vs count of listings per city/region
- [x] Wallet metrics: sum of balances, wallet payment share %, expired bonuses volume
- [x] Owner analytics: conversion rate (views -> bookings), occupancy rate, competitor benchmarking (anonymous average comparison)
- [x] Add API endpoints and admin dashboard charts
- [x] Write tests for analytics calculations
- [x] Run project test suite - must pass before next task

### Task 20: Onboarding tour and profile completeness (FR-016)

**Files:**
- Modify: `internal/domain/user.go` (add onboarding_completed flag, profile_completeness)
- Modify: `internal/service/user_service.go` (calculate completeness)
- Modify: `internal/handler/auth_handler.go` (return onboarding status)
- Modify: `frontend/src/pages/client/ClientProfile.tsx` (completeness indicator)
- Create: `frontend/src/components/OnboardingTour.tsx`

- [ ] Add profile completeness calculation (name, photo, phone, preferences)
- [ ] Add onboarding_completed flag to skip tour on subsequent visits
- [ ] Frontend: create onboarding tour component (step-by-step guide: search, book, wallet)
- [ ] Frontend: add profile completeness indicator with prompts to complete
- [ ] Write tests
- [ ] Run project test suite - must pass before next task

### Task 21: Admin справочники management (amenities, object types, holidays)

**Files:**
- Create: `frontend/src/pages/admin/AmenityManagement.tsx`
- Create: `frontend/src/pages/admin/ObjectTypeManagement.tsx`
- Create: `frontend/src/pages/admin/HolidayManagement.tsx`
- Modify: `frontend/src/router.tsx`
- If backend CRUD for amenities/object types is missing, add it

- [ ] Verify backend CRUD endpoints exist for amenities, object types, holidays; create if missing
- [ ] Admin page: manage amenities (add/edit/delete, icon assignment)
- [ ] Admin page: manage object types (bathhouse categories)
- [ ] Admin page: manage holidays per region (name, date, recurring flag)
- [ ] Write tests
- [ ] Run project test suite - must pass before next task

### Task 22: Admin financial dashboard (FR-152, FR-127)

**Files:**
- Modify: `internal/admin/pages/` (add finance page)
- Create: `internal/admin/pages/finance.go`
- Create: `internal/admin/pages/templates/finance.tmpl`

- [ ] Float dashboard: client wallets total + owner wallets total + escrow total
- [ ] Transaction reconciliation status (last run, discrepancies)
- [ ] Revenue breakdown: service fees, subscriptions, promotions
- [ ] Wallet metrics widget
- [ ] Write tests
- [ ] Run project test suite - must pass before next task

### Task 23: Chat content filtering enhancement (FR-063)

**Files:**
- Modify: `internal/antifraud/chat_filter.go` (add phone/email/URL blocking)
- Modify: `internal/service/chat_service.go`

- [ ] Verify and enhance chat filter to block: phone numbers (all formats), email addresses, URLs
- [ ] Replace blocked content with [hidden] placeholder, not reject message entirely
- [ ] Log blocked content attempts for anti-fraud review
- [ ] Write tests with various phone/email/URL formats
- [ ] Run project test suite - must pass before next task

### Task 24: Calendar bidirectional sync enhancement (FR-074)

**Files:**
- Modify: `internal/service/calendar_service.go`
- Create: `internal/cron/calendar_sync_job.go`

- [ ] Implement inbound sync: fetch external calendar (iCal URL) every 15 minutes
- [ ] Parse external calendar events and auto-create slot blocks
- [ ] Implement outbound sync: generate iCal feed with all bookings as events
- [ ] Handle conflict resolution when external event overlaps existing booking
- [ ] Write tests for sync logic
- [ ] Run project test suite - must pass before next task

### Task 25: Smart pricing recommendations (FR-089)

**Files:**
- Create: `internal/service/smart_pricing_service.go`
- Modify: `internal/handler/pricing.go`

- [ ] Calculate recommended price based on: current occupancy, average area prices, demand patterns
- [ ] Recommend coefficient range 0.8 to 1.5
- [ ] Add API endpoint: GET /api/v1/my/bathhouses/{id}/price-recommendation
- [ ] Frontend: show recommendation in pricing settings with accept/dismiss
- [ ] Write tests
- [ ] Run project test suite - must pass before next task

### Task 26: Owner wallet export and financial documents

**Files:**
- Modify: `internal/handler/wallet_handler.go`
- Modify: `internal/handler/payout_handler.go`
- Modify: `frontend/src/pages/client/PaymentHistory.tsx` (add export buttons)

- [ ] Add CSV export endpoint for wallet transactions (both client and owner)
- [ ] Add PDF export for wallet statement
- [ ] Frontend: add export buttons (CSV/PDF) on payment history and wallet pages
- [ ] Write tests for export formatting
- [ ] Run project test suite - must pass before next task

### Task 27: Verify acceptance criteria

- [ ] Manual test: create bathhouse with cancellation policy, verify correct refund calculation on cancel
- [ ] Manual test: modify booking (change time), verify price recalculation
- [ ] Manual test: CRM pages load with guest data
- [ ] Manual test: open dispute from booking, upload evidence
- [ ] Manual test: comparison of 2 bathhouses works
- [ ] Run full test suite: `go test ./... -v`
- [ ] Run linter: `make lint`
- [ ] Run frontend tests: `cd frontend && npx vitest run`
- [ ] Run frontend lint: `cd frontend && npm run lint`
- [ ] Verify test coverage meets 80%+

### Task 28: Update documentation

- [ ] Update CLAUDE.md with new features (cancellation policies, booking modifications, deposits, seasonal tariffs, bidirectional reviews, region switching, financial reports, CRM frontend, tickets frontend, disputes frontend)
- [ ] Update swagger annotations and regenerate: `make swagger`
- [ ] Regenerate frontend API client: `make frontend-generate-api`
- [ ] Move this plan to `docs/plans/completed/`
