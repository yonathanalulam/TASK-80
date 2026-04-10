#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

cleanup() {
  echo ""
  echo "=== Cleaning up test containers ==="
  docker rm -f travel-api-tests travel-web-tests 2>/dev/null || true
  docker rmi -f travel-api-tests-img travel-web-tests-img 2>/dev/null || true
}
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
