package procurement

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"travel-platform/apps/api/internal/common"
	"travel-platform/apps/api/internal/modules/finance"
	"travel-platform/apps/api/internal/modules/risk"
)

type Service struct {
	repo    *Repository
	riskSvc *risk.Service
	logger  *zap.Logger
}

func NewService(repo *Repository, riskSvc *risk.Service, logger *zap.Logger) *Service {
	return &Service{repo: repo, riskSvc: riskSvc, logger: logger}
}
func (s *Service) CreateRFQ(ctx context.Context, userID string, req CreateRFQRequest) (*RFQResponse, error) {
	if req.Title == "" {
		return nil, common.NewBadRequestError("title is required")
	}

	decision, err := s.riskSvc.EvaluateAction(ctx, userID, common.RiskActionCreateRFQ)
	if err == nil && !decision.Allowed {
		return nil, common.NewForbiddenError("action blocked by risk engine: " + decision.Reason)
	}

	deadline, err := time.Parse(time.RFC3339, req.Deadline)
	if err != nil {
		return nil, common.NewBadRequestError("deadline must be a valid RFC3339 timestamp")
	}

	rfq := &RFQ{
		CreatedBy:   userID,
		Title:       req.Title,
		Description: req.Description,
		Deadline:    deadline,
		Status:      RFQStatusDraft,
	}

	id, err := s.repo.CreateRFQ(ctx, rfq)
	if err != nil {
		return nil, common.NewInternalError("failed to create rfq", err)
	}

	created, err := s.repo.GetRFQ(ctx, id)
	if err != nil {
		return nil, common.NewInternalError("failed to fetch created rfq", err)
	}

	return &RFQResponse{
		ID:          created.ID,
		CreatedBy:   created.CreatedBy,
		Title:       created.Title,
		Description: created.Description,
		Deadline:    created.Deadline,
		Status:      created.Status,
		Items:       []RFQItem{},
		CreatedAt:   created.CreatedAt,
		UpdatedAt:   created.UpdatedAt,
	}, nil
}

func (s *Service) AddRFQItems(ctx context.Context, rfqID, userID string, items []RFQItemRequest) error {
	rfq, err := s.repo.GetRFQ(ctx, rfqID)
	if err != nil {
		return err
	}
	if rfq.Status != RFQStatusDraft {
		return common.NewConflictError("can only add items to a draft RFQ")
	}

	for _, item := range items {
		ri := &RFQItem{
			RFQID:          rfqID,
			ItemName:       item.ItemName,
			Specifications: item.Specifications,
			Quantity:       item.Quantity,
			Unit:           item.Unit,
			SortOrder:      item.SortOrder,
		}
		if err := s.repo.CreateRFQItem(ctx, ri); err != nil {
			return common.NewInternalError("failed to add rfq item", err)
		}
	}
	return nil
}

func (s *Service) IssueRFQ(ctx context.Context, rfqID, userID string, req IssueRFQRequest) error {
	if s.riskSvc != nil {
		decision, err := s.riskSvc.EvaluateAction(ctx, userID, common.RiskActionIssueRFQ)
		if err == nil && !decision.Allowed {
			return common.NewForbiddenError("action blocked by risk engine: " + decision.Reason)
		}
	}

	rfq, err := s.repo.GetRFQ(ctx, rfqID)
	if err != nil {
		return err
	}
	if rfq.Status != RFQStatusDraft {
		return common.NewConflictError("can only issue a draft RFQ")
	}

	if len(req.Items) > 0 {
		if err := s.AddRFQItems(ctx, rfqID, userID, req.Items); err != nil {
			return err
		}
	}

	for _, sid := range req.SupplierIDs {
		if err := s.repo.InviteSupplier(ctx, rfqID, sid); err != nil {
			s.logger.Error("failed to invite supplier", zap.String("supplierId", sid), zap.Error(err))
		}
	}

	return s.repo.UpdateRFQStatus(ctx, rfqID, RFQStatusIssued)
}

func (s *Service) SubmitQuote(ctx context.Context, rfqID, supplierID string, req SubmitQuoteRequest) (*QuoteResponse, error) {
	rfq, err := s.repo.GetRFQ(ctx, rfqID)
	if err != nil {
		return nil, err
	}
	if rfq.Status != RFQStatusIssued {
		return nil, common.NewConflictError("can only submit quotes for an issued RFQ")
	}

	quote := &RFQQuote{
		RFQID:        rfqID,
		SupplierID:   supplierID,
		TotalAmount:  req.TotalAmount,
		LeadTimeDays: req.LeadTimeDays,
		Notes:        req.Notes,
	}

	quoteID, err := s.repo.CreateQuote(ctx, quote)
	if err != nil {
		return nil, common.NewInternalError("failed to create quote", err)
	}

	for _, item := range req.Items {
		qi := &RFQQuoteItem{
			QuoteID:   quoteID,
			RFQItemID: item.RFQItemID,
			UnitPrice: item.UnitPrice,
			Quantity:  item.Quantity,
			Subtotal:  item.UnitPrice * float64(item.Quantity),
			Notes:     item.Notes,
		}
		if err := s.repo.CreateQuoteItem(ctx, qi); err != nil {
			s.logger.Error("failed to create quote item", zap.Error(err))
		}
	}

	created, err := s.repo.GetQuote(ctx, quoteID)
	if err != nil {
		return nil, common.NewInternalError("failed to fetch created quote", err)
	}
	items, _ := s.repo.GetQuoteItems(ctx, quoteID)
	if items == nil {
		items = []RFQQuoteItem{}
	}

	return &QuoteResponse{
		ID:           created.ID,
		RFQID:        created.RFQID,
		SupplierID:   created.SupplierID,
		TotalAmount:  created.TotalAmount,
		LeadTimeDays: created.LeadTimeDays,
		Notes:        created.Notes,
		Items:        items,
		SubmittedAt:  created.SubmittedAt,
	}, nil
}

func (s *Service) GetComparison(ctx context.Context, rfqID, userID string) (*ComparisonMatrixResponse, error) {
	rfq, err := s.repo.GetRFQ(ctx, rfqID)
	if err != nil {
		return nil, err
	}
	_ = rfq

	rfqItems, err := s.repo.GetRFQItems(ctx, rfqID)
	if err != nil {
		return nil, common.NewInternalError("failed to fetch rfq items", err)
	}
	if rfqItems == nil {
		rfqItems = []RFQItem{}
	}

	quotes, err := s.repo.GetQuotesByRFQ(ctx, rfqID)
	if err != nil {
		return nil, common.NewInternalError("failed to fetch quotes", err)
	}

	entries := make([]ComparisonMatrixEntry, 0, len(quotes))
	for _, q := range quotes {
		items, _ := s.repo.GetQuoteItems(ctx, q.ID)
		if items == nil {
			items = []RFQQuoteItem{}
		}
		entries = append(entries, ComparisonMatrixEntry{
			SupplierID:   q.SupplierID,
			TotalAmount:  q.TotalAmount,
			LeadTimeDays: q.LeadTimeDays,
			Notes:        q.Notes,
			Items:        items,
			SubmittedAt:  q.SubmittedAt,
		})
	}

	return &ComparisonMatrixResponse{
		RFQID:  rfqID,
		Items:  rfqItems,
		Quotes: entries,
	}, nil
}

func (s *Service) SelectSupplier(ctx context.Context, rfqID, userID, quoteID string) error {
	if s.riskSvc != nil {
		decision, err := s.riskSvc.EvaluateAction(ctx, userID, common.RiskActionSelectSupplier)
		if err == nil && !decision.Allowed {
			return common.NewForbiddenError("action blocked by risk engine: " + decision.Reason)
		}
	}

	rfq, err := s.repo.GetRFQ(ctx, rfqID)
	if err != nil {
		return err
	}
	if rfq.Status != RFQStatusIssued {
		return common.NewConflictError("can only select a supplier for an issued RFQ")
	}

	quote, err := s.repo.GetQuote(ctx, quoteID)
	if err != nil {
		return err
	}
	if quote.RFQID != rfqID {
		return common.NewBadRequestError("quote does not belong to this RFQ")
	}

	return s.repo.UpdateRFQStatus(ctx, rfqID, RFQStatusSelected)
}

func (s *Service) ConvertToPO(ctx context.Context, rfqID, quoteID, userID string) (*POResponse, error) {
	rfq, err := s.repo.GetRFQ(ctx, rfqID)
	if err != nil {
		return nil, err
	}
	if rfq.Status != RFQStatusSelected && rfq.Status != RFQStatusIssued {
		return nil, common.NewConflictError("RFQ must be in selected or issued status to convert to PO")
	}

	quote, err := s.repo.GetQuote(ctx, quoteID)
	if err != nil {
		return nil, err
	}

	poNumber, err := s.GeneratePONumber(ctx)
	if err != nil {
		return nil, common.NewInternalError("failed to generate PO number", err)
	}

	po := &PurchaseOrder{
		RFQID:       &rfqID,
		QuoteID:     &quoteID,
		SupplierID:  quote.SupplierID,
		CreatedBy:   userID,
		PONumber:    poNumber,
		Status:      POStatusIssued,
		TotalAmount: quote.TotalAmount,
	}

	poID, err := s.repo.CreatePO(ctx, po)
	if err != nil {
		return nil, common.NewInternalError("failed to create purchase order", err)
	}

	quoteItems, _ := s.repo.GetQuoteItems(ctx, quoteID)
	for _, qi := range quoteItems {
		poItem := &POItem{
			POID:      poID,
			ItemName:  qi.RFQItemID,
			UnitPrice: qi.UnitPrice,
			Quantity:  qi.Quantity,
			Subtotal:  qi.Subtotal,
		}
		if err := s.repo.CreatePOItem(ctx, poItem); err != nil {
			s.logger.Error("failed to create po item from quote", zap.Error(err))
		}
	}

	if err := s.repo.UpdateRFQStatus(ctx, rfqID, RFQStatusConvertedPO); err != nil {
		s.logger.Error("failed to update rfq status to converted", zap.Error(err))
	}

	created, err := s.repo.GetPO(ctx, poID)
	if err != nil {
		return nil, common.NewInternalError("failed to fetch created PO", err)
	}
	poItems, _ := s.repo.GetPOItems(ctx, poID)
	if poItems == nil {
		poItems = []POItem{}
	}

	return &POResponse{
		ID:          created.ID,
		RFQID:       created.RFQID,
		QuoteID:     created.QuoteID,
		SupplierID:  created.SupplierID,
		CreatedBy:   created.CreatedBy,
		PONumber:    created.PONumber,
		PromisedDate: created.PromisedDate,
		Status:      created.Status,
		TotalAmount: created.TotalAmount,
		Items:       poItems,
		CreatedAt:   created.CreatedAt,
		UpdatedAt:   created.UpdatedAt,
	}, nil
}

func (s *Service) AcceptPO(ctx context.Context, poID, supplierID string) error {
	po, err := s.repo.GetPO(ctx, poID)
	if err != nil {
		return err
	}
	if po.SupplierID != supplierID {
		return common.NewForbiddenError("only the assigned supplier can accept this PO")
	}
	if po.Status != POStatusIssued {
		return common.NewConflictError("can only accept an issued purchase order")
	}
	return s.repo.UpdatePOStatus(ctx, poID, POStatusAccepted)
}
func (s *Service) RecordDelivery(ctx context.Context, poID, userID string, req RecordDeliveryRequest) (*DeliveryResponse, error) {
	po, err := s.repo.GetPO(ctx, poID)
	if err != nil {
		return nil, err
	}
	if po.Status != POStatusAccepted && po.Status != POStatusPartiallyDelivered {
		return nil, common.NewConflictError("can only record delivery for an accepted or partially-delivered PO")
	}

	deliveryDate, err := time.Parse(time.RFC3339, req.DeliveryDate)
	if err != nil {
		return nil, common.NewBadRequestError("deliveryDate must be a valid RFC3339 timestamp")
	}

	d := &Delivery{
		POID:         poID,
		CourierID:    req.CourierID,
		ReceivedBy:   &userID,
		DeliveryDate: deliveryDate,
		Notes:        req.Notes,
		Status:       "delivered",
	}

	deliveryID, err := s.repo.CreateDelivery(ctx, d)
	if err != nil {
		return nil, common.NewInternalError("failed to create delivery", err)
	}

	for _, item := range req.Items {
		di := &DeliveryItem{
			DeliveryID:        deliveryID,
			POItemID:          item.POItemID,
			QuantityDelivered: item.QuantityDelivered,
			QuantityAccepted:  item.QuantityAccepted,
			QuantityRejected:  item.QuantityRejected,
		}
		if err := s.repo.CreateDeliveryItem(ctx, di); err != nil {
			s.logger.Error("failed to create delivery item", zap.Error(err))
		}
	}

	if err := s.repo.UpdatePOStatus(ctx, poID, POStatusDelivered); err != nil {
		s.logger.Error("failed to update PO status", zap.Error(err))
	}

	created, err := s.repo.GetDelivery(ctx, deliveryID)
	if err != nil {
		return nil, common.NewInternalError("failed to fetch delivery", err)
	}
	items, _ := s.repo.GetDeliveryItems(ctx, deliveryID)
	if items == nil {
		items = []DeliveryItem{}
	}

	return &DeliveryResponse{
		ID:           created.ID,
		POID:         created.POID,
		CourierID:    created.CourierID,
		ReceivedBy:   created.ReceivedBy,
		DeliveryDate: created.DeliveryDate,
		Notes:        created.Notes,
		Status:       created.Status,
		Items:        items,
		CreatedAt:    created.CreatedAt,
	}, nil
}

func (s *Service) PerformInspection(ctx context.Context, deliveryID, inspectorID string, req InspectionRequest) (*InspectionResponse, error) {
	delivery, err := s.repo.GetDelivery(ctx, deliveryID)
	if err != nil {
		return nil, err
	}

	if req.Status == InspectionStatusFailed && req.Notes == "" {
		return nil, common.NewBadRequestError("notes are required when inspection status is failed")
	}

	insp := &QualityInspection{
		DeliveryID:  &deliveryID,
		POID:        delivery.POID,
		InspectorID: inspectorID,
		Status:      req.Status,
		Notes:       req.Notes,
	}

	inspID, err := s.repo.CreateInspection(ctx, insp)
	if err != nil {
		return nil, common.NewInternalError("failed to create inspection", err)
	}

	if req.Status == InspectionStatusFailed {
		disc := &DiscrepancyTicket{
			POID:            delivery.POID,
			DeliveryID:      &deliveryID,
			InspectionID:    &inspID,
			DiscrepancyType: DiscrepancyTypeDamage,
			Description:     fmt.Sprintf("Failed inspection: %s", req.Notes),
			Notes:           req.Notes,
			CreatedBy:       inspectorID,
		}
		if _, err := s.repo.CreateDiscrepancy(ctx, disc); err != nil {
			s.logger.Error("failed to auto-create discrepancy", zap.Error(err))
		}
	}

	return &InspectionResponse{
		ID:          inspID,
		DeliveryID:  &deliveryID,
		POID:        delivery.POID,
		InspectorID: inspectorID,
		Status:      req.Status,
		Notes:       req.Notes,
		InspectedAt: time.Now(),
	}, nil
}

func (s *Service) CreateDiscrepancy(ctx context.Context, poID, userID string, req CreateDiscrepancyRequest) error {
	disc := &DiscrepancyTicket{
		POID:            req.POID,
		DeliveryID:      req.DeliveryID,
		InspectionID:    req.InspectionID,
		DiscrepancyType: req.DiscrepancyType,
		Description:     req.Description,
		Notes:           req.Notes,
		CreatedBy:       userID,
	}
	_, err := s.repo.CreateDiscrepancy(ctx, disc)
	if err != nil {
		return common.NewInternalError("failed to create discrepancy", err)
	}
	return nil
}
func (s *Service) OpenException(ctx context.Context, referenceType, referenceID, reason string) (string, error) {
	ec := &ExceptionCase{
		ReferenceType: referenceType,
		ReferenceID:   referenceID,
		Status:        ExceptionStatusOpen,
		OpenedReason:  reason,
	}
	id, err := s.repo.CreateExceptionCase(ctx, ec)
	if err != nil {
		return "", common.NewInternalError("failed to open exception", err)
	}
	return id, nil
}

func (s *Service) SubmitWaiver(ctx context.Context, exceptionID, approverID, reason string) error {
	ec, err := s.repo.GetExceptionCase(ctx, exceptionID)
	if err != nil {
		return err
	}
	if ec.Status != ExceptionStatusOpen {
		return common.NewConflictError("can only submit waivers for open exceptions")
	}

	w := &WaiverRecord{
		ExceptionCaseID: exceptionID,
		ApprovedBy:      approverID,
		WaiverReason:    reason,
	}
	return s.repo.CreateWaiver(ctx, w)
}

func (s *Service) SubmitSettlementAdjustment(ctx context.Context, exceptionID, approverID string, amount float64, direction, reason string) error {
	ec, err := s.repo.GetExceptionCase(ctx, exceptionID)
	if err != nil {
		return err
	}
	if ec.Status != ExceptionStatusOpen {
		return common.NewConflictError("can only submit adjustments for open exceptions")
	}

	tx, err := s.repo.BeginTx(ctx)
	if err != nil {
		return common.NewInternalError("failed to begin transaction", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	refUUID, _ := uuid.Parse(exceptionID)
	var lines []finance.JournalLine
	if direction == "debit" {
		lines = []finance.JournalLine{
			{AccountCode: finance.AdjustmentExpense, Direction: finance.Debit, Amount: amount},
			{AccountCode: finance.SettlementAdjustmentReserve, Direction: finance.Credit, Amount: amount},
		}
	} else {
		lines = []finance.JournalLine{
			{AccountCode: finance.SettlementAdjustmentReserve, Direction: finance.Debit, Amount: amount},
			{AccountCode: finance.AdjustmentExpense, Direction: finance.Credit, Amount: amount},
		}
	}

	err = finance.PostJournalEntry(ctx, tx, "settlement_adjustment", "exception_case", refUUID, fmt.Sprintf("Settlement adjustment for exception %s: %s", exceptionID, reason), approverID, lines)
	if err != nil {
		return common.NewInternalError("failed to post journal entry", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return common.NewInternalError("failed to commit transaction", err)
	}

	sa := &SettlementAdjustment{
		ExceptionCaseID: exceptionID,
		Amount:          amount,
		Direction:       direction,
		Reason:          reason,
		ApprovedBy:      approverID,
	}
	if _, err := s.repo.CreateSettlementAdjustment(ctx, sa); err != nil {
		return common.NewInternalError("failed to create settlement adjustment", err)
	}

	return nil
}

func (s *Service) CloseException(ctx context.Context, exceptionID, userID string) error {
	ec, err := s.repo.GetExceptionCase(ctx, exceptionID)
	if err != nil {
		return err
	}
	if ec.Status != ExceptionStatusOpen {
		return common.NewConflictError("exception is not open")
	}

	waiverCount, err := s.repo.CountWaiversByException(ctx, exceptionID)
	if err != nil {
		return common.NewInternalError("failed to count waivers", err)
	}
	adjCount, err := s.repo.CountAdjustmentsByException(ctx, exceptionID)
	if err != nil {
		return common.NewInternalError("failed to count adjustments", err)
	}

	if waiverCount == 0 && adjCount == 0 {
		return common.NewBadRequestError("cannot close exception without at least one settlement adjustment or waiver record")
	}

	return s.repo.CloseExceptionCase(ctx, exceptionID)
}
func (s *Service) GeneratePONumber(ctx context.Context) (string, error) {
	year := time.Now().Year()
	seq, err := s.repo.GetNextPOSequence(ctx, year)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("PO-%d-%04d", year, seq), nil
}

func (s *Service) GetRFQ(ctx context.Context, rfqID, userID string, roles []string) (*RFQResponse, error) {
	rfq, err := s.repo.GetRFQ(ctx, rfqID)
	if err != nil {
		return nil, err
	}

	if !common.IsAdminOrAccountant(roles) {
		if common.HasRole(roles, common.RoleSupplier) {
			invited, err := s.repo.IsSupplierInvited(ctx, rfqID, userID)
			if err != nil {
				return nil, common.NewInternalError("failed to check supplier invitation", err)
			}
			if !invited {
				return nil, common.NewForbiddenError("you do not have access to this RFQ")
			}
		} else if common.HasRole(roles, common.RoleGroupOrganizer) {
			if rfq.CreatedBy != userID {
				return nil, common.NewForbiddenError("you do not have access to this RFQ")
			}
		} else {
			return nil, common.NewForbiddenError("you do not have access to this RFQ")
		}
	}

	items, err := s.repo.GetRFQItems(ctx, rfqID)
	if err != nil {
		return nil, common.NewInternalError("failed to fetch rfq items", err)
	}
	if items == nil {
		items = []RFQItem{}
	}
	return &RFQResponse{
		ID:          rfq.ID,
		CreatedBy:   rfq.CreatedBy,
		Title:       rfq.Title,
		Description: rfq.Description,
		Deadline:    rfq.Deadline,
		Status:      rfq.Status,
		Items:       items,
		CreatedAt:   rfq.CreatedAt,
		UpdatedAt:   rfq.UpdatedAt,
	}, nil
}

func (s *Service) getPOInternal(ctx context.Context, poID string) (*POResponse, error) {
	po, err := s.repo.GetPO(ctx, poID)
	if err != nil {
		return nil, err
	}
	items, _ := s.repo.GetPOItems(ctx, poID)
	if items == nil {
		items = []POItem{}
	}
	return &POResponse{
		ID:           po.ID,
		RFQID:        po.RFQID,
		QuoteID:      po.QuoteID,
		SupplierID:   po.SupplierID,
		CreatedBy:    po.CreatedBy,
		PONumber:     po.PONumber,
		PromisedDate: po.PromisedDate,
		Status:       po.Status,
		TotalAmount:  po.TotalAmount,
		Items:        items,
		CreatedAt:    po.CreatedAt,
		UpdatedAt:    po.UpdatedAt,
	}, nil
}

func (s *Service) GetPO(ctx context.Context, poID, userID string, roles []string) (*POResponse, error) {
	po, err := s.repo.GetPO(ctx, poID)
	if err != nil {
		return nil, err
	}

	if !common.IsAdminOrAccountant(roles) {
		if common.HasRole(roles, common.RoleSupplier) {
			if po.SupplierID != userID {
				return nil, common.NewForbiddenError("you do not have access to this purchase order")
			}
		} else if common.HasRole(roles, common.RoleGroupOrganizer) {
			if po.CreatedBy != userID {
				return nil, common.NewForbiddenError("you do not have access to this purchase order")
			}
		} else {
			return nil, common.NewForbiddenError("you do not have access to this purchase order")
		}
	}

	items, _ := s.repo.GetPOItems(ctx, poID)
	if items == nil {
		items = []POItem{}
	}
	return &POResponse{
		ID:           po.ID,
		RFQID:        po.RFQID,
		QuoteID:      po.QuoteID,
		SupplierID:   po.SupplierID,
		CreatedBy:    po.CreatedBy,
		PONumber:     po.PONumber,
		PromisedDate: po.PromisedDate,
		Status:       po.Status,
		TotalAmount:  po.TotalAmount,
		Items:        items,
		CreatedAt:    po.CreatedAt,
		UpdatedAt:    po.UpdatedAt,
	}, nil
}

func (s *Service) CreatePO(ctx context.Context, userID string, req CreatePORequest) (*POResponse, error) {
	if req.SupplierID == "" {
		return nil, common.NewBadRequestError("supplierId is required")
	}

	poNumber, err := s.GeneratePONumber(ctx)
	if err != nil {
		return nil, common.NewInternalError("failed to generate PO number", err)
	}

	var promisedDate *time.Time
	if req.PromisedDate != nil {
		t, err := time.Parse(time.RFC3339, *req.PromisedDate)
		if err != nil {
			return nil, common.NewBadRequestError("promisedDate must be a valid RFC3339 timestamp")
		}
		promisedDate = &t
	}

	var totalAmount float64
	for _, item := range req.Items {
		totalAmount += item.UnitPrice * float64(item.Quantity)
	}

	po := &PurchaseOrder{
		RFQID:        req.RFQID,
		QuoteID:      req.QuoteID,
		SupplierID:   req.SupplierID,
		CreatedBy:    userID,
		PONumber:     poNumber,
		PromisedDate: promisedDate,
		Status:       POStatusIssued,
		TotalAmount:  totalAmount,
	}

	poID, err := s.repo.CreatePO(ctx, po)
	if err != nil {
		return nil, common.NewInternalError("failed to create purchase order", err)
	}

	for _, item := range req.Items {
		poItem := &POItem{
			POID:           poID,
			ItemName:       item.ItemName,
			Specifications: item.Specifications,
			UnitPrice:      item.UnitPrice,
			Quantity:       item.Quantity,
			Subtotal:       item.UnitPrice * float64(item.Quantity),
		}
		if err := s.repo.CreatePOItem(ctx, poItem); err != nil {
			s.logger.Error("failed to create po item", zap.Error(err))
		}
	}

	return s.getPOInternal(ctx, poID)
}
func (s *Service) ListRFQs(ctx context.Context, userID string, roles []string) ([]map[string]interface{}, error) {
	var query string
	var args []interface{}

	if common.IsAdminOrAccountant(roles) {
		query = `SELECT id, created_by, title, description, deadline, status, created_at, updated_at FROM rfqs ORDER BY created_at DESC`
	} else if common.HasRole(roles, common.RoleSupplier) {
		query = `SELECT r.id, r.created_by, r.title, r.description, r.deadline, r.status, r.created_at, r.updated_at
		         FROM rfqs r INNER JOIN rfq_suppliers rs ON r.id = rs.rfq_id
		         WHERE rs.supplier_id = $1 ORDER BY r.created_at DESC`
		args = append(args, userID)
	} else {
		query = `SELECT id, created_by, title, description, deadline, status, created_at, updated_at FROM rfqs WHERE created_by = $1 ORDER BY created_at DESC`
		args = append(args, userID)
	}

	rows, err := s.repo.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, common.NewInternalError("failed to list rfqs", err)
	}
	defer rows.Close()

	var result []map[string]interface{}
	for rows.Next() {
		var id, createdBy, title, status string
		var description *string
		var deadline, createdAt, updatedAt time.Time
		if err := rows.Scan(&id, &createdBy, &title, &description, &deadline, &status, &createdAt, &updatedAt); err != nil {
			return nil, common.NewInternalError("failed to scan rfq", err)
		}
		result = append(result, map[string]interface{}{
			"id": id, "createdBy": createdBy, "title": title, "description": description,
			"deadline": deadline, "status": status, "createdAt": createdAt, "updatedAt": updatedAt,
		})
	}
	if result == nil {
		result = []map[string]interface{}{}
	}
	return result, rows.Err()
}

func (s *Service) ListPOs(ctx context.Context, userID string, roles []string) ([]map[string]interface{}, error) {
	var query string
	var args []interface{}

	if common.IsAdminOrAccountant(roles) {
		query = `SELECT id, rfq_id, quote_id, supplier_id, created_by, po_number, promised_date, status, total_amount, created_at, updated_at FROM purchase_orders ORDER BY created_at DESC`
	} else if common.HasRole(roles, common.RoleSupplier) {
		query = `SELECT id, rfq_id, quote_id, supplier_id, created_by, po_number, promised_date, status, total_amount, created_at, updated_at FROM purchase_orders WHERE supplier_id = $1 ORDER BY created_at DESC`
		args = append(args, userID)
	} else {
		query = `SELECT id, rfq_id, quote_id, supplier_id, created_by, po_number, promised_date, status, total_amount, created_at, updated_at FROM purchase_orders WHERE created_by = $1 ORDER BY created_at DESC`
		args = append(args, userID)
	}

	rows, err := s.repo.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, common.NewInternalError("failed to list purchase orders", err)
	}
	defer rows.Close()

	var result []map[string]interface{}
	for rows.Next() {
		var id, supplierID, createdBy, poNumber, status string
		var rfqID, quoteID *string
		var promisedDate *time.Time
		var totalAmount float64
		var createdAt, updatedAt time.Time
		if err := rows.Scan(&id, &rfqID, &quoteID, &supplierID, &createdBy, &poNumber, &promisedDate, &status, &totalAmount, &createdAt, &updatedAt); err != nil {
			return nil, common.NewInternalError("failed to scan purchase order", err)
		}
		result = append(result, map[string]interface{}{
			"id": id, "rfqId": rfqID, "quoteId": quoteID, "supplierId": supplierID,
			"createdBy": createdBy, "poNumber": poNumber, "promisedDate": promisedDate,
			"status": status, "totalAmount": totalAmount, "createdAt": createdAt, "updatedAt": updatedAt,
		})
	}
	if result == nil {
		result = []map[string]interface{}{}
	}
	return result, rows.Err()
}

func (s *Service) ListDeliveries(ctx context.Context, userID string, roles []string) ([]map[string]interface{}, error) {
	var query string
	var args []interface{}

	if common.IsAdminOrAccountant(roles) {
		query = `SELECT id, po_id, courier_id, received_by, delivery_date, notes, status, created_at, updated_at FROM deliveries ORDER BY created_at DESC`
	} else if common.HasRole(roles, common.RoleSupplier) {
		query = `SELECT d.id, d.po_id, d.courier_id, d.received_by, d.delivery_date, d.notes, d.status, d.created_at, d.updated_at
		         FROM deliveries d INNER JOIN purchase_orders po ON d.po_id = po.id
		         WHERE po.supplier_id = $1 ORDER BY d.created_at DESC`
		args = append(args, userID)
	} else {
		query = `SELECT d.id, d.po_id, d.courier_id, d.received_by, d.delivery_date, d.notes, d.status, d.created_at, d.updated_at
		         FROM deliveries d INNER JOIN purchase_orders po ON d.po_id = po.id
		         WHERE po.created_by = $1 ORDER BY d.created_at DESC`
		args = append(args, userID)
	}

	rows, err := s.repo.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, common.NewInternalError("failed to list deliveries", err)
	}
	defer rows.Close()

	var result []map[string]interface{}
	for rows.Next() {
		var id, poID, status string
		var courierID, receivedBy *string
		var notes *string
		var deliveryDate time.Time
		var createdAt, updatedAt time.Time
		if err := rows.Scan(&id, &poID, &courierID, &receivedBy, &deliveryDate, &notes, &status, &createdAt, &updatedAt); err != nil {
			return nil, common.NewInternalError("failed to scan delivery", err)
		}
		result = append(result, map[string]interface{}{
			"id": id, "poId": poID, "courierId": courierID, "receivedBy": receivedBy,
			"deliveryDate": deliveryDate, "notes": notes, "status": status,
			"createdAt": createdAt, "updatedAt": updatedAt,
		})
	}
	if result == nil {
		result = []map[string]interface{}{}
	}
	return result, rows.Err()
}

func (s *Service) ListExceptions(ctx context.Context, userID string, roles []string) ([]map[string]interface{}, error) {
	var query string
	var args []interface{}

	if common.IsAdminOrAccountant(roles) {
		query = `SELECT id, reference_type, reference_id, status, opened_reason, opened_at, closed_at, created_at, updated_at FROM exception_cases ORDER BY created_at DESC`
	} else {
		query = `SELECT ec.id, ec.reference_type, ec.reference_id, ec.status, ec.opened_reason, ec.opened_at, ec.closed_at, ec.created_at, ec.updated_at
		         FROM exception_cases ec
		         LEFT JOIN purchase_orders po ON ec.reference_id = po.id::text
		         WHERE po.created_by = $1 OR po.supplier_id = $1
		         ORDER BY ec.created_at DESC`
		args = append(args, userID)
	}

	rows, err := s.repo.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, common.NewInternalError("failed to list exceptions", err)
	}
	defer rows.Close()

	var result []map[string]interface{}
	for rows.Next() {
		var id, referenceType, referenceID, status, openedReason string
		var openedAt, createdAt, updatedAt time.Time
		var closedAt *time.Time
		if err := rows.Scan(&id, &referenceType, &referenceID, &status, &openedReason, &openedAt, &closedAt, &createdAt, &updatedAt); err != nil {
			return nil, common.NewInternalError("failed to scan exception", err)
		}
		result = append(result, map[string]interface{}{
			"id": id, "referenceType": referenceType, "referenceId": referenceID,
			"status": status, "openedReason": openedReason, "openedAt": openedAt,
			"closedAt": closedAt, "createdAt": createdAt, "updatedAt": updatedAt,
		})
	}
	if result == nil {
		result = []map[string]interface{}{}
	}
	return result, rows.Err()
}

type SupplierQuoteView struct {
	ID          string  `json:"id"`
	RFQID       string  `json:"rfqId"`
	RFQTitle    string  `json:"rfqTitle"`
	RFQStatus   string  `json:"rfqStatus"`
	RFQDeadline string  `json:"deadline"`
	TotalAmount float64 `json:"totalQuoted"`
	SubmittedAt *string `json:"submittedAt"`
	Status      string  `json:"status"`
}

func (s *Service) ListSupplierQuotes(ctx context.Context, supplierID string) ([]SupplierQuoteView, error) {
	return s.repo.GetQuotesBySupplier(ctx, supplierID)
}
