---
name: File splitting refactoring progress
description: Status of ongoing large-file splitting refactoring in banya project
type: project
---

Split so far (all committed, build passes):
- mock_repos.go → 17 files
- interfaces.go → 8 files
- booking_service.go → 5 files (core + lifecycle + modification + checkin + admin)
- booking_handler.go → 4 files
- admin_handler.go → 4 files
- auth_handler.go → 4 files
- bot.go → 4 files
- payment_service.go → 4 files (core/initiate/hold/refund) [commit f0072e4]

**Remaining (plan: /Users/nikitaaldaev/.claude/plans/mighty-gliding-haven.md):**
1. internal/repository/postgres/bathhouse_repo.go (1001) → bathhouse_repo.go + _search.go + _scan.go
2. internal/service/bathhouse_service.go (994) → core + _lifecycle.go + _helpers.go
3. internal/service/analytics_service.go (825) → core + _admin.go + _advanced.go
4. internal/server/router.go (801) → router.go + routes_public.go + routes_api.go + routes_my.go + routes_admin.go

**Why:** "разбей все эти файлы, сделай проект читабельнее, улучши глобально файловую структуру"
**How to apply:** Continue from bathhouse_repo.go next session. Each step: create new files → trim original → go build ./...
