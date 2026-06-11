#!/bin/bash
# Regression test script — run after dev loop completes (both bug and feature mode)
# Catches: broken builds, broken unit tests, broken integration if DB is available
set -e

GO=${GO:-go}

echo "=== Regression Suite ==="

# 1. Build check — catch compilation errors across all packages
echo "→ Build check..."
$GO build ./controllers/... ./middleware/... ./models/... ./repositories/...

# 2. Unit tests — fast, no DB needed
echo "→ Unit tests..."
$GO test ./controllers/... ./middleware/... ./models/... -count=1

# 3. Integration tests — only if database is available
echo "→ Checking database availability..."
if command -v pg_isready &>/dev/null && pg_isready -h localhost -U test -d airport_test -q 2>/dev/null; then
    echo "→ Database available — running integration tests..."
    $GO test ./repositories/... -count=1
else
    echo "→ Database not available — skipping integration tests"
fi

echo "=== Regression Suite PASSED ==="
