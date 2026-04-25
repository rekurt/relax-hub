# Session 1: Pre-flight + Admin Audit + Fix-pack — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Bring up the local stack with telemetry, conduct a full-flow audit of the admin role, capture all Critical/High bugs, ship a fix-pack PR.

**Architecture:** Local docker-compose for postgres + redis + minio. Backend Go server runs on host (`go run`), Vite frontend runs on host. Browser audit driven through Chrome MCP. Telemetry via `pg_stat_statements` (server-side) and `read_network_requests` (client-side). Findings appended to `docs/audit/2026-04-25-full-audit.md` continuously; fixes shipped as one PR-per-role-bucket.

**Tech Stack:** docker-compose, Go + chi + pgx, Vite + React + Ant Design, Chrome MCP (`mcp__Claude_in_Chrome__*`), `pg_stat_statements` extension, golangci-lint, vitest.

---

## Phase 0 — File map

**Created:**
- `docs/audit/2026-04-25-full-audit.md` — bug log appended throughout.
- `docs/audit/2026-04-25-admin-role-pgss-snapshot.csv` — top-50 SQL after admin pass.

**Modified during fix-pack** (specific files unknown until audit completes — pattern: `internal/handler/*.go`, `internal/service/*.go`, `internal/repository/postgres/*.go`, `frontend/src/pages/admin/*.tsx`, `frontend/src/components/*.tsx`).

**Branches/PRs:**
- `audit/session-1-preflight` — temporary branch for the audit log itself.
- `fix/audit-admin-pack` — fix-pack PR for admin role.

---

## Phase 1 — Pre-flight: local stack

### Task 1.1: Verify Docker is running

- [ ] **Step 1.1.1: Check Docker daemon**

Run: `docker version --format '{{.Server.Version}}' 2>&1 || echo NO-DOCKER`
Expected: a version string (e.g. `27.x.x`). If `NO-DOCKER`, abort with operator note.

### Task 1.2: Bring stack up

- [ ] **Step 1.2.1: Start postgres + redis + minio detached**

Run: `docker compose up -d postgres redis minio`
Expected: three containers `Created`/`Started`. Stay on host for Go + Vite.

- [ ] **Step 1.2.2: Wait for postgres healthy**

Run:
```bash
for i in 1 2 3 4 5 6 7 8 9 10; do
  if docker compose exec -T postgres pg_isready -U postgres >/dev/null 2>&1; then
    echo "postgres ready"; break; fi
  sleep 2
done
```
Expected: `postgres ready` within 20 s.

- [ ] **Step 1.2.3: Apply migrations**

Run: `make migrate-up 2>&1 | tail -20`
Expected: last line indicates migration finished.

### Task 1.3: Enable telemetry

- [ ] **Step 1.3.1: Install pg_stat_statements**

Run:
```bash
docker compose exec -T postgres psql -U postgres -d bani -c "CREATE EXTENSION IF NOT EXISTS pg_stat_statements;"
docker compose exec -T postgres psql -U postgres -d bani -c "ALTER SYSTEM SET shared_preload_libraries = 'pg_stat_statements';"
```
Expected: `CREATE EXTENSION` (or "already exists") + silent ALTER SYSTEM.

- [ ] **Step 1.3.2: Restart postgres for `shared_preload_libraries` to take effect**

Run: `docker compose restart postgres && for i in 1 2 3 4 5; do docker compose exec -T postgres pg_isready -U postgres >/dev/null && break; sleep 2; done`
Expected: postgres back up.

- [ ] **Step 1.3.3: Verify pg_stat_statements is loaded**

Run: `docker compose exec -T postgres psql -U postgres -d bani -c "SELECT count(*) FROM pg_stat_statements;"`
Expected: a number ≥ 0 (no error).

### Task 1.4: Seed demo data

- [ ] **Step 1.4.1: Run seed-demo**

Run: `make seed-demo 2>&1 | tail -10`
Expected: log lines like `Client demo account: demo.client1@relax-hub.ru / +79990000031`.

- [ ] **Step 1.4.2: Verify seed users**

Run:
```bash
docker compose exec -T postgres psql -U postgres -d bani -tAc \
  "SELECT email, role FROM users WHERE email LIKE 'demo.%@relax-hub.ru' ORDER BY email LIMIT 10;"
```
Expected: at least `demo.admin@relax-hub.ru|admin`, `demo.owner1@relax-hub.ru|owner`, `demo.client1@relax-hub.ru|client`.

### Task 1.5: Start API + frontend

- [ ] **Step 1.5.1: Start Go server in background**

Run:
```bash
BANI_LOGGER_LEVEL=debug BANI_LOGGER_FORMAT=console go run ./cmd/server serve > /tmp/api.log 2>&1 &
```
Use `run_in_background: true`. Save PID for cleanup.

- [ ] **Step 1.5.2: Verify API healthy**

Run:
```bash
for i in 1 2 3 4 5 6 7 8 9 10; do
  curl -fs http://localhost:8080/health >/dev/null 2>&1 && echo "api ready" && break
  sleep 2
done
```
Expected: `api ready`.

- [ ] **Step 1.5.3: Start Vite in background**

Run:
```bash
(cd frontend && npm run dev > /tmp/vite.log 2>&1 &)
```
`run_in_background: true`.

- [ ] **Step 1.5.4: Verify frontend serves**

Run: `for i in 1 2 3 4 5 6 7 8 9 10; do curl -fs http://localhost:5173/ >/dev/null 2>&1 && echo vite-ready && break; sleep 2; done`
Expected: `vite-ready`.

### Task 1.6: Initialize Chrome MCP

- [ ] **Step 1.6.1: Get/create Chrome tab**

Use `mcp__Claude_in_Chrome__tabs_context_mcp` with `createIfEmpty: true`. Save returned `tabId`.

- [ ] **Step 1.6.2: Navigate to localhost frontend**

`mcp__Claude_in_Chrome__navigate` with `url: http://localhost:5173/login`. Wait 3 s.

### Task 1.7: Initialize audit MD

- [ ] **Step 1.7.1: Create audit log file**

Write `docs/audit/2026-04-25-full-audit.md` with this header:

```markdown
# RelaxHUB Full-flow Audit — 2026-04-25

## Severity legend
- Critical — 5xx on hot path, frontend crash, permission bypass, data loss, auth/TLS break.
- High — silent submit failure, wrong data shown, missing pagination on >200-row list, hot-path query > 2 s.
- Medium — empty state unclear, missing skeleton, slow query 500 ms–2 s.
- Low — color/spacing/typography drift.

## Findings

### Admin role

(populated in Phase 2)

### Owner role

### Representative role

### Client role

## Per-role pg_stat_statements snapshots
```

- [ ] **Step 1.7.2: Commit pre-flight setup**

```bash
git checkout -b audit/session-1-preflight
git add docs/audit/2026-04-25-full-audit.md
git commit -m "audit(session-1): bootstrap audit MD"
```

---

## Phase 2 — A1: Admin role audit

### Task 2.0: Reset SQL telemetry + login

- [ ] **Step 2.0.1: Reset pg_stat_statements**

Run: `docker compose exec -T postgres psql -U postgres -d bani -c "SELECT pg_stat_statements_reset();"`
Expected: a single row.

- [ ] **Step 2.0.2: Login via UI as admin**

Use Chrome MCP `browser_batch`:
1. navigate `http://localhost:5173/login?_cb=1`
2. wait 3
3. javascript_tool to set form values (autofill-bypass pattern):
```js
(() => {
  const s = Object.getOwnPropertyDescriptor(window.HTMLInputElement.prototype,'value').set;
  const e = document.querySelector('input#email');
  const p = document.querySelector('input#password');
  s.call(e, 'demo.admin@relax-hub.ru'); e.dispatchEvent(new Event('input',{bubbles:true}));
  s.call(p, 'DemoPass123!'); p.dispatchEvent(new Event('input',{bubbles:true}));
  return 'filled';
})()
```
4. find `submit button "Войти"` → click
5. wait 5
6. screenshot — confirm landed on `/admin`

If login fails, abort phase 2 with finding logged.

### Per-page audit pattern

For each admin page:

1. Navigate to URL.
2. Wait 4 s.
3. `read_network_requests` (clear=true) — note status >= 400, duplicates within 1 s, total > 800 ms, requests with unused responses.
4. Screenshot.
5. One happy-path interaction (per page).
6. `read_network_requests` again (interaction phase).
7. Append findings to audit MD under `### Admin role > <page>`.

### Task 2.1 — Batch 1: Dashboard + Moderation

Pages:
1. `/admin` — AdminDashboard
2. `/admin/moderation/bathhouses` — BathhouseModeration (interaction: filter status=pending, click first row)
3. `/admin/moderation/reviews` — ReviewModeration (interaction: open detail of first review)
4. `/admin/moderation/photos` — PhotoVerification
5. `/admin/moderation/complaints` — ComplaintManagement

- [ ] **Steps 2.1.1–2.1.5**: per-page pattern.
- [ ] **Step 2.1.6**: commit audit MD.

```bash
git add docs/audit/2026-04-25-full-audit.md
git commit -m "audit(admin): batch 1 dashboard + moderation"
```

### Task 2.2 — Batch 2: Users + Content

Pages: UserManagement, RoleManagement, CityManagement, AmenityManagement, ObjectTypeManagement, HolidayManagement.

- [ ] **Steps 2.2.1–2.2.6**: per-page pattern.
- [ ] **Step 2.2.7**: commit.

### Task 2.3 — Batch 3: Finance

Pages: AdminFinanceDashboard, BankReconciliation, WalletManagement, BookingManagement.

- [ ] **Steps 2.3.1–2.3.4**: per-page pattern.
- [ ] **Step 2.3.5**: commit.

### Task 2.4 — Batch 4: Support

Pages: TicketManagement, AdminTicketDetail, DisputeManagement, AdminDisputeDetail.

- [ ] **Steps 2.4.1–2.4.4**: pattern.
- [ ] **Step 2.4.5**: commit.

### Task 2.5 — Batch 5: Analytics

Pages: ConversionFunnels, CohortAnalysis, SupplyDemandMetrics, GeoHeatmap, AntiFraudDashboard.

- [ ] **Steps 2.5.1–2.5.5**: pattern.
- [ ] **Step 2.5.6**: commit.

### Task 2.6 — Batch 6: Configuration

Pages: PlatformSettings, FeatureFlags, ServiceFeeConfig, SubscriptionManagement, LoyaltyManagement.

- [ ] **Steps 2.6.1–2.6.5**: pattern.
- [ ] **Step 2.6.6**: commit.

### Task 2.7 — Batch 7: Misc admin

Pages: AdminProfile, AdminAuditLog, AdminNotifications, AdminNotificationCenter, ForceMajeure, FAQManagement, GlobalPromoCodes, CertificateManagement.

- [ ] **Steps 2.7.1–2.7.8**: pattern.
- [ ] **Step 2.7.9**: commit.

### Task 2.8 — Snapshot pg_stat_statements

- [ ] **Step 2.8.1: Snapshot top-50 by total time**

```bash
docker compose exec -T postgres psql -U postgres -d bani -c "
COPY (
  SELECT
    round(total_exec_time::numeric, 2) AS total_ms,
    calls,
    round(mean_exec_time::numeric, 2) AS mean_ms,
    rows,
    substring(query, 1, 200) AS query_preview
  FROM pg_stat_statements
  ORDER BY total_exec_time DESC
  LIMIT 50
) TO STDOUT WITH CSV HEADER
" > docs/audit/2026-04-25-admin-role-pgss-snapshot.csv
```

Expected: ≤50 rows in the CSV.

- [ ] **Step 2.8.2: Append snapshot summary to audit MD**

In `## Per-role pg_stat_statements snapshots` add a markdown section linking to the CSV and listing top-10 with `total_ms`.

- [ ] **Step 2.8.3: Commit + push**

```bash
git add docs/audit/2026-04-25-admin-role-pgss-snapshot.csv docs/audit/2026-04-25-full-audit.md
git commit -m "audit(admin): pg_stat_statements snapshot end-of-pass"
git push -u origin audit/session-1-preflight
```

---

## Phase 3 — Fix-pack admin

### Task 3.0: Triage

- [ ] **Step 3.0.1: List Critical/High from audit MD**

Open `docs/audit/2026-04-25-full-audit.md`. For each Critical or High in admin section, note severity, file pointer, root cause hypothesis.

- [ ] **Step 3.0.2: Open fix branch**

```bash
git checkout -b fix/audit-admin-pack
```

### Task 3.X — One sub-task PER bug (template)

```
### Task 3.X: <bug short title> (<severity>)

**Files:**
- Modify: <exact file:line>
- Test: <test file>

- [ ] **Step 3.X.1: Write failing test**
  <actual test code>
- [ ] **Step 3.X.2: Run, confirm failure**
  Run: <go test or vitest>
  Expected: FAIL with <message>.
- [ ] **Step 3.X.3: Implement fix**
  <actual code>
- [ ] **Step 3.X.4: Run test, confirm pass**
- [ ] **Step 3.X.5: Commit**
  ```
  git add <files>
  git commit -m "fix(admin): <bug title>"
  ```
```

For UI-only fixes without easy unit tests: add a vitest component test, or file a Medium "no test coverage" note and verify manually via Chrome MCP.

### Task 3.99: Full test suite + lint

- [ ] **Step 3.99.1**: `go test ./... -race 2>&1 | tail -20` → PASS.
- [ ] **Step 3.99.2**: `go vet ./... && golangci-lint run ./... 2>&1 | tail -20` → clean.
- [ ] **Step 3.99.3**: `cd frontend && npx vitest run 2>&1 | tail -10` → green.
- [ ] **Step 3.99.4**: `cd frontend && npx tsc --noEmit 2>&1 | tail -10` → no errors.

### Task 3.100: PR

- [ ] **Step 3.100.1: Push + open PR**

```bash
git push -u origin fix/audit-admin-pack
gh pr create --title "fix(audit): admin role fix-pack" --body "$(cat <<'EOF'
## Summary
First fix-pack from the full-flow audit. See \`docs/audit/2026-04-25-full-audit.md\`
for the bug list.

## Bugs fixed
<one bullet per Critical/High bug, severity-tagged>

## Test plan
- [x] go test ./... -race
- [x] go vet + golangci-lint
- [x] vitest
- [x] tsc --noEmit
- [ ] Re-smoke admin role on prod after CI deploy

EOF
)"
```

- [ ] **Step 3.100.2**: After CI green, `gh pr merge --squash --delete-branch <PR>`.

### Task 3.101: Re-smoke

- [ ] **Step 3.101.1**: After CI deploy, smoke broken paths only on `https://relax-hub.ru/admin`. Mark each fixed bug ✅ + PR link in audit MD.
- [ ] **Step 3.101.2: Final commit**

```bash
git add docs/audit/2026-04-25-full-audit.md
git commit -m "audit(admin): re-smoke complete, all admin Critical/High closed"
git push origin master
```

---

## Phase 4 — Wrap

- [ ] **Step 4.1**: Update todos.
- [ ] **Step 4.2: Stop background processes**

```bash
pkill -f "go run ./cmd/server serve" || true
pkill -f "vite" || true
docker compose down
```

- [ ] **Step 4.3**: Save next-session entry note at the bottom of audit MD: "Session 2 next: A2 owner role audit + fix-pack."

---

## Self-Review

- **Spec coverage**: Pre-flight (spec §2) → Phase 1. A1 admin (spec §3) → Phase 2. Fix-pack discipline (spec §4) → Phase 3. B/C/D/E explicitly out of scope for Session 1.
- **Placeholders**: Task 3.X is intentionally a template — concrete bugs are unknown until Phase 2 finishes; `executing-plans` instantiates per discovered bug.
- **Type consistency**: emails consistent (`demo.admin@relax-hub.ru`); branch names consistent.
- **Risk re-check**: long-session safety covered by per-batch commits in Phase 2.
