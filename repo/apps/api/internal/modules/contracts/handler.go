package contracts

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"

	"travel-platform/apps/api/internal/common"
	"travel-platform/apps/api/internal/middleware"
)

type Handler struct {
	service *ContractService
	logger  *zap.Logger
}

func NewHandler(service *ContractService, logger *zap.Logger) *Handler {
	return &Handler{service: service, logger: logger}
}

func RegisterRoutes(g *echo.Group, h *Handler) {
	g.GET("/contract-templates", h.ListTemplates)
	g.POST("/contracts/generate", h.GenerateContract)
	g.POST("/invoice-requests", h.RequestInvoice)
	g.POST("/invoice-requests/:id/approve", h.ApproveInvoiceRequest, middleware.RequireRole(common.RoleAdministrator, common.RoleAccountant))
	g.POST("/invoices/:id/generate", h.GenerateInvoice, middleware.RequireRole(common.RoleAdministrator, common.RoleAccountant))
	g.GET("/invoice-requests", h.ListInvoiceRequests)
}

func (h *Handler) ListTemplates(c echo.Context) error {
	templates, err := h.service.GetTemplates(c.Request().Context())
	if err != nil {
		return handleServiceError(c, err)
	}

	if templates == nil {
		templates = []ContractTemplate{}
	}

	return common.Success(c, templates)
}

func (h *Handler) GenerateContract(c echo.Context) error {
	userID := common.GetUserID(c)
	if userID == "" {
		return common.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required")
	}

	var req GenerateContractRequest
	if err := c.Bind(&req); err != nil {
		return common.Error(c, http.StatusBadRequest, "BAD_REQUEST", "invalid request body")
	}

	result, err := h.service.GenerateContract(c.Request().Context(), userID, req)
	if err != nil {
		return handleServiceError(c, err)
	}

	return common.Created(c, result)
}

func (h *Handler) RequestInvoice(c echo.Context) error {
	userID := common.GetUserID(c)
	if userID == "" {
		return common.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required")
	}

	var req CreateInvoiceRequestDTO
	if err := c.Bind(&req); err != nil {
		return common.Error(c, http.StatusBadRequest, "BAD_REQUEST", "invalid request body")
	}

	ir, err := h.service.RequestInvoice(c.Request().Context(), userID, req)
	if err != nil {
		return handleServiceError(c, err)
	}

	return common.Created(c, ir)
}

func (h *Handler) ApproveInvoiceRequest(c echo.Context) error {
	userID := common.GetUserID(c)
	if userID == "" {
		return common.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required")
	}

	requestID := c.Param("id")
	if requestID == "" {
		return common.Error(c, http.StatusBadRequest, "BAD_REQUEST", "request ID is required")
	}

	if err := h.service.ApproveInvoiceRequest(c.Request().Context(), requestID, userID); err != nil {
		return handleServiceError(c, err)
	}

	return common.Success(c, map[string]string{"message": "invoice request approved"})
}

func (h *Handler) GenerateInvoice(c echo.Context) error {
	userID := common.GetUserID(c)
	if userID == "" {
		return common.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required")
	}

	requestID := c.Param("id")
	if requestID == "" {
		return common.Error(c, http.StatusBadRequest, "BAD_REQUEST", "request ID is required")
	}

	result, err := h.service.GenerateInvoice(c.Request().Context(), requestID, userID)
	if err != nil {
		return handleServiceError(c, err)
	}

	return common.Created(c, result)
}

func (h *Handler) ListInvoiceRequests(c echo.Context) error {
	userID := common.GetUserID(c)
	if userID == "" {
		return common.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required")
	}

	roles := common.GetRoles(c)
	isAdmin := false
	for _, r := range roles {
		if r == common.RoleAdministrator {
			isAdmin = true
			break
		}
	}

	requests, err := h.service.GetInvoiceRequests(c.Request().Context(), userID, isAdmin)
	if err != nil {
		return handleServiceError(c, err)
	}

	if requests == nil {
		requests = []InvoiceRequest{}
	}

	return common.Success(c, requests)
}

func handleServiceError(c echo.Context, err error) error {
	if de, ok := err.(*common.DomainError); ok {
		switch de.Code {
		case common.ErrCodeNotFound:
			return common.Error(c, http.StatusNotFound, de.Code, de.Message)
		case common.ErrCodeForbidden:
			return common.Error(c, http.StatusForbidden, de.Code, de.Message)
		case common.ErrCodeConflict:
			return common.Error(c, http.StatusConflict, de.Code, de.Message)
		case common.ErrCodeBadRequest:
			return common.Error(c, http.StatusBadRequest, de.Code, de.Message)
		case common.ErrCodeValidation:
			return common.Error(c, http.StatusUnprocessableEntity, de.Code, de.Message)
		}
	}
	return common.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "an unexpected error occurred")
}
