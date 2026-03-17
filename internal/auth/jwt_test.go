package auth_test

import (
	"testing"

	"github.com/bikerental/api/internal/auth"
)

func TestJWTGenerateAndValidate(t *testing.T) {
	j := auth.NewJWTAuth("test-secret")

	token, err := j.Generate(42, "test@example.com", "John", "Doe")
	if err != nil {
		t.Fatalf("expected no error generating token, got: %v", err)
	}
	if token == "" {
		t.Fatal("expected non-empty token")
	}

	claims, err := j.Validate(token)
	if err != nil {
		t.Fatalf("expected no error validating token, got: %v", err)
	}

	if claims.Subject != "42" {
		t.Errorf("expected subject '42', got '%s'", claims.Subject)
	}
	if claims.Email != "test@example.com" {
		t.Errorf("expected email 'test@example.com', got '%s'", claims.Email)
	}
	if claims.FirstName != "John" {
		t.Errorf("expected first_name 'John', got '%s'", claims.FirstName)
	}
}

func TestJWTInvalidToken(t *testing.T) {
	j := auth.NewJWTAuth("test-secret")
	_, err := j.Validate("not.a.valid.token")
	if err == nil {
		t.Fatal("expected error for invalid token, got nil")
	}
}

func TestJWTWrongSecret(t *testing.T) {
	j1 := auth.NewJWTAuth("secret-a")
	j2 := auth.NewJWTAuth("secret-b")

	token, _ := j1.Generate(1, "a@b.com", "A", "B")
	_, err := j2.Validate(token)
	if err == nil {
		t.Fatal("expected error when validating with wrong secret")
	}
}
