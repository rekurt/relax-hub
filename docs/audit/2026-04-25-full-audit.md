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

#### A1.1 — Admin 2FA mandatory but seed-demo never enables it (Critical)

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

(more findings appended as the audit continues — the rest of the admin surface still to be walked in upcoming sessions)

### Owner role

(Session 2)

### Representative role

(Session 3)

### Client role

(Session 3)

## Per-role pg_stat_statements snapshots

(populated end of each role pass)
