package finance

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
	accountantAdmin := middleware.RequireRole(common.RoleAccountant, common.RoleAdministrator)
	courierRunner := middleware.RequireRole(common.RoleCourierRunner)

	g.GET("/wallets/:ownerId", h.GetWallet)
	g.GET("/wallets/:ownerId/transactions", h.GetTransactions)
	g.POST("/payments/record-tender", h.RecordTender, accountantAdmin)
	g.POST("/settlements/:orderId/release", h.ReleaseEscrow, accountantAdmin)
	g.POST("/refunds", h.ProcessRefund, accountantAdmin)
	g.POST("/withdrawals", h.RequestWithdrawal, courierRunner)
	g.POST("/withdrawals/:id/approve", h.ApproveWithdrawal, accountantAdmin)
	g.POST("/withdrawals/:id/reject", h.RejectWithdrawal, accountantAdmin)
	g.GET("/reconciliation", h.GetReconciliation, accountantAdmin)
	g.GET("/escrows/:ownerId", h.GetEscrows)
}

func (h *Handler) GetWallet(c echo.Context) error {
	ownerID := c.Param("ownerId")
	userID := common.GetUserID(c)
	walletType := WalletType(c.QueryParam("type"))
	if walletType == "" {
		walletType = WalletTypeCustomer
	}

	resp, err := h.svc.GetWallet(c.Request().Context(), ownerID, walletType, userID)
	if err != nil {
		return h.handleError(c, err)
	}
	return common.Success(c, resp)
}

func (h *Handler) GetTransactions(c echo.Context) error {
	ownerID := c.Param("ownerId")
	userID := common.GetUserID(c)
	page, _ := strconv.Atoi(c.QueryParam("page"))
	if page < 1 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(c.QueryParam("pageSize"))
	if pageSize < 1 {
		pageSize = 20
	}

	walletType := WalletType(c.QueryParam("type"))
	if walletType == "" {
		walletType = WalletTypeCustomer
	}
	w, err := h.svc.repo.GetWallet(c.Request().Context(), ownerID, walletType)
	if err != nil {
		return h.handleError(c, err)
	}

	resp, err := h.svc.GetTransactions(c.Request().Context(), w.ID, userID, page, pageSize)
	if err != nil {
		return h.handleError(c, err)
	}
	return common.Success(c, resp)
}

func (h *Handler) RecordTender(c echo.Context) error {
	var req RecordTenderRequest
	if err := c.Bind(&req); err != nil {
		return common.Error(c, http.StatusBadRequest, common.ErrCodeBadRequest, "invalid request body")
	}

	userID := common.GetUserID(c)
	if err := h.svc.RecordTender(c.Request().Context(), userID, req); err != nil {
		return h.handleError(c, err)
	}
	return common.Created(c, map[string]string{"message": "payment recorded"})
}

func (h *Handler) ReleaseEscrow(c echo.Context) error {
	orderID := c.Param("orderId")
	var req ReleaseEscrowRequest
	if err := c.Bind(&req); err != nil {
		return common.Error(c, http.StatusBadRequest, common.ErrCodeBadRequest, "invalid request body")
	}

	userID := common.GetUserID(c)
	if err := h.svc.ReleaseEscrow(c.Request().Context(), req.OrderType, orderID, req.Amount, userID); err != nil {
		return h.handleError(c, err)
	}
	return common.Success(c, map[string]string{"message": "escrow released"})
}

func (h *Handler) ProcessRefund(c echo.Context) error {
	var req RefundRequest
	if err := c.Bind(&req); err != nil {
		return common.Error(c, http.StatusBadRequest, common.ErrCodeBadRequest, "invalid request body")
	}

	userID := common.GetUserID(c)
	if err := h.svc.ProcessRefund(c.Request().Context(), userID, req); err != nil {
		return h.handleError(c, err)
	}
	return common.Created(c, map[string]string{"message": "refund processed"})
}

func (h *Handler) RequestWithdrawal(c echo.Context) error {
	var req WithdrawalCreateRequest
	if err := c.Bind(&req); err != nil {
		return common.Error(c, http.StatusBadRequest, common.ErrCodeBadRequest, "invalid request body")
	}

	userID := common.GetUserID(c)
	resp, err := h.svc.RequestWithdrawal(c.Request().Context(), userID, req)
	if err != nil {
		return h.handleError(c, err)
	}
	return common.Created(c, resp)
}

func (h *Handler) ApproveWithdrawal(c echo.Context) error {
	id := c.Param("id")
	userID := common.GetUserID(c)

	if err := h.svc.ApproveWithdrawal(c.Request().Context(), id, userID); err != nil {
		return h.handleError(c, err)
	}
	return common.Success(c, map[string]string{"message": "withdrawal approved"})
}

func (h *Handler) RejectWithdrawal(c echo.Context) error {
	id := c.Param("id")
	var req WithdrawalRejectRequest
	if err := c.Bind(&req); err != nil {
		return common.Error(c, http.StatusBadRequest, common.ErrCodeBadRequest, "invalid request body")
	}

	userID := common.GetUserID(c)
	if err := h.svc.RejectWithdrawal(c.Request().Context(), id, userID, req.Reason); err != nil {
		return h.handleError(c, err)
	}
	return common.Success(c, map[string]string{"message": "withdrawal rejected"})
}

func (h *Handler) GetReconciliation(c echo.Context) error {
	resp, err := h.svc.GetReconciliation(c.Request().Context())
	if err != nil {
		return h.handleError(c, err)
	}
	return common.Success(c, resp)
}

func (h *Handler) GetEscrows(c echo.Context) error {
	ownerID := c.Param("ownerId")
	userID := common.GetUserID(c)
	roles := common.GetRoles(c)

	if ownerID != userID && !common.IsAdminOrAccountant(roles) {
		return common.Error(c, http.StatusForbidden, common.ErrCodeForbidden, "access denied")
	}

	escrows, err := h.svc.GetActiveEscrows(c.Request().Context(), ownerID)
	if err != nil {
		return h.handleError(c, err)
	}
	return common.Success(c, escrows)
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
