package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"travel-platform/apps/api/internal/auth"
	"travel-platform/apps/api/internal/common"

	"github.com/labstack/echo/v4"
)

func TestJWTAuth_MissingToken(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/bookings", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Create a real auth service for token validation
	svc := auth.NewService(nil, nil, "test-secret")
	mw := JWTAuth(svc, []string{"/health"})

	handler := mw(func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	_ = handler(c)
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 for missing token, got %d", rec.Code)
	}
}

func TestJWTAuth_PublicPath(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	svc := auth.NewService(nil, nil, "test-secret")
	mw := JWTAuth(svc, []string{"/health"})

	handler := mw(func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	err := handler(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200 for public path, got %d", rec.Code)
	}
}

func TestJWTAuth_ValidToken(t *testing.T) {
	e := echo.New()

	// Generate a valid token
	svc := auth.NewService(nil, nil, "test-secret")
	token, err := svc.GenerateTestToken("user-1", "test@test.com", []string{common.RoleAdministrator})
	if err != nil {
		t.Fatalf("failed to generate test token: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/bookings", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	mw := JWTAuth(svc, []string{"/health"})

	handler := mw(func(c echo.Context) error {
		uid := common.GetUserID(c)
		if uid != "user-1" {
			t.Errorf("expected user-1, got %s", uid)
		}
		return c.String(http.StatusOK, "ok")
	})

	err = handler(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}
