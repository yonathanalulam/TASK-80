//go:build integration

package apitest

import (
	"net/http"
	"testing"

	"travel-platform/apps/api/internal/common"
)

func TestAdminListUsers(t *testing.T) {
	token := tokenFor(t, AdminUserID, AdminEmail, []string{common.RoleAdministrator})
	rec := doRequest(t, http.MethodGet, "/api/v1/admin/users", "", token)
	assertStatus(t, rec, http.StatusOK)

	resp := parseJSON(t, rec)
	data := dataMap(t, resp)

	if _, ok := data["items"]; !ok {
		t.Error("expected data to contain 'items'")
	}
	if _, ok := data["total"]; !ok {
		t.Error("expected data to contain 'total'")
	}
}

func TestAdminListUsers_NonAdmin_Forbidden(t *testing.T) {
	token := tokenFor(t, TravelerUserID, TravelerEmail, []string{common.RoleTraveler})
	rec := doRequest(t, http.MethodGet, "/api/v1/admin/users", "", token)
	assertStatus(t, rec, http.StatusForbidden)
}

func TestAdminGetAuditLogs(t *testing.T) {
	token := tokenFor(t, AdminUserID, AdminEmail, []string{common.RoleAdministrator})
	rec := doRequest(t, http.MethodGet, "/api/v1/admin/audit-logs", "", token)
	assertStatus(t, rec, http.StatusOK)

	resp := parseJSON(t, rec)
	data := dataMap(t, resp)

	if _, ok := data["items"]; !ok {
		t.Error("expected data to contain 'items'")
	}
	if _, ok := data["total"]; !ok {
		t.Error("expected data to contain 'total'")
	}
}

func TestAdminGetAuditLogs_NonAdmin_Forbidden(t *testing.T) {
	token := tokenFor(t, TravelerUserID, TravelerEmail, []string{common.RoleTraveler})
	rec := doRequest(t, http.MethodGet, "/api/v1/admin/audit-logs", "", token)
	assertStatus(t, rec, http.StatusForbidden)
}

func TestAdminGetSendLogs(t *testing.T) {
	token := tokenFor(t, AdminUserID, AdminEmail, []string{common.RoleAdministrator})
	rec := doRequest(t, http.MethodGet, "/api/v1/admin/send-logs", "", token)
	assertStatus(t, rec, http.StatusOK)
}

func TestAdminGetConfig(t *testing.T) {
	token := tokenFor(t, AdminUserID, AdminEmail, []string{common.RoleAdministrator})
	rec := doRequest(t, http.MethodGet, "/api/v1/admin/config", "", token)
	assertStatus(t, rec, http.StatusOK)

	resp := parseJSON(t, rec)
	data := dataMap(t, resp)

	expectedKeys := []string{
		"dnd_default_start",
		"dnd_default_end",
		"courier_daily_cap",
		"refund_minimum_unit",
		"download_token_ttl_min",
		"max_cancellations_24h",
		"max_rfqs_10min",
		"coupon_max_threshold",
		"coupon_max_percentage",
		"new_user_gift_exclusive",
	}
	for _, key := range expectedKeys {
		if _, ok := data[key]; !ok {
			t.Errorf("expected config to contain key %q", key)
		}
	}
}

func TestAdminGetConfig_NonAdmin_Forbidden(t *testing.T) {
	token := tokenFor(t, TravelerUserID, TravelerEmail, []string{common.RoleTraveler})
	rec := doRequest(t, http.MethodGet, "/api/v1/admin/config", "", token)
	assertStatus(t, rec, http.StatusForbidden)
}
