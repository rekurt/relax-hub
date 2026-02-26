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

### Testing

- Services tested with mock repositories from `repository/mock/`
- Handlers tested with httptest + mock services
- Middleware tested with httptest
- No integration tests (postgres repos require real DB)

### Config

Viper with env prefix `BANI_`. Nested keys use `_` separator:
`BANI_DATABASE_DSN`, `BANI_JWT_SECRET`, etc.

### Database

PostgreSQL with PostGIS for geo-queries. Migrations in `migrations/` folder.
Geo-search uses `ST_DWithin` and `ST_Distance` with `geography` type.

### Code Style

- Module path: `github.com/nikitaaldaev/bani`
- Standard Go project layout (cmd/, internal/, config/, migrations/)
- chi for routing with URL params via `chi.URLParam(r, "id")`
- UUID for entity IDs (google/uuid)
- Prices in kopecks (int64)
