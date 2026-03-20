package crypto

import (
	"testing"
)

func TestHashPassword(t *testing.T) {
	salt, err := GenerateSalt()
	if err != nil {
		t.Fatalf("GenerateSalt failed: %v", err)
	}

	password := "TestPass123"
	hash1, err := HashPassword(password, salt)
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}
	if hash1 == "" {
		t.Error("hash should not be empty")
	}

	// Same password and salt should produce same hash (deterministic)
	hash2, err := HashPassword(password, salt)
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}
	if hash1 != hash2 {
		t.Error("same password and salt should produce same hash")
	}
}

func TestVerifyPassword(t *testing.T) {
	salt, _ := GenerateSalt()
	password := "TestPass123"
	hash, _ := HashPassword(password, salt)

	// Correct password should verify
	valid, err := VerifyPassword(password, hash, salt)
	if err != nil {
		t.Fatalf("VerifyPassword failed: %v", err)
	}
	if !valid {
		t.Error("correct password should verify")
	}

	// Incorrect password should not verify
	valid, _ = VerifyPassword("WrongPass123", hash, salt)
	if valid {
		t.Error("incorrect password should not verify")
	}

	// Case-sensitive - different case should not verify
	valid, _ = VerifyPassword("testpass123", hash, salt)
	if valid {
		t.Error("password should be case-sensitive")
	}
}

func TestGenerateSalt(t *testing.T) {
	salt1, err := GenerateSalt()
	if err != nil {
		t.Fatalf("GenerateSalt failed: %v", err)
	}
	if salt1 == "" {
		t.Error("salt should not be empty")
	}
	if len(salt1) != 32 { // 16 bytes / 2 for hex encoding
		t.Errorf("salt hex should be 32 chars, got %d", len(salt1))
	}

	// Different salts should be unique
	salt2, _ := GenerateSalt()
	if salt1 == salt2 {
		t.Error("salts should be unique")
	}
}

func TestGenerateToken(t *testing.T) {
	token1, err := GenerateToken()
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}
	if token1 == "" {
		t.Error("token should not be empty")
	}
	if len(token1) != 64 { // 32 bytes / 2 for hex encoding
		t.Errorf("token hex should be 64 chars, got %d", len(token1))
	}

	// Different tokens should be unique
	token2, _ := GenerateToken()
	if token1 == token2 {
		t.Error("tokens should be unique")
	}
}

func TestValidatePasswordStrength(t *testing.T) {
	tests := []struct {
		name       string
		password   string
		shouldPass bool
	}{
		{"Valid password", "TestPass123", true},
		{"Missing uppercase", "testpass123", false},
		{"Missing lowercase", "TESTPASS123", false},
		{"Missing number", "TestPass", false},
		{"Too short", "Test12", false},
		{"All requirements", "Secure123Pass", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidatePasswordStrength(tt.password)
			if result.IsValid != tt.shouldPass {
				t.Errorf("expected IsValid=%v, got %v", tt.shouldPass, result.IsValid)
			}
			if tt.shouldPass && len(result.Errors) > 0 {
				t.Errorf("expected no errors, got %v", result.Errors)
			}
			if !tt.shouldPass && len(result.Errors) == 0 {
				t.Error("expected errors but got none")
			}
		})
	}
}

func TestGenerateID(t *testing.T) {
	id1 := GenerateID()
	if id1 == "" {
		t.Error("ID should not be empty")
	}

	id2 := GenerateID()
	if id1 == id2 {
		t.Error("IDs should be unique")
	}
}

func TestPasswordCaseSensitivity(t *testing.T) {
	salt, _ := GenerateSalt()
	password := "TeSt123Pass"
	hash, _ := HashPassword(password, salt)

	// Test various case mutations
	caseVariations := []string{
		"test123pass",
		"TEST123PASS",
		"TeSt123PASS",
		"teSt123pASS",
	}

	for _, variation := range caseVariations {
		valid, _ := VerifyPassword(variation, hash, salt)
		if valid {
			t.Errorf("password should not match case variation: %s", variation)
		}
	}
}
