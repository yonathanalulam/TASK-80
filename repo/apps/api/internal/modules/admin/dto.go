package admin

type AuditLogResponse struct {
	ID         string  `json:"id"`
	ActorID    *string `json:"actor_id,omitempty"`
	Action     string  `json:"action"`
	EntityType string  `json:"entity_type"`
	EntityID   *string `json:"entity_id,omitempty"`
	RequestID  *string `json:"request_id,omitempty"`
	CreatedAt  string  `json:"created_at"`
}

type AuditLogFilters struct {
	ActorID    string
	EntityType string
	Action     string
	Page       int
	PageSize   int
}
