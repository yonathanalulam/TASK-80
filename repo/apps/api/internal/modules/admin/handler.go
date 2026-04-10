package admin

import (
	"net/http"
	"strconv"

	"travel-platform/apps/api/internal/common"
	"travel-platform/apps/api/internal/middleware"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

type Handler struct {
	repo   *Repository
	logger *zap.Logger
}

func NewHandler(repo *Repository, logger *zap.Logger) *Handler {
	return &Handler{repo: repo, logger: logger}
}

func (h *Handler) RegisterRoutes(g *echo.Group) {
	admin := g.Group("", middleware.RequireRole("administrator"))
	admin.GET("/audit-logs", h.GetAuditLogs)
	admin.GET("/send-logs", h.GetSendLogs)
	admin.GET("/config", h.GetConfig)
}

func (h *Handler) GetAuditLogs(c echo.Context) error {
	page, _ := strconv.Atoi(c.QueryParam("page"))
	if page < 1 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(c.QueryParam("pageSize"))
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	filters := AuditLogFilters{
		ActorID:    c.QueryParam("actorId"),
		EntityType: c.QueryParam("entityType"),
		Action:     c.QueryParam("action"),
		Page:       page,
		PageSize:   pageSize,
	}

	logs, total, err := h.repo.GetAuditLogs(c.Request().Context(), filters)
	if err != nil {
		h.logger.Error("failed to get audit logs", zap.Error(err))
		return common.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to retrieve audit logs")
	}

	return common.Success(c, map[string]interface{}{
		"items":       logs,
		"total":       total,
		"page":        page,
		"page_size":   pageSize,
		"total_pages": (total + pageSize - 1) / pageSize,
	})
}

func (h *Handler) GetSendLogs(c echo.Context) error {
	page, _ := strconv.Atoi(c.QueryParam("page"))
	if page < 1 {
		page = 1
	}

	logs, total, err := h.repo.GetSendLogs(c.Request().Context(), page, 50)
	if err != nil {
		return common.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to retrieve send logs")
	}

	return common.Success(c, map[string]interface{}{
		"items": logs,
		"total": total,
		"page":  page,
	})
}

func (h *Handler) GetConfig(c echo.Context) error {
	config := map[string]interface{}{
		"dnd_default_start":       "21:00",
		"dnd_default_end":         "08:00",
		"courier_daily_cap":       2500.00,
		"refund_minimum_unit":     1.00,
		"download_token_ttl_min":  5,
		"max_cancellations_24h":   8,
		"max_rfqs_10min":          20,
		"coupon_max_threshold":    1,
		"coupon_max_percentage":   1,
		"new_user_gift_exclusive": true,
	}
	return common.Success(c, config)
}
