# Consolidated Plan: Cron Jobs + Support/Disputes + Antifraud/Audit — Test & Annotation Gaps

## Overview

All three subsystems (Cron Jobs, Support & Disputes, Anti-fraud & Audit) are fully implemented with passing tests. This plan covers the remaining gaps: missing handler tests for antifraud and audit log handlers, missing swagger annotations on antifraud endpoints, and final verification across all three subsystems.

## Context

- Files involved:
  - `internal/handler/antifraud_handler.go` — missing swagger on ListFlags, UpdateFlag; no test file
  - `internal/handler/audit_log_handler.go` — no test file
  - `internal/repository/mock/fraud_flag_repo.go` — needed for handler tests
  - `internal/repository/mock/audit_log.go` — needed for handler tests
- Related patterns: existing handler tests (ticket_handler_test.go, dispute_handler_test.go) use httptest + mock services
- Dependencies: none (all underlying code exists and passes tests)

## Development Approach

- **Testing approach**: Regular (code first, then tests)
- Complete each task fully before moving to the next
- **CRITICAL: every task MUST include new/updated tests**
- **CRITICAL: all tests must pass before starting next task**

## Implementation Steps

### Task 1: Add Swagger Annotations to Antifraud Handler

**Files:**
- Modify: `internal/handler/antifraud_handler.go`

- [x] Add swagger annotations to `ListFlags` method (@Summary, @Description, @Tags, @Produce, @Security, @Param for status/rule/user_id/page/page_size, @Success, @Failure, @Router /admin/antifraud/flags [get])
- [x] Add swagger annotations to `UpdateFlag` method (@Summary, @Description, @Tags, @Accept, @Produce, @Security, @Param id path, @Param body, @Success, @Failure, @Router /admin/antifraud/flags/{id} [patch])
- [x] Run `make swagger` to regenerate the spec
- [x] Run `go build ./...` — must pass

### Task 2: Handler Tests for AntiFraudHandler

**Files:**
- Create: `internal/handler/antifraud_handler_test.go`

- [x] Test ListFlags: default pagination, with status/rule/user_id filters, invalid status, invalid rule, invalid user_id
- [x] Test UpdateFlag: success, invalid flag ID, invalid status, flag not found
- [x] Test ListFilteredMessages: default pagination, custom page/page_size
- [x] Run `go test ./internal/handler/ -v -run TestAntiFraud` — must pass

### Task 3: Handler Tests for AuditLogHandler

**Files:**
- Create: `internal/handler/audit_log_handler_test.go`

- [ ] Test ListAuditLog: default pagination, with entity_type/entity_id/user_id/date filters
- [ ] Test ListAdminActions: default pagination, with admin_id/action/date filters
- [ ] Test GetEntityHistory: success, invalid entity ID
- [ ] Run `go test ./internal/handler/ -v -run TestAuditLog` — must pass

### Task 4: Final Verification

- [ ] Run full test suite: `go test ./... -v -race`
- [ ] Run linter: `make lint`
- [ ] Run build: `go build ./...`
- [ ] Verify swagger spec is up to date: `make swagger`

### Task 5: Update Documentation

- [ ] Move old plans (`2026-03-23-12-support-disputes.md`, `2026-03-23-13-antifraud-audit.md`, `2026-03-23-14-cron-jobs.md`) to `docs/plans/completed/`
- [ ] Move this plan to `docs/plans/completed/`
