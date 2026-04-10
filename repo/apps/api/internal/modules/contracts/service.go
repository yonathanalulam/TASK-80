package contracts

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/textproto"
	"strings"
	"time"

	"go.uber.org/zap"

	"travel-platform/apps/api/internal/common"
	"travel-platform/apps/api/internal/modules/files"
)

type ContractService struct {
	repo        *Repository
	fileService *files.FileVaultService
	logger      *zap.Logger
}

func NewContractService(repo *Repository, fileService *files.FileVaultService, logger *zap.Logger) *ContractService {
	return &ContractService{repo: repo, fileService: fileService, logger: logger}
}

func (s *ContractService) GetTemplates(ctx context.Context) ([]ContractTemplate, error) {
	return s.repo.GetActiveTemplates(ctx)
}

func (s *ContractService) GenerateContract(ctx context.Context, userID string, req GenerateContractRequest) (*GeneratedContractResponse, error) {
	if req.TemplateID == "" {
		return nil, common.NewBadRequestError("templateId is required")
	}

	tmpl, err := s.repo.GetTemplateByID(ctx, req.TemplateID)
	if err != nil {
		return nil, err
	}

	if !tmpl.Active {
		return nil, common.NewBadRequestError("template is not active")
	}

	body := tmpl.BodyTemplate
	for key, value := range req.Variables {
		placeholder := "{{" + key + "}}"
		body = strings.ReplaceAll(body, placeholder, value)
	}

	content, err := GeneratePDF(tmpl.Name, body)
	if err != nil {
		return nil, common.NewInternalError("generate contract PDF", err)
	}
	header := createMultipartHeader("contract.pdf", "application/pdf", len(content))

	fileMeta, err := s.fileService.Upload(
		ctx, userID,
		newBytesFile(content),
		header,
		"contract", "",
		true,
	)
	if err != nil {
		return nil, common.NewInternalError("store contract file", err)
	}

	varsJSON, err := json.Marshal(req.Variables)
	if err != nil {
		return nil, common.NewInternalError("marshal variables", err)
	}

	gc := &GeneratedContract{
		TemplateID:    req.TemplateID,
		VariablesJSON: varsJSON,
		FileID:        &fileMeta.ID,
		GeneratedBy:   userID,
		Version:       tmpl.Version,
	}

	if err := s.repo.CreateGeneratedContract(ctx, gc); err != nil {
		return nil, common.NewInternalError("save generated contract", err)
	}

	s.logger.Info("contract generated",
		zap.String("contractId", gc.ID),
		zap.String("templateId", req.TemplateID),
		zap.String("generatedBy", userID),
	)

	return &GeneratedContractResponse{
		ID:          gc.ID,
		TemplateID:  gc.TemplateID,
		FileID:      gc.FileID,
		GeneratedBy: gc.GeneratedBy,
		GeneratedAt: gc.GeneratedAt,
		Version:     gc.Version,
	}, nil
}

func (s *ContractService) RequestInvoice(ctx context.Context, userID string, req CreateInvoiceRequestDTO) (*InvoiceRequest, error) {
	if req.OrderType == "" || req.OrderID == "" {
		return nil, common.NewBadRequestError("orderType and orderId are required")
	}

	ir := &InvoiceRequest{
		RequesterID: userID,
		OrderType:   req.OrderType,
		OrderID:     req.OrderID,
		Status:      InvoiceRequestPending,
		Notes:       req.Notes,
	}

	if err := s.repo.CreateInvoiceRequest(ctx, ir); err != nil {
		return nil, common.NewInternalError("create invoice request", err)
	}

	s.logger.Info("invoice requested",
		zap.String("requestId", ir.ID),
		zap.String("requesterId", userID),
	)

	return ir, nil
}

func (s *ContractService) ApproveInvoiceRequest(ctx context.Context, requestID, approverID string) error {
	ir, err := s.repo.GetInvoiceRequestByID(ctx, requestID)
	if err != nil {
		return err
	}

	if ir.Status != InvoiceRequestPending {
		return common.NewBadRequestError("only pending requests can be approved")
	}

	if err := s.repo.UpdateInvoiceRequestStatus(ctx, requestID, InvoiceRequestApproved); err != nil {
		return err
	}

	s.logger.Info("invoice request approved",
		zap.String("requestId", requestID),
		zap.String("approvedBy", approverID),
	)

	return nil
}

func (s *ContractService) GenerateInvoice(ctx context.Context, requestID, userID string) (*InvoiceResponse, error) {
	ir, err := s.repo.GetInvoiceRequestByID(ctx, requestID)
	if err != nil {
		return nil, err
	}

	if ir.Status != InvoiceRequestApproved {
		return nil, common.NewBadRequestError("invoice request must be approved before generating")
	}

	year := time.Now().Year()
	seq, err := s.repo.GetNextInvoiceSequence(ctx, year)
	if err != nil {
		return nil, common.NewInternalError("generate invoice number", err)
	}
	invoiceNumber := fmt.Sprintf("INV-%d-%04d", year, seq)

	invoiceAmount, err := s.getOrderAmount(ctx, ir.OrderType, ir.OrderID)
	if err != nil {
		s.logger.Error("invoice generation aborted: could not fetch order amount",
			zap.String("orderType", ir.OrderType),
			zap.String("orderID", ir.OrderID),
			zap.String("requestId", requestID),
			zap.Error(err),
		)
		return nil, common.NewInternalError("failed to fetch source order amount for invoice", err)
	}

	invoiceBody := fmt.Sprintf(
		"Invoice Number: %s\nDate: %s\nOrder Type: %s\nOrder ID: %s\nRequester: %s\n\n--- Line Items ---\n%s / %s: %.2f\n\n--- Total ---\nAmount Due: %.2f\n\nNotes: %s\n",
		invoiceNumber,
		time.Now().Format("2006-01-02"),
		ir.OrderType, ir.OrderID,
		ir.RequesterID,
		ir.OrderType, ir.OrderID, invoiceAmount,
		invoiceAmount,
		ir.Notes,
	)

	content, err := GeneratePDF("Invoice "+invoiceNumber, invoiceBody)
	if err != nil {
		return nil, common.NewInternalError("generate invoice PDF", err)
	}
	header := createMultipartHeader(invoiceNumber+".pdf", "application/pdf", len(content))

	fileMeta, err := s.fileService.Upload(
		ctx, userID,
		newBytesFile(content),
		header,
		"invoice", ir.ID,
		true,
	)
	if err != nil {
		return nil, common.NewInternalError("store invoice file", err)
	}

	inv := &Invoice{
		RequestID:     &ir.ID,
		InvoiceNumber: invoiceNumber,
		OrderType:     ir.OrderType,
		OrderID:       ir.OrderID,
		Amount:        invoiceAmount,
		FileID:        &fileMeta.ID,
	}

	if err := s.repo.CreateInvoice(ctx, inv); err != nil {
		return nil, common.NewInternalError("save invoice", err)
	}

	if err := s.repo.UpdateInvoiceRequestStatus(ctx, requestID, InvoiceRequestGenerated); err != nil {
		s.logger.Error("failed to update invoice request status", zap.String("requestId", requestID), zap.Error(err))
	}

	s.logger.Info("invoice generated",
		zap.String("invoiceId", inv.ID),
		zap.String("invoiceNumber", invoiceNumber),
		zap.String("requestId", requestID),
	)

	return &InvoiceResponse{
		ID:            inv.ID,
		RequestID:     inv.RequestID,
		InvoiceNumber: inv.InvoiceNumber,
		OrderType:     inv.OrderType,
		OrderID:       inv.OrderID,
		Amount:        inv.Amount,
		FileID:        inv.FileID,
		GeneratedAt:   inv.GeneratedAt,
	}, nil
}

func (s *ContractService) GetInvoiceRequests(ctx context.Context, userID string, isAdmin bool) ([]InvoiceRequest, error) {
	return s.repo.ListInvoiceRequests(ctx, userID, isAdmin)
}

func (s *ContractService) getOrderAmount(ctx context.Context, orderType, orderID string) (float64, error) {
	var amount float64
	var query string
	switch orderType {
	case "booking":
		query = `SELECT COALESCE(total_amount, 0) FROM bookings WHERE id = $1`
	case "procurement":
		query = `SELECT COALESCE(total_amount, 0) FROM purchase_orders WHERE id = $1`
	default:
		return 0, fmt.Errorf("unsupported order type for invoice amount lookup: %q", orderType)
	}
	err := s.repo.pool.QueryRow(ctx, query, orderID).Scan(&amount)
	if err != nil {
		return 0, fmt.Errorf("order amount lookup failed for %s/%s: %w", orderType, orderID, err)
	}
	return amount, nil
}
type bytesFile struct {
	*bytes.Reader
}

func newBytesFile(data []byte) multipart.File {
	return &bytesFile{Reader: bytes.NewReader(data)}
}

func (f *bytesFile) Close() error { return nil }

func createMultipartHeader(filename, contentType string, size int) *multipart.FileHeader {
	h := make(textproto.MIMEHeader)
	h.Set("Content-Type", contentType)
	return &multipart.FileHeader{
		Filename: filename,
		Header:   h,
		Size:     int64(size),
	}
}
