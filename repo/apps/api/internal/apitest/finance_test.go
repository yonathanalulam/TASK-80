//go:build integration

package apitest

import (
	"net/http"
	"testing"

	"travel-platform/apps/api/internal/common"
)

func TestGetWallet(t *testing.T) {
	// Wallet service checks ownerID == requestingUserID, so we use the supplier
	// requesting their own wallet.
	token := tokenFor(t, SupplierUserID, SupplierEmail, []string{common.RoleSupplier})
	rec := doRequest(t, http.MethodGet, "/api/v1/wallets/"+SupplierUserID, "", token)
	assertStatus(t, rec, http.StatusOK)
}

func TestGetWallet_Unauthenticated(t *testing.T) {
	rec := doRequest(t, http.MethodGet, "/api/v1/wallets/"+SupplierUserID, "", "")
	assertStatus(t, rec, http.StatusUnauthorized)
}

func TestGetWalletTransactions(t *testing.T) {
	// Same ownership check applies: supplier requesting own wallet transactions.
	token := tokenFor(t, SupplierUserID, SupplierEmail, []string{common.RoleSupplier})
	rec := doRequest(t, http.MethodGet, "/api/v1/wallets/"+SupplierUserID+"/transactions", "", token)
	assertStatus(t, rec, http.StatusOK)
}

func TestGetEscrows(t *testing.T) {
	// Escrow handler allows owner or admin/accountant. Use admin here.
	token := tokenFor(t, AdminUserID, AdminEmail, []string{common.RoleAdministrator})
	rec := doRequest(t, http.MethodGet, "/api/v1/escrows/"+SupplierUserID, "", token)
	assertStatus(t, rec, http.StatusOK)
}

func TestReconciliation_Admin(t *testing.T) {
	token := tokenFor(t, AdminUserID, AdminEmail, []string{common.RoleAdministrator})
	rec := doRequest(t, http.MethodGet, "/api/v1/reconciliation", "", token)
	assertStatus(t, rec, http.StatusOK)
}

func TestReconciliation_NonAdmin(t *testing.T) {
	token := tokenFor(t, TravelerUserID, TravelerEmail, []string{common.RoleTraveler})
	rec := doRequest(t, http.MethodGet, "/api/v1/reconciliation", "", token)
	assertStatus(t, rec, http.StatusForbidden)
}

func TestRequestRefund_Unauthenticated(t *testing.T) {
	rec := doRequest(t, http.MethodPost, "/api/v1/refunds", "{}", "")
	assertStatus(t, rec, http.StatusUnauthorized)
}

func TestRequestRefund_WrongRole(t *testing.T) {
	token := tokenFor(t, TravelerUserID, TravelerEmail, []string{common.RoleTraveler})
	rec := doRequest(t, http.MethodPost, "/api/v1/refunds", "{}", token)
	assertStatus(t, rec, http.StatusForbidden)
}

func TestRequestWithdrawal_WrongRole(t *testing.T) {
	token := tokenFor(t, TravelerUserID, TravelerEmail, []string{common.RoleTraveler})
	rec := doRequest(t, http.MethodPost, "/api/v1/withdrawals", "{}", token)
	assertStatus(t, rec, http.StatusForbidden)
}

func TestProcessRefund_Accountant(t *testing.T) {
	token := tokenFor(t, AccountantUserID, AccountantEmail, []string{common.RoleAccountant})
	body := `{"orderType":"booking","orderId":"nonexistent","amount":10,"reason":"test"}`
	rec := doRequest(t, http.MethodPost, "/api/v1/refunds", body, token)
	// Real handler reached — may return 404/500 for non-existent order.
	if rec.Code == http.StatusUnauthorized || rec.Code == http.StatusForbidden {
		t.Errorf("expected handler execution, got %d (body: %s)", rec.Code, rec.Body.String())
	}
}

func TestRequestWithdrawal_Courier(t *testing.T) {
	token := tokenFor(t, CourierUserID, CourierEmail, []string{common.RoleCourierRunner})
	body := `{"amount":10}`
	rec := doRequest(t, http.MethodPost, "/api/v1/withdrawals", body, token)
	// Real handler reached.
	if rec.Code == http.StatusUnauthorized || rec.Code == http.StatusForbidden {
		t.Errorf("expected handler execution, got %d (body: %s)", rec.Code, rec.Body.String())
	}
}

func TestRecordTender_Accountant(t *testing.T) {
	token := tokenFor(t, AccountantUserID, AccountantEmail, []string{common.RoleAccountant})
	body := `{"orderId":"nonexistent","orderType":"booking","amount":50,"tenderType":"cash","reference":"ref-1"}`
	rec := doRequest(t, http.MethodPost, "/api/v1/payments/record-tender", body, token)
	if rec.Code == http.StatusUnauthorized || rec.Code == http.StatusForbidden {
		t.Errorf("expected handler execution, got %d (body: %s)", rec.Code, rec.Body.String())
	}
}

func TestReleaseEscrow_Accountant(t *testing.T) {
	token := tokenFor(t, AccountantUserID, AccountantEmail, []string{common.RoleAccountant})
	body := `{"orderType":"booking","amount":50}`
	rec := doRequest(t, http.MethodPost, "/api/v1/settlements/nonexistent/release", body, token)
	if rec.Code == http.StatusUnauthorized || rec.Code == http.StatusForbidden {
		t.Errorf("expected handler execution, got %d (body: %s)", rec.Code, rec.Body.String())
	}
}

func TestApproveWithdrawal_Accountant(t *testing.T) {
	token := tokenFor(t, AccountantUserID, AccountantEmail, []string{common.RoleAccountant})
	rec := doRequest(t, http.MethodPost, "/api/v1/withdrawals/nonexistent/approve", "", token)
	if rec.Code == http.StatusUnauthorized || rec.Code == http.StatusForbidden {
		t.Errorf("expected handler execution, got %d (body: %s)", rec.Code, rec.Body.String())
	}
}

func TestRejectWithdrawal_Accountant(t *testing.T) {
	token := tokenFor(t, AccountantUserID, AccountantEmail, []string{common.RoleAccountant})
	body := `{"reason":"test rejection"}`
	rec := doRequest(t, http.MethodPost, "/api/v1/withdrawals/nonexistent/reject", body, token)
	if rec.Code == http.StatusUnauthorized || rec.Code == http.StatusForbidden {
		t.Errorf("expected handler execution, got %d (body: %s)", rec.Code, rec.Body.String())
	}
}
