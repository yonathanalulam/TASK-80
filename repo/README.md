# TrailForge - Travel & Hospitality Procurement and Booking Operations Platform

Offline-first modular monolith for travel group itinerary management, procurement, booking, and settlement operations.

## Architecture

- **Frontend**: React 18 + TypeScript + Vite + Tailwind CSS v3 + TanStack Query + Zustand
- **Backend**: Go 1.22 + Echo v4 + pgx/v5 (PostgreSQL driver)
- **Database**: PostgreSQL 16
- **Infrastructure**: Docker Compose (single-command deployment)

### Key Design Decisions

| Decision | Rationale |
|----------|-----------|
| Offline-first | No external service dependencies; all data, files, notifications local |
| Double-entry ledger | Every financial movement creates balanced journal entries |
| Idempotent checkout | Idempotency keys prevent duplicate charges |
| Deterministic pricing | Coupon engine with explicit precedence, snapshot persistence |
| Immutable audit | Send logs, journal entries, audit logs are append-only |
| AES-256-GCM encryption | Critical attachments encrypted at rest with wrapped DEKs |
| Time-limited download tokens | File access via expiring DB-backed tokens |
| Risk engine | Heuristic-based throttling integrated into transactional actions |

### Offline-Only Constraints

- No SMS, email, payment gateways, cloud queues, cloud storage, or external auth
- Third-party tenders are manual recorded instruments only (cash, card-on-file, bank memo)
- All notifications, retry jobs, callback queues, and PDFs persist locally

## Repository Structure

```
travel-platform/
  apps/
    api/                    # Go Echo backend
      cmd/server/           # Entry point
      internal/
        app/                # Wiring, router
        auth/               # JWT auth
        common/             # Shared types, roles, errors, responses
        config/             # Env-based config
        db/                 # Connection, migrations, seeding
        middleware/          # Request ID, logging, auth, RBAC, audit, recovery
        modules/
          admin/            # Admin dashboard, audit logs
          bookings/         # Booking CRUD, checkout, escrow
          contracts/        # Contract/invoice generation (PDF)
          files/            # Encrypted file vault, download tokens
          finance/          # Wallets, ledger, refunds, withdrawals
          itineraries/      # Itinerary CRUD, checkpoints, member forms
          notifications/    # Messaging, DND, callback queue, send logs
          pricing/          # Coupon engine, stacking rules, idempotency
          procurement/      # RFQ, PO, delivery, inspection, exceptions
          reviews/          # Ratings, credit tiers, violations
          risk/             # Risk engine, throttling, blacklists
          users/            # User profiles, preferences
        worker/             # Background jobs
    web/                    # React frontend
      src/
        components/         # Layout, ProtectedRoute
        lib/                # API client, auth store, query client
        pages/              # All page components
  packages/
    shared-types/           # TypeScript DTOs and enums
    docs/                   # Architecture docs
  infra/
    sql/
      migrations/           # Numbered SQL migrations (001-012)
      seeds/                # Development seed data
    dev/                    # Legacy dev docker-compose
  scripts/                  # Dev helper scripts
  docker-compose.yml        # Production-ready single-command deployment
```

## Prerequisites

- Docker and Docker Compose (v2+)
- For local development without Docker:
  - Go 1.22+
  - Node.js 18+
  - PostgreSQL 16+

## Quick Start (One Command)

```bash
docker compose up --build
```

This starts:
1. **PostgreSQL** on port 5432 (healthchecked)
2. **API server** on port 8080 (auto-runs migrations + seeds)
3. **Web frontend** on port 3000 (nginx, proxies /api to backend)

Access the app at **http://localhost:3000**

### Seed Users (password: `password123`)

| Email | Role | Purpose |
|-------|------|---------|
| admin@travel.local | Administrator | Full access |
| organizer1@travel.local | Group Organizer | Create itineraries, bookings |
| traveler1@travel.local | Traveler | View itineraries, submit forms |
| supplier1@travel.local | Supplier | Quote on RFQs, accept POs |
| courier1@travel.local | Courier Runner | Deliveries, withdrawals |
| accountant@travel.local | Accountant | Finance, invoices, reconciliation |

## Environment Configuration

The API reads configuration from environment variables. See `apps/api/.env.example`:

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `PORT` | No | 8080 | API server port |
| `DB_HOST` | No | localhost | PostgreSQL host |
| `DB_PORT` | No | 5432 | PostgreSQL port |
| `DB_USER` | No | travel | PostgreSQL user |
| `DB_PASSWORD` | Yes | - | PostgreSQL password |
| `DB_NAME` | No | travel_platform | Database name |
| `DB_SSLMODE` | No | disable | SSL mode |
| `DATABASE_URL` | No* | - | Full connection URL (alternative to components) |
| `JWT_SECRET` | Yes | - | JWT signing secret |
| `MASTER_ENCRYPTION_KEY` | Yes | - | 64-char hex key for file encryption |
| `FILE_VAULT_PATH` | No | ./vault | Local file storage path |
| `DOWNLOAD_TOKEN_TTL` | No | 15m | Download token lifetime |
| `LOG_LEVEL` | No | info | debug/info/warn/error |
| `MIGRATIONS_DIR` | No | infra/sql/migrations | Path to migration files |
| `SEED_FILE` | No | - | Path to seed SQL file (runs once) |

## Local Development (Without Docker)

```bash
# Start PostgreSQL
docker compose up postgres -d

# Copy and edit env
cp apps/api/.env.example apps/api/.env
# Edit .env with your DB_PASSWORD, JWT_SECRET, MASTER_ENCRYPTION_KEY

# Run API (auto-migrates)
cd apps/api && go run ./cmd/server

# In another terminal, run frontend
cd apps/web && npm install && npm run dev
```

## Migrations

Migrations run automatically on API startup. They are tracked in the `schema_migrations` table and only run once.

Migration files are in `infra/sql/migrations/` and execute in lexicographic order.

To run migrations without starting the server:

```bash
cd apps/api && go run ./cmd/server --migrate-only
# Or via dev script:
./scripts/dev.sh migrate
```

## Running Tests

```bash
# Backend tests
cd apps/api && go test ./...

# Frontend type checking
cd apps/web && npx tsc --noEmit
```

## API Versioning

All endpoints are under `/api/v1/`. See the route registration in `apps/api/internal/app/router.go` for the complete route map.

## Roles and Permissions

Six roles with RBAC enforced at both middleware and service levels:

- **Administrator**: Full access to all resources
- **Group Organizer**: Create/manage itineraries, bookings, procurement
- **Traveler**: View itineraries, submit member forms, rate
- **Supplier**: Quote on RFQs, accept POs, manage deliveries
- **Courier Runner**: Handle deliveries, request withdrawals
- **Accountant**: Finance operations, invoices, reconciliation

Authorization is enforced at two levels:
1. **Role-based middleware** (`RequireRole`) on routes
2. **Object-level checks** in service layer (ownership, membership, invitation)

## Canonical Role Model

All role checks use centralized constants from `apps/api/internal/common/roles.go`:

| Constant | Value | Used By |
|----------|-------|---------|
| `RoleAdministrator` | `administrator` | Admin routes, risk, file policies, global access |
| `RoleGroupOrganizer` | `group_organizer` | Itinerary/booking/procurement creation |
| `RoleTraveler` | `traveler` | Member form submission, itinerary viewing |
| `RoleSupplier` | `supplier` | RFQ quoting, PO acceptance |
| `RoleCourierRunner` | `courier_runner` | Delivery handling, withdrawals |
| `RoleAccountant` | `accountant` | Finance, invoices, reconciliation |

Frontend role checks mirror these exact strings. No abbreviated forms (`admin`, `organizer`) are used.

## Risk Enforcement Strategy

The risk engine (`internal/modules/risk/`) evaluates sensitive actions before execution. It is integrated at the service layer across multiple domains:

| Action | Constant | Checked In |
|--------|----------|------------|
| Create RFQ | `RiskActionCreateRFQ` | Procurement service |
| Issue RFQ | `RiskActionIssueRFQ` | Procurement service |
| Select Supplier | `RiskActionSelectSupplier` | Procurement service |
| Cancel Booking | `RiskActionCancelBooking` | Bookings service |
| Process Refund | `RiskActionProcessRefund` | Finance service |
| Request Withdrawal | `RiskActionRequestWithdrawal` | Finance service |

Outcomes: `allow`, `throttle` (deny + create throttle record), `require_approval` (deny + create admin approval), `block` (blacklisted user denied).

Heuristic triggers: >8 cancellations/24h, 20 RFQs/10min, >=3 harassment flags.

## Data Sources

### Procurement Quotes
Supplier-facing "My Quotes" view uses `GET /api/v1/supplier-quotes`, which returns `SupplierQuoteView` objects joined from `rfq_quotes` and `rfqs` tables. This gives suppliers their submitted quotes with RFQ context (title, status, deadline).

### Wallet Escrows
Active escrow data in the wallet view uses `GET /api/v1/escrows/:ownerId`, which queries `escrow_accounts` joined with `bookings`/`purchase_orders` to find escrows for the user's orders. Returns `EscrowSummary` objects with held/released amounts and status.

## Security

- JWT authentication with bcrypt password hashing
- Role-based access control with object-level authorization
- AES-256-GCM file encryption at rest
- Time-limited, single-use download tokens
- Sensitive fields masked in logs (tokens, passwords, API keys)
- Parameterized SQL queries (no string concatenation)
- Immutable audit trail for all state-changing operations
- Risk engine with automatic throttling and blacklisting across procurement, finance, and booking domains
