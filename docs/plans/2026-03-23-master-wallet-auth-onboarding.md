---
# Master Implementation Plan: Wallet + Auth + Owner Onboarding

## Overview
Combined execution plan for three subsystems in dependency order:
  1. Wallet System (foundational - no deps)
  2. Auth & Security (uses wallet for welcome bonus)
  3. Owner Onboarding (depends on wallet for payouts, auth for phone/2FA)

Total: 21 tasks across 3 subsystems. Each subsystem has its own detailed plan file.

## Context
- Files involved: ~60 new/modified files across domain, service, handler, repository, migrations
- Related patterns: existing clean architecture (handler -> service -> repository), fx DI, mock repos
- Dependencies: robfig/cron (scheduling), pquerna/otp (TOTP), SMS.ru adapter (OTP)
- Existing cron system: internal/cron/ already has module.go, analytics.go, expiry.go, calendar_sync.go
- Latest migration: 000037 - new migrations start at 000038

## Development Approach
- **Testing approach**: Regular (code first, then tests)
- Complete each task fully before moving to the next
- Follow dependency order strictly: Wallet -> Auth -> Onboarding
- **CRITICAL: every task MUST include new/updated tests**
- **CRITICAL: all tests must pass before starting next task**

## Implementation Steps

### Phase 1: Wallet System (docs/plans/2026-03-23-01-wallet-system.md)

#### Task 1: Wallet Domain Models (Task 1.1)

**Files:**
- Create: `internal/domain/wallet.go`
- Modify: `internal/domain/errors.go`

- [x] Wallet, WalletTransaction, WalletHold models with enums
- [x] WalletTransactionFilter for pagination
- [x] Domain errors: ErrWalletNotFound, ErrInsufficientWalletBalance, ErrWalletLimitExceeded, ErrWalletFrozen, ErrHoldNotFound, ErrHoldExpired, ErrTopUpBelowMinimum, ErrTopUpAboveMaximum
- [x] Add error mappings to `internal/handler/response.go`

#### Task 2: Wallet Repository (Task 1.2)

**Files:**
- Modify: `internal/repository/interfaces.go`
- Create: `internal/repository/postgres/wallet_repo.go`
- Create: `internal/repository/mock/wallet_repo.go`
- Create: `migrations/000038_wallets.up.sql`, `migrations/000038_wallets.down.sql`

- [x] WalletRepository interface (20 methods: CRUD, transactions, holds, bonus expiry)
- [x] Migration: wallets, wallet_transactions, wallet_holds tables with indexes
- [x] Postgres implementation with pgx
- [x] Mock implementation for testing
- [x] Write tests, run `go test ./... -v`

#### Task 3: Wallet Service (Task 1.3)

**Files:**
- Create: `internal/service/wallet_service.go`

- [x] Core methods: CreateWallet, GetWallet, TopUp, Spend, Hold, CaptureHold, ReleaseHold, Refund, AddBonus, GetBalance, ListTransactions, ExpireBonuses
- [x] Priority spending: expiring bonuses first (by expires_at ASC), then non-bonus (FIFO)
- [x] Balance limits: max 100,000 RUB / 3,000 BYN
- [x] Top-up limits: min 500 RUB, max 30,000 RUB per tx
- [x] Write tests, run `go test ./... -v`

#### Task 4: Wallet Handler & Routes (Task 1.4)

**Files:**
- Create: `internal/handler/wallet_handler.go`
- Modify: `internal/server/router.go`

- [x] GET /api/v1/my/wallet, GET /api/v1/my/wallet/transactions, POST /api/v1/my/wallet/topup, GET /api/v1/my/wallet/holds
- [x] Swagger annotations
- [x] Register routes, write handler tests, run `go test ./... -v`

#### Task 5: Owner Payout System (Task 1.5)

**Files:**
- Create: `internal/domain/payout.go`
- Create: `internal/service/payout_service.go`
- Create: `internal/handler/payout_handler.go`
- Create: `internal/repository/postgres/payout_repo.go`
- Modify: `internal/repository/interfaces.go`
- Create: `migrations/000039_payouts.up.sql`, `migrations/000039_payouts.down.sql`

- [x] Payout model, repository, service (RequestPayout, SetAutoPayoutThreshold, ProcessPayout)
- [x] Daily/monthly payout limits
- [x] POST /api/v1/my/wallet/payout, PUT /api/v1/my/wallet/auto-payout, GET /api/v1/my/wallet/payouts
- [x] Write tests, run `go test ./... -v`

#### Task 6: Bonus Expiration Cron (Task 1.6)

**Files:**
- Create: `internal/cron/wallet_jobs.go`
- Modify: `internal/cron/module.go`

- [x] Daily cron: expire bonuses older than 180 days (configurable BANI_WALLET_BONUS_EXPIRY_DAYS)
- [x] Notification job: warn at 14 days and 3 days before expiry
- [x] Register in existing cron module
- [x] Write tests, run `go test ./... -v`

#### Task 7: Wallet fx Module (Task 1.7)

**Files:**
- Create: `internal/service/wallet_module.go` or modify existing module.go
- Create: `internal/handler/wallet_module.go` or modify existing module.go
- Modify: `internal/app/app.go`

- [x] fx.Module for wallet service + handler + payout
- [x] Register in main app container
- [x] Verify: `go build ./...`, `go test ./... -v -race`, `make lint`

### Phase 2: Authentication & Security (docs/plans/2026-03-23-02-auth-security.md)

#### Task 8: Phone + OTP Authentication (Task 2.1)

**Files:**
- Modify: `internal/domain/user.go`
- Create: `internal/service/otp_service.go`
- Modify: `internal/service/auth_service.go`
- Modify: `internal/handler/auth_handler.go`
- Create: `migrations/000040_phone_auth.up.sql`, `migrations/000040_phone_auth.down.sql`

- [x] Phone + PhoneVerified fields on User model
- [x] OTPService: Redis-backed, 6-digit codes, 5 min TTL, 3 attempts, rate limiting
- [x] SMSProvider interface + SMS.ru adapter
- [x] POST /api/v1/auth/register-phone, /login-phone, /verify-phone
- [x] Write tests, run `go test ./... -v`

#### Task 9: Two-Factor Authentication (Task 2.2)

**Files:**
- Create: `internal/service/twofa_service.go`
- Modify: `internal/handler/auth_handler.go`, `internal/domain/user.go`
- Create: `migrations/000041_two_factor_auth.up.sql`, `migrations/000041_two_factor_auth.down.sql`

- [x] TOTP via pquerna/otp + SMS 2FA
- [x] Partial token flow for 2FA during login
- [x] POST /api/v1/auth/2fa/totp/enable, /verify, DELETE /totp, POST /sms/enable, POST /2fa/verify
- [x] Write tests, run `go test ./... -v`

#### Task 10: Session Management (Task 2.3)

**Files:**
- Create: `internal/domain/session.go`, `internal/service/session_service.go`, `internal/handler/session_handler.go`
- Create: `internal/repository/postgres/session_repo.go`, `internal/repository/mock/session_repo.go`
- Modify: `internal/repository/interfaces.go`, `internal/middleware/auth.go`
- Create: `migrations/000042_sessions.up.sql`, `migrations/000042_sessions.down.sql`

- [x] Session model with device/browser/IP tracking
- [x] Session validation in auth middleware, auto-expire after 30 days
- [x] GET/DELETE /api/v1/my/sessions, DELETE /api/v1/my/sessions/{id}
- [x] Write tests, run `go test ./... -v`

#### Task 11: Password Reset (Task 2.4)

**Files:**
- Modify: `internal/service/auth_service.go`, `internal/handler/auth_handler.go`

- [x] POST /api/v1/auth/forgot-password (Redis token, email link, rate limited)
- [x] POST /api/v1/auth/reset-password (validate token, update password, terminate sessions)
- [x] Write tests, run `go test ./... -v`

#### Task 12: Account Deletion (Task 2.5)

**Files:**
- Create: `internal/service/account_deletion_service.go`
- Modify: `internal/handler/auth_handler.go`, `internal/domain/user.go`
- Create: `migrations/000043_account_deletion.up.sql`, `migrations/000043_account_deletion.down.sql`

- [x] 30-day grace period with restore option
- [x] Data anonymization on execution, fund return for top-ups, burn bonuses
- [x] POST /api/v1/auth/delete-account, /restore-account
- [x] Cron: reminders at day 0/14/27, execute at day 30
- [x] Write tests, run `go test ./... -v`

#### Task 13: Age Verification & Welcome Bonus (Task 2.6)

**Files:**
- Modify: `internal/service/auth_service.go`, `internal/handler/auth_handler.go`, `internal/domain/user.go`

- [x] age_confirmed required on registration
- [x] Auto-create wallet + credit 500 RUB welcome bonus (30 day expiry)
- [x] Config: BANI_WELCOME_BONUS_AMOUNT, BANI_WELCOME_BONUS_EXPIRY_DAYS
- [x] Write tests, run `go test ./... -v`

#### Task 14: Auth Phase Verify (Task 2.7)

- [x] `go test ./... -v -race`
- [x] `make lint`
- [x] `go build ./...`
- [x] `make swagger`

### Phase 3: Owner Onboarding (docs/plans/2026-03-23-03-owner-onboarding.md)

#### Task 15: KYC System (Task 3.1)

**Files:**
- Create: `internal/domain/kyc.go`, `internal/service/kyc_service.go`, `internal/handler/kyc_handler.go`
- Create: `internal/repository/postgres/kyc_repo.go`, `internal/repository/mock/kyc_repo.go`
- Modify: `internal/repository/interfaces.go`
- Create: `migrations/000044_kyc.up.sql`, `migrations/000044_kyc.down.sql`

- [x] KYC model with entity types (individual, sole_proprietor, self_employed, legal_entity)
- [x] Submit/approve/reject flow, expiry checking
- [x] POST /api/v1/my/kyc, GET /api/v1/my/kyc
- [x] GET /api/v1/admin/kyc/pending, PATCH /admin/kyc/{id}/approve, /reject
- [x] Write tests, run `go test ./... -v`

#### Task 16: Offer/Contract Acceptance (Task 3.2)

**Files:**
- Create: `internal/domain/offer.go`, `internal/service/offer_service.go`, `internal/handler/offer_handler.go`
- Create: `internal/repository/postgres/offer_repo.go`
- Modify: `internal/repository/interfaces.go`
- Create: `migrations/000045_offers.up.sql`, `migrations/000045_offers.down.sql`

- [x] OfferAcceptance model, versioned acceptance tracking
- [x] POST /api/v1/my/offer/accept, GET /api/v1/my/offer/status
- [x] Write tests, run `go test ./... -v`

#### Task 17: Owner Payment Details (Task 3.3)

**Files:**
- Create: `internal/domain/payment_details.go`, `internal/service/payment_details_service.go`, `internal/handler/payment_details_handler.go`
- Create: `internal/repository/postgres/payment_details_repo.go`
- Modify: `internal/repository/interfaces.go`
- Create: `migrations/000046_owner_payment_details.up.sql`, `migrations/000046_owner_payment_details.down.sql`

- [x] PaymentDetails with entity-type-specific fields and validation
- [x] PUT/GET /api/v1/my/payment-details
- [x] Write tests, run `go test ./... -v`

#### Task 18: 7-Step Listing Draft Wizard (Task 3.4)

**Files:**
- Create: `internal/domain/listing_draft.go`, `internal/service/listing_draft_service.go`, `internal/handler/listing_draft_handler.go`
- Create: `internal/repository/postgres/listing_draft_repo.go`, `internal/repository/mock/listing_draft_repo.go`
- Modify: `internal/repository/interfaces.go`
- Create: `migrations/000047_listing_drafts.up.sql`, `migrations/000047_listing_drafts.down.sql`

- [x] ListingDraft with step-based data storage (7 steps)
- [x] CRUD + submit endpoints for drafts
- [x] Write tests, run `go test ./... -v`

#### Task 19: Onboarding Gate Integration (Task 3.5)

**Files:**
- Modify: `internal/service/bathhouse_service.go`
- Modify: `internal/domain/errors.go`, `internal/handler/response.go`

- [ ] Gate CreateBathhouse behind KYC + Offer + PaymentDetails checks
- [ ] New domain errors: ErrKYCNotApproved, ErrOfferNotAccepted, ErrPaymentDetailsNotSet
- [ ] Write tests, run `go test ./... -v -race`, `make lint`

### Task 20: Final Verification

- [ ] Run full test suite: `go test ./... -v -race`
- [ ] Run linter: `make lint`
- [ ] Regenerate swagger: `make swagger`
- [ ] Verify build: `go build ./...`

### Task 21: Update Documentation

- [ ] Update CLAUDE.md with new subsystem patterns (wallet, KYC, OTP, sessions)
- [ ] Move individual plan files to `docs/plans/completed/`
- [ ] Move this master plan to `docs/plans/completed/`
