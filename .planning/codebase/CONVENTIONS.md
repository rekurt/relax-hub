# Coding Conventions

**Analysis Date:** 2026-04-17

## Naming Patterns

**Files:**
- Backend: lowercase snake_case (e.g., `booking_service.go`, `bathhouse_handler.go`)
- Frontend: lowercase with underscores for utilities, PascalCase for components (e.g., `Dashboard.tsx`, `auth.ts`)
- Test files: same name + `_test` suffix (e.g., `booking_service_test.go`, `Dashboard.test.tsx`)
- Mock files: `internal/repository/mock/` directory with entity name (e.g., `addon_repo.go`, `booking_modification_repo.go`)

**Functions:**
- Backend: camelCase (e.g., `NewBookingService`, `GetByID`, `Create`)
- Handler receivers: single letter (`h`, `w` for http.ResponseWriter) or meaningful short names
- Interface methods: verb-noun or verb-object pattern (e.g., `Create`, `GetByID`, `ListUserDisputes`)
- Private helper functions: camelCase, no export (e.g., `getIPHash`, `toBathhouseResponse`, `writeJSON`)

**Variables:**
- camelCase for all local and package variables
- UPPER_CASE for constants
- Single-letter for loop indexes (`i`, `j`, `k`)
- `ctx` for context parameter (always first parameter)
- `err` for error return values (always last return)
- Abbreviated receiver names: `r` (*http.Request), `w` (http.ResponseWriter), `m` (mock types)

**Types:**
- PascalCase for exported types (e.g., `Booking`, `BathhouseResponse`)
- lowercase for unexported response types (e.g., `bookingResponse`, `bathhouseResponse`)
- Struct field names: PascalCase, exported
- Constants: PascalCase (e.g., `BookingPending`, `WalletStatusActive`)

**Packages:**
- Lowercase, no underscores
- Clear purpose-driven names: `domain`, `service`, `handler`, `repository`, `middleware`, `logger`, `storage`, `payment`, `notification`, `antifraud`, `cron`, `geo`, `pms`, `fiscal`, `bot`, `app`, `admin`

## Code Style

**Formatting:**
- gofmt/goimports for Go (enforced by CI)
- ESLint with zero-warnings policy for TypeScript/React (CLI: `cd frontend && npm run lint`)
- 80-character soft limit for readability (not enforced, but observed)

**Linting:**
- Backend: golangci-lint v2 with `govet`, `staticcheck` (see `.golangci.yml` at repo root)
- Run: `make lint` (Go backend), `cd frontend && npm run lint` (React)
- CI enforces: zero lint/vet violations, `-race` flag on tests (see `.github/workflows/ci.yml`)

**Imports:**
- Go: Organized into groups: standard library, then external packages, then internal packages
  ```go
  import (
    "context"
    "errors"
    "net/http"
    
    "github.com/google/uuid"
    "github.com/rekurt/relax-hub/internal/domain"
  )
  ```
- TypeScript: Path aliases with `@/` prefix mapped to `src/` (in `tsconfig.json`)
  ```tsx
  import { useAuthStore } from '@/stores/auth'
  import type { InternalHandlerUserResponse } from '@/api/generated/model'
  ```

## Error Handling

**Patterns:**
- Domain errors defined as package-level vars in `internal/domain/errors.go` (e.g., `ErrNotFound`, `ErrAlreadyExists`)
- Service layer returns domain errors; handlers use `errors.Is(err, domain.ErrXxx)` for matching
- Handler layer maps domain errors to HTTP status codes via `handleServiceError()` in `internal/handler/response.go` (giant switch statement with 100+ cases)
- Wrapped errors: `PaymentFailedError` struct with `Code`, `MessageRU`, `Suggestion` fields for user-facing details
- Error wrapping: Use `errors.Unwrap()` for type assertions (e.g., `var pfe *domain.PaymentFailedError; errors.As(err, &pfe)`)

**Key mappings (from `internal/handler/response.go`):**
- 404: `ErrNotFound`, `ErrNotApproved`, `ErrSessionNotFound`, `ErrFAQNotFound`, `ErrWebhookNotFound`, etc.
- 400: `ErrInvalidInput`, `ErrBookingCancelLate`, `ErrPhoneInvalid`, `ErrOTPInvalid`, `ErrListingIncomplete`, etc.
- 409: `ErrAlreadyExists`, `ErrSlotUnavailable`, `ErrWalletConcurrentUpdate`, `ErrDisputeAlreadyExists`, etc.
- 401: `ErrUnauthorized`, `ErrSessionExpired`
- 403: `ErrForbidden`, `ErrUserBlocked`, `ErrWalletFrozen`, `ErrKYCNotApproved`, `ErrOfferNotAccepted`, etc.
- 429: `ErrOTPRateLimited`, `ErrOTPMaxAttempts`, `ErrResetRateLimited`, `ErrBroadcastRateLimit`

**Frontend (React):**
- Errors caught in try-catch blocks, displayed via RelaxHUB design-system message/notification components
- API client (axios instance in `src/api/axios-instance.ts`) handles 401 responses by calling `logout()`
- Type-safe error types via generated OpenAPI schema

## Logging

**Framework:** zap (wrapped in `internal/logger/`)

**Pattern:**
```go
log := logger.New(logger.LevelInfo) // or LevelWarn, LevelDebug
logger.Info("message", "key", value)  // also Error, Debug, Warn
```

**Injection:** Via Uber fx DI container. Logger passed to every service constructor.

**Configuration:**
- `BANI_LOGGER_LEVEL`: default `info` (options: `debug`, `info`, `warn`, `error`)
- `BANI_LOGGER_FORMAT`: `json` for production, `console` for development

**Structured logging:** Always pair keys with values (e.g., `logger.Info("booking created", "booking_id", bookingID, "user_id", userID)`)

## Comments

**When to Comment:**
- Complex business logic (e.g., pricing rules with discounts, booking status transitions)
- Non-obvious algorithmic choices
- Public API documentation (required for exported functions/types)
- Workarounds for bugs or limitations

**JSDoc/TSDoc:**
- Go: Regular comments above exported symbols (standard convention)
- TypeScript: JSDoc blocks with `/** ... */` for exported functions/types
  ```typescript
  /**
   * Конвертирует копейки в строку с рублями.
   * 15000 -> "150 ₽", 15050 -> "150,50 ₽"
   */
  export function formatPrice(kopecks: number): string
  ```

## Function Design

**Size:** Services generally 50-150 lines per method (e.g., `BookingService.Create` is ~80 lines). Complex services like `PaymentService` (~1200 lines total) are split by concern (payment processing, refunds, holds).

**Parameters:** 
- Backend: Context as first parameter always (Go convention), then domain types, then options if many
- Frontend: Component props typed as interfaces, destructured in function signature

**Return Values:**
- Backend: `(value, error)` pattern with named returns in complex functions
- Frontend: Promises for async operations, void for event handlers, union types for multiple returns

## Module Design

**Backend Exports:**
- Interfaces in `internal/service/` (e.g., `type BookingService interface { ... }`)
- Constructors exported (e.g., `func NewBookingService(...) BookingService { ... }`)
- Domain types exported from `internal/domain/`
- Implementation private (e.g., `type bookingService struct { ... }`)
- Repository interfaces centralized in single file: `internal/repository/interfaces.go`

**Frontend Structure:**
- `src/components/` — reusable, React Testing Library compatible
- `src/pages/` — route-level pages, role-based structure (`/admin/*`, `/client/*`, `/bathhouses/*`, etc.)
- `src/stores/` — Zustand stores, state persisted to localStorage
- `src/api/generated/` — auto-generated by orval, never edited manually
- `src/lib/` — utilities and helpers (e.g., `format.ts`, `constants.ts`)

**Barrel Exports:**
- Used selectively (e.g., component folders may export index.ts)
- Avoid circular dependencies

## Type Safety

**Backend (Go):**
- UUIDs: `google/uuid` package (e.g., `uuid.UUID`)
- Prices: `int64` in kopecks (not float, not rubles)
- Days of week: 0=Monday, 6=Sunday (intentionally different from Go's time.Weekday)
- Time strings: HH:MM format with string comparison (e.g., "09:00" < "14:30")
- Enums: String constants with validation

**Frontend (TypeScript):**
- Strict mode enabled (`strict: true` in `tsconfig.json`)
- Generated types from OpenAPI spec (orval) for all API responses
- Zustand store types exported as interfaces
- React component props typed as interfaces
- No `any` usage allowed (ESLint enforces via config)

## API Response Format

**Standard wrapper (all endpoints):**
```json
{
  "success": true|false,
  "data": {...},
  "error": {"code": "...", "message": "...", "suggestion": "..."},
  "meta": {"page": 1, "page_size": 20, "total_count": 100, "total_pages": 5}
}
```

**Helper functions in `internal/handler/response.go`:**
- `writeJSON(w, status, data)` — success response without pagination
- `writeJSONWithMeta(w, status, data, meta)` — success with pagination
- `writeError(w, status, code, message)` — error response
- `handleServiceError(w, err)` — maps domain errors to HTTP status codes
- `readJSON(w, r, v)` — parses JSON request body with size limit (1 MB)

## String Handling

**Encoding:**
- UTF-8 (Go default)
- Russian language support (e.g., day names in `src/lib/format.ts`, error messages in handlers)
- XSS safety: Template functions like `jsEscape` in admin panel templates (`internal/admin/pages/templates/`)

**Formatting:**
- Date format: `DD.MM.YYYY HH:mm` (dayjs with Russian locale)
- Price format: `"150 ₽"` for kopecks (via `formatPrice()` in `src/lib/format.ts`)
- Day of week: Full names (`Понедельник`) or short (`Пн`)

## Constants

**Backend locations:**
- Domain constants: `internal/domain/*.go` (e.g., `BookingPending`, `WalletStatusActive`)
- Config constants: `internal/config/` or `internal/domain/`
- Magic numbers: Named constants (e.g., `const maxBioLength = 1000`)
- Time durations: time.Duration type (e.g., `24 * time.Hour`)

**Frontend:**
- `src/lib/constants.ts` — app-wide (e.g., `AUTH_TOKEN_KEY`)
- Component-level: Defined in file (e.g., `DAYS_OF_WEEK` array)
- I18n: Hard-coded Russian strings (no translation system)

---

*Convention analysis: 2026-04-17*
