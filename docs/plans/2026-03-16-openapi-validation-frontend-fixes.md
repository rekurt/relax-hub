# Валидация OpenAPI и исправление фронтенда

## Overview

Полный аудит OpenAPI документации, сверка фронтенд запросов с бэкенд эндпоинтами,
исправление найденных проблем.

## Результаты валидации

### Swagger.json vs Router.go: СОВПАДАЮТ

Все 120+ эндпоинтов из swagger.json точно соответствуют маршрутам в router.go. Маршруты, которые
корректно отсутствуют в swagger (инфраструктурные): /health, /ready, /sitemap.xml, /calendar/{token}.ics, /swagger/*,
/ws/notifications, /widget.js, /widget.css.

### Фронтенд API клиент vs Swagger.json: СОВПАДАЮТ

Orval-генерированный клиент (33 модуля, 120+ хуков) корректно отражает swagger.json. TypeScript
компиляция проходит без ошибок.

### Все 47 страниц фронтенда: РЕАЛИЗОВАНЫ

Нет stub-страниц, нет TODO-комментариев, все маршруты из router.tsx покрыты компонентами.

### НАЙДЕННЫЕ ПРОБЛЕМЫ:

1. MISSING ENDPOINT: Нет GET /my/bathhouses/{id}/photos - PhotoManager (страница владельца) использует публичный GET /bathhouses/{id}/photos, который возвращает ТОЛЬКО verified фото. Владелец не видит фото в статусе pending и rejected.
2. TEST FAILURES: 2 теста в CertificatePurchase.test.tsx - тест "shows success result after purchase" не мокает useAuthStore, поэтому isAuthenticated = false и кнопка "Мои сертификаты" не рендерится.

## Context

- Files involved:
  - `internal/handler/photo_handler.go` (добавить handler для owner фото)
  - `internal/service/photo_service.go` (проверить/добавить метод для всех фото владельца)
  - `internal/repository/` (проверить repo метод)
  - `internal/server/router.go` (зарегистрировать новый маршрут)
  - `docs/swagger.json` (обновить через make swagger)
  - `frontend/src/api/generated/` (перегенерировать)
  - `frontend/src/pages/photos/PhotoManager.tsx` (использовать новый хук для owner)
  - `frontend/src/__tests__/CertificatePurchase.test.tsx` (исправить тест)
- Related patterns: handler -> service -> repository, swagger annotations, orval code generation
- Dependencies: нет новых зависимостей

## Development Approach

- **Testing approach**: Regular (code first, then tests)
- Complete each task fully before moving to the next
- **CRITICAL: every task MUST include new/updated tests**
- **CRITICAL: all tests must pass before starting next task**

## Implementation Steps

### Task 1: Добавить эндпоинт GET /my/bathhouses/{id}/photos для владельца

**Files:**
- Modify: `internal/handler/photo_handler.go`
- Modify: `internal/server/router.go`
- Possibly modify: `internal/service/photo_service.go`, `internal/repository/postgres/photo_repo.go`

- [x] Проверить существующий сервис/репозиторий - есть ли метод для получения ВСЕХ фото бани (не только verified)
- [x] Если нет - добавить метод в интерфейс repo и реализацию
- [x] Добавить handler метод ListByBathhouseOwner в PhotoHandler (возвращает фото всех статусов)
- [x] Добавить swagger-аннотации для нового эндпоинта
- [x] Зарегистрировать маршрут в router.go: `r.With(auth, middleware.RequireOwnerOrRepresentative()).Get("/my/bathhouses/{id}/photos", p.PhotoHandler.ListByBathhouseOwner)`
- [x] Запустить `make swagger` для обновления docs/
- [x] Запустить `go test ./internal/handler/ -v -run Photo` и `go vet ./...`

### Task 2: Перегенерировать фронтенд API клиент и обновить PhotoManager

**Files:**
- Regenerate: `frontend/src/api/generated/` (via orval)
- Modify: `frontend/src/pages/photos/PhotoManager.tsx`

- [x] Запустить `make frontend-generate-api` для перегенерации API клиента
- [x] Проверить что новый хук useGetMyBathhousesIdPhotos появился в сгенерированном коде
- [x] Обновить PhotoManager.tsx: использовать useGetMyBathhousesIdPhotos вместо useGetBathhousesIdPhotos
- [x] Обновить invalidateQueries key на `/my/bathhouses/${id}/photos`
- [x] Проверить TypeScript компиляцию: `cd frontend && npx tsc --noEmit`

### Task 3: Исправить тесты CertificatePurchase

**Files:**
- Modify: `frontend/src/__tests__/CertificatePurchase.test.tsx`

- [ ] Добавить мок useAuthStore в тест (установить isAuthenticated = true для теста "shows success result after purchase")
- [ ] Проверить второй падающий тест и исправить при необходимости
- [ ] Запустить `cd frontend && npx vitest run src/__tests__/CertificatePurchase.test.tsx`

### Task 4: Финальная верификация

- [ ] Запустить полную сборку бэкенда: `go build ./...`
- [ ] Запустить линтер: `make lint`
- [ ] Запустить бэкенд тесты: `go test ./... -v`
- [ ] Запустить фронтенд тесты: `cd frontend && npx vitest run`
- [ ] Запустить TypeScript проверку: `cd frontend && npx tsc --noEmit`
- [ ] Убедиться что все тесты проходят (561/561)

### Task 5: Update documentation

- [ ] update CLAUDE.md if internal patterns changed
- [ ] move this plan to `docs/plans/completed/`
