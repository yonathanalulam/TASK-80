package bookings

import (
	"encoding/json"
	"testing"
)

func TestCreateBookingRequest_CanonicalShape(t *testing.T) {
	payload := `{
		"title": "Mountain Adventure",
		"description": "Group trip",
		"itineraryId": "itin-123",
		"items": [
			{
				"itemType": "lodging",
				"itemName": "Hotel Room",
				"description": "Double room",
				"unitPrice": 150.00,
				"quantity": 2,
				"category": "lodging"
			}
		]
	}`

	var req CreateBookingRequest
	if err := json.Unmarshal([]byte(payload), &req); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if req.Title != "Mountain Adventure" {
		t.Errorf("Title = %q", req.Title)
	}
	if req.ItineraryID == nil || *req.ItineraryID != "itin-123" {
		t.Error("ItineraryID should be 'itin-123'")
	}
	if len(req.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(req.Items))
	}
	if req.Items[0].ItemType != "lodging" {
		t.Errorf("ItemType = %q, want 'lodging'", req.Items[0].ItemType)
	}
	if req.Items[0].ItemName != "Hotel Room" {
		t.Errorf("ItemName = %q, want 'Hotel Room'", req.Items[0].ItemName)
	}
}

func TestCreateBookingRequest_RejectsOldFieldNames(t *testing.T) {
	oldPayload := `{
		"title": "Trip",
		"itineraryLink": "itin-123",
		"lineItems": [{"type": "lodging", "name": "Room"}]
	}`

	var req CreateBookingRequest
	_ = json.Unmarshal([]byte(oldPayload), &req)

	if req.ItineraryID != nil {
		t.Error("'itineraryLink' should not populate ItineraryID")
	}
	if len(req.Items) != 0 {
		t.Error("'lineItems' should not populate Items")
	}
}

func TestCheckoutRequest_CanonicalShape(t *testing.T) {
	payload := `{
		"pricingSnapshotId": "snap-123",
		"couponCodes": ["SAVE25"],
		"idempotencyKey": "key-abc"
	}`

	var req CheckoutRequest
	if err := json.Unmarshal([]byte(payload), &req); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if req.PricingSnapshotID != "snap-123" {
		t.Errorf("PricingSnapshotID = %q", req.PricingSnapshotID)
	}
	if len(req.CouponCodes) != 1 || req.CouponCodes[0] != "SAVE25" {
		t.Errorf("CouponCodes = %v", req.CouponCodes)
	}
	if req.IdempotencyKey != "key-abc" {
		t.Errorf("IdempotencyKey = %q", req.IdempotencyKey)
	}
}
