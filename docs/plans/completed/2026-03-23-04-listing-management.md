---
# Subsystem 4: Listing Management Enhancements (FR-025, FR-029-032)

## Overview
Completeness checklist for listings, audit log for edits with auto-re-moderation on substantial changes, listing duplication, and temporary deactivation / full archival.

## Context
- Existing bathhouse CRUD: `internal/service/bathhouse_service.go`, `internal/handler/bathhouse_handler.go`
- Existing moderation: bathhouse has Status field (pending/active/rejected)
- Existing photo management: `internal/handler/photo_handler.go`
- Bathhouse model: `internal/domain/bathhouse.go`
- Existing booking check before deletion: ErrBathhouseHasBookings
- AccessChecker: `internal/service/access_checker.go`

## Dependencies
- No dependencies on other new subsystems
- Uses existing bathhouse, booking, and photo subsystems

## Development Approach
- **Testing approach**: Regular (code first, then tests)
- **CRITICAL: every task MUST include new/updated tests**
- **CRITICAL: all tests must pass before starting next task**

## Implementation Steps

### Task 4.1: Listing Completeness Check (FR-025)

**Files:**
- Modify: `internal/service/bathhouse_service.go`
- Modify: `internal/handler/bathhouse_handler.go`
- Modify: `internal/server/router.go`

- [x] Add CheckCompleteness method to BathhouseService:
  ```go
  type CompletenessItem struct {
      Field    string `json:"field"`
      Label    string `json:"label"`
      Complete bool   `json:"complete"`
      Required bool   `json:"required"`
  }
  type CompletenessResult struct {
      Items      []CompletenessItem `json:"items"`
      Score      int                `json:"score"`       // percentage 0-100
      CanSubmit  bool               `json:"can_submit"`  // all required items complete
  }
  ```
- [x] Required checks:
  - Name set (non-empty)
  - Description set (min 50 chars)
  - Type set
  - Address set + geocoded (lat/lng not zero)
  - City assigned
  - At least 3 verified photos
  - Base price > 0
  - Schedule set (at least 1 day with hours)
  - Capacity set (max_guests > 0)
- [x] Optional checks (improve score):
  - Amenities listed (>= 3)
  - Contact phone set
  - Cancellation policy set
- [x] GET /api/v1/my/bathhouses/{id}/completeness — get checklist status (RequireAuth, owner/rep)
- [x] Block moderation submission if required items incomplete
- [x] Write tests
- [x] Run `go test ./... -v` — must pass

### Task 4.2: Audit Log for Edits (FR-029)

**Files:**
- Create: `internal/domain/audit_log.go`
- Create: `internal/repository/postgres/audit_log_repo.go`
- Create: `internal/repository/mock/audit_log_repo.go`
- Create: `internal/service/audit_log_service.go`
- Modify: `internal/repository/interfaces.go`
- Modify: `internal/service/bathhouse_service.go`
- Create: `migrations/XXXXXX_audit_log.up.sql`

- [x] AuditLog model:
  - ID, EntityType (string: "bathhouse", "booking", etc.), EntityID (uuid)
  - UserID (uuid), Action (string: "create", "update", "delete", "status_change")
  - ChangedFields (JSONB: map[string]{ Old interface{}, New interface{} })
  - CreatedAt
- [x] AuditLogRepository: Create, ListByEntity(entityType, entityID, pagination), ListByUser(userID, pagination)
- [x] AuditLogService:
  - LogChange(ctx, entityType, entityID, userID, action, changes)
  - GetHistory(ctx, entityType, entityID, pagination)
  - IsSubstantialChange(changes) — returns true if address changed, type changed, or >50% photos replaced
- [x] Hook into bathhouse UpdateBathhouse service method:
  - Compare old vs new fields, log diff
  - If substantial change detected → auto-set status to pending (re-moderation)
  - Notify owner: "Your listing requires re-moderation due to substantial changes"
- [x] GET /api/v1/admin/audit-log — admin view with filters (entity_type, entity_id, user_id, date range)
- [x] GET /api/v1/my/bathhouses/{id}/history — owner view of their listing's audit log
- [x] Write tests
- [x] Run `go test ./... -v` — must pass

### Task 4.3: Listing Duplication (FR-032)

**Files:**
- Modify: `internal/service/bathhouse_service.go`
- Modify: `internal/handler/bathhouse_handler.go`

- [x] POST /api/v1/my/bathhouses/{id}/duplicate — create copy of listing (RequireAuth, owner/rep)
- [x] DuplicateBathhouse in service:
  - Load source bathhouse, verify ownership
  - Create new bathhouse with copied fields: name + " (копия)", description, type, address, city, amenities, capacity, pricing, schedule, booking settings
  - Do NOT copy: ID, slug (generate new), status (set to draft), created_at
  - Photos: reference same file URLs (no re-upload needed)
  - New listing starts as draft, goes through moderation as new
- [x] Return new bathhouse ID
- [x] Write tests
- [x] Run `go test ./... -v` — must pass

### Task 4.4: Temporary Deactivation & Full Archival (FR-030, FR-031)

**Files:**
- Modify: `internal/domain/bathhouse.go`
- Modify: `internal/service/bathhouse_service.go`
- Modify: `internal/handler/bathhouse_handler.go`

- [x] Add new bathhouse statuses: "inactive" (temporarily hidden), "archived" (permanently hidden)
- [x] POST /api/v1/my/bathhouses/{id}/deactivate — hide from search (RequireAuth, owner/rep)
  - Set status to "inactive"
  - Keep existing confirmed bookings (they proceed normally)
  - Listing not visible in search results
  - Owner can still manage existing bookings
- [x] POST /api/v1/my/bathhouses/{id}/activate — restore visibility (RequireAuth, owner/rep)
  - Set status back to "active" (only if was previously active, not if pending/rejected)
  - If was "inactive", restore to "active"
- [x] DELETE /api/v1/my/bathhouses/{id}/archive — full archive (RequireAuth, owner/rep)
  - Check for active/confirmed bookings → if any, return ErrBathhouseHasBookings
  - Set status to "archived"
  - Not reversible via API (admin can restore)
- [x] Modify search queries to exclude "inactive" and "archived" listings
- [x] Write tests
- [x] Run `go test ./... -v -race` — must pass
- [x] Run linter: `make lint`
