//go:build integration

package apitest

import (
	"net/http"
	"testing"

	"travel-platform/apps/api/internal/common"
)

func TestFileUpload_Unauthenticated(t *testing.T) {
	rec := doRequest(t, http.MethodPost, "/api/v1/files/upload", "", "")
	assertStatus(t, rec, http.StatusUnauthorized)
}

func TestFileDownloadToken_Unauthenticated(t *testing.T) {
	rec := doRequest(t, http.MethodPost, "/api/v1/files/file-123/download-token", "", "")
	assertStatus(t, rec, http.StatusUnauthorized)
}

func TestFileRecordList_Unauthenticated(t *testing.T) {
	rec := doRequest(t, http.MethodGet, "/api/v1/files/record/booking/b-1", "", "")
	assertStatus(t, rec, http.StatusUnauthorized)
}

func TestFileRecordList_Authenticated(t *testing.T) {
	token := tokenFor(t, OrganizerUserID, OrganizerEmail, []string{common.RoleGroupOrganizer})
	rec := doRequest(t, http.MethodGet, "/api/v1/files/record/booking/"+SeedBookingID, "", token)
	assertStatus(t, rec, http.StatusOK)
}

func TestFileDownload_PublicPath(t *testing.T) {
	// The download path is public (no auth required).
	// With a nonexistent token, the handler should still execute (not return 401).
	rec := doRequest(t, http.MethodGet, "/api/v1/files/download/nonexistent-token", "", "")
	if rec.Code == http.StatusUnauthorized {
		t.Errorf("expected public endpoint to not return 401, got %d (body: %s)", rec.Code, rec.Body.String())
	}
}

func TestFileUpload_Authenticated(t *testing.T) {
	// Authenticated upload reaches the real handler (may fail with 400/500 without
	// a proper multipart body, which is fine — proves handler is wired).
	token := tokenFor(t, OrganizerUserID, OrganizerEmail, []string{common.RoleGroupOrganizer})
	rec := doRequest(t, http.MethodPost, "/api/v1/files/upload", "{}", token)
	if rec.Code == http.StatusUnauthorized || rec.Code == http.StatusForbidden {
		t.Errorf("expected handler execution, got %d", rec.Code)
	}
}

func TestFileDownloadToken_Authenticated(t *testing.T) {
	// Authenticated token creation reaches the real handler.
	token := tokenFor(t, OrganizerUserID, OrganizerEmail, []string{common.RoleGroupOrganizer})
	rec := doRequest(t, http.MethodPost, "/api/v1/files/file-nonexistent/download-token",
		`{"ttlSeconds":300,"singleUse":true}`, token)
	// May return 404/500 (file not found), but NOT 401/403.
	if rec.Code == http.StatusUnauthorized || rec.Code == http.StatusForbidden {
		t.Errorf("expected handler execution, got %d", rec.Code)
	}
}
