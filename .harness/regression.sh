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

# If docker binary exists but daemon not running, try to start it (E2B sandbox)
if [ -n "$DOCKER_BIN" ]; then
    if ! "$DOCKER_BIN" info &>/dev/null 2>&1; then
        echo "  Starting Docker daemon..."
        sudo service docker start 2>/dev/null || sudo dockerd &>/dev/null &
        sleep 3
        # Add current user to docker group if needed
        sudo chmod 666 /var/run/docker.sock 2>/dev/null || true
    fi
fi

# Support both docker compose (plugin) and docker-compose (standalone)
DOCKER_COMPOSE=""
if [ -n "$DOCKER_BIN" ] && "$DOCKER_BIN" info &>/dev/null 2>&1; then
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
$GO build -buildvcs=false ./controllers/... ./middleware/... ./models/... ./repositories/...
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

echo "  Waiting for Postgres to be healthy..."
RETRIES=40
until $DOCKER_COMPOSE -f docker-compose.test.yml ps 2>/dev/null | grep "healthy" | grep -v "unhealthy" | grep -q "."; do
    RETRIES=$((RETRIES - 1))
    if [ $RETRIES -le 0 ]; then
        echo "  Postgres timed out. Current state:"
        $DOCKER_COMPOSE -f docker-compose.test.yml ps 2>/dev/null
        exit 1
    fi
    sleep 3
done
echo "  Postgres ready ✓"
echo ""

# Load test env vars AND create .env file (app uses godotenv.Load() which reads .env directly)
if [ -f .env.test ]; then
    export $(grep -v '^#' .env.test | xargs) 2>/dev/null || true
    cp .env.test .env
fi

# ── Step 4: Integration tests (real DB) ──────────────────────────────────
echo "→ Integration tests (real Postgres)..."
$GO test ./repositories/... -count=1
echo "  Integration tests OK ✓"
echo ""

# ── Step 5: Build and start app ───────────────────────────────────────────
echo "→ Building application..."
# -buildvcs=false avoids git VCS errors in environments with incomplete git state (e.g. E2B)
$GO build -buildvcs=false -o /tmp/airport-regression-app . 2>&1
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

# ── SMOKE TESTS — existing functionality that must never break ────────────
# These pass regardless of which bug is being fixed.
# Do NOT add assertions for known unfixed bugs here.
check "GET /health"                     "200" "$BASE/health"
check "GET /airlines"                   "200" "$BASE/airlines"
check "GET /airlines?page=0"            "200" "$BASE/airlines?page=0"
check "GET /airline/nonexistent → 400"  "400" "$BASE/airline/nonexistent"
check "GET /gates"                      "200" "$BASE/gates"
check "GET /gate/nonexistent → 400"     "400" "$BASE/gate/nonexistent"
check "GET /aircrafts"                  "200" "$BASE/aircrafts"
check "GET /slots"                      "200" "$BASE/slots"

# ── CONFIRMED-FIXED BUG CHECKS — added only when bug is merged to master ─
# Rule: only add an assertion here AFTER the fix is confirmed working and merged.
# Until then, track bugs in GitHub Issues, not here.
#
# Confirmed fixed (add assertions below as bugs are merged):
# check "GET /airlines?page=-1 → 400"   "400" "$BASE/airlines?page=-1"   # add when #XX merged
# check "GET /airlines?page=abc → 400"  "400" "$BASE/airlines?page=abc"  # add when #XX merged
# check "GET /gates?page=-1 → 400"      "400" "$BASE/gates?page=-1"      # add when #XX merged
# check "GET /gates?page=abc → 400"     "400" "$BASE/gates?page=abc"     # add when #XX merged

echo ""
echo "AT: $PASS passed, $FAIL failed"

[ $FAIL -gt 0 ] && echo "=== Regression FAILED ===" && exit 1
echo "=== Regression PASSED (build + unit + integration + AT) ==="
