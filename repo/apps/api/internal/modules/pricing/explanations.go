package pricing

import "fmt"

const (
	ReasonDateOutOfRange      = "DATE_OUT_OF_RANGE"
	ReasonMinSpendNotMet      = "MIN_SPEND_NOT_MET"
	ReasonMembershipRequired  = "MEMBERSHIP_REQUIRED"
	ReasonNewUserOnly         = "NEW_USER_ONLY"
	ReasonStackingNotAllowed  = "STACKING_NOT_ALLOWED"
	ReasonUsageLimitReached   = "USAGE_LIMIT_REACHED"
	ReasonExpired             = "EXPIRED"
	ReasonInactive            = "INACTIVE"
	ReasonCategoryRestricted  = "CATEGORY_RESTRICTED"
	ReasonAlreadyRedeemed     = "ALREADY_REDEEMED"
	ReasonRoomTypeRestricted  = "ROOM_TYPE_RESTRICTED"
)

var ReasonMessages = map[string]string{
	ReasonDateOutOfRange:     "This coupon is not valid for the selected dates",
	ReasonMinSpendNotMet:     "Minimum spend of $%.2f is required",
	ReasonMembershipRequired: "A %s membership is required to use this coupon",
	ReasonNewUserOnly:        "This coupon is only available for new users",
	ReasonStackingNotAllowed: "New user gift is exclusive and cannot be combined with other discounts",
	ReasonUsageLimitReached:  "This coupon has reached its usage limit",
	ReasonExpired:            "This coupon has expired",
	ReasonInactive:           "This coupon is no longer active",
	ReasonCategoryRestricted: "This coupon does not apply to the items in your cart",
	ReasonAlreadyRedeemed:    "You have already redeemed this coupon",
	ReasonRoomTypeRestricted: "This coupon is not valid for the selected room type",
}

func FormatReason(code string, args ...interface{}) string {
	tmpl, ok := ReasonMessages[code]
	if !ok {
		return code
	}
	if len(args) > 0 {
		return fmt.Sprintf(tmpl, args...)
	}
	return tmpl
}
