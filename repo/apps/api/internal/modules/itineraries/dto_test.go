package itineraries

import (
	"encoding/json"
	"testing"
)

func TestCreateItineraryRequest_AcceptsCanonicalPayload(t *testing.T) {
	payload := `{
		"title": "Test Trip",
		"meetupAt": "2026-07-14T18:30:00Z",
		"meetupLocationText": "Central Station",
		"notes": "Pack light"
	}`

	var req CreateItineraryRequest
	if err := json.Unmarshal([]byte(payload), &req); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if req.Title != "Test Trip" {
		t.Errorf("Title = %q, want %q", req.Title, "Test Trip")
	}
	if req.MeetupAt == nil || *req.MeetupAt != "2026-07-14T18:30:00Z" {
		t.Errorf("MeetupAt = %v, want %q", req.MeetupAt, "2026-07-14T18:30:00Z")
	}
	if req.MeetupLocationText != "Central Station" {
		t.Errorf("MeetupLocationText = %q, want %q", req.MeetupLocationText, "Central Station")
	}
}

func TestItineraryResponse_CanonicalFieldNames(t *testing.T) {
	resp := ItineraryResponse{
		ID:                 "test",
		Title:              "Trip",
		MeetupLocationText: "Station",
		MembersCount:       5,
	}
	data, _ := json.Marshal(resp)

	var m map[string]interface{}
	json.Unmarshal(data, &m)

	if _, ok := m["meetupAt"]; !ok {
		t.Error("response should have 'meetupAt' field")
	}
	if _, ok := m["meetupLocationText"]; !ok {
		t.Error("response should have 'meetupLocationText' field")
	}
	if _, ok := m["membersCount"]; !ok {
		t.Error("response should have 'membersCount' field")
	}

	// Old field names should not exist
	if _, ok := m["meetupDate"]; ok {
		t.Error("response should NOT have 'meetupDate' (use 'meetupAt')")
	}
	if _, ok := m["location"]; ok {
		t.Error("response should NOT have 'location' (use 'meetupLocationText')")
	}
	if _, ok := m["memberCount"]; ok {
		t.Error("response should NOT have 'memberCount' (use 'membersCount')")
	}
}

func TestCreateItineraryRequest_RejectsOldFieldNames(t *testing.T) {
	oldPayload := `{
		"title": "Trip",
		"meetupDate": "07/14/2026 6:30 PM",
		"location": "Station"
	}`

	var req CreateItineraryRequest
	_ = json.Unmarshal([]byte(oldPayload), &req)

	if req.MeetupAt != nil {
		t.Error("old field name 'meetupDate' should not populate MeetupAt")
	}
	if req.MeetupLocationText != "" {
		t.Error("old field name 'location' should not populate MeetupLocationText")
	}
}
