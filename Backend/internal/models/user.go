package models

// PasswordMetadata mirrors the passwordMetadata field stored in persons.json.
type PasswordMetadata struct {
	CaseSensitive    bool `json:"caseSensitive"`
	RequiresUpperCase bool `json:"requiresUpperCase"`
	RequiresLowerCase bool `json:"requiresLowerCase"`
	RequiresNumber   bool `json:"requiresNumber"`
	MinimumLength    int  `json:"minimumLength"`
}

// User is the full user record stored in persons.json (including secrets).
type User struct {
	ID               string           `json:"id"`
	Username         string           `json:"username"`
	Email            string           `json:"email"`
	PasswordHash     string           `json:"passwordHash"`
	Salt             string           `json:"salt"`
	PasswordMetadata PasswordMetadata `json:"passwordMetadata"`
	CreatedAt        string           `json:"createdAt"`
}

// PublicUser is the safe subset returned in API responses (no hash/salt).
type PublicUser struct {
	ID               string           `json:"id"`
	Username         string           `json:"username"`
	Email            string           `json:"email"`
	PasswordMetadata PasswordMetadata `json:"passwordMetadata"`
	CreatedAt        string           `json:"createdAt"`
}

// ToPublic strips sensitive fields from a User.
func (u User) ToPublic() PublicUser {
	return PublicUser{
		ID:               u.ID,
		Username:         u.Username,
		Email:            u.Email,
		PasswordMetadata: u.PasswordMetadata,
		CreatedAt:        u.CreatedAt,
	}
}

// SignupRequest - request body for signup
type SignupRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginRequest - request body for login
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// ErrorDetail - error response detail
type ErrorDetail struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
	Hint    string      `json:"hint,omitempty"`
}

// ErrorResponse - error response envelope
type ErrorResponse struct {
	Success bool        `json:"success"`
	Error   ErrorDetail `json:"error"`
}

// SignupResponse - response for signup
type SignupResponse struct {
	Success bool       `json:"success"`
	User    PublicUser `json:"user"`
}

// LoginResponse - response for login
type LoginResponse struct {
	Success   bool       `json:"success"`
	User      PublicUser `json:"user"`
	Token     string     `json:"token"`
	ExpiresAt string     `json:"expiresAt"`
}

// UsersResponse - response for users list
type UsersResponse struct {
	Success bool         `json:"success"`
	Users   []PublicUser `json:"users"`
}

// HealthResponse - response for health check
type HealthResponse struct {
	Status    string  `json:"status"`
	Timestamp string  `json:"timestamp"`
	Uptime    float64 `json:"uptime"`
}

// PasswordValidation - password validation result
type PasswordValidation struct {
	IsValid            bool
	Errors            []string
	HasCaseSensitivity bool
}
