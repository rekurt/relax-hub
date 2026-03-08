# Верификация фото бань

## Overview

Система проверки подлинности фото бань администратором. Бейдж "Фото проверены" на карточке бани. Требования к качеству фото, водяные знаки платформы, запрет стоковых фото. Мотивация владельцев загружать реальные фото.

**Коммерческая ценность:** Верифицированные фото увеличивают доверие и конверсию на 25%. Бейдж "Проверено" — дополнительная ценность Premium-подписки. Защита от обмана клиентов нереальными фотографиями.

## Context

- Files involved: internal/domain/bathhouse.go, internal/repository/, internal/service/, internal/handler/, migrations/
- Related patterns: Clean architecture, Uber fx DI
- Dependencies: S3 хранилище (переиспользовать из user-profiles/review-media)
- Текущее: Bathhouse.Images []string — массив URL фотографий, без проверки подлинности

## Development Approach

- **Testing approach**: Regular (code first, then tests)
- Complete each task fully before moving to the next
- **CRITICAL: every task MUST include new/updated tests**
- **CRITICAL: all tests must pass before starting next task**

## Implementation Steps

### Task 1: Доменные модели верификации фото

**Files:**
- Create: `internal/domain/photo_verification.go`
- Create: `migrations/000021_photo_verification.up.sql`
- Create: `migrations/000021_photo_verification.down.sql`

- [x] Создать модель BathhousePhoto:
  ```
  BathhousePhoto {
    ID              uuid.UUID
    BathhouseID     uuid.UUID
    URL             string
    ThumbnailURL    string
    Position        int               // порядок отображения
    Status          PhotoStatus       // pending, verified, rejected
    VerifiedByID    *uuid.UUID        // какой админ проверил
    VerifiedAt      *time.Time
    RejectionReason string
    UploadedAt      time.Time
  }
  ```
- [x] Добавить поле IsPhotoVerified bool в Bathhouse
- [x] Создать миграцию: таблица bathhouse_photos + ALTER bathhouses ADD is_photo_verified
- [x] Написать тесты
- [x] Запустить go test ./... - все тесты должны пройти

### Task 2: Репозиторий и сервис верификации

**Files:**
- Modify: `internal/repository/interfaces.go`
- Create: `internal/repository/postgres/photo_verification.go`
- Create: `internal/repository/mock/photo_verification.go`
- Create: `internal/service/photo_verification_service.go`

- [x] BathhousePhotoRepository: Create, Delete, ListByBathhouse, UpdateStatus, Reorder
- [x] PhotoVerificationService:
  - UploadPhoto(ctx, bathhouseID, file) — загрузить фото (owner/representative), создать thumbnail
  - DeletePhoto(ctx, photoID, userID) — удалить фото (owner/representative или admin)
  - ReorderPhotos(ctx, bathhouseID, photoIDs) — изменить порядок
  - VerifyPhoto(ctx, photoID, adminID) — одобрить фото (admin)
  - RejectPhoto(ctx, photoID, adminID, reason) — отклонить фото (admin)
  - GetPendingPhotos(ctx, page, pageSize) — очередь на проверку (admin)
- [x] При верификации всех фото бани — установить IsPhotoVerified = true
- [x] RBAC: загрузка — owner/representative, верификация — admin
- [x] Написать unit-тесты
- [x] Запустить go test ./... - все тесты должны пройти

### Task 3: Хендлеры управления фотографиями

**Files:**
- Create: `internal/handler/photo.go`
- Modify: `internal/server/router.go`

- [x] POST /api/v1/my/bathhouses/{id}/photos — загрузить фото (multipart, owner/rep)
- [x] DELETE /api/v1/photos/{id} — удалить фото (owner/rep/admin)
- [x] PUT /api/v1/my/bathhouses/{id}/photos/reorder — изменить порядок (owner/rep)
- [x] GET /api/v1/admin/photos/pending — очередь на проверку (admin)
- [x] PATCH /api/v1/admin/photos/{id}/verify — одобрить (admin)
- [x] PATCH /api/v1/admin/photos/{id}/reject — отклонить (admin)
- [x] Зарегистрировать маршруты, добавить fx.Module
- [x] Написать handler-тесты
- [x] Запустить go test ./... - все тесты должны пройти

### Task 4: Бейдж "Фото проверены" в выдаче

**Files:**
- Modify: `internal/handler/bathhouse.go`

- [x] Включить поле is_photo_verified в ответ карточки бани
- [x] В выдаче бань: отображать бейдж если IsPhotoVerified = true
- [x] Написать тесты
- [x] Запустить go test ./... - все тесты должны пройти

### Task 5: Верификация

- [x] Запустить полный тест-сьют: go test ./... -v
- [x] Запустить линтер: make lint
- [x] Запустить go vet ./...
- [x] Переместить этот план в docs/plans/completed/
