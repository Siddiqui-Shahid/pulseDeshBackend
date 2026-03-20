package auth

import (
	"time"
	"pulse-backend/internal/crypto"
	"pulse-backend/internal/db"
	"pulse-backend/internal/models"
)

// Signup creates a new user account.
func Signup(req models.SignupRequest) (*models.User, *models.ErrorResponse) {
	// Validate password strength
	validation := crypto.ValidatePasswordStrength(req.Password)
	if !validation.IsValid {
		return nil, &models.ErrorResponse{
			Success: false,
			Error: models.ErrorDetail{
				Code:    "WEAK_PASSWORD",
				Message: "Password does not meet requirements",
				Details: validation.Errors,
				Hint:    "Password must be at least 8 characters with uppercase, lowercase, and number",
			},
		}
	}

	// Read existing users
	users, err := db.ReadUsers()
	if err != nil {
		return nil, &models.ErrorResponse{
			Success: false,
			Error: models.ErrorDetail{
				Code:    "DB_ERROR",
				Message: "Failed to read user database",
				Details: err.Error(),
				Hint:    "Please try again later",
			},
		}
	}

	// Check if user exists
	for _, u := range users {
		if u.Email == req.Email || u.Username == req.Username {
			return nil, &models.ErrorResponse{
				Success: false,
				Error: models.ErrorDetail{
					Code:    "USER_EXISTS",
					Message: "User already exists",
					Details: map[string]string{
						"email":    req.Email,
						"username": req.Username,
					},
					Hint: "Use a different email or try logging in",
				},
			}
		}
	}

	// Generate salt and hash password
	salt, err := crypto.GenerateSalt()
	if err != nil {
		return nil, &models.ErrorResponse{
			Success: false,
			Error: models.ErrorDetail{
				Code:    "CRYPTO_ERROR",
				Message: "Failed to generate salt",
				Details: err.Error(),
				Hint:    "Please try again",
			},
		}
	}

	hash, err := crypto.HashPassword(req.Password, salt)
	if err != nil {
		return nil, &models.ErrorResponse{
			Success: false,
			Error: models.ErrorDetail{
				Code:    "CRYPTO_ERROR",
				Message: "Failed to hash password",
				Details: err.Error(),
				Hint:    "Please try again",
			},
		}
	}

	// Create new user
	newUser := models.User{
		ID:           crypto.GenerateID(),
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: hash,
		Salt:         salt,
		PasswordMetadata: models.PasswordMetadata{
			CaseSensitive:     true,
			RequiresUpperCase: true,
			RequiresLowerCase: true,
			RequiresNumber:    true,
			MinimumLength:     8,
		},
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}

	// Save user
	users = append(users, newUser)
	if err := db.WriteUsers(users); err != nil {
		return nil, &models.ErrorResponse{
			Success: false,
			Error: models.ErrorDetail{
				Code:    "DB_ERROR",
				Message: "Failed to save user",
				Details: err.Error(),
				Hint:    "Please try again later",
			},
		}
	}

	return &newUser, nil
}

// Login authenticates a user and returns a session token.
func Login(req models.LoginRequest) (*models.User, string, *models.ErrorResponse) {
	// Read users
	users, err := db.ReadUsers()
	if err != nil {
		return nil, "", &models.ErrorResponse{
			Success: false,
			Error: models.ErrorDetail{
				Code:    "DB_ERROR",
				Message: "Failed to read user database",
				Details: err.Error(),
				Hint:    "Please try again later",
			},
		}
	}

	// Find user by email
	var user *models.User
	for i := range users {
		if users[i].Email == req.Email {
			user = &users[i]
			break
		}
	}

	if user == nil {
		return nil, "", &models.ErrorResponse{
			Success: false,
			Error: models.ErrorDetail{
				Code:    "INVALID_CREDENTIALS",
				Message: "Invalid email or password",
				Hint:    "Check your credentials and try again",
			},
		}
	}

	// Verify password
	valid, err := crypto.VerifyPassword(req.Password, user.PasswordHash, user.Salt)
	if err != nil || !valid {
		return nil, "", &models.ErrorResponse{
			Success: false,
			Error: models.ErrorDetail{
				Code:    "INVALID_CREDENTIALS",
				Message: "Invalid email or password",
				Hint:    "Check your credentials and try again",
			},
		}
	}

	// Generate session token
	token, err := crypto.GenerateToken()
	if err != nil {
		return nil, "", &models.ErrorResponse{
			Success: false,
			Error: models.ErrorDetail{
				Code:    "CRYPTO_ERROR",
				Message: "Failed to generate session token",
				Details: err.Error(),
				Hint:    "Please try again",
			},
		}
	}

	return user, token, nil
}
