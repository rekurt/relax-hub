# OpenAPI документация с swaggo/swag

## Overview

Добавить автогенерируемую OpenAPI 3.0 документацию ко всем API эндпоинтам с помощью swaggo/swag. Swagger UI будет доступен по /swagger/. Аннотации добавляются в комментарии к хендлерам, спецификация генерируется командой swag init.

## Context

- Files involved: go.mod, Makefile, cmd/server/main.go, internal/server/router.go, internal/handler/*.go (58 файлов), новый docs/ пакет (автогенерируемый)
- Related patterns: все хендлеры в package handler, приватные request/response структуры, APIResponse конверт, chi роутер
- Dependencies: github.com/swaggo/swag, github.com/swaggo/http-swagger/v2

## Development Approach

- **Testing approach**: после каждого таска запускать swag init для проверки валидности аннотаций
- Аннотировать хендлеры группами по доменным областям
- **CRITICAL: swag init должен проходить без ошибок после каждого таска**
- **CRITICAL: go build ./... должен компилироваться после каждого таска**

## Implementation Steps

### Task 1: Установка зависимостей и базовая конфигурация swaggo

**Files:**
- Modify: `go.mod`
- Modify: `Makefile`
- Create: `cmd/server/docs.go` (главные аннотации API)
- Modify: `internal/server/router.go` (маршрут Swagger UI)

- [x] Добавить зависимости: `go get github.com/swaggo/swag/cmd/swag@latest`, `go get github.com/swaggo/http-swagger/v2`
- [x] Создать `cmd/server/docs.go` с главными аннотациями (@title Bani API, @version 1.0, @BasePath /api/v1, @securityDefinitions.apikey BearerAuth)
- [x] Добавить маршрут Swagger UI в router.go: `GET /swagger/*`
- [x] Добавить Makefile таргеты: `make swagger` (swag init), `make swagger-fmt` (swag fmt)
- [x] Запустить `swag init` — должен сгенерировать docs/docs.go, docs/swagger.json, docs/swagger.yaml
- [x] Добавить `docs/` в .gitignore или коммитить (решение: коммитить для CI)
- [x] Проверить: `go build ./...` компилируется, Swagger UI доступен

### Task 2: Аннотации — Auth, Users

**Files:**
- Modify: `internal/handler/auth.go`
- Modify: `internal/handler/oauth.go`

- [ ] Аннотировать Register, Login, GetMe, UpdateProfile, UploadAvatar, DeleteAvatar, GetPublicProfile, GetMyStats
- [ ] Аннотировать OAuth эндпоинты: OAuthRedirect, OAuthCallback, LinkSocialAccount, UnlinkSocialAccount, GetSocialAccounts
- [ ] Запустить `swag init` — без ошибок

### Task 3: Аннотации — Bathhouses, Cities, Photos

**Files:**
- Modify: `internal/handler/bathhouse.go`
- Modify: `internal/handler/city.go`
- Modify: `internal/handler/photo.go`

- [ ] Аннотировать List, Get, GetBySlug, Create, Update, Delete, GetAvailableSlots, GetMeta, GetSchema, GetSimilar, GetPopular, GetByCity
- [ ] Аннотировать GetMyBathhouses, WidgetKey, WidgetCode, RegenerateWidgetKey
- [ ] Аннотировать City CRUD (admin), ListCities (public)
- [ ] Аннотировать Photo эндпоинты: Upload, Reorder, GetPhotos, AdminPending, AdminVerify, AdminReject
- [ ] Запустить `swag init` — без ошибок

### Task 4: Аннотации — Bookings, Payments

**Files:**
- Modify: `internal/handler/booking.go`
- Modify: `internal/handler/payment.go`

- [ ] Аннотировать Create, List, Cancel, Confirm, Reject, Complete, GetBathhouseBookings
- [ ] Аннотировать Pay, GetPayment, ListMyPayments, YooKassaWebhook
- [ ] Запустить `swag init` — без ошибок

### Task 5: Аннотации — Reviews, Media, Complaints

**Files:**
- Modify: `internal/handler/review.go`
- Modify: `internal/handler/media.go`
- Modify: `internal/handler/complaint.go`

- [ ] Аннотировать Create/List/Update/Delete Review, RespondToReview
- [ ] Аннотировать UploadMedia, DeleteMedia, GetGallery
- [ ] Аннотировать Report (review/bathhouse/user), Admin complaints CRUD
- [ ] Аннотировать Admin review moderation: List, PendingCount, Approve, Reject, BatchApprove, BatchReject
- [ ] Запустить `swag init` — без ошибок

### Task 6: Аннотации — Loyalty, Referral, Certificates, Promo

**Files:**
- Modify: `internal/handler/loyalty.go`
- Modify: `internal/handler/referral.go`
- Modify: `internal/handler/certificate.go`
- Modify: `internal/handler/promo.go`

- [ ] Аннотировать Loyalty: GetAccount, GetTransactions, GetLevels
- [ ] Аннотировать Referral: GetCode, GetStats, GetBalance
- [ ] Аннотировать Certificates: Purchase, Redeem, GetBalance, ListMy
- [ ] Аннотировать Promo: Create, List, Delete, Validate, AdminCreate
- [ ] Запустить `swag init` — без ошибок

### Task 7: Аннотации — Chat, Notifications, Favorites, Recommendations

**Files:**
- Modify: `internal/handler/chat.go`
- Modify: `internal/handler/notification.go`
- Modify: `internal/handler/favorite.go`
- Modify: `internal/handler/recommendation.go`
- Modify: `internal/handler/device_token.go`

- [ ] Аннотировать Chat: StartChat, ListConversations, GetMessages, SendMessage, MarkRead, UnreadCount
- [ ] Аннотировать Notifications: List, UnreadCount, MarkRead, MarkAllRead, GetPreferences, UpdatePreferences
- [ ] Аннотировать Favorites: Toggle, ListMy
- [ ] Аннотировать Recommendations: Get, GetPreferences, UpdatePreferences
- [ ] Аннотировать DeviceTokens: Register, Delete
- [ ] Запустить `swag init` — без ошибок

### Task 8: Аннотации — Subscriptions, Pricing, Analytics, Calendar, Representatives, Admin, Widget

**Files:**
- Modify: `internal/handler/subscription.go`
- Modify: `internal/handler/pricing.go`
- Modify: `internal/handler/analytics.go`
- Modify: `internal/handler/calendar.go`
- Modify: `internal/handler/representative.go`
- Modify: `internal/handler/admin.go`
- Modify: `internal/handler/widget.go`
- Modify: `internal/handler/slot_block.go`
- Modify: `internal/handler/external_calendar.go`

- [ ] Аннотировать Subscriptions: Create, Get, Delete, ListMy
- [ ] Аннотировать Pricing: CreateRule, ListRules, UpdateRule, DeleteRule, PriceCalculator
- [ ] Аннотировать Analytics: GetBathhouseAnalytics, GetDailyAnalytics, AdminAnalytics, AdminTopBathhouses
- [ ] Аннотировать Calendar: GetICS, GetToken, SlotBlocks CRUD, ExternalCalendars CRUD
- [ ] Аннотировать Representatives: Add, List, Remove
- [ ] Аннотировать Admin: ListUsers, Block/Unblock, ListBathhouses, Approve/Reject
- [ ] Аннотировать Widget: GetBathhouse, GetSlots, CreateBooking
- [ ] Запустить `swag init` — без ошибок

### Task 9: Финальная верификация

- [ ] Запустить `swag init` — полная генерация без ошибок и ворнингов
- [ ] Запустить `go build ./...` — компиляция без ошибок
- [ ] Запустить `make lint` — без новых ошибок линтера
- [ ] Проверить swagger.json — все эндпоинты из router.go присутствуют в спеке
- [ ] Проверить Swagger UI — корректно рендерится, Try It Out работает с Bearer токеном

### Task 10: Обновить документацию

- [ ] Обновить CLAUDE.md: добавить секцию про OpenAPI/Swagger, команды make swagger/swagger-fmt
- [ ] Переместить план в `docs/plans/completed/`
