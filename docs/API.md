# API Documentation

Полная документация REST API B2BC маркетплейса бань.

**Base URL:** `/api/v1`

## Содержание

- [Authentication](#authentication)
- [Bathhouses](#bathhouses)
- [Bookings](#bookings)
- [Reviews](#reviews)
- [Favorites](#favorites)
- [Representatives](#representatives)
- [Cities](#cities)
- [Admin](#admin)
- [Health Checks](#health-checks)
- [Response Format](#response-format)
- [Error Codes](#error-codes)

## Response Format

Все endpoints возвращают ответ в следующем формате:

```json
{
  "success": true,
  "data": {},
  "error": null,
  "meta": {}
}
```

### Success Response

```json
{
  "success": true,
  "data": {
    "id": "uuid",
    "name": "string"
  },
  "error": null,
  "meta": {
    "timestamp": "2026-02-26T21:00:00Z"
  }
}
```

### Error Response

```json
{
  "success": false,
  "data": null,
  "error": {
    "code": "NOT_FOUND",
    "message": "Bathhouse not found"
  },
  "meta": {
    "timestamp": "2026-02-26T21:00:00Z"
  }
}
```

### Paginated Response

```json
{
  "success": true,
  "data": [
    { "id": "uuid", "name": "string" }
  ],
  "error": null,
  "meta": {
    "page": 1,
    "page_size": 20,
    "total_count": 100,
    "total_pages": 5,
    "timestamp": "2026-02-26T21:00:00Z"
  }
}
```

## Authentication

### Register

Создание нового пользователя.

**Endpoint:** `POST /auth/register`

**Access:** Public

**Request Body:**

```json
{
  "email": "user@example.com",
  "password": "securePassword123!",
  "role": "client",
  "phone": "+79991234567"
}
```

**Response:** `200 OK`

```json
{
  "success": true,
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "email": "user@example.com",
    "role": "client",
    "phone": "+79991234567",
    "created_at": "2026-02-26T21:00:00Z"
  },
  "error": null,
  "meta": {}
}
```

**Error Responses:**
- `400 Bad Request` - Invalid input
- `409 Conflict` - User already exists

---

### Login

Вход в систему и получение JWT токена.

**Endpoint:** `POST /auth/login`

**Access:** Public

**Request Body:**

```json
{
  "email": "user@example.com",
  "password": "securePassword123!"
}
```

**Response:** `200 OK`

```json
{
  "success": true,
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "token_type": "Bearer",
    "expires_in": 86400,
    "user": {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "email": "user@example.com",
      "role": "client"
    }
  },
  "error": null,
  "meta": {}
}
```

**Error Responses:**
- `401 Unauthorized` - Invalid credentials
- `403 Forbidden` - User is blocked

---

### Get Current User

Получение информации о текущем пользователе.

**Endpoint:** `GET /auth/me`

**Access:** Authenticated

**Headers:**
```
Authorization: Bearer {token}
```

**Response:** `200 OK`

```json
{
  "success": true,
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "email": "user@example.com",
    "role": "client",
    "phone": "+79991234567",
    "is_blocked": false,
    "created_at": "2026-02-26T21:00:00Z"
  },
  "error": null,
  "meta": {}
}
```

---

## Bathhouses

### List Bathhouses

Поиск бань с фильтрами.

**Endpoint:** `GET /bathhouses`

**Access:** Public

**Query Parameters:**

```
page=1                          # Номер страницы (по умолчанию: 1)
page_size=20                    # Размер страницы (по умолчанию: 20)
city_id=uuid                    # Фильтр по городу
min_price=1000                  # Минимальная цена (в копейках)
max_price=100000                # Максимальная цена (в копейках)
amenities=wifi,pool,sauna       # Фильтр по удобствам (запятая-разделённые)
latitude=55.75                  # Широта для геопоиска
longitude=37.61                 # Долгота для геопоиска
radius=10000                    # Радиус поиска (в метрах)
guest_count=4                   # Количество гостей
available_date=2026-03-01       # Дата доступности (YYYY-MM-DD)
available_time=18:00            # Время доступности (HH:MM)
open_now=true                   # Показать только открытые сейчас
search=sauna                    # Текстовый поиск по названию и описанию
```

**Example:**

```bash
curl "https://api.example.com/api/v1/bathhouses?city_id=550e8400-e29b-41d4-a716-446655440000&min_price=5000&page=1"
```

**Response:** `200 OK`

```json
{
  "success": true,
  "data": [
    {
      "id": "650e8400-e29b-41d4-a716-446655440000",
      "name": "Русская Баня",
      "description": "Классическая русская баня в центре города",
      "city_id": "550e8400-e29b-41d4-a716-446655440000",
      "city": {
        "id": "550e8400-e29b-41d4-a716-446655440000",
        "name": "Москва"
      },
      "owner_id": "750e8400-e29b-41d4-a716-446655440000",
      "price_per_hour": 5000,
      "max_guests": 10,
      "latitude": 55.7558,
      "longitude": 37.6173,
      "address": "ул. Петровка, 25",
      "amenities": ["wifi", "pool", "sauna"],
      "rating": 4.5,
      "review_count": 42,
      "status": "approved",
      "is_favorite": true,
      "created_at": "2026-02-26T21:00:00Z",
      "updated_at": "2026-02-26T21:00:00Z"
    }
  ],
  "error": null,
  "meta": {
    "page": 1,
    "page_size": 20,
    "total_count": 100,
    "total_pages": 5
  }
}
```

---

### Get Bathhouse Details

Получение детальной информации о бане.

**Endpoint:** `GET /bathhouses/{id}`

**Access:** Public

**Path Parameters:**
- `id` - UUID бани

**Response:** `200 OK`

```json
{
  "success": true,
  "data": {
    "id": "650e8400-e29b-41d4-a716-446655440000",
    "name": "Русская Баня",
    "description": "Классическая русская баня в центре города",
    "city_id": "550e8400-e29b-41d4-a716-446655440000",
    "city": {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "name": "Москва"
    },
    "owner_id": "750e8400-e29b-41d4-a716-446655440000",
    "owner": {
      "id": "750e8400-e29b-41d4-a716-446655440000",
      "email": "owner@example.com",
      "phone": "+79991234567"
    },
    "price_per_hour": 5000,
    "max_guests": 10,
    "latitude": 55.7558,
    "longitude": 37.6173,
    "address": "ул. Петровка, 25",
    "phone": "+79991234567",
    "website": "https://example.com",
    "working_hours": {
      "monday": { "open": "10:00", "close": "23:00" },
      "tuesday": { "open": "10:00", "close": "23:00" },
      "wednesday": { "open": "10:00", "close": "23:00" },
      "thursday": { "open": "10:00", "close": "23:00" },
      "friday": { "open": "10:00", "close": "23:00" },
      "saturday": { "open": "09:00", "close": "24:00" },
      "sunday": { "open": "09:00", "close": "24:00" }
    },
    "amenities": ["wifi", "pool", "sauna", "massage"],
    "rating": 4.5,
    "review_count": 42,
    "status": "approved",
    "is_favorite": true,
    "created_at": "2026-02-26T21:00:00Z",
    "updated_at": "2026-02-26T21:00:00Z"
  },
  "error": null,
  "meta": {}
}
```

---

### Get Available Slots

Получение свободных временных слотов для бани.

**Endpoint:** `GET /bathhouses/{id}/available-slots`

**Access:** Public

**Query Parameters:**
```
date=2026-03-01        # Дата в формате YYYY-MM-DD
```

**Example:**

```bash
curl "https://api.example.com/api/v1/bathhouses/650e8400-e29b-41d4-a716-446655440000/available-slots?date=2026-03-01"
```

**Response:** `200 OK`

```json
{
  "success": true,
  "data": {
    "date": "2026-03-01",
    "slots": [
      {
        "start_time": "10:00",
        "end_time": "11:00",
        "is_available": true,
        "price": 5000
      },
      {
        "start_time": "11:00",
        "end_time": "12:00",
        "is_available": true,
        "price": 5000
      },
      {
        "start_time": "12:00",
        "end_time": "13:00",
        "is_available": false,
        "price": 5000
      }
    ]
  },
  "error": null,
  "meta": {}
}
```

---

### Create Bathhouse

Создание новой бани (только для owner).

**Endpoint:** `POST /bathhouses`

**Access:** owner

**Headers:**
```
Authorization: Bearer {token}
Content-Type: application/json
```

**Request Body:**

```json
{
  "name": "Русская Баня",
  "description": "Классическая русская баня в центре города",
  "city_id": "550e8400-e29b-41d4-a716-446655440000",
  "price_per_hour": 5000,
  "max_guests": 10,
  "latitude": 55.7558,
  "longitude": 37.6173,
  "address": "ул. Петровка, 25",
  "phone": "+79991234567",
  "website": "https://example.com",
  "working_hours": {
    "monday": { "open": "10:00", "close": "23:00" },
    "tuesday": { "open": "10:00", "close": "23:00" },
    "wednesday": { "open": "10:00", "close": "23:00" },
    "thursday": { "open": "10:00", "close": "23:00" },
    "friday": { "open": "10:00", "close": "23:00" },
    "saturday": { "open": "09:00", "close": "24:00" },
    "sunday": { "open": "09:00", "close": "24:00" }
  },
  "amenities": ["wifi", "pool", "sauna", "massage"]
}
```

**Response:** `201 Created`

```json
{
  "success": true,
  "data": {
    "id": "650e8400-e29b-41d4-a716-446655440000",
    "name": "Русская Баня",
    "status": "pending",
    "created_at": "2026-02-26T21:00:00Z"
  },
  "error": null,
  "meta": {}
}
```

---

### Update Bathhouse

Обновление информации о бане.

**Endpoint:** `PUT /bathhouses/{id}`

**Access:** owner, representative

**Headers:**
```
Authorization: Bearer {token}
Content-Type: application/json
```

**Request Body:** (аналогично Create)

**Response:** `200 OK`

---

### Delete Bathhouse

Удаление бани.

**Endpoint:** `DELETE /bathhouses/{id}`

**Access:** owner

**Response:** `204 No Content`

---

### Get My Bathhouses

Получение списка моих бань.

**Endpoint:** `GET /my/bathhouses`

**Access:** owner, representative

**Query Parameters:**
```
page=1
page_size=20
status=approved           # Фильтр по статусу (pending, approved, rejected)
```

**Response:** `200 OK` (аналогично List Bathhouses)

---

## Bookings

### Create Booking

Создание бронирования.

**Endpoint:** `POST /bookings`

**Access:** client

**Headers:**
```
Authorization: Bearer {token}
Content-Type: application/json
```

**Request Body:**

```json
{
  "bathhouse_id": "650e8400-e29b-41d4-a716-446655440000",
  "start_time": "2026-03-01T18:00:00Z",
  "end_time": "2026-03-01T19:00:00Z",
  "guest_count": 6,
  "special_requests": "Пожалуйста, подготовьте ароматерапию"
}
```

**Response:** `201 Created`

```json
{
  "success": true,
  "data": {
    "id": "850e8400-e29b-41d4-a716-446655440000",
    "bathhouse_id": "650e8400-e29b-41d4-a716-446655440000",
    "client_id": "550e8400-e29b-41d4-a716-446655440000",
    "status": "pending",
    "start_time": "2026-03-01T18:00:00Z",
    "end_time": "2026-03-01T19:00:00Z",
    "guest_count": 6,
    "total_price": 5000,
    "created_at": "2026-02-26T21:00:00Z"
  },
  "error": null,
  "meta": {}
}
```

---

### Get My Bookings

Получение моих бронирований.

**Endpoint:** `GET /bookings`

**Access:** Authenticated

**Query Parameters:**
```
page=1
page_size=20
status=pending            # Фильтр по статусу (pending, confirmed, rejected, completed, cancelled)
from_date=2026-03-01      # Начальная дата
to_date=2026-03-31        # Конечная дата
```

**Response:** `200 OK` (paginated list of bookings)

---

### Cancel Booking

Отмена бронирования.

**Endpoint:** `PATCH /bookings/{id}/cancel`

**Access:** client (за 2ч), owner/representative (всегда)

**Headers:**
```
Authorization: Bearer {token}
```

**Request Body:**

```json
{
  "reason": "Изменили планы"
}
```

**Response:** `200 OK`

---

### Confirm Booking

Подтверждение бронирования.

**Endpoint:** `PATCH /bookings/{id}/confirm`

**Access:** owner, representative

**Response:** `200 OK`

---

### Reject Booking

Отклонение бронирования.

**Endpoint:** `PATCH /bookings/{id}/reject`

**Access:** owner, representative

**Request Body:**

```json
{
  "reason": "К сожалению, в это время не доступны"
}
```

**Response:** `200 OK`

---

### Complete Booking

Завершение бронирования.

**Endpoint:** `PATCH /bookings/{id}/complete`

**Access:** owner, representative

**Response:** `200 OK`

---

### Get Bathhouse Bookings

Получение бронирований бани.

**Endpoint:** `GET /bathhouses/{id}/bookings`

**Access:** owner, representative

**Query Parameters:**
```
page=1
page_size=20
status=confirmed
from_date=2026-03-01
to_date=2026-03-31
```

**Response:** `200 OK` (paginated list)

---

## Reviews

### Create Review

Создание отзыва.

**Endpoint:** `POST /bathhouses/{id}/reviews`

**Access:** client

**Headers:**
```
Authorization: Bearer {token}
```

**Request Body:**

```json
{
  "booking_id": "850e8400-e29b-41d4-a716-446655440000",
  "rating": 5,
  "text": "Отличная баня! Всё очень чистое и уютно.",
  "would_recommend": true
}
```

**Response:** `201 Created`

```json
{
  "success": true,
  "data": {
    "id": "950e8400-e29b-41d4-a716-446655440000",
    "bathhouse_id": "650e8400-e29b-41d4-a716-446655440000",
    "author_id": "550e8400-e29b-41d4-a716-446655440000",
    "rating": 5,
    "text": "Отличная баня! Всё очень чистое и уютно.",
    "would_recommend": true,
    "created_at": "2026-02-26T21:00:00Z"
  },
  "error": null,
  "meta": {}
}
```

---

### Get Bathhouse Reviews

Получение отзывов бани (только одобренные).

**Endpoint:** `GET /bathhouses/{id}/reviews`

**Access:** Public

**Query Parameters:**
```
page=1
page_size=20
sort_by=rating           # rating или date
```

**Response:** `200 OK` (paginated list)

---

### Update Review

Редактирование отзыва (автор, в течение 24ч).

**Endpoint:** `PUT /reviews/{id}`

**Access:** author

**Request Body:**

```json
{
  "rating": 4,
  "text": "Хорошая баня, но немного дорого"
}
```

**Response:** `200 OK`

---

### Delete Review

Удаление отзыва.

**Endpoint:** `DELETE /reviews/{id}`

**Access:** author, admin

**Response:** `204 No Content`

---

### Add Owner Response

Ответ владельца на отзыв.

**Endpoint:** `POST /reviews/{id}/response`

**Access:** owner, representative

**Request Body:**

```json
{
  "text": "Спасибо за ваш отзыв! Мы рады, что вам понравилось."
}
```

**Response:** `200 OK`

```json
{
  "success": true,
  "data": {
    "id": "950e8400-e29b-41d4-a716-446655440000",
    "owner_response": {
      "text": "Спасибо за ваш отзыв! Мы рады, что вам понравилось.",
      "created_at": "2026-02-26T21:05:00Z"
    }
  },
  "error": null,
  "meta": {}
}
```

---

## Favorites

### Add/Remove Favorite

Toggle добавления/удаления из избранного.

**Endpoint:** `POST /bathhouses/{id}/favorite`

**Access:** Authenticated

**Headers:**
```
Authorization: Bearer {token}
```

**Response:** `200 OK`

```json
{
  "success": true,
  "data": {
    "bathhouse_id": "650e8400-e29b-41d4-a716-446655440000",
    "is_favorite": true
  },
  "error": null,
  "meta": {}
}
```

---

### Get My Favorites

Получение списка избранных бань.

**Endpoint:** `GET /my/favorites`

**Access:** Authenticated

**Query Parameters:**
```
page=1
page_size=20
```

**Response:** `200 OK` (paginated list of bathhouses)

---

## Representatives

### Invite Representative

Приглашение представителя.

**Endpoint:** `POST /bathhouses/{id}/representatives`

**Access:** owner

**Request Body:**

```json
{
  "email": "representative@example.com"
}
```

**Response:** `201 Created`

---

### Get Bathhouse Representatives

Получение списка представителей.

**Endpoint:** `GET /bathhouses/{id}/representatives`

**Access:** owner

**Response:** `200 OK`

```json
{
  "success": true,
  "data": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "email": "representative@example.com",
      "phone": "+79991234567",
      "invited_at": "2026-02-26T21:00:00Z"
    }
  ],
  "error": null,
  "meta": {}
}
```

---

### Revoke Representative

Отозвание представителя.

**Endpoint:** `DELETE /representatives/{id}`

**Access:** owner

**Response:** `204 No Content`

---

## Cities

### Get Cities

Получение списка городов.

**Endpoint:** `GET /cities`

**Access:** Public

**Query Parameters:**
```
page=1
page_size=20
search=Москва        # Текстовый поиск
```

**Response:** `200 OK`

```json
{
  "success": true,
  "data": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "name": "Москва",
      "country": "Россия",
      "latitude": 55.7558,
      "longitude": 37.6173,
      "timezone": "Europe/Moscow",
      "created_at": "2026-02-26T21:00:00Z"
    }
  ],
  "error": null,
  "meta": {
    "page": 1,
    "page_size": 20,
    "total_count": 50,
    "total_pages": 3
  }
}
```

---

## Admin

### Get Users

Получение списка пользователей.

**Endpoint:** `GET /admin/users`

**Access:** admin

**Query Parameters:**
```
page=1
page_size=20
role=client               # Фильтр по роли
is_blocked=false          # Фильтр по статусу блокировки
```

**Response:** `200 OK`

---

### Block User

Блокировка пользователя.

**Endpoint:** `PATCH /admin/users/{id}/block`

**Access:** admin

**Request Body:**

```json
{
  "reason": "Нарушение правил сервиса"
}
```

**Response:** `200 OK`

---

### Unblock User

Разблокировка пользователя.

**Endpoint:** `PATCH /admin/users/{id}/unblock`

**Access:** admin

**Response:** `200 OK`

---

### Get Bathhouses for Moderation

Получение бань на модерации.

**Endpoint:** `GET /admin/bathhouses`

**Access:** admin

**Query Parameters:**
```
status=pending
page=1
page_size=20
```

**Response:** `200 OK`

---

### Approve Bathhouse

Одобрение бани.

**Endpoint:** `PATCH /admin/bathhouses/{id}/approve`

**Access:** admin

**Response:** `200 OK`

---

### Reject Bathhouse

Отклонение бани.

**Endpoint:** `PATCH /admin/bathhouses/{id}/reject`

**Access:** admin

**Request Body:**

```json
{
  "reason": "Неполная информация о бане"
}
```

**Response:** `200 OK`

---

### Create City

Создание города.

**Endpoint:** `POST /admin/cities`

**Access:** admin

**Request Body:**

```json
{
  "name": "Москва",
  "country": "Россия",
  "latitude": 55.7558,
  "longitude": 37.6173,
  "timezone": "Europe/Moscow"
}
```

**Response:** `201 Created`

---

### Update City

Обновление города.

**Endpoint:** `PUT /admin/cities/{id}`

**Access:** admin

**Request Body:** (аналогично Create)

**Response:** `200 OK`

---

### Delete City

Удаление города.

**Endpoint:** `DELETE /admin/cities/{id}`

**Access:** admin

**Response:** `204 No Content`

---

## Health Checks

### Liveness Probe

Проверка работоспособности приложения.

**Endpoint:** `GET /health`

**Access:** Public

**Response:** `200 OK`

```json
{
  "success": true,
  "status": "ok"
}
```

---

### Readiness Probe

Проверка готовности приложения обслуживать запросы.

**Endpoint:** `GET /ready`

**Access:** Public

**Response:** `200 OK` (when ready)

```json
{
  "success": true,
  "status": "ready",
  "database": "connected",
  "redis": "connected"
}
```

**Response:** `503 Service Unavailable` (when not ready)

```json
{
  "success": false,
  "status": "not_ready",
  "database": "disconnected",
  "redis": "connected"
}
```

---

## Error Codes

| Code | HTTP Status | Description |
|------|-------------|-------------|
| NOT_FOUND | 404 | Ресурс не найден |
| ALREADY_EXISTS | 409 | Ресурс уже существует |
| INVALID_INPUT | 400 | Некорректные входные данные |
| UNAUTHORIZED | 401 | Требуется аутентификация |
| FORBIDDEN | 403 | Доступ запрещён |
| SLOT_UNAVAILABLE | 409 | Слот недоступен |
| BOOKING_CANCEL_LATE | 400 | Слишком поздно для отмены |
| USER_BLOCKED | 403 | Пользователь заблокирован |
| BATHHOUSE_NOT_ACTIVE | 400 | Баня не активна |
| BATHHOUSE_HAS_BOOKINGS | 409 | Баня имеет бронирования |
| REVIEW_ALREADY_RESPONDED | 409 | На отзыв уже дан ответ |
| INTERNAL_SERVER_ERROR | 500 | Внутренняя ошибка сервера |

---

## Authentication

### Token Format

JWT токены отправляются в заголовке `Authorization`:

```
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

### Token Expiration

Токены действительны 24 часа с момента выдачи. По истечении времени требуется повторный вход.

### Refresh Token

В данной версии API refresh tokens не поддерживаются. Требуется повторный вход для получения нового токена.

---

## Pagination

### Query Parameters

```
page=1          # Номер страницы (по умолчанию: 1)
page_size=20    # Размер страницы (по умолчанию: 20, максимум: 100)
```

### Response Meta

```json
{
  "meta": {
    "page": 1,
    "page_size": 20,
    "total_count": 100,
    "total_pages": 5
  }
}
```

---

## Rate Limiting

На данный момент rate limiting не реализован, но планируется в будущих версиях.

---

## CORS

API поддерживает CORS для кросс-доменных запросов. Список разрешённых источников конфигурируется через переменные окружения:
- `BANI_CORS_ALLOWED_ORIGINS`
- `BANI_CORS_ALLOWED_METHODS`
- `BANI_CORS_ALLOWED_HEADERS`

---

## Версионирование

API используется версионирование через путь: `/api/v1`, `/api/v2`, и т.д.

При изменении API будут выпущены новые версии с сохранением обратной совместимости в текущей версии.
