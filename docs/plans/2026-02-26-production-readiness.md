# Подготовка B2BC маркетплейса бань к продакшену

## Overview

Полный план подготовки Go-бекенда к продакшену. Включает улучшения безопасности, мониторинга, логирования, документации, и проверку критических компонентов. План разбит на задачи, включающие: правильную обработку ошибок, структурированное логирование, конфигурацию для продакшена, оптимизацию производительности, документацию API, и финальную верификацию.

## Context

- Files involved: cmd/server/, config/, internal/server/, internal/database/, internal/middleware/, internal/handler/
- Related patterns: Clean architecture, Uber fx DI, chi router, PostgreSQL с PostGIS, Redis
- Dependencies: существующие зависимости (fx, chi, pgx, go-redis)

## Development Approach

- **Testing approach**: Regular (code first, then tests)
- Complete each task fully before moving to the next
- **CRITICAL: every task MUST include new/updated tests**
- **CRITICAL: all tests must pass before starting next task**
- Основной фокус на production-ready коде без breaking changes

## Implementation Steps

### Task 1: Структурированное логирование

**Files:**
- Create: `internal/logger/logger.go`
- Create: `internal/logger/module.go`
- Modify: `cmd/server/serve.go`
- Modify: `config/config.go`
- Modify: `internal/server/server.go`
- Modify: `internal/database/postgres.go`
- Modify: `internal/database/redis.go`

- [x] Создать пакет logger с функциями Info, Error, Debug, Warn
- [x] Добавить поддержку уровня логирования (debug/info/warn/error) через конфигурацию
- [x] Заменить log.Printf на структурированный логинг в server.go
- [x] Логировать подключение к БД и Redis в database пакете
- [x] Добавить логирование в database providers (создание connection pool)
- [x] Написать юнит-тесты для logger
- [x] Запустить go test ./... - все тесты должны пройти

### Task 2: Обработка конфигурации для продакшена

**Files:**
- Modify: `config/config.go`
- Modify: `.env.example`
- Create: `.env.production.example`
- Modify: `config/config.yaml`

- [ ] Добавить обязательную валидацию переменных: DSN, JWT Secret, Redis Addr
- [ ] Добавить поле Environment (dev/staging/production) в Config
- [ ] Валидировать, что JWT Secret в продакшене не содержит "change-me" и имеет минимальную длину 32 символа
- [ ] Проверить, что DSN использует sslmode=require в продакшене (или производная от Environment)
- [ ] Создать `.env.production.example` с правильными параметрами и комментариями
- [ ] Написать тесты для валидации конфигурации
- [ ] Запустить go test ./... - все тесты должны пройти

### Task 3: Graceful shutdown и обработка сигналов

**Files:**
- Modify: `cmd/server/serve.go`
- Modify: `cmd/server/root.go`
- Modify: `internal/app/app.go`

- [ ] Добавить обработку SIGINT и SIGTERM в serve command
- [ ] Обеспечить graceful shutdown с таймаутом 30 сек для завершения активных запросов
- [ ] Логировать получение сигнала shutdown
- [ ] Выполнить корректное закрытие database connections и redis connections
- [ ] Написать интеграционные тесты для shutdown сценария
- [ ] Запустить go test ./... - все тесты должны пройти

### Task 4: Улучшение обработки ошибок в handlers

**Files:**
- Modify: `internal/handler/response.go`
- Modify: `internal/middleware/recovery.go` (если есть или создать)
- Modify: `internal/handler/*.go` (все хендлеры)

- [ ] Добавить middleware для recovery от паник с логированием stack trace
- [ ] Убедиться, что все HTTP ошибки логируются (особенно 5xx)
- [ ] Отправлять stack trace только в dev-режиме, в продакшене скрывать детали
- [ ] Добавить request ID для трейсинга ошибок через контекст
- [ ] Написать тесты для error scenarios в хендлерах
- [ ] Запустить go test ./... - все тесты должны пройти

### Task 5: Оптимизация timeouts для продакшена

**Files:**
- Modify: `internal/server/server.go`
- Modify: `internal/database/postgres.go`
- Modify: `internal/database/redis.go`

- [ ] Убедиться, что ReadHeaderTimeout, ReadTimeout, WriteTimeout установлены правильно в server
- [ ] Добавить Context timeout для database queries (20-30 сек)
- [ ] Добавить timeout для Redis operations (5 сек)
- [ ] Логировать timeout events
- [ ] Написать тесты для timeout сценариев
- [ ] Запустить go test ./... - все тесты должны пройти

### Task 6: Healthcheck и readiness endpoints

**Files:**
- Modify: `internal/handler/health.go` (если есть или создать)
- Modify: `internal/server/router.go`

- [ ] Создать /health endpoint для liveness probe (простой 200 OK)
- [ ] Создать /ready endpoint для readiness probe (проверка DB и Redis доступности)
- [ ] Логировать неудачные health checks
- [ ] Написать тесты для health endpoints
- [ ] Запустить go test ./... - все тесты должны пройти

### Task 7: Пересмотр Dockerfile и CI/CD конфигурации

**Files:**
- Modify: `Dockerfile`
- Create: `.dockerignore`
- Create: `.github/workflows/deploy.yml` (опционально, если используется GH Actions)

- [ ] Убедиться, что Dockerfile использует multi-stage build (уже реализовано, проверить)
- [ ] Добавить healthcheck в Dockerfile для автоматических перезагрузок
- [ ] Добавить .dockerignore для уменьшения размера образа
- [ ] Убедиться, что используется non-root user в контейнере (опционально, для максимальной безопасности)
- [ ] Создать документацию для deployment процесса
- [ ] Запустить docker build и docker run локально для проверки

### Task 8: Документация API и deployment guide

**Files:**
- Create: `docs/API.md`
- Create: `docs/DEPLOYMENT.md`
- Modify: `README.md`

- [ ] Документировать все API endpoints с примерами запросов и ответов
- [ ] Написать deployment guide для Kubernetes или Docker Compose в продакшене
- [ ] Описать требования к переменным окружения для разных окружений
- [ ] Добавить примеры для health checks и monitoring
- [ ] Обновить README с информацией о production deployment
- [ ] Не требует тестов, чисто документирование

### Task 9: Финальная верификация и оптимизация

**Files:**
- N/A (только проверка существующего кода)

- [ ] Запустить полный тест-сьют: go test ./... -v
- [ ] Запустить линтер: make lint
- [ ] Запустить go vet ./...
- [ ] Проверить coverage (должно быть 70%+)
- [ ] Убедиться, что нет TODO или FIXME в production-коде
- [ ] Проверить, что миграции корректны и имеют rollback версии
- [ ] Убедиться, что нет хардкода sensitive data (passwords, tokens, endpoints)
- [ ] Проверить, что все external API вызовы имеют timeouts
- [ ] Запустить docker build и проверить размер образа

### Task 10: Завершение и подготовка к мерджу

- [ ] Обновить CLAUDE.md если появились новые паттерны
- [ ] Перевести этот план в completed
- [ ] Подготовить PR с кратким description всех изменений
- [ ] Убедиться, что все CI/CD checks проходят
