package finance

import "time"
type WalletResponse struct {
	ID         string     `json:"id"`
	OwnerID    *string    `json:"ownerId"`
	WalletType WalletType `json:"walletType"`
	Balance    float64    `json:"balance"`
	Currency   string     `json:"currency"`
	CreatedAt  time.Time  `json:"createdAt"`
	UpdatedAt  time.Time  `json:"updatedAt"`
}

type TransactionResponse struct {
	ID            string    `json:"id"`
	WalletID      string    `json:"walletId"`
	Amount        float64   `json:"amount"`
	Direction     Direction `json:"direction"`
	ReferenceType string    `json:"referenceType"`
	ReferenceID   string    `json:"referenceId"`
	Description   string    `json:"description"`
	CreatedAt     time.Time `json:"createdAt"`
}

type PaginatedTransactions struct {
	Items      []TransactionResponse `json:"items"`
	Total      int                   `json:"total"`
	Page       int                   `json:"page"`
	PageSize   int                   `json:"pageSize"`
	TotalPages int                   `json:"totalPages"`
}
type RecordTenderRequest struct {
	OrderType     string     `json:"orderType"`
	OrderID       string     `json:"orderId"`
	TenderType    TenderType `json:"tenderType"`
	Amount        float64    `json:"amount"`
	Currency      string     `json:"currency"`
	ReferenceText string     `json:"referenceText"`
}

type RecordTenderResponse struct {
	ID         string    `json:"id"`
	RecordedAt time.Time `json:"recordedAt"`
}
type RefundItemRequest struct {
	ItemID   string  `json:"itemId"`
	ItemType string  `json:"itemType"`
	Amount   float64 `json:"amount"`
}

type RefundRequest struct {
	OrderType string            `json:"orderType"`
	OrderID   string            `json:"orderId"`
	Amount    float64           `json:"amount"`
	Reason    string            `json:"reason"`
	Items     []RefundItemRequest `json:"items"`
}
type WithdrawalCreateRequest struct {
	Amount float64 `json:"amount"`
}

type WithdrawalRequestDTO struct {
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
}

type WithdrawalRejectRequest struct {
	Reason string `json:"reason"`
}
type ReconciliationReportDTO struct {
	OpeningBalance    float64                `json:"openingBalance"`
	Inflows           float64                `json:"inflows"`
	Outflows          float64                `json:"outflows"`
	HeldInEscrow      float64                `json:"heldInEscrow"`
	Released          float64                `json:"released"`
	Refunded          float64                `json:"refunded"`
	NetPayable        float64                `json:"netPayable"`
	UnreconciledItems []ReconciliationItem   `json:"unreconciledItems"`
}

type ReleaseEscrowRequest struct {
	OrderType string  `json:"orderType"`
	Amount    float64 `json:"amount"`
}

type EscrowSummary struct {
	ID             string  `json:"id"`
	OrderType      string  `json:"orderType"`
	OrderID        string  `json:"orderId"`
	AmountHeld     float64 `json:"amountHeld"`
	AmountReleased float64 `json:"amountReleased"`
	Status         string  `json:"status"`
	CreatedAt      string  `json:"createdAt"`
}
