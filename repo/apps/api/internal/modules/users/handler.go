package users

import (
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

func (h *Handler) RegisterRoutes(g *echo.Group) {
	g.GET("/:id", h.GetUser)
	g.PATCH("/:id/profile", h.UpdateProfile)
	g.PATCH("/:id/preferences", h.UpdatePreferences)
}

func (h *Handler) RegisterAdminRoutes(g *echo.Group) {
	admin := g.Group("", middleware.RequireRole("administrator"))
	admin.GET("/users", h.ListUsers)
}

func (h *Handler) GetUser(c echo.Context) error {
	id := c.Param("id")
	actorID := common.GetUserID(c)
	roles := common.GetRoles(c)

	if id != actorID && !common.HasRole(roles, common.RoleAdministrator) {
		return common.Error(c, http.StatusForbidden, common.ErrCodeForbidden, "access denied")
	}

	user, err := h.svc.GetUser(c.Request().Context(), id)
	if err != nil {
		return common.Error(c, http.StatusNotFound, "NOT_FOUND", "user not found")
	}
	return common.Success(c, user)
}

func (h *Handler) UpdateProfile(c echo.Context) error {
	id := c.Param("id")
	actorID := common.GetUserID(c)
	roles := common.GetRoles(c)

	var req UpdateProfileRequest
	if err := c.Bind(&req); err != nil {
		return common.Error(c, http.StatusBadRequest, "BAD_REQUEST", "invalid request body")
	}

	if err := h.svc.UpdateProfile(c.Request().Context(), id, actorID, roles, req); err != nil {
		if err.Error() == "forbidden" {
			return common.Error(c, http.StatusForbidden, "FORBIDDEN", "not authorized")
		}
		return common.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to update profile")
	}

	return common.Success(c, map[string]string{"message": "profile updated"})
}

func (h *Handler) UpdatePreferences(c echo.Context) error {
	id := c.Param("id")
	actorID := common.GetUserID(c)
	roles := common.GetRoles(c)

	var req UpdatePreferencesRequest
	if err := c.Bind(&req); err != nil {
		return common.Error(c, http.StatusBadRequest, "BAD_REQUEST", "invalid request body")
	}

	if err := h.svc.UpdatePreferences(c.Request().Context(), id, actorID, roles, req); err != nil {
		if err.Error() == "forbidden" {
			return common.Error(c, http.StatusForbidden, "FORBIDDEN", "not authorized")
		}
		return common.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to update preferences")
	}

	return common.Success(c, map[string]string{"message": "preferences updated"})
}

func (h *Handler) ListUsers(c echo.Context) error {
	page := 1
	pageSize := 20
	status := c.QueryParam("status")

	users, total, err := h.svc.ListUsers(c.Request().Context(), page, pageSize, status)
	if err != nil {
		return common.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to list users")
	}

	return common.Success(c, map[string]interface{}{
		"items":      users,
		"total":      total,
		"page":       page,
		"page_size":  pageSize,
		"total_pages": (total + pageSize - 1) / pageSize,
	})
}
