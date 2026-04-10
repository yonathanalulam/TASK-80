package files

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"

	"travel-platform/apps/api/internal/common"
)

type Handler struct {
	service *FileVaultService
	logger  *zap.Logger
}

func NewHandler(service *FileVaultService, logger *zap.Logger) *Handler {
	return &Handler{service: service, logger: logger}
}

func RegisterRoutes(g *echo.Group, h *Handler) {
	g.POST("/upload", h.Upload)
	g.POST("/:id/download-token", h.CreateDownloadToken)
	g.GET("/download/:token", h.Download)
	g.GET("/record/:recordType/:recordId", h.GetFilesForRecord)
}

func (h *Handler) Upload(c echo.Context) error {
	userID := common.GetUserID(c)
	if userID == "" {
		return common.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required")
	}

	file, header, err := c.Request().FormFile("file")
	if err != nil {
		return common.Error(c, http.StatusBadRequest, "BAD_REQUEST", "file is required")
	}
	defer file.Close()

	recordType := c.FormValue("recordType")
	recordID := c.FormValue("recordId")
	encrypt := c.FormValue("encrypt") == "true"

	meta, err := h.service.Upload(c.Request().Context(), userID, file, header, recordType, recordID, encrypt)
	if err != nil {
		return handleServiceError(c, err)
	}

	return common.Created(c, meta)
}

type createDownloadTokenRequest struct {
	TTLSeconds int  `json:"ttlSeconds"`
	SingleUse  bool `json:"singleUse"`
}

func (h *Handler) CreateDownloadToken(c echo.Context) error {
	userID := common.GetUserID(c)
	if userID == "" {
		return common.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required")
	}

	fileID := c.Param("id")
	if fileID == "" {
		return common.Error(c, http.StatusBadRequest, "BAD_REQUEST", "file ID is required")
	}

	var req createDownloadTokenRequest
	if err := c.Bind(&req); err != nil {
		return common.Error(c, http.StatusBadRequest, "BAD_REQUEST", "invalid request body")
	}

	roles := common.GetRoles(c)
	ttl := time.Duration(req.TTLSeconds) * time.Second

	dt, err := h.service.CreateDownloadToken(c.Request().Context(), fileID, userID, roles, ttl, req.SingleUse)
	if err != nil {
		return handleServiceError(c, err)
	}

	return common.Created(c, dt)
}

func (h *Handler) Download(c echo.Context) error {
	token := c.Param("token")
	if token == "" {
		return common.Error(c, http.StatusBadRequest, "BAD_REQUEST", "download token is required")
	}

	reader, meta, err := h.service.Download(c.Request().Context(), token)
	if err != nil {
		return handleServiceError(c, err)
	}
	defer reader.Close()

	c.Response().Header().Set("Content-Disposition", "attachment; filename=\""+meta.OriginalFilename+"\"")
	c.Response().Header().Set("Content-Type", meta.MimeType)

	return c.Stream(http.StatusOK, meta.MimeType, reader)
}

func (h *Handler) GetFilesForRecord(c echo.Context) error {
	userID := common.GetUserID(c)
	if userID == "" {
		return common.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required")
	}

	recordType := c.Param("recordType")
	recordID := c.Param("recordId")

	if recordType == "" || recordID == "" {
		return common.Error(c, http.StatusBadRequest, "BAD_REQUEST", "recordType and recordId are required")
	}

	roles := common.GetRoles(c)
	files, err := h.service.GetFilesForRecord(c.Request().Context(), recordType, recordID, userID, roles)
	if err != nil {
		return handleServiceError(c, err)
	}

	if files == nil {
		files = []FileMetadata{}
	}

	return common.Success(c, files)
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
