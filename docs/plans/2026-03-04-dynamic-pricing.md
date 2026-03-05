# Динамическое ценообразование

## Overview

Гибкая система ценообразования для владельцев бань: разные цены для будней/выходных/праздников, сезонные коэффициенты, happy hours со скидкой. Владелец настраивает ценовые правила через API, система автоматически рассчитывает стоимость при бронировании.

**Коммерческая ценность:** Увеличивает загрузку в непопулярное время (будни, утро) за счет скидок. Повышает доход в пиковые периоды (выходные, праздники). Среднее увеличение выручки бани: 15-25%.

## Context

- Files involved: internal/domain/, internal/repository/, internal/service/, internal/handler/, migrations/
- Related patterns: Clean architecture, Uber fx DI
- Dependencies: нет новых
- Текущее: Bathhouse.PricePerHour (int64, копейки) — фиксированная цена. Бронирование рассчитывает TotalPrice = PricePerHour * hours

## Development Approach

- **Testing approach**: Regular (code first, then tests)
- Complete each task fully before moving to the next
- **CRITICAL: every task MUST include new/updated tests**
- **CRITICAL: all tests must pass before starting next task**

## Implementation Steps

### Task 1: Доменные модели ценовых правил

**Files:**
- Create: `internal/domain/pricing.go`
- Create: `migrations/000008_dynamic_pricing.up.sql`
- Create: `migrations/000008_dynamic_pricing.down.sql`

- [x] Создать модель PricingRule:
  ```
  PricingRule {
    ID            uuid.UUID
    BathhouseID   uuid.UUID
    Name          string          // "Выходные", "Happy Hour", "Новый Год"
    Type          PricingRuleType // weekday, weekend, holiday, time_range, season
    Multiplier    float64         // 1.5 = +50%, 0.8 = -20%
    DaysOfWeek    []int           // [5, 6] для выходных
    TimeFrom      *string         // "08:00"
    TimeTo        *string         // "12:00"
    DateFrom      *time.Time      // для сезонных/праздничных
    DateTo        *time.Time
    Priority      int             // при конфликте правил — побеждает с высшим приоритетом
    IsActive      bool
    CreatedAt     time.Time
  }
  ```
- [x] Создать миграцию с таблицей pricing_rules
- [x] Написать тесты валидации
- [x] Запустить go test ./... - все тесты должны пройти

### Task 2: Репозиторий и сервис ценообразования

**Files:**
- Modify: `internal/repository/interfaces.go`
- Create: `internal/repository/postgres/pricing.go`
- Create: `internal/repository/mock/pricing.go`
- Create: `internal/service/pricing_service.go`

- [x] PricingRuleRepository: Create, Update, Delete, ListByBathhouse, GetActiveRules
- [x] PricingService:
  - CalculatePrice(ctx, bathhouseID, startTime, endTime) — рассчитать цену с учетом правил
  - CreateRule(ctx, rule) — создать правило
  - UpdateRule(ctx, rule) — обновить правило
  - DeleteRule(ctx, ruleID) — удалить правило
  - ListRules(ctx, bathhouseID) — правила бани
- [x] Алгоритм расчета: разбить интервал на часы, для каждого часа найти правило с наивысшим приоритетом, применить множитель к базовой цене
- [x] RBAC: только owner/representative бани
- [x] Написать unit-тесты (важно: граничные случаи с перекрытием правил)
- [x] Запустить go test ./... - все тесты должны пройти

### Task 3: Интеграция с бронированием и слотами

**Files:**
- Modify: `internal/service/booking_service.go`
- Modify: `internal/handler/bathhouse.go`

- [x] При создании бронирования: рассчитывать TotalPrice через PricingService вместо фиксированной цены
- [x] В GetAvailableSlots: показывать цену с учетом динамического ценообразования для каждого слота
- [x] Написать тесты
- [x] Запустить go test ./... - все тесты должны пройти

### Task 4: Хендлеры ценовых правил

**Files:**
- Create: `internal/handler/pricing.go`
- Modify: `internal/server/router.go`

- [x] POST /api/v1/my/bathhouses/{id}/pricing-rules — создать правило
- [x] GET /api/v1/my/bathhouses/{id}/pricing-rules — список правил
- [x] PUT /api/v1/pricing-rules/{id} — обновить правило
- [x] DELETE /api/v1/pricing-rules/{id} — удалить правило
- [x] GET /api/v1/bathhouses/{id}/price-calculator?start=...&end=... — публичный калькулятор цены
- [x] Зарегистрировать маршруты, добавить fx.Module
- [x] Написать handler-тесты
- [x] Запустить go test ./... - все тесты должны пройти

### Task 5: Верификация

- [ ] Запустить полный тест-сьют: go test ./... -v
- [ ] Запустить линтер: make lint
- [ ] Запустить go vet ./...
- [ ] Переместить этот план в docs/plans/completed/
