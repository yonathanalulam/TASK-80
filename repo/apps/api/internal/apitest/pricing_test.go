//go:build integration

package apitest

import (
	"net/http"
	"testing"

	"travel-platform/apps/api/internal/common"
)

func TestListActiveCoupons(t *testing.T) {
	token := tokenFor(t, TravelerUserID, TravelerEmail, []string{common.RoleTraveler})
	rec := doRequest(t, http.MethodGet, "/api/v1/coupons/coupons/available", "", token)
	assertStatus(t, rec, http.StatusOK)

	resp := parseJSON(t, rec)
	if !resp.Success {
		t.Fatal("expected success=true")
	}
	// Seed data includes 3 coupons (SAVE25, LODGE10, WELCOME15).
	arr, ok := resp.Data.([]interface{})
	if !ok {
		t.Fatal("expected data to be array")
	}
	if len(arr) < 1 {
		t.Error("expected at least one active coupon from seed data")
	}
}

func TestListActiveCoupons_Unauthenticated(t *testing.T) {
	rec := doRequest(t, http.MethodGet, "/api/v1/coupons/coupons/available", "", "")
	assertStatus(t, rec, http.StatusUnauthorized)
}

func TestEvaluateCoupons(t *testing.T) {
	token := tokenFor(t, OrganizerUserID, OrganizerEmail, []string{common.RoleGroupOrganizer})
	body := `{
		"couponCodes": ["SAVE25"],
		"items": [{"itemName":"Hotel","unitPrice":300,"quantity":1,"category":"lodging","itemType":"lodging"}],
		"isNewUser": false,
		"membershipTier": ""
	}`
	rec := doRequest(t, http.MethodPost, "/api/v1/coupons/coupons/evaluate", body, token)
	assertStatus(t, rec, http.StatusOK)

	resp := parseJSON(t, rec)
	if !resp.Success {
		t.Fatal("expected success=true")
	}
}

func TestEvaluateCoupons_MissingCodes(t *testing.T) {
	token := tokenFor(t, OrganizerUserID, OrganizerEmail, []string{common.RoleGroupOrganizer})
	body := `{
		"couponCodes": [],
		"items": [{"itemName":"Hotel","unitPrice":300,"quantity":1,"category":"lodging","itemType":"lodging"}]
	}`
	rec := doRequest(t, http.MethodPost, "/api/v1/coupons/coupons/evaluate", body, token)
	assertStatus(t, rec, http.StatusUnprocessableEntity)
}

func TestRedeemPreview(t *testing.T) {
	token := tokenFor(t, OrganizerUserID, OrganizerEmail, []string{common.RoleGroupOrganizer})
	body := `{
		"couponCodes": ["SAVE25"],
		"items": [{"itemName":"Hotel","unitPrice":300,"quantity":1,"category":"lodging","itemType":"lodging"}],
		"isNewUser": false
	}`
	rec := doRequest(t, http.MethodPost, "/api/v1/coupons/coupons/redeem-preview", body, token)
	assertStatus(t, rec, http.StatusOK)

	resp := parseJSON(t, rec)
	if !resp.Success {
		t.Fatal("expected success=true")
	}
	data := dataMap(t, resp)
	if data["snapshotId"] == nil {
		t.Error("expected snapshotId in redeem-preview response")
	}
}
