package authentication

import (
	"github.com/dgrijalva/jwt-go"
	"time"
)

func IssueJWT(secret string, email string, expires time.Time) (string, error) {
	key := []byte(secret)

	// Create the Claims
	claims := &jwt.StandardClaims{
		ExpiresAt: expires.Unix(),
		Subject:   email,
		Issuer:    "Helios",
		IssuedAt:  time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(key)
}
