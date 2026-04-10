package procurement

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"travel-platform/apps/api/internal/common"
	"travel-platform/apps/api/internal/middleware"

	"github.com/labstack/echo/v4"
)

func TestRFQCreation_GroupOrganizerAllowed(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/rfqs", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set(common.ContextKeyRoles, []string{common.RoleGroupOrganizer})

	handler := middleware.RequireRole(common.RoleAccountant, common.RoleAdministrator, common.RoleGroupOrganizer)(func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	err := handler(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("group_organizer should be allowed to create RFQ, got %d", rec.Code)
	}
}

func TestRFQCreation_TravelerDenied(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/rfqs", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set(common.ContextKeyRoles, []string{common.RoleTraveler})

	handler := middleware.RequireRole(common.RoleAccountant, common.RoleAdministrator, common.RoleGroupOrganizer)(func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	_ = handler(c)
	if rec.Code != http.StatusForbidden {
		t.Errorf("traveler should be denied RFQ creation, got %d", rec.Code)
	}
}

func TestRFQCreation_AccountantAllowed(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/rfqs", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set(common.ContextKeyRoles, []string{common.RoleAccountant})

	handler := middleware.RequireRole(common.RoleAccountant, common.RoleAdministrator, common.RoleGroupOrganizer)(func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	err := handler(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("accountant should be allowed to create RFQ, got %d", rec.Code)
	}
}
