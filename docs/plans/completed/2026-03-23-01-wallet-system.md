---
# Subsystem 1: Wallet System (FR-006, FR-013, FR-110-123)

## Overview
Full monetary wallet system with tagged transactions, holds, expiration, priority spending, owner payouts, and bonus lifecycle. This is the most foundational new subsystem — many other features depend on it.

## Context
- Module path: `github.com/rekurt/relax-hub`
- Architecture: handler → service → repository (clean architecture)
- DI: Uber fx with `module.go` per package
- DB: PostgreSQL + PostGIS, migrations in `migrations/`
- Prices: kopecks (int64)
- API format: `{ success, data, error: { code, message }, meta }`
- Existing interfaces: `internal/repository/interfaces.go`
- Existing domain models: `internal/domain/`
- Tests: mock repos + table-driven tests
- Currently NO wallet model exists. The loyalty system has points but the BRD requires a full monetary wallet.
- Related existing code: `internal/service/loyalty_service.go`, `internal/service/referral_service.go`, `internal/service/certificate_service.go`

## Dependencies
- No dependencies on other new subsystems (foundational)
- Depends on existing: PaymentProvider (YooKassa), notification system, user service

## Development Approach
- **Testing approach**: Regular (code first, then tests)
- Complete each task fully before moving to the next
- **CRITICAL: every task MUST include new/updated tests**
- **CRITICAL: all tests must pass before starting next task**

## Implementation Steps

### Task 1.1: Wallet Domain Models

**Files:**
- Create: `internal/domain/wallet.go`
- Modify: `internal/domain/errors.go`

- [ ] Create Wallet model: ID (uuid.UUID), UserID (uuid.UUID), Region (string: "RU"/"BY"), Balance (int64, kopecks), HeldBalance (int64), Currency (string: "RUB"/"BYN"), Status (enum: active/frozen), CreatedAt, UpdatedAt
- [ ] Create WalletTransaction model: ID, WalletID, Type (enum: topup, refund, compensation, promo, cashback, gift_cert, referral, welcome_bonus, spend, hold, unhold, expiration), Amount (int64), BalanceAfter (int64), RelatedBookingID (*uuid.UUID), Description (string), ExpiresAt (*time.Time), IsBonus (bool), CreatedAt
- [ ] Create WalletHold model: ID, WalletID, BookingID, Amount (int64), Status (enum: active/captured/released), CreatedAt, ExpiresAt
- [ ] Create WalletTransactionFilter: WalletID, Type, DateFrom, DateTo, Pagination (Page, PageSize)
- [ ] Add domain errors to `internal/domain/errors.go`:
  - ErrWalletNotFound
  - ErrInsufficientWalletBalance
  - ErrWalletLimitExceeded (balance > 100,000 RUB / 3,000 BYN)
  - ErrWalletFrozen
  - ErrHoldNotFound
  - ErrHoldExpired
  - ErrTopUpBelowMinimum (< 500 RUB)
  - ErrTopUpAboveMaximum (> 30,000 RUB)

### Task 1.2: Wallet Repository

**Files:**
- Modify: `internal/repository/interfaces.go`
- Create: `internal/repository/postgres/wallet_repo.go`
- Create: `internal/repository/mock/wallet_repo.go`
- Create: `migrations/XXXXXX_wallets.up.sql`
- Create: `migrations/XXXXXX_wallets.down.sql`

- [ ] Define WalletRepository interface in `internal/repository/interfaces.go`:
  ```go
  type WalletRepository interface {
      GetByUserID(ctx context.Context, userID uuid.UUID) (*domain.Wallet, error)
      GetByUserIDAndRegion(ctx context.Context, userID uuid.UUID, region string) (*domain.Wallet, error)
      Create(ctx context.Context, wallet *domain.Wallet) error
      UpdateBalance(ctx context.Context, walletID uuid.UUID, newBalance int64, newHeldBalance int64) error
      FreezeWallet(ctx context.Context, walletID uuid.UUID) error
      UnfreezeWallet(ctx context.Context, walletID uuid.UUID) error
      CreateTransaction(ctx context.Context, tx *domain.WalletTransaction) error
      ListTransactions(ctx context.Context, filter domain.WalletTransactionFilter) ([]domain.WalletTransaction, int, error)
      GetExpiringBonuses(ctx context.Context, walletID uuid.UUID, before time.Time) ([]domain.WalletTransaction, error)
      GetBonusTransactionsForSpending(ctx context.Context, walletID uuid.UUID) ([]domain.WalletTransaction, error)
      ExpireBonuses(ctx context.Context, transactionIDs []uuid.UUID) error
      CreateHold(ctx context.Context, hold *domain.WalletHold) error
      CaptureHold(ctx context.Context, holdID uuid.UUID) error
      ReleaseHold(ctx context.Context, holdID uuid.UUID) error
      GetActiveHolds(ctx context.Context, walletID uuid.UUID) ([]domain.WalletHold, error)
      GetHoldByID(ctx context.Context, holdID uuid.UUID) (*domain.WalletHold, error)
  }
  ```
- [ ] Create migration `migrations/XXXXXX_wallets.up.sql`:
  ```sql
  CREATE TABLE wallets (
      id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
      user_id UUID NOT NULL REFERENCES users(id),
      region VARCHAR(2) NOT NULL DEFAULT 'RU',
      balance BIGINT NOT NULL DEFAULT 0,
      held_balance BIGINT NOT NULL DEFAULT 0,
      currency VARCHAR(3) NOT NULL DEFAULT 'RUB',
      status VARCHAR(20) NOT NULL DEFAULT 'active',
      created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
      updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
      UNIQUE(user_id, region)
  );
  CREATE INDEX idx_wallets_user_id ON wallets(user_id);

  CREATE TABLE wallet_transactions (
      id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
      wallet_id UUID NOT NULL REFERENCES wallets(id),
      type VARCHAR(30) NOT NULL,
      amount BIGINT NOT NULL,
      balance_after BIGINT NOT NULL,
      related_booking_id UUID REFERENCES bookings(id),
      description TEXT,
      expires_at TIMESTAMPTZ,
      is_bonus BOOLEAN NOT NULL DEFAULT false,
      created_at TIMESTAMPTZ NOT NULL DEFAULT now()
  );
  CREATE INDEX idx_wallet_tx_wallet_type ON wallet_transactions(wallet_id, type, created_at);
  CREATE INDEX idx_wallet_tx_expires ON wallet_transactions(wallet_id, expires_at) WHERE expires_at IS NOT NULL AND is_bonus = true;

  CREATE TABLE wallet_holds (
      id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
      wallet_id UUID NOT NULL REFERENCES wallets(id),
      booking_id UUID REFERENCES bookings(id),
      amount BIGINT NOT NULL,
      status VARCHAR(20) NOT NULL DEFAULT 'active',
      created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
      expires_at TIMESTAMPTZ NOT NULL
  );
  CREATE INDEX idx_wallet_holds_wallet ON wallet_holds(wallet_id, status);
  ```
- [ ] Implement postgres WalletRepository with all interface methods using pgx
- [ ] Implement mock WalletRepository in `internal/repository/mock/wallet_repo.go`
- [ ] Write tests for wallet repository mock
- [ ] Run `go test ./... -v` — must pass

### Task 1.3: Wallet Service

**Files:**
- Create: `internal/service/wallet_service.go`

- [ ] Implement WalletService with methods:
  - CreateWallet(ctx, userID, region) — auto-detect region, set currency
  - GetWallet(ctx, userID) — return wallet with available balance (balance - held)
  - TopUp(ctx, userID, amount) — validate limits (min 500 RUB, max 30,000 RUB per tx), create payment via PaymentProvider
  - Spend(ctx, walletID, amount, bookingID) — priority spending: expiring bonuses first (sorted by expires_at ASC), then non-expiring (sorted by created_at ASC)
  - Hold(ctx, walletID, amount, bookingID) — check available balance, create hold, update held_balance
  - CaptureHold(ctx, holdID) — convert hold to spend, deduct from balance + held_balance
  - ReleaseHold(ctx, holdID) — release hold, restore held_balance
  - Refund(ctx, walletID, amount, bookingID, description) — credit wallet with refund transaction
  - AddBonus(ctx, walletID, amount, bonusType, expiresIn) — credit bonus with expiration
  - GetBalance(ctx, userID) — total, held, available, expiring_soon amounts
  - ListTransactions(ctx, userID, filter) — paginated transaction list
  - ExpireBonuses(ctx) — expire all bonuses past their expiration date
- [ ] Implement priority spending logic:
  1. Query bonus transactions sorted by expires_at ASC (expiring soonest first)
  2. Then non-bonus transactions sorted by created_at ASC (FIFO)
  3. Deduct from each until amount is fully covered
- [ ] Balance limits: max 100,000 RUB (10,000,000 kopecks) / 3,000 BYN (300,000 kopecks)
- [ ] Top-up limits: min 500 RUB (50,000 kopecks), max 30,000 RUB (3,000,000 kopecks) per transaction
- [ ] Auto-create wallet on user registration (hook into user service or call from auth service)
- [ ] Write comprehensive tests with mock repos (table-driven)
- [ ] Run `go test ./... -v` — must pass

### Task 1.4: Wallet Handler & Routes

**Files:**
- Create: `internal/handler/wallet_handler.go`
- Modify: `internal/server/router.go`

- [ ] GET /api/v1/my/wallet — get wallet balance and summary (RequireAuth)
  - Response: { balance, held_balance, available_balance, currency, expiring_soon: { amount, earliest_expiry } }
- [ ] GET /api/v1/my/wallet/transactions — list transactions with filters (RequireAuth)
  - Query params: type, date_from, date_to, page, page_size
  - Response: paginated list of transactions
- [ ] POST /api/v1/my/wallet/topup — initiate top-up (RequireAuth)
  - Request: { amount, payment_method: "card"/"sbp" }
  - Creates payment via PaymentProvider, returns redirect URL
- [ ] GET /api/v1/my/wallet/holds — list active holds (RequireAuth)
- [ ] Add swagger annotations for all endpoints
- [ ] Register routes in `internal/server/router.go` under authenticated group
- [ ] Create `internal/handler/module.go` entry or modify existing
- [ ] Write handler tests using httptest + mock services
- [ ] Run `go test ./... -v` — must pass

### Task 1.5: Owner Wallet & Payout (FR-100-104, FR-117-120)

**Files:**
- Create: `internal/domain/payout.go`
- Create: `internal/service/payout_service.go`
- Create: `internal/handler/payout_handler.go`
- Create: `internal/repository/postgres/payout_repo.go`
- Modify: `internal/repository/interfaces.go`
- Modify: `internal/server/router.go`
- Create: `migrations/XXXXXX_payouts.up.sql`

- [ ] Payout model: ID, UserID, Amount (int64), Status (pending/processing/completed/failed), BankDetails (JSONB), RequestedAt, ProcessedAt, FailureReason
- [ ] PayoutRepository interface: Create, GetByID, ListByUser, UpdateStatus
- [ ] PayoutService:
  - RequestPayout(ctx, userID, amount) — min 500 RUB, check available balance, create payout record
  - SetAutoPayoutThreshold(ctx, userID, threshold) — auto-payout when balance exceeds threshold
  - GetPayoutHistory(ctx, userID, filter) — paginated history
  - CalculateAvailableBalance(ctx, userID) — total - held - pending_payouts
  - ProcessPayout(ctx, payoutID) — execute payout (admin/cron)
- [ ] Daily payout limits: 500,000 RUB/day for individuals, 3,000,000 RUB/month
- [ ] POST /api/v1/my/wallet/payout — request payout (RequireRole: owner)
- [ ] PUT /api/v1/my/wallet/auto-payout — configure auto-payout threshold (RequireRole: owner)
- [ ] GET /api/v1/my/wallet/payouts — payout history (RequireRole: owner)
- [ ] Add swagger annotations
- [ ] Write tests
- [ ] Run `go test ./... -v` — must pass

### Task 1.6: Bonus Expiration Cron Job (FR-121-123)

**Files:**
- Create: `internal/cron/scheduler.go` (if not exists)
- Create: `internal/cron/wallet_jobs.go`
- Modify: `internal/service/wallet_service.go`
- Modify: `internal/app/` (fx module registration)

- [ ] Create daily cron job to expire bonus transactions older than 180 days (configurable via BANI_WALLET_BONUS_EXPIRY_DAYS)
- [ ] Create notification job: send notification 14 days and 3 days before expiration
  - Notification text: "У вас {amount} ₽ бонусов истекают {date}. Используйте их для бронирования!"
- [ ] Log expired amounts as wallet_transactions with type "expiration"
- [ ] Use robfig/cron or similar for scheduling
- [ ] Register cron module in fx
- [ ] Write tests
- [ ] Run `go test ./... -v` — must pass

### Task 1.7: Wallet fx Module

**Files:**
- Create: `internal/service/wallet_module.go` (or add to existing module)
- Create: `internal/handler/wallet_module.go` (or add to existing module)
- Modify: `internal/app/app.go`

- [ ] Create fx.Module for wallet service binding
- [ ] Create fx.Module for wallet handler binding
- [ ] Register in main app fx container
- [ ] Verify build: `go build ./...`
- [ ] Run full test suite: `go test ./... -v -race`
- [ ] Run linter: `make lint`
