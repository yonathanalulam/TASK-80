#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

MODE="${1:-unit}"

cleanup() {
  echo ""
  echo "=== Cleaning up test containers ==="
  docker rm -f travel-api-tests travel-web-tests 2>/dev/null || true
  docker rmi -f travel-api-tests-img travel-web-tests-img 2>/dev/null || true
}

cleanup_integration() {
  echo ""
  echo "=== Cleaning up integration test containers ==="
  docker compose -f docker-compose.test.yml down --volumes --remove-orphans 2>/dev/null || true
}

cleanup_e2e() {
  echo ""
  echo "=== Cleaning up E2E test containers ==="
  docker compose down --volumes --remove-orphans 2>/dev/null || true
  docker rm -f travel-e2e-tests 2>/dev/null || true
  docker rmi -f travel-e2e-tests-img 2>/dev/null || true
}

# ── Integration tests ──────────────────────────────────────
if [ "$MODE" = "--integration" ]; then
  trap cleanup_integration EXIT

  echo "============================================"
  echo "  Travel Platform — Integration Tests       "
  echo "============================================"
  echo ""
  echo "=== Starting test database and running API integration tests ==="
  docker compose -f docker-compose.test.yml up --build --abort-on-container-exit --exit-code-from api-integration-tests
  exit $?
fi

# ── E2E smoke tests ────────────────────────────────────────
if [ "$MODE" = "--e2e" ]; then
  trap cleanup_e2e EXIT

  echo "============================================"
  echo "  Travel Platform — E2E Smoke Tests         "
  echo "============================================"
  echo ""
  echo "=== Starting full platform ==="
  docker compose up -d --build --wait

  echo ""
  echo "=== Running E2E smoke tests ==="
  docker build -t travel-e2e-tests-img ./tests/e2e
  docker run --rm --name travel-e2e-tests \
    --network host \
    -e API_BASE=http://localhost:8080 \
    -e WEB_BASE=http://localhost:3000 \
    travel-e2e-tests-img
  E2E_EXIT=$?

  echo ""
  echo "=== Stopping platform ==="
  docker compose down --volumes

  exit $E2E_EXIT
fi

# ── Unit tests (default) ──────────────────────────────────
trap cleanup EXIT

echo "============================================"
echo "  Travel Platform — Test Runner (Docker)    "
echo "============================================"
echo ""

# ── Build & run Go API tests ────────────────────────────────
echo "=== Building Go API test image ==="
docker build -t travel-api-tests-img -f - ./apps/api <<'DOCKERFILE'
FROM golang:1.22-alpine
WORKDIR /app
RUN apk add --no-cache git
COPY go.mod go.sum ./
RUN go mod download
COPY . .
CMD ["go", "test", "-v", "-count=1", "./..."]
DOCKERFILE

echo ""
echo "=== Running Go API tests ==="
docker run --rm --name travel-api-tests travel-api-tests-img
GO_EXIT=$?

# ── Build & run Web frontend tests ─────────────────────────
echo ""
echo "=== Building Web frontend test image ==="
docker build -t travel-web-tests-img -f - ./apps/web <<'DOCKERFILE'
FROM node:18-alpine
WORKDIR /app
COPY package.json package-lock.json ./
RUN npm ci
COPY . .
CMD ["npx", "vitest", "run"]
DOCKERFILE

echo ""
echo "=== Running Web frontend tests ==="
docker run --rm --name travel-web-tests travel-web-tests-img
WEB_EXIT=$?

# ── Summary ─────────────────────────────────────────────────
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
