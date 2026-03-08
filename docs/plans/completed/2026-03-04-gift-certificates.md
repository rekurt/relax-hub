# Подарочные сертификаты

## Overview

Покупка подарочных сертификатов на определенную сумму. Уникальный код, возможность подарить другому пользователю, применение при бронировании, частичное использование, срок действия. Сертификат можно купить без регистрации и отправить по email.

**Коммерческая ценность:** Дополнительный канал привлечения. Средний чек сертификата выше обычного бронирования на 30-50%. Привлечение новых пользователей через получателей подарков.

## Context

- Files involved: internal/domain/, internal/repository/, internal/service/, internal/handler/, migrations/
- Related patterns: Clean architecture, Uber fx DI
- Dependencies: зависит от online-payments для покупки сертификатов
- Связь: применяется при бронировании как способ оплаты

## Development Approach

- **Testing approach**: Regular (code first, then tests)
- Complete each task fully before moving to the next
- **CRITICAL: every task MUST include new/updated tests**
- **CRITICAL: all tests must pass before starting next task**

## Implementation Steps

### Task 1: Доменные модели сертификатов

**Files:**
- Create: `internal/domain/certificate.go`
- Modify: `internal/domain/errors.go`
- Create: `migrations/000007_gift_certificates.up.sql`
- Create: `migrations/000007_gift_certificates.down.sql`

- [x] Создать модель GiftCertificate:
  ```
  GiftCertificate {
    ID              uuid.UUID
    Code            string          // уникальный код
    PurchaserID     *uuid.UUID      // nil если без регистрации
    PurchaserEmail  string
    RecipientEmail  string
    RecipientName   string
    Amount          int64           // номинал в копейках
    Balance         int64           // остаток в копейках
    Message         string          // поздравительное сообщение
    Status          CertificateStatus // active, used, expired
    ValidUntil      time.Time
    RedeemedByID    *uuid.UUID      // кто привязал к аккаунту
    CreatedAt       time.Time
  }
  ```
- [x] Создать модель CertificateUsage (ID, CertificateID, BookingID, Amount, UsedAt)
- [x] Добавить domain-ошибки: ErrCertificateNotFound, ErrCertificateExpired, ErrCertificateInsufficientBalance
- [x] Создать миграцию
- [x] Написать тесты
- [x] Запустить go test ./... - все тесты должны пройти

### Task 2: Репозиторий сертификатов

**Files:**
- Modify: `internal/repository/interfaces.go`
- Create: `internal/repository/postgres/certificate.go`
- Create: `internal/repository/mock/certificate.go`

- [x] GiftCertificateRepository interface: Create, GetByID, GetByCode, UpdateBalance, ListByUser, Redeem
- [x] Реализовать postgres и mock репозитории
- [x] Написать тесты
- [x] Запустить go test ./... - все тесты должны пройти

### Task 3: Сервис сертификатов

**Files:**
- Create: `internal/service/certificate_service.go`

- [x] CertificateService:
  - Purchase(ctx, amount, purchaserEmail, recipientEmail, message) — купить сертификат, инициировать оплату
  - Redeem(ctx, code, userID) — привязать сертификат к аккаунту
  - Apply(ctx, certificateID, bookingID, amount) — применить к бронированию (частично или полностью)
  - GetBalance(ctx, code) — проверить остаток
  - ListByUser(ctx, userID) — сертификаты пользователя
- [x] Генерация уникального кода (формат BANI-XXXX-XXXX)
- [x] Написать unit-тесты
- [x] Запустить go test ./... - все тесты должны пройти

### Task 4: Хендлеры сертификатов

**Files:**
- Create: `internal/handler/certificate.go`
- Modify: `internal/server/router.go`

- [x] POST /api/v1/certificates/purchase — купить сертификат (может быть без auth)
- [x] POST /api/v1/certificates/redeem — привязать сертификат к аккаунту
- [x] GET /api/v1/certificates/{code}/balance — проверить остаток
- [x] GET /api/v1/my/certificates — мои сертификаты
- [x] Зарегистрировать маршруты, добавить fx.Module
- [x] Написать handler-тесты
- [x] Запустить go test ./... - все тесты должны пройти

### Task 5: Верификация

- [x] Запустить полный тест-сьют: go test ./... -v
- [x] Запустить линтер: make lint
- [x] Запустить go vet ./...
- [x] Переместить этот план в docs/plans/completed/
