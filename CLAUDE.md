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

# OpenAPI / Swagger
make swagger                               # generate docs via swag init
make swagger-fmt                           # format swagger annotations

# Frontend (Multi-role SPA)
make frontend-dev                          # start Vite dev server
make frontend-build                        # production build
make frontend-generate-api                 # regenerate API client from swagger.json
cd frontend && npm run lint                # ESLint
cd frontend && npx vitest run              # run tests
```

## Architecture

Clean architecture: **handler → service → repository**

- `internal/domain/` — models, errors, filters. No external dependencies.
- `internal/repository/interfaces.go` — all repository interfaces in one file
- `internal/repository/postgres/` — pgx implementations, one file per entity
- `internal/repository/mock/` — in-memory mocks for testing
- `internal/service/` — business logic, RBAC checks via AccessChecker
- `internal/handler/` — HTTP handlers, one file per entity
- `internal/middleware/` — auth (JWT), RBAC, CORS (configurable origins), rate limiting, logging, panic recovery, admin audit
- `internal/server/` — chi router setup
- `internal/notification/` — dispatcher, email sender, WebSocket hub, telegram sender
- `internal/payment/` — payment provider abstraction (YooKassa integration)
- `internal/antifraud/` — fraud detection engine, rules (wallet/booking/payout), chat content filtering
- `internal/cron/` — centralized cron scheduler with distributed locking, all background jobs
- `internal/geo/` — geolocation services (isochrone zones, transport accessibility)
- `internal/pms/` — PMS integration adapters (Yclients, Restoplace)
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
Service: `AccessChecker.CanManageBathhouse()`

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
- ErrPromoNotFound→404, ErrPromoExpired→400, ErrPromoMaxUses→409, ErrPromoMinAmount→400, ErrPromoInvalid→400
- ErrMediaNotFound→404, ErrMediaFileTooLarge→400, ErrMediaInvalidType→400, ErrMediaLimitReached→409
- ErrPaymentNotFound→404, ErrPaymentAlreadyProcessed→409, ErrRefundExceedsAmount→400, ErrPaymentFailed→400
- ErrWalletNotFound→404, ErrInsufficientWalletBalance→400, ErrWalletLimitExceeded→400, ErrWalletFrozen→403
- ErrHoldNotFound→404, ErrHoldExpired→400, ErrTopUpBelowMinimum→400, ErrTopUpAboveMaximum→400
- ErrOTPRateLimited→429, ErrOTPInvalid→400, ErrOTPExpired→400, ErrOTPMaxAttempts→429
- ErrSessionNotFound→404, ErrSessionExpired→401
- ErrAccountDeletionPending→409, ErrAccountDeletionNotPending→400, ErrAccountDeleted→403
- ErrKYCNotFound→404, ErrKYCNotApproved→403, ErrKYCPending→409
- ErrOfferNotFound→404, ErrOfferNotAccepted→403, ErrOfferAlreadyAccepted→409
- ErrPaymentDetailsNotFound→404, ErrPaymentDetailsNotSet→403
- ErrAddOnNotFound→404, ErrAddOnLimitReached→409
- ErrSavedSearchNotFound→404, ErrSavedSearchLimitReached→409
- ErrSubscriptionNotFound→404, ErrSubscriptionAlreadyActive→409, ErrPromotionBudgetExhausted→409
- ErrWalletConcurrentUpdate→409
- ErrPayoutNotFound→404, ErrPayoutBelowMinimum→400, ErrPayoutDailyLimitExceeded→400, ErrPayoutMonthlyLimitExceeded→400, ErrPayoutAlreadyProcessed→409
- ErrPhoneRequired→400, ErrPhoneInvalid→400
- Err2FARequired→403, Err2FAAlreadyEnabled→409, Err2FANotEnabled→400, Err2FAInvalidCode→400, Err2FAPhoneRequired→400
- ErrResetTokenInvalid→400, ErrResetRateLimited→429
- ErrListingDraftNotFound→404, ErrListingDraftIncomplete→400, ErrListingDraftSubmitted→409, ErrListingDraftInvalidStep→400
- ErrListingIncomplete→400
- ErrCheckinTooEarly→400, ErrCheckinTooLate→400, ErrNotCheckedIn→400, ErrNoShowDisputeExpired→400
- ErrEscrowNotFound→404, ErrEscrowNotMatured→400, ErrEscrowAlreadyReleased→409, ErrEscrowDisputed→409
- ErrBroadcastNotFound→404, ErrBroadcastRateLimit→429, ErrBroadcastNotDraft→400
- ErrAutoScenarioNotFound→404
- ErrTemplateNotFound→404, ErrTemplateLimitReached→409
- ErrTicketNotFound→404, ErrTicketAlreadyClosed→409, ErrTicketAlreadyResolved→409, ErrTicketAlreadyEscalated→409
- ErrCSATAlreadySubmitted→409, ErrCSATNotResolved→400, ErrCSATInvalidScore→400
- ErrDisputeNotFound→404, ErrDisputeAlreadyExists→409, ErrDisputeAlreadyResolved→409
- ErrDisputeAlreadyClosed→409, ErrDisputeEvidenceWindowExpired→400
- ErrDisputeAppealExpired→400, ErrDisputeNotResolved→400, ErrDisputeAlreadyAppealed→409
- ErrBookingModificationLimit→400, ErrBookingNotModifiable→400
- ErrModificationRequestNotFound→404, ErrModificationRequestPending→409, ErrModificationRequestExpired→400
- ErrExtensionRequestNotFound→404, ErrExtensionRequestPending→409, ErrExtensionRequestExpired→400
- ErrSavedCardNotFound→404, ErrSavedCardLimitReached→409
- ErrRegionSwitchBlocked→409, ErrCrossRegionalBooking→403
- ErrRegionSameAsCurrent→400, ErrRegionInvalid→400
- ErrSeasonalTariffOverlap→409, ErrSeasonalTariffNotFound→404
- ErrShareTokenNotFound→404, ErrShareTokenExpired→400
- ErrImportValidationFailed→400, ErrImportFileTooLarge→400
- ErrDepositNotFound→404, ErrDepositAlreadyReleased→409, ErrDepositAlreadyClaimed→409
- ErrClientReviewNotFound→404, ErrReviewBlindPeriod→403
- ErrBankEntryNotFound→404, ErrBankEntryAlreadyMatched→409
- ErrFraudDetected→403
- ErrAdminRoleNotFound→404, ErrAdmin2FARequired→403, ErrAdminPermissionDenied→403
- ErrFAQNotFound→404
- ErrWebhookNotFound→404, ErrWebhookLimitReached→409
- ErrPMSConnectionNotFound→404, ErrPMSConnectionLimitReached→409, ErrPMSConnectionAlreadyExists→409, ErrPMSSyncFailed→400
- ErrPhotoOrderNotFound→404, ErrPhotoOrderInvalidStatus→400

### Logging

Structured logging via zap (wrapped in `internal/logger/`).

```go
logger.Info("message", "key", value)  // also Error, Debug, Warn
```

Config: `BANI_LOGGER_LEVEL` (default `info`), `BANI_LOGGER_FORMAT` (`json` for production, `console` for development). Logger injected via fx.

### Config

Viper with env prefix `BANI_`. Nested keys use `_`: `BANI_DATABASE_DSN`, `BANI_JWT_SECRET`, etc.

Key config variables:
- `BANI_CORS_ALLOWED_ORIGINS` — comma-separated allowed origins (default `*`; wildcard rejected in production)
- `BANI_DATABASE_MAX_CONNS` (default 20), `BANI_DATABASE_MIN_CONNS` (default 2), `BANI_DATABASE_MAX_CONN_LIFETIME` (default 1h) — connection pool tuning
- `BANI_LOGGER_FORMAT` — `json` (default) or `console`
- `BANI_ENVIRONMENT` — `dev` (default) or `production` (enables stricter validation)
- `BANI_ESCROW_CLAIM_HOURS` — escrow hold period before release to owner (default 48, range 24-168)
- `BANI_WALLET_REFUND_BONUS_PERCENT` — bonus % when client chooses wallet refund (default 5, range 0-15)
- `BANI_MAX_CARD_HOLD_HOURS` — max card authorization hold duration (default 72, YooKassa limit)
- `BANI_FISCAL_PROVIDER` — fiscalization provider: `none` (default) or `atol`
- `BANI_FISCAL_ATOL_LOGIN`, `BANI_FISCAL_ATOL_PASSWORD`, `BANI_FISCAL_ATOL_GROUP_CODE` — ATOL Online credentials
- `BANI_WELCOME_BONUS_AMOUNT_BY` (default 1500) — welcome bonus for BY region in kopecks (15 BYN)
- `BANI_CRON_ENABLED` (default `true`) — enable/disable cron scheduler
- `BANI_CRON_TIMEZONE` (default `Europe/Moscow`) — timezone for cron job scheduling

### Database

PostgreSQL with PostGIS. Migrations in `migrations/`. Geo-search uses `ST_DWithin`/`ST_Distance` with `geography` type. Connection pool is tuned via `BANI_DATABASE_MAX_CONNS`, `BANI_DATABASE_MIN_CONNS`, `BANI_DATABASE_MAX_CONN_LIFETIME`.

### Testing

- Services: mock repos from `repository/mock/` + table-driven tests
- Handlers: httptest + mock services
- No integration tests for postgres repos (require real DB)
- Hurl tests in `tests/hurl/` for full API endpoint testing

### CI/CD

GitHub Actions workflow (`.github/workflows/ci.yml`) runs on push/PR to `main`:
- golangci-lint v2 (govet, staticcheck enabled)
- `go vet ./...`
- `go test ./... -race` with coverage
- `go build ./...`

### OpenAPI / Swagger

All API endpoints are annotated with swaggo/swag comments. Swagger UI is served at `/swagger/`.

- `docs/` — auto-generated OpenAPI spec (docs.go, swagger.json, swagger.yaml), committed for CI
- `cmd/server/docs.go` — main API annotations (@title, @version, @BasePath, @securityDefinitions)
- Annotations live in handler method comments (`internal/handler/*.go`)
- After modifying handler annotations, run `make swagger` to regenerate the spec
- Run `make swagger-fmt` to auto-format annotation comments

## Critical Conventions

- Module path: `github.com/rekurt/relax-hub`
- chi router: `chi.URLParam(r, "id")` for URL params
- UUID (google/uuid) for all entity IDs
- Prices in **kopecks** (int64), not rubles
- Day-of-week: **0=Monday, 6=Sunday** (not Go's native 0=Sunday)
- Time format: HH:MM strings with string comparison (e.g., "09:00" < "14:30")
- Wraparound times supported: TimeFrom > TimeTo means spans midnight

## Admin Panel (GoAdmin)

Built on GoAdmin framework, enabled via `--with-admin` flag on the serve command.

- `internal/admin/` — GoAdmin engine, JWT auth bridge, fx module
- `internal/admin/pages/` — custom pages: dashboard, moderation, analytics, health, finance
- `internal/admin/pages/templates/` — Go html/template files with shared layout system
- `internal/admin/pages/static/` — self-hosted static assets (Chart.js)
- Config: `BANI_ADMIN_ENABLED` (default `false`), `BANI_ADMIN_PREFIX` (default `/admin-panel`), `BANI_ADMIN_LANGUAGE` (default `ru`), `BANI_ADMIN_THEME` (default `adminlte`)

```bash
go run ./cmd/server serve --with-admin   # start server with admin panel
```

### Shared Template System

All custom pages use a shared template hierarchy via Go `html/template`:

- `base.tmpl` — common HTML structure: sidebar navigation (links to all 4 pages + GoAdmin), breadcrumbs, shared CSS (reset, grid, cards, buttons, badges, responsive media queries), footer with "Последнее обновление"
- `components.tmpl` — reusable template blocks: `{{define "toast"}}`, `{{define "status-badge"}}`, `{{define "empty-state"}}`, `{{define "loading-spinner"}}`, `{{define "pagination"}}`
- Page templates extend base via `{{template "base" .}}` and `{{define "content"}}...{{end}}`

Template functions available: `statusRu` (EN→RU status mapping), `jsEscape` (XSS-safe JS string escaping), `relativeTime` (human-readable relative timestamps), `formatBytes` (byte size formatting).

### Custom Pages

- **Dashboard**: clickable KPI cards with trend indicators (↑/↓%), revenue sub-cards (today/week/month), relative time in activity feeds, auto-refresh toggle (60s interval)
- **Moderation Center**: inline AJAX approve/reject (no page reload), toast notifications, loading states, lightbox for review images, reject modal with reason validation and free-text comment, keyboard shortcuts, batch operations with confirmation
- **Analytics**: date presets (Today/7d/30d/Month/Year), summary row with period comparison, CSV export, self-hosted Chart.js, city filter on all charts
- **Health Monitor**: system metrics (uptime, goroutines, memory, GC), pgxpool connection stats with progress bars, filesystem check, DB size, severity-colored backlog (green/yellow/orange/red), configurable auto-refresh (15/30/60s) with pause/play
- **Finance Dashboard**: float monitoring (client wallets + owner wallets + escrow totals), transaction reconciliation status, revenue breakdown (service fees, subscriptions, promotions), wallet metrics widget

### Admin Routes

Custom pages are mounted under `{BANI_ADMIN_PREFIX}/pages/`:

- `GET /` and `GET /dashboard` — dashboard
- `GET /moderation` — moderation page
- `POST /moderation/api/{approve,reject,batch-approve,batch-reject}` — moderation actions (return JSON)
- `GET /analytics` — analytics page
- `GET /analytics/export?type={bookings|revenue|users|top_bookings|top_revenue}&date_from=&date_to=&city_id=` — CSV export
- `GET /static/*` — self-hosted static assets
- `GET /health` — health monitor
- `GET /finance` — financial dashboard

### Moderation Keyboard Shortcuts

- `a` — approve selected (or focused card)
- `r` — open reject modal
- `Escape` — close reject modal
- `Enter` in reject modal — confirm rejection
- `Ctrl+A` — select all on page
- `→` / `←` — next/previous page

Auth bridges JWT tokens from the main app to GoAdmin sessions. Menu configured in `engine.go`.

## Frontend (Multi-role SPA)

React SPA covering all three roles: client (search, book, review, loyalty), owner/representative (manage bathhouses, bookings, pricing), admin (moderation, analytics, user management). Role-based layouts with separate navigation per role.

Directory: `frontend/`

### Frontend Commands

```bash
# Development
cd frontend && npm run dev              # dev server with HMR (proxies /api -> localhost:8080)
make frontend-dev                       # same via Makefile

# Build
cd frontend && npm run build            # TypeScript check + production build to dist/
make frontend-build                     # same via Makefile

# API Client Generation
cd frontend && npm run generate:api     # regenerate API client from docs/swagger.json
make frontend-generate-api              # same via Makefile

# Lint
cd frontend && npm run lint             # ESLint (zero warnings policy)

# Test
cd frontend && npx vitest run           # run all tests
cd frontend && npx vitest run src/__tests__/Dashboard.test.tsx  # single test
```

### Frontend Tech Stack

- Vite 6 + React 18 + TypeScript 5.6 (strict mode)
- Ant Design 6 (UI components, Russian locale)
- React Router 7 (client-side routing)
- TanStack React Query 5 (server state, generated via orval)
- orval (API client generation from OpenAPI spec)
- zustand (client state: auth token, selected bathhouse, persisted to localStorage)
- Vitest + React Testing Library
- dayjs (dates, built into antd)

### Frontend Structure

```
frontend/src/
├── api/
│   ├── axios-instance.ts      # Axios with JWT interceptor, 401 -> logout
│   └── generated/             # Auto-generated by orval (DO NOT EDIT)
│       └── model/             # Shared TypeScript types
├── components/                # Reusable: AppLayout, ClientLayout, AdminLayout, BathhouseCard,
│   │                          #   BathhouseSelector, MediaUploader, NotificationBell, OAuthButtons,
│   │                          #   ProtectedRoute, ReportModal, ReviewCard, OnboardingTour,
│   │                          #   BathhouseMap, RecentlyViewed, ShareButton, ProfileCompleteness,
│   │                          #   EmptyState, ApplePayButton, GooglePayButton, SupportChatBot
├── pages/
│   ├── admin/                 # Admin role pages:
│   │   ├── AdminDashboard     #   KPI analytics, top bathhouses, support ops metrics widget
│   │   ├── UserManagement     #   User list, block/unblock, batch operations
│   │   ├── BathhouseModeration#   Approve/reject bathhouses, batch operations
│   │   ├── ReviewModeration   #   Single & batch review moderation
│   │   ├── PhotoVerification  #   Verify/reject bathhouse photos
│   │   ├── ComplaintManagement#   Resolve/dismiss complaints
│   │   ├── CityManagement     #   City CRUD
│   │   ├── GlobalPromoCodes   #   Global promo code creation
│   │   ├── AdminNotifications #   Admin notifications
│   │   ├── AdminProfile       #   Admin profile settings
│   │   ├── AmenityManagement  #   Amenity CRUD with icons
│   │   ├── ObjectTypeManagement # Bathhouse category CRUD
│   │   ├── HolidayManagement  #   Holiday management per region
│   │   ├── AntiFraudDashboard #   Fraud flag review queue
│   │   ├── TicketManagement   #   Support ticket admin queue, SLA timers
│   │   ├── AdminTicketDetail   #   Admin ticket response/escalation
│   │   ├── DisputeManagement  #   Dispute mediation queue
│   │   ├── AdminDisputeDetail #   Admin dispute resolution
│   │   ├── WalletManagement   #   Admin wallet credit/debit/freeze
│   │   ├── BookingManagement  #   Admin booking cancel/refund/status
│   │   ├── RoleManagement     #   Admin sub-role assignment
│   │   ├── AdminNotificationCenter # Critical alerts per role
│   │   ├── AdminFinanceDashboard # Float monitoring, GMV, reconciliation
│   │   ├── AdminAuditLog      #   Admin action audit trail
│   │   ├── BankReconciliation #   Bank statement import & matching
│   │   ├── PlatformSettings   #   Key-value platform config editor
│   │   ├── FeatureFlags       #   Feature flag toggles with region scoping
│   │   ├── ServiceFeeConfig   #   Service fee by region/category
│   │   ├── ConversionFunnels  #   Visit->search->book->complete funnels
│   │   ├── CohortAnalysis     #   Registration cohort retention
│   │   ├── SupplyDemandMetrics#   Listings, occupancy, DAU/MAU metrics
│   │   ├── ForceMajeure       #   Emergency mass booking cancellation
│   │   ├── SubscriptionManagement # Owner subscription tiers
│   │   ├── LoyaltyManagement  #   Loyalty tier configuration
│   │   ├── CertificateManagement # Certificate search & void
│   │   └── GeoHeatmap         #   Geographic demand/supply heatmap
│   ├── client/                # Client role pages:
│   │   ├── ClientHome         #   Home page: search, recently viewed, recommendations
│   │   ├── BathhouseSearch    #   Search with filters, geo-search, map/list/split view
│   │   ├── BathhouseDetail    #   Full info, gallery, reviews, slots, share
│   │   ├── BookingCreate      #   4-step booking wizard with combo payment
│   │   ├── BookingList        #   Client bookings with status filters
│   │   ├── BookingDetail      #   Booking info, cancel, payment, share
│   │   ├── ReviewForm         #   Create/edit review with media
│   │   ├── Favorites          #   Favorite bathhouses grid
│   │   ├── Recommendations    #   Personalized + popular bathhouses
│   │   ├── Preferences        #   User preference settings
│   │   ├── LoyaltyDashboard   #   Loyalty level, points, transactions
│   │   ├── ReferralProgram    #   Referral code, stats, balance
│   │   ├── CertificateList    #   Owned certificates
│   │   ├── CertificatePurchase#   Purchase flow
│   │   ├── PaymentHistory     #   Payment list with filters
│   │   ├── ClientProfile      #   Profile, social accounts, region switch, account deletion
│   │   ├── ClientChat         #   Real-time chat with bathhouses
│   │   ├── ClientNotifications#   Notification list
│   │   ├── SavedSearches      #   Saved search management
│   │   ├── ComparisonPage     #   Side-by-side bathhouse comparison
│   │   ├── SupportTickets     #   Support ticket list
│   │   ├── TicketDetail       #   Ticket thread with CSAT
│   │   ├── DisputeCreate      #   Open dispute with evidence
│   │   ├── DisputeDetail      #   Dispute status and appeal
│   │   ├── DisputeList        #   Client disputes list
│   │   ├── NotificationPreferences # Per-event channel toggles
│   │   ├── WalletDashboard    #   Balance, top-up, transactions, export
│   │   ├── SavedCards         #   Saved payment cards management
│   │   ├── SecuritySettings   #   Sessions, 2FA, password change
│   │   └── ActivePromoCodes   #   Available promo codes list
│   ├── bathhouses/            # Owner: BathhouseList, BathhouseForm (7-step wizard), ListingImport, AuditLog
│   ├── bookings/              # Owner: BookingList (check-in/out), BookingDetails, ExtensionRequests, ModificationRequests
│   ├── calendar/              # Owner: CalendarPage (weekly view, slot management)
│   ├── chat/                  # Owner: ChatPage, ConversationList, MessageArea (content filter warning)
│   ├── notifications/         # Owner: NotificationList
│   ├── photos/                # Owner: PhotoManager (upload, reorder, status), PhotoOrderPage (professional photography)
│   ├── promotion/             # Owner: PromotionCampaign (auction-based promotion management)
│   ├── pricing/               # Owner: PricingRules (dynamic pricing, smart pricing recommendations)
│   ├── finance/               # Owner: FinanceDashboard, PayoutPage, FinancialReports
│   ├── analytics/             # Owner: OwnerAnalytics (occupancy, income, conversion, competitor benchmarks)
│   ├── promo/                 # Owner: PromoList (promo code management)
│   ├── representatives/       # Owner: RepresentativeList
│   ├── reviews/               # Owner: ReviewList (with media display, response form)
│   ├── settings/              # Owner: ProfileSettings, WebhookSettings, PMSIntegration
│   ├── subscriptions/         # Owner: SubscriptionPage
│   ├── crm/                   # Owner CRM:
│   │   ├── GuestCardList      #   Guest card search/filter/export
│   │   ├── GuestCardDetail    #   Visit history, LTV, notes, tags
│   │   ├── SegmentList        #   Dynamic segments (new/regular/lost/VIP)
│   │   ├── BroadcastList      #   Broadcast message management
│   │   ├── BroadcastCreate    #   Create broadcast to segment
│   │   ├── AutoScenarios      #   Auto-scenario toggle/customize
│   │   ├── ResponseTemplates  #   Quick reply template CRUD
│   │   ├── RFMAnalysis        #   RFM scoring matrix
│   │   └── SegmentBuilder     #   Custom segment constructor
│   └── widget/                # Owner: WidgetSettings
├── stores/                    # Zustand: auth.ts (user/token/role), bathhouse.ts (selected bathhouse)
├── lib/                       # format.ts, constants.ts, useWebSocketNotifications.ts, useDeviceToken.ts
├── __tests__/                 # Vitest component and unit tests
├── router.tsx                 # Role-based route groups: /, /client/*, /admin/*
└── main.tsx                   # Entry point with providers
```

### Frontend Key Patterns

- Three role-based layouts: AppLayout (owner), ClientLayout (client), AdminLayout (admin) with separate navigation
- Role-based route groups: `/` for owner, `/client/*` for client, `/admin/*` for admin
- ProtectedRoute supports `allowedRoles` prop for route-level access control
- Role-based redirect after login: client -> /client, owner -> /, admin -> /admin
- OAuth social login (VK, Yandex, Google) on Login/Register pages
- API client is auto-generated: run `npm run generate:api` after changing backend swagger annotations
- `src/api/generated/` is gitignored-style: never edit manually, always regenerate
- Prices displayed via `formatPrice()` from `lib/format.ts` (kopecks -> rubles with ₽ symbol)
- Dev server proxies `/api` -> `http://localhost:8080` and `/ws` -> `ws://localhost:8080`
- Path alias: `@/` maps to `src/` in imports
- Multi-step wizard pattern: BathhouseForm (7-step), BookingCreate (4-step) use Ant Design Steps with draft save per step
- Share functionality: Web Share API with clipboard fallback via ShareButton component
- Combo payments: wallet + card split with visual slider in BookingCreate
- Push notifications: permission requested after first booking completion (not at registration)
- ShareRedirect page at `/share/booking/:token` resolves deep links

## Feature Subsystems

Each subsystem follows the same handler→service→repository pattern:

- **Notification system**: multi-channel (in-app/email/telegram), dispatcher routes by user preferences, WebSocket hub for real-time
- **OAuth**: VK, Yandex, Google social login. Config: `BANI_OAUTH_{PROVIDER}_{CLIENT_ID,CLIENT_SECRET,REDIRECT_URL}`
- **Recommendations**: collaborative filtering + user preferences scoring
- **Subscriptions**: free/premium/promoted tiers, affects feed sorting (+10 boost for premium, promoted first)
- **Dynamic pricing**: rules with priority, multipliers applied per hourly slot. Per-day pricing: individual price per day of week (Mon-Sun) via DaysOfWeek field on any rule type. Extended with long session discounts (threshold hours + discount %), extra guest surcharges (per extra guest per hour), holiday pricing (recurring holidays with per-bathhouse multipliers, default 1.5x), last-minute discounts (configurable threshold hours + discount %), seasonal tariffs (date range + multiplier). Full price breakdown in booking response: base_price, long_session_discount, extra_guest_surcharge, service_fee, holiday info, seasonal_tariff. Last-minute badge (`is_last_minute`, `last_minute_discount_percent`) included in search results
- **Loyalty**: bronze/silver/gold/platinum tiers based on visit count, points system. Cashback to wallet on booking completion: Bronze 0%, Silver 3%, Gold 5%, Platinum 10% of base price (tag: `cashback`)
- **Chat**: real-time via WebSocket, conversations tied to bathhouse+client pair
- **Telegram bot**: booking wizard with in-memory state, short ID cache for callback data (64-byte limit)
- **Complaints**: report reviews/bathhouses/users (spam, offensive, fake, fraud, other), admin moderation queue with resolve/dismiss, auto-hide reviews at 3+ reports
- **Referral program**: personal referral codes, bonus on first booking completion (500 rub default to both referrer and referee), referral balance usable on bookings
- **Gift certificates**: purchasable with or without auth, unique BANI-XXXX-XXXX codes, partial redemption with balance tracking, 365-day validity
- **Photo verification**: admin-verified bathhouse photos with pending/verified/rejected statuses, `is_photo_verified` badge on bathhouse cards, owner/representative upload with admin moderation queue
- **Promo codes**: percentage/fixed_amount/free_hour/free_addon discount types, bathhouse-scoped (owner/representative) and global (admin) codes, usage limits, validity periods, min amount checks, integrated into booking creation discount chain. `free_addon` type zeroes out a specific add-on price via `target_addon_id`
- **Review media**: photo/video attachments on reviews (max 10 photos, 1 video per review), file type/size validation, image resize to 4 sizes (thumbnail 300px, medium 800px, large 1200px, full 1920px), WebP conversion, blur-hash placeholder generation. Bathhouse gallery endpoint aggregates review media with review status filtering. Media model stores `medium_url`, `large_url`, `blur_hash` fields
- **Cancellation policies**: per-bathhouse policy (flexible/moderate/strict) with configurable time windows and refund percentages. Flexible: 100% if >24h, 50% if <24h. Moderate: 100% if >72h, 50% if 24-72h, 0% if <24h. Strict: 100% if >7d, 50% if 3-7d, 0% if <3d. `internal/domain/cancellation_policy.go`
- **Booking modifications**: two-sided approval flow — client submits modification request (date/time/duration/guests/addons), owner approves or rejects. Max 3 modifications per booking. Price difference charged or refunded on approval. 24h auto-reject timeout (cron). Statuses: pending/approved/rejected/expired. `internal/domain/booking_modification.go`, `internal/repository/postgres/booking_modification_repo.go`. Endpoints: `PUT /api/v1/bookings/{id}/modify` (create request), `POST /api/v1/my/bookings/{id}/modification/approve`, `/reject`
- **Security deposit**: per-bathhouse configurable deposit (0-50% of base price). Card hold on booking creation, auto-release 48h after check-out if no dispute, freeze on dispute. Deposit statuses: none/held/released/claimed
- **Seasonal tariffs**: date-range pricing multipliers per bathhouse (name, date_from, date_to, multiplier). Applied in price calculation pipeline after base price. CRUD endpoints for owner management. `internal/domain/seasonal_tariff.go`, `internal/repository/postgres/seasonal_tariff_repo.go`
- **Online payments**: Multi-region payment via PaymentProvider interface. RU: YooKassa, BY: bePaid — provider factory selects by user region (`internal/payment/provider_factory.go`). Payment methods: card, SBP, wallet, combo, MIR, Belkart, ERIP, Apple Pay, Google Pay. Token-based payments for Apple/Google Pay. Saved card tokenization (`internal/domain/saved_card.go`): store provider tokens with last4/brand/expiry, one-click repeat payments. Payment error retry with exponential backoff (3 attempts: 2s, 4s, 8s). Combo payments: wallet first, card remainder, rollback on failure. Payment holds for request-based bookings. Enhanced refunds: wallet refund with bonus (default 5%), proportional combo refund, admin manual refund. SBP instant payouts for RU owners via YooKassa Payouts API (auto-selected when owner has phone; fallback: bank transfer). Payout method field: `sbp` or `bank_transfer`. Config: `BANI_PAYMENT_YOOKASSA_SHOP_ID`, `BANI_PAYMENT_YOOKASSA_SECRET_KEY`, `BANI_PAYMENT_BEPAID_SHOP_ID`, `BANI_PAYMENT_BEPAID_SECRET_KEY`, `BANI_PAYMENT_RETURN_URL`
- **Wallet system**: user balance with top-up/spend/hold/refund, priority spending (expiring bonuses first), balance limits (max 100,000 RUB), top-up limits (min 500, max 30,000 RUB per tx), bonus expiration cron (180 days, configurable). Owner payouts with daily/monthly limits and auto-payout threshold. Auto-payout cron (hourly): owners with `auto_payout_threshold > 0` and `balance >= threshold` get automatic payout via `PayoutService.RequestPayout()`. `internal/cron/auto_payout.go`
- **Phone + OTP auth**: Redis-backed 6-digit codes, 5 min TTL, 3 attempts, rate limiting. SMSProvider interface + SMS.ru adapter. Config: `BANI_SMS_PROVIDER`, `BANI_SMS_API_KEY`
- **Two-factor authentication**: TOTP (pquerna/otp) + SMS 2FA, partial token flow for 2FA during login
- **Session management**: device/browser/IP tracking, auto-expire after 30 days, session validation in auth middleware
- **Password reset**: Redis token, email link, rate limited, terminates all sessions on reset
- **Account deletion**: 30-day grace period with restore, data anonymization on execution, fund return for top-ups, cron reminders at day 0/14/27
- **Age verification**: `age_confirmed` required on registration, auto-create wallet + regional welcome bonus (500 RUB for RU, 15 BYN for BY, 30 day expiry). Config: `BANI_WELCOME_BONUS_AMOUNT`, `BANI_WELCOME_BONUS_AMOUNT_BY` (default 1500 kopecks = 15 BYN), `BANI_WELCOME_BONUS_EXPIRY_DAYS`
- **KYC system**: entity types (individual, sole_proprietor, self_employed, legal_entity), submit/approve/reject flow with expiry checking, admin moderation queue
- **Offer/contract acceptance**: versioned offer acceptance tracking, required before listing creation
- **Owner payment details**: entity-type-specific fields and validation, required before listing creation
- **Listing draft wizard**: 7-step draft creation flow, step-based data storage with CRUD + submit
- **Onboarding gate**: CreateBathhouse gated behind KYC approval + offer acceptance + payment details
- **Listing completeness**: required/optional field checklist for bathhouse listings, blocks moderation submission if incomplete
- **Audit log**: tracks bathhouse edits with JSONB diff, substantial changes (address/city/photos) auto-trigger re-moderation, admin and owner history views
- **Listing duplication**: copy bathhouse with all fields as draft, name + " (копия)", new slug/UUID
- **Listing lifecycle**: deactivate (inactive, keep bookings) / activate / archive (no active bookings, irreversible), search excludes inactive/archived
- **Add-ons**: configurable per-bathhouse extras (per_item/per_hour/per_person pricing), max 20 per bathhouse, integrated into booking creation with denormalized pricing in booking_addons
- **Full-text search**: PostgreSQL tsvector with Russian language config, pg_trgm for fuzzy matching, GIN indexes, prefix search support
- **Search suggestions**: bathhouse names (trigram), city names, popular queries (Redis sorted set), 5-min cache
- **Advanced ranking**: composite score (relevance 0.30 + bayesian_rating 0.25 + conversion_rate 0.20 + occupancy_rate 0.15 + promotion_boost 0.10), extended sort options
- **Bathhouse comparison**: compare 2-3 bathhouses side by side (price, rating, capacity, amenities, etc.), optional distance calculation
- **Saved searches**: JSONB filter storage, daily cron checks for new matches with notifications, max 50 per user
- **Recently viewed**: Redis sorted set per user (last 20), recorded on bathhouse detail view
- **Service fee**: platform fee on booking base price (not add-ons), configurable by region+category with global default (10%). Admin CRUD via `ServiceFeeConfig`. Config stored in `service_fee_configs` table
- **Bidirectional reviews**: owner rates client (punctuality, cleanliness, rule_compliance). Double-blind reveal: reviews hidden until both posted OR 14 days elapsed. Auto-reveal cron job. `internal/domain/client_review.go`, `internal/service/client_review_service.go`
- **Region switching**: users can switch between RU and BY regions. Blocked if non-zero wallet balance, active bookings, open disputes, or unactivated certificates. On switch: archive old wallet, create new wallet in new currency, reset loyalty. Cross-regional bookings blocked. `internal/service/region_service.go`
- **Financial reports**: wallet history export (CSV/PDF), act generation for owners (PDF), 1C XML export for legal entities. `GET /api/v1/my/wallet/export?format=csv|pdf`, `GET /api/v1/my/finance/acts`. `internal/service/financial_report_service.go`
- **Transaction reconciliation**: daily float snapshot (client wallets + owner wallets + escrow = expected total), YooKassa transaction reconciliation with zero-tolerance discrepancy alerting. Admin dashboard widget. Daily cron job. `internal/service/reconciliation_service.go`
- **Bonus expiry notifications**: cron job sends push + email for bonuses expiring in 14 days and 3 days, with deduplication
- **Mass listing import**: CSV and Excel (.xlsx) upload with validation, creates listings as drafts (pending moderation), returns import report with per-row errors. Downloadable xlsx template. Auto-detects format by Content-Type. `POST /api/v1/my/listings/import`. `internal/service/listing_import_service.go`. Dependency: `github.com/xuri/excelize/v2`
- **Share listing/booking**: Open Graph meta tags on bathhouse pages, shareable booking links with pre-filled params. `POST /api/v1/bookings/{id}/share`, deep link resolution at `/share/booking/{token}`. `internal/handler/share_handler.go`
- **Representative sub-roles**: manager (manage bookings, reply messages) and observer (read-only). Per-bathhouse access control. Invite-by-email flow
- **Smart pricing**: recommended price based on occupancy, area averages, demand patterns. Coefficient range 0.8-1.5. `GET /api/v1/my/bathhouses/{id}/price-recommendation`. `internal/service/smart_pricing_service.go`
- **Advanced analytics**: conversion funnels (visit->search->book->complete), cohort analysis by registration month, geographic demand/supply, wallet metrics, owner analytics with competitor benchmarking
- **Onboarding tour**: profile completeness calculation (name, photo, phone, preferences), step-by-step guide component, onboarding_completed flag
- **Booking modes**: instant (default, immediate confirmation) and request (owner approval required within timeout). Request-based: status `pending_owner`, wallet/card hold, auto-reject on timeout. Booking statuses: pending, pending_owner, confirmed, cancelled, rejected, completed, no_show. Concurrent booking control: `pg_advisory_xact_lock` serializes availability check + creation within a single transaction to prevent double-booking race conditions
- **Buffer/lead time**: configurable per-bathhouse buffer between bookings (0-120 min, step 15), lead time before booking (0-48h), max advance days (7-365)
- **Check-in/check-out**: owner/rep marks guest arrival (window: start-15min to start+30min) and departure. No-show detection cron (30min after start without check-in). No-show dispute within 2h. GPS validation on no-show dispute: client coordinates checked against bathhouse location (200m radius via Haversine); distance > 200m marks evidence as weak
- **Escrow**: payment held after check-out, released to owner after claim period (default 48h). Dispute blocks release. Cron auto-releases matured escrows hourly. `internal/service/escrow_service.go`
- **Session extension**: two-sided approval flow — client requests extension (1-2h), funds are held, owner approves or rejects within 30 min timeout. On approval: funds charged, booking end time updated. On rejection/expiry: hold released. `internal/domain/extension_request.go`. Endpoints: `POST /api/v1/bookings/{id}/extend` (create request), `POST /api/v1/my/bookings/{id}/extension/approve`, `/reject`
- **Re-booking**: `GET /api/v1/bookings/{id}/rebook-data` returns past booking parameters (duration, time, guests, add-ons) for quick re-creation
- **Owner penalties**: owner cancellation credits client 10% compensation. 4+ cancellations/30 days = warning, 6+ = auto-deactivation of all bathhouses. Response rate tracking for request-mode bathhouses (daily cron): rate < 30% sets `low_response_rate_since`; if persists > 60 days → forced switch to instant booking mode or deactivation. Rate >= 30% resets tracking
- **Fiscalization**: FiscalProvider interface with ATOL placeholder and no-op provider. Receipt creation on payment success and refund (best-effort). Entity-type adapted tax system: individual (no VAT), sole_proprietor/legal_entity (VAT 20%), self_employed (NPD). `internal/fiscal/`
- **Booking reminders**: cron every 15min sends 24h reminder (push+email), 2h reminder (push), 5min owner reminder. Redis-based deduplication
- **Cron scheduler**: centralized robfig/cron scheduler (`internal/cron/`) with distributed Redis locking, panic recovery, structured logging. 29+ registered jobs: every 15min (booking timeout, no-show, reminders, modification request timeout), every 30min (extension request timeout), hourly (escrow release, review requests, auto-scenarios, anti-fraud patterns, auto-payout), daily midnight Moscow (bonus expiry, session cleanup, KYC check, owner response rate, promo deactivation, saved searches, metrics, ticket auto-close). Config: `BANI_CRON_ENABLED`, `BANI_CRON_TIMEZONE`
- **Anti-fraud engine**: rule-based fraud detection (`internal/antifraud/`). Client rules: multi-card top-up, top-up/cancel cycles, dormant balance, rapid bookings. Owner rules: self-booking, structuring, fake reviews. Listing creation rules: duplicate detection by phone/email/payment details, stoplist check (`antifraud_stoplist` table). Actions: block (prevents operation), flag (allows + notifies admin), freeze_wallet. Hooked into wallet/booking/payout/bathhouse services. Admin review at `/api/v1/admin/antifraud/flags`. Admin stoplist management: `POST/DELETE /api/v1/admin/antifraud/stoplist`
- **Reviews multi-criteria**: 4-criteria ratings (cleanliness, accuracy, communication, value_for_money) with 0.5 step, overall = average. Bayesian average rating per bathhouse (m=5, C=platform avg, cached in Redis). Auto review request cron (2h after check-out, configurable). Quality monitoring: rating < 3.0 warns owner, < 2.0 auto-depublishes. Computed badges: Verified, Top, Premium, New. NLP auto-moderation: regex profanity/spam filter, score > 0.7 → pending_moderation. Config: `BANI_REVIEW_REQUEST_DELAY_HOURS`
- **CRM for owners**: guest cards (auto-created on booking completion, visit count, LTV, avg check, notes, tags, CSV export). Dynamic segments: new (1 visit), regular (>=3), lost (>90 days), vip (>50k RUB), birthday_soon (7 days). Broadcasts to segments (rate limit 3/week per owner, 1/3 days per guest, respects notification prefs). Auto-scenarios: thank_after_visit, request_review, remind_revisit_30d, reactivate_lost_90d, birthday_greeting (hourly cron). Response templates with default seeding (max 50 per owner). Endpoints under `/api/v1/my/crm/`
- **Support tickets**: category-based (question/problem/complaint/refund_request/account_issue), auto-priority, 3-level escalation (L1→L2 at 24h, L2→L3 at 48h, critical auto-L2). Ticket messages with attachments. CSAT survey 24h after resolution. Auto-close after 7 days resolved. User endpoints: `/api/v1/my/tickets`, admin endpoints: `/api/v1/admin/tickets`
- **Dispute system**: booking disputes with evidence collection (photo/screenshot/gps/message/receipt, 72h window), blocks escrow release. Resolution: full/partial/no refund + compensation via wallet. Appeal within 7 days. Admin mediation with assignment. Statuses: open → evidence_collection → under_review → resolved/appealed → closed. Endpoints: `/api/v1/bookings/{id}/dispute`, `/api/v1/my/disputes`, `/api/v1/admin/disputes`
- **Platform settings**: admin-configurable key-value store with typed values (int/float/string/bool/json), Redis cache (5 min TTL, prefix `platform:settings:`). 12 seeded settings: service_fee_percent, welcome_bonus_amount, escrow_claim_hours, etc. Admin endpoints: `GET /api/v1/admin/settings`, `PUT /api/v1/admin/settings/{key}`
- **Feature flags**: toggleable platform features with optional region scoping (via `cities.region`), Redis cache (1 min TTL). `IsEnabled(key)` and `IsEnabledForRegion(key, region)` checks. 13 seeded flags (wallet, phone auth, 2FA, KYC, add-ons, escrow, CRM, disputes, anti-fraud, SBP, etc.). Admin endpoints: `GET /api/v1/admin/feature-flags`, `PUT /api/v1/admin/feature-flags/{key}`
- **Force majeure**: mass booking cancellation by region for emergency situations. Admin activates with region + date range + reason, system cancels all confirmed bookings in affected region, issues 100% wallet refunds, notifies clients and owners. Audit trail via `force_majeure_events` table. Booking status: `force_majeure_cancelled`. Admin endpoints: `POST /api/v1/admin/force-majeure`, `GET /api/v1/admin/force-majeure`
- **Admin sub-roles & permissions**: 6 granular admin sub-roles (super_admin, moderator, support_l1, support_l2, support_l3, finance) with 27-permission matrix. Middleware `RequireAdminPermission(permission)` for endpoint-level access control. Mandatory 2FA for all admin sub-roles. `internal/domain/admin_permission.go`, `migrations/XXXX_admin_roles.up.sql`
- **Admin notifications**: role-targeted alert system with 8 notification types (antifraud_flag, sla_violation, reconciliation_mismatch, float_drift, ticket_escalation, dispute_opened, kyc_pending, system). 4 severity levels (info, warning, critical, urgent). Daily email digest cron aggregates unread critical alerts per admin role. `internal/domain/admin_notification.go`. Endpoints: `GET /api/v1/admin/notifications`, `PUT /api/v1/admin/notifications/{id}/read`
- **Admin mass operations**: batch operations for up to 1000 records with chunked DB transactions (100 per batch). Batch approve/reject listings, block/unblock users, credit wallets. Returns succeeded/failed arrays with reasons. `POST /api/v1/admin/listings/batch`, `POST /api/v1/admin/users/batch`, `POST /api/v1/admin/wallets/batch-credit`
- **Bank reconciliation**: bank statement import (CSV, 1C XML format) with auto-matching against internal payment transactions by amount + date (±1 day) + reference number. Manual match fallback for unmatched entries. Admin endpoints: `POST /api/v1/admin/finance/bank-statement`, `GET /api/v1/admin/finance/reconciliation`, `PUT /api/v1/admin/finance/reconciliation/{id}/match`. `internal/service/bank_reconciliation_service.go`
- **Owner webhooks**: outbound webhook delivery for 5 event types (booking.created, booking.confirmed, booking.cancelled, booking.completed, payment.received). HMAC-SHA256 signature in X-Webhook-Signature header. Async delivery with retry (3 attempts, exponential backoff). Max 20 webhooks per owner. Delivery log tracking. Owner endpoints: CRUD `/api/v1/my/webhooks`, `GET /api/v1/my/webhooks/{id}/deliveries`. `internal/domain/webhook.go`
- **PMS integration**: external Property Management System connectors (Yclients, Restoplace). PMSProvider interface: SyncBookings, SyncSchedule, PushBooking, PullBookings. Bidirectional sync every 15 min via cron. Credentials stored encrypted. Owner endpoints: CRUD `/api/v1/my/pms-connections`, `POST /api/v1/my/pms-connections/{id}/sync`. `internal/pms/`
- **Isochrone search**: travel-time based search zones ("15 min by car", "30 min by transit") via OpenRouteService API. Returns GeoJSON polygons, filtered via PostGIS ST_Within. Redis cache (1h TTL). `GET /api/v1/isochrone?lat=&lon=&mode=car|transit&minutes=15`. `internal/geo/isochrone.go`
- **Transport accessibility**: nearby metro/bus/parking discovery via Yandex Maps Search API. Distance calculation (Haversine). 7-day Redis cache per bathhouse. `GET /api/v1/bathhouses/{id}/transport`. `internal/geo/transport.go`
- **SSR/prerender**: bot user-agent detection middleware serves pre-rendered HTML for SEO crawlers (Yandex, Google, social). Pre-renders bathhouse listing, detail, reviews pages. OG/Twitter Card/Schema.org meta tags. Redis cache (1h TTL, invalidated on update). `internal/seo/renderer.go`
- **Professional photography**: order professional photographers from owner cabinet. Status flow: requested -> confirmed -> completed -> cancelled. Payment from owner wallet on completion. Admin manages photographer queue. Owner endpoint: `POST /api/v1/my/bathhouses/{id}/photo-order`. Admin endpoints: `GET /api/v1/admin/photo-orders`, `PUT /api/v1/admin/photo-orders/{id}`. `internal/domain/photo_order.go`
- **FAQ bot (L1 support)**: automated FAQ matching before ticket creation. Keyword/trigram search against incoming support messages. Shows top 3 matching FAQ answers. "Not helpful" escalates to L2 (creates ticket). 6 categories (booking, payment, cancellation, wallet, account, general). Admin CRUD: `/api/v1/admin/faq`. `internal/domain/faq.go`, `internal/service/faq_bot_service.go`
- **Notification preferences**: per-event/per-channel (push/email/SMS) notification preferences. 28+ event types. Mandatory events (booking confirmed, dispute resolution, payment, system) cannot be disabled. Fallback chain: push -> email (5 min) -> SMS (critical only). FCM delivery receipt tracking. `GET/PUT /api/v1/my/notification-preferences`. `internal/domain/notification_preferences.go`
- **RFM analysis**: Recency-Frequency-Monetary scoring (1-5 each) for CRM guest cards. `GET /api/v1/my/crm/rfm`. Custom segments with multi-condition filters (visit count, avg check, last visit days, tags, RFM ranges). Dynamic evaluation resolves segment to guest list. CRUD: `/api/v1/my/crm/segments/custom`. `internal/service/rfm_service.go`, `internal/domain/custom_segment.go`
- **Broadcast personalization**: template tokens ({{guest_name}}, {{last_visit_date}}, {{visit_count}}, {{promo_code}}) replaced with guest card data on send. SMS channel via SMSProvider. Broadcast click tracking. `internal/service/broadcast_service.go`
- **Moderation SLA**: moderation queue metrics — queue size, avg wait time, SLA compliance (48h), per-moderator throughput. Alerts when items > 24h without review. `internal/admin/pages/moderation.go`
- **Promoted listing campaigns**: auction-based promotion with min 50 rub/day bid, budget management. Daily wallet deduction via cron. Auto-pause on budget exhaustion. Campaign statistics: impressions, clicks, CTR, cost. Higher bid = weighted search ranking boost. `internal/service/promotion_service.go`
- **Area average price**: city-level average base price displayed on bathhouse detail page for comparison. SQL `AVG(base_price)` for active bathhouses in same city. Redis cache (1h TTL) by city_id. Field `area_average_price` in GetByID response
- **Support operations metrics**: FCR (first contact resolution %), AHT (average handling time), SLA compliance. Admin API endpoint for support dashboard widget. `internal/service/ticket_service.go`
- **P&L / unit economics**: GMV (total bookings value), Take Rate (platform revenue / GMV), unit economics (revenue and cost per booking). Admin finance dashboard widget. `internal/service/analytics_service.go`
- **Geo heatmap**: geographic demand/supply heatmap — aggregates search queries and bathhouse locations by grid cells. `GET /api/v1/admin/analytics/heatmap`. Frontend: Yandex Maps heatmap layer. `frontend/src/pages/admin/GeoHeatmap.tsx`
- **iCal export**: calendar export for bathhouse bookings. `GET /api/v1/my/bathhouses/{id}/calendar.ics` (Content-Type: text/calendar). `internal/handler/calendar_handler.go`
- **Slug-based frontend routes**: SEO-friendly URLs `/bathhouses/:slug` in SPA. BathhouseDetail loads by slug via `getBySlug` API. All internal links use slug instead of UUID
- **Satellite map layer**: Yandex Maps layer switcher on BathhouseMap component — toggle between schema and satellite views
