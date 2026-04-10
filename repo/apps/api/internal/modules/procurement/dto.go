package procurement

import "time"
type CreateRFQRequest struct {
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Deadline    string    `json:"deadline"`
}

type RFQItemRequest struct {
	ItemName       string `json:"itemName"`
	Specifications string `json:"specifications"`
	Quantity       int    `json:"quantity"`
	Unit           string `json:"unit"`
	SortOrder      int    `json:"sortOrder"`
}

type IssueRFQRequest struct {
	SupplierIDs []string       `json:"supplierIds"`
	Items       []RFQItemRequest `json:"items"`
}

type RFQResponse struct {
	ID          string     `json:"id"`
	CreatedBy   string     `json:"createdBy"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Deadline    time.Time  `json:"deadline"`
	Status      RFQStatus  `json:"status"`
	Items       []RFQItem  `json:"items"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
}
type QuoteItemRequest struct {
	RFQItemID string  `json:"rfqItemId"`
	UnitPrice float64 `json:"unitPrice"`
	Quantity  int     `json:"quantity"`
	Notes     string  `json:"notes"`
}

type SubmitQuoteRequest struct {
	TotalAmount  float64          `json:"totalAmount"`
	LeadTimeDays int              `json:"leadTimeDays"`
	Notes        string           `json:"notes"`
	Items        []QuoteItemRequest `json:"items"`
}

type QuoteResponse struct {
	ID           string         `json:"id"`
	RFQID        string         `json:"rfqId"`
	SupplierID   string         `json:"supplierId"`
	TotalAmount  float64        `json:"totalAmount"`
	LeadTimeDays int            `json:"leadTimeDays"`
	Notes        string         `json:"notes"`
	Items        []RFQQuoteItem `json:"items"`
	SubmittedAt  time.Time      `json:"submittedAt"`
}

type ComparisonMatrixResponse struct {
	RFQID   string                  `json:"rfqId"`
	Items   []RFQItem               `json:"items"`
	Quotes  []ComparisonMatrixEntry `json:"quotes"`
}

type SelectQuoteRequest struct {
	QuoteID string `json:"quoteId"`
}
type CreatePORequest struct {
	RFQID       *string       `json:"rfqId"`
	QuoteID     *string       `json:"quoteId"`
	SupplierID  string        `json:"supplierId"`
	PromisedDate *string      `json:"promisedDate"`
	Items       []POItemRequest `json:"items"`
}

type POItemRequest struct {
	ItemName       string  `json:"itemName"`
	Specifications string  `json:"specifications"`
	UnitPrice      float64 `json:"unitPrice"`
	Quantity       int     `json:"quantity"`
}

type POResponse struct {
	ID           string     `json:"id"`
	RFQID        *string    `json:"rfqId"`
	QuoteID      *string    `json:"quoteId"`
	SupplierID   string     `json:"supplierId"`
	CreatedBy    string     `json:"createdBy"`
	PONumber     string     `json:"poNumber"`
	PromisedDate *time.Time `json:"promisedDate"`
	Status       POStatus   `json:"status"`
	TotalAmount  float64    `json:"totalAmount"`
	Items        []POItem   `json:"items"`
	CreatedAt    time.Time  `json:"createdAt"`
	UpdatedAt    time.Time  `json:"updatedAt"`
}
type DeliveryItemRequest struct {
	POItemID          string `json:"poItemId"`
	QuantityDelivered int    `json:"quantityDelivered"`
	QuantityAccepted  int    `json:"quantityAccepted"`
	QuantityRejected  int    `json:"quantityRejected"`
}

type RecordDeliveryRequest struct {
	CourierID    *string             `json:"courierId"`
	DeliveryDate string              `json:"deliveryDate"`
	Notes        string              `json:"notes"`
	Items        []DeliveryItemRequest `json:"items"`
}

type DeliveryResponse struct {
	ID           string         `json:"id"`
	POID         string         `json:"poId"`
	CourierID    *string        `json:"courierId"`
	ReceivedBy   *string        `json:"receivedBy"`
	DeliveryDate time.Time      `json:"deliveryDate"`
	Notes        string         `json:"notes"`
	Status       string         `json:"status"`
	Items        []DeliveryItem `json:"items"`
	CreatedAt    time.Time      `json:"createdAt"`
}
type InspectionRequest struct {
	Status InspectionStatus `json:"status"`
	Notes  string           `json:"notes"`
}

type InspectionResponse struct {
	ID          string           `json:"id"`
	DeliveryID  *string          `json:"deliveryId"`
	POID        string           `json:"poId"`
	InspectorID string           `json:"inspectorId"`
	Status      InspectionStatus `json:"status"`
	Notes       string           `json:"notes"`
	InspectedAt time.Time        `json:"inspectedAt"`
}
type CreateDiscrepancyRequest struct {
	POID            string          `json:"poId"`
	DeliveryID      *string         `json:"deliveryId"`
	InspectionID    *string         `json:"inspectionId"`
	DiscrepancyType DiscrepancyType `json:"discrepancyType"`
	Description     string          `json:"description"`
	Notes           string          `json:"notes"`
}
type WaiverRequest struct {
	WaiverReason string `json:"waiverReason"`
}

type SettlementAdjustmentRequest struct {
	Amount    float64 `json:"amount"`
	Direction string  `json:"direction"`
	Reason    string  `json:"reason"`
}

type ExceptionCaseResponse struct {
	ID            string              `json:"id"`
	ReferenceType string              `json:"referenceType"`
	ReferenceID   string              `json:"referenceId"`
	Status        ExceptionStatus     `json:"status"`
	OpenedReason  string              `json:"openedReason"`
	OpenedAt      time.Time           `json:"openedAt"`
	ClosedAt      *time.Time          `json:"closedAt"`
	Waivers       []WaiverRecord      `json:"waivers"`
	Adjustments   []SettlementAdjustment `json:"adjustments"`
	CreatedAt     time.Time           `json:"createdAt"`
}
