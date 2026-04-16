//go:build integration

package apitest

import (
	"net/http"
	"testing"

	"travel-platform/apps/api/internal/common"
)

func TestGetContractTemplates(t *testing.T) {
	token := tokenFor(t, AdminUserID, AdminEmail, []string{common.RoleAdministrator})
	rec := doRequest(t, http.MethodGet, "/api/v1/contract-templates", "", token)
	assertStatus(t, rec, http.StatusOK)
}

func TestGetContractTemplates_Unauthenticated(t *testing.T) {
	rec := doRequest(t, http.MethodGet, "/api/v1/contract-templates", "", "")
	assertStatus(t, rec, http.StatusUnauthorized)
}

func TestGenerateContract(t *testing.T) {
	token := tokenFor(t, AdminUserID, AdminEmail, []string{common.RoleAdministrator})
	body := `{"templateId":"test","variables":{}}`
	rec := doRequest(t, http.MethodPost, "/api/v1/contracts/generate", body, token)

	// Real handler was reached — not blocked by auth/role middleware.
	// May return 400/404/500 if template doesn't exist, which is fine.
	if rec.Code == http.StatusUnauthorized || rec.Code == http.StatusForbidden {
		t.Errorf("expected real handler execution, got status %d (body: %s)", rec.Code, rec.Body.String())
	}
}

func TestListInvoiceRequests(t *testing.T) {
	token := tokenFor(t, AccountantUserID, AccountantEmail, []string{common.RoleAccountant})
	rec := doRequest(t, http.MethodGet, "/api/v1/invoice-requests", "", token)
	assertStatus(t, rec, http.StatusOK)
}

func TestCreateInvoiceRequest_Unauthenticated(t *testing.T) {
	rec := doRequest(t, http.MethodPost, "/api/v1/invoice-requests", "{}", "")
	assertStatus(t, rec, http.StatusUnauthorized)
}

func TestCreateInvoiceRequest_Authenticated(t *testing.T) {
	token := tokenFor(t, OrganizerUserID, OrganizerEmail, []string{common.RoleGroupOrganizer})
	body := `{"orderType":"booking","orderId":"` + SeedBookingID + `","notes":"test invoice"}`
	rec := doRequest(t, http.MethodPost, "/api/v1/invoice-requests", body, token)
	// Real handler reached — may return 400/500 for incomplete data.
	if rec.Code == http.StatusUnauthorized || rec.Code == http.StatusForbidden {
		t.Errorf("expected handler execution, got %d (body: %s)", rec.Code, rec.Body.String())
	}
}

func TestApproveInvoiceRequest_Accountant(t *testing.T) {
	token := tokenFor(t, AccountantUserID, AccountantEmail, []string{common.RoleAccountant})
	rec := doRequest(t, http.MethodPost, "/api/v1/invoice-requests/nonexistent/approve", "", token)
	if rec.Code == http.StatusUnauthorized || rec.Code == http.StatusForbidden {
		t.Errorf("expected handler execution, got %d (body: %s)", rec.Code, rec.Body.String())
	}
}

func TestGenerateInvoice_Accountant(t *testing.T) {
	token := tokenFor(t, AccountantUserID, AccountantEmail, []string{common.RoleAccountant})
	rec := doRequest(t, http.MethodPost, "/api/v1/invoices/nonexistent/generate", "", token)
	if rec.Code == http.StatusUnauthorized || rec.Code == http.StatusForbidden {
		t.Errorf("expected handler execution, got %d (body: %s)", rec.Code, rec.Body.String())
	}
}
