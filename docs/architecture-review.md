# Архитектурный обзор проекта Bani

Дата: 2026-03-06

## Описание

Результаты анализа архитектуры проекта на соответствие принципам clean architecture (handler → service → repository). Проблемы ранжированы по приоритету.

---

## Critical — нарушения clean architecture

Хендлеры обращаются к репозиториям напрямую, минуя сервисный слой.

### 1. BathhouseHandler → PromotionRepository

- **Файл:** `internal/handler/bathhouse_handler.go`
- **Проблема:** хендлер инжектит `PromotionRepository` напрямую для трекинга показов и кликов промо-кампаний
- **Решение:** добавить методы `RecordImpression` и `RecordClick` в `SubscriptionService` или создать отдельный `PromotionService`

### 2. SubscriptionHandler → PromotionRepository, BathhouseRepository, *service.AccessChecker

- **Файл:** `internal/handler/subscription.go`
- **Проблема:** хендлер инжектит `PromotionRepository`, `BathhouseRepository` и конкретный тип `*service.AccessChecker` вместо интерфейса
- **Решение:** вынести логику промо-кампаний в сервис; BathhouseRepository заменить на вызов BathhouseService; AccessChecker инжектить через интерфейс

### 3. AdminHandler → ReviewRepository

- **Файл:** `internal/handler/admin_handler.go`
- **Проблема:** хендлер использует `ReviewRepository` для модерации, т.к. `ReviewService` не имеет admin-методов
- **Решение:** добавить методы модерации в `ReviewService` (например, `ListPending`, `Approve`, `Reject`)

### 4. WidgetHandler → BathhouseRepository

- **Файл:** `internal/handler/widget.go`
- **Проблема:** хендлер обращается к `BathhouseRepository` напрямую, т.к. `BathhouseService` не имеет метода `GetByAPIKey`
- **Решение:** добавить метод `GetByAPIKey` в `BathhouseService`

### 5. PricingHandler → BathhouseRepository

- **Файл:** `internal/handler/pricing.go`
- **Проблема:** хендлер обращается к `BathhouseRepository` для получения base price
- **Решение:** добавить метод получения base price в `PricingService` или `BathhouseService`

### 6. RecommendationHandler → BathhouseRepository

- **Файл:** `internal/handler/recommendation_handler.go`
- **Проблема:** `RecommendationService` возвращает `[]uuid.UUID` вместо объектов, хендлер сам обращается к `BathhouseRepository` для загрузки данных
- **Решение:** изменить сервис чтобы возвращал `[]domain.Bathhouse` (с пагинацией), инкапсулируя загрузку данных

---

## Medium — дублирование, N+1, именование

### 7. Дублирование валидации времени

- **Файлы:** `internal/domain/pricing.go`, `internal/domain/bathhouse.go`
- **Проблема:** `isValidTimeFormat` и `IsValidTimeFormat` — дублированные функции валидации HH:MM
- **Решение:** оставить одну экспортированную функцию `IsValidTimeFormat` в `domain/bathhouse.go` (или в отдельном файле `domain/time.go`), использовать из обоих мест

### 8. Двойная авторизация в SubscriptionHandler

- **Файлы:** `internal/handler/subscription.go`, `internal/service/subscription_service.go`
- **Проблема:** и хендлер, и сервис вызывают `CanManageBathhouse` — избыточная проверка
- **Решение:** оставить проверку только на уровне сервиса (canonical place for RBAC checks)

### 9. N+1 запросы в RecommendationService и RecommendationHandler

- **Файлы:** `internal/service/recommendation_service.go`, `internal/handler/recommendation_handler.go`
- **Проблема:** для каждой рекомендации выполняется отдельный запрос к БД
- **Решение:** использовать batch-запрос (`WHERE id = ANY($1)`) для загрузки нескольких бань за один SQL-запрос

### 10. Непоследовательное именование AccessChecker

- **Файлы:** множество сервисов
- **Проблема:** в одних сервисах поле называется `accessChecker`, в других — `access`
- **Решение:** унифицировать как `accessChecker` во всех сервисах

### 11. Именование файла recommendation.go

- **Файл:** `internal/repository/postgres/recommendation.go`
- **Проблема:** файл без суффикса `_repo`, хотя остальные репозитории следуют паттерну `<entity>_repo.go`
- **Решение:** переименовать в `recommendation_repo.go`

### 12. Опечатка GetActiveBybathhouse

- **Файлы:** интерфейс в `repository/interfaces.go`, реализация в `postgres/`, mock, сервис
- **Проблема:** маленькая `b` в `Bybathhouse` — нарушение Go naming convention
- **Решение:** переименовать в `GetActiveByBathhouse` во всех слоях одновременно

### 13. Пропуски в нумерации миграций

- **Директория:** `migrations/`
- **Проблема:** отсутствуют миграции с номерами 000010-000012, 000014-000016
- **Решение:** не критично для работы, но при следующих миграциях стоит продолжать с последнего номера без пропусков

---

## Low — мелкие улучшения

### 14. Модели без Validate()

- **Файлы:** `internal/domain/favorite.go`, `internal/domain/oauth.go`, `internal/domain/representative.go`
- **Проблема:** модели `Favorite`, `SocialAccount`, `Representative` не имеют метода `Validate()`, хотя остальные domain-модели его реализуют
- **Решение:** добавить `Validate()` с базовыми проверками (например, непустой UserID, BathhouseID)

### 15. RecommendationService без logger

- **Файл:** `internal/service/recommendation_service.go`
- **Проблема:** сервис не инжектит logger, хотя все остальные сервисы его используют
- **Решение:** добавить logger в конструктор и использовать для логирования ошибок

### 16. User endpoints в AuthHandler

- **Файл:** `internal/handler/auth_handler.go`
- **Проблема:** endpoints работы с профилем пользователя (GET/PUT /me) живут в AuthHandler
- **Решение:** создать отдельный `UserHandler` для /me endpoints, оставив в AuthHandler только login/register/oauth
