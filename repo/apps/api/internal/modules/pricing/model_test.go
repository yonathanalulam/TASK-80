package pricing

import (
	"encoding/json"
	"testing"
)

func TestPricingResult_UsesEligibleIneligibleCoupons(t *testing.T) {
	result := PricingResult{
		Subtotal:          850.00,
		TotalDiscount:     25.00,
		EscrowHoldAmount:  825.00,
		FinalPayable:      825.00,
		EligibleCoupons:   []AppliedCoupon{{CouponID: "c1", Code: "SAVE25", Name: "$25 Off", DiscountAmount: 25}},
		IneligibleCoupons: []IneligibleCoupon{{CouponID: "c2", Code: "VIP", Name: "VIP Only", ReasonCode: "MEMBERSHIP_REQUIRED", Message: "Membership required"}},
		AppliedDiscounts:  []AppliedDiscount{},
	}

	data, err := json.Marshal(result)
	if err != nil {
		t.Fatal(err)
	}

	var m map[string]interface{}
	_ = json.Unmarshal(data, &m)

	if _, ok := m["eligibleCoupons"]; !ok {
		t.Error("should have 'eligibleCoupons'")
	}
	if _, ok := m["ineligibleCoupons"]; !ok {
		t.Error("should have 'ineligibleCoupons'")
	}
	if _, ok := m["escrowHoldAmount"]; !ok {
		t.Error("should have 'escrowHoldAmount'")
	}
	if _, ok := m["coupons"]; ok {
		t.Error("should NOT have generic 'coupons' field")
	}
	if _, ok := m["escrowHold"]; ok {
		t.Error("should NOT have 'escrowHold' (use 'escrowHoldAmount')")
	}
}

func TestAppliedCoupon_HasDiscountAmount(t *testing.T) {
	c := AppliedCoupon{CouponID: "1", Code: "X", Name: "Test", DiscountAmount: 25}
	data, _ := json.Marshal(c)
	var m map[string]interface{}
	_ = json.Unmarshal(data, &m)

	if _, ok := m["discountAmount"]; !ok {
		t.Error("AppliedCoupon should have 'discountAmount'")
	}
}

func TestIneligibleCoupon_HasReasonFields(t *testing.T) {
	c := IneligibleCoupon{CouponID: "1", Code: "X", Name: "Test", ReasonCode: "MIN_SPEND", Message: "Min spend not met"}
	data, _ := json.Marshal(c)
	var m map[string]interface{}
	_ = json.Unmarshal(data, &m)

	if _, ok := m["reasonCode"]; !ok {
		t.Error("IneligibleCoupon should have 'reasonCode'")
	}
	if _, ok := m["message"]; !ok {
		t.Error("IneligibleCoupon should have 'message'")
	}
}
