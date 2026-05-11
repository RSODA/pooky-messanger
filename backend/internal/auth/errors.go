package auth

import "errors"

var (
	ErrCreateQuery       = errors.New("failed to create query")
	ErrInsertToDB        = errors.New("failed to insert into user")
	ErrGetFromDB         = errors.New("failed to get user")
	ErrInvalidParams     = errors.New("invalid params")
	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrLoginUser         = errors.New("login user error")
)
