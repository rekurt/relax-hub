# Рекомендательная система

## Overview

Персональные рекомендации бань на основе: истории бронирований, избранного, геолокации, предпочтений по удобствам, похожих пользователей (collaborative filtering). Эндпоинт "рекомендации для вас" и блок "похожие бани" на карточке бани.

**Коммерческая ценность:** Увеличивает конверсию на 25-35% за счет персонализации. Пользователи, получающие персональные рекомендации, бронируют в 2 раза чаще. Promoted бани можно продвигать через рекомендации (premium feature).

## Context

- Files involved: internal/domain/, internal/repository/, internal/service/, internal/handler/, migrations/
- Related patterns: Clean architecture, Uber fx DI
- Dependencies: нет новых (чистая Go-логика)
- Существующие данные для рекомендаций: bookings (UserID, BathhouseID), favorites (UserID, BathhouseID), bathhouses (amenities, city, price, rating)

## Development Approach

- **Testing approach**: Regular (code first, then tests)
- Complete each task fully before moving to the next
- **CRITICAL: every task MUST include new/updated tests**
- **CRITICAL: all tests must pass before starting next task**

## Implementation Steps

### Task 1: Сбор пользовательских предпочтений

**Files:**
- Create: `internal/domain/recommendation.go`
- Create: `migrations/000013_user_preferences.up.sql`
- Create: `migrations/000013_user_preferences.down.sql`

- [x] Создать модель UserPreferences:
  ```
  UserPreferences {
    UserID          uuid.UUID
    PreferredCityID *int64
    PriceRangeMin   *int64
    PriceRangeMax   *int64
    PreferPool      bool
    PreferSauna     bool
    PreferSteamRoom bool
    PreferHotTub    bool
    PreferBBQ       bool
    PreferKaraoke   bool
    UpdatedAt       time.Time
  }
  ```
- [x] Создать модель UserActivity (UserID, BathhouseID, Type view/book/favorite, CreatedAt) — для трекинга
- [x] Создать миграции
- [x] Написать тесты
- [x] Запустить go test ./... - все тесты должны пройти

### Task 2: Репозиторий рекомендаций

**Files:**
- Modify: `internal/repository/interfaces.go`
- Create: `internal/repository/postgres/recommendation.go`
- Create: `internal/repository/mock/recommendation.go`

- [x] RecommendationRepository:
  - GetUserPreferences(ctx, userID) — явные предпочтения
  - SaveUserPreferences(ctx, prefs) — сохранить предпочтения
  - RecordActivity(ctx, activity) — записать активность
  - GetUserBookedBathhouses(ctx, userID) — список забронированных бань
  - GetSimilarUsers(ctx, userID, limit) — пользователи с похожими бронированиями
  - GetPopularBathhouses(ctx, cityID, limit) — популярные бани в городе
  - GetSimilarBathhouses(ctx, bathhouseID, limit) — похожие бани (по удобствам, городу, цене)
- [x] Реализовать postgres и mock
- [x] Написать тесты
- [x] Запустить go test ./... - все тесты должны пройти

### Task 3: Сервис рекомендаций

**Files:**
- Create: `internal/service/recommendation_service.go`

- [ ] RecommendationService:
  - GetPersonalized(ctx, userID, page, pageSize) — персональные рекомендации
  - GetSimilar(ctx, bathhouseID, limit) — похожие бани
  - GetPopular(ctx, cityID, limit) — популярные в городе
  - UpdatePreferences(ctx, userID, prefs) — обновить предпочтения
  - RecordView(ctx, userID, bathhouseID) — зафиксировать просмотр
- [ ] Алгоритм персональных рекомендаций:
  1. Взять предпочтения пользователя (явные + из истории)
  2. Найти похожих пользователей (collaborative filtering по бронированиям)
  3. Взять бани, которые бронировали похожие пользователи, но не текущий
  4. Отфильтровать по предпочтениям (город, цена, удобства)
  5. Отсортировать по score (рейтинг * similarity * recency)
- [ ] Написать unit-тесты
- [ ] Запустить go test ./... - все тесты должны пройти

### Task 4: Хендлеры рекомендаций

**Files:**
- Create: `internal/handler/recommendation.go`
- Modify: `internal/server/router.go`
- Modify: `internal/handler/bathhouse.go`

- [ ] GET /api/v1/recommendations — персональные рекомендации (auth)
- [ ] GET /api/v1/bathhouses/{id}/similar — похожие бани (public)
- [ ] GET /api/v1/popular?city_id=1 — популярные бани в городе (public)
- [ ] GET /api/v1/my/preferences — текущие предпочтения
- [ ] PUT /api/v1/my/preferences — обновить предпочтения
- [ ] При просмотре бани (GetByID) — записать активность для авторизованных
- [ ] Зарегистрировать маршруты, добавить fx.Module
- [ ] Написать handler-тесты
- [ ] Запустить go test ./... - все тесты должны пройти

### Task 5: Верификация

- [ ] Запустить полный тест-сьют: go test ./... -v
- [ ] Запустить линтер: make lint
- [ ] Запустить go vet ./...
- [ ] Переместить этот план в docs/plans/completed/
