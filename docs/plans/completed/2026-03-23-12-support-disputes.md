---
# Subsystem 12: Support & Disputes (FR-155)

## Overview
Full support ticket system with L1/L2/L3 escalation, ticket messaging, CSAT surveys, and dispute resolution system for bookings with evidence collection, mediator review, and appeals.

## Context
- Existing complaints: `internal/service/complaint_service.go` — simple report mechanism
- Existing admin moderation: review/bathhouse moderation
- Existing booking: has cancellation and refund flows
- No support ticket system exists
- No dispute resolution system exists

## Dependencies
- Depends on: Wallet System (Subsystem 1) — for refunds/compensation during dispute resolution
- Depends on: Booking system (existing) — disputes are tied to bookings
- Depends on: Escrow (Subsystem 9, Task 9.4) — disputes block escrow release

## Development Approach
- **Testing approach**: Regular (code first, then tests)
- **CRITICAL: every task MUST include new/updated tests**
- **CRITICAL: all tests must pass before starting next task**

## Implementation Steps

### Task 12.1: Support Tickets

**Files:**
- Create: `internal/domain/ticket.go`
- Create: `internal/repository/postgres/ticket_repo.go`
- Create: `internal/repository/mock/ticket_repo.go`
- Create: `internal/service/ticket_service.go`
- Create: `internal/handler/ticket_handler.go`
- Modify: `internal/repository/interfaces.go`
- Modify: `internal/server/router.go`
- Create: `migrations/XXXXXX_support_tickets.up.sql`

- [ ] Ticket model:
  - ID (uuid), UserID (uuid), BookingID (*uuid)
  - Category (enum: question/problem/complaint/refund_request/account_issue)
  - Status (enum: open/in_progress/escalated/resolved/closed)
  - Priority (enum: low/medium/high/critical)
  - Level (enum: L1/L2/L3)
  - Subject (string), AssignedTo (*uuid — admin user)
  - CreatedAt, UpdatedAt, ResolvedAt (*time.Time)
  - CSATScore (*int — 1-5, filled after resolution)
- [ ] TicketMessage model:
  - ID, TicketID (uuid), SenderID (uuid)
  - SenderType (enum: user/support/system)
  - Body (string), Attachments ([]string — URLs)
  - CreatedAt
- [ ] TicketRepository:
  - Create, GetByID, ListByUser, ListAll (admin), UpdateStatus, Assign, AddMessage, ListMessages
  - CountByStatus (for admin dashboard)
- [ ] TicketService:
  - CreateTicket(ctx, userID, input) — create with auto-priority based on category
    - refund_request → high
    - complaint → medium
    - question → low
  - GetTicket(ctx, userID, ticketID) — with messages
  - ListUserTickets(ctx, userID, filter)
  - AddMessage(ctx, userID, ticketID, body, attachments)
  - AssignTicket(ctx, adminID, ticketID, assigneeID) — admin
  - EscalateTicket(ctx, adminID, ticketID) — L1→L2→L3
  - ResolveTicket(ctx, adminID, ticketID, resolution) — set resolved
  - CloseTicket(ctx, ticketID) — auto-close 7 days after resolution if no response
  - SubmitCSAT(ctx, userID, ticketID, score)
- [ ] User endpoints:
  - POST /api/v1/my/tickets — create ticket
  - GET /api/v1/my/tickets — list user tickets
  - GET /api/v1/my/tickets/{id} — ticket detail with messages
  - POST /api/v1/my/tickets/{id}/messages — add message
  - POST /api/v1/my/tickets/{id}/csat — submit CSAT score
- [ ] Admin endpoints:
  - GET /api/v1/admin/tickets — list all tickets with filters (status, priority, level, assigned)
  - GET /api/v1/admin/tickets/{id} — ticket detail
  - PATCH /api/v1/admin/tickets/{id}/assign — assign to admin
  - PATCH /api/v1/admin/tickets/{id}/escalate — escalate level
  - PATCH /api/v1/admin/tickets/{id}/resolve — resolve with message
  - POST /api/v1/admin/tickets/{id}/messages — admin reply
  - GET /api/v1/admin/tickets/stats — dashboard stats
- [ ] Auto-escalation rules:
  - No response in 24h → escalate L1→L2
  - No response in 48h at L2 → escalate L2→L3
  - Critical priority → auto-assign to L2
- [ ] CSAT survey: send notification 24h after resolution
- [ ] Write tests
- [ ] Run `go test ./... -v` — must pass

### Task 12.2: Disputes

**Files:**
- Create: `internal/domain/dispute.go`
- Create: `internal/service/dispute_service.go`
- Create: `internal/handler/dispute_handler.go`
- Create: `internal/repository/postgres/dispute_repo.go`
- Create: `internal/repository/mock/dispute_repo.go`
- Modify: `internal/repository/interfaces.go`
- Create: `migrations/XXXXXX_disputes.up.sql`

- [ ] Dispute model:
  - ID (uuid), BookingID (uuid)
  - InitiatorID (uuid — who opened), RespondentID (uuid — other party)
  - Reason (enum: service_not_provided/poor_quality/damage/safety_issue/billing_error/other)
  - Description (string)
  - Status (enum: open/evidence_collection/under_review/resolved/appealed/closed)
  - Resolution (*enum: full_refund/partial_refund/no_refund/compensation)
  - RefundAmount (*int64), CompensationAmount (*int64)
  - MediatorID (*uuid — admin who handles)
  - MediatorNotes (string)
  - CreatedAt, ResolvedAt (*time.Time)
  - AppealDeadline (*time.Time — 7 days after resolution)
- [ ] DisputeEvidence model:
  - ID, DisputeID (uuid), UserID (uuid)
  - Type (enum: photo/screenshot/gps/message/receipt)
  - URL (string), Description (string)
  - CreatedAt
- [ ] DisputeRepository: Create, GetByID, GetByBookingID, UpdateStatus, AddEvidence, ListEvidence, ListAll (admin)
- [ ] DisputeService:
  - OpenDispute(ctx, userID, bookingID, reason, description) — validate within claim period
    - Block escrow release (mark escrow as disputed)
    - Notify other party
    - Status: open → evidence_collection
  - SubmitEvidence(ctx, userID, disputeID, evidence) — within 72h evidence window
  - Resolve(ctx, adminID, disputeID, resolution, refundAmount, notes)
    - full_refund: refund full amount to client wallet
    - partial_refund: refund specified amount
    - no_refund: release escrow to owner
    - compensation: refund + additional compensation from platform
  - Appeal(ctx, userID, disputeID, reason) — within 7 days of resolution
    - Re-opens dispute, escalates to L3
  - CloseDispute(ctx, disputeID) — final close after appeal deadline
- [ ] User endpoints:
  - POST /api/v1/bookings/{id}/dispute — open dispute (RequireAuth)
  - POST /api/v1/disputes/{id}/evidence — submit evidence
  - GET /api/v1/disputes/{id} — get dispute details
  - POST /api/v1/disputes/{id}/appeal — appeal resolution
  - GET /api/v1/my/disputes — list user's disputes
- [ ] Admin endpoints:
  - GET /api/v1/admin/disputes — list all disputes with filters
  - GET /api/v1/admin/disputes/{id} — dispute detail with all evidence
  - PATCH /api/v1/admin/disputes/{id}/assign — assign mediator
  - PATCH /api/v1/admin/disputes/{id}/resolve — resolve dispute
- [ ] Write tests for full dispute lifecycle
- [ ] Run `go test ./... -v -race` — must pass
- [ ] Run linter: `make lint`
