package itineraries

import (
	"encoding/json"
	"time"
)

type ItineraryStatus string

const (
	StatusDraft      ItineraryStatus = "draft"
	StatusPublished  ItineraryStatus = "published"
	StatusRevised    ItineraryStatus = "revised"
	StatusInProgress ItineraryStatus = "in_progress"
	StatusCompleted  ItineraryStatus = "completed"
	StatusCancelled  ItineraryStatus = "cancelled"
	StatusArchived   ItineraryStatus = "archived"
)

type Itinerary struct {
	ID                 string          `json:"id"`
	OrganizerID        string          `json:"organizerId"`
	Title              string          `json:"title"`
	MeetupAt           *time.Time      `json:"meetupAt"`
	MeetupLocationText string          `json:"meetupLocationText"`
	Notes              string          `json:"notes"`
	Status             ItineraryStatus `json:"status"`
	PublishedAt        *time.Time      `json:"publishedAt"`
	CreatedAt          time.Time       `json:"createdAt"`
	UpdatedAt          time.Time       `json:"updatedAt"`
}

type Checkpoint struct {
	ID             string     `json:"id"`
	ItineraryID    string     `json:"itineraryId"`
	SortOrder      int        `json:"sortOrder"`
	CheckpointText string     `json:"checkpointText"`
	ETA            *time.Time `json:"eta"`
	CreatedAt      time.Time  `json:"createdAt"`
	UpdatedAt      time.Time  `json:"updatedAt"`
}

type Member struct {
	ID          string    `json:"id"`
	ItineraryID string    `json:"itineraryId"`
	UserID      string    `json:"userId"`
	Role        string    `json:"role"`
	JoinedAt    time.Time `json:"joinedAt"`
}

type FormDefinition struct {
	ID             string          `json:"id"`
	ItineraryID    string          `json:"itineraryId"`
	FieldKey       string          `json:"fieldKey"`
	FieldLabel     string          `json:"fieldLabel"`
	FieldType      string          `json:"fieldType"`
	Required       bool            `json:"required"`
	OptionsJSON    json.RawMessage `json:"options"`
	ValidationJSON json.RawMessage `json:"validation"`
	Active         bool            `json:"active"`
	SortOrder      int             `json:"sortOrder"`
	CreatedAt      time.Time       `json:"createdAt"`
	UpdatedAt      time.Time       `json:"updatedAt"`
}

type FormSubmission struct {
	ID           string          `json:"id"`
	ItineraryID  string          `json:"itineraryId"`
	MemberUserID string          `json:"memberUserId"`
	PayloadJSON  json.RawMessage `json:"payload"`
	SubmittedAt  time.Time       `json:"submittedAt"`
	UpdatedAt    time.Time       `json:"updatedAt"`
}

type ChangeEvent struct {
	ID          string          `json:"id"`
	ItineraryID string          `json:"itineraryId"`
	ActorID     string          `json:"actorId"`
	ChangeType  string          `json:"changeType"`
	Summary     string          `json:"summary"`
	DiffJSON    json.RawMessage `json:"diff"`
	VisibleFrom time.Time       `json:"visibleFrom"`
	CreatedAt   time.Time       `json:"createdAt"`
}
