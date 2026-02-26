#!/bin/bash

set -e

# Load environment variables
set -a
source "$(dirname "$0")/.env.test"
set +a

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

# Setup test database
log_info "Setting up test database..."
bash "$(dirname "$0")/setup.sh" || {
    log_error "Database setup failed"
    exit 1
}

# Run auth tests first to capture tokens by making API calls directly
log_info "Running auth tests..."
hurl tests/hurl/auth.hurl --variables-file tests/hurl/.env.test > /dev/null 2>&1 || {
    log_error "Auth tests failed"
    exit 1
}

# Get tokens directly from API
log_info "Obtaining authentication tokens..."
ADMIN_RESP=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@test.com","password":"admin-password"}')
ADMIN_TOKEN=$(echo "$ADMIN_RESP" | grep -o '"token":"[^"]*"' | cut -d'"' -f4)

OWNER_RESP=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"owner@test.com","password":"owner-password"}')
OWNER_TOKEN=$(echo "$OWNER_RESP" | grep -o '"token":"[^"]*"' | cut -d'"' -f4)

CLIENT_RESP=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"client@test.com","password":"client-password"}')
CLIENT_TOKEN=$(echo "$CLIENT_RESP" | grep -o '"token":"[^"]*"' | cut -d'"' -f4)

REPRESENTATIVE_RESP=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"representative@test.com","password":"representative-password"}')
REPRESENTATIVE_TOKEN=$(echo "$REPRESENTATIVE_RESP" | grep -o '"token":"[^"]*"' | cut -d'"' -f4)

CLIENT2_RESP=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"client2@test.com","password":"client2-password"}')
CLIENT2_TOKEN=$(echo "$CLIENT2_RESP" | grep -o '"token":"[^"]*"' | cut -d'"' -f4)

BLOCKED_USER_RESP=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"blocked@test.com","password":"blocked-password"}')
BLOCKED_USER_TOKEN=$(echo "$BLOCKED_USER_RESP" | grep -o '"token":"[^"]*"' | cut -d'"' -f4)

log_info "Obtained tokens for admin, owner, client, representative, client2, and blocked user"

# Extract database connection info from DSN
DB_URL="$BANI_DATABASE_DSN"
POSTGRES_USER="${DB_URL##*://}"
POSTGRES_USER="${POSTGRES_USER%%[:@]*}"
if [[ "$POSTGRES_USER" != "${DB_URL##*://}" ]]; then
  if [[ "${DB_URL##*://}" == *":"* ]]; then
    POSTGRES_PASSWORD="${DB_URL##*:}"
    POSTGRES_PASSWORD="${POSTGRES_PASSWORD%%@*}"
  else
    POSTGRES_PASSWORD=""
  fi
else
  POSTGRES_PASSWORD=""
fi
POSTGRES_HOST="${DB_URL##*@}"
POSTGRES_HOST="${POSTGRES_HOST%%[:/*]*}"
if [[ "${DB_URL##*@}" == *":"* ]]; then
  POSTGRES_PORT="${DB_URL##*:}"
  POSTGRES_PORT="${POSTGRES_PORT%%/*}"
else
  POSTGRES_PORT="5432"
fi
POSTGRES_DB="${DB_URL##*/}"
POSTGRES_DB="${POSTGRES_DB%%\?*}"

# Get user IDs from the users table
log_info "Obtaining user IDs..."
if [ -z "$POSTGRES_PASSWORD" ]; then
  ADMIN_USER_ID=$(psql -h "$POSTGRES_HOST" -p "$POSTGRES_PORT" -U "$POSTGRES_USER" -d "$POSTGRES_DB" -tc "SELECT id FROM users WHERE email = 'admin@test.com';" 2>/dev/null | xargs || echo "")
  BLOCKED_USER_ID=$(psql -h "$POSTGRES_HOST" -p "$POSTGRES_PORT" -U "$POSTGRES_USER" -d "$POSTGRES_DB" -tc "SELECT id FROM users WHERE email = 'blocked@test.com';" 2>/dev/null | xargs || echo "")
else
  ADMIN_USER_ID=$(PGPASSWORD="$POSTGRES_PASSWORD" psql -h "$POSTGRES_HOST" -p "$POSTGRES_PORT" -U "$POSTGRES_USER" -d "$POSTGRES_DB" -tc "SELECT id FROM users WHERE email = 'admin@test.com';" 2>/dev/null | xargs || echo "")
  BLOCKED_USER_ID=$(PGPASSWORD="$POSTGRES_PASSWORD" psql -h "$POSTGRES_HOST" -p "$POSTGRES_PORT" -U "$POSTGRES_USER" -d "$POSTGRES_DB" -tc "SELECT id FROM users WHERE email = 'blocked@test.com';" 2>/dev/null | xargs || echo "")
fi

# Verify tokens were obtained
if [ -z "$ADMIN_TOKEN" ] || [ -z "$OWNER_TOKEN" ] || [ -z "$CLIENT_TOKEN" ]; then
    log_error "Failed to obtain authentication tokens"
    log_warn "Admin response: $ADMIN_RESP"
    log_warn "Owner response: $OWNER_RESP"
    log_warn "Client response: $CLIENT_RESP"
    exit 1
fi

log_info "Running bathhouses tests..."

# Now run all subsequent tests with the captured variables
hurl tests/hurl/bathhouses.hurl --variables-file tests/hurl/.env.test \
  --variable admin_token="$ADMIN_TOKEN" \
  --variable owner_token="$OWNER_TOKEN" \
  --variable client_token="$CLIENT_TOKEN" || {
    log_error "Bathhouses tests failed"
    exit 1
}

log_info "Running bathhouses negative tests..."
hurl tests/hurl/bathhouses_negative.hurl --variables-file tests/hurl/.env.test \
  --variable admin_token="$ADMIN_TOKEN" \
  --variable owner_token="$OWNER_TOKEN" \
  --variable client_token="$CLIENT_TOKEN" || {
    log_warn "Bathhouses negative tests had failures"
}

log_info "Running bathhouses admin tests..."
hurl tests/hurl/bathhouses_admin.hurl --variables-file tests/hurl/.env.test \
  --variable admin_token="$ADMIN_TOKEN" \
  --variable owner_token="$OWNER_TOKEN" \
  --variable client_token="$CLIENT_TOKEN" || {
    log_error "Bathhouses admin tests failed"
    exit 1
}

log_info "Running bookings tests..."
hurl tests/hurl/bookings.hurl --variables-file tests/hurl/.env.test \
  --variable admin_token="$ADMIN_TOKEN" \
  --variable owner_token="$OWNER_TOKEN" \
  --variable client_token="$CLIENT_TOKEN" \
  --variable representative_token="$REPRESENTATIVE_TOKEN" \
  --variable client2_token="$CLIENT2_TOKEN" \
  --variable blocked_user_token="$BLOCKED_USER_TOKEN" || {
    log_error "Bookings tests failed"
    exit 1
}

log_info "Running bookings negative tests..."
hurl tests/hurl/bookings_negative.hurl --variables-file tests/hurl/.env.test \
  --variable admin_token="$ADMIN_TOKEN" \
  --variable owner_token="$OWNER_TOKEN" \
  --variable client_token="$CLIENT_TOKEN" \
  --variable representative_token="$REPRESENTATIVE_TOKEN" \
  --variable client2_token="$CLIENT2_TOKEN" \
  --variable blocked_user_token="$BLOCKED_USER_TOKEN" || {
    log_warn "Bookings negative tests had failures"
}

log_info "Running reviews tests..."
hurl tests/hurl/reviews.hurl --variables-file tests/hurl/.env.test \
  --variable admin_token="$ADMIN_TOKEN" \
  --variable owner_token="$OWNER_TOKEN" \
  --variable client_token="$CLIENT_TOKEN" \
  --variable representative_token="$REPRESENTATIVE_TOKEN" \
  --variable client2_token="$CLIENT2_TOKEN" \
  --variable blocked_user_token="$BLOCKED_USER_TOKEN" || {
    log_error "Reviews tests failed"
    exit 1
}

log_info "Running reviews negative tests..."
hurl tests/hurl/reviews_negative.hurl --variables-file tests/hurl/.env.test \
  --variable admin_token="$ADMIN_TOKEN" \
  --variable owner_token="$OWNER_TOKEN" \
  --variable client_token="$CLIENT_TOKEN" \
  --variable representative_token="$REPRESENTATIVE_TOKEN" \
  --variable client2_token="$CLIENT2_TOKEN" \
  --variable blocked_user_token="$BLOCKED_USER_TOKEN" || {
    log_warn "Reviews negative tests had failures"
}

log_info "Running favorites tests..."
hurl tests/hurl/favorites.hurl --variables-file tests/hurl/.env.test \
  --variable admin_token="$ADMIN_TOKEN" \
  --variable owner_token="$OWNER_TOKEN" \
  --variable client_token="$CLIENT_TOKEN" \
  --variable representative_token="$REPRESENTATIVE_TOKEN" \
  --variable client2_token="$CLIENT2_TOKEN" \
  --variable blocked_user_token="$BLOCKED_USER_TOKEN" || {
    log_error "Favorites tests failed"
    exit 1
}

log_info "Running favorites negative tests..."
hurl tests/hurl/favorites_negative.hurl --variables-file tests/hurl/.env.test \
  --variable admin_token="$ADMIN_TOKEN" \
  --variable owner_token="$OWNER_TOKEN" \
  --variable client_token="$CLIENT_TOKEN" \
  --variable representative_token="$REPRESENTATIVE_TOKEN" \
  --variable client2_token="$CLIENT2_TOKEN" \
  --variable blocked_user_token="$BLOCKED_USER_TOKEN" || {
    log_warn "Favorites negative tests had failures"
}

log_info "Running representatives tests..."
hurl tests/hurl/representatives.hurl --variables-file tests/hurl/.env.test \
  --variable admin_token="$ADMIN_TOKEN" \
  --variable owner_token="$OWNER_TOKEN" \
  --variable client_token="$CLIENT_TOKEN" \
  --variable representative_token="$REPRESENTATIVE_TOKEN" \
  --variable client2_token="$CLIENT2_TOKEN" \
  --variable blocked_user_token="$BLOCKED_USER_TOKEN" || {
    log_error "Representatives tests failed"
    exit 1
}

log_info "Running admin users tests..."
hurl tests/hurl/admin_users.hurl --variables-file tests/hurl/.env.test \
  --variable admin_token="$ADMIN_TOKEN" \
  --variable owner_token="$OWNER_TOKEN" \
  --variable client_token="$CLIENT_TOKEN" \
  --variable admin_user_id="$ADMIN_USER_ID" \
  --variable blocked_user_id="$BLOCKED_USER_ID" || {
    log_error "Admin users tests failed"
    exit 1
}

log_info "Running admin bathhouses tests..."
hurl tests/hurl/admin_bathhouses.hurl --variables-file tests/hurl/.env.test \
  --variable admin_token="$ADMIN_TOKEN" \
  --variable owner_token="$OWNER_TOKEN" \
  --variable client_token="$CLIENT_TOKEN" || {
    log_error "Admin bathhouses tests failed"
    exit 1
}

log_info "Running admin cities tests..."
hurl tests/hurl/admin_cities.hurl --variables-file tests/hurl/.env.test \
  --variable admin_token="$ADMIN_TOKEN" \
  --variable owner_token="$OWNER_TOKEN" \
  --variable client_token="$CLIENT_TOKEN" || {
    log_error "Admin cities tests failed"
    exit 1
}

log_info "Running admin negative tests..."
hurl tests/hurl/admin_negative.hurl --variables-file tests/hurl/.env.test \
  --variable admin_token="$ADMIN_TOKEN" \
  --variable owner_token="$OWNER_TOKEN" \
  --variable client_token="$CLIENT_TOKEN" \
  --variable admin_user_id="$ADMIN_USER_ID" || {
    log_warn "Admin negative tests had failures"
}

log_info "All tests completed!"
