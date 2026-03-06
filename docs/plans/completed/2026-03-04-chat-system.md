# Чат между клиентом и владельцем

## Overview

Real-time чат-система между клиентами и владельцами/представителями бань. Чат привязан к конкретной бане или бронированию. История сообщений, прочитано/не прочитано, уведомления о новых сообщениях. WebSocket для мгновенной доставки.

**Коммерческая ценность:** Снижает барьер между клиентом и баней. Увеличивает конверсию из просмотра в бронирование на 15-20%. Позволяет решать вопросы без звонков, что предпочитают 70% молодой аудитории.

## Context

- Files involved: internal/domain/, internal/repository/, internal/service/, internal/handler/, internal/server/router.go, migrations/
- Related patterns: Clean architecture, Uber fx DI, chi router
- Dependencies: gorilla/websocket (можно переиспользовать из notifications-system)
- Связь: интегрируется с notifications-system для уведомлений о новых сообщениях

## Development Approach

- **Testing approach**: Regular (code first, then tests)
- Complete each task fully before moving to the next
- **CRITICAL: every task MUST include new/updated tests**
- **CRITICAL: all tests must pass before starting next task**

## Implementation Steps

### Task 1: Доменные модели чата

**Files:**
- Create: `internal/domain/chat.go`
- Create: `migrations/000011_chat.up.sql`
- Create: `migrations/000011_chat.down.sql`

- [x] Создать модель Conversation:
  ```
  Conversation {
    ID            uuid.UUID
    BathhouseID   uuid.UUID
    ClientID      uuid.UUID
    BookingID     *uuid.UUID      // опционально привязан к бронированию
    LastMessageAt *time.Time
    CreatedAt     time.Time
  }
  ```
- [x] Создать модель Message:
  ```
  Message {
    ID              uuid.UUID
    ConversationID  uuid.UUID
    SenderID        uuid.UUID
    Text            string
    IsRead          bool
    ReadAt          *time.Time
    CreatedAt       time.Time
  }
  ```
- [x] Создать миграции с UNIQUE(bathhouse_id, client_id) на conversations
- [x] Написать тесты
- [x] Запустить go test ./... - все тесты должны пройти

### Task 2: Репозиторий чата

**Files:**
- Modify: `internal/repository/interfaces.go`
- Create: `internal/repository/postgres/chat.go`
- Create: `internal/repository/mock/chat.go`

- [x] ConversationRepository: Create, GetByID, GetByParticipants, ListByUser (paginated), GetOrCreate
- [x] MessageRepository: Create, ListByConversation (paginated, reverse chronological), MarkAsRead, CountUnread
- [x] Реализовать postgres и mock репозитории
- [x] Написать тесты
- [x] Запустить go test ./... - все тесты должны пройти

### Task 3: Сервис чата

**Files:**
- Create: `internal/service/chat_service.go`

- [x] ChatService:
  - StartConversation(ctx, bathhouseID, clientID, bookingID) — создать/получить беседу
  - SendMessage(ctx, conversationID, senderID, text) — отправить сообщение
  - ListConversations(ctx, userID, page, pageSize) — список бесед (для клиента и для владельца)
  - ListMessages(ctx, conversationID, userID, page, pageSize) — сообщения (с проверкой доступа)
  - MarkAsRead(ctx, conversationID, userID) — прочитать все сообщения в беседе
  - GetUnreadCount(ctx, userID) — общее число непрочитанных сообщений
- [x] RBAC: клиент видит только свои беседы, owner/representative — беседы по своим баням
- [x] При новом сообщении — отправить уведомление через NotificationService
- [x] Написать unit-тесты
- [x] Запустить go test ./... - все тесты должны пройти

### Task 4: WebSocket для real-time чата

**Files:**
- Modify: `internal/handler/ws.go` (расширить существующий WebSocket hub)

- [x] Расширить WebSocket hub: подписка на чат-комнаты (conversation_id)
- [x] При отправке сообщения через REST — доставить через WebSocket если получатель онлайн
- [x] Типы WS-сообщений: new_message, message_read, typing_indicator
- [x] Написать тесты
- [x] Запустить go test ./... - все тесты должны пройти

### Task 5: Хендлеры чата

**Files:**
- Create: `internal/handler/chat.go`
- Modify: `internal/server/router.go`

- [x] POST /api/v1/bathhouses/{id}/chat — начать беседу с баней
- [x] GET /api/v1/my/conversations — список бесед (paginated)
- [x] GET /api/v1/conversations/{id}/messages — сообщения (paginated)
- [x] POST /api/v1/conversations/{id}/messages — отправить сообщение
- [x] PATCH /api/v1/conversations/{id}/read — прочитать все сообщения
- [x] GET /api/v1/my/unread-messages-count — счетчик непрочитанных
- [x] Зарегистрировать маршруты, добавить fx.Module
- [x] Написать handler-тесты
- [x] Запустить go test ./... - все тесты должны пройти

### Task 6: Верификация

- [x] Запустить полный тест-сьют: go test ./... -v
- [x] Запустить линтер: make lint
- [x] Запустить go vet ./...
- [x] Переместить этот план в docs/plans/completed/
