package notifications

import (
	"encoding/json"
	"net/http"
	"strconv"

	"travel-platform/apps/api/internal/common"
	"travel-platform/apps/api/internal/middleware"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

type Handler struct {
	service *Service
	logger  *zap.Logger
}

func NewHandler(service *Service, logger *zap.Logger) *Handler {
	return &Handler{service: service, logger: logger}
}

func (h *Handler) RegisterRoutes(g *echo.Group, userGroup *echo.Group) {
	g.GET("/notifications", h.ListNotifications)
	g.POST("/notifications/:id/read", h.MarkNotificationRead)

	g.GET("/messages", h.ListMessages)
	g.POST("/messages/callback-queue/export", h.ExportCallbackQueue, middleware.RequireRole(common.RoleAdministrator, common.RoleAccountant))

	g.GET("/send-logs", h.ListSendLogs)

	userGroup.GET("/:id/dnd", h.GetDNDSettings)
	userGroup.PATCH("/:id/dnd", h.UpdateDNDSettings)
	userGroup.GET("/:id/subscriptions", h.GetSubscriptionPreferences)
	userGroup.PATCH("/:id/subscriptions", h.UpdateSubscriptionPreferences)
}
func (h *Handler) ListNotifications(c echo.Context) error {
	userID := common.GetUserID(c)
	if userID == "" {
		return common.Error(c, http.StatusUnauthorized, common.ErrCodeUnauthorized, "authentication required")
	}

	page, _ := strconv.Atoi(c.QueryParam("page"))
	pageSize, _ := strconv.Atoi(c.QueryParam("pageSize"))
	unreadOnly := c.QueryParam("unreadOnly") == "true"

	result, err := h.service.GetNotifications(c.Request().Context(), userID, page, pageSize, unreadOnly)
	if err != nil {
		h.logger.Error("list notifications failed", zap.Error(err))
		return common.Error(c, http.StatusInternalServerError, common.ErrCodeInternal, "failed to list notifications")
	}

	return common.Success(c, result)
}

func (h *Handler) MarkNotificationRead(c echo.Context) error {
	userID := common.GetUserID(c)
	if userID == "" {
		return common.Error(c, http.StatusUnauthorized, common.ErrCodeUnauthorized, "authentication required")
	}

	notificationID := c.Param("id")
	if notificationID == "" {
		return common.Error(c, http.StatusBadRequest, common.ErrCodeBadRequest, "notification id is required")
	}

	if err := h.service.MarkRead(c.Request().Context(), notificationID, userID); err != nil {
		h.logger.Error("mark read failed", zap.Error(err))
		return common.Error(c, http.StatusNotFound, common.ErrCodeNotFound, "notification not found or already read")
	}

	return common.Success(c, map[string]string{"status": "read"})
}
func (h *Handler) ListMessages(c echo.Context) error {
	userID := common.GetUserID(c)
	if userID == "" {
		return common.Error(c, http.StatusUnauthorized, common.ErrCodeUnauthorized, "authentication required")
	}

	page, _ := strconv.Atoi(c.QueryParam("page"))
	pageSize, _ := strconv.Atoi(c.QueryParam("pageSize"))

	result, err := h.service.GetMessages(c.Request().Context(), userID, page, pageSize)
	if err != nil {
		h.logger.Error("list messages failed", zap.Error(err))
		return common.Error(c, http.StatusInternalServerError, common.ErrCodeInternal, "failed to list messages")
	}

	return common.Success(c, result)
}
func (h *Handler) ListSendLogs(c echo.Context) error {
	userID := common.GetUserID(c)
	if userID == "" {
		return common.Error(c, http.StatusUnauthorized, common.ErrCodeUnauthorized, "authentication required")
	}

	roles := common.GetRoles(c)
	isAdmin := false
	for _, r := range roles {
		if r == common.RoleAdministrator {
			isAdmin = true
			break
		}
	}

	logs, err := h.service.GetSendLogs(c.Request().Context(), userID, isAdmin)
	if err != nil {
		h.logger.Error("list send logs failed", zap.Error(err))
		return common.Error(c, http.StatusInternalServerError, common.ErrCodeInternal, "failed to list send logs")
	}

	return common.Success(c, logs)
}
func (h *Handler) ExportCallbackQueue(c echo.Context) error {
	data, err := h.service.ExportCallbackQueue(c.Request().Context())
	if err != nil {
		h.logger.Error("export callback queue failed", zap.Error(err))
		return common.Error(c, http.StatusInternalServerError, common.ErrCodeInternal, "failed to export callback queue")
	}

	return c.JSONBlob(http.StatusOK, data)
}
func (h *Handler) GetDNDSettings(c echo.Context) error {
	requestingUserID := common.GetUserID(c)
	targetUserID := c.Param("id")

	if !h.canAccessUserResource(requestingUserID, targetUserID, common.GetRoles(c)) {
		return common.Error(c, http.StatusForbidden, common.ErrCodeForbidden, "access denied")
	}

	dnd, err := h.service.repo.GetDNDSettings(c.Request().Context(), targetUserID)
	if err != nil {
		h.logger.Error("get dnd settings failed", zap.Error(err))
		return common.Error(c, http.StatusInternalServerError, common.ErrCodeInternal, "failed to get DND settings")
	}

	return common.Success(c, dnd)
}

func (h *Handler) UpdateDNDSettings(c echo.Context) error {
	requestingUserID := common.GetUserID(c)
	targetUserID := c.Param("id")

	if !h.canAccessUserResource(requestingUserID, targetUserID, common.GetRoles(c)) {
		return common.Error(c, http.StatusForbidden, common.ErrCodeForbidden, "access denied")
	}

	var req UpdateDNDRequest
	if err := c.Bind(&req); err != nil {
		return common.Error(c, http.StatusBadRequest, common.ErrCodeBadRequest, "invalid request body")
	}

	if err := h.service.UpdateDND(c.Request().Context(), targetUserID, req); err != nil {
		h.logger.Error("update dnd settings failed", zap.Error(err))
		return common.Error(c, http.StatusInternalServerError, common.ErrCodeInternal, "failed to update DND settings")
	}

	return common.Success(c, map[string]string{"status": "updated"})
}
func (h *Handler) GetSubscriptionPreferences(c echo.Context) error {
	requestingUserID := common.GetUserID(c)
	targetUserID := c.Param("id")

	if !h.canAccessUserResource(requestingUserID, targetUserID, common.GetRoles(c)) {
		return common.Error(c, http.StatusForbidden, common.ErrCodeForbidden, "access denied")
	}

	prefs, err := h.service.repo.GetSubscriptionPreferences(c.Request().Context(), targetUserID)
	if err != nil {
		h.logger.Error("get subscription prefs failed", zap.Error(err))
		return common.Error(c, http.StatusInternalServerError, common.ErrCodeInternal, "failed to get subscription preferences")
	}

	return common.Success(c, prefs)
}

func (h *Handler) UpdateSubscriptionPreferences(c echo.Context) error {
	requestingUserID := common.GetUserID(c)
	targetUserID := c.Param("id")

	if !h.canAccessUserResource(requestingUserID, targetUserID, common.GetRoles(c)) {
		return common.Error(c, http.StatusForbidden, common.ErrCodeForbidden, "access denied")
	}

	var prefs []UpdateSubscriptionRequest
	if err := json.NewDecoder(c.Request().Body).Decode(&prefs); err != nil {
		return common.Error(c, http.StatusBadRequest, common.ErrCodeBadRequest, "invalid request body")
	}

	if err := h.service.UpdateSubscriptions(c.Request().Context(), targetUserID, prefs); err != nil {
		h.logger.Error("update subscriptions failed", zap.Error(err))
		return common.Error(c, http.StatusInternalServerError, common.ErrCodeInternal, "failed to update subscriptions")
	}

	return common.Success(c, map[string]string{"status": "updated"})
}
func (h *Handler) canAccessUserResource(requestingUserID, targetUserID string, roles []string) bool {
	if requestingUserID == targetUserID {
		return true
	}
	for _, r := range roles {
		if r == common.RoleAdministrator {
			return true
		}
	}
	return false
}
