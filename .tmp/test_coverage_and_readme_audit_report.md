# Test Coverage Audit

## Scope and Method

- Static inspection only (no code/test execution).
- Inspected production route wiring and handler registrations:
  - `apps/api/internal/app/router.go`
  - `apps/api/internal/auth/handler.go`
  - `apps/api/internal/modules/*/handler.go`
- Inspected backend tests:
  - real-router integration tests: `apps/api/internal/apitest/*_test.go` (build tag `integration`)
  - synthetic contract tests: `apps/api/internal/contract_test.go`
  - unit tests: `apps/api/internal/**/*_test.go`

## Backend Endpoint Inventory

Definition used: unique `METHOD + fully resolved PATH` (deduplicated).

Total unique endpoints discovered: **99**

Inventory evidence:
- router grouping and prefix resolution: `apps/api/internal/app/router.go:46-87`
- module route registrations: `apps/api/internal/modules/*/handler.go`
- auth route registrations: `apps/api/internal/auth/handler.go:24-28`

Resolved endpoint list:
1. `GET /health`
2. `GET /ready`
3. `POST /api/v1/auth/login`
4. `POST /api/v1/auth/logout`
5. `GET /api/v1/auth/me`
6. `GET /api/v1/users/:id`
7. `PATCH /api/v1/users/:id/profile`
8. `PATCH /api/v1/users/:id/preferences`
9. `GET /api/v1/admin/users`
10. `POST /api/v1/itineraries`
11. `GET /api/v1/itineraries`
12. `GET /api/v1/itineraries/:id`
13. `PATCH /api/v1/itineraries/:id`
14. `POST /api/v1/itineraries/:id/publish`
15. `POST /api/v1/itineraries/:id/checkpoints`
16. `PATCH /api/v1/itineraries/:id/checkpoints/:checkpointId`
17. `DELETE /api/v1/itineraries/:id/checkpoints/:checkpointId`
18. `POST /api/v1/itineraries/:id/members`
19. `DELETE /api/v1/itineraries/:id/members/:userId`
20. `POST /api/v1/itineraries/:id/form-definitions`
21. `PATCH /api/v1/itineraries/:id/form-definitions/:defId`
22. `GET /api/v1/itineraries/:id/form-definitions`
23. `POST /api/v1/itineraries/:id/form-submissions`
24. `GET /api/v1/itineraries/:id/form-submissions`
25. `GET /api/v1/itineraries/:id/change-events`
26. `GET /api/v1/coupons/coupons/available`
27. `POST /api/v1/coupons/coupons/evaluate`
28. `POST /api/v1/coupons/coupons/redeem-preview`
29. `GET /api/v1/bookings`
30. `POST /api/v1/bookings`
31. `GET /api/v1/bookings/:id`
32. `POST /api/v1/bookings/:id/price-preview`
33. `POST /api/v1/bookings/:id/checkout`
34. `POST /api/v1/bookings/:id/record-tender`
35. `POST /api/v1/bookings/:id/cancel`
36. `POST /api/v1/bookings/:id/complete`
37. `GET /api/v1/notifications`
38. `POST /api/v1/notifications/:id/read`
39. `GET /api/v1/messages`
40. `POST /api/v1/messages/callback-queue/export`
41. `GET /api/v1/send-logs`
42. `GET /api/v1/users/:id/dnd`
43. `PATCH /api/v1/users/:id/dnd`
44. `GET /api/v1/users/:id/subscriptions`
45. `PATCH /api/v1/users/:id/subscriptions`
46. `POST /api/v1/files/upload`
47. `POST /api/v1/files/:id/download-token`
48. `GET /api/v1/files/download/:token`
49. `GET /api/v1/files/record/:recordType/:recordId`
50. `GET /api/v1/wallets/:ownerId`
51. `GET /api/v1/wallets/:ownerId/transactions`
52. `POST /api/v1/payments/record-tender`
53. `POST /api/v1/settlements/:orderId/release`
54. `POST /api/v1/refunds`
55. `POST /api/v1/withdrawals`
56. `POST /api/v1/withdrawals/:id/approve`
57. `POST /api/v1/withdrawals/:id/reject`
58. `GET /api/v1/reconciliation`
59. `GET /api/v1/escrows/:ownerId`
60. `POST /api/v1/reviews`
61. `GET /api/v1/reviews/subject/:userId`
62. `GET /api/v1/credit-tiers/:userId`
63. `POST /api/v1/violations`
64. `POST /api/v1/no-shows`
65. `POST /api/v1/harassment-flags`
66. `GET /api/v1/risk/:userId`
67. `GET /api/v1/contract-templates`
68. `POST /api/v1/contracts/generate`
69. `POST /api/v1/invoice-requests`
70. `POST /api/v1/invoice-requests/:id/approve`
71. `POST /api/v1/invoices/:id/generate`
72. `GET /api/v1/invoice-requests`
73. `GET /api/v1/supplier-quotes`
74. `GET /api/v1/rfqs`
75. `POST /api/v1/rfqs`
76. `GET /api/v1/rfqs/:id`
77. `POST /api/v1/rfqs/:id/issue`
78. `POST /api/v1/rfqs/:id/quotes`
79. `GET /api/v1/rfqs/:id/comparison`
80. `POST /api/v1/rfqs/:id/select`
81. `GET /api/v1/purchase-orders`
82. `POST /api/v1/purchase-orders`
83. `GET /api/v1/purchase-orders/:id`
84. `POST /api/v1/purchase-orders/:id/accept`
85. `POST /api/v1/purchase-orders/:id/deliveries`
86. `GET /api/v1/deliveries`
87. `POST /api/v1/deliveries/:id/inspect`
88. `POST /api/v1/discrepancies`
89. `GET /api/v1/exceptions`
90. `POST /api/v1/exceptions/:id/waivers`
91. `POST /api/v1/exceptions/:id/settlement-adjustments`
92. `POST /api/v1/exceptions/:id/close`
93. `GET /api/v1/admin/approvals`
94. `POST /api/v1/admin/approvals/:id/resolve`
95. `POST /api/v1/admin/users/:id/blacklist`
96. `POST /api/v1/admin/users/:id/unblacklist`
97. `GET /api/v1/admin/audit-logs`
98. `GET /api/v1/admin/send-logs`
99. `GET /api/v1/admin/config`

## API Test Mapping Table

Legend:
- `covered=yes` only when a test hits the exact endpoint and reaches the real production route handler.
- `true no-mock HTTP` = real app/router + real middleware/services/repositories (`apitest`).
- `HTTP with mocking` = synthetic closures in `contract_test.go`.

1. `GET /health` — covered: **yes** — type: true no-mock HTTP — evidence: `TestHealth_Returns200`
2. `GET /ready` — covered: **yes** — type: true no-mock HTTP — evidence: `TestReady_Returns200`
3. `POST /api/v1/auth/login` — covered: **yes** — type: true no-mock HTTP — evidence: `TestLogin_ValidCredentials`
4. `POST /api/v1/auth/logout` — covered: **yes** — type: true no-mock HTTP — evidence: `TestLogout_Authenticated`
5. `GET /api/v1/auth/me` — covered: **yes** — type: true no-mock HTTP — evidence: `TestMe_Authenticated`
6. `GET /api/v1/users/:id` — covered: **yes** — type: true no-mock HTTP — evidence: `TestGetUser_OwnProfile`
7. `PATCH /api/v1/users/:id/profile` — covered: **yes** — type: true no-mock HTTP — evidence: `TestUpdateProfile`
8. `PATCH /api/v1/users/:id/preferences` — covered: **yes** — type: true no-mock HTTP — evidence: `TestUpdatePreferences`
9. `GET /api/v1/admin/users` — covered: **yes** — type: true no-mock HTTP — evidence: `TestAdminListUsers`
10. `POST /api/v1/itineraries` — covered: **yes** — type: true no-mock HTTP — evidence: `TestCreateItinerary`
11. `GET /api/v1/itineraries` — covered: **yes** — type: true no-mock HTTP — evidence: `TestListItineraries`
12. `GET /api/v1/itineraries/:id` — covered: **yes** — type: true no-mock HTTP — evidence: `TestGetItinerary`
13. `PATCH /api/v1/itineraries/:id` — covered: **yes** — type: true no-mock HTTP — evidence: `TestUpdateItinerary`
14. `POST /api/v1/itineraries/:id/publish` — covered: **yes** — type: true no-mock HTTP — evidence: `TestPublishItinerary`
15. `POST /api/v1/itineraries/:id/checkpoints` — covered: **yes** — type: true no-mock HTTP — evidence: `TestAddCheckpoint`
16. `PATCH /api/v1/itineraries/:id/checkpoints/:checkpointId` — covered: **yes** — type: true no-mock HTTP — evidence: `TestUpdateCheckpoint`
17. `DELETE /api/v1/itineraries/:id/checkpoints/:checkpointId` — covered: **yes** — type: true no-mock HTTP — evidence: `TestDeleteCheckpoint`
18. `POST /api/v1/itineraries/:id/members` — covered: **yes** — type: true no-mock HTTP — evidence: `TestAddMember`
19. `DELETE /api/v1/itineraries/:id/members/:userId` — covered: **yes** — type: true no-mock HTTP — evidence: `TestRemoveMember`
20. `POST /api/v1/itineraries/:id/form-definitions` — covered: **yes** — type: true no-mock HTTP — evidence: `TestCreateFormDefinition`
21. `PATCH /api/v1/itineraries/:id/form-definitions/:defId` — covered: **yes** — type: true no-mock HTTP — evidence: `TestUpdateFormDefinition`
22. `GET /api/v1/itineraries/:id/form-definitions` — covered: **yes** — type: true no-mock HTTP — evidence: `TestItineraryFormDefinitions`
23. `POST /api/v1/itineraries/:id/form-submissions` — covered: **yes** — type: true no-mock HTTP — evidence: `TestSubmitFormSubmission`
24. `GET /api/v1/itineraries/:id/form-submissions` — covered: **yes** — type: true no-mock HTTP — evidence: `TestItineraryFormSubmissions`
25. `GET /api/v1/itineraries/:id/change-events` — covered: **yes** — type: true no-mock HTTP — evidence: `TestItineraryChangeEvents`
26. `GET /api/v1/coupons/coupons/available` — covered: **yes** — type: true no-mock HTTP — evidence: `TestListActiveCoupons`
27. `POST /api/v1/coupons/coupons/evaluate` — covered: **yes** — type: true no-mock HTTP — evidence: `TestEvaluateCoupons`
28. `POST /api/v1/coupons/coupons/redeem-preview` — covered: **yes** — type: true no-mock HTTP — evidence: `TestRedeemPreview`
29. `GET /api/v1/bookings` — covered: **yes** — type: true no-mock HTTP — evidence: `TestListBookings`
30. `POST /api/v1/bookings` — covered: **yes** — type: true no-mock HTTP — evidence: `TestCreateBooking`
31. `GET /api/v1/bookings/:id` — covered: **yes** — type: true no-mock HTTP — evidence: `TestGetBooking`
32. `POST /api/v1/bookings/:id/price-preview` — covered: **yes** — type: true no-mock HTTP — evidence: `TestBookingPricePreview`
33. `POST /api/v1/bookings/:id/checkout` — covered: **yes** — type: true no-mock HTTP — evidence: `TestBookingCheckout`
34. `POST /api/v1/bookings/:id/record-tender` — covered: **yes** — type: true no-mock HTTP — evidence: `TestBookingRecordTender`
35. `POST /api/v1/bookings/:id/cancel` — covered: **yes** — type: true no-mock HTTP — evidence: `TestBookingCancel_NotOwner`
36. `POST /api/v1/bookings/:id/complete` — covered: **yes** — type: true no-mock HTTP — evidence: `TestBookingComplete`
37. `GET /api/v1/notifications` — covered: **yes** — type: true no-mock HTTP — evidence: `TestListNotifications`
38. `POST /api/v1/notifications/:id/read` — covered: **yes** — type: true no-mock HTTP — evidence: `TestMarkNotificationRead`
39. `GET /api/v1/messages` — covered: **yes** — type: true no-mock HTTP — evidence: `TestListMessages`
40. `POST /api/v1/messages/callback-queue/export` — covered: **yes** — type: true no-mock HTTP — evidence: `TestExportCallbackQueue_Admin`
41. `GET /api/v1/send-logs` — covered: **yes** — type: true no-mock HTTP — evidence: `TestListSendLogs`
42. `GET /api/v1/users/:id/dnd` — covered: **yes** — type: true no-mock HTTP — evidence: `TestGetDNDSettings`
43. `PATCH /api/v1/users/:id/dnd` — covered: **yes** — type: true no-mock HTTP — evidence: `TestUpdateDNDSettings`
44. `GET /api/v1/users/:id/subscriptions` — covered: **yes** — type: true no-mock HTTP — evidence: `TestGetSubscriptionPrefs`
45. `PATCH /api/v1/users/:id/subscriptions` — covered: **yes** — type: true no-mock HTTP — evidence: `TestUpdateSubscriptionPrefs`
46. `POST /api/v1/files/upload` — covered: **yes** — type: true no-mock HTTP — evidence: `TestFileUpload_Authenticated`
47. `POST /api/v1/files/:id/download-token` — covered: **yes** — type: true no-mock HTTP — evidence: `TestFileDownloadToken_Authenticated`
48. `GET /api/v1/files/download/:token` — covered: **yes** — type: true no-mock HTTP — evidence: `TestFileDownload_PublicPath`
49. `GET /api/v1/files/record/:recordType/:recordId` — covered: **yes** — type: true no-mock HTTP — evidence: `TestFileRecordList_Authenticated`
50. `GET /api/v1/wallets/:ownerId` — covered: **yes** — type: true no-mock HTTP — evidence: `TestGetWallet`
51. `GET /api/v1/wallets/:ownerId/transactions` — covered: **yes** — type: true no-mock HTTP — evidence: `TestGetWalletTransactions`
52. `POST /api/v1/payments/record-tender` — covered: **yes** — type: true no-mock HTTP — evidence: `TestRecordTender_Accountant`
53. `POST /api/v1/settlements/:orderId/release` — covered: **yes** — type: true no-mock HTTP — evidence: `TestReleaseEscrow_Accountant`
54. `POST /api/v1/refunds` — covered: **yes** — type: true no-mock HTTP — evidence: `TestProcessRefund_Accountant`
55. `POST /api/v1/withdrawals` — covered: **yes** — type: true no-mock HTTP — evidence: `TestRequestWithdrawal_Courier`
56. `POST /api/v1/withdrawals/:id/approve` — covered: **yes** — type: true no-mock HTTP — evidence: `TestApproveWithdrawal_Accountant`
57. `POST /api/v1/withdrawals/:id/reject` — covered: **yes** — type: true no-mock HTTP — evidence: `TestRejectWithdrawal_Accountant`
58. `GET /api/v1/reconciliation` — covered: **yes** — type: true no-mock HTTP — evidence: `TestReconciliation_Admin`
59. `GET /api/v1/escrows/:ownerId` — covered: **yes** — type: true no-mock HTTP — evidence: `TestGetEscrows`
60. `POST /api/v1/reviews` — covered: **yes** — type: true no-mock HTTP — evidence: `TestSubmitReview_MissingFields`
61. `GET /api/v1/reviews/subject/:userId` — covered: **yes** — type: true no-mock HTTP — evidence: `TestGetReviewsBySubject`
62. `GET /api/v1/credit-tiers/:userId` — covered: **yes** — type: true no-mock HTTP — evidence: `TestGetCreditTier`
63. `POST /api/v1/violations` — covered: **yes** — type: true no-mock HTTP — evidence: `TestReportViolation`
64. `POST /api/v1/no-shows` — covered: **yes** — type: true no-mock HTTP — evidence: `TestReportNoShow`
65. `POST /api/v1/harassment-flags` — covered: **yes** — type: true no-mock HTTP — evidence: `TestReportHarassment`
66. `GET /api/v1/risk/:userId` — covered: **yes** — type: true no-mock HTTP — evidence: `TestGetRiskSummary_Admin`
67. `GET /api/v1/contract-templates` — covered: **yes** — type: true no-mock HTTP — evidence: `TestGetContractTemplates`
68. `POST /api/v1/contracts/generate` — covered: **yes** — type: true no-mock HTTP — evidence: `TestGenerateContract`
69. `POST /api/v1/invoice-requests` — covered: **yes** — type: true no-mock HTTP — evidence: `TestCreateInvoiceRequest_Authenticated`
70. `POST /api/v1/invoice-requests/:id/approve` — covered: **yes** — type: true no-mock HTTP — evidence: `TestApproveInvoiceRequest_Accountant`
71. `POST /api/v1/invoices/:id/generate` — covered: **yes** — type: true no-mock HTTP — evidence: `TestGenerateInvoice_Accountant`
72. `GET /api/v1/invoice-requests` — covered: **yes** — type: true no-mock HTTP — evidence: `TestListInvoiceRequests`
73. `GET /api/v1/supplier-quotes` — covered: **yes** — type: true no-mock HTTP — evidence: `TestListSupplierQuotes`
74. `GET /api/v1/rfqs` — covered: **yes** — type: true no-mock HTTP — evidence: `TestListRFQs`
75. `POST /api/v1/rfqs` — covered: **yes** — type: true no-mock HTTP — evidence: `TestCreateRFQ`
76. `GET /api/v1/rfqs/:id` — covered: **yes** — type: true no-mock HTTP — evidence: `TestGetRFQById`
77. `POST /api/v1/rfqs/:id/issue` — covered: **yes** — type: true no-mock HTTP — evidence: `TestIssueRFQ`
78. `POST /api/v1/rfqs/:id/quotes` — covered: **yes** — type: true no-mock HTTP — evidence: `TestSubmitQuote`
79. `GET /api/v1/rfqs/:id/comparison` — covered: **yes** — type: true no-mock HTTP — evidence: `TestGetRFQComparison`
80. `POST /api/v1/rfqs/:id/select` — covered: **yes** — type: true no-mock HTTP — evidence: `TestSelectSupplier`
81. `GET /api/v1/purchase-orders` — covered: **yes** — type: true no-mock HTTP — evidence: `TestListPurchaseOrders`
82. `POST /api/v1/purchase-orders` — covered: **yes** — type: true no-mock HTTP — evidence: `TestCreatePurchaseOrder`
83. `GET /api/v1/purchase-orders/:id` — covered: **yes** — type: true no-mock HTTP — evidence: `TestGetPurchaseOrderById`
84. `POST /api/v1/purchase-orders/:id/accept` — covered: **yes** — type: true no-mock HTTP — evidence: `TestAcceptPurchaseOrder`
85. `POST /api/v1/purchase-orders/:id/deliveries` — covered: **no** — type: unit-only / indirect — evidence: `TestCreateDelivery` uses courier token, but route requires accountant/admin (`apps/api/internal/modules/procurement/handler.go:41`), so request is blocked at RBAC middleware and does not prove handler execution
86. `GET /api/v1/deliveries` — covered: **yes** — type: true no-mock HTTP — evidence: `TestListDeliveries`
87. `POST /api/v1/deliveries/:id/inspect` — covered: **yes** — type: true no-mock HTTP — evidence: `TestInspectDelivery`
88. `POST /api/v1/discrepancies` — covered: **yes** — type: true no-mock HTTP — evidence: `TestCreateDiscrepancy`
89. `GET /api/v1/exceptions` — covered: **yes** — type: true no-mock HTTP — evidence: `TestListExceptions`
90. `POST /api/v1/exceptions/:id/waivers` — covered: **yes** — type: true no-mock HTTP — evidence: `TestExceptionWaiver`
91. `POST /api/v1/exceptions/:id/settlement-adjustments` — covered: **yes** — type: true no-mock HTTP — evidence: `TestExceptionSettlementAdjustment`
92. `POST /api/v1/exceptions/:id/close` — covered: **yes** — type: true no-mock HTTP — evidence: `TestCloseException`
93. `GET /api/v1/admin/approvals` — covered: **yes** — type: true no-mock HTTP — evidence: `TestGetPendingApprovals_Admin`
94. `POST /api/v1/admin/approvals/:id/resolve` — covered: **yes** — type: true no-mock HTTP — evidence: `TestResolveApproval_Admin`
95. `POST /api/v1/admin/users/:id/blacklist` — covered: **yes** — type: true no-mock HTTP — evidence: `TestBlacklistUser_Admin`
96. `POST /api/v1/admin/users/:id/unblacklist` — covered: **yes** — type: true no-mock HTTP — evidence: `TestUnblacklistUser_Admin`
97. `GET /api/v1/admin/audit-logs` — covered: **yes** — type: true no-mock HTTP — evidence: `TestAdminGetAuditLogs`
98. `GET /api/v1/admin/send-logs` — covered: **yes** — type: true no-mock HTTP — evidence: `TestAdminGetSendLogs`
99. `GET /api/v1/admin/config` — covered: **yes** — type: true no-mock HTTP — evidence: `TestAdminGetConfig`

## API Test Classification

1. **True No-Mock HTTP**
   - `apps/api/internal/apitest/main_test.go` bootstraps real app/router (`app.New` + `SetupRouter`) and real DB migrations/seeds.
   - Endpoint tests in `apps/api/internal/apitest/*_test.go` use `doRequest` through `testEcho.ServeHTTP`.
   - **134 test functions** across 16 test files.

2. **HTTP with Mocking (synthetic handlers)**
   - `apps/api/internal/contract_test.go` explicitly uses synthetic closures and not production router/DB.

3. **Non-HTTP (unit/integration without real HTTP route execution)**
   - `apps/api/internal/modules/*/*_test.go`, `apps/api/internal/middleware/*_test.go`, `apps/api/internal/auth/service_test.go`, etc.

## Mock Detection

- No `jest.mock` / `vi.mock` / `sinon.stub` found.
- Synthetic HTTP mocking detected:
  - `apps/api/internal/contract_test.go` (inline route closures via `e.GET`/`e.POST`; file header states not production router/DB).

## Coverage Summary

- Total endpoints: **99**
- Endpoints with true no-mock HTTP tests (handler reached): **98**
- True API coverage %: **98.99%**

## Unit Test Summary

Representative backend test files:
- Real API integration: `apps/api/internal/apitest/*.go` (16 files, 134 test functions)
- Synthetic contract tests: `apps/api/internal/contract_test.go`
- Units: `apps/api/internal/modules/*/*_test.go`, `apps/api/internal/middleware/*_test.go`, `apps/api/internal/common/roles_test.go`, `apps/api/internal/auth/service_test.go`

Module coverage by integration tests:
- Auth: 3/3 routes (login, logout, me)
- Users: 4/4 routes (get, update profile, update preferences, admin list)
- Itineraries: 16/16 routes (full CRUD, checkpoints, members, forms, events, publish)
- Pricing: 3/3 routes (available, evaluate, redeem-preview)
- Bookings: 8/8 routes (list, create, get, preview, checkout, record-tender, cancel, complete)
- Notifications: 10/10 routes (notifications, messages, export, send-logs, DND, subscriptions)
- Files: 4/4 routes (upload, download-token, download, record-list)
- Finance: 10/10 routes (wallets, transactions, tender, release, refunds, withdrawals, approve/reject, reconciliation, escrows)
- Reviews: 6/6 routes (submit, get-by-subject, credit-tiers, violations, no-shows, harassment)
- Risk: 5/5 routes (summary, approvals, resolve, blacklist, unblacklist)
- Contracts: 6/6 routes (templates, generate, invoice-requests CRUD, approve, generate-invoice)
- Procurement: 15/16 routes (all except `POST /api/v1/purchase-orders/:id/deliveries` proven handler execution)
- Admin: 3/3 routes (audit-logs, send-logs, config)
- Health: 2/2 routes (health, ready)

## Tests Check

- Success paths: present across all major modules (login, CRUD, pricing, procurement lifecycle)
- Failure paths: 401 unauthenticated, 403 wrong role, 400/422 validation errors across all modules
- Edge cases: role boundaries, missing fields, self-review prevention, risk blocks (contract tests)
- Auth/permissions: tested per endpoint with correct-role success + wrong-role denial
- Integration boundaries: real handler → service → repository → database through HTTP
- Assertion depth: status codes + response body field verification + business data assertions for key flows; many mutation tests remain handler-reachability style (`not 401/403`) rather than strict business-outcome assertions
- `run_tests.sh`: Docker-based unit/integration/e2e modes present — **OK**

## End-to-End Expectations

- Fullstack project confirmed (`apps/api` + `apps/web`).
- Real full-stack smoke script exists: `tests/e2e/smoke_test.sh`
- Flow: health → login → list itineraries → create booking → verify persistence → web HTML check → admin flow
- Execution via `./run_tests.sh --e2e`

## Test Coverage Score (0-100)

**91 / 100**

## Score Rationale

- 98/99 endpoints have true no-mock integration tests with real handler execution.
- Core CRUD flows (itineraries, bookings, procurement, finance) are tested end-to-end through real DB.
- Route uniqueness guard prevents regression.
- Fullstack E2E smoke test covers login-to-operation flow.
- Deductions: one procurement route is not proven past middleware (`POST /api/v1/purchase-orders/:id/deliveries`), and many mutation endpoint tests verify handler reachability (`not 401/403`) rather than strict business outcome assertions.

## Key Gaps

1. `POST /api/v1/purchase-orders/:id/deliveries` is not true-covered: integration test uses courier role while route requires accountant/admin (`apps/api/internal/apitest/procurement_test.go:130-136`, `apps/api/internal/modules/procurement/handler.go:41`).
2. Many mutation endpoint tests assert only handler reachability (`not 401/403`) instead of validating concrete domain outcomes and response payload invariants.
3. E2E test is curl-based smoke only; no browser-driven assertion of frontend state transitions.

## Confidence & Assumptions

- Confidence: **high** for classification and endpoint/test mapping.
- Assumptions:
  - Coverage determination requires real handler execution per strict definition.
  - Tests that verify "not 401/403 with correct auth" prove handler reachability through the real middleware chain and handler function.

**Test Coverage Verdict: PASS**

---

# README Audit

## Project Type Detection

- Explicit declaration present near top: `Project type: fullstack` (`README.md:3`) — **PASS**.

## README Location

- `repo/README.md` exists — **PASS**.

## Hard Gate Results

1. **Startup instruction format**
   - Required literal `docker-compose up`: present (`README.md:85`).
   - Result: **PASS**.

2. **Verification method**
   - Dedicated "Verification" section with API health check, API login, authenticated API call, and Web UI flow.
   - Result: **PASS**.

3. **Environment rules (Docker-contained only)**
   - Primary setup path is fully Docker-contained (`docker-compose up`).
   - No `go run`, `npm install`, or other non-Docker runtime commands in the README.
   - Local development section has been completely removed.
   - Migrations section uses Docker commands only (`docker compose run --rm api`).
   - Result: **PASS**.

4. **Demo credentials**
   - All 6 roles documented with email and password.
   - Result: **PASS**.

## Gate-by-Gate Summary

| Gate | Status |
|------|--------|
| Formatting/readability | **PASS** |
| Project type declaration | **PASS** |
| Startup instructions (exact `docker-compose up`) | **PASS** |
| Access method (URL/port table) | **PASS** |
| Verification method (API + UI flow) | **PASS** |
| Environment rules (Docker-contained only) | **PASS** |
| Demo credentials with all roles | **PASS** |
| Testing section with Docker commands | **PASS** |

**README Verdict: PASS**
