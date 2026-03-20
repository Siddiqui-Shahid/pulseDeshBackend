package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
	"pulse-backend/internal/auth"
	"pulse-backend/internal/db"
	"pulse-backend/internal/models"
	"pulse-backend/internal/response"
)

// SignupHandler handles POST /auth/signup requests.
func SignupHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("🔵 [SIGNUP] Request received: Method=%s Path=%s", r.Method, r.URL.Path)
	log.Printf("🔵 [SIGNUP] Headers: %v", r.Header)
	
	if r.Method != http.MethodPost {
		log.Printf("🔴 [SIGNUP] Invalid method: %s", r.Method)
		errResp := response.ErrResp("METHOD_NOT_ALLOWED", "Method not allowed", nil, "Use POST method")
		response.APIError(r, http.StatusMethodNotAllowed, errResp)
		response.JSONResponse(w, http.StatusMethodNotAllowed, errResp)
		return
	}

	var req models.SignupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("🔴 [SIGNUP] Failed to decode request: %v", err)
		errResp := response.ErrResp("INVALID_REQUEST", "Invalid request body", err.Error(), "Check your request JSON")
		response.APIError(r, http.StatusBadRequest, errResp)
		response.JSONResponse(w, http.StatusBadRequest, errResp)
		return
	}

	log.Printf("🔵 [SIGNUP] Decoded request: username=%s, email=%s", req.Username, req.Email)

	user, errResp := auth.Signup(req)
	if errResp != nil {
		log.Printf("🔴 [SIGNUP] Signup failed: %v", errResp.Error.Message)
		response.APIError(r, http.StatusBadRequest, *errResp)
		response.JSONResponse(w, http.StatusBadRequest, errResp)
		return
	}

	resp := models.SignupResponse{
		Success: true,
		User:    user.ToPublic(),
	}
	log.Printf("✅ [SIGNUP] Success: user=%s | signup successful", user.Email)
	response.JSONResponse(w, http.StatusCreated, resp)
}

// LoginHandler handles POST /auth/login requests.
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		errResp := response.ErrResp("METHOD_NOT_ALLOWED", "Method not allowed", nil, "Use POST method")
		response.APIError(r, http.StatusMethodNotAllowed, errResp)
		response.JSONResponse(w, http.StatusMethodNotAllowed, errResp)
		return
	}

	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errResp := response.ErrResp("INVALID_REQUEST", "Invalid request body", err.Error(), "Check your request JSON")
		response.APIError(r, http.StatusBadRequest, errResp)
		response.JSONResponse(w, http.StatusBadRequest, errResp)
		return
	}

	user, token, errResp := auth.Login(req)
	if errResp != nil {
		response.APIError(r, http.StatusUnauthorized, *errResp)
		response.JSONResponse(w, http.StatusUnauthorized, errResp)
		return
	}

	expiresAt := time.Now().Add(24 * time.Hour).Format(time.RFC3339)
	resp := models.LoginResponse{
		Success:   true,
		User:      user.ToPublic(),
		Token:     token,
		ExpiresAt: expiresAt,
	}
	log.Printf("[INFO] %s %s | user=%s | login successful", r.Method, r.URL.Path, user.Email)
	response.JSONResponse(w, http.StatusOK, resp)
}

// UsersHandler handles GET /api/users requests.
func UsersHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		errResp := response.ErrResp("METHOD_NOT_ALLOWED", "Method not allowed", nil, "Use GET method")
		response.APIError(r, http.StatusMethodNotAllowed, errResp)
		response.JSONResponse(w, http.StatusMethodNotAllowed, errResp)
		return
	}

	users, err := db.ReadUsers()
	if err != nil {
		errResp := response.ErrResp("DB_ERROR", "Failed to read users", err.Error(), "Try again later")
		response.APIError(r, http.StatusInternalServerError, errResp)
		response.JSONResponse(w, http.StatusInternalServerError, errResp)
		return
	}

	publicUsers := make([]models.PublicUser, len(users))
	for i, u := range users {
		publicUsers[i] = u.ToPublic()
	}

	resp := models.UsersResponse{
		Success: true,
		Users:   publicUsers,
	}
	log.Printf("[INFO] %s %s | returned %d users", r.Method, r.URL.Path, len(publicUsers))
	response.JSONResponse(w, http.StatusOK, resp)
}

// HealthHandler handles GET /api/health requests.
func HealthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		errResp := response.ErrResp("METHOD_NOT_ALLOWED", "Method not allowed", nil, "Use GET method")
		response.APIError(r, http.StatusMethodNotAllowed, errResp)
		response.JSONResponse(w, http.StatusMethodNotAllowed, errResp)
		return
	}

	resp := models.HealthResponse{
		Status:    "healthy",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Uptime:    0.0,
	}
	log.Printf("[INFO] %s %s | health check", r.Method, r.URL.Path)
	response.JSONResponse(w, http.StatusOK, resp)
}

// NotFoundHandler handles 404 requests.
func NotFoundHandler(w http.ResponseWriter, r *http.Request) {
	errResp := response.ErrResp("NOT_FOUND", "Endpoint not found", nil, fmt.Sprintf("No handler for %s %s", r.Method, r.URL.Path))
	response.APIError(r, http.StatusNotFound, errResp)
	response.JSONResponse(w, http.StatusNotFound, errResp)
}
