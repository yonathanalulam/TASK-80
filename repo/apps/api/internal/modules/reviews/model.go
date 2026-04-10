package reviews

import "time"

type CreditTierName string

const (
	TierBronze   CreditTierName = "bronze"
	TierSilver   CreditTierName = "silver"
	TierGold     CreditTierName = "gold"
	TierPlatinum CreditTierName = "platinum"
)

type Review struct {
	ID             string    `json:"id"`
	ReviewerID     string    `json:"reviewerId"`
	SubjectID      string    `json:"subjectId"`
	OrderType      string    `json:"orderType"`
	OrderID        string    `json:"orderId"`
	OverallRating  float64   `json:"overallRating"`
	Comment        string    `json:"comment"`
	EditableUntil  time.Time `json:"editableUntil"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

type ReviewScore struct {
	ID          string  `json:"id"`
	ReviewID    string  `json:"reviewId"`
	DimensionID string  `json:"dimensionId"`
	Score       float64 `json:"score"`
}

type ReviewDimension struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Label  string `json:"label"`
	Active bool   `json:"active"`
}

type CreditTier struct {
	ID              string         `json:"id"`
	TierName        CreditTierName `json:"tierName"`
	MinTransactions int            `json:"minTransactions"`
	MinAvgRating    float64        `json:"minAvgRating"`
	MaxViolations   int            `json:"maxViolations"`
	Description     string         `json:"description"`
}

type UserCreditSnapshot struct {
	ID                string         `json:"id"`
	UserID            string         `json:"userId"`
	Tier              CreditTierName `json:"tier"`
	TotalTransactions int            `json:"totalTransactions"`
	AvgRating         float64        `json:"avgRating"`
	ViolationCount    int            `json:"violationCount"`
	ComputedAt        time.Time      `json:"computedAt"`
}

type ViolationRecord struct {
	ID            string    `json:"id"`
	UserID        string    `json:"userId"`
	ViolationType string    `json:"violationType"`
	Description   string    `json:"description"`
	Severity      string    `json:"severity"`
	RecordedBy    string    `json:"recordedBy"`
	CreatedAt     time.Time `json:"createdAt"`
}

type HarassmentFlag struct {
	ID             string     `json:"id"`
	ReporterID     string     `json:"reporterId"`
	SubjectID      string     `json:"subjectId"`
	Description    string     `json:"description"`
	EvidenceFileID *string    `json:"evidenceFileId,omitempty"`
	Status         string     `json:"status"`
	ReviewedBy     *string    `json:"reviewedBy,omitempty"`
	CreatedAt      time.Time  `json:"createdAt"`
	UpdatedAt      time.Time  `json:"updatedAt"`
}

type NoShowRecord struct {
	ID         string    `json:"id"`
	UserID     string    `json:"userId"`
	OrderType  string    `json:"orderType"`
	OrderID    string    `json:"orderId"`
	RecordedBy string    `json:"recordedBy"`
	CreatedAt  time.Time `json:"createdAt"`
}

type BlacklistRecord struct {
	ID            string     `json:"id"`
	UserID        string     `json:"userId"`
	Reason        string     `json:"reason"`
	BlacklistedBy string     `json:"blacklistedBy"`
	Active        bool       `json:"active"`
	CreatedAt     time.Time  `json:"createdAt"`
	LiftedAt      *time.Time `json:"liftedAt,omitempty"`
}
