# Testing Guide - Pulse Backend

## Overview
This backend is built with modular Go packages with comprehensive test coverage. Tests are organized by module for easy maintenance and execution.

## Test Structure

```
Backend/
├── internal/
│   ├── crypto/crypto_test.go       # Cryptography tests (salt, token, password hashing)
│   └── auth/service_test.go          # Authentication service tests (signup, login)
└── cmd/server/main_test.go           # Integration tests (HTTP endpoints)
```

## Running Tests

### Run All Tests
```bash
$ make test
```
or
```bash
$ go test -v ./...
```

### Run Tests for Specific Module
```bash
$ make test-module MOD=crypto
$ make test-module MOD=auth
```
or
```bash
$ go test -v ./internal/crypto
$ go test -v ./internal/auth
```

### Run Integration Tests Only
```bash
$ go test -v ./cmd/server
```

### Run Tests with Coverage Report
```bash
$ make test-coverage
```
Generated file: `coverage.html` (open in browser)

### Run Tests in Verbose Mode with Race Detection
```bash
$ make test-verbose
```

## Test Coverage Summary

### Crypto Module (`internal/crypto/crypto_test.go`)
- ✅ `TestHashPassword` - PBKDF2 hashing with salt
- ✅ `TestVerifyPassword` - Password verification and case sensitivity
- ✅ `TestGenerateSalt` - Random salt generation (16 bytes)
- ✅ `TestGenerateToken` - Random token generation (32 bytes)
- ✅ `TestValidatePasswordStrength` - Password requirement validation
- ✅ `TestGenerateID` - Unique ID generation
- ✅ `TestPasswordCaseSensitivity` - Case-sensitive password validation

### Auth Service Module (`internal/auth/service_test.go`)
- ✅ `TestSignupSuccess` - User registration flow
- ✅ `TestSignupWeakPassword` - Weak password rejection
- ✅ `TestSignupDuplicate` - Duplicate user detection
- ✅ `TestLoginSuccess` - Login with token generation
- ✅ `TestLoginInvalidCredentials` - Invalid password handling
- ✅ `TestLoginInvalidEmail` - Non-existent user handling
- ✅ `TestLoginPasswordCaseSensitivity` - Case sensitivity in login

### Integration Tests (`cmd/server/main_test.go`)
- ✅ `TestSignupIntegration` - Full signup HTTP flow
- ✅ `TestLoginIntegration` - Full login HTTP flow
- ✅ `TestHealthCheck` - Health endpoint verification
- ✅ `TestGetUsers` - Users list retrieval
- ✅ `TestInvalidPassword` - Invalid password in signup
- ✅ `TestDuplicateSignup` - Duplicate user registration

## Test Database

Each test creates its own temporary database file (`test_persons.json`) to ensure isolation. Files are automatically cleaned up after tests complete.

## Building and Running

### Build
```bash
$ make build
```
Creates executable: `./pulse-backend`

### Run
```bash
$ make run
```
Starts server on `http://localhost:3000`

### Manual Build and Run
```bash
$ go build -o pulse-backend ./cmd/server
$ ./pulse-backend
```

## Password Requirements (Validated in Tests)
- Minimum 8 characters
- At least one uppercase letter
- At least one lowercase letter  
- At least one number
- Case-sensitive comparison

## Key Test Helpers

### setupTestDB(t *testing.T) string
Creates a temporary test database and initializes the db module. Returns path for cleanup.

### Example Test Pattern
```go
func TestExample(t *testing.T) {
    dbFile := setupTestDB(t)
    defer os.Remove(dbFile)
    
    // Test code here
}
```

## Continuous Integration Tips

1. **Before committing**: Run full test suite
   ```bash
   $ make test-coverage
   ```

2. **For specific features**: Run module tests
   ```bash
   $ make test-module MOD=crypto
   ```

3. **With race detection**: Run verbose mode
   ```bash
   $ make test-verbose
   ```

## Troubleshooting

### Tests fail with "database not initialized"
Ensure `setup()` is called in tests that access database.

### Import errors
Run `go mod tidy` to ensure all dependencies are resolved.

### Port already in use
Change port in `cmd/server/main.go` if 3000 is occupied.
