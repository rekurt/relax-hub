# Бани — B2BC маркетплейс бронирования бань

Go-бекенд и React SPA для агрегатора бань с онлайн-бронированием, геопоиском на карте, RBAC с четырьмя ролями, модерацией и полным покрытием API для всех ролей (клиент, владелец, администратор).

## Стек

- **Go 1.25** — язык
- **chi** — HTTP-роутер
- **Cobra + Viper** — CLI и конфигурация
- **Uber fx** — DI-контейнер
- **pgx + PostGIS** — PostgreSQL с гео-запросами
- **go-redis** — Redis
- **golang-migrate** — миграции БД
- **golang-jwt** — JWT-аутентификация
- **bcrypt** — хэширование паролей

### Фронтенд

- **React 18** + **TypeScript 5.6** — UI
- **Vite 6** — сборщик
- **Ant Design 6** — UI-библиотека (русская локализация)
- **TanStack React Query 5** — серверное состояние
- **orval** — генерация API-клиента из OpenAPI
- **zustand** — клиентское состояние

Фронтенд — это мультиролевое SPA с тремя интерфейсами:
- **Клиент** (`/client/*`) — поиск бань, бронирование, отзывы с фото/видео, избранное, рекомендации, программа лояльности, реферальная программа, подарочные сертификаты, история платежей, чат с банями
- **Владелец/представитель** (`/`) — управление банями, бронированиями, ценообразованием, промокодами, фотографиями, чатом, подписками
- **Администратор** (`/admin/*`) — аналитика, модерация бань/отзывов/фото, управление жалобами, городами, пользователями, глобальные промокоды

## Структура проекта

```
cmd/server/          — CLI (main, serve, migrate, seed-admin)
config/              — конфигурация (Viper, config.yaml)
internal/
  app/               — DI-контейнер (fx.App)
  database/          — провайдеры PostgreSQL и Redis
  domain/            — доменные модели, ошибки, фильтры
  handler/           — HTTP-хэндлеры (REST API)
  middleware/        — auth (JWT), RBAC, CORS, logging
  repository/        — интерфейсы и PostgreSQL-реализации
    mock/            — in-memory моки для тестов
    postgres/        — SQL-реализации через pgx
  server/            — HTTP-сервер и роутер (chi)
  service/           — бизнес-логика и RBAC-проверки
migrations/          — SQL-миграции (PostGIS, таблицы, индексы)
frontend/            — React SPA (мультиролевое: клиент, владелец, админ)
  src/api/           — Axios + сгенерированный API-клиент (orval)
  src/components/    — переиспользуемые компоненты (лейауты, карточки, модалы)
  src/pages/admin/   — страницы администратора
  src/pages/client/  — страницы клиента
  src/pages/         — страницы владельца/представителя (bathhouses, bookings, etc.)
  src/stores/        — zustand stores
  src/lib/           — утилиты форматирования
```

## Роли и RBAC

| Роль | Возможности |
|------|-------------|
| **admin** | Модерация бань (approve/reject), управление пользователями (block/unblock), CRUD городов, просмотр всех данных |
| **client** | Поиск бань, бронирование, отмена своих бронирований, отзывы на завершённые бронирования (редактирование в течение 24ч, удаление), избранное |
| **owner** | CRUD своих бань, просмотр бронирований, подтверждение/отклонение бронирований, управление представителями (invite/revoke), ответы на отзывы |
| **representative** | Те же права что у owner, но только для бань, к которым привязан. Не может удалять бани и управлять другими представителями |

Регистрация доступна только как client или owner. Admin создаётся через `seed-admin`. Representative назначается через invite от owner.

## Быстрый старт

### Требования

- Go 1.25+
- PostgreSQL 17 с PostGIS 3.5
- Redis 7
- Docker и Docker Compose (опционально)

### Через Docker Compose

```bash
# Поднять PostgreSQL, Redis и приложение
make docker-up

# API в Docker доступен на http://localhost:28080
# PostgreSQL/Redis/MinIO доступны на localhost:5435, localhost:6381, localhost:9102

# Применить миграции
make migrate-up

# Создать admin-пользователя
make seed-admin
```

### Локально

```bash
# Скопировать конфигурацию
cp .env.example .env

# Собрать
make build

# Применить миграции
make migrate-up

# Создать admin
make seed-admin

# Запустить
make run
```

### Конфигурация

Через `config/config.yaml` или env-переменные с префиксом `BANI_`:

| Переменная | Описание | По умолчанию |
|------------|----------|--------------|
| `BANI_SERVER_HOST` | Хост сервера | `0.0.0.0` |
| `BANI_SERVER_PORT` | Порт сервера | `8080` |
| `BANI_DOCKER_APP_PORT` | Host-порт API при запуске через Docker Compose | `28080` |
| `BANI_DOCKER_POSTGRES_PORT` | Host-порт PostgreSQL при запуске через Docker Compose | `5435` |
| `BANI_DOCKER_REDIS_PORT` | Host-порт Redis при запуске через Docker Compose | `6381` |
| `BANI_DOCKER_MINIO_API_PORT` | Host-порт MinIO API при запуске через Docker Compose | `9102` |
| `BANI_DOCKER_MINIO_CONSOLE_PORT` | Host-порт MinIO Console при запуске через Docker Compose | `9103` |
| `BANI_BACKEND_URL` | URL API для frontend dev proxy | `http://localhost:28080` |
| `BANI_DATABASE_DSN` | PostgreSQL DSN | `postgres://postgres:postgres@localhost:5435/bani?sslmode=disable` |
| `BANI_REDIS_ADDR` | Redis адрес | `localhost:6381` |
| `BANI_JWT_SECRET` | Секрет для JWT | `change-me-in-production` |
| `BANI_JWT_TOKEN_TTL` | Время жизни токена | `24h` |
| `BANI_PAYMENT_YOOKASSA_SHOP_ID` | Shop ID в ЮKassa | `` |
| `BANI_PAYMENT_YOOKASSA_SECRET_KEY` | Секретный ключ ЮKassa | `` |
| `BANI_PAYMENT_RETURN_URL` | URL возврата после оплаты | `http://localhost:3000/payment/callback` |

## API Endpoints

Базовый URL: `/api/v1`

### Аутентификация

| Метод | Путь | Доступ | Описание |
|-------|------|--------|----------|
| POST | `/auth/register` | public | Регистрация (role: client или owner) |
| POST | `/auth/login` | public | Вход, возвращает JWT-токен |
| GET | `/auth/me` | auth | Текущий пользователь |

### Бани

| Метод | Путь | Доступ | Описание |
|-------|------|--------|----------|
| GET | `/bathhouses` | public | Поиск с фильтрами (город, цена, удобства, гео-радиус, кол-во гостей, доступность по дате/времени, открыто сейчас, текстовый поиск). При наличии JWT в ответе добавляется `is_favorite` |
| GET | `/bathhouses/{id}` | public | Детали бани |
| GET | `/bathhouses/{id}/available-slots?date=YYYY-MM-DD` | public | Свободные слоты |
| POST | `/bathhouses` | owner | Создание бани (статус: pending) |
| PUT | `/bathhouses/{id}` | owner, representative | Обновление бани |
| DELETE | `/bathhouses/{id}` | owner | Удаление бани |
| GET | `/my/bathhouses` | owner, representative | Мои бани |

### Бронирования

| Метод | Путь | Доступ | Описание |
|-------|------|--------|----------|
| POST | `/bookings` | client | Создание бронирования |
| GET | `/bookings` | auth | Мои бронирования |
| PATCH | `/bookings/{id}/cancel` | auth | Отмена (клиент за 2ч, owner/rep — всегда) |
| PATCH | `/bookings/{id}/confirm` | owner, representative | Подтверждение |
| PATCH | `/bookings/{id}/reject` | owner, representative | Отклонение |
| PATCH | `/bookings/{id}/complete` | owner, representative | Завершение бронирования |
| GET | `/bathhouses/{id}/bookings` | owner, representative | Бронирования бани |

### Отзывы

| Метод | Путь | Доступ | Описание |
|-------|------|--------|----------|
| POST | `/bathhouses/{id}/reviews` | client | Создание отзыва (только completed booking) |
| GET | `/bathhouses/{id}/reviews` | public | Список отзывов (только одобренные) |
| PUT | `/reviews/{id}` | client | Редактирование отзыва (автор, в течение 24ч) |
| DELETE | `/reviews/{id}` | auth | Удаление отзыва (автор или админ) |
| POST | `/reviews/{id}/response` | owner, representative | Ответ владельца на отзыв |

### Избранное

| Метод | Путь | Доступ | Описание |
|-------|------|--------|----------|
| POST | `/bathhouses/{id}/favorite` | auth | Добавить/убрать из избранного (toggle) |
| GET | `/my/favorites` | auth | Список избранного (paginated) |

### Рекомендации

| Метод | Путь | Доступ | Описание |
|-------|------|--------|----------|
| GET | `/recommendations` | auth | Персональные рекомендации (paginated) |
| GET | `/bathhouses/{id}/similar` | public | Похожие бани (limit, по удобствам/цене/городу) |
| GET | `/popular?city_id={id}` | public | Популярные бани в городе (limit) |
| GET | `/my/preferences` | auth | Текущие предпочтения пользователя |
| PUT | `/my/preferences` | auth | Обновить предпочтения (город, цена, удобства) |

### Представители

| Метод | Путь | Доступ | Описание |
|-------|------|--------|----------|
| POST | `/bathhouses/{id}/representatives` | owner | Пригласить представителя |
| GET | `/bathhouses/{id}/representatives` | owner | Список представителей |
| DELETE | `/representatives/{id}` | owner | Отозвать представителя |

### Динамическое ценообразование

| Метод | Путь | Доступ | Описание |
|-------|------|--------|----------|
| POST | `/my/bathhouses/{id}/pricing-rules` | owner, representative | Создать ценовое правило |
| GET | `/my/bathhouses/{id}/pricing-rules` | owner, representative | Список ценовых правил бани |
| PUT | `/pricing-rules/{id}` | owner, representative | Обновить ценовое правило |
| DELETE | `/pricing-rules/{id}` | owner, representative | Удалить ценовое правило |
| GET | `/bathhouses/{id}/price-calculator?start=...&end=...` | public | Калькулятор цены с учетом правил |

### Города

| Метод | Путь | Доступ | Описание |
|-------|------|--------|----------|
| GET | `/cities` | public | Список городов |

### Платежи

| Метод | Путь | Доступ | Описание |
|-------|------|--------|----------|
| POST | `/bookings/{id}/pay` | auth | Инициировать оплату бронирования |
| GET | `/bookings/{id}/payment` | auth | Статус оплаты бронирования |
| GET | `/my/payments` | auth | История платежей пользователя |
| POST | `/webhooks/yookassa` | public | Webhook от ЮKassa |

### Администрирование

| Метод | Путь | Доступ | Описание |
|-------|------|--------|----------|
| GET | `/admin/users` | admin | Список пользователей |
| PATCH | `/admin/users/{id}/block` | admin | Заблокировать |
| PATCH | `/admin/users/{id}/unblock` | admin | Разблокировать |
| GET | `/admin/bathhouses?status=pending` | admin | Бани на модерации |
| PATCH | `/admin/bathhouses/{id}/approve` | admin | Одобрить баню |
| PATCH | `/admin/bathhouses/{id}/reject` | admin | Отклонить баню |
| POST | `/admin/cities` | admin | Создать город |
| PUT | `/admin/cities/{id}` | admin | Обновить город |
| DELETE | `/admin/cities/{id}` | admin | Удалить город |

### Healthcheck

| Метод | Путь | Описание |
|-------|------|----------|
| GET | `/health` | Статус сервера |

### API документация (Swagger UI)

Интерактивная документация доступна по адресу `/swagger/` при запущенном сервере. Спецификация генерируется из аннотаций в коде с помощью [swaggo/swag](https://github.com/swaggo/swag).

```bash
make swagger       # перегенерировать спецификацию
make swagger-fmt   # форматировать аннотации
```

## Makefile команды

```
make build         — собрать бинарник
make run           — собрать и запустить
make test          — запустить тесты
make vet           — go vet
make lint          — golangci-lint
make migrate-up    — применить миграции
make migrate-down  — откатить миграции
make docker-up     — поднять docker compose
make docker-down   — остановить docker compose
make seed-admin    — создать admin-пользователя
make swagger       — сгенерировать OpenAPI спецификацию
make swagger-fmt   — форматировать Swagger аннотации
make clean         — очистить артефакты сборки
make frontend-dev  — запустить Vite dev-сервер
make frontend-build — production сборка фронтенда
make frontend-generate-api — перегенерировать API-клиент из swagger.json
make frontend-test — запустить тесты фронтенда
```

## Тесты

### Unit и Integration Tests

```bash
make test
```

Тесты покрывают:
- Парсинг конфигурации
- DI-провайдеры
- JWT-аутентификация и RBAC middleware
- Бизнес-логика сервисов с моками репозиториев
- HTTP-хэндлеры через httptest
- Доменные модели и валидация

### Hurl Integration Tests

Полный набор API-тестов с использованием [hurl](https://hurl.dev/):

```bash
make test-hurl
```

Тесты hurl покрывают все API-endpoints:
- **Auth** — регистрация, логин, получение профиля
- **Bathhouses** — поиск, получение деталей, CRUD операции, доступные слоты
- **Bookings** — создание, отмена, подтверждение, отклонение, завершение
- **Reviews** — список отзывов, создание, редактирование, удаление, ответы владельца
- **Favorites** — добавление в избранное, список избранного
- **Pricing** — создание/обновление/удаление ценовых правил, расчет цены с учетом правил
- **Representatives** — приглашение представителя, список, отзыв прав
- **Admin** — управление пользователями, бронированиями, городами
- **Health** — liveness и readiness probes

Hurl-тесты включают как успешные сценарии, так и проверку ошибок (403 Forbidden, 404 Not Found, 409 Conflict, и т.д.).

## Production Deployment

Приложение готово к развёртыванию в продакшене. Для развёртывания смотрите [DEPLOYMENT.md](docs/DEPLOYMENT.md).

### Healthcheck Endpoints

Приложение предоставляет два healthcheck endpoint для оркестрации контейнеров:

| Endpoint | Описание |
|----------|----------|
| `GET /health` | Liveness probe — проверяет, работает ли приложение |
| `GET /ready` | Readiness probe — проверяет доступность БД и Redis |

### Docker

```bash
# Собрать образ
docker build -t bani-api:latest .

# Запустить с Docker Compose
docker-compose up -d
```

Docker образ:
- Использует multi-stage build для минимизации размера
- Включает healthcheck директиву для автоматических перезагрузок
- Запускается от непривилегированного пользователя (app:app)
- Содержит конфиг и миграции

### Переменные окружения для продакшена

Обязательные переменные в production должны быть установлены:
- `BANI_ENVIRONMENT=production`
- `BANI_JWT_SECRET` (минимум 32 символа)
- `BANI_DATABASE_DSN` (с `sslmode=require`)
- `BANI_REDIS_ADDR`

Для подробной информации смотрите `.env.production.example`.
