package auther

import (
	"time"

	"github.com/golang-jwt/jwt/v4"
)

var (
	secretInstance string
)

func SetSecret(secret string) {
	secretInstance = secret
}

func GenerateJWTWithEmail(email string) (string, error) {
	claims := jwt.MapClaims{
		"email": email,
		"exp":   time.Now().Add(time.Hour * 72).Unix(),
		"iat":   time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString([]byte(secretInstance))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}
