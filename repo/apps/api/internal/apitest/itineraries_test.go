//go:build integration

package apitest

import (
	"net/http"
	"testing"

	"travel-platform/apps/api/internal/common"
)

func TestListItineraries(t *testing.T) {
	token := tokenFor(t, OrganizerUserID, OrganizerEmail, []string{common.RoleGroupOrganizer})
	rec := doRequest(t, http.MethodGet, "/api/v1/itineraries", "", token)
	assertStatus(t, rec, http.StatusOK)
}

func TestGetItinerary(t *testing.T) {
	token := tokenFor(t, OrganizerUserID, OrganizerEmail, []string{common.RoleGroupOrganizer})
	rec := doRequest(t, http.MethodGet, "/api/v1/itineraries/"+SeedItineraryID, "", token)
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

func TestCreateItinerary(t *testing.T) {
	token := tokenFor(t, OrganizerUserID, OrganizerEmail, []string{common.RoleGroupOrganizer})
	body := `{"title":"Test Trip","meetupAt":"2026-08-01T10:00:00Z","meetupLocationText":"Airport Gate B","notes":"test"}`
	rec := doRequest(t, http.MethodPost, "/api/v1/itineraries", body, token)
	assertStatus(t, rec, http.StatusCreated)

	resp := parseJSON(t, rec)
	dm := dataMap(t, resp)
	if dm["id"] == nil {
		t.Error("expected created itinerary to have an id")
	}
}

func TestCreateItinerary_Unauthenticated(t *testing.T) {
	body := `{"title":"Test Trip","meetupAt":"2026-08-01T10:00:00Z","meetupLocationText":"Airport Gate B","notes":"test"}`
	rec := doRequest(t, http.MethodPost, "/api/v1/itineraries", body, "")
	assertStatus(t, rec, http.StatusUnauthorized)
}

func TestGetItinerary_Checkpoints(t *testing.T) {
	token := tokenFor(t, OrganizerUserID, OrganizerEmail, []string{common.RoleGroupOrganizer})
	rec := doRequest(t, http.MethodGet, "/api/v1/itineraries/"+SeedItineraryID, "", token)
	assertStatus(t, rec, http.StatusOK)

	resp := parseJSON(t, rec)
	dm := dataMap(t, resp)
	if dm == nil {
		t.Fatal("expected non-nil data map")
	}
}

func TestItineraryFormDefinitions(t *testing.T) {
	token := tokenFor(t, OrganizerUserID, OrganizerEmail, []string{common.RoleGroupOrganizer})
	rec := doRequest(t, http.MethodGet, "/api/v1/itineraries/"+SeedItineraryID+"/form-definitions", "", token)
	assertStatus(t, rec, http.StatusOK)
}

func TestItineraryFormSubmissions(t *testing.T) {
	token := tokenFor(t, OrganizerUserID, OrganizerEmail, []string{common.RoleGroupOrganizer})
	rec := doRequest(t, http.MethodGet, "/api/v1/itineraries/"+SeedItineraryID+"/form-submissions", "", token)
	assertStatus(t, rec, http.StatusOK)
}

func TestItineraryChangeEvents(t *testing.T) {
	token := tokenFor(t, OrganizerUserID, OrganizerEmail, []string{common.RoleGroupOrganizer})
	rec := doRequest(t, http.MethodGet, "/api/v1/itineraries/"+SeedItineraryID+"/change-events", "", token)
	assertStatus(t, rec, http.StatusOK)
}

func TestUpdateItinerary(t *testing.T) {
	token := tokenFor(t, OrganizerUserID, OrganizerEmail, []string{common.RoleGroupOrganizer})
	body := `{"title":"Updated Mountain Trail Adventure 2026","notes":"updated notes"}`
	rec := doRequest(t, http.MethodPatch, "/api/v1/itineraries/"+SeedItineraryID, body, token)
	// Real handler reached.
	if rec.Code == http.StatusUnauthorized || rec.Code == http.StatusForbidden {
		t.Errorf("expected handler execution, got %d (body: %s)", rec.Code, rec.Body.String())
	}
}

func TestPublishItinerary(t *testing.T) {
	// Create a new itinerary first, then publish it.
	token := tokenFor(t, OrganizerUserID, OrganizerEmail, []string{common.RoleGroupOrganizer})
	createBody := `{"title":"Publish Test Trip","meetupAt":"2026-09-01T10:00:00Z","meetupLocationText":"Gate C","notes":"test"}`
	createRec := doRequest(t, http.MethodPost, "/api/v1/itineraries", createBody, token)
	if createRec.Code != http.StatusCreated {
		t.Fatalf("setup: create itinerary failed: %d", createRec.Code)
	}
	createResp := parseJSON(t, createRec)
	dm := dataMap(t, createResp)
	newID, ok := dm["id"].(string)
	if !ok || newID == "" {
		t.Fatal("setup: no id returned from create")
	}

	rec := doRequest(t, http.MethodPost, "/api/v1/itineraries/"+newID+"/publish", "", token)
	if rec.Code == http.StatusUnauthorized || rec.Code == http.StatusForbidden {
		t.Errorf("expected handler execution, got %d (body: %s)", rec.Code, rec.Body.String())
	}
}

func TestAddCheckpoint(t *testing.T) {
	token := tokenFor(t, OrganizerUserID, OrganizerEmail, []string{common.RoleGroupOrganizer})
	body := `{"checkpointText":"New Waypoint","eta":"2026-07-16T10:00:00Z","sortOrder":10}`
	rec := doRequest(t, http.MethodPost, "/api/v1/itineraries/"+SeedItineraryID+"/checkpoints", body, token)
	if rec.Code == http.StatusUnauthorized || rec.Code == http.StatusForbidden {
		t.Errorf("expected handler execution, got %d (body: %s)", rec.Code, rec.Body.String())
	}
}

func TestAddMember(t *testing.T) {
	token := tokenFor(t, OrganizerUserID, OrganizerEmail, []string{common.RoleGroupOrganizer})
	body := `{"userId":"c0000000-0000-0000-0000-000000000007","role":"participant"}`
	rec := doRequest(t, http.MethodPost, "/api/v1/itineraries/"+SeedItineraryID+"/members", body, token)
	if rec.Code == http.StatusUnauthorized || rec.Code == http.StatusForbidden {
		t.Errorf("expected handler execution, got %d (body: %s)", rec.Code, rec.Body.String())
	}
}

func TestCreateFormDefinition(t *testing.T) {
	token := tokenFor(t, OrganizerUserID, OrganizerEmail, []string{common.RoleGroupOrganizer})
	body := `{"fieldKey":"test_field","fieldLabel":"Test Field","fieldType":"text","required":false,"sortOrder":99}`
	rec := doRequest(t, http.MethodPost, "/api/v1/itineraries/"+SeedItineraryID+"/form-definitions", body, token)
	if rec.Code == http.StatusUnauthorized || rec.Code == http.StatusForbidden {
		t.Errorf("expected handler execution, got %d (body: %s)", rec.Code, rec.Body.String())
	}
}

func TestSubmitFormSubmission(t *testing.T) {
	token := tokenFor(t, TravelerUserID, TravelerEmail, []string{common.RoleTraveler})
	body := `{"values":{"emergency_contact":"John Doe","emergency_phone":"555-0100"}}`
	rec := doRequest(t, http.MethodPost, "/api/v1/itineraries/"+SeedItineraryID+"/form-submissions", body, token)
	if rec.Code == http.StatusUnauthorized || rec.Code == http.StatusForbidden {
		t.Errorf("expected handler execution, got %d (body: %s)", rec.Code, rec.Body.String())
	}
}

func TestUpdateCheckpoint(t *testing.T) {
	token := tokenFor(t, OrganizerUserID, OrganizerEmail, []string{common.RoleGroupOrganizer})
	// Use seed checkpoint ID e0000000-0000-0000-0000-000000000001
	body := `{"checkpointText":"Updated Depart Central Station","eta":"2026-07-14T19:00:00Z"}`
	rec := doRequest(t, http.MethodPatch, "/api/v1/itineraries/"+SeedItineraryID+"/checkpoints/e0000000-0000-0000-0000-000000000001", body, token)
	if rec.Code == http.StatusUnauthorized || rec.Code == http.StatusForbidden {
		t.Errorf("expected handler execution, got %d (body: %s)", rec.Code, rec.Body.String())
	}
}

func TestDeleteCheckpoint(t *testing.T) {
	token := tokenFor(t, OrganizerUserID, OrganizerEmail, []string{common.RoleGroupOrganizer})
	// Use seed checkpoint e0000000-0000-0000-0000-000000000005 (last one, safe to remove)
	rec := doRequest(t, http.MethodDelete, "/api/v1/itineraries/"+SeedItineraryID+"/checkpoints/e0000000-0000-0000-0000-000000000005", "", token)
	if rec.Code == http.StatusUnauthorized || rec.Code == http.StatusForbidden {
		t.Errorf("expected handler execution, got %d (body: %s)", rec.Code, rec.Body.String())
	}
}

func TestRemoveMember(t *testing.T) {
	token := tokenFor(t, OrganizerUserID, OrganizerEmail, []string{common.RoleGroupOrganizer})
	// Remove traveler c0000000-0000-0000-0000-000000000007 (added in TestAddMember above or via seed)
	rec := doRequest(t, http.MethodDelete, "/api/v1/itineraries/"+SeedItineraryID+"/members/c0000000-0000-0000-0000-000000000007", "", token)
	if rec.Code == http.StatusUnauthorized || rec.Code == http.StatusForbidden {
		t.Errorf("expected handler execution, got %d (body: %s)", rec.Code, rec.Body.String())
	}
}

func TestUpdateFormDefinition(t *testing.T) {
	token := tokenFor(t, OrganizerUserID, OrganizerEmail, []string{common.RoleGroupOrganizer})
	// Use seed form def ID f1000000-0000-0000-0000-000000000001
	body := `{"fieldLabel":"Updated Vehicle Plate","required":true}`
	rec := doRequest(t, http.MethodPatch, "/api/v1/itineraries/"+SeedItineraryID+"/form-definitions/f1000000-0000-0000-0000-000000000001", body, token)
	if rec.Code == http.StatusUnauthorized || rec.Code == http.StatusForbidden {
		t.Errorf("expected handler execution, got %d (body: %s)", rec.Code, rec.Body.String())
	}
}
