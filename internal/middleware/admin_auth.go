package middleware

import (
	"net/http"

	"github.com/bikerental/api/internal/auth"
)

type AdminAuth struct {
	basic *auth.BasicAuth
}

func NewAdminAuth(b *auth.BasicAuth) *AdminAuth {
	return &AdminAuth{basic: b}
}

func (a *AdminAuth) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username, password, ok := r.BasicAuth()
		if !ok || !a.basic.Validate(username, password) {
			w.Header().Set("WWW-Authenticate", `Basic realm="admin"`)
			http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}
