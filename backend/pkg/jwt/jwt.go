package jwt

import "time"

type JWTService interface {
	GetToken(id string) (*string, error)
	ValidateToken(tokenString string) (string, error)
}

type jwtService struct {
	secret string
	ttl    time.Duration
}

func NewGetTokenRequest(secret string, time time.Duration) JWTService {
	return &jwtService{secret, time}
}
