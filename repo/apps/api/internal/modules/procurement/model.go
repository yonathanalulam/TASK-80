package procurement

import (
	"encoding/json"
	"time"
)

type RFQStatus string

const (
	RFQStatusDraft           RFQStatus = "draft"
	RFQStatusIssued          RFQStatus = "issued"
	RFQStatusResponded       RFQStatus = "responded"
	RFQStatusComparisonReady RFQStatus = "comparison_ready"
	RFQStatusSelected        RFQStatus = "selected"
	RFQStatusClosedNoAward   RFQStatus = "closed_no_award"
	RFQStatusConvertedPO     RFQStatus = "converted_to_po"
)

type POStatus string

const (
	POStatusDraft              POStatus = "draft"
	POStatusIssued             POStatus = "issued"
	POStatusAccepted           POStatus = "accepted"
	POStatusPartiallyDelivered POStatus = "partially_delivered"
	POStatusDelivered          POStatus = "delivered"
	POStatusInspectionPending  POStatus = "inspection_pending"
	POStatusExceptionOpen      POStatus = "exception_open"
	POStatusClosed             POStatus = "closed"
)

type InspectionStatus string

const (
	InspectionStatusPending InspectionStatus = "pending"
	InspectionStatusPassed  InspectionStatus = "passed"
	InspectionStatusFailed  InspectionStatus = "failed"
)

type DiscrepancyType string

const (
	DiscrepancyTypeShortage         DiscrepancyType = "shortage"
	DiscrepancyTypeDamage           DiscrepancyType = "damage"
	DiscrepancyTypeWrongItem        DiscrepancyType = "wrong_item"
	DiscrepancyTypeLateDelivery     DiscrepancyType = "late_delivery"
	DiscrepancyTypeServiceDeviation DiscrepancyType = "service_deviation"
	DiscrepancyTypeOther            DiscrepancyType = "other"
)

type ExceptionStatus string

const (
	ExceptionStatusOpen   ExceptionStatus = "open"
	ExceptionStatusClosed ExceptionStatus = "closed"
)

type RFQ struct {
	ID          string    `json:"id"`
	CreatedBy   string    `json:"createdBy"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Deadline    time.Time `json:"deadline"`
	Status      RFQStatus `json:"status"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type RFQItem struct {
	ID             string    `json:"id"`
	RFQID          string    `json:"rfqId"`
	ItemName       string    `json:"itemName"`
	Specifications string    `json:"specifications"`
	Quantity       int       `json:"quantity"`
	Unit           string    `json:"unit"`
	SortOrder      int       `json:"sortOrder"`
	CreatedAt      time.Time `json:"createdAt"`
}

type RFQSupplier struct {
	ID         string    `json:"id"`
	RFQID      string    `json:"rfqId"`
	SupplierID string    `json:"supplierId"`
	InvitedAt  time.Time `json:"invitedAt"`
}

type RFQQuote struct {
	ID          string    `json:"id"`
	RFQID       string    `json:"rfqId"`
	SupplierID  string    `json:"supplierId"`
	TotalAmount float64   `json:"totalAmount"`
	LeadTimeDays int      `json:"leadTimeDays"`
	Notes       string    `json:"notes"`
	SubmittedAt time.Time `json:"submittedAt"`
	CreatedAt   time.Time `json:"createdAt"`
}

type RFQQuoteItem struct {
	ID        string  `json:"id"`
	QuoteID   string  `json:"quoteId"`
	RFQItemID string  `json:"rfqItemId"`
	UnitPrice float64 `json:"unitPrice"`
	Quantity  int     `json:"quantity"`
	Subtotal  float64 `json:"subtotal"`
	Notes     string  `json:"notes"`
}

type PurchaseOrder struct {
	ID           string    `json:"id"`
	RFQID        *string   `json:"rfqId"`
	QuoteID      *string   `json:"quoteId"`
	SupplierID   string    `json:"supplierId"`
	CreatedBy    string    `json:"createdBy"`
	PONumber     string    `json:"poNumber"`
	PromisedDate *time.Time `json:"promisedDate"`
	Status       POStatus  `json:"status"`
	TotalAmount  float64   `json:"totalAmount"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

type POItem struct {
	ID             string    `json:"id"`
	POID           string    `json:"poId"`
	ItemName       string    `json:"itemName"`
	Specifications string    `json:"specifications"`
	UnitPrice      float64   `json:"unitPrice"`
	Quantity       int       `json:"quantity"`
	Subtotal       float64   `json:"subtotal"`
	CreatedAt      time.Time `json:"createdAt"`
}

type Delivery struct {
	ID           string     `json:"id"`
	POID         string     `json:"poId"`
	CourierID    *string    `json:"courierId"`
	ReceivedBy   *string    `json:"receivedBy"`
	DeliveryDate time.Time  `json:"deliveryDate"`
	Notes        string     `json:"notes"`
	Status       string     `json:"status"`
	CreatedAt    time.Time  `json:"createdAt"`
	UpdatedAt    time.Time  `json:"updatedAt"`
}

type DeliveryItem struct {
	ID                string `json:"id"`
	DeliveryID        string `json:"deliveryId"`
	POItemID          string `json:"poItemId"`
	QuantityDelivered int    `json:"quantityDelivered"`
	QuantityAccepted  int    `json:"quantityAccepted"`
	QuantityRejected  int    `json:"quantityRejected"`
	CreatedAt         time.Time `json:"createdAt"`
}

type QualityInspection struct {
	ID          string           `json:"id"`
	DeliveryID  *string          `json:"deliveryId"`
	POID        string           `json:"poId"`
	InspectorID string           `json:"inspectorId"`
	Status      InspectionStatus `json:"status"`
	Notes       string           `json:"notes"`
	InspectedAt time.Time        `json:"inspectedAt"`
	CreatedAt   time.Time        `json:"createdAt"`
	UpdatedAt   time.Time        `json:"updatedAt"`
}

type DiscrepancyTicket struct {
	ID              string          `json:"id"`
	POID            string          `json:"poId"`
	DeliveryID      *string         `json:"deliveryId"`
	InspectionID    *string         `json:"inspectionId"`
	DiscrepancyType DiscrepancyType `json:"discrepancyType"`
	Description     string          `json:"description"`
	Notes           string          `json:"notes"`
	Status          string          `json:"status"`
	CreatedBy       string          `json:"createdBy"`
	CreatedAt       time.Time       `json:"createdAt"`
	UpdatedAt       time.Time       `json:"updatedAt"`
}

type ExceptionCase struct {
	ID            string          `json:"id"`
	ReferenceType string          `json:"referenceType"`
	ReferenceID   string          `json:"referenceId"`
	Status        ExceptionStatus `json:"status"`
	OpenedReason  string          `json:"openedReason"`
	OpenedAt      time.Time       `json:"openedAt"`
	ClosedAt      *time.Time      `json:"closedAt"`
	CreatedAt     time.Time       `json:"createdAt"`
	UpdatedAt     time.Time       `json:"updatedAt"`
}

type WaiverRecord struct {
	ID              string    `json:"id"`
	ExceptionCaseID string    `json:"exceptionCaseId"`
	ApprovedBy      string    `json:"approvedBy"`
	WaiverReason    string    `json:"waiverReason"`
	CreatedAt       time.Time `json:"createdAt"`
}

type SettlementAdjustment struct {
	ID              string    `json:"id"`
	ExceptionCaseID string    `json:"exceptionCaseId"`
	Amount          float64   `json:"amount"`
	Direction       string    `json:"direction"`
	Reason          string    `json:"reason"`
	ApprovedBy      string    `json:"approvedBy"`
	JournalEntryID  *string   `json:"journalEntryId"`
	CreatedAt       time.Time `json:"createdAt"`
}

type ComparisonMatrixEntry struct {
	SupplierID   string            `json:"supplierId"`
	TotalAmount  float64           `json:"totalAmount"`
	LeadTimeDays int               `json:"leadTimeDays"`
	Notes        string            `json:"notes"`
	Items        []RFQQuoteItem    `json:"items"`
	SubmittedAt  time.Time         `json:"submittedAt"`
	Metadata     json.RawMessage   `json:"metadata,omitempty"`
}
