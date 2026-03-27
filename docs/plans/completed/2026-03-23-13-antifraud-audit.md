---
# Subsystem 13: Anti-fraud & Audit (FR section 2.14)

## Overview
Anti-fraud rules engine for detecting suspicious client and owner behavior, chat content filtering to prevent off-platform contact exchange, and enhanced admin audit logging for all administrative actions.

## Context
- Existing user blocking: admin can block users
- Existing rate limiting: `internal/middleware/rate_limit.go`
- Existing chat: `internal/service/chat_service.go`
- No anti-fraud system exists
- No admin action audit log exists (listing audit log from Subsystem 4 is separate)

## Dependencies
- Depends on: Wallet System (Subsystem 1) — fraud detection on wallet operations
- Depends on: Booking system (existing) — fraud detection on booking patterns
- Depends on: Audit Log (Subsystem 4, Task 4.2) — shares AuditLog infrastructure

## Development Approach
- **Testing approach**: Regular (code first, then tests)
- **CRITICAL: every task MUST include new/updated tests**
- **CRITICAL: all tests must pass before starting next task**

## Implementation Steps

### Task 13.1: Anti-fraud Rules Engine

**Files:**
- Create: `internal/antifraud/engine.go`
- Create: `internal/antifraud/rules.go`
- Create: `internal/antifraud/module.go`
- Create: `internal/domain/fraud_flag.go`
- Create: `internal/repository/postgres/fraud_flag_repo.go`
- Modify: `internal/repository/interfaces.go`
- Create: `migrations/XXXXXX_antifraud.up.sql`

- [ ] FraudFlag model:
  - ID, UserID (uuid), Rule (string), Severity (enum: low/medium/high/critical)
  - Status (enum: pending/reviewed/dismissed/action_taken)
  - Action (string: "flag"/"freeze_wallet"/"block_user"/"notify_admin")
  - Details (JSONB — rule-specific data)
  - CreatedAt, ReviewedAt (*time.Time), ReviewedBy (*uuid)
- [ ] FraudFlagRepository: Create, ListByUser, ListPending, UpdateStatus, CountByUserAndRule
- [ ] FraudEngine interface:
  ```go
  type FraudEngine interface {
      CheckWalletTopUp(ctx context.Context, userID uuid.UUID, amount int64, cardID string) (*FraudCheckResult, error)
      CheckBookingCreate(ctx context.Context, userID uuid.UUID, bathhouseID uuid.UUID) (*FraudCheckResult, error)
      CheckBookingCancel(ctx context.Context, userID uuid.UUID, bookingID uuid.UUID) (*FraudCheckResult, error)
      CheckPayoutRequest(ctx context.Context, userID uuid.UUID, amount int64) (*FraudCheckResult, error)
  }
  type FraudCheckResult struct {
      Flagged  bool
      Rule     string
      Severity string
      Action   string
      Message  string
  }
  ```
- [ ] Client rules:
  - RULE_MULTI_CARD_TOPUP: > 3 top-ups from different cards in 24h → flag account (medium)
  - RULE_TOPUP_CANCEL_CYCLE: topup → book → cancel cycle > 2 times in 7 days → freeze wallet (high)
  - RULE_DORMANT_BALANCE: Balance > 50,000 RUB with no bookings in 30 days → flag (low)
  - RULE_RAPID_BOOKINGS: > 5 bookings in 1 hour → flag (medium)
- [ ] Owner rules:
  - RULE_SELF_BOOKING: owner/representative books own bathhouse → block booking + flag (critical)
    - Detection: booking user_id matches bathhouse owner_id or representative user_ids
  - RULE_STRUCTURING: > 3 small withdrawals (<5000 RUB) in a day → flag (high)
  - RULE_FAKE_REVIEWS: same IP for reviewer and owner → flag (high)
- [ ] Hook checks into service methods:
  - WalletService.TopUp → CheckWalletTopUp
  - BookingService.CreateBooking → CheckBookingCreate
  - BookingService.CancelBooking → CheckBookingCancel
  - PayoutService.RequestPayout → CheckPayoutRequest
- [ ] If action is "block": prevent operation and return ErrFraudDetected
- [ ] If action is "flag" or "freeze_wallet": allow operation but create flag + notify admin
- [ ] Admin dashboard:
  - GET /api/v1/admin/antifraud/flags — list flags with filters (severity, status, user)
  - PATCH /api/v1/admin/antifraud/flags/{id} — review flag (dismiss or take action)
- [ ] Write tests for each rule
- [ ] Run `go test ./... -v` — must pass

### Task 13.2: Chat Content Filtering (FR-063)

**Files:**
- Create: `internal/antifraud/chat_filter.go`
- Modify: `internal/service/chat_service.go`

- [ ] ChatFilter:
  ```go
  type ChatFilter interface {
      Filter(ctx context.Context, text string) (filtered string, wasFiltered bool, detections []string)
  }
  ```
- [ ] Detection patterns (Russian formats):
  - Phone: +7XXXXXXXXXX, 8XXXXXXXXXX, 8(XXX)XXX-XX-XX, and variations
  - Email: standard email regex
  - URLs: http/https, www., telegram.me, t.me, vk.com, wa.me
  - Messengers: "напиши в вотсап", "мой телеграм", etc. (keyword matching)
- [ ] Replace detected content with "[контактные данные скрыты]"
- [ ] Integrate into chat message sending:
  - Before saving message, run through filter
  - Save filtered version
  - Log original + detections for admin review
- [ ] Admin can view filtered messages:
  - GET /api/v1/admin/chat/filtered — list messages with filtered content
- [ ] Write tests with various Russian phone/email/URL formats
- [ ] Run `go test ./... -v` — must pass

### Task 13.3: Enhanced Admin Audit Log

**Files:**
- Create: `internal/middleware/admin_audit.go`
- Modify: `internal/service/audit_log_service.go` (from Subsystem 4)
- Modify: `internal/server/router.go`

- [ ] Create AdminAuditMiddleware for admin routes:
  - Intercept all POST/PUT/PATCH/DELETE requests under admin routes
  - Log: admin UserID, HTTP method, path, target entity (from URL), timestamp, IP, request body summary (redact sensitive fields)
  - Store as AuditLog entries with EntityType="admin_action"
- [ ] Modify AuditLogService to support admin action queries:
  - ListAdminActions(ctx, filter) — filter by admin user, action type, date range
- [ ] GET /api/v1/admin/audit-log — enhanced with admin action filters
  - Query params: admin_id, action, entity_type, date_from, date_to, page, page_size
  - Response: paginated audit log entries
- [ ] Ensure sensitive data is not logged:
  - Redact passwords, tokens, bank details from request body
  - Keep: IDs, statuses, amounts, reasons
- [ ] Write tests
- [ ] Run `go test ./... -v -race` — must pass
- [ ] Run linter: `make lint`
