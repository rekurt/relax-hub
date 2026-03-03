# Модерация отзывов

## Overview

Расширенная система модерации отзывов: очередь модерации с фильтрами, автомодерация (стоп-слова, спам-паттерны), массовые действия. Статусы отзывов: pending -> approved/rejected. Новые отзывы проходят через модерацию перед публикацией.

**Коммерческая ценность:** Качественные отзывы повышают доверие к платформе. Автомодерация снижает нагрузку на поддержку. Защита от конкурентных атак (фейковые негативные отзывы).

## Context

- Files involved: internal/domain/review.go, internal/repository/, internal/service/, internal/handler/, migrations/
- Related patterns: Clean architecture, Uber fx DI
- Dependencies: нет новых
- Текущее: Review уже имеет Status (pending/approved/rejected/hidden), но модерация не реализована — все отзывы сразу visible

## Development Approach

- **Testing approach**: Regular (code first, then tests)
- Complete each task fully before moving to the next
- **CRITICAL: every task MUST include new/updated tests**
- **CRITICAL: all tests must pass before starting next task**

## Implementation Steps

### Task 1: Автомодерация

**Files:**
- Create: `internal/moderation/filter.go`
- Create: `internal/moderation/stopwords.go`

- [ ] Создать ContentFilter:
  - CheckText(text string) (isClean bool, reasons []string) — проверка текста
  - Стоп-слова: список нецензурных слов и вариаций (транслит, замена символов)
  - Спам-паттерны: повторяющиеся символы (ааааа), CAPS LOCK, ссылки, телефоны конкурентов
  - Минимальная длина отзыва: 10 символов
  - Максимальная длина: 5000 символов
- [ ] Configurable через BANI_MODERATION_ENABLED, BANI_MODERATION_AUTO_APPROVE (bool)
- [ ] Написать тесты (важно: edge cases с обходом фильтров)
- [ ] Запустить go test ./... - все тесты должны пройти

### Task 2: Интеграция автомодерации с отзывами

**Files:**
- Modify: `internal/service/review_service.go`

- [ ] При создании отзыва:
  - Если MODERATION_ENABLED=true: status = pending, пропустить через ContentFilter
  - Если ContentFilter нашел нарушения: status = rejected, сохранить причины
  - Если MODERATION_AUTO_APPROVE=true и ContentFilter чистый: status = approved
  - Иначе: status = pending (ждет ручной модерации)
- [ ] Фильтрация в выдаче: показывать только approved отзывы (для публичных endpoint)
- [ ] Написать тесты
- [ ] Запустить go test ./... - все тесты должны пройти

### Task 3: Админ-хендлеры модерации отзывов

**Files:**
- Modify: `internal/handler/admin.go`
- Modify: `internal/server/router.go`

- [ ] GET /api/v1/admin/reviews — список отзывов с фильтрами (status, rating, date, bathhouse_id)
- [ ] GET /api/v1/admin/reviews/pending-count — счетчик ожидающих модерации
- [ ] PATCH /api/v1/admin/reviews/{id}/approve — одобрить
- [ ] PATCH /api/v1/admin/reviews/{id}/reject — отклонить (с указанием причины)
- [ ] POST /api/v1/admin/reviews/batch-approve — массовое одобрение (массив ID)
- [ ] POST /api/v1/admin/reviews/batch-reject — массовое отклонение
- [ ] Написать handler-тесты
- [ ] Запустить go test ./... - все тесты должны пройти

### Task 4: Уведомления о модерации

**Files:**
- Modify: `internal/service/review_service.go`

- [ ] Уведомить автора отзыва о результате модерации (approved/rejected + причина)
- [ ] Уведомить владельца бани о новом одобренном отзыве
- [ ] Интеграция с NotificationService (если реализован) или простой email
- [ ] Написать тесты
- [ ] Запустить go test ./... - все тесты должны пройти

### Task 5: Верификация

- [ ] Запустить полный тест-сьют: go test ./... -v
- [ ] Запустить линтер: make lint
- [ ] Запустить go vet ./...
- [ ] Переместить этот план в docs/plans/completed/
