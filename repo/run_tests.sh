#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
COMPOSE_FILE="$SCRIPT_DIR/docker-compose.test.yml"

cleanup() {
  echo ""
  echo "=== Cleaning up test containers ==="
  docker compose -f "$COMPOSE_FILE" down -v --remove-orphans 2>/dev/null || true
}
trap cleanup EXIT

echo "============================================"
echo "  Travel Platform — Test Runner (Docker)    "
echo "============================================"
echo ""

# Build and run all test services
docker compose -f "$COMPOSE_FILE" down -v --remove-orphans 2>/dev/null || true
docker compose -f "$COMPOSE_FILE" build --no-cache

echo ""
echo "=== Running Go API tests ==="
docker compose -f "$COMPOSE_FILE" run --rm api-tests
GO_EXIT=$?

echo ""
echo "=== Running Web frontend tests ==="
docker compose -f "$COMPOSE_FILE" run --rm web-tests
WEB_EXIT=$?

echo ""
echo "============================================"
if [ $GO_EXIT -eq 0 ] && [ $WEB_EXIT -eq 0 ]; then
  echo "  ALL TESTS PASSED"
else
  echo "  SOME TESTS FAILED"
  [ $GO_EXIT -ne 0 ] && echo "    - Go API tests: FAILED (exit $GO_EXIT)"
  [ $WEB_EXIT -ne 0 ] && echo "    - Web tests:    FAILED (exit $WEB_EXIT)"
fi
echo "============================================"

exit $(( GO_EXIT + WEB_EXIT ))
