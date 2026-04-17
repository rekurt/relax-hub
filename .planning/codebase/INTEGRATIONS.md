# External Integrations

**Analysis Date:** 2026-04-17

## APIs & External Services

**Payments & Fintech:**
- YooKassa (Russia)
  - Payment processing for card, SBP, wallet, combo payments
  - SDK: `github.com/rvinnie/yookassa-sdk-go` v0.1.6
  - Config: `BANI_PAYMENT_YOOKASSA_SHOP_ID`, `BANI_PAYMENT_YOOKASSA_SECRET_KEY`
  - Payout API: `BANI_PAYMENT_YOOKASSA_PAYOUT_AGENT_ID`, `BANI_PAYMENT_YOOKASSA_PAYOUT_SECRET_KEY` (SBP instant payouts)
  - Webhook: `POST /api/v1/webhooks/yookassa`
  - Methods: card, SBP, wallet, MIR (card networks)
  - Token-based payments for Apple Pay/Google Pay
  - Payment holds for booking requests (released on cancellation)

- bePaid (Belarus)
  - Alternative payment processor for BY region
  - Custom HTTP client integration in `internal/payment/bepaid_provider.go`
  - Config: `BANI_PAYMENT_BEPAID_SHOP_ID`, `BANI_PAYMENT_BEPAID_SECRET_KEY`
  - Webhook: `POST /api/v1/webhooks/bepaid`
  - Methods: Belkart, ERIP (Belarus-specific)
  - Status: production-ready

**Geolocation & Maps:**
- OpenRouteService (Isochrone API)
  - Travel-time based search zones ("15 min by car", "30 min by transit")
  - Client: HTTP client in `internal/geo/isochrone.go`
  - Config: `BANI_GEO_ISOCHRONE_API_URL`, `BANI_GEO_ISOCHRONE_API_KEY`
  - Cache: Redis (1h TTL per zone)
  - API: `GET /api/v1/isochrone?lat=&lon=&mode=car|transit&minutes=15`

- Yandex Maps Search API
  - Nearby transport discovery (metro, bus stops, parking)
  - Client: HTTP client in `internal/geo/transport.go`
  - Config: `BANI_GEO_YANDEX_SEARCH_API_KEY`
  - Cache: Redis (7-day TTL per bathhouse)
  - API: `GET /api/v1/bathhouses/{id}/transport`
  - Used for: transport accessibility feature on bathhouse detail

**SMS Delivery:**
- SMS.ru
  - SMS provider for OTP, 2FA, booking reminders, notifications
  - Client: `internal/sms/smsru.go`
  - Config: `BANI_SMS_PROVIDER=smsru`, `BANI_SMS_API_KEY`
  - No-op fallback: `internal/sms/noop.go` (development mode)
  - Used for: phone registration, 2FA codes, booking notifications

**Fiscalization (Tax Receipts):**
- ATOL Online (Russia)
  - Receipt generation for payments and refunds (best-effort)
  - Client: `internal/fiscal/atol.go`
  - Config: `BANI_FISCAL_PROVIDER=atol`, `BANI_FISCAL_ATOL_LOGIN`, `BANI_FISCAL_ATOL_PASSWORD`, `BANI_FISCAL_ATOL_GROUP_CODE`
  - Tax system adaptation: individual (no VAT), sole_proprietor/legal_entity (VAT 20%), self_employed (NPD)
  - No-op fallback: `internal/fiscal/noop.go`
  - BY provider: `internal/fiscal/by_provider.go` (logging for manual processing)

**Authentication & Social Login:**
- VK (VKontakte)
  - OAuth2 integration for Russian users
  - Client: `internal/auth/vk.go` (custom OAuth2 implementation)
  - Config: `BANI_OAUTH_VK_CLIENT_ID`, `BANI_OAUTH_VK_CLIENT_SECRET`, `BANI_OAUTH_VK_REDIRECT_URL`
  - SDK: `golang.org/x/oauth2`
  - Callback: `POST /api/v1/auth/oauth/vk/callback`

- Yandex
  - OAuth2 integration for Russian/CIS users
  - Client: `internal/auth/yandex.go`
  - Config: `BANI_OAUTH_YANDEX_CLIENT_ID`, `BANI_OAUTH_YANDEX_CLIENT_SECRET`, `BANI_OAUTH_YANDEX_REDIRECT_URL`
  - SDK: `golang.org/x/oauth2`
  - Callback: `POST /api/v1/auth/oauth/yandex/callback`

- Google
  - OAuth2 integration for international users
  - Config: `BANI_OAUTH_GOOGLE_CLIENT_ID`, `BANI_OAUTH_GOOGLE_CLIENT_SECRET`, `BANI_OAUTH_GOOGLE_REDIRECT_URL`
  - SDK: `golang.org/x/oauth2`
  - Callback: `POST /api/v1/auth/oauth/google/callback`

**Notifications:**
- Telegram Bot
  - Booking wizard, notifications, payments
  - Client: `github.com/go-telegram-bot-api/telegram-bot-api/v5` v5.5.1
  - Config: `BANI_TELEGRAM_BOT_TOKEN`, `BANI_TELEGRAM_MODE` (polling|webhook), `BANI_TELEGRAM_WEBHOOK_URL`
  - Short ID cache for callback data (max 64 bytes)
  - In-memory state machine for booking flow
  - Handler: `internal/bot/`

- Web Push (VAPID)
  - Browser push notifications via Service Worker
  - Client: `github.com/SherClockHolmes/webpush-go` v1.4.0
  - Config: `BANI_WEBPUSH_VAPID_PUBLIC_KEY`, `BANI_WEBPUSH_VAPID_PRIVATE_KEY`, `BANI_WEBPUSH_VAPID_CONTACT`
  - Subscription storage: PostgreSQL device_tokens table
  - Dispatcher: `internal/notification/dispatcher.go` (multi-channel: push, email, SMS, Telegram)
  - TTL: 24 hours per push message

**Email:**
- SMTP (configurable)
  - Email notifications for bookings, payments, account alerts
  - Client: Go `net/smtp` library with custom wrapper
  - Config: `BANI_EMAIL_HOST`, `BANI_EMAIL_PORT`, `BANI_EMAIL_USERNAME`, `BANI_EMAIL_PASSWORD`, `BANI_EMAIL_FROM`
  - Feature: async email dispatch via notification dispatcher
  - Fallback: email after push fails (5 min delay)

## Data Storage

**Databases:**
- PostgreSQL 17 with PostGIS extension
  - Primary data: users, bathhouses, bookings, payments, reviews, etc.
  - Connection: pgxpool via `internal/repository/postgres/`
  - Geographic queries: `ST_DWithin`, `ST_Distance`, `ST_Within`
  - Full-text search: tsvector with Russian language config
  - JSONB storage: discounts, audit diffs, offer acceptance
  - Advanced features: window functions, advisory locks for race prevention
  - Migrations: `golang-migrate/migrate` v4 via `migrations/` directory

**File Storage:**
- MinIO (S3-compatible)
  - Avatars, bathhouse photos, review media, imports
  - Client: `github.com/minio/minio-go/v7` v7.0.98
  - Config: `BANI_STORAGE_ENDPOINT`, `BANI_STORAGE_BUCKET`, `BANI_STORAGE_ACCESS_KEY`, `BANI_STORAGE_SECRET_KEY`, `BANI_STORAGE_REGION`, `BANI_STORAGE_USE_SSL`
  - Buckets: `bani-avatars` (dev), `bani` (docker), `bani-avatars-prod` (production)
  - Image processing: resize to 4 sizes (thumbnail 300px, medium 800px, large 1200px, full 1920px)
  - Format conversion: WebP with fallback to original
  - Metadata: blur hash placeholders for progressive image loading

**Caching & Sessions:**
- Redis 7+
  - Session tokens and device state
  - Notification queue (pending push/email/SMS)
  - Search suggestions (bathhouse names, cities, popular queries)
  - Bathhouse ratings cache (Bayesian average, 1h TTL)
  - Device tokens (push subscriptions)
  - Distributed cron locks (prevents duplicate job execution)
  - Wallet holds (temporary balance reservations)
  - Rate limit counters (OTP attempts, password reset, broadcasts)
  - Transport data cache (nearby metros/buses, 7d TTL)
  - Isochrone zones cache (travel time zones, 1h TTL)
  - Recently viewed bathhouses (per user, 20-item list)
  - Email bounce tracking (deduplication)

## Authentication & Identity

**Auth Provider:**
- Custom JWT-based authentication
  - Implementation: `internal/middleware/auth.go`, `internal/handler/auth_handler.go`
  - Token format: `{"user_id": "uuid", "role": "client|owner|representative|admin", "exp": unix_ts}`
  - TTL: `BANI_JWT_TOKEN_TTL` (default 24h)
  - Secret: `BANI_JWT_SECRET` (min 32 chars in production)
  - Middleware: `RequireAuth`, `OptionalAuth`, `RequireRole(roles...)`, `RequireOwnerOrRepresentative()`
  - Session tracking: device/browser/IP via `session` table
  - Session expiry: 30 days, auto-cleanup cron
  - Validation: token signature + expiry check on every request

- Two-Factor Authentication (2FA)
  - TOTP: Time-based One-Time Password (RFC 6238) via `github.com/pquerna/otp`
  - SMS 2FA: SMS codes via SMS.ru provider
  - Mandatory for all admin sub-roles
  - Partial token flow during 2FA verification

- Phone + OTP Authentication
  - 6-digit codes, 5 min TTL, 3 max attempts
  - Rate limiting: 3 requests per 15 min per phone
  - Redis-backed storage with atomic ops
  - SMS provider: SMS.ru or no-op

**Password Security:**
- Hashing: bcrypt via `golang.org/x/crypto/bcrypt`
- Reset flow: Redis token (1h TTL), email link
- Session termination on reset (all tokens revoked)

**Social Auth (OAuth2):**
- Providers: VK, Yandex, Google
- Implementation: `golang.org/x/oauth2` SDK
- Account linking: single account per social ID, prevent duplicate linking
- Role assignment: client by default on new registration

## Monitoring & Observability

**Error Tracking:**
- None detected - errors logged locally via zap

**Logs:**
- Structured logging via `github.com/go.uber.org/zap`
- Format: `json` (production), `console` (development)
- Level: configurable via `BANI_LOGGER_LEVEL` (debug|info|warn|error)
- Logger injected via fx DI to all services

**Metrics:**
- None detected (no Prometheus, StatsD, or similar)
- Manual analytics via SQL queries on `bookings`, `payments`, `users` tables

## CI/CD & Deployment

**Hosting:**
- Docker containers (app + postgres + redis + minio)
- Deployment target: self-hosted or cloud (Kubernetes, Docker Compose)

**CI Pipeline:**
- GitHub Actions (`.github/workflows/ci.yml`)
- Triggers: push to `main`, pull requests
- Steps:
  1. golangci-lint v2 (govet, staticcheck enabled)
  2. `go vet ./...`
  3. `go test ./... -race` with coverage
  4. `go build ./...`
- No integration tests for database (require real DB)
- Hurl tests in `tests/hurl/` for full API endpoint testing (manual via `make test-hurl`, requires running server)

**Deployment:**
- Makefile targets: `make docker-up`, `make docker-down`, `make run`
- Frontend: static files built via `npm run build` and served from backend
- Migrations: applied before server start in CI/CD

## Environment Configuration

**Required Environment Variables:**
- Database: `BANI_DATABASE_DSN`
- Redis: `BANI_REDIS_ADDR`
- JWT: `BANI_JWT_SECRET`

**Secrets Location:**
- Development: `.env` file (git-ignored)
- Production: environment variables (injected via CI/CD, Docker secrets, K8s ConfigMaps)
- Never in code or `.git/`

**Validation:**
- Strict in production (`BANI_ENVIRONMENT=production`):
  - JWT secret >= 32 characters
  - Database DSN requires `sslmode=require`
  - Storage credentials non-empty, SSL enforced
  - CORS origins no wildcard
- Dev mode (`BANI_ENVIRONMENT=dev`):
  - Relaxed validation
  - Allow `sslmode=disable`
  - Allow wildcard CORS

## Webhooks & Callbacks

**Incoming Webhooks:**
- YooKassa payment updates: `POST /api/v1/webhooks/yookassa`
- bePaid payment updates: `POST /api/v1/webhooks/bepaid`
- Telegram bot: `POST /api/v1/bot/webhook` (or polling mode)

**Outgoing Webhooks:**
- Owner webhooks: bathhouse events (booking.created, booking.confirmed, booking.cancelled, booking.completed, payment.received)
  - Delivery: async via goroutine, retry logic (3 attempts, exponential backoff)
  - Signature: HMAC-SHA256 in `X-Webhook-Signature` header
  - Storage: PostgreSQL `webhooks` and `webhook_deliveries` tables
  - Admin UI: owner can create/manage via `/api/v1/my/webhooks`

**Callbacks:**
- OAuth2 callbacks: `/api/v1/auth/oauth/{provider}/callback`
- Payment return: configurable via `BANI_PAYMENT_RETURN_URL`

## PMS & External System Integration

**PMS Connectors:**
- Yclients: booking sync, schedule pull/push
- Restoplace: booking sync, schedule pull/push
- Implementation: `internal/pms/` with `PMSProvider` interface
- Sync frequency: every 15 min via cron
- Credentials: encrypted storage in database
- Admin endpoints: `/api/v1/my/pms-connections/*`

**Format:**
- Imports: CSV, Excel (.xlsx) for batch listing creation
- Exports: CSV for guest cards, acts for owners, 1C XML for legal entities
- Endpoints: `POST /api/v1/my/listings/import`, `GET /api/v1/my/finance/acts`, `GET /api/v1/my/wallet/export`

## Feature-Specific Integrations

**Loyalty & Promotions:**
- Gift certificates: BANI-XXXX-XXXX format, 365-day validity
- Referral bonuses: configurable per-region (default 500 RUB)
- Promo codes: percentage/fixed_amount/free_hour/free_addon types
- Promotion campaigns: auction-based with daily wallet deduction

**Advanced Booking Features:**
- Booking modifications: two-sided approval (client request → owner approval)
- Extension requests: time extension with fund holds
- Concurrent booking control: PostgreSQL advisory locks prevent double-booking
- Cancellation policies: flexible/moderate/strict per bathhouse
- Buffer/lead time: configurable per bathhouse (buffer 0-120 min, lead 0-48h)

**Chat & Communication:**
- Real-time chat: WebSocket hub in `internal/notification/hub.go`
- Conversations: tied to bathhouse+client pair
- Content filtering: regex-based spam/profanity detection
- Message archival: stored in PostgreSQL

**Anti-Fraud:**
- Rule-based detection engine: `internal/antifraud/`
- Client rules: multi-card top-up, rapid bookings, suspicious patterns
- Owner rules: self-booking, structuring, fake reviews
- Listing creation rules: duplicate detection by phone/email/payment details
- Stoplist: `antifraud_stoplist` table
- Actions: block, flag (admin review), freeze_wallet
- Admin review queue: `/api/v1/admin/antifraud/flags`

---

*Integration audit: 2026-04-17*
