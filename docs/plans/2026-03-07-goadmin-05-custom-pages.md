# GoAdmin: Custom Pages (Dashboard, Moderation Center, Analytics Dashboard)

## Overview

Custom GoAdmin pages beyond CRUD tables: admin dashboard with KPIs, review moderation center, analytics dashboard with charts, platform health monitor.

## Context

- Files involved: `internal/admin/pages/`, `internal/admin/pages/templates/`, `internal/admin/engine.go`
- Related patterns: GoAdmin custom page API, Chart.js integration, Go templates
- Dependencies: Plans 1-4 must be completed first

## Development Approach

- **Testing approach**: Regular (code first, then tests)
- Complete each task fully before moving to the next
- Custom pages use GoAdmin's page API with Go HTML templates
- **CRITICAL: every task MUST include new/updated tests**
- **CRITICAL: all tests must pass before starting next task**

## Implementation Steps

### Task 1: Admin Dashboard Page

**Files:**
- Create: `internal/admin/pages/dashboard.go`
- Create: `internal/admin/pages/templates/dashboard.tmpl`

- [x] KPI cards row: total users, total bathhouses, total bookings (today/week/month), total revenue (today/week/month)
- [x] Status cards: pending bathhouses count, pending reviews count, active subscriptions count
- [x] Recent activity feed: last 10 bookings, last 10 reviews, last 10 registrations
- [x] Quick action buttons: go to pending bathhouses, go to pending reviews
- [x] Data sourced via direct SQL queries (GoAdmin DB connection)
- [x] Set as GoAdmin default landing page

### Task 2: Review Moderation Center

**Files:**
- Create: `internal/admin/pages/moderation.go`
- Create: `internal/admin/pages/templates/moderation.tmpl`

- [ ] Queue view: pending reviews listed with full text, rating, bathhouse name, user name, images
- [ ] One-click approve/reject buttons per review
- [ ] Reject with reason selection (checkboxes for common reasons)
- [ ] Batch select and approve/reject
- [ ] Filter by bathhouse, rating, date
- [ ] Statistics: approved/rejected/pending counts today/this week
- [ ] API endpoints for AJAX approve/reject actions

### Task 3: Analytics Dashboard

**Files:**
- Create: `internal/admin/pages/analytics.go`
- Create: `internal/admin/pages/templates/analytics.tmpl`

- [ ] Platform-wide charts (GoAdmin supports Chart.js):
  - Bookings per day (line chart, last 30 days)
  - Revenue per day (bar chart, last 30 days)
  - New users per day (line chart, last 30 days)
  - Top 10 bathhouses by bookings (horizontal bar)
  - Top 10 bathhouses by revenue (horizontal bar)
  - Booking status distribution (pie chart)
  - Review rating distribution (pie chart)
- [ ] Date range picker for filtering
- [ ] City filter for region-specific analytics
- [ ] Data sourced from analytics_snapshots + aggregation queries

### Task 4: Platform Health Monitor

**Files:**
- Create: `internal/admin/pages/health.go`
- Create: `internal/admin/pages/templates/health.tmpl`

- [ ] Service status: PostgreSQL connection, Redis connection, S3 storage
- [ ] System metrics: active WebSocket connections, notification queue size
- [ ] Subscription health: expiring subscriptions this week
- [ ] Moderation backlog: reviews waiting > 24h, > 48h, > 72h
- [ ] Auto-refresh every 30 seconds

### Task 5: Register Custom Pages and Menu

**Files:**
- Modify: `internal/admin/engine.go`

- [ ] Register all 4 custom pages with GoAdmin
- [ ] Menu: Dashboard as first item (home icon)
- [ ] Menu: "Operations" group with Moderation Center, Analytics Dashboard, Platform Health
- [ ] Set dashboard as default page after login

### Task 6: Final Verification

- [ ] Build: go build ./...
- [ ] Start with --with-admin
- [ ] Verify dashboard loads with real data
- [ ] Test moderation center approve/reject workflow
- [ ] Verify analytics charts render
- [ ] Check health monitor shows correct service statuses
- [ ] Run full test suite: go test ./... -v
- [ ] Run linter: make lint
- [ ] Update CLAUDE.md with admin panel section

### Task 7: Update Documentation

- [ ] Add admin panel section to CLAUDE.md: commands, config vars, architecture
- [ ] Move all 5 plan files to `docs/plans/completed/`
