package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"travel-platform/apps/api/internal/common"

	"github.com/labstack/echo/v4"
)

func TestRequireRole_Allowed(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set(common.ContextKeyRoles, []string{common.RoleAdministrator})

	handler := RequireRole(common.RoleAdministrator)(func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	err := handler(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestRequireRole_Denied_WrongRole(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set(common.ContextKeyRoles, []string{common.RoleTraveler})

	handler := RequireRole(common.RoleAdministrator)(func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	_ = handler(c)
	// The middleware returns a JSON error response with 403
	if rec.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", rec.Code)
	}
}

func TestRequireRole_Denied_NoRoles(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	// No roles set in context

	handler := RequireRole(common.RoleAdministrator)(func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	_ = handler(c)
	if rec.Code != http.StatusForbidden {
		t.Errorf("expected 403 for empty roles, got %d", rec.Code)
	}
}

func TestRequireRole_MultipleAllowed(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set(common.ContextKeyRoles, []string{common.RoleAccountant})

	// Route allows both administrator and accountant
	handler := RequireRole(common.RoleAdministrator, common.RoleAccountant)(func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	err := handler(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("accountant should be allowed, got %d", rec.Code)
	}
}

func TestRequireRole_AdminLiteral_Denied(t *testing.T) {
	// "admin" is NOT a valid role - must be "administrator"
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set(common.ContextKeyRoles, []string{"admin"}) // wrong role string

	handler := RequireRole(common.RoleAdministrator)(func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	_ = handler(c)
	if rec.Code != http.StatusForbidden {
		t.Errorf("'admin' should not match 'administrator', got %d", rec.Code)
	}
}
