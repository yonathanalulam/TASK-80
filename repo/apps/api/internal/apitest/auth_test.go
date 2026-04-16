//go:build integration

package apitest

import (
	"net/http"
	"testing"

	"travel-platform/apps/api/internal/common"
)

func TestLogin_ValidCredentials(t *testing.T) {
	body := `{"email":"` + AdminEmail + `","password":"` + SeedPassword + `"}`
	rec := doRequest(t, http.MethodPost, "/api/v1/auth/login", body, "")
	assertStatus(t, rec, http.StatusOK)

	resp := parseJSON(t, rec)
	if !resp.Success {
		t.Fatal("expected success=true")
	}

	data := dataMap(t, resp)
	token, ok := data["token"].(string)
	if !ok || token == "" {
		t.Fatal("expected non-empty token in response data")
	}
}

func TestLogin_InvalidPassword(t *testing.T) {
	body := `{"email":"` + AdminEmail + `","password":"wrong-password"}`
	rec := doRequest(t, http.MethodPost, "/api/v1/auth/login", body, "")
	assertStatus(t, rec, http.StatusUnauthorized)

	resp := parseJSON(t, rec)
	if resp.Success {
		t.Error("expected success=false")
	}
}

func TestLogin_MissingFields(t *testing.T) {
	rec := doRequest(t, http.MethodPost, "/api/v1/auth/login", `{}`, "")
	assertStatus(t, rec, http.StatusUnprocessableEntity)

	resp := parseJSON(t, rec)
	if resp.Success {
		t.Error("expected success=false")
	}
	if resp.Error == nil || resp.Error.Code != "VALIDATION_ERROR" {
		t.Errorf("expected VALIDATION_ERROR code, got %+v", resp.Error)
	}
}

func TestLogout_Authenticated(t *testing.T) {
	token := tokenFor(t, AdminUserID, AdminEmail, []string{common.RoleAdministrator})
	rec := doRequest(t, http.MethodPost, "/api/v1/auth/logout", "", token)
	assertStatus(t, rec, http.StatusOK)

	resp := parseJSON(t, rec)
	if !resp.Success {
		t.Error("expected success=true")
	}
}

func TestMe_Authenticated(t *testing.T) {
	token := tokenFor(t, AdminUserID, AdminEmail, []string{common.RoleAdministrator})
	rec := doRequest(t, http.MethodGet, "/api/v1/auth/me", "", token)
	assertStatus(t, rec, http.StatusOK)

	resp := parseJSON(t, rec)
	if !resp.Success {
		t.Fatal("expected success=true")
	}

	data := dataMap(t, resp)

	if id, ok := data["id"].(string); !ok || id == "" {
		t.Error("expected non-empty id in me response")
	}
	if email, ok := data["email"].(string); !ok || email == "" {
		t.Error("expected non-empty email in me response")
	}
	if _, ok := data["roles"]; !ok {
		t.Error("expected roles field in me response")
	}
}

func TestMe_Unauthenticated(t *testing.T) {
	rec := doRequest(t, http.MethodGet, "/api/v1/auth/me", "", "")
	assertStatus(t, rec, http.StatusUnauthorized)

	resp := parseJSON(t, rec)
	if resp.Success {
		t.Error("expected success=false")
	}
}
