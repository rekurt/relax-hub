# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build & Dev Commands

```bash
# Build
go build ./...                              # build all
make build                                  # build server binary to ./bin/bani-server

# Test
go test ./... -v                            # all tests
go test ./internal/service/ -v -run TestBooking  # single test
make test                                   # shortcut
make test-hurl                              # hurl integration tests (requires running server)

# Lint & Vet
make lint                                   # golangci-lint
go vet ./...

# Database
make migrate-up                             # apply migrations
make migrate-down                           # rollback last migration
make seed-admin                             # create admin user (interactive)

# Docker
make docker-up                              # start postgres, redis, app
make docker-down

# Run
make run                                    # build + run server
go run ./cmd/bot run                        # run telegram bot
```

## Architecture

Clean architecture: **handler → service → repository**

- `internal/domain/` — models, errors, filters. No external dependencies.
- `internal/repository/interfaces.go` — all repository interfaces in one file
- `internal/repository/postgres/` — pgx implementations, one file per entity
- `internal/repository/mock/` — in-memory mocks for testing
- `internal/service/` — business logic, RBAC checks via AccessChecker
- `internal/handler/` — HTTP handlers, one file per entity
- `internal/middleware/` — auth (JWT), RBAC, CORS, logging, panic recovery
- `internal/server/` — chi router setup
- `internal/notification/` — dispatcher, email sender, WebSocket hub, telegram sender
- `internal/bot/` — Telegram bot (separate binary: `cmd/bot/`)
- `internal/app/` — Uber fx DI container

Two binaries: `cmd/server` (HTTP API, Cobra CLI) and `cmd/bot` (Telegram bot).

## Key Patterns

### DI with Uber fx

Every layer has a `module.go` with `fx.Module`. Bind implementations to interfaces:

```go
fx.Annotate(postgres.NewUserRepo, fx.As(new(repository.UserRepository)))
```

### RBAC

Four roles: client, owner, representative, admin.

Middleware: `RequireAuth`, `OptionalAuth`, `RequireRole(roles...)`, `RequireOwnerOrRepresentative()`
Service: `AccessChecker.CanManageBathhouse()`, `AccessChecker.CanViewBathhouseBookings()`

```go
middleware.GetUserID(ctx)   // uuid.UUID
middleware.GetUserRole(ctx) // domain.UserRole
```

### API Response Format

All endpoints return `{ success, data, error: { code, message }, meta: { page, page_size, total_count, total_pages } }`.

### Error Mapping

Domain errors (`domain/errors.go`) → HTTP status codes (`handler/response.go`):
- ErrNotFound→404, ErrAlreadyExists→409, ErrInvalidInput→400
- ErrUnauthorized→401, ErrForbidden→403, ErrUserBlocked→403
- ErrSlotUnavailable→409, ErrBookingCancelLate→400, ErrBathhouseNotActive→400
- ErrBathhouseHasBookings→409, ErrReviewAlreadyResponded→409
- ErrSocialAccountAlreadyLinked→409, ErrSocialAccountNotFound→404
- ErrOAuthExchangeFailed→400, ErrInsufficientPoints→400
- ErrComplaintNotFound→404, ErrAlreadyReported→409
- ErrSelfReferral→400, ErrAlreadyReferred→409, ErrInsufficientReferralBalance→400
- ErrCertificateNotFound→404, ErrCertificateExpired→400, ErrCertificateInsufficientBalance→400
- ErrPhotoNotFound→404

### Logging

```go
logger.Info("message", "key", value)  // also Error, Debug, Warn
```

Configured via `BANI_LOGGER_LEVEL`. Logger injected via fx.

### Config

Viper with env prefix `BANI_`. Nested keys use `_`: `BANI_DATABASE_DSN`, `BANI_JWT_SECRET`, etc.

### Database

PostgreSQL with PostGIS. Migrations in `migrations/`. Geo-search uses `ST_DWithin`/`ST_Distance` with `geography` type.

### Testing

- Services: mock repos from `repository/mock/` + table-driven tests
- Handlers: httptest + mock services
- No integration tests for postgres repos (require real DB)
- Hurl tests in `tests/hurl/` for full API endpoint testing

## Critical Conventions

- Module path: `github.com/nikitaaldaev/bani`
- chi router: `chi.URLParam(r, "id")` for URL params
- UUID (google/uuid) for all entity IDs
- Prices in **kopecks** (int64), not rubles
- Day-of-week: **0=Monday, 6=Sunday** (not Go's native 0=Sunday)
- Time format: HH:MM strings with string comparison (e.g., "09:00" < "14:30")
- Wraparound times supported: TimeFrom > TimeTo means spans midnight

## Admin Panel (GoAdmin)

Built on GoAdmin framework, enabled via `--with-admin` flag on the serve command.

- `internal/admin/` — GoAdmin engine, JWT auth bridge, fx module
- `internal/admin/pages/` — custom pages: dashboard, moderation, analytics, health
- Config: `BANI_ADMIN_ENABLED` (default `false`), `BANI_ADMIN_PREFIX` (default `/admin-panel`), `BANI_ADMIN_LANGUAGE` (default `ru`), `BANI_ADMIN_THEME` (default `adminlte`)

```bash
go run ./cmd/server serve --with-admin   # start server with admin panel
```

Custom pages:
- **Dashboard**: KPI cards (users, bathhouses, bookings, revenue), status cards, recent activity feed
- **Moderation Center**: review queue with approve/reject actions, batch operations, filtering
- **Analytics**: Chart.js charts for bookings, revenue, users, top bathhouses, with date range and city filters
- **Health Monitor**: service status checks (PostgreSQL, Redis), moderation backlog, auto-refresh

Auth bridges JWT tokens from the main app to GoAdmin sessions. Menu configured in `engine.go`.

## Feature Subsystems

Each subsystem follows the same handler→service→repository pattern:

- **Notification system**: multi-channel (in-app/email/telegram), dispatcher routes by user preferences, WebSocket hub for real-time
- **OAuth**: VK, Yandex, Google social login. Config: `BANI_OAUTH_{PROVIDER}_{CLIENT_ID,CLIENT_SECRET,REDIRECT_URL}`
- **Recommendations**: collaborative filtering + user preferences scoring
- **Subscriptions**: free/premium/promoted tiers, affects feed sorting (+10 boost for premium, promoted first)
- **Dynamic pricing**: rules with priority, multipliers applied per hourly slot
- **Loyalty**: bronze/silver/gold/platinum tiers based on visit count, points system
- **Chat**: real-time via WebSocket, conversations tied to bathhouse+client pair
- **Telegram bot**: booking wizard with in-memory state, short ID cache for callback data (64-byte limit)
- **Complaints**: report reviews/bathhouses/users (spam, offensive, fake, fraud, other), admin moderation queue with resolve/dismiss, auto-hide reviews at 3+ reports
- **Referral program**: personal referral codes, bonus on first booking completion (500 rub default to both referrer and referee), referral balance usable on bookings
- **Gift certificates**: purchasable with or without auth, unique BANI-XXXX-XXXX codes, partial redemption with balance tracking, 365-day validity
- **Photo verification**: admin-verified bathhouse photos with pending/verified/rejected statuses, `is_photo_verified` badge on bathhouse cards, owner/representative upload with admin moderation queue
