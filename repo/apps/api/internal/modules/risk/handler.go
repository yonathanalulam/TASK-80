package risk

import (
	"errors"
	"net/http"

	"travel-platform/apps/api/internal/common"
	"travel-platform/apps/api/internal/middleware"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

type Handler struct {
	svc    *Service
	logger *zap.Logger
}

func NewHandler(svc *Service, logger *zap.Logger) *Handler {
	return &Handler{svc: svc, logger: logger}
}

func (h *Handler) RegisterRoutes(g *echo.Group, adminGroup *echo.Group) {
	adminOnly := middleware.RequireRole(common.RoleAdministrator)

	g.GET("/risk/:userId", h.GetRiskSummary, adminOnly)

	admin := adminGroup.Group("", adminOnly)
	admin.GET("/approvals", h.GetPendingApprovals)
	admin.POST("/approvals/:id/resolve", h.ResolveApproval)
	admin.POST("/users/:id/blacklist", h.BlacklistUser)
	admin.POST("/users/:id/unblacklist", h.UnblacklistUser)
}

func (h *Handler) GetRiskSummary(c echo.Context) error {
	userID := c.Param("userId")

	summary, err := h.svc.GetRiskSummary(c.Request().Context(), userID)
	if err != nil {
		return h.handleError(c, err)
	}

	return common.Success(c, summary)
}

func (h *Handler) GetPendingApprovals(c echo.Context) error {
	approvals, err := h.svc.GetPendingApprovals(c.Request().Context())
	if err != nil {
		return h.handleError(c, err)
	}
	return common.Success(c, approvals)
}

func (h *Handler) ResolveApproval(c echo.Context) error {
	id := c.Param("id")
	adminID := common.GetUserID(c)

	var req ResolveApprovalRequest
	if err := c.Bind(&req); err != nil {
		return common.Error(c, http.StatusBadRequest, common.ErrCodeBadRequest, "invalid request body")
	}

	if err := h.svc.ResolveApproval(c.Request().Context(), id, adminID, req.Status, req.Notes); err != nil {
		return common.Error(c, http.StatusBadRequest, common.ErrCodeBadRequest, err.Error())
	}

	return common.Success(c, map[string]string{"message": "approval resolved"})
}

func (h *Handler) BlacklistUser(c echo.Context) error {
	targetID := c.Param("id")
	adminID := common.GetUserID(c)

	var req BlacklistRequest
	if err := c.Bind(&req); err != nil {
		return common.Error(c, http.StatusBadRequest, common.ErrCodeBadRequest, "invalid request body")
	}

	if err := h.svc.BlacklistUser(c.Request().Context(), targetID, adminID, req.Reason); err != nil {
		return h.handleError(c, err)
	}

	return common.Success(c, map[string]string{"message": "user blacklisted"})
}

func (h *Handler) UnblacklistUser(c echo.Context) error {
	targetID := c.Param("id")
	adminID := common.GetUserID(c)

	if err := h.svc.UnblacklistUser(c.Request().Context(), targetID, adminID); err != nil {
		return h.handleError(c, err)
	}

	return common.Success(c, map[string]string{"message": "blacklist removed"})
}

func (h *Handler) handleError(c echo.Context, err error) error {
	var domainErr *common.DomainError
	if errors.As(err, &domainErr) {
		switch domainErr.Code {
		case common.ErrCodeNotFound:
			return common.Error(c, http.StatusNotFound, domainErr.Code, domainErr.Message)
		case common.ErrCodeForbidden:
			return common.Error(c, http.StatusForbidden, domainErr.Code, domainErr.Message)
		case common.ErrCodeBadRequest:
			return common.Error(c, http.StatusBadRequest, domainErr.Code, domainErr.Message)
		default:
			h.logger.Error("domain error", zap.String("code", domainErr.Code), zap.Error(err))
			return common.Error(c, http.StatusInternalServerError, common.ErrCodeInternal, "internal server error")
		}
	}
	h.logger.Error("unhandled error", zap.Error(err))
	return common.Error(c, http.StatusInternalServerError, common.ErrCodeInternal, "internal server error")
}
