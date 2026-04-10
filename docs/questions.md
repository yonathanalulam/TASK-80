# Project Clarification Questions

## Business Logic & Architecture Questions Log

Below are ambiguities identified in the original prompt that required interpretation. Each entry documents the question, why it was ambiguous, and the solution implemented in the codebase.

---

### 1. Frontend Rendering Approach

**Question:** What rendering strategy should the React web app use — a full single-page application with client-side routing, or server-rendered HTML with progressive enhancement?

**My Understanding:** The prompt specifies a "React web app" but does not prescribe whether it should be a traditional SPA, an SSR-based approach, or a hybrid. The mention of "guided flow" and "clear in-app change notifications visible to all participants on next login" could imply either approach.

**Solution:** Implement a full client-side SPA using React 18 with React Router for navigation. The frontend is built with Vite, uses TanStack React Query (v5) for server-state management with 5-minute stale times, Zustand for auth state, and Axios as the HTTP client. All pages are client-rendered with data fetched from the Go API over REST. The layout uses a persistent sidebar with navigation and a top header bar.

---

### 2. Authentication Mechanism

**Question:** What authentication mechanism should be used — session cookies, JWT tokens, OAuth, or something else? How should sessions be managed in the offline-first environment?

**My Understanding:** The prompt mentions roles and access control but does not specify an authentication protocol. The "fully offline environment" constraint rules out external identity providers but leaves open the question of token format and session management.

**Solution:** Implement JWT-based authentication with Bearer tokens. The `POST /api/v1/auth/login` endpoint accepts email/password credentials and returns a JWT. Tokens are stored in localStorage on the client and sent via the `Authorization: Bearer <token>` header. The backend maintains a `sessions` table with token hashes and expiration timestamps. A `GET /api/v1/auth/me` endpoint allows the frontend to rehydrate user state on page load. Logout invalidates the session server-side via `POST /api/v1/auth/logout`.

---

### 3. API Versioning and Structure

**Question:** Should the API be versioned, and what URL structure should it follow?

**My Understanding:** The prompt says "REST-style APIs consumed by the React frontend over a decoupled architecture" but does not specify versioning strategy, URL conventions, or response envelope format.

**Solution:** All API endpoints are prefixed with `/api/v1/`. Responses use a standard envelope: `{ "success": true, "data": ... }` for success and `{ "success": false, "error": { "code": "...", "message": "...", "details": {...} } }` for errors. HTTP status codes follow REST conventions (200 OK, 201 Created, 400 Bad Request, 401 Unauthorized, 403 Forbidden, 404 Not Found, 422 Unprocessable Entity, 500 Internal Server Error).

---

### 4. Coupon Stacking Precedence When Multiple of Same Type

**Question:** When a user submits multiple coupons of the same discount type (e.g., two threshold_fixed coupons), which one wins?

**My Understanding:** The prompt states "at most 1 threshold discount plus 1 percentage discount; new-user gifts never stack" but does not clarify the selection strategy when multiple coupons of the same type are submitted.

**Solution:** The pricing engine evaluates all submitted coupons and selects the one with the highest discount amount within each type. For `threshold_fixed` coupons, the one yielding the largest fixed discount is applied. For `percentage` coupons, the one yielding the largest percentage-based discount is applied. Non-selected coupons of the same type receive a `STACKING_NOT_ALLOWED` reason code. The `new_user_gift` type is exclusive — when applied, it blocks both threshold and percentage coupons entirely.

---

### 5. Member Tier Discount Rates

**Question:** What discount percentages should membership tiers provide, and how do they interact with coupons?

**My Understanding:** The prompt mentions "member pricing" as a coupon type but does not define specific tier discount rates or whether membership discounts stack with coupons.

**Solution:** Membership tier discounts are hardcoded in the pricing engine: Silver = 3% off, Gold = 5% off, Platinum = 8% off. The tier discount is applied first to the subtotal before coupon evaluation begins. It acts as a separate discount layer that stacks with coupon discounts — it is listed as an `AppliedDiscount` of type `"member_tier"` in the pricing result.

---

### 6. Chart of Accounts for Double-Entry Ledger

**Question:** What account codes and structure should the double-entry ledger use?

**My Understanding:** The prompt requires "a full double-entry ledger for every movement" but does not specify the chart of accounts, account naming conventions, or the journal entry structure.

**Solution:** Implement a 10-account chart of accounts: `1000` (CashOnHand), `1100` (ManualTenderClearing), `2000` (EscrowLiability), `2100` (SupplierPayable), `2200` (CourierPayable), `2300` (CustomerWalletLiability), `2400` (RefundLiability), `2500` (SettlementAdjustmentReserve), `4000` (Revenue), `4100` (FeeRevenue), `5000` (AdjustmentExpense). Each financial operation posts a `journal_entry` with balanced `journal_lines` (total debits = total credits). Counterparties are tracked in a dedicated `counterparties` table for reconciliation.

---

### 7. Wallet Types and Platform Accounts

**Question:** How many wallet types are needed, and should the platform itself have wallets for clearing/escrow operations?

**My Understanding:** The prompt mentions "an internal wallet" and "escrow" but does not specify whether the platform needs its own system wallets for clearing operations, or if each role type needs a distinct wallet type.

**Solution:** Implement 7 wallet types: `customer`, `supplier`, `courier`, `platform_clearing`, `escrow_control`, `refund_clearing`, and `fee_revenue`. Platform wallets (clearing, escrow, refund, fee) have no `owner_id` and serve as system-level accounts. User wallets are created per-user for customers, suppliers, and couriers. The seed data pre-creates 4 platform wallets and 4 user wallets for test data.

---

### 8. Credit Tier Definitions and Thresholds

**Question:** What credit tiers should exist, and what thresholds determine tier placement?

**My Understanding:** The prompt says reviews "feed visible credit tiers and can block repeat offenders" but does not define the tier names, thresholds, or how they are computed.

**Solution:** Implement 5 credit tiers: Bronze (min 0 transactions, 0.0 avg rating, max 10 violations), Silver (min 5 transactions, 3.0 avg rating, max 5 violations), Gold (min 15 transactions, 3.5 avg rating, max 3 violations), Platinum (min 30 transactions, 4.0 avg rating, max 1 violation), and Restricted (min 0 transactions, 0.0 avg rating, max 0 violations — for sanctioned accounts). Tiers are computed as snapshots in `user_credit_snapshots` with total transaction count, average rating, and violation count.

