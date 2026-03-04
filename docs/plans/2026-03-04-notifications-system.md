# Система уведомлений

## Overview

In-app уведомления + email + push. Типы событий: статус бронирования, новый отзыв, ответ на отзыв, промо-акции, напоминание о визите. WebSocket для real-time доставки, настройки каналов для пользователя (выбор: in-app, email, push).

**Коммерческая ценность:** Увеличивает возврат пользователей (retention). Push-уведомления о промо-акциях повышают конверсию на 20-30%. Напоминания снижают no-show на 40%.

## Context

- Files involved: internal/domain/, internal/repository/, internal/service/, internal/handler/, internal/server/router.go, migrations/
- Related patterns: Clean architecture, Uber fx DI, chi router
- Dependencies: gorilla/websocket (WebSocket), smtp/sendgrid (email)
- Связь: интегрируется с booking, review, promo системами

## Development Approach

- **Testing approach**: Regular (code first, then tests)
- Complete each task fully before moving to the next
- **CRITICAL: every task MUST include new/updated tests**
- **CRITICAL: all tests must pass before starting next task**

## Implementation Steps

### Task 1: Доменные модели уведомлений

**Files:**
- Create: `internal/domain/notification.go`
- Create: `migrations/000010_notifications.up.sql`
- Create: `migrations/000010_notifications.down.sql`

- [x] Создать модель Notification:
  ```
  Notification {
    ID          uuid.UUID
    UserID      uuid.UUID
    Type        NotificationType  // booking_confirmed, booking_cancelled, new_review, review_response, promo, reminder, system
    Title       string
    Body        string
    Data        map[string]string // доп. данные (booking_id, bathhouse_id, etc.)
    IsRead      bool
    ReadAt      *time.Time
    CreatedAt   time.Time
  }
  ```
- [x] Создать модель NotificationPreferences:
  ```
  NotificationPreferences {
    UserID        uuid.UUID
    InApp         bool  // default true
    Email         bool  // default true
    Push          bool  // default false
    BookingEvents bool  // default true
    ReviewEvents  bool  // default true
    PromoEvents   bool  // default true
    Reminders     bool  // default true
  }
  ```
- [x] Создать миграции
- [x] Написать тесты
- [x] Запустить go test ./... - все тесты должны пройти

### Task 2: Репозиторий уведомлений

**Files:**
- Modify: `internal/repository/interfaces.go`
- Create: `internal/repository/postgres/notification.go`
- Create: `internal/repository/mock/notification.go`

- [ ] NotificationRepository: Create, GetByID, ListByUser (paginated), MarkAsRead, MarkAllAsRead, CountUnread, GetPreferences, UpdatePreferences
- [ ] Реализовать postgres и mock репозитории
- [ ] Написать тесты
- [ ] Запустить go test ./... - все тесты должны пройти

### Task 3: Сервис уведомлений и каналы доставки

**Files:**
- Create: `internal/service/notification_service.go`
- Create: `internal/notification/email.go`
- Create: `internal/notification/dispatcher.go`

- [ ] NotificationService:
  - Send(ctx, userID, type, title, body, data) — создать уведомление и отправить по настроенным каналам
  - List(ctx, userID, page, pageSize) — список уведомлений
  - MarkAsRead(ctx, userID, notificationID) — прочитать
  - MarkAllAsRead(ctx, userID) — прочитать все
  - GetUnreadCount(ctx, userID) — счетчик непрочитанных
  - UpdatePreferences(ctx, userID, prefs) — настройки каналов
- [ ] Dispatcher: маршрутизация по каналам (in-app, email) с учетом preferences
- [ ] EmailSender: отправка email через SMTP/SendGrid
- [ ] Написать unit-тесты
- [ ] Запустить go test ./... - все тесты должны пройти

### Task 4: WebSocket для real-time уведомлений

**Files:**
- Create: `internal/handler/ws.go`
- Modify: `internal/server/router.go`

- [ ] WebSocket endpoint: GET /api/v1/ws/notifications — подключение по JWT-токену
- [ ] Hub-паттерн: регистрация/отключение клиентов, отправка уведомлений подключенным пользователям
- [ ] При создании in-app уведомления: отправить через WebSocket если пользователь онлайн
- [ ] Heartbeat/ping-pong для поддержания соединения
- [ ] Написать тесты
- [ ] Запустить go test ./... - все тесты должны пройти

### Task 5: Хендлеры уведомлений и интеграция

**Files:**
- Create: `internal/handler/notification.go`
- Modify: `internal/server/router.go`
- Modify: `internal/service/booking_service.go`
- Modify: `internal/service/review_service.go`

- [ ] GET /api/v1/my/notifications — список уведомлений (paginated)
- [ ] GET /api/v1/my/notifications/unread-count — счетчик непрочитанных
- [ ] PATCH /api/v1/my/notifications/{id}/read — прочитать
- [ ] PATCH /api/v1/my/notifications/read-all — прочитать все
- [ ] GET /api/v1/my/notification-preferences — текущие настройки
- [ ] PUT /api/v1/my/notification-preferences — обновить настройки
- [ ] Интегрировать отправку уведомлений: при подтверждении/отмене бронирования, при новом отзыве, при ответе на отзыв
- [ ] Зарегистрировать маршруты, добавить fx.Module
- [ ] Написать handler-тесты
- [ ] Запустить go test ./... - все тесты должны пройти

### Task 6: Верификация

- [ ] Запустить полный тест-сьют: go test ./... -v
- [ ] Запустить линтер: make lint
- [ ] Запустить go vet ./...
- [ ] Обновить CLAUDE.md
- [ ] Переместить этот план в docs/plans/completed/
