package pricing

import (
	"context"
	"fmt"
	"math"
	"time"
)

var MemberTierDiscounts = map[string]float64{
	"silver":   0.03,
	"gold":     0.05,
	"platinum": 0.08,
}

func EvaluateCheckout(
	ctx context.Context,
	repo *Repository,
	items []BookingItem,
	couponCodes []string,
	userID string,
	isNewUser bool,
	membershipTier string,
	scopeKey string,
) (*PricingResult, error) {
	now := time.Now()

	var subtotal float64
	for i := range items {
		items[i].Subtotal = roundMoney(float64(items[i].Quantity) * items[i].UnitPrice)
		subtotal += items[i].Subtotal
	}
	subtotal = roundMoney(subtotal)

	result := &PricingResult{
		Subtotal:          subtotal,
		EligibleCoupons:   make([]AppliedCoupon, 0),
		IneligibleCoupons: make([]IneligibleCoupon, 0),
		AppliedDiscounts:  make([]AppliedDiscount, 0),
	}

	workingTotal := subtotal

	if pct, ok := MemberTierDiscounts[membershipTier]; ok {
		discount := roundMoney(subtotal * pct)
		if discount > 0 {
			workingTotal = roundMoney(workingTotal - discount)
			result.AppliedDiscounts = append(result.AppliedDiscounts, AppliedDiscount{
				Type:        "member_tier",
				Description: fmt.Sprintf("%s member discount (%.0f%%)", membershipTier, pct*100),
				Amount:      discount,
			})
		}
	}

	if len(couponCodes) == 0 {
		result.TotalDiscount = roundMoney(subtotal - workingTotal)
		result.FinalPayable = roundMoney(workingTotal)
		result.EscrowHoldAmount = result.FinalPayable
		return result, nil
	}

	coupons, err := repo.GetActiveCouponsByCodes(ctx, couponCodes)
	if err != nil {
		return nil, fmt.Errorf("load coupons: %w", err)
	}

	couponMap := make(map[string]Coupon, len(coupons))
	for _, c := range coupons {
		couponMap[c.Code] = c
	}

	var eligible []CouponEligibilityResult
	for _, code := range couponCodes {
		coupon, found := couponMap[code]
		if !found {
			result.IneligibleCoupons = append(result.IneligibleCoupons, IneligibleCoupon{
				Code:       code,
				ReasonCode: ReasonInactive,
				Message:    FormatReason(ReasonInactive),
			})
			continue
		}

		res := evaluateSingleCoupon(ctx, repo, coupon, items, workingTotal, userID, isNewUser, membershipTier, scopeKey, now)
		if !res.Eligible {
			result.IneligibleCoupons = append(result.IneligibleCoupons, IneligibleCoupon{
				CouponID:   res.CouponID,
				Code:       res.Code,
				Name:       res.Name,
				ReasonCode: res.ReasonCode,
				Message:    res.Reason,
			})
		} else {
			eligible = append(eligible, res)
		}
	}

	var appliedThreshold *CouponEligibilityResult
	var appliedPercentage *CouponEligibilityResult
	var appliedNewUserGift *CouponEligibilityResult

	for i := range eligible {
		e := &eligible[i]
		coupon := couponMap[e.Code]
		switch coupon.DiscountType {
		case "new_user_gift":
			appliedNewUserGift = e
		case "threshold_fixed":
			if appliedThreshold == nil || e.DiscountAmount > appliedThreshold.DiscountAmount {
				appliedThreshold = e
			}
		case "percentage":
			if appliedPercentage == nil || e.DiscountAmount > appliedPercentage.DiscountAmount {
				appliedPercentage = e
			}
		}
	}

	if appliedNewUserGift != nil {
		if appliedThreshold != nil {
			result.IneligibleCoupons = append(result.IneligibleCoupons, IneligibleCoupon{
				CouponID:   appliedThreshold.CouponID,
				Code:       appliedThreshold.Code,
				Name:       appliedThreshold.Name,
				ReasonCode: ReasonStackingNotAllowed,
				Message:    FormatReason(ReasonStackingNotAllowed),
			})
		}
		if appliedPercentage != nil {
			result.IneligibleCoupons = append(result.IneligibleCoupons, IneligibleCoupon{
				CouponID:   appliedPercentage.CouponID,
				Code:       appliedPercentage.Code,
				Name:       appliedPercentage.Name,
				ReasonCode: ReasonStackingNotAllowed,
				Message:    FormatReason(ReasonStackingNotAllowed),
			})
		}

		discount := roundMoney(appliedNewUserGift.DiscountAmount)
		workingTotal = roundMoney(workingTotal - discount)
		if workingTotal < 0 {
			workingTotal = 0
		}
		result.EligibleCoupons = append(result.EligibleCoupons, AppliedCoupon{
			CouponID:       appliedNewUserGift.CouponID,
			Code:           appliedNewUserGift.Code,
			Name:           appliedNewUserGift.Name,
			DiscountAmount: discount,
		})
		result.AppliedDiscounts = append(result.AppliedDiscounts, AppliedDiscount{
			Type:        "new_user_gift",
			Description: fmt.Sprintf("New user gift: %s", appliedNewUserGift.Name),
			Amount:      discount,
		})
	} else {
		if appliedThreshold != nil {
			discount := roundMoney(appliedThreshold.DiscountAmount)
			workingTotal = roundMoney(workingTotal - discount)
			if workingTotal < 0 {
				workingTotal = 0
			}
			result.EligibleCoupons = append(result.EligibleCoupons, AppliedCoupon{
				CouponID:       appliedThreshold.CouponID,
				Code:           appliedThreshold.Code,
				Name:           appliedThreshold.Name,
				DiscountAmount: discount,
			})
			result.AppliedDiscounts = append(result.AppliedDiscounts, AppliedDiscount{
				Type:        "threshold_fixed",
				Description: fmt.Sprintf("Threshold discount: %s", appliedThreshold.Name),
				Amount:      discount,
			})
		}

		if appliedPercentage != nil {
			discount := roundMoney(appliedPercentage.DiscountAmount)
			workingTotal = roundMoney(workingTotal - discount)
			if workingTotal < 0 {
				workingTotal = 0
			}
			result.EligibleCoupons = append(result.EligibleCoupons, AppliedCoupon{
				CouponID:       appliedPercentage.CouponID,
				Code:           appliedPercentage.Code,
				Name:           appliedPercentage.Name,
				DiscountAmount: discount,
			})
			result.AppliedDiscounts = append(result.AppliedDiscounts, AppliedDiscount{
				Type:        "percentage",
				Description: fmt.Sprintf("Percentage discount: %s", appliedPercentage.Name),
				Amount:      discount,
			})
		}

		for _, e := range eligible {
			coupon := couponMap[e.Code]
			if coupon.DiscountType == "threshold_fixed" && appliedThreshold != nil && e.CouponID != appliedThreshold.CouponID {
				result.IneligibleCoupons = append(result.IneligibleCoupons, IneligibleCoupon{
					CouponID:   e.CouponID,
					Code:       e.Code,
					Name:       e.Name,
					ReasonCode: ReasonStackingNotAllowed,
					Message:    "Only one threshold discount can be applied",
				})
			}
			if coupon.DiscountType == "percentage" && appliedPercentage != nil && e.CouponID != appliedPercentage.CouponID {
				result.IneligibleCoupons = append(result.IneligibleCoupons, IneligibleCoupon{
					CouponID:   e.CouponID,
					Code:       e.Code,
					Name:       e.Name,
					ReasonCode: ReasonStackingNotAllowed,
					Message:    "Only one percentage discount can be applied",
				})
			}
		}
	}

	result.TotalDiscount = roundMoney(subtotal - workingTotal)
	result.FinalPayable = roundMoney(workingTotal)
	result.EscrowHoldAmount = result.FinalPayable

	return result, nil
}

func evaluateSingleCoupon(
	ctx context.Context,
	repo *Repository,
	coupon Coupon,
	items []BookingItem,
	currentTotal float64,
	userID string,
	isNewUser bool,
	membershipTier string,
	scopeKey string,
	now time.Time,
) CouponEligibilityResult {
	res := CouponEligibilityResult{
		CouponID: coupon.ID,
		Code:     coupon.Code,
		Name:     coupon.Name,
	}

	if !coupon.Active {
		res.ReasonCode = ReasonInactive
		res.Reason = FormatReason(ReasonInactive)
		return res
	}

	if ok, code := checkDateRange(coupon, now); !ok {
		res.ReasonCode = code
		res.Reason = FormatReason(code)
		return res
	}

	if ok, code := checkMinSpend(coupon, currentTotal); !ok {
		res.ReasonCode = code
		res.Reason = FormatReason(code, coupon.MinSpend)
		return res
	}

	if ok, code := checkCategoryRestriction(coupon, items); !ok {
		res.ReasonCode = code
		res.Reason = FormatReason(code)
		return res
	}

	if ok, code := checkNewUserOnly(coupon, isNewUser); !ok {
		res.ReasonCode = code
		res.Reason = FormatReason(code)
		return res
	}

	if ok, code := checkMembershipRequired(coupon, membershipTier); !ok {
		elig := parseCouponEligibility(coupon)
		res.ReasonCode = code
		res.Reason = FormatReason(code, elig.MembershipRequired)
		return res
	}

	if ok, code := checkUsageLimitTotal(ctx, repo, coupon.ID, coupon.UsageLimitTotal); !ok {
		res.ReasonCode = code
		res.Reason = FormatReason(code)
		return res
	}

	if ok, code := checkUsageLimitPerUser(ctx, repo, coupon.ID, userID, coupon.UsageLimitPerUser); !ok {
		res.ReasonCode = code
		res.Reason = FormatReason(code)
		return res
	}

	if ok, code := checkAlreadyRedeemed(ctx, repo, coupon.ID, userID, scopeKey); !ok {
		res.ReasonCode = code
		res.Reason = FormatReason(code)
		return res
	}

	res.Eligible = true
	switch coupon.DiscountType {
	case "threshold_fixed":
		res.DiscountAmount = roundMoney(coupon.Amount)
	case "percentage":
		res.DiscountAmount = roundMoney(currentTotal * coupon.PercentOff / 100)
	case "new_user_gift":
		res.DiscountAmount = roundMoney(coupon.Amount)
	}

	return res
}

func roundMoney(v float64) float64 {
	return math.Round(v*100) / 100
}
