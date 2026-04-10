package reviews

import "time"
type DimensionScore struct {
	DimensionName string  `json:"dimensionName"`
	Score         float64 `json:"score"`
}

type CreateReviewRequest struct {
	SubjectID     string           `json:"subjectId"`
	OrderType     string           `json:"orderType"`
	OrderID       string           `json:"orderId"`
	OverallRating float64          `json:"overallRating"`
	Comment       string           `json:"comment"`
	Scores        []DimensionScore `json:"scores"`
}

type CreateViolationRequest struct {
	UserID        string `json:"userId"`
	ViolationType string `json:"violationType"`
	Description   string `json:"description"`
	Severity      string `json:"severity"`
}

type CreateNoShowRequest struct {
	UserID    string `json:"userId"`
	OrderType string `json:"orderType"`
	OrderID   string `json:"orderId"`
}

type CreateHarassmentFlagRequest struct {
	SubjectID      string  `json:"subjectId"`
	Description    string  `json:"description"`
	EvidenceFileID *string `json:"evidenceFileId,omitempty"`
}
type ReviewResponse struct {
	ID            string         `json:"id"`
	ReviewerID    string         `json:"reviewerId"`
	SubjectID     string         `json:"subjectId"`
	OrderType     string         `json:"orderType"`
	OrderID       string         `json:"orderId"`
	OverallRating float64        `json:"overallRating"`
	Comment       string         `json:"comment"`
	Scores        []ScoreDetail  `json:"scores,omitempty"`
	EditableUntil time.Time      `json:"editableUntil"`
	CreatedAt     time.Time      `json:"createdAt"`
}

type ScoreDetail struct {
	DimensionName string  `json:"dimensionName"`
	Score         float64 `json:"score"`
}

type CreditSnapshotResponse struct {
	UserID            string         `json:"userId"`
	Tier              CreditTierName `json:"tier"`
	TotalTransactions int            `json:"totalTransactions"`
	AvgRating         float64        `json:"avgRating"`
	ViolationCount    int            `json:"violationCount"`
	ComputedAt        time.Time      `json:"computedAt"`
}

type RiskSummaryResponse struct {
	UserID          string                  `json:"userId"`
	CreditSnapshot  *CreditSnapshotResponse `json:"creditSnapshot,omitempty"`
	ViolationCount  int                     `json:"violationCount"`
	NoShowCount     int                     `json:"noShowCount"`
	HarassmentFlags int                     `json:"harassmentFlags"`
	IsBlacklisted   bool                    `json:"isBlacklisted"`
}
