package itineraries

import (
	"errors"
	"net/http"
	"strconv"

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
	organizer := middleware.RequireRole("group_organizer", "administrator")
	traveler := middleware.RequireRole("traveler", "group_organizer", "administrator")

	g.POST("", h.Create, organizer)
	g.GET("", h.List)
	g.GET("/:id", h.GetByID)
	g.PATCH("/:id", h.Update, organizer)
	g.POST("/:id/publish", h.Publish, organizer)

	g.POST("/:id/checkpoints", h.AddCheckpoint, organizer)
	g.PATCH("/:id/checkpoints/:checkpointId", h.UpdateCheckpoint, organizer)
	g.DELETE("/:id/checkpoints/:checkpointId", h.DeleteCheckpoint, organizer)

	g.POST("/:id/members", h.AddMember, organizer)
	g.DELETE("/:id/members/:userId", h.RemoveMember, organizer)

	g.POST("/:id/form-definitions", h.CreateFormDefinition, organizer)
	g.PATCH("/:id/form-definitions/:defId", h.UpdateFormDefinition, organizer)
	g.GET("/:id/form-definitions", h.GetFormDefinitions)

	g.POST("/:id/form-submissions", h.SubmitForm, traveler)
	g.GET("/:id/form-submissions", h.GetFormSubmissions, organizer)

	g.GET("/:id/change-events", h.GetChangeEvents)
}
func (h *Handler) Create(c echo.Context) error {
	var req CreateItineraryRequest
	if err := c.Bind(&req); err != nil {
		return common.Error(c, http.StatusBadRequest, common.ErrCodeBadRequest, "invalid request body")
	}

	userID := common.GetUserID(c)
	resp, err := h.svc.CreateItinerary(c.Request().Context(), userID, req)
	if err != nil {
		return h.handleError(c, err)
	}
	return common.Created(c, resp)
}

func (h *Handler) List(c echo.Context) error {
	userID := common.GetUserID(c)
	roles := common.GetRoles(c)
	page, _ := strconv.Atoi(c.QueryParam("page"))
	if page < 1 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(c.QueryParam("pageSize"))
	if pageSize < 1 {
		pageSize = 20
	}
	status := c.QueryParam("status")

	resp, err := h.svc.ListItineraries(c.Request().Context(), userID, roles, page, pageSize, status)
	if err != nil {
		return h.handleError(c, err)
	}
	return common.Success(c, resp)
}

func (h *Handler) GetByID(c echo.Context) error {
	id := c.Param("id")
	userID := common.GetUserID(c)
	roles := common.GetRoles(c)

	resp, err := h.svc.GetItinerary(c.Request().Context(), id, userID, roles)
	if err != nil {
		return h.handleError(c, err)
	}
	return common.Success(c, resp)
}

func (h *Handler) Update(c echo.Context) error {
	id := c.Param("id")
	var req UpdateItineraryRequest
	if err := c.Bind(&req); err != nil {
		return common.Error(c, http.StatusBadRequest, common.ErrCodeBadRequest, "invalid request body")
	}

	userID := common.GetUserID(c)
	roles := common.GetRoles(c)
	if err := h.svc.UpdateItinerary(c.Request().Context(), id, userID, roles, req); err != nil {
		return h.handleError(c, err)
	}
	return common.Success(c, map[string]string{"message": "itinerary updated"})
}

func (h *Handler) Publish(c echo.Context) error {
	id := c.Param("id")
	userID := common.GetUserID(c)
	roles := common.GetRoles(c)

	if err := h.svc.PublishItinerary(c.Request().Context(), id, userID, roles); err != nil {
		return h.handleError(c, err)
	}
	return common.Success(c, map[string]string{"message": "itinerary published"})
}
func (h *Handler) AddCheckpoint(c echo.Context) error {
	itineraryID := c.Param("id")
	var req CreateCheckpointRequest
	if err := c.Bind(&req); err != nil {
		return common.Error(c, http.StatusBadRequest, common.ErrCodeBadRequest, "invalid request body")
	}

	userID := common.GetUserID(c)
	roles := common.GetRoles(c)
	if err := h.svc.AddCheckpoint(c.Request().Context(), itineraryID, userID, roles, req); err != nil {
		return h.handleError(c, err)
	}
	return common.Created(c, map[string]string{"message": "checkpoint added"})
}

func (h *Handler) UpdateCheckpoint(c echo.Context) error {
	itineraryID := c.Param("id")
	checkpointID := c.Param("checkpointId")
	var req UpdateCheckpointRequest
	if err := c.Bind(&req); err != nil {
		return common.Error(c, http.StatusBadRequest, common.ErrCodeBadRequest, "invalid request body")
	}

	userID := common.GetUserID(c)
	roles := common.GetRoles(c)
	if err := h.svc.UpdateCheckpoint(c.Request().Context(), itineraryID, checkpointID, userID, roles, req); err != nil {
		return h.handleError(c, err)
	}
	return common.Success(c, map[string]string{"message": "checkpoint updated"})
}

func (h *Handler) DeleteCheckpoint(c echo.Context) error {
	itineraryID := c.Param("id")
	checkpointID := c.Param("checkpointId")
	userID := common.GetUserID(c)
	roles := common.GetRoles(c)

	if err := h.svc.DeleteCheckpoint(c.Request().Context(), itineraryID, checkpointID, userID, roles); err != nil {
		return h.handleError(c, err)
	}
	return common.Success(c, map[string]string{"message": "checkpoint deleted"})
}
func (h *Handler) AddMember(c echo.Context) error {
	itineraryID := c.Param("id")
	var req AddMemberRequest
	if err := c.Bind(&req); err != nil {
		return common.Error(c, http.StatusBadRequest, common.ErrCodeBadRequest, "invalid request body")
	}

	userID := common.GetUserID(c)
	roles := common.GetRoles(c)
	if err := h.svc.AddMember(c.Request().Context(), itineraryID, userID, roles, req); err != nil {
		return h.handleError(c, err)
	}
	return common.Created(c, map[string]string{"message": "member added"})
}

func (h *Handler) RemoveMember(c echo.Context) error {
	itineraryID := c.Param("id")
	targetUserID := c.Param("userId")
	userID := common.GetUserID(c)
	roles := common.GetRoles(c)

	if err := h.svc.RemoveMember(c.Request().Context(), itineraryID, userID, targetUserID, roles); err != nil {
		return h.handleError(c, err)
	}
	return common.Success(c, map[string]string{"message": "member removed"})
}
func (h *Handler) CreateFormDefinition(c echo.Context) error {
	itineraryID := c.Param("id")
	var req CreateFormDefinitionRequest
	if err := c.Bind(&req); err != nil {
		return common.Error(c, http.StatusBadRequest, common.ErrCodeBadRequest, "invalid request body")
	}

	userID := common.GetUserID(c)
	roles := common.GetRoles(c)
	if err := h.svc.CreateFormDefinition(c.Request().Context(), itineraryID, userID, roles, req); err != nil {
		return h.handleError(c, err)
	}
	return common.Created(c, map[string]string{"message": "form definition created"})
}

func (h *Handler) UpdateFormDefinition(c echo.Context) error {
	itineraryID := c.Param("id")
	defID := c.Param("defId")
	var req UpdateFormDefinitionRequest
	if err := c.Bind(&req); err != nil {
		return common.Error(c, http.StatusBadRequest, common.ErrCodeBadRequest, "invalid request body")
	}

	userID := common.GetUserID(c)
	roles := common.GetRoles(c)
	if err := h.svc.UpdateFormDefinition(c.Request().Context(), itineraryID, defID, userID, roles, req); err != nil {
		return h.handleError(c, err)
	}
	return common.Success(c, map[string]string{"message": "form definition updated"})
}

func (h *Handler) GetFormDefinitions(c echo.Context) error {
	itineraryID := c.Param("id")
	userID := common.GetUserID(c)
	roles := common.GetRoles(c)

	defs, err := h.svc.GetFormDefinitions(c.Request().Context(), itineraryID, userID, roles)
	if err != nil {
		return h.handleError(c, err)
	}
	return common.Success(c, defs)
}
func (h *Handler) SubmitForm(c echo.Context) error {
	itineraryID := c.Param("id")
	var req SubmitFormRequest
	if err := c.Bind(&req); err != nil {
		return common.Error(c, http.StatusBadRequest, common.ErrCodeBadRequest, "invalid request body")
	}

	userID := common.GetUserID(c)
	if err := h.svc.SubmitForm(c.Request().Context(), itineraryID, userID, req); err != nil {
		return h.handleError(c, err)
	}
	return common.Created(c, map[string]string{"message": "form submitted"})
}

func (h *Handler) GetFormSubmissions(c echo.Context) error {
	itineraryID := c.Param("id")
	userID := common.GetUserID(c)
	roles := common.GetRoles(c)

	subs, err := h.svc.GetFormSubmissions(c.Request().Context(), itineraryID, userID, roles)
	if err != nil {
		return h.handleError(c, err)
	}
	return common.Success(c, subs)
}
func (h *Handler) GetChangeEvents(c echo.Context) error {
	itineraryID := c.Param("id")
	userID := common.GetUserID(c)
	roles := common.GetRoles(c)

	events, err := h.svc.GetChangeEvents(c.Request().Context(), itineraryID, userID, roles)
	if err != nil {
		return h.handleError(c, err)
	}
	return common.Success(c, events)
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
