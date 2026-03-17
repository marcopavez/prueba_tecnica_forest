package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWTAuth struct {
	secret []byte
}

type Claims struct {
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	jwt.RegisteredClaims
}

func NewJWTAuth(secret string) *JWTAuth {
	return &JWTAuth{secret: []byte(secret)}
}

func (j *JWTAuth) Generate(userID int64, email, firstName, lastName string) (string, error) {
	claims := Claims{
		Email:     email,
		FirstName: firstName,
		LastName:  lastName,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   fmt.Sprintf("%d", userID),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(30 * 24 * time.Hour)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.secret)
}

func (j *JWTAuth) Validate(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return j.secret, nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}
	return claims, nil
}
