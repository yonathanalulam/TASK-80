package bookings

import (
	"net/http"
	"strconv"

	"travel-platform/apps/api/internal/common"

	"github.com/labstack/echo/v4"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(g *echo.Group) {
	g.GET("", h.ListBookings)
	g.POST("", h.Create)
	g.GET("/:id", h.GetByID)
	g.POST("/:id/price-preview", h.PricePreview)
	g.POST("/:id/checkout", h.Checkout)
	g.POST("/:id/record-tender", h.RecordTender)
	g.POST("/:id/cancel", h.Cancel)
	g.POST("/:id/complete", h.Complete)
}

func (h *Handler) Create(c echo.Context) error {
	var req CreateBookingRequest
	if err := c.Bind(&req); err != nil {
		return common.Error(c, http.StatusBadRequest, "BAD_REQUEST", "invalid request body")
	}

	errs := make(map[string]string)
	if req.Title == "" {
		errs["title"] = "title is required"
	}
	if len(req.Items) == 0 {
		errs["items"] = "at least one item is required"
	}
	for i, item := range req.Items {
		if item.ItemName == "" {
			errs[formatItemField(i, "itemName")] = "item name is required"
		}
		if item.UnitPrice <= 0 {
			errs[formatItemField(i, "unitPrice")] = "unit price must be positive"
		}
		if item.Quantity <= 0 {
			errs[formatItemField(i, "quantity")] = "quantity must be positive"
		}
	}
	if len(errs) > 0 {
		return common.ValidationError(c, errs)
	}

	userID := common.GetUserID(c)
	if userID == "" {
		return common.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", "not authenticated")
	}

	resp, err := h.service.CreateBooking(c.Request().Context(), userID, req)
	if err != nil {
		return common.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
	}

	return common.Created(c, resp)
}

func (h *Handler) GetByID(c echo.Context) error {
	id := c.Param("id")
	userID := common.GetUserID(c)
	if userID == "" {
		return common.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", "not authenticated")
	}

	resp, err := h.service.GetBooking(c.Request().Context(), id, userID)
	if err != nil {
		if err.Error() == "booking not found" {
			return common.Error(c, http.StatusNotFound, "NOT_FOUND", "booking not found")
		}
		if err.Error() == "forbidden: not the organizer" {
			return common.Error(c, http.StatusForbidden, "FORBIDDEN", "you do not own this booking")
		}
		return common.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
	}

	return common.Success(c, resp)
}

type PricePreviewRequest struct {
	CouponCodes    []string `json:"couponCodes"`
	MembershipTier string   `json:"membershipTier"`
	IsNewUser      bool     `json:"isNewUser"`
}

func (h *Handler) PricePreview(c echo.Context) error {
	id := c.Param("id")
	userID := common.GetUserID(c)
	if userID == "" {
		return common.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", "not authenticated")
	}

	var req PricePreviewRequest
	if err := c.Bind(&req); err != nil {
		return common.Error(c, http.StatusBadRequest, "BAD_REQUEST", "invalid request body")
	}

	result, err := h.service.PricePreview(c.Request().Context(), id, userID, req.CouponCodes, req.IsNewUser, req.MembershipTier)
	if err != nil {
		if err.Error() == "booking not found" {
			return common.Error(c, http.StatusNotFound, "NOT_FOUND", "booking not found")
		}
		if err.Error() == "forbidden: not the organizer" {
			return common.Error(c, http.StatusForbidden, "FORBIDDEN", "you do not own this booking")
		}
		return common.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
	}

	return common.Success(c, result)
}

func (h *Handler) Checkout(c echo.Context) error {
	id := c.Param("id")
	userID := common.GetUserID(c)
	if userID == "" {
		return common.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", "not authenticated")
	}

	var req CheckoutRequest
	if err := c.Bind(&req); err != nil {
		return common.Error(c, http.StatusBadRequest, "BAD_REQUEST", "invalid request body")
	}

	if req.IdempotencyKey == "" {
		req.IdempotencyKey = c.Request().Header.Get("Idempotency-Key")
	}
	if req.IdempotencyKey == "" {
		return common.Error(c, http.StatusBadRequest, "BAD_REQUEST", "Idempotency-Key header is required")
	}

	resp, err := h.service.Checkout(c.Request().Context(), id, userID, req)
	if err != nil {
		msg := err.Error()
		switch msg {
		case "booking not found":
			return common.Error(c, http.StatusNotFound, "NOT_FOUND", msg)
		case "forbidden: not the organizer":
			return common.Error(c, http.StatusForbidden, "FORBIDDEN", "you do not own this booking")
		case "idempotency key reused with different request body":
			return common.Error(c, http.StatusConflict, "CONFLICT", msg)
		case "idempotency key is required":
			return common.Error(c, http.StatusBadRequest, "BAD_REQUEST", msg)
		case "booking has no items":
			return common.Error(c, http.StatusBadRequest, "BAD_REQUEST", msg)
		}
		if len(msg) > 7 && msg[:7] == "booking" {
			return common.Error(c, http.StatusConflict, "CONFLICT", msg)
		}
		return common.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "checkout failed")
	}

	return common.Success(c, resp)
}

func (h *Handler) RecordTender(c echo.Context) error {
	id := c.Param("id")
	userID := common.GetUserID(c)
	if userID == "" {
		return common.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", "not authenticated")
	}

	var req TenderRecordRequest
	if err := c.Bind(&req); err != nil {
		return common.Error(c, http.StatusBadRequest, "BAD_REQUEST", "invalid request body")
	}

	errs := make(map[string]string)
	if req.TenderType == "" {
		errs["tenderType"] = "tender type is required"
	}
	if req.Amount <= 0 {
		errs["amount"] = "amount must be positive"
	}
	if len(errs) > 0 {
		return common.ValidationError(c, errs)
	}

	err := h.service.RecordTender(c.Request().Context(), id, userID, req)
	if err != nil {
		if err.Error() == "booking not found" {
			return common.Error(c, http.StatusNotFound, "NOT_FOUND", "booking not found")
		}
		if err.Error() == "forbidden: not the organizer" {
			return common.Error(c, http.StatusForbidden, "FORBIDDEN", "you do not own this booking")
		}
		return common.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
	}

	return common.Success(c, map[string]string{"message": "tender recorded"})
}

func (h *Handler) Cancel(c echo.Context) error {
	id := c.Param("id")
	userID := common.GetUserID(c)
	if userID == "" {
		return common.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", "not authenticated")
	}

	err := h.service.CancelBooking(c.Request().Context(), id, userID)
	if err != nil {
		if err.Error() == "booking not found" {
			return common.Error(c, http.StatusNotFound, "NOT_FOUND", "booking not found")
		}
		if err.Error() == "forbidden: not the organizer" {
			return common.Error(c, http.StatusForbidden, "FORBIDDEN", "you do not own this booking")
		}
		return common.Error(c, http.StatusConflict, "CONFLICT", err.Error())
	}

	return common.Success(c, map[string]string{"message": "booking cancelled"})
}

func (h *Handler) Complete(c echo.Context) error {
	id := c.Param("id")
	userID := common.GetUserID(c)
	if userID == "" {
		return common.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", "not authenticated")
	}

	err := h.service.CompleteBooking(c.Request().Context(), id, userID)
	if err != nil {
		if err.Error() == "booking not found" {
			return common.Error(c, http.StatusNotFound, "NOT_FOUND", "booking not found")
		}
		if err.Error() == "forbidden: not the organizer" {
			return common.Error(c, http.StatusForbidden, "FORBIDDEN", "you do not own this booking")
		}
		return common.Error(c, http.StatusConflict, "CONFLICT", err.Error())
	}

	return common.Success(c, map[string]string{"message": "booking completed"})
}

func (h *Handler) ListBookings(c echo.Context) error {
	userID := common.GetUserID(c)
	page, _ := strconv.Atoi(c.QueryParam("page"))
	pageSize, _ := strconv.Atoi(c.QueryParam("pageSize"))
	result, err := h.service.ListBookings(c.Request().Context(), userID, page, pageSize)
	if err != nil {
		return common.Error(c, http.StatusInternalServerError, common.ErrCodeInternal, "failed to list bookings")
	}
	return common.Success(c, result)
}

func formatItemField(index int, field string) string {
	return "items[" + itoa(index) + "]." + field
}

func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	s := ""
	for i > 0 {
		s = string(rune('0'+i%10)) + s
		i /= 10
	}
	return s
}
