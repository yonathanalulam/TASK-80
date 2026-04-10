package finance

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"travel-platform/apps/api/internal/common"
	"travel-platform/apps/api/internal/modules/risk"
)

const MaxDailyWithdrawal = 2500.00

const MinRefundAmount = 1.00

type Service struct {
	repo    *Repository
	riskSvc *risk.Service
	logger  *zap.Logger
}

func NewService(repo *Repository, riskSvc *risk.Service, logger *zap.Logger) *Service {
	return &Service{repo: repo, riskSvc: riskSvc, logger: logger}
}
func (s *Service) GetWallet(ctx context.Context, ownerID string, walletType WalletType, requestingUserID string) (*WalletResponse, error) {
	if ownerID != requestingUserID {
		return nil, common.NewForbiddenError("you can only view your own wallet")
	}

	w, err := s.repo.GetWallet(ctx, ownerID, walletType)
	if err != nil {
		return nil, err
	}

	return &WalletResponse{
		ID:         w.ID,
		OwnerID:    w.OwnerID,
		WalletType: w.WalletType,
		Balance:    w.Balance,
		Currency:   w.Currency,
		CreatedAt:  w.CreatedAt,
		UpdatedAt:  w.UpdatedAt,
	}, nil
}

func (s *Service) GetTransactions(ctx context.Context, walletID, userID string, page, pageSize int) (*PaginatedTransactions, error) {
	w, err := s.repo.GetWalletByID(ctx, walletID)
	if err != nil {
		return nil, err
	}
	if w.OwnerID != nil && *w.OwnerID != userID {
		return nil, common.NewForbiddenError("you can only view your own transactions")
	}

	txns, total, err := s.repo.GetWalletTransactions(ctx, walletID, page, pageSize)
	if err != nil {
		return nil, common.NewInternalError("failed to fetch transactions", err)
	}

	items := make([]TransactionResponse, 0, len(txns))
	for _, t := range txns {
		items = append(items, TransactionResponse{
			ID:            t.ID,
			WalletID:      t.WalletID,
			Amount:        t.Amount,
			Direction:     t.Direction,
			ReferenceType: t.ReferenceType,
			ReferenceID:   t.ReferenceID,
			Description:   t.Description,
			CreatedAt:     t.CreatedAt,
		})
	}

	totalPages := 0
	if pageSize > 0 {
		totalPages = (total + pageSize - 1) / pageSize
	}

	return &PaginatedTransactions{
		Items:      items,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}
func (s *Service) RecordTender(ctx context.Context, userID string, req RecordTenderRequest) error {
	if req.Amount <= 0 {
		return common.NewBadRequestError("amount must be positive")
	}
	if req.OrderID == "" || req.OrderType == "" {
		return common.NewBadRequestError("orderType and orderId are required")
	}

	p := &PaymentRecord{
		OrderType:     req.OrderType,
		OrderID:       req.OrderID,
		TenderType:    req.TenderType,
		Amount:        req.Amount,
		Currency:      req.Currency,
		ReferenceText: req.ReferenceText,
		RecordedBy:    userID,
	}

	paymentID, err := s.repo.CreatePaymentRecord(ctx, p)
	if err != nil {
		return common.NewInternalError("failed to record payment", err)
	}

	tx, err := s.repo.BeginTx(ctx)
	if err != nil {
		return common.NewInternalError("failed to begin transaction", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	refID, _ := uuid.Parse(paymentID)
	err = PostJournalEntry(ctx, tx, "tender_recording", "payment_record", refID, fmt.Sprintf("Tender recorded for %s %s", req.OrderType, req.OrderID), userID, []JournalLine{
		{AccountCode: ManualTenderClearing, Direction: Debit, Amount: req.Amount},
		{AccountCode: CashOnHand, Direction: Credit, Amount: req.Amount},
	})
	if err != nil {
		return common.NewInternalError("failed to post journal entry", err)
	}

	return tx.Commit(ctx)
}
func (s *Service) ProcessRefund(ctx context.Context, userID string, req RefundRequest) error {
	if s.riskSvc != nil {
		decision, err := s.riskSvc.EvaluateAction(ctx, userID, common.RiskActionProcessRefund)
		if err == nil && !decision.Allowed {
			return common.NewForbiddenError("action blocked by risk engine: " + decision.Reason)
		}
	}

	if req.Amount < MinRefundAmount {
		return common.NewBadRequestError(fmt.Sprintf("refund amount must be at least $%.2f", MinRefundAmount))
	}

	if math.Mod(req.Amount*100, 100) != 0 {
		return common.NewBadRequestError("refund amount must be in whole dollars (multiples of 1.00)")
	}

	if req.OrderID == "" || req.OrderType == "" {
		return common.NewBadRequestError("orderType and orderId are required")
	}

	refund := &Refund{
		OrderType:    req.OrderType,
		OrderID:      req.OrderID,
		RefundAmount: req.Amount,
		RefundReason: req.Reason,
		CreatedBy:    userID,
		Status:       RefundStatusApproved,
	}

	refundID, err := s.repo.CreateRefund(ctx, refund)
	if err != nil {
		return common.NewInternalError("failed to create refund", err)
	}

	for _, item := range req.Items {
		ri := &RefundItem{
			RefundID: refundID,
			ItemID:   item.ItemID,
			ItemType: item.ItemType,
			Amount:   item.Amount,
		}
		if err := s.repo.CreateRefundItem(ctx, ri); err != nil {
			s.logger.Error("failed to create refund item", zap.Error(err))
		}
	}

	tx, err := s.repo.BeginTx(ctx)
	if err != nil {
		return common.NewInternalError("failed to begin transaction", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	refUUID, _ := uuid.Parse(refundID)
	err = PostJournalEntry(ctx, tx, "refund", "refund", refUUID, fmt.Sprintf("Refund for %s %s: %s", req.OrderType, req.OrderID, req.Reason), userID, []JournalLine{
		{AccountCode: EscrowLiability, Direction: Debit, Amount: req.Amount},
		{AccountCode: CashOnHand, Direction: Credit, Amount: req.Amount},
	})
	if err != nil {
		return common.NewInternalError("failed to post refund journal entry", err)
	}

	escrow, err := s.repo.GetEscrow(ctx, req.OrderType, req.OrderID)
	if err == nil {
		if err := s.repo.UpdateEscrowRefunded(ctx, tx, escrow.ID, req.Amount); err != nil {
			s.logger.Error("failed to update escrow refunded", zap.Error(err))
		}
	}

	return tx.Commit(ctx)
}
func (s *Service) RequestWithdrawal(ctx context.Context, courierID string, req WithdrawalCreateRequest) (*WithdrawalRequestDTO, error) {
	if s.riskSvc != nil {
		decision, err := s.riskSvc.EvaluateAction(ctx, courierID, common.RiskActionRequestWithdrawal)
		if err == nil && !decision.Allowed {
			return nil, common.NewForbiddenError("action blocked by risk engine: " + decision.Reason)
		}
	}

	if req.Amount <= 0 {
		return nil, common.NewBadRequestError("withdrawal amount must be positive")
	}

	now := time.Now()
	dailyTotal, err := s.repo.GetDailyWithdrawalTotal(ctx, courierID, now)
	if err != nil {
		return nil, common.NewInternalError("failed to check daily withdrawal total", err)
	}

	if dailyTotal+req.Amount > MaxDailyWithdrawal {
		remaining := MaxDailyWithdrawal - dailyTotal
		return nil, common.NewBadRequestError(fmt.Sprintf("daily withdrawal cap exceeded; remaining allowance: $%.2f", remaining))
	}

	wr := &WithdrawalRequest{
		CourierID:     courierID,
		RequestAmount: req.Amount,
		Status:        WithdrawalStatusRequested,
	}

	id, err := s.repo.CreateWithdrawalRequest(ctx, wr)
	if err != nil {
		return nil, common.NewInternalError("failed to create withdrawal request", err)
	}

	created, err := s.repo.GetWithdrawalRequest(ctx, id)
	if err != nil {
		return nil, common.NewInternalError("failed to fetch withdrawal request", err)
	}

	return &WithdrawalRequestDTO{
		ID:            created.ID,
		CourierID:     created.CourierID,
		RequestAmount: created.RequestAmount,
		Status:        created.Status,
		RequestedAt:   created.RequestedAt,
		CreatedAt:     created.CreatedAt,
	}, nil
}

func (s *Service) ApproveWithdrawal(ctx context.Context, withdrawalID, approverID string) error {
	wr, err := s.repo.GetWithdrawalRequest(ctx, withdrawalID)
	if err != nil {
		return err
	}

	if wr.Status != WithdrawalStatusRequested {
		return common.NewConflictError("withdrawal request is not in a requestable state")
	}

	now := time.Now()
	dailyTotal, err := s.repo.GetDailyWithdrawalTotal(ctx, wr.CourierID, now)
	if err != nil {
		return common.NewInternalError("failed to check daily withdrawal total", err)
	}
	if dailyTotal > MaxDailyWithdrawal {
		return common.NewBadRequestError("daily withdrawal cap would be exceeded at approval time")
	}

	tx, err := s.repo.BeginTx(ctx)
	if err != nil {
		return common.NewInternalError("failed to begin transaction", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	refUUID, _ := uuid.Parse(withdrawalID)
	err = PostJournalEntry(ctx, tx, "withdrawal_disbursement", "withdrawal_request", refUUID, fmt.Sprintf("Withdrawal disbursement for courier %s", wr.CourierID), approverID, []JournalLine{
		{AccountCode: CourierPayable, Direction: Debit, Amount: wr.RequestAmount},
		{AccountCode: CashOnHand, Direction: Credit, Amount: wr.RequestAmount},
	})
	if err != nil {
		return common.NewInternalError("failed to post withdrawal journal entry", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return common.NewInternalError("failed to commit transaction", err)
	}

	if err := s.repo.CreateWithdrawalDisbursement(ctx, withdrawalID, wr.RequestAmount); err != nil {
		s.logger.Error("failed to create disbursement record", zap.Error(err))
	}

	if err := s.repo.UpdateWithdrawalApproval(ctx, withdrawalID, approverID, string(WithdrawalStatusSettled)); err != nil {
		return common.NewInternalError("failed to update withdrawal status", err)
	}

	return nil
}

func (s *Service) RejectWithdrawal(ctx context.Context, withdrawalID, approverID, reason string) error {
	wr, err := s.repo.GetWithdrawalRequest(ctx, withdrawalID)
	if err != nil {
		return err
	}

	if wr.Status != WithdrawalStatusRequested {
		return common.NewConflictError("withdrawal request is not in a requestable state")
	}

	if reason == "" {
		return common.NewBadRequestError("rejection reason is required")
	}

	return s.repo.UpdateWithdrawalRejection(ctx, withdrawalID, approverID, reason)
}

func (s *Service) GetActiveEscrows(ctx context.Context, ownerID string) ([]EscrowSummary, error) {
	return s.repo.GetActiveEscrowsByOwner(ctx, ownerID)
}
func (s *Service) ReleaseEscrow(ctx context.Context, orderType, orderID string, amount float64, createdBy string) error {
	if amount <= 0 {
		return common.NewBadRequestError("release amount must be positive")
	}

	escrow, err := s.repo.GetEscrow(ctx, orderType, orderID)
	if err != nil {
		return err
	}

	remaining := escrow.AmountHeld - escrow.AmountReleased - escrow.AmountRefunded
	if amount > remaining {
		return common.NewBadRequestError(fmt.Sprintf("release amount exceeds remaining escrow balance of $%.2f", remaining))
	}

	tx, err := s.repo.BeginTx(ctx)
	if err != nil {
		return common.NewInternalError("failed to begin transaction", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if err := s.repo.UpdateEscrowReleased(ctx, tx, escrow.ID, amount); err != nil {
		return common.NewInternalError("failed to update escrow", err)
	}

	refUUID, _ := uuid.Parse(orderID)
	err = PostJournalEntry(ctx, tx, "escrow_release", orderType, refUUID, fmt.Sprintf("Escrow release for %s %s", orderType, orderID), createdBy, []JournalLine{
		{AccountCode: EscrowLiability, Direction: Debit, Amount: amount},
		{AccountCode: SupplierPayable, Direction: Credit, Amount: amount},
	})
	if err != nil {
		return common.NewInternalError("failed to post escrow release journal entry", err)
	}

	return tx.Commit(ctx)
}
func (s *Service) GetReconciliation(ctx context.Context) (*ReconciliationReportDTO, error) {
	report, err := s.repo.GetReconciliationSummary(ctx)
	if err != nil {
		return nil, common.NewInternalError("failed to generate reconciliation report", err)
	}
	return report, nil
}
