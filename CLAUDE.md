# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build & Dev Commands

Go 1.25.1. Module path `github.com/rekurt/relax-hub`. Two binaries: `cmd/server` (HTTP API + admin panel + Cobra CLI) and `cmd/bot` (Telegram bot).

```bash
# Build & run
make build                                       # build ./bin/bani-server from ./cmd/server
make run                                         # build + run server
go run ./cmd/server serve --with-admin           # serve with GoAdmin panel mounted
go run ./cmd/bot run                             # run telegram bot (shares BANI_* config)

# Test
make test                                        # go test ./... -v
go test ./internal/service/ -v -run TestBooking  # single test
make test-hurl                                   # tests/hurl/run_all_tests.sh (needs running server)
make frontend-test                               # cd frontend && npx vitest run

# Lint & vet
make lint                                        # golangci-lint v2 (govet, staticcheck enabled)
make vet                                         # go vet

# Database
make migrate-up                                  # apply migrations
make migrate-down                                # rollback last migration
make seed-admin                                  # interactive: create admin user
make seed-demo                                   # populate demo data

# Docker (override file remaps ports to avoid local clashes)
make docker-up        # app:28080, postgres:5435, redis:6381, minio:9102 (override defaults)
make docker-down
make docker-build

# OpenAPI / Swagger (commit regenerated docs/)
make swagger                                     # swag init from cmd/server/docs.go
make swagger-fmt                                 # format annotations

# Frontend (frontend/ dir, multi-role SPA)
make frontend-dev                                # vite dev server with HMR
make frontend-build                              # tsc + production build to dist/
make frontend-generate-api                       # orval regen client from docs/swagger.json
cd frontend && npm run lint                      # ESLint (zero-warnings policy)
```

CI (`.github/workflows/ci.yml`, push/PR to `main`): golangci-lint v2 → `go vet` → `go test -race` with coverage → `go build`.

## Repository Layout

```
cmd/server/         # main binary: serve, migrate, seed-admin, seed-demo CLI
cmd/bot/            # telegram bot binary (separate process)
config/             # Viper config + tests
internal/
  app/              # Uber fx DI container
  domain/           # models, errors, filters (no external deps)
  repository/       # interfaces.go + postgres/ (pgx) + mock/ (in-memory)
  service/          # business logic, AccessChecker for RBAC
  handler/          # HTTP handlers (one file per entity)
  middleware/       # auth (JWT), RBAC, CORS, rate limit, logging, panic, admin audit
  server/           # chi router setup
  notification/     # dispatcher, email, WS hub, telegram sender
  payment/          # PaymentProvider interface (YooKassa, bePaid)
  antifraud/        # rule engine (wallet/booking/payout/listing rules)
  cron/             # robfig/cron scheduler with Redis distributed locking
  geo/              # isochrone (OpenRouteService), transport (Yandex Maps)
  pms/              # external PMS adapters (Yclients, Restoplace)
  bot/              # Telegram bot logic
  fiscal/           # FiscalProvider (ATOL placeholder, no-op)
  seo/              # SSR prerender for crawler user-agents
  admin/            # GoAdmin engine, JWT bridge, custom pages
migrations/         # 230+ SQL migrations (PostGIS, pg_trgm, tsvector)
docs/               # generated OpenAPI (swagger.json/yaml committed for CI), API.md, DEPLOYMENT.md
tests/hurl/         # black-box HTTP contract tests
design/             # RelaxHUB Design System source (JSX/CSS) — populates frontend/src/components/design
frontend/           # React 18 + Vite 6 + TS 5.6 SPA
widget/             # standalone embeddable booking widget (own build.sh)
```

## Key Patterns

### DI with Uber fx

Every layer ships a `module.go` exposing `fx.Module`. Bind implementations to interfaces explicitly:

```go
fx.Annotate(postgres.NewUserRepo, fx.As(new(repository.UserRepository)))
```

### RBAC

Four roles: `client`, `owner`, `representative`, `admin`. Registration is client/owner only; admin via `seed-admin`; representative via owner invite.

```go
middleware.RequireAuth, OptionalAuth, RequireRole(roles...), RequireOwnerOrRepresentative()
middleware.GetUserID(ctx)   // uuid.UUID
middleware.GetUserRole(ctx) // domain.UserRole
service.AccessChecker.CanManageBathhouse(...)
```

Admin sub-roles: 6 granular roles (super_admin, moderator, support_l1/l2/l3, finance) with a 27-permission matrix and mandatory 2FA. Use `RequireAdminPermission(perm)` for endpoint gating.

### API Response Format

All endpoints return `{ success, data, error: { code, message }, meta: { page, page_size, total_count, total_pages } }`.

### Domain Errors

Source of truth: `internal/domain/errors.go` (definitions) and `internal/handler/response.go` (`mapError` → HTTP status). When adding a new domain error, update both files together — there is no separate registry to keep in sync.

### Logging

Zap, wrapped in `internal/logger/`. Injected via fx.

```go
logger.Info("message", "key", value)  // also Error, Debug, Warn
```

### Config

Viper with env prefix `BANI_`; nested keys join with `_` (`BANI_DATABASE_DSN`, `BANI_JWT_SECRET`). See **Configuration Reference** below for the full env var list.

### Database

PostgreSQL + PostGIS. Geo-search uses `ST_DWithin`/`ST_Distance` with `geography` type. Pool tuned via `BANI_DATABASE_MAX_CONNS` / `MIN_CONNS` / `MAX_CONN_LIFETIME`. Concurrent booking creation is serialized via `pg_advisory_xact_lock` inside the booking transaction to prevent double-book races.

### Testing

- Services: `repository/mock/` repos + table-driven tests
- Handlers: `httptest` + mock services
- No integration tests against real Postgres in `internal/`; use `tests/hurl/*.hurl` for HTTP contract checks
- Cross-module integration tests live in `tests/`

### OpenAPI / Swagger

swaggo annotations live in handler comments and `cmd/server/docs.go` (root spec). Swagger UI served at `/swagger/`. After editing annotations: `make swagger`. The generated `docs/` directory is committed (CI consumes it; frontend `orval` consumes `docs/swagger.json`).

## Critical Conventions

- **Prices in kopecks** (int64), never rubles
- **Day-of-week: 0=Monday, 6=Sunday** in pricing/schedule code (NOT Go's native `time.Weekday` where Sunday=0). When you serialize a `time.Weekday` to this convention, convert.
- **Time format**: `HH:MM` strings compared lexicographically (`"09:00" < "14:30"`)
- **Wraparound times**: when `TimeFrom > TimeTo`, the range spans midnight
- **UUIDs** (google/uuid) for all entity IDs
- **chi**: `chi.URLParam(r, "id")` for URL params

## Admin Panel (GoAdmin)

Enabled with `--with-admin` on serve. Config: `BANI_ADMIN_ENABLED` (default `false`), `BANI_ADMIN_PREFIX` (default `/admin-panel`), `BANI_ADMIN_LANGUAGE` (`ru`), `BANI_ADMIN_THEME` (`adminlte`).

Custom pages mounted under `{PREFIX}/pages/`: `/dashboard`, `/moderation` (+ `POST /moderation/api/{approve,reject,batch-approve,batch-reject}`), `/analytics` (+ `/analytics/export?type=...`), `/health`, `/finance`, `/static/*`. JWT from the main app bridges to GoAdmin sessions.

All custom pages share `base.tmpl` (sidebar/breadcrumbs/CSS) and `components.tmpl` (toast, status-badge, empty-state, loading-spinner, pagination). Template helpers: `statusRu`, `jsEscape`, `relativeTime`, `formatBytes`. Source in `internal/admin/pages/`.

Moderation keyboard shortcuts: `a` approve, `r` reject modal, `Esc` close, `Enter` confirm reject, `Ctrl+A` select all, `←/→` paginate.

## Frontend (frontend/)

Multi-role SPA: client (`/client/*`), owner/representative (`/`), admin (`/admin/*`). Three layouts (`AppLayout`, `ClientLayout`, `AdminLayout`) with separate navigation. `ProtectedRoute` accepts `allowedRoles`. Post-login redirect is role-based (client→`/client`, owner→`/`, admin→`/admin`).

**Stack**: Vite 6, React 18, TypeScript 5.6 strict, React Router 7, TanStack React Query 5, orval (API client), zustand (auth + selected bathhouse, persisted to localStorage), Vitest + RTL, dayjs.

**Dir conventions** (`frontend/src/`):
- `api/axios-instance.ts` — JWT interceptor; 401 logs out
- `api/generated/` — orval output, **never edit by hand**; regenerate via `npm run generate:api` after backend swagger changes
- `components/` — shared (layouts, cards, modals, ShareButton, OnboardingTour, BathhouseMap, etc.)
- `components/design/` — RelaxHUB Design System (sourced from repo-root `design/`)
- `pages/admin/`, `pages/client/`, plus owner pages at top level (`bathhouses/`, `bookings/`, `calendar/`, `chat/`, `crm/`, `pricing/`, `finance/`, `analytics/`, `promotion/`, `photos/`, `settings/`, ...)
- `stores/` — `auth.ts`, `bathhouse.ts`
- `lib/` — `format.ts` (kopecks→₽ via `formatPrice`), `useWebSocketNotifications`, `useDeviceToken`, constants
- `router.tsx` — role-grouped routes
- Path alias `@/` → `src/`

**Key patterns**:
- Dev proxy: `/api` → `localhost:8080`, `/ws` → `ws://localhost:8080`
- Multi-step wizards: `BathhouseForm` (7 steps), `BookingCreate` (4 steps) — design-system Steps with per-step draft persistence
- Combo payments: wallet+card split with visual slider in `BookingCreate`
- Web Share API w/ clipboard fallback via `ShareButton`; deep links resolved at `/share/booking/:token`
- Push permission requested only after first completed booking, not at registration
- Slug-based bathhouse URLs (`/bathhouses/:slug`); detail page loads via `getBySlug`
- OAuth (VK, Yandex, Google) on login/register

## Feature Subsystems

Every subsystem follows handler→service→repository. Below grouped by domain. Endpoint paths refer to `/api/v1/...`.

### Auth, identity, accounts
- **OAuth**: VK, Yandex, Google
- **Phone + OTP**: Redis-backed 6-digit codes (5 min TTL, 3 attempts), SMSProvider interface (SMS.ru adapter)
- **2FA**: TOTP (pquerna/otp) + SMS, partial-token flow during login
- **Sessions**: device/browser/IP tracking, 30-day expiry
- **Password reset**: Redis token via email link, terminates all sessions on use
- **Account deletion**: 30-day grace with restore, anonymization on execution, cron reminders day 0/14/27
- **Age verification**: required at registration; auto-creates wallet + welcome bonus (RU 500₽ / BY 15 BYN, 30-day expiry)
- **KYC**: `individual | sole_proprietor | self_employed | legal_entity` with submit/approve/reject + expiry
- **Offer & owner payment details**: versioned acceptance, required before listing creation
- **Region switching** (RU↔BY): blocked if non-zero balance, active bookings, open disputes, or unactivated certificates; archives old wallet, resets loyalty (`internal/service/region_service.go`)

### Bathhouses & listings
- **Listing draft wizard**: 7-step CRUD with submit
- **Onboarding gate**: `CreateBathhouse` requires KYC approval + offer + payment details
- **Listing completeness**: required/optional checklist; blocks moderation submission
- **Audit log**: JSONB diff per edit; substantial changes (address/city/photos) auto re-trigger moderation
- **Lifecycle**: deactivate / activate / archive (irreversible, no active bookings); search excludes inactive/archived
- **Duplication**: copy → draft, name + " (копия)", new slug/UUID
- **Add-ons**: per-bathhouse extras (per_item / per_hour / per_person), max 20; denormalized pricing in `booking_addons`
- **Photo verification**: pending/verified/rejected; `is_photo_verified` badge
- **Mass import**: CSV / xlsx (`excelize/v2`) → drafts (`POST /my/listings/import`)
- **Slug + SEO routes**, **OG/share links** (`POST /bookings/{id}/share` + `/share/booking/{token}`)

### Search, ranking, recommendations
- **Full-text**: tsvector (Russian config) + pg_trgm GIN, prefix support
- **Suggestions**: bathhouse names (trigram), city names, popular queries (Redis sorted set), 5-min cache
- **Ranking**: composite (relevance 0.30 + bayesian_rating 0.25 + conversion 0.20 + occupancy 0.15 + promo 0.10)
- **Bathhouse comparison**: 2–3 side-by-side
- **Saved searches**: JSONB filters, daily cron checks for new matches, max 50/user
- **Recently viewed**: Redis sorted set per user (last 20)
- **Recommendations**: collaborative filtering + preferences scoring
- **Isochrone search**: travel-time zones via OpenRouteService → GeoJSON → PostGIS `ST_Within` (`/isochrone`, 1h Redis cache)
- **Transport accessibility**: nearby metro/bus/parking via Yandex Maps Search (`/bathhouses/{id}/transport`, 7-day cache)
- **Geo heatmap** (admin): grid-cell aggregates of search queries vs. listings

### Bookings & calendar
- **Modes**: instant (default) and request (owner approval w/ timeout); statuses include `pending`, `pending_owner`, `confirmed`, `cancelled`, `rejected`, `completed`, `no_show`, `force_majeure_cancelled`
- **Buffer/lead time**: per-bathhouse buffer 0–120 min step 15, lead time 0–48h, advance 7–365 days
- **Check-in/check-out**: arrival window start-15 to start+30; no-show cron 30 min after start; no-show dispute within 2h with GPS validation (200m Haversine radius)
- **Modifications** (two-sided): client request → owner approve/reject; max 3/booking; 24h auto-reject; price diff charged or refunded (`PUT /bookings/{id}/modify`, `POST /my/bookings/{id}/modification/{approve,reject}`)
- **Extension** (two-sided): 1–2h, 30 min owner timeout, hold released on reject/expire
- **Re-booking**: `GET /bookings/{id}/rebook-data`
- **Cancellation policies**: flexible / moderate / strict — see `internal/domain/cancellation_policy.go`
- **Owner penalties**: cancellation = 10% client compensation; 4+/30d = warning, 6+ = auto-deactivate; response rate <30% for 60d → forced switch to instant or deactivation
- **iCal export**: `GET /my/bathhouses/{id}/calendar.ics`

### Pricing
- **Dynamic pricing**: prioritized rules with multipliers per hourly slot. Per-day pricing via `DaysOfWeek`. Long-session discounts, extra-guest surcharges, holiday multipliers (default 1.5×), last-minute discounts, seasonal tariffs (date range)
- **Smart pricing**: occupancy + area + demand → coefficient 0.8–1.5 (`GET /my/bathhouses/{id}/price-recommendation`)
- **Service fee**: platform fee on base price (not add-ons), region+category configurable, default 10% (admin `ServiceFeeConfig`)
- **Area average price** on bathhouse detail (city-level avg, 1h Redis cache)
- Booking response carries full price breakdown: `base_price`, `long_session_discount`, `extra_guest_surcharge`, `service_fee`, `holiday`, `seasonal_tariff`, last-minute fields

### Payments, wallet, payouts
- **Multi-region payments**: `PaymentProvider` factory selects by user region — RU: YooKassa, BY: bePaid (`internal/payment/provider_factory.go`)
- **Methods**: card, SBP, wallet, combo, MIR, Belkart, ERIP, Apple Pay, Google Pay (token-based)
- **Saved cards**: provider tokens with last4/brand/expiry for one-click repeat
- **Combo**: wallet first, card remainder, atomic rollback on card failure
- **Holds**: card auth holds for request-mode bookings (max 72h via `BANI_MAX_CARD_HOLD_HOURS`)
- **Refunds**: wallet refund + bonus (default 5%), proportional combo split, admin manual refund; payment retries 3× exponential (2s/4s/8s)
- **Wallet**: priority spend (expiring bonuses first), balance cap 100k RUB, top-up 500–30k RUB; bonus expiry cron 180d
- **Payouts**: SBP instant for RU owners w/ phone (YooKassa Payouts API) else bank transfer; daily/monthly limits + auto-payout threshold (hourly cron, `internal/cron/auto_payout.go`)
- **Escrow**: held after check-out, released after claim period (default 48h, `BANI_ESCROW_CLAIM_HOURS`); disputes block release; hourly auto-release cron
- **Security deposit**: 0–50% of base, card-held on creation, auto-released 48h after check-out unless dispute
- **Fiscalization**: `FiscalProvider` interface, ATOL placeholder + no-op; entity-aware tax (individual no VAT, SP/legal VAT 20%, self_employed NPD)
- **Bank reconciliation**: CSV / 1C-XML import, auto-match by amount + date (±1d) + reference; manual fallback (admin `/finance/...`)
- **Daily float reconciliation**: client wallets + owner wallets + escrow = expected; zero-tolerance discrepancy alerts

### Reviews & loyalty
- **Reviews multi-criteria**: cleanliness / accuracy / communication / value (0.5 step, overall = avg); Bayesian site-wide rating (m=5, cached)
- **Auto review request** cron (default 2h after check-out, `BANI_REVIEW_REQUEST_DELAY_HOURS`)
- **NLP auto-moderation**: regex profanity/spam, score >0.7 → `pending_moderation`
- **Quality monitoring**: rating <3.0 warns owner, <2.0 auto-depublishes
- **Bidirectional reviews**: owner rates client (punctuality, cleanliness, rule_compliance); double-blind reveal (both posted OR 14d, auto-reveal cron)
- **Review media**: 10 photos + 1 video per review; resize to 4 sizes (300/800/1200/1920) + WebP + blur-hash placeholder
- **Loyalty**: bronze/silver/gold/platinum by visit count; cashback 0/3/5/10% to wallet on completion (tag `cashback`)
- **Referrals**: personal codes, 500₽ bonus to both on referee's first completion (default)
- **Gift certificates**: `BANI-XXXX-XXXX`, partial redemption, 365d validity, purchasable without auth
- **Promo codes**: `percentage | fixed_amount | free_hour | free_addon` (the last zeroes a specific add-on via `target_addon_id`); bathhouse-scoped or global; usage/validity/min-amount checks

### Notifications & chat
- **Multi-channel**: in-app / email / telegram / push / SMS
- **Notification preferences**: per-event/channel; 28+ event types; mandatory events can't be disabled; fallback push→email(5min)→SMS(critical)
- **Chat**: WebSocket, conversations keyed by bathhouse+client; content filter warnings
- **Telegram bot**: separate binary, in-memory wizard state, short-ID cache for 64-byte callback limit
- **Booking reminders cron** (15 min): 24h push+email, 2h push, 5min owner reminder; Redis dedup
- **Bonus expiry notifications**: 14d + 3d push+email, deduped

### Owner CRM, promotion, analytics
- **CRM**: guest cards (auto-created on completion: visits, LTV, avg check, notes, tags, CSV export); dynamic segments (new/regular/lost/vip/birthday_soon); broadcasts to segments (rate limit 3/wk owner, 1/3d guest, respects prefs); auto-scenarios (thank/review/revisit/reactivate/birthday, hourly cron); response templates (max 50)
- **RFM analysis**: 1–5 each, custom multi-condition segments
- **Broadcast personalization**: `{{guest_name}}`, `{{last_visit_date}}`, `{{visit_count}}`, `{{promo_code}}`; click tracking
- **Subscriptions**: free / premium / promoted (premium +10 boost, promoted first)
- **Promoted campaigns**: auction (min 50₽/d), daily wallet deduction, auto-pause on budget exhaustion, weighted ranking boost
- **Owner webhooks**: 5 events (booking.{created,confirmed,cancelled,completed}, payment.received), HMAC-SHA256, async retry 3× expo, max 20/owner
- **PMS**: Yclients / Restoplace via `PMSProvider` (SyncBookings, SyncSchedule, PushBooking, PullBookings); 15-min bidir cron; encrypted creds
- **Owner analytics**: occupancy, income, conversion, competitor benchmarks
- **Financial reports**: wallet CSV/PDF export, owner acts PDF, 1C XML export for legal entities
- **Professional photography**: order from cabinet, requested→confirmed→completed→cancelled; pays from owner wallet on completion

### Support, disputes, anti-fraud
- **Support tickets**: category-based, auto-priority, 3-level escalation (24h L1→L2, 48h L2→L3, critical auto-L2); CSAT 24h post-resolution; auto-close 7d
- **FAQ bot (L1)**: trigram match before ticket creation; "not helpful" escalates to L2 (creates ticket)
- **Dispute system**: evidence collection 72h (photo/screenshot/gps/message/receipt) blocks escrow; resolution = full/partial/no refund + compensation; appeal within 7d
- **Anti-fraud**: rule engine — client (multi-card top-up, top-up/cancel cycles, dormant balance, rapid bookings), owner (self-booking, structuring, fake reviews), listing (duplicate detect by phone/email/payment, stoplist `antifraud_stoplist`); actions: block/flag/freeze_wallet
- **Complaints**: spam/offensive/fake/fraud/other on reviews/bathhouses/users; auto-hide review at 3+ reports
- **Force majeure** (admin): mass cancel by region+date with 100% wallet refund and audit (`force_majeure_events`)

### Admin tooling & platform config
- **Mass operations**: up to 1000 records, chunked tx (100/batch); succeeded/failed arrays
- **Admin notifications**: 8 types × 4 severity, daily critical-digest cron per role
- **Audit log**: admin actions
- **Platform settings**: typed key/value (int/float/string/bool/json), Redis cache 5min, prefix `platform:settings:` (12 seeded)
- **Feature flags**: optional region scoping (via `cities.region`), Redis cache 1min; `IsEnabled(key)` / `IsEnabledForRegion(key, region)` (13 seeded)
- **Moderation SLA**: queue size, avg wait, 48h SLA, per-moderator throughput; alerts at 24h
- **Conversion funnels, cohort analysis, supply/demand metrics, P&L unit economics, support ops (FCR/AHT)**
- **SSR/prerender**: bot-UA middleware serves pre-rendered HTML for SEO crawlers (Yandex, Google, social) with OG/Twitter/Schema.org meta; 1h Redis cache, invalidated on update

### Cron scheduler
Centralized robfig/cron in `internal/cron/` with Redis distributed locking, panic recovery, structured logging. 29+ jobs:
- **15 min**: booking timeout, no-show, reminders, modification request timeout
- **30 min**: extension request timeout
- **Hourly**: escrow release, review requests, auto-scenarios, anti-fraud patterns, auto-payout
- **Daily midnight Moscow**: bonus expiry, session cleanup, KYC check, owner response rate, promo deactivation, saved searches, metrics, ticket auto-close

Toggle: `BANI_CRON_ENABLED` (default `true`); timezone: `BANI_CRON_TIMEZONE` (default `Europe/Moscow`).

## Configuration Reference

All env vars use prefix `BANI_`. Highlights (defaults in parentheses):

**Environment & networking**
- `ENVIRONMENT` (`dev` / `production` — `production` enables stricter validation)
- `CORS_ALLOWED_ORIGINS` (`*`; wildcard rejected in production)
- `DATABASE_DSN`, `DATABASE_MAX_CONNS` (20), `DATABASE_MIN_CONNS` (2), `DATABASE_MAX_CONN_LIFETIME` (1h)
- `LOGGER_LEVEL` (`info`), `LOGGER_FORMAT` (`json` for prod, `console` for dev)
- `JWT_SECRET`

**Cron & schedules**
- `CRON_ENABLED` (`true`), `CRON_TIMEZONE` (`Europe/Moscow`)
- `REVIEW_REQUEST_DELAY_HOURS`

**Wallet, escrow, refunds**
- `WALLET_REFUND_BONUS_PERCENT` (5, range 0–15)
- `MAX_CARD_HOLD_HOURS` (72, YooKassa limit)
- `ESCROW_CLAIM_HOURS` (48, range 24–168)
- `WELCOME_BONUS_AMOUNT` (RU), `WELCOME_BONUS_AMOUNT_BY` (BY, default 1500 kopecks = 15 BYN), `WELCOME_BONUS_EXPIRY_DAYS`

**Payments**
- `PAYMENT_YOOKASSA_SHOP_ID`, `PAYMENT_YOOKASSA_SECRET_KEY`
- `PAYMENT_BEPAID_SHOP_ID`, `PAYMENT_BEPAID_SECRET_KEY`
- `PAYMENT_RETURN_URL`

**Fiscalization**
- `FISCAL_PROVIDER` (`none` / `atol`)
- `FISCAL_ATOL_LOGIN`, `FISCAL_ATOL_PASSWORD`, `FISCAL_ATOL_GROUP_CODE`

**OAuth & SMS**
- `OAUTH_{VK,YANDEX,GOOGLE}_{CLIENT_ID,CLIENT_SECRET,REDIRECT_URL}`
- `SMS_PROVIDER`, `SMS_API_KEY`

**Admin panel**
- `ADMIN_ENABLED` (`false`), `ADMIN_PREFIX` (`/admin-panel`), `ADMIN_LANGUAGE` (`ru`), `ADMIN_THEME` (`adminlte`)

**Docker port overrides** (set by `docker-compose.override.yml`)
- `DOCKER_APP_PORT` (28080), `DOCKER_POSTGRES_PORT` (5435), `DOCKER_REDIS_PORT` (6381), `DOCKER_MINIO_API_PORT` (9102)

## Commit & PR Conventions

- **Conventional Commits**: `feat:`, `fix:`, `refactor:`, `style:`, `ci:`, `chore:` (see git log for established style)
- Scope each commit to one logical change
- PRs should include behaviour summary, migration/config notes if any, test evidence (`make test` / `make test-hurl` output), and example req/res for API changes
- Auth, payments, RBAC, webhook changes warrant extra review — touching surfaces with security or financial impact
