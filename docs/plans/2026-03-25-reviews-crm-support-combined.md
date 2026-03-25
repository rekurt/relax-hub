# Combined Plan: Reviews & Rating + CRM for Owners + Support & Disputes

## Overview

Three interconnected subsystems: (1) Reviews & Rating enhancements with multi-criteria ratings, Bayesian averages, auto review requests, quality monitoring, and NLP auto-moderation; (2) CRM for Owners with guest cards, segments, broadcasts, auto-scenarios, and response templates; (3) Support & Disputes with ticket system, escalation, CSAT surveys, and dispute resolution with evidence collection and appeals.

## Context

- Existing reviews: `internal/service/review_service.go`, `internal/handler/review_handler.go`, `internal/domain/review.go`
- Existing booking: `internal/service/booking_service.go`
- Existing chat: `internal/service/chat_service.go`
- Existing notifications: `internal/notification/`
- Existing complaints: `internal/service/complaint_service.go`
- Existing escrow: `internal/service/escrow_service.go`
- Wallet system (existing) for refunds/compensation
- Related patterns: handler→service→repository, fx DI modules, domain errors → HTTP status mapping
- Dependencies: Escrow system (existing) — disputes block escrow release; Wallet system (existing) — for refunds

## Development Approach

- **Testing approach**: Regular (code first, then tests)
- Complete each task fully before moving to the next
- Subsystems are ordered by dependency: Reviews first (no deps), CRM second (uses booking data), Support & Disputes third (uses wallet + escrow)
- **CRITICAL: every task MUST include new/updated tests**
- **CRITICAL: all tests must pass before starting next task**

## Implementation Steps

### Task 1: Multi-criteria Rating (Reviews)

**Files:**
- Modify: `internal/domain/review.go`
- Modify: `internal/service/review_service.go`
- Modify: `internal/handler/review_handler.go`
- Modify: `internal/repository/postgres/review_repo.go`
- Create: `migrations/XXXXXX_review_criteria.up.sql`

- [x] Add criteria fields to Review model: Cleanliness, Accuracy, Communication, ValueForMoney (float64, 1.0-5.0, step 0.5)
- [x] Overall Rating = average of 4 criteria (rounded to 1 decimal)
- [x] Migration: add `cleanliness DECIMAL(2,1)`, `accuracy DECIMAL(2,1)`, `communication DECIMAL(2,1)`, `value_for_money DECIMAL(2,1)` to reviews
- [x] Modify review creation: require all 4 criteria (or overall rating for backwards compat)
- [x] Modify review response: include criteria breakdown
- [x] Modify bathhouse detail: include average per criteria (avg_cleanliness, avg_accuracy, avg_communication, avg_value_for_money)
- [x] Write tests
- [x] Run `go test ./... -v` — must pass

### Task 2: Bayesian Average Rating (Reviews)

**Files:**
- Modify: `internal/service/review_service.go`
- Modify: `internal/repository/postgres/review_repo.go`
- Modify: `internal/domain/bathhouse.go`

- [x] Implement Bayesian average: R_bayesian = (n * R + m * C) / (n + m), where m = min reviews threshold (default 5), C = platform-wide average
- [x] Display rules: < 3 reviews show "Новое" badge, >= 3 show Bayesian average
- [x] Recalculate on every review create/update/delete; cache platform average C in Redis (recalc daily)
- [x] Add BayesianRating field to Bathhouse
- [x] Write tests with edge cases (0, 1, many reviews)
- [x] Run `go test ./... -v` — must pass

### Task 3: Auto Review Request (Reviews)

**Files:**
- Create: `internal/cron/review_jobs.go`
- Modify: `internal/service/review_service.go`

- [x] Cron job (hourly): find completed bookings with check-out + configurable delay (BANI_REVIEW_REQUEST_DELAY_HOURS, default 2), no existing review, no request sent (Redis dedup)
- [x] Send push + email notification with pre-filled data (booking ID, bathhouse name)
- [x] Max 1 review request per booking
- [x] Write tests
- [x] Run `go test ./... -v` — must pass

### Task 4: Quality Monitoring & Badges (Reviews)

**Files:**
- Modify: `internal/service/review_service.go`
- Modify: `internal/service/bathhouse_service.go`
- Modify: `internal/domain/bathhouse.go`

- [x] On rating recalculation: rating < 3.0 (>= 10 reviews) → warn owner; rating < 2.0 (>= 10 reviews) → auto-depublish + notify owner + admin
- [x] Computed badges: "Verified" (photos moderated + KYC), "Top" (Bayesian >= 4.5 AND count >= 10), "Premium" (active subscription), "New" (< 30 days AND < 3 reviews)
- [x] Add Badges ([]string) to bathhouse response (computed, not stored)
- [x] Write tests
- [x] Run `go test ./... -v` — must pass

### Task 5: NLP Auto-moderation Placeholder (Reviews)

**Files:**
- Create: `internal/moderation/text_moderator.go`
- Modify: `internal/service/review_service.go`

- [x] Define TextModerationService interface with Analyze(ctx, text) -> ModerationResult (Flagged, Score, Flags)
- [x] Regex-based v1: profanity filter, URL/phone/email detection, spam patterns (repeated chars, ALL CAPS > 50%)
- [x] Integration: score > 0.7 → set review to "pending_moderation"; otherwise auto-publish
- [x] Store moderation result for admin reference
- [x] Write tests with various text samples
- [x] Run `go test ./... -v -race` — must pass

### Task 6: Guest Cards (CRM)

**Files:**
- Create: `internal/domain/guest_card.go`
- Create: `internal/repository/postgres/guest_card_repo.go`
- Create: `internal/repository/mock/guest_card_repo.go`
- Create: `internal/service/guest_card_service.go`
- Create: `internal/handler/guest_card_handler.go`
- Modify: `internal/repository/interfaces.go`
- Modify: `internal/server/router.go`
- Create: `migrations/XXXXXX_guest_cards.up.sql`

- [x] GuestCard model: ID, OwnerID, ClientID, BathhouseID, FirstVisitAt, LastVisitAt, VisitCount, TotalSpent, AvgCheck, Notes, Tags, timestamps
- [x] GuestCardRepository: Upsert, GetByOwnerAndClient, ListByOwner (search, tag filter, date range, LTV sort), UpdateNotes, GetStats
- [x] GuestCardService: RecordVisit (on booking completion), ListGuests, GetGuestDetail, UpdateGuestNotes, ExportCSV
- [x] Hook into booking completion: call RecordVisit
- [x] Endpoints (RequireRole: owner/representative): GET/PUT /api/v1/my/crm/guests, GET /api/v1/my/crm/guests/export, GET /api/v1/my/crm/stats
- [x] Write tests
- [x] Run `go test ./... -v` — must pass

### Task 7: Segments (CRM)

**Files:**
- Modify: `internal/service/guest_card_service.go`
- Modify: `internal/handler/guest_card_handler.go`

- [x] Predefined segments (dynamic): "new" (1 visit), "regular" (>= 3), "lost" (> 90 days), "vip" (> 50,000 RUB), "birthday_soon" (7 days)
- [x] SegmentService: ListSegments (with counts), GetGuestsInSegment
- [x] Endpoints: GET /api/v1/my/crm/segments, GET /api/v1/my/crm/segments/{slug}/guests
- [x] Write tests
- [x] Run `go test ./... -v` — must pass

### Task 8: Broadcasts (CRM)

**Files:**
- Create: `internal/domain/broadcast.go`
- Create: `internal/service/broadcast_service.go`
- Create: `internal/handler/broadcast_handler.go`
- Create: `internal/repository/postgres/broadcast_repo.go`
- Modify: `internal/repository/interfaces.go`
- Create: `migrations/XXXXXX_broadcasts.up.sql`

- [x] Broadcast model: ID, OwnerID, Segment, Title, Body, ImageURL, PromoCodeID, Channels, Status (draft/sending/sent/failed), Stats (Delivered, Read)
- [x] BroadcastRepository: Create, GetByID, ListByOwner, UpdateStatus, UpdateStats
- [x] BroadcastService: Create, Send (rate limit: 3/week per owner, 1/3 days per guest), ListBroadcasts, TrackDelivery
- [x] Endpoints: POST/GET /api/v1/my/crm/broadcasts, POST /api/v1/my/crm/broadcasts/{id}/send, GET /api/v1/my/crm/broadcasts/{id}
- [x] Respect user notification preferences (unsubscribe)
- [x] Write tests
- [x] Run `go test ./... -v` — must pass

### Task 9: Auto-scenarios (CRM)

**Files:**
- Create: `internal/domain/auto_scenario.go`
- Create: `internal/service/auto_scenario_service.go`
- Create: `internal/handler/auto_scenario_handler.go`
- Create: `internal/repository/postgres/auto_scenario_repo.go`
- Modify: `internal/repository/interfaces.go`
- Create: `migrations/XXXXXX_auto_scenarios.up.sql`

- [x] Predefined scenarios: thank_after_visit (1h), request_review (2h), remind_revisit_30d, reactivate_lost_90d, birthday_greeting
- [x] AutoScenario model: ID, OwnerID, Type, Enabled, CustomText, Channel, DelayHours, PromoCodeID, timestamps
- [x] AutoScenarioRepository: Upsert, ListByOwner, GetByOwnerAndType
- [x] AutoScenarioService: ListScenarios, UpdateScenario, ExecuteScenarios (hourly cron)
- [x] Endpoints: GET /api/v1/my/crm/auto-scenarios, PUT /api/v1/my/crm/auto-scenarios/{type}
- [x] Write tests
- [x] Run `go test ./... -v` — must pass

### Task 10: Response Templates (CRM)

**Files:**
- Create: `internal/domain/response_template.go`
- Create: `internal/service/template_service.go`
- Create: `internal/handler/template_handler.go`
- Create: `internal/repository/postgres/template_repo.go`
- Modify: `internal/repository/interfaces.go`
- Create: `migrations/XXXXXX_response_templates.up.sql`

- [x] ResponseTemplate model: ID, OwnerID, Title, Body, IsDefault, SortOrder, CreatedAt
- [x] TemplateRepository: Create, Update, Delete, ListByOwner, GetByID
- [x] TemplateService: Create, Update, Delete, List, SeedDefaults (3 default templates on first use)
- [x] Endpoints: POST/GET/PUT/DELETE /api/v1/my/crm/templates
- [x] Write tests
- [x] Run `go test ./... -v` — must pass

### Task 11: Support Tickets

**Files:**
- Create: `internal/domain/ticket.go`
- Create: `internal/repository/postgres/ticket_repo.go`
- Create: `internal/repository/mock/ticket_repo.go`
- Create: `internal/service/ticket_service.go`
- Create: `internal/handler/ticket_handler.go`
- Modify: `internal/repository/interfaces.go`
- Modify: `internal/server/router.go`
- Create: `migrations/XXXXXX_support_tickets.up.sql`

- [x] Ticket model: ID, UserID, BookingID, Category (question/problem/complaint/refund_request/account_issue), Status (open/in_progress/escalated/resolved/closed), Priority (low/medium/high/critical), Level (L1/L2/L3), Subject, AssignedTo, CSATScore, timestamps
- [x] TicketMessage model: ID, TicketID, SenderID, SenderType, Body, Attachments, CreatedAt
- [x] TicketRepository: Create, GetByID, ListByUser, ListAll, UpdateStatus, Assign, AddMessage, ListMessages, CountByStatus
- [x] TicketService: CreateTicket (auto-priority by category), GetTicket, ListUserTickets, AddMessage, AssignTicket, EscalateTicket, ResolveTicket, CloseTicket (auto 7 days), SubmitCSAT
- [x] User endpoints: POST/GET /api/v1/my/tickets, GET /api/v1/my/tickets/{id}, POST messages, POST csat
- [x] Admin endpoints: GET /api/v1/admin/tickets, PATCH assign/escalate/resolve, POST messages, GET stats
- [x] Auto-escalation: no response 24h → L2, 48h at L2 → L3, critical → auto L2
- [x] CSAT survey notification 24h after resolution
- [x] Write tests
- [x] Run `go test ./... -v` — must pass

### Task 12: Disputes

**Files:**
- Create: `internal/domain/dispute.go`
- Create: `internal/service/dispute_service.go`
- Create: `internal/handler/dispute_handler.go`
- Create: `internal/repository/postgres/dispute_repo.go`
- Create: `internal/repository/mock/dispute_repo.go`
- Modify: `internal/repository/interfaces.go`
- Create: `migrations/XXXXXX_disputes.up.sql`

- [x] Dispute model: ID, BookingID, InitiatorID, RespondentID, Reason (service_not_provided/poor_quality/damage/safety_issue/billing_error/other), Status (open/evidence_collection/under_review/resolved/appealed/closed), Resolution, RefundAmount, CompensationAmount, MediatorID, MediatorNotes, AppealDeadline (7 days)
- [x] DisputeEvidence model: ID, DisputeID, UserID, Type (photo/screenshot/gps/message/receipt), URL, Description
- [x] DisputeRepository: Create, GetByID, GetByBookingID, UpdateStatus, AddEvidence, ListEvidence, ListAll
- [x] DisputeService: OpenDispute (block escrow), SubmitEvidence (72h window), Resolve (full/partial/no refund, compensation), Appeal (within 7 days), CloseDispute
- [x] User endpoints: POST /api/v1/bookings/{id}/dispute, POST evidence, GET detail, POST appeal, GET /api/v1/my/disputes
- [x] Admin endpoints: GET /api/v1/admin/disputes, PATCH assign, PATCH resolve
- [x] Write tests for full dispute lifecycle
- [x] Run `go test ./... -v -race` — must pass

### Task 13: Verify acceptance criteria

- [x] Manual test: create review with multi-criteria, verify Bayesian rating recalculation
- [x] Manual test: create guest card via booking completion, verify segment assignment
- [x] Manual test: open support ticket, escalate, resolve, submit CSAT
- [x] Manual test: open dispute on booking, submit evidence, resolve, test appeal
- [x] Run full test suite: `go test ./... -v -race`
- [x] Run linter: `make lint`
- [x] Verify test coverage meets 80%+

### Task 14: Update documentation

- [ ] Update CLAUDE.md: add Reviews multi-criteria, CRM subsystem, Support & Disputes subsystem descriptions
- [ ] Run `make swagger` to regenerate OpenAPI spec with all new endpoints
- [ ] Move this plan to `docs/plans/completed/`
