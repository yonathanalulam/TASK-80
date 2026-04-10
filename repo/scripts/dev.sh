#!/bin/bash

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(dirname "$SCRIPT_DIR")"

case "$1" in
  db:up)
    echo "Starting PostgreSQL..."
    docker compose -f "$ROOT_DIR/infra/dev/docker-compose.yml" up -d
    echo "Waiting for PostgreSQL to be ready..."
    sleep 3
    echo "PostgreSQL is ready on port 5432"
    ;;
  db:down)
    echo "Stopping PostgreSQL..."
    docker compose -f "$ROOT_DIR/infra/dev/docker-compose.yml" down
    ;;
  db:reset)
    echo "Resetting database..."
    docker compose -f "$ROOT_DIR/infra/dev/docker-compose.yml" down -v
    docker compose -f "$ROOT_DIR/infra/dev/docker-compose.yml" up -d
    sleep 3
    echo "Database reset complete"
    ;;
  api:run)
    echo "Starting API server..."
    cd "$ROOT_DIR/apps/api"
    export PATH="/c/Go/bin:$PATH"
    go run ./cmd/server
    ;;
  web:run)
    echo "Starting web dev server..."
    cd "$ROOT_DIR/apps/web"
    npm run dev
    ;;
  migrate)
    echo "Running migrations..."
    cd "$ROOT_DIR/apps/api"
    export PATH="/c/Go/bin:$PATH"
    go run ./cmd/server --migrate-only
    ;;
  seed)
    echo "Running seed data..."
    PGPASSWORD=travel_dev_pass psql -h localhost -U travel -d travel_platform -f "$ROOT_DIR/infra/sql/seeds/seed.sql"
    ;;
  *)
    echo "Usage: $0 {db:up|db:down|db:reset|api:run|web:run|migrate|seed}"
    exit 1
    ;;
esac
