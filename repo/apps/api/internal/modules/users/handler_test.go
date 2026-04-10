package users

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"travel-platform/apps/api/internal/common"

	"github.com/labstack/echo/v4"
)

func TestGetUser_CrossUser_Denied(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/users/user-2", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("user-2")
	c.Set(common.ContextKeyUserID, "user-1")
	c.Set(common.ContextKeyRoles, []string{common.RoleTraveler})

	h := &Handler{svc: nil}
	_ = h.GetUser(c)

	if rec.Code != http.StatusForbidden {
		t.Errorf("cross-user access by traveler should be 403, got %d", rec.Code)
	}
}

func TestGetUser_CrossUser_SupplierDenied(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/users/user-2", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("user-2")
	c.Set(common.ContextKeyUserID, "user-1")
	c.Set(common.ContextKeyRoles, []string{common.RoleSupplier})

	h := &Handler{svc: nil}
	_ = h.GetUser(c)

	if rec.Code != http.StatusForbidden {
		t.Errorf("cross-user access by supplier should be 403, got %d", rec.Code)
	}
}

func TestGetUser_AdminLiteral_Denied(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/users/user-2", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("user-2")
	c.Set(common.ContextKeyUserID, "user-1")
	c.Set(common.ContextKeyRoles, []string{"admin"})

	h := &Handler{svc: nil}
	_ = h.GetUser(c)

	if rec.Code != http.StatusForbidden {
		t.Errorf("'admin' (not 'administrator') should be denied, got %d", rec.Code)
	}
}

func TestGetUser_AuthzCheck_PrecedesServiceCall(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/users/target-user", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("target-user")
	c.Set(common.ContextKeyUserID, "other-user")
	c.Set(common.ContextKeyRoles, []string{common.RoleTraveler})

	h := &Handler{svc: nil}
	_ = h.GetUser(c)

	if rec.Code != http.StatusForbidden {
		t.Errorf("authz should reject before reaching service (nil svc would panic otherwise), got %d", rec.Code)
	}
}
