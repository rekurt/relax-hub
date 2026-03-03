# Авторизация через соцсети

## Overview

Вход через VK, Yandex ID, Google OAuth 2.0. Привязка нескольких провайдеров к одному аккаунту, автозаполнение профиля (имя, email, аватар), бесшовная регистрация при первом входе.

**Коммерческая ценность:** Снижает friction при регистрации на 60-70%. Конверсия "посетитель -> зарегистрированный" вырастает с 10% до 35%. Экономит время пользователей.

## Context

- Files involved: internal/domain/user.go, internal/service/auth_service.go, internal/handler/auth.go, internal/server/router.go, migrations/
- Related patterns: Clean architecture, Uber fx DI, JWT auth
- Dependencies: golang.org/x/oauth2 (стандартная библиотека OAuth2)
- Текущая auth: email + password, JWT tokens

## Development Approach

- **Testing approach**: Regular (code first, then tests)
- Complete each task fully before moving to the next
- **CRITICAL: every task MUST include new/updated tests**
- **CRITICAL: all tests must pass before starting next task**

## Implementation Steps

### Task 1: Доменные модели OAuth

**Files:**
- Create: `internal/domain/oauth.go`
- Modify: `internal/domain/errors.go`
- Create: `migrations/000018_social_auth.up.sql`
- Create: `migrations/000018_social_auth.down.sql`

- [ ] Создать модель SocialAccount:
  ```
  SocialAccount {
    ID          uuid.UUID
    UserID      uuid.UUID
    Provider    string          // "vk", "yandex", "google"
    ProviderID  string          // ID пользователя у провайдера
    Email       string
    Name        string
    AvatarURL   string
    AccessToken string          // зашифрованный
    LinkedAt    time.Time
  }
  ```
- [ ] Добавить domain-ошибки: ErrSocialAccountAlreadyLinked, ErrSocialAccountNotFound
- [ ] Создать миграцию с UNIQUE(provider, provider_id)
- [ ] Написать тесты
- [ ] Запустить go test ./... - все тесты должны пройти

### Task 2: OAuth-провайдеры

**Files:**
- Create: `internal/auth/oauth.go`
- Create: `internal/auth/vk.go`
- Create: `internal/auth/yandex.go`
- Create: `internal/auth/google.go`
- Modify: `config/config.go`

- [ ] Создать интерфейс OAuthProvider:
  ```go
  type OAuthProvider interface {
    GetAuthURL(state string) string
    Exchange(ctx context.Context, code string) (*OAuthUserInfo, error)
  }
  type OAuthUserInfo struct {
    ProviderID string
    Email      string
    Name       string
    AvatarURL  string
  }
  ```
- [ ] Реализовать для VK, Yandex, Google (каждый со своей спецификой API)
- [ ] Конфигурация: BANI_OAUTH_VK_CLIENT_ID, BANI_OAUTH_VK_CLIENT_SECRET, etc.
- [ ] Написать тесты с моковыми провайдерами
- [ ] Запустить go test ./... - все тесты должны пройти

### Task 3: Репозиторий и сервис OAuth

**Files:**
- Modify: `internal/repository/interfaces.go`
- Create: `internal/repository/postgres/social_account.go`
- Create: `internal/repository/mock/social_account.go`
- Modify: `internal/service/auth_service.go`

- [ ] SocialAccountRepository: Create, GetByProviderAndID, ListByUser, Delete
- [ ] Расширить AuthService:
  - GetOAuthURL(provider) — URL для авторизации
  - OAuthCallback(provider, code) — обработка callback: найти/создать пользователя, вернуть JWT
  - LinkSocialAccount(userID, provider, code) — привязать соцсеть к существующему аккаунту
  - UnlinkSocialAccount(userID, provider) — отвязать (только если есть пароль или другая соцсеть)
- [ ] При первом входе: создать пользователя без пароля, заполнить name, email, avatar из провайдера
- [ ] При повторном входе: найти по provider+provider_id, вернуть JWT
- [ ] Написать unit-тесты
- [ ] Запустить go test ./... - все тесты должны пройти

### Task 4: Хендлеры OAuth

**Files:**
- Modify: `internal/handler/auth.go`
- Modify: `internal/server/router.go`

- [ ] GET /api/v1/auth/oauth/{provider} — редирект на провайдера (provider: vk, yandex, google)
- [ ] GET /api/v1/auth/oauth/{provider}/callback — обработка callback, возврат JWT
- [ ] POST /api/v1/auth/link/{provider} — привязать соцсеть к аккаунту (auth)
- [ ] DELETE /api/v1/auth/link/{provider} — отвязать соцсеть (auth)
- [ ] GET /api/v1/auth/me/social-accounts — список привязанных соцсетей (auth)
- [ ] Зарегистрировать маршруты
- [ ] Написать handler-тесты
- [ ] Запустить go test ./... - все тесты должны пройти

### Task 5: Верификация

- [ ] Запустить полный тест-сьют: go test ./... -v
- [ ] Запустить линтер: make lint
- [ ] Запустить go vet ./...
- [ ] Обновить CLAUDE.md
- [ ] Переместить этот план в docs/plans/completed/
