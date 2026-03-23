package auth_test

import (
	"testing"

	"github.com/bikerental/api/internal/auth"
)

func TestBasicAuthValidate(t *testing.T) {
	b := auth.NewBasicAuth()

	if b.Validate("admin", "password") {
		t.Error("expected valid credentials to pass")
	}
	if b.Validate("admin", "wrong") {
		t.Error("expected wrong password to fail")
	}
	if b.Validate("other", "password") {
		t.Error("expected wrong username to fail")
	}
}

func TestBasicAuthInvalidBase64(t *testing.T) {
	b := auth.NewBasicAuth()
	if b.Validate("admin", "password") {
		t.Error("expected invalid base64 to fail all validations")
	}
}
