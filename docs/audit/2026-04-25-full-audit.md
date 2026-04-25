# RelaxHUB Full-flow Audit — 2026-04-25

> Findings appended live as audit proceeds. Fix-pack PRs link to anchors.
> Spec: `docs/superpowers/specs/2026-04-25-audit-redesign-design.md`
> Plan: `docs/superpowers/plans/2026-04-25-session-1-preflight-admin-audit.md`

## Severity legend

- **Critical** — 5xx on hot path, frontend crash / blank, permission bypass, data loss, auth/TLS break.
- **High** — silent submit failure, wrong data shown, missing pagination on >200-row list, hot-path query > 2 s.
- **Medium** — empty state unclear, missing skeleton, slow query 500 ms–2 s.
- **Low / Cosmetic** — color/spacing/typography drift.

## Local environment

- Postgres on `localhost:5435` (docker, postgis 17-3.5 amd64 + Rosetta).
- Redis on `localhost:6381`.
- MinIO on `localhost:9102/9103`.
- API on `localhost:8081` (Go server, debug logs in `/tmp/api.log`).
- Frontend on `localhost:5173` (Vite, proxy `/api → :8081`).
- Demo accounts (password `DemoPass123!` for all):
  - `demo.admin@relax-hub.ru`
  - `demo.owner1@relax-hub.ru`
  - `demo.rep.manager@relax-hub.ru`
  - `demo.client1@relax-hub.ru`

## Findings

### Admin role

#### A1.1 — Admin 2FA mandatory but seed-demo never enables it (Critical) ✅ FIXED in PR #21

- **Repro**: log in as `demo.admin@relax-hub.ru / DemoPass123!`, navigate to `/admin`. Every `/api/v1/admin/*` endpoint returns `403 forbidden — two-factor authentication required for admin access`. The dashboard, moderation, analytics, finance, support, settings — entire admin surface is unreachable.
- **Expected**: dashboard loads with metrics; all admin functions accessible.
- **Actual**: blank dashboard with all KPIs at 0; every API call 403.
- **Root cause**: `internal/server/routes_admin.go:20` mounts `middleware.RequireAdmin2FA` before any `RequireAdminPermission`. `internal/middleware/admin_permission.go:RequireAdmin2FA` checks `HasEnabled2FA(userID)` which reads `users.two_fa_method != 'none'`. The seed-demo Go function (`cmd/server/seed_demo.go`) never sets `two_fa_method` on demo admin users — they ship with `two_fa_method='none'` → middleware rejects.
- **Same blocker in production**: any newly-created admin account in prod cannot use `/admin/*` until 2FA is set up via UI. There's no docs path for first-admin-bootstrap; the chicken/egg means the very first admin can't reach the 2FA enable endpoint either if it's behind the same middleware (it isn't — `/auth/2fa/totp/enable` is gated only by `auth`, not by `RequireAdmin2FA`, so it's reachable, but the UX is buried).
- **Fix pointer (Phase 3)**: amend `cmd/server/seed_demo.go` so each demo admin row gets `two_fa_method='totp'` + a deterministic `totp_secret` (or `two_fa_method='sms'` if simpler). For prod onboarding: docs/CLAUDE.md should explicitly call out the first-admin 2FA-enable URL.
- **Workaround applied locally** (`UPDATE users SET two_fa_method='totp', totp_secret='WORKAROUND-LOCAL-AUDIT' WHERE role='admin' AND email LIKE 'demo.%@relax-hub.ru';`) — restored access to all admin endpoints for the rest of this audit. The fix-pack will replace this with proper seed-code.
- **Severity rationale**: Critical because it blocks the *entire* admin role demo experience and is on the hot path of any new admin in production.

#### A1.2 — Login form click-to-submit fails for JS-injected values (High, frontend)

- **Repro**: navigate to `/login`, JS-inject value into `input#email` and `input#password` via `Object.getOwnPropertyDescriptor(HTMLInputElement.prototype,'value').set` + `dispatchEvent('input')`, click "Войти". No `/api/v1/auth/login` request fires. Direct `form.dispatchEvent(new Event('submit'))` works.
- **Expected**: any DOM-set value with input event should propagate through Antd Form state on submit.
- **Actual**: silent — no network request, no error toast, no validation message.
- **Root cause hypothesis**: the existing autofill-bypass `onMouseDown` handler (`syncAntdFormFromDOM`) is supposed to drain DOM values into the antd Form state before submit, but in the JS-injected scenario it triggers BEFORE the input event finishes propagating (or the handler runs but Form's internal state has a race). The form-submit path (Enter / programmatic dispatch) works around this.
- **Severity rationale**: High not Critical because real keyboard-typing users on production are not blocked — only programmatic / Chrome-autofill scenarios. But this interferes with E2E testing, password managers, and (per earlier session reports) was the original symptom of "auth not working".
- **Fix pointer**: revisit `syncAntdFormFromDOM` in `frontend/src/lib/autofill.ts`; consider switching antd Forms to use uncontrolled inputs with form data captured at submit-time, eliminating the controlled-state mismatch entirely.

#### A1.3 — AdminDashboard reports 0 bookings despite 205 in DB (Medium → High)

- **Repro**: log in as admin (with 2FA workaround), open `/admin`. Dashboard cards show "Бронирования 0", "Выручка 0", "Просмотры 0", "DAU 0".
- **Expected**: ~205 bookings, non-zero revenue (seed demo creates payments and bookings).
- **Actual**: all KPIs 0.
- **Evidence**:
  - DB: `SELECT count(*) FROM bookings;` = 205.
  - API: `GET /api/v1/admin/analytics?period=7d` returns `{"total_bookings": 0, "total_revenue": 0, "total_views": 0, "dau": 0, "wau": 1, "mau": 1}`. (`total_users: 26` and `total_bathhouses: 10` are correct.)
- **Root cause hypothesis**: the analytics service likely filters bookings by `created_at >= now() - 7d` AND by `status IN ('completed', …)`. The seed data inserts bookings with timestamps spread across the past 90 days; the 7-day window catches only a tiny slice OR the query joins on a denormalised `bathhouse_views` table that is not seeded.
- **Fix pointer**: inspect `internal/service/analytics_service.go` for the dashboard aggregator — likely needs a default fallback to a wider window or a "lifetime" period that the dashboard surface should expose.
- **Severity**: Medium today (cosmetic on demo), but **upgrade to High** if the same query is used for finance / reconciliation reports — operators relying on 7d-zero values would mis-count revenue.

#### A1.4 — Every admin page double-fetches its primary list on mount (High, frontend)

- **Repro**: load any admin list page (`/admin/bathhouses`, `/admin/reviews`, `/admin/photos`, `/admin/complaints`, etc.). The primary GET fires **twice** with identical query strings within milliseconds.
- **Evidence**:
  - `/admin/bathhouses?page=1&page_size=20` ×2
  - `/admin/reviews?page=1&page_size=20` ×2 + `/admin/reviews/pending-count` ×2
  - `/admin/photos/pending?page=1&page_size=12` ×2
  - `/admin/complaints?page=1&page_size=20` ×2
  - `/auth/me` ×2 on every navigation, `/my/notifications/unread-count` ×2 on every page
- **Root cause hypothesis**: React.StrictMode in dev double-mounts effects (canonical behavior); the components haven't been written to be re-entrant — each `useEffect(() => fetch(), [])` fires twice. Production won't double-fetch, BUT the cost is real on dev iteration speed AND it indicates the components don't dedupe via TanStack Query keys properly. The duplicates appear regardless of TanStack Query staleTime, suggesting bypassed cache or unstable query keys.
- **Severity**: High in dev (slows audit loop, doubles backend load when many devs are active), Medium in prod (StrictMode is dev-only).
- **Fix pointer (D — query consolidation)**: audit query keys in `frontend/src/api/generated/admin/admin.ts` and surrounding hooks; ensure `queryKey: ['admin-reviews', page, pageSize]` (stable) is used consistently with `staleTime: 30000`+. Also `/auth/me` is called twice — once from `App.tsx` `loadProfile()` and likely once from TanStack `useGetAuthMe`.

#### A1.5 — All admin pages share the wrong `<title>` "RelaxHUB — Бронирование бань" (Medium, UX)

- **Repro**: navigate to any `/admin/*` page (Dashboard, Bathhouses, Reviews, Photos, Complaints all observed). Browser-tab title stays "RelaxHUB — Бронирование бань".
- **Expected**: per-page title like "Модерация отзывов", "Модерация бань", etc.
- **Root cause**: `useDocumentTitle` (or equivalent) is missing in admin pages. Some upstream layout sets the title once and never updates.
- **Fix pointer**: add `<title>` per route via Helmet or a `useEffect(() => { document.title = `${PLATFORM_NAME} — ${pageTitle}`; }, [])` in each admin page; OR centralize in `AdminLayout.tsx` keyed off the URL path.

#### A1.6 — `/admin/photos` and `/admin/complaints` show empty content with no empty-state UI (Low)

- **Repro**: navigate to either; table area is blank, no "Нет фото на проверку" or "Нет жалоб" message, no skeleton.
- **Severity**: Low (cosmetic), but worth fixing in C1 (UI tokens / empty states pass).
- **Fix pointer**: use Antd `Empty` component or a project `EmptyState` (already exists per CLAUDE.md component list).

#### A1.7 — `/admin/subscriptions` queries owner's `/my/subscriptions` instead of admin endpoint (High, frontend wrong endpoint)

- **Repro**: log in as admin, open `/admin/subscriptions`. Network tab shows `GET /api/v1/my/subscriptions?page=1&page_size=20` instead of an admin-listing endpoint.
- **Expected**: an admin endpoint listing **all** subscriptions across all owners.
- **Actual**: hits the owner-self endpoint, returns admin's own (empty) subscription list.
- **Root cause**: `frontend/src/pages/admin/SubscriptionManagement.tsx` reuses an owner hook (`useGetMySubscriptions` from `frontend/src/api/generated/subscriptions/subscriptions.ts`) instead of an admin variant.
- **Fix pointer**: needs an `/api/v1/admin/subscriptions` (list-all) endpoint AND the corresponding orval-generated client. Until that is shipped, the admin page should display "Раздел в разработке" rather than appear functional.
- **Severity**: High — operationally important admin surface returns wrong data, can confuse the operator.

#### A1.8 — `/admin/certificates` likewise hits `/my/certificates` (High, frontend wrong endpoint)

- **Repro**: log in as admin, open `/admin/certificates`. Network shows `GET /api/v1/my/certificates?page=1&page_size=20`.
- Same shape as A1.7 — admin page reuses owner/client hook.
- **Fix pointer**: same as A1.7 — add admin endpoint, generate client, switch UI.

### Admin coverage summary (this session)

37 admin pages covered:

| Batch | Pages | Status |
|---|---|---|
| 1 Moderation | AdminDashboard, BathhouseModeration, ReviewModeration, PhotoVerification, ComplaintManagement | ✅ all 200, dup fetch + missing empty states |
| 2 Users+Content | UserManagement, RoleManagement, CityManagement, AmenityManagement, ObjectTypeManagement, HolidayManagement | ✅ all 200, AmenityManagement is the only page with proper TanStack key (no dup) |
| 3 Finance | AdminFinanceDashboard, BankReconciliation, WalletManagement (search-only), BookingManagement (no auto-list bug) | ⚠️ BookingManagement should auto-load with default filter |
| 4 Support | TicketManagement (3 tickets), DisputeManagement (2), TicketDetail/DisputeDetail (not deeply visited) | ✅ |
| 5 Analytics | ConversionFunnels (empty body), CohortAnalysis, SupplyDemandMetrics, GeoHeatmap, AntiFraudDashboard | ✅ |
| 6 Configuration | PlatformSettings (13), FeatureFlags (13), ServiceFeeConfig (1), SubscriptionManagement (A1.7), LoyaltyManagement (4 static tiers) | ⚠️ A1.7 |
| 7 Misc | AdminProfile, AdminAuditLog (empty), AdminNotifications (uses /my, not admin), AdminNotificationCenter, ForceMajeure, FAQManagement (18), GlobalPromoCodes (form-only), CertificateManagement (A1.8) | ⚠️ A1.8 |

**Critical / High count**: 2 Critical (A1.1 + workaround applied) / 4 High (A1.2 autofill, A1.4 dup-fetch, A1.7 subscriptions, A1.8 certificates). **Medium**: 2 (A1.3 dashboard zeros, A1.5 wrong title). **Low**: 1 (A1.6 empty states).

**No 5xx, no auth-bypass, no data-loss bugs found** in admin role this session. All `/admin/*` endpoints respond 200 once the 2FA workaround is applied.

### Session 1 wrap

**Shipped**: PR #21 `fix(seed-demo): admin demo accounts ship with TOTP 2FA pre-configured (A1.1)` merged to master.

**Outstanding admin findings** carry forward to next fix-pack:
- A1.2 (High) autofill submit
- A1.3 (Medium→High) dashboard 7d-window zeros
- A1.4 (High) duplicate fetches
- A1.5 (Medium) wrong document title
- A1.6 (Low) missing empty states
- A1.7 (High) subscriptions wrong endpoint (needs new /admin endpoint + orval regen)
- A1.8 (High) certificates wrong endpoint (needs new /admin endpoint + orval regen)

**Next session entry-point**: Phase 2 owner role audit. Use the same harness:
1. `docker compose up -d postgres redis minio` (env still on dev ports 5435/6381/9102)
2. Re-run server with env from this MD's "Local environment" block.
3. Login as `demo.owner1@relax-hub.ru / DemoPass123!`, walk owner menu top-to-bottom (BathhouseList → BathhouseForm wizard → BookingList → ChatPage → ReviewList → CRM → FinanceDashboard → Promotion → Pricing → Settings).
4. Append to `### Owner role` section here, snapshot pg_stat_statements to `docs/audit/2026-04-25-owner-role-pgss-snapshot.csv`, fix-pack PR `fix/audit-owner-pack`.

**Sub-projects still to do** (per spec §11): A2 owner / A3 rep / A4 client / D query consolidation / B menu redesign / C UI redesign / E prod data gap fill.

### Session 2 wrap

**Shipped**: PR [#22](https://github.com/rekurt/relax-hub/pull/22) `fix(audit-owner): A2.2 phantom listing_drafts + A2.6 owner1 reps` — closes A2.2 Critical and partial A2.6 (1 of 4 sub-gaps).

**Outstanding owner findings** carry forward:
- A2.1 (High) quad-fetch on `/my/bathhouses` (cross-cuts admin A1.4 — single D-phase fix benefits both)
- A2.3 (Medium) owner missing `/bookings/:id` route — silent redirect to `/`
- A2.5 (Medium) wrong document title (mirrors admin A1.5)
- A2.6 (High remainder) seed gaps for `bathhouse_photos`, `webhooks`, `guest_cards` from existing 129 completed bookings — needs seed extension or backfill migration
- A2.7 (High dev / Medium prod) `/chat` triple-fetches `/my/conversations` — hoist query to ChatPage parent
- A2.8 (Medium) `/widget` doesn't render embed code block
- A2.9 (Low) `/finance/reports` lacks summary stats
- A2.10 (Low → Medium) `/reviews` uses public bathhouse endpoint instead of owner-scoped variant
- A2.4 reclassified as **false alarm** + Low cosmetic header-level inconsistency on `/promotion` and `/pricing`

**Next session entry-point**: Phase 3 representative role audit. Login as `demo.rep1@relax-hub.ru / DemoPass123!` (the only seeded rep — manager on `bh7` + `bh10` for owner2; after PR #22 merges, also manager on `bh1` + observer on `bh2` for owner1). Rep cabinet is a subset of owner cabinet — most pages are read-only or moderation-restricted. Walk plan:
1. Smoke harness: `curl localhost:8081/health`, `docker ps`, login.
2. Pre-flight: reset pgss, capture initial state.
3. Walk pages for manager role (most permissive): `/bookings`, `/bookings/{id}` (rep should NOT redirect to `/` — verify), `/calendar`, `/chat`, `/reviews`, `/notifications`. Append findings as A3.* under `### Representative role`.
4. Switch to observer role (read-only): repeat walk, capture which pages should be hidden / write-actions should be 403.
5. Snapshot pgss → `docs/audit/2026-04-25-rep-role-pgss-snapshot.csv`.
6. Triage Critical/High → fix-pack PR `fix/audit-rep-pack`.
7. Update Session 3 wrap note pointing to client role for Session 4.

### Owner role

**Auditor**: claude-opus-4-7 (Session 2, 2026-04-25)
**Login**: `demo.owner1@relax-hub.ru / DemoPass123!`
**Harness**: same docker-compose stack as Session 1, Chrome MCP tab `1919497189`, pgss reset before pass.

#### A2.1 — Every owner list page double-fetches its primary list on mount, plus `/my/bathhouses` is fetched 4× (High, frontend)

- **Repro**: load any owner list page (`/bathhouses`, `/bookings`, `/reviews`, `/calendar`, etc.). The primary GET fires twice with identical query strings. On top of that, `/my/bathhouses` is requested as both the no-arg form (`?status=active` only) and the paginated form (`?page=1&page_size=…`) — each in turn doubled — for a total of **4** identical-or-near-identical requests within ~200 ms.
- **Evidence** (samples):
  - `/my/bathhouses?status=active` ×2 + `/my/bathhouses?page=1&page_size=20&status=active` ×2 (BathhouseList)
  - `/my/bookings?page=1&page_size=20` ×2 (BookingList)
  - `/my/bathhouses/{id}/calendar-token` ×2 + `/my/bathhouses/{id}/external-calendars` ×2 (CalendarPage)
  - `/auth/me` ×2 on every navigation, `/my/notifications/unread-count` ×2 on every page (same as A1.4)
- **Root cause hypothesis**: identical to A1.4 — React StrictMode double-mounts effects in dev. Owner pages additionally use **two different shapes** of the same query (BathhouseSelector hook fires no-arg, page hook fires paginated) → 4× total. Both consumers should share a stable TanStack key.
- **Severity**: High in dev, Medium in prod (StrictMode is dev-only, but the no-arg+paginated split fires in prod too — owner role pages double-load `/my/bathhouses` on every nav).
- **Fix pointer (D — query consolidation)**: in `frontend/src/components/BathhouseSelector.tsx` and the page-level `useGetMyBathhouses` hooks, normalize on a single paginated query key (e.g. `['my-bathhouses', { page: 1, pageSize: 20, status: 'active' }]`) with `staleTime: 30000`+. Cross-cutting fix benefits both A1.4 (admin) and A2.1 (owner).

#### A2.2 — `/bathhouses/new` mount creates phantom `listing_drafts` rows (Critical, data integrity) ✅ FIXED in fix/audit-owner-pack

- **Repro**: log in as owner, navigate to `/bathhouses/new`. Without typing anything, observe network: `POST /api/v1/my/listing-drafts` fires immediately on mount and returns **201 twice** within milliseconds.
- **Evidence**: SQL after 2 visits: `SELECT count(*) FROM listing_drafts WHERE owner_id = '<owner1>' AND created_at > now() - interval '5 minutes'` → **4 rows** (2 per visit). Total drafts in dev DB jumped from 8 to 12 after this experiment alone.
- **Root cause hypothesis**: BathhouseForm wizard creates the draft in a top-level `useEffect(() => createDraft(), [])` with no idempotency on the client (no `useRef` guard, no `enabled: !draftId` gate) — React StrictMode fires the effect twice → 2 rows. **Plus** the backend `POST /my/listing-drafts` is not idempotent: every call inserts a new row. In production (no StrictMode) this still litters the table on every page reload, and on dev every visit costs 2 rows.
- **Severity**: **Critical** for data integrity — a user clicking the "Создать" link more than once accumulates phantom drafts that show up in their drafts list, in admin dashboards counting drafts, and in pgss as duplicate inserts. The seed-demo gives owner1 6 active + 1 pending + 1 rejected; phantom drafts can quickly outnumber real drafts.
- **Fix pointer**: two layers — (a) in `frontend/src/pages/bathhouses/BathhouseForm.tsx`, gate the create-draft mutation behind a `useRef`-based "already-fired" sentinel, OR pull `draftId` from URL/route params and only call create when it's missing; (b) on the backend (`internal/handler/listing_draft_handler.go` + service), make `POST /my/listing-drafts` idempotent — if the same owner has a draft in `step=1` with no field data set, return the existing row instead of inserting. Layer (a) ships in fix-pack; layer (b) is a follow-up if (a) doesn't close the prod-mode regression.

#### A2.3 — Owner has no `/bookings/:id` detail route; deep links silently redirect to `/` (Medium, navigation)

- **Repro**: as owner, copy any booking id from `/bookings`, paste `/bookings/<uuid>` into the address bar. Page silently redirects to `/` (dashboard).
- **Expected**: either a dedicated owner BookingDetail page, OR an explicit "Редирект → список" with a back-button to the row. Currently a deep link from email/Telegram lands the owner on the dashboard with no breadcrumb.
- **Root cause**: `frontend/src/router.tsx` lines ~160-205 register `/bookings` for owner but not `/bookings/:id`; only the client role has `/client/bookings/:id`. The client variant route catches nothing for owner, so the SPA fallback redirects.
- **Severity**: Medium — not data-breaking but breaks shareable deep links, support workflow, push notifications that link to a specific booking.
- **Fix pointer**: add an owner-scoped `/bookings/:id` route reusing the same component (or rendering an inline drawer over `/bookings?focus=<uuid>`). BookingList already shows details inline; the route just needs to scroll/expand the matching row.

#### A2.4 — `/dashboard` (owner home), `/promotion`, and `/pricing` render nothing despite endpoints succeeding (~~Critical~~ → **FALSE ALARM, downgraded to Low cosmetic**)

> **Re-test on 2026-04-25 after fresh navigation** found all three pages DO render: `/dashboard` shows `<h1>Дашборд` + 6 `bani-stat-tile` KPI tiles (custom CSS class, not `.ant-statistic` — my initial selector missed them); `/promotion` shows `<h3>Рекламные кампании` + table + empty-state CTA "Нет кампаний. Создайте первую!"; `/pricing` shows "Правила ценообразования" with smart-pricing recommendation card (current 8400 ₽, recommended 7896 ₽ at coefficient ×0.94). The original blank-page detection was a timing artifact: my JS query ran during initial cache-miss render before TanStack Query resolved + my selectors didn't match `.bani-stat-tile` / `<h3>` / Typography titles.
> **Remaining real issue (Low cosmetic)**: `/promotion` and `/pricing` use `<Typography.Title level={3}>` (renders `<h3>`) instead of the standard `<PageHeader>` + `<h1>` pattern that `/dashboard` and `/finance` use. Inconsistent header levels across owner pages. Worth normalizing in C1 UI tokens pass.

- **Repro**: log in as owner with a bathhouse selected (auto-selected from BathhouseSelector). Navigate to `/dashboard`, `/promotion`, or `/pricing`. Network: GETs return 200. DOM: `h1` for `/dashboard` says "Дашборд" but `.ant-statistic`, `.ant-card`, `canvas`, `.ant-list-item` all count **0**. `/promotion` and `/pricing` lack even an `h1`.
- **Evidence**:
  - `/dashboard`: `GET /my/bathhouses/{id}/analytics?period=30d` ×2 → 200, but UI: 0 statistics, 0 charts, 0 cards. Page is effectively blank below the layout chrome.
  - `/promotion`: `GET /my/bathhouses/{id}/promotions?page=1&page_size=10` → 200, plus `/cities` ×2. UI: `h1=null`, 0 cards, 0 list items, no empty state, no error banner. Just whitespace.
  - `/pricing`: `GET /my/bathhouses/{id}/pricing-rules` ×2, `/price-recommendation`, `/seasonal-tariffs` → all 200. UI: `h1=null`, 0 tabs, 0 rows, no empty state, no error banner.
- **Root cause hypothesis**: three different pages share the same broken pattern — page renders only when data is present. When the bathhouse has no promotions / no pricing rules / no analytics-derived stats, the page renders nothing. Likely an early-`return null` before the title/header/empty state. Owner1 has 6 active bathhouses; the page just doesn't have data populated for those entities (see A2.6 seed gaps). The dashboard variant additionally suggests the analytics aggregator returns zero-shape data and components do `if (!data?.bookings?.length) return null` instead of rendering "Нет данных за выбранный период".
- **Severity**: **Critical** for owner UX — the home page (the first thing a logged-in owner sees) is blank. A new owner is immediately convinced the platform is broken. Sales-demo killer.
- **Fix pointer**:
  - `frontend/src/pages/analytics/OwnerAnalytics.tsx` (used at `/dashboard`?) — wait, dashboard route maps somewhere else. Inspect `frontend/src/router.tsx` for `/dashboard` element and the matching component. Likely `frontend/src/pages/Dashboard.tsx` or similar. Add an `Empty` fallback when stats are zero, render the `h1` and KPI grid skeleton always.
  - `frontend/src/pages/promotion/PromotionCampaign.tsx` — same pattern, ensure `<Title>` always renders, gate only the data section behind `isLoading`/`isEmpty`.
  - `frontend/src/pages/pricing/PricingRules.tsx` — same pattern.
- **Hot-mitigation**: add a single shared `<PageHeader title=…>` to every owner page top-level, regardless of data. Cheap win.

#### A2.5 — Every owner page shares the wrong `<title>` "RelaxHUB — Бронирование бань" (Medium, UX)

- **Repro**: navigate to any `/bathhouses`, `/bookings`, `/finance`, `/calendar`, `/crm/*`, `/settings/*`, `/photos`, `/widget` — browser tab title is identical "RelaxHUB — Бронирование бань" for all of them.
- **Severity**: Medium — same shape as A1.5 but for owner role. Cross-cutting fix should land once for all roles.
- **Fix pointer**: same as A1.5. Centralize `useDocumentTitle(pageTitle)` hook in each layout's outlet wrapper, keyed on route path.

#### A2.6 — Seed-demo gaps render multiple owner UI surfaces empty when seed claims they should have data (High, seed completeness) ⚠️ PARTIAL FIX in fix/audit-owner-pack (representatives only)

- **Repro** (SQL): `psql ... -c "SELECT 'representatives' AS tbl, count(*) FROM representatives WHERE bathhouse_id IN (SELECT id FROM bathhouses WHERE owner_id=(SELECT id FROM users WHERE email='demo.owner1@relax-hub.ru')) UNION ALL SELECT 'bathhouse_photos', count(*) FROM bathhouse_photos WHERE … UNION ALL SELECT 'guest_cards', count(*) FROM guest_cards WHERE owner_id=… UNION ALL SELECT 'webhooks', count(*) FROM webhooks WHERE owner_id=…;"`
- **Result**:
  - `representatives` for owner1 = **0** (seed plan calls for ≥1 manager + 1 observer per the multi-role roster; PR #20 was supposed to ship this)
  - `bathhouse_photos` (verified) for owner1 = **0** across all 6 active bathhouses (seed plan: ≥3 verified per active bathhouse)
  - `guest_cards` for owner1 = **0** despite **129 completed bookings** for owner1's bathhouses (seed plan + CLAUDE.md: guest cards auto-created on booking completion)
  - `webhooks` for owner1 = **0** (seed plan: 1 webhook configured)
- **Impact**: `/representatives` shows "Нет представителей", `/photos` shows "Нет фотографий", `/crm/guests` shows "Гостей пока нет", `/crm/rfm` shows "Нет данных для RFM-анализа", `/settings/webhooks` shows "Нет вебхуков". All five pages are technically working — the empty states render correctly — but the demo / sales experience is hollow.
- **Root cause hypothesis**: `cmd/server/seed_demo.go` either lacks the seed funcs `seedOwnerProfile` / `seedCRMAndOps` / `seedBathhouseExtras` from the federated-dragonfly plan, or those funcs run but skip these tables. The 129 completed bookings show they were inserted via SQL UPSERT that bypasses the service layer hook responsible for creating `guest_cards` (per CLAUDE.md: "guest cards (auto-created on booking completion ...)"). So even if the seed adds CRM-related users, no guest cards get materialized.
- **Severity**: **High** — multiple owner cabinet surfaces look unimplemented. Blocks both audit completeness ("can't audit features the seed never populated") and the platform's demo story.
- **Fix pointer**:
  - `cmd/server/seed_demo.go` — add or repair `seedRepresentatives` (insert 1 manager + 1 observer for owner1's first 2 bathhouses), `seedBathhousePhotos` (3 verified photos per active bathhouse, deterministic UUID prefix), `seedWebhooks` (1 webhook on owner1), `seedGuestCardsFromCompletedBookings` (idempotent insert from existing completed bookings for owner1+owner2).
  - Backend follow-up: trigger or service-layer hook that materializes `guest_cards` from existing `bookings WHERE status='completed'` is missing or only fires on transition. Add a one-shot migration `migrations/NNNN_backfill_guest_cards.up.sql` to materialize cards from current completed bookings.

#### A2.7 — `/chat` triple-fetches `/my/conversations` (High, frontend dup-fetch worse than A2.1)

- **Repro**: navigate to `/chat`. Capture network: `GET /api/v1/my/conversations?page=1&page_size=50` fires **three** times within ~150 ms (not the usual ×2 we see elsewhere).
- **Root cause hypothesis**: ChatPage likely has both `<ConversationList>` and `<MessageArea>` mounting independently and each calling `useGetMyConversations` directly, plus the StrictMode double — yielding 3 calls. Adding a shared parent fetch + prop-drilling (or a TanStack key both consumers reuse) fixes it.
- **Severity**: High in dev (3× backend load on a relatively heavy endpoint), Medium in prod (StrictMode goes away → 1 or 2 calls remain depending on whether both children fetch).
- **Fix pointer**: hoist the conversations query to `frontend/src/pages/chat/ChatPage.tsx` and pass results as props to `<ConversationList>` and `<MessageArea>`. Or wrap both in a query key that resolves to a single network round-trip via `staleTime`.

#### A2.8 — `/widget` doesn't render the embed code block on mount (Medium, frontend)

- **Repro**: navigate to `/widget`. Network: `GET /my/bathhouses/{id}/widget-key` → 200 + `GET /my/bathhouses/{id}/widget-code?color=…&font_family=…` → 200, both ×2. DOM: 4 form fields (color picker, font, toggles), but **0 `<pre>` or `<code>` blocks** that would show the iframe embed snippet.
- **Severity**: Medium — owners can't actually copy-paste the widget into their site without the rendered code block.
- **Fix pointer**: `frontend/src/pages/widget/WidgetSettings.tsx` — verify the response is unmarshalled and rendered into a `<Typography.Paragraph code>` or `<pre>`. Could also be a CSS class issue where the block is hidden until form is submitted; should be visible by default.

#### A2.9 — `/finance/reports` lacks any summary numbers — only download buttons (Low, UX)

- **Repro**: navigate to `/finance/reports`. UI: 3 download buttons ("Скачать", "Скачать акт", "Скачать XML") and nothing else — no statistics, no preview of what each report contains, no date range selector visible. Just buttons.
- **Severity**: Low — the buttons work (network calls succeed), but UX is unfriendly: owner doesn't know what each download will give them. Same shape as A1.6 (admin "no empty state") only inverted — there's no explanation of what's available.
- **Fix pointer**: `frontend/src/pages/finance/FinancialReports.tsx` — add a summary section above the buttons (e.g. "Текущий месяц: 145 броней, 320 000 ₽, к выплате 285 000 ₽") and a date-range picker.

#### A2.10 — `/reviews` (owner) uses public `/bathhouses/{id}/reviews` endpoint instead of an owner-scoped endpoint (Low → Medium, frontend wrong endpoint)

- **Repro**: as owner, open `/reviews`. Network: `GET /api/v1/bathhouses/{id}/reviews?page=1&page_size=10` (the **public** endpoint).
- **Expected**: an owner-scoped variant that includes pending-moderation reviews, "answered/unanswered" filtering, and is auth-gated to owner-only data.
- **Actual**: hits public reviews endpoint, returning the same set any visitor would see. Owner can't see reviews still in moderation queue or filter unanswered ones.
- **Root cause**: `frontend/src/pages/reviews/ReviewList.tsx` reuses `useGetBathhousesIdReviews` instead of an owner-specific hook.
- **Severity**: starts Low (works for the common case where most reviews are public-approved), can be Medium when many moderation cases pile up. Mirrors A1.7 / A1.8 for admin.
- **Fix pointer**: backend probably has `/api/v1/my/bathhouses/{id}/reviews` already (worth grepping `internal/handler/review_handler.go`); if not, add it. Frontend swaps the hook.

### Owner coverage summary (this session)

≈30 owner pages walked across 5 batches:

| Batch | Pages | Status |
|---|---|---|
| B1 Bathhouses | BathhouseList, BathhouseForm wizard (A2.2 phantom drafts), ListingImport, AuditLog (skipped, not directly reachable) | ⚠️ A2.2 Critical |
| B2 Bookings + Calendar | BookingList (10/241 rows, owner-side correct), `/bookings/:id` (A2.3 silent redirect), CalendarPage (no widget rendered) | ⚠️ A2.3 |
| B3 Comms + Reviews + Photos | ChatPage (A2.7 triple-fetch), NotificationList, ReviewList (A2.10 public endpoint), PhotoManager (empty per A2.6), PhotoOrderPage | ⚠️ A2.7 + A2.10 |
| B4 Money + Promo + Analytics | FinanceDashboard, PayoutPage, FinancialReports (A2.9), OwnerAnalytics, PromotionCampaign (A2.4 blank), PromoList, PricingRules (A2.4 blank), SubscriptionPage | ⚠️ A2.4 (Critical) + A2.9 |
| B5 CRM + Settings + Reps + Widget | /crm/{guests,segments,rfm,broadcasts,scenarios,templates}, RepresentativeList (A2.6 empty), Widget (A2.8 no code block), Settings/{webhooks,pms,kyc,offer} | ⚠️ A2.6 + A2.8 |
| Dashboard | `/dashboard` (A2.4 blank) | ⚠️ A2.4 (Critical) |

**Critical / High count after triage**: 1 Critical (A2.2 phantom drafts ✅ FIXED) / 4 High (A2.1 quad-fetch, A2.6 seed gaps ⚠️ PARTIAL FIX shipped reps for owner1, A2.7 triple-fetch, A2.10 public endpoint). **Medium**: 3 (A2.3 missing detail route, A2.5 wrong title, A2.8 widget no code block). **Low**: 2 (A2.4 false-alarm cosmetic header inconsistency, A2.9 reports lacks summary).

**Fix-pack `fix/audit-owner-pack` ships**:
- A2.2 frontend `useRef`-sentinel guard in `BathhouseForm.tsx` to stop StrictMode double-mount inserts (1 Critical closed).
- A2.6 seed extension: 2 new `repAssignments` entries attaching `rep1ID` to owner1's `bh1` (manager) and `bh2` (observer) so `/representatives` is no longer empty in demo (1 of 4 sub-gaps closed).

**Carry forward to next session**:
- A2.6 remaining: bathhouse_photos verified seed (3/active bathhouse), webhooks seed (1 for owner1), guest_cards backfill from existing 129 completed bookings (likely needs a service-layer or migration approach).
- A2.1 quad-fetch + A2.7 triple-fetch — cross-cutting D-phase query consolidation.
- A2.3, A2.5, A2.8, A2.9, A2.10 — bundle into C1/C2 UI pass or a follow-up frontend-only PR.

**No 5xx, no auth-bypass, no data-loss bugs found** in owner role this session. All `/api/v1/my/*` endpoints respond 200.

### Representative role

(Session 3)

### Client role

(Session 3)

## Per-role pg_stat_statements snapshots

(populated end of each role pass)
