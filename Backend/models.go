package main

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

// toPublic strips sensitive fields from a User.
func (u User) toPublic() PublicUser {
	return PublicUser{
		ID:               u.ID,
		Username:         u.Username,
		Email:            u.Email,
		PasswordMetadata: u.PasswordMetadata,
		CreatedAt:        u.CreatedAt,
	}
}

// ---------- request bodies ----------

type signupRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// ---------- response envelopes ----------

type errorDetail struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
	Hint    string      `json:"hint,omitempty"`
}

type errorResponse struct {
	Success bool        `json:"success"`
	Error   errorDetail `json:"error"`
}

type signupResponse struct {
	Success bool       `json:"success"`
	User    PublicUser `json:"user"`
}

type loginResponse struct {
	Success   bool       `json:"success"`
	User      PublicUser `json:"user"`
	Token     string     `json:"token"`
	ExpiresAt string     `json:"expiresAt"`
}

type usersResponse struct {
	Success bool         `json:"success"`
	Users   []PublicUser `json:"users"`
}

type healthResponse struct {
	Status    string  `json:"status"`
	Timestamp string  `json:"timestamp"`
	Uptime    float64 `json:"uptime"`
}

// passwordValidation holds the result of validatePasswordStrength.
type passwordValidation struct {
	IsValid          bool
	Errors           []string
	HasCaseSensitivity bool
}
