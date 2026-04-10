package reviews

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"

	"travel-platform/apps/api/internal/common"
	"travel-platform/apps/api/internal/middleware"
)

type Handler struct {
	service *ReviewService
	logger  *zap.Logger
}

func NewHandler(service *ReviewService, logger *zap.Logger) *Handler {
	return &Handler{service: service, logger: logger}
}

func RegisterRoutes(g *echo.Group, h *Handler) {
	g.POST("/reviews", h.SubmitReview)
	g.GET("/reviews/subject/:userId", h.GetReviewsForSubject)
	g.GET("/credit-tiers/:userId", h.GetCreditSnapshot)
	g.POST("/violations", h.RecordViolation, middleware.RequireRole(common.RoleAdministrator, common.RoleAccountant))
	g.POST("/no-shows", h.RecordNoShow, middleware.RequireRole(common.RoleAdministrator, common.RoleAccountant))
	g.POST("/harassment-flags", h.FlagHarassment)
	g.GET("/risk/:userId", h.GetRiskSummary, middleware.RequireRole(common.RoleAdministrator))
}

func (h *Handler) SubmitReview(c echo.Context) error {
	userID := common.GetUserID(c)
	if userID == "" {
		return common.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required")
	}

	var req CreateReviewRequest
	if err := c.Bind(&req); err != nil {
		return common.Error(c, http.StatusBadRequest, "BAD_REQUEST", "invalid request body")
	}

	if err := h.service.SubmitReview(c.Request().Context(), userID, req); err != nil {
		return handleServiceError(c, err)
	}

	return common.Created(c, map[string]string{"message": "review submitted"})
}

func (h *Handler) GetReviewsForSubject(c echo.Context) error {
	subjectID := c.Param("userId")
	if subjectID == "" {
		return common.Error(c, http.StatusBadRequest, "BAD_REQUEST", "userId is required")
	}

	page, _ := strconv.Atoi(c.QueryParam("page"))
	pageSize, _ := strconv.Atoi(c.QueryParam("pageSize"))

	reviews, total, err := h.service.GetReviewsForSubject(c.Request().Context(), subjectID, page, pageSize)
	if err != nil {
		return handleServiceError(c, err)
	}

	if reviews == nil {
		reviews = []ReviewResponse{}
	}

	return common.Success(c, map[string]interface{}{
		"items": reviews,
		"total": total,
		"page":  page,
	})
}

func (h *Handler) GetCreditSnapshot(c echo.Context) error {
	userID := c.Param("userId")
	if userID == "" {
		return common.Error(c, http.StatusBadRequest, "BAD_REQUEST", "userId is required")
	}

	snap, err := h.service.GetCreditSnapshot(c.Request().Context(), userID)
	if err != nil {
		return handleServiceError(c, err)
	}

	return common.Success(c, snap)
}

func (h *Handler) RecordViolation(c echo.Context) error {
	userID := common.GetUserID(c)
	if userID == "" {
		return common.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required")
	}

	var req CreateViolationRequest
	if err := c.Bind(&req); err != nil {
		return common.Error(c, http.StatusBadRequest, "BAD_REQUEST", "invalid request body")
	}

	if err := h.service.RecordViolation(c.Request().Context(), userID, req); err != nil {
		return handleServiceError(c, err)
	}

	return common.Created(c, map[string]string{"message": "violation recorded"})
}

func (h *Handler) RecordNoShow(c echo.Context) error {
	userID := common.GetUserID(c)
	if userID == "" {
		return common.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required")
	}

	var req CreateNoShowRequest
	if err := c.Bind(&req); err != nil {
		return common.Error(c, http.StatusBadRequest, "BAD_REQUEST", "invalid request body")
	}

	if err := h.service.RecordNoShow(c.Request().Context(), userID, req); err != nil {
		return handleServiceError(c, err)
	}

	return common.Created(c, map[string]string{"message": "no-show recorded"})
}

func (h *Handler) FlagHarassment(c echo.Context) error {
	userID := common.GetUserID(c)
	if userID == "" {
		return common.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required")
	}

	var req CreateHarassmentFlagRequest
	if err := c.Bind(&req); err != nil {
		return common.Error(c, http.StatusBadRequest, "BAD_REQUEST", "invalid request body")
	}

	if err := h.service.FlagHarassment(c.Request().Context(), userID, req); err != nil {
		return handleServiceError(c, err)
	}

	return common.Created(c, map[string]string{"message": "harassment flag submitted"})
}

func (h *Handler) GetRiskSummary(c echo.Context) error {
	userID := c.Param("userId")
	if userID == "" {
		return common.Error(c, http.StatusBadRequest, "BAD_REQUEST", "userId is required")
	}

	summary, err := h.service.GetRiskSummary(c.Request().Context(), userID)
	if err != nil {
		return handleServiceError(c, err)
	}

	return common.Success(c, summary)
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
