#!/bin/bash
# .harness/test.sh — Unit test script
# Run by the agent harness at Pass 1 after every code change.
# Fast: no database needed (uses gomock for all repository calls).
# Exit 0 = all tests passed. Exit 1 = failure.
set -e

# Use GO env var if set (harness injects the correct path), else fall back to PATH
GO=${GO:-$(which go 2>/dev/null || echo go)}

echo "=== Unit Tests ==="
echo "Go binary: $GO"
echo "Package: ./controllers/... ./middleware/... ./models/..."
echo ""

$GO test ./controllers/... ./middleware/... ./models/... -v -count=1

echo ""
echo "=== Unit Tests PASSED ==="
