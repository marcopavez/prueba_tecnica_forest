package auth_test

import (
	"testing"

	"github.com/bikerental/api/internal/auth"
)

func TestJWTGenerateAndValidate(t *testing.T) {
	j := auth.NewJWTAuth()

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

	if claims.Subject != 42 {
		t.Errorf("expected subject '42', got '%d'", claims.Subject)
	}
	if claims.Email != "test@example.com" {
		t.Errorf("expected email 'test@example.com', got '%s'", claims.Email)
	}
	if claims.FirstName != "John" {
		t.Errorf("expected first_name 'John', got '%s'", claims.FirstName)
	}
}
