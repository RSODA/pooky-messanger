package jwt

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func (t *jwtService) GetToken(id string) (*string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(t.ttl)),
		Subject:   id,
	})

	tokenString, err := token.SignedString([]byte(t.secret))
	if err != nil {
		return nil, err
	}

	return &tokenString, nil
}
