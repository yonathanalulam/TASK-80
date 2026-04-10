package contracts

import "time"
type GenerateContractRequest struct {
	TemplateID string            `json:"templateId"`
	Variables  map[string]string `json:"variables"`
}

type CreateInvoiceRequestDTO struct {
	OrderType string `json:"orderType"`
	OrderID   string `json:"orderId"`
	Notes     string `json:"notes"`
}
type ContractTemplateResponse struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Active  bool   `json:"active"`
	Version int    `json:"version"`
}

type GeneratedContractResponse struct {
	ID          string    `json:"id"`
	TemplateID  string    `json:"templateId"`
	FileID      *string   `json:"fileId,omitempty"`
	GeneratedBy string    `json:"generatedBy"`
	GeneratedAt time.Time `json:"generatedAt"`
	Version     int       `json:"version"`
}

type InvoiceResponse struct {
	ID            string    `json:"id"`
	RequestID     *string   `json:"requestId,omitempty"`
	InvoiceNumber string    `json:"invoiceNumber"`
	OrderType     string    `json:"orderType"`
	OrderID       string    `json:"orderId"`
	Amount        float64   `json:"amount"`
	FileID        *string   `json:"fileId,omitempty"`
	GeneratedAt   time.Time `json:"generatedAt"`
}

type InvoiceRequestResponse struct {
	ID          string               `json:"id"`
	RequesterID string               `json:"requesterId"`
	OrderType   string               `json:"orderType"`
	OrderID     string               `json:"orderId"`
	Status      InvoiceRequestStatus `json:"status"`
	Notes       string               `json:"notes"`
	CreatedAt   time.Time            `json:"createdAt"`
	UpdatedAt   time.Time            `json:"updatedAt"`
}
