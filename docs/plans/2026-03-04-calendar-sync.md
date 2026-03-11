# Синхронизация календаря

## Overview

iCal-экспорт расписания бронирований для владельцев. Импорт занятости из Google Calendar и Яндекс.Календарь. Автоматическая блокировка слотов при внешних событиях. Двусторонняя синхронизация.

**Коммерческая ценность:** Решает проблему двойного бронирования для бань, работающих на нескольких платформах. Снижает отмены из-за конфликтов на 80%. Premium-фича для подписчиков.

## Context

- Files involved: internal/domain/, internal/repository/, internal/service/, internal/handler/, migrations/
- Related patterns: Clean architecture, Uber fx DI
- Dependencies: нет новых Go-зависимостей (iCal — простой текстовый формат)
- Текущее: BookingRepository.CheckAvailability проверяет пересечения слотов

## Development Approach

- **Testing approach**: Regular (code first, then tests)
- Complete each task fully before moving to the next
- **CRITICAL: every task MUST include new/updated tests**
- **CRITICAL: all tests must pass before starting next task**

## Implementation Steps

### Task 1: iCal-экспорт (выдача)

**Files:**
- Create: `internal/calendar/ical.go`
- Create: `internal/handler/calendar.go`
- Modify: `internal/server/router.go`

- [x] Генератор iCal-формата:
  ```
  BEGIN:VCALENDAR
  VERSION:2.0
  PRODID:-//Bani//Booking Calendar//RU
  BEGIN:VEVENT
  DTSTART:20260304T100000Z
  DTEND:20260304T120000Z
  SUMMARY:Бронирование - Баня "Название"
  DESCRIPTION:Гостей: 4, Контакт: +7...
  END:VEVENT
  END:VCALENDAR
  ```
- [x] GET /api/v1/my/bathhouses/{id}/calendar.ics — экспорт бронирований в iCal (auth + уникальный токен)
- [x] Уникальный URL для подписки (без JWT, по секретному токену): /calendar/{secret_token}.ics
- [x] Написать тесты
- [x] Запустить go test ./... - все тесты должны пройти

### Task 2: Внешние блокировки слотов

**Files:**
- Create: `internal/domain/slot_block.go`
- Create: `migrations/000019_slot_blocks.up.sql`
- Create: `migrations/000019_slot_blocks.down.sql`

- [ ] Модель SlotBlock:
  ```
  SlotBlock {
    ID            uuid.UUID
    BathhouseID   uuid.UUID
    StartTime     time.Time
    EndTime       time.Time
    Source        string    // "manual", "google_calendar", "yandex_calendar"
    ExternalID    string    // ID события во внешнем календаре
    Description   string
    CreatedAt     time.Time
  }
  ```
- [ ] Создать миграцию
- [ ] Интегрировать с CheckAvailability: проверять и slot_blocks при бронировании
- [ ] Написать тесты
- [ ] Запустить go test ./... - все тесты должны пройти

### Task 3: Импорт из внешних календарей

**Files:**
- Create: `internal/calendar/sync.go`
- Create: `internal/calendar/parser.go`
- Modify: `internal/repository/interfaces.go`
- Create: `internal/repository/postgres/slot_block.go`
- Create: `internal/repository/mock/slot_block.go`

- [ ] Парсер iCal-формата (для Google Calendar и Яндекс.Календарь — оба используют iCal)
- [ ] CalendarSync service:
  - AddExternalCalendar(ctx, bathhouseID, calendarURL) — подписаться на внешний календарь
  - SyncCalendar(ctx, bathhouseID) — синхронизировать (вызывается по cron)
  - RemoveExternalCalendar(ctx, bathhouseID, calendarID) — отписаться
- [ ] SlotBlockRepository: Create, Delete, ListByBathhouse, GetByExternalID, DeleteBySource
- [ ] Cron-задача: синхронизация каждые 15 минут
- [ ] Написать тесты
- [ ] Запустить go test ./... - все тесты должны пройти

### Task 4: Хендлеры управления календарями

**Files:**
- Modify: `internal/handler/calendar.go`
- Modify: `internal/server/router.go`

- [ ] POST /api/v1/my/bathhouses/{id}/external-calendars — добавить внешний календарь (URL)
- [ ] GET /api/v1/my/bathhouses/{id}/external-calendars — список внешних календарей
- [ ] DELETE /api/v1/my/external-calendars/{id} — удалить внешний календарь
- [ ] POST /api/v1/my/bathhouses/{id}/external-calendars/sync — принудительная синхронизация
- [ ] POST /api/v1/my/bathhouses/{id}/slot-blocks — ручная блокировка слота
- [ ] DELETE /api/v1/my/slot-blocks/{id} — снять ручную блокировку
- [ ] Зарегистрировать маршруты, добавить fx.Module
- [ ] Написать handler-тесты
- [ ] Запустить go test ./... - все тесты должны пройти

### Task 5: Верификация

- [ ] Запустить полный тест-сьют: go test ./... -v
- [ ] Запустить линтер: make lint
- [ ] Запустить go vet ./...
- [ ] Переместить этот план в docs/plans/completed/
