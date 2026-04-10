package itineraries

import (
	"encoding/json"
	"time"
)
type CreateItineraryRequest struct {
	Title              string  `json:"title"`
	MeetupAt           *string `json:"meetupAt"`
	MeetupLocationText string  `json:"meetupLocationText"`
	Notes              string  `json:"notes"`
}

type UpdateItineraryRequest struct {
	Title              *string `json:"title"`
	MeetupAt           *string `json:"meetupAt"`
	MeetupLocationText *string `json:"meetupLocationText"`
	Notes              *string `json:"notes"`
}

type CreateCheckpointRequest struct {
	CheckpointText string  `json:"checkpointText"`
	SortOrder      int     `json:"sortOrder"`
	ETA            *string `json:"eta"`
}

type UpdateCheckpointRequest struct {
	CheckpointText *string `json:"checkpointText"`
	SortOrder      *int    `json:"sortOrder"`
	ETA            *string `json:"eta"`
}

type AddMemberRequest struct {
	UserID string `json:"userId"`
	Role   string `json:"role"`
}

type CreateFormDefinitionRequest struct {
	FieldKey   string          `json:"fieldKey"`
	FieldLabel string          `json:"fieldLabel"`
	FieldType  string          `json:"fieldType"`
	Required   bool            `json:"required"`
	Options    json.RawMessage `json:"options"`
	Validation json.RawMessage `json:"validation"`
	SortOrder  int             `json:"sortOrder"`
}

type UpdateFormDefinitionRequest struct {
	FieldLabel *string          `json:"fieldLabel"`
	FieldType  *string          `json:"fieldType"`
	Required   *bool            `json:"required"`
	Options    *json.RawMessage `json:"options"`
	Validation *json.RawMessage `json:"validation"`
	Active     *bool            `json:"active"`
	SortOrder  *int             `json:"sortOrder"`
}

type SubmitFormRequest struct {
	Payload map[string]interface{} `json:"payload"`
}
type ItineraryResponse struct {
	ID                 string          `json:"id"`
	OrganizerID        string          `json:"organizerId"`
	Title              string          `json:"title"`
	MeetupAt           *time.Time      `json:"meetupAt"`
	MeetupLocationText string          `json:"meetupLocationText"`
	Notes              string          `json:"notes"`
	Status             ItineraryStatus `json:"status"`
	PublishedAt        *time.Time      `json:"publishedAt"`
	CheckpointsCount   int             `json:"checkpointsCount"`
	MembersCount       int             `json:"membersCount"`
	CreatedAt          time.Time       `json:"createdAt"`
	UpdatedAt          time.Time       `json:"updatedAt"`
}

type ItineraryDetailResponse struct {
	ItineraryResponse
	Checkpoints     []Checkpoint     `json:"checkpoints"`
	Members         []Member         `json:"members"`
	FormDefinitions []FormDefinition `json:"formDefinitions"`
}

type ChangeEventResponse struct {
	ID         string          `json:"id"`
	ActorID    string          `json:"actorId"`
	ChangeType string          `json:"changeType"`
	Summary    string          `json:"summary"`
	Diff       json.RawMessage `json:"diff"`
	CreatedAt  time.Time       `json:"createdAt"`
}

type PaginatedResponse struct {
	Items      interface{} `json:"items"`
	Total      int         `json:"total"`
	Page       int         `json:"page"`
	PageSize   int         `json:"pageSize"`
	TotalPages int         `json:"totalPages"`
}
