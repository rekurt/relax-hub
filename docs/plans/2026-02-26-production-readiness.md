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

- [x] Добавить обязательную валидацию переменных: DSN, JWT Secret, Redis Addr
- [x] Добавить поле Environment (dev/staging/production) в Config
- [x] Валидировать, что JWT Secret в продакшене не содержит "change-me" и имеет минимальную длину 32 символа
- [x] Проверить, что DSN использует sslmode=require в продакшене (или производная от Environment)
- [x] Создать `.env.production.example` с правильными параметрами и комментариями
- [x] Написать тесты для валидации конфигурации
- [x] Запустить go test ./... - все тесты должны пройти

### Task 3: Graceful shutdown и обработка сигналов

**Files:**
- Modify: `cmd/server/serve.go`
- Modify: `cmd/server/root.go`
- Modify: `internal/app/app.go`

- [x] Добавить обработку SIGINT и SIGTERM в serve command
- [x] Обеспечить graceful shutdown с таймаутом 30 сек для завершения активных запросов
- [x] Логировать получение сигнала shutdown
- [x] Выполнить корректное закрытие database connections и redis connections
- [x] Написать интеграционные тесты для shutdown сценария
- [x] Запустить go test ./... - все тесты должны пройти

### Task 4: Улучшение обработки ошибок в handlers

**Files:**
- Modify: `internal/handler/response.go`
- Modify: `internal/middleware/recovery.go` (если есть или создать)
- Modify: `internal/handler/*.go` (все хендлеры)

- [x] Добавить middleware для recovery от паник с логированием stack trace
- [x] Убедиться, что все HTTP ошибки логируются (особенно 5xx)
- [x] Отправлять stack trace только в dev-режиме, в продакшене скрывать детали
- [x] Добавить request ID для трейсинга ошибок через контекст
- [x] Написать тесты для error scenarios в хендлерах
- [x] Запустить go test ./... - все тесты должны пройти

### Task 5: Оптимизация timeouts для продакшена

**Files:**
- Modify: `internal/server/server.go`
- Modify: `internal/database/postgres.go`
- Modify: `internal/database/redis.go`

- [x] Убедиться, что ReadHeaderTimeout, ReadTimeout, WriteTimeout установлены правильно в server
- [x] Добавить Context timeout для database queries (20-30 сек)
- [x] Добавить timeout для Redis operations (5 сек)
- [x] Логировать timeout events
- [x] Написать тесты для timeout сценариев
- [x] Запустить go test ./... - все тесты должны пройти

### Task 6: Healthcheck и readiness endpoints

**Files:**
- Modify: `internal/handler/health.go` (если есть или создать)
- Modify: `internal/server/router.go`

- [x] Создать /health endpoint для liveness probe (простой 200 OK)
- [x] Создать /ready endpoint для readiness probe (проверка DB и Redis доступности)
- [x] Логировать неудачные health checks
- [x] Написать тесты для health endpoints
- [x] Запустить go test ./... - все тесты должны пройти

### Task 7: Пересмотр Dockerfile и CI/CD конфигурации

**Files:**
- Modify: `Dockerfile`
- Create: `.dockerignore`
- Create: `.github/workflows/deploy.yml` (опционально, если используется GH Actions)

- [x] Убедиться, что Dockerfile использует multi-stage build (уже реализовано, проверить)
- [x] Добавить healthcheck в Dockerfile для автоматических перезагрузок
- [x] Добавить .dockerignore для уменьшения размера образа
- [x] Убедиться, что используется non-root user в контейнере (опционально, для максимальной безопасности)
- [x] Создать документацию для deployment процесса
- [x] Запустить docker build и docker run локально для проверки

### Task 8: Документация API и deployment guide

**Files:**
- Create: `docs/API.md`
- Create: `docs/DEPLOYMENT.md`
- Modify: `README.md`

- [x] Документировать все API endpoints с примерами запросов и ответов
- [x] Написать deployment guide для Kubernetes или Docker Compose в продакшене
- [x] Описать требования к переменным окружения для разных окружений
- [x] Добавить примеры для health checks и monitoring
- [x] Обновить README с информацией о production deployment
- [x] Не требует тестов, чисто документирование

### Task 9: Финальная верификация и оптимизация

**Files:**
- N/A (только проверка существующего кода)

- [x] Запустить полный тест-сьют: go test ./... -v
- [x] Запустить линтер: make lint
- [x] Запустить go vet ./...
- [x] Проверить coverage (должно быть 70%+)
- [x] Убедиться, что нет TODO или FIXME в production-коде
- [x] Проверить, что миграции корректны и имеют rollback версии
- [x] Убедиться, что нет хардкода sensitive data (passwords, tokens, endpoints)
- [x] Проверить, что все external API вызовы имеют timeouts
- [x] Запустить docker build и проверить размер образа

### Task 10: Завершение и подготовка к мерджу

- [x] Обновить CLAUDE.md если появились новые паттерны
- [x] Перевести этот план в completed
- [x] Подготовить PR с кратким description всех изменений
- [x] Убедиться, что все CI/CD checks проходят
