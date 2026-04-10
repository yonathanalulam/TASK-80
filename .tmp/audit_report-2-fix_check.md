# Fix Check Report - Round 2 (Against `audit_report_2_fix-check_round1.md`)

## Scope
- Static re-check of the 6 issues from `.tmp/audit_report_2_fix-check_round1.md`.
- No runtime execution performed.

## Results Summary
- Fixed: **6**
- Partially fixed: **0**
- Not fixed: **0**

## Issue-by-Issue Verification

1. **Review submission contract mismatch (frontend payload vs backend DTO)**
   - **Status:** **Fixed**
   - **Evidence:** frontend submits canonical payload with `orderType`, `orderId`, `scores[]` and renders `scores` from list items (`apps/web/src/pages/ReviewDashboard.tsx:136`, `apps/web/src/pages/ReviewDashboard.tsx:375`, `apps/web/src/pages/ReviewDashboard.tsx:378`, `apps/web/src/pages/ReviewDashboard.tsx:265`), matching backend DTO (`apps/api/internal/modules/reviews/dto.go:11`, `apps/api/internal/modules/reviews/dto.go:12`, `apps/api/internal/modules/reviews/dto.go:15`, `apps/api/internal/modules/reviews/dto.go:44`).

2. **Booking price-preview response shape mismatch in detail page**
   - **Status:** **Fixed**
   - **Evidence:** frontend now uses `eligibleCoupons`, `ineligibleCoupons`, and `escrowHoldAmount` (`apps/web/src/pages/BookingDetail.tsx:350`, `apps/web/src/pages/BookingDetail.tsx:362`, `apps/web/src/pages/BookingDetail.tsx:383`), aligned with backend pricing result model (`apps/api/internal/modules/pricing/model.go:59`, `apps/api/internal/modules/pricing/model.go:61`, `apps/api/internal/modules/pricing/model.go:62`).

3. **Document center file/invoice list model mismatch with backend response models**
   - **Status:** **Fixed**
   - **Evidence:** frontend interfaces and table fields now use backend names (`originalFilename/mimeType/byteSize`, `orderType/orderId/notes`) (`apps/web/src/pages/DocumentCenter.tsx:7`, `apps/web/src/pages/DocumentCenter.tsx:8`, `apps/web/src/pages/DocumentCenter.tsx:9`, `apps/web/src/pages/DocumentCenter.tsx:25`, `apps/web/src/pages/DocumentCenter.tsx:26`, `apps/web/src/pages/DocumentCenter.tsx:28`, `apps/web/src/pages/DocumentCenter.tsx:371`, `apps/web/src/pages/DocumentCenter.tsx:381`), matching backend models (`apps/api/internal/modules/files/model.go:8`, `apps/api/internal/modules/files/model.go:10`, `apps/api/internal/modules/contracts/model.go:41`, `apps/api/internal/modules/contracts/model.go:42`, `apps/api/internal/modules/contracts/model.go:44`).

4. **RFQ create role affordance mismatch between frontend and backend**
   - **Status:** **Fixed**
   - **Evidence:** frontend allows organizer/admin/accountant create (`apps/web/src/pages/ProcurementDashboard.tsx:98`) and backend RFQ route now permits organizer via `rfqCreator` middleware (`apps/api/internal/modules/procurement/handler.go:25`, `apps/api/internal/modules/procurement/handler.go:30`).

5. **Test suite depth mostly unit/shape-focused with limited integration checks**
   - **Status:** **Fixed (material improvement)**
   - **Evidence:** repository now includes broad integration-path tests covering authz boundaries, risk-blocked workflows, file token flow, procurement lifecycle, refund workflow, review submit contracts, and booking preview/checkout contract paths (`apps/api/internal/integration_test.go:63`, `apps/api/internal/integration_test.go:270`, `apps/api/internal/integration_test.go:468`, `apps/api/internal/integration_test.go:703`, `apps/api/internal/integration_test.go:1012`, `apps/api/internal/integration_test.go:1256`, `apps/api/internal/integration_test.go:1404`). Frontend contract tests were also added (`apps/web/src/__tests__/api-contracts.test.ts:1`).

6. **DocumentCenter record linking/listing gap**
   - **Status:** **Fixed**
   - **Evidence:** record selection state now drives both upload link fields and list query path (`apps/web/src/pages/DocumentCenter.tsx:68`, `apps/web/src/pages/DocumentCenter.tsx:69`, `apps/web/src/pages/DocumentCenter.tsx:72`, `apps/web/src/pages/DocumentCenter.tsx:74`, `apps/web/src/pages/DocumentCenter.tsx:100`, `apps/web/src/pages/DocumentCenter.tsx:101`), and backend links files when both values are present (`apps/api/internal/modules/files/service.go:100`).

## Final Check Verdict
- All issues listed in round 1 are now statically verified as **fixed**.
- Remaining validation is runtime-only (UI behavior under real data, download behavior in browser, and end-to-end operational workflows).
