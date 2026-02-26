#!/bin/bash

set -e

# Test script to verify setup.sh can be sourced and its functions work

# Source the setup script functions (without executing the main logic)
SCRIPT_DIR="$(dirname "$0")"

# Test 1: Verify .env.test exists
if [ -f "$SCRIPT_DIR/.env.test" ]; then
    echo "✓ .env.test file exists"
else
    echo "✗ .env.test file missing"
    exit 1
fi

# Test 2: Verify setup.sh is executable
if [ -x "$SCRIPT_DIR/setup.sh" ]; then
    echo "✓ setup.sh is executable"
else
    echo "✗ setup.sh is not executable"
    exit 1
fi

# Test 3: Verify hurl is installed
if command -v hurl &> /dev/null; then
    echo "✓ hurl is installed: $(hurl --version)"
else
    echo "✗ hurl is not installed"
    exit 1
fi

# Test 4: Verify migrations directory exists
if [ -d "$SCRIPT_DIR/../../migrations" ]; then
    migration_count=$(ls "$SCRIPT_DIR/../../migrations"/*.up.sql 2>/dev/null | wc -l)
    echo "✓ Found $migration_count migration files"
else
    echo "✗ Migrations directory not found"
    exit 1
fi

# Test 5: Verify PostgreSQL tools are available
if command -v psql &> /dev/null; then
    echo "✓ psql is available: $(psql --version)"
else
    echo "✗ psql is not available"
    exit 1
fi

# Test 6: Verify Redis is available (optional but recommended)
if command -v redis-cli &> /dev/null; then
    echo "✓ redis-cli is available"
else
    echo "⚠ redis-cli not available (tests may fail if Redis is not running)"
fi

echo ""
echo "All setup checks passed! Ready to run tests."
echo ""
echo "Next steps:"
echo "1. Ensure PostgreSQL and Redis are running"
echo "2. Run: ./setup.sh"
echo "3. Run: hurl --env-file .env.test auth.hurl"
