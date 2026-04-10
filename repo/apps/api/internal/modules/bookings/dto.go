package bookings

import "time"
type CreateBookingRequest struct {
	Title       string                   `json:"title"`
	Description string                   `json:"description"`
	ItineraryID *string                  `json:"itineraryId"`
	Items       []CreateBookingItemRequest `json:"items"`
}

type CreateBookingItemRequest struct {
	ItemType    string  `json:"itemType"`
	ItemName    string  `json:"itemName"`
	Description string  `json:"description"`
	UnitPrice   float64 `json:"unitPrice"`
	Quantity    int     `json:"quantity"`
	Category    string  `json:"category"`
}

type CheckoutRequest struct {
	PricingSnapshotID string   `json:"pricingSnapshotId"`
	CouponCodes       []string `json:"couponCodes"`
	IdempotencyKey    string   `json:"idempotencyKey"`
	MembershipTier    string   `json:"membershipTier"`
	IsNewUser         bool     `json:"isNewUser"`
}

type TenderRecordRequest struct {
	TenderType    string  `json:"tenderType"`
	Amount        float64 `json:"amount"`
	ReferenceText string  `json:"referenceText"`
}
type BookingResponse struct {
	ID                string        `json:"id"`
	OrganizerID       string        `json:"organizerId"`
	ItineraryID       *string       `json:"itineraryId,omitempty"`
	Title             string        `json:"title"`
	Description       string        `json:"description"`
	Status            BookingStatus `json:"status"`
	TotalAmount       float64       `json:"totalAmount"`
	DiscountAmount    float64       `json:"discountAmount"`
	EscrowAmount      float64       `json:"escrowAmount"`
	PricingSnapshotID *string       `json:"pricingSnapshotId,omitempty"`
	Items             []BookingItemResponse `json:"items,omitempty"`
	CreatedAt         time.Time     `json:"createdAt"`
	UpdatedAt         time.Time     `json:"updatedAt"`
}

type BookingItemResponse struct {
	ID          string  `json:"id"`
	ItemType    string  `json:"itemType"`
	ItemName    string  `json:"itemName"`
	Description string  `json:"description"`
	UnitPrice   float64 `json:"unitPrice"`
	Quantity    int     `json:"quantity"`
	Subtotal    float64 `json:"subtotal"`
	Category    string  `json:"category"`
}

type CheckoutResponse struct {
	BookingID      string  `json:"bookingId"`
	Status         string  `json:"status"`
	TotalAmount    float64 `json:"totalAmount"`
	DiscountAmount float64 `json:"discountAmount"`
	EscrowAmount   float64 `json:"escrowAmount"`
	SnapshotID     string  `json:"snapshotId"`
}
