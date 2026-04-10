package procurement

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"

	"travel-platform/apps/api/internal/common"
	"travel-platform/apps/api/internal/middleware"
)

type Handler struct {
	svc    *Service
	logger *zap.Logger
}

func NewHandler(svc *Service, logger *zap.Logger) *Handler {
	return &Handler{svc: svc, logger: logger}
}

func (h *Handler) RegisterRoutes(g *echo.Group) {
	accountantAdmin := middleware.RequireRole(common.RoleAccountant, common.RoleAdministrator)
	rfqCreator := middleware.RequireRole(common.RoleAccountant, common.RoleAdministrator, common.RoleGroupOrganizer)
	supplier := middleware.RequireRole(common.RoleSupplier, common.RoleAdministrator)

	g.GET("/supplier-quotes", h.ListSupplierQuotes)
	g.GET("/rfqs", h.ListRFQs)
	g.POST("/rfqs", h.CreateRFQ, rfqCreator)
	g.GET("/rfqs/:id", h.GetRFQ)
	g.POST("/rfqs/:id/issue", h.IssueRFQ, accountantAdmin)
	g.POST("/rfqs/:id/quotes", h.SubmitQuote, supplier)
	g.GET("/rfqs/:id/comparison", h.GetComparison, accountantAdmin)
	g.POST("/rfqs/:id/select", h.SelectSupplier, accountantAdmin)

	g.GET("/purchase-orders", h.ListPOs)
	g.POST("/purchase-orders", h.CreatePO, accountantAdmin)
	g.GET("/purchase-orders/:id", h.GetPO)
	g.POST("/purchase-orders/:id/accept", h.AcceptPO, supplier)
	g.POST("/purchase-orders/:id/deliveries", h.RecordDelivery, accountantAdmin)

	g.GET("/deliveries", h.ListDeliveries)
	g.POST("/deliveries/:id/inspect", h.PerformInspection, accountantAdmin)
	g.POST("/discrepancies", h.CreateDiscrepancy, accountantAdmin)

	g.GET("/exceptions", h.ListExceptions)
	g.POST("/exceptions/:id/waivers", h.SubmitWaiver, accountantAdmin)
	g.POST("/exceptions/:id/settlement-adjustments", h.SubmitSettlementAdjustment, accountantAdmin)
	g.POST("/exceptions/:id/close", h.CloseException, accountantAdmin)
}

func (h *Handler) CreateRFQ(c echo.Context) error {
	var req CreateRFQRequest
	if err := c.Bind(&req); err != nil {
		return common.Error(c, http.StatusBadRequest, common.ErrCodeBadRequest, "invalid request body")
	}
	userID := common.GetUserID(c)
	resp, err := h.svc.CreateRFQ(c.Request().Context(), userID, req)
	if err != nil {
		return h.handleError(c, err)
	}
	return common.Created(c, resp)
}

func (h *Handler) GetRFQ(c echo.Context) error {
	id := c.Param("id")
	userID := common.GetUserID(c)
	roles := common.GetRoles(c)
	resp, err := h.svc.GetRFQ(c.Request().Context(), id, userID, roles)
	if err != nil {
		return h.handleError(c, err)
	}
	return common.Success(c, resp)
}

func (h *Handler) IssueRFQ(c echo.Context) error {
	id := c.Param("id")
	var req IssueRFQRequest
	if err := c.Bind(&req); err != nil {
		return common.Error(c, http.StatusBadRequest, common.ErrCodeBadRequest, "invalid request body")
	}
	userID := common.GetUserID(c)
	if err := h.svc.IssueRFQ(c.Request().Context(), id, userID, req); err != nil {
		return h.handleError(c, err)
	}
	return common.Success(c, map[string]string{"message": "rfq issued"})
}

func (h *Handler) SubmitQuote(c echo.Context) error {
	rfqID := c.Param("id")
	var req SubmitQuoteRequest
	if err := c.Bind(&req); err != nil {
		return common.Error(c, http.StatusBadRequest, common.ErrCodeBadRequest, "invalid request body")
	}
	supplierID := common.GetUserID(c)
	resp, err := h.svc.SubmitQuote(c.Request().Context(), rfqID, supplierID, req)
	if err != nil {
		return h.handleError(c, err)
	}
	return common.Created(c, resp)
}

func (h *Handler) GetComparison(c echo.Context) error {
	rfqID := c.Param("id")
	userID := common.GetUserID(c)
	resp, err := h.svc.GetComparison(c.Request().Context(), rfqID, userID)
	if err != nil {
		return h.handleError(c, err)
	}
	return common.Success(c, resp)
}

func (h *Handler) SelectSupplier(c echo.Context) error {
	rfqID := c.Param("id")
	var req SelectQuoteRequest
	if err := c.Bind(&req); err != nil {
		return common.Error(c, http.StatusBadRequest, common.ErrCodeBadRequest, "invalid request body")
	}
	userID := common.GetUserID(c)
	if err := h.svc.SelectSupplier(c.Request().Context(), rfqID, userID, req.QuoteID); err != nil {
		return h.handleError(c, err)
	}
	return common.Success(c, map[string]string{"message": "supplier selected"})
}

func (h *Handler) CreatePO(c echo.Context) error {
	var req CreatePORequest
	if err := c.Bind(&req); err != nil {
		return common.Error(c, http.StatusBadRequest, common.ErrCodeBadRequest, "invalid request body")
	}
	userID := common.GetUserID(c)
	resp, err := h.svc.CreatePO(c.Request().Context(), userID, req)
	if err != nil {
		return h.handleError(c, err)
	}
	return common.Created(c, resp)
}

func (h *Handler) GetPO(c echo.Context) error {
	id := c.Param("id")
	userID := common.GetUserID(c)
	roles := common.GetRoles(c)
	resp, err := h.svc.GetPO(c.Request().Context(), id, userID, roles)
	if err != nil {
		return h.handleError(c, err)
	}
	return common.Success(c, resp)
}

func (h *Handler) AcceptPO(c echo.Context) error {
	id := c.Param("id")
	supplierID := common.GetUserID(c)
	if err := h.svc.AcceptPO(c.Request().Context(), id, supplierID); err != nil {
		return h.handleError(c, err)
	}
	return common.Success(c, map[string]string{"message": "purchase order accepted"})
}

func (h *Handler) RecordDelivery(c echo.Context) error {
	poID := c.Param("id")
	var req RecordDeliveryRequest
	if err := c.Bind(&req); err != nil {
		return common.Error(c, http.StatusBadRequest, common.ErrCodeBadRequest, "invalid request body")
	}
	userID := common.GetUserID(c)
	resp, err := h.svc.RecordDelivery(c.Request().Context(), poID, userID, req)
	if err != nil {
		return h.handleError(c, err)
	}
	return common.Created(c, resp)
}

func (h *Handler) PerformInspection(c echo.Context) error {
	deliveryID := c.Param("id")
	var req InspectionRequest
	if err := c.Bind(&req); err != nil {
		return common.Error(c, http.StatusBadRequest, common.ErrCodeBadRequest, "invalid request body")
	}
	inspectorID := common.GetUserID(c)
	resp, err := h.svc.PerformInspection(c.Request().Context(), deliveryID, inspectorID, req)
	if err != nil {
		return h.handleError(c, err)
	}
	return common.Created(c, resp)
}

func (h *Handler) CreateDiscrepancy(c echo.Context) error {
	var req CreateDiscrepancyRequest
	if err := c.Bind(&req); err != nil {
		return common.Error(c, http.StatusBadRequest, common.ErrCodeBadRequest, "invalid request body")
	}
	userID := common.GetUserID(c)
	if err := h.svc.CreateDiscrepancy(c.Request().Context(), req.POID, userID, req); err != nil {
		return h.handleError(c, err)
	}
	return common.Created(c, map[string]string{"message": "discrepancy created"})
}

func (h *Handler) SubmitWaiver(c echo.Context) error {
	exceptionID := c.Param("id")
	var req WaiverRequest
	if err := c.Bind(&req); err != nil {
		return common.Error(c, http.StatusBadRequest, common.ErrCodeBadRequest, "invalid request body")
	}
	userID := common.GetUserID(c)
	if err := h.svc.SubmitWaiver(c.Request().Context(), exceptionID, userID, req.WaiverReason); err != nil {
		return h.handleError(c, err)
	}
	return common.Created(c, map[string]string{"message": "waiver submitted"})
}

func (h *Handler) SubmitSettlementAdjustment(c echo.Context) error {
	exceptionID := c.Param("id")
	var req SettlementAdjustmentRequest
	if err := c.Bind(&req); err != nil {
		return common.Error(c, http.StatusBadRequest, common.ErrCodeBadRequest, "invalid request body")
	}
	userID := common.GetUserID(c)
	if err := h.svc.SubmitSettlementAdjustment(c.Request().Context(), exceptionID, userID, req.Amount, req.Direction, req.Reason); err != nil {
		return h.handleError(c, err)
	}
	return common.Created(c, map[string]string{"message": "settlement adjustment submitted"})
}

func (h *Handler) CloseException(c echo.Context) error {
	exceptionID := c.Param("id")
	userID := common.GetUserID(c)
	if err := h.svc.CloseException(c.Request().Context(), exceptionID, userID); err != nil {
		return h.handleError(c, err)
	}
	return common.Success(c, map[string]string{"message": "exception closed"})
}

func (h *Handler) ListRFQs(c echo.Context) error {
	userID := common.GetUserID(c)
	roles := common.GetRoles(c)
	result, err := h.svc.ListRFQs(c.Request().Context(), userID, roles)
	if err != nil {
		return h.handleError(c, err)
	}
	return common.Success(c, result)
}

func (h *Handler) ListPOs(c echo.Context) error {
	userID := common.GetUserID(c)
	roles := common.GetRoles(c)
	result, err := h.svc.ListPOs(c.Request().Context(), userID, roles)
	if err != nil {
		return h.handleError(c, err)
	}
	return common.Success(c, result)
}

func (h *Handler) ListDeliveries(c echo.Context) error {
	userID := common.GetUserID(c)
	roles := common.GetRoles(c)
	result, err := h.svc.ListDeliveries(c.Request().Context(), userID, roles)
	if err != nil {
		return h.handleError(c, err)
	}
	return common.Success(c, result)
}

func (h *Handler) ListExceptions(c echo.Context) error {
	userID := common.GetUserID(c)
	roles := common.GetRoles(c)
	result, err := h.svc.ListExceptions(c.Request().Context(), userID, roles)
	if err != nil {
		return h.handleError(c, err)
	}
	return common.Success(c, result)
}

func (h *Handler) ListSupplierQuotes(c echo.Context) error {
	userID := common.GetUserID(c)
	quotes, err := h.svc.ListSupplierQuotes(c.Request().Context(), userID)
	if err != nil {
		return h.handleError(c, err)
	}
	return common.Success(c, quotes)
}

func (h *Handler) handleError(c echo.Context, err error) error {
	var domainErr *common.DomainError
	if errors.As(err, &domainErr) {
		switch domainErr.Code {
		case common.ErrCodeNotFound:
			return common.Error(c, http.StatusNotFound, domainErr.Code, domainErr.Message)
		case common.ErrCodeForbidden:
			return common.Error(c, http.StatusForbidden, domainErr.Code, domainErr.Message)
		case common.ErrCodeConflict:
			return common.Error(c, http.StatusConflict, domainErr.Code, domainErr.Message)
		case common.ErrCodeBadRequest:
			return common.Error(c, http.StatusBadRequest, domainErr.Code, domainErr.Message)
		case common.ErrCodeValidation:
			return common.Error(c, http.StatusUnprocessableEntity, domainErr.Code, domainErr.Message)
		default:
			h.logger.Error("domain error", zap.String("code", domainErr.Code), zap.Error(err))
			return common.Error(c, http.StatusInternalServerError, common.ErrCodeInternal, "internal server error")
		}
	}
	h.logger.Error("unhandled error", zap.Error(err))
	return common.Error(c, http.StatusInternalServerError, common.ErrCodeInternal, "internal server error")
}
