# GoAdmin Infrastructure Setup

## Overview

Base GoAdmin integration: dependency, config, cobra command, fx module, chi adapter, authentication adapter, basic layout.

## Context

- Files involved: `go.mod`, `internal/config/config.go`, `internal/admin/`, `internal/server/router.go`, `internal/app/app.go`, `cmd/server/serve.go`, `migrations/`
- Related patterns: Uber fx DI modules, cobra+viper CLI, chi router mounting
- Dependencies: `github.com/GoAdminGroup/go-admin`, `github.com/GoAdminGroup/themes/adminlte`, `github.com/GoAdminGroup/go-admin/adapter/chi`

## Development Approach

- **Testing approach**: Regular (code first, then tests)
- Complete each task fully before moving to the next
- GoAdmin integration is infrastructure-heavy; focus on correct wiring before UI
- **CRITICAL: every task MUST include new/updated tests**
- **CRITICAL: all tests must pass before starting next task**

## Implementation Steps

### Task 1: Add GoAdmin Dependencies

**Files:**
- Modify: `go.mod`

- [x] go get github.com/GoAdminGroup/go-admin
- [x] go get github.com/GoAdminGroup/themes/adminlte (or sword)
- [x] go get github.com/GoAdminGroup/go-admin/adapter/chi
- [x] go get github.com/GoAdminGroup/go-admin/engine
- [x] go mod tidy

### Task 2: GoAdmin Config Structure

**Files:**
- Modify: `internal/config/config.go`

- [ ] Add Admin section to Config struct: Enabled bool, Prefix string ("/admin-panel"), Language string ("ru"), Theme string
- [ ] Add BANI_ADMIN_ENABLED, BANI_ADMIN_PREFIX, BANI_ADMIN_LANGUAGE, BANI_ADMIN_THEME env vars
- [ ] Default: Enabled=false, Prefix="/admin-panel", Language="ru", Theme="adminlte"

### Task 3: GoAdmin Engine Module

**Files:**
- Create: `internal/admin/module.go`
- Create: `internal/admin/engine.go`
- Create: `internal/admin/auth.go`

- [ ] Create fx.Module for admin panel
- [ ] Initialize GoAdmin engine with chi adapter
- [ ] Configure GoAdmin connection to existing PostgreSQL (reuse BANI_DATABASE_DSN)
- [ ] Create custom auth adapter that bridges existing JWT auth (domain.UserRole admin) to GoAdmin session
- [ ] Configure GoAdmin file upload to use existing S3/MinIO storage config
- [ ] Set Russian language, timezone, logo, mini-logo, footer

### Task 4: Mount Admin Router

**Files:**
- Modify: `internal/server/router.go`
- Modify: `internal/app/app.go`

- [ ] Conditionally mount GoAdmin engine routes under config.Admin.Prefix when Admin.Enabled=true
- [ ] Add admin fx.Module to app.go (conditional on config)
- [ ] Ensure GoAdmin static assets are served correctly

### Task 5: Cobra Flag Integration

**Files:**
- Modify: `cmd/server/serve.go`

- [ ] Add --with-admin flag to serve command
- [ ] When flag is set, override BANI_ADMIN_ENABLED=true
- [ ] Log admin panel URL on startup when enabled

### Task 6: GoAdmin Migration

**Files:**
- Create: `migrations/000025_goadmin_tables.up.sql`
- Create: `migrations/000025_goadmin_tables.down.sql`

- [ ] Create GoAdmin system tables (goadmin_users, goadmin_roles, goadmin_permissions, goadmin_role_users, goadmin_user_permissions, goadmin_menu, goadmin_operation_log, goadmin_session)
- [ ] Seed default admin role and menu structure
- [ ] Bridge: create GoAdmin admin user linked to existing admin users in users table

### Task 7: Verify Infrastructure

- [ ] Build project: go build ./...
- [ ] Run migrations: make migrate-up
- [ ] Start server with --with-admin flag
- [ ] Access /admin-panel/ in browser - login page renders
- [ ] Login with admin credentials
- [ ] Run tests: go test ./... -v
- [ ] Run linter: make lint
