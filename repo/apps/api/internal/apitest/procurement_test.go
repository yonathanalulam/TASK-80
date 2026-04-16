//go:build integration

package apitest

import (
	"net/http"
	"testing"

	"travel-platform/apps/api/internal/common"
)

func TestListRFQs(t *testing.T) {
	token := tokenFor(t, OrganizerUserID, OrganizerEmail, []string{common.RoleGroupOrganizer})
	rec := doRequest(t, http.MethodGet, "/api/v1/rfqs", "", token)
	assertStatus(t, rec, http.StatusOK)
}

func TestListRFQs_Unauthenticated(t *testing.T) {
	rec := doRequest(t, http.MethodGet, "/api/v1/rfqs", "", "")
	assertStatus(t, rec, http.StatusUnauthorized)
}

func TestCreateRFQ(t *testing.T) {
	token := tokenFor(t, OrganizerUserID, OrganizerEmail, []string{common.RoleGroupOrganizer})
	body := `{"title":"Test RFQ","deadline":"2026-12-01","description":"Test","items":[{"description":"Item 1","quantity":10}]}`
	rec := doRequest(t, http.MethodPost, "/api/v1/rfqs", body, token)
	assertStatus(t, rec, http.StatusCreated)
}

func TestCreateRFQ_WrongRole(t *testing.T) {
	token := tokenFor(t, TravelerUserID, TravelerEmail, []string{common.RoleTraveler})
	body := `{"title":"Test RFQ","deadline":"2026-12-01","description":"Test","items":[{"description":"Item 1","quantity":10}]}`
	rec := doRequest(t, http.MethodPost, "/api/v1/rfqs", body, token)
	assertStatus(t, rec, http.StatusForbidden)
}

func TestListSupplierQuotes(t *testing.T) {
	token := tokenFor(t, SupplierUserID, SupplierEmail, []string{common.RoleSupplier})
	rec := doRequest(t, http.MethodGet, "/api/v1/supplier-quotes", "", token)
	assertStatus(t, rec, http.StatusOK)
}

func TestListPurchaseOrders(t *testing.T) {
	token := tokenFor(t, AdminUserID, AdminEmail, []string{common.RoleAdministrator})
	rec := doRequest(t, http.MethodGet, "/api/v1/purchase-orders", "", token)
	assertStatus(t, rec, http.StatusOK)
}

func TestListDeliveries(t *testing.T) {
	token := tokenFor(t, AdminUserID, AdminEmail, []string{common.RoleAdministrator})
	rec := doRequest(t, http.MethodGet, "/api/v1/deliveries", "", token)
	assertStatus(t, rec, http.StatusOK)
}

func TestListExceptions(t *testing.T) {
	token := tokenFor(t, AdminUserID, AdminEmail, []string{common.RoleAdministrator})
	rec := doRequest(t, http.MethodGet, "/api/v1/exceptions", "", token)
	assertStatus(t, rec, http.StatusOK)
}

func TestGetRFQById(t *testing.T) {
	// First create an RFQ, then fetch it by ID.
	token := tokenFor(t, OrganizerUserID, OrganizerEmail, []string{common.RoleGroupOrganizer})
	createBody := `{"title":"Get Test RFQ","deadline":"2026-12-15","description":"Test","items":[{"description":"Item A","quantity":5}]}`
	createRec := doRequest(t, http.MethodPost, "/api/v1/rfqs", createBody, token)
	if createRec.Code != http.StatusCreated {
		t.Fatalf("setup: create RFQ failed: %d", createRec.Code)
	}
	createResp := parseJSON(t, createRec)
	dm := dataMap(t, createResp)
	rfqID, _ := dm["id"].(string)

	rec := doRequest(t, http.MethodGet, "/api/v1/rfqs/"+rfqID, "", token)
	assertStatus(t, rec, http.StatusOK)
}

func TestIssueRFQ(t *testing.T) {
	// Create, then issue.
	orgToken := tokenFor(t, OrganizerUserID, OrganizerEmail, []string{common.RoleGroupOrganizer})
	createBody := `{"title":"Issue Test RFQ","deadline":"2026-12-20","description":"Issue test","items":[{"description":"Item B","quantity":3}]}`
	createRec := doRequest(t, http.MethodPost, "/api/v1/rfqs", createBody, orgToken)
	if createRec.Code != http.StatusCreated {
		t.Fatalf("setup: create RFQ failed: %d", createRec.Code)
	}
	createResp := parseJSON(t, createRec)
	dm := dataMap(t, createResp)
	rfqID, _ := dm["id"].(string)

	acctToken := tokenFor(t, AccountantUserID, AccountantEmail, []string{common.RoleAccountant})
	body := `{"supplierIds":["` + SupplierUserID + `"]}`
	rec := doRequest(t, http.MethodPost, "/api/v1/rfqs/"+rfqID+"/issue", body, acctToken)
	if rec.Code == http.StatusUnauthorized || rec.Code == http.StatusForbidden {
		t.Errorf("expected handler execution, got %d (body: %s)", rec.Code, rec.Body.String())
	}
}

func TestGetRFQComparison(t *testing.T) {
	token := tokenFor(t, AdminUserID, AdminEmail, []string{common.RoleAdministrator})
	rec := doRequest(t, http.MethodGet, "/api/v1/rfqs/nonexistent/comparison", "", token)
	if rec.Code == http.StatusUnauthorized || rec.Code == http.StatusForbidden {
		t.Errorf("expected handler execution, got %d (body: %s)", rec.Code, rec.Body.String())
	}
}

func TestCreatePurchaseOrder(t *testing.T) {
	token := tokenFor(t, AccountantUserID, AccountantEmail, []string{common.RoleAccountant})
	body := `{"quoteId":"nonexistent","rfqId":"nonexistent"}`
	rec := doRequest(t, http.MethodPost, "/api/v1/purchase-orders", body, token)
	if rec.Code == http.StatusUnauthorized || rec.Code == http.StatusForbidden {
		t.Errorf("expected handler execution, got %d (body: %s)", rec.Code, rec.Body.String())
	}
}

func TestGetPurchaseOrderById(t *testing.T) {
	token := tokenFor(t, AdminUserID, AdminEmail, []string{common.RoleAdministrator})
	rec := doRequest(t, http.MethodGet, "/api/v1/purchase-orders/nonexistent", "", token)
	if rec.Code == http.StatusUnauthorized || rec.Code == http.StatusForbidden {
		t.Errorf("expected handler execution, got %d (body: %s)", rec.Code, rec.Body.String())
	}
}

func TestAcceptPurchaseOrder(t *testing.T) {
	token := tokenFor(t, SupplierUserID, SupplierEmail, []string{common.RoleSupplier})
	rec := doRequest(t, http.MethodPost, "/api/v1/purchase-orders/nonexistent/accept", "", token)
	if rec.Code == http.StatusUnauthorized || rec.Code == http.StatusForbidden {
		t.Errorf("expected handler execution, got %d (body: %s)", rec.Code, rec.Body.String())
	}
}

func TestCreateDelivery(t *testing.T) {
	token := tokenFor(t, CourierUserID, CourierEmail, []string{common.RoleCourierRunner})
	body := `{"notes":"delivery test"}`
	rec := doRequest(t, http.MethodPost, "/api/v1/purchase-orders/nonexistent/deliveries", body, token)
	if rec.Code == http.StatusUnauthorized || rec.Code == http.StatusForbidden {
		t.Errorf("expected handler execution, got %d (body: %s)", rec.Code, rec.Body.String())
	}
}

func TestInspectDelivery(t *testing.T) {
	token := tokenFor(t, AdminUserID, AdminEmail, []string{common.RoleAdministrator})
	body := `{"passed":true,"notes":"looks good"}`
	rec := doRequest(t, http.MethodPost, "/api/v1/deliveries/nonexistent/inspect", body, token)
	if rec.Code == http.StatusUnauthorized || rec.Code == http.StatusForbidden {
		t.Errorf("expected handler execution, got %d (body: %s)", rec.Code, rec.Body.String())
	}
}

func TestCreateDiscrepancy(t *testing.T) {
	token := tokenFor(t, AdminUserID, AdminEmail, []string{common.RoleAdministrator})
	body := `{"deliveryId":"nonexistent","description":"test discrepancy"}`
	rec := doRequest(t, http.MethodPost, "/api/v1/discrepancies", body, token)
	if rec.Code == http.StatusUnauthorized || rec.Code == http.StatusForbidden {
		t.Errorf("expected handler execution, got %d (body: %s)", rec.Code, rec.Body.String())
	}
}

func TestExceptionWaiver(t *testing.T) {
	token := tokenFor(t, AdminUserID, AdminEmail, []string{common.RoleAdministrator})
	body := `{"reason":"test waiver"}`
	rec := doRequest(t, http.MethodPost, "/api/v1/exceptions/nonexistent/waivers", body, token)
	if rec.Code == http.StatusUnauthorized || rec.Code == http.StatusForbidden {
		t.Errorf("expected handler execution, got %d (body: %s)", rec.Code, rec.Body.String())
	}
}

func TestExceptionSettlementAdjustment(t *testing.T) {
	token := tokenFor(t, AdminUserID, AdminEmail, []string{common.RoleAdministrator})
	body := `{"amount":50,"reason":"test adjustment"}`
	rec := doRequest(t, http.MethodPost, "/api/v1/exceptions/nonexistent/settlement-adjustments", body, token)
	if rec.Code == http.StatusUnauthorized || rec.Code == http.StatusForbidden {
		t.Errorf("expected handler execution, got %d (body: %s)", rec.Code, rec.Body.String())
	}
}

func TestCloseException(t *testing.T) {
	token := tokenFor(t, AdminUserID, AdminEmail, []string{common.RoleAdministrator})
	body := `{"resolution":"resolved via test"}`
	rec := doRequest(t, http.MethodPost, "/api/v1/exceptions/nonexistent/close", body, token)
	if rec.Code == http.StatusUnauthorized || rec.Code == http.StatusForbidden {
		t.Errorf("expected handler execution, got %d (body: %s)", rec.Code, rec.Body.String())
	}
}

func TestSubmitQuote(t *testing.T) {
	token := tokenFor(t, SupplierUserID, SupplierEmail, []string{common.RoleSupplier})
	body := `{"totalAmount":5000,"leadTimeDays":7,"notes":"test quote"}`
	rec := doRequest(t, http.MethodPost, "/api/v1/rfqs/nonexistent/quotes", body, token)
	if rec.Code == http.StatusUnauthorized || rec.Code == http.StatusForbidden {
		t.Errorf("expected handler execution, got %d (body: %s)", rec.Code, rec.Body.String())
	}
}

func TestSelectSupplier(t *testing.T) {
	token := tokenFor(t, AccountantUserID, AccountantEmail, []string{common.RoleAccountant})
	body := `{"quoteId":"nonexistent"}`
	rec := doRequest(t, http.MethodPost, "/api/v1/rfqs/nonexistent/select", body, token)
	if rec.Code == http.StatusUnauthorized || rec.Code == http.StatusForbidden {
		t.Errorf("expected handler execution, got %d (body: %s)", rec.Code, rec.Body.String())
	}
}
