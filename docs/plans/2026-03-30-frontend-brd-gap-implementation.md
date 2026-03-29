# Frontend Gap Analysis & Implementation Plan - BRD RelaxHub v6

## Overview
Comprehensive frontend implementation plan to close all gaps between the BRD RelaxHub v6 requirements and the current SPA. The existing frontend has ~84 pages covering most core flows, but ~25 pages/features are missing or incomplete. This plan addresses every gap systematically.

## Context
- Files involved: `frontend/src/pages/`, `frontend/src/components/`, `frontend/src/router.tsx`, `frontend/src/stores/`, `frontend/src/lib/`
- Related patterns: Ant Design 6 + React Query (orval-generated) + Zustand stores, role-based layouts (AppLayout/ClientLayout/AdminLayout)
- Dependencies: All backend API endpoints exist in swagger.json, orval API client already generated with 71 modules
- All new pages follow existing pattern: React Query hooks from `@/api/generated/*`, Ant Design components, Russian locale

## Development Approach
- **Testing approach**: Regular (code first, then tests via Vitest + React Testing Library)
- Complete each task fully before moving to the next
- Each page follows existing patterns: generated API hooks, Ant Design UI, role-based routing
- **CRITICAL: every task MUST include new/updated tests**
- **CRITICAL: all tests must pass before starting next task**
- **CRITICAL: run `cd frontend && npm run lint` after each task - zero warnings policy**

## Implementation Steps

### Task 1: Fix PMSIntegration Routing & Owner Settings Restructure

**Files:**
- Modify: `frontend/src/router.tsx`
- Modify: `frontend/src/components/AppLayout.tsx`

- [x] Add `/settings/pms` route pointing to existing PMSIntegration component
- [x] Add PMS Integration link to AppLayout sidebar under Settings section
- [x] Verify PMSIntegration page renders correctly with API calls
- [x] Write test for route accessibility and rendering
- [x] Run project test suite + lint

### Task 2: Client Wallet Dashboard (FR-110-116)

**Files:**
- Create: `frontend/src/pages/client/WalletDashboard.tsx`
- Modify: `frontend/src/router.tsx`
- Modify: `frontend/src/components/ClientLayout.tsx`

- [x] Build wallet page with: balance display (single number per FR-116), top-up form with amount validation (min 500 / max 30,000 RUB), transaction history table with type/date/amount/description/linked booking
- [x] Add filters: by type (topup/refund/cashback/promo/referral/welcome_bonus/gift_cert), date range
- [x] Add bonus expiry warnings section (show bonuses expiring within 30 days with amounts and dates)
- [x] Add CSV/PDF export buttons for wallet history (FR-115)
- [x] Add hold indicator showing frozen amounts (FR-113)
- [x] Add route `/client/wallet` and sidebar link in ClientLayout
- [x] Write tests for wallet dashboard rendering, top-up validation, filter interactions
- [x] Run project test suite + lint

### Task 3: Client Saved Cards Management (FR-004, FR-092)

**Files:**
- Create: `frontend/src/pages/client/SavedCards.tsx`
- Modify: `frontend/src/router.tsx`

- [x] Build saved cards list with: last4, brand icon, expiry, default card toggle
- [x] Add card deletion with confirmation dialog
- [x] Add route `/client/cards` and link from ClientProfile
- [x] Write tests for card list rendering and delete flow
- [x] Run project test suite + lint

### Task 4: Client Session Management & Security (FR-007, FR-008, FR-015)

**Files:**
- Create: `frontend/src/pages/client/SecuritySettings.tsx`
- Modify: `frontend/src/router.tsx`
- Modify: `frontend/src/components/ClientLayout.tsx`

- [x] Build security settings page with three sections:
  - Active sessions list (device, browser, IP, last active, terminate button) per FR-008
  - 2FA setup: TOTP QR code generation + verification code input, SMS 2FA toggle per FR-007
  - Password change form (current + new + confirm) per FR-015
- [x] Add route `/client/security` and sidebar link
- [x] Write tests for session list, 2FA toggle, password change form
- [x] Run project test suite + lint

### Task 5: Client Account Deletion & Region Switching (FR-014, FR-017)

**Files:**
- Modify: `frontend/src/pages/client/ClientProfile.tsx`

- [x] Add account deletion section: confirmation modal with 30-day grace period explanation, info about fund return (topup refunded, bonuses lost) per FR-014
- [x] Add restore account button if deletion is pending (from profile state)
- [x] Add region switching section: current region display, switch button, validation checks (non-zero balance, active bookings, certificates, disputes) with error messages per FR-017
- [x] Write tests for deletion flow modal, region switch validation
- [x] Run project test suite + lint

### Task 6: Client Recently Viewed & Home Page Enhancement (FR-051, FR-016)

**Files:**
- Modify: `frontend/src/pages/client/ClientHome.tsx`
- Create: `frontend/src/components/RecentlyViewed.tsx`

- [x] Create RecentlyViewed component using Redis-backed API (last 20 items): horizontal scrollable cards with photo, name, price, rating
- [x] Enhance ClientHome with sections: search bar, recently viewed, recommendations, promotional banners, popular nearby (by geolocation)
- [x] Record bathhouse view on BathhouseDetail page open (API call to record recently viewed)
- [x] Write tests for RecentlyViewed component rendering
- [x] Run project test suite + lint

### Task 7: Share Booking & Listing (FR-047, FR-081)

**Files:**
- Create: `frontend/src/components/ShareButton.tsx`
- Modify: `frontend/src/pages/client/BookingDetail.tsx`
- Modify: `frontend/src/pages/client/BathhouseDetail.tsx`
- Create: `frontend/src/pages/ShareRedirect.tsx`
- Modify: `frontend/src/router.tsx`

- [x] Create ShareButton component: native Web Share API with fallback to copy-to-clipboard, generates shareable URL
- [x] Add share button to BookingDetail (generates share token via POST /api/v1/bookings/{id}/share)
- [x] Add share button to BathhouseDetail (copy URL with slug)
- [x] Create ShareRedirect page at `/share/booking/:token` that resolves deep link and redirects
- [x] Write tests for ShareButton component and ShareRedirect
- [x] Run project test suite + lint

### Task 8: Owner Finance & Wallet Dashboard (FR-100-104, FR-117-120)

**Files:**
- Create: `frontend/src/pages/finance/FinanceDashboard.tsx`
- Create: `frontend/src/pages/finance/PayoutPage.tsx`
- Create: `frontend/src/pages/finance/FinancialReports.tsx`
- Modify: `frontend/src/router.tsx`
- Modify: `frontend/src/components/AppLayout.tsx`

- [x] Build FinanceDashboard with: wallet balance (available + frozen), income chart (week/month/year), recent transactions table with type/date/amount/booking link, filters by type and date
- [x] Build PayoutPage with: payout form (amount, method: SBP/bank transfer), daily/monthly limit display, payout history table, auto-payout threshold toggle (FR-102)
- [x] Build FinancialReports with: act generation (PDF), wallet history export (CSV/PDF), 1C XML export for legal entities (FR-104)
- [x] Add routes `/finance`, `/finance/payouts`, `/finance/reports` and sidebar section in AppLayout
- [x] Write tests for finance dashboard, payout form validation, export actions
- [x] Run project test suite + lint

### Task 9: Owner Smart Pricing & Analytics (FR-089, owner analytics)

**Files:**
- Create: `frontend/src/pages/analytics/OwnerAnalytics.tsx`
- Modify: `frontend/src/pages/pricing/PricingRules.tsx`
- Modify: `frontend/src/router.tsx`
- Modify: `frontend/src/components/AppLayout.tsx`

- [x] Add smart pricing recommendation widget to PricingRules: shows recommended price coefficient (0.8-1.5) based on occupancy/demand/competitors, accept/dismiss buttons (FR-089)
- [x] Build OwnerAnalytics page with: occupancy chart (% filled slots by period), income dynamics (week/month/year line chart), page views & conversion rate, anonymous competitor comparison (avg price, occupancy, rating in area), CRM metrics (unique guests, repeat rate, avg LTV)
- [x] Add routes `/analytics` and sidebar link
- [x] Write tests for analytics page rendering, smart pricing widget
- [x] Run project test suite + lint

### Task 10: Owner Listing Enhancements - Import, Duplicate, Completeness (FR-025, FR-032, FR-033)

**Files:**
- Create: `frontend/src/pages/bathhouses/ListingImport.tsx`
- Modify: `frontend/src/pages/bathhouses/BathhouseList.tsx`
- Modify: `frontend/src/pages/bathhouses/BathhouseForm.tsx`
- Modify: `frontend/src/router.tsx`

- [x] Build ListingImport page: file upload (CSV/XLSX), template download link, validation results with per-row errors, import progress (FR-033)
- [x] Add "Duplicate" button to BathhouseList per listing row - calls API and redirects to edit form (FR-032)
- [x] Add listing completeness checklist to BathhouseForm: shows required vs optional fields, blocks submission if incomplete (FR-025). Visual checklist sidebar with check marks
- [x] Add route `/bathhouses/import`
- [x] Write tests for import page, duplicate action, completeness checklist
- [x] Run project test suite + lint

### Task 11: Owner Audit Log & Check-in/Check-out (FR-029, FR-066)

**Files:**
- Create: `frontend/src/pages/bathhouses/AuditLog.tsx`
- Modify: `frontend/src/pages/bookings/BookingList.tsx` (owner)
- Modify: `frontend/src/router.tsx`

- [x] Build AuditLog page: diff-based change history for selected bathhouse, filter by date, shows who/what/when with JSONB diff visualization (FR-029)
- [x] Enhance owner BookingList with explicit check-in/check-out buttons (not just "complete"): "Гость прибыл" button (available from start-15min to start+30min), "Гость ушёл" button (FR-066)
- [x] Add visual check-in status indicators on booking cards
- [x] Add route `/bathhouses/:id/audit`
- [x] Write tests for audit log rendering, check-in/out button visibility logic
- [x] Run project test suite + lint

### Task 12: Owner Calendar Enhancements (FR-076, FR-074)

**Files:**
- Modify: `frontend/src/pages/calendar/CalendarPage.tsx`

- [x] Add consolidated multi-bathhouse view: when owner has multiple bathhouses, show all bookings on one calendar with color coding per bathhouse (FR-076)
- [x] Add day view mode (hourly grid) in addition to existing week/month views
- [x] Enhance external calendar sync UI: connect/disconnect Google Calendar, Yandex Calendar with sync status indicator (FR-074)
- [x] Write tests for multi-bathhouse view toggle, day view rendering
- [x] Run project test suite + lint

### Task 13: Owner BathhouseForm 7-Step Wizard Enhancement (FR-034)

**Files:**
- Modify: `frontend/src/pages/bathhouses/BathhouseForm.tsx`

- [ ] Restructure form into proper 7-step wizard with stepper UI:
  1. Welcome - video/intro and process overview
  2. Object info - name, address (with map marker drag), type, description, capacity, amenities, rules
  3. Photos - upload, drag-to-reorder, cover selection
  4. Pricing - base price with area average recommendation hint
  5. Schedule - working days/hours with template presets ("standard work week")
  6. Cancellation policy - three options with visual comparison
  7. Preview - how card looks to clients
- [ ] Add save draft capability on each step with progress indicator
- [ ] Add back/forward navigation between steps
- [ ] Write tests for step navigation, draft save, preview rendering
- [ ] Run project test suite + lint

### Task 14: Owner Extension Request Handling (FR-065)

**Files:**
- Modify: `frontend/src/pages/bookings/BookingList.tsx` (owner)
- Create: `frontend/src/pages/bookings/ExtensionRequests.tsx`

- [ ] Build ExtensionRequests component (similar to ModificationRequests): shows pending extension requests with requested hours, held amount, approve/reject buttons, 30-min timeout warning
- [ ] Integrate extension request notifications into owner BookingList
- [ ] Write tests for extension request approval/rejection flow
- [ ] Run project test suite + lint

### Task 15: Admin Platform Settings (FR-097 service fee, feature flags, platform settings)

**Files:**
- Create: `frontend/src/pages/admin/PlatformSettings.tsx`
- Create: `frontend/src/pages/admin/FeatureFlags.tsx`
- Create: `frontend/src/pages/admin/ServiceFeeConfig.tsx`
- Modify: `frontend/src/router.tsx`
- Modify: `frontend/src/components/AdminLayout.tsx`

- [ ] Build PlatformSettings page: key-value editor for 12+ platform settings (service_fee_percent, welcome_bonus_amount, escrow_claim_hours, etc.), typed inputs (int/float/string/bool/json)
- [ ] Build FeatureFlags page: toggle list for 13+ feature flags, region scoping selector, enabled/disabled status with description
- [ ] Build ServiceFeeConfig page: fee configuration by region and category, global default, CRUD table
- [ ] Add routes `/admin/settings`, `/admin/feature-flags`, `/admin/service-fees` and sidebar links
- [ ] Write tests for settings editor, feature flag toggles, service fee CRUD
- [ ] Run project test suite + lint

### Task 16: Admin Finance Dashboard (FR-124-128, FR-150-152)

**Files:**
- Create: `frontend/src/pages/admin/FinanceDashboard.tsx`
- Create: `frontend/src/pages/admin/BankReconciliation.tsx`
- Create: `frontend/src/pages/admin/AuditLog.tsx`
- Modify: `frontend/src/router.tsx`
- Modify: `frontend/src/components/AdminLayout.tsx`

- [ ] Build FinanceDashboard with: float monitoring widget (client wallets + owner wallets + escrow totals), daily reconciliation status, GMV/Take Rate/Revenue KPIs, revenue breakdown (service fees, subscriptions, promotions), transaction reconciliation alerts (FR-125, FR-127, FR-152)
- [ ] Build BankReconciliation page: bank statement upload (CSV, 1C XML), auto-matched transactions table, unmatched entries queue with manual match interface (FR-128)
- [ ] Build AuditLog page: searchable log of all admin actions (who/what/when), filter by action type, user, date range
- [ ] Add routes `/admin/finance`, `/admin/finance/reconciliation`, `/admin/audit-log` and sidebar links
- [ ] Write tests for finance widgets, reconciliation upload, audit log filtering
- [ ] Run project test suite + lint

### Task 17: Admin Advanced Analytics (FR-147-154)

**Files:**
- Create: `frontend/src/pages/admin/ConversionFunnels.tsx`
- Create: `frontend/src/pages/admin/CohortAnalysis.tsx`
- Create: `frontend/src/pages/admin/SupplyDemandMetrics.tsx`
- Modify: `frontend/src/router.tsx`

- [ ] Build ConversionFunnels page: visual funnel chart (visit -> search -> view -> book -> complete), filter by region/period/type (FR-147)
- [ ] Build CohortAnalysis page: registration cohort table, retention/frequency/avg check over time (FR-154)
- [ ] Build SupplyDemandMetrics page: active listings, avg occupancy, ADR, new/deactivated listings, DAU/MAU, wallet metrics (FR-148, FR-149, FR-151)
- [ ] Add routes under `/admin/analytics/*` and sidebar sub-menu
- [ ] Write tests for funnel chart, cohort table, metrics rendering
- [ ] Run project test suite + lint

### Task 18: Admin Force Majeure & Mass Operations (FR-080)

**Files:**
- Create: `frontend/src/pages/admin/ForceMajeure.tsx`
- Modify: `frontend/src/pages/admin/BathhouseModeration.tsx`
- Modify: `frontend/src/pages/admin/UserManagement.tsx`
- Modify: `frontend/src/router.tsx`

- [ ] Build ForceMajeure page: activate force majeure with region selector, date range, reason text, preview affected bookings count, confirmation dialog, history of past force majeure events (FR-080)
- [ ] Enhance BathhouseModeration with batch approve/reject (select multiple, act on all) - already partially exists but verify completeness per FR-144
- [ ] Enhance UserManagement with batch block/unblock operations
- [ ] Add route `/admin/force-majeure`
- [ ] Write tests for force majeure activation flow, batch operations
- [ ] Run project test suite + lint

### Task 19: Admin Subscription & Loyalty Management

**Files:**
- Create: `frontend/src/pages/admin/SubscriptionManagement.tsx`
- Create: `frontend/src/pages/admin/LoyaltyManagement.tsx`
- Create: `frontend/src/pages/admin/CertificateManagement.tsx`
- Modify: `frontend/src/router.tsx`

- [ ] Build SubscriptionManagement: view owner subscriptions, tiers, status, manual activation/deactivation
- [ ] Build LoyaltyManagement: configure tiers (bronze/silver/gold/platinum thresholds), cashback percentages, points multipliers
- [ ] Build CertificateManagement: view all certificates, search by code, view redemption history, void certificates
- [ ] Add routes `/admin/subscriptions`, `/admin/loyalty`, `/admin/certificates`
- [ ] Write tests for subscription CRUD, loyalty tier editing, certificate search
- [ ] Run project test suite + lint

### Task 20: Search View Modes & Map Interactions (FR-048-050)

**Files:**
- Modify: `frontend/src/pages/client/BathhouseSearch.tsx`
- Modify: `frontend/src/components/BathhouseMap.tsx`

- [ ] Verify and enhance three view modes (List / Map / Split-view) per FR-050
- [ ] Add map cluster behavior when >50 markers (FR-048)
- [ ] Add "Search in this area" button on map pan (FR-049)
- [ ] Add hover synchronization: hover on list item highlights map marker and vice versa (FR-049)
- [ ] Add mini-card popup on marker click (photo, name, price, rating) with link to full page (FR-049)
- [ ] Write tests for view mode switching, map interactions
- [ ] Run project test suite + lint

### Task 21: Booking Flow Stepper Enhancement (FR-055-058)

**Files:**
- Modify: `frontend/src/pages/client/BookingCreate.tsx`

- [ ] Restructure into stepped wizard: Step 1 (Date/Time/Duration/Guests) -> Step 2 (Add-ons selection) -> Step 3 (Promo/Certificate/Wallet) -> Step 4 (Price breakdown + Payment method) per FR-055
- [ ] Add payment method selection with visual cards: full wallet / card / SBP / Apple Pay / Google Pay / combo (FR-057)
- [ ] Add combo payment slider: drag to choose wallet vs card split amount (FR-093)
- [ ] Show clear breakdown per source: "X from wallet + Y from card" (FR-057)
- [ ] Handle request-mode bookings: show "request" flow with hold explanation (FR-058)
- [ ] Handle concurrent slot conflict: show "Slot just taken" message with nearest alternatives (FR-061)
- [ ] Write tests for stepper flow, combo payment, request mode
- [ ] Run project test suite + lint

### Task 22: Push Notification Permission & Onboarding (FR-016, FR-142)

**Files:**
- Modify: `frontend/src/components/OnboardingTour.tsx`
- Modify: `frontend/src/lib/useDeviceToken.ts`
- Modify: `frontend/src/pages/client/BookingDetail.tsx`

- [ ] Implement push permission request after first booking completion (not at registration) per FR-142
- [ ] Verify OnboardingTour shows: how to search, how to book, what is wallet, nearby recommendations (FR-016)
- [ ] Add welcome bonus display in onboarding (500 RUB / 15 BYN) per FR-016
- [ ] Write tests for push permission timing, onboarding steps
- [ ] Run project test suite + lint

### Task 23: Client Active Promo Codes & Chat Content Filter (FR-099, FR-063)

**Files:**
- Create: `frontend/src/pages/client/ActivePromoCodes.tsx`
- Modify: `frontend/src/router.tsx`
- Modify: `frontend/src/pages/chat/MessageArea.tsx`

- [ ] Build ActivePromoCodes page: list of available promo codes for the client with conditions (min amount, valid until, discount type), copy code button
- [ ] Add route `/client/promos`
- [ ] Verify chat message area handles blocked content (phone numbers, emails, URLs) per FR-063 - show warning when message is filtered
- [ ] Write tests for promo list rendering, chat filter warning display
- [ ] Run project test suite + lint

### Task 24: Admin Support Operations Metrics

**Files:**
- Modify: `frontend/src/pages/admin/TicketManagement.tsx`
- Modify: `frontend/src/pages/admin/AdminDashboard.tsx`

- [ ] Add support metrics dashboard widget to AdminDashboard: FCR%, AHT, SLA compliance, queue size
- [ ] Enhance TicketManagement with: per-agent throughput stats, SLA countdown timers, escalation indicators
- [ ] Write tests for metrics widget, SLA timer display
- [ ] Run project test suite + lint

### Task 25: Verify Acceptance Criteria & Final Integration

- [ ] Manual test: complete client registration -> search -> book -> pay -> review flow
- [ ] Manual test: owner onboarding -> create listing -> manage bookings -> finance -> payout flow
- [ ] Manual test: admin moderation -> settings -> analytics -> force majeure flow
- [ ] Run full test suite: `cd frontend && npx vitest run`
- [ ] Run linter: `cd frontend && npm run lint`
- [ ] Run build: `cd frontend && npm run build`
- [ ] Verify all new routes are accessible from navigation menus
- [ ] Cross-reference every FR in BRD sections 2.1-2.18 against implemented pages

### Task 26: Update Documentation

- [ ] Update CLAUDE.md Frontend Structure section with new pages
- [ ] Update CLAUDE.md Frontend Key Patterns with any new patterns introduced
- [ ] Move this plan to `docs/plans/completed/`
