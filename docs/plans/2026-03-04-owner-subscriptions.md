# Подписки для владельцев бань

## Overview

Трехуровневая модель подписок для владельцев бань: Free (базовое размещение), Premium (приоритет в выдаче, бейдж, расширенная статистика), Promoted (рекламный блок, таргетинг по городу, аналитика показов/кликов). Это основной источник дохода платформы от B2B-сегмента.

**Коммерческая ценность:** Рекуррентный доход от подписок. Premium ~2000-5000 руб/мес, Promoted — CPC/CPM модель. При 100 банях на Premium = 200-500k руб/мес recurring.

## Context

- Files involved: internal/domain/, internal/repository/, internal/service/, internal/handler/, internal/server/router.go, migrations/
- Related patterns: Clean architecture, Uber fx DI, chi router
- Dependencies: зависит от online-payments для автоматического списания
- Текущая модель Bathhouse имеет поля Status, Rating, и выдача сортируется по rating/price/distance

## Development Approach

- **Testing approach**: Regular (code first, then tests)
- Complete each task fully before moving to the next
- **CRITICAL: every task MUST include new/updated tests**
- **CRITICAL: all tests must pass before starting next task**

## Implementation Steps

### Task 1: Доменные модели подписок

**Files:**
- Create: `internal/domain/subscription.go`
- Modify: `internal/domain/errors.go`
- Create: `migrations/000005_subscriptions.up.sql`
- Create: `migrations/000005_subscriptions.down.sql`

- [x] Создать модель Subscription:
  ```
  Subscription {
    ID            uuid.UUID
    BathhouseID   uuid.UUID
    OwnerID       uuid.UUID
    Plan          SubscriptionPlan  // free, premium, promoted
    Status        SubscriptionStatus // active, expired, cancelled
    StartDate     time.Time
    EndDate       *time.Time        // nil для free
    AutoRenew     bool
    PriceKopecks  int64
    CreatedAt     time.Time
    UpdatedAt     time.Time
  }
  ```
- [x] Создать модель Promotion (для Promoted плана):
  ```
  Promotion {
    ID              uuid.UUID
    BathhouseID     uuid.UUID
    BudgetKopecks   int64
    SpentKopecks    int64
    StartDate       time.Time
    EndDate         time.Time
    TargetCityID    *int64
    Status          PromotionStatus  // active, paused, exhausted, expired
    ImpressionCount int64
    ClickCount      int64
    CreatedAt       time.Time
  }
  ```
- [x] Добавить domain-ошибки: ErrSubscriptionNotFound, ErrSubscriptionAlreadyActive, ErrPromotionBudgetExhausted
- [x] Создать миграцию с таблицами subscriptions и promotions
- [x] Написать тесты валидации моделей
- [x] Запустить go test ./... - все тесты должны пройти

### Task 2: Репозиторий подписок и промо

**Files:**
- Modify: `internal/repository/interfaces.go`
- Create: `internal/repository/postgres/subscription.go`
- Create: `internal/repository/postgres/promotion.go`
- Create: `internal/repository/mock/subscription.go`
- Create: `internal/repository/mock/promotion.go`

- [x] Добавить SubscriptionRepository interface:
  ```go
  type SubscriptionRepository interface {
    Create(ctx context.Context, sub *domain.Subscription) error
    GetByID(ctx context.Context, id uuid.UUID) (*domain.Subscription, error)
    GetActiveBybathhouse(ctx context.Context, bathhouseID uuid.UUID) (*domain.Subscription, error)
    Update(ctx context.Context, sub *domain.Subscription) error
    ListByOwner(ctx context.Context, ownerID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.Subscription], error)
    GetExpiring(ctx context.Context, before time.Time) ([]domain.Subscription, error)
  }
  ```
- [x] Добавить PromotionRepository interface
- [x] Реализовать postgres и mock репозитории
- [x] Написать тесты
- [x] Запустить go test ./... - все тесты должны пройти

### Task 3: Сервис подписок

**Files:**
- Create: `internal/service/subscription_service.go`

- [x] Создать SubscriptionService:
  - Subscribe(ctx, bathhouseID, plan) — создать подписку, списать оплату
  - Cancel(ctx, subscriptionID) — отменить автопродление
  - GetActive(ctx, bathhouseID) — текущая активная подписка
  - ListByOwner(ctx, ownerID, page, pageSize) — все подписки владельца
- [x] RBAC: только owner бани может управлять подписками
- [x] Написать unit-тесты
- [x] Запустить go test ./... - все тесты должны пройти

### Task 4: Влияние подписки на выдачу

**Files:**
- Modify: `internal/repository/postgres/bathhouse_repo.go`
- Modify: `internal/domain/filter.go`

- [x] При выдаче бань (List) учитывать подписку:
  - Premium: +10 к sort score
  - Promoted: отдельный блок "Рекомендованные" (первые N результатов, помеченные is_promoted)
- [x] Трекинг показов (impression) для promoted бань
- [x] Трекинг кликов (при GetByID) для promoted бань
- [x] Написать тесты
- [x] Запустить go test ./... - все тесты должны пройти

### Task 5: Хендлеры подписок

**Files:**
- Create: `internal/handler/subscription.go`
- Modify: `internal/server/router.go`

- [x] POST /api/v1/my/bathhouses/{id}/subscription - оформить подписку
- [x] GET /api/v1/my/bathhouses/{id}/subscription - текущая подписка
- [x] DELETE /api/v1/my/bathhouses/{id}/subscription - отменить автопродление
- [x] GET /api/v1/my/subscriptions - все подписки владельца
- [x] POST /api/v1/my/bathhouses/{id}/promotion - создать рекламную кампанию
- [x] GET /api/v1/my/bathhouses/{id}/promotion - статистика промо
- [x] Зарегистрировать маршруты, добавить fx.Module
- [x] Написать handler-тесты
- [x] Запустить go test ./... - все тесты должны пройти

### Task 6: Верификация

- [ ] Запустить полный тест-сьют: go test ./... -v
- [ ] Запустить линтер: make lint
- [ ] Запустить go vet ./...
- [ ] Обновить CLAUDE.md если появились новые паттерны
- [ ] Переместить этот план в docs/plans/completed/
