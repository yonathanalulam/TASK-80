package risk

import "time"

type RiskSummaryDTO struct {
	UserID          string           `json:"userId"`
	Score           float64          `json:"score"`
	IsBlacklisted   bool             `json:"isBlacklisted"`
	ActiveThrottles []ThrottleActionDTO `json:"activeThrottles"`
	RecentEvents    []RiskEventDTO   `json:"recentEvents"`
}

type ThrottleActionDTO struct {
	ID         string    `json:"id"`
	ActionType string    `json:"actionType"`
	Reason     string    `json:"reason"`
	ExpiresAt  time.Time `json:"expiresAt"`
	Active     bool      `json:"active"`
	CreatedAt  time.Time `json:"createdAt"`
}

type AdminApprovalDTO struct {
	ID              string     `json:"id"`
	UserID          *string    `json:"userId"`
	ActionType      string     `json:"actionType"`
	ReferenceType   string     `json:"referenceType"`
	ReferenceID     string     `json:"referenceId"`
	Status          string     `json:"status"`
	RequestedBy     string     `json:"requestedBy"`
	ResolvedBy      *string    `json:"resolvedBy"`
	ResolutionNotes string     `json:"resolutionNotes"`
	CreatedAt       time.Time  `json:"createdAt"`
	ResolvedAt      *time.Time `json:"resolvedAt"`
}

type RiskEventDTO struct {
	ID          string    `json:"id"`
	EventType   string    `json:"eventType"`
	Description string    `json:"description"`
	Severity    string    `json:"severity"`
	CreatedAt   time.Time `json:"createdAt"`
}

type ResolveApprovalRequest struct {
	Status string `json:"status"`
	Notes  string `json:"notes"`
}

type BlacklistRequest struct {
	Reason string `json:"reason"`
}
