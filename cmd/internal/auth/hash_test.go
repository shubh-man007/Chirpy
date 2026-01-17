package auth

import (
	"testing"
)

func TestHashPasswordSuccess(t *testing.T) {
	password := "supersecret123"

	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if hash == "" {
		t.Fatalf("expected hash to be non-empty")
	}

	if hash == password {
		t.Fatalf("hash should not equal original password")
	}
}

func TestCheckPasswordHashCorrect(t *testing.T) {
	password := "mypassword"

	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}

	match, err := CheckPasswordHash(password, hash)
	if err != nil {
		t.Fatalf("expected no error comparing hash: %v", err)
	}

	if !match {
		t.Fatalf("expected password to match hash, but it did not")
	}
}

func TestCheckPasswordHashIncorrect(t *testing.T) {
	password := "correct"
	wrongPassword := "wrongpass"

	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}

	match, err := CheckPasswordHash(wrongPassword, hash)
	if err != nil {
		t.Fatalf("expected no error comparing hash: %v", err)
	}

	if match {
		t.Fatalf("expected password NOT to match hash")
	}
}

func TestCheckPasswordHashInvalidHash(t *testing.T) {
	password := "something"
	badHash := "this-is-not-a-valid-argon2-hash"

	match, err := CheckPasswordHash(password, badHash)
	if err == nil {
		t.Fatalf("expected error for invalid hash format, got none")
	}

	if match {
		t.Fatalf("expected match to be false for invalid hash")
	}
}
