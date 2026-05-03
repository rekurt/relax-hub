# Technology Stack

**Analysis Date:** 2026-04-17

## Languages

**Primary:**
- Go 1.25.1 - Backend API server and Telegram bot

**Secondary:**
- TypeScript 5.6 (strict mode) - React frontend SPA
- JavaScript/JSX - React components (compiled via TypeScript)

## Runtime

**Backend:**
- Go 1.25.1 runtime

**Frontend:**
- Node.js 18+ (for development and build)
- Browser runtime (React 18)

**Package Manager:**
- `go mod` for Go dependencies (managed via `go.mod`)
- npm 10+ for JavaScript dependencies (managed via `frontend/package.json`)
- Lockfile: `frontend/package-lock.json` (committed)

## Frameworks

**Backend (Go):**
- Chi v5.2.5 - HTTP router and middleware framework
- Uber fx v1.24.0 - Dependency injection container with module-based DI pattern
- golang-migrate v4.19.1 - Database schema migrations
- Go Admin v1.2.26 - Admin dashboard framework (optional, via `--with-admin` flag)
- cobra v1.10.2 - CLI command framework for server/bot CLI

**Frontend (React):**
- React 18.3.1 - UI library
- React Router 7.13.1 - Client-side routing (role-based route groups)
- Vite 6.0.5 - Build tool and dev server
- TypeScript 5.6.2 - Static type checking (strict mode enforced in CI)

**UI & State Management:**
- RelaxHUB Design System (`frontend/src/components/design`) - repo-owned component library adapted from `./design`, with Russian locale and native React controls
- Zustand 5.0.11 - Client state (auth token, selected bathhouse, persisted to localStorage)
- TanStack React Query 5.90.21 - Server state management, automatic caching and synchronization

## Key Dependencies

**Critical Backend:**
- jackc/pgx v5.8.0 - PostgreSQL driver with connection pooling (pgxpool)
- redis/go-redis v9.18.0 - Redis client
- golang-jwt/jwt v5.3.1 - JWT token generation and validation
- google/uuid v1.6.0 - UUID generation (all entity IDs)

**Backend - Storage & Media:**
- minio/minio-go v7.0.98 - S3-compatible object storage client (avatars, photos)
- disintegration/imaging v1.6.2 - Image resizing and processing
- chai2010/webp v1.4.0 - WebP format conversion
- buckket/go-blurhash v1.1.0 - Blur hash generation for placeholder images
- xuri/excelize v2.10.1 - XLSX file reading/writing for listing imports

**Backend - Payments:**
- rvinnie/yookassa-sdk-go v0.1.6 - YooKassa payment gateway for Russia
- (bePaid integration via custom HTTP client in `internal/payment/bepaid_provider.go`)

**Backend - Communication:**
- go-telegram-bot-api/telegram-bot-api v5.5.1 - Telegram bot
- SherClockHolmes/webpush-go v1.4.0 - Web Push notifications (VAPID-based)
- gorilla/websocket v1.5.3 - WebSocket for real-time chat and notifications

**Backend - Utilities:**
- pquerna/otp v1.5.0 - TOTP for two-factor authentication
- robfig/cron v3.0.1 - Scheduled jobs and background tasks
- go-chi/cors v1.2.2 - CORS middleware
- spf13/viper v1.21.0 - Configuration management (reads Viper with env prefix `BANI_`)
- spf13/cobra v1.10.2 - CLI framework
- joho/godotenv v1.5.1 - `.env` file loading
- golang.org/x/oauth2 v0.35.0 - OAuth2 for social login (VK, Yandex, Google)
- golang.org/x/crypto v0.48.0 - Password hashing and encryption
- go.uber.org/zap v1.26.0 - Structured logging framework
- swaggo/swag v1.16.6 - OpenAPI/Swagger documentation generation

**Frontend - Data/API:**
- axios 1.13.6 - HTTP client for API calls (configured with JWT interceptor)
- orval 7.13.2 - OpenAPI client code generator (auto-generates `src/api/generated/`)
- dayjs 1.11.20 - Date/time formatting used by design-system controls and business flows

**Testing Backend:**
- stretchr/testify v1.11.1 - Assertion library for Go tests
- alicebob/miniredis v2.37.0 - In-memory Redis mock for testing

**Testing Frontend:**
- vitest 4.1.0 - Unit test runner
- @testing-library/react 16.3.2 - React component testing utilities
- @testing-library/jest-dom 6.9.1 - Custom Jest matchers
- jsdom 28.1.0 - DOM implementation for Node.js

**Build/Development Frontend:**
- @vitejs/plugin-react 4.3.4 - Vite plugin for React Fast Refresh
- ESLint 9.39.4 - Linting with zero-warnings enforcement
- @typescript-eslint/eslint-plugin 8.57.0 - TypeScript ESLint rules
- @tanstack/react-query-devtools 5.91.3 - React Query debugging tools

## Configuration

**Environment:**
- Viper-based configuration with env prefix `BANI_`
- Environment variables take precedence over file-based config
- Development: `.env` file (git-ignored)
- Production template: `.env.production.example` (checked in)

**Key Configuration Variables:**
- `BANI_ENVIRONMENT` - `dev` or `production` (controls validation strictness)
- `BANI_SERVER_HOST` - Server bind address (default `0.0.0.0`)
- `BANI_SERVER_PORT` - Server port (default `8080`)
- `BANI_DATABASE_DSN` - PostgreSQL connection string
- `BANI_DATABASE_MAX_CONNS` - Connection pool size (default 20, max 100)
- `BANI_DATABASE_MIN_CONNS` - Min pool connections (default 2)
- `BANI_DATABASE_MAX_CONN_LIFETIME` - Connection reuse limit (default 1h)
- `BANI_REDIS_ADDR` - Redis address with port
- `BANI_REDIS_PASSWORD` - Redis auth password (optional)
- `BANI_JWT_SECRET` - JWT signing key (min 32 chars in production)
- `BANI_JWT_TOKEN_TTL` - Token expiry (default 24h)
- `BANI_LOGGER_LEVEL` - Log level: debug, info, warn, error (default info)
- `BANI_LOGGER_FORMAT` - json (production) or console (development)
- `BANI_CORS_ALLOWED_ORIGINS` - Comma-separated origins (wildcard `*` rejected in production)
- `BANI_STORAGE_ENDPOINT` - S3-compatible endpoint
- `BANI_STORAGE_BUCKET` - Bucket name
- `BANI_STORAGE_ACCESS_KEY` - S3 access key
- `BANI_STORAGE_SECRET_KEY` - S3 secret
- `BANI_STORAGE_REGION` - Region code
- `BANI_STORAGE_USE_SSL` - SSL requirement (must be true in production)
- `BANI_OAUTH_{VK,YANDEX,GOOGLE}_{CLIENT_ID,CLIENT_SECRET,REDIRECT_URL}` - OAuth provider configs
- `BANI_PAYMENT_YOOKASSA_{SHOP_ID,SECRET_KEY,PAYOUT_AGENT_ID,PAYOUT_SECRET_KEY}` - YooKassa config
- `BANI_PAYMENT_BEPAID_{SHOP_ID,SECRET_KEY}` - bePaid config
- `BANI_PAYMENT_RETURN_URL` - Post-payment redirect
- `BANI_PAYMENT_WALLET_REFUND_BONUS_PERCENT` - Wallet refund bonus (default 5%, range 0-15)
- `BANI_FISCAL_PROVIDER` - none, atol, or by
- `BANI_FISCAL_ATOL_{LOGIN,PASSWORD,GROUP_CODE}` - ATOL Online credentials
- `BANI_SMS_PROVIDER` - sms provider (smsru or noop)
- `BANI_SMS_API_KEY` - SMS.ru API key
- `BANI_TELEGRAM_BOT_TOKEN` - Telegram bot token
- `BANI_TELEGRAM_WEBHOOK_URL` - Webhook URL for Telegram updates
- `BANI_TELEGRAM_MODE` - polling or webhook
- `BANI_ADMIN_ENABLED` - Enable/disable admin panel (default false)
- `BANI_ADMIN_PREFIX` - Admin URL prefix (default /admin-panel)
- `BANI_ADMIN_LANGUAGE` - ru or en (default ru)
- `BANI_ADMIN_THEME` - adminlte (default)
- `BANI_WEBPUSH_VAPID_PUBLIC_KEY` - VAPID public key
- `BANI_WEBPUSH_VAPID_PRIVATE_KEY` - VAPID private key
- `BANI_WEBPUSH_VAPID_CONTACT` - VAPID contact email
- `BANI_ESCROW_CLAIM_HOURS` - Escrow hold period (default 48, range 24-168)
- `BANI_WELCOME_BONUS_AMOUNT` - RU welcome bonus in kopecks (default 50000 = 500 RUB)
- `BANI_WELCOME_BONUS_AMOUNT_BY` - BY welcome bonus in kopecks (default 1500 = 15 BYN)
- `BANI_WELCOME_BONUS_EXPIRY_DAYS` - Bonus expiry period (default 30)
- `BANI_REVIEW_REQUEST_DELAY_HOURS` - Hours after check-out to send review request (default 2)
- `BANI_WALLET_BONUS_EXPIRY_DAYS` - Bonus expiry from creation (default 180)
- `BANI_CRON_ENABLED` - Enable/disable background jobs (default true)
- `BANI_CRON_TIMEZONE` - Timezone for cron scheduling (default Europe/Moscow)
- `BANI_GEO_ISOCHRONE_API_URL` - OpenRouteService base URL
- `BANI_GEO_ISOCHRONE_API_KEY` - OpenRouteService API key
- `BANI_GEO_YANDEX_SEARCH_API_KEY` - Yandex Maps Search API key
- `BANI_EMAIL_HOST` - SMTP server hostname
- `BANI_EMAIL_PORT` - SMTP port
- `BANI_EMAIL_USERNAME` - SMTP username
- `BANI_EMAIL_PASSWORD` - SMTP password
- `BANI_EMAIL_FROM` - Sender email address

**Build:**
- Backend: `go build ./...` outputs to `./bin/`
- Frontend: Vite config at `frontend/vite.config.ts`
- TypeScript config at `frontend/tsconfig.json` (strict mode)
- ESLint config via `@eslint/js` flat config (zero warnings enforced in CI)

## Database

**Type:** PostgreSQL 17+ with PostGIS extension

**Connection:**
- Pooled via pgxpool (configurable size)
- Production: enforced SSL (sslmode=require)
- Development: optional SSL

**Features Used:**
- PostGIS for geographic queries (`ST_DWithin`, `ST_Distance` with `geography` type)
- Full-text search (tsvector, pg_trgm for fuzzy matching)
- JSON/JSONB columns for flexible data (discounts, metadata, audit diffs)
- Distributed advisory locks (`pg_advisory_xact_lock`) for race condition prevention
- Window functions for analytics

**Migrations:**
- golang-migrate v4 for versioned schema changes
- Migration files in `migrations/` directory
- Applied on startup via CI/CD and local setup

**Backup Storage:**
- MinIO (S3-compatible object storage)
- Stores avatars, bathhouse photos, review media, imports
- Two image size variants: medium (800px), large (1200px), plus blur hash

## Caching & Session Storage

**Redis 7+:**
- Session tokens
- Notifications queue
- Search suggestions
- Cached bathhouse ratings (Bayesian average)
- Device tokens (push subscriptions)
- Distributed cron locks
- Expense tracking (wallet holds, card authorizations)
- Recently viewed bathhouses (per user)
- Rate limit counters (OTP, password reset, broadcasts)
- API response cache (search results, transport data, isochrones)

## Container Orchestration

**Development:**
- `docker-compose.yml` - Full stack (app, postgres:17-postgis, redis:7, minio:latest)
- `docker-compose.deps.yml` - Dependencies only (postgres, redis, minio) for local development

**Production:**
- Containerized via Dockerfile (Go binary, not alpine due to CGO dependencies)
- Docker image: Go 1.25 base with migrations included

## Frontend Build & Serving

**Dev Server:**
- Vite dev server on `http://localhost:5173`
- Proxies `/api` → `http://localhost:8080` (backend)
- Proxies `/ws` → `ws://localhost:8080` (WebSocket)
- React Fast Refresh enabled

**Production Build:**
- `npm run build` produces optimized bundle in `dist/`
- Static files served by backend from `frontend/dist/`
- SPA routing via frontend router (all routes except `/api`, `/ws`, `/admin` go to index.html)

## API Documentation

**Swagger/OpenAPI:**
- Generated from inline swaggo comments in `internal/handler/*.go`
- Specification: `docs/swagger.json` (committed for CI)
- UI: Served at `/swagger/` endpoint
- Command to regenerate: `make swagger` (runs `swag init`)

## Module System

**Backend:**
- Go module path: `github.com/rekurt/relax-hub`
- fx modules per layer: `internal/{domain,repository,service,handler,middleware,cron,payment,notification}/module.go`
- Central DI container in `internal/app/app.go`

**Frontend:**
- Path alias: `@/` maps to `src/` for cleaner imports
- API client auto-generated to `src/api/generated/` (never edited manually)
- Component structure: `src/components/`, `src/pages/`, `src/stores/`

---

*Stack analysis: 2026-04-17*
