//go:build integration

package apitest

import (
	"net/http"
	"testing"

	"travel-platform/apps/api/internal/common"
)

func TestGetUser_OwnProfile(t *testing.T) {
	token := tokenFor(t, TravelerUserID, TravelerEmail, []string{common.RoleTraveler})
	rec := doRequest(t, http.MethodGet, "/api/v1/users/"+TravelerUserID, "", token)
	assertStatus(t, rec, http.StatusOK)

	resp := parseJSON(t, rec)
	data := dataMap(t, resp)

	if id, ok := data["id"].(string); !ok || id != TravelerUserID {
		t.Errorf("expected id=%s, got %v", TravelerUserID, data["id"])
	}
	if email, ok := data["email"].(string); !ok || email != TravelerEmail {
		t.Errorf("expected email=%s, got %v", TravelerEmail, data["email"])
	}
}

func TestGetUser_OtherUser_Forbidden(t *testing.T) {
	token := tokenFor(t, TravelerUserID, TravelerEmail, []string{common.RoleTraveler})
	rec := doRequest(t, http.MethodGet, "/api/v1/users/"+AdminUserID, "", token)
	assertStatus(t, rec, http.StatusForbidden)
}

func TestGetUser_Admin_CanViewOthers(t *testing.T) {
	token := tokenFor(t, AdminUserID, AdminEmail, []string{common.RoleAdministrator})
	rec := doRequest(t, http.MethodGet, "/api/v1/users/"+TravelerUserID, "", token)
	assertStatus(t, rec, http.StatusOK)
}

func TestGetUser_Unauthenticated(t *testing.T) {
	rec := doRequest(t, http.MethodGet, "/api/v1/users/"+TravelerUserID, "", "")
	assertStatus(t, rec, http.StatusUnauthorized)
}

func TestUpdateProfile(t *testing.T) {
	token := tokenFor(t, TravelerUserID, TravelerEmail, []string{common.RoleTraveler})
	body := `{"displayName":"Test Name"}`
	rec := doRequest(t, http.MethodPatch, "/api/v1/users/"+TravelerUserID+"/profile", body, token)
	assertStatus(t, rec, http.StatusOK)
}

func TestUpdatePreferences(t *testing.T) {
	token := tokenFor(t, TravelerUserID, TravelerEmail, []string{common.RoleTraveler})
	body := `{"timezone":"UTC"}`
	rec := doRequest(t, http.MethodPatch, "/api/v1/users/"+TravelerUserID+"/preferences", body, token)
	assertStatus(t, rec, http.StatusOK)
}
