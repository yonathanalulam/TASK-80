package contracts

import (
	"encoding/json"
	"testing"
)

func TestCreateInvoiceRequestDTO_CanonicalShape(t *testing.T) {
	payload := `{
		"orderType": "booking",
		"orderId": "booking-123",
		"notes": "Please generate invoice"
	}`

	var req CreateInvoiceRequestDTO
	if err := json.Unmarshal([]byte(payload), &req); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if req.OrderType != "booking" {
		t.Errorf("OrderType = %q, want 'booking'", req.OrderType)
	}
	if req.OrderID != "booking-123" {
		t.Errorf("OrderID = %q, want 'booking-123'", req.OrderID)
	}
}

func TestCreateInvoiceRequestDTO_RejectsOldFieldNames(t *testing.T) {
	oldPayload := `{
		"bookingId": "booking-123",
		"amount": 850.00
	}`

	var req CreateInvoiceRequestDTO
	_ = json.Unmarshal([]byte(oldPayload), &req)

	if req.OrderType != "" {
		t.Error("'bookingId' should not populate OrderType")
	}
	if req.OrderID != "" {
		t.Error("'bookingId' should not populate OrderID")
	}
}
