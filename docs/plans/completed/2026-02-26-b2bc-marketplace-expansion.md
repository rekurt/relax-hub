# B2BC Marketplace: Расширение функционала (Полный план)

## Overview

Полный план расширения функционала маркетплейса бань. План разбит на 4 фазы. В рамках текущей реализации выполняется только Фаза 1 (Пользовательский опыт). Остальные фазы задокументированы для последующей реализации.

## Фазы (Roadmap)

- **Фаза 1 (РЕАЛИЗУЕМ СЕЙЧАС):** Пользовательский опыт - расширенные фильтры, избранное, улучшенные отзывы
- **Фаза 2 (позже):** Модерация и безопасность - модерация отзывов, система жалоб, админ-панель
- **Фаза 3 (позже):** Монетизация - платные размещения, пакеты подписок, аналитика
- **Фаза 4 (позже):** Продвинутые фичи - рекомендации, уведомления, SEO

---

## ФАЗА 1: Пользовательский опыт (реализуем сейчас)

### Context

- Files involved: internal/domain/, internal/repository/, internal/service/, internal/handler/, migrations/
- Related patterns: Clean architecture handler->service->repository, Uber fx DI, chi router
- Dependencies: нет новых внешних зависимостей

### Development Approach

- **Testing approach**: Regular (code first, then tests)
- Complete each task fully before moving to the next
- **CRITICAL: every task MUST include new/updated tests**
- **CRITICAL: all tests must pass before starting next task**

### Task 1: Расширение модели отзывов - ответы владельцев и модерация

**Files:**
- Modify: `internal/domain/review.go`
- Modify: `internal/domain/errors.go`
- Create: `migrations/000002_reviews_enhancement.up.sql`
- Create: `migrations/000002_reviews_enhancement.down.sql`

- [x] Добавить в модель Review поля: UpdatedAt, Status (pending/approved/rejected/hidden), OwnerResponse, OwnerResponseAt, Images []string
- [x] Добавить ReviewStatus тип с константами
- [x] Добавить ReviewFilter struct (BathhouseID, Status, MinRating, Page, PageSize)
- [x] Добавить domain-ошибки: ErrReviewAlreadyResponded, ErrReviewNotFound
- [x] Создать миграцию: ALTER reviews ADD COLUMN status, owner_response, owner_response_at, updated_at, images
- [x] Написать тесты для валидации новых доменных типов

### Task 2: Репозиторий и сервис отзывов - расширение

**Files:**
- Modify: `internal/repository/interfaces.go`
- Modify: `internal/repository/postgres/review.go`
- Modify: `internal/repository/mock/review.go`
- Modify: `internal/service/review.go`
- Modify: `internal/service/interfaces.go`

- [x] Расширить ReviewRepository: Update, Delete, GetByID, ListByBathhouse(с фильтром), UpdateStatus, AddOwnerResponse
- [x] Реализовать новые методы в postgres и mock репозиториях
- [x] Расширить ReviewService: Update (автор в течение 24ч), Delete (автор или админ), GetByID, AddOwnerResponse (владелец/представитель бани)
- [x] Добавить RBAC-проверки для ответов владельцев через AccessChecker
- [x] Написать unit-тесты для сервиса с mock-репозиторием
- [x] Запустить go test ./... - все тесты должны пройти

### Task 3: Хендлеры отзывов - новые эндпоинты

**Files:**
- Modify: `internal/handler/review.go`
- Modify: `internal/server/router.go`

- [x] PUT /api/v1/reviews/{id} - редактирование отзыва (автор, 24ч)
- [x] DELETE /api/v1/reviews/{id} - удаление (автор или админ)
- [x] POST /api/v1/reviews/{id}/response - ответ владельца
- [x] Зарегистрировать новые маршруты в роутере
- [x] Написать handler-тесты с httptest
- [x] Запустить go test ./... - все тесты должны пройти

### Task 4: Система избранного (Favorites)

**Files:**
- Create: `internal/domain/favorite.go`
- Modify: `internal/repository/interfaces.go`
- Create: `internal/repository/postgres/favorite.go`
- Create: `internal/repository/mock/favorite.go`
- Create: `internal/service/favorite.go`
- Modify: `internal/service/interfaces.go`
- Create: `internal/handler/favorite.go`
- Modify: `internal/server/router.go`
- Create: `migrations/000003_favorites.up.sql`
- Create: `migrations/000003_favorites.down.sql`

- [x] Создать модель Favorite (ID, UserID, BathhouseID, CreatedAt) + миграцию с UNIQUE(user_id, bathhouse_id)
- [x] Создать FavoriteRepository: Add, Remove, ListByUser (paginated), IsFavorite, CountByUser
- [x] Реализовать postgres и mock репозитории
- [x] Создать FavoriteService: Toggle, List, IsFavorite
- [x] Создать FavoriteHandler с эндпоинтами:
  - POST /api/v1/bathhouses/{id}/favorite - добавить/убрать из избранного (toggle)
  - GET /api/v1/my/favorites - список избранного (paginated)
- [x] Добавить fx.Module в DI-контейнер
- [x] Написать тесты для всех слоев
- [x] Запустить go test ./... - все тесты должны пройти

### Task 5: Расширенные фильтры для бань

**Files:**
- Modify: `internal/domain/bathhouse.go`
- Modify: `internal/repository/postgres/bathhouse.go`
- Modify: `internal/handler/bathhouse.go`

- [x] Расширить BathhouseFilter: AvailableDate (*time.Time), AvailableTimeFrom/To (*string), GuestCount (*int), OpenNow (*bool), SearchQuery (*string - полнотекстовый поиск по названию и описанию)
- [x] Реализовать фильтрацию в postgres-репозитории:
  - AvailableDate: JOIN с bookings, исключить занятые слоты
  - OpenNow: сравнение с working_hours текущего дня/времени
  - SearchQuery: ILIKE по name и description
  - GuestCount: max_guests >= filter.GuestCount
- [x] Обновить handler для парсинга новых query-параметров
- [x] Написать тесты для новых фильтров
- [x] Запустить go test ./... - все тесты должны пройти

### Task 6: Флаг "в избранном" в выдаче бань

**Files:**
- Modify: `internal/service/bathhouse.go`
- Modify: `internal/handler/bathhouse.go`

- [x] Добавить в ответ BathhouseResponse поле IsFavorite bool
- [x] При авторизованном запросе GET /api/v1/bathhouses и GET /api/v1/bathhouses/{id} проверять через FavoriteRepository
- [x] Написать тесты
- [x] Запустить go test ./... - все тесты должны пройти

### Task 7: Верификация и финализация Фазы 1

- [x] Запустить полный тест-сьют: go test ./... -v
- [x] Запустить линтер: make lint
- [x] Запустить go vet ./...
- [x] Проверить что все миграции корректны (up и down)
- [x] Обновить CLAUDE.md если появились новые паттерны
- [x] Переместить этот план в `docs/plans/completed/`

---

## ФАЗА 2: Модерация и безопасность (план на будущее, не реализуем сейчас)

### Task 2.1: Система жалоб (Complaints/Reports)

- Модель Complaint: ID, ReporterID, TargetType (review/bathhouse), TargetID, Reason (enum: spam/offensive/fake/other), Description, Status (pending/resolved/dismissed), ResolvedBy, ResolvedAt
- Миграция для таблицы complaints
- CRUD репозиторий, сервис, хендлер
- Эндпоинты:
  - POST /api/v1/reviews/{id}/report - пожаловаться на отзыв
  - POST /api/v1/bathhouses/{id}/report - пожаловаться на баню
  - GET /api/v1/admin/complaints - список жалоб (фильтр по статусу/типу)
  - PATCH /api/v1/admin/complaints/{id}/resolve - решить жалобу
  - PATCH /api/v1/admin/complaints/{id}/dismiss - отклонить жалобу

### Task 2.2: Модерация отзывов (Admin)

- Расширить админ-эндпоинты:
  - GET /api/v1/admin/reviews - список отзывов с фильтрами (статус, рейтинг, дата)
  - PATCH /api/v1/admin/reviews/{id}/approve - одобрить
  - PATCH /api/v1/admin/reviews/{id}/reject - отклонить
  - DELETE /api/v1/admin/reviews/{id} - удалить
- Опционально: автомодерация (стоп-слова, спам-паттерны)

### Task 2.3: Расширенная админ-панель

- GET /api/v1/admin/stats - общая статистика (кол-во бань, пользователей, бронирований, отзывов, жалоб)
- Фильтрация бань по статусу, городу, рейтингу
- История действий модератора (audit log)

---

## ФАЗА 3: Монетизация (план на будущее, не реализуем сейчас)

### Концепция платных размещений

Трехуровневая модель подписок для владельцев бань:

**Free (бесплатно):**
- Базовое размещение в каталоге
- До 5 фото
- Стандартная позиция в выдаче

**Premium (подписка, ежемесячно):**
- До 20 фото
- Бейдж "Проверено"
- Приоритет в выдаче (boost_score +10)
- Статистика просмотров за 30 дней

**Promoted (платное продвижение, за показы/клики):**
- Отдельный блок "Рекомендованные" вверху выдачи
- Настраиваемый бюджет и период
- Таргетинг по городу
- Расширенная аналитика (показы, клики, CTR, конверсии)

### Task 3.1: Модель подписок (Subscription)

- Subscription: ID, BathhouseID, Plan (free/premium), StartDate, EndDate, IsActive, AutoRenew
- Миграция, CRUD
- Интеграция с выдачей бань (boost_score для Premium)

### Task 3.2: Промо-размещения (Promotion)

- Promotion: ID, BathhouseID, BudgetKopecks, SpentKopecks, StartDate, EndDate, TargetCityID, Status, ImpressionCount, ClickCount
- PromotionImpression: для трекинга показов
- Логика отображения: отдельный блок promoted в ответе search

### Task 3.3: Аналитика для владельцев

- BathhouseView: трекинг просмотров (BathhouseID, ViewerID, ViewedAt, Source)
- GET /api/v1/my/bathhouses/{id}/analytics - просмотры, бронирования, конверсии, отзывы за период
- Агрегация по дням/неделям/месяцам

---

## ФАЗА 4: Продвинутые фичи (план на будущее, не реализуем сейчас)

### Task 4.1: Рекомендательная система

- Рекомендации на основе: истории бронирований, избранного, геолокации, похожих пользователей
- GET /api/v1/recommendations - персональные рекомендации

### Task 4.2: Уведомления

- Модель Notification (in-app)
- Типы: новое бронирование, статус бронирования изменен, новый отзыв, ответ на отзыв, жалоба решена
- WebSocket или polling для real-time

### Task 4.3: SEO и публичные страницы

- Slug для бань (ЧПУ-ссылки)
- Мета-данные для поисковиков
- Sitemap генерация
