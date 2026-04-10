package pricing

import (
	"context"
	"encoding/json"
	"time"
)

func checkDateRange(coupon Coupon, now time.Time) (bool, string) {
	if coupon.ValidFrom != nil && now.Before(*coupon.ValidFrom) {
		return false, ReasonDateOutOfRange
	}
	if coupon.ValidTo != nil && now.After(*coupon.ValidTo) {
		return false, ReasonExpired
	}
	return true, ""
}

func checkMinSpend(coupon Coupon, subtotal float64) (bool, string) {
	if coupon.MinSpend > 0 && subtotal < coupon.MinSpend {
		return false, ReasonMinSpendNotMet
	}
	return true, ""
}

func checkCategoryRestriction(coupon Coupon, items []BookingItem) (bool, string) {
	elig := parseCouponEligibility(coupon)
	if len(elig.Categories) == 0 {
		return true, ""
	}

	allowed := make(map[string]bool, len(elig.Categories))
	for _, cat := range elig.Categories {
		allowed[cat] = true
	}

	for _, item := range items {
		if allowed[item.Category] {
			return true, ""
		}
	}
	return false, ReasonCategoryRestricted
}

func checkNewUserOnly(coupon Coupon, isNewUser bool) (bool, string) {
	elig := parseCouponEligibility(coupon)
	if elig.NewUserOnly && !isNewUser {
		return false, ReasonNewUserOnly
	}
	return true, ""
}

func checkMembershipRequired(coupon Coupon, membershipTier string) (bool, string) {
	elig := parseCouponEligibility(coupon)
	if elig.MembershipRequired != "" && membershipTier != elig.MembershipRequired {
		return false, ReasonMembershipRequired
	}
	return true, ""
}

func checkUsageLimitTotal(ctx context.Context, repo *Repository, couponID string, limit *int) (bool, string) {
	if limit == nil {
		return true, ""
	}
	count, err := repo.GetCouponRedemptionCount(ctx, couponID)
	if err != nil {
		return false, ReasonUsageLimitReached
	}
	if count >= *limit {
		return false, ReasonUsageLimitReached
	}
	return true, ""
}

func checkUsageLimitPerUser(ctx context.Context, repo *Repository, couponID, userID string, limit *int) (bool, string) {
	if limit == nil {
		return true, ""
	}
	count, err := repo.GetUserCouponRedemptionCount(ctx, couponID, userID)
	if err != nil {
		return false, ReasonUsageLimitReached
	}
	if count >= *limit {
		return false, ReasonUsageLimitReached
	}
	return true, ""
}

func checkAlreadyRedeemed(ctx context.Context, repo *Repository, couponID, userID, scopeKey string) (bool, string) {
	if scopeKey == "" {
		return true, ""
	}
	exists, err := repo.HasRedemption(ctx, couponID, userID, scopeKey)
	if err != nil {
		return false, ReasonAlreadyRedeemed
	}
	if exists {
		return false, ReasonAlreadyRedeemed
	}
	return true, ""
}

func parseCouponEligibility(coupon Coupon) CouponEligibility {
	var elig CouponEligibility
	if len(coupon.EligibilityJSON) > 0 {
		_ = json.Unmarshal(coupon.EligibilityJSON, &elig)
	}
	return elig
}
