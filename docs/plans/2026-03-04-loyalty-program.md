# Программа лояльности

## Overview

Накопительная система баллов за бронирования. Уровни: Бронза/Серебро/Золото/Платина с прогрессирующими привилегиями. Списание баллов при оплате бронирований. Баллы начисляются после завершения визита.

**Коммерческая ценность:** Увеличивает LTV (Lifetime Value) пользователя в 2-3 раза. Повышает частоту повторных визитов. Создает switching cost — пользователям невыгодно уходить к конкурентам.

## Context

- Files involved: internal/domain/, internal/repository/, internal/service/, internal/handler/, migrations/
- Related patterns: Clean architecture, Uber fx DI
- Dependencies: нет новых
- Связь: начисление при completed booking, списание при создании нового booking

## Development Approach

- **Testing approach**: Regular (code first, then tests)
- Complete each task fully before moving to the next
- **CRITICAL: every task MUST include new/updated tests**
- **CRITICAL: all tests must pass before starting next task**

## Implementation Steps

### Task 1: Доменные модели лояльности

**Files:**
- Create: `internal/domain/loyalty.go`
- Modify: `internal/domain/errors.go`
- Create: `migrations/000012_loyalty.up.sql`
- Create: `migrations/000012_loyalty.down.sql`

- [x] Создать модель LoyaltyAccount:
  ```
  LoyaltyAccount {
    UserID        uuid.UUID
    Level         LoyaltyLevel    // bronze, silver, gold, platinum
    Points        int64           // текущий баланс баллов
    TotalEarned   int64           // всего заработано
    TotalSpent    int64           // всего потрачено
    VisitCount    int             // количество завершенных визитов
    UpdatedAt     time.Time
    CreatedAt     time.Time
  }
  ```
- [x] Уровни и пороги:
  - Bronze: 0+ визитов (1 балл = 1 рубль, скидка 0%)
  - Silver: 5+ визитов (1.2 балла = 1 рубль, скидка 3%)
  - Gold: 15+ визитов (1.5 балла = 1 рубль, скидка 5%)
  - Platinum: 30+ визитов (2 балла = 1 рубль, скидка 10%)
- [x] Создать модель LoyaltyTransaction (ID, UserID, Type earn/spend, Amount, BookingID, Description, CreatedAt)
- [x] Добавить domain-ошибку: ErrInsufficientPoints
- [x] Создать миграции
- [x] Написать тесты
- [x] Запустить go test ./... - все тесты должны пройти

### Task 2: Репозиторий лояльности

**Files:**
- Modify: `internal/repository/interfaces.go`
- Create: `internal/repository/postgres/loyalty.go`
- Create: `internal/repository/mock/loyalty.go`

- [x] LoyaltyRepository: GetAccount, CreateAccount, AddPoints, SpendPoints, UpdateLevel, ListTransactions (paginated)
- [x] Реализовать postgres и mock репозитории
- [x] Написать тесты
- [x] Запустить go test ./... - все тесты должны пройти

### Task 3: Сервис лояльности

**Files:**
- Create: `internal/service/loyalty_service.go`

- [x] LoyaltyService:
  - GetAccount(ctx, userID) — аккаунт лояльности (создать если нет)
  - EarnPoints(ctx, userID, bookingID) — начислить баллы (вызывается при booking completed)
  - SpendPoints(ctx, userID, amount, bookingID) — списать баллы при бронировании
  - GetDiscount(ctx, userID) — текущая скидка по уровню
  - RecalculateLevel(ctx, userID) — пересчитать уровень по количеству визитов
  - ListTransactions(ctx, userID, page, pageSize) — история операций
- [x] Формула начисления: TotalPrice / 100 * multiplier (зависит от уровня)
- [x] Написать unit-тесты
- [x] Запустить go test ./... - все тесты должны пройти

### Task 4: Интеграция с бронированием

**Files:**
- Modify: `internal/service/booking_service.go`
- Modify: `internal/handler/booking.go`

- [x] При complete бронирования — начислить баллы, пересчитать уровень
- [x] При создании бронирования — возможность оплатить баллами (частично)
- [x] В ответе бронирования: earned_points, loyalty_discount
- [x] Написать тесты
- [x] Запустить go test ./... - все тесты должны пройти

### Task 5: Хендлеры лояльности

**Files:**
- Create: `internal/handler/loyalty.go`
- Modify: `internal/server/router.go`

- [x] GET /api/v1/my/loyalty — аккаунт лояльности (уровень, баллы, привилегии)
- [x] GET /api/v1/my/loyalty/transactions — история операций (paginated)
- [x] GET /api/v1/my/loyalty/levels — описание всех уровней и привилегий
- [x] Зарегистрировать маршруты, добавить fx.Module
- [x] Написать handler-тесты
- [x] Запустить go test ./... - все тесты должны пройти

### Task 6: Верификация

- [ ] Запустить полный тест-сьют: go test ./... -v
- [ ] Запустить линтер: make lint
- [ ] Запустить go vet ./...
- [ ] Переместить этот план в docs/plans/completed/
