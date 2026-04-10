package reviews

import (
	"encoding/json"
	"testing"
)

func TestCreateReviewRequest_CanonicalShape(t *testing.T) {
	payload := `{
		"subjectId": "user-2",
		"orderType": "booking",
		"orderId": "booking-123",
		"overallRating": 4.5,
		"comment": "Great service",
		"scores": [
			{"dimensionName": "punctuality", "score": 5.0},
			{"dimensionName": "quality", "score": 4.0}
		]
	}`

	var req CreateReviewRequest
	if err := json.Unmarshal([]byte(payload), &req); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if req.SubjectID != "user-2" {
		t.Errorf("SubjectID = %q", req.SubjectID)
	}
	if req.OrderType != "booking" {
		t.Errorf("OrderType = %q", req.OrderType)
	}
	if req.OrderID != "booking-123" {
		t.Errorf("OrderID = %q", req.OrderID)
	}
	if len(req.Scores) != 2 {
		t.Fatalf("expected 2 scores, got %d", len(req.Scores))
	}
	if req.Scores[0].DimensionName != "punctuality" {
		t.Errorf("Scores[0].DimensionName = %q", req.Scores[0].DimensionName)
	}
}

func TestCreateReviewRequest_RejectsOldShape(t *testing.T) {
	oldPayload := `{
		"subjectId": "user-2",
		"overallRating": 4.5,
		"dimensions": {"punctuality": 5, "quality": 4},
		"comment": "Good"
	}`

	var req CreateReviewRequest
	_ = json.Unmarshal([]byte(oldPayload), &req)

	if req.OrderType != "" {
		t.Error("old payload without orderType should leave it empty")
	}
	if req.OrderID != "" {
		t.Error("old payload without orderId should leave it empty")
	}
	if len(req.Scores) != 0 {
		t.Error("'dimensions' should NOT populate Scores (different json key)")
	}
}

func TestReviewResponse_CanonicalFieldNames(t *testing.T) {
	resp := ReviewResponse{
		ID:            "r1",
		ReviewerID:    "u1",
		SubjectID:     "u2",
		OrderType:     "booking",
		OrderID:       "b1",
		OverallRating: 4.0,
		Comment:       "Good",
		Scores:        []ScoreDetail{{DimensionName: "quality", Score: 4.0}},
	}

	data, _ := json.Marshal(resp)
	var m map[string]interface{}
	_ = json.Unmarshal(data, &m)

	if _, ok := m["reviewerId"]; !ok {
		t.Error("response should have 'reviewerId'")
	}
	if _, ok := m["subjectId"]; !ok {
		t.Error("response should have 'subjectId'")
	}
	if _, ok := m["scores"]; !ok {
		t.Error("response should have 'scores'")
	}

	if _, ok := m["reviewerName"]; ok {
		t.Error("response should NOT have 'reviewerName'")
	}
	if _, ok := m["subjectName"]; ok {
		t.Error("response should NOT have 'subjectName'")
	}
	if _, ok := m["dimensions"]; ok {
		t.Error("response should NOT have 'dimensions'")
	}
}
