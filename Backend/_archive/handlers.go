package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

// jsonResponse writes a JSON body with the given HTTP status code.
func jsonResponse(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(payload) //nolint:errcheck
}

// errResp is a convenience constructor for errorResponse.
func errResp(code, message string, details interface{}, hint string) errorResponse {
	return errorResponse{
		Success: false,
		Error: errorDetail{
			Code:    code,
			Message: message,
			Details: details,
			Hint:    hint,
		},
	}
}

// apiError logs a structured error entry and returns the errorResponse.
// Format:
//
//	[ERROR] POST /auth/login | 401 | code=INVALID_PASSWORD | msg=invalid email or password | hint=Password is case-sensitive...
func apiError(r *http.Request, status int, resp errorResponse) errorResponse {
	hint := ""
	if resp.Error.Hint != "" {
		hint = fmt.Sprintf(" | hint=%s", resp.Error.Hint)
	}
	log.Printf("[ERROR] %s %s | %d | code=%s | msg=%s%s",
		r.Method, r.URL.Path, status, resp.Error.Code, resp.Error.Message, hint)
	return resp
}

// ---------------------------------------------------------------------------
// POST /signup  &  POST /auth/signup
// ---------------------------------------------------------------------------

func handleSignup(w http.ResponseWriter, r *http.Request) {
	var req signupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonResponse(w, http.StatusBadRequest,
			apiError(r, http.StatusBadRequest,
				errResp("INVALID_JSON", "request body is not valid JSON", nil, "")))
		return
	}

	log.Printf("[INFO]  %s %s | attempting signup for username=%q email=%q",
		r.Method, r.URL.Path, req.Username, req.Email)

	if req.Username == "" || req.Email == "" || req.Password == "" {
		jsonResponse(w, http.StatusBadRequest,
			apiError(r, http.StatusBadRequest,
				errResp("MISSING_FIELDS",
					"username, email and password are required",
					[]string{"username", "email", "password"}, "")))
		return
	}

	pv := validatePasswordStrength(req.Password)
	if !pv.IsValid {
		jsonResponse(w, http.StatusBadRequest,
			apiError(r, http.StatusBadRequest,
				errResp("PASSWORD_WEAK",
					"password does not meet requirements",
					pv.Errors,
					"Password must contain uppercase, lowercase, a number, and be at least 8 characters long")))
		return
	}

	users, err := readUsers()
	if err != nil {
		log.Printf("[ERROR] %s %s | DB read failed: %v", r.Method, r.URL.Path, err)
		jsonResponse(w, http.StatusInternalServerError,
			apiError(r, http.StatusInternalServerError,
				errResp("SERVER_ERROR", "internal server error", nil,
					"Something went wrong. Please try again later.")))
		return
	}

	for _, u := range users {
		if u.Email == req.Email || u.Username == req.Username {
			hint := "Username already taken"
			if u.Email == req.Email {
				hint = "Email already registered"
			}
			jsonResponse(w, http.StatusConflict,
				apiError(r, http.StatusConflict,
					errResp("USER_EXISTS",
						"user with that email or username already exists",
						nil, hint)))
			return
		}
	}

	salt, err := generateSalt()
	if err != nil {
		log.Printf("[ERROR] %s %s | salt generation failed: %v", r.Method, r.URL.Path, err)
		jsonResponse(w, http.StatusInternalServerError,
			apiError(r, http.StatusInternalServerError,
				errResp("SERVER_ERROR", "internal server error", nil, "")))
		return
	}

	hash, err := hashPassword(req.Password, salt)
	if err != nil {
		log.Printf("[ERROR] %s %s | password hashing failed: %v", r.Method, r.URL.Path, err)
		jsonResponse(w, http.StatusInternalServerError,
			apiError(r, http.StatusInternalServerError,
				errResp("SERVER_ERROR", "internal server error", nil, "")))
		return
	}

	user := User{
		ID:           generateID(),
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: hash,
		Salt:         salt,
		PasswordMetadata: PasswordMetadata{
			CaseSensitive:    true,
			RequiresUpperCase: true,
			RequiresLowerCase: true,
			RequiresNumber:   true,
			MinimumLength:    8,
		},
		CreatedAt: time.Now().UTC().Format(time.RFC3339Nano),
	}

	users = append(users, user)
	if err := writeUsers(users); err != nil {
		log.Printf("[ERROR] %s %s | DB write failed: %v", r.Method, r.URL.Path, err)
		jsonResponse(w, http.StatusInternalServerError,
			apiError(r, http.StatusInternalServerError,
				errResp("SERVER_ERROR", "internal server error", nil,
					"Something went wrong. Please try again later.")))
		return
	}

	log.Printf("[INFO]  %s %s | 201 | signup OK email=%q username=%q id=%s",
		r.Method, r.URL.Path, req.Email, req.Username, user.ID)
	jsonResponse(w, http.StatusCreated, signupResponse{
		Success: true,
		User:    user.toPublic(),
	})
}

// ---------------------------------------------------------------------------
// POST /login  &  POST /auth/login
// ---------------------------------------------------------------------------

func handleLogin(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonResponse(w, http.StatusBadRequest,
			apiError(r, http.StatusBadRequest,
				errResp("INVALID_JSON", "request body is not valid JSON", nil, "")))
		return
	}

	log.Printf("[INFO]  %s %s | attempting login for email=%q", r.Method, r.URL.Path, req.Email)

	if req.Email == "" || req.Password == "" {
		jsonResponse(w, http.StatusBadRequest,
			apiError(r, http.StatusBadRequest,
				errResp("MISSING_FIELDS",
					"email and password are required",
					[]string{"email", "password"}, "")))
		return
	}

	users, err := readUsers()
	if err != nil {
		log.Printf("[ERROR] %s %s | DB read failed: %v", r.Method, r.URL.Path, err)
		jsonResponse(w, http.StatusInternalServerError,
			apiError(r, http.StatusInternalServerError,
				errResp("SERVER_ERROR", "internal server error", nil,
					"Something went wrong. Please try again later.")))
		return
	}

	var found *User
	for i := range users {
		if users[i].Email == req.Email {
			found = &users[i]
			break
		}
	}

	if found == nil {
		jsonResponse(w, http.StatusUnauthorized,
			apiError(r, http.StatusUnauthorized,
				errResp("INVALID_CREDENTIALS", "invalid email or password", nil, "")))
		return
	}

	ok, err := verifyPassword(req.Password, found.PasswordHash, found.Salt)
	if err != nil || !ok {
		jsonResponse(w, http.StatusUnauthorized,
			apiError(r, http.StatusUnauthorized,
				errResp("INVALID_PASSWORD",
					"invalid email or password",
					nil,
					"Password is case-sensitive. Check uppercase, lowercase, numbers, and length (min 8 characters).")),
		)
		return
	}

	token, err := generateToken()
	if err != nil {
		log.Printf("[ERROR] %s %s | token generation failed: %v", r.Method, r.URL.Path, err)
		jsonResponse(w, http.StatusInternalServerError,
			apiError(r, http.StatusInternalServerError,
				errResp("SERVER_ERROR", "internal server error", nil, "")))
		return
	}

	log.Printf("[INFO]  %s %s | 200 | login OK email=%q id=%s",
		r.Method, r.URL.Path, req.Email, found.ID)
	jsonResponse(w, http.StatusOK, loginResponse{
		Success:   true,
		User:      found.toPublic(),
		Token:     token,
		ExpiresAt: time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339Nano),
	})
}

// ---------------------------------------------------------------------------
// GET /users
// ---------------------------------------------------------------------------

func handleUsers(w http.ResponseWriter, r *http.Request) {
	users, err := readUsers()
	if err != nil {
		log.Printf("[ERROR] %s %s | DB read failed: %v", r.Method, r.URL.Path, err)
		jsonResponse(w, http.StatusInternalServerError,
			apiError(r, http.StatusInternalServerError,
				errResp("SERVER_ERROR", "internal server error", nil, "Failed to fetch users")))
		return
	}

	pub := make([]PublicUser, len(users))
	for i, u := range users {
		pub[i] = u.toPublic()
	}
	log.Printf("[INFO]  %s %s | 200 | listed %d users", r.Method, r.URL.Path, len(pub))
	jsonResponse(w, http.StatusOK, usersResponse{Success: true, Users: pub})
}

// ---------------------------------------------------------------------------
// GET /health
// ---------------------------------------------------------------------------

func handleHealth(startTime time.Time) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uptime := time.Since(startTime).Seconds()
		log.Printf("[INFO]  %s %s | 200 | uptime=%.2fs", r.Method, r.URL.Path, uptime)
		jsonResponse(w, http.StatusOK, healthResponse{
			Status:    "ok",
			Timestamp: time.Now().UTC().Format(time.RFC3339Nano),
			Uptime:    uptime,
		})
	}
}

// ---------------------------------------------------------------------------
// 404 fallback
// ---------------------------------------------------------------------------

func handleNotFound(w http.ResponseWriter, r *http.Request) {
	jsonResponse(w, http.StatusNotFound,
		apiError(r, http.StatusNotFound,
			errResp("NOT_FOUND",
				"endpoint not found",
				nil,
				fmt.Sprintf("%s %s does not exist", r.Method, r.URL.Path))))
}
