---
# Subsystem 3: Owner Onboarding (FR-009-013, FR-034)

## Overview
KYC verification system, offer/contract acceptance, owner payment details management, and 7-step listing creation wizard with draft support. These features gate the owner flow: KYC + offer acceptance required before listing creation.

## Context
- Existing bathhouse CRUD: `internal/service/bathhouse_service.go`, `internal/handler/bathhouse_handler.go`
- Existing media upload: `internal/handler/media_handler.go`
- Roles: owner creates bathhouses, representative manages them
- Moderation flow exists: bathhouse status goes pending → active/rejected
- AccessChecker for RBAC: `internal/service/access_checker.go`

## Dependencies
- Depends on: Wallet System (Task 1) — wallet auto-created during registration
- No other new subsystem dependencies

## Development Approach
- **Testing approach**: Regular (code first, then tests)
- Complete each task fully before moving to the next
- **CRITICAL: every task MUST include new/updated tests**
- **CRITICAL: all tests must pass before starting next task**

## Implementation Steps

### Task 3.1: KYC System (FR-010)

**Files:**
- Create: `internal/domain/kyc.go`
- Create: `internal/service/kyc_service.go`
- Create: `internal/handler/kyc_handler.go`
- Create: `internal/repository/postgres/kyc_repo.go`
- Create: `internal/repository/mock/kyc_repo.go`
- Modify: `internal/repository/interfaces.go`
- Create: `migrations/XXXXXX_kyc.up.sql`
- Create: `migrations/XXXXXX_kyc.down.sql`

- [ ] KYC model:
  - ID (uuid), UserID (uuid), Status (enum: pending/approved/rejected/expired)
  - EntityType (enum: individual/sole_proprietor/self_employed/legal_entity)
  - FullName, INN, OGRNIP (for sole_proprietor), CompanyName (for legal_entity)
  - DocumentURLs ([]string — passport scan, registration docs, etc.)
  - SubmittedAt, ReviewedAt, ExpiresAt (*time.Time), RejectionReason (string)
- [ ] KYCRepository interface: Create, GetByUserID, Update, ListPending, Approve, Reject
- [ ] KYCService:
  - Submit(ctx, userID, data) — validate required fields by entity type, set status=pending
  - GetStatus(ctx, userID) — return current KYC status
  - Approve(ctx, kycID, adminID) — set approved, set ExpiresAt = now + 1 year
  - Reject(ctx, kycID, adminID, reason) — set rejected with reason
  - CheckExpiry(ctx) — cron: find KYC expiring in 30 days → notify, expired → block listing creation
- [ ] Endpoints:
  - POST /api/v1/my/kyc — submit KYC documents (RequireRole: owner)
  - GET /api/v1/my/kyc — get KYC status (RequireRole: owner)
  - GET /api/v1/admin/kyc/pending — list pending KYC (RequireRole: admin)
  - PATCH /api/v1/admin/kyc/{id}/approve — approve (RequireRole: admin)
  - PATCH /api/v1/admin/kyc/{id}/reject — reject with reason (RequireRole: admin)
- [ ] Block bathhouse creation if KYC not approved
- [ ] SumSub integration placeholder (manual review for now, interface ready for future)
- [ ] Write tests
- [ ] Run `go test ./... -v` — must pass

### Task 3.2: Offer/Contract Acceptance (FR-011)

**Files:**
- Create: `internal/domain/offer.go`
- Create: `internal/service/offer_service.go`
- Create: `internal/handler/offer_handler.go`
- Create: `internal/repository/postgres/offer_repo.go`
- Modify: `internal/repository/interfaces.go`
- Create: `migrations/XXXXXX_offers.up.sql`

- [ ] OfferAcceptance model: ID, UserID, OfferVersion (string, e.g. "v1.0"), AcceptedAt, IP (string)
- [ ] OfferRepository: Create, GetLatestByUser, GetCurrentVersion
- [ ] OfferService:
  - GetCurrentVersion(ctx) — return current offer version from config (BANI_OFFER_VERSION)
  - Accept(ctx, userID, version, ip) — create acceptance record
  - IsAccepted(ctx, userID) — check if user accepted current version
  - RequireReacceptance(ctx) — called when offer version changes, marks all as needing re-acceptance
- [ ] POST /api/v1/my/offer/accept — accept current offer version
- [ ] GET /api/v1/my/offer/status — check acceptance status, return { accepted, version, accepted_at }
- [ ] Block listing creation if offer not accepted (check in bathhouse service)
- [ ] On offer version update: existing acceptances for old version become invalid
- [ ] Write tests
- [ ] Run `go test ./... -v` — must pass

### Task 3.3: Owner Payment Details (FR-012)

**Files:**
- Create: `internal/domain/payment_details.go`
- Create: `internal/service/payment_details_service.go`
- Create: `internal/handler/payment_details_handler.go`
- Create: `internal/repository/postgres/payment_details_repo.go`
- Modify: `internal/repository/interfaces.go`
- Create: `migrations/XXXXXX_owner_payment_details.up.sql`

- [ ] PaymentDetails model:
  - ID, UserID, EntityType (matches KYC entity type)
  - For individual: BankCardNumber, CardHolderName
  - For sole_proprietor/legal_entity: BankAccount, BIK, INN, CorrespondentAccount, BankName
  - IsVerified (bool), CreatedAt, UpdatedAt
- [ ] PaymentDetailsRepository: Upsert, GetByUserID
- [ ] PaymentDetailsService:
  - Set(ctx, userID, details) — validate by entity type (different required fields)
  - Get(ctx, userID) — return current details (mask sensitive data in response)
  - Validate(ctx, userID) — check all required fields present for entity type
- [ ] PUT /api/v1/my/payment-details — set/update payment details (RequireRole: owner)
- [ ] GET /api/v1/my/payment-details — get current details (RequireRole: owner)
- [ ] Validation rules:
  - Individual: bank card number required (16 digits)
  - Sole proprietor: INN (12 digits), bank account (20 digits), BIK (9 digits)
  - Legal entity: INN (10 digits), bank account, BIK, correspondent account, company name
- [ ] Write tests
- [ ] Run `go test ./... -v` — must pass

### Task 3.4: 7-Step Wizard Draft Support (FR-034)

**Files:**
- Create: `internal/domain/listing_draft.go`
- Create: `internal/service/listing_draft_service.go`
- Create: `internal/handler/listing_draft_handler.go`
- Create: `internal/repository/postgres/listing_draft_repo.go`
- Create: `internal/repository/mock/listing_draft_repo.go`
- Modify: `internal/repository/interfaces.go`
- Create: `migrations/XXXXXX_listing_drafts.up.sql`

- [ ] ListingDraft model: ID, UserID, CurrentStep (int, 1-7), Data (map[int]json.RawMessage — step data per step number), CreatedAt, UpdatedAt
- [ ] Steps:
  1. Basic info (name, type, description)
  2. Location (address, city, coordinates)
  3. Amenities & capacity
  4. Photos (min 3)
  5. Pricing & schedule
  6. Booking settings (mode, buffer, lead time)
  7. Review & submit
- [ ] ListingDraftRepository: Create, GetByID, GetByUserID, UpdateStep, Delete
- [ ] ListingDraftService:
  - Create(ctx, userID) — check KYC approved + offer accepted, create draft
  - SaveStep(ctx, draftID, userID, step, data) — validate step data, save
  - GetDraft(ctx, draftID, userID) — return draft with all step data
  - ListDrafts(ctx, userID) — return user's drafts
  - Submit(ctx, draftID, userID) — validate all steps complete, convert to bathhouse, set status=pending
  - Delete(ctx, draftID, userID) — discard draft
- [ ] POST /api/v1/my/listing-drafts — create draft
- [ ] PUT /api/v1/my/listing-drafts/{id}/step/{step} — save step data
- [ ] GET /api/v1/my/listing-drafts/{id} — get draft with all steps
- [ ] GET /api/v1/my/listing-drafts — list user's drafts
- [ ] POST /api/v1/my/listing-drafts/{id}/submit — convert draft to bathhouse + send to moderation
- [ ] DELETE /api/v1/my/listing-drafts/{id} — discard draft
- [ ] Write tests
- [ ] Run `go test ./... -v` — must pass

### Task 3.5: Onboarding Gate Integration

**Files:**
- Modify: `internal/service/bathhouse_service.go`

- [ ] Modify CreateBathhouse to check prerequisites:
  1. KYC approved (call KYCService.GetStatus)
  2. Offer accepted (call OfferService.IsAccepted)
  3. Payment details set (call PaymentDetailsService.Validate)
- [ ] Return appropriate error if any prerequisite not met (ErrKYCNotApproved, ErrOfferNotAccepted, ErrPaymentDetailsNotSet)
- [ ] Add these errors to domain/errors.go and handler/response.go mapping
- [ ] Write tests
- [ ] Run `go test ./... -v -race` — must pass
- [ ] Run linter: `make lint`
