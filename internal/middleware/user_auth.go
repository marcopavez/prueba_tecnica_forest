package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/bikerental/api/internal/auth"
)

type contextKey string

const UserIDKey contextKey = "id"

type UserAuth struct {
	jwt *auth.JWTAuth
}

func NewUserAuth(j *auth.JWTAuth) *UserAuth {
	return &UserAuth{jwt: j}
}

func (u *UserAuth) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			http.Error(w, `{"error":"missing or invalid authorization header"}`, http.StatusUnauthorized)
			return
		}
		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		claims, err := u.jwt.Validate(tokenStr)
		if err != nil {
			http.Error(w, `{"error":"invalid or expired token"}`, http.StatusUnauthorized)
			return
		}

		if claims.Subject == 0 {
			http.Error(w, `{"error":"invalid token subject"}`, http.StatusUnauthorized)
			return
		}
		ctx := context.WithValue(r.Context(), UserIDKey, claims.Subject)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetUserID(r *http.Request) int64 {
	v, _ := r.Context().Value(UserIDKey).(int64)
	return v
}
