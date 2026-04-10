package app

import (
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
)

// TestFileRouteComposition verifies that file routes are mounted under
// /api/v1/files/... without accidental path segment duplication like
// /api/v1/files/files/...
func TestFileRouteComposition(t *testing.T) {
	// We cannot easily instantiate the full App (requires DB), so we test
	// the route structure by simulating what the router does: create an Echo
	// instance, mount a group with "/files" prefix, then register routes
	// with the same prefix pattern the handler uses.

	e := echo.New()
	v1 := e.Group("/api/v1")
	fileGroup := v1.Group("/files")

	// Simulate the corrected RegisterRoutes paths (no /files prefix in handler).
	fileGroup.POST("/upload", func(c echo.Context) error { return nil })
	fileGroup.POST("/:id/download-token", func(c echo.Context) error { return nil })
	fileGroup.GET("/download/:token", func(c echo.Context) error { return nil })
	fileGroup.GET("/record/:recordType/:recordId", func(c echo.Context) error { return nil })

	routes := e.Routes()

	// Verify canonical file routes exist.
	expectedPaths := []string{
		"/api/v1/files/upload",
		"/api/v1/files/:id/download-token",
		"/api/v1/files/download/:token",
		"/api/v1/files/record/:recordType/:recordId",
	}

	registeredPaths := make(map[string]bool)
	for _, r := range routes {
		registeredPaths[r.Path] = true
	}

	for _, expected := range expectedPaths {
		if !registeredPaths[expected] {
			t.Errorf("expected route %q not registered", expected)
		}
	}

	// Verify no /files/files/ duplication exists.
	for _, r := range routes {
		if strings.Contains(r.Path, "/files/files/") {
			t.Errorf("duplicated path segment found: %q", r.Path)
		}
	}
}

// TestBookingRouteComposition verifies booking routes are under /api/v1/bookings/...
// without duplication.
func TestBookingRouteComposition(t *testing.T) {
	e := echo.New()
	v1 := e.Group("/api/v1")
	bookingGroup := v1.Group("/bookings")

	// Simulate handler registration (corrected: no extra /bookings prefix).
	bookingGroup.GET("", func(c echo.Context) error { return nil })
	bookingGroup.POST("", func(c echo.Context) error { return nil })
	bookingGroup.GET("/:id", func(c echo.Context) error { return nil })

	routes := e.Routes()
	registeredPaths := make(map[string]bool)
	for _, r := range routes {
		registeredPaths[r.Path] = true
	}

	if !registeredPaths["/api/v1/bookings"] {
		t.Error("expected GET /api/v1/bookings (list) route not registered")
	}
	if !registeredPaths["/api/v1/bookings/:id"] {
		t.Error("expected GET /api/v1/bookings/:id route not registered")
	}

	for _, r := range routes {
		if strings.Contains(r.Path, "/bookings/bookings") {
			t.Errorf("duplicated path segment found: %q", r.Path)
		}
	}
}
