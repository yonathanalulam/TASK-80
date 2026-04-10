package contracts

import (
	"encoding/json"
	"time"
)

type ContractTemplate struct {
	ID                 string          `json:"id"`
	Name               string          `json:"name"`
	BodyTemplate       string          `json:"bodyTemplate"`
	VariableSchemaJSON json.RawMessage `json:"variableSchemaJson"`
	Active             bool            `json:"active"`
	Version            int             `json:"version"`
	CreatedAt          time.Time       `json:"createdAt"`
	UpdatedAt          time.Time       `json:"updatedAt"`
}

type GeneratedContract struct {
	ID            string          `json:"id"`
	TemplateID    string          `json:"templateId"`
	VariablesJSON json.RawMessage `json:"variablesJson"`
	FileID        *string         `json:"fileId,omitempty"`
	GeneratedBy   string          `json:"generatedBy"`
	GeneratedAt   time.Time       `json:"generatedAt"`
	Version       int             `json:"version"`
}

type InvoiceRequestStatus string

const (
	InvoiceRequestPending   InvoiceRequestStatus = "pending"
	InvoiceRequestApproved  InvoiceRequestStatus = "approved"
	InvoiceRequestRejected  InvoiceRequestStatus = "rejected"
	InvoiceRequestGenerated InvoiceRequestStatus = "generated"
)

type InvoiceRequest struct {
	ID          string               `json:"id"`
	RequesterID string               `json:"requesterId"`
	OrderType   string               `json:"orderType"`
	OrderID     string               `json:"orderId"`
	Status      InvoiceRequestStatus `json:"status"`
	Notes       string               `json:"notes"`
	CreatedAt   time.Time            `json:"createdAt"`
	UpdatedAt   time.Time            `json:"updatedAt"`
}

type Invoice struct {
	ID            string    `json:"id"`
	RequestID     *string   `json:"requestId,omitempty"`
	InvoiceNumber string    `json:"invoiceNumber"`
	OrderType     string    `json:"orderType"`
	OrderID       string    `json:"orderId"`
	Amount        float64   `json:"amount"`
	FileID        *string   `json:"fileId,omitempty"`
	GeneratedAt   time.Time `json:"generatedAt"`
	CreatedAt     time.Time `json:"createdAt"`
}
