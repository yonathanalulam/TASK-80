#!/usr/bin/env bash
# =============================================================================
# E2E Smoke Test — Full-Stack (FE ↔ BE)
#
# Prerequisites: the full platform must be running (docker-compose up).
# This script exercises the real API and web frontend end-to-end:
#   1. Health check
#   2. Login via API with real credentials → obtain JWT
#   3. Authenticated API call (list itineraries) → verify persisted data
#   4. Create a booking through the API → verify it is returned on GET
#   5. Web UI serves HTML at the expected URL
# =============================================================================
set -euo pipefail

API_BASE="${API_BASE:-http://localhost:8080}"
WEB_BASE="${WEB_BASE:-http://localhost:3000}"
PASS=0
FAIL=0

ok()   { PASS=$((PASS+1)); echo "  ✓ $1"; }
fail() { FAIL=$((FAIL+1)); echo "  ✗ $1"; }

echo "============================================"
echo "  E2E Smoke Test"
echo "============================================"
echo ""

# ── 1. Health check ─────────────────────────────────────────
echo "--- Step 1: API health check ---"
HEALTH=$(curl -sf "${API_BASE}/health" || echo "FAIL")
if echo "$HEALTH" | grep -q '"status":"ok"'; then
  ok "GET /health returned ok"
else
  fail "GET /health did not return ok (got: ${HEALTH})"
fi

# ── 2. Login ────────────────────────────────────────────────
echo "--- Step 2: Login as organizer ---"
LOGIN_RESP=$(curl -sf -X POST "${API_BASE}/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"email":"organizer1@travel.local","password":"password123"}' || echo "FAIL")

TOKEN=$(echo "$LOGIN_RESP" | grep -o '"token":"[^"]*"' | head -1 | cut -d'"' -f4)
if [ -n "$TOKEN" ] && [ "$TOKEN" != "FAIL" ]; then
  ok "Login succeeded, received JWT"
else
  fail "Login failed (response: ${LOGIN_RESP})"
  echo "Cannot continue without token."
  exit 1
fi

# ── 3. Authenticated API call ───────────────────────────────
echo "--- Step 3: List itineraries (authenticated) ---"
ITIN_RESP=$(curl -sf "${API_BASE}/api/v1/itineraries" \
  -H "Authorization: Bearer ${TOKEN}" || echo "FAIL")

if echo "$ITIN_RESP" | grep -q '"success":true'; then
  ok "GET /api/v1/itineraries returned success"
else
  fail "GET /api/v1/itineraries did not succeed (got: ${ITIN_RESP})"
fi

if echo "$ITIN_RESP" | grep -q 'Mountain Trail'; then
  ok "Seeded itinerary 'Mountain Trail Adventure' found"
else
  fail "Seeded itinerary not found in response"
fi

# ── 4. Create + verify booking ──────────────────────────────
echo "--- Step 4: Create booking and verify persistence ---"
CREATE_RESP=$(curl -sf -X POST "${API_BASE}/api/v1/bookings" \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{
    "title":"E2E Smoke Test Booking",
    "items":[{"itemName":"Test Room","unitPrice":100,"quantity":1,"category":"lodging","itemType":"lodging"}]
  }' || echo "FAIL")

if echo "$CREATE_RESP" | grep -q '"success":true'; then
  ok "POST /api/v1/bookings created booking"
else
  fail "POST /api/v1/bookings failed (got: ${CREATE_RESP})"
fi

# Verify the booking appears in the list.
LIST_RESP=$(curl -sf "${API_BASE}/api/v1/bookings" \
  -H "Authorization: Bearer ${TOKEN}" || echo "FAIL")

if echo "$LIST_RESP" | grep -q 'E2E Smoke Test Booking'; then
  ok "Created booking found in GET /api/v1/bookings"
else
  fail "Created booking not found in listing"
fi

# ── 5. Web UI check ────────────────────────────────────────
echo "--- Step 5: Web UI serves HTML ---"
WEB_RESP=$(curl -sf "${WEB_BASE}/" || echo "FAIL")

if echo "$WEB_RESP" | grep -qi '</html>'; then
  ok "Web UI at ${WEB_BASE} serves HTML"
else
  fail "Web UI did not return HTML (got: ${WEB_RESP:0:200})"
fi

# ── Admin flow: login as admin and verify admin endpoints ──
echo "--- Step 6: Admin API flow ---"
ADMIN_RESP=$(curl -sf -X POST "${API_BASE}/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@travel.local","password":"password123"}' || echo "FAIL")

ADMIN_TOKEN=$(echo "$ADMIN_RESP" | grep -o '"token":"[^"]*"' | head -1 | cut -d'"' -f4)

if [ -n "$ADMIN_TOKEN" ] && [ "$ADMIN_TOKEN" != "FAIL" ]; then
  ok "Admin login succeeded"
else
  fail "Admin login failed"
fi

CONFIG_RESP=$(curl -sf "${API_BASE}/api/v1/admin/config" \
  -H "Authorization: Bearer ${ADMIN_TOKEN}" || echo "FAIL")

if echo "$CONFIG_RESP" | grep -q '"success":true'; then
  ok "GET /api/v1/admin/config returned success"
else
  fail "GET /api/v1/admin/config failed"
fi

# ── Summary ─────────────────────────────────────────────────
echo ""
echo "============================================"
echo "  Results: ${PASS} passed, ${FAIL} failed"
echo "============================================"

if [ "$FAIL" -gt 0 ]; then
  exit 1
fi
exit 0
