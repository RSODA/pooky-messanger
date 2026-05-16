package user

import "errors"

var (
	ErrCreateQuery  = errors.New("error creating query")
	ErrUserNotFound = errors.New("user not found")
	ErrInvalidToken = errors.New("invalid token")
	ErrUpdateAvatar = errors.New("error updating avatar")
)
