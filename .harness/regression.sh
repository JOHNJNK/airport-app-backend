#!/bin/bash
# .harness/regression.sh — Regression suite
# Run by the agent harness at Pass 2 after dev loop completes.
# Catches: broken builds, broken unit tests, broken integration (if DB available).
# Runs for BOTH bug fixes and feature additions.
# Exit 0 = nothing broken. Exit 1 = regression detected.
set -e

GO=${GO:-$(which go 2>/dev/null || echo go)}

echo "=== Regression Suite ==="
echo "Go binary: $GO"
echo ""

# Step 1: Build check — catch compilation errors across all non-server packages
# (server/ imports docs/ which needs a database, skip it)
echo "→ Build check (controllers, middleware, models, repositories)..."
$GO build ./controllers/... ./middleware/... ./models/... ./repositories/...
echo "  Build OK ✓"
echo ""

# Step 2: Unit tests — fast, no DB needed
echo "→ Unit tests (controllers, middleware)..."
$GO test ./controllers/... ./middleware/... -count=1
echo "  Unit tests OK ✓"
echo ""

# Step 3: Integration tests — only if database is reachable
echo "→ Checking database availability..."
if command -v pg_isready &>/dev/null && \
   pg_isready -h "${AIRPORT_HOST:-localhost}" \
              -U "${AIRPORT_POSTGRES_USER:-test}" \
              -d "${AIRPORT_DB_NAME:-airport_test}" \
              -q 2>/dev/null; then
    echo "  Database available — running integration tests..."
    $GO test ./repositories/... -count=1
    echo "  Integration tests OK ✓"
else
    echo "  Database not available — skipping integration tests (not a failure)"
fi

echo ""
echo "=== Regression Suite PASSED ==="
