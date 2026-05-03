# Frontend BRD Gap Implementation Plan

## Overview

Comprehensive gap analysis and implementation plan to bring the React frontend into full compliance with BRD RelaxHub v6. The analysis covers all three roles (client, owner, admin) across 116+ functional requirements.

## Context

- Files involved: `frontend/src/pages/`, `frontend/src/components/`, `frontend/src/lib/`, `frontend/src/router.tsx`
- Related patterns: RelaxHUB Design System (`@/components/design`), TanStack React Query (orval-generated), zustand stores, role-based routing
- Dependencies: Generated API client (orval), Yandex Maps API, WebSocket, dayjs

## Gap Summary

### Critical Missing Features (not implemented at all)
1. Phone+OTP registration/login (FR-001) - Login only supports email, no phone/OTP flow
2. Password reset flow (FR-015) - no "Forgot password" page
3. 2FA partial token login flow (FR-007) - Login doesn't handle 2FA challenge response
4. Evidence upload in dispute creation (FR-spor) - DisputeCreate has no file/photo upload
5. Multi-criteria review ratings (FR-132) - ReviewForm uses single Rate, not 4 criteria (cleanliness, accuracy, communication, value)
6. Calendar views: day/week/month (FR-076) - CalendarPage only has weekly view, missing day/month/svodny
7. Booking modification request from client side (FR-067) - no UI to modify a booking
8. Booking extension request from client side (FR-065) - no UI to extend a booking
9. Owner dashboard (FR-2.15) - Dashboard.tsx is minimal, missing today's bookings, rating, wallet balance, quick actions
10. Isochrone search UI (FR-053) - BathhouseSearch has placeholder for isochrone but no actual integration
11. ForgotPassword page - missing entirely

### Significantly Incomplete Pages (exist but missing BRD features)
12. BathhouseDetail - missing: transport accessibility block, similar bathhouses block, cancellation policy display, booking mode indicator, area average price comparison, full price breakdown preview, owner profile block
13. BookingDetail (client) - missing: modification request, extension request, re-book button, share button, dispute button based on timing, route/navigation button
14. BookingCreate - missing: certificate code redemption, saved card selection, re-booking pre-fill from query params, 3DS redirect handling
15. BathhouseSearch - missing: search suggestions/autocomplete (FR-039), "search in this area" button on map, marker clustering, marker price labels, hover highlight sync between list and map, date/time availability filter
16. ClientHome - missing: OnboardingTour integration for first-time users, action banners for promotions
17. WalletDashboard - missing: bonus expiry info display, CSV/PDF export
18. ClientProfile - missing: region switch feature (FR-017), account deletion request (FR-014)
19. Register - missing: phone+OTP option, age confirmation checkbox (FR-005)
20. DisputeDetail (client) - missing: evidence upload during 72h window, appeal button after resolution
21. DisputeCreate - missing: evidence upload (photos/screenshots/GPS)
22. CalendarPage (owner) - missing: day view, month view, multi-bathhouse svodny view, manual slot blocking
23. Owner BookingList - missing: check-in/check-out buttons (FR-066)
24. Owner BookingDetails - missing: check-in/check-out flow, no-show handling
25. PricingRules - missing: seasonal tariff management (FR-087), holiday pricing per bathhouse (FR-088), last-minute discount settings (FR-090), smart pricing recommendations display (FR-089)
26. RepresentativeList - missing: role selection (manager/observer), per-bathhouse access control
27. FinanceDashboard (owner) - missing: auto-payout threshold setting (FR-102), acts download
28. ReviewList (owner) - missing: owner response form within the list
29. BathhouseForm - missing: step 1 video/welcome, step 5 schedule templates, step 7 client-side preview, address drag-to-correct map interaction (FR-027)
30. GuestCardDetail (CRM) - missing: notes, tags, LTV display
31. BroadcastCreate (CRM) - missing: personalization tokens ({{guest_name}}, etc.), image attachment, SMS channel option
32. Login - missing: phone/OTP tab, 2FA challenge handling, forgot password link

### Admin Pages Needing Enhancement
33. AdminDashboard - missing: support ops metrics widget (FCR, AHT, SLA)
34. DisputeManagement - missing: mediator assignment, evidence review inline
35. WalletManagement - missing: freeze/unfreeze wallet
36. BookingManagement - missing: manual refund action with comment
37. UserManagement - missing: batch operations (block/unblock multiple), user detail modal with full profile
38. BathhouseModeration - missing: batch operations, SLA countdown display
39. RoleManagement - missing: 27-permission matrix display, 2FA enforcement indicator
40. AdminNotificationCenter - missing: severity-based filtering, email digest settings

### Missing Components
41. ForgotPassword page - new page needed
42. TwoFactorChallenge component - for login 2FA flow
43. PhoneOTPInput component - for phone registration/login
44. BookingModificationModal - client-side booking modification
45. BookingExtensionModal - client-side extension request
46. EvidenceUploader component - for dispute evidence (photo/GPS/screenshot)
47. TransportAccessibility component - for bathhouse detail page
48. SimilarBathhouses component - for bathhouse detail page
49. PriceBreakdown component - reusable price breakdown display
50. SearchSuggestions component - autocomplete dropdown for search
51. OwnerReviewResponse component - inline review response form

## Development Approach

- **Testing approach**: Regular (code first, then tests)
- Complete each task fully before moving to the next
- Follow existing patterns: RelaxHUB Design System (`@/components/design`), orval-generated hooks, zustand stores
- **CRITICAL: every task MUST include new/updated tests**
- **CRITICAL: all tests must pass before starting next task**

## Implementation Steps

### Task 1: Auth - Phone/OTP Registration + Login + Forgot Password + 2FA Challenge

**Files:**
- Create: `frontend/src/pages/ForgotPassword.tsx`
- Create: `frontend/src/components/PhoneOTPInput.tsx`
- Create: `frontend/src/components/TwoFactorChallenge.tsx`
- Modify: `frontend/src/pages/Login.tsx`
- Modify: `frontend/src/pages/Register.tsx`
- Modify: `frontend/src/router.tsx`

- [x] Create PhoneOTPInput component (phone input + "Send OTP" button + 6-digit code input + countdown timer)
- [x] Create TwoFactorChallenge component (TOTP code input + SMS fallback, used when login returns partial_token)
- [x] Update Login: add tab switcher email/phone, phone+OTP login flow, 2FA challenge handling, "Forgot password" link
- [x] Update Register: add tab switcher email/phone, phone+OTP registration, age confirmation checkbox (FR-005), password field only for email
- [x] Create ForgotPassword page (email input -> send reset link -> success message)
- [x] Add /forgot-password route to router.tsx
- [x] Write tests for Login, Register, ForgotPassword, PhoneOTPInput, TwoFactorChallenge
- [x] Run project test suite - must pass before task 2

### Task 2: Client Home + Onboarding Tour

**Files:**
- Modify: `frontend/src/pages/client/ClientHome.tsx`
- Modify: `frontend/src/components/OnboardingTour.tsx`

- [x] Integrate OnboardingTour trigger for first-time users (check onboarding_completed flag)
- [x] Add city selection when GPS unavailable (FR-036)
- [x] Improve PromoBanner to show actual active promotions from API
- [x] Write tests for ClientHome onboarding integration
- [x] Run project test suite - must pass before task 3

### Task 3: Search - Autocomplete, Filters, Map Interaction

**Files:**
- Create: `frontend/src/components/SearchSuggestions.tsx`
- Modify: `frontend/src/pages/client/BathhouseSearch.tsx`
- Modify: `frontend/src/components/BathhouseMap.tsx`
- Modify: `frontend/src/components/BathhouseCard.tsx`

- [x] Create SearchSuggestions dropdown (popular queries, bathhouse names, cities from API with debounce 300ms)
- [x] Add date+time filter to search (FR-037) - DatePicker + TimePicker that filters by available slots
- [x] Add "Search in this area" button on map pan (FR-049)
- [x] Add marker price labels on map markers
- [x] Add marker clustering for 50+ results (FR-048)
- [x] Add hover highlight sync: list item <-> map marker (FR-049)
- [x] Add booking type filter (instant/request) (FR-040)
- [x] Add minimum rating filter (FR-040)
- [x] Add status filter (verified/top/premium) (FR-040)
- [x] Add last-minute badge on cards with discount (FR-090)
- [x] Write tests for SearchSuggestions, updated BathhouseSearch filters and map interactions
- [x] Run project test suite - must pass before task 4

### Task 4: Bathhouse Detail - Complete Feature Set

**Files:**
- Create: `frontend/src/components/TransportAccessibility.tsx`
- Create: `frontend/src/components/SimilarBathhouses.tsx`
- Create: `frontend/src/components/PriceBreakdown.tsx`
- Modify: `frontend/src/pages/client/BathhouseDetail.tsx`

- [x] Create TransportAccessibility component (nearby metro/bus/parking from API) (FR-054)
- [x] Create SimilarBathhouses component (up to 6 similar, from API) (FR-044)
- [x] Create PriceBreakdown component (base price, add-ons, service fee, discounts, area average comparison, total) (FR-044)
- [x] Add owner profile block (rating, objects count, registration date) (FR-044)
- [x] Add cancellation policy display (FR-022)
- [x] Add booking mode indicator (instant/request) (FR-058)
- [x] Add area average price comparison (FR-044)
- [x] Add full price breakdown preview for selected slot (FR-044)
- [x] Add visiting rules display (FR-021)
- [x] Write tests for all new components and updated BathhouseDetail
- [x] Run project test suite - must pass before task 5

### Task 5: Booking Create - Certificate, Saved Cards, Re-booking

**Files:**
- Modify: `frontend/src/pages/client/BookingCreate.tsx`

- [ ] Add certificate code input and balance check (FR-056)
- [ ] Add saved card selection from user's saved cards (FR-004)
- [ ] Add re-booking pre-fill from query params (date excluded, rest pre-filled) (FR-059)
- [ ] Add hold indicator for request-based bookings (FR-058)
- [ ] Ensure full price breakdown shows all line items: base, add-ons per item, long session discount, extra guest surcharge, holiday surcharge, seasonal tariff, service fee, promo discount, certificate, wallet, total (FR-055)
- [ ] Write tests for certificate flow, saved cards, re-booking pre-fill
- [ ] Run project test suite - must pass before task 6

### Task 6: Booking Detail + Modification + Extension + Re-book

**Files:**
- Create: `frontend/src/components/BookingModificationModal.tsx`
- Create: `frontend/src/components/BookingExtensionModal.tsx`
- Modify: `frontend/src/pages/client/BookingDetail.tsx`
- Modify: `frontend/src/pages/client/BookingList.tsx`

- [ ] Create BookingModificationModal (change date/time/duration/guests/addons, max 3 per booking) (FR-067)
- [ ] Create BookingExtensionModal (request 1-2h extension) (FR-065)
- [ ] Add "Modify" button to BookingDetail (visible if booking is modifiable)
- [ ] Add "Extend" button to BookingDetail (visible during active booking)
- [ ] Add "Re-book" button on completed bookings (FR-059) linking to BookingCreate with pre-filled params
- [ ] Add "Share booking" button (FR-081)
- [ ] Add "Open dispute" button with timing logic (only after check-out, within 72h) (FR-spor)
- [ ] Add cancellation with wallet/card refund choice and bonus display (FR-068)
- [ ] Add route/navigation button (link to Yandex Maps)
- [ ] Add booking status timeline display
- [ ] Update BookingList: re-book button on completed, status filters
- [ ] Write tests for BookingModificationModal, BookingExtensionModal, updated BookingDetail
- [ ] Run project test suite - must pass before task 7

### Task 7: Review Form - Multi-Criteria Ratings

**Files:**
- Modify: `frontend/src/pages/client/ReviewForm.tsx`

- [ ] Replace single Rate with 4 criteria: cleanliness, accuracy, communication, value_for_money (FR-132)
- [ ] Add 0.5 step to each rating slider (FR-132)
- [ ] Show calculated overall rating as average
- [ ] Add completeness indicator encouraging text review
- [ ] Write tests for multi-criteria ReviewForm
- [ ] Run project test suite - must pass before task 8

### Task 8: Dispute System - Evidence Upload + Appeal

**Files:**
- Create: `frontend/src/components/EvidenceUploader.tsx`
- Modify: `frontend/src/pages/client/DisputeCreate.tsx`
- Modify: `frontend/src/pages/client/DisputeDetail.tsx`

- [ ] Create EvidenceUploader component (photo/screenshot upload + GPS location capture + type selector)
- [ ] Update DisputeCreate: integrate EvidenceUploader, show 72h evidence window timer
- [ ] Update DisputeDetail: add evidence upload during 72h window, appeal button (within 7 days after resolution), show evidence timeline, show resolution details
- [ ] Write tests for EvidenceUploader, updated DisputeCreate and DisputeDetail
- [ ] Run project test suite - must pass before task 9

### Task 9: Wallet - Export + Bonus Expiry Display

**Files:**
- Modify: `frontend/src/pages/client/WalletDashboard.tsx`

- [ ] Add bonus expiry info display: show nearest expiring bonuses with dates and amounts (FR-122)
- [ ] Add CSV and PDF export buttons for wallet history (FR-115)
- [ ] Add filter by transaction type (FR-115)
- [ ] Add frozen/held amount display (FR-113)
- [ ] Write tests for updated WalletDashboard
- [ ] Run project test suite - must pass before task 10

### Task 10: Client Profile - Region Switch + Account Deletion

**Files:**
- Modify: `frontend/src/pages/client/ClientProfile.tsx`

- [ ] Add region switch section (BY <-> RU) with blocking condition checks (balance, bookings, disputes, certificates) (FR-017)
- [ ] Add account deletion request section with 30-day grace period explanation (FR-014)
- [ ] Add profile completeness indicator linking to missing fields
- [ ] Write tests for region switch and account deletion flows
- [ ] Run project test suite - must pass before task 11

### Task 11: Owner Dashboard - Complete

**Files:**
- Modify: `frontend/src/pages/Dashboard.tsx`

- [ ] Add today's bookings summary with quick actions
- [ ] Add revenue widget (today/week/month)
- [ ] Add current rating display
- [ ] Add wallet balance with quick payout link
- [ ] Add notifications/alerts section
- [ ] Add CRM summary (new guests, pending messages)
- [ ] Add empty states with CTAs for each section (FR-2.15)
- [ ] Write tests for enhanced Dashboard
- [ ] Run project test suite - must pass before task 12

### Task 12: Owner Calendar - Day/Month/Svodny Views + Slot Blocking

**Files:**
- Modify: `frontend/src/pages/calendar/CalendarPage.tsx`

- [ ] Add day view (hourly grid) (FR-076)
- [ ] Add month view (FR-076)
- [ ] Add color coding by status: confirmed=green, pending=yellow, cancelled=red, blocked=grey (FR-076)
- [ ] Add multi-bathhouse svodny view for owners with multiple objects (FR-076)
- [ ] Add manual slot blocking with reason (FR-072)
- [ ] Write tests for new calendar views and slot blocking
- [ ] Run project test suite - must pass before task 13

### Task 13: Owner Bookings - Check-in/Check-out + No-show

**Files:**
- Modify: `frontend/src/pages/bookings/BookingList.tsx`
- Modify: `frontend/src/pages/bookings/BookingDetails.tsx`
- Modify: `frontend/src/pages/bookings/ExtensionRequests.tsx`
- Modify: `frontend/src/pages/bookings/ModificationRequests.tsx`

- [ ] Add "Guest arrived" button (check-in window: start-15min to start+30min) (FR-066)
- [ ] Add "Guest left" button (check-out) (FR-066)
- [ ] Add no-show indicator with timing (FR-070)
- [ ] Add request-based booking approve/reject actions (FR-058)
- [ ] Add push reminder 5 min before start (FR-066)
- [ ] Update ExtensionRequests: approve/reject with 30min timeout display
- [ ] Update ModificationRequests: approve/reject with price diff display, 24h timeout
- [ ] Write tests for check-in/check-out flow and request handling
- [ ] Run project test suite - must pass before task 14

### Task 14: Pricing - Seasonal Tariffs + Holiday + Last-Minute + Smart Pricing

**Files:**
- Modify: `frontend/src/pages/pricing/PricingRules.tsx`

- [ ] Add seasonal tariff CRUD section (date range + multiplier) (FR-087)
- [ ] Add per-bathhouse holiday pricing multiplier settings (FR-088)
- [ ] Add last-minute discount settings (threshold hours + discount %) (FR-090)
- [ ] Add smart pricing recommendation display (show recommended coefficient 0.8-1.5) (FR-089)
- [ ] Write tests for new pricing sections
- [ ] Run project test suite - must pass before task 15

### Task 15: Owner BathhouseForm - Wizard Completion

**Files:**
- Modify: `frontend/src/pages/bathhouses/BathhouseForm.tsx`

- [ ] Step 1: Add video/intro overview section (FR-034 step 1)
- [ ] Step 4: Add area average price hint (FR-034 step 4)
- [ ] Step 5: Add schedule templates ("standard work week" etc.) (FR-034 step 5)
- [ ] Step 7: Add client-facing preview rendering (FR-034 step 7)
- [ ] Add address map with draggable marker for correction (FR-027)
- [ ] Write tests for wizard improvements
- [ ] Run project test suite - must pass before task 16

### Task 16: Owner Representatives - Sub-roles + Per-Bathhouse Access

**Files:**
- Modify: `frontend/src/pages/representatives/RepresentativeList.tsx`

- [ ] Add role selection per representative: manager (manage bookings, reply) / observer (read-only) (FR-2.15)
- [ ] Add per-bathhouse access control (select which bathhouses the rep can access)
- [ ] Add invite-by-email flow
- [ ] Add revoke access action
- [ ] Write tests for representative sub-roles
- [ ] Run project test suite - must pass before task 17

### Task 17: Owner Finance - Auto-payout + Acts + Reports Export

**Files:**
- Modify: `frontend/src/pages/finance/FinanceDashboard.tsx`
- Modify: `frontend/src/pages/finance/PayoutPage.tsx`
- Modify: `frontend/src/pages/finance/FinancialReports.tsx`

- [ ] Add auto-payout threshold setting (FR-102)
- [ ] Add SBP instant payout indicator (FR-101)
- [ ] Add acts download (PDF, XML for legal entities) (FR-104)
- [ ] Add payout history with status tracking
- [ ] Add frozen amount display during claim period
- [ ] Write tests for finance enhancements
- [ ] Run project test suite - must pass before task 18

### Task 18: Owner Reviews - Response Form + Bidirectional Reviews

**Files:**
- Modify: `frontend/src/pages/reviews/ReviewList.tsx`

- [ ] Add inline owner response form (one response per review) (FR-135)
- [ ] Add bidirectional review section: owner rates client (punctuality, cleanliness, rule_compliance) (FR-130)
- [ ] Add blind period indicator (reviews hidden until both posted or 14 days) (FR-131)
- [ ] Write tests for review response and bidirectional review
- [ ] Run project test suite - must pass before task 19

### Task 19: CRM - Guest Card + Broadcast + Auto-scenarios Completion

**Files:**
- Modify: `frontend/src/pages/crm/GuestCardDetail.tsx`
- Modify: `frontend/src/pages/crm/BroadcastCreate.tsx`
- Modify: `frontend/src/pages/crm/AutoScenarios.tsx`

- [ ] GuestCardDetail: add notes field, tags management, LTV display, visit history, avg check (FR-CRM)
- [ ] BroadcastCreate: add personalization tokens ({{guest_name}}, {{last_visit_date}}, etc.), image attachment, SMS channel option, preview (FR-CRM)
- [ ] AutoScenarios: add per-scenario customization (text, timing, channel), enable/disable toggle (FR-CRM)
- [ ] Write tests for CRM enhancements
- [ ] Run project test suite - must pass before task 20

### Task 20: Admin Dashboard - Support Ops Metrics

**Files:**
- Modify: `frontend/src/pages/admin/AdminDashboard.tsx`

- [ ] Add support operations metrics widget: FCR, AHT, SLA compliance (FR-153 operational)
- [ ] Add P&L/unit economics widget (FR-150)
- [ ] Add wallet metrics widget (FR-151)
- [ ] Write tests for admin dashboard enhancements
- [ ] Run project test suite - must pass before task 21

### Task 21: Admin Dispute + Moderation Enhancements

**Files:**
- Modify: `frontend/src/pages/admin/DisputeManagement.tsx`
- Modify: `frontend/src/pages/admin/AdminDisputeDetail.tsx`
- Modify: `frontend/src/pages/admin/BathhouseModeration.tsx`
- Modify: `frontend/src/pages/admin/UserManagement.tsx`
- Modify: `frontend/src/pages/admin/WalletManagement.tsx`
- Modify: `frontend/src/pages/admin/BookingManagement.tsx`

- [ ] DisputeManagement: add mediator assignment dropdown
- [ ] AdminDisputeDetail: add evidence gallery review, resolution form (full/partial/no refund + compensation), appeal review
- [ ] BathhouseModeration: add SLA countdown (48h), batch operations, moderation queue metrics
- [ ] UserManagement: add batch block/unblock, user detail modal with full profile, KYC status
- [ ] WalletManagement: add freeze/unfreeze wallet action
- [ ] BookingManagement: add manual refund with required comment and audit log
- [ ] Write tests for admin enhancements
- [ ] Run project test suite - must pass before task 22

### Task 22: Admin Roles + Notification Center Enhancements

**Files:**
- Modify: `frontend/src/pages/admin/RoleManagement.tsx`
- Modify: `frontend/src/pages/admin/AdminNotificationCenter.tsx`

- [ ] RoleManagement: display 27-permission matrix for 6 admin sub-roles, 2FA enforcement indicator (FR-2.13)
- [ ] AdminNotificationCenter: add severity-based filtering (info/warning/critical/urgent), email digest toggle
- [ ] Write tests for admin role and notification enhancements
- [ ] Run project test suite - must pass before task 23

### Task 23: Isochrone Search + Geolocation

**Files:**
- Modify: `frontend/src/pages/client/BathhouseSearch.tsx`
- Modify: `frontend/src/components/BathhouseMap.tsx`

- [ ] Add isochrone search UI: transport mode selector (car/transit) + minutes slider (FR-053)
- [ ] Display isochrone polygon on map (GeoJSON from API)
- [ ] Add GPS location detection with fallback to city picker (FR-036)
- [ ] Write tests for isochrone integration
- [ ] Run project test suite - must pass before task 24

### Task 24: Verify Acceptance Criteria

- [ ] Manual test: full client booking flow (search -> detail -> book -> pay -> view in history)
- [ ] Manual test: owner onboarding flow (register -> KYC -> offer -> wizard -> submit)
- [ ] Manual test: admin moderation flow (queue -> approve/reject -> notification)
- [ ] Manual test: dispute flow (create with evidence -> admin review -> resolve -> appeal)
- [ ] Run full test suite (`cd frontend && npx vitest run`)
- [ ] Run linter (`cd frontend && npm run lint`)
- [ ] Verify test coverage meets 80%+

### Task 25: Update Documentation

- [ ] Update CLAUDE.md if internal patterns changed (new components, new pages)
- [ ] Move this plan to `docs/plans/completed/`
