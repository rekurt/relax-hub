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

### Admin Routes

Custom pages are mounted under `{BANI_ADMIN_PREFIX}/pages/`:

- `GET /` and `GET /dashboard` — dashboard
- `GET /moderation` — moderation page
- `POST /moderation/api/{approve,reject,batch-approve,batch-reject}` — moderation actions (return JSON)
- `GET /analytics` — analytics page
- `GET /analytics/export?type={bookings|revenue|users|top_bookings|top_revenue}&date_from=&date_to=&city_id=` — CSV export
- `GET /static/*` — self-hosted static assets
- `GET /health` — health monitor

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
│   │                          #   ProtectedRoute, ReportModal, ReviewCard
├── pages/
│   ├── admin/                 # Admin role pages:
│   │   ├── AdminDashboard     #   KPI analytics, top bathhouses
│   │   ├── UserManagement     #   User list, block/unblock
│   │   ├── BathhouseModeration#   Approve/reject bathhouses
│   │   ├── ReviewModeration   #   Single & batch review moderation
│   │   ├── PhotoVerification  #   Verify/reject bathhouse photos
│   │   ├── ComplaintManagement#   Resolve/dismiss complaints
│   │   ├── CityManagement     #   City CRUD
│   │   ├── GlobalPromoCodes   #   Global promo code creation
│   │   ├── AdminNotifications #   Admin notifications
│   │   └── AdminProfile       #   Admin profile settings
│   ├── client/                # Client role pages:
│   │   ├── BathhouseSearch    #   Search with filters, geo-search
│   │   ├── BathhouseDetail    #   Full info, gallery, reviews, slots
│   │   ├── BookingCreate      #   Booking flow with promo/certificate/loyalty
│   │   ├── BookingList        #   Client bookings with status filters
│   │   ├── BookingDetail      #   Booking info, cancel, payment
│   │   ├── ReviewForm         #   Create/edit review with media
│   │   ├── Favorites          #   Favorite bathhouses grid
│   │   ├── Recommendations    #   Personalized + popular bathhouses
│   │   ├── Preferences        #   User preference settings
│   │   ├── LoyaltyDashboard   #   Loyalty level, points, transactions
│   │   ├── ReferralProgram    #   Referral code, stats, balance
│   │   ├── CertificateList    #   Owned certificates
│   │   ├── CertificatePurchase#   Purchase flow
│   │   ├── PaymentHistory     #   Payment list with filters
│   │   ├── ClientProfile      #   Profile, social accounts, notifications
│   │   ├── ClientChat         #   Real-time chat with bathhouses
│   │   └── ClientNotifications#   Notification list
│   ├── bathhouses/            # Owner: BathhouseList, BathhouseForm (create/edit)
│   ├── bookings/              # Owner: BookingList (with payment info), BookingDetails
│   ├── calendar/              # Owner: CalendarPage (weekly view, slot management)
│   ├── chat/                  # Owner: ChatPage, ConversationList, MessageArea
│   ├── notifications/         # Owner: NotificationList
│   ├── photos/                # Owner: PhotoManager (upload, reorder, status)
│   ├── pricing/               # Owner: PricingRules (dynamic pricing)
│   ├── promo/                 # Owner: PromoList (promo code management)
│   ├── representatives/       # Owner: RepresentativeList
│   ├── reviews/               # Owner: ReviewList (with media display, response form)
│   ├── settings/              # Owner: ProfileSettings
│   ├── subscriptions/         # Owner: SubscriptionPage
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

## Feature Subsystems

Each subsystem follows the same handler→service→repository pattern:

- **Notification system**: multi-channel (in-app/email/telegram), dispatcher routes by user preferences, WebSocket hub for real-time
- **OAuth**: VK, Yandex, Google social login. Config: `BANI_OAUTH_{PROVIDER}_{CLIENT_ID,CLIENT_SECRET,REDIRECT_URL}`
- **Recommendations**: collaborative filtering + user preferences scoring
- **Subscriptions**: free/premium/promoted tiers, affects feed sorting (+10 boost for premium, promoted first)
- **Dynamic pricing**: rules with priority, multipliers applied per hourly slot. Extended with long session discounts (threshold hours + discount %), extra guest surcharges (per extra guest per hour), holiday pricing (recurring holidays with per-bathhouse multipliers, default 1.5x), last-minute discounts (configurable threshold hours + discount %). Full price breakdown in booking response: base_price, long_session_discount, extra_guest_surcharge, service_fee, holiday info
- **Loyalty**: bronze/silver/gold/platinum tiers based on visit count, points system
- **Chat**: real-time via WebSocket, conversations tied to bathhouse+client pair
- **Telegram bot**: booking wizard with in-memory state, short ID cache for callback data (64-byte limit)
- **Complaints**: report reviews/bathhouses/users (spam, offensive, fake, fraud, other), admin moderation queue with resolve/dismiss, auto-hide reviews at 3+ reports
- **Referral program**: personal referral codes, bonus on first booking completion (500 rub default to both referrer and referee), referral balance usable on bookings
- **Gift certificates**: purchasable with or without auth, unique BANI-XXXX-XXXX codes, partial redemption with balance tracking, 365-day validity
- **Photo verification**: admin-verified bathhouse photos with pending/verified/rejected statuses, `is_photo_verified` badge on bathhouse cards, owner/representative upload with admin moderation queue
- **Promo codes**: percentage/fixed_amount/free_hour discount types, bathhouse-scoped (owner/representative) and global (admin) codes, usage limits, validity periods, min amount checks, integrated into booking creation discount chain
- **Review media**: photo/video attachments on reviews (max 10 photos, 1 video per review), file type/size validation, image resize and thumbnail generation, bathhouse gallery endpoint aggregates review media with review status filtering
- **Online payments**: YooKassa integration via PaymentProvider interface, automatic refund on booking cancellation (100% if >24h, 50% if 2-24h, 0% if <2h), webhook processing. Payment methods: card, SBP (sbp), wallet, combo (wallet + card/SBP). Combo payments: wallet debited first, card payment for remainder, rollback on failure. Payment holds for request-based bookings (capture=false). Enhanced refunds: wallet refund with bonus (default 5%), proportional combo refund, admin manual refund. Config: `BANI_PAYMENT_YOOKASSA_SHOP_ID`, `BANI_PAYMENT_YOOKASSA_SECRET_KEY`, `BANI_PAYMENT_RETURN_URL`
- **Wallet system**: user balance with top-up/spend/hold/refund, priority spending (expiring bonuses first), balance limits (max 100,000 RUB), top-up limits (min 500, max 30,000 RUB per tx), bonus expiration cron (180 days, configurable). Owner payouts with daily/monthly limits and auto-payout threshold
- **Phone + OTP auth**: Redis-backed 6-digit codes, 5 min TTL, 3 attempts, rate limiting. SMSProvider interface + SMS.ru adapter. Config: `BANI_SMS_PROVIDER`, `BANI_SMS_API_KEY`
- **Two-factor authentication**: TOTP (pquerna/otp) + SMS 2FA, partial token flow for 2FA during login
- **Session management**: device/browser/IP tracking, auto-expire after 30 days, session validation in auth middleware
- **Password reset**: Redis token, email link, rate limited, terminates all sessions on reset
- **Account deletion**: 30-day grace period with restore, data anonymization on execution, fund return for top-ups, cron reminders at day 0/14/27
- **Age verification**: `age_confirmed` required on registration, auto-create wallet + 500 RUB welcome bonus (30 day expiry). Config: `BANI_WELCOME_BONUS_AMOUNT`, `BANI_WELCOME_BONUS_EXPIRY_DAYS`
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
- **Booking modes**: instant (default, immediate confirmation) and request (owner approval required within timeout). Request-based: status `pending_owner`, wallet/card hold, auto-reject on timeout. Booking statuses: pending, pending_owner, confirmed, cancelled, rejected, completed, no_show
- **Buffer/lead time**: configurable per-bathhouse buffer between bookings (0-120 min, step 15), lead time before booking (0-48h), max advance days (7-365)
- **Check-in/check-out**: owner/rep marks guest arrival (window: start-15min to start+30min) and departure. No-show detection cron (30min after start without check-in). No-show dispute within 2h
- **Escrow**: payment held after check-out, released to owner after claim period (default 48h). Dispute blocks release. Cron auto-releases matured escrows hourly. `internal/service/escrow_service.go`
- **Session extension**: client extends active booking by 1-2h if next slots available, respects buffer/working hours. Creates separate extension payment
- **Re-booking**: `GET /api/v1/bookings/{id}/rebook-data` returns past booking parameters (duration, time, guests, add-ons) for quick re-creation
- **Owner penalties**: owner cancellation credits client 10% compensation. 4+ cancellations/30 days = warning, 6+ = auto-deactivation of all bathhouses. Response rate tracking for request-mode bathhouses (daily cron), low rate warnings and enforcement
- **Fiscalization**: FiscalProvider interface with ATOL placeholder and no-op provider. Receipt creation on payment success and refund (best-effort). `internal/fiscal/`
- **Booking reminders**: cron every 15min sends 24h reminder (push+email), 2h reminder (push), 5min owner reminder. Redis-based deduplication
- **Cron scheduler**: centralized robfig/cron scheduler (`internal/cron/`) with distributed Redis locking, panic recovery, structured logging. 26 registered jobs: every 15min (booking timeout, no-show, reminders), hourly (escrow release, review requests, auto-scenarios, anti-fraud patterns), daily midnight Moscow (bonus expiry, session cleanup, KYC check, owner response rate, promo deactivation, saved searches, metrics, ticket auto-close). Config: `BANI_CRON_ENABLED`, `BANI_CRON_TIMEZONE`
- **Anti-fraud engine**: rule-based fraud detection (`internal/antifraud/`). Client rules: multi-card top-up, top-up/cancel cycles, dormant balance, rapid bookings. Owner rules: self-booking, structuring, fake reviews. Actions: block (prevents operation), flag (allows + notifies admin), freeze_wallet. Hooked into wallet/booking/payout services. Admin review at `/api/v1/admin/antifraud/flags`
- **Reviews multi-criteria**: 4-criteria ratings (cleanliness, accuracy, communication, value_for_money) with 0.5 step, overall = average. Bayesian average rating per bathhouse (m=5, C=platform avg, cached in Redis). Auto review request cron (2h after check-out, configurable). Quality monitoring: rating < 3.0 warns owner, < 2.0 auto-depublishes. Computed badges: Verified, Top, Premium, New. NLP auto-moderation: regex profanity/spam filter, score > 0.7 → pending_moderation. Config: `BANI_REVIEW_REQUEST_DELAY_HOURS`
- **CRM for owners**: guest cards (auto-created on booking completion, visit count, LTV, avg check, notes, tags, CSV export). Dynamic segments: new (1 visit), regular (>=3), lost (>90 days), vip (>50k RUB), birthday_soon (7 days). Broadcasts to segments (rate limit 3/week per owner, 1/3 days per guest, respects notification prefs). Auto-scenarios: thank_after_visit, request_review, remind_revisit_30d, reactivate_lost_90d, birthday_greeting (hourly cron). Response templates with default seeding (max 50 per owner). Endpoints under `/api/v1/my/crm/`
- **Support tickets**: category-based (question/problem/complaint/refund_request/account_issue), auto-priority, 3-level escalation (L1→L2 at 24h, L2→L3 at 48h, critical auto-L2). Ticket messages with attachments. CSAT survey 24h after resolution. Auto-close after 7 days resolved. User endpoints: `/api/v1/my/tickets`, admin endpoints: `/api/v1/admin/tickets`
- **Dispute system**: booking disputes with evidence collection (photo/screenshot/gps/message/receipt, 72h window), blocks escrow release. Resolution: full/partial/no refund + compensation via wallet. Appeal within 7 days. Admin mediation with assignment. Statuses: open → evidence_collection → under_review → resolved/appealed → closed. Endpoints: `/api/v1/bookings/{id}/dispute`, `/api/v1/my/disputes`, `/api/v1/admin/disputes`
- **Platform settings**: admin-configurable key-value store with typed values (int/float/string/bool/json), Redis cache (5 min TTL, prefix `platform:settings:`). 12 seeded settings: service_fee_percent, welcome_bonus_amount, escrow_claim_hours, etc. Admin endpoints: `GET /api/v1/admin/settings`, `PUT /api/v1/admin/settings/{key}`
- **Feature flags**: toggleable platform features with optional region scoping (via `cities.region`), Redis cache (1 min TTL). `IsEnabled(key)` and `IsEnabledForRegion(key, region)` checks. 13 seeded flags (wallet, phone auth, 2FA, KYC, add-ons, escrow, CRM, disputes, anti-fraud, SBP, etc.). Admin endpoints: `GET /api/v1/admin/feature-flags`, `PUT /api/v1/admin/feature-flags/{key}`
- **Force majeure**: mass booking cancellation by region for emergency situations. Admin activates with region + date range + reason, system cancels all confirmed bookings in affected region, issues 100% wallet refunds, notifies clients and owners. Audit trail via `force_majeure_events` table. Booking status: `force_majeure_cancelled`. Admin endpoints: `POST /api/v1/admin/force-majeure`, `GET /api/v1/admin/force-majeure`
