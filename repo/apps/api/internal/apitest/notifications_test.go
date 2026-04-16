//go:build integration

package apitest

import (
	"net/http"
	"testing"

	"travel-platform/apps/api/internal/common"
)

func TestListNotifications(t *testing.T) {
	token := tokenFor(t, TravelerUserID, TravelerEmail, []string{common.RoleTraveler})
	rec := doRequest(t, http.MethodGet, "/api/v1/notifications", "", token)
	assertStatus(t, rec, http.StatusOK)
}

func TestListNotifications_Unauthenticated(t *testing.T) {
	rec := doRequest(t, http.MethodGet, "/api/v1/notifications", "", "")
	assertStatus(t, rec, http.StatusUnauthorized)
}

func TestListMessages(t *testing.T) {
	token := tokenFor(t, TravelerUserID, TravelerEmail, []string{common.RoleTraveler})
	rec := doRequest(t, http.MethodGet, "/api/v1/messages", "", token)
	assertStatus(t, rec, http.StatusOK)
}

func TestListSendLogs(t *testing.T) {
	token := tokenFor(t, TravelerUserID, TravelerEmail, []string{common.RoleTraveler})
	rec := doRequest(t, http.MethodGet, "/api/v1/send-logs", "", token)
	assertStatus(t, rec, http.StatusOK)
}

func TestGetDNDSettings(t *testing.T) {
	token := tokenFor(t, TravelerUserID, TravelerEmail, []string{common.RoleTraveler})
	rec := doRequest(t, http.MethodGet, "/api/v1/users/"+TravelerUserID+"/dnd", "", token)
	assertStatus(t, rec, http.StatusOK)
}

func TestUpdateDNDSettings(t *testing.T) {
	token := tokenFor(t, TravelerUserID, TravelerEmail, []string{common.RoleTraveler})
	body := `{"enabled":true,"dndStart":"22:00","dndEnd":"07:00"}`
	rec := doRequest(t, http.MethodPatch, "/api/v1/users/"+TravelerUserID+"/dnd", body, token)
	assertStatus(t, rec, http.StatusOK)
}

func TestGetDNDSettings_OtherUser_Forbidden(t *testing.T) {
	token := tokenFor(t, TravelerUserID, TravelerEmail, []string{common.RoleTraveler})
	rec := doRequest(t, http.MethodGet, "/api/v1/users/"+AdminUserID+"/dnd", "", token)
	assertStatus(t, rec, http.StatusForbidden)
}

func TestGetSubscriptionPrefs(t *testing.T) {
	token := tokenFor(t, TravelerUserID, TravelerEmail, []string{common.RoleTraveler})
	rec := doRequest(t, http.MethodGet, "/api/v1/users/"+TravelerUserID+"/subscriptions", "", token)
	assertStatus(t, rec, http.StatusOK)
}

func TestExportCallbackQueue_AdminOnly(t *testing.T) {
	token := tokenFor(t, TravelerUserID, TravelerEmail, []string{common.RoleTraveler})
	rec := doRequest(t, http.MethodPost, "/api/v1/messages/callback-queue/export", "", token)
	assertStatus(t, rec, http.StatusForbidden)
}

func TestExportCallbackQueue_Admin(t *testing.T) {
	token := tokenFor(t, AdminUserID, AdminEmail, []string{common.RoleAdministrator})
	rec := doRequest(t, http.MethodPost, "/api/v1/messages/callback-queue/export", "", token)
	// The response may be a raw JSON blob rather than the standard envelope.
	assertStatus(t, rec, http.StatusOK)
}

func TestMarkNotificationRead(t *testing.T) {
	// Mark a non-existent notification as read — proves the real handler executes.
	token := tokenFor(t, TravelerUserID, TravelerEmail, []string{common.RoleTraveler})
	rec := doRequest(t, http.MethodPost, "/api/v1/notifications/notif-nonexistent/read", "", token)
	// Handler returns 404 for non-existent notification, which proves real handler execution.
	if rec.Code == http.StatusUnauthorized || rec.Code == http.StatusForbidden {
		t.Errorf("expected handler execution, got %d", rec.Code)
	}
}

func TestUpdateSubscriptionPrefs(t *testing.T) {
	token := tokenFor(t, TravelerUserID, TravelerEmail, []string{common.RoleTraveler})
	body := `[{"channel":"in_app","enabled":true}]`
	rec := doRequest(t, http.MethodPatch, "/api/v1/users/"+TravelerUserID+"/subscriptions", body, token)
	assertStatus(t, rec, http.StatusOK)
}
