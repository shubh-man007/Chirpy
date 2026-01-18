package auth

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestMakeJWT(t *testing.T) {
	userID := uuid.New()
	secret := "test-secret"
	expiresIn := time.Hour

	token, err := MakeJWT(userID, secret, expiresIn)
	if err != nil {
		t.Fatalf("MakeJWT failed: %v", err)
	}

	if token == "" {
		t.Error("Expected non-empty token")
	}

	// Token should be a valid JWT format (3 parts separated by dots)
	// This is a basic check - proper validation happens in ValidateJWT tests
	if len(token) < 10 {
		t.Error("Token seems too short to be valid")
	}
}

func TestValidateJWT_Success(t *testing.T) {
	userID := uuid.New()
	secret := "test-secret"
	expiresIn := time.Hour

	// Create a valid token
	token, err := MakeJWT(userID, secret, expiresIn)
	if err != nil {
		t.Fatalf("MakeJWT failed: %v", err)
	}

	// Validate the token
	parsedUserID, err := ValidateJWT(token, secret)
	if err != nil {
		t.Fatalf("ValidateJWT failed: %v", err)
	}

	// Check that we get back the same user ID
	if parsedUserID != userID {
		t.Errorf("Expected user ID %v, got %v", userID, parsedUserID)
	}
}

func TestValidateJWT_ExpiredToken(t *testing.T) {
	userID := uuid.New()
	secret := "test-secret"
	expiresIn := -time.Hour // Token expired 1 hour ago

	// Create an already-expired token
	token, err := MakeJWT(userID, secret, expiresIn)
	if err != nil {
		t.Fatalf("MakeJWT failed: %v", err)
	}

	// Try to validate the expired token
	_, err = ValidateJWT(token, secret)
	if err == nil {
		t.Error("Expected error for expired token, got nil")
	}
}

func TestValidateJWT_WrongSecret(t *testing.T) {
	userID := uuid.New()
	correctSecret := "correct-secret"
	wrongSecret := "wrong-secret"
	expiresIn := time.Hour

	// Create token with correct secret
	token, err := MakeJWT(userID, correctSecret, expiresIn)
	if err != nil {
		t.Fatalf("MakeJWT failed: %v", err)
	}

	// Try to validate with wrong secret
	_, err = ValidateJWT(token, wrongSecret)
	if err == nil {
		t.Error("Expected error for wrong secret, got nil")
	}
}

func TestValidateJWT_InvalidToken(t *testing.T) {
	secret := "test-secret"
	invalidToken := "this.is.invalid"

	_, err := ValidateJWT(invalidToken, secret)
	if err == nil {
		t.Error("Expected error for invalid token format, got nil")
	}
}

func TestValidateJWT_EmptyToken(t *testing.T) {
	secret := "test-secret"
	emptyToken := ""

	_, err := ValidateJWT(emptyToken, secret)
	if err == nil {
		t.Error("Expected error for empty token, got nil")
	}
}

func TestValidateJWT_MalformedToken(t *testing.T) {
	secret := "test-secret"
	malformedToken := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.invalid.signature"

	_, err := ValidateJWT(malformedToken, secret)
	if err == nil {
		t.Error("Expected error for malformed token, got nil")
	}
}

func TestMakeJWT_ShortExpiration(t *testing.T) {
	userID := uuid.New()
	secret := "test-secret"
	expiresIn := time.Millisecond * 100

	token, err := MakeJWT(userID, secret, expiresIn)
	if err != nil {
		t.Fatalf("MakeJWT failed: %v", err)
	}

	// Wait for token to expire
	time.Sleep(time.Millisecond * 200)

	// Token should be expired now
	_, err = ValidateJWT(token, secret)
	if err == nil {
		t.Error("Expected error for expired token, got nil")
	}
}

func TestMakeJWT_DifferentUserIDs(t *testing.T) {
	secret := "test-secret"
	expiresIn := time.Hour

	userID1 := uuid.New()
	userID2 := uuid.New()

	token1, err := MakeJWT(userID1, secret, expiresIn)
	if err != nil {
		t.Fatalf("MakeJWT failed for user 1: %v", err)
	}

	token2, err := MakeJWT(userID2, secret, expiresIn)
	if err != nil {
		t.Fatalf("MakeJWT failed for user 2: %v", err)
	}

	// Tokens should be different
	if token1 == token2 {
		t.Error("Expected different tokens for different users")
	}

	// Validate both tokens return correct user IDs
	parsedID1, err := ValidateJWT(token1, secret)
	if err != nil {
		t.Fatalf("ValidateJWT failed for token1: %v", err)
	}
	if parsedID1 != userID1 {
		t.Errorf("Expected user ID %v, got %v", userID1, parsedID1)
	}

	parsedID2, err := ValidateJWT(token2, secret)
	if err != nil {
		t.Fatalf("ValidateJWT failed for token2: %v", err)
	}
	if parsedID2 != userID2 {
		t.Errorf("Expected user ID %v, got %v", userID2, parsedID2)
	}
}
