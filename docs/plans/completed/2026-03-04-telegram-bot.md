# Telegram-бот

## Overview

Telegram-бот для поиска бань, просмотра доступных слотов и бронирования. Уведомления о статусе бронирования, напоминания за день до визита. Inline-режим для быстрого поиска. Привязка Telegram-аккаунта к аккаунту платформы.

**Коммерческая ценность:** Telegram — основной мессенджер в России (90M+ пользователей). Бот снижает барьер входа: бронирование без установки приложения. Увеличивает reach на 30-40%.

## Context

- Files involved: создается как отдельный модуль cmd/bot/, internal/bot/
- Related patterns: Clean architecture, Uber fx DI
- Dependencies: go-telegram-bot-api/telegram-bot-api (Telegram Bot API)
- Связь: использует существующие сервисы (BathhouseService, BookingService) через DI

## Development Approach

- **Testing approach**: Regular (code first, then tests)
- Complete each task fully before moving to the next
- **CRITICAL: every task MUST include new/updated tests**
- **CRITICAL: all tests must pass before starting next task**

## Implementation Steps

### Task 1: Инфраструктура бота

**Files:**
- Create: `cmd/bot/main.go`
- Create: `internal/bot/bot.go`
- Create: `internal/bot/handler.go`
- Create: `internal/bot/module.go`
- Modify: `config/config.go`

- [x] Создать точку входа cmd/bot/main.go с Uber fx
- [x] Конфигурация: BANI_TELEGRAM_BOT_TOKEN, BANI_TELEGRAM_WEBHOOK_URL (опционально)
- [x] Bot struct с зависимостями на сервисы
- [x] Long polling режим для разработки, webhook для production
- [x] Написать тесты инициализации
- [x] Запустить go test ./... - все тесты должны пройти

### Task 2: Привязка Telegram-аккаунта

**Files:**
- Create: `internal/domain/telegram.go`
- Create: `migrations/000016_telegram_links.up.sql`
- Create: `migrations/000016_telegram_links.down.sql`
- Modify: `internal/repository/interfaces.go`
- Create: `internal/repository/postgres/telegram.go`
- Create: `internal/repository/mock/telegram.go`

- [x] Модель TelegramLink (UserID, TelegramID int64, TelegramUsername, LinkedAt)
- [x] TelegramLinkRepository: Create, GetByTelegramID, GetByUserID, Delete
- [x] Команда /link <token> — привязать через одноразовый токен, генерируемый в веб-интерфейсе
- [x] Создать миграции
- [x] Написать тесты
- [x] Запустить go test ./... - все тесты должны пройти

### Task 3: Основные команды бота

**Files:**
- Modify: `internal/bot/handler.go`
- Create: `internal/bot/keyboards.go`

- [x] /start — приветствие, описание возможностей
- [x] /search <город> — поиск бань в городе (inline-кнопки с результатами)
- [x] /book <id> — начать бронирование бани (пошаговый wizard: дата -> время -> гости -> подтверждение)
- [x] /mybookings — мои бронирования (список с кнопками управления)
- [x] /cancel <id> — отменить бронирование
- [x] /favorites — мое избранное
- [x] Inline-кнопки (callback_query) для навигации по результатам
- [x] Написать тесты
- [x] Запустить go test ./... - все тесты должны пройти

### Task 4: Inline-режим и уведомления

**Files:**
- Modify: `internal/bot/handler.go`
- Create: `internal/bot/notifications.go`

- [x] Inline-режим: @bani_bot <запрос> — поиск бань с inline-результатами
- [x] Уведомления через бота (если привязан Telegram):
  - Подтверждение/отмена бронирования
  - Напоминание за 24 часа до визита
  - Новый отзыв на баню (для владельцев)
- [x] Интеграция с NotificationService: добавить канал "telegram"
- [x] Написать тесты
- [x] Запустить go test ./... - все тесты должны пройти

### Task 5: Верификация

- [x] Запустить полный тест-сьют: go test ./... -v
- [x] Запустить линтер: make lint
- [x] Запустить go vet ./...
- [x] Переместить этот план в docs/plans/completed/
