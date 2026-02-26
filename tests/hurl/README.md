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

## Test Data Setup and Teardown

### Test Database Initialization

The `setup.sh` script handles all test database operations:

1. **Database Creation** - Creates `bani_test` database if needed
2. **Schema Reset** - Drops and recreates the public schema
3. **Migrations** - Runs all database migrations in order
4. **Seed Data** - Inserts test users, cities, bathhouses, bookings, and reviews

### Teardown

To reset the test database between test runs:

```bash
cd tests/hurl
./setup.sh
```

This safely clears all test data and re-seeds from scratch.

### Test User Credentials

Use these credentials in hurl tests to authenticate:

```
admin@test.com / admin-password (admin role)
owner@test.com / owner-password (owner role)
representative@test.com / rep-password (representative role)
client@test.com / client-password (client role)
client2@test.com / client2-password (client role, for multi-user tests)
blocked@test.com / blocked-password (client role, blocked status)
```

## Running Individual Test Suites

Each test module can be run independently after setup:

### Authentication Tests
```bash
hurl --env-file tests/hurl/.env.test tests/hurl/auth.hurl
hurl --env-file tests/hurl/.env.test tests/hurl/auth_negative.hurl
```

### Bathhouse Tests
```bash
hurl --env-file tests/hurl/.env.test tests/hurl/bathhouses.hurl
hurl --env-file tests/hurl/.env.test tests/hurl/bathhouses_admin.hurl
```

### Booking Tests
```bash
hurl --env-file tests/hurl/.env.test tests/hurl/bookings.hurl
hurl --env-file tests/hurl/.env.test tests/hurl/bookings_negative.hurl
```

### Review Tests
```bash
hurl --env-file tests/hurl/.env.test tests/hurl/reviews.hurl
hurl --env-file tests/hurl/.env.test tests/hurl/reviews_negative.hurl
```

### Favorite and Representative Tests
```bash
hurl --env-file tests/hurl/.env.test tests/hurl/favorites.hurl
hurl --env-file tests/hurl/.env.test tests/hurl/representatives.hurl
```

### Admin Tests
```bash
hurl --env-file tests/hurl/.env.test tests/hurl/admin_users.hurl
hurl --env-file tests/hurl/.env.test tests/hurl/admin_bathhouses.hurl
hurl --env-file tests/hurl/.env.test tests/hurl/admin_cities.hurl
```

### Public Endpoint Tests
```bash
hurl --env-file tests/hurl/.env.test tests/hurl/cities.hurl
hurl --env-file tests/hurl/.env.test tests/hurl/health.hurl
```

## Test Execution Order

For a complete test run, execute tests in this order to ensure dependencies are satisfied:

1. Task 1: Setup (setup.sh)
2. Task 2: Auth endpoints
3. Task 3: Bathhouses endpoints
4. Task 4: Bookings endpoints
5. Task 5: Reviews endpoints
6. Task 6: Favorites and Representatives
7. Task 7: Admin endpoints
8. Task 8: Cities and Health endpoints

Each task must pass completely before moving to the next. Use `make test-hurl` to run the complete suite in order.

## Test Data Dependencies

Tests have implicit dependencies on test data state:

- Auth tests create and validate new users
- Bathhouse tests depend on users created by auth tests
- Booking tests depend on bathhouses created by bathhouse tests
- Review tests depend on bookings created by booking tests
- Favorite tests can run independently once users exist
- Representative tests depend on bathhouses existing
- Admin tests can run independently (they operate on any data)
- Health and Cities tests have no dependencies

The `run_all_tests.sh` script manages these dependencies by executing tests in the correct order and capturing IDs that are passed between test files.

## Common Issues and Solutions

### "Connection refused" error
- Ensure PostgreSQL is running: `pg_isready`
- Ensure Redis is running: `redis-cli ping`
- Check that the test database exists: `psql -l | grep bani_test`

### "Permission denied" on setup.sh
- Make script executable: `chmod +x tests/hurl/setup.sh`

### Tests fail after code changes
- Re-run setup to reset test data: `./setup.sh`
- Verify server is running: `curl http://localhost:8080/health`

### "Assertion failed" errors
- Run with verbose output: `hurl --verbose ...`
- Check that expected HTTP status codes are returned
- Verify response format matches API contract

### Token expiration
- Tokens are generated during test execution
- Tests login and capture tokens for immediate use
- No token expiration issues during a single test run
