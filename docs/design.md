# Design Document

## Travel & Hospitality Procurement and Booking Operations Platform

---

## 1. Architecture Overview

The platform follows a **decoupled frontend/backend architecture**:

- **Frontend:** React 18 SPA (single-page application) built with Vite, served as static assets
- **Backend:** Go API server using the Echo v4 framework, exposing REST endpoints under `/api/v1/`
- **Database:** PostgreSQL with `uuid-ossp` and `pgcrypto` extensions for UUID generation and cryptographic functions
- **Shared Types:** A `packages/shared-types` TypeScript package defines DTOs and enums shared across the frontend

```
repo/
├── apps/
│   ├── api/                    # Go backend (Echo)
│   │   ├── cmd/server/         # Entry point (main.go)
│   │   └── internal/
│   │       ├── app/            # App bootstrap, router setup
│   │       ├── auth/           # Authentication handler & service
│   │       ├── common/         # Shared types (response envelope, roles, errors, context)
│   │       ├── config/         # Environment-based configuration
│   │       ├── db/             # Database connection, migrations, seed
│   │       ├── middleware/     # Auth JWT, RBAC, audit logging, recovery, request ID, logger
│   │       ├── modules/        # Domain modules (see below)
│   │       └── worker/         # Background job worker
│   └── web/                    # React frontend (Vite + TypeScript)
│       └── src/
│           ├── components/     # Layout, ProtectedRoute
│           ├── lib/            # API client (Axios), auth store (Zustand), query config
│           ├── pages/          # Page components organized by feature
│           └── __tests__/      # API contract tests
├── infra/
│   └── sql/
│       ├── migrations/         # 12 sequential SQL migrations (001–012)
│       └── seeds/              # Seed data for development
└── packages/
    └── shared-types/           # Shared TypeScript DTOs and enums
```

---

## 2. Backend Module Design

The backend is organized into **domain modules**, each following a consistent layered pattern:

```
modules/<domain>/
├── model.go        # Database models and domain constants
├── dto.go          # Request/response DTOs with JSON tags
├── repository.go   # Database access layer (raw SQL via pgx)
├── service.go      # Business logic orchestration
└── handler.go      # HTTP handler (Echo route registration, request binding, response writing)
```

### Domain Modules

| Module | Responsibility |
|---|---|
| `users` | User profiles, preferences, CRUD |
| `itineraries` | Itinerary lifecycle, checkpoints, members, form definitions/submissions, change events |
| `bookings` | Booking creation, checkout with pricing, tender recording, cancellation, completion |
| `pricing` | Coupon management, pricing engine, stacking rules, idempotency, snapshots |
| `finance` | Wallets, transactions, escrow, refunds, withdrawals, double-entry ledger, reconciliation |
| `contracts` | Contract templates, PDF generation, invoice requests, invoice generation |
| `files` | File upload with AES-256-GCM encryption, download tokens, role-based access |
| `notifications` | Notification events, messages, DND settings, subscriptions, callback queue, send logs |
| `procurement` | RFQ workflow, PO management, deliveries, inspections, discrepancies, exceptions |
| `reviews` | Multi-dimensional reviews, credit tiers, violations, no-shows, harassment flags |
| `risk` | Risk scoring engine, throttle actions, admin approvals, blacklist management |
| `admin` | Audit logs, send logs, system configuration |

---

## 3. Authentication & Authorization

### Authentication Flow

1. User submits email/password to `POST /api/v1/auth/login`
2. Backend verifies credentials against `bcrypt` password hashes in the `users` table
3. A JWT token is issued and a session record is created in the `sessions` table
4. The frontend stores the token in localStorage and attaches it as `Authorization: Bearer <token>` on every request
5. The `JWTAuth` middleware validates the token, extracts user ID, email, and roles into the Echo context
6. Public paths (health checks, login, file downloads with token) bypass JWT validation

### Role-Based Access Control (RBAC)

Six predefined roles with a fine-grained permission system (37 permissions across 9+ resources):

| Role | Key Capabilities |
|---|---|
| **Traveler** | View itineraries/bookings, submit forms, write reviews, request invoices, view wallet |
| **Group Organizer** | Full itinerary management, create bookings, initiate RFQs, procurement read access |
| **Supplier** | Submit quotes on RFQs, accept POs, receive deliveries, manage reviews, view wallet |
| **Courier/Runner** | Receive deliveries, manage files, write reviews, request withdrawals |
| **Accountant** | Record tenders, approve refunds/withdrawals, reconciliation, financial operations |
| **Administrator** | All permissions — user management, audit logs, risk management, system config |

RBAC is enforced at two levels:
- **Middleware level:** `RequireRole(...)` middleware on route groups checks role membership
- **Service level:** Policy functions (e.g., `CanManageItinerary`, `CanViewItinerary`) enforce ownership and membership rules

---

## 4. Database Design

### Migration Strategy

12 sequential SQL migrations applied in order at startup via the `db/migrate.go` module. Each migration is idempotent and uses `IF NOT EXISTS` / `CREATE TYPE ... AS ENUM` patterns.

### Key Design Decisions

- **UUIDs as primary keys** — generated via `uuid_generate_v4()` for all entities
- **JSONB for flexible data** — used for form submissions, eligibility rules, notification payloads, change event diffs, risk metadata, and user preferences
- **Enum types** — PostgreSQL enums for booking status, wallet types, tender types, RFQ/PO status, inspection status, discrepancy types, exception status, credit tiers, and review dimensions
- **Soft deletes** — `deleted_at` timestamps on users; most other entities use status-based lifecycle management
- **Unique constraints** — enforced on coupon redemptions `(coupon_id, user_id, redemption_scope_key)`, form submissions `(itinerary_id, member_user_id)`, reviews `(reviewer_id, order_type, order_id)`, and idempotency keys `(actor_id, route, key)`
- **Immutable audit tables** — `audit_logs`, `journal_entries`, `journal_lines`, `send_logs`, and `domain_events` are append-only

### Schema Overview

```
Users & Auth:       users, roles, permissions, role_permissions, user_roles, 
                    user_profiles, user_preferences, sessions,
                    do_not_disturb_settings, subscription_preferences

Itineraries:        itineraries, itinerary_checkpoints, itinerary_members,
                    itinerary_member_form_definitions, itinerary_member_form_submissions,
                    itinerary_change_events, itinerary_notifications

Bookings:           bookings, booking_items, checkout_pricing_snapshots

Coupons & Pricing:  coupons, coupon_packs, coupon_pack_items, coupon_redemptions,
                    pricing_memberships, idempotency_keys

Finance:            wallets, wallet_transactions, escrow_accounts, payment_records,
                    refunds, refund_items, withdrawal_requests, withdrawal_disbursements,
                    journal_entries, journal_lines, counterparties,
                    reconciliation_runs, reconciliation_items

Notifications:      message_templates, messages, notification_events,
                    notification_recipients, callback_queue_entries, send_logs,
                    message_read_receipts

Files & Contracts:  file_metadata, file_record_links, file_access_policies,
                    download_tokens, contract_templates, generated_contracts,
                    invoice_requests, invoices

Procurement:        rfqs, rfq_items, rfq_suppliers, rfq_quotes, rfq_quote_items,
                    purchase_orders, po_items, deliveries, delivery_items,
                    quality_inspections, discrepancy_tickets, exception_cases,
                    waiver_records, settlement_adjustments

Reviews & Risk:     reviews, review_dimensions, review_scores, credit_tiers,
                    user_credit_snapshots, violation_records, harassment_flags,
                    no_show_records, blacklist_records, risk_events, risk_scores,
                    throttle_actions, admin_approvals

Audit:              audit_logs, entity_versions, domain_events
```

---

## 5. Pricing & Coupon Engine

### Coupon Types

| Type | Behavior |
|---|---|
| `threshold_fixed` | Fixed dollar discount (e.g., "$25 off $200") when subtotal meets `min_spend` |
| `percentage` | Percentage discount on current total (e.g., "10% off lodging") |
| `new_user_gift` | Fixed dollar gift for new users, valid for 14 days from account creation |

### Stacking Rules

- At most **1 threshold_fixed** coupon per checkout (highest discount wins)
- At most **1 percentage** coupon per checkout (highest discount wins)
- `new_user_gift` is **exclusive** — blocks both threshold and percentage coupons
- Threshold and percentage coupons can stack with each other (unless new_user_gift applies)
- Member tier discounts (Silver 3%, Gold 5%, Platinum 8%) are applied first and stack with coupons

### Eligibility Evaluation

The engine evaluates each coupon against these checks in order:
1. Active status
2. Date range validity (`valid_from` / `valid_to`)
3. Minimum spend threshold
4. Category restrictions (items must match eligible categories)
5. New user flag
6. Membership tier requirement
7. Platform-wide usage limit
8. Per-user usage limit
9. Already redeemed in same scope (idempotency)

Ineligible coupons return a human-readable reason with a machine-readable reason code (e.g., `REASON_MIN_SPEND_NOT_MET`, `REASON_CATEGORY_RESTRICTED`, `REASON_STACKING_NOT_ALLOWED`).

### Idempotency

Checkout requests require an `Idempotency-Key` header. The system:
1. Checks for an existing response matching `(actor_id, route, key)`
2. Validates the request hash to detect payload mutation
3. Locks the key with `ON CONFLICT DO NOTHING` for concurrency safety
4. Caches the response on completion
5. Expires keys after 24 hours (background worker cleanup every 15 min)

---

## 6. Finance & Settlement

### Wallet System

Seven wallet types supporting the full funds lifecycle:

| Wallet Type | Purpose |
|---|---|
| `customer` | End-user balance |
| `supplier` | Supplier receivables |
| `courier` | Courier earnings |
| `platform_clearing` | Incoming payment clearing |
| `escrow_control` | Funds held pending completion |
| `refund_clearing` | Refund processing |
| `fee_revenue` | Platform fee collection |

### Double-Entry Ledger

Every financial movement is recorded as a balanced journal entry:

| Operation | Debit Account | Credit Account |
|---|---|---|
| Record tender | ManualTenderClearing (1100) | CashOnHand (1000) |
| Release escrow | EscrowLiability (2000) | SupplierPayable (2100) |
| Process refund | EscrowLiability (2000) | CashOnHand (1000) |
| Approve withdrawal | CourierPayable (2200) | CashOnHand (1000) |
| Settlement adjustment | SettlementAdjustmentReserve (2500) | AdjustmentExpense (5000) |

### Constraints

- **Minimum refund amount:** $1.00
- **Courier daily withdrawal cap:** $2,500.00 per courier per day (checked at both request and approval time)
- **Escrow release validation:** release amount must not exceed remaining held amount
- **Manual instruments only:** cash, card-on-file (recorded), bank transfer (recorded), other manual — no network authorization

---

## 7. Procurement Workflow

The procurement module implements a full RFQ-to-PO-to-delivery lifecycle:

```
RFQ: draft → issued → responded → comparison_ready → selected → converted_to_po
                                                    → closed_no_award

PO:  draft → issued → accepted → partially_delivered → delivered 
     → inspection_pending → exception_open → closed

Exception: open → pending_financial_resolution → ready_to_close → closed
                → pending_waiver → ready_to_close → closed
```

### Key Workflow Steps

1. **RFQ Creation** — Organizer/accountant creates RFQ with items and deadline (risk-checked)
2. **RFQ Issue** — Items finalized, suppliers invited, status → `issued`
3. **Quote Submission** — Suppliers submit total amount, lead time, and per-item pricing
4. **Quote Comparison** — Matrix view of all supplier quotes against RFQ items
5. **Supplier Selection** — Best quote selected, triggers PO creation path
6. **PO Conversion** — RFQ quote converts to PO with auto-generated PO number (`PO-YYYY-XXXX`)
7. **PO Acceptance** — Supplier accepts terms
8. **Delivery Recording** — Items received with quantity tracking (delivered/accepted/rejected)
9. **Quality Inspection** — Pass/fail with required notes; failure auto-creates discrepancy ticket
10. **Discrepancy Handling** — Types: shortage, damage, wrong_item, late_delivery, service_deviation, other
11. **Exception Resolution** — Requires at least one waiver or settlement adjustment before closure

---

## 8. Risk Control Engine

### Automated Threat Detection

| Trigger | Threshold | Action | Duration |
|---|---|---|---|
| Excessive cancellations | >8 in 24 hours | Throttle + admin approval | 6 hours |
| RFQ spam | ≥20 in 10 minutes | Throttle | 1 hour |
| Harassment flags | ≥3 open flags | Freeze all actions | 24 hours |
| Blacklisted account | Active blacklist record | Block all actions | Until lifted |

### Risk Scoring

A background worker recomputes risk scores every 10 minutes for users with recent events. Severity weights: critical=10, high=5, medium=2, low=1. Scores are persisted in `risk_scores` with `factors_json` for transparency.

### Event Severity Classification

- `harassment_flag` → high
- `cancellation` → medium
- `rfq_creation` → low
- Other events → low

### Admin Escalation

When a throttled action requires approval, an `admin_approvals` record is created. Administrators can view pending approvals via `GET /api/v1/admin/approvals` and resolve them with approval or denial plus notes.

---

## 9. File Management & Encryption

### Upload Flow

1. Client uploads file via multipart form to `POST /api/v1/files/upload`
2. If `encrypt=true`, a per-file DEK is generated (32 bytes via `crypto/rand`)
3. File content is encrypted with AES-256-GCM (random nonce prepended)
4. DEK is wrapped under the master key (from `MASTER_ENCRYPTION_KEY` env var)
5. Metadata stored: original filename, MIME type, byte size, SHA-256 hash, encryption status, wrapped DEK
6. File linked to a record via `file_record_links` (record_type + record_id)

### Download Flow

1. Authorized user requests `POST /api/v1/files/:id/download-token` with TTL and single-use flag
2. System generates a UUID token with `expires_at` timestamp
3. File is served via `GET /api/v1/files/download/:token` (public — no JWT required)
4. If encrypted, file is decrypted on-the-fly using unwrapped DEK
5. Single-use tokens are marked as used after first download
6. Background worker cleans up expired tokens every 5 minutes

---

## 10. Notification & Messaging

### Architecture

- **Notification Events** — triggered by system actions, stored with `source_type`, `source_id`, and `payload_json`
- **Notification Recipients** — per-event, per-user delivery tracking with status (pending/delivered) and optional `deferred_until`
- **Messages** — templated in-app messages with subject/body, sender/recipient
- **Send Logs** — immutable delivery audit trail for all notification events
- **Callback Queue** — entries queued for offline export with attempt tracking

### DND Compliance

- Per-user DND settings with configurable start/end times (default: 9:00 PM – 8:00 AM)
- Notifications during DND are deferred (not dropped)
- Background worker processes deferred notifications every 1 minute

### Subscription Preferences

Users can configure per-channel, per-event-type subscription preferences. Disabled subscriptions suppress delivery for that event type on that channel.

---

## 11. Review & Credit System

### Multi-Dimensional Reviews

- Reviews are unique per `(reviewer_id, order_type, order_id)`
- 9 scoring dimensions available: punctuality, communication, quality, compliance, professionalism, accuracy, cleanliness, route_adherence, delivery_integrity
- Each review has an `editable_until` window after creation
- Overall rating (1-5) plus per-dimension scores

### Credit Tiers

| Tier | Min Transactions | Min Avg Rating | Max Violations |
|---|---|---|---|
| Bronze | 0 | 0.0 | 10 |
| Silver | 5 | 3.0 | 5 |
| Gold | 15 | 3.5 | 3 |
| Platinum | 30 | 4.0 | 1 |
| Restricted | 0 | 0.0 | 0 |

### Moderation

- **Violations** — recorded by admin/accountant with severity levels
- **No-Shows** — tracked per order with user attribution
- **Harassment Flags** — any user can flag with description and optional evidence file
- **Blacklist** — admin can blacklist/unblacklist accounts, blocking all risk-gated actions

---

## 12. Frontend Design

### Technology Stack

| Library | Purpose |
|---|---|
| React 18 | UI framework |
| React Router | Client-side routing |
| TanStack React Query v5 | Server state management (5-min stale time, 1 retry) |
| Zustand | Auth state management |
| Axios | HTTP client with interceptors |
| React Hook Form + Zod | Form handling and validation |
| Tailwind CSS | Utility-first styling |

### Page Structure

| Page | Key Features |
|---|---|
| **Dashboard** | Welcome message, quick-action cards for main features |
| **Login** | Email/password with Zod validation, redirect on success |
| **Itinerary List** | Paginated grid with status badges, create link |
| **Itinerary Detail** | 5 tabs (Overview, Checkpoints, Members, Forms, Change History) |
| **Itinerary Wizard** | 7-step guided form (title, date, location, checkpoints, notes, member forms, review) |
| **Booking List** | Paginated table with status/amount columns |
| **Booking Detail** | Line items, coupon application with preview, checkout, tender recording |
| **Booking New** | Dynamic line items form with subtotal calculation |
| **Procurement Dashboard** | 4 tabs (RFQs, POs, Deliveries, Exceptions), supplier quotes section |
| **Notification Center** | Filtered notifications, mark-as-read, DND settings, callback export |
| **Wallet Dashboard** | Balance, escrows, transaction history, withdrawal/refund modals |
| **Document Center** | File upload (drag-and-drop), contract generation, invoice requests |
| **Review Dashboard** | Credit tier display, review submission with star ratings, harassment flagging |
| **Admin Overview** | System configuration display |
| **Admin Users** | User listing with status |
| **Admin Settings** | Key-value configuration display |

### API Client

- Base URL: `/api/v1`
- Request interceptor: attaches `Authorization: Bearer <token>` from localStorage
- Response interceptor: unwraps `{ success: true, data: ... }` envelope
- 401 responses: clear auth state and redirect to `/login`

---

## 13. Background Workers

Four periodic background jobs run in goroutines:

| Job | Interval | Purpose |
|---|---|---|
| Process deferred notifications | 1 minute | Deliver notifications whose DND deferral has expired |
| Clean up download tokens | 5 minutes | Delete expired and used download tokens |
| Recompute risk scores | 10 minutes | Recalculate risk scores for users with recent events |
| Clean up idempotency keys | 15 minutes | Delete expired idempotency entries (>24h old) |

---

## 14. Audit & Traceability

### Audit Logging Middleware

- Captures all mutating requests (POST, PATCH, PUT, DELETE) asynchronously
- Records: action (method + path), actor ID, request ID, IP address, response status
- Stored in immutable `audit_logs` table

### Entity Versioning

- `entity_versions` table tracks version history with `version_number`, `data_json`, `changed_by`
- Enables point-in-time reconstruction of entity state

### Domain Events

- `domain_events` table provides event sourcing: `event_type`, `aggregate_type`, `aggregate_id`, `payload_json`, `actor_id`
- All events are append-only for full traceability

### Send Logs

- Every notification delivery attempt is recorded in `send_logs`
- Tracks recipient, message ID, event type, channel, status, and payload summary
- Immutable — provides complete notification delivery audit trail
