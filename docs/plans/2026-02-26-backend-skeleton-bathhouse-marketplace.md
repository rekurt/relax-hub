# Скелет бекенда агрегатора бань (B2BC маркетплейс)

## Overview

Создание скелета Go-бекенда для B2BC маркетплейса услуг бронирования бань. Включает структуру проекта, конфигурацию через Cobra+Viper, DI через Uber fx, REST API через chi, работу с PostgreSQL и Redis, геопоиск бань на карте, фильтрацию, онлайн-бронирование, RBAC с четырьмя ролями (admin, client, owner, representative).

## Context

- Files involved: проект создается с нуля
- Related patterns: стандартная Go project layout (cmd/, internal/, pkg/)
- Dependencies: go-chi/chi, spf13/cobra, spf13/viper, uber-go/fx, jackc/pgx, redis/go-redis, golang-migrate/migrate, golang-jwt/jwt

## Development Approach

- Testing approach: Regular (code first, then tests)
- Complete each task fully before moving to the next
- Используем clean architecture: handler -> service -> repository
- RBAC через middleware + per-handler проверки
- **CRITICAL: every task MUST include new/updated tests**
- **CRITICAL: all tests must pass before starting next task**

## Роли и флоу

### Admin (администратор платформы)
- Модерация бань (approve/reject pending бань)
- Просмотр всех пользователей, бань, бронирований
- Блокировка/разблокировка пользователей
- Управление справочником городов

### Client (клиент)
- Поиск и фильтрация бань на карте
- Онлайн-бронирование
- Отмена своих бронирований
- Просмотр своих бронирований
- Оставление отзывов на завершённые бронирования

### Owner (хозяин бани)
- CRUD своих бань
- Просмотр бронирований своих бань
- Подтверждение/отклонение бронирований
- Управление представителями (invite/revoke)

### Representative (представитель хозяина бани)
- Те же права что у owner, но только для бань, к которым привязан
- Не может управлять другими представителями
- Не может удалять бани

## Implementation Steps

### Task 1: Инициализация Go-модуля и структура проекта

**Files:**
- Create: `go.mod`
- Create: `cmd/server/main.go`
- Create: `cmd/server/root.go`
- Create: `cmd/server/serve.go`
- Create: `config/config.go`
- Create: `config/config.yaml`
- Create: `.env.example`

- [x] Инициализировать go module
- [x] Настроить Cobra CLI: root command + serve subcommand
- [x] Настроить Viper для чтения config.yaml и env-переменных
- [x] Конфиг: server (host, port), database (DSN), redis (addr, password, db), jwt (secret, token_ttl)
- [x] Написать тесты для парсинга конфигурации
- [x] Запустить тесты - должны проходить

### Task 2: DI-контейнер (Uber fx) и подключение к БД

**Files:**
- Create: `internal/app/app.go`
- Create: `internal/database/postgres.go`
- Create: `internal/database/redis.go`

- [x] Создать fx.Module для PostgreSQL (pgxpool)
- [x] Создать fx.Module для Redis (go-redis client)
- [x] Собрать основной fx.App в internal/app/app.go
- [x] Интегрировать fx.App в Cobra serve command
- [x] Написать тесты для создания провайдеров (с моками)
- [x] Запустить тесты - должны проходить

### Task 3: HTTP-сервер, middleware и RBAC

**Files:**
- Create: `internal/server/server.go`
- Create: `internal/server/router.go`
- Create: `internal/middleware/logging.go`
- Create: `internal/middleware/cors.go`
- Create: `internal/middleware/auth.go`
- Create: `internal/middleware/rbac.go`

- [x] Создать HTTP-сервер на chi с graceful shutdown
- [x] Добавить fx.Module для HTTP-сервера
- [x] Middleware: structured logging, CORS, recovery, request ID
- [x] Middleware auth.go: JWT-аутентификация - парсинг токена, извлечение user_id и role в context
- [x] Middleware rbac.go: RBAC middleware с функцией RequireRole(roles ...domain.UserRole) - проверяет роль из контекста:
```go
// Извлечение данных из контекста
func GetUserID(ctx context.Context) uuid.UUID
func GetUserRole(ctx context.Context) domain.UserRole

// Middleware
func RequireAuth(authService service.AuthService) func(http.Handler) http.Handler
func RequireRole(roles ...domain.UserRole) func(http.Handler) http.Handler
func RequireOwnerOrRepresentative() func(http.Handler) http.Handler
```
- [x] Роутер: /api/v1 группа, healthcheck endpoint GET /health
- [x] Написать тесты для middleware: auth, rbac (проверка доступа по ролям, запрет для неавторизованных)
- [x] Запустить тесты - должны проходить

### Task 4: Domain models и миграции

**Files:**
- Create: `internal/domain/user.go`
- Create: `internal/domain/bathhouse.go`
- Create: `internal/domain/booking.go`
- Create: `internal/domain/city.go`
- Create: `internal/domain/review.go`
- Create: `internal/domain/representative.go`
- Create: `internal/domain/errors.go`
- Create: `internal/domain/filter.go`
- Create: `migrations/000001_init.up.sql`
- Create: `migrations/000001_init.down.sql`

- [x] Модель User:
```go
type UserRole string
const (
    RoleClient         UserRole = "client"
    RoleOwner          UserRole = "owner"
    RoleRepresentative UserRole = "representative"
    RoleAdmin          UserRole = "admin"
)

type User struct {
    ID           uuid.UUID
    Email        string
    PasswordHash string
    Name         string
    Phone        string
    Role         UserRole
    IsActive     bool
    CreatedAt    time.Time
    UpdatedAt    time.Time
}
```

- [x] Модель Representative (связь представителя с баней):
```go
type Representative struct {
    ID          uuid.UUID
    UserID      uuid.UUID  // пользователь с ролью representative
    BathhouseID uuid.UUID  // к какой бане привязан
    OwnerID     uuid.UUID  // кто назначил (owner)
    CreatedAt   time.Time
}
```

- [x] Модель City:
```go
type City struct {
    ID        int64
    Name      string
    Slug      string    // "moscow", "spb"
    Latitude  float64
    Longitude float64
}
```

- [x] Модель Bathhouse:
```go
type BathhouseStatus string
const (
    BathhouseStatusActive   BathhouseStatus = "active"
    BathhouseStatusInactive BathhouseStatus = "inactive"
    BathhouseStatusPending  BathhouseStatus = "pending"   // на модерации
    BathhouseStatusRejected BathhouseStatus = "rejected"  // отклонена админом
)

type WorkingHours struct {
    DayOfWeek int    // 0=Пн, 6=Вс
    OpenTime  string // "09:00"
    CloseTime string // "23:00"
}

type Bathhouse struct {
    ID            uuid.UUID
    OwnerID       uuid.UUID
    Name          string
    Description   string
    Address       string
    CityID        int64
    Latitude      float64
    Longitude     float64
    PricePerHour  int64            // в копейках
    MinDuration   int              // минимальная длительность в часах
    MaxGuests     int
    HasPool       bool
    HasSauna      bool
    HasSteamRoom  bool
    HasHotTub     bool
    HasBBQ        bool
    HasKaraoke    bool
    Rating        float64
    ReviewCount   int
    Images        []string
    WorkingHours  []WorkingHours
    Status        BathhouseStatus
    CreatedAt     time.Time
    UpdatedAt     time.Time
}
```

- [x] Модель Booking:
```go
type BookingStatus string
const (
    BookingPending   BookingStatus = "pending"
    BookingConfirmed BookingStatus = "confirmed"
    BookingCancelled BookingStatus = "cancelled"
    BookingRejected  BookingStatus = "rejected"
    BookingCompleted BookingStatus = "completed"
)

type Booking struct {
    ID          uuid.UUID
    UserID      uuid.UUID
    BathhouseID uuid.UUID
    StartTime   time.Time
    EndTime     time.Time
    GuestCount  int
    TotalPrice  int64
    Status      BookingStatus
    Comment     string
    CreatedAt   time.Time
    UpdatedAt   time.Time
}
```

- [x] Модель Review:
```go
type Review struct {
    ID          uuid.UUID
    UserID      uuid.UUID
    BathhouseID uuid.UUID
    BookingID   uuid.UUID
    Rating      int          // 1-5
    Text        string
    CreatedAt   time.Time
}
```

- [x] Фильтры и пагинация (filter.go):
```go
type BathhouseFilter struct {
    CityID       *int64
    CitySlug     *string
    PriceMin     *int64
    PriceMax     *int64
    MinGuests    *int
    HasPool      *bool
    HasSauna     *bool
    HasSteamRoom *bool
    HasHotTub    *bool
    HasBBQ       *bool
    HasKaraoke   *bool
    MinRating    *float64
    Latitude     *float64
    Longitude    *float64
    RadiusKm     *float64
    Status       *BathhouseStatus  // для admin-фильтрации
    SortBy       string            // "price", "rating", "distance"
    SortOrder    string            // "asc", "desc"
    Page         int
    PageSize     int
}

type PaginatedResult[T any] struct {
    Items      []T
    TotalCount int64
    Page       int
    PageSize   int
    TotalPages int
}
```

- [x] Domain errors:
```go
var (
    ErrNotFound          = errors.New("not found")
    ErrAlreadyExists     = errors.New("already exists")
    ErrInvalidInput      = errors.New("invalid input")
    ErrUnauthorized      = errors.New("unauthorized")
    ErrForbidden         = errors.New("forbidden")
    ErrSlotUnavailable   = errors.New("time slot is unavailable")
    ErrBookingCancelLate = errors.New("too late to cancel booking")
    ErrUserBlocked       = errors.New("user is blocked")
    ErrBathhouseNotActive = errors.New("bathhouse is not active")
)
```

- [x] SQL-миграция:
  - таблицы: users, cities, bathhouses, bookings, reviews, representatives
  - PostGIS extension для гео-поиска
  - representatives: unique constraint на (user_id, bathhouse_id)
  - индексы: city_id, coordinates (GIST), status, user_id, bathhouse_id, owner_id
  - reviews: unique constraint на (user_id, booking_id)
- [x] Настроить golang-migrate для запуска миграций из CLI (cobra subcommand migrate)
- [x] Написать тесты для валидации доменных моделей
- [x] Запустить тесты - должны проходить

### Task 5: Repository layer (интерфейсы и PostgreSQL-реализация)

**Files:**
- Create: `internal/repository/interfaces.go`
- Create: `internal/repository/postgres/user_repo.go`
- Create: `internal/repository/postgres/bathhouse_repo.go`
- Create: `internal/repository/postgres/booking_repo.go`
- Create: `internal/repository/postgres/review_repo.go`
- Create: `internal/repository/postgres/city_repo.go`
- Create: `internal/repository/postgres/representative_repo.go`

- [x] Интерфейсы репозиториев в interfaces.go:
```go
type UserRepository interface {
    Create(ctx context.Context, user *domain.User) error
    GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
    GetByEmail(ctx context.Context, email string) (*domain.User, error)
    Update(ctx context.Context, user *domain.User) error
    List(ctx context.Context, page, pageSize int) (*domain.PaginatedResult[domain.User], error)  // admin
    SetActive(ctx context.Context, id uuid.UUID, active bool) error  // admin: block/unblock
}

type CityRepository interface {
    Create(ctx context.Context, city *domain.City) error      // admin
    GetAll(ctx context.Context) ([]domain.City, error)
    GetBySlug(ctx context.Context, slug string) (*domain.City, error)
    GetByID(ctx context.Context, id int64) (*domain.City, error)
    Update(ctx context.Context, city *domain.City) error      // admin
    Delete(ctx context.Context, id int64) error               // admin
}

type BathhouseRepository interface {
    Create(ctx context.Context, bh *domain.Bathhouse) error
    GetByID(ctx context.Context, id uuid.UUID) (*domain.Bathhouse, error)
    Update(ctx context.Context, bh *domain.Bathhouse) error
    Delete(ctx context.Context, id uuid.UUID) error
    List(ctx context.Context, filter domain.BathhouseFilter) (*domain.PaginatedResult[domain.Bathhouse], error)
    ListByOwner(ctx context.Context, ownerID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.Bathhouse], error)
    UpdateRating(ctx context.Context, bathhouseID uuid.UUID) error
    UpdateStatus(ctx context.Context, id uuid.UUID, status domain.BathhouseStatus) error  // admin moderation
}

type BookingRepository interface {
    Create(ctx context.Context, booking *domain.Booking) error
    GetByID(ctx context.Context, id uuid.UUID) (*domain.Booking, error)
    ListByUser(ctx context.Context, userID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.Booking], error)
    ListByBathhouse(ctx context.Context, bathhouseID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.Booking], error)
    UpdateStatus(ctx context.Context, id uuid.UUID, status domain.BookingStatus) error
    CheckAvailability(ctx context.Context, bathhouseID uuid.UUID, startTime, endTime time.Time) (bool, error)
    GetOverlapping(ctx context.Context, bathhouseID uuid.UUID, startTime, endTime time.Time) ([]domain.Booking, error)
}

type ReviewRepository interface {
    Create(ctx context.Context, review *domain.Review) error
    ListByBathhouse(ctx context.Context, bathhouseID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.Review], error)
    GetByBookingID(ctx context.Context, bookingID uuid.UUID) (*domain.Review, error)
}

type RepresentativeRepository interface {
    Create(ctx context.Context, rep *domain.Representative) error
    Delete(ctx context.Context, id uuid.UUID) error
    GetByUserAndBathhouse(ctx context.Context, userID, bathhouseID uuid.UUID) (*domain.Representative, error)
    ListByBathhouse(ctx context.Context, bathhouseID uuid.UUID) ([]domain.Representative, error)
    ListByUser(ctx context.Context, userID uuid.UUID) ([]domain.Representative, error)
    ListBathhouseIDsByUser(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error) // для быстрой проверки доступа
}
```

- [x] Реализация UserRepository: SQL-запросы через pgx, List для админки с пагинацией, SetActive для блокировки
- [x] Реализация CityRepository: полный CRUD для справочника городов (admin)
- [x] Реализация BathhouseRepository: List с динамическим построением WHERE-условий из BathhouseFilter, гео-запрос через ST_DWithin/ST_Distance (PostGIS), сортировка по distance/price/rating, фильтрация по status для админа
- [x] Реализация BookingRepository: CheckAvailability через проверку пересечения интервалов OVERLAPS, GetOverlapping для показа занятых слотов
- [x] Реализация ReviewRepository: Create с пересчётом рейтинга бани, листинг с пагинацией
- [x] Реализация RepresentativeRepository: CRUD для привязки представителей к баням, ListBathhouseIDsByUser для быстрой проверки доступа
- [x] Зарегистрировать все репозитории как fx-провайдеры
- [x] Написать тесты для репозиториев (мокаем через интерфейсы)
- [x] Запустить тесты - должны проходить

### Task 6: Service layer (бизнес-логика + RBAC)

**Files:**
- Create: `internal/service/auth_service.go`
- Create: `internal/service/user_service.go`
- Create: `internal/service/bathhouse_service.go`
- Create: `internal/service/booking_service.go`
- Create: `internal/service/review_service.go`
- Create: `internal/service/city_service.go`
- Create: `internal/service/representative_service.go`
- Create: `internal/service/access.go`

- [x] AccessChecker (access.go) - централизованная проверка доступа:
```go
type AccessChecker struct {
    repRepo repository.RepresentativeRepository
    bhRepo  repository.BathhouseRepository
}

// CanManageBathhouse - проверяет может ли пользователь управлять баней
// owner - если он владелец бани
// representative - если привязан к этой бане
// admin - всегда может
func (a *AccessChecker) CanManageBathhouse(ctx context.Context, userID uuid.UUID, userRole domain.UserRole, bathhouseID uuid.UUID) error

// CanViewBathhouseBookings - может ли видеть бронирования бани
// owner/representative этой бани, admin
func (a *AccessChecker) CanViewBathhouseBookings(ctx context.Context, userID uuid.UUID, userRole domain.UserRole, bathhouseID uuid.UUID) error
```

- [x] AuthService:
```go
type AuthService interface {
    Register(ctx context.Context, input RegisterInput) (*domain.User, string, error)
    Login(ctx context.Context, email, password string) (*domain.User, string, error)
    ParseToken(ctx context.Context, token string) (uuid.UUID, domain.UserRole, error)
}

type RegisterInput struct {
    Email    string
    Password string
    Name     string
    Phone    string
    Role     domain.UserRole // client или owner (admin создается через seed, representative - через invite)
}
```
Логика: хэширование пароля bcrypt, генерация JWT с claims {user_id, role, exp}, валидация email уникальности, регистрация только как client или owner, проверка IsActive при логине

- [x] UserService:
```go
type UserService interface {
    GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
    Update(ctx context.Context, id uuid.UUID, input UpdateUserInput) (*domain.User, error)
    // Admin methods:
    List(ctx context.Context, page, pageSize int) (*domain.PaginatedResult[domain.User], error)
    Block(ctx context.Context, id uuid.UUID) error
    Unblock(ctx context.Context, id uuid.UUID) error
}
```

- [x] BathhouseService:
```go
type BathhouseService interface {
    Create(ctx context.Context, ownerID uuid.UUID, input CreateBathhouseInput) (*domain.Bathhouse, error)
    GetByID(ctx context.Context, id uuid.UUID) (*domain.Bathhouse, error)
    Update(ctx context.Context, userID uuid.UUID, role domain.UserRole, id uuid.UUID, input UpdateBathhouseInput) (*domain.Bathhouse, error)
    Delete(ctx context.Context, ownerID uuid.UUID, id uuid.UUID) error  // только owner, не representative
    Search(ctx context.Context, filter domain.BathhouseFilter) (*domain.PaginatedResult[domain.Bathhouse], error)
    ListByOwner(ctx context.Context, ownerID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.Bathhouse], error)
    // Admin moderation:
    Approve(ctx context.Context, id uuid.UUID) error
    Reject(ctx context.Context, id uuid.UUID) error
}

type CreateBathhouseInput struct {
    Name         string
    Description  string
    Address      string
    CityID       int64
    Latitude     float64
    Longitude    float64
    PricePerHour int64
    MinDuration  int
    MaxGuests    int
    HasPool      bool
    HasSauna     bool
    HasSteamRoom bool
    HasHotTub    bool
    HasBBQ       bool
    HasKaraoke   bool
    Images       []string
    WorkingHours []domain.WorkingHours
}
```
Логика: при Create статус = pending, Update проверяет доступ через AccessChecker (owner или representative), Delete только owner, Approve/Reject только admin

- [x] BookingService:
```go
type BookingService interface {
    Create(ctx context.Context, userID uuid.UUID, input CreateBookingInput) (*domain.Booking, error)
    Cancel(ctx context.Context, userID uuid.UUID, role domain.UserRole, bookingID uuid.UUID) error
    Confirm(ctx context.Context, userID uuid.UUID, role domain.UserRole, bookingID uuid.UUID) error
    Reject(ctx context.Context, userID uuid.UUID, role domain.UserRole, bookingID uuid.UUID) error
    ListByUser(ctx context.Context, userID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.Booking], error)
    ListByBathhouse(ctx context.Context, userID uuid.UUID, role domain.UserRole, bathhouseID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.Booking], error)
    GetAvailableSlots(ctx context.Context, bathhouseID uuid.UUID, date time.Time) ([]TimeSlot, error)
}

type CreateBookingInput struct {
    BathhouseID uuid.UUID
    StartTime   time.Time
    EndTime     time.Time
    GuestCount  int
    Comment     string
}

type TimeSlot struct {
    StartTime time.Time
    EndTime   time.Time
    Available bool
}
```
Логика: Cancel - клиент отменяет свою бронь (минимум за 2ч) или owner/representative отменяет бронь своей бани, Confirm/Reject - только owner/representative бани через AccessChecker, проверка что баня active

- [x] RepresentativeService:
```go
type RepresentativeService interface {
    Invite(ctx context.Context, ownerID uuid.UUID, input InviteRepresentativeInput) (*domain.Representative, error)
    Revoke(ctx context.Context, ownerID uuid.UUID, representativeID uuid.UUID) error
    ListByBathhouse(ctx context.Context, ownerID uuid.UUID, bathhouseID uuid.UUID) ([]domain.Representative, error)
    GetMyBathhouses(ctx context.Context, userID uuid.UUID) ([]domain.Bathhouse, error)
}

type InviteRepresentativeInput struct {
    UserEmail   string     // email существующего пользователя
    BathhouseID uuid.UUID
}
```
Логика: только owner может invite/revoke, при invite меняем роль пользователя на representative (если был client), проверяем что баня принадлежит owner

- [x] ReviewService:
```go
type ReviewService interface {
    Create(ctx context.Context, userID uuid.UUID, input CreateReviewInput) (*domain.Review, error)
    ListByBathhouse(ctx context.Context, bathhouseID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.Review], error)
}

type CreateReviewInput struct {
    BookingID   uuid.UUID
    Rating      int
    Text        string
}
```
Логика: только клиент с completed booking, один отзыв на бронирование

- [x] CityService:
```go
type CityService interface {
    GetAll(ctx context.Context) ([]domain.City, error)
    GetBySlug(ctx context.Context, slug string) (*domain.City, error)
    // Admin:
    Create(ctx context.Context, input CreateCityInput) (*domain.City, error)
    Update(ctx context.Context, id int64, input UpdateCityInput) (*domain.City, error)
    Delete(ctx context.Context, id int64) error
}
```

- [x] Зарегистрировать все сервисы и AccessChecker как fx-провайдеры
- [x] Написать unit-тесты для каждого сервиса (мокаем репозитории), особенно RBAC-проверки:
  - admin может approve/reject бани, block/unblock пользователей
  - owner может CRUD своих бань, invite/revoke представителей
  - representative может update бани и manage бронирования только привязанных бань
  - client не может выполнять owner/admin действия
- [x] Запустить тесты - должны проходить

### Task 7: HTTP handlers (API endpoints с RBAC)

**Files:**
- Create: `internal/handler/auth_handler.go`
- Create: `internal/handler/bathhouse_handler.go`
- Create: `internal/handler/booking_handler.go`
- Create: `internal/handler/review_handler.go`
- Create: `internal/handler/city_handler.go`
- Create: `internal/handler/representative_handler.go`
- Create: `internal/handler/admin_handler.go`
- Create: `internal/handler/response.go`

- [x] response.go: единый формат ответа:
```go
type APIResponse struct {
    Success bool        `json:"success"`
    Data    interface{} `json:"data,omitempty"`
    Error   *APIError   `json:"error,omitempty"`
    Meta    *Meta       `json:"meta,omitempty"`
}

type APIError struct {
    Code    string `json:"code"`
    Message string `json:"message"`
}

type Meta struct {
    Page       int   `json:"page"`
    PageSize   int   `json:"page_size"`
    TotalCount int64 `json:"total_count"`
    TotalPages int   `json:"total_pages"`
}
```

- [x] Auth endpoints (public):
  - POST /api/v1/auth/register - регистрация (role: client или owner)
  - POST /api/v1/auth/login - вход -> token
  - GET /api/v1/auth/me - текущий пользователь (RequireAuth)

- [x] Bathhouse endpoints:
  - GET /api/v1/bathhouses - список с фильтрами (public, только active бани)
  - GET /api/v1/bathhouses/:id - детали (public, только active или свои)
  - POST /api/v1/bathhouses - создание (RequireRole: owner)
  - PUT /api/v1/bathhouses/:id - обновление (RequireRole: owner, representative + AccessChecker)
  - DELETE /api/v1/bathhouses/:id - удаление (RequireRole: owner, проверка владения)
  - GET /api/v1/bathhouses/:id/available-slots?date=YYYY-MM-DD - свободные слоты (public)
  - GET /api/v1/my/bathhouses - бани текущего owner/representative (RequireRole: owner, representative)

- [x] Booking endpoints (RequireAuth):
  - POST /api/v1/bookings - создание бронирования (RequireRole: client)
  - GET /api/v1/bookings - список бронирований текущего пользователя
  - PATCH /api/v1/bookings/:id/cancel - отмена (client свою, owner/representative свою баню)
  - PATCH /api/v1/bookings/:id/confirm - подтверждение (RequireRole: owner, representative + AccessChecker)
  - PATCH /api/v1/bookings/:id/reject - отклонение (RequireRole: owner, representative + AccessChecker)
  - GET /api/v1/bathhouses/:id/bookings - бронирования бани (RequireRole: owner, representative, admin + AccessChecker)

- [x] Review endpoints:
  - POST /api/v1/bathhouses/:id/reviews - создание отзыва (RequireRole: client)
  - GET /api/v1/bathhouses/:id/reviews - список отзывов (public)

- [x] Representative endpoints (RequireRole: owner):
  - POST /api/v1/bathhouses/:id/representatives - пригласить представителя
  - GET /api/v1/bathhouses/:id/representatives - список представителей
  - DELETE /api/v1/representatives/:id - отозвать представителя

- [x] City endpoints:
  - GET /api/v1/cities - список городов (public)

- [x] Admin endpoints (RequireRole: admin):
  - GET /api/v1/admin/users - список пользователей с пагинацией
  - PATCH /api/v1/admin/users/:id/block - заблокировать
  - PATCH /api/v1/admin/users/:id/unblock - разблокировать
  - PATCH /api/v1/admin/bathhouses/:id/approve - одобрить баню
  - PATCH /api/v1/admin/bathhouses/:id/reject - отклонить баню
  - GET /api/v1/admin/bathhouses?status=pending - бани на модерации
  - POST /api/v1/admin/cities - создать город
  - PUT /api/v1/admin/cities/:id - обновить город
  - DELETE /api/v1/admin/cities/:id - удалить город

- [x] Зарегистрировать хэндлеры как fx-провайдеры, привязать к роутеру с RBAC middleware
- [x] Написать тесты для хэндлеров (httptest), включая проверку RBAC (403 при неправильной роли)
- [x] Запустить тесты - должны проходить

### Task 8: Docker и docker-compose для разработки

**Files:**
- Create: `Dockerfile`
- Create: `docker-compose.yml`
- Create: `Makefile`

- [ ] Multi-stage Dockerfile для Go-приложения
- [ ] docker-compose: app, postgres (с PostGIS), redis
- [ ] Makefile: build, run, test, migrate-up, migrate-down, lint, docker-up, docker-down, seed-admin
- [ ] seed-admin: cobra subcommand для создания первого admin-пользователя
- [ ] Обновить .gitignore
- [ ] Проверить что docker-compose.yml валиден
- [ ] Запустить go vet и тесты - должны проходить

### Task 9: Verify acceptance criteria

- [ ] go build ./... компилируется без ошибок
- [ ] go test ./... все тесты проходят
- [ ] go vet ./... без предупреждений
- [ ] Проверить что cobra CLI работает: help, serve --help, migrate --help
- [ ] Проверить структуру проекта соответствует плану
- [ ] Проверить тестовое покрытие >= 80%
- [ ] Проверить RBAC: admin endpoints недоступны для client/owner, owner endpoints недоступны для client

### Task 10: Update documentation

- [ ] Обновить README.md: описание проекта, структура, роли и RBAC, запуск, API endpoints
- [ ] Создать CLAUDE.md с паттернами проекта
- [ ] Переместить план в docs/plans/completed/
