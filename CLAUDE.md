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
- ErrOAuthExchangeFailed -> 400
- ErrInsufficientPoints -> 400

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

### Recommendation Engine

Personalized bathhouse recommendations using collaborative filtering and user preferences:

- **Models**: UserPreferences (explicit preferences: city, price range, amenities), UserActivity (view/booking/favorite tracking)
- **Repository** (`internal/repository/postgres/recommendation.go`): GetUserPreferences, SaveUserPreferences, RecordActivity, GetUserBookedBathhouses, GetSimilarUsers, GetPopularBathhouses, GetSimilarBathhouses
- **Service** (`internal/service/recommendation_service.go`):
  - GetPersonalized(userID, page, pageSize) — algorithm: get user preferences → find similar users (collaborative filtering) → get bathhouses booked by similar users but not current user → filter by preferences (city, price, amenities) → score by rating × similarity weight × recency bonus
  - GetSimilar(bathhouseID, limit) — bathhouses with similar amenities/price/city
  - GetPopular(cityID, limit) — highest-rated bathhouses in city
  - UpdatePreferences(userID, prefs) — save explicit preferences
  - RecordView(userID, bathhouseID) — track activity for recommendations

Routes:
- `GET /api/v1/recommendations?page=1&page_size=20` — personalized (auth required)
- `GET /api/v1/bathhouses/{id}/similar?limit=10` — similar bathhouses (public)
- `GET /api/v1/popular?city_id=1&limit=10` — popular in city (public)
- `GET /api/v1/my/preferences` — get preferences (auth required)
- `PUT /api/v1/my/preferences` — update preferences (auth required)

Activity tracking: RecordView called on bathhouse detail retrieval for authenticated users.

Recommendation scoring formula: rating × similarity_weight × recency_bonus

### Subscriptions and Promotions

Three-tier subscription model for bathhouse monetization:

**Models:**
- **Subscription** (`internal/domain/subscription.go`): tracks plan type (free/premium/promoted), status (active/expired/cancelled), dates, pricing, and auto-renewal
- **Promotion** (`internal/domain/promotion.go`): tracks advertising campaigns with budget, impressions, clicks, target city, and status

**Plans:**
- Free: no charge, basic listing
- Premium: 5000 kopecks/month, +10 sort score boost in feed
- Promoted: 10000 kopecks/month, dedicated "Recommended" block, impression/click tracking

**Repositories:**
- `internal/repository/postgres/subscription.go` — CRUD operations, fetch active subscriptions, list by owner with pagination, find expiring subscriptions
- `internal/repository/postgres/promotion.go` — budget/metrics tracking, impression/click counting
- Mock implementations for testing in `internal/repository/mock/`

**Service:**
- `internal/service/subscription_service.go` — Subscribe (create + charge), Cancel (disable auto-renewal), GetActive, ListByOwner
- RBAC: only bathhouse owners can manage their subscriptions
- Error: `ErrSubscriptionNotFound`, `ErrSubscriptionAlreadyActive`, `ErrPromotionBudgetExhausted`

**Feed Impact:**
- Premium subscriptions: bathhousses get +10 points in sort scoring
- Promoted subscriptions: appear first in all results regardless of sort order, includes "is_promoted" flag in response
- Impression tracking: recorded when promoted bathhouses appear in List results
- Click tracking: recorded when promoted bathhouses are viewed via GetByID

**Promotion Management:**
- Promotions can only be created for bathhouses with active Promoted subscription
- Promotions have budget tracking with impression/click counting
- Budget model: tracks budget_kopecks and spent_kopecks for campaign management

**Handlers:**
- `POST /api/v1/my/bathhouses/{id}/subscription` — subscribe to plan (owner auth required)
- `GET /api/v1/my/bathhouses/{id}/subscription` — get current subscription (owner auth)
- `DELETE /api/v1/my/bathhouses/{id}/subscription` — cancel auto-renewal (owner auth)
- `GET /api/v1/my/subscriptions?page=1&page_size=20` — list all owner subscriptions (owner auth)
- `POST /api/v1/my/bathhouses/{id}/promotion` — create promotion campaign (owner auth)
- `GET /api/v1/my/bathhouses/{id}/promotion` — get promotion analytics (owner auth)

Database migrations in `migrations/000005_subscriptions.up.sql` create subscriptions and promotions tables with indexes on bathhouse_id, owner_id, and status for efficient queries.

### Dynamic Pricing System

Flexible pricing rules allow bathhouse owners to set different prices for different times/days:

**Models:**
- **PricingRule** (`internal/domain/pricing.go`): ID, BathhouseID, Name, Type (weekday/weekend/holiday/time_range/season), Multiplier (1.5 = +50%, 0.8 = -20%), DaysOfWeek, TimeFrom/TimeTo (HH:MM format), DateFrom/DateTo, Priority (higher wins on conflict), IsActive, CreatedAt

**Repository:**
- `internal/repository/postgres/pricing.go` — Create, Update, Delete, ListByBathhouse, GetActiveRules
- Mock implementation for testing in `internal/repository/mock/`

**Service:**
- `internal/service/pricing_service.go` — CalculatePrice, CreateRule, UpdateRule, DeleteRule, ListRules
- Algorithm: Split booking interval into hourly slots, find highest-priority applicable rule for each hour, apply multiplier to base price
- Day-of-week convention: 0=Monday, 6=Sunday (matches WorkingHours)
- Handles wraparound time ranges (e.g., 22:00-06:00 overnight shifts)

**Integration:**
- BookingService.CreateBooking uses PricingService.CalculatePrice instead of fixed multiplier
- GetAvailableSlots returns calculated price per slot from pricing rules
- Database migration `migrations/000008_dynamic_pricing.up.sql` creates pricing_rules table with indexes

**Handlers:**
- `POST /api/v1/my/bathhouses/{id}/pricing-rules` — create rule (owner/rep auth required)
- `GET /api/v1/my/bathhouses/{id}/pricing-rules` — list rules (owner/rep auth)
- `PUT /api/v1/pricing-rules/{id}` — update rule (owner/rep auth)
- `DELETE /api/v1/pricing-rules/{id}` — delete rule (owner/rep auth)
- `GET /api/v1/bathhouses/{id}/price-calculator?start=...&end=...` — public price quote calculator

**Critical Details:**
- Time format validation: HH:MM (00:00-23:59)
- Day-of-week uses app convention (0=Monday), not Go's native (0=Sunday)
- Wraparound times supported: from > to means rule spans midnight
- String comparison works for HH:MM format (e.g., "09:00" < "14:30")

### Loyalty Program

Cumulative points system rewarding users for completed bookings with tiered benefits:

**Models:**
- **LoyaltyAccount** (`internal/domain/loyalty.go`): UserID, Level (bronze/silver/gold/platinum), Points, TotalEarned, TotalSpent, VisitCount
- **LoyaltyTransaction** (`internal/domain/loyalty.go`): ID, UserID, Type (earn/spend), Amount, BookingID, Description

**Tiers:**
- Bronze: 0+ visits (1x multiplier, 0% discount)
- Silver: 5+ visits (1.2x multiplier, 3% discount)
- Gold: 15+ visits (1.5x multiplier, 5% discount)
- Platinum: 30+ visits (2x multiplier, 10% discount)

**Repository** (`internal/repository/postgres/loyalty_repo.go`): GetAccount, CreateAccount, AddPoints, SpendPoints, IncrementVisitCount, UpdateLevel, ListTransactions

**Service** (`internal/service/loyalty_service.go`):
- GetAccount (auto-creates at Bronze if missing)
- EarnPoints — returns points awarded; increments visit count
- SpendPoints, GetDiscount, RecalculateLevel, ListTransactions
- Points formula: TotalPrice / 100 * level multiplier

**Integration with BookingService:**
- On completion: EarnPoints + IncrementVisitCount + RecalculateLevel (non-blocking, logged on failure)
- On creation: optional partial payment with loyalty points (use_points field), booking ID generated before spend

**Handlers:**
- `GET /api/v1/my/loyalty` — loyalty account with privileges (auth required)
- `GET /api/v1/my/loyalty/transactions` — paginated transaction history (auth required)
- `GET /api/v1/my/loyalty/levels` — all levels with thresholds (auth required)

Database migration: `migrations/000012_loyalty.up.sql`

### Code Style

- Module path: `github.com/nikitaaldaev/bani`
- Standard Go project layout (cmd/, internal/, config/, migrations/)
- chi for routing with URL params via `chi.URLParam(r, "id")`
- UUID for entity IDs (google/uuid)
- Prices in kopecks (int64)
