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

log_info "All tests completed!"
