//go:build integration

package apitest

import (
	"net/http"
	"testing"

	"travel-platform/apps/api/internal/common"
)

func TestGetRiskSummary_Admin(t *testing.T) {
	token := tokenFor(t, AdminUserID, AdminEmail, []string{common.RoleAdministrator})
	rec := doRequest(t, http.MethodGet, "/api/v1/risk/"+TravelerUserID, "", token)
	assertStatus(t, rec, http.StatusOK)
}

func TestGetRiskSummary_NonAdmin(t *testing.T) {
	token := tokenFor(t, TravelerUserID, TravelerEmail, []string{common.RoleTraveler})
	rec := doRequest(t, http.MethodGet, "/api/v1/risk/"+TravelerUserID, "", token)
	assertStatus(t, rec, http.StatusForbidden)
}

func TestGetRiskSummary_Unauthenticated(t *testing.T) {
	rec := doRequest(t, http.MethodGet, "/api/v1/risk/"+TravelerUserID, "", "")
	assertStatus(t, rec, http.StatusUnauthorized)
}

func TestGetPendingApprovals_Admin(t *testing.T) {
	token := tokenFor(t, AdminUserID, AdminEmail, []string{common.RoleAdministrator})
	rec := doRequest(t, http.MethodGet, "/api/v1/admin/approvals", "", token)
	assertStatus(t, rec, http.StatusOK)
}

func TestBlacklistUser_Admin(t *testing.T) {
	token := tokenFor(t, AdminUserID, AdminEmail, []string{common.RoleAdministrator})
	body := `{"reason":"test"}`
	rec := doRequest(t, http.MethodPost, "/api/v1/admin/users/"+TravelerUserID+"/blacklist", body, token)
	// The real handler should execute (not 401/403). It may return 200/201 or
	// an application-level error, but auth/authz should pass.
	if rec.Code == http.StatusUnauthorized || rec.Code == http.StatusForbidden {
		t.Errorf("expected admin to pass auth, got %d (body: %s)", rec.Code, rec.Body.String())
	}
}

func TestUnblacklistUser_Admin(t *testing.T) {
	token := tokenFor(t, AdminUserID, AdminEmail, []string{common.RoleAdministrator})
	body := `{"reason":"test"}`
	rec := doRequest(t, http.MethodPost, "/api/v1/admin/users/"+TravelerUserID+"/unblacklist", body, token)
	// The real handler should execute (not 401/403).
	if rec.Code == http.StatusUnauthorized || rec.Code == http.StatusForbidden {
		t.Errorf("expected admin to pass auth, got %d (body: %s)", rec.Code, rec.Body.String())
	}
}

func TestResolveApproval_Admin(t *testing.T) {
	token := tokenFor(t, AdminUserID, AdminEmail, []string{common.RoleAdministrator})
	body := `{"status":"approved","notes":"test resolution"}`
	rec := doRequest(t, http.MethodPost, "/api/v1/admin/approvals/nonexistent/resolve", body, token)
	if rec.Code == http.StatusUnauthorized || rec.Code == http.StatusForbidden {
		t.Errorf("expected handler execution, got %d (body: %s)", rec.Code, rec.Body.String())
	}
}
