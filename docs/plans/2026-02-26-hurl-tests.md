---
# Hurl Integration Tests for API Endpoints

## Overview
Create comprehensive hurl test suite covering all main API endpoints (auth, bathhouses, bookings, reviews, favorites, representatives, admin, cities). Hurl tests will validate API contracts, authentication flows, authorization, error handling, and data integrity.

## Context
- Files involved: tests/hurl/ (new directory structure)
- API routes defined in internal/server/router.go
- Authentication via JWT middleware
- Role-based access control (client, owner, representative, admin)
- Response format: {success, data, error, meta}
- Database: PostgreSQL with test data

## Development Approach
- Test-driven approach: write hurl tests covering happy paths and error cases
- Organize tests by feature (auth, bathhouses, bookings, etc.)
- Each feature has separate .hurl files
- Tests validate both successful requests and error scenarios
- Setup script to prepare test database state
- Complete each feature fully before moving to next

## Implementation Steps

### Task 1: Project Structure and Setup

**Files:**
- Create: tests/hurl/setup.sh (database setup for tests)
- Create: tests/hurl/.env.test (test environment variables)
- Create: tests/hurl/README.md (hurl tests documentation)

- [x] create tests/hurl directory structure
- [x] create setup.sh script to reset test database and seed initial data
- [x] create .env.test with database and server URL configuration
- [x] verify hurl is installed locally (brew install hurl or equivalent)
- [x] write tests for setup task

### Task 2: Auth Endpoints Tests

**Files:**
- Create: tests/hurl/auth.hurl
- Create: tests/hurl/auth_negative.hurl

- [x] test POST /auth/register (valid user, email already exists)
- [x] test POST /auth/login (valid credentials, wrong password, user not found)
- [x] test GET /auth/me (with valid token, without token, with invalid token)
- [x] verify JWT token structure in responses
- [x] test token expiration scenarios if applicable
- [x] write integration tests for auth
- [x] run hurl tests - must pass before task 3

### Task 3: Bathhouses Endpoints Tests

**Files:**
- Create: tests/hurl/bathhouses.hurl
- Create: tests/hurl/bathhouses_admin.hurl

- [x] test GET /bathhouses (search with filters, pagination, is_favorite enrichment with/without auth)
- [x] test GET /bathhouses/{id} (existing bathhouse, non-existent, with/without auth for is_favorite)
- [x] test GET /bathhouses/{id}/available-slots (date range, available slots response format)
- [x] test POST /bathhouses (owner creates bathhouse, invalid role, missing fields)
- [x] test PUT /bathhouses/{id} (owner/representative updates, forbidden access)
- [x] test DELETE /bathhouses/{id} (owner deletes, forbidden, bathhouse with bookings)
- [x] test GET /my/bathhouses (owner/representative lists, client forbidden)
- [x] write integration tests for bathhouses
- [x] run hurl tests - must pass before task 4

### Task 4: Bookings Endpoints Tests

**Files:**
- Create: tests/hurl/bookings.hurl
- Create: tests/hurl/bookings_negative.hurl

- [x] test POST /bookings (create valid booking, slot unavailable, wrong role)
- [x] test GET /bookings (user lists own bookings, pagination)
- [x] test PATCH /bookings/{id}/cancel (user cancels within deadline, too late)
- [x] test PATCH /bookings/{id}/confirm (owner/rep confirms, wrong role, booking not found)
- [x] test PATCH /bookings/{id}/reject (owner/rep rejects with reason)
- [x] test PATCH /bookings/{id}/complete (owner/rep marks complete)
- [x] test GET /bathhouses/{id}/bookings (owner/rep views bathhouse bookings, forbidden)
- [x] write integration tests for bookings
- [x] run hurl tests - must pass before task 5

### Task 5: Reviews Endpoints Tests

**Files:**
- Create: tests/hurl/reviews.hurl
- Create: tests/hurl/reviews_negative.hurl

- [x] test GET /bathhouses/{id}/reviews (public access, pagination)
- [x] test POST /bathhouses/{id}/reviews (client creates review, wrong role, duplicate)
- [x] test PUT /reviews/{id} (update own review, update others forbidden)
- [x] test DELETE /reviews/{id} (delete own review, forbidden)
- [x] test POST /reviews/{id}/response (owner responds to review, wrong role, already responded)
- [x] verify review response structure and metadata
- [x] write integration tests for reviews
- [x] run hurl tests - must pass before task 6

### Task 6: Favorites and Representatives Endpoints Tests

**Files:**
- Create: tests/hurl/favorites.hurl
- Create: tests/hurl/representatives.hurl

- [x] test POST /bathhouses/{id}/favorite (toggle favorite, non-existent bathhouse)
- [x] test GET /my/favorites (list user favorites with pagination)
- [x] test POST /bathhouses/{id}/representatives (owner invites representative)
- [x] test GET /bathhouses/{id}/representatives (owner lists representatives)
- [x] test DELETE /representatives/{id} (owner revokes, forbidden)
- [x] verify favorite toggle state
- [x] write integration tests for favorites and representatives
- [x] run hurl tests - must pass before task 7

### Task 7: Admin Endpoints Tests

**Files:**
- Create: tests/hurl/admin_users.hurl
- Create: tests/hurl/admin_bathhouses.hurl
- Create: tests/hurl/admin_cities.hurl

- [x] test GET /admin/users (admin lists users, non-admin forbidden)
- [x] test PATCH /admin/users/{id}/block (admin blocks user, user already blocked)
- [x] test PATCH /admin/users/{id}/unblock (admin unblocks user)
- [x] test GET /admin/bathhouses (admin lists pending bathhouses)
- [x] test PATCH /admin/bathhouses/{id}/approve (admin approves bathhouse)
- [x] test PATCH /admin/bathhouses/{id}/reject (admin rejects with reason)
- [x] test POST /admin/cities (admin creates city)
- [x] test PUT /admin/cities/{id} (admin updates city)
- [x] test DELETE /admin/cities/{id} (admin deletes city)
- [x] verify non-admin access is forbidden
- [x] write integration tests for admin endpoints
- [x] run hurl tests - must pass before task 8

### Task 8: Cities and Health Endpoints Tests

**Files:**
- Create: tests/hurl/cities.hurl
- Create: tests/hurl/health.hurl

- [ ] test GET /cities (public list of cities)
- [ ] test GET /health (health check, expect OK)
- [ ] test GET /ready (readiness check, expect ready or not ready)
- [ ] write integration tests for cities and health
- [ ] run hurl tests - must pass before task 9

### Task 9: Complete Integration Test Suite

- [ ] create Makefile target: make test-hurl (runs all hurl tests)
- [ ] create script to run setup and all tests in order
- [ ] verify all hurl tests pass
- [ ] run full project test suite (go test ./... -v) - must pass
- [ ] run linter (make lint)
- [ ] verify test coverage meets 80%+ (coverage.out)
- [ ] document hurl test execution in README.md
- [ ] update CLAUDE.md with hurl testing conventions if applicable

### Task 10: Documentation and Cleanup

- [ ] update tests/hurl/README.md with complete hurl testing guide
- [ ] document test data setup and teardown procedures
- [ ] document how to run individual test suites
- [ ] document environment variables and configuration
- [ ] move this plan to docs/plans/completed/
