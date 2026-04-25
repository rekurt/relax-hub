# RelaxHUB — Full Audit, Fix-pack, Redesign & Prod Data Gap

**Date:** 2026-04-25
**Status:** Approved (Section 1) — autonomous execution authorized
**Owner:** Operator (delegating)
**Branch base:** `master`

## 0. Context

`relax-hub.ru` is a multi-role marketplace (client / owner / representative / admin) deployed to VK Cloud K8s. Frontend: Vite + React + Ant Design 6 + TanStack Query, ~117 screens across 5 layouts. Backend: Go + chi + pgx + Uber fx. Live URL https://relax-hub.ru is the source of truth; localhost will be the audit/fix workshop.

This spec consolidates 5 sub-projects into one execution pipeline:

```
A. Full audit (per role, full-flow click-through)
   ├─ records bugs + network captures + SQL telemetry
   └─ feeds D (queries) and B (menu) and C (UI)
D. Query consolidation (frontend HTTP + backend SQL)
B. Main menu redesign (top-nav UX, role-aware)
C. Professional UI overhaul (design tokens + per-layout)
E. Data-gap fill: local vs prod schema/seed-demo, idempotent Job
```

## 1. Goals & Non-goals

### Goals
- Markdown audit report `docs/audit/2026-04-25-full-audit.md` with bug list categorized by severity
  (Critical / High / Medium / Low / Cosmetic), one section per role.
- All Critical + High bugs fixed via fix-packs (one PR per role-bucket).
- Query consolidation: documented top-50 SQL by total_time + top-20 N+1 patterns + duplicate frontend
  HTTP fan-out report.
- New menu shipped, replacing the current sidebar/top-nav split.
- UI design system extracted to tokens + at least one layout (ClientLayout) fully revamped.
- Prod data gap closed: idempotent Job fills missing reference rows
  (amenities, object_types, holidays, platform_settings, feature_flags, service_fee_configs)
  + demo users/bathhouses needed for sales/feature demos.

### Non-goals
- Real payments (YooKassa stays in stub mode locally; no confirmable payment in production).
- Full DB dump from local→prod. Only idempotent UPSERT-by-UUID via existing
  `seed_demo.go`. Live user data (`bookings`, `payments`, `wallets`, `wallet_transactions`,
  `reviews`, `client_reviews`, `disputes`, `support_tickets`, `complaints`, `audit_logs`)
  is never written from outside.
- Production destructive testing (delete bathhouse, force-cancel) — only local.
- Generating large synthetic data sets — `seed_demo.go` is amended, not replaced.
- Backend rewrite. Single-line index fixes are applied; deep refactors are filed as Medium and skipped.

## 2. Pre-flight & Tooling

### Local stack
1. `make docker-up` — postgres 16, redis 7, minio.
2. `make migrate-up` — apply `migrations/*.up.sql`.
3. `go run ./cmd/server seed-demo --password=DemoPass123!` — populate fixtures.
4. `go run ./cmd/server serve --with-admin` — API on :8080 with admin panel.
5. `make frontend-dev` — Vite on :5173 (proxies `/api → :8080`, `/ws → :8080`).

### Telemetry to enable BEFORE the audit starts
- Postgres: `CREATE EXTENSION IF NOT EXISTS pg_stat_statements;` + `pg_stat_statements_reset()`
  before each role pass, snapshot after.
- App: `BANI_LOGGER_LEVEL=debug`, JSON logs.
- Browser: Chrome MCP (`mcp__Claude_in_Chrome__*`). For each page visit, capture
  `read_network_requests` post-load AND post-interaction.

### Audit log keeper
- Single file `docs/audit/2026-04-25-full-audit.md`, sections by role.
- Each finding: severity, role, screen URL, repro steps, expected vs actual,
  network/SQL evidence, root cause hypothesis, fix pointer.

## 3. Audit method (P1 — linear by role)

Order: **admin → owner → representative → client**.

For each role:

1. Login as the role's demo user:
   - `demo.admin@relax-hub.ru / DemoPass123!`
   - `demo.owner1@relax-hub.ru / DemoPass123!`
   - `demo.rep.manager@relax-hub.ru / DemoPass123!`
   - `demo.client1@relax-hub.ru / DemoPass123!`
2. Reset SQL telemetry: `SELECT pg_stat_statements_reset();`
3. Walk the menu top-to-bottom:
   1. Load → wait for idle → screenshot.
   2. Read network: flag status >= 400, duplicates within 1 s, total > 800 ms,
      requests fired but unused on screen.
   3. Interact: each form's happy-path submit, each list's primary CRUD action.
   4. Append findings to audit MD.
4. End-of-role: snapshot top 50 SQL, triage Critical/High → fix-pack PR; Medium/Low → backlog.
   Open `fix/audit-<role>-pack`, land changes, push, open PR, squash-merge.
   Re-smoke critical paths after CI deploy.

### Severity rules
- Critical: 5xx on non-edge case; frontend crash / blank / infinite spinner;
  permission bypass; data corruption; auth/TLS break.
- High: silent submit failure; wrong data shown; missing pagination on >200-row list;
  hot-path query > 2 s.
- Medium: unclear empty states; missing skeletons; slow query 500 ms–2 s.
- Low / Cosmetic: color / spacing / typography drift.

## 4. Fix-pack discipline

- One PR per role-bucket (`fix/audit-admin-pack`, etc.).
- Each PR: bug list (severity + 1-line repro), commit-per-bug, link to audit MD.
- After merge + CI deploy → audit MD marks the bug fixed and links the PR.
- Critical bugs that need >2 h: hot-mitigation (feature-flag off / temporary error)
  + follow-up issue.

## 5. seed-demo amendments

Extend `cmd/server/seed_demo.go` only when the audit finds incomplete fixtures:
- Missing role coverage (e.g., one rep but two sub-roles).
- Missing edge case (e.g., booking in `force_majeure_cancelled` to exercise admin UI).
- Missing reference rows (e.g., `service_fee_config` for region BY).

UPSERT-by-UUID, deterministic UUIDs in the existing namespace prefixes
(`30000000-…` clients, `44000000-…` wallets, etc.).

## 6. D — Query consolidation

After all 4 role audits:
- Backend: dedupe `pg_stat_statements` snapshots, rank by `total_exec_time` desc.
  Top 20 → fix decision per query (index, rewrite, denormalize, or accepted).
  Index-only fixes → `migrations/NNN_query_audit_indexes.up.sql`.
- Frontend: from accumulated network captures, identify repeat-fetch (mount→fetch X→
  unmount→mount→fetch X again) — fix in TanStack Query staleTime / shared keys.
  Identify dead requests — remove.

Output: `docs/audit/2026-04-25-query-audit.md`.

## 7. B — Main menu redesign

Reference brands: Booking.com (clean horizontal), yandex-travel (search-prominent),
Avito (right-side avatar). Constraints:

- Single top bar across all roles. Width-responsive: hamburger <768 px.
- Left: logo + ≤5 visible primary anchors at lg.
- Center: contextual quick-search (catalog / bookings / moderation queue per role).
- Right: notifications bell + avatar menu.
- Sub-navigation: tabs inside each page (no sidebar on mobile, drawer on tablet).
- Role switching for admin via avatar menu (read-only).

Implementation: new components `Topbar.tsx`, `TopbarSearch.tsx`, `TopbarAvatarMenu.tsx`
in `frontend/src/components/topbar/`. Replace header sections in
`AppLayout`/`ClientLayout`/`AdminLayout`; keep content slots untouched.

## 8. C — UI redesign

Two phases:

C1: Design tokens — extract Ant theme into a tokens layer:
- Color: brand primary (current teal), accent, 6-step grayscale, semantic.
- Typography: ru-friendly (Inter or similar), 6-step scale.
- Spacing: 4-px base, 8 step.
- Radius: 4 / 8 / 12 / 16.
- Shadows: 3 elevations + focus ring.
- Motion: 150 ms hover, 300 ms ease-out sheets.

Apply via `ConfigProvider.theme.token` so antd inherits; expose as CSS vars.

C2: ClientLayout deep revamp — hero search, results grid, bathhouse detail, booking
wizard, profile, wallet — all restyled. Owner / Admin: token retrofit (look feels new)
without structural changes.

Out of scope: full revamp of every Owner / Admin screen.

## 9. E — Prod data gap fill

Last step. Order:

1. Run `scripts/audit/diff-reference-tables.sh` comparing row counts of allowed reference
   tables between local and prod.
2. For each gap, ensure `seed_demo.go` (or `seed_reference.go`) covers it idempotently.
3. Build + push API image via existing CI.
4. Run `kubectl create job --from=cronjob/relax-hub-seed-reference …` (new CronJob,
   reference-only).

Allowed (whitelist):
```
amenities, object_types, holidays, platform_settings, feature_flags,
service_fee_configs, cities, faq, response_templates, admin_permissions, admin_roles
```

Disallowed (hard refusal in code):
```
users, wallets, wallet_transactions, payments, bookings, reviews,
client_reviews, disputes, support_tickets, complaints, audit_logs,
saved_cards, sessions, kyc_applications, owner_payment_details
```

## 10. Risks & mitigations

| Risk | Mitigation |
|---|---|
| Audit takes 10+ h, context dies mid-session | Commit + push after each role's fix-pack |
| Fix breaks something else | `go test ./... -race` + `npx vitest run` before merge; CI gate |
| seed-demo UUIDs collide | Deterministic UUIDs, `INSERT … ON CONFLICT DO NOTHING` |
| Prod data fill writes to disallowed table | Whitelist + integration test fails build on disallowed reference |
| Menu redesign breaks current users | Feature-flag `ui.new_menu` defaults off; opt-in until verified |
| UI tokens break existing antd themes | Apply incrementally, visual smoke per layout |

## 11. Sequencing & checkpoints

```
Session 1 (today):
  ├─ Pre-flight: stack up, telemetry on, seed loaded
  ├─ A1: audit admin
  └─ Fix-pack admin → PR → merge
Session 2:
  ├─ A2: audit owner
  └─ Fix-pack owner
Session 3:
  ├─ A3: rep + A4: client
  └─ Fix-pack client
Session 4: D query consolidation
Sessions 5–6: B menu + C1 tokens
Session 7+: C2 ClientLayout revamp
Final: E prod data gap fill
```

## 12. Acceptance

- [ ] Zero open Critical or High in audit MD across four roles.
- [ ] Top-20 SQL fixed or accepted with reasoning.
- [ ] New menu live in all five layouts (or feature-flagged for staging).
- [ ] Design tokens layer in place; ClientLayout visually new.
- [ ] Prod missing-reference-row count = 0 against the whitelist.
