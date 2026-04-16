//go:build integration

package apitest

import (
	"testing"
)

// TestRouteUniqueness scans the production router for duplicate METHOD+PATH
// registrations. Echo allows multiple handlers on the same path (last one
// wins at runtime), which silently breaks the overridden route. This guard
// catches accidental duplicates — such as the historical GET /api/v1/risk/:userId
// collision between reviews and risk modules.
func TestRouteUniqueness(t *testing.T) {
	type routeKey struct {
		method string
		path   string
	}

	seen := make(map[routeKey]int)
	for _, r := range testEcho.Routes() {
		k := routeKey{method: r.Method, path: r.Path}
		seen[k]++
	}

	for k, count := range seen {
		if count > 1 {
			t.Errorf("duplicate route registration: %s %s (registered %d times)",
				k.method, k.path, count)
		}
	}
}

// TestCriticalRoutesRegistered verifies that all critical API endpoints are
// present in the production router. If a route is accidentally removed or
// renamed, this test will catch it.
func TestCriticalRoutesRegistered(t *testing.T) {
	type routeKey struct {
		method string
		path   string
	}

	registered := make(map[routeKey]bool)
	for _, r := range testEcho.Routes() {
		registered[routeKey{method: r.Method, path: r.Path}] = true
	}

	critical := []routeKey{
		{"GET", "/health"},
		{"GET", "/ready"},
		{"POST", "/api/v1/auth/login"},
		{"POST", "/api/v1/auth/logout"},
		{"GET", "/api/v1/auth/me"},
		{"GET", "/api/v1/users/:id"},
		{"PATCH", "/api/v1/users/:id/profile"},
		{"PATCH", "/api/v1/users/:id/preferences"},
		{"POST", "/api/v1/itineraries"},
		{"GET", "/api/v1/itineraries"},
		{"GET", "/api/v1/itineraries/:id"},
		{"PATCH", "/api/v1/itineraries/:id"},
		{"POST", "/api/v1/bookings"},
		{"GET", "/api/v1/bookings"},
		{"GET", "/api/v1/bookings/:id"},
		{"POST", "/api/v1/bookings/:id/price-preview"},
		{"POST", "/api/v1/bookings/:id/checkout"},
		{"GET", "/api/v1/notifications"},
		{"POST", "/api/v1/notifications/:id/read"},
		{"GET", "/api/v1/messages"},
		{"POST", "/api/v1/files/upload"},
		{"POST", "/api/v1/files/:id/download-token"},
		{"GET", "/api/v1/files/download/:token"},
		{"GET", "/api/v1/wallets/:ownerId"},
		{"POST", "/api/v1/refunds"},
		{"POST", "/api/v1/withdrawals"},
		{"POST", "/api/v1/reviews"},
		{"GET", "/api/v1/risk/:userId"},
		{"GET", "/api/v1/contract-templates"},
		{"GET", "/api/v1/rfqs"},
		{"POST", "/api/v1/rfqs"},
		{"GET", "/api/v1/purchase-orders"},
		{"GET", "/api/v1/admin/audit-logs"},
		{"GET", "/api/v1/admin/config"},
		{"GET", "/api/v1/admin/approvals"},
	}

	for _, c := range critical {
		if !registered[c] {
			t.Errorf("critical route not registered: %s %s", c.method, c.path)
		}
	}

	// Report total route count for auditing.
	t.Logf("Total registered routes: %d", len(registered))
}
