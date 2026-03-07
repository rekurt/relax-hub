# Система промокодов и скидок

## Overview

Создание и применение промокодов для скидок на бронирования. Владельцы бань создают промокоды для своих заведений, админы — глобальные промокоды платформы. Типы скидок: процент, фиксированная сумма, бесплатный час. Лимиты по количеству использований, сроку действия, минимальной сумме.

**Коммерческая ценность:** Инструмент маркетинга для владельцев. Привлечение новых клиентов, увеличение повторных визитов. Промокоды платформы стимулируют первые бронирования.

## Context

- Files involved: internal/domain/, internal/repository/, internal/service/, internal/handler/, migrations/
- Related patterns: Clean architecture, Uber fx DI
- Dependencies: нет новых внешних зависимостей
- Связь: применяется при создании бронирования, влияет на TotalPrice

## Development Approach

- **Testing approach**: Regular (code first, then tests)
- Complete each task fully before moving to the next
- **CRITICAL: every task MUST include new/updated tests**
- **CRITICAL: all tests must pass before starting next task**

## Implementation Steps

### Task 1: Доменные модели промокодов

**Files:**
- Create: `internal/domain/promo.go`
- Modify: `internal/domain/errors.go`
- Create: `migrations/000010_promo_codes.up.sql`
- Create: `migrations/000010_promo_codes.down.sql`

- [x] Создать модель PromoCode:
  ```
  PromoCode {
    ID              uuid.UUID
    Code            string          // уникальный код, uppercase
    Type            PromoType       // percentage, fixed_amount, free_hour
    Value           int64           // процент (10 = 10%) или сумма в копейках
    BathhouseID     *uuid.UUID      // nil = глобальный
    CreatorID       uuid.UUID       // owner или admin
    MaxUses         int             // 0 = безлимит
    CurrentUses     int
    MinAmount       int64           // минимальная сумма бронирования в копейках
    ValidFrom       time.Time
    ValidUntil      time.Time
    IsActive        bool
    CreatedAt       time.Time
  }
  ```
- [x] Создать модель PromoUsage (ID, PromoCodeID, UserID, BookingID, DiscountAmount, UsedAt)
- [x] Добавить domain-ошибки: ErrPromoNotFound, ErrPromoExpired, ErrPromoMaxUses, ErrPromoMinAmount, ErrPromoInvalid
- [x] Создать миграцию с таблицами promo_codes и promo_usages
- [x] Написать тесты
- [x] Запустить go test ./... - все тесты должны пройти

### Task 2: Репозиторий промокодов

**Files:**
- Modify: `internal/repository/interfaces.go`
- Create: `internal/repository/postgres/promo.go`
- Create: `internal/repository/mock/promo.go`

- [ ] PromoCodeRepository interface: Create, GetByID, GetByCode, Update, ListByBathhouse, ListByCreator, IncrementUses, RecordUsage
- [ ] Реализовать postgres и mock репозитории
- [ ] Написать тесты
- [ ] Запустить go test ./... - все тесты должны пройти

### Task 3: Сервис промокодов

**Files:**
- Create: `internal/service/promo_service.go`

- [ ] PromoService:
  - Create(ctx, promo) — создать промокод (owner для своей бани, admin для глобальных)
  - Validate(ctx, code, bathhouseID, amount) — проверить промокод и вернуть сумму скидки
  - Apply(ctx, code, bookingID) — применить промокод к бронированию
  - Deactivate(ctx, promoID) — деактивировать промокод
  - ListByBathhouse(ctx, bathhouseID) — промокоды бани
- [ ] RBAC: owner создает для своей бани, admin — глобальные
- [ ] Логика расчета скидки: percentage от суммы, fixed_amount вычитается, free_hour уменьшает время
- [ ] Написать unit-тесты
- [ ] Запустить go test ./... - все тесты должны пройти

### Task 4: Интеграция с бронированием

**Files:**
- Modify: `internal/service/booking_service.go`
- Modify: `internal/handler/booking.go`

- [ ] Добавить поле PromoCode в запрос создания бронирования
- [ ] При создании бронирования: если указан промокод — валидировать, рассчитать скидку, сохранить PromoUsage
- [ ] В ответе бронирования отображать original_price и discount_amount
- [ ] Написать тесты
- [ ] Запустить go test ./... - все тесты должны пройти

### Task 5: Хендлеры промокодов

**Files:**
- Create: `internal/handler/promo.go`
- Modify: `internal/server/router.go`

- [ ] POST /api/v1/my/bathhouses/{id}/promo-codes — создать промокод для бани
- [ ] GET /api/v1/my/bathhouses/{id}/promo-codes — список промокодов бани
- [ ] DELETE /api/v1/promo-codes/{id} — деактивировать промокод
- [ ] POST /api/v1/promo-codes/validate — проверить промокод (публичный)
- [ ] POST /api/v1/admin/promo-codes — создать глобальный промокод (admin)
- [ ] Зарегистрировать маршруты, добавить fx.Module
- [ ] Написать handler-тесты
- [ ] Запустить go test ./... - все тесты должны пройти

### Task 6: Верификация

- [ ] Запустить полный тест-сьют: go test ./... -v
- [ ] Запустить линтер: make lint
- [ ] Запустить go vet ./...
- [ ] Обновить CLAUDE.md
- [ ] Переместить этот план в docs/plans/completed/
