# Интеграция онлайн-оплаты

## Overview

Подключение платежного шлюза (ЮKassa) для онлайн-оплаты бронирований. Клиенты смогут оплатить бронирование при создании, получить автоматический возврат при отмене, просматривать историю платежей. Владельцы видят статусы оплаты по своим баням.

**Коммерческая ценность:** Ключевая фича для монетизации. Позволяет платформе брать комиссию с каждого бронирования (3-5%). Снижает no-show, повышает доверие клиентов.

## Context

- Files involved: internal/domain/, internal/repository/, internal/service/, internal/handler/, internal/server/router.go, migrations/
- Related patterns: Clean architecture handler->service->repository, Uber fx DI, chi router
- Dependencies: yookassa-sdk-go (ЮKassa Go SDK)
- Текущая модель Booking: ID, UserID, BathhouseID, StartTime, EndTime, GuestCount, TotalPrice, Status, Comment
- Цены в копейках (int64)

## Development Approach

- **Testing approach**: Regular (code first, then tests)
- Complete each task fully before moving to the next
- **CRITICAL: every task MUST include new/updated tests**
- **CRITICAL: all tests must pass before starting next task**

## Implementation Steps

### Task 1: Доменные модели платежей

**Files:**
- Create: `internal/domain/payment.go`
- Modify: `internal/domain/errors.go`
- Create: `migrations/000004_payments.up.sql`
- Create: `migrations/000004_payments.down.sql`

- [x] Создать модель Payment:
  ```
  Payment {
    ID            uuid.UUID
    BookingID     uuid.UUID
    UserID        uuid.UUID
    Amount        int64           // в копейках
    Currency      string          // "RUB"
    Status        PaymentStatus   // pending, processing, succeeded, failed, refunded, partially_refunded
    Provider      string          // "yookassa"
    ExternalID    string          // ID транзакции в платежной системе
    RefundAmount  int64           // сумма возврата
    RefundedAt    *time.Time
    Metadata      map[string]string
    CreatedAt     time.Time
    UpdatedAt     time.Time
  }
  ```
- [x] Создать PaymentStatus с валидацией IsValid()
- [x] Добавить domain-ошибки: ErrPaymentNotFound, ErrPaymentAlreadyProcessed, ErrRefundExceedsAmount, ErrPaymentFailed
- [x] Создать миграцию:
  ```sql
  CREATE TABLE payments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    booking_id UUID NOT NULL REFERENCES bookings(id),
    user_id UUID NOT NULL REFERENCES users(id),
    amount BIGINT NOT NULL,
    currency VARCHAR(3) NOT NULL DEFAULT 'RUB',
    status VARCHAR(30) NOT NULL DEFAULT 'pending',
    provider VARCHAR(30) NOT NULL DEFAULT 'yookassa',
    external_id VARCHAR(255),
    refund_amount BIGINT DEFAULT 0,
    refunded_at TIMESTAMPTZ,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
  );
  CREATE INDEX idx_payments_booking_id ON payments(booking_id);
  CREATE INDEX idx_payments_user_id ON payments(user_id);
  CREATE INDEX idx_payments_external_id ON payments(external_id);
  ```
- [x] Написать тесты для валидации модели Payment
- [x] Запустить go test ./... - все тесты должны пройти

### Task 2: Репозиторий платежей

**Files:**
- Modify: `internal/repository/interfaces.go`
- Create: `internal/repository/postgres/payment.go`
- Create: `internal/repository/mock/payment.go`

- [x] Добавить PaymentRepository interface:
  ```go
  type PaymentRepository interface {
    Create(ctx context.Context, payment *domain.Payment) error
    GetByID(ctx context.Context, id uuid.UUID) (*domain.Payment, error)
    GetByBookingID(ctx context.Context, bookingID uuid.UUID) (*domain.Payment, error)
    GetByExternalID(ctx context.Context, externalID string) (*domain.Payment, error)
    UpdateStatus(ctx context.Context, id uuid.UUID, status domain.PaymentStatus, externalID string) error
    UpdateRefund(ctx context.Context, id uuid.UUID, refundAmount int64, refundedAt time.Time) error
    ListByUser(ctx context.Context, userID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.Payment], error)
  }
  ```
- [x] Реализовать postgres-репозиторий
- [x] Реализовать mock-репозиторий для тестов
- [x] Написать тесты для mock-репозитория
- [x] Запустить go test ./... - все тесты должны пройти

### Task 3: Интеграция с ЮKassa

**Files:**
- Create: `internal/payment/yookassa.go`
- Create: `internal/payment/provider.go` (интерфейс)
- Modify: `config/config.go`

- [x] Создать интерфейс PaymentProvider:
  ```go
  type PaymentProvider interface {
    CreatePayment(ctx context.Context, amount int64, currency string, description string, returnURL string, metadata map[string]string) (externalID string, confirmationURL string, err error)
    GetPaymentStatus(ctx context.Context, externalID string) (status string, err error)
    CreateRefund(ctx context.Context, externalID string, amount int64) error
  }
  ```
- [x] Реализовать YooKassaProvider с SDK
- [x] Добавить конфигурацию: BANI_PAYMENT_YOOKASSA_SHOP_ID, BANI_PAYMENT_YOOKASSA_SECRET_KEY, BANI_PAYMENT_RETURN_URL
- [x] Написать тесты с моковым провайдером
- [x] Запустить go test ./... - все тесты должны пройти

### Task 4: Сервис платежей

**Files:**
- Create: `internal/service/payment_service.go`

- [x] Создать PaymentService interface:
  ```go
  type PaymentService interface {
    InitiatePayment(ctx context.Context, bookingID uuid.UUID) (confirmationURL string, err error)
    HandleWebhook(ctx context.Context, event WebhookEvent) error
    RefundPayment(ctx context.Context, bookingID uuid.UUID) error
    GetPaymentByBooking(ctx context.Context, bookingID uuid.UUID) (*domain.Payment, error)
    ListUserPayments(ctx context.Context, userID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.Payment], error)
  }
  ```
- [x] InitiatePayment: создать Payment в БД, вызвать PaymentProvider.CreatePayment, вернуть URL для оплаты
- [x] HandleWebhook: обработать webhook от ЮKassa, обновить статус Payment, при успехе обновить Booking status на confirmed
- [x] RefundPayment: при отмене бронирования создать возврат, если платеж был успешен (полный возврат если > 24ч до визита, без возврата если < 2ч)
- [x] Написать unit-тесты с mock-провайдером и mock-репозиторием
- [x] Запустить go test ./... - все тесты должны пройти

### Task 5: Хендлеры платежей и webhook

**Files:**
- Create: `internal/handler/payment.go`
- Modify: `internal/server/router.go`

- [x] POST /api/v1/bookings/{id}/pay - инициировать оплату бронирования (возвращает confirmation_url)
- [x] POST /api/v1/webhooks/yookassa - webhook от ЮKassa (верификация подписи, обработка события)
- [x] GET /api/v1/my/payments - история платежей текущего пользователя (paginated)
- [x] GET /api/v1/bookings/{id}/payment - статус оплаты бронирования
- [x] Зарегистрировать маршруты в роутере
- [x] Добавить fx.Module для payment-слоя
- [x] Написать handler-тесты с httptest
- [x] Запустить go test ./... - все тесты должны пройти

### Task 6: Интеграция с бронированием

**Files:**
- Modify: `internal/service/booking_service.go`
- Modify: `internal/handler/booking.go`

- [ ] При отмене бронирования автоматически инициировать возврат если платеж был
- [ ] Добавить поле payment_status в ответ booking endpoint
- [ ] Написать тесты интеграции booking + payment
- [ ] Запустить go test ./... - все тесты должны пройти

### Task 7: Верификация

- [ ] Запустить полный тест-сьют: go test ./... -v
- [ ] Запустить линтер: make lint
- [ ] Запустить go vet ./...
- [ ] Проверить миграции (up и down)
- [ ] Обновить CLAUDE.md если появились новые паттерны
- [ ] Переместить этот план в docs/plans/completed/
