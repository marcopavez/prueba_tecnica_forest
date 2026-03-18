package auth

import (
	"encoding/base64"
	"strings"
)

type BasicAuth struct {
	username string
	password string
}

func NewBasicAuth(adminCredentials string) *BasicAuth {
	decoded, err := base64.StdEncoding.DecodeString(adminCredentials)
	if err != nil {
		return &BasicAuth{}
	}
	parts := strings.SplitN(string(decoded), ":", 2)
	if len(parts) != 2 {
		return &BasicAuth{}
	}
	return &BasicAuth{username: parts[0], password: parts[1]}
}

func (b *BasicAuth) Validate(username, password string) bool {
	return username == b.username && password == b.password
}
