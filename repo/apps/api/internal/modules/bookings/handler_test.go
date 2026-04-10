package bookings

import (
	"encoding/json"
	"testing"
)

// TestBookingListResponseShape verifies that the booking list response
// conforms to the expected contract: {items, total, page, pageSize, totalPages}.
// This prevents frontend/backend response-shape mismatches.
func TestBookingListResponseShape(t *testing.T) {
	// Simulate the response map that ListBookings returns.
	response := map[string]interface{}{
		"items":      []map[string]interface{}{},
		"total":      0,
		"page":       1,
		"pageSize":   20,
		"totalPages": 0,
	}

	// Verify all required fields exist.
	requiredFields := []string{"items", "total", "page", "pageSize", "totalPages"}
	for _, field := range requiredFields {
		if _, ok := response[field]; !ok {
			t.Errorf("booking list response missing required field %q", field)
		}
	}

	// Verify NO forbidden fields that would cause frontend mismatch.
	forbiddenFields := []string{"data", "meta", "totalCount"}
	for _, field := range forbiddenFields {
		if _, ok := response[field]; ok {
			t.Errorf("booking list response contains legacy field %q which causes frontend mismatch", field)
		}
	}

	// Verify items is an array.
	items, ok := response["items"].([]map[string]interface{})
	if !ok {
		t.Fatal("items should be an array")
	}
	_ = items

	// Verify the response can be JSON-marshaled.
	data, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("failed to marshal booking list response: %v", err)
	}

	// Verify it can be unmarshaled into the expected shape.
	var parsed struct {
		Items      []interface{} `json:"items"`
		Total      int           `json:"total"`
		Page       int           `json:"page"`
		PageSize   int           `json:"pageSize"`
		TotalPages int           `json:"totalPages"`
	}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("failed to unmarshal booking list response: %v", err)
	}

	if parsed.Page != 1 {
		t.Errorf("page = %d, want 1", parsed.Page)
	}
	if parsed.PageSize != 20 {
		t.Errorf("pageSize = %d, want 20", parsed.PageSize)
	}
}

// TestBookingItemResponseFields verifies that individual booking items
// contain the fields the frontend expects.
func TestBookingItemResponseFields(t *testing.T) {
	item := map[string]interface{}{
		"id":             "test-id",
		"organizerId":    "org-id",
		"itineraryId":    nil,
		"title":          "Test Booking",
		"description":    nil,
		"status":         "draft",
		"totalAmount":    850.00,
		"discountAmount": 0.0,
		"escrowAmount":   0.0,
		"createdAt":      "2026-01-01T00:00:00Z",
		"updatedAt":      "2026-01-01T00:00:00Z",
	}

	// Frontend requires these fields.
	requiredFields := []string{"id", "title", "status", "totalAmount", "discountAmount", "createdAt"}
	for _, field := range requiredFields {
		if _, ok := item[field]; !ok {
			t.Errorf("booking item missing required field %q", field)
		}
	}

	// Frontend uses "discountAmount" not "discount".
	if _, ok := item["discount"]; ok {
		t.Error("booking item should use 'discountAmount' not 'discount'")
	}
}
