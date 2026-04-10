package auth

import (
	"net/http"

	"travel-platform/apps/api/internal/common"

	"github.com/labstack/echo/v4"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *Handler) RegisterRoutes(g *echo.Group) {
	g.POST("/login", h.Login)
	g.POST("/logout", h.Logout)
	g.GET("/me", h.Me)
}

func (h *Handler) Login(c echo.Context) error {
	var req LoginRequest
	if err := c.Bind(&req); err != nil {
		return common.Error(c, http.StatusBadRequest, "BAD_REQUEST", "invalid request body")
	}

	if req.Email == "" || req.Password == "" {
		return common.ValidationError(c, map[string]string{
			"email":    "email is required",
			"password": "password is required",
		})
	}

	token, err := h.service.Login(c.Request().Context(), req.Email, req.Password)
	if err != nil {
		return common.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", "invalid email or password")
	}

	return common.Success(c, map[string]string{
		"token": token,
	})
}

func (h *Handler) Logout(c echo.Context) error {
	return common.Success(c, map[string]string{
		"message": "logged out successfully",
	})
}

func (h *Handler) Me(c echo.Context) error {
	userID := common.GetUserID(c)
	if userID == "" {
		return common.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", "not authenticated")
	}

	user, err := h.service.GetUserByID(c.Request().Context(), userID)
	if err != nil {
		return common.Error(c, http.StatusNotFound, "NOT_FOUND", "user not found")
	}

	return common.Success(c, user)
}
