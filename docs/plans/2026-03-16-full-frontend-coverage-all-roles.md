# Full Frontend Coverage - All Roles (Owner, Client, Admin)

## Overview

Extend the existing Owner Dashboard SPA to cover 100% of the backend API for all three roles: client (search, book, review), owner/representative (existing + enhancements), admin (moderation, management). Role-based layouts with separate navigation per role.

## Context

- Files involved: `frontend/src/` (router, components, pages, stores, lib)
- Related patterns: Ant Design 6, React Query (orval-generated hooks), Zustand stores, role-based guards
- Dependencies: All API hooks already generated in `src/api/generated/`

## Development Approach

- **Testing approach**: Regular (code first, then tests)
- Complete each task fully before moving to the next
- Reuse existing patterns (BathhouseSelector, AppLayout sidebar, formatPrice, etc.)
- **CRITICAL: every task MUST include new/updated tests**
- **CRITICAL: all tests must pass before starting next task**

## Implementation Steps

### Task 1: Multi-role architecture foundation

**Files:**
- Modify: `frontend/src/router.tsx`
- Modify: `frontend/src/components/ProtectedRoute.tsx`
- Create: `frontend/src/components/ClientLayout.tsx`
- Create: `frontend/src/components/AdminLayout.tsx`
- Modify: `frontend/src/components/AppLayout.tsx` (rename concept to OwnerLayout)
- Modify: `frontend/src/stores/auth.ts` (role-based redirects)

- [x] Add role-based route groups: `/client/*`, `/admin/*`, existing owner routes stay at `/`
- [x] Create ClientLayout with client-specific navigation (search, bookings, favorites, loyalty, chat, profile)
- [x] Create AdminLayout with admin-specific navigation (dashboard, users, bathhouses, reviews, photos, complaints, cities, promos)
- [x] Update ProtectedRoute to support role restrictions (allowedRoles prop)
- [x] Add role-based redirect after login (client -> /client, owner -> /, admin -> /admin)
- [x] Write tests for role-based routing and redirects
- [x] Run project test suite - must pass before task 2

### Task 2: Client - Public bathhouse search and detail

**Files:**
- Create: `frontend/src/pages/client/BathhouseSearch.tsx`
- Create: `frontend/src/pages/client/BathhouseDetail.tsx`
- Create: `frontend/src/components/BathhouseCard.tsx`

- [x] BathhouseSearch page: search with filters (city, price range, amenities, rating), geo-search, pagination, sorting
- [x] BathhouseCard component: photo, name, rating, price, city, amenities preview, favorite toggle
- [x] BathhouseDetail page: full info, photo gallery, reviews list, available slots, similar bathhouses, JSON-LD schema
- [x] API hooks: `useGetBathhouses`, `useGetBathhousesId`, `useGetBathhousesIdPhotos`, `useGetBathhousesIdReviews`, `useGetBathhousesIdSimilar`, `useGetBathhousesIdAvailableSlots`, `useGetBathhousesIdGallery`, `useGetCities`, `usePostBathhousesIdFavorite`
- [x] Write tests for search filters, card rendering, detail page
- [x] Run project test suite - must pass before task 3

### Task 3: Client - Booking creation flow

**Files:**
- Create: `frontend/src/pages/client/BookingCreate.tsx`
- Create: `frontend/src/pages/client/BookingList.tsx`
- Create: `frontend/src/pages/client/BookingDetail.tsx`

- [x] BookingCreate: date picker, slot selection, promo code input with validation, gift certificate input, loyalty points toggle, referral balance toggle, price calculator display, payment initiation
- [x] BookingList: client's own bookings with status filters, cancel action, payment status
- [x] BookingDetail: full booking info with payment status, cancel button (with refund policy display)
- [x] API hooks: `useGetBathhousesIdAvailableSlots`, `useGetBathhousesIdPriceCalculator`, `usePostBookings`, `useGetBookings`, `usePatchBookingsIdCancel`, `usePostBookingsIdPay`, `useGetBookingsIdPayment`, `usePostPromoCodesValidate`, `useGetCertificatesCodeBalance`
- [x] Write tests for booking flow, promo validation, payment display
- [x] Run project test suite - must pass before task 4

### Task 4: Client - Reviews with media

**Files:**
- Create: `frontend/src/pages/client/ReviewForm.tsx`
- Create: `frontend/src/components/ReviewCard.tsx`
- Create: `frontend/src/components/MediaUploader.tsx`

- [x] ReviewForm: rating input (1-5 stars), text, photo/video upload (max 10 photos, 1 video)
- [x] ReviewCard reusable component: rating, text, media gallery, owner response, report button
- [x] Media upload with type/size validation, thumbnail preview
- [x] API hooks: `usePostBathhousesIdReviews`, `usePutReviewsId`, `useDeleteReviewsId`, `usePostReviewsIdMedia`, `useDeleteMediaId`, `usePostReviewsIdReport`
- [x] Write tests for review creation, media upload validation
- [x] Run project test suite - must pass before task 5

### Task 5: Client - Favorites and recommendations

**Files:**
- Create: `frontend/src/pages/client/Favorites.tsx`
- Create: `frontend/src/pages/client/Recommendations.tsx`
- Create: `frontend/src/pages/client/Preferences.tsx`

- [x] Favorites page: grid of favorite bathhouses with unfavorite toggle
- [x] Recommendations page: personalized recommendations grid, popular bathhouses section
- [x] Preferences page: user preference settings that improve recommendations
- [x] API hooks: `useGetMyFavorites`, `usePostBathhousesIdFavorite`, `useGetRecommendations`, `useGetPopular`, `useGetMyPreferences`, `usePutMyPreferences`
- [x] Write tests for favorites toggle, recommendations display
- [x] Run project test suite - must pass before task 6

### Task 6: Client - Loyalty program

**Files:**
- Create: `frontend/src/pages/client/LoyaltyDashboard.tsx`

- [x] Loyalty dashboard: current level (bronze/silver/gold/platinum) with progress bar to next level, points balance, privileges list
- [x] Transactions history table: earned/spent points, timestamps, reasons
- [x] Level info cards showing requirements and benefits for each tier
- [x] API hooks: `useGetMyLoyalty`, `useGetMyLoyaltyTransactions`, `useGetMyLoyaltyLevels`
- [x] Write tests for level display, transaction list, progress calculation
- [x] Run project test suite - must pass before task 7

### Task 7: Client - Referral program

**Files:**
- Create: `frontend/src/pages/client/ReferralProgram.tsx`

- [x] Personal referral code with copy-to-clipboard and share buttons
- [x] Referral stats: invited count, completed bookings count
- [x] Referral balance display with usage on bookings
- [x] API hooks: `useGetMyReferral`, `useGetMyReferralStats`, `useGetMyReferralBalance`
- [x] Write tests for referral code display, stats rendering
- [x] Run project test suite - must pass before task 8

### Task 8: Client - Gift certificates

**Files:**
- Create: `frontend/src/pages/client/CertificateList.tsx`
- Create: `frontend/src/pages/client/CertificatePurchase.tsx`

- [x] CertificateList: owned certificates with code, balance, expiry, status
- [x] CertificatePurchase: amount input, purchase flow (works without auth too)
- [x] Certificate redemption UI (redeem code input, balance check)
- [x] API hooks: `useGetMyCertificates`, `usePostCertificatesPurchase`, `usePostCertificatesRedeem`, `useGetCertificatesCodeBalance`
- [x] Write tests for purchase flow, balance display, expiry handling
- [x] Run project test suite - must pass before task 9

### Task 9: Client - Payments history

**Files:**
- Create: `frontend/src/pages/client/PaymentHistory.tsx`

- [x] Payment history table: date, amount, status, booking link, refund info
- [x] Status badges: pending, succeeded, refunded, cancelled
- [x] Filter by date range and status
- [x] API hooks: `useGetMyPayments`
- [x] Write tests for payment list, status display, filtering
- [x] Run project test suite - must pass before task 10

### Task 10: Client - User profile and stats

**Files:**
- Create: `frontend/src/pages/client/ClientProfile.tsx`

- [x] Profile settings (reuse ProfileSettings pattern): name, email, phone, avatar
- [x] User statistics: total bookings, reviews count
- [x] Social account linking (VK, Yandex, Google) with OAuth flow
- [x] Notification preferences
- [x] API hooks: `useGetAuthMe`, `usePutAuthMe`, `usePostAuthMeAvatar`, `useDeleteAuthMeAvatar`, `useGetMyStats`, `useGetAuthMeSocialAccounts`, `usePostAuthLinkProvider`, `useDeleteAuthLinkProvider`, `useGetMyNotificationPreferences`, `usePutMyNotificationPreferences`
- [x] Write tests for profile edit, social link/unlink
- [x] Run project test suite - must pass before task 11

### Task 11: Admin - Dashboard and user management

**Files:**
- Create: `frontend/src/pages/admin/AdminDashboard.tsx`
- Create: `frontend/src/pages/admin/UserManagement.tsx`

- [x] Admin dashboard: KPI analytics (bookings, revenue, users, bathhouses), top bathhouses table
- [x] User management: searchable user table, block/unblock actions with confirmation
- [x] API hooks: `useGetAdminAnalytics`, `useGetAdminAnalyticsTop`, `useGetAdminUsers`, `usePatchAdminUsersIdBlock`, `usePatchAdminUsersIdUnblock`
- [x] Write tests for analytics display, user block/unblock flow
- [x] Run project test suite - must pass before task 12

### Task 12: Admin - Bathhouse moderation

**Files:**
- Create: `frontend/src/pages/admin/BathhouseModeration.tsx`

- [x] Bathhouse list with status filter (pending/active/rejected/inactive)
- [x] Detail view with all bathhouse info
- [x] Approve/reject actions with confirmation
- [x] API hooks: `useGetAdminBathhouses`, `usePatchAdminBathhousesIdApprove`, `usePatchAdminBathhousesIdReject`
- [x] Write tests for moderation actions, status filtering
- [x] Run project test suite - must pass before task 13

### Task 13: Admin - Review moderation

**Files:**
- Create: `frontend/src/pages/admin/ReviewModeration.tsx`

- [x] Review list with status filter, pending count badge
- [x] Review detail with media, user info, bathhouse info
- [x] Approve/reject single and batch actions
- [x] API hooks: `useGetAdminReviews`, `useGetAdminReviewsPendingCount`, `usePatchAdminReviewsIdApprove`, `usePatchAdminReviewsIdReject`, `usePostAdminReviewsBatchApprove`, `usePostAdminReviewsBatchReject`
- [x] Write tests for single and batch moderation flows
- [x] Run project test suite - must pass before task 14

### Task 14: Admin - Photo verification

**Files:**
- Create: `frontend/src/pages/admin/PhotoVerification.tsx`

- [x] Pending photos grid with bathhouse context
- [x] Full-size photo lightbox view
- [x] Verify/reject actions per photo
- [x] API hooks: `useGetAdminPhotosPending`, `usePatchAdminPhotosIdVerify`, `usePatchAdminPhotosIdReject`
- [x] Write tests for verification flow, photo display
- [x] Run project test suite - must pass before task 15

### Task 15: Admin - Complaint management

**Files:**
- Create: `frontend/src/pages/admin/ComplaintManagement.tsx`

- [x] Complaint list with type filter (spam, offensive, fake, fraud, other) and status filter
- [x] Complaint detail: reporter info, target (review/bathhouse/user), reason, description
- [x] Resolve/dismiss actions with notes
- [x] API hooks: `useGetAdminComplaints`, `useGetAdminComplaintsId`, `usePatchAdminComplaintsIdResolve`, `usePatchAdminComplaintsIdDismiss`
- [x] Write tests for complaint list, resolve/dismiss actions
- [x] Run project test suite - must pass before task 16

### Task 16: Admin - City management and global promo codes

**Files:**
- Create: `frontend/src/pages/admin/CityManagement.tsx`
- Create: `frontend/src/pages/admin/GlobalPromoCodes.tsx`

- [x] City CRUD: table with name, slug, create/edit/delete
- [x] Global promo code creation: type (percentage/fixed/free_hour), limits, validity
- [x] API hooks: `useGetCities`, `usePostAdminCities`, `usePutAdminCitiesId`, `useDeleteAdminCitiesId`, `usePostAdminPromoCodes`
- [x] Write tests for city CRUD, promo code creation
- [x] Run project test suite - must pass before task 17

### Task 17: Owner enhancements - Payment display and complaint submission

**Files:**
- Modify: `frontend/src/pages/bookings/BookingList.tsx` (add payment info)
- Modify: `frontend/src/pages/reviews/ReviewList.tsx` (add media display, report button)
- Create: `frontend/src/components/ReportModal.tsx`

- [x] Add payment status column/badge to owner booking list
- [x] Add payment initiation button for unpaid confirmed bookings
- [x] Display review media (photos/videos) in owner review list
- [x] Add ReportModal component for reporting reviews/bathhouses/users
- [x] API hooks: `useGetBookingsIdPayment`, `usePostBookingsIdPay`, `usePostReviewsIdReport`, `usePostBathhousesIdReport`, `usePostUsersIdReport`
- [x] Write tests for payment display, report modal
- [x] Run project test suite - must pass before task 18

### Task 18: OAuth social login and device tokens

**Files:**
- Modify: `frontend/src/pages/Login.tsx` (add OAuth buttons)
- Modify: `frontend/src/pages/Register.tsx` (add OAuth buttons)
- Create: `frontend/src/lib/useDeviceToken.ts`

- [x] Add VK, Yandex, Google OAuth login buttons to Login/Register pages
- [x] OAuth redirect flow: redirect to provider, handle callback with token
- [x] Device token registration hook for push notification support
- [x] API hooks: `useGetAuthOauthProvider`, `usePostDeviceTokens`, `useDeleteDeviceTokensId`
- [x] Write tests for OAuth button rendering, device token hook
- [x] Run project test suite - must pass before task 19

### Task 19: Client chat and notifications integration

**Files:**
- Modify: `frontend/src/components/ClientLayout.tsx` (add notification bell, chat badge)
- Create: `frontend/src/pages/client/ClientChat.tsx`
- Create: `frontend/src/pages/client/ClientNotifications.tsx`

- [x] Client chat page: start conversation with bathhouse, existing conversations list, real-time messaging
- [x] Client notifications page (reuse notification patterns from owner)
- [x] Notification bell in ClientLayout header
- [x] API hooks: `usePostBathhousesIdChat`, `useGetMyConversations`, `useGetConversationsIdMessages`, `usePostConversationsIdMessages`, `usePatchConversationsIdRead`, `useGetMyNotifications`, `usePatchMyNotificationsIdRead`, `usePatchMyNotificationsReadAll`
- [x] Write tests for chat initiation, notification display
- [x] Run project test suite - must pass before task 20

### Task 20: Admin notifications, chat, and admin profile

**Files:**
- Create: `frontend/src/pages/admin/AdminNotifications.tsx`
- Create: `frontend/src/pages/admin/AdminProfile.tsx`

- [x] Admin notifications page with notification preferences
- [x] Admin profile settings (name, email, avatar)
- [x] Notification bell in AdminLayout header
- [x] Write tests for admin notifications, profile
- [x] Run project test suite - must pass before task 21

### Task 21: Verify acceptance criteria

- [x] Manual test: login as client, search bathhouse, create booking, pay, leave review with photos
- [x] Manual test: login as owner, manage bathhouse, confirm booking, respond to review
- [x] Manual test: login as admin, moderate bathhouse, approve review, manage complaints
- [x] Verify all 120+ API endpoints have corresponding UI coverage
- [x] Run full test suite: `cd frontend && npx vitest run`
- [x] Run linter: `cd frontend && npm run lint`
- [x] Run build: `cd frontend && npm run build`

### Task 22: Update documentation

- [ ] Update CLAUDE.md frontend section with new pages and role-based architecture
- [ ] Update README.md with multi-role SPA description
- [ ] Move this plan to `docs/plans/completed/`
