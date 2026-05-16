package middleware

import (
	"pooky-messanger/pkg/jwt"
)

type middleware struct {
	t jwt.JWTService
}

func NewMiddleware(t jwt.JWTService) middleware {
	return middleware{t: t}
}
