package auth

import (
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWTAuth struct {
	secret []byte
}

type Claims struct {
	Subject   int64            `json:"sub"`
	ExpiresAt *jwt.NumericDate `json:"exp"`
	Email     string           `json:"email"`
	FirstName string           `json:"first_name"`
	LastName  string           `json:"last_name"`
	jwt.RegisteredClaims
}

func NewJWTAuth() *JWTAuth {

	jwtSecret := os.Getenv("JWT_SECRET")

	return &JWTAuth{secret: []byte(jwtSecret)}
}

func (j *JWTAuth) Generate(userID int64, email, firstName, lastName string) (string, error) {

	claims := Claims{
		Subject:   userID,
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(30 * 24 * time.Hour)),
		Email:     email,
		FirstName: firstName,
		LastName:  lastName,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signedToken, err := token.SignedString(j.secret)
	if err != nil {
		return "", err
	}

	return signedToken, nil
}

func (j *JWTAuth) Validate(tokenString string) (*Claims, error) {
	parsedToken, err := jwt.ParseWithClaims(tokenString, &Claims{}, j.provideSigningKey)
	if err != nil {
		return nil, err
	}

	claims, err := j.extractClaims(parsedToken)
	if err != nil {
		return nil, err
	}

	return claims, nil
}

func (j *JWTAuth) provideSigningKey(token *jwt.Token) (interface{}, error) {
	_, isHMAC := token.Method.(*jwt.SigningMethodHMAC)
	if !isHMAC {
		return nil, fmt.Errorf("algoritmo de firma inesperado: %v", token.Header["alg"])
	}
	return j.secret, nil
}

func (j *JWTAuth) extractClaims(parsedToken *jwt.Token) (*Claims, error) {
	claims, ok := parsedToken.Claims.(*Claims)
	if !ok || !parsedToken.Valid {
		return nil, fmt.Errorf("token inválido o claims malformados")
	}
	return claims, nil
}
