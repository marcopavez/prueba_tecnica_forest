package auth_test

import (
	"encoding/base64"
	"testing"

	"github.com/bikerental/api/internal/auth"
)

func TestBasicAuthValidate(t *testing.T) {
	encoded := base64.StdEncoding.EncodeToString([]byte("admin:password"))
	b := auth.NewBasicAuth(encoded)

	if !b.Validate("admin", "password") {
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
	b := auth.NewBasicAuth("!!!not-base64!!!")
	if b.Validate("admin", "password") {
		t.Error("expected invalid base64 to fail all validations")
	}
}
