package chat

import "errors"

var (
	ErrCreateQuery      = errors.New("error create query to database")
	ErrInvalidArguments = errors.New("error invalid(-s) arguments")
)
