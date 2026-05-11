package config

import (
	"errors"
	"os"
	"strconv"
	"time"
)

type TokenConfig interface {
	GetDuration() time.Duration
	GetSecret() string
}

type tokenConfig struct {
	secret   string
	duration time.Duration
}

func NewTokenService() (TokenConfig, error) {
	secret := os.Getenv("JWT_SECRET")
	if len(secret) == 0 {
		return nil, errors.New("TOKEN_SECRET environment variable not set")
	}

	duration := os.Getenv("JWT_TTL")
	if len(duration) == 0 {
		return nil, errors.New("JWT_TTL environment variable not set")
	}

	durationInt, err := strconv.Atoi(duration)
	if err != nil {
		return nil, errors.New("JWT_TTL environment variable not set")
	}

	return &tokenConfig{
		secret:   secret,
		duration: time.Duration(durationInt) * time.Minute,
	}, nil
}

func (t *tokenConfig) GetDuration() time.Duration {
	return t.duration
}

func (t *tokenConfig) GetSecret() string {
	return t.secret
}
