#!/bin/bash
# Unit test script — run by the agent harness after every code change
# Fast path: no database needed (uses gomock)
set -e

GO=${GO:-go}

echo "Running unit tests..."
$GO test ./controllers/... ./middleware/... ./models/... -v -count=1

EXIT=$?
if [ $EXIT -eq 0 ]; then
    echo "All unit tests passed ✓"
else
    echo "Unit tests FAILED ✗"
    exit $EXIT
fi
