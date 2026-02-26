# Бани — B2BC маркетплейс бронирования бань

Go-бекенд для агрегатора бань с онлайн-бронированием, геопоиском на карте, RBAC с четырьмя ролями и модерацией.

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
| `BANI_DATABASE_DSN` | PostgreSQL DSN | `postgres://postgres:postgres@localhost:5432/bani?sslmode=disable` |
| `BANI_REDIS_ADDR` | Redis адрес | `localhost:6379` |
| `BANI_JWT_SECRET` | Секрет для JWT | `change-me-in-production` |
| `BANI_JWT_TOKEN_TTL` | Время жизни токена | `24h` |

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

### Представители

| Метод | Путь | Доступ | Описание |
|-------|------|--------|----------|
| POST | `/bathhouses/{id}/representatives` | owner | Пригласить представителя |
| GET | `/bathhouses/{id}/representatives` | owner | Список представителей |
| DELETE | `/representatives/{id}` | owner | Отозвать представителя |

### Города

| Метод | Путь | Доступ | Описание |
|-------|------|--------|----------|
| GET | `/cities` | public | Список городов |

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
make clean         — очистить артефакты сборки
```

## Тесты

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
