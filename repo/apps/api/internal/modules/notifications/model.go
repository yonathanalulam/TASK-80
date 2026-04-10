package notifications

import (
	"encoding/json"
	"time"
)

type Message struct {
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

type NotificationEvent struct {
	ID          string          `json:"id"`
	EventType   string          `json:"eventType"`
	SourceType  string          `json:"sourceType"`
	SourceID    string          `json:"sourceId"`
	PayloadJSON json.RawMessage `json:"payloadJson"`
	CreatedAt   time.Time       `json:"createdAt"`
}

type NotificationRecipient struct {
	ID            string     `json:"id"`
	EventID       string     `json:"eventId"`
	UserID        string     `json:"userId"`
	Channel       string     `json:"channel"`
	Status        string     `json:"status"`
	DeliveredAt   *time.Time `json:"deliveredAt"`
	ReadAt        *time.Time `json:"readAt"`
	DeferredUntil *time.Time `json:"deferredUntil"`
	CreatedAt     time.Time  `json:"createdAt"`

	EventType   string          `json:"eventType,omitempty"`
	SourceType  string          `json:"sourceType,omitempty"`
	SourceID    string          `json:"sourceId,omitempty"`
	PayloadJSON json.RawMessage `json:"payloadJson,omitempty"`
}

type CallbackQueueEntry struct {
	ID              string          `json:"id"`
	EventID         *string         `json:"eventId"`
	RecipientID     *string         `json:"recipientId"`
	PayloadJSON     json.RawMessage `json:"payloadJson"`
	Status          string          `json:"status"`
	Attempts        int             `json:"attempts"`
	LastAttemptedAt *time.Time      `json:"lastAttemptedAt"`
	ExportedAt      *time.Time      `json:"exportedAt"`
	CreatedAt       time.Time       `json:"createdAt"`
}

type SendLog struct {
	ID                string          `json:"id"`
	RecipientUserID   string          `json:"recipientUserId"`
	MessageID         *string         `json:"messageId"`
	EventType         string          `json:"eventType"`
	ChannelType       string          `json:"channelType"`
	Status            string          `json:"status"`
	PayloadSummaryJSON json.RawMessage `json:"payloadSummaryJson"`
	CreatedAt         time.Time       `json:"createdAt"`
}

type MessageTemplate struct {
	ID              string    `json:"id"`
	TemplateKey     string    `json:"templateKey"`
	SubjectTemplate string    `json:"subjectTemplate"`
	BodyTemplate    string    `json:"bodyTemplate"`
	ChannelType     string    `json:"channelType"`
	Active          bool      `json:"active"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
}

type SubscriptionPreference struct {
	ID          string `json:"id"`
	UserID      string `json:"userId"`
	ChannelType string `json:"channelType"`
	EventType   string `json:"eventType"`
	Enabled     bool   `json:"enabled"`
}

type DNDSetting struct {
	UserID   string `json:"userId"`
	DNDStart string `json:"dndStart"`
	DNDEnd   string `json:"dndEnd"`
	Enabled  bool   `json:"enabled"`
}
