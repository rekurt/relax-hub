# Architecture

**Analysis Date:** 2026-04-17

## Pattern Overview

**Overall:** Clean Architecture with dependency injection via Uber fx.

**Key Characteristics:**
- Strict separation: `handler` → `service` → `repository` layers
- Domain models in `internal/domain/` (no external dependencies)
- All repository interfaces in single file: `internal/repository/interfaces.go`
- Uber fx for DI: each subsystem provides a `Module` with bindings
- Strongly typed errors in domain layer for HTTP status mapping
- Context-based auth and RBAC via middleware
- PostgreSQL + Redis for persistence

## Layers

**Handler Layer:**
- Purpose: HTTP request/response handling, input validation, RBAC enforcement
- Location: `internal/handler/`
- Contains: One file per entity (e.g., `booking_handler.go`, `bathhouse_handler.go`)
- Depends on: Service layer, middleware (auth context), logger
- Used by: Server router (`internal/server/router.go`)

**Service Layer:**
- Purpose: Business logic, validation, orchestration, RBAC checks via `AccessChecker`
- Location: `internal/service/`
- Contains: One service per domain aggregate (e.g., `BookingService`, `BathhouseService`)
- Depends on: Repository interfaces, domain models, external services (payment, notification, etc.)
- Used by: Handlers, cron jobs, bot handlers
- Pattern: Constructor injection of repositories + external services; `CreateX(ctx, input)` methods

**Repository Layer:**
- Purpose: Database queries, transactions, persistence
- Location: `internal/repository/`
  - `internal/repository/interfaces.go` — all 50+ repository interfaces in one file
  - `internal/repository/postgres/` — pgx implementations (one file per entity)
  - `internal/repository/mock/` — in-memory mocks for unit testing
- Depends on: Domain models only
- Used by: Service layer

**Domain Layer:**
- Purpose: Core models, business rules, error types
- Location: `internal/domain/`
- Contains: Entity models (`user.go`, `booking.go`, `bathhouse.go`), filters, value objects, error definitions
- Depends on: Go stdlib only
- Used by: All layers

## Data Flow

**HTTP Request (Booking Creation Example):**

1. **Router** (`internal/server/router.go`) — HTTP handler receives `POST /api/v1/bookings`
2. **Middleware** (`internal/middleware/auth.go`) — `RequireAuth` extracts JWT token → sets userID/role in context
3. **Handler** (`internal/handler/booking_handler.go`) — parses request body, calls `BookingService.Create(ctx, input)`
4. **Service** (`internal/service/booking_service.go`) — performs RBAC check, calculates pricing (dynamic rules + discounts), validates availability via repo, calls payment service, calls repository
5. **Repository** (`internal/repository/postgres/booking_repo.go`) — uses `CreateWithAvailabilityCheck()` with `pg_advisory_xact_lock` for race condition safety, executes INSERT + creates related records (booking_addons, hold)
6. **Response** — handler receives `BookingResult{Booking, AddOns, Prices}`, encodes to JSON with `APIResponse{Success: true, Data: ...}`

**State Management:**
- **Auth state:** JWT token in Authorization header, parsed by auth middleware, userID/role stored in context
- **Session state:** Optional `SessionValidator` middleware validates session still active (via Redis/DB)
- **Wallet/hold state:** Atomic transactions with locking (pg_advisory locks for race conditions, SERIALIZABLE isolation where needed)
- **Distributed jobs:** Redis-backed cron with distributed locking (`SETNX`) to prevent duplicate execution across instances

## Key Abstractions

**Service Layer Patterns:**

1. **Input/Result Types** — Each major operation has input and output structs:
   - `CreateBookingInput` + `BookingResult`
   - `TopUpWalletInput` + success/error
   - Decouples handler from service implementation

2. **AccessChecker** — RBAC abstraction in service layer:
   - `accessChecker.CanManageBathhouse(ctx, bathhouseID)` — checks ownership
   - `accessChecker.CanManageBooking(ctx, bookingID)` — checks role + ownership
   - Used by all services to enforce multi-role authorization (client, owner, representative, admin)

3. **Payment Provider** — Multi-region payment abstraction:
   - `internal/payment/provider_factory.go` — selects YooKassa (RU) or bePaid (BY) by user region
   - `PaymentProvider.Charge()`, `Refund()`, `CreateHold()` — consistent interface
   - Integration points in `BookingService.Create()`, `PaymentHandler.Confirm()`

4. **Notification Dispatcher** — Multi-channel notification:
   - `NotificationDispatcher.Dispatch(ctx, notif)` — routes by user preferences
   - Channels: in-app (`NotificationRepository`), email (`EmailSender`), SMS (`SMSProvider`), Telegram (`TelegramSender`)
   - Respects user preferences and fallback chain

5. **Pricing Engine** — Layered price calculation:
   - Dynamic pricing rules per hour + day-of-week multipliers
   - Long-session discounts (>N hours), extra guest surcharges, holiday pricing, seasonal tariffs, last-minute discounts
   - Applied in `BookingService.CalculatePrice()` with additive breakdown
   - Service fee applied separately per region (platform fee on base price, configurable)

6. **Anti-Fraud Engine** — Rule-based fraud detection:
   - Client rules: multi-card top-ups, top-up/cancel cycles, rapid bookings
   - Owner rules: self-booking, structuring, fake reviews
   - Listing rules: duplicate detection by phone/email/payment details
   - Actions: BLOCK (prevents operation), FLAG (allows + notifies admin), FREEZE_WALLET
   - Hooked into wallet/booking/payout/bathhouse services

7. **CRM Guest Cards** — Auto-created on booking completion:
   - Visit count, LTV (lifetime value), last visit date, average check, notes, tags
   - Dynamic segments: new (1 visit), regular (>=3), lost (>90 days), vip (>50k RUB)
   - RFM scoring (Recency-Frequency-Monetary, 1-5 scale each)
   - Broadcasts to segments with rate limiting
   - Auto-scenarios (thank, request review, reactivate lost)

## Entry Points

**HTTP Server Entry Point:**
- Location: `cmd/server/main.go` → `Execute()` → `serveCmd.RunE`
- Loads config via `config.Load()` (Viper, env prefix `BANI_`)
- Creates fx App via `app.New(cfg)` with all modules
- Registers server lifecycle hooks: `RegisterServer()` binds HTTP server to fx lifecycle
- Graceful shutdown: 30-second timeout, signal handling for SIGINT/SIGTERM

**Telegram Bot Entry Point:**
- Location: `cmd/bot/main.go` → `cmd/bot/run.go`
- Separate binary (`go run ./cmd/bot run`)
- Reuses same config loading, DI, database connection
- Implements booking wizard with in-memory state machine

**Router Entry Point:**
- Location: `internal/server/router.go` → `NewRouter(params RouterParams)`
- Chi router v5 with middleware stack:
  1. RequestID (chi built-in)
  2. Logging (structured zap)
  3. Recovery (panic recovery)
  4. CORS (configurable allowed origins)
- Global routes: `/health`, `/ready`, `/sitemap.xml`, `/swagger/*` (dev only)
- Auth middleware applied per route group: `RequireAuth`, `OptionalAuth`, `RequireRole(roles...)`
- Rate limiters per operation: auth (login/register), widget, webhook, promo validation
- 120+ endpoint routes organized by resource

## Error Handling

**Strategy:** Domain-layer errors mapped to HTTP status codes in handler layer.

**Patterns:**

1. **Domain Errors** (`internal/domain/errors.go`):
   - Sentinel errors: `var ErrNotFound = errors.New("not found")`
   - Checked with `errors.Is(err, domain.ErrNotFound)`

2. **Error → Status Code Mapping** (`internal/handler/response.go`):
   ```go
   switch {
   case errors.Is(err, domain.ErrNotFound):
       status = http.StatusNotFound
   case errors.Is(err, domain.ErrUnauthorized):
       status = http.StatusUnauthorized
   case errors.Is(err, domain.ErrForbidden):
       status = http.StatusForbidden
   case errors.Is(err, domain.ErrAlreadyExists):
       status = http.StatusConflict
   }
   ```

3. **API Error Response**:
   ```json
   {
     "success": false,
     "error": {
       "code": "NOT_FOUND",
       "message": "bathhouse not found"
     }
   }
   ```

4. **Service Layer Error Wrapping**:
   - Services return domain errors or wrapped errors with context
   - Handlers convert to API responses with appropriate HTTP status

## Cross-Cutting Concerns

**Logging:** Structured logging via `internal/logger/` (zap wrapper)
- Config: `BANI_LOGGER_LEVEL` (default `info`), `BANI_LOGGER_FORMAT` (`json` or `console`)
- Injected via fx in all layers
- Pattern: `logger.Info("msg", "key", value)`, `logger.Error("msg", "error", err)`

**Validation:**
- Input validation in handlers (e.g., email format, required fields)
- Business rule validation in services (e.g., booking availability, pricing constraints)
- Domain models may have validation methods

**Authentication:**
- JWT tokens (HS256) parsed by `AuthService.ParseToken()`
- Context-based: `middleware.GetUserID(ctx)`, `middleware.GetUserRole(ctx)`
- Session validation: optional `SessionValidator` middleware touches Redis on each request
- Two-factor auth: TOTP + SMS, checked before session creation

**Authorization (RBAC):**
- Middleware: `RequireRole([]*roles)`, `RequireAdminPermission(perm)`
- Service layer: `AccessChecker.Can*()` methods for resource-level checks
- Four base roles: client, owner, representative, admin
- Six admin sub-roles: super_admin, moderator, support_l1/l2/l3, finance
- Mandatory 2FA for all admin sub-roles

**Caching:**
- Redis via `internal/storage/` — holds recently viewed, search suggestions, feature flags, platform settings, loyalty tiers
- In-memory caching in services (e.g., last 20 recently viewed per user)
- TTL varies: 1 min (feature flags), 5 min (platform settings), 1h (search suggestions, area avg price)

**Database Transactions:**
- Booking creation uses `pg_advisory_xact_lock` + `CreateWithAvailabilityCheck()` to prevent double-booking
- Wallet operations use optimistic concurrency control (version field) with retry logic
- Some operations use SERIALIZABLE isolation level for strong consistency

---

*Architecture analysis: 2026-04-17*
