package users

type UserResponse struct {
	ID          string   `json:"id"`
	Email       string   `json:"email"`
	Status      string   `json:"status"`
	DisplayName string   `json:"display_name"`
	Roles       []string `json:"roles"`
}

type UpdateProfileRequest struct {
	DisplayName *string `json:"display_name,omitempty"`
	PhoneMasked *string `json:"phone_masked,omitempty"`
}

type UpdatePreferencesRequest struct {
	Preferences map[string]interface{} `json:"preferences"`
}
