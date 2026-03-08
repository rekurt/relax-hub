# SEO-оптимизация и ЧПУ

## Overview

SEO-оптимизация для поисковых систем: slug (ЧПУ) для бань и городов, мета-теги (title, description, og:image), генерация sitemap.xml, микроразметка Schema.org (LocalBusiness), canonical URLs. Цель — органический трафик из поисковиков.

**Коммерческая ценность:** Органический (бесплатный) трафик из поисковиков. SEO — основной канал привлечения для маркетплейсов. Правильная микроразметка дает rich snippets в Google (рейтинг, цена, адрес), повышая CTR на 30-50%.

## Context

- Files involved: internal/domain/, internal/repository/, internal/service/, internal/handler/, internal/server/router.go, migrations/
- Related patterns: Clean architecture, Uber fx DI
- Dependencies: нет новых
- Текущее: City уже имеет Slug, бани идентифицируются по UUID. Нужно добавить slug для бань

## Development Approach

- **Testing approach**: Regular (code first, then tests)
- Complete each task fully before moving to the next
- **CRITICAL: every task MUST include new/updated tests**
- **CRITICAL: all tests must pass before starting next task**

## Implementation Steps

### Task 1: Slug для бань

**Files:**
- Modify: `internal/domain/bathhouse.go`
- Create: `internal/seo/slug.go`
- Create: `migrations/000023_bathhouse_slugs.up.sql`
- Create: `migrations/000023_bathhouse_slugs.down.sql`

- [x] Добавить поле Slug string в Bathhouse
- [x] Генератор slug из названия (транслитерация ru->en, lowercase, дефисы):
  - "Баня на Липовой" -> "banya-na-lipovoy"
  - При конфликте: добавить суффикс "-2", "-3" etc.
- [x] Миграция: ALTER bathhouses ADD COLUMN slug VARCHAR(255) UNIQUE
- [x] Заполнить slug для существующих бань из name
- [x] Написать тесты (транслитерация, коллизии, спецсимволы)
- [x] Запустить go test ./... - все тесты должны пройти

### Task 2: ЧПУ-маршруты

**Files:**
- Modify: `internal/repository/postgres/bathhouse.go`
- Modify: `internal/handler/bathhouse.go`
- Modify: `internal/server/router.go`

- [x] Добавить GetBySlug в BathhouseRepository
- [x] GET /api/v1/bathhouses/by-slug/{slug} — получить баню по slug
- [x] GET /api/v1/cities/{slug}/bathhouses — бани в городе по slug (уже есть city_slug filter)
- [x] При создании/обновлении бани — автоматически генерировать/обновлять slug
- [x] Написать тесты
- [x] Запустить go test ./... - все тесты должны пройти

### Task 3: Мета-теги и Open Graph

**Files:**
- Create: `internal/seo/meta.go`
- Modify: `internal/handler/bathhouse.go`

- [x] Генератор мета-тегов для бани:
  ```
  MetaTags {
    Title       string  // "Баня на Липовой в Москве — Bani.ru"
    Description string  // "Баня на Липовой: русская баня, бассейн, мангал. Цена от 2000 руб/ч. Рейтинг 4.8. Бронируйте онлайн."
    OGImage     string  // первое фото бани
    OGType      string  // "business.business"
    Canonical   string  // "https://bani.ru/moscow/banya-na-lipovoy"
  }
  ```
- [x] Включить meta в ответ GetByID и GetBySlug
- [x] GET /api/v1/bathhouses/{id}/meta — мета-теги для SSR/prerender
- [x] Написать тесты
- [x] Запустить go test ./... - все тесты должны пройти

### Task 4: Sitemap и Schema.org

**Files:**
- Create: `internal/handler/sitemap.go`
- Create: `internal/seo/schema.go`
- Modify: `internal/server/router.go`

- [ ] GET /sitemap.xml — XML sitemap:
  ```xml
  <urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
    <url>
      <loc>https://bani.ru/moscow/banya-na-lipovoy</loc>
      <lastmod>2026-03-01</lastmod>
      <changefreq>weekly</changefreq>
      <priority>0.8</priority>
    </url>
    ...
  </urlset>
  ```
- [ ] Включать: все активные бани, все города, главную страницу
- [ ] Кешировать sitemap в Redis (обновлять раз в сутки)
- [ ] Schema.org JSON-LD для бань (тип LocalBusiness):
  ```json
  {
    "@context": "https://schema.org",
    "@type": "LocalBusiness",
    "name": "Баня на Липовой",
    "address": { "@type": "PostalAddress", "addressLocality": "Москва" },
    "geo": { "@type": "GeoCoordinates", "latitude": 55.7558, "longitude": 37.6173 },
    "aggregateRating": { "@type": "AggregateRating", "ratingValue": "4.8", "reviewCount": "42" },
    "priceRange": "$$"
  }
  ```
- [ ] GET /api/v1/bathhouses/{id}/schema — JSON-LD для бани
- [ ] Написать тесты
- [ ] Запустить go test ./... - все тесты должны пройти

### Task 5: Верификация

- [ ] Запустить полный тест-сьют: go test ./... -v
- [ ] Запустить линтер: make lint
- [ ] Запустить go vet ./...
- [ ] Переместить этот план в docs/plans/completed/
