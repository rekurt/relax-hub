# Repository Guidelines

## Project Structure & Module Organization
`cmd/server` contains the main CLI entrypoints: `serve`, `migrate`, and `seed-admin`; `cmd/bot` runs the Telegram bot. Core backend code lives in `internal/`: `handler` for HTTP endpoints, `service` for business logic, `repository/postgres` for persistence, `middleware` for auth/RBAC, and `domain` for entities and validation. Configuration is in `config/`, SQL migrations in `migrations/`, API and deployment docs in `docs/`, and black-box API checks in `tests/hurl/`. The booking widget is a separate frontend artifact in `widget/`.

## Build, Test, and Development Commands
Use the `Makefile` as the default entrypoint:

- `make build` builds `./bin/bani-server` from `./cmd/server`
- `make run` builds and starts the API locally
- `make test` runs all Go tests with verbose output
- `make test-hurl` runs Hurl API scenarios from `tests/hurl/`
- `make migrate-up` / `make migrate-down` applies or rolls back DB migrations
- `make docker-up` / `make docker-down` starts or stops PostgreSQL, Redis, and the app via Docker Compose
- `make seed-admin` creates the initial admin user interactively

## Coding Style & Naming Conventions
Follow idiomatic Go. Keep packages focused and dependency direction layered: handlers call services, services call repositories. Use tabs for Go indentation, lowercase package names, `camelCase` for unexported symbols, and `PascalCase` for exported types and functions. Prefer table-driven tests in `*_test.go`. Run `gofmt` on edited files; run `golangci-lint run ./...` when touching shared logic.

## Testing Guidelines
Place unit tests next to the code they verify, for example `internal/service/payment_service_test.go`. Use `tests/` for cross-module and integration coverage, and `tests/hurl/*.hurl` for HTTP contract checks. Add tests for new handlers, services, and repository behavior; changes to migrations or public API should include either Go integration coverage or Hurl scenarios.

## Commit & Pull Request Guidelines
Recent history uses Conventional Commits such as `feat: ...` and `fix: ...`; keep that format and scope messages to one logical change. Pull requests should include a short description of behavior changes, migration/config notes if applicable, test evidence (`make test`, `make test-hurl`, or focused commands), and example requests/responses for API changes.

## Security & Configuration Tips
Do not commit real secrets from `.env`; use `.env.example` and `config/config.yaml` as templates. Validate changes touching auth, payments, RBAC, and webhooks carefully, since this repository includes JWT auth, YooKassa integration, Redis, and admin surfaces.
