package admin

import "time"

type AuditLog struct {
	ID            string    `json:"id"`
	ActorID       *string   `json:"actor_id,omitempty"`
	Action        string    `json:"action"`
	EntityType    string    `json:"entity_type"`
	EntityID      *string   `json:"entity_id,omitempty"`
	BeforeSummary *string   `json:"before_summary,omitempty"`
	AfterSummary  *string   `json:"after_summary,omitempty"`
	RequestID     *string   `json:"request_id,omitempty"`
	IPAddress     *string   `json:"ip_address,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
}

type DomainEvent struct {
	ID            string    `json:"id"`
	EventType     string    `json:"event_type"`
	AggregateType string    `json:"aggregate_type"`
	AggregateID   string    `json:"aggregate_id"`
	PayloadJSON   *string   `json:"payload_json,omitempty"`
	ActorID       *string   `json:"actor_id,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
}
