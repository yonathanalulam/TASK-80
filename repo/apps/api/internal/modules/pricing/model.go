package pricing

import (
	"encoding/json"
	"time"
)

type BookingItem struct {
	ID          string  `json:"id"`
	BookingID   string  `json:"bookingId"`
	ItemType    string  `json:"itemType"`
	ItemName    string  `json:"itemName"`
	Description string  `json:"description"`
	UnitPrice   float64 `json:"unitPrice"`
	Quantity    int     `json:"quantity"`
	Subtotal    float64 `json:"subtotal"`
	Category    string  `json:"category"`
}

type Coupon struct {
	ID               string          `json:"id"`
	Code             string          `json:"code"`
	Name             string          `json:"name"`
	DiscountType     string          `json:"discountType"`
	Amount           float64         `json:"amount"`
	MinSpend         float64         `json:"minSpend"`
	PercentOff       float64         `json:"percentOff"`
	ValidFrom        *time.Time      `json:"validFrom"`
	ValidTo          *time.Time      `json:"validTo"`
	EligibilityJSON  json.RawMessage `json:"eligibilityJson"`
	StackGroup       string          `json:"stackGroup"`
	Exclusive        bool            `json:"exclusive"`
	UsageLimitTotal  *int            `json:"usageLimitTotal"`
	UsageLimitPerUser *int           `json:"usageLimitPerUser"`
	Active           bool            `json:"active"`
	CreatedAt        time.Time       `json:"createdAt"`
	UpdatedAt        time.Time       `json:"updatedAt"`
}

type CouponEligibility struct {
	Categories         []string `json:"categories,omitempty"`
	NewUserOnly        bool     `json:"newUserOnly,omitempty"`
	MembershipRequired string   `json:"membershipRequired,omitempty"`
}

type CouponEligibilityResult struct {
	CouponID       string  `json:"couponId"`
	Code           string  `json:"code"`
	Name           string  `json:"name"`
	Eligible       bool    `json:"eligible"`
	Reason         string  `json:"reason"`
	ReasonCode     string  `json:"reasonCode,omitempty"`
	DiscountAmount float64 `json:"discountAmount"`
}

type PricingResult struct {
	Subtotal          float64            `json:"subtotal"`
	TotalDiscount     float64            `json:"totalDiscount"`
	EscrowHoldAmount  float64            `json:"escrowHoldAmount"`
	FinalPayable      float64            `json:"finalPayable"`
	EligibleCoupons   []AppliedCoupon    `json:"eligibleCoupons"`
	IneligibleCoupons []IneligibleCoupon `json:"ineligibleCoupons"`
	AppliedDiscounts  []AppliedDiscount  `json:"appliedDiscounts"`
	SnapshotID        string             `json:"snapshotId,omitempty"`
}

type AppliedCoupon struct {
	CouponID       string  `json:"couponId"`
	Code           string  `json:"code"`
	Name           string  `json:"name"`
	DiscountAmount float64 `json:"discountAmount"`
}

type IneligibleCoupon struct {
	CouponID   string `json:"couponId"`
	Code       string `json:"code"`
	Name       string `json:"name"`
	ReasonCode string `json:"reasonCode"`
	Message    string `json:"message"`
}

type AppliedDiscount struct {
	Type        string  `json:"type"`
	Description string  `json:"description"`
	Amount      float64 `json:"amount"`
}

type IdempotencyResponse struct {
	ResponseCode int             `json:"responseCode"`
	ResponseBody json.RawMessage `json:"responseBody"`
}
