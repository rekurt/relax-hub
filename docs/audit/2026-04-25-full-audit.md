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

#### A2.2 — `/bathhouses/new` mount creates phantom `listing_drafts` rows (Critical, data integrity)

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

### Representative role

(Session 3)

### Client role

(Session 3)

## Per-role pg_stat_statements snapshots

(populated end of each role pass)
