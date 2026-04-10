package finance

import (
	"encoding/json"
	"time"
)

type WalletType string

const (
	WalletTypeCustomer WalletType = "customer"
	WalletTypeCourier  WalletType = "courier"
	WalletTypeSupplier WalletType = "supplier"
	WalletTypeSystem   WalletType = "system"
)

type EscrowStatus string

const (
	EscrowStatusHeld     EscrowStatus = "held"
	EscrowStatusPartial  EscrowStatus = "partially_released"
	EscrowStatusReleased EscrowStatus = "released"
	EscrowStatusRefunded EscrowStatus = "refunded"
)

type TenderType string

const (
	TenderTypeCash     TenderType = "cash"
	TenderTypeBank     TenderType = "bank_transfer"
	TenderTypeMobile   TenderType = "mobile_money"
	TenderTypeInternal TenderType = "internal"
)

type RefundStatus string

const (
	RefundStatusPending  RefundStatus = "pending"
	RefundStatusApproved RefundStatus = "approved"
	RefundStatusRejected RefundStatus = "rejected"
)

type WithdrawalStatus string

const (
	WithdrawalStatusRequested WithdrawalStatus = "requested"
	WithdrawalStatusApproved  WithdrawalStatus = "approved"
	WithdrawalStatusRejected  WithdrawalStatus = "rejected"
	WithdrawalStatusSettled   WithdrawalStatus = "settled"
)

type ReconciliationStatus string

const (
	ReconciliationStatusPending   ReconciliationStatus = "pending"
	ReconciliationStatusCompleted ReconciliationStatus = "completed"
)

type Wallet struct {
	ID         string     `json:"id"`
	OwnerID    *string    `json:"ownerId"`
	WalletType WalletType `json:"walletType"`
	Balance    float64    `json:"balance"`
	Currency   string     `json:"currency"`
	CreatedAt  time.Time  `json:"createdAt"`
	UpdatedAt  time.Time  `json:"updatedAt"`
}

type WalletTransaction struct {
	ID            string    `json:"id"`
	WalletID      string    `json:"walletId"`
	Amount        float64   `json:"amount"`
	Direction     Direction `json:"direction"`
	ReferenceType string    `json:"referenceType"`
	ReferenceID   string    `json:"referenceId"`
	Description   string    `json:"description"`
	CreatedAt     time.Time `json:"createdAt"`
}

type EscrowAccount struct {
	ID             string       `json:"id"`
	OrderType      string       `json:"orderType"`
	OrderID        string       `json:"orderId"`
	AmountHeld     float64      `json:"amountHeld"`
	AmountReleased float64      `json:"amountReleased"`
	AmountRefunded float64      `json:"amountRefunded"`
	Status         EscrowStatus `json:"status"`
	CreatedAt      time.Time    `json:"createdAt"`
	UpdatedAt      time.Time    `json:"updatedAt"`
}

type PaymentRecord struct {
	ID            string     `json:"id"`
	OrderType     string     `json:"orderType"`
	OrderID       string     `json:"orderId"`
	TenderType    TenderType `json:"tenderType"`
	Amount        float64    `json:"amount"`
	Currency      string     `json:"currency"`
	ReferenceText string     `json:"referenceText"`
	RecordedBy    string     `json:"recordedBy"`
	RecordedAt    time.Time  `json:"recordedAt"`
}

type Refund struct {
	ID           string       `json:"id"`
	OrderType    string       `json:"orderType"`
	OrderID      string       `json:"orderId"`
	RefundAmount float64      `json:"refundAmount"`
	RefundReason string       `json:"refundReason"`
	CreatedBy    string       `json:"createdBy"`
	ApprovedBy   *string      `json:"approvedBy"`
	Status       RefundStatus `json:"status"`
	CreatedAt    time.Time    `json:"createdAt"`
	UpdatedAt    time.Time    `json:"updatedAt"`
}

type RefundItem struct {
	ID        string    `json:"id"`
	RefundID  string    `json:"refundId"`
	ItemID    string    `json:"itemId"`
	ItemType  string    `json:"itemType"`
	Amount    float64   `json:"amount"`
	CreatedAt time.Time `json:"createdAt"`
}

type WithdrawalRequest struct {
	ID             string           `json:"id"`
	CourierID      string           `json:"courierId"`
	RequestAmount  float64          `json:"requestAmount"`
	Status         WithdrawalStatus `json:"status"`
	RequestedAt    time.Time        `json:"requestedAt"`
	ReviewedBy     *string          `json:"reviewedBy"`
	ApprovedBy     *string          `json:"approvedBy"`
	RejectedReason *string          `json:"rejectedReason"`
	SettledAt      *time.Time       `json:"settledAt"`
	CreatedAt      time.Time        `json:"createdAt"`
	UpdatedAt      time.Time        `json:"updatedAt"`
}

type WithdrawalDisbursement struct {
	ID           string    `json:"id"`
	WithdrawalID string    `json:"withdrawalId"`
	Amount       float64   `json:"amount"`
	DisbursedAt  time.Time `json:"disbursedAt"`
}

type ReconciliationRun struct {
	ID          string               `json:"id"`
	RunDate     time.Time            `json:"runDate"`
	Status      ReconciliationStatus `json:"status"`
	SummaryJSON json.RawMessage      `json:"summaryJson"`
	CreatedBy   string               `json:"createdBy"`
	CreatedAt   time.Time            `json:"createdAt"`
}

type ReconciliationItem struct {
	ID             string  `json:"id"`
	RunID          string  `json:"runId"`
	ItemType       string  `json:"itemType"`
	ReferenceID    string  `json:"referenceId"`
	ExpectedAmount float64 `json:"expectedAmount"`
	ActualAmount   float64 `json:"actualAmount"`
	Difference     float64 `json:"difference"`
	Status         string  `json:"status"`
	Notes          string  `json:"notes"`
}
