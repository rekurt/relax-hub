# Встраиваемый виджет бронирования

## Overview

JavaScript-виджет для встраивания на сайт бани. Показ доступных слотов, бронирование без перехода на платформу. Настраиваемый дизайн (цвета, шрифты), генерация кода виджета в личном кабинете владельца.

**Коммерческая ценность:** Расширяет охват платформы: бани с собственными сайтами получают бронирования через платформу. Premium/Promoted фича для владельцев. Каждое бронирование через виджет = комиссия платформе.

## Context

- Files involved: internal/handler/, internal/server/router.go, widget/ (новая директория)
- Related patterns: Clean architecture, Uber fx DI, CORS
- Dependencies: нет новых для backend; JS-виджет — vanilla JS или preact
- Связь: использует существующий API для слотов и бронирований

## Development Approach

- **Testing approach**: Regular (code first, then tests)
- Complete each task fully before moving to the next
- **CRITICAL: every task MUST include new/updated tests**
- **CRITICAL: all tests must pass before starting next task**

## Implementation Steps

### Task 1: API для виджета

**Files:**
- Create: `internal/handler/widget.go`
- Modify: `internal/server/router.go`

- [ ] Создать публичные эндпоинты для виджета (без auth, с API key бани):
  - GET /api/v1/widget/{api_key}/bathhouse — информация о бане (название, фото, цена)
  - GET /api/v1/widget/{api_key}/slots?date=... — доступные слоты на дату
  - POST /api/v1/widget/{api_key}/booking — создать бронирование (имя, телефон, email, дата, время, гости)
- [ ] API key генерируется для каждой бани (UUID, хранится в БД)
- [ ] Rate limiting для widget endpoints
- [ ] CORS: разрешить встраивание с любого домена (отдельная CORS-политика)
- [ ] Написать тесты
- [ ] Запустить go test ./... - все тесты должны пройти

### Task 2: Генерация API-ключей

**Files:**
- Create: `migrations/000017_widget_api_keys.up.sql`
- Create: `migrations/000017_widget_api_keys.down.sql`
- Modify: `internal/domain/bathhouse.go`

- [ ] Добавить поле ApiKey string в Bathhouse
- [ ] Миграция: ALTER bathhouses ADD COLUMN api_key VARCHAR(64) UNIQUE
- [ ] Генерация ключа при создании бани или по запросу владельца
- [ ] GET /api/v1/my/bathhouses/{id}/widget-key — получить API key
- [ ] POST /api/v1/my/bathhouses/{id}/widget-key/regenerate — пересоздать ключ
- [ ] Написать тесты
- [ ] Запустить go test ./... - все тесты должны пройти

### Task 3: JavaScript виджет

**Files:**
- Create: `widget/src/widget.js`
- Create: `widget/src/styles.css`
- Create: `widget/build.sh`

- [ ] Vanilla JS виджет (без фреймворков, минимальный размер):
  - Календарь с доступными датами
  - Слоты на выбранную дату
  - Форма бронирования (имя, телефон, email, кол-во гостей)
  - Подтверждение бронирования
- [ ] Настраиваемые параметры: primaryColor, fontFamily, language
- [ ] Код встраивания:
  ```html
  <div id="bani-widget" data-api-key="xxx" data-color="#4CAF50"></div>
  <script src="https://api.bani.ru/widget.js"></script>
  ```
- [ ] Сборка: минификация, один файл widget.min.js + widget.min.css
- [ ] Написать тесты

### Task 4: Генератор кода виджета в ЛК

**Files:**
- Modify: `internal/handler/widget.go`
- Modify: `internal/server/router.go`

- [ ] GET /api/v1/my/bathhouses/{id}/widget-code — сгенерировать HTML-код для встраивания с превью настроек
- [ ] Параметры: color, fontFamily, showPrice, showRating
- [ ] Эндпоинт для раздачи статики виджета: GET /widget.js, GET /widget.css
- [ ] Написать тесты
- [ ] Запустить go test ./... - все тесты должны пройти

### Task 5: Верификация

- [ ] Запустить полный тест-сьют: go test ./... -v
- [ ] Запустить линтер: make lint
- [ ] Запустить go vet ./...
- [ ] Переместить этот план в docs/plans/completed/
