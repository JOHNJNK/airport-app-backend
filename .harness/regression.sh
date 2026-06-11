#!/bin/bash
# .harness/regression.sh — Full Regression + AT (Automation Testing)
# Run by the harness after every change (bug or feature).
#
# Covers:
#   1. Build check
#   2. Unit tests (mocked, no DB)
#   3. Integration tests (real Postgres via docker-compose.test.yml)
#   4. AT — starts the app and tests HTTP endpoints end-to-end

set -e
REPO_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$REPO_DIR"

GO=${GO:-$(which go 2>/dev/null || echo go)}
APP_PORT=8080
APP_PID=""

# Find docker — check common Mac/Linux paths since PATH may not be set in sandbox
DOCKER_BIN=""
for d in docker /usr/local/bin/docker /opt/homebrew/bin/docker /usr/bin/docker; do
    if command -v "$d" &>/dev/null 2>&1 || [ -x "$d" ]; then
        DOCKER_BIN="$d"
        break
    fi
done

# Support both docker compose (plugin) and docker-compose (standalone)
DOCKER_COMPOSE=""
if [ -n "$DOCKER_BIN" ]; then
    if "$DOCKER_BIN" compose version &>/dev/null 2>&1; then
        DOCKER_COMPOSE="$DOCKER_BIN compose"
    elif command -v docker-compose &>/dev/null 2>&1; then
        DOCKER_COMPOSE="docker-compose"
    fi
fi

cleanup() {
    [ -n "$APP_PID" ] && kill "$APP_PID" 2>/dev/null || true
    [ -n "$DOCKER_COMPOSE" ] && $DOCKER_COMPOSE -f docker-compose.test.yml down -v -q 2>/dev/null || true
}
trap cleanup EXIT

echo "=== Regression + AT Suite ==="
echo "Go binary: $GO"
echo ""

# ── Step 1: Build check ───────────────────────────────────────────────────
echo "→ Build check..."
$GO build ./controllers/... ./middleware/... ./models/... ./repositories/...
echo "  Build OK ✓"
echo ""

# ── Step 2: Unit tests (no DB needed) ────────────────────────────────────
echo "→ Unit tests (mocked)..."
$GO test ./controllers/... ./middleware/... -count=1
echo "  Unit tests OK ✓"
echo ""

# ── Step 3: DB-dependent tests (if docker available) ─────────────────────
if [ -z "$DOCKER_COMPOSE" ]; then
    echo "→ Docker not available — skipping integration tests and AT"
    echo "  Install Docker to run the full regression suite"
    echo ""
    echo "=== Regression PASSED (build + unit only — no Docker) ==="
    exit 0
fi

echo "→ Starting test database (docker-compose.test.yml)..."
$DOCKER_COMPOSE -f docker-compose.test.yml up -d
echo "  Waiting for Postgres..."
RETRIES=30
until $DOCKER_COMPOSE -f docker-compose.test.yml ps 2>/dev/null | grep -q "healthy"; do
    RETRIES=$((RETRIES - 1))
    [ $RETRIES -le 0 ] && echo "  Postgres timed out" && exit 1
    sleep 2
done
echo "  Postgres ready ✓"
echo ""

[ -f .env.test ] && export $(grep -v '^#' .env.test | xargs)

# ── Step 4: Integration tests (real DB) ──────────────────────────────────
echo "→ Integration tests (real Postgres)..."
$GO test ./repositories/... -count=1
echo "  Integration tests OK ✓"
echo ""

# ── Step 5: Build and start app ───────────────────────────────────────────
echo "→ Starting application on :$APP_PORT..."
$GO build -o /tmp/airport-regression-app . 2>&1
/tmp/airport-regression-app &
APP_PID=$!

RETRIES=20
until curl -sf "http://localhost:$APP_PORT/health" > /dev/null 2>&1; do
    RETRIES=$((RETRIES - 1))
    [ $RETRIES -le 0 ] && echo "  App failed to start" && exit 1
    sleep 1
done
echo "  Application ready ✓"
echo ""

# ── Step 6: AT — HTTP endpoint assertions ────────────────────────────────
echo "→ Running AT (HTTP endpoint assertions)..."
PASS=0; FAIL=0

check() {
    local desc="$1" expected="$2" url="$3"
    local actual; actual=$(curl -s -o /dev/null -w "%{http_code}" "$url" 2>/dev/null)
    if [ "$actual" = "$expected" ]; then
        echo "  ✓ $desc ($actual)"
        PASS=$((PASS+1))
    else
        echo "  ✗ $desc — expected $expected got $actual"
        FAIL=$((FAIL+1))
    fi
}

BASE="http://localhost:$APP_PORT"

check "GET /health"                    "200" "$BASE/health"
check "GET /airlines"                  "200" "$BASE/airlines"
check "GET /airlines?page=0"           "200" "$BASE/airlines?page=0"
check "GET /airlines?page=-1 → 400"   "400" "$BASE/airlines?page=-1"
check "GET /airlines?page=abc → 400"  "400" "$BASE/airlines?page=abc"
check "GET /airline/nonexistent → 400" "400" "$BASE/airline/nonexistent"
check "GET /gates"                     "200" "$BASE/gates"
check "GET /aircrafts"                 "200" "$BASE/aircrafts"
check "GET /slots"                     "200" "$BASE/slots"

echo ""
echo "AT: $PASS passed, $FAIL failed"

[ $FAIL -gt 0 ] && echo "=== Regression FAILED ===" && exit 1
echo "=== Regression PASSED (build + unit + integration + AT) ==="
