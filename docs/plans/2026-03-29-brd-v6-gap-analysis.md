# Доработка сервиса RelaxHub до полного соответствия BRD v6

## Overview
Комплексный план по устранению всех выявленных расхождений между текущей
реализацией и бизнес-требованиями BRD RelaxHub v6. Охватывает backend, frontend, cron-задачи,
антифрод, платежи, аналитику и SEO.

## Методология анализа
Проведено систематическое сопоставление каждого FR-требования (FR-001 — FR-155) из BRD v6 с
кодовой базой. Ниже — только реальные расхождения, сгруппированные по приоритету и
модулю.

## Context
- Архитектура: handler → service → repository (clean architecture, Uber fx DI)
- Backend: Go, chi router, pgx, Redis
- Frontend: React 18, Ant Design, TanStack Query, orval
- Тесты: table-driven unit tests с mock-репозиториями
- Related patterns: все новые модули следуют паттерну domain → repository → service → handler с fx.Module
- Dependencies: go-blurhash (C1), excelize уже есть

## Development Approach
- **Testing approach**: Regular (code first, then tests)
- Complete each task fully before moving to the next
- Каждый блок (A, B, C, D, E) может выполняться независимо, но внутри блока задачи последовательны
- **CRITICAL: every task MUST include new/updated tests**
- **CRITICAL: all tests must pass before starting next task**

---

## Implementation Steps

## БЛОК A: КРИТИЧЕСКИЕ РАСХОЖДЕНИЯ (бизнес-логика нарушена или отсутствует)

### Task A1: Конкурентный контроль бронирования — row-level locking (FR-061)

**Проблема:** CheckAvailability() использует простой COUNT-запрос без блокировки. Race condition: два
параллельных запроса могут создать двойное бронирование одного слота.

**Files:**
- Modify: `internal/repository/postgres/booking_repo.go` — добавить SELECT ... FOR UPDATE или pg_advisory_lock в CheckAvailability + Create
- Modify: `internal/service/booking_service.go` — обернуть check+create в единую транзакцию

- [x] Добавить advisory lock или FOR UPDATE SKIP LOCKED при проверке доступности слота
- [x] Обернуть CheckAvailability + Create в одну транзакцию
- [x] Написать тест на параллельное бронирование одного слота
- [x] Нагрузочный тест (опционально): 10 горутин на один слот
- [x] run project test suite — must pass before Task A2

### Task A2: Подтверждение изменения бронирования обеими сторонами (FR-067)

**Проблема:** Сейчас клиент модифицирует бронирование без подтверждения владельца.
BRD требует: "Изменение требует подтверждения обеих сторон".

**Files:**
- Modify: `internal/domain/booking.go` — добавить статус `pending_modification` и структуру ModificationRequest
- Create: `internal/domain/booking_modification.go` — модель запроса на изменение
- Modify: `internal/service/booking_service.go` — Modify() создаёт запрос, новые методы ApproveModification/RejectModification
- Modify: `internal/handler/booking_handler.go` — эндпоинты approve/reject modification
- Create: `internal/repository/postgres/booking_modification_repo.go`
- Create: `migrations/XXXXXX_booking_modification_requests.up.sql`

- [x] Создать домен модели ModificationRequest (old values, new values, status: pending/approved/rejected)
- [x] Миграция: таблица booking_modification_requests
- [x] Репозиторий: CRUD для modification requests
- [x] Сервис: Modify() → создаёт pending request, отправляет уведомление владельцу
- [x] Сервис: ApproveModification() → применяет изменения, пересчитывает цену
- [x] Сервис: RejectModification() → уведомляет клиента
- [x] Cron: таймаут на ответ владельца (24ч) → авто-отклонение
- [x] Handler: POST /api/v1/my/bookings/{id}/modification/approve, /reject
- [x] Тесты сервиса и хендлера
- [x] Frontend: UI подтверждения для владельца
- [x] run project test suite — must pass before Task A3

### Task A3: Подтверждение продления сеанса обеими сторонами (FR-065)

**Проблема:** Продление выполняется клиентом односторонне. BRD: "Продление
подтверждается обеими сторонами".

**Files:**
- Modify: `internal/service/booking_service.go` — Extend() создаёт запрос на продление
- Modify: `internal/handler/booking_handler.go` — эндпоинты approve/reject extension
- Create: `migrations/XXXXXX_extension_requests.up.sql`

- [ ] Модель ExtensionRequest (booking_id, hours, status)
- [ ] Extend() → создаёт pending request + холдирует средства + уведомляет владельца
- [ ] ApproveExtension() → применяет, списывает средства
- [ ] RejectExtension() → снимает холд, уведомляет клиента
- [ ] Таймаут 30 минут (cron или отложенная задача)
- [ ] Тесты
- [ ] run project test suite — must pass before Task A4

### Task A4: Антифрод при создании листинга (FR-028)

**Проблема:** При создании листинга нет проверки на дубликаты аккаунтов (телефон, email,
реквизиты) и стоп-лист.

**Files:**
- Modify: `internal/antifraud/rules.go` — добавить правило listing_creation
- Create: `internal/antifraud/listing_rules.go` — проверка дубликатов и стоп-листа
- Modify: `internal/service/bathhouse_service.go` — вызвать антифрод перед созданием
- Create: `migrations/XXXXXX_antifraud_stoplist.up.sql` — таблица стоп-листа

- [ ] Таблица antifraud_stoplist (phone, email, payment_details, reason, blocked_at)
- [ ] Правило: поиск дубликатов по телефону/email/реквизитам среди существующих владельцев
- [ ] Правило: проверка по стоп-листу
- [ ] Интеграция в bathhouse_service.Create()
- [ ] Admin-эндпоинт для управления стоп-листом
- [ ] Тесты
- [ ] run project test suite — must pass before Task A5

### Task A5: GPS-валидация при оспаривании неявки (FR-070)

**Проблема:** GPS-координаты при оспаривании no-show фиксируются, но не проверяются по
радиусу 200м от объекта.

**Files:**
- Modify: `internal/service/booking_service.go` — DisputeNoShow() добавить проверку расстояния
- Modify: `internal/domain/bathhouse.go` — убедиться что Latitude/Longitude доступны

- [ ] Haversine-функция или PostGIS ST_DWithin для проверки расстояния (200м)
- [ ] Если расстояние > 200м → отклонить спор автоматически или пометить как слабое доказательство
- [ ] Тесты с координатами внутри/вне радиуса
- [ ] run project test suite — must pass before Block B

---

## БЛОК B: ВАЖНЫЕ ФУНКЦИОНАЛЬНЫЕ ПРОБЕЛЫ

### Task B1: Индивидуальные цены по дням недели (FR-086)

**Проблема:** Текущая система поддерживает только weekday/weekend группировку. BRD требует
возможность задать разную цену для каждого дня (Пн, Вт, ..., Вс).

**Files:**
- Modify: `internal/domain/pricing.go` — новый RuleType `per_day` или расширение weekday/weekend для произвольных дней
- Modify: `internal/service/pricing_service.go` — калькуляция по конкретному дню
- Modify: `frontend/src/pages/pricing/PricingRules.tsx` — UI для задания цен по дням

- [ ] Расширить PricingRule: разрешить произвольный набор дней в DaysOfWeek для любого типа
- [ ] Или добавить RuleType `per_day` с валидацией
- [ ] Обновить CalculatePrice для матчинга по конкретному дню недели
- [ ] Frontend: 7-дневная сетка цен
- [ ] Тесты
- [ ] run project test suite — must pass before Task B2

### Task B2: Cron-задача для автовыплат (FR-102)

**Проблема:** Настройки автовыплат сохраняются, но нет cron-задачи, которая
автоматически инициирует выплату при превышении порога.

**Files:**
- Create: `internal/cron/auto_payout.go`
- Modify: `internal/cron/scheduler.go` — зарегистрировать задачу

- [ ] Cron-задача (каждый час): выбрать владельцев с auto_payout_threshold > 0 и balance >= threshold
- [ ] Для каждого: вызвать PayoutService.RequestPayout()
- [ ] Redis-лок для предотвращения дублей
- [ ] Тесты
- [ ] run project test suite — must pass before Task B3

### Task B3: Принудительное переключение в instant-режим при низком проценте ответов (FR-079)

**Проблема:** При response rate <30% за 60 дней отправляется только предупреждение. BRD требует
принудительный перевод в instant или деактивацию.

**Files:**
- Modify: `internal/domain/bathhouse.go` — поле low_response_rate_since (timestamp)
- Modify: `internal/service/booking_service.go` — CheckOwnerResponseRates: логика 60-дневного окна
- Create: `migrations/XXXXXX_low_response_rate_tracking.up.sql`

- [ ] Миграция: добавить low_response_rate_since в bathhouses
- [ ] При rate < 30%: если low_response_rate_since null → установить текущую дату
- [ ] Если low_response_rate_since > 60 дней → переключить booking_mode на instant или деактивировать
- [ ] При rate >= 30%: сбросить low_response_rate_since
- [ ] Тесты
- [ ] run project test suite — must pass before Task B4

### Task B4: Лояльность — кэшбэк на кошелёк (FR-112, FR-111)

**Проблема:** BRD упоминает cashback как тип зачисления на кошелёк (tag: cashback). Текущая
система только начисляет points, но не конвертирует их в реальный кэшбэк на кошелёк.

**Files:**
- Modify: `internal/service/loyalty_service.go` — после начисления points, зачислить cashback на кошелёк
- Modify: `internal/domain/loyalty.go` — определить % кэшбэка по уровню

- [ ] Определить механику: points → скидка ИЛИ cashback → кошелёк (BRD указывает cashback)
- [ ] Добавить зачисление на кошелёк с тегом "cashback" при завершении бронирования
- [ ] Кэшбэк % зависит от уровня лояльности (Bronze 0%, Silver 3%, Gold 5%, Platinum 10%)
- [ ] Тесты
- [ ] run project test suite — must pass before Task B5

### Task B5: Средняя цена по району на карточке объекта (FR-044)

**Проблема:** BRD требует показывать среднюю цену по району для сравнения на карточке.
Не реализовано.

**Files:**
- Modify: `internal/service/bathhouse_service.go` — метод GetAreaAveragePrice()
- Modify: `internal/handler/bathhouse_handler.go` — добавить в ответ GetByID
- Modify: `internal/repository/postgres/bathhouse_repo.go` — SQL-запрос средней цены в радиусе

- [ ] SQL: AVG(base_price) WHERE city_id = X AND status = active AND id != current
- [ ] Добавить поле area_average_price в ответ детальной карточки
- [ ] Redis-кеш (1ч) по city_id
- [ ] Frontend: отображение на странице BathhouseDetail
- [ ] Тесты
- [ ] run project test suite — must pass before Task B6

### Task B6: Региональный приветственный бонус (FR-016)

**Проблема:** BRD указывает 500₽ для РФ и 15 BYN для РБ. Сейчас только глобальная настройка 500₽.

**Files:**
- Modify: `internal/service/user_service.go` — начисление бонуса с учётом региона
- Modify: `internal/config/config.go` — отдельная настройка для BY

- [ ] Добавить конфиг BANI_WELCOME_BONUS_AMOUNT_BY (default 1500 = 15 BYN в копейках)
- [ ] При регистрации: определить регион → выбрать сумму
- [ ] Тесты
- [ ] run project test suite — must pass before Task B7

### Task B7: Мгновенные выплаты через СБП для владельцев (FR-101)

**Проблема:** СБП работает для приёма платежей, но не для выплат владельцам. BRD:
"мгновенно через СБП в РФ".

**Files:**
- Modify: `internal/payment/yookassa.go` — метод PayoutViaSBP (YooKassa Payouts API)
- Modify: `internal/service/payout_service.go` — выбор метода выплаты (SBP vs bank transfer)
- Modify: `internal/domain/payout.go` — добавить payout_method: sbp, bank_transfer

- [ ] Интеграция с YooKassa Payouts API для СБП-выплат
- [ ] Поле payout_method в domain + миграция
- [ ] Логика выбора: если owner из РФ и указан телефон → СБП, иначе bank transfer
- [ ] Тесты
- [ ] run project test suite — must pass before Block C

---

## БЛОК C: ПРОБЕЛЫ В ОБРАБОТКЕ МЕДИА И SEO

### Task C1: WebP-конвертация и множественные размеры фото (FR-023)

**Проблема:** Фото только в JPEG (1920px + 300px thumbnail). BRD требует: 4 размера (thumbnail, medium, large, full),
WebP-конвертация, blur-hash.

**Files:**
- Modify: `internal/service/media_service.go` — генерация 4 размеров, WebP, blur-hash
- Modify: `go.mod` — добавить зависимость для blur-hash (github.com/buckket/go-blurhash)

- [ ] 4 размера: thumbnail (300px), medium (800px), large (1200px), full (1920px)
- [ ] WebP-конвертация (golang.org/x/image/webp или imaging)
- [ ] Blur-hash генерация для placeholder
- [ ] Обновить domain/media: хранить пути ко всем размерам + blur_hash
- [ ] Миграция: добавить поля medium_url, large_url, blur_hash в media
- [ ] Тесты
- [ ] run project test suite — must pass before Task C2

### Task C2: Satellite-слой карты (FR-048)

**Проблема:** Карта поддерживает только стандартный вид. BRD требует: "два слоя — схема
и спутник".

**Files:**
- Modify: `frontend/src/components/BathhouseMap.tsx` — добавить переключатель слоёв

- [ ] Добавить Yandex Maps LayerSwitch (yandex.map.layer.satellite)
- [ ] Кнопка переключения "Схема / Спутник"
- [ ] Тесты
- [ ] run project test suite — must pass before Task C3

### Task C3: Frontend slug-based маршруты (FR-047, SEO)

**Проблема:** Backend поддерживает slug-маршруты (/city/bathhouse-slug), но frontend SPA использует UUID. Для
SEO и ЧПУ нужны slug-маршруты в SPA.

**Files:**
- Modify: `frontend/src/router.tsx` — slug-based routes
- Modify: `frontend/src/pages/client/BathhouseDetail.tsx` — загрузка по slug
- Modify: `frontend/src/pages/client/BathhouseSearch.tsx` — ссылки через slug

- [ ] Добавить маршрут /bathhouses/:slug в router
- [ ] BathhouseDetail: загрузка через getBySlug API
- [ ] Обновить все ссылки на bathhouse: использовать slug вместо UUID
- [ ] Тесты
- [ ] run project test suite — must pass before Task C4

### Task C4: Бейдж "Last minute" в результатах поиска (FR-090)

**Проблема:** Скидка last-minute рассчитывается, но бейдж не отображается в списке поиска.

**Files:**
- Modify: `internal/handler/bathhouse_handler.go` — добавить is_last_minute в ответ поиска
- Modify: `frontend/src/components/BathhouseCard.tsx` — бейдж "Last minute"

- [ ] В ответе поиска: для каждого объекта вычислить, есть ли last-minute слот сегодня
- [ ] Frontend: бейдж/тег "Last minute -X%" на карточке
- [ ] Тесты
- [ ] run project test suite — must pass before Block D

---

## БЛОК D: ПРОБЕЛЫ В АНАЛИТИКЕ И АДМИНКЕ

### Task D1: Операционные метрики поддержки (FR-155, раздел "Операционная")

**Проблема:** BRD требует CSAT, FCR (first contact resolution), AHT (average handling time), SLA compliance метрики для
поддержки. Нужно проверить полноту.

**Files:**
- Modify: `internal/service/ticket_service.go` — добавить FCR, AHT расчёт
- Modify: `internal/admin/pages/` — дашборд операционных метрик

- [ ] FCR: % тикетов закрытых после первого ответа оператора
- [ ] AHT: среднее время от создания тикета до закрытия
- [ ] Агрегация в admin API endpoint
- [ ] Frontend admin: виджет с метриками
- [ ] Тесты
- [ ] run project test suite — must pass before Task D2

### Task D2: Метрики предложения и спроса (FR-148, FR-149)

**Проблема:** BRD требует: ADR (средняя стоимость бронирования), DAU/MAU, CAC, LTV, churn. Нужно
проверить полноту аналитических эндпоинтов.

**Files:**
- Modify: `internal/service/analytics_service.go` — убедиться что все метрики реализованы
- Modify: `internal/handler/analytics_handler.go`

- [ ] Проверить наличие: ADR, DAU/MAU, churn rate (90 дней без визита)
- [ ] Добавить недостающие метрики
- [ ] Тесты
- [ ] run project test suite — must pass before Task D3

### Task D3: P&L и unit-экономика (FR-150, "Операционная")

**Проблема:** BRD требует: GMV, Take Rate, Revenue (сбор + подписки + продвижение), EBITDA, unit-экономика
(доход/расход на бронирование).

**Files:**
- Modify: `internal/service/analytics_service.go`
- Modify: `internal/admin/pages/finance.go` — виджет P&L

- [ ] GMV = сумма всех бронирований за период
- [ ] Take Rate = (сервисный сбор + подписки + продвижение) / GMV
- [ ] Unit economics: доход и расход на одно бронирование
- [ ] Admin UI: P&L дашборд
- [ ] Тесты
- [ ] run project test suite — must pass before Task D4

### Task D4: Тепловая карта спроса/предложения (FR-153)

**Проблема:** BRD требует географическую тепловую карту: где больше объектов vs где
больше запросов.

**Files:**
- Create: `internal/handler/analytics_handler.go` — эндпоинт heatmap data (или добавить метод в существующий)
- Create: `frontend/src/pages/admin/GeoHeatmap.tsx`

- [ ] Backend: агрегация по координатам (кластеризация поисковых запросов и объектов по ячейкам)
- [ ] API endpoint: GET /api/v1/admin/analytics/heatmap
- [ ] Frontend: тепловая карта на Yandex Maps
- [ ] Тесты
- [ ] run project test suite — must pass before Block E

---

## БЛОК E: МЕЛКИЕ РАСХОЖДЕНИЯ И ДОРАБОТКИ

### Task E1: Промокод типа "бесплатный add-on" (FR-099)

**Проблема:** Есть percentage, fixed_amount, free_hour. BRD упоминает "бесплатный add-on" как тип
промокода.

**Files:**
- Modify: `internal/domain/promo.go` — добавить PromoTypeFreeAddon
- Modify: `internal/service/promo_service.go` — логика применения
- Modify: `internal/service/booking_service.go` — обнуление цены выбранного add-on

- [ ] Новый тип промо: free_addon с полем target_addon_id
- [ ] При применении: скидка = цена указанного add-on
- [ ] Валидация: add-on должен быть в списке bathhouse add-ons
- [ ] Тесты
- [ ] run project test suite — must pass before Task E2

### Task E2: Фискализация — адаптация по типу налогового статуса (FR-105)

**Проблема:** Фискальные чеки формируются одинаково для всех. BRD: система адаптирует
формы отчётности под налоговый статус (ИП, самозанятый, физлицо, юрлицо).

**Files:**
- Modify: `internal/fiscal/atol.go` — разные НДС-ставки и параметры по entity_type

- [ ] Определить НДС-ставку по entity_type (физлицо: без НДС, ИП/юрлицо: 20%, самозанятый: НПД)
- [ ] Передавать корректный tax_system в ATOL
- [ ] Тесты
- [ ] run project test suite — must pass before Task E3

### Task E3: iCal export endpoint (FR-074)

**Проблема:** iCal-генератор существует, но HTTP-эндпоинт для скачивания .ics отсутствует
или не зарегистрирован.

**Files:**
- Modify: `internal/handler/calendar_handler.go` — зарегистрировать GET /api/v1/my/bathhouses/{id}/calendar.ics
- Modify: `internal/server/router.go`

- [ ] Проверить наличие маршрута calendar.ics
- [ ] Если отсутствует — зарегистрировать, Content-Type: text/calendar
- [ ] Тест
- [ ] run project test suite — must pass before Task E4

### Task E4: Пустые состояния (empty states) на всех экранах (FR раздел 2.15, 2.18)

**Проблема:** BRD требует информативные пустые состояния на каждом экране без данных с
призывом к действию.

**Files:**
- Modify: множество frontend компонентов (BookingList, ReviewList, Favorites, Wallet, CRM pages, etc.)

- [ ] Аудит всех страниц: проверить наличие empty state
- [ ] Добавить недостающие empty states с поясняющим текстом и CTA-кнопкой
- [ ] Использовать Ant Design Empty компонент
- [ ] run project test suite — must pass before Task E5

### Task E5: Счётчик результатов в реальном времени при фильтрации (FR-040)

**Проблема:** BRD: "Счётчик результатов обновляется при каждом изменении фильтра".
Проверить реализацию.

**Files:**
- Modify: `frontend/src/pages/client/BathhouseSearch.tsx`

- [ ] Проверить наличие live-счётчика при смене каждого фильтра
- [ ] Если отсутствует — добавить debounced запрос total_count при изменении фильтров
- [ ] Тест
- [ ] run project test suite — must pass before Task E6

### Task E6: Ограничение 3 promoted объектов на страницу поиска (FR-043)

**Проблема:** BRD: "на одной странице результатов не более 3 продвинутых объектов".
Проверить реализацию лимита.

**Files:**
- Modify: `internal/repository/postgres/bathhouse_repo.go` — ограничение promoted в выдаче

- [ ] Проверить SQL-запрос поиска: лимитированы ли promoted записи до 3 на страницу
- [ ] Если нет — добавить WITH promoted AS (... LIMIT 3) UNION ALL unpromoted
- [ ] Тест
- [ ] run project test suite — must pass before Block F

---

## БЛОК F: ФИНАЛЬНАЯ ВЕРИФИКАЦИЯ

### Task F1: Прогон тестов и линтера

- [ ] `go test ./... -race`
- [ ] `make lint`
- [ ] `cd frontend && npm run lint`
- [ ] `cd frontend && npx vitest run`

### Task F2: Обновление swagger-документации

- [ ] `make swagger` для всех новых/изменённых эндпоинтов
- [ ] Regenerate frontend API client: `make frontend-generate-api`

### Task F3: Обновление CLAUDE.md

- [ ] Добавить описание новых подсистем (booking modification requests, extension requests, auto-payout cron, etc.)
- [ ] Обновить список ошибок в Error Mapping

### Task F4: Перенос плана в completed

- [ ] Переместить этот план в `docs/plans/completed/`
