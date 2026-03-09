# Фото и видео в отзывах

## Overview

Загрузка фотографий и видео к отзывам. S3-совместимое хранилище, ресайз/оптимизация изображений, модерация медиа. Галерея фото бани от посетителей на карточке бани.

**Коммерческая ценность:** Отзывы с фото получают в 3 раза больше просмотров. Пользовательский контент (UGC) повышает доверие. Галерея реальных фото привлекает новых клиентов лучше, чем фото от владельцев.

## Context

- Files involved: internal/domain/review.go, internal/repository/, internal/service/, internal/handler/, migrations/
- Related patterns: Clean architecture, Uber fx DI
- Dependencies: S3-совместимое хранилище (переиспользовать из user-profiles), imaging library для ресайза
- Текущее: Review.Images []string — уже есть поле для URL изображений, но нет загрузки

## Development Approach

- **Testing approach**: Regular (code first, then tests)
- Complete each task fully before moving to the next
- **CRITICAL: every task MUST include new/updated tests**
- **CRITICAL: all tests must pass before starting next task**

## Implementation Steps

### Task 1: Доменные модели медиа

**Files:**
- Create: `internal/domain/media.go`
- Create: `migrations/000015_review_media.up.sql`
- Create: `migrations/000015_review_media.down.sql`

- [x] Создать модель Media:
  ```
  Media {
    ID            uuid.UUID
    OwnerType     string          // "review", "bathhouse"
    OwnerID       uuid.UUID
    UserID        uuid.UUID
    Type          MediaType       // image, video
    URL           string
    ThumbnailURL  string
    OriginalName  string
    Size          int64           // bytes
    MimeType      string
    Width         int
    Height        int
    Status        MediaStatus     // pending, approved, rejected
    CreatedAt     time.Time
  }
  ```
- [x] Создать миграцию с таблицей media
- [x] Ограничения: макс 10 фото на отзыв, макс 1 видео, макс размер фото 10MB, видео 50MB
- [x] Написать тесты
- [x] Запустить go test ./... - все тесты должны пройти

### Task 2: Сервис загрузки и обработки медиа

**Files:**
- Modify: `internal/repository/interfaces.go`
- Create: `internal/repository/postgres/media.go`
- Create: `internal/repository/mock/media.go`
- Create: `internal/service/media_service.go`

- [x] MediaRepository: Create, GetByID, ListByOwner, Delete, UpdateStatus
- [x] MediaService:
  - Upload(ctx, ownerType, ownerID, userID, file) — загрузить файл, создать thumbnail, сохранить в S3
  - Delete(ctx, mediaID, userID) — удалить (автор или админ)
  - ListByReview(ctx, reviewID) — медиа отзыва
  - ListByBathhouse(ctx, bathhouseID, page, pageSize) — галерея бани (все фото из отзывов)
- [x] Обработка изображений: ресайз до макс 1920px, thumbnail 300x300, JPEG оптимизация
- [x] Валидация: проверка mime type, размера файла
- [x] Написать unit-тесты
- [x] Запустить go test ./... - все тесты должны пройти

### Task 3: Интеграция с отзывами

**Files:**
- Modify: `internal/handler/review.go`
- Modify: `internal/service/review_service.go`

- [ ] POST /api/v1/reviews/{id}/media — загрузить фото/видео к отзыву (multipart/form-data)
- [ ] DELETE /api/v1/media/{id} — удалить медиа
- [ ] В ответе отзыва: включить массив media с URL и thumbnail_url
- [ ] Написать тесты
- [ ] Запустить go test ./... - все тесты должны пройти

### Task 4: Галерея бани от посетителей

**Files:**
- Create: `internal/handler/media.go`
- Modify: `internal/server/router.go`

- [ ] GET /api/v1/bathhouses/{id}/gallery — публичная галерея фото от посетителей (paginated)
- [ ] Включить preview-фото в ответ карточки бани (первые 4 фото из отзывов)
- [ ] Зарегистрировать маршруты, добавить fx.Module
- [ ] Написать handler-тесты
- [ ] Запустить go test ./... - все тесты должны пройти

### Task 5: Верификация

- [ ] Запустить полный тест-сьют: go test ./... -v
- [ ] Запустить линтер: make lint
- [ ] Запустить go vet ./...
- [ ] Переместить этот план в docs/plans/completed/
