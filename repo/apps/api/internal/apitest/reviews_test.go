//go:build integration

package apitest

import (
	"net/http"
	"testing"

	"travel-platform/apps/api/internal/common"
)

func TestGetReviewsBySubject(t *testing.T) {
	token := tokenFor(t, TravelerUserID, TravelerEmail, []string{common.RoleTraveler})
	rec := doRequest(t, http.MethodGet, "/api/v1/reviews/subject/"+TravelerUserID, "", token)
	assertStatus(t, rec, http.StatusOK)
}

func TestGetCreditTier(t *testing.T) {
	token := tokenFor(t, AdminUserID, AdminEmail, []string{common.RoleAdministrator})
	rec := doRequest(t, http.MethodGet, "/api/v1/credit-tiers/"+TravelerUserID, "", token)
	assertStatus(t, rec, http.StatusOK)
}

func TestSubmitReview_Unauthenticated(t *testing.T) {
	rec := doRequest(t, http.MethodPost, "/api/v1/reviews", "{}", "")
	assertStatus(t, rec, http.StatusUnauthorized)
}

func TestSubmitReview_MissingFields(t *testing.T) {
	token := tokenFor(t, TravelerUserID, TravelerEmail, []string{common.RoleTraveler})
	rec := doRequest(t, http.MethodPost, "/api/v1/reviews", "{}", token)

	// The handler should return 400 or 422 for missing required fields.
	if rec.Code != http.StatusBadRequest && rec.Code != http.StatusUnprocessableEntity {
		t.Errorf("status = %d, want 400 or 422 (body: %s)", rec.Code, rec.Body.String())
	}
}

func TestReportViolation(t *testing.T) {
	token := tokenFor(t, AdminUserID, AdminEmail, []string{common.RoleAdministrator})
	body := `{"userId":"c0000000-0000-0000-0000-000000000004","type":"test","description":"test violation"}`
	rec := doRequest(t, http.MethodPost, "/api/v1/violations", body, token)

	// Real handler was reached — not blocked by auth/role middleware.
	if rec.Code == http.StatusUnauthorized || rec.Code == http.StatusForbidden {
		t.Errorf("expected real handler execution, got status %d (body: %s)", rec.Code, rec.Body.String())
	}
}

func TestReportNoShow(t *testing.T) {
	token := tokenFor(t, AdminUserID, AdminEmail, []string{common.RoleAdministrator})
	body := `{"userId":"c0000000-0000-0000-0000-000000000004","orderId":"order-1","orderType":"booking"}`
	rec := doRequest(t, http.MethodPost, "/api/v1/no-shows", body, token)

	// Real handler was reached — not blocked by auth/role middleware.
	if rec.Code == http.StatusUnauthorized || rec.Code == http.StatusForbidden {
		t.Errorf("expected real handler execution, got status %d (body: %s)", rec.Code, rec.Body.String())
	}
}

func TestReportHarassment(t *testing.T) {
	token := tokenFor(t, AdminUserID, AdminEmail, []string{common.RoleAdministrator})
	body := `{"reportedUserId":"c0000000-0000-0000-0000-000000000004","description":"test"}`
	rec := doRequest(t, http.MethodPost, "/api/v1/harassment-flags", body, token)

	// Real handler was reached — not blocked by auth/role middleware.
	if rec.Code == http.StatusUnauthorized || rec.Code == http.StatusForbidden {
		t.Errorf("expected real handler execution, got status %d (body: %s)", rec.Code, rec.Body.String())
	}
}
