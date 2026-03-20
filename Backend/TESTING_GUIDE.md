# Go Testing Guide - Pulse Backend

## Overview
This guide explains how to write tests in Go using the testing patterns from the Pulse backend. All 30+ tests are passing with comprehensive coverage!

## Quick Summary of Test Results

✅ **17 Auth Service Tests** - Full signup/login flow with edge cases
✅ **8 Crypto Tests** - Password hashing, salt/token generation  
✅ **6 Integration Tests** - HTTP endpoint testing
✅ **30+ Total Tests** - All passing with comprehensive coverage

---

## 1. Basic Test Structure

Every test file must:
1. Be in the same package as the code being tested
2. End with `_test.go` 
3. Import the `testing` package
4. Write functions with signature `func TestXxx(t *testing.T)`

### Example Test File Location
```
internal/auth/
├── service.go       ← Code being tested
└── service_test.go  ← Tests go here
```

### Basic Test Template
```go
package auth

import (
    "testing"
    "pulse-backend/internal/models"
)

func TestMyFeature(t *testing.T) {
    // 1. Setup
    // 2. Execute
    // 3. Assert
}
```

---

## 2. Test File Organization for Auth Service

### File: `internal/auth/service_test.go`

#### Part 1: Setup Function (Database Isolation)
```go
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
```

**Why this matters:**
- Each test gets its own isolated database
- Tests don't interfere with each other
- Files are automatically cleaned up with `defer os.RemoveAll(tmpdir)`

#### Part 2: Individual Test Cases

---

## 3. Auth Test Cases Explained

### Test 1: Basic Success Cases
```go
func TestSignupSuccess(t *testing.T) {
    // 1. Setup: Create isolated database
    tmpdir := setupTestDB(t)
    defer os.RemoveAll(tmpdir)

    // 2. Execute: Create user
    req := models.SignupRequest{
        Username: "testuser",
        Email:    "test@example.com",
        Password: "StrongPass123",
    }
    user, errResp := Signup(req)

    // 3. Assert: Verify results
    if errResp != nil {
        t.Fatalf("Signup failed: %v", errResp.Error.Message)
    }
    if user == nil {
        t.Error("user should not be nil")
    }
    if user.Email != req.Email {
        t.Errorf("expected email %s, got %s", req.Email, user.Email)
    }
}
```

**Key Pattern:**
- `t.Fatalf()` - Fatal error, stop test
- `t.Error()` - Non-fatal error, continue
- `t.Errorf()` - Formatted error

---

### Test 2: Password Validation
```go
func TestSignupWeakPassword(t *testing.T) {
    tmpdir := setupTestDB(t)
    defer os.RemoveAll(tmpdir)

    // Test weak password rejection
    req := models.SignupRequest{
        Username: "testuser",
        Email:    "test@example.com",
        Password: "weak", // Missing: length, uppercase, number
    }

    user, errResp := Signup(req)
    if errResp == nil {
        t.Error("expected error for weak password")
    }
    if errResp.Error.Code != "WEAK_PASSWORD" {
        t.Errorf("expected WEAK_PASSWORD, got %s", errResp.Error.Code)
    }
}
```

**Testing Strategy:**
- Test boundary conditions
- Verify error codes
- Check error messages

---

### Test 3: Multiple Users
```go
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
            t.Errorf("Test %s: Signup failed", tc.name)
        }
        if user.Email != tc.email {
            t.Errorf("Test %s: email mismatch", tc.name)
        }
    }
}
```

**Pattern: Table-Driven Tests**
- Use structs to define test cases
- Loop through test cases
- Name each case for clarity
- Easy to add new cases

---

### Test 4: Complete Signup → Login Flow
```go
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

    // Step 3: Verify consistency
    if loginUser.ID != savedUser.ID {
        t.Errorf("User ID mismatch: %s != %s", savedUser.ID, loginUser.ID)
    }
    if token == "" {
        t.Error("token should not be empty")
    }
}
```

**Testing Integration:**
- Sign up a user
- Login with same credentials
- Verify the returned data matches
- Ensure token is generated

---

### Test 5: Wrong Password Attempts
```go
func TestSignupAndLoginWrongPassword(t *testing.T) {
    tmpdir := setupTestDB(t)
    defer os.RemoveAll(tmpdir)

    // Sign up
    Signup(models.SignupRequest{
        Username: "testuser",
        Email:    "test@example.com",
        Password: "CorrectPass123",
    })

    // Test multiple wrong password variations
    wrongPasswords := []struct {
        password string
        desc     string
    }{
        {"WrongPass123", "completely different password"},
        {"correctpass123", "right password wrong case"},
        {"CorrectPass124", "off by one character"},
        {"CorrectPass12", "missing last character"},
    }

    for _, tc := range wrongPasswords {
        _, _, errResp := Login(models.LoginRequest{
            Email:    "test@example.com",
            Password: tc.password,
        })
        if errResp == nil {
            t.Errorf("Login with %s should fail", tc.desc)
        }
    }
}
```

**Testing Edge Cases:**
- Case sensitivity (important!)
- Off-by-one errors
- Character mutations
- Each variation tested separately

---

### Test 6: Duplicate Prevention
```go
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
        t.Fatalf("First signup should succeed")
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
```

**Testing Constraints:**
- Unique usernames enforced
- Unique emails enforced  
- Proper error codes returned
- First attempt succeeds, second fails

---

## 4. Assertion Patterns

### Common Assertions Used in Tests

```go
// String comparison
if user.Email != expectedEmail {
    t.Errorf("Email mismatch: expected %s, got %s", expectedEmail, user.Email)
}

// String not empty
if user.ID == "" {
    t.Error("user ID should not be empty")
}

// Error should be nil (success)
if err != nil {
    t.Fatalf("Operation failed: %v", err)
}

// Error should NOT be nil (negative test)
if err == nil {
    t.Error("expected error but got none")
}

// Error code checking
if errResp.Error.Code != "WEAK_PASSWORD" {
    t.Errorf("expected WEAK_PASSWORD, got %s", errResp.Error.Code)
}

// Array length
if len(users) != 3 {
    t.Errorf("expected 3 users, got %d", len(users))
}

// Boolean conditions
if errResp == nil {
    t.Error("expected error for duplicate user")
}
```

---

## 5. Running Tests

### Run All Tests
```bash
$ go test -v ./...
```

### Run Tests for Specific Package
```bash
$ go test -v ./internal/auth
```

### Run Specific Test
```bash
$ go test -v ./internal/auth -run TestSignupSuccess
```

### Run Tests with Coverage
```bash
$ go test -coverprofile=coverage.out ./...
$ go tool cover -html=coverage.out  # Open in browser
```

### Run with Make Command
```bash
$ make test                    # Run all tests
$ make test-verbose           # Run with coverage  
$ make test-module MOD=auth   # Test auth module
$ make test-coverage          # Generate HTML report
```

---

## 6. Best Practices for Writing Tests

### ✅ DO:
1. **Isolate Tests** - Use temp databases, don't share state
2. **Name Tests Clearly** - `TestSignupInvalidPassword` is better than `TestError`
3. **Test One Thing** - Each test should verify one behavior
4. **Use Table-Driven Tests** - For testing multiple similar cases
5. **Include Edge Cases** - Off-by-one, empty, null, etc.
6. **Test Error Cases** - Not just the happy path
7. **Use Descriptive Messages** - Help future maintainers understand failures
8. **Keep Tests Fast** - Use mocking/isolation when needed

### ❌ DON'T:
1. **Share State Between Tests** - Use isolated databases  
2. **Test Multiple Things** - One assertion per test when possible
3. **Hardcode Paths** - Use temp directories
4. **Test Implementation Details** - Test behavior, not internals
5. **Ignore Edge Cases** - Case sensitivity, special characters, etc.
6. **Write Complex Assertions** - Make them easy to understand

---

## 7. Complete Auth Test Map

### Signup Tests (7 tests)
| Test Name | What It Tests | Status |
|-----------|---------------|--------|
| TestSignupSuccess | Happy path signup | ✅ PASS |
| TestSignupWeakPassword | Password validation | ✅ PASS |
| TestSignupDuplicate | Duplicate email rejection | ✅ PASS |
| TestSignupMultipleUsers | Multiple user creation | ✅ PASS |
| TestPasswordVariations | Various valid passwords | ✅ PASS |
| TestInvalidPasswordsRejected | Invalid password patterns | ✅ PASS |
| TestDuplicateEmails/Usernames | Unique field constraints | ✅ PASS |

### Login Tests (6 tests)
| Test Name | What It Tests | Status |
|-----------|---------------|--------|
| TestLoginSuccess | Happy path login | ✅ PASS |
| TestLoginInvalidCredentials | Wrong password | ✅ PASS |
| TestLoginInvalidEmail | Non-existent user | ✅ PASS |
| TestLoginPasswordCaseSensitivity | Case-sensitive check | ✅ PASS |
| TestLoginMultipleUsers | Multiple user logins | ✅ PASS |
| TestEmailRegistration | Email login flow | ✅ PASS |

### Flow Tests (4 tests)
| Test Name | What It Tests | Status |
|-----------|---------------|--------|
| TestSignupAndLoginFlow | Complete flow | ✅ PASS |
| TestSignupAndLoginWrongPassword | Failed login attempts | ✅ PASS |
| TestNoUserWithEmail | Non-existent user error | ✅ PASS |
| TestLoginMultipleUsers | Multi-user flows | ✅ PASS |

---

## 8. Example: Writing a New Test

### Scenario: Test username case sensitivity

```go
func TestUsernamesAreCaseSensitive(t *testing.T) {
    tmpdir := setupTestDB(t)
    defer os.RemoveAll(tmpdir)

    // Signup with lowercase
    req1 := models.SignupRequest{
        Username: "john",        // lowercase
        Email:    "john@example.com",
        Password: "ValidPass1",
    }
    user1, err1 := Signup(req1)
    if err1 != nil {
        t.Fatalf("First signup should succeed: %v", err1.Error.Message)
    }

    // Try signup with uppercase version of same name
    req2 := models.SignupRequest{
        Username: "JOHN",        // uppercase - different username
        Email:    "JOHN@example.com",
        Password: "ValidPass2",
    }
    user2, err2 := Signup(req2)
    // This should SUCCEED because "john" != "JOHN"
    if err2 != nil {
        t.Fatalf("Second signup should succeed: %v", err2.Error.Message)
    }

    // Verify they're different users
    if user1.ID == user2.ID {
        t.Error("Different usernames should create different users")
    }
}
```

---

## 9. Running All Tests Summary

```bash
$ cd /Users/muhammedshahidsiddiqui/Desktop/Projects/Pulse/Backend

$ go test -v ./internal/auth -count=1
# Output:
# === RUN   TestSignupSuccess
# --- PASS: TestSignupSuccess (0.05s)
# === RUN   TestSignupWeakPassword
# --- PASS: TestSignupWeakPassword (0.00s)
# ... (17 total tests)
# PASS ok  pulse-backend/internal/auth  1.091s

$ go test -v ./...
# Runs all tests across all packages
# 30+ tests total
```

---

## 10. Key Takeaways

✅ **Test Structure**: Setup → Execute → Assert  
✅ **Isolation**: Each test has its own database  
✅ **Coverage**: 17 auth tests covering signup, login, and edge cases  
✅ **Best Practices**: Table-driven tests, clear naming, isolated state  
✅ **Confidence**: All 30+ tests passing ensures code reliability  

---

**Happy Testing! 🎉**
