# Реферальная программа

## Overview

Персональная реферальная ссылка для каждого пользователя. Бонус приглашающему и приглашенному при первом бронировании. Статистика приглашений, вывод бонусов на бронирования.

**Коммерческая ценность:** Органический рост пользовательской базы с минимальными затратами на привлечение. CAC (Cost of Customer Acquisition) через рефералы в 3-5 раз ниже, чем через рекламу.

## Context

- Files involved: internal/domain/, internal/repository/, internal/service/, internal/handler/, migrations/
- Related patterns: Clean architecture, Uber fx DI
- Dependencies: нет новых
- Связь: бонус применяется при бронировании

## Development Approach

- **Testing approach**: Regular (code first, then tests)
- Complete each task fully before moving to the next
- **CRITICAL: every task MUST include new/updated tests**
- **CRITICAL: all tests must pass before starting next task**

## Implementation Steps

### Task 1: Доменные модели реферальной программы

**Files:**
- Create: `internal/domain/referral.go`
- Modify: `internal/domain/errors.go`
- Create: `migrations/000009_referrals.up.sql`
- Create: `migrations/000009_referrals.down.sql`

- [x] Создать модель Referral:
  ```
  Referral {
    ID            uuid.UUID
    ReferrerID    uuid.UUID       // кто пригласил
    RefereeID     uuid.UUID       // кого пригласили
    ReferralCode  string          // уникальный код приглашающего
    Status        ReferralStatus  // pending, completed, expired
    BonusAmount   int64           // бонус в копейках
    CompletedAt   *time.Time      // когда реферал сделал первое бронирование
    CreatedAt     time.Time
  }
  ```
- [x] Создать модель ReferralBalance (UserID, Balance int64, TotalEarned int64)
- [x] Добавить поле ReferralCode в модель User
- [x] Добавить domain-ошибки: ErrSelfReferral, ErrAlreadyReferred, ErrInsufficientReferralBalance
- [x] Создать миграции
- [x] Написать тесты
- [x] Запустить go test ./... - все тесты должны пройти

### Task 2: Репозиторий и сервис рефералов

**Files:**
- Modify: `internal/repository/interfaces.go`
- Create: `internal/repository/postgres/referral.go`
- Create: `internal/repository/mock/referral.go`
- Create: `internal/service/referral_service.go`

- [x] ReferralRepository: Create, GetByReferee, ListByReferrer, UpdateStatus, GetBalance, UpdateBalance
- [x] ReferralService:
  - GenerateCode(ctx, userID) — генерация уникального кода
  - RegisterReferral(ctx, referralCode, newUserID) — регистрация реферала при регистрации
  - CompleteReferral(ctx, refereeID) — начислить бонус при первом бронировании
  - GetBalance(ctx, userID) — баланс бонусов
  - UseBalance(ctx, userID, amount, bookingID) — списание бонусов при бронировании
  - GetStats(ctx, userID) — статистика приглашений
- [x] Написать unit-тесты
- [x] Запустить go test ./... - все тесты должны пройти

### Task 3: Интеграция с регистрацией и бронированием

**Files:**
- Modify: `internal/service/auth_service.go`
- Modify: `internal/handler/auth.go`
- Modify: `internal/service/booking_service.go`

- [x] При регистрации: если указан referral_code, создать Referral запись
- [x] При первом бронировании: начислить бонус обоим (configurable: default 500 руб)
- [x] При бронировании: возможность оплатить реферальными бонусами (частично)
- [x] Написать тесты
- [x] Запустить go test ./... - все тесты должны пройти

### Task 4: Хендлеры реферальной программы

**Files:**
- Create: `internal/handler/referral.go`
- Modify: `internal/server/router.go`

- [ ] GET /api/v1/my/referral — мой реферальный код и ссылка
- [ ] GET /api/v1/my/referral/stats — статистика (приглашено, завершено, заработано)
- [ ] GET /api/v1/my/referral/balance — текущий баланс бонусов
- [ ] Зарегистрировать маршруты, добавить fx.Module
- [ ] Написать handler-тесты
- [ ] Запустить go test ./... - все тесты должны пройти

### Task 5: Верификация

- [ ] Запустить полный тест-сьют: go test ./... -v
- [ ] Запустить линтер: make lint
- [ ] Запустить go vet ./...
- [ ] Переместить этот план в docs/plans/completed/
