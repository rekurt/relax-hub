# Расширенные профили пользователей

## Overview

Расширение пользовательских профилей: аватар, bio, предпочтения (тип бани, удобства, ценовой диапазон). История посещений, статистика (количество визитов, средний чек). Публичный профиль с отзывами пользователя.

**Коммерческая ценность:** Повышает вовлеченность и привязку к платформе. Публичные профили с отзывами увеличивают доверие к отзывам. Подробные предпочтения улучшают качество рекомендаций.

## Context

- Files involved: internal/domain/user.go, internal/repository/, internal/service/, internal/handler/, migrations/
- Related patterns: Clean architecture, Uber fx DI
- Dependencies: S3-совместимое хранилище для аватаров (MinIO/AWS S3)
- Текущая модель User: ID, Email, PasswordHash, Name, Phone, Role, IsActive

## Development Approach

- **Testing approach**: Regular (code first, then tests)
- Complete each task fully before moving to the next
- **CRITICAL: every task MUST include new/updated tests**
- **CRITICAL: all tests must pass before starting next task**

## Implementation Steps

### Task 1: Расширение модели пользователя

**Files:**
- Modify: `internal/domain/user.go`
- Create: `migrations/000014_user_profiles.up.sql`
- Create: `migrations/000014_user_profiles.down.sql`

- [x] Расширить модель User:
  ```
  + AvatarURL     string
  + Bio           string
  + CityID        *int64
  ```
- [x] Создать модель UserProfile (агрегат для публичного профиля):
  ```
  UserProfile {
    ID          uuid.UUID
    Name        string
    AvatarURL   string
    Bio         string
    CityName    string
    MemberSince time.Time
    ReviewCount int
    VisitCount  int
    AvgRating   float64   // средний рейтинг оставленных отзывов
  }
  ```
- [x] Создать миграцию: ALTER users ADD COLUMN avatar_url, bio, city_id
- [x] Написать тесты
- [x] Запустить go test ./... - все тесты должны пройти

### Task 2: Загрузка аватаров (File Storage)

**Files:**
- Create: `internal/storage/s3.go`
- Create: `internal/storage/provider.go` (интерфейс)
- Modify: `config/config.go`

- [x] Создать интерфейс FileStorage:
  ```go
  type FileStorage interface {
    Upload(ctx context.Context, filename string, data io.Reader, contentType string) (url string, err error)
    Delete(ctx context.Context, filename string) error
  }
  ```
- [x] Реализовать S3Storage (совместим с MinIO и AWS S3)
- [x] Конфигурация: BANI_STORAGE_ENDPOINT, BANI_STORAGE_BUCKET, BANI_STORAGE_ACCESS_KEY, BANI_STORAGE_SECRET_KEY
- [x] Ресайз аватаров: 200x200 и 50x50 (thumbnail)
- [x] Написать тесты с моковым хранилищем
- [x] Запустить go test ./... - все тесты должны пройти

### Task 3: Расширение сервиса и хендлера пользователей

**Files:**
- Modify: `internal/service/user_service.go`
- Modify: `internal/handler/auth.go`

- [x] Расширить UserService: UpdateProfile, UploadAvatar, GetPublicProfile, GetUserStats
- [x] PUT /api/v1/auth/me — обновить профиль (name, bio, phone, city_id)
- [x] POST /api/v1/auth/me/avatar — загрузить аватар (multipart/form-data)
- [x] DELETE /api/v1/auth/me/avatar — удалить аватар
- [x] GET /api/v1/users/{id}/profile — публичный профиль (имя, аватар, bio, кол-во отзывов)
- [x] Написать тесты
- [x] Запустить go test ./... - все тесты должны пройти

### Task 4: Статистика пользователя

**Files:**
- Modify: `internal/repository/interfaces.go`
- Modify: `internal/repository/postgres/booking.go`
- Modify: `internal/repository/postgres/review.go`

- [ ] Добавить в BookingRepository: GetUserStats(ctx, userID) — количество визитов, средний чек, общая сумма
- [ ] Добавить в ReviewRepository: GetUserReviewStats(ctx, userID) — кол-во отзывов, средний рейтинг
- [ ] GET /api/v1/my/stats — персональная статистика
- [ ] Написать тесты
- [ ] Запустить go test ./... - все тесты должны пройти

### Task 5: Верификация

- [ ] Запустить полный тест-сьют: go test ./... -v
- [ ] Запустить линтер: make lint
- [ ] Запустить go vet ./...
- [ ] Переместить этот план в docs/plans/completed/
