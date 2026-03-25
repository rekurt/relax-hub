# Support, Disputes, Anti-fraud & Cron Jobs (Subsystems 12-14)

## Overview
Combined implementation plan covering: support ticket system with escalation and CSAT, booking dispute resolution with evidence and appeals, anti-fraud rules engine with chat content filtering, admin audit logging, and centralized cron scheduler for all background tasks.

## Context
- Files involved:
  - Domain: `internal/domain/ticket.go`, `internal/domain/dispute.go`, `internal/domain/fraud_flag.go`
  - Repository: `internal/repository/interfaces.go`, `internal/repository/postgres/ticket_repo.go`, `internal/repository/postgres/dispute_repo.go`, `internal/repository/postgres/fraud_flag_repo.go`
  - Mocks: `internal/repository/mock/ticket_repo.go`, `internal/repository/mock/dispute_repo.go`
  - Services: `internal/service/ticket_service.go`, `internal/service/dispute_service.go`
  - Handlers: `internal/handler/ticket_handler.go`, `internal/handler/dispute_handler.go`
  - Anti-fraud: `internal/antifraud/engine.go`, `internal/antifraud/rules.go`, `internal/antifraud/chat_filter.go`, `internal/antifraud/module.go`
  - Audit: `internal/middleware/admin_audit.go`
  - Cron: `internal/cron/scheduler.go`, `internal/cron/module.go`, `internal/cron/*_jobs.go`
  - Router: `internal/server/router.go`
  - App: `internal/app/app.go`
- Related patterns: handler->service->repository, fx modules, mock repos for testing
- Dependencies: Wallet System, Booking system, Escrow, Chat service, Audit Log (Subsystem 4)

## Development Approach
- **Testing approach**: Regular (code first, then tests)
- Complete each task fully before moving to the next
- **CRITICAL: every task MUST include new/updated tests**
- **CRITICAL: all tests must pass before starting next task**

## Implementation Steps

### Task 1: Support Tickets — Domain, Repository, Migration

**Files:**
- Create: `internal/domain/ticket.go`
- Create: `internal/repository/postgres/ticket_repo.go`
- Create: `internal/repository/mock/ticket_repo.go`
- Modify: `internal/repository/interfaces.go`
- Create: `migrations/XXXXXX_support_tickets.up.sql`

- [x] Ticket model: ID, UserID, BookingID(*uuid), Category(question/problem/complaint/refund_request/account_issue), Status(open/in_progress/escalated/resolved/closed), Priority(low/medium/high/critical), Level(L1/L2/L3), Subject, AssignedTo(*uuid), CreatedAt, UpdatedAt, ResolvedAt, CSATScore(*int 1-5)
- [x] TicketMessage model: ID, TicketID, SenderID, SenderType(user/support/system), Body, Attachments([]string), CreatedAt
- [x] Migration: tickets and ticket_messages tables with indexes
- [x] TicketRepository interface + postgres implementation: Create, GetByID, ListByUser, ListAll, UpdateStatus, Assign, AddMessage, ListMessages, CountByStatus
- [x] Mock repository for testing
- [x] Write tests for repository
- [x] Run `go test ./... -v` — must pass

### Task 2: Support Tickets — Service & Handler

**Files:**
- Create: `internal/service/ticket_service.go`
- Create: `internal/handler/ticket_handler.go`
- Modify: `internal/server/router.go`

- [x] TicketService: CreateTicket (auto-priority by category), GetTicket, ListUserTickets, AddMessage, AssignTicket, EscalateTicket (L1->L2->L3), ResolveTicket, CloseTicket, SubmitCSAT
- [x] Auto-escalation: no response 24h -> L2, 48h at L2 -> L3, critical -> auto L2
- [x] CSAT: notification 24h after resolution, score validation (1-5), ErrCSATAlreadySubmitted/ErrCSATNotResolved/ErrCSATInvalidScore
- [x] User endpoints: POST/GET /api/v1/my/tickets, GET /api/v1/my/tickets/{id}, POST .../messages, POST .../csat
- [x] Admin endpoints: GET /api/v1/admin/tickets, GET .../{id}, PATCH .../assign, PATCH .../escalate, PATCH .../resolve, POST .../messages, GET .../stats
- [x] Wire into fx module and router
- [x] Write tests for service and handler
- [x] Run `go test ./... -v` — must pass

### Task 3: Disputes — Domain, Repository, Migration

**Files:**
- Create: `internal/domain/dispute.go`
- Create: `internal/repository/postgres/dispute_repo.go`
- Create: `internal/repository/mock/dispute_repo.go`
- Modify: `internal/repository/interfaces.go`
- Create: `migrations/XXXXXX_disputes.up.sql`

- [x] Dispute model: ID, BookingID, InitiatorID, RespondentID, Reason(service_not_provided/poor_quality/damage/safety_issue/billing_error/other), Description, Status(open/evidence_collection/under_review/resolved/appealed/closed), Resolution(*full_refund/partial_refund/no_refund/compensation), RefundAmount, CompensationAmount, MediatorID, MediatorNotes, CreatedAt, ResolvedAt, AppealDeadline
- [x] DisputeEvidence model: ID, DisputeID, UserID, Type(photo/screenshot/gps/message/receipt), URL, Description, CreatedAt
- [x] Migration: disputes, dispute_evidence tables with indexes
- [x] DisputeRepository: Create, GetByID, GetByBookingID, UpdateStatus, AddEvidence, ListEvidence, ListAll
- [x] Mock repository
- [x] Write tests
- [x] Run `go test ./... -v` — must pass

### Task 4: Disputes — Service & Handler

**Files:**
- Create: `internal/service/dispute_service.go`
- Create: `internal/handler/dispute_handler.go`
- Modify: `internal/server/router.go`

- [x] DisputeService: OpenDispute (validate claim period, block escrow, notify other party, status -> evidence_collection), SubmitEvidence (within 72h window), Resolve (full/partial/no refund or compensation via wallet), Appeal (within 7 days, re-opens to L3), CloseDispute
- [x] User endpoints: POST /api/v1/bookings/{id}/dispute, POST /api/v1/disputes/{id}/evidence, GET /api/v1/disputes/{id}, POST .../appeal, GET /api/v1/my/disputes
- [x] Admin endpoints: GET /api/v1/admin/disputes, GET .../{id}, PATCH .../assign, PATCH .../resolve
- [x] Wire into fx module and router
- [x] Write tests for full dispute lifecycle
- [x] Run `go test ./... -v` — must pass

### Task 5: Anti-fraud Rules Engine

**Files:**
- Create: `internal/antifraud/engine.go`
- Create: `internal/antifraud/rules.go`
- Create: `internal/antifraud/module.go`
- Create: `internal/domain/fraud_flag.go`
- Create: `internal/repository/postgres/fraud_flag_repo.go`
- Modify: `internal/repository/interfaces.go`
- Create: `migrations/XXXXXX_antifraud.up.sql`

- [x] FraudFlag model: ID, UserID, Rule, Severity(low/medium/high/critical), Status(pending/reviewed/dismissed/action_taken), Action(flag/freeze_wallet/block_user/notify_admin), Details(JSONB), CreatedAt, ReviewedAt, ReviewedBy
- [x] FraudFlagRepository: Create, ListByUser, ListPending, UpdateStatus, CountByUserAndRule
- [x] FraudEngine interface: CheckWalletTopUp, CheckBookingCreate, CheckBookingCancel, CheckPayoutRequest
- [x] Client rules: RULE_MULTI_CARD_TOPUP (>3 cards/24h), RULE_TOPUP_CANCEL_CYCLE (>2 cycles/7d), RULE_DORMANT_BALANCE (>50k no bookings/30d), RULE_RAPID_BOOKINGS (>5/hour)
- [x] Owner rules: RULE_SELF_BOOKING (owner books own bathhouse), RULE_STRUCTURING (>3 small withdrawals/day), RULE_FAKE_REVIEWS (same IP)
- [x] Hook into WalletService.TopUp, BookingService.CreateBooking, BookingService.CancelBooking, PayoutService.RequestPayout
- [x] Action handling: "block" prevents operation (ErrFraudDetected), "flag"/"freeze_wallet" allows but creates flag + notifies admin
- [x] Admin endpoints: GET /api/v1/admin/antifraud/flags, PATCH .../{id}
- [x] Write tests for each rule
- [x] Run `go test ./... -v` — must pass

### Task 6: Chat Content Filtering

**Files:**
- Create: `internal/antifraud/chat_filter.go`
- Modify: `internal/service/chat_service.go`

- [x] ChatFilter interface: Filter(ctx, text) -> (filtered, wasFiltered, detections)
- [x] Detection patterns (Russian formats): phone (+7/8 variations), email, URLs (http/https/www/t.me/vk.com/wa.me), messenger keywords ("напиши в вотсап", "мой телеграм")
- [x] Replace detected content with "[контактные данные скрыты]"
- [x] Integrate into chat message sending: filter before save, log original + detections
- [x] Admin endpoint: GET /api/v1/admin/chat/filtered
- [x] Write tests with various Russian phone/email/URL formats
- [x] Run `go test ./... -v` — must pass

### Task 7: Enhanced Admin Audit Log

**Files:**
- Create: `internal/middleware/admin_audit.go`
- Modify: `internal/service/audit_log_service.go`
- Modify: `internal/server/router.go`

- [ ] AdminAuditMiddleware: intercept POST/PUT/PATCH/DELETE on admin routes, log admin UserID, method, path, target entity, timestamp, IP, redacted request body
- [ ] Store as AuditLog entries with EntityType="admin_action"
- [ ] Modify AuditLogService: ListAdminActions with filters (admin_id, action, entity_type, date range)
- [ ] GET /api/v1/admin/audit-log with admin action filters
- [ ] Sensitive data redaction: strip passwords, tokens, bank details; keep IDs, statuses, amounts, reasons
- [ ] Write tests
- [ ] Run `go test ./... -v` — must pass

### Task 8: Cron Scheduler Infrastructure

**Files:**
- Create: `internal/cron/scheduler.go`
- Create: `internal/cron/module.go`
- Modify: `internal/app/app.go`

- [ ] CronScheduler struct wrapping robfig/cron: NewCronScheduler, Register(spec, name, job), Start, Stop
- [ ] Job wrapper: panic recovery, structured logging (start/end/duration/error), distributed lock via Redis SETNX with TTL, context with timeout
- [ ] Config: BANI_CRON_ENABLED (default true), BANI_CRON_TIMEZONE (default "Europe/Moscow")
- [ ] fx module: register scheduler, start on app start, stop on shutdown
- [ ] Write tests for scheduler infrastructure
- [ ] Run `go test ./... -v` — must pass

### Task 9: Register All Cron Jobs

**Files:**
- Create: `internal/cron/wallet_jobs.go`
- Create: `internal/cron/booking_jobs.go`
- Create: `internal/cron/review_jobs.go`
- Create: `internal/cron/crm_jobs.go`
- Create: `internal/cron/auth_jobs.go`
- Create: `internal/cron/escrow_jobs.go`
- Create: `internal/cron/antifraud_jobs.go`
- Create: `internal/cron/search_jobs.go`
- Create: `internal/cron/owner_jobs.go`
- Create: `internal/cron/ticket_jobs.go`

- [ ] Every 15 minutes: BookingRequestTimeout, NoShowDetection, BookingReminders
- [ ] Hourly: EscrowRelease, ReviewRequest, AutoScenarioExecution, AntiFraudPatternDetection
- [ ] Daily (midnight Moscow): BonusExpiration, BonusExpiryNotification, AccountDeletionExecution, SessionCleanup, KYCExpiryCheck, OwnerResponseRateMonitoring, PromoCodeDeactivation, SavedSearchNotification, BathhouseMetricsUpdate, PlatformAverageRating, TicketAutoClose
- [ ] Each job calls relevant service method, all use distributed locking
- [ ] Write tests for job registration
- [ ] Run `go test ./... -v -race` — must pass

### Task 10: Verify acceptance criteria

- [ ] Manual test: create support ticket, escalate, resolve, submit CSAT
- [ ] Manual test: open dispute on booking, submit evidence, admin resolves, appeal
- [ ] Manual test: trigger anti-fraud rule, verify flag created, admin reviews
- [ ] Manual test: send chat message with phone number, verify filtered
- [ ] Manual test: verify admin audit log captures admin actions
- [ ] Manual test: verify cron scheduler starts and runs jobs on schedule
- [ ] Run full test suite: `go test ./... -race`
- [ ] Run linter: `make lint`
- [ ] Verify build: `go build ./...`
- [ ] Verify test coverage meets 80%+

### Task 11: Update documentation

- [ ] Update CLAUDE.md if internal patterns changed
- [ ] Run `make swagger` to regenerate API docs
- [ ] Move this plan to `docs/plans/completed/`
