package pricing

import (
	"net/http"

	"travel-platform/apps/api/internal/common"

	"github.com/labstack/echo/v4"
)

type Handler struct {
	repo *Repository
}

func NewHandler(repo *Repository) *Handler {
	return &Handler{repo: repo}
}

func (h *Handler) RegisterRoutes(g *echo.Group) {
	coupons := g.Group("/coupons")
	coupons.GET("/available", h.ListActiveCoupons)
	coupons.POST("/evaluate", h.EvaluateCoupons)
	coupons.POST("/redeem-preview", h.RedeemPreview)
}

type EvaluateCouponsRequest struct {
	CouponCodes    []string      `json:"couponCodes"`
	Items          []BookingItem `json:"items"`
	MembershipTier string        `json:"membershipTier"`
	IsNewUser      bool          `json:"isNewUser"`
}

type RedeemPreviewRequest struct {
	CouponCodes    []string      `json:"couponCodes"`
	Items          []BookingItem `json:"items"`
	MembershipTier string        `json:"membershipTier"`
	IsNewUser      bool          `json:"isNewUser"`
}

func (h *Handler) ListActiveCoupons(c echo.Context) error {
	coupons, err := h.repo.GetAllActiveCoupons(c.Request().Context())
	if err != nil {
		return common.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to load coupons")
	}
	return common.Success(c, coupons)
}

func (h *Handler) EvaluateCoupons(c echo.Context) error {
	var req EvaluateCouponsRequest
	if err := c.Bind(&req); err != nil {
		return common.Error(c, http.StatusBadRequest, "BAD_REQUEST", "invalid request body")
	}

	if len(req.CouponCodes) == 0 {
		return common.ValidationError(c, map[string]string{
			"couponCodes": "at least one coupon code is required",
		})
	}
	if len(req.Items) == 0 {
		return common.ValidationError(c, map[string]string{
			"items": "at least one item is required",
		})
	}

	userID := common.GetUserID(c)

	result, err := EvaluateCheckout(
		c.Request().Context(),
		h.repo,
		req.Items,
		req.CouponCodes,
		userID,
		req.IsNewUser,
		req.MembershipTier,
		"",
	)
	if err != nil {
		return common.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "pricing evaluation failed")
	}

	return common.Success(c, result)
}

func (h *Handler) RedeemPreview(c echo.Context) error {
	var req RedeemPreviewRequest
	if err := c.Bind(&req); err != nil {
		return common.Error(c, http.StatusBadRequest, "BAD_REQUEST", "invalid request body")
	}

	if len(req.Items) == 0 {
		return common.ValidationError(c, map[string]string{
			"items": "at least one item is required",
		})
	}

	userID := common.GetUserID(c)

	result, err := EvaluateCheckout(
		c.Request().Context(),
		h.repo,
		req.Items,
		req.CouponCodes,
		userID,
		req.IsNewUser,
		req.MembershipTier,
		"",
	)
	if err != nil {
		return common.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "pricing evaluation failed")
	}

	snapshotJSON, err := CreateSnapshot(result, req.Items)
	if err != nil {
		return common.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to create pricing snapshot")
	}

	snapshotID, err := h.repo.SavePricingSnapshot(c.Request().Context(), nil, snapshotJSON)
	if err != nil {
		return common.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to save pricing snapshot")
	}

	result.SnapshotID = snapshotID
	return common.Success(c, result)
}
