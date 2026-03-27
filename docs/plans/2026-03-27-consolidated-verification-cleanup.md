---
# Consolidated Verification & Cleanup: Subsystems 7-11

## Overview
All five subsystems (Booking Enhancements, Pricing, Payment/Financial, Reviews/Rating, CRM for Owners) are fully implemented in the codebase. Plans 10 and 11 have unchecked task items despite complete implementation. This plan covers verification, test validation, and plan file cleanup.

## Context
- Files involved: all files listed in plans 7-11 (already exist and implemented)
- Related patterns: handler->service->repository, fx DI, cron jobs in internal/cron/
- Dependencies: none outstanding - all subsystems are wired up

## Development Approach
- **Testing approach**: Verification-only (run existing tests, lint, build)
- No new code to write - all features are implemented
- Focus on confirming everything compiles, passes tests, and lint is clean
- **CRITICAL: every task MUST include new/updated tests**
- **CRITICAL: all tests must pass before starting next task**

## Implementation Steps

### Task 1: Run full test suite to confirm all subsystems work

**Files:**
- All test files across internal/service/, internal/handler/, internal/cron/, internal/moderation/

- [x] run `go test ./... -v -race` and confirm all tests pass
- [x] run `make lint` and confirm no lint errors
- [x] run `go build ./...` and confirm clean build

### Task 2: Verify swagger annotations are up to date

**Files:**
- Modify: `docs/swagger.json` (regenerate if needed)

- [x] run `make swagger` to regenerate OpenAPI spec
- [x] verify new CRM endpoints (/my/crm/*) appear in swagger
- [x] verify review criteria fields appear in swagger models
- [x] run `make swagger-fmt`

### Task 3: Update plan files - mark completed tasks

**Files:**
- Modify: `docs/plans/2026-03-23-10-reviews-rating.md`
- Modify: `docs/plans/2026-03-23-11-crm-owners.md`

- [x] mark all tasks in plan 10 as [x] (all implemented)
- [x] mark all tasks in plan 11 as [x] (all implemented)

### Task 4: Archive all 5 completed plans

**Files:**
- Move: all 5 plan files to `docs/plans/completed/`

- [ ] move 2026-03-23-07-booking-enhancements.md to completed/
- [ ] move 2026-03-23-08-pricing-enhancements.md to completed/
- [ ] move 2026-03-23-09-payment-financial.md to completed/
- [ ] move 2026-03-23-10-reviews-rating.md to completed/
- [ ] move 2026-03-23-11-crm-owners.md to completed/

### Task 5: Verify acceptance criteria

- [ ] run full test suite: `go test ./... -v -race`
- [ ] run linter: `make lint`
- [ ] confirm all 5 plan files archived to docs/plans/completed/

### Task 6: Update documentation

- [ ] update CLAUDE.md if internal patterns changed
- [ ] move this plan to `docs/plans/completed/`
