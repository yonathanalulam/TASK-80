package notifications

import (
	"encoding/json"
	"time"
)
type NotificationDTO struct {
	ID          string          `json:"id"`
	EventType   string          `json:"eventType"`
	SourceType  string          `json:"sourceType"`
	SourceID    string          `json:"sourceId"`
	Channel     string          `json:"channel"`
	Status      string          `json:"status"`
	Payload     json.RawMessage `json:"payload"`
	DeliveredAt *time.Time      `json:"deliveredAt"`
	ReadAt      *time.Time      `json:"readAt"`
	CreatedAt   time.Time       `json:"createdAt"`
}

type MessageDTO struct {
	ID           string          `json:"id"`
	SenderID     *string         `json:"senderId"`
	RecipientID  string          `json:"recipientId"`
	Subject      string          `json:"subject"`
	Body         string          `json:"body"`
	TemplateID   *string         `json:"templateId"`
	MetadataJSON json.RawMessage `json:"metadataJson"`
	ReadAt       *time.Time      `json:"readAt"`
	CreatedAt    time.Time       `json:"createdAt"`
}

type SendLogDTO struct {
	ID                string          `json:"id"`
	RecipientUserID   string          `json:"recipientUserId"`
	MessageID         *string         `json:"messageId"`
	EventType         string          `json:"eventType"`
	ChannelType       string          `json:"channelType"`
	Status            string          `json:"status"`
	PayloadSummaryJSON json.RawMessage `json:"payloadSummaryJson"`
	CreatedAt         time.Time       `json:"createdAt"`
}

type PaginatedResponse struct {
	Items      interface{} `json:"items"`
	Total      int         `json:"total"`
	Page       int         `json:"page"`
	PageSize   int         `json:"pageSize"`
	TotalPages int         `json:"totalPages"`
}
type UpdateSubscriptionRequest struct {
	ChannelType string `json:"channelType"`
	EventType   string `json:"eventType"`
	Enabled     bool   `json:"enabled"`
}

type UpdateDNDRequest struct {
	DNDStart string `json:"dndStart"`
	DNDEnd   string `json:"dndEnd"`
	Enabled  *bool  `json:"enabled"`
}

type CallbackExportRequest struct {
	Status string `json:"status"`
	Limit  int    `json:"limit"`
}

type CallbackExportResponse struct {
	Entries    []CallbackQueueEntry `json:"entries"`
	Count      int                  `json:"count"`
	ExportedAt time.Time            `json:"exportedAt"`
}
