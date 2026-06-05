package security_test

import (
	"testing"

	"golang.org/x/crypto/bcrypt"

	"github.com/StepanKomis/Ticketa/src/internal/security"
)

func TestHashPassword_EmptyString(t *testing.T) {
	_, err := security.HashPassword("")
	if err == nil {
		t.Fatal("expected error for empty password, got nil")
	}
}

func TestHashPassword_ReturnsNonEmptyHash(t *testing.T) {
	hash, err := security.HashPassword("Secret1!")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if hash == "" {
		t.Fatal("expected non-empty hash")
	}
}

func TestHashPassword_ProducesBcryptHash(t *testing.T) {
	raw := "Secret1!"
	hash, err := security.HashPassword(raw)
	if err != nil {
		t.Fatalf("HashPassword: %v", err)
	}
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(raw)); err != nil {
		t.Errorf("hash is not verifiable by bcrypt: %v", err)
	}
}

func TestHashPassword_UniquePerCall(t *testing.T) {
	h1, _ := security.HashPassword("Secret1!")
	h2, _ := security.HashPassword("Secret1!")
	if h1 == h2 {
		t.Error("expected different hashes for the same input (bcrypt salting)")
	}
}

func TestCheckPassword_CorrectPassword(t *testing.T) {
	raw := "Secret1!"
	hash, _ := security.HashPassword(raw)
	if err := security.CheckPassword(raw, hash); err != nil {
		t.Errorf("expected nil for correct password, got: %v", err)
	}
}

func TestCheckPassword_WrongPassword(t *testing.T) {
	hash, _ := security.HashPassword("Secret1!")
	if err := security.CheckPassword("Wrong1!", hash); err == nil {
		t.Fatal("expected error for wrong password, got nil")
	}
}

func TestCheckPassword_InvalidHash(t *testing.T) {
	if err := security.CheckPassword("Secret1!", "not-a-bcrypt-hash"); err == nil {
		t.Fatal("expected error for invalid hash, got nil")
	}
}
