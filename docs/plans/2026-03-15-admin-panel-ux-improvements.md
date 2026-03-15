# Admin Panel UX Improvements - Расширенный план

## Overview

Комплексное улучшение пользовательского опыта админ-панели:
устранение критических UX-проблем
(разрозненные страницы без навигации, reload после каждого действия, alert-ы вместо
нормальных уведомлений), добавление новых функций (CSV экспорт, расширенный мониторинг,
аудит действий), исправление багов (XSS в графиках, некорректные статусы Redis,
нелокализованные статусы) и визуальная унификация всех разделов.

## Context

- Files involved:
  - `internal/admin/pages/dashboard.go`, `internal/admin/pages/templates/dashboard.tmpl`
  - `internal/admin/pages/moderation.go`, `internal/admin/pages/templates/moderation.tmpl`
  - `internal/admin/pages/analytics.go`, `internal/admin/pages/templates/analytics.tmpl`
  - `internal/admin/pages/health.go`, `internal/admin/pages/templates/health.tmpl`
  - `internal/admin/pages/templates/base.tmpl` (новый - shared layout)
  - `internal/admin/pages/templates/components.tmpl` (новый - переиспользуемые компоненты)
  - `internal/admin/engine.go` (menu, routing)
  - `internal/admin/pages/*_test.go` (tests for all pages)
- Related patterns: GoAdmin framework, html/template, Chart.js, vanilla JS fetch API
- Dependencies: no new external dependencies (Chart.js будет self-hosted вместо CDN)

## Development Approach

- **Testing approach**: Regular (code first, then tests)
- Complete each task fully before moving to the next
- **CRITICAL: every task MUST include new/updated tests**
- **CRITICAL: all tests must pass before starting next task**

## Implementation Steps

### Task 1: Shared base layout и компонентная система шаблонов

Сейчас каждая из 4 страниц - полностью самостоятельный HTML-документ с дублированным CSS.
Навигация между страницами отсутствует - при переходе с GoAdmin на кастомную страницу
пропадает sidebar. Визуальный разрыв между GoAdmin-страницами и кастомными страницами.

**Files:**
- Create: `internal/admin/pages/templates/base.tmpl`
- Create: `internal/admin/pages/templates/components.tmpl`
- Modify: `internal/admin/pages/templates/dashboard.tmpl`
- Modify: `internal/admin/pages/templates/moderation.tmpl`
- Modify: `internal/admin/pages/templates/analytics.tmpl`
- Modify: `internal/admin/pages/templates/health.tmpl`
- Modify: `internal/admin/pages/dashboard.go`
- Modify: `internal/admin/pages/moderation.go`
- Modify: `internal/admin/pages/analytics.go`
- Modify: `internal/admin/pages/health.go`

- [x] Создать `base.tmpl` с общей HTML-структурой: `<!DOCTYPE html>`, `<head>` (meta, title, shared CSS),
    `<body>` с sidebar навигацией (ссылки на все 4 страницы + GoAdmin), breadcrumb, content block, footer
- [x] Вынести в `base.tmpl` общий CSS: reset, typography, `.container`, `.btn`, `.badge`, `.filter-bar`,
    `.filter-group`, card styles, grid system - убрать дублирование из 4 шаблонов
- [x] Создать `components.tmpl` с переиспользуемыми блоками:
    - `{{define "toast"}}` - система toast-уведомлений (success/error/warning, auto-dismiss 3s)
    - `{{define "status-badge"}}` - цветной бейдж статуса с локализацией (EN→RU)
    - `{{define "empty-state"}}` - пустое состояние с иконкой и текстом
    - `{{define "loading-spinner"}}` - анимация загрузки
    - `{{define "pagination"}}` - универсальный пагинатор
- [x] Добавить боковую навигацию в base.tmpl: список ссылок (Дашборд, Модерация, Аналитика,
    Мониторинг) с иконками, активный пункт подсвечивается, ссылка "Назад в GoAdmin"
- [x] Добавить breadcrumb навигацию: Главная > Операции > Текущая страница
- [x] Добавить динамический `<title>` для каждой страницы (сейчас - generic)
- [x] Добавить footer с "Последнее обновление: {{.GeneratedAt}}" на всех страницах
- [x] Унифицировать `max-width` контейнера (сейчас 1200px на health, 1400px на остальных)
- [x] Перевести все 4 шаблона на использование `base.tmpl` через `{{template "base" .}}`
    и `{{define "content"}}...{{end}}`
- [x] Обновить парсинг шаблонов во всех 4 handler-ах: ParseFS должен включать base.tmpl и
    components.tmpl вместе с page-specific шаблоном
- [x] Write tests: проверить рендеринг каждой страницы с новым layout, наличие sidebar,
    breadcrumb, title, footer в HTML output
- [x] Run project test suite - must pass before task 2

### Task 2: Локализация статусов и исправление XSS

Статусы бронирований и отзывов отображаются на английском (pending, confirmed, approved).
В analytics шаблоне названия бань вставляются в JS без экранирования - XSS вектор.
Redis в статусе "unconfigured" отображается как "Недоступен".

**Files:**
- Modify: `internal/admin/pages/templates/dashboard.tmpl`
- Modify: `internal/admin/pages/templates/moderation.tmpl`
- Modify: `internal/admin/pages/templates/analytics.tmpl`
- Modify: `internal/admin/pages/templates/health.tmpl`
- Modify: `internal/admin/pages/dashboard.go`
- Modify: `internal/admin/pages/analytics.go`
- Modify: `internal/admin/pages/health.go`

- [x] Создать template function `statusRu` для маппинга статусов EN→RU:
    pending→"Ожидает", confirmed→"Подтверждено", approved→"Одобрено", rejected→"Отклонено",
    cancelled→"Отменено", completed→"Завершено", hidden→"Скрыто"
- [x] Применить `statusRu` ко всем бейджам статусов на dashboard и moderation
- [x] Исправить XSS в analytics: экранировать `.Name` в chart.js данных через template
    function `jsEscape` (заменить `"`, `\`, `<`, `>` и т.д.)
- [x] Исправить health page: добавить третье состояние "unconfigured" для Redis -
    показывать "Не настроен" (серый бейдж) вместо "Недоступен" (красный)
- [x] Показывать реальное сообщение об ошибке в health page service cards вместо
    generic "Сервис недоступен" - отображать `{{.Error}}` в collapsed блоке
- [x] Исправить экранирование `pagesPrefix` в JS на moderation page (использовать
    template function для безопасной вставки в JS строку)
- [x] Write tests: проверить маппинг всех статусов, XSS экранирование для спецсимволов
    в названиях бань, корректное отображение 3 состояний Redis
- [x] Run project test suite - must pass before task 3

### Task 3: Moderation - inline AJAX без перезагрузки страницы

Approve/reject вызывают полную перезагрузку страницы. Ошибки показываются через alert().
Нет подтверждения для batch approve. Нет lightbox для картинок. Нет кастомных причин отказа.

**Files:**
- Modify: `internal/admin/pages/templates/moderation.tmpl`
- Modify: `internal/admin/pages/moderation.go`

- [x] Заменить `location.reload()` после approve/reject на DOM-манипуляцию:
    - При approve: плавно скрыть карточку (CSS transition opacity→0, затем remove),
      обновить счетчики в stats-cards
    - При reject: аналогично, обновить счетчики
    - Обновить номер текущей страницы и total если нужно
- [x] Реализовать toast-уведомления вместо alert():
    - Успех: зеленый toast "Отзыв одобрен" / "Отзыв отклонен" (auto-dismiss 3s)
    - Ошибка: красный toast с текстом ошибки от сервера (auto-dismiss 5s)
    - Для сетевых ошибок: красный toast "Ошибка сети. Попробуйте снова."
- [x] Добавить loading state на кнопки действий:
    - При клике: кнопка disabled + spinner внутри
    - После завершения: восстановить кнопку (или удалить карточку при успехе)
    - Предотвращение двойных кликов
- [x] Добавить confirmation dialog для batch approve (сейчас есть только для reject):
    "Одобрить N отзывов? Это действие нельзя отменить."
- [x] Динамически обновлять stats-карточки после каждого действия:
    pending -1, approved/rejected +1, без перезагрузки
- [x] Для batch операций: удалять обработанные карточки из DOM по одной с анимацией,
    показывать toast "Обработано X из Y отзывов" (учитывая что часть могла быть
    уже обработана другим админом)
- [x] Добавить поле free-text в reject modal: textarea "Дополнительный комментарий"
    (необязательное) помимо чекбоксов с причинами
- [x] Добавить валидацию: при отклонении должна быть выбрана хотя бы одна причина
    или заполнен комментарий (сейчас можно отклонить без причины)
- [x] Добавить lightbox для изображений отзывов: клик по 80x80 thumbnail открывает
    полноразмерное изображение в overlay с кнопкой закрытия и навигацией стрелками
- [x] Добавить опцию "Скрытые" в фильтр по статусу (CSS стили для hidden уже есть,
    но в select нет такой опции)
- [x] Добавить усечение длинного текста отзыва: показывать первые 200 символов
    с кнопкой "Показать полностью" (expand/collapse)
- [x] Write tests: проверить JSON-ответы API для inline обновлений, валидацию причин
    отклонения, корректность новых response fields
- [x] Run project test suite - must pass before task 4

### Task 4: Dashboard - кликабельные KPI, тренды и улучшенная лента активности

KPI-карточки статичны и не кликабельны. Нет трендов (сравнение с предыдущим периодом).
Статусы бронирований на английском. Нет автообновления. Activity feed не информативен.

**Files:**
- Modify: `internal/admin/pages/templates/dashboard.tmpl`
- Modify: `internal/admin/pages/dashboard.go`

- [x] Сделать KPI-карточки кликабельными:
    - "Ожидающие бани" → ссылка на moderation page с фильтром (или GoAdmin bathhouses
      с фильтром pending, исправив текущий некорректный URL с `__goadmin_edit_pk=`)
    - "Ожидающие отзывы" → ссылка на moderation page с status=pending
    - "Бронирования сегодня" → ссылка на analytics с date_from=today
    - "Пользователи" → ссылка на GoAdmin users list
- [x] Добавить индикаторы трендов на KPI-карточки:
    - Рассчитывать процент изменения vs предыдущий период (вчера, прошлая неделя, прошлый месяц)
    - Отображать ↑12% (зеленый) или ↓5% (красный) рядом с числом
    - Добавить SQL запросы для получения данных предыдущего периода
- [x] Улучшить карточку выручки: разделить Today/Week/Month на 3 отдельные sub-cards
    вместо одной строки через "/" (сейчас трудно читать)
- [x] Добавить относительное время в ленте активности: "5 мин назад", "2 часа назад"
    (Go template function `relativeTime` + JS для обновления без reload)
- [x] Добавить preview текста отзыва в ленте последних отзывов (первые 100 символов)
- [x] Добавить empty-state для каждой ленты: вместо "Нет данных" показывать
    иконку + текст "За выбранный период нет бронирований"
- [x] Добавить toggle автообновления: кнопка play/pause, интервал 60 сек, по умолчанию
    выключено, рядом текст "Обновлено: HH:MM:SS"
- [x] Применить `text-truncate` к именам пользователей (сейчас только на банях)
- [x] Добавить быструю ссылку "Посмотреть все" внизу каждой ленты активности
    (→ на соответствующий раздел GoAdmin)
- [x] Write tests: проверить расчет трендов, форматирование relative time,
    корректность ссылок на KPI карточках, новые SQL запросы для prev period
- [x] Run project test suite - must pass before task 5

### Task 5: Analytics - быстрые пресеты дат, CSV экспорт и сводка

Нет быстрого выбора периода (надо вручную вводить даты). Нет экспорта данных.
Нет суммарных чисел - только графики. Chart.js загружается с CDN (может быть недоступен).
Фильтр по городу не применяется к графику новых пользователей.

**Files:**
- Modify: `internal/admin/pages/analytics.go`
- Modify: `internal/admin/pages/templates/analytics.tmpl`
- Create: `internal/admin/pages/static/chart.min.js` (self-hosted Chart.js)

**5a: Quick date presets и UI улучшения**

- [x] Добавить кнопки быстрого выбора периода над date picker:
    "Сегодня", "7 дней", "30 дней", "Этот месяц", "Прошлый месяц", "Этот год"
    - Каждая кнопка - ссылка с предвычисленными date_from/date_to query params
    - Активная кнопка подсвечена (определяется по совпадению дат в фильтре)
- [x] Добавить сводную строку под графиками: "Итого за период: X бронирований,
    Y ₽ выручки, Z новых пользователей" - крупный шрифт, отдельный блок
- [x] Добавить процент изменения в сводке: сравнение с предыдущим аналогичным
    периодом (например, если выбрано 7 дней, сравнить с предыдущими 7 днями)
    - ↑12% зеленый / ↓5% красный / = 0% серый
- [x] Исправить: применить фильтр по городу к графику новых пользователей
    (сейчас `loadNewUsersPerDay` игнорирует `cityID`)

**5b: CSV Export**

- [x] Добавить CSV export endpoint: `GET /pages/analytics/export`
    - Query params: `type` (bookings|revenue|users|top_bookings|top_revenue),
      `from`, `to`, `city_id`
    - Response: `Content-Type: text/csv`, `Content-Disposition: attachment;
      filename="analytics_bookings_2026-03-01_2026-03-15.csv"`
    - Формат CSV: заголовок + данные, BOM для корректного открытия в Excel
- [x] Добавить кнопку "Скачать CSV" (иконка download) рядом с каждым графиком
    - Кнопка формирует URL с текущими фильтрами и type для этого графика
- [x] Добавить маршрут в engine.go: `GET /pages/analytics/export`

**5c: Self-hosted Chart.js**

- [x] Скопировать Chart.js 4.4.7 UMD bundle в `internal/admin/pages/static/chart.min.js`
- [x] Добавить route для раздачи статики: `GET /pages/static/*`
- [x] Заменить CDN ссылку в analytics.tmpl на локальный путь
- [x] Обновить embed.FS для включения static директории
- [x] Write tests: CSV export endpoint (формат, headers, фильтрация), date presets
    (корректность дат для каждого пресета), period comparison расчеты
- [x] Run project test suite - must pass before task 6

### Task 6: Health Monitor - расширенный мониторинг и системные метрики

Только 2 сервиса мониторятся. Нет системных метрик (память, горутины). Нет истории
состояний. Авто-обновление агрессивное (30 сек) без возможности паузы. Backlog
не линкуется на модерацию.

**Files:**
- Modify: `internal/admin/pages/health.go`
- Modify: `internal/admin/pages/templates/health.tmpl`

**6a: Системные метрики**

- [x] Добавить секцию "Система" с метриками из `runtime` пакета:
    - Uptime сервера (время с момента запуска)
    - Go version (`runtime.Version()`)
    - Количество горутин (`runtime.NumGoroutine()`)
    - Использование памяти: Alloc, TotalAlloc, Sys, NumGC (`runtime.MemStats`)
    - Количество CPU (`runtime.NumCPU()`)
- [x] Добавить метрики connection pool из pgxpool:
    - Active connections / Max connections (progress bar с процентом)
    - Idle connections
    - Total connections acquired
    - Wait count и wait duration
    - Конструктивные прогресс бары (зеленый <50%, желтый 50-80%, красный >80%)

**6b: Улучшение UX мониторинга**

- [x] Color-code backlog с порогами severity:
    - 0: зеленый (ok)
    - 1-5: желтый (warning)
    - 6-20: оранжевый (elevated)
    - >20: красный (critical)
    - Добавить пульсирующую анимацию для critical
- [x] Сделать числа backlog кликабельными → ссылка на модерацию с предфильтром
    (pending + created_at <= порог)
- [x] Улучшить auto-refresh UX:
    - Кнопка play/pause (сейчас нельзя остановить)
    - Настраиваемый интервал: 15/30/60 секунд (dropdown)
    - Показывать "Последнее обновление: HH:MM:SS" prominently
    - Countdown bar вместо текстового счетчика
- [x] Показывать latency как "2 ms" / "150 ms" (человекочитаемый формат)
    вместо Go-формата `2ms` / `1.234µs`
- [x] Добавить visual status indicator: зеленая точка (пульсирующая) для up,
    красная для down, серая для unconfigured

**6c: Дополнительные проверки**

- [x] Добавить проверку доступности файловой системы (запись temp файла)
- [x] Добавить отображение размера БД: `SELECT pg_database_size(current_database())`
- [x] Показывать количество активных WebSocket соединений с контекстом
    (min/max за последний час, если данные доступны)
- [x] Write tests: системные метрики (runtime данные), pool stats, threshold
    color-coding logic, новые health checks, форматирование latency
- [x] Run project test suite - must pass before task 7

### Task 7: Keyboard shortcuts и accessibility для модерации

Модерация - основной рабочий инструмент администратора. Keyboard shortcuts
значительно ускорят обработку очереди.

**Files:**
- Modify: `internal/admin/pages/templates/moderation.tmpl`

- [x] Добавить keyboard shortcuts:
    - `a` — одобрить выбранные (или текущий focused если ничего не выбрано)
    - `r` — открыть reject modal для выбранных
    - `Escape` — закрыть reject modal
    - `Enter` в reject modal — подтвердить отклонение
    - `Ctrl+A` — выбрать все на странице
    - `→` / `←` — следующая/предыдущая страница пагинации
- [x] Добавить focus management: Tab навигация по карточкам, outline для
    focused карточки
- [x] Добавить подсказку по горячим клавишам: маленькая иконка "?" в правом
    нижнем углу, при клике/hover показывает список shortcuts
- [x] Write tests: проверить наличие keyboard event listeners в шаблоне,
    корректность data-атрибутов для shortcuts
- [x] Run project test suite - must pass before task 8

### Task 8: Responsive design для всех страниц

Сейчас все страницы используют CSS Grid с `auto-fit/minmax()` без media queries.
На мобильных устройствах графики и таблицы выходят за пределы экрана.

**Files:**
- Modify: `internal/admin/pages/templates/base.tmpl` (shared CSS)
- Modify: `internal/admin/pages/templates/dashboard.tmpl`
- Modify: `internal/admin/pages/templates/analytics.tmpl`
- Modify: `internal/admin/pages/templates/moderation.tmpl`

- [ ] Добавить media queries в shared CSS:
    - `@media (max-width: 768px)`: sidebar collapse в hamburger menu,
      single column layout для всех grid-ов
    - `@media (max-width: 1024px)`: 2-column layout для KPI карточек,
      уменьшить minmax значения для графиков
- [ ] Dashboard: уменьшить `minmax(400px, 1fr)` до `minmax(280px, 1fr)` для feed-grid
- [ ] Analytics: уменьшить `minmax(500px, 1fr)` до `minmax(300px, 1fr)` для chart-grid,
    добавить `overflow-x: auto` для графиков которые не помещаются
- [ ] Moderation: адаптировать filter-bar для вертикального layout на мобильных,
    image thumbnails уменьшить до 60x60 на мобильных
- [ ] Таблицы: добавить `overflow-x: auto` wrapper для горизонтального скролла
    на узких экранах
- [ ] Write tests: проверить наличие media queries в rendered HTML
- [ ] Run project test suite - must pass before task 9

### Task 9: Verify acceptance criteria

- [ ] Manual test: навигация через sidebar между всеми 4 страницами
- [ ] Manual test: breadcrumb навигация работает корректно
- [ ] Manual test: approve/reject отзыва без перезагрузки страницы, toast уведомление
- [ ] Manual test: batch approve с confirmation dialog
- [ ] Manual test: reject с валидацией причин (нельзя отклонить без причины)
- [ ] Manual test: lightbox для картинок отзывов
- [ ] Manual test: keyboard shortcuts на модерации (a, r, Escape, стрелки)
- [ ] Manual test: CSV export из analytics (открыть в Excel, проверить кодировку)
- [ ] Manual test: date presets на analytics (7 дней, 30 дней, этот месяц)
- [ ] Manual test: health page показывает системные метрики, pool stats, 3 состояния Redis
- [ ] Manual test: health page pause/resume auto-refresh
- [ ] Manual test: dashboard KPI кликабельны и ведут на правильные страницы
- [ ] Manual test: dashboard тренды отображаются корректно (↑/↓ процент)
- [ ] Manual test: все статусы отображаются на русском языке
- [ ] Manual test: responsive - проверить на 375px, 768px, 1024px, 1440px viewports
- [ ] Run full test suite (`make test`)
- [ ] Run linter (`make lint`)
- [ ] Verify test coverage meets 80%+

### Task 10: Update documentation

- [ ] Update CLAUDE.md: добавить описание shared template system (base.tmpl, components.tmpl),
    new routes (analytics/export, static/*), keyboard shortcuts
- [ ] Move this plan to `docs/plans/completed/`
