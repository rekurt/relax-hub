# Codebase Structure

**Analysis Date:** 2026-04-17

## Directory Layout

```
banya/
├── cmd/                           # Executable entry points (two binaries)
│   ├── server/                    # HTTP API server (Cobra CLI)
│   │   ├── main.go               # Entry point
│   │   ├── root.go               # Root Cobra command
│   │   ├── serve.go              # `serve` subcommand: starts HTTP server
│   │   ├── migrate.go            # `migrate` subcommand: DB migrations
│   │   ├── seed.go               # `seed` subcommand: seed data
│   │   └── docs.go               # Swagger annotations (auto-generated)
│   └── bot/                       # Telegram bot binary
│       ├── main.go               # Entry point
│       └── run.go                # Bot command handler with state machine
│
├── internal/                      # Go backend (clean architecture)
│   ├── domain/                    # Core entities, value objects, errors
│   │   ├── errors.go             # 100+ sentinel error types
│   │   ├── user.go               # User entity
│   │   ├── booking.go            # Booking entity
│   │   ├── bathhouse.go          # Bathhouse listing entity
│   │   ├── review.go             # Review entity
│   │   ├── wallet.go             # User wallet with balance, holds
│   │   ├── payment.go            # Payment entity
│   │   ├── dispute.go            # Booking dispute entity
│   │   ├── ticket.go             # Support ticket entity
│   │   ├── guest_card.go         # CRM guest card entity
│   │   └── [40+ more domain models]
│   │
│   ├── repository/                # Data access layer
│   │   ├── interfaces.go         # All 50+ repository interfaces in one file
│   │   ├── postgres/             # PostgreSQL implementations via pgx
│   │   │   ├── user_repo.go
│   │   │   ├── booking_repo.go
│   │   │   ├── bathhouse_repo.go
│   │   │   ├── payment_repo.go
│   │   │   ├── wallet_repo.go
│   │   │   └── [60+ more repos]
│   │   └── mock/                 # In-memory test mocks
│   │       ├── user_mock.go
│   │       └── [all entity mocks]
│   │
│   ├── service/                   # Business logic layer
│   │   ├── access.go             # AccessChecker RBAC abstraction
│   │   ├── booking_service.go    # Booking CRUD + pricing logic
│   │   ├── bathhouse_service.go  # Bathhouse CRUD + validation
│   │   ├── payment_service.go    # Payment processing orchestration
│   │   ├── wallet_service.go     # Wallet top-up, spend, refund, holds
│   │   ├── pricing_service.go    # Dynamic pricing rules engine
│   │   ├── review_service.go     # Review CRUD + moderation
│   │   ├── notification_service.go # Multi-channel notifications
│   │   ├── analytics_service.go  # Business metrics (GMV, conversion)
│   │   ├── crm_service.go        # Guest cards, segments, broadcasts
│   │   ├── loyalty_service.go    # Loyalty points, tiers, cashback
│   │   └── [40+ more services]
│   │
│   ├── handler/                   # HTTP handlers (request/response)
│   │   ├── response.go           # APIResponse, error mapping helpers
│   │   ├── auth_handler.go       # Login, register, password reset
│   │   ├── booking_handler.go    # Booking endpoints
│   │   ├── bathhouse_handler.go  # Bathhouse CRUD + search
│   │   ├── review_handler.go     # Review CRUD + moderation
│   │   ├── wallet_handler.go     # Wallet operations
│   │   ├── payment_handler.go    # Payment processing
│   │   ├── admin_handler.go      # Admin user/listing/booking management
│   │   ├── analytics.go          # Analytics endpoints
│   │   └── [60+ more handlers]
│   │
│   ├── middleware/                # HTTP middleware
│   │   ├── auth.go               # JWT parsing, context injection
│   │   ├── rbac.go               # Role-based access control
│   │   ├── admin_permission.go   # Admin sub-role permission checks
│   │   ├── cors.go               # CORS headers
│   │   ├── logging.go            # Request/response logging
│   │   ├── recovery.go           # Panic recovery
│   │   ├── rate_limit.go         # Token bucket rate limiter
│   │   ├── admin_audit.go        # Admin action audit logging
│   │   └── prerender.go          # SEO bot detection for SSR
│   │
│   ├── server/                    # HTTP server setup
│   │   ├── server.go             # fx.Lifecycle hooks, HTTP server config
│   │   ├── router.go             # Chi router + 120+ route definitions
│   │   └── health_integration_test.go
│   │
│   ├── database/                  # Database setup
│   │   ├── postgres.go           # pgx connection pool, migrations
│   │   └── redis.go              # Redis client setup
│   │
│   ├── logger/                    # Structured logging wrapper
│   │   └── logger.go             # Zap logger factory + levels
│   │
│   ├── app/                       # DI container
│   │   └── app.go                # Uber fx App setup with all modules
│   │
│   ├── payment/                   # Payment processing
│   │   ├── provider_factory.go   # Selects YooKassa (RU) or bePaid (BY)
│   │   ├── yookassa.go           # YooKassa implementation
│   │   ├── bepaid.go             # bePaid implementation
│   │   └── models.go             # Common payment models
│   │
│   ├── notification/              # Multi-channel notifications
│   │   ├── dispatcher.go         # Routes by user preferences
│   │   ├── email_sender.go       # SMTP email delivery
│   │   ├── telegram_sender.go    # Telegram bot API
│   │   ├── websocket_hub.go      # Real-time WebSocket server
│   │   └── models.go             # Notification types
│   │
│   ├── sms/                       # SMS provider
│   │   └── sms_ru.go             # SMS.ru adapter
│   │
│   ├── antifraud/                 # Fraud detection
│   │   ├── engine.go             # Rule evaluation
│   │   ├── rules.go              # Fraud rules (client, owner, listing)
│   │   └── models.go             # Fraud flag models
│   │
│   ├── cron/                      # Background job scheduler
│   │   ├── scheduler.go          # Robfig/cron + distributed locking
│   │   ├── booking_jobs.go       # Booking timeouts, check-in reminders
│   │   ├── wallet_jobs.go        # Bonus expiry, auto-payout
│   │   ├── review_jobs.go        # Review request, blind reveal
│   │   ├── moderation_jobs.go    # Escrow release, KYC expiry checks
│   │   └── [10+ more job files]
│   │
│   ├── geo/                       # Geolocation services
│   │   ├── isochrone.go          # Travel time zones via OpenRouteService
│   │   ├── transport.go          # Transit/parking discovery
│   │   └── helpers.go            # Haversine distance, GeoJSON
│   │
│   ├── fiscal/                    # Tax/receipt generation
│   │   ├── provider.go           # FiscalProvider interface
│   │   └── atol.go               # ATOL Online adapter
│   │
│   ├── pms/                       # Property management system integrations
│   │   ├── provider.go           # PMSProvider interface
│   │   ├── yclients.go           # Yclients adapter
│   │   └── restoplace.go         # Restoplace adapter
│   │
│   ├── storage/                   # Redis storage layer
│   │   └── redis.go              # Key-value helpers (search suggestions, recommendations)
│   │
│   ├── seo/                       # Search engine optimization
│   │   └── renderer.go           # Pre-rendering for bots (Open Graph, Schema.org)
│   │
│   ├── admin/                     # Admin panel (GoAdmin framework)
│   │   ├── engine.go             # GoAdmin setup, JWT bridge
│   │   ├── pages/                # Custom HTML pages
│   │   │   ├── dashboard.go      # KPI cards, revenue, activity
│   │   │   ├── moderation.go     # Bathhouse/review approval queue
│   │   │   ├── analytics.go      # Business metrics, charts
│   │   │   ├── health.go         # System metrics, DB stats
│   │   │   ├── finance.go        # Float monitoring, reconciliation
│   │   │   └── templates/        # Go html/template files
│   │   │       ├── base.tmpl     # Shared layout
│   │   │       ├── components.tmpl  # Reusable blocks
│   │   │       ├── dashboard.tmpl
│   │   │       └── [more templates]
│   │   └── static/               # Self-hosted assets (Chart.js, CSS)
│   │
│   ├── auth/                      # Authentication logic
│   │   ├── jwt.go                # JWT token generation/parsing
│   │   ├── otp.go                # OTP code generation/validation
│   │   └── session.go            # Session creation/validation
│   │
│   ├── moderation/                # Content moderation
│   │   ├── profanity.go          # Regex-based spam/profanity filter
│   │   └── nlp.go                # Sentiment/toxicity scoring
│   │
│   ├── calendar/                  # External calendar sync
│   │   └── sync_service.go       # iCal/Google Calendar bidirectional sync
│   │
│   └── config/                    # Configuration management
│       └── config.go             # Viper config loader (BANI_* env vars)
│
├── frontend/                      # React TypeScript SPA
│   ├── src/
│   │   ├── main.tsx              # Entry point with providers
│   │   ├── router.tsx            # React Router v7 role-based routes
│   │   ├── App.tsx               # Root component
│   │   ├── api/
│   │   │   ├── axios-instance.ts # Axios with JWT interceptor
│   │   │   └── generated/        # Auto-generated by orval (DO NOT EDIT)
│   │   │       ├── model/        # Shared TypeScript types
│   │   │       ├── auth/
│   │   │       ├── bookings/
│   │   │       ├── listings/
│   │   │       ├── wallet/
│   │   │       ├── admin/        # 20+ admin endpoints
│   │   │       └── [50+ API modules]
│   │   ├── components/           # Reusable React components
│   │   │   ├── AppLayout.tsx     # Owner/rep sidebar layout
│   │   │   ├── ClientLayout.tsx  # Client nav layout
│   │   │   ├── AdminLayout.tsx   # Admin dashboard layout
│   │   │   ├── ProtectedRoute.tsx # Role-based route guard
│   │   │   ├── BathhouseCard.tsx # Listing card component
│   │   │   ├── MediaUploader.tsx # File upload with validation
│   │   │   ├── OAuthButtons.tsx  # VK/Yandex/Google login
│   │   │   └── [50+ more components]
│   │   ├── pages/                # Page components (role-based)
│   │   │   ├── Login.tsx         # Login + OAuth
│   │   │   ├── Register.tsx      # Registration + age verification
│   │   │   ├── Dashboard.tsx     # Owner KPI dashboard
│   │   │   ├── OAuthCallback.tsx # OAuth redirect handler
│   │   │   ├── bathhouses/       # Owner listing management
│   │   │   │   ├── BathhouseList.tsx
│   │   │   │   ├── BathhouseForm.tsx # 7-step wizard
│   │   │   │   ├── ListingImport.tsx
│   │   │   │   └── AuditLog.tsx
│   │   │   ├── bookings/         # Owner booking management
│   │   │   │   ├── BookingList.tsx
│   │   │   │   └── BookingDetails.tsx
│   │   │   ├── calendar/         # Owner weekly calendar
│   │   │   │   └── CalendarPage.tsx
│   │   │   ├── pricing/          # Dynamic pricing rules
│   │   │   │   └── PricingRules.tsx
│   │   │   ├── client/           # Client pages
│   │   │   │   ├── ClientHome.tsx
│   │   │   │   ├── BathhouseSearch.tsx # Filters, map, list view
│   │   │   │   ├── BathhouseDetail.tsx # Gallery, reviews, booking
│   │   │   │   ├── BookingCreate.tsx   # 4-step wizard
│   │   │   │   ├── BookingList.tsx
│   │   │   │   ├── Favorites.tsx
│   │   │   │   ├── ReviewForm.tsx
│   │   │   │   ├── WalletDashboard.tsx
│   │   │   │   └── [30+ more client pages]
│   │   │   ├── admin/            # Admin pages
│   │   │   │   ├── AdminDashboard.tsx # KPI, top listings, support metrics
│   │   │   │   ├── UserManagement.tsx
│   │   │   │   ├── BathhouseModeration.tsx # Batch approve/reject
│   │   │   │   ├── ReviewModeration.tsx
│   │   │   │   ├── DisputeManagement.tsx
│   │   │   │   ├── TicketManagement.tsx
│   │   │   │   ├── AntiFraudDashboard.tsx
│   │   │   │   ├── AdminFinanceDashboard.tsx # Float, GMV, reconciliation
│   │   │   │   ├── GeoHeatmap.tsx # Geographic demand/supply
│   │   │   │   └── [30+ more admin pages]
│   │   │   └── [20+ other pages]
│   │   ├── stores/               # Zustand state management
│   │   │   ├── auth.ts           # User token, role, login state
│   │   │   └── bathhouse.ts      # Selected bathhouse for owner
│   │   ├── lib/                  # Utility functions
│   │   │   ├── format.ts         # Price formatting, date formatting
│   │   │   ├── constants.ts      # App constants (roles, booking statuses)
│   │   │   ├── useWebSocketNotifications.ts # Real-time notifications hook
│   │   │   └── useDeviceToken.ts # Push notification device token
│   │   ├── __tests__/            # Vitest unit/component tests
│   │   │   ├── Dashboard.test.tsx
│   │   │   ├── BookingCreate.test.tsx
│   │   │   └── [test files]
│   │   ├── vite-env.d.ts         # Vite type definitions
│   │   └── test-setup.ts         # Vitest configuration
│   ├── index.html                # SPA entry point
│   ├── vite.config.ts            # Vite + plugin config, /api proxy
│   ├── tsconfig.json             # TypeScript strict mode config
│   ├── package.json              # Dependencies (React Query, TanStack, RelaxHUB design system)
│   └── vitest.config.ts          # Test runner configuration
│
├── widget/                        # Embedded booking widget (separate SPA)
│   ├── src/
│   │   ├── main.tsx              # Widget entry point
│   │   └── WidgetApp.tsx         # Self-contained booking form
│   ├── vite.config.ts
│   └── package.json
│
├── migrations/                    # SQL schema migrations
│   ├── 000001_init.up.sql        # Initial schema (users, bathouses, bookings)
│   ├── 000002_reviews_enhancement.up.sql
│   ├── 000003_favorites.up.sql
│   ├── [100+ migration files with up/down pairs]
│   └── 000150_admin_roles.up.sql # Latest schema version
│
├── tests/                         # Integration tests
│   ├── hurl/                      # HTTP request/response tests (hurl format)
│   │   ├── auth.hurl
│   │   ├── bookings.hurl
│   │   ├── payments.hurl
│   │   └── [integration tests]
│   └── fixtures/                  # Test data
│
├── docs/                          # Auto-generated OpenAPI spec
│   ├── docs.go                    # Generated swagger comments
│   ├── swagger.json               # OpenAPI v3 spec (committed)
│   └── swagger.yaml
│
├── go.mod                         # Go module definition
├── go.sum                         # Go dependencies
├── Makefile                       # Build, test, dev commands
├── docker-compose.yml             # PostgreSQL, Redis, app services
├── CLAUDE.md                      # Instructions for Claude AI
└── config.yaml                    # Local development config template
```

## Directory Purposes

**`cmd/`:**
- Purpose: Executable entry points (two separate binaries)
- Contains: Cobra CLI command definitions, flag parsing
- Key files:
  - `cmd/server/` — HTTP API server with migrate/seed/serve subcommands
  - `cmd/bot/` — Telegram bot with state machine

**`internal/domain/`:**
- Purpose: Core business entities, value objects, error types
- Contains: 50+ domain models (no dependencies on other layers)
- Key files: `errors.go` (100+ sentinel errors), `user.go`, `booking.go`, `bathhouse.go`, `payment.go`, `wallet.go`, `dispute.go`
- No external imports — only Go stdlib

**`internal/repository/`:**
- Purpose: Data persistence abstraction
- Contains: Interface definitions + PostgreSQL + test mocks
- Key files:
  - `interfaces.go` — All 50+ repository interfaces in one file for visibility
  - `postgres/` — One file per entity (pgx implementations)
  - `mock/` — In-memory implementations for unit testing

**`internal/service/`:**
- Purpose: Business logic, validation, RBAC, orchestration
- Contains: One service per domain aggregate
- Key files: `access.go` (RBAC checker), `booking_service.go`, `payment_service.go`, `pricing_service.go`
- Pattern: Constructor injection; input/output structs for each operation

**`internal/handler/`:**
- Purpose: HTTP request/response handling
- Contains: One file per resource (e.g., `booking_handler.go` handles all `/bookings/*` endpoints)
- Key files: `response.go` (error mapping, APIResponse builder), 60+ handler files
- Pattern: Methods named `Create()`, `GetByID()`, `List()`, etc.

**`internal/middleware/`:**
- Purpose: HTTP middleware (auth, RBAC, logging, CORS, rate limiting)
- Contains: 9 middleware files
- Key files: `auth.go` (JWT), `rbac.go` (role checks), `rate_limit.go`, `logging.go`
- Used by: `internal/server/router.go` in middleware stack

**`internal/server/`:**
- Purpose: HTTP server setup and routing
- Contains: Chi router with 120+ route definitions
- Key files:
  - `server.go` — fx.Lifecycle hooks, HTTP server config (timeouts, listeners)
  - `router.go` — Chi v5 router + complete route tree

**`internal/database/`:**
- Purpose: Database connection setup, migrations
- Contains: pgx pool, Redis client, flyway/sql-migrate integration
- Key files: `postgres.go`, `redis.go`

**`internal/payment/`:**
- Purpose: Payment processing abstraction
- Contains: Provider factory, YooKasha impl, bePaid impl
- Key files:
  - `provider_factory.go` — Selects provider by user region
  - `yookassa.go`, `bepaid.go` — Payment implementations
  - Integration: `BookingService.Create()` calls `PaymentProvider.CreateHold()`

**`internal/notification/`:**
- Purpose: Multi-channel notifications
- Contains: Dispatcher, email/SMS/Telegram senders, WebSocket hub
- Key files:
  - `dispatcher.go` — Routes by user preferences + fallback chain
  - `websocket_hub.go` — Real-time in-app notifications

**`internal/cron/`:**
- Purpose: Background job scheduling
- Contains: Robfig/cron + distributed Redis locking
- Key files:
  - `scheduler.go` — Job registration + lock management
  - `booking_jobs.go`, `wallet_jobs.go`, `review_jobs.go`, etc.
  - 30+ registered jobs: every 15min, hourly, daily (midnight Moscow)

**`internal/admin/`:**
- Purpose: Admin panel (GoAdmin framework)
- Contains: Custom pages, templates, static assets
- Key files:
  - `engine.go` — GoAdmin setup + JWT auth bridge
  - `pages/dashboard.go`, `pages/moderation.go`, `pages/analytics.go`
  - `pages/templates/` — Go html/template (base.tmpl, components.tmpl, page templates)

**`frontend/src/`:**
- Purpose: React SPA (owner, client, admin roles)
- Contains: Components, pages, stores, API client
- Key files:
  - `router.tsx` — React Router v7 with role-based route groups
  - `api/generated/` — Orval-generated API client (DO NOT EDIT)
  - `pages/` — Organized by role: owner (/), client (/client/*), admin (/admin/*)
  - `stores/` — Zustand stores (auth token, selected bathhouse)
  - `components/` — Reusable RelaxHUB design-system components and app-level wrappers

**`migrations/`:**
- Purpose: Database schema version control
- Contains: SQL migration pairs (up/down) in flyway format
- Key files: 150+ migrations (000001_init.up.sql through 000150_admin_roles.up.sql)
- Pattern: Sequential numbers, descriptive names, reversible

**`tests/hurl/`:**
- Purpose: HTTP integration tests
- Contains: .hurl files (request/response assertions)
- Key files: `auth.hurl`, `bookings.hurl`, `payments.hurl`

**`docs/`:**
- Purpose: OpenAPI specification (auto-generated)
- Contains: Swagger annotations in Go code → swag generates this
- Key files: `swagger.json`, `swagger.yaml` (committed to git)
- Generation: `make swagger` regenerates from handler annotations

## Key File Locations

**Entry Points:**
- `cmd/server/main.go` — HTTP API server entry point
- `cmd/bot/main.go` — Telegram bot entry point
- `frontend/src/main.tsx` — React SPA entry point

**Configuration:**
- `config.yaml` — Local dev config template
- `internal/config/config.go` — Viper loader (env prefix `BANI_`)
- `.env` — Environment variables (not committed, contains secrets)

**Core Logic:**
- `internal/service/booking_service.go` — Booking creation + pricing
- `internal/service/payment_service.go` — Payment orchestration
- `internal/service/wallet_service.go` — Wallet operations
- `internal/service/pricing_service.go` — Dynamic pricing rules
- `internal/handler/booking_handler.go` — Booking HTTP endpoints

**Database Schema:**
- `migrations/000001_init.up.sql` — Core tables (users, bathouses, bookings, reviews)
- `migrations/000002_reviews_enhancement.up.sql` — Review criteria, moderation
- `migrations/000010_promo_codes.up.sql` — Promo code system
- `migrations/000050_wallets.up.sql` — Wallet + holds + bonuses
- `migrations/000100_disputes.up.sql` — Dispute system
- `migrations/000150_admin_roles.up.sql` — Admin sub-roles + permissions

**Testing:**
- `internal/repository/mock/` — Mock implementations for unit tests
- `internal/service/*_test.go` — Table-driven service tests
- `frontend/src/__tests__/` — Vitest component tests
- `tests/hurl/` — HTTP integration tests

**API Documentation:**
- `cmd/server/docs.go` — Swagger annotation comments
- `docs/swagger.json` — OpenAPI v3 spec (regenerated via `make swagger`)
- Frontend: `/swagger/` endpoint serves Swagger UI (dev only)

## Naming Conventions

**Files:**
- `*_handler.go` — HTTP handlers (e.g., `booking_handler.go`)
- `*_service.go` — Business logic services (e.g., `payment_service.go`)
- `*_repo.go` — Repository implementations (e.g., `user_repo.go`)
- `*_test.go` — Unit tests (e.g., `booking_service_test.go`)
- `.test.tsx` — React component tests (e.g., `Dashboard.test.tsx`)

**Directories:**
- Lowercase with underscores: `internal/admin/pages/`, `frontend/src/api/generated/`
- Short names for packages: `auth`, `cron`, `geo`, `pms`, `seo`

**Functions/Methods:**
- camelCase: `CreateBooking()`, `GetByID()`, `ValidateBooking()`
- Prefixes for interfaces: `Is*()`, `Can*()`, `NewX()` (constructors)
- Services: `CreateX()`, `UpdateX()`, `DeleteX()`, `GetX()` (CRUD pattern)

**Types (Go):**
- PascalCase: `Booking`, `BathhouseService`, `PaymentProvider`
- Interfaces end in pattern: `Repository`, `Service`, `Provider`, `Handler`
- Request/response types: `CreateBookingInput`, `BookingResult`, `PaginatedResult[T]`

**Variables (Go):**
- camelCase: `userID`, `bathhouseID`, `bookingService`
- Context keys: const `contextKey` string type
- Error variables: `ErrNotFound`, `ErrAlreadyExists` (domain/errors.go)

**Routes (HTTP):**
- RESTful: `/api/v1/bookings`, `/api/v1/bookings/{id}`, `/api/v1/my/bookings`
- Prefix `/my/` for authenticated user's own resources
- Prefix `/admin/` for admin-only endpoints
- Versioning: `/api/v1/` (single version)

**React Components:**
- PascalCase: `BathhouseCard`, `BookingCreate`, `AdminDashboard`
- Hooks start with `use`: `useBookings()`, `useWebSocketNotifications()`
- Pages in `pages/` directory with path matching: `pages/client/BookingCreate.tsx` → route `/client/bookings/create`

## Where to Add New Code

**New Feature (e.g., "Add gift certificates"):**
- Domain model: `internal/domain/certificate.go` (Cert entity, errors)
- Repository interface: Add `CertificateRepository` to `internal/repository/interfaces.go`
- Repository implementation: `internal/repository/postgres/certificate_repo.go`
- Service: `internal/service/certificate_service.go` with CRUD methods + pricing logic
- Handler: `internal/handler/certificate_handler.go` with HTTP endpoints
- Routes: Add to `internal/server/router.go` under `/api/v1/certificates` group
- Frontend page: `frontend/src/pages/client/CertificatePurchase.tsx`
- Tests: `internal/service/certificate_service_test.go` with table-driven tests + mocks
- Migrations: `migrations/XXXXXX_certificates.up.sql` + reverse

**New Component/Module:**
- Implementation: `internal/[module_name]/`
- Create `module.go` with `fx.Module` definition (export all public constructors)
- Add to `app.New()` DI container in `internal/app/app.go`
- Example: `internal/geo/` has isochrone + transport services

**Utilities/Helpers:**
- Shared across domain: `internal/lib/[name].go` (if needed; most go in domain/service)
- Format/validation: `lib/` in frontend (`frontend/src/lib/format.ts`)
- Re-export via package `init()` or barrel file if used widely

**Database Migrations:**
- Create: `migrations/XXXXXX_description.up.sql` + `XXXXXX_description.down.sql`
- Number: Increment from latest (e.g., 000151_add_xyz.up.sql)
- Pattern: Use PostGIS (bathhouses have lat/lng), indexes on foreign keys, constraints
- Add-on pattern: `ALTER TABLE` then create new table if major schema change
- Run: `make migrate-up` to apply, committed to git

**Tests:**
- Unit: Mock repos from `repository/mock/`, table-driven assertions
- Integration: Use hurl in `tests/hurl/` (requires running server)
- Component: Vitest + React Testing Library in `frontend/src/__tests__/`
- No in-process postgres tests (would require docker or test container)

## Special Directories

**`frontend/src/api/generated/`:**
- Purpose: Auto-generated API client via orval (from OpenAPI spec)
- Generated: Yes (do NOT edit manually)
- Committed: No (built during CI/CD)
- Regenerate: `make frontend-generate-api` after backend swagger changes
- Structure: One folder per API tag (auth, bookings, admin, etc.)

**`docs/`:**
- Purpose: Auto-generated OpenAPI/Swagger specification
- Generated: Yes (via `make swagger` from Go annotations)
- Committed: Yes (needed for CI/CD and frontend generation)
- Update: After modifying handler annotations, run `make swagger`

**`migrations/`:**
- Purpose: Database schema version control
- Generated: No
- Committed: Yes (mandatory)
- Pattern: Flyway format (VVVV__Description.sql), sequential numbers

**`widget/`:**
- Purpose: Separate embedded booking widget (iframe-able)
- Generated: No (source code)
- Committed: Yes
- Build: Separate SPA, compiled to `widget/dist/`
- Entry: `widget/src/WidgetApp.tsx`

**`tests/hurl/`:**
- Purpose: Integration tests (HTTP request/response)
- Generated: No
- Committed: Yes
- Run: `make test-hurl` (requires running server on localhost:8080)

**`.github/workflows/`:**
- Purpose: CI/CD pipeline (GitHub Actions)
- Generated: No
- Committed: Yes
- Runs: golangci-lint, go test, go build, frontend build on push/PR to main

---

*Structure analysis: 2026-04-17*
