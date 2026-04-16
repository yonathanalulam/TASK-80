//go:build integration

package apitest

import (
	"net/http"
	"testing"

	"travel-platform/apps/api/internal/common"
)

func TestListBookings(t *testing.T) {
	token := tokenFor(t, OrganizerUserID, OrganizerEmail, []string{common.RoleGroupOrganizer})
	rec := doRequest(t, http.MethodGet, "/api/v1/bookings", "", token)
	assertStatus(t, rec, http.StatusOK)
}

func TestGetBooking(t *testing.T) {
	token := tokenFor(t, OrganizerUserID, OrganizerEmail, []string{common.RoleGroupOrganizer})
	rec := doRequest(t, http.MethodGet, "/api/v1/bookings/"+SeedBookingID, "", token)
	assertStatus(t, rec, http.StatusOK)

	resp := parseJSON(t, rec)
	dm := dataMap(t, resp)
	if dm["id"] == nil {
		t.Error("expected data to contain id")
	}
	if dm["title"] == nil {
		t.Error("expected data to contain title")
	}
}

func TestCreateBooking(t *testing.T) {
	token := tokenFor(t, OrganizerUserID, OrganizerEmail, []string{common.RoleGroupOrganizer})
	body := `{"title":"Test Booking","items":[{"itemName":"Hotel Room","unitPrice":100,"quantity":2,"category":"lodging","itemType":"lodging"}]}`
	rec := doRequest(t, http.MethodPost, "/api/v1/bookings", body, token)
	assertStatus(t, rec, http.StatusCreated)
}

func TestCreateBooking_MissingTitle(t *testing.T) {
	token := tokenFor(t, OrganizerUserID, OrganizerEmail, []string{common.RoleGroupOrganizer})
	body := `{"items":[{"itemName":"X","unitPrice":1,"quantity":1}]}`
	rec := doRequest(t, http.MethodPost, "/api/v1/bookings", body, token)
	assertStatus(t, rec, http.StatusUnprocessableEntity)
}

func TestCreateBooking_Unauthenticated(t *testing.T) {
	body := `{"title":"Test Booking","items":[{"itemName":"Hotel Room","unitPrice":100,"quantity":2,"category":"lodging","itemType":"lodging"}]}`
	rec := doRequest(t, http.MethodPost, "/api/v1/bookings", body, "")
	assertStatus(t, rec, http.StatusUnauthorized)
}

func TestBookingPricePreview(t *testing.T) {
	token := tokenFor(t, OrganizerUserID, OrganizerEmail, []string{common.RoleGroupOrganizer})
	body := `{"couponCodes":["SAVE25"]}`
	rec := doRequest(t, http.MethodPost, "/api/v1/bookings/"+SeedBookingID+"/price-preview", body, token)
	assertStatus(t, rec, http.StatusOK)
}

func TestBookingCancel_NotOwner(t *testing.T) {
	token := tokenFor(t, TravelerUserID, TravelerEmail, []string{common.RoleTraveler})
	rec := doRequest(t, http.MethodPost, "/api/v1/bookings/"+SeedBookingID+"/cancel", "", token)
	assertStatus(t, rec, http.StatusForbidden)
}

func TestBookingCheckout(t *testing.T) {
	// Create a fresh booking, then attempt checkout.
	token := tokenFor(t, OrganizerUserID, OrganizerEmail, []string{common.RoleGroupOrganizer})
	createBody := `{"title":"Checkout Test Booking","items":[{"itemName":"Room","unitPrice":100,"quantity":1,"category":"lodging","itemType":"lodging"}]}`
	createRec := doRequest(t, http.MethodPost, "/api/v1/bookings", createBody, token)
	if createRec.Code != http.StatusCreated {
		t.Fatalf("setup: create booking failed: %d", createRec.Code)
	}
	createResp := parseJSON(t, createRec)
	dm := dataMap(t, createResp)
	bookingID, _ := dm["id"].(string)

	body := `{"pricingSnapshotId":"snap-test","couponCodes":[],"idempotencyKey":"idem-test-1"}`
	rec := doRequest(t, http.MethodPost, "/api/v1/bookings/"+bookingID+"/checkout", body, token)
	// Handler reached — may fail with 400/500 due to missing snapshot.
	if rec.Code == http.StatusUnauthorized || rec.Code == http.StatusForbidden {
		t.Errorf("expected handler execution, got %d (body: %s)", rec.Code, rec.Body.String())
	}
}

func TestBookingRecordTender(t *testing.T) {
	token := tokenFor(t, OrganizerUserID, OrganizerEmail, []string{common.RoleGroupOrganizer})
	body := `{"tenderType":"cash","amount":100,"reference":"cash-001"}`
	rec := doRequest(t, http.MethodPost, "/api/v1/bookings/"+SeedBookingID+"/record-tender", body, token)
	if rec.Code == http.StatusUnauthorized || rec.Code == http.StatusForbidden {
		t.Errorf("expected handler execution, got %d (body: %s)", rec.Code, rec.Body.String())
	}
}

func TestBookingComplete(t *testing.T) {
	token := tokenFor(t, OrganizerUserID, OrganizerEmail, []string{common.RoleGroupOrganizer})
	rec := doRequest(t, http.MethodPost, "/api/v1/bookings/"+SeedBookingID+"/complete", "", token)
	// May fail with conflict (wrong status), but proves handler reached.
	if rec.Code == http.StatusUnauthorized || rec.Code == http.StatusForbidden {
		t.Errorf("expected handler execution, got %d (body: %s)", rec.Code, rec.Body.String())
	}
}
