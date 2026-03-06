# Анализ согласованности проекта, .claudeignore и e2e тесты

## Overview

Три направления работы:
1. Документирование найденных архитектурных несогласованностей (отчет)
2. Создание .claudeignore для оптимизации работы Claude с проектом
3. Создание e2e (hurl) тестов для фич, не покрытых тестами: notifications, subscriptions/promotions, pricing rules, recommendations, OAuth, widget API

## Результаты архитектурного анализа

### Критические нарушения clean architecture (handler -> repository напрямую):
- `handler/bathhouse_handler.go` — инжектит `PromotionRepository` напрямую (нет PromotionService)
- `handler/subscription.go` — инжектит `PromotionRepository`, `BathhouseRepository`, `*service.AccessChecker` (конкретный тип)
- `handler/admin_handler.go` — инжектит `ReviewRepository` для модерации (ReviewService не имеет admin-методов)
- `handler/widget.go` — инжектит `BathhouseRepository` (BathhouseService не имеет GetByAPIKey)
- `handler/pricing.go` — инжектит `BathhouseRepository` для получения base price
- `handler/recommendation_handler.go` — инжектит `BathhouseRepository` (сервис возвращает []uuid.UUID вместо объектов)

### Средние проблемы:
- Дублирование `isValidTimeFormat` / `IsValidTimeFormat` в domain/pricing.go и domain/bathhouse.go
- Двойная авторизация в SubscriptionHandler (handler + service оба вызывают CanManageBathhouse)
- N+1 запросы в RecommendationService и RecommendationHandler
- Непоследовательное именование поля AccessChecker: `accessChecker` vs `access`
- Файл `recommendation.go` без суффикса `_repo` в postgres/
- Опечатка `GetActiveBybathhouse` (маленькая b) во всех слоях
- Пропуски в нумерации миграций (000010-000012, 000014-000016)

### Мелкие проблемы:
- `Favorite`, `SocialAccount`, `Representative` без Validate()
- RecommendationService не инжектит logger
- User endpoints живут в AuthHandler вместо отдельного UserHandler

## Context

- Existing hurl tests: auth, bathhouses, bookings, cities, favorites, representatives, reviews, admin, health
- Missing hurl tests: notifications, subscriptions, promotions, pricing rules, recommendations, OAuth, widget
- Existing test infrastructure: `tests/hurl/run_all_tests.sh`, `tests/hurl/setup.sh`

## Development Approach

- **Testing approach**: Regular (code first)
- Each task is independent and can be done in any order
- Follow existing hurl test patterns from `tests/hurl/`
- **CRITICAL: every task MUST include new/updated tests**
- **CRITICAL: all tests must pass before starting next task**

## Implementation Steps

### Task 1: Создать отчет об архитектурном анализе

**Files:**
- Create: `docs/architecture-review.md`

- [x] Оформить все найденные проблемы в структурированный markdown-отчет
- [x] Разделить по приоритетам: critical, medium, low
- [x] Для каждой проблемы указать файлы и предложить решение

### Task 2: Создать .claudeignore

**Files:**
- Create: `.claudeignore`

- [x] Создать .claudeignore с исключениями:
  - `bin/` — скомпилированные бинарники
  - `bani` и `bani-server` — бинарники в корне
  - `coverage.out` — файл покрытия
  - `widget/dist/` — собранный виджет
  - `migrations/` — SQL миграции (редко нужны для code review)
  - `go.sum` — lock-файл зависимостей
  - `.git/` — git internal
  - `.DS_Store` — OS файлы
  - `.ralphex/` — ralphex internal
  - `docs/plans/` — планы реализации (не код)
  - `*.out` — coverage и прочие output файлы

### Task 3: e2e тесты — Subscriptions и Promotions

**Files:**
- Create: `tests/hurl/subscriptions.hurl`
- Create: `tests/hurl/subscriptions_negative.hurl`

- [x] Happy path: создать подписку (free/premium/promoted), получить, отменить, список подписок владельца
- [x] Создать промо-кампанию для bathhouse с promoted-подпиской
- [x] Негативные: подписка без авторизации, на чужую баню, дубликат active подписки, промо без promoted-плана

### Task 4: e2e тесты — Pricing Rules

**Files:**
- Create: `tests/hurl/pricing_rules.hurl`
- Create: `tests/hurl/pricing_rules_negative.hurl`

- [x] Happy path: создать правило (weekday/weekend/time_range), список, обновить, удалить
- [x] Публичный price calculator: расчет цены с правилами
- [x] Негативные: создание без авторизации, невалидное время, несуществующая баня

### Task 5: e2e тесты — Recommendations

**Files:**
- Create: `tests/hurl/recommendations.hurl`

- [x] Персонализированные рекомендации (auth required)
- [x] Похожие бани (public)
- [x] Популярные в городе (public)
- [x] Получение/обновление пользовательских предпочтений

### Task 6: e2e тесты — Notifications

**Files:**
- Create: `tests/hurl/notifications.hurl`

- [ ] Получить список уведомлений (auth required)
- [ ] Пометить как прочитанное
- [ ] Получить/обновить настройки уведомлений
- [ ] Получить кол-во непрочитанных

### Task 7: e2e тесты — Widget API

**Files:**
- Create: `tests/hurl/widget.hurl`

- [ ] Получить данные бани по API key
- [ ] Получить доступные слоты
- [ ] Создать бронирование через виджет

### Task 8: Обновить run_all_tests.sh

**Files:**
- Modify: `tests/hurl/run_all_tests.sh`

- [ ] Добавить новые hurl-файлы в скрипт запуска
- [ ] Убедиться что порядок выполнения корректен (subscriptions после bathhouses, pricing после subscriptions и т.д.)

### Task 9: Verify acceptance criteria

- [ ] Все новые hurl-файлы синтаксически корректны (hurl --check)
- [ ] .claudeignore создан и корректен
- [ ] Отчет об архитектуре полный и структурированный
- [ ] go build ./... проходит без ошибок
- [ ] go test ./... проходит без ошибок

### Task 10: Update documentation

- [ ] Обновить CLAUDE.md если внутренние паттерны изменились
- [ ] Переместить этот план в `docs/plans/completed/`
