# Travel & Hospitality Platform - Architecture

## Overview
Offline-first modular monolith for travel group itinerary management, procurement, booking, and settlement.

## Stack
- **Frontend**: React + TypeScript + Vite + Tailwind CSS + TanStack Query + Zustand
- **Backend**: Go 1.22 + Echo + pgx
- **Database**: PostgreSQL 15+
- **Infrastructure**: Docker Compose (local dev)

## Module Structure (Backend)
Each domain module follows: handler → service → repository → model pattern.

## Key Design Decisions
1. **Offline-first**: No external service dependencies. All data local.
2. **Double-entry ledger**: Every financial movement creates balanced journal entries.
3. **Idempotent checkout**: Idempotency keys prevent duplicate charges.
4. **Deterministic pricing**: Coupon engine with explicit precedence and snapshot persistence.
5. **Immutable audit**: Send logs, journal entries, and audit logs are append-only.
6. **AES-256 file encryption**: Critical attachments encrypted at rest with wrapped DEKs.
7. **Time-limited download tokens**: File access through expiring DB-backed tokens.

## Roles
Traveler, Group Organizer, Supplier, CourierRunner, Accountant, Administrator

## API Versioning
All endpoints under `/api/v1/`.
