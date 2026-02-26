# Hurl Integration Tests

This directory contains integration tests for the Bani API using Hurl, a command-line HTTP client for running HTTP tests.

## Overview

Hurl tests validate API contracts, authentication flows, authorization, error handling, and data integrity. Tests are organized by feature and cover both happy paths and error scenarios.

## Prerequisites

1. Install Hurl: `brew install hurl` (macOS) or see https://hurl.dev/docs/installation.html
2. PostgreSQL 12+ running locally
3. Redis running locally
4. Go 1.19+ (for running the server)

## Test Files

- `setup.sh` - Database setup script that creates and seeds the test database
- `.env.test` - Environment variables for test configuration
- `auth.hurl` - Authentication endpoints tests (register, login, me)
- `auth_negative.hurl` - Authentication error cases
- `bathhouses.hurl` - Bathhouse endpoints tests
- `bathhouses_admin.hurl` - Admin bathhouse operations
- `bookings.hurl` - Booking endpoints tests
- `bookings_negative.hurl` - Booking error cases
- `reviews.hurl` - Review endpoints tests
- `reviews_negative.hurl` - Review error cases
- `favorites.hurl` - Favorite endpoints tests
- `representatives.hurl` - Representative endpoints tests
- `admin_users.hurl` - Admin user management tests
- `admin_bathhouses.hurl` - Admin bathhouse management tests
- `admin_cities.hurl` - Admin city management tests
- `cities.hurl` - Public cities endpoints tests
- `health.hurl` - Health and readiness checks

## Running Tests

### Setup Test Database

Before running any tests, initialize the test database:

```bash
cd tests/hurl
./setup.sh
```

This script will:
1. Create the test database if it doesn't exist
2. Drop and recreate the public schema
3. Run all migrations
4. Seed test data (users, cities, bathhouses, bookings, reviews)

### Run All Hurl Tests

```bash
make test-hurl
```

This runs all hurl tests in sequence, requiring each to pass before continuing.

### Run Individual Test Suites

```bash
# From project root
hurl --env-file tests/hurl/.env.test tests/hurl/auth.hurl
hurl --env-file tests/hurl/.env.test tests/hurl/bathhouses.hurl
hurl --env-file tests/hurl/.env.test tests/hurl/bookings.hurl
# ... etc
```

### Run with Verbose Output

For debugging, use the `--verbose` flag:

```bash
hurl --verbose --env-file tests/hurl/.env.test tests/hurl/auth.hurl
```

### Continue on Failure

To continue running tests even if one fails:

```bash
hurl --continue --env-file tests/hurl/.env.test tests/hurl/*.hurl
```

## Environment Variables

The `.env.test` file contains:

- `BANI_ENVIRONMENT` - Set to 'test'
- `BANI_SERVER_HOST` - localhost
- `BANI_SERVER_PORT` - 8080
- `BANI_DATABASE_DSN` - PostgreSQL connection string for test database
- `BANI_REDIS_ADDR` - Redis connection address
- `BANI_JWT_SECRET` - JWT secret for test tokens
- `BANI_LOGGER_LEVEL` - Logging level (debug for tests)

## Test Data

The `setup.sh` script seeds the following test users:

| Email | Password | Role | Status |
|-------|----------|------|--------|
| admin@test.com | admin-password | admin | active |
| owner@test.com | owner-password | owner | active |
| representative@test.com | rep-password | representative | active |
| client@test.com | client-password | client | active |
| client2@test.com | client2-password | client | active |
| blocked@test.com | blocked-password | client | blocked |

Test cities:
- Moscow (latitude: 55.7558, longitude: 37.6173)
- Saint Petersburg (latitude: 59.9311, longitude: 30.3609)
- Novosibirsk (latitude: 55.0415, longitude: 82.9346)

Test bathhouses created for the owner user (owner@test.com).

## Test Structure

Each test file follows this pattern:

1. **Login** - Authenticate and store JWT token
2. **Positive Tests** - Happy path scenarios
3. **Negative Tests** - Error cases and validation

Response validation includes:
- HTTP status codes
- Response structure (success, data, error, meta fields)
- Data integrity and required fields
- Error codes and messages

## API Response Format

All endpoints return:

```json
{
  "success": true/false,
  "data": { ... },
  "error": { "code": "...", "message": "..." },
  "meta": { "page": 1, "page_size": 20, "total_count": 100, "total_pages": 5 }
}
```

## Debugging

### Check if Server is Running

```bash
curl -X GET http://localhost:8080/health
```

### Check Database Connection

```bash
PGPASSWORD=postgres psql -h localhost -U postgres -d bani_test -c "SELECT COUNT(*) FROM users;"
```

### View Test Database

```bash
PGPASSWORD=postgres psql -h localhost -U postgres -d bani_test
```

### Clear Test Database

```bash
./setup.sh
```

## Hurl Documentation

For more information on Hurl syntax and features:
- https://hurl.dev/docs/tutorial.html
- https://hurl.dev/docs/request.html
- https://hurl.dev/docs/response.html
- https://hurl.dev/docs/assert.html

## Integration with CI/CD

Tests can be integrated into CI/CD pipelines by:
1. Setting up PostgreSQL and Redis services
2. Running `./setup.sh` before tests
3. Running `make test-hurl` or individual hurl commands
4. Collecting test results and artifacts

## Test Execution Order

Tests should be executed in this order to ensure dependencies are satisfied:

1. Task 1: Setup (setup.sh)
2. Task 2: Auth endpoints
3. Task 3: Bathhouses endpoints
4. Task 4: Bookings endpoints
5. Task 5: Reviews endpoints
6. Task 6: Favorites and Representatives
7. Task 7: Admin endpoints
8. Task 8: Cities and Health endpoints

Each task must pass completely before moving to the next.
