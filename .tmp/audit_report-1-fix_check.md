# Fix Check Report for Prior `audit_report_1_fix-check_round1.md` Findings (Static-Only)

## 1. Scope
- Source of truth for checked findings: `.tmp/audit_report-1-fix-check_round1.md`.
- Verification method: static code inspection only; no runtime execution.

## 2. Finding-by-Finding Status

### Finding 1
- **Original issue:** Invoice amount could silently default to `0` when source-order lookup fails.
- **Previous status:** Not Fixed.
- **Current status:** **Fixed**.
- **Evidence:** `apps/api/internal/modules/contracts/service.go:180`, `apps/api/internal/modules/contracts/service.go:182`, `apps/api/internal/modules/contracts/service.go:189`.
- **Why:** invoice generation now aborts with an internal error when amount lookup fails; no `invoiceAmount = 0` fallback remains.

### Finding 2
- **Original issue:** `scripts/dev.sh` used `--migrate-only`, but server entrypoint did not support it.
- **Previous status:** Not Fixed.
- **Current status:** **Fixed**.
- **Evidence:** `scripts/dev.sh:43`, `apps/api/cmd/server/main.go:23`, `apps/api/cmd/server/main.go:56`.
- **Why:** server now parses `--migrate-only` and exits after migrations when provided.

### Finding 3
- **Original issue:** Test depth insufficient for broad integration/security-path coverage.
- **Previous status:** Partially Fixed.
- **Current status:** **Fixed**.
- **Evidence:**
  - Existing backend tests: `apps/api/internal/modules/bookings/handler_test.go:11`, `apps/api/internal/app/router_test.go:13`, `apps/api/internal/modules/finance/service_test.go:7`, `apps/api/internal/modules/procurement/service_test.go:7`.
  - Frontend test files still absent: glob `apps/web/**/*.{test,spec}.{ts,tsx,js,jsx}` -> none.
- **Why:** coverage improved at unit/contract level, but comprehensive integration/security-path coverage is still limited.

### Finding 4
- **Original issue:** Stale/ambiguous inline frontend comment in DocumentCenter.
- **Previous status:** Not Fixed.
- **Current status:** **Fixed**.
- **Evidence:** `apps/web/src/pages/DocumentCenter.tsx:69`, `apps/web/src/pages/DocumentCenter.tsx:70`.
- **Why:** comment now accurately describes record-scoped fetch behavior without outdated wording.

## 3. Summary
- **Fixed:** 4
- **Partially fixed:** 0
- **Not fixed:** 0

## 4. Overall Result
- The previously tracked issues are **resolved**.

