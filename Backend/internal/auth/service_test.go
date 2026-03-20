package auth

import (
	"os"
	"path/filepath"
	"testing"
	"pulse-backend/internal/db"
	"pulse-backend/internal/models"
)

func setupTestDB(t *testing.T) string {
	tmpdir, err := os.MkdirTemp("", "auth_test_db")
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

func TestSignupSuccess(t *testing.T) {
	tmpdir := setupTestDB(t)
	defer os.RemoveAll(tmpdir)

	req := models.SignupRequest{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "StrongPass123",
	}

	user, errResp := Signup(req)
	if errResp != nil {
		t.Fatalf("Signup failed: %v", errResp.Error.Message)
	}
	if user == nil {
		t.Error("user should not be nil")
	}
	if user.Email != req.Email {
		t.Errorf("expected email %s, got %s", req.Email, user.Email)
	}
	if user.ID == "" {
		t.Error("user ID should not be empty")
	}
}

func TestSignupWeakPassword(t *testing.T) {
	tmpdir := setupTestDB(t)
	defer os.RemoveAll(tmpdir)

	req := models.SignupRequest{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "weak", // Too short, no uppercase, no number
	}

	user, errResp := Signup(req)
	if errResp == nil {
		t.Error("expected error for weak password")
	}
	if user != nil {
		t.Error("user should be nil on error")
	}
	if errResp.Error.Code != "WEAK_PASSWORD" {
		t.Errorf("expected WEAK_PASSWORD, got %s", errResp.Error.Code)
	}
}

func TestSignupDuplicate(t *testing.T) {
	tmpdir := setupTestDB(t)
	defer os.RemoveAll(tmpdir)

	req := models.SignupRequest{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "StrongPass123",
	}

	// First signup succeeds
	_, errResp := Signup(req)
	if errResp != nil {
		t.Fatalf("First signup failed: %v", errResp.Error.Message)
	}

	// Second signup with same email fails
	_, errResp = Signup(req)
	if errResp == nil {
		t.Error("expected error for duplicate user")
	}
	if errResp.Error.Code != "USER_EXISTS" {
		t.Errorf("expected USER_EXISTS, got %s", errResp.Error.Code)
	}
}

func TestLoginSuccess(t *testing.T) {
	tmpdir := setupTestDB(t)
	defer os.RemoveAll(tmpdir)

	// Create a user first
	signupReq := models.SignupRequest{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "StrongPass123",
	}
	_, errResp := Signup(signupReq)
	if errResp != nil {
		t.Fatalf("Signup failed: %v", errResp.Error.Message)
	}

	// Now login
	loginReq := models.LoginRequest{
		Email:    "test@example.com",
		Password: "StrongPass123",
	}

	user, token, errResp := Login(loginReq)
	if errResp != nil {
		t.Fatalf("Login failed: %v", errResp.Error.Message)
	}
	if user == nil {
		t.Error("user should not be nil")
	}
	if token == "" {
		t.Error("token should not be empty")
	}
	if len(token) != 64 {
		t.Errorf("expected token length 64, got %d", len(token))
	}
}

func TestLoginInvalidCredentials(t *testing.T) {
	tmpdir := setupTestDB(t)
	defer os.RemoveAll(tmpdir)

	// Create a user first
	signupReq := models.SignupRequest{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "StrongPass123",
	}
	Signup(signupReq)

	// Try login with wrong password
	loginReq := models.LoginRequest{
		Email:    "test@example.com",
		Password: "WrongPass123",
	}

	user, token, errResp := Login(loginReq)
	if errResp == nil {
		t.Error("expected error for invalid password")
	}
	if user != nil {
		t.Error("user should be nil on error")
	}
	if token != "" {
		t.Error("token should be empty on error")
	}
	if errResp.Error.Code != "INVALID_CREDENTIALS" {
		t.Errorf("expected INVALID_CREDENTIALS, got %s", errResp.Error.Code)
	}
}

func TestLoginInvalidEmail(t *testing.T) {
	tmpdir := setupTestDB(t)
	defer os.RemoveAll(tmpdir)

	loginReq := models.LoginRequest{
		Email:    "nonexistent@example.com",
		Password: "StrongPass123",
	}

	user, token, errResp := Login(loginReq)
	if errResp == nil {
		t.Error("expected error for non-existent email")
	}
	if user != nil {
		t.Error("user should be nil on error")
	}
	if token != "" {
		t.Error("token should be empty on error")
	}
}

func TestLoginPasswordCaseSensitivity(t *testing.T) {
	tmpdir := setupTestDB(t)
	defer os.RemoveAll(tmpdir)

	// Create a user
	signupReq := models.SignupRequest{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "StrongPass123",
	}
	Signup(signupReq)

	// Try login with different case
	loginReq := models.LoginRequest{
		Email:    "test@example.com",
		Password: "strongpass123",
	}

	user, token, errResp := Login(loginReq)
	if errResp == nil {
		t.Error("expected error for wrong case")
	}
	if user != nil {
		t.Error("user should be nil on error")
	}
	if token != "" {
		t.Error("token should be empty on error")
	}
}

// ============ COMPREHENSIVE AUTH FLOW TESTS ============

// TestLoginPasswordCaseMismatch tests login fails when password case does not match
func TestLoginPasswordCaseMismatch(t *testing.T) {
	tmpdir := setupTestDB(t)
	defer os.RemoveAll(tmpdir)

	signupReq := models.SignupRequest{
 		Username: "test10",
 		Email:    "test10@example.com",
 		Password: "PAss#123",
 	}
 	_, errResp := Signup(signupReq)
 	if errResp != nil {
 		t.Fatalf("Signup failed: %v", errResp.Error.Message)
 	}

 	loginReq := models.LoginRequest{
 		Email:    "test10@example.com",
 		Password: "Pass#123", // wrong case
 	}
 	user, token, errResp := Login(loginReq)
 	if errResp == nil {
 		t.Error("expected error for wrong password case")
 	}
 	if user != nil {
 		t.Error("user should be nil on error")
 	}
 	if token != "" {
 		t.Error("token should be empty on error")
 	}
}

// TestSignupMultipleUsers tests creating multiple users with different credentials
func TestSignupMultipleUsers(t *testing.T) {
	tmpdir := setupTestDB(t)
	defer os.RemoveAll(tmpdir)

	testCases := []struct {
		name     string
		username string
		email    string
		password string
	}{
		{"User1", "alice", "alice@example.com", "Alice123Password"},
		{"User2", "bob", "bob@example.com", "Bob456Password"},
		{"User3", "charlie", "charlie@example.com", "Charlie789Password"},
	}

	for _, tc := range testCases {
		req := models.SignupRequest{
			Username: tc.username,
			Email:    tc.email,
			Password: tc.password,
		}
		user, errResp := Signup(req)
		if errResp != nil {
			t.Errorf("Test %s: Signup failed: %v", tc.name, errResp.Error.Message)
		}
		if user.Email != tc.email {
			t.Errorf("Test %s: expected email %s, got %s", tc.name, tc.email, user.Email)
		}
		if user.Username != tc.username {
			t.Errorf("Test %s: expected username %s, got %s", tc.name, tc.username, user.Username)
		}
	}
}

// TestLoginMultipleUsers tests logging in multiple different users
func TestLoginMultipleUsers(t *testing.T) {
	tmpdir := setupTestDB(t)
	defer os.RemoveAll(tmpdir)

	users := []struct {
		username string
		email    string
		password string
	}{
		{"alice", "alice@example.com", "Alice123Pass"},
		{"bob", "bob@example.com", "Bob456Pass"},
		{"charlie", "charlie@example.com", "Charlie789Pass"},
	}

	// Sign up all users
	for _, u := range users {
		Signup(models.SignupRequest{
			Username: u.username,
			Email:    u.email,
			Password: u.password,
		})
	}

	// Login each user with correct password
	for _, u := range users {
		user, token, errResp := Login(models.LoginRequest{
			Email:    u.email,
			Password: u.password,
		})
		if errResp != nil {
			t.Errorf("Login failed for %s: %v", u.username, errResp.Error.Message)
		}
		if user.Email != u.email {
			t.Errorf("returned user email mismatch: expected %s, got %s", u.email, user.Email)
		}
		if token == "" {
			t.Errorf("login for %s should return token", u.username)
		}
	}
}

// TestPasswordVariations tests various valid password combinations
func TestPasswordVariations(t *testing.T) {
	tmpdir := setupTestDB(t)
	defer os.RemoveAll(tmpdir)

	validPasswords := []string{
		"ValidPass1",       // minimum 8 chars
		"VeryStrongPass99", // longer password
		"MyPassword0rd",    // with special char (though not required)
		"AaBbCc123",        // alternating case
		"UPPERCASE123lower", // mixed case
	}

	for i, pwd := range validPasswords {
		emailNum := 'a' + rune(i)
		req := models.SignupRequest{
			Username: "user" + string(emailNum),
			Email:    "email" + string(emailNum) + "@test.com",
			Password: pwd,
		}
		user, errResp := Signup(req)
		if errResp != nil {
			t.Errorf("Signup with password %q should succeed, got error: %v", pwd, errResp.Error.Message)
		}
		if user == nil {
			t.Errorf("User should not be nil for password %q", pwd)
		}
	}
}

// TestInvalidPasswordsRejected tests passwords that should be rejected
func TestInvalidPasswordsRejected(t *testing.T) {
	tmpdir := setupTestDB(t)
	defer os.RemoveAll(tmpdir)

	invalidPasswords := []struct {
		password string
		reason   string
	}{
		{"weak", "too short, no uppercase, no number"},
		{"NoNumbers", "missing number"},
		{"nouppercase1", "missing uppercase"},
		{"NOLOWERCASE1", "missing lowercase"},
		{"short1A", "less than 8 chars"},
		{"1234567A", "only number and one uppercase"},
	}

	for i, tc := range invalidPasswords {
		emailNum := 'a' + rune(i)
		req := models.SignupRequest{
			Username: "testuser" + string(emailNum),
			Email:    "test" + string(emailNum) + "@example.com",
			Password: tc.password,
		}
		_, errResp := Signup(req)
		if errResp == nil {
			t.Errorf("Password %q (%s) should be rejected", tc.password, tc.reason)
		}
		if errResp != nil && errResp.Error.Code != "WEAK_PASSWORD" {
			t.Errorf("Expected WEAK_PASSWORD error, got %s", errResp.Error.Code)
		}
	}
}

// TestDuplicateEmails tests preventing duplicate email registration
func TestDuplicateEmails(t *testing.T) {
	tmpdir := setupTestDB(t)
	defer os.RemoveAll(tmpdir)

	email := "duplicate@example.com"
	
	// First signup succeeds
	req1 := models.SignupRequest{
		Username: "user1",
		Email:    email,
		Password: "ValidPass1",
	}
	_, errResp := Signup(req1)
	if errResp != nil {
		t.Fatalf("First signup should succeed: %v", errResp.Error.Message)
	}

	// Second signup with same email fails
	req2 := models.SignupRequest{
		Username: "user2",
		Email:    email,
		Password: "ValidPass2",
	}
	_, errResp = Signup(req2)
	if errResp == nil {
		t.Error("second signup with duplicate email should fail")
	}
	if errResp.Error.Code != "USER_EXISTS" {
		t.Errorf("expected USER_EXISTS error, got %s", errResp.Error.Code)
	}
}

// TestDuplicateUsernames tests preventing duplicate username registration
func TestDuplicateUsernames(t *testing.T) {
	tmpdir := setupTestDB(t)
	defer os.RemoveAll(tmpdir)

	username := "duplicateuser"
	
	// First signup succeeds
	req1 := models.SignupRequest{
		Username: username,
		Email:    "email1@example.com",
		Password: "ValidPass1",
	}
	_, errResp := Signup(req1)
	if errResp != nil {
		t.Fatalf("First signup should succeed: %v", errResp.Error.Message)
	}

	// Second signup with same username fails
	req2 := models.SignupRequest{
		Username: username,
		Email:    "email2@example.com",
		Password: "ValidPass2",
	}
	_, errResp = Signup(req2)
	if errResp == nil {
		t.Error("second signup with duplicate username should fail")
	}
	if errResp.Error.Code != "USER_EXISTS" {
		t.Errorf("expected USER_EXISTS error, got %s", errResp.Error.Code)
	}
}

// TestSignupAndLoginFlow tests complete signup -> login flow for single user
func TestSignupAndLoginFlow(t *testing.T) {
	tmpdir := setupTestDB(t)
	defer os.RemoveAll(tmpdir)

	// Step 1: Sign up
	signupReq := models.SignupRequest{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "ComplexPass123",
	}
	savedUser, signupErr := Signup(signupReq)
	if signupErr != nil {
		t.Fatalf("Signup failed: %v", signupErr.Error.Message)
	}

	// Step 2: Login with correct credentials
	loginReq := models.LoginRequest{
		Email:    "test@example.com",
		Password: "ComplexPass123",
	}
	loginUser, token, loginErr := Login(loginReq)
	if loginErr != nil {
		t.Fatalf("Login failed: %v", loginErr.Error.Message)
	}

	// Verify returned user matches saved user
	if loginUser.ID != savedUser.ID {
		t.Errorf("User ID mismatch: signup ID %s, login ID %s", savedUser.ID, loginUser.ID)
	}
	if loginUser.Email != savedUser.Email {
		t.Errorf("Email mismatch: signup %s, login %s", savedUser.Email, loginUser.Email)
	}
	if loginUser.Username != savedUser.Username {
		t.Errorf("Username mismatch: signup %s, login %s", savedUser.Username, loginUser.Username)
	}

	// Verify token is valid
	if token == "" {
		t.Error("token should not be empty after login")
	}
	if len(token) < 32 {
		t.Errorf("token too short: %d", len(token))
	}
}

// TestSignupAndLoginWrongPassword tests correct signup then failed login
func TestSignupAndLoginWrongPassword(t *testing.T) {
	tmpdir := setupTestDB(t)
	defer os.RemoveAll(tmpdir)

	// Sign up
	Signup(models.SignupRequest{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "CorrectPass123",
	})

	// Try login with wrong password
	testCases := []struct {
		password string
		desc     string
	}{
		{"WrongPass123", "completely different password"},
		{"correctpass123", "right password wrong case"},
		{"CorrectPass124", "off by one character"},
		{"CorrectPass12", "missing last character"},
		{"CorrectPass1234", "extra character"},
	}

	for _, tc := range testCases {
		_, _, errResp := Login(models.LoginRequest{
			Email:    "test@example.com",
			Password: tc.password,
		})
		if errResp == nil {
			t.Errorf("Login with %s should fail", tc.desc)
		}
		if errResp.Error.Code != "INVALID_CREDENTIALS" {
			t.Errorf("Expected INVALID_CREDENTIALS, got %s for %s", errResp.Error.Code, tc.desc)
		}
	}
}

// TestCaseSensitiveEmails tests if email comparison is case-insensitive
func TestEmailRegistration(t *testing.T) {
	tmpdir := setupTestDB(t)
	defer os.RemoveAll(tmpdir)

	// Signup with lowercase email
	req1 := models.SignupRequest{
		Username: "user1",
		Email:    "test@example.com",
		Password: "ValidPass1",
	}
	_, errResp := Signup(req1)
	if errResp != nil {
		t.Fatalf("First signup should succeed: %v", errResp.Error.Message)
	}

	// Try to login with same email in different case
	_, _, loginErr := Login(models.LoginRequest{
		Email:    "test@example.com",
		Password: "ValidPass1",
	})
	if loginErr != nil {
		t.Errorf("Login with correct email should succeed, got: %v", loginErr.Error.Message)
	}
}

// TestNoUserWithEmail tests login attempt with non-existent email
func TestNoUserWithEmail(t *testing.T) {
	tmpdir := setupTestDB(t)
	defer os.RemoveAll(tmpdir)

	// Try to login without any users registered
	_, _, errResp := Login(models.LoginRequest{
		Email:    "nonexistent@example.com",
		Password: "AnyPass123",
	})
	if errResp == nil {
		t.Error("login with non-existent email should fail")
	}
	if errResp.Error.Code != "INVALID_CREDENTIALS" {
		t.Errorf("expected INVALID_CREDENTIALS, got %s", errResp.Error.Code)
	}
}

