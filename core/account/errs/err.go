package errs

import "errors"

var (
	// ErrUserNotFound user not found error
	ErrUserNotFound = errors.New("user not found")
)
