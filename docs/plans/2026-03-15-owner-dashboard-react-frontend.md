# Фронтенд личного кабинета владельца бань (React + Ant Design + orval)

## Overview

React SPA для владельцев бань: управление банями, бронированиями, отзывами, ценами, промокодами, аналитикой, чатом и настройками. API-клиент генерируется из OpenAPI спецификации через orval (TanStack Query хуки + типы).

## Context

- OpenAPI spec: `docs/swagger.json` / `docs/swagger.yaml`
- Backend API base: `/api/v1`, все ответы в формате `{success, data, error, meta}`
- Цены в копейках (int64), день недели 0=Пн..6=Вс
- Роли: owner, representative (оба имеют доступ к большинству эндпоинтов ЛК)
- Существующего фронтенда нет - создаем с нуля
- Директория: `frontend/` в корне проекта

## Tech Stack

- Vite + React 18 + TypeScript
- Ant Design 5 (UI-компоненты)
- React Router 6 (маршрутизация)
- TanStack Query (серверное состояние, через orval-генерацию)
- orval (генерация API-клиента из OpenAPI)
- dayjs (даты, встроен в antd)
- zustand (минимальный клиентский стейт: auth токен, текущий пользователь)

## Development Approach

- **Testing approach**: Regular (code first, then tests)
- Каждый таск - самостоятельный функциональный блок
- Полностью завершаем таск перед переходом к следующему
- **CRITICAL: every task MUST include new/updated tests**
- **CRITICAL: all tests must pass before starting next task**

## Implementation Steps

### Task 1: Scaffolding проекта

**Files:**
- Create: `frontend/package.json`
- Create: `frontend/vite.config.ts`
- Create: `frontend/tsconfig.json`
- Create: `frontend/index.html`
- Create: `frontend/src/main.tsx`
- Create: `frontend/src/App.tsx`
- Create: `frontend/.eslintrc.cjs`

- [x] Инициализировать Vite проект с React + TypeScript
- [x] Установить зависимости: antd, react-router-dom, @tanstack/react-query, zustand, dayjs
- [x] Установить dev-зависимости: orval, @tanstack/react-query-devtools
- [x] Настроить vite.config.ts с proxy на `/api` -> `http://localhost:8080`
- [x] Настроить базовый tsconfig.json с path aliases (`@/` -> `src/`)
- [x] Создать точку входа main.tsx с провайдерами (QueryClient, Router, ConfigProvider antd с русской локалью)
- [x] Добавить в корневой Makefile команды: `make frontend-dev`, `make frontend-build`, `make frontend-generate-api`
- [x] Проверить что `npm run dev` запускается без ошибок

### Task 2: Генерация API-клиента через orval

**Files:**
- Create: `frontend/orval.config.ts`
- Create: `frontend/src/api/` (автогенерация)
- Create: `frontend/src/api/axios-instance.ts`

- [x] Создать orval.config.ts: input из `../docs/swagger.json`, output в `src/api/generated/`
- [x] Настроить orval: генерировать React Query хуки, TypeScript типы, использовать кастомный axios instance
- [x] Создать axios-instance.ts: baseURL `/api/v1`, interceptor для JWT токена из localStorage, interceptor для обработки 401 (редирект на login)
- [x] Запустить `npx orval` и убедиться что API-клиент сгенерирован
- [x] Добавить npm script `generate:api` в package.json
- [x] Создать утилиту `src/lib/format.ts`: formatPrice (копейки -> рубли), formatDayOfWeek (0=Пн), formatDateTime

### Task 3: Аутентификация и auth store

**Files:**
- Create: `frontend/src/stores/auth.ts`
- Create: `frontend/src/pages/Login.tsx`
- Create: `frontend/src/pages/Register.tsx`
- Create: `frontend/src/components/ProtectedRoute.tsx`

- [x] Создать zustand store для auth: user, token, login(), logout(), isAuthenticated
- [x] При инициализации проверять JWT в localStorage, загружать профиль через GET /auth/me
- [x] Страница Login: форма email + password, вызов POST /auth/login, сохранение JWT
- [x] Страница Register: форма с role=owner, email, password, name, phone
- [x] ProtectedRoute: редирект на /login если не авторизован, проверка роли owner/representative
- [x] Обработка ошибок авторизации: показ antd message при неверных данных

### Task 4: Layout и навигация

**Files:**
- Create: `frontend/src/components/AppLayout.tsx`
- Create: `frontend/src/components/BathhouseSelector.tsx`
- Create: `frontend/src/router.tsx`

- [ ] AppLayout: antd Layout с Sider (sidebar меню), Header (имя пользователя, аватар, выход), Content
- [ ] Sidebar меню: Дашборд, Бани, Бронирования, Отзывы, Календарь, Цены, Промокоды, Чат, Представители, Подписки, Виджет, Настройки профиля
- [ ] BathhouseSelector: выпадающий список бань владельца (GET /my/bathhouses) в header, сохранение выбранной бани в zustand
- [ ] Настроить React Router: вложенные маршруты внутри AppLayout
- [ ] Мобильная адаптивность sidebar (collapsible)
- [ ] Breadcrumbs на основе текущего маршрута

### Task 5: Дашборд (аналитика)

**Files:**
- Create: `frontend/src/pages/Dashboard.tsx`

- [ ] Карточки KPI: бронирования, выручка, просмотры, рейтинг, средний чек, конверсия (GET /my/bathhouses/{id}/analytics)
- [ ] Индикаторы изменений (+/-% по сравнению с прошлым периодом) с цветовой индикацией
- [ ] Переключатель периода: 1д, 7д, 30д, 90д
- [ ] Если нет выбранной бани - показать сводку или подсказку выбрать баню
- [ ] Обработка состояний загрузки (antd Skeleton) и ошибок

### Task 6: Управление банями

**Files:**
- Create: `frontend/src/pages/bathhouses/BathhouseList.tsx`
- Create: `frontend/src/pages/bathhouses/BathhouseForm.tsx`

- [ ] BathhouseList: таблица бань (GET /my/bathhouses) со статусом, рейтингом, числом отзывов
- [ ] Кнопка создания бани (только для owner)
- [ ] BathhouseForm: создание/редактирование с полями: название, описание, адрес, город, координаты, цена за час, мин. длительность, макс. гостей, удобства (чекбоксы), изображения (список URL)
- [ ] Настройка рабочих часов: таблица по дням недели (0=Пн..6=Вс) с временем открытия/закрытия
- [ ] Статус бани (pending/active/rejected) - отображение с Badge
- [ ] Удаление бани с подтверждением (Popconfirm)

### Task 7: Управление бронированиями

**Files:**
- Create: `frontend/src/pages/bookings/BookingList.tsx`
- Create: `frontend/src/pages/bookings/BookingDetails.tsx`

- [ ] BookingList: таблица бронирований (GET /bathhouses/{id}/bookings) с пагинацией
- [ ] Колонки: дата/время, гость, кол-во гостей, сумма, статус, действия
- [ ] Фильтры: по статусу (pending/confirmed/completed/cancelled/rejected), по дате
- [ ] Действия: подтвердить (confirm), отклонить (reject), завершить (complete), отменить (cancel)
- [ ] BookingDetails: модальное окно с полной информацией о бронировании и платеже
- [ ] Цветовая кодировка статусов (Tag с цветами)

### Task 8: Календарь и расписание

**Files:**
- Create: `frontend/src/pages/calendar/CalendarPage.tsx`

- [ ] Отображение бронирований в виде календаря (antd Calendar или недельный вид)
- [ ] Визуализация слотов: занятые (бронирования), заблокированные (slot-blocks), свободные
- [ ] Создание блокировки слотов: модальное окно с выбором периода и описанием (POST /my/bathhouses/{id}/slot-blocks)
- [ ] Удаление блокировки
- [ ] Экспорт iCal: кнопка получения ссылки (GET /my/bathhouses/{id}/calendar-token)
- [ ] Подключение внешних календарей: форма добавления URL (Google Calendar, Airbnb)
- [ ] Список внешних календарей с кнопкой синхронизации и удаления

### Task 9: Управление отзывами

**Files:**
- Create: `frontend/src/pages/reviews/ReviewList.tsx`

- [ ] Список отзывов по бане (GET /bathhouses/{id}/reviews) с пагинацией
- [ ] Отображение: рейтинг (звёзды), текст, фото/видео, дата, ответ владельца
- [ ] Форма ответа на отзыв: текстовое поле + кнопка отправки (POST /reviews/{id}/response)
- [ ] Фильтр: без ответа / с ответом
- [ ] Статистика: средний рейтинг, распределение оценок

### Task 10: Правила ценообразования

**Files:**
- Create: `frontend/src/pages/pricing/PricingRules.tsx`

- [ ] Список правил (GET /my/bathhouses/{id}/pricing-rules) в виде таблицы
- [ ] Создание/редактирование правила: модальная форма с полями name, type (день недели/время суток/диапазон дат), multiplier, условия в зависимости от типа
- [ ] Приоритет правил (drag-and-drop или числовой ввод)
- [ ] Переключатель активности (is_active)
- [ ] Удаление правила с подтверждением

### Task 11: Промокоды

**Files:**
- Create: `frontend/src/pages/promo/PromoList.tsx`

- [ ] Список промокодов бани (GET /my/bathhouses/{id}/promo-codes) с пагинацией
- [ ] Создание промокода: форма с code, type (percentage/fixed_amount/free_hour), value, max_uses, min_amount, validity period
- [ ] Отображение: код, тип скидки, использовано/максимум, период, статус
- [ ] Деактивация промокода (DELETE /promo-codes/{id})
- [ ] Копирование кода в буфер обмена

### Task 12: Управление представителями

**Files:**
- Create: `frontend/src/pages/representatives/RepresentativeList.tsx`

- [ ] Список представителей бани (GET /bathhouses/{id}/representatives)
- [ ] Приглашение нового представителя: модалка с полем email (POST /bathhouses/{id}/representatives)
- [ ] Удаление представителя с подтверждением (DELETE /representatives/{id})
- [ ] Доступно только для роли owner

### Task 13: Чат

**Files:**
- Create: `frontend/src/pages/chat/ChatPage.tsx`
- Create: `frontend/src/pages/chat/ConversationList.tsx`
- Create: `frontend/src/pages/chat/MessageArea.tsx`

- [ ] Двухпанельный layout: список бесед слева, сообщения справа
- [ ] ConversationList: список бесед (GET /my/conversations) с последним сообщением и непрочитанными
- [ ] MessageArea: история сообщений (GET /conversations/{id}/messages) со скроллом, отправка (POST /conversations/{id}/messages)
- [ ] Отметка прочитанным (PATCH /conversations/{id}/read)
- [ ] Счётчик непрочитанных в sidebar (GET /my/unread-messages-count)
- [ ] WebSocket подключение для real-time обновлений (GET /ws/notifications с JWT)

### Task 14: Уведомления и настройки профиля

**Files:**
- Create: `frontend/src/pages/notifications/NotificationList.tsx`
- Create: `frontend/src/pages/settings/ProfileSettings.tsx`
- Create: `frontend/src/components/NotificationBell.tsx`

- [ ] NotificationBell: иконка в header со счётчиком непрочитанных, dropdown со списком
- [ ] NotificationList: полный список уведомлений с пагинацией, кнопка "прочитать все"
- [ ] ProfileSettings: редактирование профиля (имя, телефон, bio, город), загрузка/удаление аватара
- [ ] Настройки уведомлений: переключатели каналов (in_app, email, push) и событий (booking, review, promo, reminders)
- [ ] Привязка/отвязка социальных аккаунтов (VK, Yandex, Google)

### Task 15: Подписки и виджет

**Files:**
- Create: `frontend/src/pages/subscriptions/SubscriptionPage.tsx`
- Create: `frontend/src/pages/widget/WidgetSettings.tsx`

- [ ] SubscriptionPage: текущая подписка бани, выбор плана (free/premium/promoted), отмена
- [ ] Список всех подписок (GET /my/subscriptions)
- [ ] Промо-кампании: создание (бюджет, длительность, целевой город), статистика (показы, клики, потрачено)
- [ ] WidgetSettings: получение embed-кода (GET /my/bathhouses/{id}/widget-code), предпросмотр
- [ ] Кнопка перегенерации API-ключа (POST /my/bathhouses/{id}/widget-key/regenerate) с подтверждением
- [ ] Настройки виджета: цвет, шрифт, показывать цену/рейтинг

### Task 16: Фото-менеджмент

**Files:**
- Create: `frontend/src/pages/photos/PhotoManager.tsx`

- [ ] Сетка фотографий бани (GET /bathhouses/{id}/photos) с drag-and-drop сортировкой
- [ ] Загрузка новых фото (POST /my/bathhouses/{id}/photos) с preview
- [ ] Статусы фото: pending (на модерации), verified, rejected - с Badge
- [ ] Перестановка порядка (PUT /my/bathhouses/{id}/photos/reorder)
- [ ] Удаление фото (DELETE /photos/{id})

### Task 17: Верификация и финальная проверка

- [ ] Проверить навигацию по всем разделам
- [ ] Проверить обработку ошибок API (401 -> редирект, 403 -> сообщение, 404, 500)
- [ ] Проверить работу с пустыми состояниями (нет бань, нет бронирований и т.д.)
- [ ] Проверить мобильную адаптивность
- [ ] Запустить `npm run build` - убедиться что сборка проходит без ошибок
- [ ] Запустить линтер - убедиться что нет ошибок

### Task 18: Обновление документации

- [ ] Обновить CLAUDE.md с информацией о фронтенде (команды, структура)
- [ ] Переместить этот план в `docs/plans/completed/`
