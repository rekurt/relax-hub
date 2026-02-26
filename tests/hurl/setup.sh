#!/bin/bash

set -e

# Load test environment variables
set -a
source "$(dirname "$0")/.env.test"
set +a

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

# Extract database connection info from DSN
# Format: postgres://user:password@host:port/dbname?sslmode=disable or postgres://user@host/dbname
DB_URL="$BANI_DATABASE_DSN"
# Extract user (before : or @)
POSTGRES_USER="${DB_URL##*://}"
POSTGRES_USER="${POSTGRES_USER%%[:@]*}"
# Extract password (if present, between : and @)
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
# Extract host (after @ before : or /)
POSTGRES_HOST="${DB_URL##*@}"
POSTGRES_HOST="${POSTGRES_HOST%%[:/*]*}"
# Extract port (after host: before /)
if [[ "${DB_URL##*@}" == *":"* ]]; then
  POSTGRES_PORT="${DB_URL##*:}"
  POSTGRES_PORT="${POSTGRES_PORT%%/*}"
else
  POSTGRES_PORT="5432"
fi
# Extract database name
POSTGRES_DB="${DB_URL##*/}"
POSTGRES_DB="${POSTGRES_DB%%\?*}"

log_info "Connecting to PostgreSQL at $POSTGRES_HOST:$POSTGRES_PORT"

# Create test database if it doesn't exist
log_info "Creating test database '$POSTGRES_DB'..."
if [ -z "$POSTGRES_PASSWORD" ]; then
  psql -h "$POSTGRES_HOST" -p "$POSTGRES_PORT" -U "$POSTGRES_USER" -d postgres -tc "SELECT 1 FROM pg_database WHERE datname = '$POSTGRES_DB'" | grep -q 1 || \
    psql -h "$POSTGRES_HOST" -p "$POSTGRES_PORT" -U "$POSTGRES_USER" -d postgres -c "CREATE DATABASE \"$POSTGRES_DB\";"
else
  PGPASSWORD="$POSTGRES_PASSWORD" psql -h "$POSTGRES_HOST" -p "$POSTGRES_PORT" -U "$POSTGRES_USER" -d postgres -tc "SELECT 1 FROM pg_database WHERE datname = '$POSTGRES_DB'" | grep -q 1 || \
    PGPASSWORD="$POSTGRES_PASSWORD" psql -h "$POSTGRES_HOST" -p "$POSTGRES_PORT" -U "$POSTGRES_USER" -d postgres -c "CREATE DATABASE \"$POSTGRES_DB\";"
fi

log_info "Resetting test database schema..."
# Drop all tables and extensions
if [ -z "$POSTGRES_PASSWORD" ]; then
  psql -h "$POSTGRES_HOST" -p "$POSTGRES_PORT" -U "$POSTGRES_USER" -d "$POSTGRES_DB" << 'EOF'
DROP SCHEMA IF EXISTS public CASCADE;
CREATE SCHEMA public;
EOF
else
  PGPASSWORD="$POSTGRES_PASSWORD" psql -h "$POSTGRES_HOST" -p "$POSTGRES_PORT" -U "$POSTGRES_USER" -d "$POSTGRES_DB" << 'EOF'
DROP SCHEMA IF EXISTS public CASCADE;
CREATE SCHEMA public;
EOF
fi

log_info "Installing PostGIS extension..."
# Try to install PostGIS (may fail if not available)
if [ -z "$POSTGRES_PASSWORD" ]; then
  psql -h "$POSTGRES_HOST" -p "$POSTGRES_PORT" -U "$POSTGRES_USER" -d "$POSTGRES_DB" -c "CREATE EXTENSION IF NOT EXISTS postgis;" > /dev/null 2>&1 || true
else
  PGPASSWORD="$POSTGRES_PASSWORD" psql -h "$POSTGRES_HOST" -p "$POSTGRES_PORT" -U "$POSTGRES_USER" -d "$POSTGRES_DB" -c "CREATE EXTENSION IF NOT EXISTS postgis;" > /dev/null 2>&1 || true
fi

log_info "Running migrations..."
# Apply migrations (use no_postgis version for init)
MIGRATION_DIR="$(cd "$(dirname "$0")"/../../migrations && pwd)"
# Apply init_no_postgis.sql first, then other migrations
for migration in "$MIGRATION_DIR"/000001_init_no_postgis.up.sql "$MIGRATION_DIR"/000002_*.up.sql "$MIGRATION_DIR"/000003_*.up.sql; do
    [ -f "$migration" ] || continue
    log_info "Applying $(basename "$migration")..."
    if [ -z "$POSTGRES_PASSWORD" ]; then
      psql -h "$POSTGRES_HOST" -p "$POSTGRES_PORT" -U "$POSTGRES_USER" -d "$POSTGRES_DB" -f "$migration" > /dev/null 2>&1 || {
          log_error "Failed to apply migration: $(basename "$migration")"
          exit 1
      }
    else
      PGPASSWORD="$POSTGRES_PASSWORD" psql -h "$POSTGRES_HOST" -p "$POSTGRES_PORT" -U "$POSTGRES_USER" -d "$POSTGRES_DB" -f "$migration" > /dev/null 2>&1 || {
          log_error "Failed to apply migration: $(basename "$migration")"
          exit 1
      }
    fi
done

log_info "Seeding test data..."

# Create temporary SQL file for test data
TEST_DATA_FILE=$(mktemp)
cat > "$TEST_DATA_FILE" << 'EOF'
-- Create test users
INSERT INTO users (email, password_hash, name, phone, role, is_active) VALUES
  ('admin@test.com', '$2a$10$2R./p9GTXF/325PJnx/tvOXaxMjJZhC/mEaAZJ3FRGQiZ4U6f/JBy', 'Admin User', '+1234567890', 'admin', true),
  ('owner@test.com', '$2a$10$WiOza7gHYyBIT2dfFJsW9eo.7hS8.kuwt05VdT3tQI2Mz8BwCoBX6', 'Owner User', '+1234567891', 'owner', true),
  ('representative@test.com', '$2a$10$5etUHu7KX4g4nel2jxrXgO06z.QkkqAGC25dCGPDcTZplJz3CXFY6', 'Representative User', '+1234567892', 'representative', true),
  ('client@test.com', '$2a$10$xh7pDp6sitliSK6TJLzXj.O7dwHd8FoA.aN2B3KX6UxThvmySPvNu', 'Client User', '+1234567893', 'client', true),
  ('client2@test.com', '$2a$10$zrLx8jTPUlADKRGqWjlRM.A95yOBZnoM6LdE/eW9tMhfiZKeNoGve', 'Client Two', '+1234567894', 'client', true),
  ('blocked@test.com', '$2a$10$m5SaOpY7.Xm0/kIYrxVPhu0CsgEq2yi9dMKUsco37j0rYgDNFs/Hi', 'Blocked User', '+1234567895', 'client', false);

-- Create test cities
INSERT INTO cities (name, slug, latitude, longitude) VALUES
  ('Moscow', 'moscow', 55.7558, 37.6173),
  ('Saint Petersburg', 'saint-petersburg', 59.9311, 30.3609),
  ('Novosibirsk', 'novosibirsk', 55.0415, 82.9346);

-- Create test bathhouses (using first 3 city IDs which should exist)
INSERT INTO bathhouses (owner_id, name, description, address, city_id, latitude, longitude, price_per_hour, min_duration, max_guests, has_pool, has_sauna, has_steam_room, rating, status) VALUES
  ((SELECT id FROM users WHERE email = 'owner@test.com' LIMIT 1), 'Premium Bathhouse', 'A premium bathhouse', 'Red Square 1', 1, 55.7558, 37.6173, 500000, 1, 10, true, true, true, 4.5, 'active'),
  ((SELECT id FROM users WHERE email = 'owner@test.com' LIMIT 1), 'Cozy Sauna', 'A cozy sauna', 'Nevsky Prospect 1', 2, 59.9311, 30.3609, 300000, 1, 6, false, true, true, 4.0, 'active'),
  ((SELECT id FROM users WHERE email = 'owner@test.com' LIMIT 1), 'Pending Bathhouse', 'Waiting for approval', 'Future Street 1', 3, 55.0415, 82.9346, 250000, 1, 8, false, true, false, 0, 'pending');

-- Create test bookings (pending confirmation)
INSERT INTO bookings (user_id, bathhouse_id, start_time, end_time, guest_count, total_price, status) VALUES
  ((SELECT id FROM users WHERE email = 'client@test.com' LIMIT 1),
   (SELECT id FROM bathhouses WHERE name = 'Premium Bathhouse' LIMIT 1),
   now() + interval '1 day',
   now() + interval '1 day 2 hours',
   4,
   1000000,
   'pending');

-- Create test review
INSERT INTO reviews (user_id, bathhouse_id, booking_id, rating, text, status) VALUES
  ((SELECT id FROM users WHERE email = 'client@test.com' LIMIT 1),
   (SELECT id FROM bathhouses WHERE name = 'Premium Bathhouse' LIMIT 1),
   (SELECT id FROM bookings LIMIT 1),
   5,
   'Great bathhouse!',
   'approved');

-- Create test representative
INSERT INTO representatives (user_id, bathhouse_id, owner_id) VALUES
  ((SELECT id FROM users WHERE email = 'representative@test.com' LIMIT 1),
   (SELECT id FROM bathhouses WHERE name = 'Premium Bathhouse' LIMIT 1),
   (SELECT id FROM users WHERE email = 'owner@test.com' LIMIT 1));

-- Create test favorite
INSERT INTO favorites (user_id, bathhouse_id) VALUES
  ((SELECT id FROM users WHERE email = 'client@test.com' LIMIT 1),
   (SELECT id FROM bathhouses WHERE name = 'Premium Bathhouse' LIMIT 1));
EOF

# Execute test data SQL
if [ -z "$POSTGRES_PASSWORD" ]; then
  psql -h "$POSTGRES_HOST" -p "$POSTGRES_PORT" -U "$POSTGRES_USER" -d "$POSTGRES_DB" -f "$TEST_DATA_FILE"
else
  PGPASSWORD="$POSTGRES_PASSWORD" psql -h "$POSTGRES_HOST" -p "$POSTGRES_PORT" -U "$POSTGRES_USER" -d "$POSTGRES_DB" -f "$TEST_DATA_FILE"
fi

# Clean up temporary file
rm -f "$TEST_DATA_FILE"

log_info "Test database setup completed successfully!"
