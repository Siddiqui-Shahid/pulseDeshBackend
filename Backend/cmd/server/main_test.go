package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"pulse-backend/internal/db"
	"pulse-backend/internal/handlers"
	"pulse-backend/internal/models"
)

func setupTestDB(t *testing.T) string {
	tmpdir, err := os.MkdirTemp("", "test_db")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	dbPath := filepath.Join(tmpdir, "test_users.json")
	
	// Create empty database file
	f, err := os.Create(dbPath)
	if err != nil {
		t.Fatalf("Failed to create DB file: %v", err)
	}
	f.WriteString("[]")
	f.Close()
	
	db.Init(dbPath)
	return tmpdir
}

func TestSignupIntegration(t *testing.T) {
	tmpdir := setupTestDB(t)
	defer os.RemoveAll(tmpdir)

	reqBody := models.SignupRequest{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "GoodPass123",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/auth/signup", bytes.NewReader(body))
	w := httptest.NewRecorder()

	handlers.SignupHandler(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected status %d, got %d", http.StatusCreated, w.Code)
	}

	var resp models.SignupResponse
	json.NewDecoder(w.Body).Decode(&resp)
	if !resp.Success || resp.User.Email != reqBody.Email {
		t.Errorf("unexpected response: %v", resp)
	}
}

func TestLoginIntegration(t *testing.T) {
	tmpdir := setupTestDB(t)
	defer os.RemoveAll(tmpdir)

	// Signup first
	signupReq := models.SignupRequest{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "GoodPass123",
	}
	signupBody, _ := json.Marshal(signupReq)
	signupHTTPReq := httptest.NewRequest("POST", "/auth/signup", bytes.NewReader(signupBody))
	w := httptest.NewRecorder()
	handlers.SignupHandler(w, signupHTTPReq)

	// Login
	loginReq := models.LoginRequest{
		Email:    "test@example.com",
		Password: "GoodPass123",
	}
	loginBody, _ := json.Marshal(loginReq)
	loginHTTPReq := httptest.NewRequest("POST", "/auth/login", bytes.NewReader(loginBody))
	w = httptest.NewRecorder()
	handlers.LoginHandler(w, loginHTTPReq)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var resp models.LoginResponse
	json.NewDecoder(w.Body).Decode(&resp)
	if !resp.Success || resp.Token == "" {
		t.Errorf("unexpected response: %v", resp)
	}
}

func TestHealthCheck(t *testing.T) {
	tmpdir := setupTestDB(t)
	defer os.RemoveAll(tmpdir)

	req := httptest.NewRequest("GET", "/api/health", nil)
	w := httptest.NewRecorder()

	handlers.HealthHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var resp models.HealthResponse
	json.NewDecoder(w.Body).Decode(&resp)
	if resp.Status != "healthy" {
		t.Errorf("expected healthy status, got %s", resp.Status)
	}
}

func TestGetUsers(t *testing.T) {
	tmpdir := setupTestDB(t)
	defer os.RemoveAll(tmpdir)

	// Create some users
	for i := 0; i < 3; i++ {
		reqBody := models.SignupRequest{
			Username: fmt.Sprintf("user%d", i),
			Email:    fmt.Sprintf("user%d@example.com", i),
			Password: "GoodPass123",
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/auth/signup", bytes.NewReader(body))
		w := httptest.NewRecorder()
		handlers.SignupHandler(w, req)
	}

	// Get users
	req := httptest.NewRequest("GET", "/api/users", nil)
	w := httptest.NewRecorder()
	handlers.UsersHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var resp models.UsersResponse
	json.NewDecoder(w.Body).Decode(&resp)
	if len(resp.Users) != 3 {
		t.Errorf("expected 3 users, got %d", len(resp.Users))
	}
}

func TestInvalidPassword(t *testing.T) {
	tmpdir := setupTestDB(t)
	defer os.RemoveAll(tmpdir)

	reqBody := models.SignupRequest{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "weak",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/auth/signup", bytes.NewReader(body))
	w := httptest.NewRecorder()

	handlers.SignupHandler(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}

	var resp models.ErrorResponse
	json.NewDecoder(w.Body).Decode(&resp)
	if resp.Error.Code != "WEAK_PASSWORD" {
		t.Errorf("expected WEAK_PASSWORD, got %s", resp.Error.Code)
	}
}

func TestDuplicateSignup(t *testing.T) {
	tmpdir := setupTestDB(t)
	defer os.RemoveAll(tmpdir)

	reqBody := models.SignupRequest{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "GoodPass123",
	}

	// First signup
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/auth/signup", bytes.NewReader(body))
	w := httptest.NewRecorder()
	handlers.SignupHandler(w, req)

	// Duplicate signup
	body, _ = json.Marshal(reqBody)
	req = httptest.NewRequest("POST", "/auth/signup", bytes.NewReader(body))
	w = httptest.NewRecorder()
	handlers.SignupHandler(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}

	var resp models.ErrorResponse
	json.NewDecoder(w.Body).Decode(&resp)
	if resp.Error.Code != "USER_EXISTS" {
		t.Errorf("expected USER_EXISTS, got %s", resp.Error.Code)
	}
}
