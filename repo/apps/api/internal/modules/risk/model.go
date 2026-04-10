package risk

import (
	"encoding/json"
	"time"
)

type RiskEvent struct {
	ID           string          `json:"id"`
	UserID       string          `json:"userId"`
	EventType    string          `json:"eventType"`
	Description  string          `json:"description"`
	Severity     string          `json:"severity"`
	MetadataJSON json.RawMessage `json:"metadata"`
	CreatedAt    time.Time       `json:"createdAt"`
}

type RiskScore struct {
	ID          string          `json:"id"`
	UserID      string          `json:"userId"`
	Score       float64         `json:"score"`
	FactorsJSON json.RawMessage `json:"factors"`
	ComputedAt  time.Time       `json:"computedAt"`
}

type ThrottleAction struct {
	ID         string     `json:"id"`
	UserID     string     `json:"userId"`
	ActionType string     `json:"actionType"`
	Reason     string     `json:"reason"`
	ExpiresAt  *time.Time `json:"expiresAt"`
	Active     bool       `json:"active"`
	CreatedBy  string     `json:"createdBy"`
	CreatedAt  time.Time  `json:"createdAt"`
}

type AdminApproval struct {
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

type BlacklistRecord struct {
	ID            string     `json:"id"`
	UserID        string     `json:"userId"`
	Reason        string     `json:"reason"`
	BlacklistedBy string     `json:"blacklistedBy"`
	Active        bool       `json:"active"`
	CreatedAt     time.Time  `json:"createdAt"`
	LiftedAt      *time.Time `json:"liftedAt"`
}

type RiskDecision struct {
	Allowed         bool   `json:"allowed"`
	RequireApproval bool   `json:"requireApproval,omitempty"`
	Reason          string `json:"reason"`
}
