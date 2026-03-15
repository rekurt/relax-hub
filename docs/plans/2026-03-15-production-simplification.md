# Production Simplification: Architecture Cleanup + Production Readiness

## Overview

Comprehensive simplification and hardening of the banya project: removing unnecessary abstractions, fixing architectural leaks (handlers bypassing services), replacing the custom logger with structured logging, securing CORS/rate limiting, enabling disabled linters, and adding CI/CD.

## Context

- Files involved: ~40 files across internal/service, internal/handler, internal/middleware, internal/repository, internal/logger, internal/app, internal/database, config/, .golangci.yml, .github/workflows
- Related patterns: handler->service->repository clean architecture, fx DI, chi router
- Dependencies: go.uber.org/zap (already in go.mod as indirect dep)

## Development Approach

- **Testing approach**: Regular (code first, then tests)
- Complete each task fully before moving to the next
- **CRITICAL: every task MUST include new/updated tests**
- **CRITICAL: all tests must pass before starting next task**

## Implementation Steps

### Task 1: Replace custom logger with zap (structured JSON)

**Files:**
- Rewrite: `internal/logger/logger.go`
- Modify: `internal/logger/module.go`
- Modify: `config/config.go` (add log format config)

The current custom logger uses `fmt.Sprintf("[%s] %s", level, msg)` with `log.New()` on every call. Fields are dumped as raw `%v` slices. zap is already in go.mod as indirect dep.

- [x] Replace custom Logger implementation with a thin wrapper around `zap.SugaredLogger`
- [x] Keep existing API signature (`Info(msg, key, val...)`, `Error(msg, key, val...)`, etc.) to minimize caller changes
- [x] Add `BANI_LOGGER_FORMAT` config: `json` (default/production) or `console` (development)
- [x] Remove per-call `log.New()` allocation
- [x] Update module.go fx provider
- [x] Run `go test ./...` to verify no breakage

### Task 2: Fix CORS to be configurable

**Files:**
- Modify: `internal/middleware/cors.go`
- Modify: `config/config.go` (add CORS config)

CORS is hardcoded as `AllowedOrigins: ["*"]`. Must be configurable per environment.

- [x] Add `BANI_CORS_ALLOWED_ORIGINS` config (comma-separated list, default `["*"]` for dev)
- [x] Pass allowed origins from config into CORS middleware constructor
- [x] Add production validation: reject `["*"]` when `BANI_ENVIRONMENT=production`
- [x] Test with config validation tests

### Task 3: Add rate limiting to auth and public endpoints

**Files:**
- Modify: `internal/middleware/rate_limit.go`
- Modify: `internal/server/router.go`

Auth/register, auth/login, webhook, promo-codes/validate, certificates/purchase have no rate limiting. The existing rate limiter is in-memory only.

- [x] Add `Retry-After` and `X-RateLimit-Remaining` headers to rate limit responses
- [x] Fix `r.RemoteAddr` port issue: use IP-only key (strip port)
- [x] Apply rate limiting middleware to auth group (register: 5/min, login: 10/min)
- [x] Apply rate limiting to webhooks (30/min) and promo validation (20/min)
- [x] Add tests for rate limit middleware

### Task 4: Enable disabled linters and fix violations

**Files:**
- Modify: `.golangci.yml`
- Modify: various files with lint violations

govet, staticcheck, gosimple, and typecheck are all disabled. These are critical correctness linters.

- [x] Enable `govet` and fix all violations
- [x] Enable `staticcheck` and fix all violations
- [x] Enable `gosimple` and fix all violations
- [x] Enable `typecheck` and fix all violations
- [x] Run `make lint` clean

### Task 5: Remove unnecessary abstractions

**Files:**
- Modify: `internal/service/access.go`
- Modify: `internal/handler/ws.go`
- Modify: `internal/service/chat_service.go`
- Modify: `internal/service/chat_service_test.go`
- Modify: `internal/app/app.go`
- Modify: `internal/notification/module.go`
- Modify: `internal/notification/email.go`
- Modify: `internal/notification/push.go`
- Modify: `internal/storage/noop.go`
- Modify: `internal/bot/module.go`
- Modify: `internal/service/city_service.go`
- Modify: `internal/service/favorite_service.go`

The codebase has accumulated multiple layers of abstraction that add indirection without providing real value. This task removes wrapper interfaces, alias methods, unnecessary adapter functions, and simplifies the DI wiring.

#### 5.1: Remove AccessChecker alias method `CanViewBathhouseBookings`

`CanViewBathhouseBookings()` in `access.go` is a 1-line pass-through to `CanManageBathhouse()`. Zero additional logic.

- [x] Find all callers of `CanViewBathhouseBookings` across the codebase
- [x] Replace every call with `CanManageBathhouse`
- [x] Delete `CanViewBathhouseBookings` method from AccessChecker
- [x] Remove from AccessChecker interface definition if present
- [x] Run tests to verify

#### 5.2: Remove `ConversationAccessChecker` wrapper interface

`ConversationAccessChecker` in `handler/ws.go` is a single-method interface that wraps `ChatService.CanAccessConversation()`. An adapter in `app.go` (`func(svc service.ChatService) handler.ConversationAccessChecker { return svc }`) exists solely to bridge this.

- [x] Change `WSHandler` to depend on `service.ChatService` directly instead of `ConversationAccessChecker`
- [x] Update WSHandler constructor and all call sites from `checker.CanAccessConversation(...)` to `chatService.CanAccessConversation(...)`
- [x] Delete the `ConversationAccessChecker` interface definition from `handler/ws.go`
- [x] Remove the adapter function from `internal/app/app.go`
- [x] Update DI module to inject ChatService into WSHandler
- [x] Run tests to verify

#### 5.3: Remove `ChatBroadcaster` wrapper interface

`ChatBroadcaster` in `service/chat_service.go` wraps 2 methods from `notification.Hub` (`BroadcastNewMessage`, `BroadcastMessageRead`). An adapter in `app.go` (`func(hub *notification.Hub) service.ChatBroadcaster { return hub }`) exists to bridge this.

- [x] Change `ChatService` to depend on `*notification.Hub` directly instead of `ChatBroadcaster`
- [x] Update ChatService constructor signature and field type
- [x] Delete the `ChatBroadcaster` interface definition from `chat_service.go`
- [x] Remove the adapter function from `internal/app/app.go`
- [x] Remove `noopChatBroadcaster` test struct from `chat_service_test.go` and replace with a nil Hub or a test Hub instance
- [x] Update DI module
- [x] Run tests to verify

#### 5.4: Remove noop implementations for optional features

4 noop structs exist: `NoopTelegramSender`, `NoopEmailSender`, `NoopPushSender`, `NoopStorage`. These are used to satisfy DI when features are disabled but add boilerplate. Replace with nil-safe patterns.

- [x] Add nil-receiver guards to TelegramSender, EmailSender, PushSender, Storage method calls (e.g., `if s == nil { return nil }`)
- [x] Remove `NoopTelegramSender` from `notification/module.go`
- [x] Remove `NoopEmailSender` from `notification/email.go`
- [x] Remove `NoopPushSender` from `notification/push.go`
- [x] Remove `NoopStorage` from `storage/noop.go`
- [x] Update DI modules to provide `nil` when feature is disabled instead of noop struct
- [x] Run tests to verify

#### 5.5: Simplify `ProvideBotConfig` factory function

`ProvideBotConfig()` in `bot/module.go` is a 1-line field extraction: `return &cfg.Telegram`. This adds unnecessary indirection.

- [x] Inline config access where bot config is needed, or use `fx.Supply` with direct field extraction
- [x] Remove `ProvideBotConfig` function
- [x] Update bot module DI wiring
- [x] Run tests to verify

#### 5.6: Remove pure pass-through service methods

Several service methods add zero business logic and just forward to repository:
- `FavoriteService.IsFavorite()` -> direct repo call
- `CityService.GetAll()`, `CityService.GetBySlug()`, `CityService.Delete()` -> direct repo calls
- `SubscriptionService.GetActive()`, `SubscriptionService.ListByOwner()` -> direct repo calls

- [x] Audit each pass-through method to confirm it has no business logic, no access checks, no logging
- [x] For confirmed pass-throughs: remove the service method and have the handler call the service's other methods or restructure so the repo call is part of a meaningful service method
- [x] Note: do NOT remove service methods that perform access checks or validation, even if they look simple
- [x] Run tests to verify
- [x] Final: run full test suite after all sub-tasks to confirm nothing is broken

### Task 6: Fix handler-to-repo bypasses

**Files:**
- Modify: `internal/handler/bathhouse_handler.go` (remove direct PromotionRepository, CityRepository)
- Modify: `internal/service/bathhouse_service.go` or create promotion-related service methods
- Modify: `internal/handler/subscription.go`, `pricing.go`, `recommendation_handler.go`, `widget.go`, `admin_handler.go`
- Create: `internal/handler/device_token_handler.go` service or inline into existing service

6 handlers bypass the service layer with direct repo injections. This violates the handler->service->repository pattern.

- [x] Move `PromotionRepository.RecordImpression/RecordClick/GetActiveByBathhouse` calls behind BathhouseService or a new PromotionService
- [x] Move `CityRepository` usage in BathhouseHandler.buildMeta() behind CityService (it already exists)
- [x] Move direct `BathhouseRepository` in SubscriptionHandler, PricingHandler, RecommendationHandler, WidgetHandler behind BathhouseService
- [x] Move direct `ReviewRepository` in AdminHandler behind ReviewService
- [x] Create minimal DeviceTokenService (or add methods to UserService) for DeviceTokenHandler
- [x] Update handler constructors, DI module, and router params
- [x] Update existing tests for modified handlers

### Task 7: Simplify mock pagination boilerplate

**Files:**
- Create: `internal/repository/mock/helpers.go`
- Modify: all mock files with repeated pagination logic

15+ mock methods repeat the same 6-line pagination slice pattern.

- [ ] Create generic `paginate[T any](items []T, page, pageSize int) *domain.PaginatedResult[T]` helper
- [ ] Replace all manual pagination logic in mock files with the helper
- [ ] Run tests to verify mocks still work correctly

### Task 8: Tune database connection pool

**Files:**
- Modify: `internal/database/postgres.go`
- Modify: `config/config.go` (add pool config)

Pool uses pgx defaults with no tuning. No MaxConns, MaxConnLifetime, MinConns configured.

- [ ] Add pool config: `BANI_DATABASE_MAX_CONNS` (default 20), `BANI_DATABASE_MAX_CONN_LIFETIME` (default 1h), `BANI_DATABASE_MIN_CONNS` (default 2)
- [ ] Apply pool config to pgxpool before creating the pool
- [ ] Add pool config to config validation (warn if MaxConns > 100)
- [ ] Test pool creation with config

### Task 9: Fix silent panics and config inconsistencies

**Files:**
- Modify: `internal/notification/hub.go` (log recovered panics)
- Modify: `internal/middleware/recovery.go` (use config instead of env var)
- Modify: `config/config.go` (validate payment/telegram config)

- [ ] In hub.go: replace `_ = recover()` with logged recovery using the logger
- [ ] In recovery.go: replace `os.Getenv("BANI_ENVIRONMENT")` with config-based check
- [ ] Add startup validation: if `BANI_PAYMENT_YOOKASSA_SHOP_ID` is empty, log a warning (not error, since payments may be optional)
- [ ] Test recovery middleware behavior

### Task 10: Add CI/CD pipeline (GitHub Actions)

**Files:**
- Create: `.github/workflows/ci.yml`

No automated CI/CD exists.

- [ ] Create workflow: on push/PR to main
- [ ] Steps: checkout, setup-go, go mod download, make lint, go vet ./..., go test ./... -race -coverprofile=coverage.out
- [ ] Add coverage threshold check (80%+)
- [ ] Add go build ./... step
- [ ] Test workflow by pushing to a branch

### Task 11: Clean up committed secrets

**Files:**
- Modify: `config/config.yaml` (remove or replace with env-var references)
- Modify: `docker-compose.yml` (use env_file instead of hardcoded secrets)
- Modify: `.gitignore` (add config/config.local.yaml)

config.yaml is committed with minioadmin credentials and change-me-in-production JWT. docker-compose.yml hardcodes a JWT secret that bypasses the validator.

- [ ] Replace `config/config.yaml` with `config/config.yaml.example` containing only structure and env var references
- [ ] Add `config/config.yaml` to .gitignore
- [ ] In docker-compose.yml, replace hardcoded `BANI_JWT_SECRET` with `${BANI_JWT_SECRET}` and add `env_file: .env`
- [ ] Verify app still starts with example config + env overrides

### Task 12: Verify acceptance criteria

- [ ] manual test: start server, verify structured JSON logs appear
- [ ] manual test: verify CORS rejects unauthorized origins when configured
- [ ] manual test: verify rate limiting returns 429 on auth endpoints
- [ ] run full test suite: `go test ./... -race`
- [ ] run linter: `make lint` (with govet, staticcheck enabled)
- [ ] verify no direct repo imports in handler constructors (except via services)

### Task 13: Update documentation

- [ ] update CLAUDE.md: add logger format config, CORS config, pool config, CI/CD info
- [ ] update CLAUDE.md: remove references to disabled linters
- [ ] move this plan to `docs/plans/completed/`
