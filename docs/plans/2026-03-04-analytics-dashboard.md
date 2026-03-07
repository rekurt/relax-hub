# Аналитика для владельцев и админов

## Overview

Аналитический дашборд: для владельцев — просмотры, бронирования, конверсия, выручка, динамика рейтинга. Для админов — общая статистика платформы, топ-бани, активность пользователей, финансовые метрики. Агрегация по дням/неделям/месяцам с кешированием в Redis.

**Коммерческая ценность:** Аналитика — ключевой аргумент для Premium-подписки. Владельцы принимают решения на основе данных, повышая качество сервиса. Админская аналитика нужна для управления платформой и отчетности перед инвесторами.

## Context

- Files involved: internal/domain/, internal/repository/, internal/service/, internal/handler/, migrations/
- Related patterns: Clean architecture, Uber fx DI, Redis caching
- Dependencies: нет новых (Redis уже в проекте)
- Данные для аналитики: bookings, reviews, users, bathhouses — все уже есть

## Development Approach

- **Testing approach**: Regular (code first, then tests)
- Complete each task fully before moving to the next
- **CRITICAL: every task MUST include new/updated tests**
- **CRITICAL: all tests must pass before starting next task**

## Implementation Steps

### Task 1: Трекинг просмотров

**Files:**
- Create: `internal/domain/analytics.go`
- Create: `migrations/000022_analytics.up.sql`
- Create: `migrations/000022_analytics.down.sql`

- [x] Создать модель BathhouseView:
  ```
  BathhouseView {
    ID          uuid.UUID
    BathhouseID uuid.UUID
    ViewerID    *uuid.UUID      // nil для анонимных
    Source      string          // "search", "direct", "widget", "telegram"
    IPHash      string          // для подсчета уникальных без хранения IP
    ViewedAt    time.Time
  }
  ```
- [x] Создать модель AnalyticsSnapshot (агрегированные данные за период):
  ```
  AnalyticsSnapshot {
    BathhouseID   uuid.UUID
    Date          time.Time       // день
    Views         int64
    UniqueViews   int64
    Bookings      int64
    Revenue       int64           // в копейках
    ReviewCount   int
    AvgRating     float64
  }
  ```
- [x] Создать миграции
- [x] Написать тесты
- [x] Запустить go test ./... - все тесты должны пройти

### Task 2: Репозиторий аналитики

**Files:**
- Modify: `internal/repository/interfaces.go`
- Create: `internal/repository/postgres/analytics.go`
- Create: `internal/repository/mock/analytics.go`

- [x] AnalyticsRepository:
  - RecordView(ctx, view) — записать просмотр
  - GetBathhouseStats(ctx, bathhouseID, from, to) — агрегированная статистика за период
  - GetDailyStats(ctx, bathhouseID, from, to) — по дням
  - GetPlatformStats(ctx, from, to) — общая статистика платформы
  - GetTopBathhouses(ctx, metric, limit) — топ по метрике (views, bookings, revenue, rating)
  - CreateSnapshot(ctx, snapshot) — сохранить агрегированный снапшот
- [x] Реализовать postgres и mock
- [x] Написать тесты
- [x] Запустить go test ./... - все тесты должны пройти

### Task 3: Сервис аналитики

**Files:**
- Create: `internal/service/analytics_service.go`

- [ ] AnalyticsService:
  - RecordView(ctx, bathhouseID, viewerID, source) — записать просмотр (дедупликация по IP за 30 мин)
  - GetOwnerDashboard(ctx, bathhouseID, period) — дашборд владельца:
    - Просмотры (total, unique), бронирования, конверсия (bookings/views)
    - Выручка, средний чек
    - Рейтинг (текущий, динамика)
    - Сравнение с предыдущим периодом (%)
  - GetAdminDashboard(ctx, period) — дашборд админа:
    - Общие метрики: пользователи, бани, бронирования, выручка
    - Новые регистрации за период
    - Топ-10 бань по бронированиям/выручке
    - Активность (DAU/WAU/MAU)
  - AggregateDaily(ctx) — cron: агрегация за прошедший день в snapshots
- [ ] Кеширование в Redis: дашборды кешируются на 15 минут
- [ ] RBAC: owner — только свои бани, admin — все
- [ ] Написать unit-тесты
- [ ] Запустить go test ./... - все тесты должны пройти

### Task 4: Хендлеры аналитики

**Files:**
- Create: `internal/handler/analytics.go`
- Modify: `internal/server/router.go`
- Modify: `internal/handler/bathhouse.go`

- [ ] GET /api/v1/my/bathhouses/{id}/analytics?period=30d — дашборд владельца (owner/rep)
- [ ] GET /api/v1/my/bathhouses/{id}/analytics/daily?from=...&to=... — по дням (owner/rep)
- [ ] GET /api/v1/admin/analytics?period=30d — дашборд платформы (admin)
- [ ] GET /api/v1/admin/analytics/top?metric=bookings&limit=10 — топ бань (admin)
- [ ] При просмотре бани (GetByID) — записать view
- [ ] Зарегистрировать маршруты, добавить fx.Module
- [ ] Написать handler-тесты
- [ ] Запустить go test ./... - все тесты должны пройти

### Task 5: Cron-задачи

**Files:**
- Create: `internal/cron/analytics.go`

- [ ] Ежедневная агрегация в 02:00: подсчет views, bookings, revenue за прошлый день
- [ ] Еженедельная очистка старых записей BathhouseView (хранить 90 дней детальных)
- [ ] Интеграция с Uber fx lifecycle
- [ ] Написать тесты
- [ ] Запустить go test ./... - все тесты должны пройти

### Task 6: Верификация

- [ ] Запустить полный тест-сьют: go test ./... -v
- [ ] Запустить линтер: make lint
- [ ] Запустить go vet ./...
- [ ] Переместить этот план в docs/plans/completed/
