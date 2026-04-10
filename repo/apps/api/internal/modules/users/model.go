package users

import "time"

type User struct {
	ID        string     `json:"id"`
	Email     string     `json:"email"`
	Status    string     `json:"status"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}

type UserProfile struct {
	UserID      string  `json:"user_id"`
	DisplayName string  `json:"display_name"`
	PhoneMasked *string `json:"phone_masked,omitempty"`
}

type UserPreference struct {
	UserID          string `json:"user_id"`
	PreferencesJSON string `json:"preferences_json"`
}
