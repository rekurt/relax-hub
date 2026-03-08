# Система жалоб и репортов

## Overview

Жалобы на отзывы и бани: спам, нецензурная лексика, фейковые отзывы, мошенничество. Очередь модерации для админа с фильтрами и массовыми действиями. Автоматические действия при множественных жалобах (автоскрытие). Уведомления о решении жалобы.

**Коммерческая ценность:** Доверие к платформе. Чистый каталог без спама и фейков повышает конверсию. Снижение негативного UGC защищает репутацию владельцев бань и платформы.

## Context

- Files involved: internal/domain/, internal/repository/, internal/service/, internal/handler/, migrations/
- Related patterns: Clean architecture, Uber fx DI
- Dependencies: нет новых
- Уже существует: ReviewStatus (pending/approved/rejected/hidden), Admin handler

## Development Approach

- **Testing approach**: Regular (code first, then tests)
- Complete each task fully before moving to the next
- **CRITICAL: every task MUST include new/updated tests**
- **CRITICAL: all tests must pass before starting next task**

## Implementation Steps

### Task 1: Доменные модели жалоб

**Files:**
- Create: `internal/domain/complaint.go`
- Modify: `internal/domain/errors.go`
- Create: `migrations/000020_complaints.up.sql`
- Create: `migrations/000020_complaints.down.sql`

- [x] Создать модель Complaint:
  ```
  Complaint {
    ID            uuid.UUID
    ReporterID    uuid.UUID
    TargetType    ComplaintTargetType  // "review", "bathhouse", "user"
    TargetID      uuid.UUID
    Reason        ComplaintReason      // spam, offensive, fake, fraud, other
    Description   string
    Status        ComplaintStatus      // pending, resolved, dismissed
    ResolvedByID  *uuid.UUID
    Resolution    string               // описание решения
    ResolvedAt    *time.Time
    CreatedAt     time.Time
  }
  ```
- [x] Добавить domain-ошибки: ErrComplaintNotFound, ErrAlreadyReported (один пользователь — одна жалоба на объект)
- [x] Создать миграцию с UNIQUE(reporter_id, target_type, target_id)
- [x] Написать тесты
- [x] Запустить go test ./... - все тесты должны пройти

### Task 2: Репозиторий жалоб

**Files:**
- Modify: `internal/repository/interfaces.go`
- Create: `internal/repository/postgres/complaint.go`
- Create: `internal/repository/mock/complaint.go`

- [x] ComplaintRepository: Create, GetByID, List (paginated, с фильтрами), UpdateStatus, CountByTarget, CheckExists
- [x] Фильтры: status, target_type, reason, date range
- [x] Реализовать postgres и mock
- [x] Написать тесты
- [x] Запустить go test ./... - все тесты должны пройти

### Task 3: Сервис жалоб

**Files:**
- Create: `internal/service/complaint_service.go`

- [ ] ComplaintService:
  - Report(ctx, reporterID, targetType, targetID, reason, description) — подать жалобу
  - Resolve(ctx, complaintID, adminID, resolution) — решить жалобу
  - Dismiss(ctx, complaintID, adminID) — отклонить жалобу
  - List(ctx, filter, page, pageSize) — список жалоб (admin)
  - GetByID(ctx, id) — деталь жалобы
- [ ] Автоматические действия:
  - 3+ жалоб на отзыв — автоскрытие (status = hidden)
  - 5+ жалоб на баню — уведомление админу, пометка для проверки
- [ ] RBAC: подать жалобу — любой auth user, управление — только admin
- [ ] Написать unit-тесты
- [ ] Запустить go test ./... - все тесты должны пройти

### Task 4: Хендлеры жалоб

**Files:**
- Create: `internal/handler/complaint.go`
- Modify: `internal/server/router.go`

- [ ] POST /api/v1/reviews/{id}/report — пожаловаться на отзыв (auth)
- [ ] POST /api/v1/bathhouses/{id}/report — пожаловаться на баню (auth)
- [ ] POST /api/v1/users/{id}/report — пожаловаться на пользователя (auth)
- [ ] GET /api/v1/admin/complaints — список жалоб с фильтрами (admin)
- [ ] GET /api/v1/admin/complaints/{id} — детали жалобы (admin)
- [ ] PATCH /api/v1/admin/complaints/{id}/resolve — решить (admin)
- [ ] PATCH /api/v1/admin/complaints/{id}/dismiss — отклонить (admin)
- [ ] Зарегистрировать маршруты, добавить fx.Module
- [ ] Написать handler-тесты
- [ ] Запустить go test ./... - все тесты должны пройти

### Task 5: Верификация

- [ ] Запустить полный тест-сьют: go test ./... -v
- [ ] Запустить линтер: make lint
- [ ] Запустить go vet ./...
- [ ] Переместить этот план в docs/plans/completed/
