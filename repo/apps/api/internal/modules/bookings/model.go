package bookings

import (
	"time"
)

type BookingStatus string

const (
	StatusDraft              BookingStatus = "draft"
	StatusPendingPricing     BookingStatus = "pending_pricing"
	StatusPaidHeldInEscrow   BookingStatus = "paid_held_in_escrow"
	StatusCompleted          BookingStatus = "completed"
	StatusCancelled          BookingStatus = "cancelled"
	StatusRefunded           BookingStatus = "refunded"
)

type EscrowStatus string

const (
	EscrowHeld     EscrowStatus = "held"
	EscrowReleased EscrowStatus = "released"
	EscrowRefunded EscrowStatus = "refunded"
)

type Booking struct {
	ID                string        `json:"id"`
	OrganizerID       string        `json:"organizerId"`
	ItineraryID       *string       `json:"itineraryId"`
	Title             string        `json:"title"`
	Description       string        `json:"description"`
	Status            BookingStatus `json:"status"`
	TotalAmount       float64       `json:"totalAmount"`
	DiscountAmount    float64       `json:"discountAmount"`
	EscrowAmount      float64       `json:"escrowAmount"`
	PricingSnapshotID *string       `json:"pricingSnapshotId"`
	CreatedAt         time.Time     `json:"createdAt"`
	UpdatedAt         time.Time     `json:"updatedAt"`
}

type BookingItem struct {
	ID          string    `json:"id"`
	BookingID   string    `json:"bookingId"`
	ItemType    string    `json:"itemType"`
	ItemName    string    `json:"itemName"`
	Description string    `json:"description"`
	UnitPrice   float64   `json:"unitPrice"`
	Quantity    int       `json:"quantity"`
	Subtotal    float64   `json:"subtotal"`
	Category    string    `json:"category"`
	CreatedAt   time.Time `json:"createdAt"`
}

type Escrow struct {
	ID             string       `json:"id"`
	OrderType      string       `json:"orderType"`
	OrderID        string       `json:"orderId"`
	AmountHeld     float64      `json:"amountHeld"`
	AmountReleased float64      `json:"amountReleased"`
	AmountRefunded float64      `json:"amountRefunded"`
	Status         EscrowStatus `json:"status"`
}

type PaymentRecord struct {
	ID            string    `json:"id"`
	OrderType     string    `json:"orderType"`
	OrderID       string    `json:"orderId"`
	TenderType    string    `json:"tenderType"`
	Amount        float64   `json:"amount"`
	Currency      string    `json:"currency"`
	ReferenceText string    `json:"referenceText"`
	RecordedBy    string    `json:"recordedBy"`
	RecordedAt    time.Time `json:"recordedAt"`
}

type CouponRedemption struct {
	ID                 string  `json:"id"`
	CouponID           string  `json:"couponId"`
	UserID             string  `json:"userId"`
	BookingID          *string `json:"bookingId"`
	RedemptionScopeKey string  `json:"redemptionScopeKey"`
	DiscountAmount     float64 `json:"discountAmount"`
}
