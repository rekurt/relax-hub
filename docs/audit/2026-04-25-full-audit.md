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

(populated in Phase 2)

### Owner role

(Session 2)

### Representative role

(Session 3)

### Client role

(Session 3)

## Per-role pg_stat_statements snapshots

(populated end of each role pass)
