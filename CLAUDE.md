# CLAUDE.md — Project Patterns

## Build & Test Commands

```bash
go build ./...          # build
go test ./... -v        # run tests
go vet ./...            # vet
make test               # shortcut for tests
make lint               # golangci-lint
```

## Architecture

Clean architecture: handler -> service -> repository

- **domain/** — models, errors, filters. No external dependencies.
- **repository/interfaces.go** — all repository interfaces in one file
- **repository/postgres/** — pgx implementations, one file per entity
- **repository/mock/** — in-memory mock implementations for testing
- **service/** — business logic, RBAC checks via AccessChecker, one file per entity
- **handler/** — HTTP handlers, one file per entity, uses service interfaces
- **middleware/** — auth (JWT), RBAC (role-based), CORS, logging
- **server/** — chi router, HTTP server with graceful shutdown
- **notification/** — delivery channels: dispatcher, email sender, WebSocket hub
- **app/** — Uber fx DI container, assembles all modules

## Key Patterns

### DI with Uber fx

Every layer has a `module.go` with `fx.Module` that provides implementations.
Use `fx.Annotate` + `fx.As` to bind implementations to interfaces:

```go
fx.Annotate(postgres.NewUserRepo, fx.As(new(repository.UserRepository)))
```

### RBAC

Four roles: client, owner, representative, admin.

Middleware layer:
- `RequireAuth` — extracts user_id and role from JWT into context
- `OptionalAuth` — extracts user_id and role from JWT if present, but allows unauthenticated requests (used for is_favorite enrichment on public endpoints)
- `RequireRole(roles...)` — checks role from context
- `RequireOwnerOrRepresentative()` — allows owner or representative roles

Service layer:
- `AccessChecker.CanManageBathhouse()` — checks if user can manage a specific bathhouse
- `AccessChecker.CanViewBathhouseBookings()` — checks if user can view bookings

### Context Helpers

```go
middleware.GetUserID(ctx)   // uuid.UUID
middleware.GetUserRole(ctx) // domain.UserRole
```

### API Response Format

All endpoints return:
```json
{
  "success": true/false,
  "data": { ... },
  "error": { "code": "...", "message": "..." },
  "meta": { "page": 1, "page_size": 20, "total_count": 100, "total_pages": 5 }
}
```

### Error Mapping

Domain errors (domain/errors.go) map to HTTP status codes in handler/response.go:
- ErrNotFound -> 404
- ErrAlreadyExists -> 409
- ErrInvalidInput -> 400
- ErrUnauthorized -> 401
- ErrForbidden -> 403
- ErrSlotUnavailable -> 409
- ErrBookingCancelLate -> 400
- ErrUserBlocked -> 403
- ErrBathhouseNotActive -> 400
- ErrBathhouseHasBookings -> 409
- ErrReviewAlreadyResponded -> 409
- ErrSocialAccountAlreadyLinked -> 409
- ErrSocialAccountNotFound -> 404

### Structured Logging

All logging uses the `internal/logger` package with structured log levels (debug/info/warn/error).
Configure via `BANI_LOGGER_LEVEL` environment variable. Logger is injected via Uber fx:

```go
logger.Info("message", "key", value)
logger.Error("error", "key", value)
logger.Debug("debug info", "key", value)
logger.Warn("warning", "key", value)
```

### Error Handling and Recovery

- `internal/middleware/recovery.go` — panic recovery with stack traces in dev mode
- Stack traces logged with request ID for tracing
- Production mode hides implementation details
- Use `IsDevEnvironment()` to check BANI_ENVIRONMENT value

### Production Configuration

- Environment field in config (dev/staging/production)
- `Validate()` method ensures required fields: DSN, JWT Secret, Redis Addr
- Production constraints: JWT Secret 32+ chars, DSN uses sslmode=require
- Config validation happens on application startup

### Health Checks

- `/health` — liveness probe (simple 200 OK)
- `/ready` — readiness probe (checks DB and Redis connectivity)
- Returns 503 Service Unavailable if dependencies down
- Structured logging for failed checks

### Timeouts

HTTP server timeouts configured in internal/server/server.go:
- ReadHeaderTimeout: 5 seconds
- ReadTimeout: 10 seconds
- WriteTimeout: 30 seconds
- IdleTimeout: 120 seconds

Context timeouts:
- Database queries: 30 seconds (via context.WithTimeout in handlers)
- Redis operations: 5 seconds (client-level timeout)
- Use `context.WithTimeout()` for queries exceeding base timeout

### Testing

Unit Tests:
- Services tested with mock repositories from `repository/mock/`
- Handlers tested with httptest + mock services
- Middleware tested with httptest
- Error handling and panic recovery tested with realistic scenarios
- No integration tests (postgres repos require real DB)

Hurl Integration Tests:
- All API endpoints tested with real HTTP requests
- Located in `tests/hurl/` directory
- Use `run_all_tests.sh` to run full suite with setup
- Database setup: `bash tests/hurl/setup.sh`
- Individual test file format: `### Test N: description` with assertions
- JWT tokens captured with `--variable` flag in run_all_tests.sh
- Response assertions: `jsonpath`, `exists`, `isString`, `isNumber` predicates
- Error cases validated with HTTP status codes and error codes
- Run with `make test-hurl`

### Config

Viper with env prefix `BANI_`. Nested keys use `_` separator:
`BANI_DATABASE_DSN`, `BANI_JWT_SECRET`, etc.

### Database

PostgreSQL with PostGIS for geo-queries. Migrations in `migrations/` folder.
Geo-search uses `ST_DWithin` and `ST_Distance` with `geography` type.

### Notification System

Event-driven notifications with multi-channel delivery:
- **Types**: booking_confirmed, booking_cancelled, new_review, review_response, promo, reminder, system
- **Channels**: in-app (DB + WebSocket), email (SMTP/SendGrid)
- **Dispatcher** (`internal/notification/dispatcher.go`): routes to channels based on user preferences
- **WebSocket Hub** (`internal/notification/hub.go`): Hub pattern for real-time delivery to connected clients
- **Integration**: BookingService and ReviewService call NotificationService.Send() on key events
- **Preferences**: per-user channel/event-type settings in NotificationPreferences model

WebSocket endpoint: `GET /api/v1/ws/notifications?token=<JWT>` with ping/pong heartbeat.

### OAuth / Social Auth

Social login via VK, Yandex ID, Google OAuth 2.0. Multiple providers per account.

- **auth/** — OAuth provider implementations (VK, Yandex, Google), each with `GetAuthURL` and `Exchange` methods
- **domain/oauth.go** — SocialAccount model, linked to User via UserID
- **repository/postgres/social_account.go** — CRUD for social_accounts table
- **service/oauth_service.go** — OAuthCallback (find/create user + JWT), LinkSocialAccount, UnlinkSocialAccount

Routes:
- `GET /api/v1/auth/oauth/{provider}` — redirect to provider auth page
- `GET /api/v1/auth/oauth/{provider}/callback` — handle callback, return JWT
- `POST /api/v1/auth/link/{provider}` — link social account (auth required)
- `DELETE /api/v1/auth/link/{provider}` — unlink social account (auth required)
- `GET /api/v1/auth/me/social-accounts` — list linked accounts (auth required)

Config: `BANI_OAUTH_VK_CLIENT_ID`, `BANI_OAUTH_VK_CLIENT_SECRET`, `BANI_OAUTH_VK_REDIRECT_URL` (same pattern for YANDEX, GOOGLE)

### Code Style

- Module path: `github.com/nikitaaldaev/bani`
- Standard Go project layout (cmd/, internal/, config/, migrations/)
- chi for routing with URL params via `chi.URLParam(r, "id")`
- UUID for entity IDs (google/uuid)
- Prices in kopecks (int64)
