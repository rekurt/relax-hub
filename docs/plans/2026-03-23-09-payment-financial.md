---
# Subsystem 9: Payment & Financial Enhancements (FR-092-109, FR-124)

## Overview
SBP payment support, combo payments (wallet + card), payment holds for request bookings, escrow system, enhanced refund logic with wallet bonus, and fiscalization placeholder.

## Context
- Existing payments: `internal/payment/` — PaymentProvider interface, YooKassa implementation
- Existing payment service: `internal/service/payment_service.go`
- Existing payment handler: `internal/handler/payment_handler.go`
- Existing payment model: `internal/domain/payment.go`
- Existing refund logic: time-based tiers in booking service
- YooKassa supports SBP, card holds (capture=false)

## Dependencies
- Depends on: Wallet System (Subsystem 1) — for wallet holds, combo payments, wallet refunds
- Depends on: Booking Enhancements (Subsystem 7) — for request-based booking holds

## Development Approach
- **Testing approach**: Regular (code first, then tests)
- **CRITICAL: every task MUST include new/updated tests**
- **CRITICAL: all tests must pass before starting next task**

## Implementation Steps

### Task 9.1: SBP Payment Support (FR-092)

**Files:**
- Modify: `internal/payment/yookassa.go` (or provider file)
- Modify: `internal/domain/payment.go`
- Modify: `internal/handler/payment_handler.go`

- [x] Add PaymentMethod enum to payment model: "card", "sbp", "wallet", "combo"
- [x] YooKassa SBP: set `confirmation.type = "redirect"` for SBP payments
  - SBP returns redirect URL to bank app selection
- [x] Modify payment creation handler to accept `payment_method` field
- [x] Modify PaymentProvider interface if needed to accept method parameter
- [x] Update payment webhook handler to handle SBP-specific callbacks
- [x] Write tests
- [x] Run `go test ./... -v` — must pass

### Task 9.2: Combo Payment — Wallet + Card/SBP (FR-093)

**Files:**
- Modify: `internal/service/payment_service.go`
- Modify: `internal/handler/payment_handler.go`
- Modify: `internal/domain/payment.go`
- Create: `migrations/XXXXXX_combo_payments.up.sql`

- [x] Accept split payment request:
  ```json
  {
    "wallet_amount": 50000,
    "card_amount": 100000,
    "payment_method": "card"
  }
  ```
- [x] Add to Payment model: WalletAmount (int64), CardAmount (int64), PaymentMethod (string)
- [x] Migration: add `wallet_amount BIGINT DEFAULT 0`, `card_amount BIGINT DEFAULT 0`, `payment_method VARCHAR(10) DEFAULT 'card'` to payments
- [x] Combo payment flow:
  1. Validate: wallet_amount + card_amount == total
  2. Validate: user has sufficient wallet balance
  3. Debit wallet first (WalletService.Spend)
  4. Create card/SBP payment for remaining amount
  5. If card payment fails → refund wallet immediately (WalletService.Refund)
  6. If card payment succeeds → complete booking
- [x] For full wallet payments: no card charge needed, complete immediately
- [x] Store payment breakdown in payment record
- [x] Write tests for all combinations: full card, full wallet, combo, combo with card failure
- [x] Run `go test ./... -v` — must pass

### Task 9.3: Payment Hold for Request Bookings (FR-094)

**Files:**
- Modify: `internal/service/payment_service.go`
- Modify: `internal/service/booking_service.go`

- [x] For request-based bookings, create hold instead of charge:
  - Card: use YooKassa `capture: false` (creates authorization hold)
  - Wallet: use WalletService.Hold
  - Combo: wallet hold + card authorization
- [x] On owner approve: capture payment
  - Card: YooKassa capture API
  - Wallet: WalletService.CaptureHold
- [x] On owner reject or timeout: release hold
  - Card: YooKassa cancel API
  - Wallet: WalletService.ReleaseHold
- [x] Max hold duration: 72 hours (YooKassa limit for card holds)
- [x] If hold expires before owner response: auto-reject booking
- [x] Write tests
- [x] Run `go test ./... -v` — must pass

### Task 9.4: Escrow System (FR-124)

**Files:**
- Create: `internal/domain/escrow.go`
- Create: `internal/service/escrow_service.go`
- Create: `internal/repository/postgres/escrow_repo.go`
- Create: `internal/repository/mock/escrow_repo.go`
- Modify: `internal/repository/interfaces.go`
- Create: `migrations/XXXXXX_escrow.up.sql`

- [x] Escrow model:
  - ID, BookingID (uuid), Amount (int64), ServiceFee (int64)
  - Status (enum: held/released/disputed/refunded)
  - ClaimPeriodEndsAt (time.Time — default 48h after check-out)
  - ReleasedAt (*time.Time)
  - CreatedAt
- [x] EscrowRepository: Create, GetByBookingID, GetByID, UpdateStatus, ListMatured
- [x] EscrowService:
  - CreateEscrow(ctx, bookingID, amount, serviceFee) — on booking completion (check-out)
  - ReleaseToOwner(ctx, escrowID) — after claim period with no dispute
    - Credit owner wallet: amount - serviceFee
    - Credit platform: serviceFee
  - MarkDisputed(ctx, escrowID) — when dispute opened
  - ProcessRefund(ctx, escrowID, refundAmount) — on dispute resolution
  - ProcessMaturedEscrows(ctx) — cron: find escrows where claim_period_ends_at < now AND status=held
- [x] Integrate with booking check-out flow: on check-out → create escrow
- [x] Configurable claim period: BANI_ESCROW_CLAIM_HOURS (default 48, range 24-168)
- [x] Cron: hourly check for matured escrow entries
- [x] Write tests
- [x] Run `go test ./... -v` — must pass

### Task 9.5: Enhanced Refund Logic (FR-106-109)

**Files:**
- Modify: `internal/service/booking_service.go`
- Modify: `internal/service/payment_service.go`
- Modify: `internal/handler/booking_handler.go`

- [x] Refund destination choice:
  - PATCH /api/v1/bookings/{id}/cancel now accepts `refund_to: "wallet" | "card"`
  - Wallet refund: instant, + bonus (configurable BANI_WALLET_REFUND_BONUS_PERCENT, default 5, range 0-15)
    - Example: 1000 RUB refund to wallet = 1050 RUB credited
  - Card refund: 3-10 business days (existing flow via YooKassa)
- [x] Combo refund: proportional to original payment sources
  - If paid 30% wallet + 70% card → refund 30% to wallet + 70% to card
  - Wallet portion gets bonus, card portion doesn't
- [x] Manual refund by L2+ support (admin endpoint):
  - POST /api/v1/admin/bookings/{id}/refund
  - Request: { amount, reason, refund_to }
  - RequireRole: admin
  - Full audit log entry
- [x] Log all refunds in audit_log (from Subsystem 4) with: amount, reason, source, destination
- [x] Write tests
- [x] Run `go test ./... -v` — must pass

### Task 9.6: Fiscalization Placeholder (FR-095)

**Files:**
- Create: `internal/fiscal/fiscal.go` (interface)
- Create: `internal/fiscal/atol.go` (placeholder implementation)
- Create: `internal/fiscal/module.go`

- [x] Define FiscalProvider interface:
  ```go
  type FiscalProvider interface {
      CreateReceipt(ctx context.Context, req ReceiptRequest) (*Receipt, error)
  }
  type ReceiptRequest struct {
      Type     ReceiptType // advance, advance_credit, full_payment, refund
      Amount   int64       // kopecks
      Email    string
      Phone    string
      Items    []ReceiptItem
  }
  type ReceiptItem struct {
      Name     string
      Quantity int
      Price    int64
      VAT      string // none, vat0, vat10, vat20
  }
  ```
- [x] ATOL Online placeholder: log receipt request, return mock receipt ID
  - Config: BANI_FISCAL_PROVIDER (default "none"), BANI_FISCAL_ATOL_LOGIN, BANI_FISCAL_ATOL_PASSWORD, BANI_FISCAL_ATOL_GROUP_CODE
- [x] Hook into payment flow: create receipt after successful payment
  - On payment success: create "advance" receipt
  - On refund: create "refund" receipt
- [x] Register in fx module
- [x] Write tests
- [x] Run `go test ./... -v -race` — must pass
- [x] Run linter: `make lint`
